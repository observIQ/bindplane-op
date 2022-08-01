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

package copy

import (
	"errors"
	"fmt"

	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/spf13/cobra"
)

// ConfigurationCommand returns the BindPlane Copy Configuration cobra command.
func ConfigurationCommand(bindplane *cli.BindPlane) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configuration",
		Aliases: []string{"config"},
		Short:   "Copy a configuration resource.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("missing required arguments, must specify the configuration name and the desired name of the copy")
			}

			c, err := bindplane.Client()
			if err != nil {
				return err
			}

			if err := c.CopyConfig(cmd.Context(), args[0], args[1]); err != nil {
				return err
			}

			fmt.Printf("Successfully copied configuration %s as %s.\n", args[0], args[1])
			return nil
		},
	}

	return cmd
}
