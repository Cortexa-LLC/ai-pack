#!/usr/bin/env python3
"""
Coordination Timer - Periodic Agent Monitoring

Triggers coordination check-ins every 30 seconds to monitor background agents.
Cross-platform Python implementation.

Usage: coordination-timer.py [interval_seconds] [max_checks]

Default: 30 seconds, 1200 checks (10 hours)
"""

import sys
import time
import subprocess
from datetime import datetime
from pathlib import Path

def main():
    """Run coordination timer."""
    # Parse arguments
    interval_seconds = int(sys.argv[1]) if len(sys.argv) > 1 else 30
    max_checks = int(sys.argv[2]) if len(sys.argv) > 2 else 1200

    checkpoint_file = Path('.claude/.coordination-checkpoint')
    log_file = Path('.claude/.coordination.log')

    # Ensure .claude directory exists
    checkpoint_file.parent.mkdir(exist_ok=True)

    # Initialize
    with open(log_file, 'a') as f:
        f.write(f"Coordination timer starting (interval: {interval_seconds}s, max: {max_checks} checks)\n")

    # Run timer loop
    check_count = 0
    while check_count < max_checks:
        time.sleep(interval_seconds)
        check_count += 1

        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")

        # Update checkpoint file
        with open(checkpoint_file, 'w') as f:
            f.write(f"{check_count}\n")
            f.write(f"{timestamp}\n")
            f.write(f"⏰ Coordination checkpoint {check_count} triggered at {timestamp}\n")

        # Log checkpoint
        with open(log_file, 'a') as f:
            f.write(f"[{timestamp}] Coordination checkpoint {check_count}\n")

        # Run coordination check (lightweight)
        try:
            subprocess.run(
                [sys.executable, '.claude/scripts/coordinator-check.py'],
                capture_output=True,
                timeout=10
            )
        except Exception:
            pass  # Continue even if check fails

    # Completed
    with open(log_file, 'a') as f:
        f.write(f"Coordination timer completed {max_checks} checks\n")

if __name__ == '__main__':
    main()
