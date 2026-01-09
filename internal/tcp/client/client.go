package client

import (
	"log"
	"net"
	"sync"

	"github.com/briheet/kquetolk/internal/resp"
)

type Conner interface {
	HandleConnection() error
	IsWriteable() bool
	Close() error
	Write() error
}

type Conn struct {
	conn net.Conn

	mu       sync.RWMutex
	readBuf  []byte
	writeBuf []byte

	resp resp.Resp
}

func NewClientConnection(conn net.Conn) *Conn {
	return &Conn{
		conn:     conn,
		readBuf:  make([]byte, 1024),
		writeBuf: make([]byte, 1024),
		resp:     resp.NewResp(),
	}
}

func (c *Conn) HandleConnection() error {

	readBuf := make([]byte, 1024)

	for {
		n, err := c.conn.Read(readBuf)
		if err != nil {
			return err
		}

		c.mu.Lock()
		c.readBuf = append(c.readBuf, readBuf[:n]...)

		for {

			out, consumed, err := c.resp.HandleParsing(&c.readBuf)
			if err != nil {
				c.mu.Unlock()
				return err
			}

			if consumed == 0 {
				break
			}

			c.readBuf = c.readBuf[consumed:]
			c.writeBuf = append(c.writeBuf, out...)
		}
		c.mu.Unlock()

		if err := c.flush(); err != nil {
			return err
		}
	}

}

func (c *Conn) flush() error {

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.writeBuf) == 0 {
		return nil
	}

	// Write back to connection for now
	if err := c.writeToTcp(); err != nil {
		log.Printf("error writing to conn: %v", err)
	}

	return nil
}

func (c *Conn) Close() error {
	defer c.conn.Close()
	return nil
}

func (c *Conn) writeToTcp() error {

	log.Println("reaching the write, dont worry")

	_, err := c.conn.Write(c.writeBuf)
	if err != nil {
		return err
	}

	return nil
}
