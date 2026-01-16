# Workflow Validation Guide

**Version:** 1.0
**Date:** 2026-01-15
**Purpose:** Ensure test quality and coverage for all changes to ai-pack

---

## Overview

The **Pre-Change Validation Rule** ensures that:
1. All tests pass before any change is committed
2. Tests exist for all roles
3. Role modifications are accompanied by test updates
4. Test structure remains intact

---

## Quick Start

### Run Validation Manually

```bash
cd tests/

# Full validation (includes integration tests)
python3 pre-change-validation.py

# Quick validation (skip slow integration tests)
python3 pre-change-validation.py --quick

# Check only (report issues but don't block)
python3 pre-change-validation.py --check
```

### Expected Output

**Success:**
```
======================================================================
VALIDATION SUMMARY
======================================================================

✅ ALL VALIDATIONS PASSED

You may proceed with changes:
  - All tests passing
  - All roles have tests
  - Test structure verified

✅ Ready to commit
```

**Failure:**
```
======================================================================
VALIDATION SUMMARY
======================================================================

❌ VALIDATION FAILED

Errors (2):
  Missing tests for roles:
    - new-role.md -> test_role_new.py
  2 tests failing (not the known settings.json issue)

❌ CHANGES BLOCKED

Fix the errors above before committing changes.
```

---

## What Gets Validated

### 1. All Tests Pass

**Validation:**
- Runs `python3 run_tests.py` (or `--quick` mode)
- Verifies all tests pass
- Allows 1 known failure: `.claude/settings.json` (infrastructure issue)

**Failure Conditions:**
- Any test fails (except known settings.json issue)
- Test runner errors

**Fix:**
```bash
# Run tests to see failures
cd tests/
python3 run_tests.py -v

# Fix failing tests
# Re-run validation
```

---

### 2. Role Test Coverage

**Validation:**
- Checks that every role file has a corresponding test file
- Verifies test files exist and are correctly named

**Current Mapping:**
```
roles/engineer.md          → tests/test_role_engineer.py
roles/reviewer.md          → tests/test_role_reviewer.py
roles/tester.md            → tests/test_role_tester.py
roles/cartographer.md      → tests/test_role_specialists.py
roles/architect.md         → tests/test_role_specialists.py
roles/designer.md          → tests/test_role_specialists.py
roles/inspector.md         → tests/test_role_specialists.py
roles/orchestrator.md      → tests/test_orchestrator_delegation.py
```

**Failure Conditions:**
- New role added without tests
- Role test file deleted

**Fix (New Role Added):**
```bash
# If you added roles/new-role.md, create test file:
cd tests/

# Copy template from existing role test
cp test_role_engineer.py test_role_new.py

# Update test to validate new role
# Re-run validation
```

**Fix (Role Removed):**
```bash
# If role removed, remove corresponding test (or update mapping)
# Re-run validation
```

---

### 3. Role Modifications

**Validation:**
- Detects modified role files via `git diff`
- Warns if no test files were modified alongside role changes

**Warning Conditions:**
- Role file modified but no test modified

**Fix:**
```bash
# If you modified roles/engineer.md:
# Update tests/test_role_engineer.py to match changes

# Example: If you added new Engineer responsibility
# Add test to verify new responsibility
```

**Note:** This is a WARNING, not a blocker. Consider whether your role change requires test updates.

---

### 4. Test Structure Integrity

**Validation:**
- Verifies all required test files exist
- Checks test directory structure

**Required Files:**
```
tests/
  run_tests.py                      # Test runner
  test_role_engineer.py             # Engineer role tests
  test_role_reviewer.py             # Reviewer role tests
  test_role_tester.py               # Tester role tests
  test_role_specialists.py          # Specialist roles
  test_orchestrator_delegation.py   # Orchestrator tests
  test_beads_integration.py         # Beads integration tests
  pre-change-validation.py          # This validator
```

**Failure Conditions:**
- Required test file missing
- Test directory corrupted

**Fix:**
```bash
# Restore missing files from git
git restore tests/test_role_*.py

# Or recreate from backup
```

---

## When to Run Validation

### Required (MUST Run)

