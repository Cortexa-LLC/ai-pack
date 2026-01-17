# AI-Pack Testing - Quick Start Guide

**Purpose:** Get started with AI-Pack validation testing quickly

---

## Prerequisites

- Python 3.7+ installed
- AI-Pack repository cloned
- Terminal/command prompt access

---

## Quick Commands

### List All Available Tests

```bash
cd tests/
python3 tools/run-validation.py --list
```

**Output:**
```
Available Test Cases:

Category: background-agents
  [Critical] TC-BA-001: Spawned Agent File Persistence Verification
  [Critical] TC-BA-002: Token Limit Detection and Error Reporting

Category: orchestrator
  [Critical] TC-OR-001: Orchestrator Completion Verification Protocol

...
```

### Run Critical Tests Only

```bash
python3 tools/run-validation.py --critical
```

This runs all tests marked as **Priority: Critical** - the must-pass tests before any workflow deployment.

### Run Specific Category

```bash
# Test spawned agents
python3 tools/run-validation.py --category background-agents

# Test orchestrator
python3 tools/run-validation.py --category orchestrator

# Test gates
python3 tools/run-validation.py --category gates
```

### Run Specific Test

```bash
python3 tools/run-validation.py --test TC-BA-001
```

---

## Test Execution Flow

When you run a test, here's what happens:

1. **Test Selection**
   - Script finds matching test cases
   - Displays test metadata (priority, status, title)

2. **Manual Execution Prompt**
   ```
   Running: TC-BA-001
     Priority: Critical
     Status: Active
     Title: TC-BA-001: Spawned Agent File Persistence Verification

     Test case: /path/to/TC-BA-001-file-persistence.md
     Follow manual execution steps in test case document

     Execute test now? (y/n/s=skip):
   ```

3. **Test Case Opens**
   - Opens in VS Code (or default editor)
   - Follow the scenario steps in the test case
   - Perform actions as described

4. **Record Result**
   ```
   Test result? (p=pass/f=fail): p
   ```
   - Enter `p` if test passed
   - Enter `f` if test failed
   - If failed, provide failure notes

5. **Automatic Reporting**
   - Results recorded to `reports/YYYY-MM-DD-HHMMSS-test-run.md`
   - Summary displayed at end

---

## Critical Test Cases (Must Pass)

These tests MUST pass before deploying workflow changes:

### Spawned Agents
```bash
python3 tools/run-validation.py --test TC-BA-001  # File Persistence
python3 tools/run-validation.py --test TC-BA-002  # Token Limit Detection
```

### Orchestrator
```bash
python3 tools/run-validation.py --test TC-OR-001  # Completion Verification
python3 tools/run-validation.py --test TC-OR-005  # Task Decomposition
```

**OR run all critical at once:**
```bash
python3 tools/run-validation.py --critical
```

---

## Agent Output Verification

For testing orchestrator verification protocol:

```bash
# Verify agent output file
python3 tools/verify-agent-output.py /path/to/agent-output.txt

# Specify working directory for file paths
python3 tools/verify-agent-output.py /path/to/agent-output.txt \
  --working-dir /Users/user/project

# Generate markdown report
python3 tools/verify-agent-output.py /path/to/agent-output.txt --markdown
```

**Output:**
```
Running agent output verification...

Step 1: Error Pattern Check
  ✓ No error patterns found

Step 2: Write() Call Analysis
  ✓ Write() calls detected: 2

Step 3: Claimed Files Extraction
  ✓ Files claimed: 2
    src/UserService.cs
    tests/UserServiceTests.cs

Step 4: File Existence Verification
  ✓ All 2 file(s) exist and not empty
    ✓ src/UserService.cs (2,048 bytes)
    ✓ tests/UserServiceTests.cs (1,536 bytes)

Step 5: Overall Status
  ✅ VERIFIED - Agent completed successfully
    All verification steps passed
    Files persisted successfully
```

---

## Reading Test Reports

After running tests, check the report:

```bash
# View latest report
ls -t reports/*.md | head -1 | xargs cat

# Or navigate to reports/
cd reports/
ls -t *.md | head -1  # Latest report
```

