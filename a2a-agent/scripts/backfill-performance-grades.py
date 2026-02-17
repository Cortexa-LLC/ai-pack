#!/usr/bin/env python3
"""
Backfill performance grades from historical task metadata.

Analyzes completed tasks to generate performance grades per model/role/project.
"""

import json
import os
import sys
from datetime import datetime
from pathlib import Path
from collections import defaultdict
from typing import Dict, List

# Project roots to scan for tasks
PROJECT_ROOTS = [
    "/Users/bryanw/Projects/Vibe/ai-pack/a2a-agent",
    "/Users/bryanw/Projects/Vintage/tools/xasm++",
]


def sanitize_filename(s: str) -> str:
    """Sanitize string for use in filename."""
    replacements = {
        '/': '_', '\\': '_', ':': '_', '*': '_',
        '?': '_', '"': '_', '<': '_', '>': '_', '|': '_'
    }
    for old, new in replacements.items():
        s = s.replace(old, new)
    return s


def calculate_grade(success_rate: float, retry_rate: float) -> str:
    """Calculate letter grade based on success and retry rates."""
    # Grade A: 95%+ success, <5% retry
    if success_rate >= 0.95 and retry_rate < 0.05:
        return "A"

    # Grade B: 85%+ success, <15% retry
    if success_rate >= 0.85 and retry_rate < 0.15:
        return "B"

    # Grade C: 70%+ success, <30% retry
    if success_rate >= 0.70 and retry_rate < 0.30:
        return "C"

    # Grade D: 50%+ success, <50% retry
    if success_rate >= 0.50 and retry_rate < 0.50:
        return "D"

    # Grade F: Everything else
    return "F"


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
                    'role': metadata.get('role', 'unknown'),
                    'status': metadata.get('status', ''),
                    'spawned_at': metadata.get('spawned_at'),
                    'updated_at': metadata.get('updated_at'),
                    'model': metadata.get('metadata', {}).get('model', 'claude-sonnet-4-5-20250929'),
                }

                # Parse timestamps for duration calculation
                if task['spawned_at'] and task['updated_at']:
                    try:
                        start = datetime.fromisoformat(task['spawned_at'].replace('Z', '+00:00'))
                        end = datetime.fromisoformat(task['updated_at'].replace('Z', '+00:00'))
                        task['duration_ms'] = int((end - start).total_seconds() * 1000)
                        task['first_used'] = start
                        task['last_used'] = end
                    except Exception:
                        task['duration_ms'] = 0

                # Determine success/failure
                task['success'] = task['status'] == 'completed'
                task['failure'] = task['status'] == 'failed'

                tasks.append(task)
            except Exception as e:
                print(f"Warning: Failed to parse {metadata_file}: {e}", file=sys.stderr)
                continue

    return tasks


def aggregate_grades(tasks: List[Dict]) -> Dict[str, Dict]:
    """Aggregate tasks into performance grades by model:role:project."""
    grades = defaultdict(lambda: {
        'model_id': '',
        'role_id': '',
        'project_id': '',
        'total_attempts': 0,
        'successes': 0,
        'failures': 0,
        'retries': 0,
        'total_tokens_used': 0,
        'total_execution_time_ms': 0,
        'escalation_count': 0,
        'downgrade_count': 0,
        'first_used': None,
        'last_used': None,
        'last_task_id': '',
    })

    for task in tasks:
        # Skip tasks without required fields
        if not task.get('role') or not task.get('project_root'):
            continue

        model = task.get('model', 'claude-sonnet-4-5-20250929')
        role = task['role']
        project = task['project_root']

        key = f"{model}:{role}:{project}"
        grade = grades[key]

        # Initialize if first time
        if not grade['model_id']:
            grade['model_id'] = model
            grade['role_id'] = role
            grade['project_id'] = project

        # Update counts
        grade['total_attempts'] += 1
        if task.get('success'):
            grade['successes'] += 1
        if task.get('failure'):
            grade['failures'] += 1

        # Update time tracking
        if task.get('duration_ms'):
            grade['total_execution_time_ms'] += task['duration_ms']

        if task.get('first_used'):
            if grade['first_used'] is None or task['first_used'] < grade['first_used']:
                grade['first_used'] = task['first_used']

        if task.get('last_used'):
            if grade['last_used'] is None or task['last_used'] > grade['last_used']:
                grade['last_used'] = task['last_used']
                grade['last_task_id'] = task['task_id']

        # Note: Token data not available in backfilled historical data
        # Will be tracked for new tasks going forward

    return grades


