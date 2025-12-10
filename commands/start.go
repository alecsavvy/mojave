package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/alecsavvy/mojave/app"
	"github.com/alecsavvy/mojave/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run the Sonata node",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, _ := zap.NewDevelopment()
			defer logger.Sync()

			home, err := cmd.Flags().GetString("home")
			if err != nil {
				return err
			}

			config, err := config.ReadConfig(home)
			if err != nil {
				return err
			}

			app, err := app.NewApp(config, logger)
			if err != nil {
				return err
			}

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
