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

	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/internal/cli/printer"
	"github.com/spf13/cobra"
)

// DestinationsCommand returns the BindPlane get destinations cobra command
func DestinationsCommand(bindplane *cli.BindPlane) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "destinations [id]",
		Aliases: []string{"destination"},
		Short:   "Displays the destinations",
		Long:    `A destination is a service that receives logs, metrics, and traces.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := bindplane.Client()
			if err != nil {
				return fmt.Errorf("error creating client: %w", err)
			}

			if len(args) > 0 {
				name := args[0]
				destination, err := c.Destination(cmd.Context(), name)
				if err != nil {
					return err
				}

				if destination == nil {
					return fmt.Errorf("no destination found with name %s", name)
				}

				printer.PrintResource(bindplane.Printer(), destination)
				return nil
			}

			destinations, err := c.Destinations(cmd.Context())
			if err != nil {
				return err
			}

			printer.PrintResources(bindplane.Printer(), destinations)
			return nil
		},
	}
	return cmd
}
