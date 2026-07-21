package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newApkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apk",
		Short: "Build APK and create a patch",
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
