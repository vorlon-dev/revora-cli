package commands

import (
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/revora/revora/internal/di"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"go.uber.org/zap"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check system prerequisites",
		RunE: func(cmd *cobra.Command, args []string) error {
			container := getContainer(cmd)
			if container == nil {
				return fmt.Errorf("container not initialised")
			}
			return runDoctor(container)
		},
	}
}

func runDoctor(container *di.Container) error {
	logger := container.Logger
	var issues []string

	// 1. Check git
	logger.Info("Checking git...")
	if _, err := exec.LookPath("git"); err != nil {
		issues = append(issues, "git not found in PATH")
		logger.Warn("git not found")
	} else {
		logger.Info("git found")
	}

	// 2. Check patch engine
	logger.Info("Checking patch engine...")
	patchBinary := container.Config.PatchEngine
	if patchBinary == "" {
		patchBinary = "revora-patch"
	}
	if _, err := exec.LookPath(patchBinary); err != nil {
		issues = append(issues, fmt.Sprintf("%s not found in PATH (required for patch operations)", patchBinary))
		logger.Warn("patch engine not found")
	} else {
		logger.Info("patch engine found")
	}

	// 3. Check GitHub token
	logger.Info("Checking GitHub token...")
	token := container.Config.GitHubToken
	if token == "" {
		// try keyring
		tok, err := keyring.Get(serviceName, "github_token")
		if err == nil && tok != "" {
			token = tok
			container.Config.GitHubToken = tok
		}
	}
	if token == "" {
		issues = append(issues, "GitHub token not set (run revora login or set REVORA_GITHUB_TOKEN)")
		logger.Warn("GitHub token not found")
	} else {
		// Validate token by calling GitHub API
		client := &http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
		if err != nil {
			issues = append(issues, "Failed to create token validation request")
		} else {
			req.Header.Set("Authorization", "token "+token)
			req.Header.Set("User-Agent", "revora-cli")
			resp, err := client.Do(req)
			if err != nil {
				issues = append(issues, fmt.Sprintf("Token validation failed: %v", err))
				logger.Warn("Token validation failed", zap.Error(err))
			} else {
				defer resp.Body.Close()
				if resp.StatusCode != 200 {
					issues = append(issues, fmt.Sprintf("GitHub token invalid (HTTP %d)", resp.StatusCode))
					logger.Warn("GitHub token invalid", zap.Int("status", resp.StatusCode))
				} else {
					logger.Info("GitHub token valid")
				}
			}
		}
	}

	// Print results
	fmt.Println("\nRevora Doctor Results:")
	if len(issues) == 0 {
		fmt.Println("✓ All checks passed. Your environment is ready.")
		return nil
	}
	fmt.Println("✗ Some issues found:")
	for _, issue := range issues {
		fmt.Printf("  - %s\n", issue)
	}
	return fmt.Errorf("doctor found %d issues", len(issues))
}
