#!/usr/bin/env python3
"""
AI-Pack Task Packet Gate Enforcement

Checks if task packet exists before allowing implementation work.
This script is called by Claude Code hooks to enforce the Task Packet gate.

Exit codes:
  0 - Allow (gate passed)
  1 - Error (technical failure)
  2 - Block (gate violation)
"""

import os
import sys
import json
from pathlib import Path


def check_task_packet(user_input):
    """Check if task packet exists when user requests implementation work."""

    # Keywords that indicate implementation work
    implementation_keywords = [
        "implement",
        "code",
        "write",
        "build",
        "create",
        "develop",
        "engineer",
        "orchestrate"
    ]

    # Check if user input contains implementation keywords
    user_input_lower = user_input.lower()
    is_implementation_request = any(
        keyword in user_input_lower
        for keyword in implementation_keywords
    )

    # If not an implementation request, allow
    if not is_implementation_request:
        return 0

    # Check if .ai/tasks directory exists and has tasks
    tasks_dir = Path(".ai/tasks")

    if not tasks_dir.exists():
        print("⚠️  GATE VIOLATION: No Task Packet")
        print()
        print("Before implementation, create a task packet:")
        print("  /ai-pack task-init <task-name>")
        print()
        print("This is MANDATORY for all non-trivial tasks.")
        print()
        print("See: .ai-pack/gates/00-global-gates.md")
        return 2

    # Get task directories
    task_dirs = [d for d in tasks_dir.iterdir() if d.is_dir() and not d.name.startswith('.')]

    if not task_dirs:
        print("⚠️  GATE VIOLATION: No Active Task Packet")
        print()
        print("Before implementation, create a task packet:")
        print("  /ai-pack task-init <task-name>")
        print()
        print("This is MANDATORY for all non-trivial tasks.")
        print()
        print("See: .ai-pack/gates/00-global-gates.md")
        return 2

    # Gate passed
    return 0


if __name__ == "__main__":
    # Read user input from stdin (provided by Claude Code hook)
    try:
        data = json.load(sys.stdin)
        user_input = data.get("user_input", "")
    except:
        # If no JSON input, check command line args
        user_input = " ".join(sys.argv[1:]) if len(sys.argv) > 1 else ""

    sys.exit(check_task_packet(user_input))
