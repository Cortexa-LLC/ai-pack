#!/usr/bin/env python3
"""
Tier 2 Test Environment Setup Script

Ensures all required configuration is in place for running tier 2 validation tests.
Creates .claude/settings.json with proper permissions if missing.
"""

import sys
import json
from pathlib import Path


def create_claude_directory():
    """Create .claude directory if it doesn't exist"""
    claude_dir = Path(".claude")
    if not claude_dir.exists():
        print("Creating .claude/ directory...")
        claude_dir.mkdir(parents=True)
        print("✓ Created .claude/")
        return True
    else:
        print("✓ .claude/ directory exists")
        return False


def setup_settings_json():
    """Create or update .claude/settings.json with required permissions"""
    settings_path = Path(".claude/settings.json")

    required_config = {
        "permissions": {
            "allow": [
                "Write(*)",
                "Edit(*)",
                "Read(*)"
            ],
            "defaultMode": "bypassPermissions"
        }
    }

    if not settings_path.exists():
        print("\nCreating .claude/settings.json with background agent permissions...")
        with open(settings_path, 'w') as f:
            json.dump(required_config, f, indent=2)
        print("✓ Created .claude/settings.json")
        return True

    # File exists, validate it
    print("\n.claude/settings.json exists, validating...")
    try:
        with open(settings_path) as f:
            settings = json.load(f)

        permissions = settings.get("permissions", {})
        allowed = permissions.get("allow", [])
        default_mode = permissions.get("defaultMode", "")

        required_perms = ["Write(*)", "Edit(*)", "Read(*)"]
        has_all_perms = all(perm in allowed for perm in required_perms)
        has_bypass = default_mode == "bypassPermissions"

        if has_all_perms and has_bypass:
            print("✓ .claude/settings.json has all required permissions")
            return False
        else:
            print("⚠ .claude/settings.json missing required permissions")
            print("\nCurrent configuration:")
            print(json.dumps(settings, indent=2))
            print("\nRequired configuration:")
            print(json.dumps(required_config, indent=2))

            response = input("\nUpdate settings.json with required permissions? [y/N]: ")
            if response.lower() == 'y':
                # Merge configurations
                if "permissions" not in settings:
                    settings["permissions"] = {}

                settings["permissions"]["allow"] = required_config["permissions"]["allow"]
                settings["permissions"]["defaultMode"] = required_config["permissions"]["defaultMode"]

                with open(settings_path, 'w') as f:
                    json.dump(settings, f, indent=2)

                print("✓ Updated .claude/settings.json")
                return True
            else:
                print("✗ Skipped updating settings.json")
                print("  Warning: Tier 2 tests may fail without proper permissions")
                return False

    except Exception as e:
        print(f"✗ Error reading settings.json: {e}")
        print("\nCurrent file may be invalid JSON")

        response = input("Replace with valid configuration? [y/N]: ")
        if response.lower() == 'y':
            with open(settings_path, 'w') as f:
                json.dump(required_config, f, indent=2)
            print("✓ Replaced .claude/settings.json")
            return True
        else:
            print("✗ Skipped replacing settings.json")
            return False


def create_test_artifacts_directory():
    """Create .ai/test-artifacts directory if it doesn't exist"""
    artifacts_dir = Path(".ai/test-artifacts")
    if not artifacts_dir.exists():
        print("\nCreating .ai/test-artifacts/ directory...")
        artifacts_dir.mkdir(parents=True)
        print("✓ Created .ai/test-artifacts/")
        return True
    else:
        print("✓ .ai/test-artifacts/ directory exists")
        return False


def verify_test_files():
    """Verify tier 2 test files are present"""
    print("\nVerifying tier 2 test files...")
    tests_dir = Path("tests")

    if not tests_dir.exists():
        print("✗ tests/ directory not found")
        print("  You may not be in the ai-pack root directory")
        return False

    tier2_tests = list(tests_dir.glob("test_tier2*.py"))
    expected_count = 7

    if len(tier2_tests) >= expected_count:
        print(f"✓ Found {len(tier2_tests)} tier 2 test files")
        return True
    else:
        print(f"⚠ Found {len(tier2_tests)} tier 2 test files (expected {expected_count})")
        print("  You may need to pull latest changes from git")
        return False


def main():
    """Run setup"""
    print("=" * 60)
    print("TIER 2 TEST ENVIRONMENT SETUP")
    print("=" * 60)

    # Check we're in the right directory
    if not Path("README.md").exists() or not Path("tests").exists():
        print("\n✗ ERROR: Must run from ai-pack root directory")
        print("\nCurrent directory:", Path.cwd())
        print("Expected: */ai-pack")
        print("\nRun: cd path/to/ai-pack && python3 tests/tools/setup-tier2-environment.py")
        return 1

    changes_made = []

    # 1. Create .claude directory
    if create_claude_directory():
        changes_made.append("Created .claude/ directory")

    # 2. Setup settings.json
    if setup_settings_json():
        changes_made.append("Configured .claude/settings.json")

    # 3. Create test artifacts directory
    if create_test_artifacts_directory():
        changes_made.append("Created .ai/test-artifacts/ directory")

    # 4. Verify test files
    verify_test_files()

    # Summary
    print("\n" + "=" * 60)
    print("SETUP COMPLETE")
    print("=" * 60)

    if changes_made:
        print("\nChanges made:")
        for change in changes_made:
            print(f"  • {change}")

        print("\n⚠ IMPORTANT: Restart Claude Code for settings to take effect")
    else:
        print("\n✓ Environment already configured - no changes needed")

    print("\n" + "=" * 60)
    print("NEXT STEPS")
    print("=" * 60)
    print("\n1. Restart Claude Code (if settings.json was created/updated)")
    print("2. Run diagnostic to verify:")
    print("   python3 tests/tools/check-tier2-environment.py")
    print("\n3. Run tier 2 tests:")
    print("   python3 tests/test_tier2_real_execution.py")
    print("\nOr run all tier 2 tests:")
    print("   python3 tests/run_tests.py --tier2")
    print("\n" + "=" * 60)

    return 0


if __name__ == "__main__":
    sys.exit(main())
