package cmd

import (
	"github.com/spf13/cobra"
)

func InitCommands(cmd *cobra.Command) {
	cmd.AddCommand(ExchangeCommand())
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
