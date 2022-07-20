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

package common

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	registerWindowsSink()
}

// NewLogger returns a new Logger for the specified config and level
func NewLogger(config Common, level zapcore.Level) (*zap.Logger, error) {
	logPath := config.BindPlaneLogFilePath()
	if config.LogOutput == LogOutputStdout {
		return NewStdoutLogger(level)
	}
	return NewFileLogger(level, logPath)
}

// NewFileLogger takes a logging level and log file path and returns a zip.Logger
func NewFileLogger(level zapcore.Level, path string) (*zap.Logger, error) {
	writer := &lumberjack.Logger{
		Filename:   pathToURI(path),
		MaxSize:    100, // mb
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   true,
	}
	core := zapcore.NewCore(newEncoder(), zapcore.AddSync(writer), level)
	return zap.New(core), validatePath(path)
}

// NewStdoutLogger returns a new Logger with the specified level, writing to stdout
func NewStdoutLogger(level zapcore.Level) (*zap.Logger, error) {
	core := zapcore.NewCore(newEncoder(), zapcore.Lock(os.Stdout), level)
	return zap.New(core), nil
}

func newEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.CallerKey = ""
	encoderConfig.StacktraceKey = ""
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.MessageKey = "message"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

func pathToURI(path string) string {
	return pathToURIByOS(path, runtime.GOOS)
}

func pathToURIByOS(path, goos string) string {
	switch goos {
	case "windows":
		return "winfile:///" + filepath.ToSlash(path)
	default:
		return filepath.ToSlash(path)
	}
}

var registerSyncsOnce sync.Once

func registerWindowsSink() {
	registerSyncsOnce.Do(func() {
		if runtime.GOOS == "windows" {
			err := zap.RegisterSink("winfile", newWinFileSink)
			if err != nil {
				panic(err)
			}
		}
	})
}

func newWinFileSink(u *url.URL) (zap.Sink, error) {
	// Ensure permissions restrict access to the running user only
	return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
}

// validatePath attempts to create a temp file under the log
// directory.
func validatePath(p string) error {
	dir, _ := filepath.Split(p)

	// If directory is set, ensure it exists
	if dir != "" {
		if _, err := os.Stat(dir); err != nil {
			return fmt.Errorf("log directory: %w", err)
		}
	}

	// Create test file in directory
	f, err := os.CreateTemp(dir, "validate")
	if err != nil {
		return fmt.Errorf("log file creation: %w", err)
	}

	// Grab file path and close right away
	validationPath := f.Name()
	if err := f.Close(); err != nil {
		return fmt.Errorf("close log file %s: %w", validationPath, err)
	}

	if err := os.Remove(f.Name()); err != nil {
		return fmt.Errorf("cleanup log file %s: %w", validationPath, err)
	}

	return nil
}
