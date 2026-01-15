package server

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"

	"github.com/briheet/kquetolk/internal/tcp/client"
)

type Server interface {
	Execute(context.Context) error
	Start() error
	Shutdown() error
}

var _ Server = (*TcpServer)(nil)

type options struct {
	host string
	port string
}

type Option interface {
	apply(*options)
}

type hostOption string

func WithHost(host string) Option {
	return hostOption(host)
}

func (h hostOption) apply(opts *options) {
	opts.host = string(h)
}

type portOption string

func WithPort(port string) Option {
	return portOption(port)
}

func (p portOption) apply(opts *options) {
	opts.port = string(p)
}

type TcpServer struct {
	// Server specific stuff
	addr    *net.TCPAddr
	Listner *net.TCPListener

	// To manage all connections
	mu          sync.RWMutex
	connections map[net.Conn]*client.Conn
}

func NewTcpServer(opts ...Option) (*TcpServer, error) {

	// Defaults
	cfg := options{
		host: "localhost",
		port: "6379",
	}

	for _, opt := range opts {
		opt.apply(&cfg)
	}

	addr := net.JoinHostPort(cfg.host, cfg.port)

	newtcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	newListener, err := net.ListenTCP("tcp", newtcpAddr)
	if err != nil {
		return nil, err
	}

	return &TcpServer{
		addr:        newtcpAddr,
		Listner:     newListener,
		mu:          sync.RWMutex{},
		connections: make(map[net.Conn]*client.Conn),
	}, nil
}

func (s *TcpServer) Execute(ctx context.Context) error {

	errChan := make(chan error, 1)

	go func() {
		if err := s.Start(); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return s.Shutdown()
	}
}

func (s *TcpServer) Start() error {

	for {
		conn, err := s.Listner.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			log.Printf("Error connecting to server: %v", err)
			return err
		}

		clientConn, err := client.NewClientConnection(
			client.WithClientConn(conn),
		)
		if err != nil {
			log.Println(err)
		}

		s.mu.Lock()
		s.connections[conn] = clientConn
		s.mu.Unlock()

		go clientConn.HandleConnection()
	}

}

func (s *TcpServer) Shutdown() error {

	if err := s.Listner.Close(); err != nil {
		return err
	}

	// Lock and get the connections, append and leave global lock and close.
	s.mu.Lock()

	connections := make([]*client.Conn, 0, len(s.connections))
	for _, conn := range s.connections {
		connections = append(connections, conn)
	}

	s.mu.Unlock()

	for _, c := range connections {
		c.Close()
	}

	return nil
}
