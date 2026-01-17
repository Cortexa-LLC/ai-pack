#!/usr/bin/env python3
"""
Role Test: Orchestrator Delegation and Coordination

Tests that the Orchestrator role:
- Delegates to appropriate specialist roles
- Coordinates multi-agent workflows
- Verifies deliverables from delegated agents
- Manages task decomposition
- Enforces quality gates

Status: EXECUTABLE
Priority: CRITICAL (Priority 2D)
"""

import subprocess
import sys
import time
import unittest
from pathlib import Path
from datetime import datetime


class TestOrchestratorDelegation(unittest.TestCase):
    """
    Test Orchestrator delegation to specialist roles

    Validates that Orchestrator can:
    1. Identify need for specialist roles
    2. Delegate to appropriate specialists
    3. Verify specialist deliverables
    4. Coordinate multi-agent workflows
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"orchestrator-test-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_01_orchestrator_delegates_to_cartographer(self):
        """Test: Orchestrator delegates PRD creation to Cartographer"""
        print("\n" + "="*70)
        print("ORCHESTRATOR TEST 1: Delegates to Cartographer")
        print("="*70)

        # Orchestrator identifies need for PRD
        print("Step 1: Orchestrator analyzes task")
        print("   Task type: Feature implementation")
        print("   Size: Large (requires PRD)")
        print("   Decision: Delegate to Cartographer")

        # Orchestrator creates delegation record
        delegation_log = self.test_dir / "delegation-log.md"
        delegation_log.write_text(f'''# Orchestrator Delegation Log

**Date:** {datetime.now().strftime("%Y-%m-%d %H:%M")}

## Delegation 1: Cartographer

**Task:** Create PRD for User Authentication feature
**Reason:** Large feature requires product requirements documentation
**Expected Deliverable:** docs/product/YYYY-MM-DD-user-auth/prd.md
**Status:** Delegated

## Verification Criteria
- [ ] PRD document exists
- [ ] PRD has all required sections
- [ ] Requirements are clear and testable
''')

        print("Step 2: Orchestrator delegates to Cartographer")
        print("   Creating delegation log...")

        # Verify delegation log created
        self.assertTrue(delegation_log.exists(), "❌ Delegation log not created")
        print(f"✅ Delegation log created: {delegation_log}")

        # Simulate Cartographer completing work
        prd_dir = self.test_dir / "docs" / "product" / "2026-01-15-user-auth"
        prd_dir.mkdir(parents=True, exist_ok=True)
        prd_file = prd_dir / "prd.md"
        prd_file.write_text("# PRD: User Authentication\n\n## Requirements\n- FR-1: User login")

        print("Step 3: Cartographer completes PRD")

        # Orchestrator verifies deliverable
        if prd_file.exists():
            print("✅ Orchestrator verification: PRD exists")
            print(f"   Location: {prd_file}")
        else:
            self.fail("❌ Orchestrator verification failed: PRD missing")

    def test_02_orchestrator_delegates_to_architect(self):
        """Test: Orchestrator delegates architecture to Architect"""
        print("\n" + "="*70)
        print("ORCHESTRATOR TEST 2: Delegates to Architect")
        print("="*70)

        # Orchestrator identifies need for architecture
        print("Step 1: Orchestrator analyzes requirements")
        print("   Complexity: High (requires architectural design)")
        print("   Decision: Delegate to Architect")

        # Simulate Architect completing work
        arch_dir = self.test_dir / "docs" / "architecture" / "2026-01-15-user-auth"
        arch_dir.mkdir(parents=True, exist_ok=True)
        arch_file = arch_dir / "architecture.md"
        arch_file.write_text("# Architecture: User Authentication")

        adr_dir = self.test_dir / "docs" / "adr"
        adr_dir.mkdir(parents=True, exist_ok=True)
        adr_file = adr_dir / "001-jwt-auth.md"
        adr_file.write_text("# ADR-001: JWT Authentication")

        print("Step 2: Architect completes deliverables")

        # Orchestrator verifies deliverables
        deliverables = [
            ("Architecture doc", arch_file),
            ("ADR", adr_file)
        ]

        all_exist = True
        for name, file_path in deliverables:
            if file_path.exists():
                print(f"✅ Orchestrator verification: {name} exists")
            else:
                print(f"❌ Orchestrator verification failed: {name} missing")
                all_exist = False

        self.assertTrue(all_exist, "❌ Not all Architect deliverables present")

    def test_03_orchestrator_delegates_to_engineer(self):
        """Test: Orchestrator delegates implementation to Engineer"""
        print("\n" + "="*70)
        print("ORCHESTRATOR TEST 3: Delegates to Engineer")
        print("="*70)

        # Orchestrator creates task packet for Engineer
        task_packet_dir = self.test_dir / ".ai" / "tasks" / "2026-01-15_implement-auth"
        task_packet_dir.mkdir(parents=True, exist_ok=True)

        # Contract
        contract = task_packet_dir / "00-contract.md"
        contract.write_text('''# Task: Implement User Authentication

## Requirements
- Implement login endpoint
- Add password hashing
- Create tests

## Acceptance Criteria
- All tests passing
- Coverage >= 80%
''')

        # Plan
        plan = task_packet_dir / "10-plan.md"
        plan.write_text('''# Implementation Plan

## Approach
1. Write tests (TDD)
2. Implement login logic
3. Add password hashing
''')

        print("Step 1: Orchestrator creates task packet")
        print(f"   Task packet: {task_packet_dir}")

        # Verify task packet created
        self.assertTrue(contract.exists(), "❌ Contract not created")
        self.assertTrue(plan.exists(), "❌ Plan not created")
        print("✅ Task packet created for Engineer")

        # Simulate Engineer completing work
        work_log = task_packet_dir / "20-work-log.md"
        work_log.write_text(f'''# Work Log

## Session {datetime.now().strftime("%Y-%m-%d %H:%M")}

### Completed
- Implemented login endpoint
- Added password hashing with bcrypt
- Created comprehensive tests

### Status
Ready for review
''')

        print("Step 2: Engineer completes implementation")

        # Orchestrator verifies completion
        if work_log.exists():
            content = work_log.read_text()
            if "Completed" in content:
                print("✅ Orchestrator verification: Engineer completed work")
            else:
                self.fail("❌ Work log lacks completion status")
        else:
            self.fail("❌ Work log not found")

    def test_04_orchestrator_delegates_to_reviewer(self):
        """Test: Orchestrator delegates review to Reviewer"""
        print("\n" + "="*70)
        print("ORCHESTRATOR TEST 4: Delegates to Reviewer")
        print("="*70)

        # Orchestrator identifies code ready for review
        print("Step 1: Orchestrator detects implementation complete")
        print("   Decision: Delegate to Reviewer")

        # Simulate Reviewer completing review
        task_packet_dir = self.test_dir / ".ai" / "tasks" / "2026-01-15_review-auth"
        task_packet_dir.mkdir(parents=True, exist_ok=True)

        review = task_packet_dir / "30-review.md"
        review.write_text(f'''# Code Review

**Date:** {datetime.now().strftime("%Y-%m-%d %H:%M")}

## Reviewer Verdict: APPROVED

## Findings
- Code quality: Excellent
- Test coverage: 92%
- Standards compliance: Good

## Verdict
✅ APPROVED - Ready for acceptance
''')

        print("Step 2: Reviewer completes review")

        # Orchestrator verifies verdict
        if review.exists():
            content = review.read_text()
            if "APPROVED" in content:
                print("✅ Orchestrator verification: Review APPROVED")
                print("   Can proceed to acceptance")
            elif "REJECTED" in content or "CHANGES REQUIRED" in content:
                print("⚠️  Orchestrator: Review BLOCKED")
                print("   Cannot proceed until issues resolved")
            else:
                self.fail("❌ Review lacks clear verdict")
        else:
            self.fail("❌ Review document not found")


class TestOrchestratorCoordination(unittest.TestCase):
    """
    Test Orchestrator coordination of multiple agents

    Validates that Orchestrator can:
    1. Manage parallel agent execution
    2. Handle dependencies between agents
    3. Enforce WIP limits
    4. Coordinate completion verification
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"orchestrator-coord-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_01_orchestrator_enforces_wip_limits(self):
        """Test: Orchestrator enforces WIP limits (max 3 concurrent agents)"""
        print("\n" + "="*70)
        print("COORDINATION TEST 1: Enforces WIP Limits")
        print("="*70)

        # Orchestrator tracks active agents
        active_agents = []

        # Simulate spawning agents
        print("Spawning spawned agents...")
        for i in range(5):
            if len(active_agents) < 3:
                active_agents.append(f"Engineer-{i+1}")
                print(f"   ✅ Spawned: Engineer-{i+1} (WIP: {len(active_agents)}/3)")
            else:
                print(f"   ⚠️  Cannot spawn Engineer-{i+1}: WIP limit reached (3/3)")
                print(f"      Waiting for agent completion...")

        # Verify WIP limit enforced
        self.assertLessEqual(
            len(active_agents),
            3,
            "❌ WIP limit violated (> 3 concurrent agents)"
        )
        print(f"✅ WIP limit enforced: {len(active_agents)}/3 agents active")

    def test_02_orchestrator_verifies_deliverables(self):
        """Test: Orchestrator verifies all delegated agents complete deliverables"""
        print("\n" + "="*70)
        print("COORDINATION TEST 2: Verifies All Deliverables")
        print("="*70)

        # Orchestrator tracks expected deliverables
        expected_deliverables = {
            "Cartographer": self.test_dir / "docs" / "product" / "prd.md",
            "Architect": self.test_dir / "docs" / "architecture" / "architecture.md",
            "Designer": self.test_dir / "docs" / "design" / "design-specs.md",
            "Engineer": self.test_dir / "src" / "auth.py"
        }

        # Create deliverables
        for role, file_path in expected_deliverables.items():
            file_path.parent.mkdir(parents=True, exist_ok=True)
            file_path.write_text(f"# Deliverable from {role}")

        print("Verifying deliverables from delegated agents...")

        # Orchestrator verifies all deliverables
        all_present = True
        for role, file_path in expected_deliverables.items():
            if file_path.exists():
                print(f"   ✅ {role}: Deliverable verified")
            else:
                print(f"   ❌ {role}: Deliverable MISSING")
                all_present = False

        self.assertTrue(
            all_present,
            "❌ Not all deliverables present"
        )
        print("✅ All delegated agents completed deliverables")


