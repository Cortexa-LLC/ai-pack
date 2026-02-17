#!/usr/bin/env python3
"""
Backfill historical metrics from Anthropic CSV exports.

Distributes daily token/cost aggregates across tasks proportionally by duration,
then writes per-project metrics files.
"""

import csv
import json
import os
import sys
from datetime import datetime, timedelta
from pathlib import Path
from collections import defaultdict
from typing import Dict, List, Tuple

# Project roots to scan for tasks
PROJECT_ROOTS = [
    "/Users/bryanw/Projects/Vibe/ai-pack/a2a-agent",
    "/Users/bryanw/Projects/Vintage/tools/xasm++",
]

# CSV file paths
CSV_DIR = Path.home() / "Downloads"
TOKEN_CSVS = [
    CSV_DIR / "claude_api_tokens_2026_01.csv",
    CSV_DIR / "claude_api_tokens_2026_02.csv",
]
COST_CSVS = [
    CSV_DIR / "claude_api_cost_2026_01_17_to_2026_02_01.csv",
    CSV_DIR / "claude_api_cost_2026_02_01_to_2026_02_16.csv",
]


def parse_tokens_csv(csv_path: Path) -> Dict[str, Dict]:
    """Parse tokens CSV into daily totals by model."""
    daily_data = {}

    with open(csv_path, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            date = row['usage_date_utc']
            model = row['model_version']
            key = f"{date}:{model}"

            # Sum all token types
            input_tokens = (
                int(row.get('usage_input_tokens_no_cache', 0) or 0) +
                int(row.get('usage_input_tokens_cache_write_5m', 0) or 0) +
                int(row.get('usage_input_tokens_cache_write_1h', 0) or 0) +
                int(row.get('usage_input_tokens_cache_read', 0) or 0)
            )
            output_tokens = int(row.get('usage_output_tokens', 0) or 0)

            daily_data[key] = {
                'date': date,
                'model': model,
                'input_tokens': input_tokens,
                'output_tokens': output_tokens,
            }

    return daily_data


def parse_costs_csv(csv_path: Path) -> Dict[str, float]:
    """Parse costs CSV into daily totals by model."""
    daily_costs = defaultdict(float)

    with open(csv_path, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            date = row['usage_date_utc']
            # Map display name to model version
            model_map = {
                'Claude Sonnet 4.5': 'claude-sonnet-4-5-20250929',
                'Claude Haiku 4.5': 'claude-haiku-4-5-20251001',
            }
            model = model_map.get(row['model'], row['model'])
            cost = float(row.get('cost_usd', 0) or 0)

            key = f"{date}:{model}"
            daily_costs[key] += cost

    return daily_costs


def find_all_task_metadata(project_roots: List[str]) -> List[Dict]:
    """Find all task metadata files across projects."""
    tasks = []

    for project_root in project_roots:
        tasks_dir = Path(project_root) / ".beads" / "tasks"
        if not tasks_dir.exists():
            continue

        for metadata_file in tasks_dir.rglob("00-metadata.json"):
            try:
                with open(metadata_file, 'r') as f:
                    metadata = json.load(f)

                # Extract key fields
                task = {
                    'task_id': metadata.get('task_id', ''),
                    'project_root': metadata.get('metadata', {}).get('project_root') or project_root,
                    'spawned_at': metadata.get('spawned_at'),
                    'updated_at': metadata.get('updated_at'),
                    'status': metadata.get('status', ''),
                }

                # Parse timestamps
                if task['spawned_at']:
                    task['start'] = datetime.fromisoformat(task['spawned_at'].replace('Z', '+00:00'))
                if task['updated_at']:
                    task['end'] = datetime.fromisoformat(task['updated_at'].replace('Z', '+00:00'))

                # Calculate duration in seconds
                if 'start' in task and 'end' in task:
                    task['duration_sec'] = (task['end'] - task['start']).total_seconds()
                else:
                    task['duration_sec'] = 0

                tasks.append(task)
            except Exception as e:
                print(f"Warning: Failed to parse {metadata_file}: {e}", file=sys.stderr)
                continue

    return tasks


def get_tasks_for_day(tasks: List[Dict], date_str: str) -> List[Dict]:
    """Find all tasks that ran during the given day."""
    from datetime import timezone
    date = datetime.strptime(date_str, "%Y-%m-%d")
    # Make timezone-aware (UTC)
    day_start = date.replace(hour=0, minute=0, second=0, tzinfo=timezone.utc)
    day_end = date.replace(hour=23, minute=59, second=59, tzinfo=timezone.utc)

    matching = []
    for task in tasks:
        if 'start' not in task or 'end' not in task:
            continue

        # Task overlaps with day if:
        # task_start <= day_end AND task_end >= day_start
        if task['start'] <= day_end and task['end'] >= day_start:
            matching.append(task)

    return matching


def distribute_daily_usage(daily_tokens: Dict, daily_costs: Dict, tasks: List[Dict]) -> Dict[str, Dict]:
    """Distribute daily usage across tasks proportionally by duration."""
    project_daily = defaultdict(lambda: defaultdict(lambda: {
        'input_tokens': 0,
        'output_tokens': 0,
        'cost': 0.0,
        'calls': 0,
        'models': defaultdict(lambda: {
            'input_tokens': 0,
            'output_tokens': 0,
            'cost': 0.0,
            'calls': 0,
        })
    }))

    # Process each day's data
    for key, token_data in daily_tokens.items():
        date = token_data['date']
        model = token_data['model']
        input_tokens = token_data['input_tokens']
        output_tokens = token_data['output_tokens']
        cost = daily_costs.get(key, 0.0)

        # Find tasks that ran on this day
        day_tasks = get_tasks_for_day(tasks, date)

        if not day_tasks:
            print(f"Warning: No tasks found for {date} with model {model}", file=sys.stderr)
            continue

        # Calculate total duration for proportional split
        total_duration = sum(t['duration_sec'] for t in day_tasks if t['duration_sec'] > 0)

        if total_duration == 0:
            # Equal split if no duration data
            total_duration = len(day_tasks)
            for task in day_tasks:
                task['duration_sec'] = 1

        # Distribute proportionally
        for task in day_tasks:
            if task['duration_sec'] <= 0:
                continue

            proportion = task['duration_sec'] / total_duration
            project = task['project_root']

            # Calculate this task's share
            task_input = int(input_tokens * proportion)
            task_output = int(output_tokens * proportion)
            task_cost = cost * proportion

            # Add to project's daily total
            project_daily[project][date]['input_tokens'] += task_input
            project_daily[project][date]['output_tokens'] += task_output
            project_daily[project][date]['cost'] += task_cost
            project_daily[project][date]['calls'] += max(1, int(proportion * 10))  # Estimate calls

            # Add to model breakdown
            project_daily[project][date]['models'][model]['input_tokens'] += task_input
            project_daily[project][date]['models'][model]['output_tokens'] += task_output
            project_daily[project][date]['models'][model]['cost'] += task_cost
            project_daily[project][date]['models'][model]['calls'] += max(1, int(proportion * 10))

    return project_daily


def write_project_metrics(project_daily: Dict[str, Dict]):
    """Write per-project daily metrics files."""
    for project_root, daily_data in project_daily.items():
        metrics_dir = Path(project_root) / ".claude" / "metrics" / "daily"
        metrics_dir.mkdir(parents=True, exist_ok=True)

        for date, data in daily_data.items():
            output_file = metrics_dir / f"{date}.json"

            # Build provider breakdown
            provider_breakdown = {}
            for model, model_data in data['models'].items():
                provider = "anthropic" if "claude" in model.lower() else "openai"
                key = f"{provider}:{model}"
                provider_breakdown[key] = {
                    "provider": provider,
                    "model": model,
                    "calls": model_data['calls'],
                    "input_tokens": model_data['input_tokens'],
                    "output_tokens": model_data['output_tokens'],
                    "cost": round(model_data['cost'], 6)
                }

            metrics_content = {
                "date": date,
                "total_input_tokens": data['input_tokens'],
                "total_output_tokens": data['output_tokens'],
                "provider_breakdown": provider_breakdown,
                "last_updated": datetime.now().astimezone().isoformat(),
                "backfilled": True,
                "backfill_method": "proportional_by_duration"
            }

            with open(output_file, 'w') as f:
                json.dump(metrics_content, f, indent=2)

            print(f"✅ Wrote {output_file}")
            print(f"   {data['input_tokens']:,} input + {data['output_tokens']:,} output tokens")
            print(f"   ${data['cost']:.4f}")


def main():
    print("🔄 Starting metrics backfill...\n")

    # Load CSV data
    print("📥 Loading Anthropic CSV exports...")
    daily_tokens = {}
    for csv_path in TOKEN_CSVS:
        if csv_path.exists():
            daily_tokens.update(parse_tokens_csv(csv_path))
            print(f"   ✅ Loaded {csv_path.name}")

    daily_costs = {}
    for csv_path in COST_CSVS:
        if csv_path.exists():
            daily_costs.update(parse_costs_csv(csv_path))
            print(f"   ✅ Loaded {csv_path.name}")

    print(f"\n📊 Found {len(daily_tokens)} day+model combinations")

    # Find all tasks
    print(f"\n🔍 Scanning for task metadata in {len(PROJECT_ROOTS)} projects...")
    tasks = find_all_task_metadata(PROJECT_ROOTS)
    print(f"   ✅ Found {len(tasks)} tasks")

    # Distribute usage
    print("\n⚙️  Distributing usage across tasks by duration...")
    project_daily = distribute_daily_usage(daily_tokens, daily_costs, tasks)

    # Write metrics
    print(f"\n💾 Writing per-project metrics files...")
    write_project_metrics(project_daily)

    print(f"\n✨ Backfill complete!")
    print(f"   Projects updated: {len(project_daily)}")
    total_days = sum(len(dates) for dates in project_daily.values())
    print(f"   Daily files created: {total_days}")


if __name__ == "__main__":
    main()
