package client

import (
	"errors"
	"net"
	"strings"
	"sync"

	"github.com/briheet/kquetolk/internal/resp"
)

var (
	ErrNilClient = errors.New("net.Conn is nil")
)

type Client interface {
	HandleConnection() error
	handleCommand(resp.Value) []byte
	flush() error
	Close()
}

var _ Client = (*Conn)(nil)

type Conn struct {
	conn net.Conn

	mu       sync.RWMutex
	readBuf  []byte
	writeBuf []byte
}

type options struct {
	conn net.Conn
}

type Option interface {
	apply(*options)
}

type clientConnOption struct {
	conn net.Conn
}

func WithClientConn(conn net.Conn) Option {
	return clientConnOption{conn: conn}
}

func (o clientConnOption) apply(opts *options) {
	opts.conn = o.conn
}

func NewClientConnection(opts ...Option) (*Conn, error) {

	// Default config
	cfg := options{
		conn: nil,
	}

	for _, opt := range opts {
		opt.apply(&cfg)
	}

	if cfg.conn == nil {
		return nil, ErrNilClient
	}

	return &Conn{
		conn:     cfg.conn,
		readBuf:  make([]byte, 0, 4096),
		writeBuf: make([]byte, 0, 4096),
	}, nil
}

func (c *Conn) HandleConnection() error {
	tmp := make([]byte, 4096)

	for {
		n, err := c.conn.Read(tmp)
		if err != nil {
			return err
		}

		c.readBuf = append(c.readBuf, tmp[:n]...)

		for {
			val, consumed, err := resp.Parse(c.readBuf)

			if err == resp.ErrIncomplete {
				break
			}
			if err != nil {
				return err
			}

			c.readBuf = c.readBuf[consumed:]

			reply := c.handleCommand(val)

			c.writeBuf = append(c.writeBuf, reply...)
		}

		if err := c.flush(); err != nil {
			return err
		}
	}
}

func (c *Conn) handleCommand(v resp.Value) []byte {
	if v.Type != '*' || len(v.Array) == 0 {
		return resp.EncodeError("ERR invalid command")
	}

	cmd := strings.ToUpper(v.Array[0].Str)

	switch cmd {
	case "PING":
		return resp.EncodeSimpleString("PONG")
	case "ECHO":
		if len(v.Array) != 2 {
			return resp.EncodeError("ERR wrong number of arguments")
		}
		return resp.EncodeBulkString(v.Array[1].Str)
	default:
		return resp.EncodeError("ERR unknown command")
	}
}

func (c *Conn) flush() error {
	if len(c.writeBuf) == 0 {
		return nil
	}

	_, err := c.conn.Write(c.writeBuf)
	if err != nil {
		return err
	}

	c.writeBuf = c.writeBuf[:0]
	return nil
}

func (c *Conn) Close() {
	defer c.conn.Close()
}
