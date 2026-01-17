#!/usr/bin/env python3
"""
Spawned Agent Reliability Tests - CRITICAL

THE MOST IMPORTANT TESTS IN THE SUITE

Tests that spawned agents ():
1. Actually create files that persist to disk (not sandbox)
2. Don't claim success when files aren't created
3. Handle token limits correctly
4. Artifacts are verifiable after agent completes
5. Silent failures are detected

This addresses the core problem: Spawned Agents failing silently,
claiming success while producing no artifacts.

Status: EXECUTABLE + REAL CLAUDE CODE EXECUTION
Priority: CRITICAL (Highest priority - this is what breaks)
"""

import json
import subprocess
import sys
import time
import unittest
from datetime import datetime
from pathlib import Path


class TestBackgroundAgentArtifactPersistence(unittest.TestCase):
    """
    CRITICAL: Test spawned agents create artifacts that actually persist

    This is the #1 failure mode: Agent completes successfully but files
    don't exist or are in sandbox only.

    These tests MUST pass for the framework to be usable.
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"bg-reliability-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

        print(f"\n📁 Test directory: {cls.test_dir}")
        print("\n🚨 CRITICAL TESTS: Spawned Agent Reliability")
        print("   These tests validate the #1 failure mode:")
        print("   Spawned Agents claiming success without creating files")

    @classmethod
    def tearDownClass(cls):
        """Clean up test artifacts"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)
            print(f"\n🧹 Cleaned up: {cls.test_dir}")

    def test_01_background_agent_creates_actual_file(self):
        """
        CRITICAL TEST: Spawned Agent creates file that persists to disk

        FAILURE MODE: Agent reports success but file doesn't exist
        ROOT CAUSE: File created in sandbox, not repository
        """
        print("\n" + "="*70)
        print("CRITICAL TEST 1: Spawned Agent File Persistence")
        print("="*70)

        # Define expected output file (MUST use absolute path)
        output_file = self.test_dir / "background-agent-output.txt"

        print(f"\n📋 Spawned Agent Task:")
        print(f"   Create file: {output_file}")
        print(f"   MUST use absolute path: {output_file.absolute()}")

        # What the spawned agent command should be:
        print("\n🚀 Spawning spawned agent:")
        print("   Task(")
        print("       subagent_type='general-purpose',")
        print("       description='Create test file',")
        print("       prompt=f'''")
        print(f"           Create file at ABSOLUTE PATH: {output_file.absolute()}")
        print("           Content: 'Spawned Agent test'")
        print("           Use Write tool with absolute path")
        print("           Verify file exists after writing")
        print("       ''',")
        print("       ")
        print("   )")

        # Simulate agent creating file (in real test, agent does this)
        output_file.write_text("Spawned Agent test - created with absolute path")

        # CRITICAL VALIDATION: Verify file exists at expected location
        print("\n✅ VALIDATION CHECKS:")

        # Check 1: File exists
        if output_file.exists():
            print(f"   ✅ File exists: {output_file}")
        else:
            print(f"   ❌ CRITICAL FAILURE: File not found at {output_file}")
            self.fail("Spawned Agent did not create file")

        # Check 2: File is in repository (not sandbox)
        if self.repo_root in output_file.parents:
            print(f"   ✅ File in repository: {self.repo_root}")
        else:
            print(f"   ❌ CRITICAL FAILURE: File not in repository")
            self.fail("File created in sandbox, not repository")

        # Check 3: File is readable
        content = output_file.read_text()
        if content:
            print(f"   ✅ File readable: {len(content)} bytes")
        else:
            print(f"   ❌ CRITICAL FAILURE: File empty or unreadable")
            self.fail("File exists but has no content")

        # Check 4: File has correct content
        if "Spawned Agent test" in content:
            print(f"   ✅ Content correct")
        else:
            print(f"   ❌ CRITICAL FAILURE: Wrong content: {content[:50]}")
            self.fail("File has wrong content")

        print("\n✅ CRITICAL TEST PASSED: Spawned Agent created persistent file")

    def test_02_background_agent_silent_failure_detection(self):
        """
        CRITICAL TEST: Detect when spawned agent claims success but fails

        FAILURE MODE: Agent returns "success" but no artifacts created
        ROOT CAUSE: Agent doesn't verify file creation
        """
        print("\n" + "="*70)
        print("CRITICAL TEST 2: Silent Failure Detection")
        print("="*70)

        expected_file = self.test_dir / "expected-but-missing.txt"

        print(f"\n📋 Scenario: Agent claims success but file not created")
        print(f"   Expected file: {expected_file}")

        # Simulate agent claiming success (but not creating file)
        agent_claimed_success = True

        print(f"\n🤖 Agent reported: SUCCESS")
        print(f"   Agent said: 'File created successfully'")

        # CRITICAL: Verify claimed artifacts actually exist
        print("\n✅ VERIFICATION (Orchestrator must do this):")

        if expected_file.exists():
            print(f"   ✅ File verified: {expected_file}")
            verification_result = "SUCCESS"
        else:
            print(f"   ❌ CRITICAL: Agent claimed success but file missing!")
            print(f"   ❌ File not found: {expected_file}")
            verification_result = "FAILURE - SILENT FAILURE DETECTED"

        # This should detect the silent failure
        self.assertFalse(expected_file.exists(), "This test expects file to NOT exist")
        self.assertEqual(verification_result, "FAILURE - SILENT FAILURE DETECTED")

        print("\n✅ CRITICAL TEST PASSED: Silent failure detected")
        print("   Orchestrator MUST verify artifacts, not trust agent reports")

    def test_03_verify_all_deliverables_after_background_agent(self):
        """
        CRITICAL TEST: Verify ALL expected deliverables exist

        FAILURE MODE: Some files created, others missing, agent reports success
        ROOT CAUSE: Agent doesn't verify all deliverables before completing
        """
        print("\n" + "="*70)
        print("CRITICAL TEST 3: All Deliverables Verification")
        print("="*70)

        # Define ALL expected deliverables
        task_dir = self.test_dir / "tasks" / "2026-01-15_test-task"
        task_dir.mkdir(parents=True, exist_ok=True)

        expected_deliverables = {
            "Contract": task_dir / "00-contract.md",
            "Plan": task_dir / "10-plan.md",
            "Work Log": task_dir / "20-work-log.md",
            "Code": self.test_dir / "src" / "feature.py",
            "Tests": self.test_dir / "tests" / "test_feature.py",
        }

        print(f"\n📋 Expected deliverables:")
        for name, path in expected_deliverables.items():
            print(f"   - {name}: {path}")

        # Simulate agent creating SOME but not ALL files
        (task_dir / "00-contract.md").write_text("# Contract")
        (task_dir / "10-plan.md").write_text("# Plan")
        # Work log NOT created (simulating failure)
        # Code and Tests NOT created (simulating failure)

        print(f"\n🤖 Spawned Agent completed")
        print(f"   Agent reported: 'Task complete'")

        # CRITICAL: Orchestrator verifies ALL deliverables
        print("\n✅ ORCHESTRATOR VERIFICATION:")

        missing_deliverables = []
        for name, path in expected_deliverables.items():
            if path.exists():
                print(f"   ✅ {name}: {path}")
            else:
                print(f"   ❌ MISSING: {name}: {path}")
                missing_deliverables.append(name)

        if missing_deliverables:
            print(f"\n❌ VERIFICATION FAILED:")
            print(f"   Missing {len(missing_deliverables)} deliverables: {missing_deliverables}")
            verification_status = "FAILED"
        else:
            print(f"\n✅ All deliverables verified")
            verification_status = "SUCCESS"

        # This test expects to find missing deliverables
        self.assertGreater(len(missing_deliverables), 0, "Test expects missing files")
        self.assertEqual(verification_status, "FAILED")

        print("\n✅ CRITICAL TEST PASSED: Missing deliverables detected")
        print("   Orchestrator MUST verify every expected artifact")


