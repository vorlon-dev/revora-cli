package main

import (
"fmt"
"os"

"github.com/revora/revora/internal/commands"
)

func main() {
if err := commands.NewRootCmd().Execute(); err != nil {
fmt.Fprintln(os.Stderr, err)
os.Exit(1)
}
}
