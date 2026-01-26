#!/bin/bash

# Benchmark script for Kquetolk - generates throughput vs client count data
# Usage: ./run_benchmark.sh [server_port] [output_file]
#
# Best practices from Redis/memtier documentation:
# - Use multiple iterations (-x) for reliable results
# - Pipeline 10-20 for high throughput testing
# - Capture multiple latency percentiles
# - Default 1:10 SET:GET ratio

SERVER_HOST="127.0.0.1"
SERVER_PORT="${1:-6379}"
OUTPUT_FILE="${2:-benchmark_results.csv}"

# Client counts to test (threads * clients per thread)
# Using 4 threads, varying clients per thread
THREADS=4
CLIENTS_PER_THREAD=(25 50 125 250 500 625 1250 2500)  # Total: 100, 200, 500, 1000, 2000, 2500, 5000, 10000

# Benchmark settings
PIPELINE=16
REQUESTS=10000      # Per client - 10k is memtier default, good for burst testing
ITERATIONS=3        # Run 3x and take best result for reliability

echo "client_count,ops_sec,avg_latency_ms,p50_latency_ms,p99_latency_ms,p999_latency_ms" > "$OUTPUT_FILE"

for c in "${CLIENTS_PER_THREAD[@]}"; do
    total_clients=$((THREADS * c))
    echo "Testing with $total_clients clients ($THREADS threads x $c clients)..."

    # Run memtier_benchmark with multiple iterations and percentile output
    output=$(memtier_benchmark \
        -s "$SERVER_HOST" \
        -p "$SERVER_PORT" \
        -t "$THREADS" \
        -c "$c" \
        --pipeline="$PIPELINE" \
        -n "$REQUESTS" \
        -x "$ITERATIONS" \
        --print-percentiles=50,99,99.9 \
        --hide-histogram \
        2>&1)

    # Parse the BEST RUN Totals line (memtier outputs this with -x > 1)
    # If no BEST RUN, fall back to regular Totals
    if echo "$output" | grep -q "BEST RUN"; then
        totals=$(echo "$output" | sed -n '/BEST RUN/,/^$/p' | grep "^Totals" | tail -1)
    else
        totals=$(echo "$output" | grep "^Totals" | tail -1)
    fi

    # Parse percentile lines
    # Format: p50     0.12300
    p50=$(echo "$output" | grep -E "^p50\s" | tail -1 | awk '{print $2}')
    p99=$(echo "$output" | grep -E "^p99\s" | tail -1 | awk '{print $2}')
    p999=$(echo "$output" | grep -E "^p99\.9\s" | tail -1 | awk '{print $2}')

    if [ -n "$totals" ]; then
        ops_sec=$(echo "$totals" | awk '{print $4}')
        avg_latency=$(echo "$totals" | awk '{print $5}')

        # Use parsed percentiles, fallback to totals p99 if not found
        [ -z "$p50" ] && p50="$avg_latency"
        [ -z "$p99" ] && p99=$(echo "$totals" | awk '{print $6}')
        [ -z "$p999" ] && p999="$p99"

        echo "$total_clients,$ops_sec,$avg_latency,$p50,$p99,$p999" >> "$OUTPUT_FILE"
        echo "  -> $ops_sec ops/sec, avg: ${avg_latency}ms, p50: ${p50}ms, p99: ${p99}ms, p99.9: ${p999}ms"
    else
        echo "  -> Failed to parse output"
        echo "DEBUG: $output" | head -50
    fi

    # Brief pause between tests
    sleep 2
done

echo ""
echo "Results saved to $OUTPUT_FILE"
echo "Run: python3 benchmark/plot_benchmark.py $OUTPUT_FILE"
