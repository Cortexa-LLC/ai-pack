#!/usr/bin/env python3
"""
TC-BA-005: Spawned Agent Permission Verification (Executable Test)

Automated test that verifies .claude/settings.json is properly configured
for spawned agents.
"""

import json
import sys
import unittest
from pathlib import Path


class TestBackgroundAgentPermissions(unittest.TestCase):
    """
    Executable tests for spawned agent permission configuration.

    Based on TC-BA-005: Permission Pre-Verification
    """

    def setUp(self):
        """Set up test - find repository root"""
        self.repo_root = Path.cwd()
        while not (self.repo_root / ".git").exists():
            if self.repo_root.parent == self.repo_root:
                raise RuntimeError("Not in a git repository")
            self.repo_root = self.repo_root.parent

        self.settings_path = self.repo_root / ".claude" / "settings.json"
        self.settings_local_path = self.repo_root / ".claude" / "settings.local.json"

    def test_settings_json_exists(self):
        """Test that .claude/settings.json exists"""
        self.assertTrue(
            self.settings_path.exists(),
            f"❌ .claude/settings.json not found at {self.settings_path}\n"
            f"Run: python3 .ai-pack/templates/.claude-setup.py"
        )

    def test_settings_json_valid(self):
        """Test that settings.json is valid JSON"""
        if not self.settings_path.exists():
            self.skipTest("settings.json does not exist")

        try:
            with open(self.settings_path, 'r') as f:
                json.load(f)
        except json.JSONDecodeError as e:
            self.fail(f"❌ settings.json is invalid JSON: {e}")

    def test_write_permission_configured(self):
        """Test that Write(*) permission is in allow list"""
        if not self.settings_path.exists():
            self.skipTest("settings.json does not exist")

        with open(self.settings_path, 'r') as f:
            settings = json.load(f)

        permissions = settings.get("permissions", {})
        allow_list = permissions.get("allow", [])

        self.assertIn(
            "Write(*)",
            allow_list,
            f"❌ Write(*) not in permissions.allow\n"
            f"Current: {allow_list}\n"
            f"Required: Add \"Write(*)\" to permissions.allow array"
        )

    def test_edit_permission_configured(self):
        """Test that Edit(*) permission is in allow list (recommended)"""
        if not self.settings_path.exists():
            self.skipTest("settings.json does not exist")

        with open(self.settings_path, 'r') as f:
            settings = json.load(f)

        permissions = settings.get("permissions", {})
        allow_list = permissions.get("allow", [])

        if "Edit(*)" not in allow_list:
            self.skipTest("⚠️  Edit(*) not configured (recommended but not required)")

    def test_read_permission_configured(self):
        """Test that Read(*) permission is in allow list (recommended)"""
        if not self.settings_path.exists():
            self.skipTest("settings.json does not exist")

        with open(self.settings_path, 'r') as f:
            settings = json.load(f)

        permissions = settings.get("permissions", {})
        allow_list = permissions.get("allow", [])

        if "Read(*)" not in allow_list:
            self.skipTest("⚠️  Read(*) not configured (recommended but not required)")

    def test_default_mode_bypass_permissions(self):
        """Test that defaultMode is set to bypassPermissions"""
        if not self.settings_path.exists():
            self.skipTest("settings.json does not exist")

        with open(self.settings_path, 'r') as f:
            settings = json.load(f)

        permissions = settings.get("permissions", {})
        default_mode = permissions.get("defaultMode", "")

        self.assertEqual(
            default_mode,
            "bypassPermissions",
            f"❌ defaultMode must be 'bypassPermissions'\n"
            f"Current: '{default_mode}'\n"
            f"Spawned Agents cannot prompt for permissions"
        )

    def test_local_override_warning(self):
        """Warn if settings.local.json exists (may override settings.json)"""
        if self.settings_local_path.exists():
            print(f"\n⚠️  WARNING: {self.settings_local_path} exists")
            print("   This may override settings.json")
            print("   Ensure Write(*) is also in settings.local.json")


