#!/usr/bin/env python3
"""
AI-Pack Codex Health Check

Reports basic repository health for Codex workflows without Claude tooling.
"""

from __future__ import annotations

import json
import subprocess
import sys
import time
from pathlib import Path


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


def format_status(ok: bool) -> str:
    return "OK" if ok else "WARN"


def check_git_tracking(repo_root: Path, rel_path: str) -> str:
    try:
        result = subprocess.run(
            ["git", "ls-files", rel_path],
            cwd=repo_root,
            capture_output=True,
            text=True,
            check=False,
        )
        return "OK" if result.stdout.strip() else "WARN"
    except Exception:
        return "WARN"


def load_jsonl_count(path: Path) -> int:
    try:
        with open(path, "r", encoding="utf-8") as handle:
            return sum(1 for line in handle if line.strip())
    except Exception:
        return 0


def main() -> int:
    repo_root = find_repo_root()
    print("AI-Pack Codex Health Check")
    print("=" * 32)
    print(f"Repository: {repo_root}")
    print()

    checks = []

    ai_pack = repo_root / ".ai-pack"
    checks.append(("ai-pack submodule", ai_pack.exists()))

    agents_file = repo_root / "AGENTS.md"
    checks.append(("AGENTS.md present", agents_file.exists()))

    codex_dir = repo_root / ".codex"
    checks.append((".codex directory", codex_dir.exists()))

    tasks_dir = repo_root / ".ai" / "tasks"
    checks.append(("task packets dir", tasks_dir.exists()))

    beads_file = repo_root / ".beads" / "issues.jsonl"
    checks.append(("Beads issues.jsonl", beads_file.exists()))

    print("Core Checks")
    for label, ok in checks:
        print(f"- {label:22} {format_status(ok)}")
    print()

    if tasks_dir.exists():
        task_dirs = sorted([p for p in tasks_dir.iterdir() if p.is_dir()])
        missing_packets = 0
        missing_files = 0
        recent_logs = 0
        now = time.time()

        for task_dir in task_dirs:
            missing = [name for name in REQUIRED_TASK_FILES if not (task_dir / name).exists()]
            if missing:
                missing_packets += 1
                missing_files += len(missing)

            work_log = task_dir / "20-work-log.md"
            if work_log.exists():
                age_minutes = (now - work_log.stat().st_mtime) / 60
                if age_minutes <= 30:
                    recent_logs += 1

        print("Task Packet Health")
        print(f"- task packets         {len(task_dirs)}")
        print(f"- packets incomplete   {missing_packets}")
        print(f"- missing files        {missing_files}")
        print(f"- recent work logs     {recent_logs}")
        print()
    else:
        print("Task Packet Health")
        print("- task packets         0")
        print()

    if beads_file.exists():
        task_count = load_jsonl_count(beads_file)
        tracked = check_git_tracking(repo_root, ".beads/issues.jsonl")
        print("Beads Health")
        print(f"- issues.jsonl entries {task_count}")
        print(f"- git tracked          {tracked}")
        print()

    print("Next Steps")
    if not ai_pack.exists():
        print("- Add ai-pack submodule: git submodule add <url> .ai-pack")
    if not agents_file.exists():
        print("- Install Codex assets: python3 .ai-pack/templates/.codex-setup.py")
    if not tasks_dir.exists():
        print("- Create task packets: mkdir -p .ai/tasks")
    if beads_file.exists() is False:
        print("- Initialize Beads: bd init (if using task tracking)")

    return 0


if __name__ == "__main__":
    sys.exit(main())
