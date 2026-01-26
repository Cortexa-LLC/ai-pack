#!/usr/bin/env python3
"""
A2A Agent Server Performance Monitor

Monitors and reports on A2A agent server performance including:
- API call latency statistics
- Token usage (input/output)
- Task completion metrics
- Prompt caching effectiveness
- Server health status

Usage:
    python monitor-performance.py [options]

Options:
    --server-url URL    Server URL (default: http://localhost:8080)
    --log-file PATH     Log file path (default: /tmp/agent-server.log)
    --tail N            Show last N API calls (default: 30)
    --json              Output as JSON
    --watch             Continuous monitoring mode (updates every 10s)
"""

import argparse
import json
import re
import sys
import time
from collections import defaultdict
from datetime import datetime
from typing import Dict, List, Optional, Tuple
from urllib.request import urlopen
from urllib.error import URLError


class PerformanceMonitor:
    def __init__(self, server_url: str, log_file: str):
        self.server_url = server_url.rstrip('/')
        self.log_file = log_file

    def fetch_metrics(self) -> Optional[Dict]:
        """Fetch metrics from the server /metrics endpoint."""
        try:
            with urlopen(f"{self.server_url}/metrics", timeout=5) as response:
                return json.loads(response.read().decode())
        except (URLError, Exception) as e:
            print(f"⚠️  Could not fetch metrics: {e}", file=sys.stderr)
            return None

    def parse_log_file(self, tail_count: int = 30) -> Tuple[List[Dict], List[Dict]]:
        """Parse the log file for API calls and task completions."""
        api_calls = []
        task_completions = []

        try:
            with open(self.log_file, 'r') as f:
                lines = f.readlines()

                # Process lines in reverse to get most recent first
                for line in reversed(lines):
                    try:
                        log_entry = json.loads(line)

                        # Parse API call timing
                        if log_entry.get('msg') == 'agent_log':
                            message = log_entry.get('message', '')

                            # API call pattern: "API: 4580ms | in:59286 out:89"
                            api_match = re.search(r'API: (\d+)ms \| in:(\d+) out:(\d+)', message)
                            if api_match:
                                api_calls.append({
                                    'task_id': log_entry.get('task_id'),
                                    'time': log_entry.get('time'),
                                    'duration_ms': int(api_match.group(1)),
                                    'tokens_in': int(api_match.group(2)),
                                    'tokens_out': int(api_match.group(3)),
                                })

                                if len(api_calls) >= tail_count:
                                    break

                        # Task completion
                        elif log_entry.get('msg') == 'task_completed':
                            task_completions.append({
                                'task_id': log_entry.get('task_id'),
                                'role': log_entry.get('role'),
                                'status': log_entry.get('status'),
                                'duration_ms': log_entry.get('duration_ms'),
                                'time': log_entry.get('time'),
                            })

                    except json.JSONDecodeError:
                        continue

        except FileNotFoundError:
            print(f"⚠️  Log file not found: {self.log_file}", file=sys.stderr)
            return [], []

        # Reverse to get chronological order
        api_calls.reverse()
        task_completions.reverse()

        return api_calls, task_completions

    def calculate_api_statistics(self, api_calls: List[Dict]) -> Dict:
        """Calculate statistics from API calls."""
        if not api_calls:
            return {
                'count': 0,
                'avg_duration_ms': 0,
                'min_duration_ms': 0,
                'max_duration_ms': 0,
                'total_tokens_in': 0,
                'total_tokens_out': 0,
                'avg_tokens_in': 0,
                'avg_tokens_out': 0,
            }

        durations = [call['duration_ms'] for call in api_calls]
        tokens_in = [call['tokens_in'] for call in api_calls]
        tokens_out = [call['tokens_out'] for call in api_calls]

        return {
            'count': len(api_calls),
            'avg_duration_ms': int(sum(durations) / len(durations)),
            'min_duration_ms': min(durations),
            'max_duration_ms': max(durations),
            'total_tokens_in': sum(tokens_in),
            'total_tokens_out': sum(tokens_out),
            'avg_tokens_in': int(sum(tokens_in) / len(tokens_in)),
            'avg_tokens_out': int(sum(tokens_out) / len(tokens_out)),
        }

    def generate_report(self, tail_count: int = 30, json_output: bool = False) -> Dict:
        """Generate a comprehensive performance report."""
        metrics = self.fetch_metrics()
        api_calls, task_completions = self.parse_log_file(tail_count)
        api_stats = self.calculate_api_statistics(api_calls)

        # Calculate cache hit rate (approximation based on token ratio)
        cache_efficiency = "Unknown"
        if api_stats['avg_tokens_in'] > 0 and api_stats['avg_tokens_out'] > 0:
            ratio = api_stats['avg_tokens_in'] / api_stats['avg_tokens_out']
            if ratio > 50:
                cache_efficiency = "Excellent (likely caching)"
            elif ratio > 20:
                cache_efficiency = "Good"
            else:
                cache_efficiency = "Low (check caching)"

        report = {
            'timestamp': datetime.now().isoformat(),
            'server_url': self.server_url,
            'metrics': metrics,
            'api_statistics': {
                **api_stats,
                'sample_size': tail_count,
                'cache_efficiency': cache_efficiency,
            },
            'recent_tasks': task_completions[-10:] if task_completions else [],
        }

        return report

    def print_report(self, report: Dict):
        """Print a human-readable performance report."""
        print("=" * 70)
        print("A2A AGENT SERVER PERFORMANCE REPORT")
        print("=" * 70)
        print(f"Generated: {report['timestamp']}")
        print(f"Server: {report['server_url']}")
        print()

        # Server Metrics
        metrics = report['metrics']
        if metrics:
            print("📊 SERVER METRICS")
            print("-" * 70)
            print(f"  Tasks Spawned:     {metrics['tasks_spawned']}")
            print(f"  Tasks Completed:   {metrics['tasks_completed']} ✅")
            print(f"  Tasks Failed:      {metrics['tasks_failed']}")
            print(f"  Tasks In Progress: {metrics['tasks_in_progress']}")

            if metrics['tasks_completed'] > 0:
                print(f"  Average Duration:  {metrics['avg_duration_ms'] / 1000:.1f}s")

            print(f"  API Calls Total:   {metrics['api_calls_total']}")
            print(f"  API Calls Success: {metrics['api_calls_success']}")
            print(f"  API Calls Failed:  {metrics['api_calls_failed']}")

            if metrics['api_calls_total'] > 0:
                success_rate = (metrics['api_calls_success'] / metrics['api_calls_total']) * 100
                print(f"  Success Rate:      {success_rate:.1f}%")

            print(f"  Rate Limit Hits:   {metrics['rate_limit_violations']}")
            print()
        else:
            print("⚠️  Server metrics unavailable (is server running?)")
            print()

        # API Call Statistics
        api_stats = report['api_statistics']
        if api_stats['count'] > 0:
            print(f"🔥 API CALL PERFORMANCE (last {api_stats['sample_size']} calls)")
            print("-" * 70)
            print(f"  Calls Analyzed:    {api_stats['count']}")
            print(f"  Average Latency:   {api_stats['avg_duration_ms'] / 1000:.1f}s")
            print(f"  Min Latency:       {api_stats['min_duration_ms'] / 1000:.1f}s")
            print(f"  Max Latency:       {api_stats['max_duration_ms'] / 1000:.1f}s")
            print()

            print("💾 TOKEN USAGE")
            print("-" * 70)
            print(f"  Total Input:       {api_stats['total_tokens_in']:,} tokens")
            print(f"  Total Output:      {api_stats['total_tokens_out']:,} tokens")
            print(f"  Avg Input/Call:    {api_stats['avg_tokens_in']:,} tokens")
            print(f"  Avg Output/Call:   {api_stats['avg_tokens_out']:,} tokens")

            if api_stats['avg_tokens_out'] > 0:
                ratio = api_stats['avg_tokens_in'] / api_stats['avg_tokens_out']
                print(f"  Input/Output Ratio: {ratio:.1f}:1")

            print(f"  Cache Efficiency:  {api_stats['cache_efficiency']}")
            print()
        else:
            print("⚠️  No recent API calls found in logs")
            print()

        # Recent Tasks
        recent_tasks = report['recent_tasks']
        if recent_tasks:
            print(f"📋 RECENT TASKS (last {len(recent_tasks)})")
            print("-" * 70)
            for task in recent_tasks:
                duration_s = task['duration_ms'] / 1000
                status_icon = "✅" if task['status'] == 'completed' else "❌"
                print(f"  {status_icon} {task['role']:10} {task['task_id']:30} {duration_s:6.1f}s")
            print()

        print("=" * 70)


