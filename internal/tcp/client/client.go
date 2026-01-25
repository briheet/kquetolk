package client

import (
	"bytes"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/briheet/kquetolk/internal/resp"
	"github.com/briheet/kquetolk/internal/storage"
)

var (
	ErrNilClient = errors.New("net.Conn is nil")
)

var dataPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

type Client interface {
	HandleConnection() error
	handleCommand(resp.Value) []byte
	flush() error
	Close()
}

var _ Client = (*Conn)(nil)

type Conn struct {
	mu       sync.RWMutex
	writeBuf []byte
	readBuf  []byte
	conn     net.Conn
	store    *storage.ShardedMap
}

type options struct {
	conn  net.Conn
	store *storage.ShardedMap
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

type storageOption struct {
	store *storage.ShardedMap
}

func WithStorage(store *storage.ShardedMap) Option {
	return storageOption{store: store}
}

func (o storageOption) apply(opts *options) {
	opts.store = o.store
}

func NewClientConnection(opts ...Option) (*Conn, error) {

	// Default config
	cfg := options{
		conn:  nil,
		store: nil,
	}

	for _, opt := range opts {
		opt.apply(&cfg)
	}

	if cfg.conn == nil {
		return nil, ErrNilClient
	}

	return &Conn{
		conn:     cfg.conn,
		store:    cfg.store,
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
		return c.handlePing()
	case "ECHO":
		return c.handleEcho(v.Array)
	case "SET":
		return c.handleSet(v.Array[1:])
	case "GET":
		return c.handleGet(v.Array[1:])
	default:
		return resp.EncodeError("ERR unknown command")
	}
}

func (c *Conn) handlePing() []byte {
	return resp.EncodeSimpleString("PONG")
}

func (c *Conn) handleEcho(args []resp.Value) []byte {
	if len(args) != 2 {
		return resp.EncodeError("ERR wrong number of arguments")
	}
	return resp.EncodeBulkString(args[0].Str)
}

func (c *Conn) handleSet(args []resp.Value) []byte {
	if len(args) < 2 {
		return resp.EncodeError("ERR wrong number of arguments for 'set' command")
	}

	key := args[0].Str
	value := []byte(args[1].Str)
	var ttl time.Duration

	// Parse optional arguments: EX seconds, PX milliseconds
	for i := 2; i < len(args); i++ {
		opt := strings.ToUpper(args[i].Str)
		switch opt {
		case "EX":
			if i+1 >= len(args) {
				return resp.EncodeError("ERR syntax error")
			}
			secs, err := strconv.Atoi(args[i+1].Str)
			if err != nil {
				return resp.EncodeError("ERR value is not an integer or out of range")
			}
			ttl = time.Duration(secs) * time.Second
			i++
		case "PX":
			if i+1 >= len(args) {
				return resp.EncodeError("ERR syntax error")
			}
			ms, err := strconv.Atoi(args[i+1].Str)
			if err != nil {
				return resp.EncodeError("ERR value is not an integer or out of range")
			}
			ttl = time.Duration(ms) * time.Millisecond
			i++
		}
	}

	c.store.Set(key, value, ttl)
	return resp.EncodeSimpleString("OK")
}

func (c *Conn) handleGet(args []resp.Value) []byte {
	if len(args) != 1 {
		return resp.EncodeError("ERR wrong number of arguments for 'get' command")
	}

	key := args[0].Str
	value, ok := c.store.Get(key)
	if !ok {
		return resp.EncodeNullBulkString()
	}

	return resp.EncodeBulkString(string(value))
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