def calculate_derived_metrics(grades: Dict[str, Dict]) -> None:
    """Calculate derived metrics for each grade."""
    for grade in grades.values():
        if grade['total_attempts'] == 0:
            continue

        # Calculate rates
        grade['success_rate'] = grade['successes'] / grade['total_attempts']
        grade['error_rate'] = grade['failures'] / grade['total_attempts']
        grade['retry_rate'] = grade['retries'] / grade['total_attempts']

        # Calculate averages (only for data we have)
        grade['average_execution_time'] = (grade['total_execution_time_ms'] / grade['total_attempts']) / 1000.0
        # Note: average_tokens not included - no historical token data available

        # Calculate confidence score (0.0 to 1.0, full confidence at 20+ samples)
        grade['confidence_score'] = min(1.0, grade['total_attempts'] / 20.0)

        # Calculate letter grade
        grade['grade'] = calculate_grade(grade['success_rate'], grade['retry_rate'])

        # Convert timestamps to ISO format
        if grade['first_used']:
            grade['first_used'] = grade['first_used'].isoformat()
        if grade['last_used']:
            grade['last_used'] = grade['last_used'].isoformat()


def write_grade_files(grades: Dict[str, Dict]) -> None:
    """Write performance grade files to each project."""
    # Group by project
    by_project = defaultdict(list)
    for key, grade in grades.items():
        by_project[grade['project_id']].append(grade)

    for project_root, project_grades in by_project.items():
        grades_dir = Path(project_root) / ".claude" / "performance_grades"
        grades_dir.mkdir(parents=True, exist_ok=True)

        for grade in project_grades:
            filename = f"{sanitize_filename(grade['model_id'])}_{sanitize_filename(grade['role_id'])}_{sanitize_filename(grade['project_id'])}.json"
            output_file = grades_dir / filename

            with open(output_file, 'w') as f:
                json.dump(grade, f, indent=2)

            print(f"✅ Wrote {output_file}")
            print(f"   {grade['role_id']}@{grade['model_id']}: Grade {grade['grade']} ({grade['total_attempts']} attempts, {grade['success_rate']:.1%} success)")


def main():
    print("🔄 Starting performance grades backfill...\n")

    # Find all tasks
    print(f"🔍 Scanning for task metadata in {len(PROJECT_ROOTS)} projects...")
    tasks = find_all_task_metadata(PROJECT_ROOTS)
    print(f"   ✅ Found {len(tasks)} tasks")

    # Aggregate into grades
    print("\n⚙️  Aggregating tasks into performance grades...")
    grades = aggregate_grades(tasks)
    print(f"   ✅ Created {len(grades)} unique grade combinations")

    # Calculate derived metrics
    print("\n📊 Calculating grades and metrics...")
    calculate_derived_metrics(grades)

    # Write grade files
    print(f"\n💾 Writing performance grade files...")
    write_grade_files(grades)

    print(f"\n✨ Backfill complete!")
    print(f"   Grade combinations created: {len(grades)}")

    # Show grade distribution
    grade_dist = defaultdict(int)
    for grade in grades.values():
        grade_dist[grade['grade']] += 1

    print(f"\n   Grade distribution:")
    for letter in ['A', 'B', 'C', 'D', 'F']:
        if letter in grade_dist:
            print(f"     {letter}: {grade_dist[letter]}")


if __name__ == "__main__":
    main()
