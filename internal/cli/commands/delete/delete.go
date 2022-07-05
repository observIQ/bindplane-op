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

package delete

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/model"
)

var file string

// Command returns the bindplane delete cobra command
func Command(bindplane *cli.BindPlane) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete bindplane resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := bindplane.Client()
			if err != nil {
				return fmt.Errorf("error creating client: %w", err)
			}

			if file == "" {
				// This doesn't appear to ever return an error.
				_ = cmd.Help()
				return nil
			}

			resources, err := model.ResourcesFromFile(file)
			if err != nil {
				return fmt.Errorf("error unmarshaling file: %s, %w", file, err)
			}

			resourceStatuses, err := c.Delete(cmd.Context(), resources)
			if err != nil {
				return err
			}

			model.PrintResourceUpdates(cmd.OutOrStdout(), resourceStatuses)
			return nil
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "delete resources from a file")

	cmd.AddCommand(
		deleteResourceCommand(bindplane, "agent", []string{"agents"}),
		deleteResourceCommand(bindplane, "configuration", []string{"configurations", "configs", "config"}),
		deleteResourceCommand(bindplane, "source", []string{"sources"}),
		deleteResourceCommand(bindplane, "source-type", []string{"source-types", "sourceType", "sourceTypes"}),
		deleteResourceCommand(bindplane, "destination", []string{"destinations"}),
		deleteResourceCommand(bindplane, "destination-type", []string{"destination-types", "destinationType", "destinationTypes"}),
	)

	return cmd
}

func deleteResourceCommand(bindplane *cli.BindPlane, resourceType string, aliases []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s <name>", resourceType),
		Aliases: aliases,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := bindplane.Client()
			if err != nil {
				return fmt.Errorf("error creating client: %w", err)
			}

			if len(args) == 0 {
				return fmt.Errorf("missing required argument <name>")
			}

			ctx := cmd.Context()
			name := args[0]
			batch := false

			switch resourceType {
			case "agent":
				_, err = c.DeleteAgents(ctx, args)
				batch = true
			case "configuration":
				err = c.DeleteConfiguration(ctx, name)
			case "source":
				err = c.DeleteSource(ctx, name)
			case "source-type":
				err = c.DeleteSourceType(ctx, name)
			case "destination":
				err = c.DeleteDestination(ctx, name)
			case "destination-type":
				err = c.DeleteDestinationType(ctx, name)
			default:
				return fmt.Errorf("unknown type, unable to delete %s '%s'", resourceType, name)
			}

			if err != nil {
				return err
			}

			if batch {
				for _, name := range args {
					fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted %s '%s'\n", resourceType, name)
				}
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted %s '%s'\n", resourceType, name)
			}
			return nil
		},
	}

	return cmd
}
