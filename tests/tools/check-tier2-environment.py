#!/usr/bin/env python3
"""
Tier 2 Test Environment Diagnostic
Run this on both machines and compare results
"""

import sys
import os
import subprocess
import json
from pathlib import Path

def check_settings():
    """Check .claude/settings.json configuration"""
    print("\n" + "=" * 60)
    print("1. Claude Code Settings (CRITICAL)")
    print("=" * 60)

    settings_path = Path(".claude/settings.json")

    if not settings_path.exists():
        print("✗ .claude/settings.json MISSING")
        print("\nTHIS IS LIKELY THE PROBLEM!")
        print("\nCreate .claude/settings.json with:")
        print(json.dumps({
            "permissions": {
                "allow": ["Write(*)", "Edit(*)", "Read(*)"],
                "defaultMode": "bypassPermissions"
            }
        }, indent=2))
        print("\nSee: docs/CLAUDE-CODE-CONFIGURATION.md")
        return False

    print("✓ .claude/settings.json exists")

    try:
        with open(settings_path) as f:
            settings = json.load(f)

        print("\nContent:")
        print(json.dumps(settings, indent=2))

        # Check required permissions
        permissions = settings.get("permissions", {})
        allowed = permissions.get("allow", [])
        default_mode = permissions.get("defaultMode", "")

        required = ["Write(*)", "Edit(*)", "Read(*)"]
        has_all = all(perm in allowed for perm in required)
        has_bypass = default_mode == "bypassPermissions"

        if has_all and has_bypass:
            print("\n✓ All required permissions present")
            return True
        else:
            print("\n✗ MISSING REQUIRED PERMISSIONS")
            print("  Background agents will fail without:")
            if not has_all:
                print("  - Write(*), Edit(*), Read(*)")
            if not has_bypass:
                print("  - defaultMode: bypassPermissions")
            return False

    except Exception as e:
        print(f"✗ Error reading settings.json: {e}")
        return False

def check_working_directory():
    """Check current working directory"""
    print("\n" + "=" * 60)
    print("2. Working Directory")
    print("=" * 60)

    cwd = Path.cwd()
    print(f"Current directory: {cwd}")
    print(f"Expected: */ai-pack")

    if cwd.name == "ai-pack":
        print("✓ In ai-pack directory")
        return True
    else:
        print("⚠ Not in ai-pack directory")
        return False

def check_git_state():
    """Check git branch and status"""
    print("\n" + "=" * 60)
    print("3. Git State")
    print("=" * 60)

    try:
        branch = subprocess.check_output(
            ["git", "branch", "--show-current"],
            text=True
        ).strip()
        print(f"Branch: {branch}")

        commit = subprocess.check_output(
            ["git", "log", "-1", "--oneline"],
            text=True
        ).strip()
        print(f"Latest commit: {commit}")

        status = subprocess.check_output(
            ["git", "status", "--short"],
            text=True
        ).strip()
        if status:
            print(f"\nUncommitted changes:")
            for line in status.split('\n')[:5]:
                print(f"  {line}")
        else:
            print("\nNo uncommitted changes")

        return True
    except Exception as e:
        print(f"✗ Git check failed: {e}")
        return False

def check_python():
    """Check Python version"""
    print("\n" + "=" * 60)
    print("4. Python Environment")
    print("=" * 60)

    print(f"Python version: {sys.version}")
    print(f"Python executable: {sys.executable}")
    return True

def check_test_files():
    """Check for tier 2 test files"""
    print("\n" + "=" * 60)
    print("5. Test Files")
    print("=" * 60)

    tests_dir = Path("tests")
    if not tests_dir.exists():
        print("✗ tests/ directory not found")
        return False

    tier2_tests = list(tests_dir.glob("test_tier2*.py"))
    print(f"Tier 2 test files found: {len(tier2_tests)}")
    print(f"Expected: 6")

    if tier2_tests:
        print("\nAvailable tier 2 tests:")
        for test in sorted(tier2_tests):
            print(f"  {test.name}")

    return len(tier2_tests) == 7

def check_test_artifacts():
    """Check test artifacts directory"""
    print("\n" + "=" * 60)
    print("6. Test Artifacts")
    print("=" * 60)

    artifacts_dir = Path(".ai/test-artifacts")
    if artifacts_dir.exists():
        print("✓ Test artifacts directory exists")

        tier2_runs = list(artifacts_dir.glob("tier2-*"))
        print(f"Existing test runs: {len(tier2_runs)}")
    else:
        print("No .ai/test-artifacts directory (will be created on first run)")

    return True

def check_test_import():
    """Check if tests can be imported"""
    print("\n" + "=" * 60)
    print("7. Test Import Check")
    print("=" * 60)

    sys.path.insert(0, "tests")
    try:
        import test_tier2_real_execution
        print("✓ test_tier2_real_execution can be imported")
        return True
    except Exception as e:
        print(f"✗ Import failed: {e}")
        return False

def check_beads():
    """Check if beads is installed"""
    print("\n" + "=" * 60)
    print("8. Beads Installation (Optional)")
    print("=" * 60)

    try:
        subprocess.run(
            ["bd", "--version"],
            capture_output=True,
            check=True
        )
        print("✓ beads (bd) is installed")
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        print("⚠ beads (bd) not found in PATH")
        print("  Install: https://github.com/steveyegge/beads")
        print("  (Optional for tier 2 tests)")
        return False

def main():
    """Run all checks"""
    print("=" * 60)
    print("TIER 2 TEST ENVIRONMENT DIAGNOSTIC")
    print("=" * 60)

    checks = [
        ("Settings", check_settings),
        ("Working Directory", check_working_directory),
        ("Git State", check_git_state),
        ("Python", check_python),
        ("Test Files", check_test_files),
        ("Test Artifacts", check_test_artifacts),
        ("Test Import", check_test_import),
        ("Beads", check_beads),
    ]

    results = {}
    for name, check_func in checks:
        try:
            results[name] = check_func()
        except Exception as e:
            print(f"\n✗ {name} check failed with exception: {e}")
            results[name] = False

    print("\n" + "=" * 60)
    print("DIAGNOSTIC SUMMARY")
    print("=" * 60)

    for name, passed in results.items():
        status = "✓" if passed else "✗"
        print(f"{status} {name}")

    critical_failed = not results.get("Settings", False)

    print("\n" + "=" * 60)
    if critical_failed:
        print("CRITICAL ISSUE DETECTED")
        print("Most common problem: Missing .claude/settings.json")
        print("See: docs/CLAUDE-CODE-CONFIGURATION.md")
    else:
        print("Environment looks good!")
        print("\nTo run tier 2 tests:")
        print("  python3 tests/test_tier2_real_execution.py")
    print("=" * 60)

    return 0 if not critical_failed else 1

if __name__ == "__main__":
    sys.exit(main())
