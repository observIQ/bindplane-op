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

// DestinationTypesCommand returns the BindPlane get destination-types cobra command
func DestinationTypesCommand(bindplane *cli.BindPlane) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "destination-types [id]",
		Aliases: []string{"destination-type"},
		Short:   "Displays the destination types",
		Long:    `A destination type is a type of service that receives logs, metrics, and traces.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := bindplane.Client()
			if err != nil {
				return fmt.Errorf("error creating client: %w", err)
			}

			if len(args) > 0 {
				name := args[0]
				destinationType, err := c.DestinationType(cmd.Context(), name)
				if err != nil {
					return err
				}

				if destinationType == nil {
					return fmt.Errorf("no destination-type found with name %s", name)
				}

				printer.PrintResource(bindplane.Printer(), destinationType)
				return nil
			}

			destinationTypes, err := c.DestinationTypes(cmd.Context())
			if err != nil {
				return err
			}

			printer.PrintResources(bindplane.Printer(), destinationTypes)
			return nil
		},
	}
	return cmd
}