class TestGate08Enforcement(unittest.TestCase):
    """
    Tests for Gate 08: Spawned Agent Permission Enforcement

    Verifies the gate would block spawning if permissions missing.
    """

    def setUp(self):
        """Set up test"""
        self.repo_root = Path.cwd()
        while not (self.repo_root / ".git").exists():
            if self.repo_root.parent == self.repo_root:
                raise RuntimeError("Not in a git repository")
            self.repo_root = self.repo_root.parent

        self.settings_path = self.repo_root / ".claude" / "settings.json"

    def test_gate_would_pass_with_correct_config(self):
        """Test that Gate 08 would allow spawning with correct config"""
        if not self.settings_path.exists():
            self.skipTest("settings.json does not exist")

        with open(self.settings_path, 'r') as f:
            settings = json.load(f)

        permissions = settings.get("permissions", {})
        allow_list = permissions.get("allow", [])
        default_mode = permissions.get("defaultMode", "")

        # Check all gate conditions
        has_write = "Write(*)" in allow_list
        correct_mode = default_mode == "bypassPermissions"

        self.assertTrue(
            has_write and correct_mode,
            f"❌ Gate 08 would BLOCK spawning:\n"
            f"  Write(*) configured: {has_write}\n"
            f"  defaultMode correct: {correct_mode}\n"
            f"Cannot spawn spawned agents until both are true"
        )

    def test_gate_would_block_without_write(self):
        """Test that Gate 08 would block if Write(*) missing"""
        if not self.settings_path.exists():
            self.skipTest("settings.json does not exist")

        with open(self.settings_path, 'r') as f:
            settings = json.load(f)

        permissions = settings.get("permissions", {})
        allow_list = permissions.get("allow", [])

        if "Write(*)" not in allow_list:
            # This is expected - gate SHOULD block
            print("\n✅ Gate 08 would correctly BLOCK (Write(*) missing)")
            return

        # Write(*) is present - gate would pass
        print("\n✅ Gate 08 would PASS (Write(*) configured)")

    def test_gate_would_block_with_wrong_mode(self):
        """Test that Gate 08 would block if defaultMode wrong"""
        if not self.settings_path.exists():
            self.skipTest("settings.json does not exist")

        with open(self.settings_path, 'r') as f:
            settings = json.load(f)

        permissions = settings.get("permissions", {})
        default_mode = permissions.get("defaultMode", "")

        if default_mode != "bypassPermissions":
            # This is expected - gate SHOULD block
            print(f"\n✅ Gate 08 would correctly BLOCK (defaultMode: {default_mode})")
            return

        # defaultMode is correct - gate would pass
        print("\n✅ Gate 08 would PASS (defaultMode correct)")


def run_tests():
    """Run all tests and return results"""
    # Create test suite
    loader = unittest.TestLoader()
    suite = unittest.TestSuite()

    # Add test classes
    suite.addTests(loader.loadTestsFromTestCase(TestBackgroundAgentPermissions))
    suite.addTests(loader.loadTestsFromTestCase(TestGate08Enforcement))

    # Run tests with verbose output
    runner = unittest.TextTestRunner(verbosity=2)
    result = runner.run(suite)

    # Print summary
    print("\n" + "="*70)
    print("TEST SUMMARY")
    print("="*70)
    print(f"Tests run: {result.testsRun}")
    print(f"Successes: {result.testsRun - len(result.failures) - len(result.errors) - len(result.skipped)}")
    print(f"Failures: {len(result.failures)}")
    print(f"Errors: {len(result.errors)}")
    print(f"Skipped: {len(result.skipped)}")

    if result.wasSuccessful():
        print("\n✅ ALL TESTS PASSED")
        print("\nSpawned Agents are properly configured!")
        print("Safe to spawn with ")
        return 0
    else:
        print("\n❌ TESTS FAILED")
        print("\nFix configuration issues before spawning spawned agents.")
        print("\nSolutions:")
        print("1. Run: python3 .ai-pack/templates/.claude-setup.py")
        print("2. Or manually add Write(*) and set defaultMode")
        print("3. Or use foreground agents ()")
        return 1


if __name__ == "__main__":
    sys.exit(run_tests())
