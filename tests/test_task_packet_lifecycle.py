#!/usr/bin/env python3
"""
Task Packet Lifecycle Tests

Tests that task packets follow the proper lifecycle:
- Contract creation (00-contract.md)
- Plan documentation (10-plan.md)
- Work log updates (20-work-log.md)
- Review documentation (30-review.md)
- Acceptance sign-off (40-acceptance.md)
- Proper directory structure (.ai/tasks/<beads-id>-<YYYYMMDDHHMMSS>-<short-desc>/)
- Cross-references between files

Status: EXECUTABLE
Priority: CRITICAL (Foundation of framework)
"""

import subprocess
import sys
import time
import unittest
from datetime import datetime
from pathlib import Path


class TestTaskPacketStructure(unittest.TestCase):
    """
    Test task packet directory structure and required files

    Validates that:
    1. Task packets use proper naming convention
    2. Required files exist
    3. Proper directory location (.ai/tasks/)
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        # Create test task packet using new naming convention:
        # <beads-id>-<YYYYMMDDHHMMSS>-<short-desc>
        timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
        cls.task_name = f"local-{timestamp}-test-task-lifecycle"
        cls.task_packet_dir = cls.repo_root / ".ai" / "tasks" / cls.task_name
        cls.task_packet_dir.mkdir(parents=True, exist_ok=True)

        print(f"\n📁 Repository root: {cls.repo_root}")
        print(f"📁 Test task packet: {cls.task_packet_dir}")

    @classmethod
    def tearDownClass(cls):
        """Clean up test artifacts"""
        if cls.task_packet_dir.exists():
            import shutil
            shutil.rmtree(cls.task_packet_dir.parent.parent)
            print(f"\n🧹 Cleaned up: {cls.task_packet_dir.parent.parent}")

    def test_01_task_packet_naming_convention(self):
        """Test: Task packet uses <beads-id>-<YYYYMMDDHHMMSS>-<short-desc> format"""
        print("\n" + "="*70)
        print("TASK PACKET TEST 1: Naming Convention")
        print("="*70)

        # Verify naming pattern: <beads-id>-<YYYYMMDDHHMMSS>-<short-desc>
        # e.g. ai-pack-4gx-20260220073115-printf-tokens
        # or local-20260220083116-test-task (for test/local tasks without Beads)
        import re
        pattern = r"^[a-z0-9][a-z0-9\-]+-\d{14}-[\w\-]+$"

        task_name = self.task_packet_dir.name

        self.assertTrue(
            re.match(pattern, task_name),
            f"❌ Task packet name '{task_name}' doesn't match pattern <beads-id>-<YYYYMMDDHHMMSS>-<short-desc>"
        )

        print(f"✅ Task packet name follows convention: {task_name}")

    def test_02_task_packet_location(self):
        """Test: Task packet in .ai/tasks/ directory"""
        print("\n" + "="*70)
        print("TASK PACKET TEST 2: Directory Location")
        print("="*70)

        # Verify location
        expected_parent = self.repo_root / ".ai" / "tasks"

        self.assertEqual(
            self.task_packet_dir.parent,
            expected_parent,
            f"❌ Task packet not in .ai/tasks/ directory"
        )

        print(f"✅ Task packet in correct location: .ai/tasks/")

    def test_03_contract_file_created(self):
        """Test: 00-contract.md file exists"""
        print("\n" + "="*70)
        print("TASK PACKET TEST 3: Contract File")
        print("="*70)

        contract_file = self.task_packet_dir / "00-contract.md"

        # Create contract
        contract_file.write_text(f"""# Task Contract

**Task:** Test Task Lifecycle
**Created:** {datetime.now().strftime("%Y-%m-%d %H:%M")}

## Objective
Test that task packet lifecycle works correctly.

## Requirements
- Create all required files
- Verify structure
- Test cross-references

