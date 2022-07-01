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

package apply

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/model"
)

// Command returns the bindplane apply cobra command.
func Command(bindplane *cli.BindPlane) *cobra.Command {
	var fileFlag []string

	cmd := &cobra.Command{
		Use:   "apply [file]",
		Short: "Apply resources",
		Long:  `Apply resources from a file with a filepath or use 'bindplane apply -' to apply resources from stdin.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := bindplane.Client()
			if err != nil {
				return fmt.Errorf("error creating client: %w", err)
			}

			// any positional args are treated as if they were prefixed with -f/--file. this allows shell globs to be used
			// with or without -f. for example, the following two commands are the same "apply -f *.yaml" and "apply *.yaml"
			fileArgs := fileFlag
			fileArgs = append(fileArgs, args...)

			if len(fileArgs) == 0 {
				// This will not return an error for the default help function.
				_ = cmd.Help()
				return nil
			}

			var errs error
			var resources []*model.AnyResource

			// read all of the files
			for _, fileArg := range fileArgs {
				fileResources, err := readResources(cmd, fileArg)
				if err != nil {
					errs = multierror.Append(errs, err)
					continue
				}
				resources = append(resources, fileResources...)
			}

			// fail if any file cannot be read
			if errs != nil {
				return errs
			}

			// apply them all together
			resourceStatuses, err := c.Apply(cmd.Context(), resources)
			if err != nil {
				return err
			}

			model.PrintResourceUpdates(cmd.OutOrStdout(), resourceStatuses)
			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&fileFlag, "file", "f", []string{}, "path to a yaml file that specifies bindplane resources")

	return cmd
}

func readResources(cmd *cobra.Command, fileArg string) ([]*model.AnyResource, error) {
	if fileArg == "-" {
		return model.ResourcesFromReader(cmd.InOrStdin())
	}
	return model.ResourcesFromFile(fileArg)
}