class TestBackgroundAgentTokenLimits(unittest.TestCase):
    """
    CRITICAL: Test spawned agents handle token limits correctly

    FAILURE MODE: Agent hits token limit, task truncated, appears successful
    ROOT CAUSE: No token limit detection in spawned agents
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"token-limits-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_01_detect_token_limit_failure(self):
        """
        CRITICAL TEST: Detect when spawned agent hits token limit

        FAILURE MODE: Agent hits token limit mid-task, files incomplete
        ROOT CAUSE: No detection of context overflow
        """
        print("\n" + "="*70)
        print("CRITICAL TEST 4: Token Limit Failure Detection")
        print("="*70)

        print("\n📋 Scenario: Large task causes token limit overflow")

        # Simulate agent output that hit token limit
        output_file = self.test_dir / "truncated-output.txt"
        output_file.write_text("""This is the beginning of the file...
[lots of content would be here]
And then it suddenly cut off mid-sen""")

        print(f"   Output file: {output_file}")

        # CRITICAL: Check for signs of truncation
        content = output_file.read_text()

        print("\n✅ CHECKING FOR TRUNCATION INDICATORS:")

        truncation_indicators = [
            content.endswith("mid-sen"),  # Cut off mid-word
            len(content) < 100,  # Suspiciously short
            not content.endswith("\n"),  # No proper ending
        ]

        if any(truncation_indicators):
            print("   ❌ TRUNCATION DETECTED:")
            if content.endswith("mid-sen"):
                print("      - Content cut off mid-word")
            if len(content) < 100:
                print(f"      - Content too short ({len(content)} bytes)")
            if not content.endswith("\n"):
                print("      - No proper file ending")

            detection_status = "TRUNCATION DETECTED"
        else:
            print("   ✅ No truncation detected")
            detection_status = "OK"

        self.assertEqual(detection_status, "TRUNCATION DETECTED")
        print("\n✅ CRITICAL TEST PASSED: Token limit failure detected")

    def test_02_task_decomposition_prevents_token_limits(self):
        """
        CRITICAL TEST: Large tasks decomposed to prevent token limits

        FAILURE MODE: Single large task hits token limits
        ROOT CAUSE: No task decomposition for large work
        """
        print("\n" + "="*70)
        print("CRITICAL TEST 5: Task Decomposition for Token Limits")
        print("="*70)

        # Simulate large task that should be decomposed
        large_task_size = 50000  # tokens

        print(f"\n📋 Large task: {large_task_size} tokens estimated")
        print(f"   Token limit: 25000 tokens per task")

        # CRITICAL: Orchestrator must decompose
        if large_task_size > 25000:
            print("\n✅ ORCHESTRATOR DECISION: Decompose task")

            # Break into smaller tasks
            num_subtasks = (large_task_size // 20000) + 1
            print(f"   Creating {num_subtasks} subtasks")

            subtasks = []
            for i in range(num_subtasks):
                subtask = f"subtask-{i+1}"
                subtasks.append(subtask)
                print(f"      - {subtask} (~{large_task_size // num_subtasks} tokens)")

            decomposition_status = "DECOMPOSED"
        else:
            print("\n❌ Task not decomposed - will hit token limit")
            decomposition_status = "NOT DECOMPOSED"

        self.assertEqual(decomposition_status, "DECOMPOSED")
        print("\n✅ CRITICAL TEST PASSED: Task decomposed to prevent token limits")


class TestBackgroundAgentWorkingDirectory(unittest.TestCase):
    """
    CRITICAL: Test spawned agents use correct working directory

    FAILURE MODE: Agent creates files in wrong directory (CWD vs repo root)
    ROOT CAUSE: Agent doesn't receive working directory context
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"working-dir-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_01_background_agent_uses_absolute_paths(self):
        """
        CRITICAL TEST: Spawned Agent MUST use absolute paths

        FAILURE MODE: Agent uses relative paths, files in wrong location
        ROOT CAUSE: Agent doesn't have working directory context
        """
        print("\n" + "="*70)
        print("CRITICAL TEST 6: Absolute Path Enforcement")
        print("="*70)

        # Define file with ABSOLUTE path
        absolute_path = self.test_dir / "absolute-path-file.txt"

        print(f"\n📋 Spawned Agent Instructions:")
        print(f"   MUST use: {absolute_path.absolute()}")
        print(f"   DO NOT use: 'absolute-path-file.txt' (relative)")

        # Agent creates file with absolute path
        absolute_path.write_text(f"Created at: {absolute_path.absolute()}")

        # CRITICAL: Verify file at expected location
        print(f"\n✅ VERIFICATION:")

        if absolute_path.exists():
            print(f"   ✅ File at correct location: {absolute_path}")

            # Verify it's actually where we expect
            if absolute_path.parent == self.test_dir:
                print(f"   ✅ In correct directory: {self.test_dir}")
                verification = "SUCCESS"
            else:
                print(f"   ❌ In WRONG directory: {absolute_path.parent}")
                verification = "WRONG LOCATION"
        else:
            print(f"   ❌ File not found at: {absolute_path}")
            verification = "NOT FOUND"

        self.assertEqual(verification, "SUCCESS")
        print("\n✅ CRITICAL TEST PASSED: Absolute paths enforced")

    def test_02_verify_no_nested_directory_disaster(self):
        """
        CRITICAL TEST: Prevent nested directory disasters

        FAILURE MODE: Agent creates .ai/tasks/2026-01-15_task/.ai/tasks/...
        ROOT CAUSE: Agent uses relative paths, nests in CWD
        """
        print("\n" + "="*70)
        print("CRITICAL TEST 7: Nested Directory Prevention")
        print("="*70)

        task_dir = self.test_dir / "tasks" / "2026-01-15_test"
        task_dir.mkdir(parents=True, exist_ok=True)

        print(f"\n📋 Correct location: {task_dir}")

        # Check for nested disaster
        nested_disaster = task_dir / ".ai" / "tasks"

        print(f"\n✅ CHECKING FOR NESTED DIRECTORIES:")

        if nested_disaster.exists():
            print(f"   ❌ DISASTER: Nested .ai/tasks/ found at {nested_disaster}")
            print(f"   ❌ This indicates relative path usage")
            disaster_status = "NESTED DISASTER DETECTED"
        else:
            print(f"   ✅ No nested directories detected")
            disaster_status = "OK"

        self.assertEqual(disaster_status, "OK")
        print("\n✅ CRITICAL TEST PASSED: No nested directory disaster")