1. **Before Committing Role Changes**
   ```bash
   # Modified roles/engineer.md
   cd tests/
   python3 pre-change-validation.py --quick
   git add .
   git commit -m "Update Engineer role"
   ```

2. **After Adding New Role**
   ```bash
   # Created roles/new-role.md
   # Create tests/test_role_new.py
   cd tests/
   python3 pre-change-validation.py
   # Fix any errors
   git add .
   git commit -m "Add New role with tests"
   ```

3. **After Modifying Tests**
   ```bash
   # Modified tests/test_role_engineer.py
   cd tests/
   python3 pre-change-validation.py --quick
   git add .
   git commit -m "Update Engineer tests"
   ```

### Recommended (SHOULD Run)

1. **Before Pushing to Remote**
   ```bash
   cd tests/
   python3 pre-change-validation.py
   git push
   ```

2. **After Pulling Changes**
   ```bash
   git pull
   cd tests/
   python3 pre-change-validation.py --quick
   ```

3. **Daily Development**
   ```bash
   # Start of day
   cd tests/
   python3 pre-change-validation.py --check
   ```

---

## Automation Options

### Option 1: Git Pre-Commit Hook (Recommended)

**Setup:**
```bash
# From ai-pack root
cp tests/hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

**Behavior:**
- Runs automatically before every `git commit`
- Blocks commit if validation fails
- Can bypass with `git commit --no-verify` (NOT recommended)

**Hook Template:**
```bash
#!/bin/bash
# .git/hooks/pre-commit

cd tests/ || exit 1

echo "Running pre-commit validation..."
python3 pre-change-validation.py --quick

if [ $? -ne 0 ]; then
    echo ""
    echo "❌ Pre-commit validation failed"
    echo "Fix errors above or use 'git commit --no-verify' to bypass (NOT recommended)"
    exit 1
fi

echo "✅ Pre-commit validation passed"
exit 0
```

### Option 2: Manual Workflow Rule

**Add to your workflow:**
```bash
# Before every commit:
make validate  # or python3 tests/pre-change-validation.py --quick
git add .
git commit -m "message"
```

### Option 3: CI/CD Integration

**In GitHub Actions / GitLab CI:**
```yaml
test:
  script:
    - cd tests/
    - python3 pre-change-validation.py
