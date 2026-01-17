#!/usr/bin/env python3
"""
AI-Pack Codex Task Queue

Uses Beads (bd) when available, with a JSONL fallback.
"""

from __future__ import annotations

import json
import shutil
import subprocess
import sys
from pathlib import Path


def find_repo_root() -> Path:
    cwd = Path.cwd()
    for path in [cwd, *cwd.parents]:
        if (path / ".git").exists():
            return path
    return cwd


def run_bd(command: list[str], repo_root: Path) -> int:
    result = subprocess.run(
        command,
        cwd=repo_root,
        text=True,
        capture_output=True,
        check=False,
    )
    if result.stdout:
        print(result.stdout.rstrip())
    if result.stderr:
        print(result.stderr.rstrip())
    return result.returncode


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


def main() -> int:
    repo_root = find_repo_root()
    beads_file = repo_root / ".beads" / "issues.jsonl"

    print("AI-Pack Task Queue")
    print("=" * 19)
    print()

    if shutil.which("bd"):
        print("Beads Ready Tasks")
        run_bd(["bd", "ready"], repo_root)
        print()

        print("Open Tasks")
        run_bd(["bd", "list", "--status", "open"], repo_root)
        print()

        print("Blocked Tasks")
        run_bd(["bd", "list", "--status", "blocked"], repo_root)
        print()

        return 0

    if not beads_file.exists():
        print("No Beads database found.")
        print("Run 'bd init' to initialize task tracking.")
        return 0

    tasks = read_tasks(beads_file)
    open_tasks = [t for t in tasks if t.get("status") == "open"]
    blocked_tasks = [t for t in tasks if t.get("status") == "blocked"]

    print("Open Tasks (JSONL fallback)")
    for task in open_tasks:
        task_id = task.get("id", "unknown")
        title = task.get("title", "unknown")
        assignee = task.get("assignee", "unassigned")
        priority = task.get("priority", "unknown")
        print(f"- {task_id}: {title} [{priority}] ({assignee})")
    if not open_tasks:
        print("- none")
    print()

    print("Blocked Tasks (JSONL fallback)")
    for task in blocked_tasks:
        task_id = task.get("id", "unknown")
        title = task.get("title", "unknown")
        reason = task.get("blocked_reason", "unknown")
        print(f"- {task_id}: {title} (reason: {reason})")
    if not blocked_tasks:
        print("- none")
    print()

    print("Tip: Install Beads (bd) for dependency-aware queue ordering.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
