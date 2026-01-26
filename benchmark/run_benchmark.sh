#!/bin/bash

# Benchmark script for Kquetolk - generates throughput vs client count data
# Usage: ./run_benchmark.sh [server_port] [output_file]

SERVER_HOST="127.0.0.1"
SERVER_PORT="${1:-6379}"
OUTPUT_FILE="${2:-benchmark_results.csv}"

# Client counts to test (threads * clients per thread)
# Using 4 threads, varying clients per thread
THREADS=4
CLIENTS_PER_THREAD=(25 50 125 250 500 625 1250 2500)  # Total: 100, 200, 500, 1000, 2000, 2500, 5000, 10000

# Benchmark settings
PIPELINE=16
REQUESTS=100000

echo "client_count,ops_sec,avg_latency_ms,p99_latency_ms" > "$OUTPUT_FILE"

for c in "${CLIENTS_PER_THREAD[@]}"; do
    total_clients=$((THREADS * c))
    echo "Testing with $total_clients clients ($THREADS threads x $c clients)..."

    # Run memtier_benchmark and capture output
    output=$(memtier_benchmark \
        -s "$SERVER_HOST" \
        -p "$SERVER_PORT" \
        -t "$THREADS" \
        -c "$c" \
        --pipeline="$PIPELINE" \
        -n "$REQUESTS" \
        --hide-histogram \
        2>&1)

    # Parse the Totals line for ops/sec and latency
    # Format: Totals    227272.73    2272727.27    2500000.00    6.390    9.600
    totals=$(echo "$output" | grep "^Totals" | tail -1)

    if [ -n "$totals" ]; then
        ops_sec=$(echo "$totals" | awk '{print $4}')
        avg_latency=$(echo "$totals" | awk '{print $5}')
        p99_latency=$(echo "$totals" | awk '{print $6}')

        echo "$total_clients,$ops_sec,$avg_latency,$p99_latency" >> "$OUTPUT_FILE"
        echo "  -> $ops_sec ops/sec, avg: ${avg_latency}ms, p99: ${p99_latency}ms"
    else
        echo "  -> Failed to parse output"
    fi

    # Brief pause between tests
    sleep 2
done

echo ""
echo "Results saved to $OUTPUT_FILE"
echo "Run: python3 benchmark/plot_benchmark.py $OUTPUT_FILE"
