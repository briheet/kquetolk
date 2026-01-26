package gnettcp

import (
	"github.com/briheet/kquetolk/internal/storage"
	"github.com/panjf2000/gnet/v2"
)

var _ gnet.EventHandler = (*GnetServer)(nil)

type Option interface {
	apply(*options)
}

type options struct {
	multiCore bool
	addr      string
	store     *storage.ShardedMap
}

type multicoreOption bool

func (m multicoreOption) apply(opts *options) {
	opts.multiCore = bool(m)
}

func WithMultiCore(multicore bool) Option {
	return multicoreOption(multicore)
}

type addrOption string

func (a addrOption) apply(opts *options) {
	opts.addr = string(a)
}

func WithAddr(addr string) Option {
	return addrOption(addr)
}

type storeOption struct {
	store *storage.ShardedMap
}

func (s storeOption) apply(opts *options) {
	opts.store = s.store
}

func WithStore(store *storage.ShardedMap) Option {
	return storeOption{store: store}
}
