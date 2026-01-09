# Code Quality Review Gate

**Version:** 1.0.0
**Last Updated:** 2026-01-08
**Type:** Mandatory Enforcement Gate

## Overview

The Code Quality Review Gate is a **mandatory checkpoint** that enforces formal code review and test validation for all work packages involving code changes. This gate ensures that both test discipline (TDD) and code quality standards are verified before work is considered complete.

**Purpose:** Ensure that all code changes undergo mandatory verification by specialized review agents (Tester and Reviewer) before acceptance. If either review finds blocking issues, the work is considered incomplete.

---

## Trigger Conditions

This gate activates when:
```
IF work package includes code changes THEN
  TRIGGER code quality review gate
  REQUIRE Tester validation (TDD and test sufficiency)
  REQUIRE Reviewer validation (code quality and standards)
  BLOCK completion until both validations pass
END IF
```

**Specifically triggers when:**
- Implementation phase complete (Phase 3 exit)
- Code files modified (not just documentation/configuration)
- Entering review phase (Phase 4)
- Before work acceptance and sign-off

**Does NOT trigger when:**
- Pure documentation changes (no code modified)
- Configuration-only changes (no logic modified)
- Research or exploration tasks (no implementation)
- Simple file operations (moving, renaming without logic changes)

---

## Mandatory Review Procedure

The orchestrator MUST complete this review procedure before accepting work:

### Step 1: Determine Review Requirements

```
PROCEDURE determine_review_requirements:
  IF work_package includes code changes THEN
    code_changes = identify_modified_code_files()
    has_logic_changes = check_for_logic_changes(code_changes)
    has_test_changes = check_for_test_changes(code_changes)

    IF has_logic_changes OR has_test_changes THEN
      REQUIRE tester_validation = true
      REQUIRE reviewer_validation = true
    END IF
  END IF

  RETURN (tester_validation_required, reviewer_validation_required)
END PROCEDURE
```

**Code Changes Include:**
```
✓ New source files (.js, .ts, .py, .rs, .go, .java, etc.)
✓ Modified source files (logic changes)
✓ New test files
✓ Modified test files
✓ Deleted source or test files
✓ Refactored code
```

**Exempt from Gate:**
```
⊗ Documentation only (.md files, comments only)
⊗ Configuration only (.json, .yaml, .toml) without logic
⊗ Build scripts without logic changes
⊗ README/CHANGELOG updates
⊗ File renames without content changes
```

---

### Step 2: Delegate to Tester Agent (MANDATORY)

```
PROCEDURE delegate_to_tester:
  STEP 1: Create tester task description
    focus_areas = [
      "TDD process compliance",
      "Test coverage requirements (80-90%)",
      "Test quality assessment",
      "Test type coverage (unit/integration/e2e)"
    ]

  STEP 2: Spawn Tester agent
    tester = Task(
      subagent_type="general-purpose",
      prompt="You are the Tester role from .ai-pack/roles/tester.md.
              Validate TDD compliance and test sufficiency for this work package.
              Focus areas: ${focus_areas}
              Report findings in .ai/tasks/${task_id}/30-review.md"
    )

  STEP 3: Wait for Tester validation
    tester_result = wait_for_completion(tester)

  STEP 4: Parse Tester verdict
    IF tester_result contains "TEST VALIDATION: APPROVED" THEN
      tester_passed = true
    ELSE IF tester_result contains "TEST VALIDATION: CHANGES REQUIRED" THEN
      tester_passed = false
      tester_blockers = extract_critical_major_findings(tester_result)
    END IF

  RETURN (tester_passed, tester_blockers)
END PROCEDURE
```

---

### Step 3: Delegate to Reviewer Agent (MANDATORY)

```
PROCEDURE delegate_to_reviewer:
  STEP 1: Create reviewer task description
    focus_areas = [
      "Code quality against standards",
      "Architecture consistency",
      "Security concerns",
      "Documentation adequacy"
    ]

  STEP 2: Spawn Reviewer agent
    reviewer = Task(
      subagent_type="general-purpose",
      prompt="You are the Reviewer role from .ai-pack/roles/reviewer.md.
              Review code quality and standards compliance for this work package.
              Focus areas: ${focus_areas}
              Report findings in .ai/tasks/${task_id}/30-review.md"
    )

  STEP 3: Wait for Reviewer validation
    reviewer_result = wait_for_completion(reviewer)

  STEP 4: Parse Reviewer verdict
    IF reviewer_result contains "APPROVED" THEN
      reviewer_passed = true
    ELSE IF reviewer_result contains "CHANGES REQUESTED" THEN
      reviewer_passed = false
      reviewer_blockers = extract_critical_major_findings(reviewer_result)
    END IF

  RETURN (reviewer_passed, reviewer_blockers)
END PROCEDURE
```

