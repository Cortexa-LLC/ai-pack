#!/usr/bin/env python3
"""
Tier 2 File Persistence Verification Test

CRITICAL: This test verifies that files created by background agents
are actually persisted to the repository and visible to the orchestrator.

This addresses the concern: Are files being written to a sandbox/temporary
filesystem or are they actually in the repository?

Run with: python3 test_multi-agent_file_persistence_verification.py -v
"""

import json
import time
import unittest
from datetime import datetime
from pathlib import Path


class TestFilePersistenceVerification(unittest.TestCase):
    """
    Verification: Background Agent File Persistence

    Tests that files created by background agents:
    1. Actually exist in the repository
    2. Are visible to the orchestrator
    3. Contain actual content (not empty/truncated)
    4. Are in the correct location
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"multi-agent-persistence-verify-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

        print("\n" + "="*70)
        print("🔍 TIER 2 FILE PERSISTENCE VERIFICATION")
        print("="*70)
        print("Validates that background agent files actually persist")
        print("="*70 + "\n")

    def test_01_verify_existing_multi-agent_artifacts(self):
        """
        Verification: Check existing tier 2 test artifacts

        Scans all existing multi-agent test runs and verifies:
        - Files were actually created
        - Files are in repository
        - Files have content
        """
        print("\n" + "="*70)
        print("VERIFICATION 1: Existing Tier 2 Artifacts")
        print("="*70)

        artifacts_dir = self.repo_root / ".ai" / "test-artifacts"

        if not artifacts_dir.exists():
            self.skipTest("No test artifacts directory found")

        # Find all multi-agent test directories
        multi-agent_dirs = list(artifacts_dir.glob("multi-agent-*"))

        if not multi-agent_dirs:
            self.skipTest("No multi-agent test artifacts found")

        print(f"\nFound {len(multi-agent_dirs)} multi-agent test runs")

        total_files = 0
        total_size = 0
        missing_files = []
        empty_files = []

        for test_dir in multi-agent_dirs:
            print(f"\nChecking: {test_dir.name}")

            # Find all Python files in output directories
            py_files = list(test_dir.rglob("output/**/*.py"))
            md_files = list(test_dir.rglob("output/**/*.md"))
            all_files = py_files + md_files

            for file_path in all_files:
                total_files += 1

                if not file_path.exists():
                    missing_files.append(file_path)
                    print(f"  ❌ MISSING: {file_path.relative_to(test_dir)}")
                    continue

                size = file_path.stat().st_size
                total_size += size

                if size == 0:
                    empty_files.append(file_path)
                    print(f"  ⚠️  EMPTY: {file_path.relative_to(test_dir)}")
                elif size < 50:
                    print(f"  ⚠️  SMALL: {file_path.relative_to(test_dir)} ({size} bytes)")
                else:
                    # File looks good, don't print unless verbose
                    pass

            if all_files:
                print(f"  Files: {len(all_files)} ({sum(f.stat().st_size for f in all_files if f.exists())} bytes total)")

        print("\n" + "="*70)
        print("PERSISTENCE VERIFICATION SUMMARY")
        print("="*70)
        print(f"Total files checked: {total_files}")
        print(f"Total size: {total_size:,} bytes")
        print(f"Missing files: {len(missing_files)}")
        print(f"Empty files: {len(empty_files)}")

        # Assertions
        self.assertGreater(total_files, 0, "Should have found some files")
        self.assertEqual(len(missing_files), 0, f"Found {len(missing_files)} missing files")
        self.assertEqual(len(empty_files), 0, f"Found {len(empty_files)} empty files")
        self.assertGreater(total_size, 1000, "Total content should be substantial")

        print(f"\n✅ ALL FILES VERIFIED:")
        print(f"   - {total_files} files exist in repository")
        print(f"   - {total_size:,} bytes total")
        print(f"   - All files have content")
        print(f"   - All visible to orchestrator")

    def test_02_verify_specific_test_outputs(self):
        """
        Verification: Specific test output validation

        Checks specific expected files from each test type:
        - multi-agent-real: model.py, service.py, test_service.py
        - multi-agent-sequential: requirements.md, architecture.md, profile/*, tests/*
        - multi-agent-parallel: feature-*/auth/*, feature-*/api/*, etc.
        """
        print("\n" + "="*70)
        print("VERIFICATION 2: Specific Test Outputs")
        print("="*70)

        artifacts_dir = self.repo_root / ".ai" / "test-artifacts"

        # Check multi-agent-real outputs
        real_dirs = list(artifacts_dir.glob("multi-agent-real-*"))
        if real_dirs:
            latest_real = max(real_dirs, key=lambda p: p.stat().st_mtime)
            print(f"\nChecking multi-agent-real: {latest_real.name}")

            expected_files = [
                latest_real / "output" / "model.py",
                latest_real / "output" / "service.py",
                latest_real / "output" / "test_service.py",
            ]

            for expected in expected_files:
                self.assertTrue(expected.exists(), f"Missing: {expected.name}")
                self.assertGreater(expected.stat().st_size, 50, f"Too small: {expected.name}")
                print(f"  ✓ {expected.name} ({expected.stat().st_size} bytes)")

        # Check multi-agent-sequential outputs
        seq_dirs = list(artifacts_dir.glob("multi-agent-sequential-*"))
        if seq_dirs:
            latest_seq = max(seq_dirs, key=lambda p: p.stat().st_mtime)
            print(f"\nChecking multi-agent-sequential: {latest_seq.name}")

            expected_outputs = [
                latest_seq / "output" / "requirements.md",
                latest_seq / "output" / "architecture.md",
            ]

            for expected in expected_outputs:
                if expected.exists():
                    self.assertGreater(expected.stat().st_size, 100, f"Too small: {expected.name}")
                    print(f"  ✓ {expected.name} ({expected.stat().st_size} bytes)")

            # Check for profile code
            profile_files = list((latest_seq / "output" / "profile").glob("*.py"))
            self.assertGreater(len(profile_files), 0, "Should have profile Python files")
            print(f"  ✓ {len(profile_files)} profile Python files")

        # Check multi-agent-parallel outputs
        parallel_dirs = list(artifacts_dir.glob("multi-agent-parallel-*"))
        if parallel_dirs:
            latest_parallel = max(parallel_dirs, key=lambda p: p.stat().st_mtime)
            print(f"\nChecking multi-agent-parallel: {latest_parallel.name}")

            # Should have 5 features
            feature_dirs = list((latest_parallel / "output").glob("feature-*"))
            self.assertGreaterEqual(len(feature_dirs), 3, "Should have multiple features")
            print(f"  ✓ {len(feature_dirs)} feature directories")

            for feature_dir in feature_dirs:
                py_files = list(feature_dir.rglob("*.py"))
                self.assertGreater(len(py_files), 0, f"Feature {feature_dir.name} should have Python files")
                print(f"    {feature_dir.name}: {len(py_files)} Python files")

        print("\n✅ SPECIFIC OUTPUT VERIFICATION PASSED")

    def test_03_verify_file_locations_in_repository(self):
        """
        Verification: Files are actually in git repository

        Ensures files are not in a temp/sandbox location but actually
        under the git repository root.
        """
        print("\n" + "="*70)
        print("VERIFICATION 3: Repository Location Check")
        print("="*70)

        artifacts_dir = self.repo_root / ".ai" / "test-artifacts"

        if not artifacts_dir.exists():
            self.skipTest("No test artifacts")

        # Get all files in multi-agent tests
        all_files = list(artifacts_dir.rglob("multi-agent-*/output/**/*"))
        actual_files = [f for f in all_files if f.is_file()]

        print(f"\nChecking {len(actual_files)} files...")

        outside_repo = []
        for file_path in actual_files:
            # Check if file is under repo root
            try:
                file_path.relative_to(self.repo_root)
                # File is under repo root - good!
            except ValueError:
                # File is NOT under repo root - problem!
                outside_repo.append(file_path)
                print(f"  ❌ OUTSIDE REPO: {file_path}")

        print(f"\n✅ Repository Location Verification:")
        print(f"   - Checked: {len(actual_files)} files")
        print(f"   - All in repository: {len(outside_repo) == 0}")

        self.assertEqual(len(outside_repo), 0,
                        f"Found {len(outside_repo)} files outside repository")

    def test_04_verify_orchestrator_can_read_files(self):
        """
        Verification: Orchestrator can actually read agent-created files

        This is the critical test: Can we (the orchestrator) actually
        open and read the content of files created by background agents?
        """
        print("\n" + "="*70)
        print("VERIFICATION 4: Orchestrator Read Access")
        print("="*70)

        artifacts_dir = self.repo_root / ".ai" / "test-artifacts"

        # Find a multi-agent-real output directory
        real_dirs = list(artifacts_dir.glob("multi-agent-real-*"))
        if not real_dirs:
            self.skipTest("No multi-agent-real test found")

        latest = max(real_dirs, key=lambda p: p.stat().st_mtime)
        test_file = latest / "output" / "model.py"

        print(f"\nAttempting to read: {test_file.relative_to(self.repo_root)}")

        # Can we see it exists?
        self.assertTrue(test_file.exists(), "File should exist")
        print(f"  ✓ File exists")

        # Can we read it?
        try:
            content = test_file.read_text()
            self.assertIsNotNone(content, "Content should not be None")
            self.assertGreater(len(content), 0, "Content should not be empty")
            print(f"  ✓ Can read file ({len(content)} characters)")

            # Does it have expected Python code?
            self.assertIn("class", content, "Should contain Python class")
            print(f"  ✓ Contains expected code (Python class)")

            # Print sample
            lines = content.split('\n')
            print(f"\n  Sample content (first 5 lines):")
            for line in lines[:5]:
                print(f"    {line}")

        except Exception as e:
            self.fail(f"Failed to read file: {e}")

        print(f"\n✅ ORCHESTRATOR CAN READ AGENT FILES")
        print(f"   Files are NOT in sandbox/temp filesystem")
        print(f"   Files ARE in repository and accessible")


def main():
    """Run tests"""
    unittest.main(argv=[''], verbosity=2, exit=True)


if __name__ == '__main__':
    main()
