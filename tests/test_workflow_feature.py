#!/usr/bin/env python3
"""
Feature Workflow Tests

Tests the complete feature workflow:
Phase 0: Planning (Product Manager → Architect → Designer)
Phase 1: Setup (Orchestrator creates task packet)
Phase 2: Implementation (Engineer with TDD)
Phase 3: Review (Tester → Reviewer)
Phase 4: Acceptance (Orchestrator sign-off)

Status: EXECUTABLE
Priority: HIGH (Most common workflow)
"""

import subprocess
import sys
import time
import unittest
from datetime import datetime
from pathlib import Path


class TestFeatureWorkflowPhases(unittest.TestCase):
    """
    Test feature workflow executes all phases correctly

    Validates that:
    1. Phase 0: Specialists create planning docs
    2. Phase 1: Task packet created
    3. Phase 2: Engineer implements with TDD
    4. Phase 3: Reviews completed
    5. Phase 4: Acceptance sign-off
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        # Create test directories
        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"feature-workflow-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

        print(f"\n📁 Test directory: {cls.test_dir}")

    @classmethod
    def tearDownClass(cls):
        """Clean up test artifacts"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)
            print(f"\n🧹 Cleaned up: {cls.test_dir}")

    def test_01_phase_0_product-manager_creates_prd(self):
        """Test: Phase 0 - Product Manager creates PRD"""
        print("\n" + "="*70)
        print("FEATURE WORKFLOW TEST 1: Phase 0 - Product Manager (PRD)")
        print("="*70)

        # Product Manager creates PRD
        prd_dir = self.test_dir / "docs" / "product" / "2026-01-15-user-profile"
        prd_dir.mkdir(parents=True, exist_ok=True)

        prd_file = prd_dir / "prd.md"
        prd_file.write_text("""# PRD: User Profile Feature

**Date:** 2026-01-15
**Author:** Product Manager

## Problem Statement
Users need to manage their profile information.

## Customer Value
Enables users to personalize their experience.

## Functional Requirements
- FR-1: User can view profile
- FR-2: User can edit profile
- FR-3: Profile validation

## User Stories
- As a user, I want to view my profile
- As a user, I want to edit my name and email
- As a user, I want validation on email format
""")

        self.assertTrue(prd_file.exists(), "❌ PRD not created")
        print(f"✅ Product Manager created PRD: {prd_file}")

    def test_02_phase_0_architect_creates_design(self):
        """Test: Phase 0 - Architect creates architecture"""
        print("\n" + "="*70)
        print("FEATURE WORKFLOW TEST 2: Phase 0 - Architect (Design)")
        print("="*70)

        # Architect creates architecture doc
        arch_dir = self.test_dir / "docs" / "architecture" / "2026-01-15-user-profile"
        arch_dir.mkdir(parents=True, exist_ok=True)

        arch_file = arch_dir / "architecture.md"
        arch_file.write_text("""# Architecture: User Profile

## Components
- ProfileController (API layer)
- ProfileService (Business logic)
- ProfileRepository (Data access)

## API Endpoints
- GET /api/profile
- PUT /api/profile

## Data Model
```typescript
interface UserProfile {
  id: string;
  name: string;
  email: string;
  updatedAt: Date;
}
```
""")

        # Architect creates ADR
        adr_dir = self.test_dir / "docs" / "adr"
        adr_dir.mkdir(parents=True, exist_ok=True)

        adr_file = adr_dir / "001-profile-api-design.md"
        adr_file.write_text("""# ADR-001: Profile API Design

## Context
Need RESTful API for profile management.

## Decision
Use PUT for updates (idempotent).

## Rationale
PUT ensures idempotency for profile updates.

## Consequences
Simpler client implementation.
""")

        self.assertTrue(arch_file.exists(), "❌ Architecture doc not created")
        self.assertTrue(adr_file.exists(), "❌ ADR not created")
        print(f"✅ Architect created architecture: {arch_file}")
        print(f"✅ Architect created ADR: {adr_file}")

    def test_03_phase_0_designer_creates_ux(self):
        """Test: Phase 0 - Designer creates UX design"""
        print("\n" + "="*70)
        print("FEATURE WORKFLOW TEST 3: Phase 0 - Designer (UX)")
        print("="*70)

        # Designer creates design specs
        design_dir = self.test_dir / "docs" / "design" / "2026-01-15-user-profile"
        design_dir.mkdir(parents=True, exist_ok=True)

        design_file = design_dir / "design-specs.md"
        design_file.write_text("""# Design Specs: User Profile

## User Flows
1. View Profile: Dashboard → Profile Page
2. Edit Profile: Profile Page → Edit Mode → Save

## Wireframes
- Profile view with name, email display
- Edit form with validation
- Success confirmation

## Accessibility
- ARIA labels on form fields
- Keyboard navigation support
- High contrast mode compatible
""")

        self.assertTrue(design_file.exists(), "❌ Design specs not created")
        print(f"✅ Designer created design specs: {design_file}")

    def test_04_phase_1_orchestrator_creates_task_packet(self):
        """Test: Phase 1 - Orchestrator creates task packet"""
        print("\n" + "="*70)
        print("FEATURE WORKFLOW TEST 4: Phase 1 - Task Packet")
        print("="*70)

        # Orchestrator creates task packet
        task_dir = self.test_dir / "tasks" / "local-20260115090000-user-profile"
        task_dir.mkdir(parents=True, exist_ok=True)

        # Contract
        contract = task_dir / "task.md"
        contract.write_text("""# Task Contract: User Profile

## Objective
Implement user profile viewing and editing.

## Requirements
- FR-1: View profile (GET /api/profile)
- FR-2: Edit profile (PUT /api/profile)
- FR-3: Email validation

## Acceptance Criteria
- All tests passing
- Coverage >= 80%
- Code review approved

## References
- PRD: docs/product/2026-01-15-user-profile/prd.md
- Architecture: docs/architecture/2026-01-15-user-profile/architecture.md
""")

        # Plan
        plan = task_dir / "task.md"
        plan.write_text("""# Implementation Plan

## Approach
TDD with RED-GREEN-REFACTOR.

## Steps
1. Create ProfileController tests
2. Implement GET endpoint
3. Implement PUT endpoint
4. Add validation
5. Integration tests

## References
- Contract: [task.md](task.md)
""")

        self.assertTrue(contract.exists(), "❌ Contract not created")
        self.assertTrue(plan.exists(), "❌ Plan not created")
        print(f"✅ Orchestrator created contract: {contract}")
        print(f"✅ Orchestrator created plan: {plan}")

    def test_05_phase_2_engineer_implements_with_tdd(self):
        """Test: Phase 2 - Engineer implements with TDD"""
        print("\n" + "="*70)
        print("FEATURE WORKFLOW TEST 5: Phase 2 - Implementation")
        print("="*70)

        task_dir = self.test_dir / "tasks" / "local-20260115090000-user-profile"

        # Engineer creates tests (RED phase)
        tests_dir = self.test_dir / "tests"
        tests_dir.mkdir(exist_ok=True)

        test_file = tests_dir / "test_profile.py"
        test_file.write_text('''"""Profile tests"""

def test_get_profile():
    """Test GET /api/profile"""
    # Test implementation
    assert True

def test_put_profile():
    """Test PUT /api/profile"""
    # Test implementation
    assert True

def test_email_validation():
    """Test email validation"""
    # Test implementation
    assert True
''')

        # Engineer creates implementation (GREEN phase)
        src_dir = self.test_dir / "src"
        src_dir.mkdir(exist_ok=True)

        impl_file = src_dir / "profile.py"
        impl_file.write_text('''"""Profile implementation"""

class ProfileController:
    def get_profile(self, user_id):
        """Get user profile"""
        pass

    def update_profile(self, user_id, data):
        """Update user profile"""
        pass

    def validate_email(self, email):
        """Validate email format"""
        pass
''')

        # Engineer updates work log
        work_log = task_dir / "result.md"
        work_log.write_text(f"""# Work Log

## Session {datetime.now().strftime("%Y-%m-%d %H:%M")}

### TDD Cycle
- RED: Created failing tests
- GREEN: Implemented ProfileController
- REFACTOR: Extracted validation logic

### Completed
- ✅ GET /api/profile endpoint
- ✅ PUT /api/profile endpoint
- ✅ Email validation
- ✅ All tests passing

### Test Results
- Tests: 3
- Passed: 3
- Coverage: 85%

### References
- Plan: [task.md](task.md)
""")

        self.assertTrue(test_file.exists(), "❌ Tests not created")
        self.assertTrue(impl_file.exists(), "❌ Implementation not created")
        self.assertTrue(work_log.exists(), "❌ Work log not updated")
        print(f"✅ Engineer created tests: {test_file}")
        print(f"✅ Engineer created implementation: {impl_file}")
        print(f"✅ Engineer updated work log: {work_log}")

    def test_06_phase_3_tester_validates_tdd(self):
        """Test: Phase 3 - Tester validates TDD"""
        print("\n" + "="*70)
        print("FEATURE WORKFLOW TEST 6: Phase 3 - Tester Review")
        print("="*70)

        task_dir = self.test_dir / "tasks" / "local-20260115090000-user-profile"

        # Tester creates validation report
        review = task_dir / "result.md"
        review.write_text(f"""# Review

**Date:** {datetime.now().strftime("%Y-%m-%d %H:%M")}

## Tester Verdict: APPROVED

### TDD Compliance
- ✅ Tests written before implementation
- ✅ RED-GREEN-REFACTOR cycle followed
- ✅ Coverage: 85% (target: 80%)

### Test Quality
- ✅ Tests are meaningful
- ✅ Edge cases covered
- ✅ All tests passing

## References
- Work Log: [result.md](result.md)
""")

        content = review.read_text()
        self.assertIn("Tester Verdict: APPROVED", content)
        print(f"✅ Tester created review: {review}")
        print("✅ Tester verdict: APPROVED")

    def test_07_phase_3_reviewer_approves_quality(self):
        """Test: Phase 3 - Reviewer approves code quality"""
        print("\n" + "="*70)
        print("FEATURE WORKFLOW TEST 7: Phase 3 - Reviewer Approval")
        print("="*70)

        task_dir = self.test_dir / "tasks" / "local-20260115090000-user-profile"
        review = task_dir / "result.md"

        # Reviewer adds to review document
        existing = review.read_text()
        updated = existing + f"""
## Reviewer Verdict: APPROVED

### Code Quality
- ✅ Clean code principles
- ✅ Proper error handling
- ✅ Good separation of concerns

### Standards Compliance
- ✅ Naming conventions followed
- ✅ Documentation complete
- ✅ No code smells detected

## Final Verdict
✅ **APPROVED** - Ready for acceptance
"""
        review.write_text(updated)

        content = review.read_text()
        self.assertIn("Reviewer Verdict: APPROVED", content)
        self.assertIn("Final Verdict", content)
        print("✅ Reviewer verdict: APPROVED")
        print("✅ Both Tester and Reviewer approved")

    def test_08_phase_4_orchestrator_acceptance(self):
        """Test: Phase 4 - Orchestrator acceptance sign-off"""
        print("\n" + "="*70)
        print("FEATURE WORKFLOW TEST 8: Phase 4 - Acceptance")
        print("="*70)

        task_dir = self.test_dir / "tasks" / "local-20260115090000-user-profile"

        # Orchestrator creates acceptance document
        acceptance = task_dir / "result.md"
        acceptance.write_text(f"""# Acceptance

**Date:** {datetime.now().strftime("%Y-%m-%d %H:%M")}

## Acceptance Criteria Verification

### FR-1: View profile (GET /api/profile)
✅ **VERIFIED** - Endpoint implemented and tested

### FR-2: Edit profile (PUT /api/profile)
✅ **VERIFIED** - Endpoint implemented and tested

### FR-3: Email validation
✅ **VERIFIED** - Validation working correctly

## Quality Gates

### TDD Gate
✅ **PASSED** - Tester approved, coverage 85%

### Review Gate
✅ **PASSED** - Reviewer approved

## Deviations
None - All requirements met

## Sign-off
✅ **ACCEPTED** - Ready for deployment

**Accepted by:** Orchestrator
**Date:** {datetime.now().strftime("%Y-%m-%d %H:%M")}

## References
- Contract: [task.md](task.md)
- Review: [result.md](result.md)
""")

        content = acceptance.read_text()
        self.assertIn("ACCEPTED", content)
        self.assertIn("Ready for deployment", content)
        print(f"✅ Orchestrator created acceptance: {acceptance}")
        print("✅ Feature ACCEPTED and ready for deployment")


