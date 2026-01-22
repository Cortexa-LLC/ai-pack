---
sidebar_position: 1
title: "Bugfix Workflow"
---

# Bugfix Workflow

**Version:** 1.0.0
**Last Updated:** 2026-01-07

## Overview

The Bugfix Workflow is specialized for identifying, analyzing, and fixing defects. It emphasizes root cause analysis, regression prevention, and thorough verification.

**Extends:** [Standard Workflow](standard.md)

**Use for:** Fixing bugs, addressing defects, resolving errors, correcting unexpected behavior.

---

## Key Differences from Standard Workflow

1. **Root Cause Analysis** - Must understand why bug occurred
2. **Reproduction First** - Must reproduce bug before fixing
3. **Regression Test** - Must add test that would have caught the bug
4. **Prevention Focus** - Look for similar bugs
5. **Fast Track Option** - Critical bugs may skip some gates

---

## Bugfix-Specific Phases

### Phase 1: Bug Triage & Reproduction

**Objective:** Understand, reproduce, and assess the bug.

**DELEGATION STRATEGY:**

**Option A: Delegate to Inspector (Static Code Analysis for Complex Bugs)**
```text
IF bug is complex AND root cause unclear AND reproducible locally THEN
  Orchestrator delegates to Inspector
  Inspector conducts static code investigation and retrospective
  Inspector creates task packet for Engineer
  Orchestrator delegates to Engineer with task packet
END IF
```text

**Option B: Delegate to Spelunker (Runtime Investigation for Production/Live System Issues)**
```text
IF production-only issue OR runtime behavior investigation needed THEN
  Orchestrator delegates to Spelunker
  Spelunker investigates runtime behavior (traces execution, inspects state)
  Spelunker creates runtime report for Engineer
  Orchestrator delegates to Engineer with runtime findings
END IF
```text

**Option C: Hybrid Approach (Both Inspector and Spelunker)**
```text
IF complex issue requires both runtime and static analysis THEN
  Orchestrator delegates to Spelunker (runtime investigation)
  Orchestrator delegates to Inspector (static code analysis)
  Wait for combined findings
  Orchestrator delegates to Engineer with full context
END IF
```text

**Option D: Engineer Self-Investigation (For Simple Bugs)**
```text
IF bug is simple OR root cause obvious THEN
  Orchestrator delegates to Engineer
  Engineer follows bugfix workflow phases 1-4
END IF
```text

**Selection Criteria:**
- **Complex bug with static code analysis → Inspector:**
  - Root cause unknown but likely in code logic
  - Bug reproducible locally
  - Similar bugs may exist in codebase
  - Investigation requires code forensic analysis

