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

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/internal/cli/printer"
)

// AgentsCommand returns the BindPlane get agents cobra command
func AgentsCommand(bindplane *cli.BindPlane) *cobra.Command {
	var (
		selector string
		query    string
		limit    int
		offset   int
	)
	cmd := &cobra.Command{
		Use:     "agents [id]",
		Aliases: []string{"agent"},
		Short:   "Displays the agents",
		Long:    `An agent collects logs, metrics, and traces for sources and sends them to destinations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := bindplane.Client()
			if err != nil {
				return fmt.Errorf("error creating client: %w", err)
			}

			if len(args) > 0 {
				id := args[0]
				agent, err := c.Agent(cmd.Context(), id)
				if err != nil {
					return err
				}

				if agent == nil {
					return fmt.Errorf("no agent found with ID %s", id)
				}

				printer.PrintResource(bindplane.Printer(), agent)
				return nil
			}

			agents, err := c.Agents(cmd.Context(),
				client.WithSelector(selector),
				client.WithQuery(query),
				client.WithOffset(offset),
				client.WithLimit(limit),
			)
			if err != nil {
				return err
			}

			printer.PrintResources(bindplane.Printer(), agents)
			return nil
		},
	}

	cmd.Flags().StringVarP(&selector, "selector", "l", "", "label selector to filter agents by label, e.g. name=value")
	cmd.Flags().StringVarP(&query, "query", "q", "", "search query to filter agents")
	cmd.Flags().IntVar(&offset, "offset", 0, "number of agents to skip for paging")
	cmd.Flags().IntVar(&limit, "limit", 100, "maximum number of agents to return")

	return cmd
}
