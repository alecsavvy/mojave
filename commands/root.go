package commands

import (
	"github.com/spf13/cobra"
)

// RootCmd is the root command.
var RootCmd = &cobra.Command{
	Use:   "mojave",
	Short: "Mojave node",
}

func init() {
	RootCmd.AddCommand(StartCmd)
}

// Execute runs the root command.
func Execute() error {
	return RootCmd.Execute()
}