---

### Step 4: Evaluate Combined Results

```
PROCEDURE evaluate_review_results:
  (tester_passed, tester_blockers) = delegate_to_tester()
  (reviewer_passed, reviewer_blockers) = delegate_to_reviewer()

  IF tester_passed AND reviewer_passed THEN
    RETURN GATE_PASSED
  ELSE
    all_blockers = tester_blockers + reviewer_blockers
    RETURN GATE_BLOCKED(all_blockers)
  END IF
END PROCEDURE
```

---

### Step 5: Handle Gate Results

```
PROCEDURE handle_gate_results:
  result = evaluate_review_results()

  IF result == GATE_PASSED THEN
    proceed_to_acceptance()
    update_task_packet_acceptance()
    mark_work_complete()
  ELSE IF result == GATE_BLOCKED THEN
    blockers = result.get_blockers()
    report_blockers_to_user()
    coordinate_fixes(blockers)
    resubmit_for_review()
  END IF
END PROCEDURE
```

---

## Enforcement Rules

### Rule 1: Tester Validation is Mandatory for Code Changes

```
ENFORCEMENT:
  IF work includes code changes THEN
    REQUIRE Tester delegation
    REQUIRE Tester completion
    REQUIRE Tester verdict (APPROVED or CHANGES REQUIRED)

    IF Tester not invoked THEN
      GATE VIOLATION: "Code Quality Review Gate - Tester validation missing"
      BLOCK work acceptance
      HALT completion
    END IF

    IF Tester verdict == "CHANGES REQUIRED" THEN
      WORK INCOMPLETE
      REQUIRE fixes for Critical and Major findings
      REQUIRE re-validation by Tester
    END IF
  END IF
```

**Tester Blocking Conditions:**
```
❌ BLOCKERS (work incomplete):
- TDD process not followed
- Test coverage < 80%
- Critical business logic untested (<95%)
- Error handling untested (<90%)
- Integration points untested (<100%)
- Tests failing
- Flaky tests present
- Security test scenarios missing
```

---

### Rule 2: Reviewer Validation is Mandatory for Code Changes

```
ENFORCEMENT:
  IF work includes code changes THEN
    REQUIRE Reviewer delegation
    REQUIRE Reviewer completion
    REQUIRE Reviewer verdict (APPROVED or CHANGES REQUESTED)

    IF Reviewer not invoked THEN
      GATE VIOLATION: "Code Quality Review Gate - Reviewer validation missing"
      BLOCK work acceptance
      HALT completion
    END IF

    IF Reviewer verdict == "CHANGES REQUESTED" THEN
      WORK INCOMPLETE
      REQUIRE fixes for Critical and Major findings
      REQUIRE re-validation by Reviewer
    END IF
  END IF
```

**Reviewer Blocking Conditions:**
```
❌ BLOCKERS (work incomplete):
- Critical security vulnerabilities
- Major standards violations
- Architecture violations
- Poor error handling
- Missing critical tests
- Acceptance criteria not met
- Code smells (significant)
- Documentation inadequate (public APIs)
```

---

### Rule 3: Both Validations Must Pass for Completion

```
ENFORCEMENT:
  result = evaluate_review_results()

  IF NOT tester_passed THEN
    BLOCK completion
    REPORT: "Work incomplete - Tester validation failed"
    REQUIRE: Fix test-related issues
  END IF

  IF NOT reviewer_passed THEN
    BLOCK completion
    REPORT: "Work incomplete - Reviewer validation failed"
    REQUIRE: Fix code quality issues
  END IF

  IF NOT (tester_passed AND reviewer_passed) THEN
    WORK STATUS = INCOMPLETE
    BLOCK acceptance
    BLOCK sign-off
  END IF
```

---

### Rule 4: Fixes Require Re-Review

```
ENFORCEMENT:
  IF issues found by Tester OR Reviewer THEN
    worker_fixes_issues()

    IF Tester found issues THEN
      REQUIRE Tester re-validation
    END IF

    IF Reviewer found issues THEN
      REQUIRE Reviewer re-validation
    END IF

    REPEAT until both validations pass
  END IF
```

