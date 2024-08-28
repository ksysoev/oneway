package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/ksysoev/oneway/pkg/cmd"
	"github.com/spf13/cobra"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	defer cancel()

	rootCmd := &cobra.Command{
		Use:   "oneway",
		Short: "oneway is mesh network",
		Long:  "oneway is a mesh network that allows services to communicate with each other",
	}

	cmd.InitCommands(rootCmd)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		slog.Error("failed to execute command", slog.Any("error", err))
		os.Exit(1)
	}
}
