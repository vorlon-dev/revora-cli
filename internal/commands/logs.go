package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logs",
		Short: "Show update logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Logs require Revora Cloud integration – coming soon.")
			return nil
		},
	}
}
