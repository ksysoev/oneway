package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/ksysoev/oneway/pkg/cmd"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	rootCmd := cmd.InitCommands()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		slog.Error("failed to execute command", slog.Any("error", err))
		cancel()
		os.Exit(1)
	}

	cancel()
}
