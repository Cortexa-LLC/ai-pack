#!/usr/bin/env python3
"""
Stop monitoring timers on session end
Cross-platform Python implementation
"""

import subprocess
import platform
import sys

def kill_process_by_name(process_name):
    """Kill a process by name (cross-platform)."""
    try:
        if platform.system() == "Windows":
            # Windows: find and kill Python processes running the script
            # First, get PIDs of Python processes with the script name in command line
            result = subprocess.run(
                ['wmic', 'process', 'where', f"commandline like '%{process_name}%'", 'get', 'processid'],
                capture_output=True,
                text=True
            )

            # Extract PIDs from output
            pids = []
            for line in result.stdout.split('\n'):
                line = line.strip()
                if line and line.isdigit():
                    pids.append(line)

            # Kill each PID
            for pid in pids:
                subprocess.run(['taskkill', '/F', '/PID', pid], capture_output=True)

            return len(pids) > 0
        else:
            # Unix-like: use pkill
            result = subprocess.run(
                ['pkill', '-f', process_name],
                capture_output=True
            )
            return result.returncode == 0
    except Exception:
        return False

def main():
    """Stop monitoring timers."""
    print("🛑 Stopping ai-pack monitoring timers...")

    # Stop coordination timer
    if kill_process_by_name('coordination-timer.py'):
        print("  ✅ Coordination timer stopped")
    else:
        print("  ⏩ No coordination timer running")

    # Stop watchdog timer
    if kill_process_by_name('watchdog-timer.py'):
        print("  ✅ Watchdog timer stopped")
    else:
        print("  ⏩ No watchdog timer running")

    print("✅ ai-pack monitoring stopped")

if __name__ == '__main__':
    main()
