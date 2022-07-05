// Copyright  observIQ, Inc.
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

package commands

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.uber.org/zap"
	"google.golang.org/api/option"

	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/internal/cli/commands/profile"
	"github.com/observiq/bindplane-op/internal/cli/flags"
	v "github.com/observiq/bindplane-op/internal/version"
)

// BindplaneHome returns the value of the homeArg, BINDPLANE_CONFIG_HOME,
// or a default of $HOME/.bindplane
func BindplaneHome() string {
	if homeEnv, ok := os.LookupEnv("BINDPLANE_CONFIG_HOME"); ok {
		return homeEnv
	}
	return common.DefaultBindPlaneHomePath()
}

// Command is the root command that represents the base command, in this function we add persistent flags,
// and bind them to viper.
// The persistent pre run function here is where we read the profile file and set the
// values for bindplane.Config
func Command(bindplane *cli.BindPlane, name string) *cobra.Command {
	var configArg string
	var profileArg string

	cmd := &cobra.Command{
		Use:   name,
		Short: "Next generation agent management platform",
		// cobra.CheckErr will print the returned error with exit status,
		// so we disable errors on this and child commands so error message isn't repeated
		SilenceErrors: true,
		// This will prevent child commands from printing the help message on error.
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := readConfig(bindplane, configArg, profileArg)
			if err != nil {
				return fmt.Errorf("error while trying to read configuration, %w", err)
			}

			err = initViper(bindplane.Config)
			if err != nil {
				return fmt.Errorf("error while trying to unmarshal configuration, %w", err)
			}

			initTracing(bindplane)
			return nil
		},
	}

	// Set global flags
	flags.Global(cmd)

	// These flags are command line only
	cmd.PersistentFlags().StringVarP(&configArg, "config", "c", "", "full path to configuration file")
	cmd.PersistentFlags().StringVar(&profileArg, "profile", "", "configuration profile name to use")

	return cmd
}

// readConfig reads configuration files from a .yaml file and merges those values with flags and environment variables.
// We don't treat a missing config file to be an error.
func readConfig(bindplane *cli.BindPlane, configFlagValue string, profileFlagValue string) error {
	f := profile.NewHelper(BindplaneHome()).Folder()

	configFile, err := configFilePath(f, configFlagValue, profileFlagValue)
	if err != nil {
		return err
	}
	if configFile != "" {
		viper.SetConfigFile(configFile)

		// Read values from file, its okay if there is no config file found
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("error reading in configuration file: %s, %w", viper.GetViper().ConfigFileUsed(), err)
		}
	}

	// set the ConfigFile and ProfileName so that commands can reference them
	bindplane.ConfigFile = configFile
	if configFlagValue == "" {
		if profileFlagValue != "" {
			bindplane.ProfileName = profileFlagValue
		} else {
			bindplane.ProfileName, _ = f.CurrentProfileName()
		}
	}

	return nil
}

// This does a couple things:
// 1.  If the --config flag is set we assume its a full filepath to a configuration file and use that directly.
// 2.  If the --profile flag is set we look for a profile specified in ~/.bindplane/profiles/[name].yaml
// 3.  If neither flag is not set we look for a current profile specified in ~/.bindplane/profiles/current
// Note: ~/.bindplane/profiles/ is changed by setting BINDPLANE_CONFIG_HOME and will use $BINDPLANE_CONFIG_HOME/profiles/
func configFilePath(f profile.Folder, configFlagValue string, profileFlagValue string) (string, error) {
	// --config [path]
	if configFlagValue != "" {
		// Assume its a full filepath to a users config file
		return configFlagValue, nil
	}

	// --profile [name]
	if profileFlagValue != "" {
		// ensure the profile exists before using it
		if !f.ProfileExists(profileFlagValue) {
			return "", fmt.Errorf("no profile found with name '%s'", profileFlagValue)
		}
		return f.ProfilePath(profileFlagValue), nil
	}

	// use the current profile if there is one, ignore the error because we might not have a current profile
	profilePath, err := f.CurrentProfilePath()
	if err == nil {
		return profilePath, nil
	}

	// no configFilePath to use
	return "", nil
}

// Called before run time to populate the Config struct with values from viper
// We unmarshal twice here to get the squashed common values, then again
// to load in the nested values
func initViper(conf *common.Config) error {
	err := viper.Unmarshal(conf, func(dc *mapstructure.DecoderConfig) {
		dc.Squash = true
	})
	if err != nil {
		return err
	}
	return viper.Unmarshal(conf)
}

func initTracing(bindplane *cli.BindPlane) {
	config := bindplane.Config
	tracing := config.Server.GoogleCloudTracing
	if tracing == nil || !tracing.Enabled {
		return
	}
	bindplane.Logger().Info("Enabling tracing to Google Cloud")

	hostname, _ := os.Hostname()
	exporter, err := texporter.New(
		texporter.WithProjectID(tracing.ProjectID),
		texporter.WithTraceClientOptions([]option.ClientOption{
			option.WithCredentialsFile(tracing.CredentialsFile),
		}),
	)
	if err != nil {
		bindplane.Logger().Error("Failed to create trace exporter, tracing disabled", zap.Error(err))
		return
	}

	tp := trace.NewTracerProvider(
		// batch traces before sending
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("bindplane"),
			semconv.ServiceVersionKey.String(v.NewVersion().String()),
			semconv.HostArchKey.String(runtime.GOARCH),
			semconv.HostNameKey.String(hostname),
		)),
	)

	// cleanly shutdown and flush telemetry when the application exits.
	bindplane.AddShutdownHook(func() {
		bindplane.Logger().Info("flushing traces to Google Cloud before shutdown")

		// do not make the application hang when it is shutdown.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			bindplane.Logger().Error("error during shutdown of the otel trace provider", zap.Error(err))
		}
	})

	// set the TracerProvider as the global so that we can use it as needed
	otel.SetTracerProvider(tp)
}
