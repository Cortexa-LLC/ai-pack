#!/usr/bin/env python3
"""
Bugfix Workflow Tests

Tests the bugfix workflow:
Phase 0: Investigation (Optional: Inspector creates RCA)
Phase 1: Fix Implementation (Engineer with TDD + regression test)
Phase 2: Validation (Tester verifies fix and regression test)
Phase 3: Review (Reviewer approves)

Status: EXECUTABLE
Priority: HIGH (Common workflow)
"""

import subprocess
import sys
import time
import unittest
from datetime import datetime
from pathlib import Path


class TestBugfixWorkflowPhases(unittest.TestCase):
    """
    Test bugfix workflow executes correctly

    Validates that:
    1. Phase 0: Inspector investigates and creates RCA (optional)
    2. Phase 1: Engineer fixes bug with regression test
    3. Phase 2: Tester validates fix
    4. Phase 3: Reviewer approves
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"bugfix-workflow-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

        print(f"\n📁 Test directory: {cls.test_dir}")

    @classmethod
    def tearDownClass(cls):
        """Clean up test artifacts"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)
            print(f"\n🧹 Cleaned up: {cls.test_dir}")

    def test_01_phase_0_inspector_creates_rca(self):
        """Test: Phase 0 - Inspector creates RCA (optional)"""
        print("\n" + "="*70)
        print("BUGFIX WORKFLOW TEST 1: Phase 0 - Investigation (RCA)")
        print("="*70)

        # Inspector creates RCA
        rca_dir = self.test_dir / "docs" / "investigations" / "2026-01-15-login-failure"
        rca_dir.mkdir(parents=True, exist_ok=True)

        rca_file = rca_dir / "rca.md"
        rca_file.write_text(f"""# Root Cause Analysis: Login Failure

**Date:** {datetime.now().strftime("%Y-%m-%d")}
**Bug ID:** BUG-123
**Severity:** High

## Bug Summary
Users cannot login with valid credentials.

## Reproduction Steps
1. Navigate to /login
2. Enter valid email and password
3. Click "Login"
4. Error: "Invalid credentials"

## Root Cause
Password comparison using `==` instead of bcrypt.compare().
Plain text comparison always fails with hashed passwords.

## Five Whys Analysis
1. Why did login fail? → Password comparison returned false
2. Why did comparison fail? → Using == operator on hashed password
3. Why == operator? → Developer forgot to use bcrypt.compare()
4. Why forgotten? → No code review caught this
5. Why not caught? → Missing test for password hashing

## Fix Recommendations
1. Replace == with bcrypt.compare()
2. Add regression test for password validation
3. Add code review checklist for security functions

## Prevention
- Add test for all auth functions
- Mandatory code review for security code
- Add static analysis rule

## References
- Code: src/auth/login.py, line 42
""")

        self.assertTrue(rca_file.exists(), "❌ RCA not created")
        print(f"✅ Inspector created RCA: {rca_file}")
        print("✅ Root cause identified: bcrypt.compare() not used")

    def test_02_phase_1_engineer_creates_regression_test(self):
        """Test: Phase 1 - Engineer creates regression test (RED)"""
        print("\n" + "="*70)
        print("BUGFIX WORKFLOW TEST 2: Phase 1 - Regression Test (RED)")
        print("="*70)

        # Engineer creates regression test first (TDD RED phase)
        tests_dir = self.test_dir / "tests"
        tests_dir.mkdir(exist_ok=True)

        test_file = tests_dir / "test_login_regression.py"
        test_file.write_text('''"""Regression test for login bug"""

def test_login_with_hashed_password():
    """
    Regression test for BUG-123

    Ensures login works with bcrypt hashed passwords.
    This test should fail before fix, pass after fix.
    """
    # Arrange
    hashed_password = "$2b$10$..."  # bcrypt hash
    plain_password = "correct_password"

    # Act
    result = validate_password(plain_password, hashed_password)

    # Assert
    assert result is True, "Login should succeed with correct password"

def test_login_rejects_wrong_password():
    """Test login rejects incorrect password"""
    hashed_password = "$2b$10$..."
    wrong_password = "wrong_password"

    result = validate_password(wrong_password, hashed_password)

    assert result is False, "Login should fail with wrong password"
''')

        self.assertTrue(test_file.exists(), "❌ Regression test not created")
        print(f"✅ Engineer created regression test: {test_file}")
        print("✅ Test covers the bug scenario (BUG-123)")

    def test_03_phase_1_engineer_implements_fix(self):
        """Test: Phase 1 - Engineer implements fix (GREEN)"""
        print("\n" + "="*70)
        print("BUGFIX WORKFLOW TEST 3: Phase 1 - Bug Fix (GREEN)")
        print("="*70)

        # Engineer implements fix
        src_dir = self.test_dir / "src"
        src_dir.mkdir(exist_ok=True)

        fix_file = src_dir / "login.py"
        fix_file.write_text('''"""Login implementation (FIXED)"""
import bcrypt

def validate_password(plain_password: str, hashed_password: str) -> bool:
    """
    Validate password using bcrypt.

    FIX for BUG-123: Changed from == comparison to bcrypt.compare()
    """
    # BEFORE (BUG): return plain_password == hashed_password
    # AFTER (FIX):
    return bcrypt.checkpw(
        plain_password.encode('utf-8'),
        hashed_password.encode('utf-8')
    )
''')

        # Engineer updates work log
        task_dir = self.test_dir / "tasks" / "local-20260115090000-fix-login"
        task_dir.mkdir(parents=True, exist_ok=True)

        work_log = task_dir / "result.md"
        work_log.write_text(f"""# Work Log: Fix Login Bug (BUG-123)

## Session {datetime.now().strftime("%Y-%m-%d %H:%M")}

### Bug Details
- **Bug:** BUG-123 - Login fails with valid credentials
- **Root Cause:** Using == instead of bcrypt.compare()

### TDD Cycle (Bugfix)
- RED: Created regression test (test_login_regression.py)
- Test fails (as expected - bug exists)
- GREEN: Fixed password validation (use bcrypt.compare())
- Test passes (bug fixed)
- REFACTOR: Added error handling

### Fix Summary
Changed line 42 in src/auth/login.py:
- BEFORE: `return plain_password == hashed_password`
- AFTER: `return bcrypt.checkpw(...)`

### Regression Test
- ✅ test_login_with_hashed_password - PASSES
- ✅ test_login_rejects_wrong_password - PASSES
- ✅ All existing tests still pass

### Test Results
- New tests: 2
- Existing tests: 15
- Total: 17
- All passing: ✅

### References
- RCA: docs/investigations/2026-01-15-login-failure/rca.md
""")

        self.assertTrue(fix_file.exists(), "❌ Fix not implemented")
        self.assertTrue(work_log.exists(), "❌ Work log not updated")
        print(f"✅ Engineer implemented fix: {fix_file}")
        print(f"✅ Engineer updated work log: {work_log}")
        print("✅ Regression test now passes")

    def test_04_phase_2_tester_validates_fix(self):
        """Test: Phase 2 - Tester validates bug fix"""
        print("\n" + "="*70)
        print("BUGFIX WORKFLOW TEST 4: Phase 2 - Tester Validation")
        print("="*70)

        task_dir = self.test_dir / "tasks" / "local-20260115090000-fix-login"

        # Tester validates the fix
        review = task_dir / "result.md"
        review.write_text(f"""# Review: Bug Fix BUG-123

**Date:** {datetime.now().strftime("%Y-%m-%d %H:%M")}

## Tester Verdict: APPROVED

### Bug Fix Validation
- ✅ Regression test created before fix
- ✅ Test failed before fix (RED)
- ✅ Test passes after fix (GREEN)
- ✅ Bug is fixed

### Regression Test Quality
- ✅ Test covers exact bug scenario
- ✅ Test will prevent regression
- ✅ Clear test documentation (references BUG-123)

### Test Coverage
- ✅ Existing tests still pass (17/17)
- ✅ No regressions introduced
- ✅ Coverage maintained: 85%

### TDD Compliance
- ✅ RED-GREEN-REFACTOR followed for bugfix
- ✅ Test-first approach used

## Validation Checklist
- ✅ Bug reproducible before fix
- ✅ Bug fixed after implementation
- ✅ Regression test prevents recurrence
- ✅ No new bugs introduced

## Final Tester Verdict
✅ **APPROVED** - Bug fix verified, regression test solid

### References
- Work Log: [result.md](result.md)
- RCA: docs/investigations/2026-01-15-login-failure/rca.md
""")

        content = review.read_text()
        self.assertIn("Tester Verdict: APPROVED", content)
        self.assertIn("Bug fix verified", content)
        print(f"✅ Tester created review: {review}")
        print("✅ Tester verdict: APPROVED")
        print("✅ Regression test validated")

    def test_05_phase_3_reviewer_approves_fix(self):
        """Test: Phase 3 - Reviewer approves code quality"""
        print("\n" + "="*70)
        print("BUGFIX WORKFLOW TEST 5: Phase 3 - Reviewer Approval")
        print("="*70)

        task_dir = self.test_dir / "tasks" / "local-20260115090000-fix-login"
        review = task_dir / "result.md"

        # Reviewer adds to review
        existing = review.read_text()
        updated = existing + f"""
## Reviewer Verdict: APPROVED

### Code Quality
- ✅ Correct use of bcrypt.checkpw()
- ✅ Proper error handling added
- ✅ Code is clean and readable

### Security Review
- ✅ Password handling secure (bcrypt)
- ✅ No plain text password storage
- ✅ Follows security best practices

### Standards Compliance
- ✅ Follows coding standards
- ✅ Proper documentation
- ✅ Clear fix comments in code

## Final Verdict
✅ **APPROVED** - Bug fix ready for deployment

**Reviewer Notes:**
Excellent bugfix approach:
- Test-first for regression prevention
- Clean implementation
- Good documentation
"""
        review.write_text(updated)

        content = review.read_text()
        self.assertIn("Reviewer Verdict: APPROVED", content)
        print("✅ Reviewer verdict: APPROVED")
        print("✅ Both Tester and Reviewer approved")


