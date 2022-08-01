package copy

import (
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/spf13/cobra"
)

// Command returns the BindPlane copy cobra command.
func Command(bindplane *cli.BindPlane) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "copy",
		Aliases: []string{"cp"},
		Short:   "Make a copy of a resource",
		Example: "bindplanectl copy config my-config my-config-copy",
	}

	cmd.AddCommand(
		ConfigurationCommand(bindplane),
	)

	return cmd
}
