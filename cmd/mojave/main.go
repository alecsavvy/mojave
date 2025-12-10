package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecsavvy/mojave/commands"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := commands.Execute(ctx); err != nil {
		fmt.Printf("boom: %v\n", err)
		os.Exit(1)
	}
}