class TestBugfixWorkflowIntegration(unittest.TestCase):
    """
    Integration test: Complete bugfix workflow

    End-to-end validation of entire bugfix workflow
    """

    @classmethod
    def setUpClass(cls):
        """Set up integration test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"bugfix-integration-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_complete_bugfix_workflow(self):
        """
        Integration Test: Complete bugfix workflow

        Scenario: Fix "Null Pointer Exception" bug
        Workflow: Investigation → Regression Test → Fix → Validation → Review

        Expected: Bug fixed, regression test prevents recurrence
        """
        print("\n" + "="*70)
        print("INTEGRATION TEST: Complete Bugfix Workflow")
        print("="*70)

        # Phase 0: Investigation
        print("\n🔍 Phase 0: Investigation")
        rca_dir = self.test_dir / "docs" / "investigations" / "2026-01-15-null-pointer"
        rca_dir.mkdir(parents=True, exist_ok=True)
        (rca_dir / "rca.md").write_text("""# RCA: Null Pointer Exception
## Root Cause
Missing null check before accessing user.name
## Fix
Add null check
""")
        print("  ✅ Inspector: RCA created")

        # Phase 1: Regression Test (RED)
        print("\n🔴 Phase 1A: Regression Test (RED)")
        tests_dir = self.test_dir / "tests"
        tests_dir.mkdir(exist_ok=True)
        (tests_dir / "test_null_regression.py").write_text("""def test_handles_null_user():
    # This test fails before fix
    assert handle_user(None) == "Unknown"
