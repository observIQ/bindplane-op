// Copyright  observIQ, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"io"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/cli/printer"
)

// BindPlane is an instance of the BindPlane command line client
type BindPlane struct {
	Config *common.Config

	// ProfileName represents the name of the current profile in use. If empty, either a --config flag is being used or
	// there is no current profile name. This is set during initialization in root.go.
	ProfileName string

	// ConfigFile is the path to the configuration file in use. This is either the value of the --config flag or the
	// config file based on the current profile. It will be empty if there is no --config flag or current profile.
	ConfigFile string

	client client.BindPlane

	writer      io.Writer
	printer     printer.Printer
	initPrinter sync.Once

	logger     *zap.Logger
	initLogger sync.Once

	shutdownHooks []ShutdownHook
}

// NewBindPlane TODO(doc)
func NewBindPlane(config *common.Config, writer io.Writer) *BindPlane {
	return &BindPlane{
		Config:        config,
		writer:        writer,
		shutdownHooks: []ShutdownHook{},
	}
}

// NewBindPlaneForTesting is primarily used for testing and sets an empty config, the writer to Stdout, and a nop logger.
func NewBindPlaneForTesting() *BindPlane {
	return &BindPlane{
		Config:        common.InitConfig(""),
		writer:        os.Stdout,
		shutdownHooks: []ShutdownHook{},
		logger:        zap.NewNop(),
	}
}

// Client initializes and returns the client. The client will be initialized once.
func (i *BindPlane) Client() (client.BindPlane, error) {
	// don't override a client provided to SetClient
	if i.client == nil {
		var err error
		i.client, err = client.NewBindPlane(&i.Config.Client, i.Logger(), i.Config.Server.BindPlaneURL())
		if err != nil {
			return nil, err
		}
	}

	return i.client, nil
}

// SetClient will set the client used by the BindPlane client. Primarily used for testing.
func (i *BindPlane) SetClient(client client.BindPlane) {
	i.client = client
}

// LogLevel returns the zapcore.Level that will used to initialize the logger. For Production, we use zapcore.InfoLevel.
// Test and Development both use zapcore.DebugLevel.
func (i *BindPlane) LogLevel() zapcore.Level {
	// TODO(andy): add a --log-level flag
	switch i.Config.Server.BindPlaneEnv() {
	case common.EnvProduction:
		return zapcore.InfoLevel
	default:
		return zapcore.DebugLevel
	}
}

// Logger returns a logger that can be used for logging
func (i *BindPlane) Logger() *zap.Logger {
	i.initLogger.Do(func() {
		// we delay the initialization of the logger until it is requested because the Config is not currently fully loaded
		// when NewBindPlane is called because flags have not yet been processed
		if i.logger == nil {
			logger, err := common.NewLogger(i.Config.Client.Common, i.LogLevel())
			if err != nil {
				panic(fmt.Errorf("failed to initialize logger: %w", err))
			}
			i.logger = logger
		}
	})
	return i.logger
}

// Printer initializes and returns the printer. The printer will be initialized once.
func (i *BindPlane) Printer() printer.Printer {
	i.initPrinter.Do(func() {
		if i.printer == nil {
			switch i.Config.Output {
			case "json":
				i.printer = printer.NewJSONPrinter(i.writer, i.Logger())
			case "yaml":
				i.printer = printer.NewYamlPrinter(i.writer, i.Logger())
			case "table":
				fallthrough
			default:
				i.printer = printer.NewTablePrinter(i.writer)
			}
		}
	})
	return i.printer
}

// ShutdownHook is called at shutdown when added using AddShutdownHook
type ShutdownHook func()

// Shutdown will be called when the command completes
func (i *BindPlane) Shutdown() {
	for _, hook := range i.shutdownHooks {
		hook()
	}
}

// AddShutdownHook adds a hook to be called at shutdown. It is expected to be used during initialization and is not safe
// to call from multiple goroutines. Shutdown hooks will be called on shutdown in the order that they were added (i.e. FIFO).
func (i *BindPlane) AddShutdownHook(hook ShutdownHook) {
	i.shutdownHooks = append(i.shutdownHooks, hook)
}
