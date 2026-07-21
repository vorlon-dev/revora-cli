package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newAabCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "aab",
		Short: "Build AAB and create a patch",
		RunE: func(cmd *cobra.Command, args []string) error {
			container := getContainer(cmd)
			if container == nil {
				return fmt.Errorf("container not initialised")
			}
			if err := buildProject(container); err != nil {
				return err
			}
			return createPatch(cmd.Context(), container)
		},
	}
}
