#!/usr/bin/env python3
"""
Plot benchmark results with comprehensive visualizations.

Usage:
    python3 plot_benchmark.py kquetolk_results.csv [redis_results.csv]

Generates:
    1. Throughput vs Clients (side-by-side comparison)
    2. Throughput overlay (both on same graph)
    3. P99 Latency vs Clients (critical for SLA evaluation)
    4. Latency percentile breakdown (p50 vs p99 vs p99.9)
    5. Throughput vs Latency tradeoff curve
"""

import sys
import csv
import matplotlib.pyplot as plt
import matplotlib.ticker as ticker


def read_csv(filename):
    """Read benchmark CSV and return data dict."""
    data = {
        'clients': [],
        'ops_sec': [],
        'avg_latency': [],
        'p50_latency': [],
        'p99_latency': [],
        'p999_latency': []
    }

    with open(filename, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            data['clients'].append(int(row['client_count']))
            data['ops_sec'].append(float(row['ops_sec']))
            data['avg_latency'].append(float(row.get('avg_latency_ms', 0)))
            # Handle both old format (only p99) and new format (p50, p99, p99.9)
            data['p50_latency'].append(float(row.get('p50_latency_ms', row.get('avg_latency_ms', 0))))
            data['p99_latency'].append(float(row.get('p99_latency_ms', 0)))
            data['p999_latency'].append(float(row.get('p999_latency_ms', row.get('p99_latency_ms', 0))))

    return data


def format_ops(x, p):
    """Format ops/sec for axis labels."""
    if x >= 1e6:
        return f'{x/1e6:.1f}M'
    elif x >= 1e3:
        return f'{x/1e3:.0f}K'
    return f'{x:.0f}'


def format_latency(x, p):
    """Format latency for axis labels."""
    if x >= 1000:
        return f'{x/1000:.1f}s'
    return f'{x:.0f}ms'


def plot_throughput_comparison(k_data, r_data, output_file='throughput_comparison.png'):
    """Plot side-by-side throughput comparison."""
    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(14, 5))

    # Kquetolk
    ax1.plot(k_data['clients'], k_data['ops_sec'], 'o-', linewidth=2, markersize=8, color='#2563eb')
    ax1.set_xlabel('Number of Clients', fontsize=11)
    ax1.set_ylabel('Throughput (ops/sec)', fontsize=11)
    ax1.set_title('Kquetolk', fontsize=13, fontweight='bold')
    ax1.set_xscale('log')
    ax1.xaxis.set_major_formatter(ticker.ScalarFormatter())
    ax1.yaxis.set_major_formatter(ticker.FuncFormatter(format_ops))
    ax1.grid(True, alpha=0.3)

    # Redis
    ax2.plot(r_data['clients'], r_data['ops_sec'], 'o-', linewidth=2, markersize=8, color='#dc2626')
    ax2.set_xlabel('Number of Clients', fontsize=11)
    ax2.set_ylabel('Throughput (ops/sec)', fontsize=11)
    ax2.set_title('Redis', fontsize=13, fontweight='bold')
    ax2.set_xscale('log')
    ax2.xaxis.set_major_formatter(ticker.ScalarFormatter())
    ax2.yaxis.set_major_formatter(ticker.FuncFormatter(format_ops))
    ax2.grid(True, alpha=0.3)

    # Match y-axis scales
    max_ops = max(max(k_data['ops_sec']), max(r_data['ops_sec'])) * 1.1
    ax1.set_ylim(0, max_ops)
    ax2.set_ylim(0, max_ops)

    plt.suptitle('Throughput vs Client Count Comparison', fontsize=14, y=1.02)
    plt.tight_layout()
    plt.savefig(output_file, dpi=150, bbox_inches='tight')
    print(f"Saved: {output_file}")
    plt.close()


