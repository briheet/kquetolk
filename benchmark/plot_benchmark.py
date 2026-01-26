#!/usr/bin/env python3
"""
Plot benchmark results: Throughput vs Client Count

Usage:
    python3 plot_benchmark.py kquetolk_results.csv [redis_results.csv]

This generates a graph similar to typical Redis benchmark comparisons.
"""

import sys
import csv
import matplotlib.pyplot as plt
import matplotlib.ticker as ticker


def read_csv(filename):
    """Read benchmark CSV and return data."""
    clients = []
    ops_sec = []

    with open(filename, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            clients.append(int(row['client_count']))
            ops_sec.append(float(row['ops_sec']))

    return clients, ops_sec


def plot_single(kquetolk_file, output_file='throughput_graph.png'):
    """Plot single server results."""
    clients, ops = read_csv(kquetolk_file)

    fig, ax = plt.subplots(figsize=(10, 6))

    ax.plot(clients, ops, 'o-', linewidth=2, markersize=8,
            color='#2563eb', label='Kquetolk')

    ax.set_xlabel('Number of Clients', fontsize=12)
    ax.set_ylabel('Throughput (ops/sec)', fontsize=12)
    ax.set_title('Kquetolk: Throughput vs Client Count', fontsize=14)

    ax.set_xscale('log')
    ax.xaxis.set_major_formatter(ticker.ScalarFormatter())
    ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x, p: f'{x/1e6:.1f}M' if x >= 1e6 else f'{x/1e3:.0f}K'))

    ax.grid(True, alpha=0.3)
    ax.legend(loc='best')

    plt.tight_layout()
    plt.savefig(output_file, dpi=150)
    print(f"Graph saved to {output_file}")
    plt.show()


def plot_comparison(kquetolk_file, redis_file, output_file='throughput_comparison.png'):
    """Plot comparison between Kquetolk and Redis."""
    k_clients, k_ops = read_csv(kquetolk_file)
    r_clients, r_ops = read_csv(redis_file)

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(14, 5))

    # Left plot: Kquetolk
    ax1.plot(k_clients, k_ops, 'o-', linewidth=2, markersize=8, color='#2563eb')
    ax1.set_xlabel('Number of Clients', fontsize=11)
    ax1.set_ylabel('Throughput (ops/sec)', fontsize=11)
    ax1.set_title('Kquetolk', fontsize=13, fontweight='bold')
    ax1.set_xscale('log')
    ax1.xaxis.set_major_formatter(ticker.ScalarFormatter())
    ax1.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x, p: f'{x/1e6:.1f}M' if x >= 1e6 else f'{x/1e3:.0f}K'))
    ax1.grid(True, alpha=0.3)

    # Right plot: Redis
    ax2.plot(r_clients, r_ops, 'o-', linewidth=2, markersize=8, color='#dc2626')
    ax2.set_xlabel('Number of Clients', fontsize=11)
    ax2.set_ylabel('Throughput (ops/sec)', fontsize=11)
    ax2.set_title('Redis', fontsize=13, fontweight='bold')
    ax2.set_xscale('log')
    ax2.xaxis.set_major_formatter(ticker.ScalarFormatter())
    ax2.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x, p: f'{x/1e6:.1f}M' if x >= 1e6 else f'{x/1e3:.0f}K'))
    ax2.grid(True, alpha=0.3)

    # Match y-axis scales for fair comparison
    max_ops = max(max(k_ops), max(r_ops)) * 1.1
    ax1.set_ylim(0, max_ops)
    ax2.set_ylim(0, max_ops)

    plt.suptitle('Throughput vs Client Count Comparison', fontsize=14, y=1.02)
    plt.tight_layout()
    plt.savefig(output_file, dpi=150, bbox_inches='tight')
    print(f"Comparison graph saved to {output_file}")
    plt.show()


def plot_overlay(kquetolk_file, redis_file, output_file='throughput_overlay.png'):
    """Plot both servers on the same graph."""
    k_clients, k_ops = read_csv(kquetolk_file)
    r_clients, r_ops = read_csv(redis_file)

    fig, ax = plt.subplots(figsize=(10, 6))

    ax.plot(k_clients, k_ops, 'o-', linewidth=2, markersize=8,
            color='#2563eb', label='Kquetolk')
    ax.plot(r_clients, r_ops, 's-', linewidth=2, markersize=8,
            color='#dc2626', label='Redis')

    ax.set_xlabel('Number of Clients', fontsize=12)
    ax.set_ylabel('Throughput (ops/sec)', fontsize=12)
    ax.set_title('Throughput vs Client Count', fontsize=14)

    ax.set_xscale('log')
    ax.xaxis.set_major_formatter(ticker.ScalarFormatter())
    ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x, p: f'{x/1e6:.1f}M' if x >= 1e6 else f'{x/1e3:.0f}K'))

    ax.grid(True, alpha=0.3)
    ax.legend(loc='best', fontsize=11)

    plt.tight_layout()
    plt.savefig(output_file, dpi=150)
    print(f"Overlay graph saved to {output_file}")
    plt.show()


if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Usage:")
        print("  python3 plot_benchmark.py kquetolk_results.csv")
        print("  python3 plot_benchmark.py kquetolk_results.csv redis_results.csv")
        sys.exit(1)

    if len(sys.argv) == 2:
        plot_single(sys.argv[1])
    else:
        plot_comparison(sys.argv[1], sys.argv[2])
        plot_overlay(sys.argv[1], sys.argv[2])
