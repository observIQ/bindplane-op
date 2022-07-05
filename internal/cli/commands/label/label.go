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

package label

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/model"
)

var (
	overwriteFlag bool
	listFlag      bool
)

// Command returns the BindPlane label resource cobra command.
func Command(bindplane *cli.BindPlane) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "label type name/id [key0=value0 ... keyN=valueN]",
		Aliases: []string{"labels"},
		Short:   "List or modify the labels of a resource",
		Long:    `Agents are identified by id. All other resources are identified by name.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return label(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), args, bindplane)
		},
	}

	cmd.Flags().BoolVar(&overwriteFlag, "overwrite", false, "If true, then existing labels will be overwritten. Defaults to false which will produce an error if a label with the same name and a different value already exists.")
	cmd.Flags().BoolVar(&listFlag, "list", false, "If true, list the labels for the resource.")

	return cmd
}

func label(ctx context.Context, stdout io.Writer, stderr io.Writer, args []string, bindplane *cli.BindPlane) error {
	if len(args) == 0 {
		return fmt.Errorf("missing resource type name")
	}
	resourceType := model.ParseKind(args[0])
	if resourceType != model.KindAgent {
		// TODO(andy): Support other resource types
		return fmt.Errorf("only agent labels are currently supported, other resources coming soon")
	}
	if len(args) == 1 {
		if args[0] == "agent" {
			return fmt.Errorf("missing agent id")
		}
		return fmt.Errorf("missing %s name", args[0])
	}

	resources, changes, err := splitArgs(args[1:])
	if err != nil {
		return err
	}

	client, err := bindplane.Client()
	if err != nil {
		return err
	}

	// no resources and --list
	if len(changes) == 0 && listFlag {
		for _, resource := range resources {
			labels, err := client.AgentLabels(ctx, resource)
			if err != nil {
				fmt.Fprintf(stderr, "error: %s\n", err.Error())
				continue
			}
			printLabels(stdout, resourceType, resource, len(resources) > 1, labels)
		}
		return nil
	}

	// no changes
	if len(changes) == 0 {
		return fmt.Errorf("no label updates specified, use --list to show labels")
	}

	// no resources
	if len(resources) == 0 {
		return fmt.Errorf("no resources specified")
	}

	// resources and changes
	labels, err := newLabels(changes)
	if err != nil {
		return err
	}
	for _, resource := range resources {
		labels, err := client.ApplyAgentLabels(ctx, resource, &labels, overwriteFlag)
		if err != nil {
			// adjust &overwrite=true feedback that comes from REST API call
			errorMessage := strings.Replace(err.Error(), "?overwrite=true", "--overwrite", 1)
			fmt.Fprintf(stderr, "error: %s\n", errorMessage)
			printLabels(stdout, resourceType, resource, true, labels)
			continue
		}
		if labels == nil {
			fmt.Fprintf(stderr, "error: no labels returned\n")
			continue
		}
		if listFlag {
			printLabels(stdout, resourceType, resource, len(resources) > 1, labels)
		} else {
			fmt.Fprintf(stdout, "%s %s labeled\n", resourceType, resource)
		}
	}

	return nil
}

func printLabels(stdout io.Writer, resourceType model.Kind, resource string, printHeading bool, labels *model.Labels) {
	if labels == nil {
		return
	}
	if printHeading {
		fmt.Fprintf(stdout, "Labels for %s %s:\n", resourceType, resource)
	}
	for name, value := range labels.Set {
		if printHeading {
			fmt.Fprintf(stdout, " %s=%s\n", name, value)
		} else {
			fmt.Fprintf(stdout, "%s=%s\n", name, value)
		}
	}
}

func splitArgs(args []string) (resources []string, changes []*labelChange, err error) {
	resources = []string{}
	changes = []*labelChange{}
	for _, arg := range args {
		change := parseChange(arg)
		if change != nil {
			changes = append(changes, change)
			continue
		}
		if len(changes) > 0 {
			// resources appears after changes
			return nil, nil, fmt.Errorf("resource name/id must be before label changes: %s", arg)
		}
		resources = append(resources, arg)
	}
	return resources, changes, nil
}

type labelChange struct {
	name  string
	value string
}

// parseChange returns the name/value pair for a label change or nil if this is not a label change
// "name=value" will have name and value parsed
// "name-" will have name parsed and value ""
func parseChange(arg string) *labelChange {
	pair := strings.Split(arg, "=")
	if len(pair) > 2 {
		return nil
	}
	if len(pair) == 2 {
		return &labelChange{
			name:  pair[0],
			value: pair[1],
		}
	}
	minus := strings.TrimSuffix(arg, "-")
	if arg != minus {
		// string was modified, suffix removed
		return &labelChange{
			name:  minus,
			value: "",
		}
	}
	return nil
}

func (l labelChange) String() string {
	return fmt.Sprintf("%s=%s", l.name, l.value)
}

func newLabels(changes []*labelChange) (model.Labels, error) {
	labels := map[string]string{}
	for _, change := range changes {
		labels[change.name] = change.value
	}
	return model.LabelsFromMap(labels)
}