---

## Review Orchestration

### Sequential Review Pattern (Recommended)

Execute reviews sequentially to allow test fixes before code review:

```
SEQUENTIAL REVIEW:
  STEP 1: Tester validation first
    tester_result = delegate_to_tester()

    IF tester_result == CHANGES_REQUIRED THEN
      fix_test_issues()
      re_run_tester_validation()
    END IF

  STEP 2: Reviewer validation after tests pass
    // Only proceed when tests validated
    IF tester_passed THEN
      reviewer_result = delegate_to_reviewer()

      IF reviewer_result == CHANGES_REQUESTED THEN
        fix_code_issues()
        re_run_tester_validation()  // Verify tests still pass
        re_run_reviewer_validation()
      END IF
    END IF
```

**Rationale:**
- Tester catches test issues first
- Reviewer sees code with validated tests
- Avoids reviewer time on code with test problems
- Natural progression: tests → code → acceptance

---

### Parallel Review Pattern (Alternative)

Execute reviews in parallel for faster feedback:

```
PARALLEL REVIEW:
  Single message with two Task calls:
    - Task(tester, "Validate TDD and test sufficiency")
    - Task(reviewer, "Review code quality and standards")

  Wait for both to complete

  Evaluate combined results:
    IF any blocker found THEN
      consolidate_feedback()
      coordinate_fixes()
      re_review_both()
    END IF
```

**Rationale:**
- Faster feedback (parallel execution)
- Comprehensive feedback in one cycle
- Use when high confidence in test quality

**Recommendation:** Use sequential for first-time reviews, parallel for re-reviews.

---

## Integration with Workflow

### Phase 4: Review (Enhanced with Mandatory Validation)

**Phase 4 Entry Requirements:**
```
✓ Phase 3 (Implementation) complete
✓ All planned steps executed
✓ Code changes committed
✓ Worker self-review complete
```

**Phase 4 Activities (UPDATED):**
```
4.1 Worker Self-Review (existing)
    □ Run full test suite locally
    □ Check test coverage
    □ Verify build succeeds
    □ Run linters
    □ Basic quality checks

4.2 Tester Validation (NEW - MANDATORY)
    □ Orchestrator delegates to Tester agent
    □ Tester validates TDD compliance
    □ Tester verifies test coverage (80-90%)
    □ Tester assesses test quality
    □ Tester provides verdict: APPROVED or CHANGES REQUIRED

4.3 Test Issue Resolution (NEW - IF NEEDED)
    □ Worker addresses Tester findings (Critical/Major)
    □ Worker re-runs tests and coverage
    □ Worker requests Tester re-validation
    □ Repeat until Tester approves

4.4 Reviewer Validation (NEW - MANDATORY)
    □ Orchestrator delegates to Reviewer agent
    □ Reviewer validates code quality
    □ Reviewer verifies standards compliance
    □ Reviewer assesses architecture/security
    □ Reviewer provides verdict: APPROVED or CHANGES REQUESTED

4.5 Code Issue Resolution (NEW - IF NEEDED)
    □ Worker addresses Reviewer findings (Critical/Major)
    □ Worker re-runs tests (ensure still passing)
    □ Worker requests Tester + Reviewer re-validation
    □ Repeat until Reviewer approves

4.6 User Acceptance (existing)
    □ All acceptance criteria met
    □ User satisfied (if applicable)
    □ Documentation complete
```

**Phase 4 Exit Criteria (UPDATED):**
```
✓ Worker self-review complete
✓ Tester validation: APPROVED (MANDATORY)
✓ Reviewer validation: APPROVED (MANDATORY)
✓ All Critical/Major findings resolved
✓ Tests passing (100%)
✓ Coverage meets target (80-90%)
✓ Standards compliance verified
✓ User acceptance confirmed
✓ Documentation complete
```

---

## Blocking vs. Non-Blocking Issues

### Blocking Issues (Work Incomplete)

Issues that MUST be fixed before approval:

**From Tester:**
```
❌ CRITICAL (blocks acceptance):
- TDD not followed (implementation before tests)
- Coverage < 80%
- Tests failing
- Critical logic untested (<95%)
- Error handling untested (<90%)
- Integration points untested (<100%)

❌ MAJOR (blocks acceptance):
- Flaky/unreliable tests
- Test quality poor (tests implementation, not behavior)
- Missing edge case tests
- Security scenarios untested
- Integration tests missing for integrations
```

