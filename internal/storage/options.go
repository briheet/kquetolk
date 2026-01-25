package storage

import "time"

type config struct {
	numShards       uint64
	cleanupInterval time.Duration
}

type Option interface {
	apply(*config)
}

type shardsOption uint64

func (s shardsOption) apply(c *config) {
	c.numShards = uint64(s)
}

func WithShards(n uint64) Option {
	return shardsOption(n)
}

type cleanupIntervalOption time.Duration

func (ci cleanupIntervalOption) apply(c *config) {
	c.cleanupInterval = time.Duration(ci)
}

func WithCleanupInterval(d time.Duration) Option {
	return cleanupIntervalOption(d)
}
