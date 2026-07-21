// internal/commands/login.go
package commands

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/revora/revora/internal/di"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Store a GitHub personal access token",
		Long: `Authenticate with GitHub by providing a personal access token.
The token must have 'repo' and 'workflow' scopes.
You can create one at https://github.com/settings/tokens`,
		RunE: func(cmd *cobra.Command, args []string) error {
			container := getContainer(cmd)
			if container == nil || container.Logger == nil {
				return fmt.Errorf("container not initialised")
			}
			return loginInteractive(container)
		},
	}
}

func loginInteractive(container *di.Container) error {
	logger := container.Logger
	logger.Info("Prompting for GitHub token")

	fmt.Print("GitHub personal access token: ")
	byteToken, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("read token: %w", err)
	}
	fmt.Println() // newline after hidden input
	token := strings.TrimSpace(string(byteToken))

	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	if err := keyring.Set(serviceName, "github_token", token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	logger.Info("Token saved successfully")
	fmt.Println("Login successful. Token stored in system keychain.")
	return nil
}