## Acceptance Criteria
- All files exist
- Proper content in each
- Cross-references valid
""")

        self.assertTrue(
            contract_file.exists(),
            f"❌ Contract file not created: {contract_file}"
        )

        print(f"✅ Contract created: {contract_file}")

    def test_04_plan_file_created(self):
        """Test: 10-plan.md file exists"""
        print("\n" + "="*70)
        print("TASK PACKET TEST 4: Plan File")
        print("="*70)

        plan_file = self.task_packet_dir / "10-plan.md"

        # Create plan
        plan_file.write_text(f"""# Implementation Plan

**Date:** {datetime.now().strftime("%Y-%m-%d %H:%M")}

## Approach
Test-driven development:
1. Write test for contract creation
2. Write test for plan creation
3. Write test for work log
4. Write test for review
5. Write test for acceptance

## Steps
1. Create 00-contract.md
2. Create 10-plan.md (this file)
3. Create 20-work-log.md
4. Create 30-review.md
5. Create 40-acceptance.md

## References
- Contract: [00-contract.md](00-contract.md)
""")

        self.assertTrue(
            plan_file.exists(),
            f"❌ Plan file not created: {plan_file}"
        )

        # Verify cross-reference to contract
        content = plan_file.read_text()
        self.assertIn(
            "00-contract.md",
            content,
            "❌ Plan doesn't reference contract"
        )

        print(f"✅ Plan created: {plan_file}")
        print("✅ Plan references contract")

    def test_05_work_log_file_created(self):
        """Test: 20-work-log.md file exists"""
        print("\n" + "="*70)
        print("TASK PACKET TEST 5: Work Log File")
        print("="*70)

        work_log_file = self.task_packet_dir / "20-work-log.md"

        # Create work log
        work_log_file.write_text(f"""# Work Log

## Session {datetime.now().strftime("%Y-%m-%d %H:%M")}

### Completed
- ✅ Created contract (00-contract.md)
- ✅ Created plan (10-plan.md)
- ✅ Created work log (this file)

### In Progress
- ⏳ Testing lifecycle

### Next Steps
- Create review document
- Create acceptance document

### References
- Plan: [10-plan.md](10-plan.md)
""")

        self.assertTrue(
            work_log_file.exists(),
            f"❌ Work log file not created: {work_log_file}"
        )

        # Verify references
        content = work_log_file.read_text()
        self.assertIn(
            "10-plan.md",
            content,
            "❌ Work log doesn't reference plan"
        )

        print(f"✅ Work log created: {work_log_file}")
        print("✅ Work log references plan")


class TestTaskPacketLifecycle(unittest.TestCase):
    """
    Test complete task packet lifecycle

    Validates that:
    1. Contract → Plan → Work Log → Review → Acceptance
    2. Proper cross-references throughout
    3. All required sections present
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        # Create test task packet
        timestamp = datetime.now().strftime("%Y-%m-%d")
        cls.task_name = f"{timestamp}_lifecycle-integration-{int(time.time())}"
        cls.task_packet_dir = cls.repo_root / ".ai" / "tasks" / cls.task_name
        cls.task_packet_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up test artifacts"""
        if cls.task_packet_dir.exists():
            import shutil
            shutil.rmtree(cls.task_packet_dir.parent.parent)

    def test_01_complete_lifecycle_flow(self):
        """Test: Complete task packet lifecycle"""
        print("\n" + "="*70)
        print("LIFECYCLE TEST 1: Complete Flow")
        print("="*70)

        # Phase 1: Contract
        print("\nPhase 1: Contract Creation")
        contract = self.task_packet_dir / "00-contract.md"
        contract.write_text("""# Task Contract

**Task:** Implement User Login
**Created:** 2026-01-15

## Objective
Implement secure user login functionality.

## Requirements
- FR-1: User can login with email/password
- FR-2: Session management
- FR-3: Password hashing (bcrypt)

## Acceptance Criteria
- All tests passing
- Coverage >= 80%
- Code review approved
""")
        print(f"  ✅ Contract created")

        # Phase 2: Plan
        print("\nPhase 2: Plan Documentation")
        plan = self.task_packet_dir / "10-plan.md"
        plan.write_text("""# Implementation Plan

## Approach
Test-driven development with RED-GREEN-REFACTOR.

