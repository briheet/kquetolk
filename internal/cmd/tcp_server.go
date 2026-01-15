package cmd

import (
	"context"

	"github.com/briheet/kquetolk/internal/tcp/server"
	"github.com/spf13/cobra"
)

func TcpServerCmd(ctx context.Context) *cobra.Command {
	tcpServerCmd := &cobra.Command{
		Use:   "tcpServer",
		Short: "Tcp server redis resp compatible.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			// New tcp server obj
			tcpServer, err := server.NewTcpServer(
				server.WithHost("localhost"),
				server.WithPort("6379"),
			)
			if err != nil {
				return err
			}

			if err := tcpServer.Execute(ctx); err != nil {
				return err
			}

			return nil
		},
	}

	return tcpServerCmd
}