```

---

## Common Scenarios

### Scenario 1: Adding a New Role

**Steps:**
1. Create role file: `roles/my-new-role.md`
2. Create test file: `tests/test_role_my_new.py`
3. Write tests validating role deliverables
4. Run validation:
   ```bash
   cd tests/
   python3 pre-change-validation.py
   ```
5. Fix any errors
6. Commit:
   ```bash
   git add roles/my-new-role.md tests/test_role_my_new.py
   git commit -m "Add My New role with tests"
   ```

### Scenario 2: Modifying Existing Role

**Steps:**
1. Modify role file: `roles/engineer.md`
2. Update corresponding test: `tests/test_role_engineer.py`
3. Run validation:
   ```bash
   cd tests/
   python3 pre-change-validation.py --quick
   ```
4. Fix any test failures
5. Commit:
   ```bash
   git add roles/engineer.md tests/test_role_engineer.py
   git commit -m "Update Engineer role and tests"
   ```

### Scenario 3: Removing a Role

**Steps:**
1. Remove role file: `git rm roles/deprecated-role.md`
2. Remove or update test mapping in `pre-change-validation.py`
3. Run validation:
   ```bash
   cd tests/
   python3 pre-change-validation.py
   ```
4. Commit:
   ```bash
   git add .
   git commit -m "Remove deprecated role"
   ```

### Scenario 4: Test Failures

**Steps:**
1. Run validation and see failures
2. Identify failing test:
   ```bash
   cd tests/
   python3 run_tests.py -v | grep FAIL
   ```
3. Fix the issue in code or tests
4. Re-run validation:
   ```bash
   python3 pre-change-validation.py --quick
   ```
5. Repeat until all pass

---

## Troubleshooting

### "1 tests failing" (settings.json)

**Issue:** Pre-existing `.claude/settings.json` infrastructure issue

**Solution:** This is allowed. Validation should show:
```
⚠️  Allowing 1 pre-existing .claude/settings.json failure
✅ All role tests passing (1 known infrastructure issue)
```

If you see this error, the validator is working correctly.

### "Missing tests for roles"

**Issue:** A role file exists but no corresponding test

**Solution:**
1. Check role-to-test mapping in `pre-change-validation.py`
2. Create missing test file
3. Re-run validation

### "Test execution failed"

**Issue:** Test runner error (not test failure)

**Solution:**
1. Check test runner exists: `ls tests/run_tests.py`
2. Verify Python installation: `python3 --version`
3. Check test file syntax
4. Review full test output

### "Not in a git repository"

**Issue:** Running outside of git repo

**Solution:**
```bash
# Navigate to ai-pack root
cd /path/to/ai-pack
cd tests/
python3 pre-change-validation.py
```

---

## Best Practices

### 1. Always Run Before Committing

**Good:**
```bash
cd tests/
python3 pre-change-validation.py --quick
git add .
git commit -m "..."
```

**Bad:**
```bash
git add .
git commit -m "..."
# Oops, forgot to validate
```

### 2. Fix Failures Immediately

**Good:**
```bash
python3 pre-change-validation.py --quick
# Validation failed
# Fix issues NOW
python3 pre-change-validation.py --quick
# All pass
git commit
```

**Bad:**
```bash
python3 pre-change-validation.py --quick
# Validation failed
# I'll fix it later... (never does)
```

### 3. Update Tests with Role Changes

**Good:**
```bash
# Modified roles/engineer.md to add new responsibility
vim tests/test_role_engineer.py
# Add test for new responsibility
python3 pre-change-validation.py
```

**Bad:**
```bash
# Modified roles/engineer.md
# Commit without updating tests
# Tests now incomplete
```

### 4. Use --quick for Rapid Iteration

**During development:**
```bash
# Fast validation during active development
python3 pre-change-validation.py --quick
```

**Before pushing:**
```bash
# Full validation before push
python3 pre-change-validation.py
```

### 5. Never Bypass Without Reason

**Good:**
```bash
# Normal commit
python3 pre-change-validation.py --quick
git commit -m "..."
```

**Acceptable (rare):**
```bash
# Emergency hotfix, tests unrelated
git commit --no-verify -m "HOTFIX: Critical production issue"
# THEN fix tests immediately after
```

**Bad:**
```bash
# Lazy bypass because tests failing
git commit --no-verify -m "Whatever"
# Tests never fixed
```

---

## Integration with CI/CD

### GitHub Actions

```yaml
name: Tests

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-python@v2
        with:
          python-version: '3.9'
      - name: Run validation
        run: |
          cd tests/
          python3 pre-change-validation.py
```

### GitLab CI

```yaml
test:
  stage: test
  script:
    - cd tests/
    - python3 pre-change-validation.py
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
```

---

## Extending the Validator

### Add New Validation Check

**Example: Verify documentation exists for new roles**

```python
def _verify_role_documentation(self) -> bool:
    """Verify all roles have documentation"""
    print("\nStep X: Verifying role documentation")
    print("-" * 70)

    roles_dir = self.repo_root / "roles"
    docs_dir = self.repo_root / "docs" / "roles"

    for role_file in roles_dir.glob("*.md"):
        doc_file = docs_dir / role_file.name
        if not doc_file.exists():
            self.errors.append(f"Missing documentation for {role_file.name}")

    return len(self.errors) == 0
```

Add to `run()` method:
```python
# Step 5: Verify documentation
if not self._verify_role_documentation():
    return False
```

---

## Summary

**Pre-Change Validation Rule ensures:**
✅ All tests pass before commits
✅ All roles have corresponding tests
✅ Role changes accompanied by test updates
✅ Test structure remains intact

**Commands:**
```bash
python3 pre-change-validation.py         # Full validation
python3 pre-change-validation.py --quick # Quick validation
python3 pre-change-validation.py --check # Check only (don't block)
```

**Integration:**
- Git pre-commit hook (recommended)
- Manual workflow rule
- CI/CD pipeline

**Benefit:**
Prevents broken tests from being committed, ensures comprehensive test coverage remains 100%, validates workflow quality continuously.

---

**Version:** 1.0
**Last Updated:** 2026-01-15
**Next Review:** When new validation rules needed