def main():
    parser = argparse.ArgumentParser(
        description='Monitor A2A Agent Server performance',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__
    )
    parser.add_argument(
        '--server-url',
        default='http://localhost:8080',
        help='Server URL (default: http://localhost:8080)'
    )
    parser.add_argument(
        '--log-file',
        default='/tmp/agent-server.log',
        help='Log file path (default: /tmp/agent-server.log)'
    )
    parser.add_argument(
        '--tail',
        type=int,
        default=30,
        help='Show last N API calls (default: 30)'
    )
    parser.add_argument(
        '--json',
        action='store_true',
        help='Output as JSON'
    )
    parser.add_argument(
        '--watch',
        action='store_true',
        help='Continuous monitoring mode (updates every 10s)'
    )

    args = parser.parse_args()

    monitor = PerformanceMonitor(args.server_url, args.log_file)

    try:
        if args.watch:
            print("🔄 Continuous monitoring mode (Ctrl+C to exit)")
            print()
            while True:
                if not args.json:
                    # Clear screen (ANSI escape code)
                    print("\033[2J\033[H", end='')

                report = monitor.generate_report(args.tail, args.json)

                if args.json:
                    print(json.dumps(report, indent=2))
                else:
                    monitor.print_report(report)

                time.sleep(10)
        else:
            report = monitor.generate_report(args.tail, args.json)

            if args.json:
                print(json.dumps(report, indent=2))
            else:
                monitor.print_report(report)

    except KeyboardInterrupt:
        print("\n👋 Monitoring stopped")
        sys.exit(0)
    except Exception as e:
        print(f"❌ Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()
