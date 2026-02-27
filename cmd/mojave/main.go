package main

import (
	"context"
	"log"
	"os"

	"github.com/alecsavvy/mojave/app"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:    "run",
		Aliases: []string{"r"},
		Usage:   "run the mojave node",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			app := app.NewApp()

			if err := app.Run(ctx); err != nil {
				return err
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
