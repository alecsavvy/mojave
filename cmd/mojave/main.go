package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecsavvy/mojave/app"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:    "run",
		Aliases: []string{"r"},
		Usage:   "run the mojave node",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// random tmp dir
			homeDir := os.TempDir() + "/mojave-dev-" + time.Now().Format("20060102150405")
			app := app.NewApp(homeDir)

			if err := app.Run(ctx); err != nil {
				return err
			}

			return nil
		},
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := cmd.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