**Report format:**
```markdown
# AI-Pack Test Execution Report

**Date:** 2026-01-15
**Time:** 14:30:00

---

## Test Results Summary

### TC-BA-001: Spawned Agent File Persistence Verification

**Status:** ✅ PASSED
**Notes:** Manual execution successful

---

### TC-BA-002: Token Limit Detection and Error Reporting

**Status:** ❌ FAILED
**Notes:** Token limit not detected correctly

---

## Summary Statistics

- **Total Tests:** 2
- **Passed:** 1
- **Failed:** 1
- **Pass Rate:** 50%

## Recommendations

⚠️ **CRITICAL:** 1 test(s) failed.

**Required Actions:**
1. Review failed test cases above
2. Identify root causes
3. Fix identified issues
4. Re-run test suite
5. Verify all tests pass before deployment

**DO NOT deploy workflow changes until all critical tests pass.**
```

---

## Typical Workflow

### Before Making Changes

1. **Establish baseline:**
   ```bash
   python3 tools/run-validation.py --critical
   ```
   - All critical tests should pass
   - Creates baseline report

### After Making Changes

2. **Run regression tests:**
   ```bash
   python3 tools/run-validation.py --critical
   ```
   - Compare results to baseline
   - Ensure no regressions

3. **If tests fail:**
   - Review test case scenarios
   - Identify what broke
   - Fix the issue
   - Re-run tests

4. **When all pass:**
   - Commit changes
   - Include test report in commit message or PR

---

## Common Scenarios

### Scenario 1: Testing Spawned Agent Fix

```bash
# 1. Make changes to orchestrator.md or agent behavior

# 2. Run spawned agent tests
python3 tools/run-validation.py --category background-agents

# 3. Follow test scenarios for TC-BA-001 and TC-BA-002

# 4. Verify:
#    - Files persist correctly (TC-BA-001)
#    - Token limits detected (TC-BA-002)

# 5. Check report
cat reports/$(ls -t reports/*.md | head -1)
```

### Scenario 2: Testing Orchestrator Completion

```bash
# 1. Make changes to orchestrator verification protocol

# 2. Run orchestrator tests
python3 tools/run-validation.py --test TC-OR-001

# 3. Follow test scenario:
#    - Scenario A: Failed agent (should detect)
#    - Scenario B: Successful agent (should pass)

# 4. Verify both scenarios work correctly
```

### Scenario 3: Full Regression Suite

```bash
# Run ALL critical tests before release
python3 tools/run-validation.py --critical

# Review report
cat reports/$(ls -t reports/*.md | head -1)

# Must see:
# - Pass Rate: 100%
# - All tests passed
```

---

## Test Case Structure

Each test case contains:

1. **Objective** - What we're validating
2. **Background** - Why this test exists (production failures)
3. **Prerequisites** - What you need before testing
4. **Test Scenario** - Step-by-step execution
5. **Expected Behavior** - What should happen
6. **Pass/Fail Criteria** - Clear success criteria
7. **Known Issues** - Related problems and mitigations

---

## Tips

### Faster Testing

- Use `--test` for single test execution
- Use `--category` to focus on one area
- Use `--critical` for pre-deployment checks

### Understanding Failures

- Read the test case Background section
- Review actual production failures referenced
- Check commit references for fixes

### Contributing Test Cases

1. Create new test in appropriate category
2. Follow test case template
3. Base on real production scenarios
4. Include clear pass/fail criteria

---

## Help

```bash
# View all options
python3 tools/run-validation.py --help

# Agent verification help
python3 tools/verify-agent-output.py --help
```

---

## Next Steps

1. **Run critical tests now:**
   ```bash
   cd tests/
   python3 tools/run-validation.py --critical
   ```

2. **Review test cases:**
   - Read `validation/background-agents/TC-BA-001-file-persistence.md`
   - Understand the production failures
   - See how tests prevent regressions

3. **Add new tests:**
   - When you find a bug, create a test case
   - Follow template in existing test cases
   - Add to appropriate category

---

**Questions?** See [README.md](README.md) for complete documentation.
