#!/usr/bin/env python3
"""
AI-Pack Task Status Script

Displays current status of active task packets in .ai/tasks/
"""

import os
import sys
from pathlib import Path
from datetime import datetime


def get_task_progress(task_dir):
    """Analyze task packet progress (2-file format)."""
    files = {
        "task": task_dir / "task.md",
        "result": task_dir / "result.md",
    }

    progress = {}
    for key, filepath in files.items():
        exists = filepath.exists()
        progress[key] = {"exists": exists, "path": filepath}

    return progress


def determine_phase(progress):
    """Determine task phase from task.md / result.md existence."""
    if progress["result"]["exists"]:
        # Try to extract Status field from result.md
        try:
            content = progress["result"]["path"].read_text()
            for line in content.splitlines():
                if line.strip().startswith("**Status:**"):
                    status_value = line.split("**Status:**", 1)[1].strip()
                    return "Complete", status_value
        except Exception:
            pass
        return "Complete", "result.md present"

    if progress["task"]["exists"]:
        return "In Progress", "task.md present, result.md not yet written"

    return "Not Started", "task.md missing"


def show_task_status():
    """Display status of all task packets."""
    tasks_dir = Path(".ai/tasks")

    # Check if .ai/tasks exists
    if not tasks_dir.exists():
        print()
        print("⚠️  No Active Task Packets")
        print()
        print("Before starting work, create a task packet:")
        print("  /ai-pack task-init <task-name>")
        print()
        print("This is MANDATORY for all non-trivial tasks.")
        print()
        return 0

    # Get all task directories
    task_dirs = [d for d in tasks_dir.iterdir() if d.is_dir() and not d.name.startswith('.')]

    if not task_dirs:
        print()
        print("⚠️  No Active Task Packets")
        print()
        print("Before starting work, create a task packet:")
        print("  /ai-pack task-init <task-name>")
        print()
        return 0

    # Display header
    print()
    print("📋 Active Task Packets")
    print("━" * 80)
    print()

    # Display each task
    for task_dir in sorted(task_dirs, reverse=True):
        # Parse display name from directory name
        task_name = task_dir.name
        if "_" in task_name:
            date_str, name = task_name.split("_", 1)
            display_name = name.replace("-", " ").title()
        else:
            display_name = task_name

        print(f"Task: {display_name}")
        print(f"Path: {task_dir}")

        # Get progress and phase
        progress = get_task_progress(task_dir)
        phase, phase_desc = determine_phase(progress)

        print(f"Status: {phase} — {phase_desc}")
        print()
        print("Files:")
        task_icon = "✅" if progress["task"]["exists"] else "⏸️"
        result_icon = "✅" if progress["result"]["exists"] else "⏸️"
        print(f"  {task_icon} task.md    (requirements, acceptance criteria)")
        print(f"  {result_icon} result.md  (written by agent on completion)")
        print()

        # Next steps
        print("Next Steps:")
        if phase == "Complete":
            print("  ✅ Task complete! Consider archiving to .ai/archive/")
        elif phase == "In Progress":
            print("  - Agent is working; result.md will appear when done")
            print("  - Monitor via: agent list --running")
        else:
            print(f"  - Edit {task_dir}/task.md")
            print("  - Fill in: description, files to change, acceptance criteria, constraints")
            print("  - Then choose role: /ai-pack engineer or /ai-pack orchestrate")

        print()
        print("─" * 80)
        print()

    return 0


if __name__ == "__main__":
    sys.exit(show_task_status())
