package commands

import (
	"github.com/spf13/cobra"
)

// StartCmd is the start subcommand (runs the node/server).
var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the node",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
