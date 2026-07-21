// internal/commands/logout.go
package commands

import (
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"go.uber.org/zap"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored GitHub credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			container := getContainer(cmd)
			if container == nil || container.Logger == nil {
				// fallback: just try to delete the token without logger
				return keyring.Delete(serviceName, "github_token")
			}
			logger := container.Logger
			if err := keyring.Delete(serviceName, "github_token"); err != nil {
				logger.Warn("No token to delete", zap.Error(err))
				return nil
			}
			logger.Info("Successfully logged out")
			return nil
		},
	}
}
