package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback",
		Short: "Rollback to a previous release",
		RunE: func(cmd *cobra.Command, args []string) error {
			container := getContainer(cmd)
			if container == nil {
				return fmt.Errorf("container not initialised")
			}
			if container.GitHubClient == nil {
				return fmt.Errorf("GitHub client not available – run revora login first")
			}
			container.Logger.Info("Rollback not yet implemented")
			return fmt.Errorf("rollback command not yet implemented")
		},
	}
}
