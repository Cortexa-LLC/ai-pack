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

#### 1.1 Bug Understanding
```
□ What is the expected behavior?
□ What is the actual behavior?
□ How to reproduce?
□ What are the symptoms?
□ When did it start occurring?
□ How many users affected?
```

**Severity Classification:**
```
CRITICAL: System down, data loss, security breach
  → Immediate fix required
  → May skip some gates

MAJOR: Core functionality broken, many users affected
  → Fix within 24 hours
  → Fast-track through workflow

MINOR: Edge case, workaround exists, few users affected
  → Fix in normal cycle
  → Follow standard workflow
```

---

#### 1.2 Bug Reproduction
```
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
```

**Reproduction Script Example:**
```javascript
// Bug: Division by zero in calculation
describe('BUG-123: Division by zero crash', () => {
  it('should handle zero denominator gracefully', () => {
    const result = calculate(10, 0);  // Currently crashes
    expect(result).toBeNull();         // Should return null
  });
});
```

---

#### 1.3 Root Cause Analysis
```
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
```

**Root Cause Documentation:**
```
Bug: [Summary]
Location: [file:line]
Root Cause: [Why it happened]
Contributing Factors: [What enabled it]
Why Missed: [Why tests didn't catch it]
```

---

### Phase 2: Fix Strategy & Impact Assessment

**Objective:** Design fix that addresses root cause without side effects.

#### 2.1 Fix Strategy Selection
```
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
```

---

#### 2.2 Impact Assessment
```
□ What code is affected by fix?
□ Any breaking changes?
□ Performance implications?
□ Need database changes?
□ Affects other features?
□ Backward compatibility?
```

**Risk Assessment:**
```
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
```

---

### Phase 3: Fix Implementation & Testing

**Objective:** Fix bug and prove it won't recur.

#### 3.1 Implementation Approach
```
Test-Driven Bug Fix:
1. Write failing test (reproduces bug) → RED
2. Implement minimal fix → GREEN
3. Verify test now passes
4. Add edge case tests
5. Refactor for quality
6. Run full test suite
7. Verify no regressions
```

#### 3.2 Regression Prevention
```
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
```

---

#### 3.3 Similar Bug Prevention
```
After fixing bug, check for similar issues:

□ Search codebase for similar patterns
□ Review related functionality
□ Check for copy-paste code
□ Verify consistent error handling
□ Update conventions if pattern emerges
```

**Example:**
```javascript
// Found bug: Array access without bounds check
// FIX:
if (index >= 0 && index < array.length) {
  return array[index];
}

// NOW SEARCH: Find all array accesses
// Add bounds checks where missing
```

---

### Phase 4: Verification & Documentation

**Objective:** Prove bug is fixed and document findings.

#### 4.1 Fix Verification
```
□ Original test case passes
□ All new tests pass
□ All existing tests pass (no regressions)
□ Manual testing confirms fix
□ Edge cases handled
□ Original reporter confirms (if possible)
```

---

#### 4.2 Documentation of Fix
```
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
```

---

#### 4.3 Lessons Learned
```
Document for team knowledge:
□ Why bug occurred
□ How to prevent similar bugs
□ Gaps in test coverage
□ Process improvements needed
□ Convention updates needed
```

---

## Fast-Track for Critical Bugs

For CRITICAL severity bugs only:

### Expedited Workflow
```
1. Reproduce (required, no shortcut)
2. Assess impact (quick assessment)
3. Implement hotfix
4. Test fix directly
5. Deploy immediately
6. Monitor closely
7. Follow up with proper fix later if hotfix was workaround
```

### Critical Bug Gates
```
✓ Bug reproduced
✓ Fix tested
✓ No worse than current state
✓ Rollback plan ready

Can defer:
- Comprehensive testing (do after deploy)
- Full root cause analysis (do after fix)
- Perfect solution (hotfix acceptable)
```

---

## Common Bug Patterns

### Off-by-One Errors
```javascript
// ❌ Bug
for (let i = 0; i <= array.length; i++) { ... }

// ✅ Fix
for (let i = 0; i < array.length; i++) { ... }
```

### Null/Undefined Handling
```javascript
// ❌ Bug
const name = user.profile.name;  // Crashes if profile is null

// ✅ Fix
const name = user?.profile?.name ?? 'Unknown';
```

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
```

---

## Bugfix-Specific Checklist

### Before Declaring Fixed
```
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
```

---

## Example Bugfix: Login Timeout Issue

### Phase 1: Triage & Reproduction
```
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
```

### Phase 2: Fix Strategy
```
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
```

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
```

### Phase 4: Verification
```
✓ Configuration now loaded correctly
✓ Tests pass (config loading verified)
✓ Manual testing: 30-minute timeout works
✓ No regressions in other functionality
✓ Logging confirms correct timeout value

Deploy: Rolled out to production
Result: Issue resolved, no recurrence
```

---

## Success Criteria

A bugfix is complete when:
```
✓ Bug reproduced and understood
✓ Root cause identified
✓ Fix implemented and tested
✓ Regression test added
✓ All tests passing
✓ No new bugs introduced
✓ Bug verified fixed
✓ Documentation complete
✓ Lessons learned captured
```

---

## References

- [Standard Workflow](standard.md)
- [Refactoring Guide](../quality/clean-code/03-refactoring.md)
- [Testing Guidelines](../quality/clean-code/04-testing.md)
- [Verification Gates](../gates/30-verification.md)

---

**Last reviewed:** 2026-01-07
**Next review:** Quarterly or when bugfix practices evolve
