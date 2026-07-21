package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View or set configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			container := getContainer(cmd)
			if container == nil {
				return fmt.Errorf("container not initialised")
			}
			// For now just display the config file location and contents
			configFile := filepath.Join(container.Config.ProjectDir, "revora.yaml")
			if _, err := os.Stat(configFile); os.IsNotExist(err) {
				fmt.Println("No revora.yaml found in project. Run 'revora init' first.")
				return nil
			}
			data, err := os.ReadFile(configFile)
			if err != nil {
				return err
			}
			fmt.Printf("Config file: %s\n", configFile)
			fmt.Println(string(data))
			return nil
		},
	}
	// future: add set subcommand
	return cmd
}