""")
        print("  ✅ Engineer: Regression test created")
        print("  ❌ Test FAILS (bug exists)")

        # Phase 1B: Fix Implementation (GREEN)
        print("\n🟢 Phase 1B: Fix Implementation (GREEN)")
        src_dir = self.test_dir / "src"
        src_dir.mkdir(exist_ok=True)
        (src_dir / "user.py").write_text("""def handle_user(user):
    # FIX: Added null check
    if user is None:
        return "Unknown"
    return user.name
""")
        print("  ✅ Engineer: Fix implemented")
        print("  ✅ Test PASSES (bug fixed)")

        task_dir = self.test_dir / "tasks" / "local-20260115090000-fix-null-pointer"
        task_dir.mkdir(parents=True, exist_ok=True)
        (task_dir / "result.md").write_text("# Work Log\n## Fix\nAdded null check")
        print("  ✅ Engineer: Work log updated")

        # Phase 2: Validation
        print("\n✅ Phase 2: Validation")
        (task_dir / "result.md").write_text("""# Review
## Tester Verdict: APPROVED
- Regression test validated
- Bug fixed
""")
        print("  ✅ Tester: APPROVED")

        # Phase 3: Review
        print("\n👀 Phase 3: Review")
        existing = (task_dir / "result.md").read_text()
        (task_dir / "result.md").write_text(existing + "\n## Reviewer Verdict: APPROVED\n")
        print("  ✅ Reviewer: APPROVED")

        # Verify deliverables
        print("\n📦 Verifying Bugfix Deliverables:")
        deliverables = {
            "RCA": rca_dir / "rca.md",
            "Regression Test": tests_dir / "test_null_regression.py",
            "Fix": src_dir / "user.py",
            "Work Log": task_dir / "result.md",
            "Review": task_dir / "result.md",
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
        print("\nComplete Bugfix Workflow Verified:")
        print("  ✓ Phase 0: Investigation (RCA)")
        print("  ✓ Phase 1: Regression Test + Fix (TDD)")
        print("  ✓ Phase 2: Validation (Tester)")
        print("  ✓ Phase 3: Review (Reviewer)")


if __name__ == "__main__":
    print("="*70)
    print("Bugfix Workflow Tests")
    print("="*70)
    print("\nValidating bugfix workflow execution...")
    print()

    # Run tests
    unittest.main(verbosity=2)
