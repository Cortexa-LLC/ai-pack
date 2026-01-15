#!/usr/bin/env python3
"""
Migrate Agent Status from JSON to Beads

Migrates existing .claude/.agent-status.json (legacy) to Beads tasks.

Usage:
    python3 scripts/migrate-agent-status-to-beads.py [--dry-run]

Options:
    --dry-run    Show what would be done without making changes
"""

import json
import subprocess
import sys
from pathlib import Path
from datetime import datetime


def run_bd_command(args, dry_run=False):
    """Run bd command and return output."""
    if dry_run:
        print(f"[DRY RUN] Would run: bd {' '.join(args)}")
        return 0, ""

    try:
        result = subprocess.run(
            ["bd"] + args,
            capture_output=True,
            text=True,
            check=True
        )
        return result.returncode, result.stdout.strip()
    except subprocess.CalledProcessError as e:
        print(f"Error running bd command: {e}")
        return e.returncode, e.stderr
    except FileNotFoundError:
        print("Error: 'bd' command not found. Install Beads first:")
        print("  https://github.com/steveyegge/beads")
        sys.exit(1)


def migrate_agents(dry_run=False):
    """Migrate agent status from JSON to Beads."""
    status_file = Path(".claude/.agent-status.json")

    if not status_file.exists():
        print("No .claude/.agent-status.json found - nothing to migrate")
        return

    print(f"Reading agent status from {status_file}")

    try:
        with open(status_file) as f:
            data = json.load(f)
    except json.JSONDecodeError:
        print(f"Error: Invalid JSON in {status_file}")
        sys.exit(1)

    agents = data.get("agents", {})

    if not agents:
        print("No agents found in status file")
        return

    print(f"\nFound {len(agents)} agents to migrate\n")

    migrated_count = 0

    for agent_id, agent in agents.items():
        role = agent.get("role", "Worker")
        task = agent.get("task", "Unknown task")
        status = agent.get("status", "active")

        # Create Beads task title
        title = f"Agent: {role} - {task}"
        assignee = f"{role}-{agent_id}"

        print(f"Migrating: {agent_id}")
        print(f"  Title: {title}")
        print(f"  Assignee: {assignee}")
        print(f"  Status: {status}")

        # Create Beads task
        returncode, task_id = run_bd_command([
            "create",
            title,
            "--assignee", assignee,
            "--priority", "normal"
        ], dry_run)

        if returncode != 0:
            print(f"  ❌ Failed to create Beads task")
            continue

        if not dry_run:
            # Extract task ID from output (last word in "Created task: bd-a1b2")
            if task_id:
                parts = task_id.split()
                if parts:
                    bd_task_id = parts[-1]
                    print(f"  ✓ Created Beads task: {bd_task_id}")

                    # Update status based on original status
                    if status == "active" or status == "in_progress":
                        run_bd_command(["start", bd_task_id], dry_run)
                        print(f"  ✓ Marked as in_progress")
                    elif status == "completed" or status == "closed":
                        run_bd_command(["start", bd_task_id], dry_run)
                        run_bd_command(["close", bd_task_id], dry_run)
                        print(f"  ✓ Marked as closed")
                    elif status == "blocked":
                        run_bd_command(["start", bd_task_id], dry_run)
                        blocker_reason = "Migrated from legacy system"
                        if agent.get("blockers"):
                            blocker_reason = agent["blockers"][0].get("blocker", blocker_reason)
                        run_bd_command(["block", bd_task_id, blocker_reason], dry_run)
                        print(f"  ✓ Marked as blocked: {blocker_reason}")

                    migrated_count += 1
        else:
            print(f"  [DRY RUN] Would create and update task")
            migrated_count += 1

        print()

    print(f"\nMigration complete: {migrated_count}/{len(agents)} agents migrated")

    if not dry_run:
        backup_file = status_file.with_suffix(".json.backup")
        print(f"\nTo backup old file, run:")
        print(f"  mv {status_file} {backup_file}")


def main():
    """Main entry point."""
    dry_run = "--dry-run" in sys.argv

    print("=" * 70)
    print("AI-Pack Agent Status Migration: JSON → Beads")
    print("=" * 70)
    print()

    if dry_run:
        print("Running in DRY RUN mode - no changes will be made\n")

    # Check Beads is installed
    try:
        subprocess.run(["bd", "--version"], capture_output=True, check=True)
    except (subprocess.CalledProcessError, FileNotFoundError):
        print("Error: Beads not installed")
        print("Install from: https://github.com/steveyegge/beads")
        sys.exit(1)

    # Check Beads is initialized
    if not Path(".beads/issues.jsonl").exists():
        print("Error: Beads not initialized in this project")
        print("Run: bd init")
        sys.exit(1)

    migrate_agents(dry_run)

    if not dry_run:
        print("\nNext steps:")
        print("  1. Verify migrated agents: bd list --json | jq '.[] | select(.title | startswith(\"Agent:\"))'")
        print("  2. Check with: /ai-pack agents")
        print("  3. Backup old file: mv .claude/.agent-status.json .claude/.agent-status.json.backup")


if __name__ == "__main__":
    main()
