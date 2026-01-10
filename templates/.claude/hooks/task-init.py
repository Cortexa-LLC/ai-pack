#!/usr/bin/env python3
"""
AI-Pack Task Initialization Script

Creates a new task packet with templates from .ai-pack/templates/task-packet/
"""

import os
import sys
import shutil
from datetime import datetime
from pathlib import Path


def slugify(text):
    """Convert text to URL-friendly slug."""
    import re
    text = text.lower()
    text = re.sub(r'[^a-z0-9]+', '-', text)
    text = text.strip('-')
    return text


def create_task_packet(task_name):
    """Create task packet directory and copy templates."""

    # Validate task name
    if not task_name:
        print("❌ Error: Task name is required")
        print("Usage: /ai-pack task-init <task-name>")
        return 1

    # Create slug from task name
    slug = slugify(task_name)
    if not slug:
        print(f"❌ Error: Invalid task name '{task_name}'")
        return 1

    # Create task ID with timestamp
    timestamp = datetime.now().strftime("%Y-%m-%d")
    task_id = f"{timestamp}_{slug}"

    # Define paths
    tasks_dir = Path(".ai/tasks")
    task_dir = tasks_dir / task_id
    template_dir = Path(".ai-pack/templates/task-packet")

    # Check if .ai-pack exists
    if not template_dir.exists():
        print("❌ Error: .ai-pack/templates/task-packet/ not found")
        print("Is this repository configured with ai-pack framework?")
        return 1

    # Check if task already exists
    if task_dir.exists():
        print(f"❌ Error: Task packet already exists: {task_dir}")
        return 1

    # Create .ai/tasks directory if needed
    tasks_dir.mkdir(parents=True, exist_ok=True)

    # Create task directory
    task_dir.mkdir(parents=True, exist_ok=True)

    # Copy templates
    try:
        templates = [
            "00-contract.md",
            "10-plan.md",
            "20-work-log.md",
            "30-review.md",
            "40-acceptance.md"
        ]

        for template in templates:
            src = template_dir / template
            dst = task_dir / template

            if src.exists():
                shutil.copy2(src, dst)
                print(f"✅ Created: {dst}")
            else:
                print(f"⚠️  Template not found: {src}")

    except Exception as e:
        print(f"❌ Error copying templates: {e}")
        return 1

    # Success message
    print()
    print(f"✅ Task packet created: {task_dir}")
    print()
    print("Next steps:")
    print(f"  1. Edit {task_dir}/00-contract.md")
    print("     - Define requirements and acceptance criteria")
    print()
    print(f"  2. Edit {task_dir}/10-plan.md")
    print("     - Document implementation approach")
    print()
    print("  3. Choose your role:")
    print("     - Simple task: /ai-pack engineer")
    print("     - Complex task: /ai-pack orchestrate")
    print("     - Bug investigation: /ai-pack inspect")
    print()

    return 0


if __name__ == "__main__":
    # Get task name from command line
    if len(sys.argv) < 2:
        print("❌ Error: Task name is required")
        print("Usage: /ai-pack task-init <task-name>")
        sys.exit(1)

    task_name = " ".join(sys.argv[1:])
    sys.exit(create_task_packet(task_name))
