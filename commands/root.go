package commands

import (
	"context"
	"fmt"

	"github.com/sonata-labs/sonata/config"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "sonata",
		Short: "Sonata is an open distribution platform for the music industry",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello, World!")
		},
	}

	root.PersistentFlags().String("home", config.DefaultHomeDirPath(), "home directory (default is $HOME/.sonata)")

	root.AddCommand(NewStartCommand())

	return root
}

func Execute(ctx context.Context) error {
	return NewRootCommand().ExecuteContext(ctx)
}