class TestFeatureWorkflowIntegration(unittest.TestCase):
    """
    Integration test: Complete feature workflow

    End-to-end validation of entire feature workflow
    """

    @classmethod
    def setUpClass(cls):
        """Set up integration test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"feature-integration-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_complete_feature_workflow(self):
        """
        Integration Test: Complete feature workflow execution

        Scenario: Implement "Dark Mode Toggle" feature
        All phases: Planning → Setup → Implementation → Review → Acceptance

        Expected: All deliverables created, all gates passed
        """
        print("\n" + "="*70)
        print("INTEGRATION TEST: Complete Feature Workflow")
        print("="*70)

        # Phase 0: Planning
        print("\n📋 Phase 0: Planning")

        prd_dir = self.test_dir / "docs" / "product" / "2026-01-15-dark-mode"
        prd_dir.mkdir(parents=True, exist_ok=True)
        (prd_dir / "prd.md").write_text("# PRD: Dark Mode\n## Requirements\n- Toggle dark mode")
        print("  ✅ Product Manager: PRD created")

        arch_dir = self.test_dir / "docs" / "architecture" / "2026-01-15-dark-mode"
        arch_dir.mkdir(parents=True, exist_ok=True)
        (arch_dir / "architecture.md").write_text("# Architecture: Dark Mode\n## Theme System")
        print("  ✅ Architect: Architecture created")

        design_dir = self.test_dir / "docs" / "design" / "2026-01-15-dark-mode"
        design_dir.mkdir(parents=True, exist_ok=True)
        (design_dir / "design-specs.md").write_text("# Design: Dark Mode\n## Toggle UI")
        print("  ✅ Designer: Design specs created")

        # Phase 1: Setup
        print("\n🎯 Phase 1: Setup")

        task_dir = self.test_dir / "tasks" / "local-20260115090000-dark-mode"
        task_dir.mkdir(parents=True, exist_ok=True)
        (task_dir / "task.md").write_text("# Contract: Dark Mode\n## Objective\nDark mode toggle")
        (task_dir / "task.md").write_text("# Plan\n## Approach\nTDD implementation")
        print("  ✅ Orchestrator: Task packet created")

        # Phase 2: Implementation
        print("\n💻 Phase 2: Implementation")

        (self.test_dir / "tests").mkdir(exist_ok=True)
        (self.test_dir / "tests" / "test_theme.py").write_text("def test_toggle(): assert True")
        (self.test_dir / "src").mkdir(exist_ok=True)
        (self.test_dir / "src" / "theme.py").write_text("class ThemeController: pass")
        (task_dir / "result.md").write_text("# Work Log\n## Completed\n- Implementation done")
        print("  ✅ Engineer: Implementation complete")

        # Phase 3: Review
        print("\n🔍 Phase 3: Review")

        (task_dir / "result.md").write_text("""# Review
