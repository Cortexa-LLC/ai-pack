# AI-Pack Executable Test Suite

**Status:** Active
**Last Updated:** 2026-01-15
**Type:** Automated Python Tests

---

## Overview

This document describes the **executable test suite** for AI-Pack - actual Python unit and integration tests that run automatically without manual intervention.

**Key Difference from validation/:**
- `validation/` = Test case **documentation** (manual execution guides)
- `test_*.py` = **Executable tests** (automated validation)

---

## Executable Test Files

### 1. `test_background_agent_permissions.py`

**Type:** Unit Tests
**Purpose:** Validate .claude/settings.json configuration
**Based On:** TC-BA-005 (Permission Pre-Verification)

**What It Tests:**
```python
✅ test_settings_json_exists()
   - Verifies .claude/settings.json exists
   - Fails if missing with setup instructions

✅ test_settings_json_valid()
   - Validates JSON syntax
   - Fails if malformed

✅ test_write_permission_configured()
   - Checks Write(*) in permissions.allow
   - BLOCKING if missing

✅ test_edit_permission_configured()
   - Checks Edit(*) in permissions.allow
   - Warning if missing (recommended)

✅ test_read_permission_configured()
   - Checks Read(*) in permissions.allow
   - Warning if missing (recommended)

✅ test_default_mode_bypass_permissions()
   - Verifies defaultMode: "bypassPermissions"
   - BLOCKING if wrong

✅ test_local_override_warning()
   - Detects settings.local.json
   - Warns about potential override

✅ test_gate_would_pass_with_correct_config()
   - Simulates Gate 08 enforcement
   - Verifies gate would allow spawning

✅ test_gate_would_block_without_write()
   - Simulates Gate 08 blocking
   - Verifies gate would correctly block

✅ test_gate_would_block_with_wrong_mode()
   - Simulates Gate 08 blocking on wrong mode
   - Verifies gate enforcement
```

**Run:**
```bash
python3 -m unittest test_background_agent_permissions
```

**Expected:** All tests pass if permissions configured correctly

---

### 2. `test_integration_background_agent_spawn.py`

**Type:** Integration Tests
**Purpose:** Verify actual file creation in repository
**Based On:** TC-INT-001 (Spawned Agent File Persistence)

**What It Tests:**
```python
✅ test_01_create_simple_file()
   - Creates actual text file in repository
   - Verifies file exists at expected location
   - Validates file content
   - Confirms file in repository (not sandbox)

✅ test_02_create_subdirectory_structure()
   - Creates nested directories
   - Creates JSON file in deep structure
   - Validates JSON syntax
   - Verifies directory hierarchy

✅ test_03_create_multiple_files_atomically()
   - Creates 3 files in one operation
   - Verifies all files present
   - Tests atomic file creation

✅ test_04_verify_no_sandbox_pollution()
   - Searches common sandbox locations (/tmp, etc.)
   - Fails if test files found outside repository
   - Validates isolation

✅ test_05_absolute_path_resolution()
   - Creates file with explicit absolute path
   - Verifies path resolves correctly
   - Confirms no relative path issues
```

**Run:**
```bash
python3 -m unittest test_integration_background_agent_spawn
```

**Expected:** All tests pass, files created in `.ai/test-artifacts/`, cleaned up after

---

## Test Runner: `run_tests.py`

**Purpose:** Unified test execution and reporting

**Features:**
- Auto-discovers all `test_*.py` files
- Runs tests with unittest framework
- Generates markdown reports
- Provides pass/fail summary
- Cross-platform compatible

**Usage:**

```bash
# Run all tests
python3 run_tests.py

# Run only unit tests
python3 run_tests.py --unit

# Run only integration tests
python3 run_tests.py --integration

# Quick tests (skip slow integration)
python3 run_tests.py --quick

# Verbose output
python3 run_tests.py -v
```

**Output:**
```
======================================================================
AI-Pack Automated Test Suite
======================================================================

Running: All tests
Tests directory: /path/to/tests

test_default_mode_bypass_permissions ... ok
test_settings_json_exists ... ok
test_write_permission_configured ... ok
test_01_create_simple_file ... ok
test_02_create_subdirectory_structure ... ok
test_03_create_multiple_files_atomically ... ok
test_04_verify_no_sandbox_pollution ... ok
test_05_absolute_path_resolution ... ok

----------------------------------------------------------------------
Ran 13 tests in 3.245s

OK

======================================================================
TEST REPORT
======================================================================

# AI-Pack Test Execution Report

**Generated:** 2026-01-15 14:30:22
**Tests Run:** 13

---

## Summary

- **Total Tests:** 13
- **Passed:** 13
- **Failed:** 0
- **Errors:** 0
- **Skipped:** 0
- **Success Rate:** 100%

---

## Status

✅ **ALL TESTS PASSED**

The framework is working correctly. All automated validations passed.

Safe to proceed with workflow changes.

---

**Report Location:** reports/2026-01-15-143022-test-run.md
**Generated:** 2026-01-15 14:30:22

✅ Test report saved to: reports/2026-01-15-143022-test-run.md
```

---

## Test Reports

**Location:** `reports/YYYY-MM-DD-HHMMSS-test-run.md`

**Format:** Markdown

**Contents:**
- Test execution timestamp
- Total tests run
- Pass/fail breakdown
- Detailed failure information
- Error tracebacks
- Skipped tests with reasons
- Recommendations for fixes

