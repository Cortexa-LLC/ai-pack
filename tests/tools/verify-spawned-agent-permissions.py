#!/usr/bin/env python3
"""
Spawned Agent Permission Verification Utility

Verifies that .claude/settings.json is properly configured for spawned agents
that need to write files to the repository.

Based on TC-BA-005: Permission Pre-Verification test case.

Usage:
    python3 verify-background-agent-permissions.py [--repo-root PATH]
"""

import argparse
import json
import sys
from pathlib import Path
from typing import Dict, List, Tuple


class PermissionVerifier:
    """Verifies spawned agent permissions are configured correctly"""

    def __init__(self, repo_root: Path):
        self.repo_root = repo_root
        self.settings_path = repo_root / ".claude" / "settings.json"
        self.settings_local_path = repo_root / ".claude" / "settings.local.json"
        self.errors: List[str] = []
        self.warnings: List[str] = []
        self.settings: Dict = {}

    def verify_all(self) -> Tuple[bool, List[str], List[str]]:
        """
        Run all permission checks.

        Returns:
            Tuple of (success, errors, warnings)
        """
        print("🔍 Checking spawned agent permissions...\n")

        # Step 1: Check settings.json exists
        if not self._check_settings_exists():
            return False, self.errors, self.warnings

        # Step 2: Load and parse settings
        if not self._load_settings():
            return False, self.errors, self.warnings

        # Step 3: Check for settings.local.json override
        self._check_local_override()

        # Step 4: Verify Write(*) permission
        self._check_write_permission()

        # Step 5: Verify Edit(*) permission
        self._check_edit_permission()

        # Step 6: Verify Read(*) permission
        self._check_read_permission()

        # Step 7: Verify defaultMode
        self._check_default_mode()

        # Step 8: Check for common git permissions
        self._check_git_permissions()

        # Return results
        success = len(self.errors) == 0
        return success, self.errors, self.warnings

    def _check_settings_exists(self) -> bool:
        """Check if .claude/settings.json exists"""
        if not self.settings_path.exists():
            self.errors.append(
                f"❌ ERROR: .claude/settings.json not found\n"
                f"   Location: {self.settings_path}\n"
                f"\n"
                f"   SOLUTIONS:\n"
                f"   Option 1: Use ai-pack setup script (Recommended)\n"
                f"     python3 .ai-pack/templates/.claude-setup.py\n"
                f"\n"
                f"   Option 2: Manual setup\n"
                f"     mkdir -p .claude\n"
                f"     cp .ai-pack/templates/.claude/settings.json .claude/\n"
                f"\n"
                f"   Option 3: Run agents in FOREGROUND\n"
                f"     Use \n"
                f"     Agents will prompt for permissions interactively"
            )
            return False

        print(f"✅ .claude/settings.json exists")
        return True

    def _load_settings(self) -> bool:
        """Load and parse settings.json"""
        try:
            with open(self.settings_path, 'r') as f:
                self.settings = json.load(f)
            print(f"✅ settings.json is valid JSON")
            return True
        except json.JSONDecodeError as e:
            self.errors.append(
                f"❌ ERROR: .claude/settings.json is invalid JSON\n"
                f"   Error: {e}\n"
                f"   Fix: Validate JSON syntax and retry"
            )
            return False
        except Exception as e:
            self.errors.append(
                f"❌ ERROR: Could not read .claude/settings.json\n"
                f"   Error: {e}"
            )
            return False

    def _check_local_override(self):
        """Check if settings.local.json exists (may override settings.json)"""
        if self.settings_local_path.exists():
            self.warnings.append(
                f"⚠️  WARNING: .claude/settings.local.json exists\n"
                f"   This file may override settings.json\n"
                f"   Ensure Write(*) is also in settings.local.json\n"
                f"   OR remove settings.local.json to use settings.json only"
            )
            print(f"⚠️  settings.local.json detected (may override)")

    def _check_write_permission(self):
        """Verify Write(*) permission in allow list"""
        permissions = self.settings.get("permissions", {})
        allow_list = permissions.get("allow", [])

        if "Write(*)" not in allow_list:
            self.errors.append(
                f"❌ ERROR: Write(*) not in permissions.allow\n"
                f"   Current allow list: {allow_list}\n"
                f"\n"
                f"   REQUIRED: Add \"Write(*)\" to permissions.allow array\n"
                f"\n"
                f"   Example:\n"
                f"   {{\n"
                f"     \"permissions\": {{\n"
                f"       \"allow\": [\n"
                f"         \"Write(*)\",  // ← Add this\n"
                f"         \"Edit(*)\",\n"
                f"         \"Read(*)\"\n"
                f"       ]\n"
                f"     }}\n"
                f"   }}\n"
                f"\n"
                f"   WHY: Background agents need Write(*) to persist files"
            )
        else:
            print(f"✅ Write(*) permission configured")

    def _check_edit_permission(self):
        """Verify Edit(*) permission in allow list"""
        permissions = self.settings.get("permissions", {})
        allow_list = permissions.get("allow", [])

        if "Edit(*)" not in allow_list:
            self.warnings.append(
                f"⚠️  WARNING: Edit(*) not in permissions.allow\n"
                f"   Recommended: Add \"Edit(*)\" for file modifications\n"
                f"   Background agents may need to edit existing files"
            )
        else:
            print(f"✅ Edit(*) permission configured")

    def _check_read_permission(self):
        """Verify Read(*) permission in allow list"""
        permissions = self.settings.get("permissions", {})
        allow_list = permissions.get("allow", [])

        if "Read(*)" not in allow_list:
            self.warnings.append(
                f"⚠️  WARNING: Read(*) not in permissions.allow\n"
                f"   Recommended: Add \"Read(*)\" for reading files\n"
                f"   Background agents may need to read existing code"
            )
        else:
            print(f"✅ Read(*) permission configured")

    def _check_default_mode(self):
        """Verify defaultMode is set to bypassPermissions"""
        permissions = self.settings.get("permissions", {})
        default_mode = permissions.get("defaultMode", "")

        if default_mode != "bypassPermissions":
            self.errors.append(
                f"❌ ERROR: defaultMode not set to \"bypassPermissions\"\n"
                f"   Current value: \"{default_mode}\"\n"
                f"\n"
                f"   REQUIRED: Set defaultMode to \"bypassPermissions\"\n"
                f"\n"
                f"   Example:\n"
                f"   {{\n"
                f"     \"permissions\": {{\n"
                f"       \"defaultMode\": \"bypassPermissions\"  // ← Set this\n"
                f"     }}\n"
                f"   }}\n"
                f"\n"
                f"   WHY: Background agents cannot prompt for permissions\n"
                f"        Must have permissions pre-configured"
            )
        else:
            print(f"✅ defaultMode: bypassPermissions")

    def _check_git_permissions(self):
        """Check for common git permissions (recommended but not required)"""
        permissions = self.settings.get("permissions", {})
        allow_list = permissions.get("allow", [])

        # Check for git permissions
        has_git = any("git" in perm.lower() for perm in allow_list)

        if not has_git:
            self.warnings.append(
                f"⚠️  WARNING: No git permissions found\n"
                f"   Recommended: Add \"Bash(git:*)\" for git operations\n"
                f"   Background agents may need to run git commands"
            )
        else:
            print(f"✅ Git permissions configured")

    def print_results(self):
        """Print formatted results"""
        print("\n" + "="*60)
        print("PERMISSION VERIFICATION RESULTS")
        print("="*60 + "\n")

        if self.errors:
            print("ERRORS (must fix):")
            print("-" * 60)
            for error in self.errors:
                print(error)
                print()

        if self.warnings:
            print("WARNINGS (recommended fixes):")
            print("-" * 60)
            for warning in self.warnings:
                print(warning)
                print()

        if not self.errors and not self.warnings:
            print("✅ ALL CHECKS PASSED")
            print("\nBackground agents are properly configured for:")
            print("  - Writing files to repository")
            print("  - Editing existing files")
            print("  - Reading project files")
            print("  - Running git commands")
            print("\nSafe to spawn spawned agents with ")

        print()


