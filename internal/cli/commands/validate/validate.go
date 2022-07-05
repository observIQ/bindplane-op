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

package validate

import (
	"fmt"
	"os"

	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/spf13/cobra"
)

// Command returns the BindPlane validate get cobra command
func Command(bindplane *cli.BindPlane) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "validate the current profile",
		Run: func(cmd *cobra.Command, args []string) {
			if err := bindplane.Config.Validate(); err != nil {
				fmt.Fprint(cmd.OutOrStdout(), err)
				os.Exit(1)
			}
			fmt.Fprint(cmd.OutOrStdout(), "configuration is valid\n")
			os.Exit(0)
		},
	}

	return cmd
}