**Example Report:**
```markdown
# AI-Pack Test Execution Report

**Generated:** 2026-01-15 14:30:22
**Tests Run:** 13

## Summary

- **Total Tests:** 13
- **Passed:** 12
- **Failed:** 1
- **Success Rate:** 92%

## Status

❌ **TESTS FAILED**

1 test(s) failed, 0 error(s) occurred.

**Required Actions:**
1. Review failures below
2. Fix identified issues
3. Re-run tests
4. Verify all pass before deployment

## Failures

### test_write_permission_configured

```
AssertionError: ❌ Write(*) not in permissions.allow
Current: ['Edit(*)', 'Read(*)', 'Bash(git:*)']
Required: Add "Write(*)" to permissions.allow array
```

## Recommendations

Fix Write(*) permission:
1. Add "Write(*)" to .claude/settings.json
2. Or run: python3 .ai-pack/templates/.claude-setup.py
3. Re-run tests
```

---

## Comparison: Executable vs Manual Tests

| Aspect | Executable Tests | Manual Tests (validation/) |
|--------|------------------|---------------------------|
| **Execution** | Automatic (Python unittest) | Manual (human follows steps) |
| **Speed** | Seconds | Minutes to hours |
| **Repeatability** | 100% consistent | Varies by human |
| **CI/CD** | ✅ Can integrate | ❌ Cannot automate |
| **Reporting** | Auto-generated markdown | Manual documentation |
| **Coverage** | Permission checks, file creation | Full workflows, agent behavior |
| **Best For** | Regression testing, CI/CD | Exploratory testing, complex scenarios |

**Recommendation:** Use **executable tests** for continuous validation, **manual tests** for complex workflow validation.

---

## Adding New Executable Tests

### Step 1: Create Test File

```python
# tests/test_new_feature.py

import unittest
from pathlib import Path

class TestNewFeature(unittest.TestCase):
    """Test new feature functionality"""

    def test_feature_works(self):
        """Test that feature works correctly"""
        # Arrange
        expected = "result"

        # Act
        actual = some_function()

        # Assert
        self.assertEqual(actual, expected)

if __name__ == "__main__":
    unittest.main()
```

### Step 2: Follow Naming Convention

**Files:** `test_*.py`
**Classes:** `Test*`
**Methods:** `test_*`

### Step 3: Run Test

```bash
# Run specific test
python3 -m unittest test_new_feature

# Run all tests (includes new one)
python3 run_tests.py
```

### Step 4: Document Test

Create corresponding documentation in `validation/` directory if needed for complex scenarios.

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: AI-Pack Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v2

    - name: Set up Python
      uses: actions/setup-python@v2
      with:
        python-version: '3.9'

    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        # Add any dependencies

    - name: Run tests
      run: |
        cd tests
        python3 run_tests.py

    - name: Upload test report
      uses: actions/upload-artifact@v2
      with:
        name: test-report
        path: tests/reports/
```

---

## Test Coverage

### Current Coverage

**Unit Tests:**
- ✅ Permission configuration validation
- ✅ Gate 08 enforcement logic
- ✅ Settings.json parsing
- ✅ Local override detection

**Integration Tests:**
- ✅ File creation in repository
- ✅ Subdirectory structure creation
- ✅ Multiple file atomic operations
- ✅ Sandbox isolation verification
- ✅ Absolute path resolution

### Planned Coverage

**Unit Tests (Future):**
- [ ] Batch size validation
- [ ] WIP limit enforcement
- [ ] Token budget estimation
- [ ] Task decomposition logic

**Integration Tests (Future):**
- [ ] Actual spawned agent spawn (requires Claude API)
- [ ] Multi-file task execution
- [ ] Parallel agent coordination
- [ ] Gate blocking verification

---

## Troubleshooting

### Test Failure: "settings.json not found"

**Problem:** .claude/settings.json doesn't exist

**Solution:**
```bash
python3 .ai-pack/templates/.claude-setup.py
```

---

### Test Failure: "Write(*) not configured"

**Problem:** Missing Write permission

**Solution:**
```bash
# Add to .claude/settings.json
{
  "permissions": {
    "allow": [
      "Write(*)",  // ← Add this
      ...
    ]
  }
}
```

---

### Test Failure: "Files not in repository"

**Problem:** Integration test created files outside repo

**Solution:**
- Check working directory context
- Verify absolute paths used
- Review test_dir resolution

---

### All Tests Skipped

**Problem:** Pre-requisites not met

**Solution:**
- Run unit tests first
- Fix permission configuration
- Then run integration tests

---

## Success Criteria

**Tests are successful when:**

✅ All unit tests pass (100%)
✅ All integration tests pass (100%)
✅ Reports generated automatically
✅ No manual intervention needed
✅ Tests complete in <5 seconds (unit), <30 seconds (integration)
✅ CI/CD can run tests

**Leading Indicators:**
- Tests run on every commit
- Failures detected immediately
- Clear error messages
- Quick fixes possible
- No regression issues

---

## References

**Documentation:**
- `validation/background-agents/TC-BA-005-permission-verification.md`
- `validation/integration/TC-INT-001-background-agent-file-persistence.md`
- `gates/08-background-agent-permissions.md`

**Tools:**
- `tools/verify-background-agent-permissions.py`
- `tools/verify-agent-output.py`

**Python unittest:** https://docs.python.org/3/library/unittest.html

---

**Version:** 1.0.0
**Status:** Active
**Maintainer:** AI-Pack Team
