# AI-Pack Test Suite

**Purpose:** Comprehensive validation test suite for AI-Pack workflows, roles, gates, and integrations.

**Last Updated:** 2026-01-15
**Version:** 2.0
**Status:** ✅ Production Ready

---

## Directory Structure

```
tests/
├── README.md                           # This file
├── run_tests.py                        # ✨ AUTOMATED test runner
├── pre-change-validation.py            # ✨ Pre-commit validation (NEW v2.0)
├── test_*.py                           # ✨ EXECUTABLE test files
│   ├── test_background_agent_permissions.py     # Spawned Agent tests
│   ├── test_integration_background_agent_spawn.py  # Integration tests
│   ├── test_role_engineer.py           # ✨ Engineer role tests (NEW)
│   ├── test_role_reviewer.py           # ✨ Reviewer role tests (NEW)
│   ├── test_role_tester.py             # ✨ Tester role tests (NEW)
│   ├── test_role_specialists.py        # ✨ Specialist roles tests (NEW)
│   ├── test_orchestrator_delegation.py # ✨ Orchestrator tests (NEW)
│   └── test_beads_integration.py       # ✨ Beads integration tests (NEW v2.0)
├── hooks/                              # ✨ Git hooks (NEW v2.0)
│   └── pre-commit                      # Pre-commit validation hook
├── validation/                         # Test case documentation
│   ├── background-agents/              # Spawned Agent test cases
│   ├── orchestrator/                   # Orchestrator behavior test cases
│   ├── gates/                          # Gate enforcement test cases
│   ├── workflows/                      # Workflow compliance test cases
│   └── integration/                    # End-to-end integration tests
├── fixtures/                           # Test fixtures and mock data
│   ├── task-packets/                   # Sample task packets
│   ├── code-samples/                   # Sample code for testing
│   └── agent-outputs/                  # Sample agent outputs
├── reports/                            # Test execution reports
│   └── YYYY-MM-DD-HHMMSS-test-run.md  # Automated test reports
└── tools/                              # Test utilities
    ├── run-validation.py               # Manual test runner (deprecated)
    ├── verify-agent-output.py          # Agent output verification
    └── verify-background-agent-permissions.py  # Permission checker
```

---

## Test Case Categories

### 1. Spawned Agent Tests (`validation/background-agents/`)

**Focus:** Spawned Agent file persistence, error detection, and completion verification

**Critical Issues Being Tested:**
- Agent claims success but files not persisted
- Token limit failures not detected
- Silent permission failures
- Sandbox isolation issues

**Test Cases:**
- `TC-BA-001` - File Persistence Verification
- `TC-BA-002` - Token Limit Detection
- `TC-BA-003` - Working Directory Context
- `TC-BA-004` - Absolute Path Requirements
- `TC-BA-005` - Permission Pre-Verification
- `TC-BA-006` - Partial Completion Detection (TODO)

### 2. Orchestrator Tests (`validation/orchestrator/`)

**Focus:** Orchestrator coordination, delegation, and completion reporting

**Critical Issues Being Tested:**
- False success reporting
- Incomplete work marked as complete
- Missing artifact persistence verification
- Parallel execution not triggered

**Test Cases:**
- `TC-OR-001` - Completion Verification Protocol
- `TC-OR-002` - Artifact Persistence Enforcement (TODO)
- `TC-OR-003` - Parallel Execution Trigger (TODO)
- `TC-OR-004` - Agent Registration Protocol (TODO)
- `TC-OR-005` - Task Decomposition Size Limits
- `TC-OR-006` - Cross-Reference Verification (TODO)

### 3. Gate Enforcement Tests (`validation/gates/`)

**Focus:** Quality gate enforcement and blocking mechanisms

**Test Cases:**
- `TC-GT-001` - Task Packet Gate
- `TC-GT-002` - TDD Enforcement Gate
- `TC-GT-003` - Code Quality Review Gate
- `TC-GT-004` - Artifact Persistence Gate
- `TC-GT-005` - Execution Strategy Gate

### 4. Workflow Compliance Tests (`validation/workflows/`)

**Focus:** Workflow phase execution and compliance

**Test Cases:**
- `TC-WF-001` - Feature Workflow Compliance
- `TC-WF-002` - Bugfix Workflow Compliance
- `TC-WF-003` - Refactor Workflow Compliance
- `TC-WF-004` - Specialist Integration (Product Manager, Architect, Designer)

### 5. Integration Tests (`validation/integration/`)

**Focus:** End-to-end workflow execution with multiple agents

**Test Cases:**
- `TC-INT-001` - Full Feature Development Cycle
- `TC-INT-002` - Parallel Multi-Agent Coordination
- `TC-INT-003` - Bugfix Investigation to Resolution
- `TC-INT-004` - Large Task Decomposition

---

## Test Execution

### ✨ Automated Testing (PRIMARY METHOD)

**Run all executable tests:**
```bash
cd tests/
python3 run_tests.py
```

**Quick tests only (no integration):**
```bash
python3 run_tests.py --quick
```

**Integration tests only:**
```bash
python3 run_tests.py --integration
```

**What gets tested:**
- ✅ Permission configuration validation
- ✅ Gate enforcement logic
- ✅ File persistence verification
- ✅ Working directory context
- ✅ Absolute path resolution
- ✅ Sandbox isolation prevention

**Output:**
- Detailed test results in console
- Automated report in `reports/YYYY-MM-DD-HHMMSS-test-run.md`
- Pass/fail status with actionable guidance

