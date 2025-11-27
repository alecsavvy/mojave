package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/sonata-labs/sonata/app"
	"github.com/sonata-labs/sonata/config"
	"github.com/spf13/cobra"
)

func NewStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run the Sonata node",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString("home")
			if err != nil {
				return err
			}

			config, err := config.ReadConfig(home)
			if err != nil {
				return err
			}

			app := app.NewApp(config)
			if err := app.Run(cmd.Context()); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				return fmt.Errorf("failed to run app: %w", err)
			}

			return nil
		},
	}
}