def main():
    parser = argparse.ArgumentParser(
        description="Verify spawned agent permissions are configured correctly"
    )
    parser.add_argument(
        "--repo-root",
        type=Path,
        default=Path.cwd(),
        help="Path to repository root (default: current directory)"
    )

    args = parser.parse_args()

    # Verify repo root exists
    if not args.repo_root.exists():
        print(f"❌ ERROR: Repository root does not exist: {args.repo_root}")
        sys.exit(1)

    # Run verification
    verifier = PermissionVerifier(args.repo_root)
    success, errors, warnings = verifier.verify_all()

    # Print results
    verifier.print_results()

    # Exit with appropriate code
    if not success:
        print("❌ PERMISSION VERIFICATION FAILED")
        print("\nCannot safely spawn spawned agents until errors are fixed.")
        print("\nOptions:")
        print("1. Fix errors above and retry")
        print("2. Run: python3 .ai-pack/templates/.claude-setup.py")
        print("3. Use foreground agents ()")
        sys.exit(1)
    elif warnings:
        print("⚠️  PERMISSION VERIFICATION PASSED WITH WARNINGS")
        print("\nBackground agents will work but recommended fixes available.")
        sys.exit(0)
    else:
        print("✅ PERMISSION VERIFICATION PASSED")
        print("\nSafe to spawn spawned agents!")
        sys.exit(0)


if __name__ == "__main__":
    main()
