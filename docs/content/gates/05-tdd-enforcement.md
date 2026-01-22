---
sidebar_position: 3
title: "TDD Enforcement (MANDATORY, BLOCKING)"
---

# TDD Enforcement Gate (MANDATORY, BLOCKING)

**Version:** 1.0.0
**Last Updated:** 2026-01-14

## Overview

Test-Driven Development (TDD) is a MANDATORY, ENFORCED requirement for all code changes in ai-pack. This gate ensures the RED-GREEN-REFACTOR cycle is followed without exception.

**Severity:** BLOCKING - Work cannot proceed without TDD compliance

---

## Core Principle

> "The act of writing a unit test is more an act of design than of verification. It is also more an act of documentation than of verification. The act of writing a unit test closes a remarkable number of feedback loops, the least of which is the one pertaining to verification of function."
> — Robert C. Martin, *The Clean Coder*

**TDD is not optional. It is not a suggestion. It is MANDATORY and ENFORCED.**

---

## The TDD Cycle (MANDATORY)

### RED-GREEN-REFACTOR

All code MUST follow this cycle:

```text
1. RED: Write a failing test
   ↓
2. GREEN: Write minimal code to make it pass
   ↓
3. REFACTOR: Clean up while keeping tests green
   ↓
   (Repeat for next requirement)
```text

**This cycle is BLOCKING. Engineers cannot skip it.**

---

## Gate Requirements (BLOCKING)

### Requirement 1: Tests Written BEFORE Implementation

**Rule:** Tests MUST exist and MUST be committed BEFORE the implementation code.

**Verification:**
```text
FOR each feature or bug fix:
  REQUIRE:
    ✓ Test file created/modified BEFORE implementation file
    ✓ Test commit timestamp < implementation commit timestamp
    ✓ Git history shows test-first pattern
    ✓ Tests define expected behavior clearly

  IF tests written after implementation THEN
    BLOCK approval
    REJECT work
    REQUIRE: Rewrite with TDD approach
  END IF
END FOR
```text

**Evidence Required:**
- Git commit history showing tests before implementation
- Test files with earlier timestamps than implementation
- Commit messages indicating TDD cycle (e.g., "Add failing test for...", "Make test pass", "Refactor...")

---

### Requirement 2: RED Phase Evidence

**Rule:** Tests MUST initially fail (RED phase) before implementation.

**Verification:**
```text
FOR each test:
  REQUIRE:
    ✓ Test initially failed when first committed
    ✓ Failure was intentional (not due to bugs)
    ✓ Failure message indicates unimplemented feature
    ✓ Test defines precise expected behavior

  IF no evidence of RED phase THEN
    BLOCK approval
    REJECT work
    REQUIRE: Demonstrate TDD cycle
  END IF
END FOR
```text

**Evidence Required:**
- Commit showing test that fails
- Test output or commit message indicating failure
- Implementation commit that makes test pass

---

### Requirement 3: Test Pyramid Adherence

**Rule:** Tests MUST follow proper test pyramid distribution (Fowler, Cohn).

**Pyramid Structure:**
```text
        /\
       /E2E\      ← Very few (5-10%)
      /------\
     / Integ  \   ← Some (15-25%)
    /----------\
   /   Unit     \  ← Lots (65-80%)
  /--------------\
```text

**Verification:**
```text
WHEN analyzing test suite:
  CALCULATE:
    unit_percentage = unit_tests / total_tests * 100
    integration_percentage = integration_tests / total_tests * 100
    e2e_percentage = e2e_tests / total_tests * 100

  REQUIRE:
    ✓ unit_percentage >= 65% AND unit_percentage <= 80%
    ✓ integration_percentage >= 15% AND integration_percentage <= 25%
    ✓ e2e_percentage >= 5% AND e2e_percentage <= 10%

  IF pyramid inverted OR disproportionate THEN
    WARN about test distribution
    REQUEST: Rebalance test suite
  END IF
END WHEN
```text