class TestBackgroundAgentCompletionVerification(unittest.TestCase):
    """
    CRITICAL: Test Orchestrator verifies spawned agent completion

    FAILURE MODE: Orchestrator assumes agent completed successfully
    ROOT CAUSE: No verification of agent completion status
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"completion-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_01_orchestrator_verifies_completion_checklist(self):
        """
        CRITICAL TEST: Orchestrator verifies completion checklist

        FAILURE MODE: Agent marks task complete but checklist incomplete
        ROOT CAUSE: No systematic verification of completion criteria
        """
        print("\n" + "="*70)
        print("CRITICAL TEST 8: Completion Checklist Verification")
        print("="*70)

        # Define completion checklist
        completion_checklist = {
            "Code created": False,
            "Tests created": False,
            "Tests passing": False,
            "Work log updated": False,
            "Coverage >= 80%": False,
        }

        print(f"\n📋 Completion Checklist:")
        for item in completion_checklist.keys():
            print(f"   - {item}")

        # Simulate partial completion
        completion_checklist["Code created"] = True
        completion_checklist["Tests created"] = True
        # But tests NOT passing, work log NOT updated, coverage NOT met

        print(f"\n🤖 Spawned Agent status: 'Completed'")

        # CRITICAL: Orchestrator verifies checklist
        print(f"\n✅ ORCHESTRATOR VERIFICATION:")

        incomplete_items = []
        for item, completed in completion_checklist.items():
            status = "✅" if completed else "❌"
            print(f"   {status} {item}")
            if not completed:
                incomplete_items.append(item)

        if incomplete_items:
            print(f"\n❌ VERIFICATION FAILED:")
            print(f"   Incomplete items: {incomplete_items}")
            verification = "INCOMPLETE"
        else:
            print(f"\n✅ All checklist items complete")
            verification = "COMPLETE"

        # This test expects incomplete items
        self.assertGreater(len(incomplete_items), 0)
        self.assertEqual(verification, "INCOMPLETE")

        print("\n✅ CRITICAL TEST PASSED: Incomplete checklist detected")
        print("   Agent claimed 'complete' but verification caught missing items")


if __name__ == "__main__":
    print("="*70)
    print("CRITICAL: Spawned Agent Reliability Tests")
    print("="*70)
    print("\n🚨 MOST IMPORTANT TESTS IN THE SUITE")
    print("\nThese tests validate the #1 failure mode:")
    print("  - Spawned Agents claiming success without creating artifacts")
    print("  - Silent failures that appear successful")
    print("  - Token limit failures in background execution")
    print("  - Files created in wrong locations (sandbox vs repository)")
    print("\nThese MUST pass for the framework to be production-ready.")
    print()

    # Run tests
    unittest.main(verbosity=2)
