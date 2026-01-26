package gnettcp

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/briheet/kquetolk/internal/resp"
	"github.com/briheet/kquetolk/internal/storage"
	"github.com/panjf2000/gnet/v2"
)

type GnetServer struct {
	eng       gnet.Engine
	store     *storage.ShardedMap
	MultiCore bool
	Addr      string
}

func NewGnetServer(opts ...Option) *GnetServer {
	cfg := &options{
		multiCore: true,
		addr:      ":6379",
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return &GnetServer{
		store:     cfg.store,
		MultiCore: cfg.multiCore,
		Addr:      cfg.addr,
	}
}

func (gs *GnetServer) OnBoot(eng gnet.Engine) gnet.Action {
	gs.eng = eng
	return gnet.None
}

func (gs *GnetServer) OnShutdown(eng gnet.Engine) {
}

func (gs *GnetServer) Stop(ctx context.Context) error {
	return gs.eng.Stop(ctx)
}

func (gs *GnetServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	return nil, gnet.None
}

func (gs *GnetServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	return gnet.None
}

func (gs *GnetServer) OnTraffic(c gnet.Conn) gnet.Action {
	buf, _ := c.Peek(-1)
	if len(buf) == 0 {
		return gnet.None
	}

	var responses []byte
	offset := 0

	for offset < len(buf) {
		val, n, err := resp.Parse(buf[offset:])
		if err != nil {
			if err == resp.ErrIncomplete {
				break
			}
			responses = append(responses, resp.EncodeError("ERR "+err.Error())...)
			c.Discard(-1)
			c.Write(responses)
			return gnet.None
		}

		offset += n
		responses = append(responses, gs.handleCommand(val)...)
	}

	if offset > 0 {
		c.Discard(offset)
		c.Write(responses)
	}

	return gnet.None
}

func (gs *GnetServer) OnTick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.None
}

func (gs *GnetServer) handleCommand(val resp.Value) []byte {
	if val.Type != '*' || len(val.Array) == 0 {
		return resp.EncodeError("ERR invalid command format")
	}

	cmd := strings.ToUpper(val.Array[0].Str)
	args := val.Array[1:]

	switch cmd {
	case "PING":
		return gs.cmdPing(args)
	case "ECHO":
		return gs.cmdEcho(args)
	case "SET":
		return gs.cmdSet(args)
	case "GET":
		return gs.cmdGet(args)
	case "DEL":
		return gs.cmdDel(args)
	case "EXISTS":
		return gs.cmdExists(args)
	case "INCR":
		return gs.cmdIncr(args)
	case "KEYS":
		return gs.cmdKeys(args)
	case "COMMAND":
		return resp.EncodeSimpleString("OK")
	default:
		return resp.EncodeError("ERR unknown command '" + cmd + "'")
	}
}

func (gs *GnetServer) cmdPing(args []resp.Value) []byte {
	if len(args) == 0 {
		return resp.EncodeSimpleString("PONG")
	}
	return resp.EncodeBulkString(args[0].Str)
}

func (gs *GnetServer) cmdEcho(args []resp.Value) []byte {
	if len(args) != 1 {
		return resp.EncodeError("ERR wrong number of arguments for 'echo' command")
	}
	return resp.EncodeBulkString(args[0].Str)
}

func (gs *GnetServer) cmdSet(args []resp.Value) []byte {
	if len(args) < 2 {
		return resp.EncodeError("ERR wrong number of arguments for 'set' command")
	}

	key := args[0].Str
	value := args[1].Str
	var ttl time.Duration

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

	gs.store.Set(key, []byte(value), ttl)
	return resp.EncodeSimpleString("OK")
}

func (gs *GnetServer) cmdGet(args []resp.Value) []byte {
	if len(args) != 1 {
		return resp.EncodeError("ERR wrong number of arguments for 'get' command")
	}

	value, ok := gs.store.Get(args[0].Str)
	if !ok {
		return resp.EncodeNullBulkString()
	}
	return resp.EncodeBulkString(string(value))
}

func (gs *GnetServer) cmdDel(args []resp.Value) []byte {
	if len(args) == 0 {
		return resp.EncodeError("ERR wrong number of arguments for 'del' command")
	}

	count := 0
	for _, arg := range args {
		if gs.store.Del(arg.Str) {
			count++
		}
	}
	return resp.EncodeInteger(int64(count))
}

func (gs *GnetServer) cmdExists(args []resp.Value) []byte {
	if len(args) == 0 {
		return resp.EncodeError("ERR wrong number of arguments for 'exists' command")
	}

	count := 0
	for _, arg := range args {
		if gs.store.Exists(arg.Str) {
			count++
		}
	}
	return resp.EncodeInteger(int64(count))
}

func (gs *GnetServer) cmdIncr(args []resp.Value) []byte {
	if len(args) != 1 {
		return resp.EncodeError("ERR wrong number of arguments for 'incr' command")
	}

	val, err := gs.store.Incr(args[0].Str)
	if err != nil {
		return resp.EncodeError("ERR " + err.Error())
	}
	return resp.EncodeInteger(val)
}

func (gs *GnetServer) cmdKeys(args []resp.Value) []byte {
	if len(args) != 1 {
		return resp.EncodeError("ERR wrong number of arguments for 'keys' command")
	}

	keys := gs.store.Keys()
	return resp.EncodeArray(keys)
}
