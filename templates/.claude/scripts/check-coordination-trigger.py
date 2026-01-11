#!/usr/bin/env python3
"""
Coordination Trigger Checker
Checks if a coordination checkpoint has been reached and provides guidance.
"""

import os
import sys
from pathlib import Path

def check_coordination_trigger():
    """Check if coordination checkpoint has been triggered."""
    checkpoint_file = Path(".claude/.coordination-checkpoint")
    last_check_file = Path(".claude/.coordination-last-check")

    # If checkpoint file doesn't exist, no timer running
    if not checkpoint_file.exists():
        return {
            "triggered": False,
            "reason": "No coordination timer running",
            "action": "none"
        }

    # Read current checkpoint
    try:
        with open(checkpoint_file, 'r') as f:
            lines = f.readlines()
            if len(lines) < 3:
                return {"triggered": False, "reason": "Invalid checkpoint file", "action": "none"}

            current_count = int(lines[0].strip())
            timestamp = lines[1].strip()
            message = lines[2].strip()
    except Exception as e:
        return {"triggered": False, "reason": f"Error reading checkpoint: {e}", "action": "none"}

    # Read last processed checkpoint
    last_count = 0
    if last_check_file.exists():
        try:
            with open(last_check_file, 'r') as f:
                last_count = int(f.read().strip())
        except:
            last_count = 0

    # Check if new checkpoint reached
    if current_count > last_count:
        # Update last check
        with open(last_check_file, 'w') as f:
            f.write(str(current_count))

        return {
            "triggered": True,
            "checkpoint": current_count,
            "timestamp": timestamp,
            "message": message,
            "action": "coordinate"
        }

    return {
        "triggered": False,
        "reason": "No new checkpoint",
        "action": "none"
    }

def main():
    """Main entry point."""
    result = check_coordination_trigger()

    if result["triggered"]:
        print(f"\n{'='*60}")
        print(f"🚨 COORDINATION CHECK-IN REQUIRED")
        print(f"{'='*60}")
        print(f"Checkpoint: {result['checkpoint']}")
        print(f"Timestamp: {result['timestamp']}")
        print(f"Message: {result['message']}")
        print(f"\n{'='*60}")
        print(f"ACTION REQUIRED:")
        print(f"1. Check agent status: /ai-pack agents")
        print(f"2. Review work logs: Read .ai/tasks/*/20-work-log.md")
        print(f"3. Check git activity: git log -5 --oneline")
        print(f"4. Identify any blocked/stuck agents")
        print(f"5. Intervene if needed")
        print(f"6. Report status")
        print(f"{'='*60}\n")
        sys.exit(1)  # Exit with error code to trigger hook attention
    else:
        # Silent success - no coordination needed yet
        sys.exit(0)

if __name__ == "__main__":
    main()