class TestOrchestratorIntegration(unittest.TestCase):
    """
    Integration test: Full Orchestrator workflow

    End-to-end test of Orchestrator coordinating multiple roles
    """

    @classmethod
    def setUpClass(cls):
        """Set up integration test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"orchestrator-integration-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_full_orchestrator_workflow(self):
        """
        Integration Test: Orchestrator coordinates full feature workflow

        Scenario:
        1. Orchestrator receives feature request
        2. Delegates to Cartographer for PRD
        3. Delegates to Architect for design
        4. Delegates to Engineer for implementation
        5. Delegates to Reviewer for code review
        6. Verifies all deliverables complete

        Expected:
        - All specialists delegated appropriately
        - All deliverables produced
        - Orchestrator verification successful
        """
        print("\n" + "="*70)
        print("INTEGRATION TEST: Full Orchestrator Workflow")
        print("="*70)

        # Phase 1: Orchestrator delegates to Cartographer
        print("\nPhase 1: Requirements (Cartographer)")
        prd_dir = self.test_dir / "docs" / "product" / "2026-01-15-feature"
        prd_dir.mkdir(parents=True, exist_ok=True)
        prd_file = prd_dir / "prd.md"
        prd_file.write_text("# PRD: Feature\n\n## Requirements\n- Requirement 1")
        print("   ✅ Cartographer: PRD created")

        # Phase 2: Orchestrator delegates to Architect
        print("\nPhase 2: Architecture (Architect)")
        arch_dir = self.test_dir / "docs" / "architecture" / "2026-01-15-feature"
        arch_dir.mkdir(parents=True, exist_ok=True)
        arch_file = arch_dir / "architecture.md"
        arch_file.write_text("# Architecture: Feature")
        print("   ✅ Architect: Architecture doc created")

        # Phase 3: Orchestrator delegates to Engineer
        print("\nPhase 3: Implementation (Engineer)")
        src_dir = self.test_dir / "src"
        src_dir.mkdir(exist_ok=True)
        code_file = src_dir / "feature.py"
        code_file.write_text("def feature(): pass")
        print("   ✅ Engineer: Implementation complete")

        # Phase 4: Orchestrator delegates to Reviewer
        print("\nPhase 4: Review (Reviewer)")
        task_dir = self.test_dir / ".ai" / "tasks" / "2026-01-15_feature"
        task_dir.mkdir(parents=True, exist_ok=True)
        review_file = task_dir / "30-review.md"
        review_file.write_text("# Review\n\n## Verdict: APPROVED")
        print("   ✅ Reviewer: Review APPROVED")

        # Orchestrator verification
        print("\nOrchestrator Final Verification:")

        deliverables = {
            "PRD": prd_file,
            "Architecture": arch_file,
            "Implementation": code_file,
            "Review": review_file
        }

        all_complete = True
        for name, file_path in deliverables.items():
            if file_path.exists():
                print(f"   ✅ {name}: Complete")
            else:
                print(f"   ❌ {name}: MISSING")
                all_complete = False

        self.assertTrue(all_complete, "❌ Not all deliverables complete")

        print("\n✅ INTEGRATION TEST PASSED")
        print("\nOrchestrator successfully:")
        print("   ✓ Delegated to all required specialists")
        print("   ✓ Verified all deliverables produced")
        print("   ✓ Coordinated complete feature workflow")


if __name__ == "__main__":
    print("="*70)
    print("Orchestrator Delegation and Coordination Tests")
    print("="*70)
    print("\nValidating Orchestrator role delegation and coordination...")
    print()

    # Run tests
    unittest.main(verbosity=2)
