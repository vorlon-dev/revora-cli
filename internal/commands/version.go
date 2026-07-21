package commands

import (
"fmt"

"github.com/spf13/cobra"
)

var Version = "0.1.0"

func newVersionCmd() *cobra.Command {
return &cobra.Command{
Use:   "version",
Short: "Print the version number of Revora CLI",
Run: func(cmd *cobra.Command, args []string) {
fmt.Println("Revora CLI version", Version)
},
}
}
