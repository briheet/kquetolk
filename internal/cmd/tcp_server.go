package cmd

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	gnettcp "github.com/briheet/kquetolk/internal/gnetTcp"
	"github.com/briheet/kquetolk/internal/storage"
	"github.com/panjf2000/gnet/v2"
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
			store := storage.New()
			defer store.Close()

			gnetServer := gnettcp.NewGnetServer(
				gnettcp.WithStore(store),
				gnettcp.WithAddr(":6379"),
				gnettcp.WithMultiCore(true),
			)

			log.Println("Starting server on tcp://:6379")

			errChan := make(chan error, 1)
			go func() {
				errChan <- gnet.Run(
					gnetServer,
					fmt.Sprintf("tcp://%s", gnetServer.Addr),
					gnet.WithMulticore(gnetServer.MultiCore),
					gnet.WithLockOSThread(true),
					gnet.WithLoadBalancing(gnet.LeastConnections),
					gnet.WithReuseAddr(true),
					gnet.WithReusePort(true),
					gnet.WithTCPNoDelay(gnet.TCPNoDelay),
					gnet.WithTCPKeepAlive(time.Minute*5),
					gnet.WithReadBufferCap(64*1024),
					gnet.WithWriteBufferCap(64*1024),
					gnet.WithSocketRecvBuffer(64*1024),
					gnet.WithSocketSendBuffer(64*1024),
					// gnet.WithEdgeTriggeredIO(true), // disabled - can hurt perf
					gnet.WithNumEventLoop(runtime.NumCPU()),
				)
			}()

			select {
			case err := <-errChan:
				return err
			case <-ctx.Done():
				log.Println("Shutting down server...")
				return gnetServer.Stop(context.Background())
			}
		},
	}

	return tcpServerCmd
}
