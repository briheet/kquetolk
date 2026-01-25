# Kquetolk

A Redis-compatible in-memory server implementing the RESP protocol in Go.

## Features

- RESP protocol parser
- Sharded in-memory storage (32 shards, FNV-1a hashing)
- TTL support with lazy + parallel eviction (one worker per shard)
- Commands: `PING`, `ECHO`, `SET`, `GET`

## Usage

```sh
make build && make run
```

## Benchmark

Tested on Apple M4 Pro with `memtier_benchmark` (pipeline=16, 1:10 SET:GET ratio).

### 1,000 Concurrent Connections

```
memtier_benchmark -s 127.0.0.1 -p 6379 -t 4 -c 250 --pipeline=16 -n 100000
```

| Operation | Ops/sec | Avg Latency | p99 Latency |
|-----------|---------|-------------|-------------|
| SET | 227K | 6.39 ms | 9.66 ms |
| GET | 2.27M | 6.39 ms | 9.60 ms |
| **Total** | **2.50M** | **6.39 ms** | **9.60 ms** |

### 10,000 Concurrent Connections

```
memtier_benchmark -s 127.0.0.1 -p 6379 -t 4 -c 2500 --pipeline=16
```

| Operation | Ops/sec | Avg Latency | p99 Latency |
|-----------|---------|-------------|-------------|
| SET | 130K | 133 ms | 930 ms |
| GET | 1.30M | 133 ms | 934 ms |
| **Total** | **1.43M** | **133 ms** | **930 ms** |

### Comparison with Redis

Benchmarked against `redis:latest` Docker image under identical conditions.

#### 1,000 Connections

| Server | Ops/sec | Avg Latency | p99 Latency |
|--------|---------|-------------|-------------|
| **Kquetolk** | **2.50M** | **6.39 ms** | **9.60 ms** |
| Redis | 1.71M | 10.18 ms | 18.43 ms |

#### 10,000 Connections

| Server | Ops/sec | Avg Latency | p99 Latency |
|--------|---------|-------------|-------------|
| Kquetolk | 1.43M | 133 ms | 930 ms |
| **Redis** | **1.76M** | **107 ms** | **545 ms** |

### Analysis

- **1K connections**: Kquetolk is **46% faster** with 37% lower latency — goroutines handle moderate concurrency efficiently
- **10K connections**: Redis is **23% faster** with better tail latency — its single-threaded event loop (epoll/kqueue) scales better at high connection counts
- Goroutine-per-connection incurs overhead from context switching and lock contention at scale
