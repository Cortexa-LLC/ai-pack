#!/usr/bin/env python3
"""
GitHub Integration Timer - Periodic sync and CI monitoring
Runs only during Claude Code sessions
"""

import sys
import time
import subprocess
from pathlib import Path
from datetime import datetime

def log(message):
    """Log with timestamp."""
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    print(f"[{timestamp}] [GitHub] {message}", flush=True)

def detect_ai_pack_root():
    """Detect AI-Pack installation location."""
    # Method 1: Check git submodules
    try:
        result = subprocess.run(
            ['git', 'config', '--file', '.gitmodules', '--get-regexp', 'path'],
            capture_output=True,
            text=True,
            cwd=Path(__file__).parent.parent.parent.parent
        )
        for line in result.stdout.split('\n'):
            if 'ai-pack' in line.lower():
                path = line.split()[-1]
                ai_pack_path = Path(__file__).parent.parent.parent.parent / path
                if ai_pack_path.exists():
                    return ai_pack_path
    except:
        pass

    # Method 2: Search for ai-pack directory
    search_root = Path(__file__).parent.parent.parent.parent
    for candidate in ['.ai-pack', 'ai-pack', 'tools/ai-pack']:
        ai_pack_path = search_root / candidate
        if ai_pack_path.exists():
            return ai_pack_path

    # Not found
    return None

def run_integration_command(command):
    """Run a GitHub integration command."""
    ai_pack_root = detect_ai_pack_root()

    if not ai_pack_root:
        log("ERROR: AI-Pack installation not found")
        return False

    script_path = ai_pack_root / 'scripts' / 'github-integration.py'

    if not script_path.exists():
        log(f"ERROR: github-integration.py not found at {script_path}")
        return False

    try:
        result = subprocess.run(
            ['python3', str(script_path), command],
            capture_output=True,
            text=True,
            timeout=60
        )

        if result.returncode == 0:
            return True
        else:
            log(f"Command '{command}' failed: {result.stderr}")
            return False
    except subprocess.TimeoutExpired:
        log(f"Command '{command}' timed out")
        return False
    except Exception as e:
        log(f"Error running '{command}': {e}")
        return False

def check_integration_enabled():
    """Check if GitHub integration is enabled."""
    ai_pack_root = detect_ai_pack_root()

    if not ai_pack_root:
        return False

    script_path = ai_pack_root / 'scripts' / 'github-integration.py'

    try:
        result = subprocess.run(
            ['python3', str(script_path), 'status'],
            capture_output=True,
            text=True,
            timeout=10
        )
        return 'ENABLED' in result.stdout
    except:
        return False

def main():
    """Main timer loop."""
    if len(sys.argv) < 3:
        print("Usage: github-integration-timer.py <interval_seconds> <max_iterations>")
        print("Example: github-integration-timer.py 300 120")
        sys.exit(1)

    interval = int(sys.argv[1])
    max_iterations = int(sys.argv[2])

    log(f"Starting GitHub integration timer (interval={interval}s, max_iterations={max_iterations})")

    # Check if integration is enabled
    if not check_integration_enabled():
        log("GitHub integration is not enabled - timer exiting")
        sys.exit(0)

    iteration = 0

    try:
        while iteration < max_iterations:
            iteration += 1

            log(f"Iteration {iteration}/{max_iterations}")

            # Run sync command
            log("Running sync...")
            if run_integration_command('sync'):
                log("Sync completed successfully")
            else:
                log("Sync failed or had errors")

            # Run CI check
            log("Checking CI status...")
            if run_integration_command('check-ci'):
                log("CI check completed")
            else:
                log("CI check failed")

            # Wait for next iteration
            if iteration < max_iterations:
                log(f"Sleeping for {interval} seconds...")
                time.sleep(interval)

        log(f"Max iterations ({max_iterations}) reached - timer exiting")

    except KeyboardInterrupt:
        log("Interrupted - timer exiting")
    except Exception as e:
        log(f"Error: {e}")
        sys.exit(1)

if __name__ == '__main__':
    main()