def plot_throughput_overlay(k_data, r_data, output_file='throughput_overlay.png'):
    """Plot both servers on the same throughput graph."""
    fig, ax = plt.subplots(figsize=(10, 6))

    ax.plot(k_data['clients'], k_data['ops_sec'], 'o-', linewidth=2, markersize=8,
            color='#2563eb', label='Kquetolk')
    ax.plot(r_data['clients'], r_data['ops_sec'], 's-', linewidth=2, markersize=8,
            color='#dc2626', label='Redis')

    ax.set_xlabel('Number of Clients', fontsize=12)
    ax.set_ylabel('Throughput (ops/sec)', fontsize=12)
    ax.set_title('Throughput vs Client Count', fontsize=14)
    ax.set_xscale('log')
    ax.xaxis.set_major_formatter(ticker.ScalarFormatter())
    ax.yaxis.set_major_formatter(ticker.FuncFormatter(format_ops))
    ax.grid(True, alpha=0.3)
    ax.legend(loc='best', fontsize=11)

    plt.tight_layout()
    plt.savefig(output_file, dpi=150)
    print(f"Saved: {output_file}")
    plt.close()


def plot_latency_comparison(k_data, r_data, output_file='latency_p99_comparison.png'):
    """Plot P99 latency comparison - critical for SLA evaluation."""
    fig, ax = plt.subplots(figsize=(10, 6))

    ax.plot(k_data['clients'], k_data['p99_latency'], 'o-', linewidth=2, markersize=8,
            color='#2563eb', label='Kquetolk P99')
    ax.plot(r_data['clients'], r_data['p99_latency'], 's-', linewidth=2, markersize=8,
            color='#dc2626', label='Redis P99')

    ax.set_xlabel('Number of Clients', fontsize=12)
    ax.set_ylabel('P99 Latency (ms)', fontsize=12)
    ax.set_title('P99 Latency vs Client Count', fontsize=14)
    ax.set_xscale('log')
    ax.xaxis.set_major_formatter(ticker.ScalarFormatter())
    ax.grid(True, alpha=0.3)
    ax.legend(loc='best', fontsize=11)

    # Add SLA reference line at 10ms (common Redis SLA target)
    ax.axhline(y=10, color='green', linestyle='--', alpha=0.5, label='10ms SLA target')

    plt.tight_layout()
    plt.savefig(output_file, dpi=150)
    print(f"Saved: {output_file}")
    plt.close()


def plot_latency_percentiles(data, name, color, output_file='latency_percentiles.png'):
    """Plot latency percentile breakdown for a single server."""
    fig, ax = plt.subplots(figsize=(10, 6))

    ax.plot(data['clients'], data['p50_latency'], 'o-', linewidth=2, markersize=6,
            color=color, alpha=0.5, label='P50 (median)')
    ax.plot(data['clients'], data['p99_latency'], 's-', linewidth=2, markersize=6,
            color=color, alpha=0.75, label='P99')
    ax.plot(data['clients'], data['p999_latency'], '^-', linewidth=2, markersize=6,
            color=color, label='P99.9 (tail)')

    ax.set_xlabel('Number of Clients', fontsize=12)
    ax.set_ylabel('Latency (ms)', fontsize=12)
    ax.set_title(f'{name}: Latency Percentiles vs Client Count', fontsize=14)
    ax.set_xscale('log')
    ax.xaxis.set_major_formatter(ticker.ScalarFormatter())
    ax.grid(True, alpha=0.3)
    ax.legend(loc='upper left', fontsize=11)

    plt.tight_layout()
    plt.savefig(output_file, dpi=150)
    print(f"Saved: {output_file}")
    plt.close()


