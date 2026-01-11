#!/usr/bin/env python3
"""
Watchdog Health Check - Periodic System Validation

Runs every 5 minutes to detect systemic issues.
Outputs status to log file, only alerts user if critical issues found.
"""

import os
import sys
import glob
import subprocess
from datetime import datetime, timedelta
from pathlib import Path

def check_coordination_timer():
    """Verify coordination timer is running when agents exist."""
    result = subprocess.run(
        ['pgrep', '-f', 'coordination-timer.sh'],
        capture_output=True
    )
    return result.returncode == 0

def check_agent_work_logs():
    """Check if agents are making progress."""
    work_logs = glob.glob('.ai/tasks/*/20-work-log.md')
    if not work_logs:
        return True  # No agents, no problem

    # Check if work logs updated recently (within 10 minutes)
    stale_logs = []
    for log in work_logs:
        mtime = os.path.getmtime(log)
        age_minutes = (datetime.now().timestamp() - mtime) / 60

        if age_minutes > 10:
            stale_logs.append((log, age_minutes))

    if stale_logs:
        print(f"⚠️ WATCHDOG: {len(stale_logs)} agent work logs stale (>10 min)")
        for log, age in stale_logs:
            print(f"  - {log}: {age:.1f} minutes old")
        return False

    return True

def check_orchestrator_role_violations():
    """Check if Orchestrator is doing implementation work."""
    work_logs = glob.glob('.ai/tasks/*/20-work-log.md')

    violations = []
    for log in work_logs:
        with open(log, 'r') as f:
            content = f.read()

        # Check for Orchestrator + Engineer in same work log
        if 'Orchestrator Role' in content and 'Engineer Role' in content:
            # Same heading means role switch in same session
            if '## ' in content:
                violations.append(log)

    if violations:
        print(f"🚨 WATCHDOG: {len(violations)} Orchestrator role violations detected")
        for log in violations:
            print(f"  - {log}: Orchestrator switching to Engineer role")
        return False

    return True

def check_permissions():
    """Verify permissions configured for background agents."""
    settings_file = '.claude/settings.json'
    if not os.path.exists(settings_file):
        print("⚠️ WATCHDOG: settings.json missing")
        return False

    with open(settings_file, 'r') as f:
        content = f.read()

    # Check for required permissions
    required = ['Write(*)', 'Edit(*)', 'defaultMode']
    missing = [perm for perm in required if perm not in content]

    if missing:
        print(f"⚠️ WATCHDOG: Missing permissions in settings.json: {missing}")
        return False

    return True

def main():
    """Run all health checks."""
    print(f"\n[{datetime.now().strftime('%H:%M:%S')}] Watchdog health check...")

    checks = {
        'Coordination Timer': check_coordination_timer,
        'Agent Work Logs': check_agent_work_logs,
        'Role Boundaries': check_orchestrator_role_violations,
        'Permissions': check_permissions,
    }

    results = {}
    for name, check_fn in checks.items():
        try:
            results[name] = check_fn()
        except Exception as e:
            print(f"❌ {name}: Check failed with error: {e}")
            results[name] = False

    # Summary
    passed = sum(1 for v in results.values() if v)
    total = len(results)

    if passed == total:
        print(f"✅ All checks passed ({passed}/{total})")
    else:
        failed = total - passed
        print(f"⚠️ {failed}/{total} checks failed")

        # Log failures
        for name, passed in results.items():
            if not passed:
                print(f"  ❌ {name}")

    return 0 if passed == total else 1

if __name__ == '__main__':
    sys.exit(main())
