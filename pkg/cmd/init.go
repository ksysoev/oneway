package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func InitCommands() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "oneway",
		Short: "oneway is mesh network",
		Long:  "oneway is a mesh network that allows services to communicate with each other",
	}

	cmd.AddCommand(ExchangeCommand(&configPath))
	cmd.AddCommand(RevProxyCommand(&configPath))

	cmd.PersistentFlags().StringVar(&configPath, "config", "./runtime/config.yaml", "config file path")

	return cmd
}

func ExchangeCommand(cfgPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "exchange",
		Short: "Start the Exchange server",
		Long:  "Start the Exchange server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := initConfig(*cfgPath)
			if err != nil {
				return fmt.Errorf("failed to inititialize config: %w", err)
			}

			return runExchange(cmd.Context(), cfg.Exchange)
		},
	}
}

func RevProxyCommand(cfgPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "revproxy",
		Short: "Start the RevProxy server",
		Long:  "Start the RevProxy server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := initConfig(*cfgPath)
			if err != nil {
				return fmt.Errorf("failed to inititialize config: %w", err)
			}

			return runRevProxy(cmd.Context(), cfg.Revproxy)
		},
	}
}
