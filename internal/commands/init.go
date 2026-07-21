// internal/commands/init.go
package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/revora/revora/internal/crypto"
	"github.com/revora/revora/internal/di"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var platformFlag string

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Revora in the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			container := getContainer(cmd)
			if container == nil {
				return fmt.Errorf("container not initialised")
			}
			return initProject(cmd, container)
		},
	}
	cmd.Flags().StringVarP(&platformFlag, "platform", "p", "", "Force a platform (android, flutter, react-native, ios, kotlin, compose, unity)")
	return cmd
}

func initProject(cmd *cobra.Command, container *di.Container) error {
	logger := container.Logger

	// Force the current working directory as the project root
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	// Determine platform
	var platform string
	if platformFlag != "" {
		platform = platformFlag
		logger.Info("Using platform from flag", zap.String("platform", platform))
	} else {
		detected, err := container.PlatformDetector.Detect(cwd)
		if err != nil {
			logger.Warn("Could not auto-detect platform", zap.Error(err))
			fmt.Println("Auto-detection failed. Supported platforms: android, flutter, react-native, ios, kotlin, compose, unity, unknown")
			fmt.Print("Enter platform: ")
			if _, scanErr := fmt.Scanln(&platform); scanErr != nil {
				return fmt.Errorf("read platform: %w", scanErr)
			}
		} else {
			platform = detected
		}
	}

	allowed := map[string]bool{
		"android": true, "flutter": true, "react-native": true,
		"ios": true, "kotlin": true, "compose": true, "unity": true, "unknown": true,
	}
	if !allowed[platform] {
		return fmt.Errorf("unsupported platform '%s'", platform)
	}
	logger.Info("Using platform", zap.String("platform", platform))

	// Create .revora directory inside current directory
	revoraDir := filepath.Join(cwd, ".revora")
	if err := os.MkdirAll(revoraDir, 0755); err != nil {
		return fmt.Errorf("create .revora: %w", err)
	}

	// Generate signing keys
	if container.KeyManager != nil {
		// Recreate key manager for the correct directory
		container.KeyManager = crypto.NewKeyManager(cwd) // force keys into current dir
		if err := container.KeyManager.Generate(); err != nil {
			return fmt.Errorf("generate keys: %w", err)
		}
		logger.Info("Signing keys generated")
	}

	// Write revora.yaml in current directory
	configFile := filepath.Join(cwd, "revora.yaml")
	publicKeyPath := filepath.Join(revoraDir, "keys", "public.pem")
	content := fmt.Sprintf("platform: %s\nsigning_key: %s\n", platform, publicKeyPath)
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write revora.yaml: %w", err)
	}
	logger.Info("revora.yaml created")

	// Create GitHub Actions workflow
	workflowDir := filepath.Join(cwd, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return fmt.Errorf("create workflows dir: %w", err)
	}
	workflowFile := filepath.Join(workflowDir, "revora.yml")
	workflowContent := `name: Revora Release
on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to release'
        required: true
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Revora
        run: |
          curl -sSL https://install.revora.dev | bash
      - name: Build and Patch
        run: revora patch --ci
        env:
          REVORA_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
`
	if err := os.WriteFile(workflowFile, []byte(workflowContent), 0644); err != nil {
		return fmt.Errorf("write workflow: %w", err)
	}
	logger.Info("GitHub Actions workflow created")

	fmt.Println("Revora initialized successfully!")
	fmt.Printf("Platform: %s\n", platform)
	fmt.Printf("Configuration: %s\n", configFile)
	fmt.Printf("Keys: %s\n", revoraDir)
	return nil
}
