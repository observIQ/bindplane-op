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

package get

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/internal/cli/printer"
)

// ConfigurationsCommand returns the BindPlane get configurations cobra command
func ConfigurationsCommand(bindplane *cli.BindPlane) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configurations",
		Aliases: []string{"configuration", "configs", "config"},
		Short:   "Displays the configurations",
		Long:    "A configuration provides a complete agent configuration to ship logs metrics, and traces",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := bindplane.Client()
			if err != nil {
				return fmt.Errorf("error creating client: %w", err)
			}

			if len(args) > 0 {
				name := args[0]
				configuration, err := c.Configuration(cmd.Context(), name)
				if err != nil {
					return err
				}
				if configuration == nil {
					return fmt.Errorf("no Configuration found with name %s", name)
				}

				if bindplane.Config.Output == "raw" {
					raw, err := c.RawConfiguration(cmd.Context(), name)
					if err != nil {
						return err
					}
					if _, err := cmd.OutOrStdout().Write([]byte(raw)); err != nil {
						return err
					}
					return nil
				}

				printer.PrintResource(bindplane.Printer(), configuration)
				return nil
			}

			configurations, err := c.Configurations(cmd.Context())
			if err != nil {
				return err
			}

			printer.PrintResources(bindplane.Printer(), configurations)
			return nil
		},
	}

	return cmd
}
