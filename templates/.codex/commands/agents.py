#!/usr/bin/env python3
"""
AI-Pack Codex Agent Status

Reads .beads/issues.jsonl to show agent tasks without Claude tooling.
"""

from __future__ import annotations

import json
import sys
from pathlib import Path


MAX_BACKGROUND_AGENTS = 3


def find_repo_root() -> Path:
    cwd = Path.cwd()
    for path in [cwd, *cwd.parents]:
        if (path / ".git").exists():
            return path
    return cwd


def read_tasks(path: Path) -> list[dict]:
    tasks = []
    with open(path, "r", encoding="utf-8") as handle:
        for line in handle:
            line = line.strip()
            if not line:
                continue
            try:
                tasks.append(json.loads(line))
            except json.JSONDecodeError:
                continue
    return tasks


def is_agent_task(task: dict) -> bool:
    title = str(task.get("title", ""))
    return title.startswith("Agent:")


def format_task(task: dict) -> str:
    task_id = task.get("id", "unknown")
    assignee = task.get("assignee", "unknown")
    title = task.get("title", "unknown")
    status = task.get("status", "unknown")
    created_at = task.get("created_at", "unknown")
    return (
        f"{task_id}:\n"
        f"  Assignee: {assignee}\n"
        f"  Task:     {title}\n"
        f"  Status:   {status}\n"
        f"  Started:  {created_at}\n"
    )


def main() -> int:
    repo_root = find_repo_root()
    beads_file = repo_root / ".beads" / "issues.jsonl"

    if not beads_file.exists():
        print("No Beads database found.")
        print("Run 'bd init' if you want agent tracking.")
        return 0

    tasks = [task for task in read_tasks(beads_file) if is_agent_task(task)]

    active = [t for t in tasks if t.get("status") == "in_progress"]
    blocked = [t for t in tasks if t.get("status") == "blocked"]
    completed = [t for t in tasks if t.get("status") == "closed"]
    open_tasks = [t for t in tasks if t.get("status") == "open"]

    print("AI-Pack Agent Status (via .beads/issues.jsonl)")
    print("=" * 46)
    print()
    print(f"Active Agents: {len(active)} / {MAX_BACKGROUND_AGENTS} maximum")
    print()

    if active:
        for task in active:
            print(format_task(task))
    else:
        print("No active agents.")
        print()

    if completed:
        print(f"Completed: {len(completed)}")
        for task in completed:
            task_id = task.get("id", "unknown")
            title = task.get("title", "unknown")
            assignee = task.get("assignee", "unknown")
            print(f"  - {task_id}: {title} ({assignee})")
        print()

    if blocked:
        print(f"Blocked: {len(blocked)}")
        for task in blocked:
            task_id = task.get("id", "unknown")
            title = task.get("title", "unknown")
            reason = task.get("blocked_reason", "unknown")
            print(f"  - {task_id}: {title} - Reason: {reason}")
        print()

    if open_tasks:
        print(f"Open: {len(open_tasks)}")
        print()

    available = max(MAX_BACKGROUND_AGENTS - len(active), 0)
    print(f"Available capacity: {available} slots")
    print()
    print("Shared Context Reminder:")
    print("- All agents share the same repository")
    print("- Coordinate builds and test runs")
    print("- No per-agent git branches")

    return 0


if __name__ == "__main__":
    sys.exit(main())
