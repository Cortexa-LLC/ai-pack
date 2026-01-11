#!/usr/bin/env python3
"""
Watchdog Timer - Periodic System Health Checks

Runs system health checks every 5 minutes to detect systemic issues.
Cross-platform Python implementation.

Usage: watchdog-timer.py [interval_seconds] [max_checks]

Default: 300 seconds (5 minutes), 120 checks (10 hours)
"""

import sys
import time
import subprocess
from datetime import datetime
from pathlib import Path

def main():
    """Run watchdog timer."""
    # Parse arguments
    interval_seconds = int(sys.argv[1]) if len(sys.argv) > 1 else 300
    max_checks = int(sys.argv[2]) if len(sys.argv) > 2 else 120

    checkpoint_file = Path('.claude/.watchdog-checkpoint')
    log_file = Path('.claude/.watchdog.log')

    # Ensure .claude directory exists
    checkpoint_file.parent.mkdir(exist_ok=True)

    # Initialize
    with open(log_file, 'a') as f:
        f.write(f"Watchdog timer starting (interval: {interval_seconds}s, max: {max_checks} checks)\n")

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
            f.write(f"🐕 Watchdog checkpoint {check_count} triggered at {timestamp}\n")

        # Log checkpoint
        with open(log_file, 'a') as f:
            f.write(f"[{timestamp}] Watchdog checkpoint {check_count}\n")

        # Run health check (lightweight)
        try:
            result = subprocess.run(
                [sys.executable, '.claude/scripts/watchdog-check.py'],
                capture_output=True,
                text=True,
                timeout=30
            )
            # Log output
            with open(log_file, 'a') as f:
                f.write(result.stdout)
                if result.stderr:
                    f.write(result.stderr)
        except Exception as e:
            with open(log_file, 'a') as f:
                f.write(f"Error running health check: {e}\n")

    # Completed
    with open(log_file, 'a') as f:
        f.write(f"Watchdog timer completed {max_checks} checks\n")

if __name__ == '__main__':
    main()