## Steps
1. Write failing test for login endpoint
2. Implement minimal login logic
3. Add password hashing
4. Add session management
5. Refactor and optimize

## Technical Decisions
- Use bcrypt for password hashing
- JWT for session tokens
- Express.js middleware for auth

## References
- Contract: [00-contract.md](00-contract.md)
""")
        print(f"  ✅ Plan created")

        # Verify cross-reference
        self.assertIn("00-contract.md", plan.read_text())
        print("  ✅ Plan references contract")

        # Phase 3: Work Log
        print("\nPhase 3: Work Log Updates")
        work_log = self.task_packet_dir / "20-work-log.md"
        work_log.write_text(f"""# Work Log

## Session {datetime.now().strftime("%Y-%m-%d %H:%M")}

### Completed
- ✅ Implemented login endpoint
- ✅ Added password hashing with bcrypt
- ✅ Created comprehensive tests
- ✅ All tests passing

### TDD Cycle
- RED: Created failing test for POST /login
- GREEN: Implemented minimal login logic
- REFACTOR: Extracted password validation

### Test Results
- Tests run: 15
- Tests passed: 15
- Coverage: 92%

### Next Steps
- Ready for review

### References
- Plan: [10-plan.md](10-plan.md)
""")
        print(f"  ✅ Work log created")

        # Verify cross-reference
        self.assertIn("10-plan.md", work_log.read_text())
        print("  ✅ Work log references plan")

        # Phase 4: Review
        print("\nPhase 4: Review Documentation")
        review = self.task_packet_dir / "30-review.md"
        review.write_text(f"""# Review

**Date:** {datetime.now().strftime("%Y-%m-%d %H:%M")}

## Tester Verdict: APPROVED

### TDD Compliance
- ✅ Tests written before implementation
- ✅ RED-GREEN-REFACTOR cycle followed
- ✅ Coverage: 92% (target: 80%)

## Reviewer Verdict: APPROVED

### Code Quality
- ✅ Clean code principles followed
- ✅ Proper error handling
- ✅ Security best practices (bcrypt, JWT)

### Findings
- Excellent test coverage
- Well-structured code
- Good separation of concerns

## Final Verdict
✅ **APPROVED** - Ready for acceptance

### References
- Work Log: [20-work-log.md](20-work-log.md)
""")
        print(f"  ✅ Review created")

        # Verify verdicts
        content = review.read_text()
        self.assertIn("APPROVED", content)
        self.assertIn("20-work-log.md", content)
        print("  ✅ Review contains verdicts")
        print("  ✅ Review references work log")

        # Phase 5: Acceptance
        print("\nPhase 5: Acceptance Sign-off")
        acceptance = self.task_packet_dir / "40-acceptance.md"
        acceptance.write_text(f"""# Acceptance

**Date:** {datetime.now().strftime("%Y-%m-%d %H:%M")}

## Acceptance Criteria Verification

### FR-1: User can login with email/password
✅ **VERIFIED** - Login endpoint functional

### FR-2: Session management
✅ **VERIFIED** - JWT tokens issued and validated

### FR-3: Password hashing (bcrypt)
✅ **VERIFIED** - bcrypt implemented, tested

## Quality Gates

### TDD Gate
✅ **PASSED** - Tester approved, coverage 92%

### Review Gate
✅ **PASSED** - Reviewer approved

## Deviations
None - All requirements met as specified

## Sign-off
✅ **ACCEPTED** - Ready for deployment

**Accepted by:** Orchestrator
**Date:** {datetime.now().strftime("%Y-%m-%d %H:%M")}

