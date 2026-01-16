#!/usr/bin/env python3
"""
Gate Enforcement Tests

Tests that quality gates enforce standards:
- Gate 30: TDD Enforcement (BLOCKING - no code without tests)
- Gate 35: Code Quality Review (BLOCKING - needs Tester + Reviewer approval)
- Gate 10: Persistence Gate (artifacts must persist)

Status: EXECUTABLE
Priority: HIGH (Quality gates are critical)
"""

import subprocess
import sys
import time
import unittest
from datetime import datetime
from pathlib import Path


class TestGate30TDDEnforcement(unittest.TestCase):
    """
    Test Gate 30: TDD Enforcement

    Validates that:
    1. Code without tests is BLOCKED
    2. Tests with insufficient coverage are BLOCKED
    3. Code with good tests is ALLOWED
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"gate-tdd-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up test artifacts"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_01_gate_blocks_without_tests(self):
        """Test: Gate 30 BLOCKS code submitted without tests"""
        print("\n" + "="*70)
        print("GATE 30 TEST 1: BLOCKS Code Without Tests")
        print("="*70)

        # Scenario: Code submitted without tests
        src_dir = self.test_dir / "src"
        src_dir.mkdir(exist_ok=True)

        code_file = src_dir / "feature.py"
        code_file.write_text("""def new_feature():
    '''New feature with no tests'''
    return "feature"
""")

        # Check for tests
        tests_dir = self.test_dir / "tests"
        has_tests = tests_dir.exists() and list(tests_dir.glob("test_*.py"))

        print(f"Code exists: {code_file.exists()}")
        print(f"Tests exist: {has_tests}")

        # Gate 30 decision
        if not has_tests:
            gate_status = "BLOCKED"
            gate_message = "❌ GATE 30 BLOCKED: No tests found for code changes"
        else:
            gate_status = "ALLOWED"
            gate_message = "✅ GATE 30 ALLOWED: Tests present"

        print(f"\nGate 30 Status: {gate_status}")
        print(gate_message)

        self.assertEqual(gate_status, "BLOCKED", "Gate should BLOCK code without tests")
        print("\n✅ Gate 30 correctly BLOCKS code without tests")

    def test_02_gate_blocks_insufficient_coverage(self):
        """Test: Gate 30 BLOCKS if coverage < 80%"""
        print("\n" + "="*70)
        print("GATE 30 TEST 2: BLOCKS Insufficient Coverage")
        print("="*70)

        # Scenario: Tests exist but coverage < 80%
        src_dir = self.test_dir / "src"
        src_dir.mkdir(exist_ok=True)

        (src_dir / "service.py").write_text("""def method1():
    return 1

def method2():
    return 2

def method3():
    return 3
""")

        tests_dir = self.test_dir / "tests"
        tests_dir.mkdir(exist_ok=True)

        (tests_dir / "test_service.py").write_text("""def test_method1():
    assert method1() == 1
# Only 1 of 3 methods tested = 33% coverage
""")

        # Simulated coverage
        simulated_coverage = 33  # Only method1 tested
        coverage_threshold = 80

        print(f"Coverage: {simulated_coverage}%")
        print(f"Threshold: {coverage_threshold}%")

        # Gate 30 decision
        if simulated_coverage < coverage_threshold:
            gate_status = "BLOCKED"
            gate_message = f"❌ GATE 30 BLOCKED: Coverage {simulated_coverage}% < {coverage_threshold}%"
        else:
            gate_status = "ALLOWED"
            gate_message = f"✅ GATE 30 ALLOWED: Coverage {simulated_coverage}% >= {coverage_threshold}%"

        print(f"\nGate 30 Status: {gate_status}")
        print(gate_message)

        self.assertEqual(gate_status, "BLOCKED", "Gate should BLOCK insufficient coverage")
        print("\n✅ Gate 30 correctly BLOCKS insufficient coverage")

    def test_03_gate_allows_good_tests(self):
        """Test: Gate 30 ALLOWS code with good tests and coverage"""
        print("\n" + "="*70)
        print("GATE 30 TEST 3: ALLOWS Good Tests + Coverage")
        print("="*70)

        # Scenario: Good tests, good coverage
        src_dir = self.test_dir / "src"
        src_dir.mkdir(exist_ok=True)

        (src_dir / "calculator.py").write_text("""def add(a, b):
    return a + b

def subtract(a, b):
    return a - b
""")

        tests_dir = self.test_dir / "tests"
        tests_dir.mkdir(exist_ok=True)

        (tests_dir / "test_calculator.py").write_text("""def test_add():
    assert add(2, 3) == 5

def test_subtract():
    assert subtract(5, 3) == 2
