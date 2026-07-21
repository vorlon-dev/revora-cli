package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/revora/revora/internal/di"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newBuildCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Build the project for current platform",
		RunE: func(cmd *cobra.Command, args []string) error {
			container := getContainer(cmd)
			if container == nil {
				return fmt.Errorf("container not initialised")
			}
			return buildProject(container)
		},
	}
}

func buildProject(container *di.Container) error {
	logger := container.Logger
	platform := container.Config.Platform

	logger.Info("Starting build", zap.String("platform", platform))

	var cmd *exec.Cmd
	switch platform {
	case "android", "kotlin", "compose":
		// Assume gradle wrapper is present
		cmd = exec.Command("./gradlew", "assembleRelease")
	case "flutter":
		cmd = exec.Command("flutter", "build", "apk")
	case "react-native":
		cmd = exec.Command("npx", "react-native", "run-android", "--variant=release")
	case "ios":
		cmd = exec.Command("xcodebuild", "-workspace", "ios/App.xcworkspace", "-scheme", "App", "-configuration", "Release")
	default:
		return fmt.Errorf("unsupported platform for building: %s", platform)
	}

	cmd.Dir = container.Config.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	logger.Info("Build successful")
	return nil
}