**From Reviewer:**
```
❌ CRITICAL (blocks acceptance):
- Security vulnerabilities
- Data corruption risks
- System stability issues
- Breaking changes without approval

❌ MAJOR (blocks acceptance):
- Standards violations (significant)
- Architecture violations
- Missing tests for critical paths
- Poor error handling
- Acceptance criteria not met
```

---

### Non-Blocking Issues (Can Approve with Suggestions)

Issues that don't prevent approval but should be noted:

**From Tester:**
```
⚠ MINOR (suggestions):
- Test naming improvements
- Test refactoring opportunities
- Additional edge cases (non-critical)
- Test documentation
- Performance optimizations
```

**From Reviewer:**
```
⚠ MINOR (suggestions):
- Style inconsistencies (minor)
- Missing comments (non-complex code)
- Naming improvements
- Refactoring opportunities
- Non-critical optimizations
```

**Approval with Suggestions:**
```
Both Tester and Reviewer can approve despite minor issues:

"APPROVED with suggestions:

Consider these improvements for future:
- [Minor suggestion 1]
- [Minor suggestion 2]

These don't block approval - excellent work!"
```

---

## Documentation Requirements

### Review Findings Location

All review findings MUST be documented in:
```
.ai/tasks/${task_id}/30-review.md
```

**Required Sections:**
```markdown
# Review Report

## Tester Validation

**Verdict:** APPROVED | CHANGES REQUIRED

**TDD Compliance:** PASS | FAIL
[Details of TDD analysis]

**Coverage Metrics:**
- Overall: X% (target: 80-90%)
- Business Logic: X% (target: 95%+)
- Error Handling: X% (target: 90%+)

**Findings:**
[Critical/Major/Minor findings with severity, location, issue, recommendation]

**Status:** APPROVED | CHANGES REQUIRED

---

## Reviewer Validation

**Verdict:** APPROVED | CHANGES REQUESTED

**Code Quality:** PASS | FAIL
[Assessment of code quality]

**Standards Compliance:** PASS | FAIL
[Assessment of standards adherence]

**Security:** PASS | FAIL
[Security review findings]

**Findings:**
[Critical/Major/Minor findings with severity, location, issue, recommendation]

**Status:** APPROVED | CHANGES REQUESTED

---

## Combined Result

**Overall Verdict:** APPROVED | WORK INCOMPLETE

**Blocking Issues:** [count]
**Must Fix:** [list of critical/major issues]

**Next Steps:**
[What needs to happen for approval]
```

---

## Gate Compliance Checklist

Before marking work complete, verify:

```
□ Code changes identified
□ Tester agent delegated (if code changes present)
□ Tester validation completed
□ Tester verdict documented in 30-review.md
□ Test issues resolved (if any Critical/Major found)
□ Tester re-validation completed (if issues found)
□ Tester final verdict: APPROVED

□ Reviewer agent delegated (if code changes present)
□ Reviewer validation completed
□ Reviewer verdict documented in 30-review.md
□ Code issues resolved (if any Critical/Major found)
□ Reviewer re-validation completed (if issues found)
□ Reviewer final verdict: APPROVED

□ Both validations: APPROVED
□ All blocking issues resolved
□ Work packet 30-review.md complete
□ Ready for acceptance sign-off

IF all checked THEN
  GATE PASSED - proceed to acceptance
ELSE
  GATE BLOCKED - work incomplete
END IF
```

---

## Violation Handling

### Violation Type 1: Review Skipped

```
Orchestrator attempts to mark work complete without Tester/Reviewer validation

Response:
- GATE VIOLATION: "Code Quality Review Gate - Required reviews missing"
- HALT acceptance
- REPORT to orchestrator: "Code changes require mandatory Tester and Reviewer validation"
- REQUIRE: Delegate to both Tester and Reviewer before proceeding
```

---

### Violation Type 2: Completion Despite Blocking Issues

```
Orchestrator attempts to mark work complete when Tester or Reviewer found blocking issues

Response:
- GATE VIOLATION: "Code Quality Review Gate - Blocking issues unresolved"
- HALT acceptance
- REPORT to orchestrator: "Work incomplete - Critical/Major findings must be addressed"
- REQUIRE: Fix blocking issues and obtain re-approval
```

---

### Violation Type 3: Insufficient Review

