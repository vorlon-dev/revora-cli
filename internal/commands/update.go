package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update Revora CLI to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Self-update not yet implemented – please download the latest release from GitHub.")
			return nil
		},
	}
}
