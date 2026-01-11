package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

func Execute(ctx context.Context) int {

	rootCmd := &cobra.Command{
		Use:   "kquetolk",
		Short: "Kquetolk is a redis resp compatible server written in go.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Profiling and stuff
			return nil
		},
	}

	rootCmd.AddCommand(TcpServerCmd(ctx))

	if err := rootCmd.Execute(); err != nil {
		return 1
	}

	return 0
}