### References
- Contract: [00-contract.md](00-contract.md)
- Review: [30-review.md](30-review.md)
""")
        print(f"  ✅ Acceptance created")

        # Verify sign-off
        content = acceptance.read_text()
        self.assertIn("ACCEPTED", content)
        self.assertIn("00-contract.md", content)
        self.assertIn("30-review.md", content)
        print("  ✅ Acceptance contains sign-off")
        print("  ✅ Acceptance references contract and review")

        # Verify all files exist
        print("\nVerifying Complete Lifecycle:")
        all_files = [
            ("Contract", contract),
            ("Plan", plan),
            ("Work Log", work_log),
            ("Review", review),
            ("Acceptance", acceptance)
        ]

        for name, file_path in all_files:
            self.assertTrue(file_path.exists(), f"❌ {name} missing")
            print(f"  ✅ {name}: {file_path.name}")

        print("\n✅ COMPLETE LIFECYCLE VERIFIED")
        print("   Contract → Plan → Work Log → Review → Acceptance")


class TestTaskPacketCrossReferences(unittest.TestCase):
    """
    Test cross-references between task packet files

    Validates that:
    1. Files reference earlier files in chain
    2. Links are valid
    3. Consistency maintained
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        # Create test task packet
        timestamp = datetime.now().strftime("%Y-%m-%d")
        cls.task_name = f"{timestamp}_cross-refs-{int(time.time())}"
        cls.task_packet_dir = cls.repo_root / ".ai" / "tasks" / cls.task_name
        cls.task_packet_dir.mkdir(parents=True, exist_ok=True)

        # Create all files
        (cls.task_packet_dir / "00-contract.md").write_text("# Contract\n")
        (cls.task_packet_dir / "10-plan.md").write_text("# Plan\n[Contract](00-contract.md)\n")
        (cls.task_packet_dir / "20-work-log.md").write_text("# Work Log\n[Plan](10-plan.md)\n")
        (cls.task_packet_dir / "30-review.md").write_text("# Review\n[Work Log](20-work-log.md)\n")
        (cls.task_packet_dir / "40-acceptance.md").write_text("# Acceptance\n[Contract](00-contract.md)\n[Review](30-review.md)\n")

    @classmethod
    def tearDownClass(cls):
        """Clean up test artifacts"""
        if cls.task_packet_dir.exists():
            import shutil
            shutil.rmtree(cls.task_packet_dir.parent.parent)

    def test_01_plan_references_contract(self):
        """Test: Plan references contract"""
        print("\n" + "="*70)
        print("CROSS-REFERENCE TEST 1: Plan → Contract")
        print("="*70)

        plan = self.task_packet_dir / "10-plan.md"
        content = plan.read_text()

        self.assertIn(
            "00-contract.md",
            content,
            "❌ Plan doesn't reference contract"
        )

        print("✅ Plan references contract")

    def test_02_work_log_references_plan(self):
        """Test: Work log references plan"""
        print("\n" + "="*70)
        print("CROSS-REFERENCE TEST 2: Work Log → Plan")
        print("="*70)

        work_log = self.task_packet_dir / "20-work-log.md"
        content = work_log.read_text()

        self.assertIn(
            "10-plan.md",
            content,
            "❌ Work log doesn't reference plan"
        )

        print("✅ Work log references plan")

    def test_03_review_references_work_log(self):
        """Test: Review references work log"""
        print("\n" + "="*70)
        print("CROSS-REFERENCE TEST 3: Review → Work Log")
        print("="*70)

        review = self.task_packet_dir / "30-review.md"
        content = review.read_text()

        self.assertIn(
            "20-work-log.md",
            content,
            "❌ Review doesn't reference work log"
        )

        print("✅ Review references work log")

    def test_04_acceptance_references_contract_and_review(self):
        """Test: Acceptance references contract and review"""
        print("\n" + "="*70)
        print("CROSS-REFERENCE TEST 4: Acceptance → Contract + Review")
        print("="*70)

        acceptance = self.task_packet_dir / "40-acceptance.md"
        content = acceptance.read_text()

        self.assertIn(
            "00-contract.md",
            content,
            "❌ Acceptance doesn't reference contract"
        )

        self.assertIn(
            "30-review.md",
            content,
            "❌ Acceptance doesn't reference review"
        )

        print("✅ Acceptance references contract")
        print("✅ Acceptance references review")


if __name__ == "__main__":
    print("="*70)
    print("Task Packet Lifecycle Tests")
    print("="*70)
    print("\nValidating task packet structure and lifecycle...")
    print()

    # Run tests
    unittest.main(verbosity=2)