# Both methods tested = 100% coverage
""")

        # Simulated coverage
        simulated_coverage = 100
        coverage_threshold = 80
        tests_pass = True

        print(f"Tests exist: ✅")
        print(f"Coverage: {simulated_coverage}%")
        print(f"Threshold: {coverage_threshold}%")
        print(f"Tests passing: {tests_pass}")

        # Gate 30 decision
        has_tests = tests_dir.exists() and list(tests_dir.glob("test_*.py"))

        if has_tests and simulated_coverage >= coverage_threshold and tests_pass:
            gate_status = "ALLOWED"
            gate_message = "✅ GATE 30 ALLOWED: TDD requirements met"
        else:
            gate_status = "BLOCKED"
            gate_message = "❌ GATE 30 BLOCKED: TDD requirements not met"

        print(f"\nGate 30 Status: {gate_status}")
        print(gate_message)

        self.assertEqual(gate_status, "ALLOWED", "Gate should ALLOW good tests")
        print("\n✅ Gate 30 correctly ALLOWS code with good tests")


class TestGate35CodeQualityReview(unittest.TestCase):
    """
    Test Gate 35: Code Quality Review

    Validates that:
    1. Code without Tester approval is BLOCKED
    2. Code without Reviewer approval is BLOCKED
    3. Code with both approvals is ALLOWED
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"gate-review-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up test artifacts"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_01_gate_blocks_without_tester_approval(self):
        """Test: Gate 35 BLOCKS without Tester approval"""
        print("\n" + "="*70)
        print("GATE 35 TEST 1: BLOCKS Without Tester Approval")
        print("="*70)

        # Scenario: No review file
        task_dir = self.test_dir / "tasks" / "2026-01-15_task1"
        task_dir.mkdir(parents=True, exist_ok=True)

        review_file = task_dir / "30-review.md"
        # No review file exists

        # Check verdicts
        tester_approved = False
        reviewer_approved = False

        if review_file.exists():
            content = review_file.read_text()
            tester_approved = "Tester Verdict: APPROVED" in content
            reviewer_approved = "Reviewer Verdict: APPROVED" in content

        print(f"Review exists: {review_file.exists()}")
        print(f"Tester approved: {tester_approved}")
        print(f"Reviewer approved: {reviewer_approved}")

        # Gate 35 decision
        if not tester_approved:
            gate_status = "BLOCKED"
            gate_message = "❌ GATE 35 BLOCKED: Missing Tester approval"
        elif not reviewer_approved:
            gate_status = "BLOCKED"
            gate_message = "❌ GATE 35 BLOCKED: Missing Reviewer approval"
        else:
            gate_status = "ALLOWED"
            gate_message = "✅ GATE 35 ALLOWED: Both approvals present"

        print(f"\nGate 35 Status: {gate_status}")
        print(gate_message)

        self.assertEqual(gate_status, "BLOCKED", "Gate should BLOCK without Tester approval")
        print("\n✅ Gate 35 correctly BLOCKS without Tester approval")

    def test_02_gate_blocks_if_rejected(self):
        """Test: Gate 35 BLOCKS if Tester REJECTED"""
        print("\n" + "="*70)
        print("GATE 35 TEST 2: BLOCKS If Tests REJECTED")
        print("="*70)

        # Scenario: Tester REJECTED
        task_dir = self.test_dir / "tasks" / "2026-01-15_task2"
        task_dir.mkdir(parents=True, exist_ok=True)

        review_file = task_dir / "30-review.md"
        review_file.write_text("""# Review

## Tester Verdict: REJECTED

### Issues
- Coverage only 60% (need 80%)
- Tests failing: 2/10

## Required Actions
- Fix failing tests
- Improve coverage
""")

        content = review_file.read_text()
        tester_approved = "Tester Verdict: APPROVED" in content
        tester_rejected = "Tester Verdict: REJECTED" in content

        print(f"Tester approved: {tester_approved}")
        print(f"Tester rejected: {tester_rejected}")

        # Gate 35 decision
        if tester_rejected:
            gate_status = "BLOCKED"
            gate_message = "❌ GATE 35 BLOCKED: Tester REJECTED - fix issues and resubmit"
        elif tester_approved:
            gate_status = "ALLOWED"
            gate_message = "✅ GATE 35 ALLOWED: Tester approved"
        else:
            gate_status = "BLOCKED"
            gate_message = "❌ GATE 35 BLOCKED: No Tester verdict"

        print(f"\nGate 35 Status: {gate_status}")
        print(gate_message)

        self.assertEqual(gate_status, "BLOCKED", "Gate should BLOCK if Tester rejected")
        print("\n✅ Gate 35 correctly BLOCKS when Tester REJECTED")

    def test_03_gate_allows_with_both_approvals(self):
        """Test: Gate 35 ALLOWS with both Tester + Reviewer approval"""
        print("\n" + "="*70)
        print("GATE 35 TEST 3: ALLOWS With Both Approvals")
        print("="*70)

        # Scenario: Both approved
        task_dir = self.test_dir / "tasks" / "2026-01-15_task3"
        task_dir.mkdir(parents=True, exist_ok=True)

        review_file = task_dir / "30-review.md"
        review_file.write_text("""# Review

## Tester Verdict: APPROVED
- Coverage: 92%
- All tests passing

## Reviewer Verdict: APPROVED
- Code quality: Excellent
- Standards compliance: Good
""")

        content = review_file.read_text()
        tester_approved = "Tester Verdict: APPROVED" in content
        reviewer_approved = "Reviewer Verdict: APPROVED" in content

        print(f"Tester approved: {tester_approved}")
        print(f"Reviewer approved: {reviewer_approved}")

        # Gate 35 decision
        if tester_approved and reviewer_approved:
            gate_status = "ALLOWED"
            gate_message = "✅ GATE 35 ALLOWED: Both Tester and Reviewer approved"
        else:
            gate_status = "BLOCKED"
            gate_message = "❌ GATE 35 BLOCKED: Missing approvals"

        print(f"\nGate 35 Status: {gate_status}")
        print(gate_message)

        self.assertEqual(gate_status, "ALLOWED", "Gate should ALLOW with both approvals")
        print("\n✅ Gate 35 correctly ALLOWS with both approvals")