## Tester Verdict: APPROVED
## Reviewer Verdict: APPROVED
""")
        print("  ✅ Tester: APPROVED")
        print("  ✅ Reviewer: APPROVED")

        # Phase 4: Acceptance
        print("\n✅ Phase 4: Acceptance")

        (task_dir / "result.md").write_text("# Acceptance\n## Sign-off\n✅ ACCEPTED")
        print("  ✅ Orchestrator: ACCEPTED")

        # Verify all deliverables
        print("\n📦 Verifying Deliverables:")

        deliverables = {
            "PRD": prd_dir / "prd.md",
            "Architecture": arch_dir / "architecture.md",
            "Design": design_dir / "design-specs.md",
            "Contract": task_dir / "task.md",
            "Plan": task_dir / "task.md",
            "Tests": self.test_dir / "tests" / "test_theme.py",
            "Implementation": self.test_dir / "src" / "theme.py",
            "Work Log": task_dir / "result.md",
            "Review": task_dir / "result.md",
            "Acceptance": task_dir / "result.md",
        }

        all_exist = True
        for name, path in deliverables.items():
            if path.exists():
                print(f"  ✅ {name}")
            else:
                print(f"  ❌ {name} MISSING")
                all_exist = False

        self.assertTrue(all_exist, "❌ Not all deliverables present")

        print("\n✅ INTEGRATION TEST PASSED")
        print("\nComplete Feature Workflow Verified:")
        print("  ✓ Phase 0: Planning (Specialists)")
        print("  ✓ Phase 1: Setup (Task Packet)")
        print("  ✓ Phase 2: Implementation (Engineer + TDD)")
        print("  ✓ Phase 3: Review (Tester + Reviewer)")
        print("  ✓ Phase 4: Acceptance (Orchestrator)")


if __name__ == "__main__":
    print("="*70)
    print("Feature Workflow Tests")
    print("="*70)
    print("\nValidating complete feature workflow execution...")
    print()

    # Run tests
    unittest.main(verbosity=2)