def plot_throughput_latency_tradeoff(k_data, r_data, output_file='throughput_latency_tradeoff.png'):
    """Plot throughput vs latency - the classic tradeoff curve.

    This is one of the most important visualizations:
    - X-axis: Throughput (what you get)
    - Y-axis: P99 Latency (what you pay)
    - Lower and to the right is better
    """
    fig, ax = plt.subplots(figsize=(10, 6))

    # Plot with client count as annotation
    ax.plot(k_data['ops_sec'], k_data['p99_latency'], 'o-', linewidth=2, markersize=10,
            color='#2563eb', label='Kquetolk')
    ax.plot(r_data['ops_sec'], r_data['p99_latency'], 's-', linewidth=2, markersize=10,
            color='#dc2626', label='Redis')

    # Annotate key points with client counts
    for i, clients in enumerate(k_data['clients']):
        if clients in [100, 1000, 10000]:
            ax.annotate(f'{clients}c', (k_data['ops_sec'][i], k_data['p99_latency'][i]),
                       textcoords="offset points", xytext=(5, 5), fontsize=8, color='#2563eb')
    for i, clients in enumerate(r_data['clients']):
        if clients in [100, 1000, 10000]:
            ax.annotate(f'{clients}c', (r_data['ops_sec'][i], r_data['p99_latency'][i]),
                       textcoords="offset points", xytext=(5, 5), fontsize=8, color='#dc2626')

    ax.set_xlabel('Throughput (ops/sec)', fontsize=12)
    ax.set_ylabel('P99 Latency (ms)', fontsize=12)
    ax.set_title('Throughput vs Latency Tradeoff\n(lower-right is better)', fontsize=14)
    ax.xaxis.set_major_formatter(ticker.FuncFormatter(format_ops))
    ax.grid(True, alpha=0.3)
    ax.legend(loc='upper left', fontsize=11)

    # Add quadrant annotations
    ax.text(0.95, 0.05, 'IDEAL\n(high throughput,\nlow latency)',
            transform=ax.transAxes, fontsize=9, alpha=0.5, ha='right', va='bottom')

    plt.tight_layout()
    plt.savefig(output_file, dpi=150)
    print(f"Saved: {output_file}")
    plt.close()


def plot_single(kquetolk_file, suffix=''):
    """Plot single server results."""
    k_data = read_csv(kquetolk_file)

    # Throughput graph
    fig, ax = plt.subplots(figsize=(10, 6))
    ax.plot(k_data['clients'], k_data['ops_sec'], 'o-', linewidth=2, markersize=8, color='#2563eb')
    ax.set_xlabel('Number of Clients', fontsize=12)
    ax.set_ylabel('Throughput (ops/sec)', fontsize=12)
    ax.set_title('Kquetolk: Throughput vs Client Count', fontsize=14)
    ax.set_xscale('log')
    ax.xaxis.set_major_formatter(ticker.ScalarFormatter())
    ax.yaxis.set_major_formatter(ticker.FuncFormatter(format_ops))
    ax.grid(True, alpha=0.3)
    plt.tight_layout()
    plt.savefig(f'throughput_graph{suffix}.png', dpi=150)
    print(f"Saved: throughput_graph{suffix}.png")
    plt.close()

    # Latency percentiles
    plot_latency_percentiles(k_data, 'Kquetolk', '#2563eb', f'latency_percentiles{suffix}.png')


def plot_all(kquetolk_file, redis_file, suffix=''):
    """Generate all comparison plots."""
    k_data = read_csv(kquetolk_file)
    r_data = read_csv(redis_file)

    # 1. Throughput comparison (side-by-side)
    plot_throughput_comparison(k_data, r_data, f'throughput_comparison{suffix}.png')

    # 2. Throughput overlay
    plot_throughput_overlay(k_data, r_data, f'throughput_overlay{suffix}.png')

    # 3. P99 Latency comparison
    plot_latency_comparison(k_data, r_data, f'latency_p99_comparison{suffix}.png')

    # 4. Latency percentiles for each
    plot_latency_percentiles(k_data, 'Kquetolk', '#2563eb', f'latency_percentiles_kquetolk{suffix}.png')
    plot_latency_percentiles(r_data, 'Redis', '#dc2626', f'latency_percentiles_redis{suffix}.png')

    # 5. Throughput vs Latency tradeoff
    plot_throughput_latency_tradeoff(k_data, r_data, f'throughput_latency_tradeoff{suffix}.png')

    print(f"\nGenerated 6 graphs with suffix '{suffix}'")


if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Usage:")
        print("  python3 plot_benchmark.py kquetolk_results.csv")
        print("  python3 plot_benchmark.py kquetolk_results.csv redis_results.csv")
        print("  python3 plot_benchmark.py kquetolk_results.csv redis_results.csv _gnet")
        sys.exit(1)

    suffix = ''
    if len(sys.argv) >= 4:
        suffix = sys.argv[3]

    if len(sys.argv) == 2:
        plot_single(sys.argv[1], suffix)
    else:
        plot_all(sys.argv[1], sys.argv[2], suffix)