---

### Manual Testing (DEPRECATED - Use automated tests instead)

The `tools/run-validation.py` script provides manual test execution with prompts.
This is being phased out in favor of automated tests.

**Only use manual testing for:**
- Tests not yet automated
- Exploratory testing
- Documentation verification

---

## Test Case Format

Each test case follows this structure:

```markdown
# TC-XX-NNN: Test Case Title

**Category:** [Spawned Agents | Orchestrator | Gates | Workflows | Integration]
**Priority:** [Critical | High | Medium | Low]
**Status:** [Draft | Active | Deprecated]
**Last Updated:** YYYY-MM-DD

---

## Objective

What this test validates.

## Background

Context and motivation for this test case.

## Prerequisites

- System state requirements
- Required permissions
- Setup steps

## Test Scenario

Step-by-step scenario to execute.

## Expected Behavior

What SHOULD happen if framework is working correctly.

## Actual Behavior (Execution Record)

What actually happened during test execution.

## Pass/Fail Criteria

Clear criteria for determining success.

## Known Issues

Related bugs or limitations.

## References

- Related commits
- Related issues
- Related documentation
```

---

## Test Execution Reports

After each test run, create a report in `reports/`:

**Format:** `YYYY-MM-DD-test-run.md`

**Contents:**
- Test cases executed
- Pass/fail summary
- Issues discovered
- Recommendations
- Regression status

---

## Critical Test Cases (Must Pass)

These test cases MUST pass before any workflow change is deployed:

### Spawned Agents
- **TC-BA-001** - File Persistence Verification
- **TC-BA-002** - Token Limit Detection
- **TC-BA-003** - Working Directory Context
- **TC-BA-004** - Absolute Path Requirements
- **TC-BA-005** - Permission Pre-Verification

### Orchestrator
- **TC-OR-001** - Completion Verification Protocol
- **TC-OR-002** - Artifact Persistence Enforcement (TODO)
- **TC-OR-005** - Task Decomposition Size Limits

### Gates
- **TC-GT-002** - TDD Enforcement Gate
- **TC-GT-003** - Code Quality Review Gate

---

## Test Case Development Guidelines

### When to Add a New Test Case

Add a new test case when:
1. **Production failure** occurs - Document the failure scenario
2. **New feature** added - Validate it works as expected
3. **Gate changed** - Ensure enforcement still works
4. **Edge case** discovered - Prevent regression

### Test Case Naming

- **Prefix:** `TC-{CATEGORY}-{NUMBER}`
- **Categories:** BA (Spawned Agents), OR (Orchestrator), GT (Gates), WF (Workflows), INT (Integration)
- **Number:** Sequential 001, 002, 003...

**Examples:**
- `TC-BA-001` - First spawned agent test
- `TC-OR-005` - Fifth orchestrator test
- `TC-GT-002` - Second gate test

### Test Case Priority

- **Critical:** Production failures, data loss risks, security issues
- **High:** Common workflows, user-facing features
- **Medium:** Edge cases, optimization checks
- **Low:** Documentation validation, style checks

---

## Quick Start

### 1. Run All Automated Tests ✨ RECOMMENDED

```bash
cd tests/
python3 run_tests.py
```

**Expected output:**
```
====================================================================================
AI-Pack Automated Test Suite
========================================== ==========================================

Running: All tests
Tests directory: /path/to/tests

test_default_mode_bypass_permissions ... ok
test_edit_permission_configured ... ok
test_settings_json_exists ... ok
test_write_permission_configured ... ok
test_01_create_simple_file ... ok
test_02_create_subdirectory_structure ... ok
...

----------------------------------------------------------------------
Ran 10 tests in 2.456s

OK

✅ ALL TESTS PASSED

Test report saved to: reports/2026-01-15-143022-test-run.md
```

### 2. Quick Validation (Fast Tests Only)

```bash
# Skip slow integration tests
python3 run_tests.py --quick
```

### 3. Run Specific Test File

```bash
# Run only permission tests
python3 -m unittest test_background_agent_permissions

# Run only integration tests
python3 -m unittest test_integration_background_agent_spawn
```

### 4. View Test Documentation

```bash
# View test case documentation
cat validation/background-agents/TC-BA-005-permission-verification.md

# View integration test documentation
cat validation/integration/TC-INT-001-background-agent-file-persistence.md
```

---

## Regression Testing

After any change to:
- `.ai-pack/roles/orchestrator.md`
- `.ai-pack/gates/*.md`
- `.ai-pack/workflows/*.md`
- `.claude/skills/**/*.md`

**MUST run full critical test suite** and verify no regressions.

---

## Contributing Test Cases

1. Create test case in appropriate `validation/` subdirectory
2. Follow test case format (see above)
3. Include realistic scenario based on production usage
4. Document expected vs actual behavior
5. Link to related issues/commits
6. Add to critical suite if applicable

---

## References

- **Recent Failures:**
  - Commit `edf1d8a` - Token limit + false success (consumer-project 2026-01-15)
  - Commit `e1764ec` - Task decomposition failures
  - Commit `801fb14` - Working directory context

- **Framework Documentation:**
  - [Orchestrator Role](../roles/orchestrator.md)
  - [Gates](../gates/)
  - [Workflows](../workflows/)

---

**Version:** 1.0.0
**Status:** Active
**Maintainer:** Bryan Woodruff