- **Production/runtime issue → Spelunker:**
  - Production-only problem (can't reproduce locally)
  - Performance issue requiring profiling
  - Intermittent bug (timing, race conditions, Heisenbugs)
  - Complex distributed system issue
  - Need to understand actual runtime behavior
  - Deep call stack mysteries
  - External integration failures
  - Unfamiliar live system investigation

- **Complex requiring both → Hybrid (Inspector + Spelunker):**
  - Production behavior mysterious AND code analysis needed
  - Runtime findings inform static analysis
  - Static analysis informs runtime investigation

- **Simple/obvious bug → Engineer directly:**
  - Bug is obvious (typo, simple logic error)
  - Root cause immediately apparent
  - Fix is straightforward

---

#### 1.1 Bug Understanding
```text
□ What is the expected behavior?
□ What is the actual behavior?
□ How to reproduce?
□ What are the symptoms?
□ When did it start occurring?
□ How many users affected?
```text

**Severity Classification:**
```text
CRITICAL: System down, data loss, security breach
  → Immediate fix required
  → May skip some gates

MAJOR: Core functionality broken, many users affected
  → Fix within 24 hours
  → Fast-track through workflow

MINOR: Edge case, workaround exists, few users affected
  → Fix in normal cycle
  → Follow standard workflow
```text

---

#### 1.2 Bug Reproduction
```text
CRITICAL: Must reproduce bug before fixing

Reproduction Steps:
1. Create minimal test case
2. Document exact steps
3. Identify conditions required
4. Note any workarounds
5. Verify reproduction consistent

Write Failing Test:
□ Test that demonstrates the bug
□ Test should fail before fix
□ Test should pass after fix
```text

**Reproduction Script Example:**
```javascript
// Bug: Division by zero in calculation
describe('BUG-123: Division by zero crash', () => {
  it('should handle zero denominator gracefully', () => {
    const result = calculate(10, 0);  // Currently crashes
    expect(result).toBeNull();         // Should return null
  });
});
```text

---

#### 1.3 Root Cause Analysis
```text
Investigation Questions:
□ Where is the bug in the code?
□ Why did it occur?
□ Why wasn't it caught by tests?
□ Are there similar bugs elsewhere?
□ What conditions trigger it?

Tools for Investigation:
- Debuggers
- Log analysis
- Stack traces
- Git blame (when introduced?)
- Code review
```text

**Root Cause Documentation:**
```text
Bug: [Summary]
Location: [file:line]
Root Cause: [Why it happened]
Contributing Factors: [What enabled it]
Why Missed: [Why tests didn't catch it]
```text

---

### Phase 2: Fix Strategy & Impact Assessment

**Objective:** Design fix that addresses root cause without side effects.

#### 2.1 Fix Strategy Selection
```text
Options:
1. Minimal fix - Address immediate issue only
2. Comprehensive fix - Address root cause properly
3. Workaround - Temporary mitigation

Selection Criteria:
□ Risk of side effects
□ Time sensitivity
□ Code complexity
□ Test coverage
□ Long-term maintenance

Recommend: Comprehensive fix unless critical time pressure
```text

---

#### 2.2 Impact Assessment
```text
□ What code is affected by fix?
□ Any breaking changes?
□ Performance implications?
□ Need database changes?
□ Affects other features?
□ Backward compatibility?
```text

**Risk Assessment:**
```text
Low Risk:
- Isolated change
- Well-tested area
- No dependencies
- Easy rollback

High Risk:
- Core functionality
- Many dependencies
- Complex interaction
- Difficult rollback

→ High risk fixes require extra testing
```text

---

### Phase 3: Fix Implementation & Testing

**Objective:** Fix bug and prove it won't recur.

#### 3.1 Implementation Approach
```text
Test-Driven Bug Fix:
1. Write failing test (reproduces bug) → RED
2. Implement minimal fix → GREEN
3. Verify test now passes
4. Add edge case tests
5. Refactor for quality
6. Run full test suite
7. Verify no regressions
```text

#### 3.2 Regression Prevention
```text
Required Tests:
□ Test that reproduces original bug
□ Tests for related edge cases
□ Tests for similar scenarios
□ Integration tests affected path

Test should:
✓ Fail before fix applied
✓ Pass after fix applied
✓ Prevent bug from recurring
✓ Be maintainable
```text

---

#### 3.3 Similar Bug Prevention
```text
After fixing bug, check for similar issues:

□ Search codebase for similar patterns
□ Review related functionality
□ Check for copy-paste code
□ Verify consistent error handling
□ Update conventions if pattern emerges
```text

**Example:**
```javascript
// Found bug: Array access without bounds check
// FIX:
if (index >= 0 && index < array.length) {
  return array[index];
}

// NOW SEARCH: Find all array accesses
// Add bounds checks where missing
```text

---

### Phase 4: Verification & Documentation

**Objective:** Prove bug is fixed and document findings.

#### 4.1 Fix Verification
```text
□ Original test case passes
□ All new tests pass
□ All existing tests pass (no regressions)
□ Manual testing confirms fix
□ Edge cases handled
□ Original reporter confirms (if possible)
```text

---

#### 4.2 Documentation of Fix
```text
Required Documentation:
□ Bug description and root cause
□ Fix approach and rationale
□ Tests added
□ Related changes made
□ Known limitations (if any)

Commit Message:
fix: [short description of bug]

- Root cause: [why bug occurred]
- Fix: [what was changed]
- Tests: [tests added]
- Impact: [who was affected]

Fixes #[issue-number]
```text

---

#### 4.3 Lessons Learned
```text
Document for team knowledge:
□ Why bug occurred
□ How to prevent similar bugs
□ Gaps in test coverage
□ Process improvements needed
□ Convention updates needed
```text

---

## Fast-Track for Critical Bugs

For CRITICAL severity bugs only:

### Expedited Workflow
```text
1. Reproduce (required, no shortcut)
2. Assess impact (quick assessment)
3. Implement hotfix
4. Test fix directly
5. Deploy immediately
6. Monitor closely
7. Follow up with proper fix later if hotfix was workaround
```text

### Critical Bug Gates
```text
✓ Bug reproduced
✓ Fix tested
✓ No worse than current state
✓ Rollback plan ready

Can defer:
- Comprehensive testing (do after deploy)
- Full root cause analysis (do after fix)
- Perfect solution (hotfix acceptable)
```text

---

## Common Bug Patterns

### Off-by-One Errors
```javascript
// ❌ Bug
for (let i = 0; i <= array.length; i++) { ... }

// ✅ Fix
for (let i = 0; i < array.length; i++) { ... }
```text

### Null/Undefined Handling
```javascript
// ❌ Bug
const name = user.profile.name;  // Crashes if profile is null

// ✅ Fix
const name = user?.profile?.name ?? 'Unknown';
```text

### Race Conditions
```javascript
// ❌ Bug
async function loadData() {
  data = await fetch('/api/data');  // Multiple calls overwrite
}

// ✅ Fix
async function loadData() {
  if (loading) return;
  loading = true;
  data = await fetch('/api/data');
  loading = false;
}
```text

---

## Bugfix-Specific Checklist

### Before Declaring Fixed
```text
□ Bug reproduced successfully
□ Root cause identified
□ Fix implemented
□ Regression test added
□ All tests passing
□ Similar bugs checked
□ No new issues introduced
□ Original reporter notified (if applicable)
□ Documentation updated
□ Lessons learned documented
```text

---

## Example Bugfix: Login Timeout Issue

### Phase 1: Triage & Reproduction
```text
Bug Report:
Users getting logged out after 5 minutes of inactivity

Expected: 30-minute session timeout
Actual: 5-minute timeout

Reproduction:
1. Login to application
2. Leave idle for 5 minutes
3. Try to interact
4. Result: Logged out

Root Cause Analysis:
- Session timeout hardcoded to 300 seconds (5 min)
- Configuration file has 1800 (30 min) but not read
- Bug introduced in commit abc123 (refactoring)
```text

### Phase 2: Fix Strategy
```text
Strategy: Fix configuration reading

Impact Assessment:
- Low risk (isolated config loading)
- No breaking changes
- Easy rollback
- Affects all users (but positively)

Fix Plan:
1. Repair configuration file reading
2. Add test for configuration loading
3. Add logging for timeout value
4. Verify with various config values
```text

### Phase 3: Implementation
```javascript
// Before (bug):
const SESSION_TIMEOUT = 300;

// After (fix):
const SESSION_TIMEOUT = config.get('sessionTimeout', 1800);

// Test added:
describe('Session timeout configuration', () => {
  it('should read timeout from config file', () => {
    const timeout = loadConfig().sessionTimeout;
    expect(timeout).toBe(1800);
  });

  it('should fall back to default if config missing', () => {
    const timeout = loadConfig({}).sessionTimeout;
    expect(timeout).toBe(1800);
  });
});
```text

### Phase 4: Verification
```text
✓ Configuration now loaded correctly
✓ Tests pass (config loading verified)
✓ Manual testing: 30-minute timeout works
✓ No regressions in other functionality
✓ Logging confirms correct timeout value

Deploy: Rolled out to production
Result: Issue resolved, no recurrence
```text

---

## Post-Fix: Retrospective Artifact Persistence

**CRITICAL:** After bug is fixed and verified, bug investigation retrospective MUST be persisted to repository for organizational learning.

### When to Persist Retrospective

```text
WHEN bug fix verified and accepted THEN
  IF Inspector was used THEN
    Inspector persists retrospective to docs/investigations/
  ELSE IF Engineer did investigation THEN
    Engineer (or Orchestrator) persists retrospective to docs/investigations/
  END IF

  persist: Retrospective document with lessons learned
  commit: Add to investigation knowledge base
  see roles/inspector.md "Artifact Persistence" section
END
```text

### Why This Matters

**Organizational Learning:**
- Captures knowledge about system failure modes
- Patterns emerge across multiple bug investigations
- Prevents repeating same bugs
- Informs architecture improvements

**Pattern Detection:**
- Categorized retrospectives reveal systemic issues
- Multiple similar bugs → systemic improvement needed
- Retrospective index helps diagnose similar symptoms faster

**Repository Structure After Bugfix:**
```text
project-root/
├── docs/
│   ├── investigations/
│   │   ├── BUG-123-null-pointer-in-payment.md
│   │   ├── BUG-145-race-condition-order-processing.md
│   │   └── README.md (index by root cause category)
│   └── ...
└── .ai/
    └── tasks/ (temporary work-in-progress)
```text

**Communication Pattern:**
```text
"Bug fix complete and retrospective committed to repository.

Location: docs/investigations/[bug-id]-[description].md

Root Cause: [Brief explanation]
Category: [Pattern category]
Lessons Learned: [Summary]

This retrospective is now part of the organizational knowledge base."
```text

---

## Success Criteria

A bugfix is complete when:
```text
✓ Bug reproduced and understood
✓ Root cause identified
✓ Fix implemented and tested
✓ Regression test added
✓ All tests passing
✓ No new bugs introduced
✓ Bug verified fixed
✓ Documentation complete
✓ Lessons learned captured
✓ Retrospective persisted to docs/investigations/ (for non-trivial bugs)
```text

---

## References

- [Standard Workflow](standard.md)
- [Refactoring Guide](../quality/clean-code/03-refactoring.md)
- [Testing Guidelines](../quality/clean-code/04-testing.md)
- [Verification Gates](../gates/30-verification.md)

---

**Last reviewed:** 2026-01-07
**Next review:** Quarterly or when bugfix practices evolve
