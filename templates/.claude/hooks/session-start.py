#!/usr/bin/env python3
"""
Auto-start monitoring timers on Claude Code session launch
Cross-platform Python implementation
"""

import os
import sys
import subprocess
import platform
from pathlib import Path

def is_process_running(process_name):
    """Check if a process is running (cross-platform)."""
    try:
        if platform.system() == "Windows":
            # Windows: use tasklist
            result = subprocess.run(
                ['tasklist', '/FI', f'IMAGENAME eq python*'],
                capture_output=True,
                text=True
            )
            # Check if process_name appears in command line
            result2 = subprocess.run(
                ['wmic', 'process', 'where', f"commandline like '%{process_name}%'", 'get', 'processid'],
                capture_output=True,
                text=True
            )
            return process_name in result2.stdout
        else:
            # Unix-like: use pgrep
            result = subprocess.run(
                ['pgrep', '-f', process_name],
                capture_output=True
            )
            return result.returncode == 0
    except Exception:
        return False

def start_background_script(script_name, *args):
    """Start a Python script in the background (cross-platform)."""
    script_dir = Path(__file__).parent.parent / 'scripts'
    script_path = script_dir / script_name

    if not script_path.exists():
        print(f"  ⚠️ Script not found: {script_path}")
        return None

    try:
        if platform.system() == "Windows":
            # Windows: use CREATE_NO_WINDOW flag
            import subprocess
            proc = subprocess.Popen(
                [sys.executable, str(script_path)] + list(args),
                creationflags=subprocess.CREATE_NO_WINDOW if platform.system() == "Windows" else 0,
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL
            )
        else:
            # Unix-like: standard background process
            proc = subprocess.Popen(
                [sys.executable, str(script_path)] + list(args),
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL
            )
        return proc.pid
    except Exception as e:
        print(f"  ❌ Failed to start {script_name}: {e}")
        return None

def main():
    """Start monitoring timers."""
    print("🚀 Starting ai-pack monitoring timers...")

    # Start coordination timer (30-second intervals, 10-hour max)
    if not is_process_running('coordination-timer.py'):
        pid = start_background_script('coordination-timer.py', '30', '1200')
        if pid:
            print(f"✅ Coordination timer started (PID {pid})")
        else:
            print("❌ Failed to start coordination timer")
    else:
        print("⏩ Coordination timer already running")

    # Start watchdog timer (5-minute intervals, 10-hour max)
    if not is_process_running('watchdog-timer.py'):
        pid = start_background_script('watchdog-timer.py', '300', '120')
        if pid:
            print(f"✅ Watchdog timer started (PID {pid})")
        else:
            print("❌ Failed to start watchdog timer")
    else:
        print("⏩ Watchdog timer already running")

    print("✅ ai-pack monitoring active")

if __name__ == '__main__':
    main()