**Reference:** [Practical Test Pyramid](https://martinfowler.com/articles/practical-test-pyramid.html) - Martin Fowler

---

### Requirement 4: Coverage Targets

**Rule:** Code coverage MUST meet minimum thresholds.

**Coverage Requirements:**
```text
MANDATORY Targets:
  ✓ Overall coverage: 80-90% (BLOCKING if < 80%)
  ✓ Critical business logic: 95%+ (BLOCKING if < 95%)
  ✓ Error handling paths: 90%+ (BLOCKING if < 90%)
  ✓ Integration points: 100% (BLOCKING if < 100%)
  ✓ Public APIs: 100% (BLOCKING if < 100%)

IF coverage < target THEN
  BLOCK approval
  REJECT work
  REQUIRE: Write missing tests
END IF
```text

**Reference:** [Test Coverage](https://martinfowler.com/bliki/TestCoverage.html) - Martin Fowler

---

## Enforcement Mechanisms

### Engineer Role (MANDATORY)

Engineers MUST follow TDD cycle:

```text
BEFORE writing any implementation code:
  STEP 1: Write failing test (RED)
    - Create test that fails
    - Commit test with message: "Add failing test for [feature]"
    - Verify test fails

  STEP 2: Write minimal implementation (GREEN)
    - Write simplest code to pass test
    - Commit with message: "Make [feature] test pass"
    - Verify all tests pass

  STEP 3: Refactor (REFACTOR)
    - Clean up code while keeping tests green
    - Commit with message: "Refactor [feature]"
    - Verify all tests still pass

  IF Engineer skips TDD THEN
    Tester BLOCKS approval
    Work is REJECTED
    Engineer must redo with TDD
  END IF
END BEFORE
```text

---

### Tester Role (MANDATORY, BLOCKING)

Tester MUST validate TDD compliance:

```text
WHEN validating work:
  STEP 1: Verify TDD Process
    - Check git history for test-first pattern
    - Verify RED phase (tests initially failed)
    - Verify GREEN phase (tests now pass)
    - Verify REFACTOR phase (code cleaned up)

  STEP 2: Check Test Pyramid
    - Analyze test distribution
    - Verify proper pyramid structure
    - Check for inverted pyramid anti-pattern

  STEP 3: Validate Coverage
    - Run coverage tools
    - Check against mandatory targets
    - Identify critical gaps

  IF TDD not followed OR coverage insufficient THEN
    STATUS = "CHANGES REQUIRED"
    BLOCK approval
    REJECT work
    DOCUMENT violations in 30-review.md
    REQUIRE: Engineer redo with proper TDD
  ELSE
    STATUS = "APPROVED"
    Allow work to proceed
  END IF
END WHEN
```text

**Tester verdict MUST be:**
- **"APPROVED"** - TDD followed, coverage met, tests sufficient
- **"CHANGES REQUIRED"** - TDD violated, coverage insufficient, or test quality poor

**No middle ground. No exceptions.**

---

### Orchestrator Role (MANDATORY)

Orchestrator MUST enforce TDD delegation:

```text
BEFORE delegating to Engineer:
  REMIND: "TDD is MANDATORY. Follow RED-GREEN-REFACTOR cycle."

AFTER Engineer completes:
  REQUIRE: Tester validation

  IF Tester verdict == "CHANGES REQUIRED" THEN
    BLOCK task completion
    REQUIRE: Engineer fix violations
    RE-DELEGATE to Tester for re-validation
  END IF

  Task is NOT complete until Tester approves
END
```text

---

## Anti-Patterns (BLOCKING VIOLATIONS)

### ❌ Tests Written After Implementation

**Violation:**
```text
git log:
  commit abc123: "Implement user authentication"  ← Implementation FIRST
  commit def456: "Add tests for authentication"    ← Tests SECOND
```text

**Why This Fails:** This is NOT TDD. Tests written after lose the design benefit of TDD.

**Enforcement:** BLOCKED by Tester. Work REJECTED.

---

### ❌ No RED Phase Evidence

**Violation:**
```text
git log:
  commit abc123: "Add authentication tests"  ← Tests added
  (no evidence tests ever failed)
  commit def456: "Implement authentication"  ← Implementation
```text

**Why This Fails:** No proof of RED phase. Tests may have been written to match existing code.

**Enforcement:** BLOCKED by Tester. Work REJECTED.

---

### ❌ Inverted Test Pyramid

**Violation:**
```text
Test Distribution:
  Unit tests: 20%        ← Too few!
  Integration tests: 30%
  E2E tests: 50%         ← Too many!
```text

**Why This Fails:** Slow, brittle tests. Violates Fowler's test pyramid guidance.

**Enforcement:** BLOCKED by Tester. Request rebalancing.

---

### ❌ Insufficient Coverage

**Violation:**
```text
Coverage Report:
  Overall: 65%           ← BLOCKING (< 80%)
  Critical logic: 88%    ← BLOCKING (< 95%)
```text

**Why This Fails:** Does not meet mandatory coverage targets.

**Enforcement:** BLOCKED by Tester. Work REJECTED.

---

## Consequences of Violations

### Immediate Consequences

```text
IF TDD not followed THEN
  Tester verdict = "CHANGES REQUIRED"
  Work status = "BLOCKED"
  Task marked = "INCOMPLETE"

  Engineer MUST:
    1. Revert implementation
    2. Start over with proper TDD cycle
    3. Write tests FIRST
    4. Follow RED-GREEN-REFACTOR
    5. Re-submit for Tester validation

  Work cannot proceed until compliant
END IF
```text

### No Exceptions

**There are NO exceptions to TDD requirements:**
- ❌ "It's a small change" - Still requires TDD
- ❌ "Just a bug fix" - Still requires TDD
- ❌ "Time pressure" - Still requires TDD
- ❌ "Legacy code" - Still requires TDD for new changes
- ❌ "Prototyping" - Use spikes, then implement with TDD

**TDD is not negotiable.**

---

## Acceptable Exceptions (Very Limited)

Only these scenarios MAY skip TDD (with documentation):

### 1. Spike Solutions (Exploratory Code)

```text
IF work type == "spike" OR "proof of concept" THEN
  TDD may be skipped
  REQUIRE:
    - Clearly marked as spike in task packet
    - Code not intended for production
    - Will be discarded after exploration
    - Production version will use TDD
END IF
```text

### 2. Generated Code

```text
IF code is machine-generated THEN
  TDD may not apply directly
  REQUIRE:
    - Mark as generated in comments
    - Generator itself must be tested
    - Generated output spot-checked
END IF
```text

### 3. Legacy Code Characterization Tests

```text
IF adding tests to legacy code without tests THEN
  Tests may be written after (characterization tests)
  REQUIRE:
    - Clearly documented as characterization
    - Covers existing behavior
    - Future changes use TDD
END IF
```text

**All exceptions MUST be documented in task packet and approved by Reviewer.**

---

## Success Criteria

Work is TDD-compliant when:

- ✓ Git history shows test-first commits
- ✓ Evidence of RED phase (initial test failures)
- ✓ Evidence of GREEN phase (tests passing)
- ✓ Evidence of REFACTOR phase (code improvements)
- ✓ Test pyramid properly structured (65-80% unit, 15-25% integration, 5-10% e2e)
- ✓ Coverage meets mandatory targets (80%+ overall, 95%+ critical)
- ✓ All tests pass
- ✓ Test quality is high (clear, fast, independent)
- ✓ Tester verdict == "APPROVED"

**If any criterion fails, work is BLOCKED.**

---

## References

**Martin Fowler (martinfowler.com):**
- [Practical Test Pyramid](https://martinfowler.com/articles/practical-test-pyramid.html)
- [Test Coverage](https://martinfowler.com/bliki/TestCoverage.html)
- [Unit Test](https://martinfowler.com/bliki/UnitTest.html)
- [Test Double](https://martinfowler.com/bliki/TestDouble.html)
- [Test-Driven Development](https://martinfowler.com/bliki/TestDrivenDevelopment.html)

**Robert C. Martin:**
- *The Clean Coder* - Chapter 5: Test-Driven Development
- *Clean Code* - Chapter 9: Unit Tests

**Kent Beck:**
- *Test-Driven Development: By Example*

**Mike Cohn:**
- *Succeeding with Agile* - Test Pyramid concept

---

## Implementation Checklist

**For Engineers:**
- [ ] Understand TDD is MANDATORY
- [ ] Write test BEFORE implementation
- [ ] Verify test fails (RED)
- [ ] Write minimal code to pass (GREEN)
- [ ] Refactor while keeping tests green (REFACTOR)
- [ ] Commit shows clear TDD cycle
- [ ] Test pyramid properly structured
- [ ] Coverage meets targets

**For Testers:**
- [ ] Verify git history shows TDD pattern
- [ ] Check for RED phase evidence
- [ ] Validate test pyramid structure
- [ ] Run coverage analysis
- [ ] Check against mandatory targets
- [ ] BLOCK if TDD not followed
- [ ] Document verdict in 30-review.md

**For Orchestrators:**
- [ ] Remind Engineers of TDD requirement
- [ ] Require Tester validation
- [ ] BLOCK completion if Tester rejects
- [ ] Enforce re-work if violations found

---

**Last reviewed:** 2026-01-14
**Next review:** When TDD practices evolve or violations increase
