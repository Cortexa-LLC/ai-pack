#!/usr/bin/env python3
"""
AI-Pack Task Status Script

Displays current status of active task packets in .ai/tasks/
"""

import os
import sys
from pathlib import Path
from datetime import datetime


def get_file_status(filepath):
    """Check if file exists and has content beyond template."""
    if not filepath.exists():
        return "⏸️", "Not started"

    # Check if file has meaningful content (not just template)
    try:
        content = filepath.read_text()
        # Simple heuristic: >200 chars and not mostly template markers
        if len(content) > 200 and content.count("[") < 5:
            return "✅", "Completed"
        else:
            return "🔄", "In progress"
    except:
        return "⏸️", "Not started"


def get_task_progress(task_dir):
    """Analyze task packet progress."""
    files = {
        "contract": task_dir / "00-contract.md",
        "plan": task_dir / "10-plan.md",
        "work_log": task_dir / "20-work-log.md",
        "review": task_dir / "30-review.md",
        "acceptance": task_dir / "40-acceptance.md"
    }

    progress = {}
    for key, filepath in files.items():
        icon, status = get_file_status(filepath)
        progress[key] = {"icon": icon, "status": status, "path": filepath}

    return progress


def determine_current_phase(progress):
    """Determine which phase the task is in."""
    if progress["acceptance"]["icon"] == "✅":
        return "Complete", "Task is complete!"

    if progress["review"]["icon"] in ["🔄", "✅"]:
        return "Review", "Code is under review"

    if progress["work_log"]["icon"] in ["🔄", "✅"]:
        return "Implementation", "Work in progress"

    if progress["plan"]["icon"] == "✅":
        return "Ready to Start", "Plan complete, ready for implementation"

    if progress["contract"]["icon"] == "✅":
        return "Planning", "Contract defined, plan needs completion"

    return "Not Started", "Contract needs to be filled out"


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
        # Parse task name
        task_name = task_dir.name
        if "_" in task_name:
            date_str, name = task_name.split("_", 1)
            display_name = name.replace("-", " ").title()
        else:
            display_name = task_name

        print(f"Task: {display_name}")
        print(f"Path: {task_dir}")

        # Get progress
        progress = get_task_progress(task_dir)
        phase, phase_desc = determine_current_phase(progress)

        print(f"Status: {phase}")
        print()
        print("Progress:")
        print(f"  {progress['contract']['icon']} 00-contract.md   (Requirements)")
        print(f"  {progress['plan']['icon']} 10-plan.md       (Implementation plan)")
        print(f"  {progress['work_log']['icon']} 20-work-log.md   (Execution progress)")
        print(f"  {progress['review']['icon']} 30-review.md     (Quality assurance)")
        print(f"  {progress['acceptance']['icon']} 40-acceptance.md (Sign-off)")
        print()

        # Determine next steps
        print("Next Steps:")
        if phase == "Complete":
            print("  ✅ Task complete! Consider archiving to .ai/archive/")
        elif phase == "Review":
            print("  - Address review feedback if needed")
            print("  - Run /ai-pack review to re-validate")
        elif phase == "Implementation":
            print("  - Continue implementation")
            print("  - Update work log regularly")
            print("  - Run /ai-pack test when ready for validation")
        elif phase == "Ready to Start":
            print("  - Choose role: /ai-pack engineer or /ai-pack orchestrate")
            print("  - Begin implementation following the plan")
        elif phase == "Planning":
            print(f"  - Edit {task_dir}/10-plan.md")
            print("  - Document implementation approach")
            print("  - Then choose role to begin work")
        else:
            print(f"  - Edit {task_dir}/00-contract.md")
            print("  - Define requirements and acceptance criteria")
            print(f"  - Then edit {task_dir}/10-plan.md")

        print()
        print("─" * 80)
        print()

    return 0


if __name__ == "__main__":
    sys.exit(show_task_status())
