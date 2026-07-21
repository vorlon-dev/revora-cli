// internal/commands/root.go
package commands

import (
	"context"

	"github.com/revora/revora/internal/config"
	"github.com/revora/revora/internal/di"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"go.uber.org/zap"
)

type contextKey string

const containerKey contextKey = "revora-container"

// NewRootCmd creates the root command with all subcommands.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revora",
		Short: "Revora OTA update delivery platform",
		Long:  `Revora allows developers to push over-the-air updates to mobile apps using GitHub Releases.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip container init for commands that don't need it
			switch cmd.Name() {
			case "help", "completion", "version":
				return nil
			}

			// Already initialised?
			if _, ok := cmd.Context().Value(containerKey).(*di.Container); ok {
				return nil
			}

			logger, err := zap.NewProduction()
			if err != nil {
				return err
			}

			// Load config
			cfg, err := config.Load()
			if err != nil {
				// Allow fallback with minimal config
				logger.Warn("Failed to load config, using defaults", zap.Error(err))
				cfg = &config.Manager{
					ProjectDir: ".",
				}
			}

			// If token not in env, try keyring
			if cfg.GitHubToken == "" {
				tok, err := keyring.Get(serviceName, "github_token")
				if err == nil && tok != "" {
					cfg.GitHubToken = tok
					logger.Info("Loaded GitHub token from keychain")
				}
			}

			container, err := di.NewContainer(logger, cfg)
			if err != nil {
				logger.Warn("Failed to initialise container", zap.Error(err))
				// still store a minimal container so commands can handle gracefully
				container = &di.Container{
					Config: cfg,
					Logger: logger,
				}
			}

			ctx := context.WithValue(cmd.Context(), containerKey, container)
			cmd.SetContext(ctx)
			return nil
		},
	}

	// subcommands
	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newLogoutCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newDoctorCmd())
	cmd.AddCommand(newPatchCmd())
	cmd.AddCommand(newBuildCmd())
	cmd.AddCommand(newApkCmd())
	cmd.AddCommand(newAabCmd())
	cmd.AddCommand(newReleaseCmd())
	cmd.AddCommand(newRollbackCmd())
	cmd.AddCommand(newVersionsCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newLogsCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newVersionCmd())

	return cmd
}

func getContainer(cmd *cobra.Command) *di.Container {
	if v, ok := cmd.Context().Value(containerKey).(*di.Container); ok {
		return v
	}
	return nil
}
