package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show project status",
		RunE: func(cmd *cobra.Command, args []string) error {
			container := getContainer(cmd)
			if container == nil {
				return fmt.Errorf("container not initialised")
			}
			cfg := container.Config
			tokenStatus := "not set"
			if cfg.GitHubToken != "" {
				tokenStatus = "set"
			}
			fmt.Printf("Project directory: %s\n", cfg.ProjectDir)
			fmt.Printf("Platform: %s\n", cfg.Platform)
			fmt.Printf("GitHub token: %s\n", tokenStatus)
			fmt.Printf("Signing key: %s\n", cfg.SigningKey)
			return nil
		},
	}
}
