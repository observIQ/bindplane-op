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