```
Review performed but findings not documented properly

Response:
- GATE VIOLATION: "Code Quality Review Gate - Review documentation insufficient"
- REQUIRE: Complete review documentation in 30-review.md
- REQUIRE: Clear verdict from both Tester and Reviewer
```

---

## Integration with Other Gates

This gate integrates with:

- **[Global Gates](00-global-gates.md)** - Quality baseline, TDD requirement
- **[Verification Gates](30-verification.md)** - Post-implementation verification
- **[Persistence Gates](10-persistence.md)** - Version control integration

**Relationship:**
```
Implementation Complete (Phase 3 exit)
     ↓
Global Gates → Verify tests passing, quality baseline
     ↓
Code Quality Review Gate → Tester + Reviewer validation
     ↓
Verification Gates → Final verification checks
     ↓
Acceptance (Phase 4 exit)
```

---

## Success Criteria

This gate is successful when:

- ✓ All code changes undergo mandatory review
- ✓ Tester validation automatic for every code change
- ✓ Reviewer validation automatic for every code change
- ✓ Blocking issues prevent work completion
- ✓ Quality maintained consistently
- ✓ No quality regressions slip through
- ✓ Team maintains high standards automatically

---

## Examples

### Example 1: Feature Implementation (Both Reviews Required)

**Scenario:** Implement user authentication feature

**Work Package:**
- Code changes: src/auth/login.js, src/auth/session.js
- Test changes: tests/auth/login.test.js, tests/auth/session.test.js

**Review Process:**
```
STEP 1: Worker completes implementation
STEP 2: Worker self-review (tests pass locally, coverage 89%)

STEP 3: Orchestrator delegates to Tester
  Tester Result: "CHANGES REQUIRED"
  Issues:
    - [C1] Coverage only 73% (below 80% threshold)
    - [M1] Error handling tests missing
  Action: Worker adds tests, coverage → 89%
  Re-validation: "APPROVED"

STEP 4: Orchestrator delegates to Reviewer
  Reviewer Result: "CHANGES REQUESTED"
  Issues:
    - [C1] Password stored in plaintext (security)
    - [M1] No input validation on login endpoint
  Action: Worker fixes security issues, adds validation
  Re-validation: "APPROVED"

STEP 5: Gate passes, work proceeds to acceptance
```

---

### Example 2: Bug Fix (Both Reviews Required)

**Scenario:** Fix memory leak in data processing

**Work Package:**
- Code changes: src/processors/dataProcessor.js
- Test changes: tests/processors/dataProcessor.test.js (new regression test)

**Review Process:**
```
STEP 1: Worker completes fix
STEP 2: Worker self-review (regression test passes)

STEP 3: Orchestrator delegates to Tester
  Tester Result: "APPROVED"
  Comments:
    - TDD followed (test first, then fix)
    - Regression test comprehensive
    - Coverage maintained at 91%

STEP 4: Orchestrator delegates to Reviewer
  Reviewer Result: "APPROVED"
  Comments:
    - Root cause correctly identified
    - Fix properly implemented
    - No side effects observed
    - [m1] Consider adding memory monitoring (minor)

STEP 5: Gate passes (both approved), work proceeds to acceptance
```

---

### Example 3: Documentation Only (Reviews Skipped)

**Scenario:** Update README with installation instructions

**Work Package:**
- Documentation changes: README.md, docs/installation.md
- No code changes

**Review Process:**
```
STEP 1: Worker completes documentation
STEP 2: Orchestrator checks for code changes: NONE FOUND

STEP 3: Gate check: Code changes present? NO
  Result: Skip Tester validation (no code/tests changed)
  Result: Skip Reviewer validation (no code changed)

STEP 4: Proceed directly to user acceptance
  (Gate does not trigger for documentation-only changes)
```

---

## References

- **[Tester Role](../roles/tester.md)** - TDD validation and test sufficiency verification
- **[Reviewer Role](../roles/reviewer.md)** - Code quality and standards verification
- **[Orchestrator Role](../roles/orchestrator.md)** - Review coordination responsibilities
- **[Standard Workflow](../workflows/standard.md)** - Phase 4: Review integration
- **[Global Gates](00-global-gates.md)** - Gate 7: Test-Driven Development
- **[Verification Gates](30-verification.md)** - Post-implementation verification

---

**Last reviewed:** 2026-01-08
**Next review:** Quarterly or when review practices evolve
**Status:** ACTIVE - Mandatory enforcement gate
