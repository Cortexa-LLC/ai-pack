#!/usr/bin/env python3
"""
Coordinator Check-in Script

Called by coordination timer every 30 seconds.
Reads agent work logs and outputs progress updates.
Cross-platform Python implementation.
"""

import os
import glob
from datetime import datetime
from pathlib import Path

def check_agent_progress():
    """Check all agent work logs and report status."""

    # Find all work logs
    work_logs = glob.glob('.ai/tasks/*/20-work-log.md')

    if not work_logs:
        # No active work packages
        return

    print(f"\n⏱️ Agent Progress Update ({datetime.now().strftime('%H:%M:%S')})\n")

    agents_found = False
    for log_path in work_logs:
        # Read log file
        try:
            with open(log_path, 'r', encoding='utf-8') as f:
                content = f.read()

            # Check if this work log has agent activity
            # Look for role markers: Engineer Role, Reviewer Role, Tester Role
            if any(role in content for role in ['Engineer Role', 'Reviewer Role', 'Tester Role', 'Coordinator Role']):
                agents_found = True

                # Extract task name from path
                task_name = Path(log_path).parent.name

                # Get last modification time
                mtime = os.path.getmtime(log_path)
                age_seconds = datetime.now().timestamp() - mtime
                age_minutes = age_seconds / 60

                # Determine status
                if age_minutes < 2:
                    status = "✅"
                    health = "healthy"
                elif age_minutes < 5:
                    status = "⚠️"
                    health = "slow"
                else:
                    status = "🚨"
                    health = "stalled"

                # Count commits (rough approximation)
                commit_count = content.count('- [x]')

                # Extract most recent progress entry
                lines = content.split('\n')
                recent_progress = "Working..."
                for line in reversed(lines[-50:]):  # Check last 50 lines
                    if line.startswith('###') and any(word in line.lower() for word in ['progress', 'complete', 'status', 'update']):
                        recent_progress = line.replace('###', '').strip()
                        break

                print(f"{status} {task_name}: {recent_progress} • {commit_count} tasks • Last update: {age_minutes:.0f}m ago ({health})")

        except Exception as e:
            print(f"❌ Error reading {log_path}: {e}")

    if not agents_found:
        # Work logs exist but no agent activity
        return

    print("\nNext check-in: 30s\n")

if __name__ == '__main__':
    check_agent_progress()
