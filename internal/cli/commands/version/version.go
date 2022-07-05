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

package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/internal/version"
)

// Command returns the BindPlane versions cobra command.
func Command(bindplane *cli.BindPlane) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints BindPlane version",
		Long:  `Prints BindPlane build version (commit or tag).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := bindplane.Client()
			if err != nil {
				return fmt.Errorf("error creating client: %w", err)
			}

			sv, err := c.Version(cmd.Context())
			if err != nil {
				m := fmt.Sprintf("Failed to get version from server: %s\n", err.Error())
				fmt.Fprint(cmd.OutOrStdout(), m)
			}

			fmt.Fprintf(cmd.OutOrStdout(),
				"client: %s\nserver: %s\n",
				version.NewVersion().String(),
				sv.String())

			return nil
		},
	}
	return cmd
}
