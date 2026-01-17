#!/usr/bin/env python3
"""
AI-Pack Codex Orchestration Snapshot

Summarizes task packets and agent activity for Orchestrator workflows.
"""

from __future__ import annotations

import json
import sys
import time
from pathlib import Path


MAX_BACKGROUND_AGENTS = 3
REQUIRED_TASK_FILES = [
    "00-contract.md",
    "10-plan.md",
    "20-work-log.md",
    "30-review.md",
    "40-acceptance.md",
]


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
    return str(task.get("title", "")).startswith("Agent:")


def summarize_task_packets(tasks_dir: Path) -> dict:
    summary = {
        "count": 0,
        "incomplete": 0,
        "missing_files": 0,
        "recent_logs": 0,
        "recent_tasks": [],
    }

    if not tasks_dir.exists():
        return summary

    task_dirs = sorted([p for p in tasks_dir.iterdir() if p.is_dir()])
    summary["count"] = len(task_dirs)
    now = time.time()

    for task_dir in task_dirs:
        missing = [name for name in REQUIRED_TASK_FILES if not (task_dir / name).exists()]
        if missing:
            summary["incomplete"] += 1
            summary["missing_files"] += len(missing)

        work_log = task_dir / "20-work-log.md"
        if work_log.exists():
            age_minutes = (now - work_log.stat().st_mtime) / 60
            if age_minutes <= 60:
                summary["recent_logs"] += 1

    summary["recent_tasks"] = task_dirs[-5:]
    return summary


def main() -> int:
    repo_root = find_repo_root()
    tasks_dir = repo_root / ".ai" / "tasks"
    beads_file = repo_root / ".beads" / "issues.jsonl"

    print("AI-Pack Orchestration Snapshot")
    print("=" * 34)
    print(f"Repository: {repo_root}")
    print()

    summary = summarize_task_packets(tasks_dir)
    print("Task Packets")
    print(f"- total packets        {summary['count']}")
    print(f"- incomplete packets   {summary['incomplete']}")
    print(f"- missing files        {summary['missing_files']}")
    print(f"- recent work logs     {summary['recent_logs']}")
    print()

    if summary["recent_tasks"]:
        print("Recent Tasks")
        for task_dir in summary["recent_tasks"]:
            work_log = task_dir / "20-work-log.md"
            status = "no-log"
            if work_log.exists():
                age_minutes = int((time.time() - work_log.stat().st_mtime) / 60)
                status = f"log {age_minutes}m ago"
            print(f"- {task_dir.name} ({status})")
        print()

    if beads_file.exists():
        tasks = [task for task in read_tasks(beads_file) if is_agent_task(task)]
        active = [t for t in tasks if t.get("status") == "in_progress"]
        blocked = [t for t in tasks if t.get("status") == "blocked"]
        completed = [t for t in tasks if t.get("status") == "closed"]

        print("Agent Activity")
        print(f"- active agents        {len(active)} / {MAX_BACKGROUND_AGENTS}")
        print(f"- blocked agents       {len(blocked)}")
        print(f"- completed agents     {len(completed)}")
        print()

        if len(active) > MAX_BACKGROUND_AGENTS:
            print("Warning: Active agents exceed WIP limits.")
            print()
    else:
        print("Agent Activity")
        print("- no Beads database found")
        print()

    print("Next Steps")
    print("- Use task queue: python3 .codex/commands/task-queue.py")
    print("- Update work logs for active tasks")
    print("- Verify artifacts persisted for background agents")

    return 0


if __name__ == "__main__":
    sys.exit(main())