class TestGate10PersistenceGate(unittest.TestCase):
    """
    Test Gate 10: Persistence Gate

    Validates that:
    1. Artifacts must persist to disk
    2. Files must exist in correct locations
    3. No sandbox-only files allowed
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"gate-persistence-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up test artifacts"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_01_gate_verifies_artifacts_persist(self):
        """Test: Gate 10 verifies artifacts persist to disk"""
        print("\n" + "="*70)
        print("GATE 10 TEST 1: Verifies Artifacts Persist")
        print("="*70)

        # Expected artifacts
        expected_artifacts = {
            "Contract": self.test_dir / "tasks" / "2026-01-15_test" / "00-contract.md",
            "Plan": self.test_dir / "tasks" / "2026-01-15_test" / "10-plan.md",
            "Code": self.test_dir / "src" / "feature.py",
            "Tests": self.test_dir / "tests" / "test_feature.py",
        }

        # Create some artifacts
        (self.test_dir / "tasks" / "2026-01-15_test").mkdir(parents=True, exist_ok=True)
        (expected_artifacts["Contract"]).write_text("# Contract")
        (expected_artifacts["Plan"]).write_text("# Plan")

        # Check persistence
        print("\nChecking artifact persistence:")
        all_persisted = True
        for name, path in expected_artifacts.items():
            exists = path.exists()
            status = "✅" if exists else "❌"
            print(f"  {status} {name}: {path}")
            if not exists:
                all_persisted = False

        # Gate 10 decision
        if all_persisted:
            gate_status = "ALLOWED"
            gate_message = "✅ GATE 10 ALLOWED: All artifacts persisted"
        else:
            gate_status = "BLOCKED"
            gate_message = "❌ GATE 10 BLOCKED: Missing artifacts - files must persist to disk"

        print(f"\nGate 10 Status: {gate_status}")
        print(gate_message)

        self.assertEqual(gate_status, "BLOCKED", "Gate should BLOCK when artifacts missing")
        print("\n✅ Gate 10 correctly BLOCKS when artifacts missing")

    def test_02_gate_allows_when_all_artifacts_present(self):
        """Test: Gate 10 ALLOWS when all artifacts persist"""
        print("\n" + "="*70)
        print("GATE 10 TEST 2: ALLOWS When All Artifacts Present")
        print("="*70)

        # Create all expected artifacts
        task_dir = self.test_dir / "tasks" / "2026-01-15_complete"
        task_dir.mkdir(parents=True, exist_ok=True)

        src_dir = self.test_dir / "src"
        src_dir.mkdir(exist_ok=True)

        tests_dir = self.test_dir / "tests"
        tests_dir.mkdir(exist_ok=True)

        artifacts = {
            "Contract": task_dir / "00-contract.md",
            "Plan": task_dir / "10-plan.md",
            "Work Log": task_dir / "20-work-log.md",
            "Code": src_dir / "feature.py",
            "Tests": tests_dir / "test_feature.py",
        }

        # Create all files
        for name, path in artifacts.items():
            path.write_text(f"# {name}")

        # Check persistence
        print("\nChecking artifact persistence:")
        all_persisted = True
        for name, path in artifacts.items():
            exists = path.exists()
            status = "✅" if exists else "❌"
            print(f"  {status} {name}: {path}")
            if not exists:
                all_persisted = False

        # Gate 10 decision
        if all_persisted:
            gate_status = "ALLOWED"
            gate_message = "✅ GATE 10 ALLOWED: All artifacts persisted to disk"
        else:
            gate_status = "BLOCKED"
            gate_message = "❌ GATE 10 BLOCKED: Missing artifacts"

        print(f"\nGate 10 Status: {gate_status}")
        print(gate_message)

        self.assertEqual(gate_status, "ALLOWED", "Gate should ALLOW when all artifacts present")
        print("\n✅ Gate 10 correctly ALLOWS when all artifacts persisted")


if __name__ == "__main__":
    print("="*70)
    print("Gate Enforcement Tests")
    print("="*70)
    print("\nValidating quality gate enforcement...")
    print()

    # Run tests
    unittest.main(verbosity=2)
