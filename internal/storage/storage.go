package storage

import (
	"errors"
	"hash/fnv"
	"strconv"
	"sync"
	"time"
)

var (
	ErrNotInteger = errors.New("value is not an integer")
)

const (
	defaultNumShards       = 32
	defaultCleanupInterval = time.Minute
)

type Shard struct {
	mu   sync.RWMutex
	data map[string]Entry
}

type ShardedMap struct {
	shards    []*Shard
	numShards uint64
	stopCh    chan struct{}
}

func New(opts ...Option) *ShardedMap {
	cfg := config{
		numShards:       defaultNumShards,
		cleanupInterval: defaultCleanupInterval,
	}

	for _, opt := range opts {
		opt.apply(&cfg)
	}

	shards := make([]*Shard, cfg.numShards)
	for i := range shards {
		shards[i] = &Shard{
			data: make(map[string]Entry),
		}
	}

	sm := &ShardedMap{
		shards:    shards,
		numShards: cfg.numShards,
		stopCh:    make(chan struct{}),
	}

	sm.startCleanup(cfg.cleanupInterval)

	return sm
}

func (sm *ShardedMap) getShard(key string) *Shard {
	h := fnv.New64a()
	h.Write([]byte(key))
	return sm.shards[h.Sum64()%sm.numShards]
}

func (sm *ShardedMap) Set(key string, value []byte, ttl time.Duration) {
	shard := sm.getShard(key)

	var expiresAt int64
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl).UnixNano()
	}

	shard.mu.Lock()
	defer shard.mu.Unlock()

	shard.data[key] = Entry{
		Value:     value,
		ExpiresAt: expiresAt,
	}
}

func (sm *ShardedMap) Get(key string) ([]byte, bool) {
	shard := sm.getShard(key)

	shard.mu.RLock()
	entry, ok := shard.data[key]
	expired := ok && entry.IsExpired()
	shard.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if expired {
		sm.Del(key)
		return nil, false
	}

	return entry.Value, true
}

func (sm *ShardedMap) Del(key string) bool {
	shard := sm.getShard(key)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	_, ok := shard.data[key]
	delete(shard.data, key)
	return ok
}

func (sm *ShardedMap) Exists(key string) bool {
	shard := sm.getShard(key)

	shard.mu.RLock()
	entry, ok := shard.data[key]
	expired := ok && entry.IsExpired()
	shard.mu.RUnlock()

	if !ok {
		return false
	}

	if expired {
		sm.Del(key)
		return false
	}

	return true
}

func (sm *ShardedMap) Incr(key string) (int64, error) {
	shard := sm.getShard(key)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	entry, ok := shard.data[key]

	if !ok || entry.IsExpired() {
		shard.data[key] = Entry{
			Value:     []byte("1"),
			ExpiresAt: 0,
		}
		return 1, nil
	}

	val, err := strconv.ParseInt(string(entry.Value), 10, 64)
	if err != nil {
		return 0, ErrNotInteger
	}

	val++
	shard.data[key] = Entry{
		Value:     []byte(strconv.FormatInt(val, 10)),
		ExpiresAt: entry.ExpiresAt,
	}

	return val, nil
}

func (sm *ShardedMap) Keys() []string {
	now := time.Now().UnixNano()
	var keys []string

	for _, shard := range sm.shards {
		keys = shard.collectKeys(keys, now)
	}

	return keys
}

func (s *Shard) collectKeys(keys []string, now int64) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for k, v := range s.data {
		if v.ExpiresAt == 0 || now <= v.ExpiresAt {
			keys = append(keys, k)
		}
	}
	return keys
}

func (sm *ShardedMap) startCleanup(interval time.Duration) {
	for i := range sm.shards {
		go sm.shardCleaner(sm.shards[i], interval)
	}
}

func (sm *ShardedMap) shardCleaner(shard *Shard, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			shard.evictExpired()
		case <-sm.stopCh:
			return
		}
	}
}

func (s *Shard) evictExpired() {
	now := time.Now().UnixNano()

	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range s.data {
		if v.ExpiresAt != 0 && now > v.ExpiresAt {
			delete(s.data, k)
		}
	}
}

func (sm *ShardedMap) Close() {
	close(sm.stopCh)
}
