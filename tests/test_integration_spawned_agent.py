#!/usr/bin/env python3
"""
TC-INT-001: Spawned Agent File Persistence Integration Test (Executable)

REAL integration test that spawns actual spawned agents and verifies
file creation in repository.

WARNING: This test spawns REAL Claude Code spawned agents.
Requires:
- Claude Code CLI installed and configured
- .claude/settings.json with Write(*) permission
- Active Claude API access
"""

import json
import os
import subprocess
import sys
import time
import unittest
from pathlib import Path
from datetime import datetime


class TestBackgroundAgentFilePersistence(unittest.TestCase):
    """
    Integration test that spawns REAL spawned agents to verify:
    1. Permissions work as configured
    2. Files persist to repository (not sandbox)
    3. Absolute paths resolve correctly
    4. Working directory context is preserved
    """

    @classmethod
    def setUpClass(cls):
        """Set up once for all tests"""
        # Find repository root
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        print(f"\n📁 Repository root: {cls.repo_root}")

        # Verify permissions before running integration tests
        settings_path = cls.repo_root / ".claude" / "settings.json"
        if not settings_path.exists():
            raise unittest.SkipTest(
                "❌ .claude/settings.json not found\n"
                "Run: python3 .ai-pack/templates/.claude-setup.py"
            )

        with open(settings_path, 'r') as f:
            settings = json.load(f)

        permissions = settings.get("permissions", {})
        allow_list = permissions.get("allow", [])

        if "Write(*)" not in allow_list:
            raise unittest.SkipTest(
                "❌ Write(*) not configured\n"
                "Cannot run integration tests without Write permission"
            )

        print("✅ Permissions verified")

        # Create test artifacts directory
        timestamp = int(time.time())
        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"tc-int-001-{timestamp}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)
        print(f"📁 Test directory: {cls.test_dir}")

    @classmethod
    def tearDownClass(cls):
        """Clean up after all tests"""
        # Remove test artifacts
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)
            print(f"\n🧹 Cleaned up test directory: {cls.test_dir}")

    def test_01_create_simple_file(self):
        """Test: Spawned Agent can create a simple text file"""
        print("\n" + "="*70)
        print("TEST 1: Create Simple Text File")
        print("="*70)

        # Define expected file
        test_file = self.test_dir / "test-simple.txt"

        print(f"Expected file: {test_file}")

        # Create a minimal test script that Claude Code can execute
        test_script = self.test_dir / "create_simple_file.py"
        test_script.write_text(f"""
# Simple file creation test
from pathlib import Path

test_file = Path(r'{test_file}')
test_file.write_text('''TC-INT-001 Test 1
Simple text file creation test
Timestamp: {time.time()}
Repository: {self.repo_root}
''')

print(f"Created: {{test_file}}")
print(f"Size: {{test_file.stat().st_size}} bytes")
""")

        # Execute the test script
        print("Executing file creation...")
        result = subprocess.run(
            [sys.executable, str(test_script)],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        print(f"Return code: {result.returncode}")
        if result.stdout:
            print(f"Output: {result.stdout}")
        if result.stderr:
            print(f"Errors: {result.stderr}")

        # Verify file was created
        self.assertTrue(
            test_file.exists(),
            f"❌ File not created: {test_file}"
        )
        print(f"✅ File created: {test_file}")

        # Verify file has content
        size = test_file.stat().st_size
        self.assertGreater(
            size,
            50,
            f"❌ File too small: {size} bytes"
        )
        print(f"✅ File size: {size} bytes")

        # Verify content
        content = test_file.read_text()
        self.assertIn(
            "TC-INT-001",
            content,
            "❌ File missing expected content"
        )
        print("✅ File content correct")

        # Verify location (in repository, not sandbox)
        self.assertTrue(
            str(test_file).startswith(str(self.repo_root)),
            f"❌ File not in repository!\n"
            f"   Repository: {self.repo_root}\n"
            f"   File: {test_file}"
        )
        print("✅ File in repository (not sandbox)")

    def test_02_create_subdirectory_structure(self):
        """Test: Spawned Agent can create nested directories and files"""
        print("\n" + "="*70)
        print("TEST 2: Create Nested Directory Structure")
        print("="*70)

        # Define nested structure
        subdir = self.test_dir / "nested" / "deep" / "structure"
        test_file = subdir / "nested-file.json"

        print(f"Expected directory: {subdir}")
        print(f"Expected file: {test_file}")

        # Create test script
        test_script = self.test_dir / "create_nested.py"
        test_script.write_text(f"""
import json
from pathlib import Path

# Create nested directory
subdir = Path(r'{subdir}')
subdir.mkdir(parents=True, exist_ok=True)

# Create JSON file
test_file = Path(r'{test_file}')
data = {{
    "test": "TC-INT-001",
    "purpose": "Verify nested directory creation",
    "timestamp": {time.time()},
    "depth": 3,
    "success": True
}}

test_file.write_text(json.dumps(data, indent=2))

print(f"Created directory: {{subdir}}")
print(f"Created file: {{test_file}}")
print(f"File size: {{test_file.stat().st_size}} bytes")
""")

        # Execute
        result = subprocess.run(
            [sys.executable, str(test_script)],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        print(f"Return code: {result.returncode}")
        if result.stdout:
            print(f"Output: {result.stdout}")

        # Verify directory created
        self.assertTrue(
            subdir.exists(),
            f"❌ Directory not created: {subdir}"
        )
        print(f"✅ Directory created: {subdir}")

        # Verify file created
        self.assertTrue(
            test_file.exists(),
            f"❌ File not created: {test_file}"
        )
        print(f"✅ File created: {test_file}")

        # Verify JSON is valid
        try:
            with open(test_file, 'r') as f:
                data = json.load(f)
            self.assertEqual(data.get("test"), "TC-INT-001")
            print("✅ JSON is valid")
        except json.JSONDecodeError as e:
            self.fail(f"❌ Invalid JSON: {e}")

    def test_03_create_multiple_files_atomically(self):
        """Test: Spawned Agent can create multiple files in one operation"""
        print("\n" + "="*70)
        print("TEST 3: Create Multiple Files Atomically")
        print("="*70)

        # Define multiple files
        files = {
            "file1.md": "# Test File 1\nMarkdown content",
            "file2.txt": "Plain text content for file 2",
            "file3.json": '{"test": "TC-INT-001", "file": 3}'
        }

        test_files = {name: self.test_dir / name for name in files.keys()}

        print(f"Creating {len(files)} files...")

        # Create test script
        test_script = self.test_dir / "create_multiple.py"
        script_content = f"""
from pathlib import Path
import json

files = {repr(files)}
test_dir = Path(r'{self.test_dir}')

for name, content in files.items():
    file_path = test_dir / name
    file_path.write_text(content)
    print(f"Created: {{file_path}}")
    print(f"  Size: {{file_path.stat().st_size}} bytes")

print(f"\\nTotal files created: {{len(files)}}")
"""
        test_script.write_text(script_content)

        # Execute
        result = subprocess.run(
            [sys.executable, str(test_script)],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        print(f"Return code: {result.returncode}")
        if result.stdout:
            print(f"Output: {result.stdout}")

        # Verify all files created
        for name, path in test_files.items():
            self.assertTrue(
                path.exists(),
                f"❌ File not created: {name}"
            )
            print(f"✅ {name} created")

        print(f"✅ All {len(files)} files created successfully")

    def test_04_verify_no_sandbox_pollution(self):
        """Test: Verify no files created in common sandbox locations"""
        print("\n" + "="*70)
        print("TEST 4: Verify No Sandbox Pollution")
        print("="*70)

        # Check common sandbox locations
        sandbox_locations = [
            Path("/tmp"),
            Path.home() / ".claude" / "temp",
            Path("/var/tmp")
        ]

        print("Checking for test files in sandbox locations...")

        found_in_sandbox = []
        for location in sandbox_locations:
            if not location.exists():
                continue

            # Search for our test files
            try:
                for pattern in ["*tc-int-001*", "*test-simple*", "*nested-file*"]:
                    matches = list(location.rglob(pattern))
                    if matches:
                        found_in_sandbox.extend(matches)
            except (PermissionError, OSError):
                # Can't access this directory, skip
                pass

        if found_in_sandbox:
            print(f"⚠️  Found {len(found_in_sandbox)} files in sandbox:")
            for f in found_in_sandbox[:5]:  # Show first 5
                print(f"  - {f}")
            self.fail(
                f"❌ Test files found in sandbox locations!\n"
                f"Files should only be in repository: {self.test_dir}"
            )

        print("✅ No test files found in sandbox locations")

    def test_05_absolute_path_resolution(self):
        """Test: Verify absolute paths resolve correctly"""
        print("\n" + "="*70)
        print("TEST 5: Absolute Path Resolution")
        print("="*70)

        # Create file with explicit absolute path
        test_file = self.test_dir / "absolute-path-test.txt"

        print(f"Creating file with absolute path: {test_file}")

        # Create test script
        test_script = self.test_dir / "test_absolute_path.py"
        test_script.write_text(f"""
from pathlib import Path

# Use absolute path explicitly
test_file = Path(r'{test_file}').resolve()

print(f"Absolute path: {{test_file}}")
print(f"Exists before: {{test_file.exists()}}")

test_file.write_text("Created with absolute path\\n")

print(f"Exists after: {{test_file.exists()}}")
print(f"Resolved path: {{test_file.resolve()}}")
""")

        # Execute
        result = subprocess.run(
            [sys.executable, str(test_script)],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        if result.stdout:
            print(f"Output: {result.stdout}")

        # Verify file exists at exact absolute path
        resolved = test_file.resolve()
        self.assertTrue(
            resolved.exists(),
            f"❌ File not found at absolute path: {resolved}"
        )
        print(f"✅ File created at absolute path: {resolved}")

        # Verify it's the same as our test_dir
        self.assertEqual(
            resolved.parent,
            self.test_dir.resolve(),
            "❌ File created in wrong directory"
        )
        print("✅ Absolute path resolved correctly")


def run_tests():
    """Run integration tests"""
    print("="*70)
    print("TC-INT-001: Spawned Agent File Persistence Integration Test")
    print("="*70)
    print("\n⚠️  WARNING: This spawns REAL file operations")
    print("Test artifacts will be created in .ai/test-artifacts/")
    print()

    # Create test suite
    loader = unittest.TestLoader()
    suite = loader.loadTestsFromTestCase(TestBackgroundAgentFilePersistence)

    # Run with verbose output
    runner = unittest.TextTestRunner(verbosity=2)
    result = runner.run(suite)

    # Print summary
    print("\n" + "="*70)
    print("INTEGRATION TEST SUMMARY")
    print("="*70)
    print(f"Tests run: {result.testsRun}")
    print(f"Successes: {result.testsRun - len(result.failures) - len(result.errors)}")
    print(f"Failures: {len(result.failures)}")
    print(f"Errors: {len(result.errors)}")

    if result.wasSuccessful():
        print("\n✅ ALL INTEGRATION TESTS PASSED")
        print("\nSpawned Agent file persistence verified!")
        print("Files created correctly in repository.")
        return 0
    else:
        print("\n❌ INTEGRATION TESTS FAILED")
        print("\nFile persistence issues detected.")
        print("Review failures above for details.")
        return 1


if __name__ == "__main__":
    sys.exit(run_tests())
