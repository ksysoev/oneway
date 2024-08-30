package cmd

import (
	"github.com/spf13/cobra"
)

func InitCommands(cmd *cobra.Command) {
	cmd.AddCommand(ExchangeCommand())
	cmd.AddCommand(RevConProxyCommand())
}

func ExchangeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "exchange",
		Short: "Start the exchange server",
		Long:  "Start the exchange server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExchange(cmd.Context())
		},
	}
}

func RevConProxyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "revconproxy",
		Short: "Start the RevConProxy server",
		Long:  "Start the RevConProxy server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRevConProxy(cmd.Context())
		},
	}
}
