# Bug Investigation Retrospective: [Bug Title]

**Bug ID:** BUG-YYYY-NNN
**Severity:** [Critical | High | Medium | Low]
**Status:** [Fixed | In Progress | Won't Fix | Duplicate]
**Date Reported:** YYYY-MM-DD
**Date Fixed:** YYYY-MM-DD
**Time to Fix:** [X days]

---

## Executive Summary

[1-2 sentence summary of the bug, root cause, and fix]

---

## Bug Report

### Original Description

[Copy of the original bug report]

**Reporter:** [Name/Team]
**Environment:** [Production | Staging | Development]
**Affected Version:** [Version number]

### Symptoms

**User-Visible Behavior:**
- [What the user experiences]
- [Error messages shown]
- [Unexpected behavior]

**System Behavior:**
- [What the system actually does]
- [Errors in logs]
- [Metrics/monitoring observations]

### Impact

**User Impact:**
- **Affected Users:** [Number/percentage]
- **Frequency:** [Always | Sometimes | Rare]
- **Workaround Available:** [Yes/No - Description]

**Business Impact:**
- [Revenue/reputation/compliance impact]

---

## Reproduction

### Steps to Reproduce

**Environment Setup:**
```bash
# Required environment
[Setup commands]
```

**Reproduction Steps:**
1. [Specific step 1]
2. [Specific step 2]
3. [Specific step 3]

**Expected Result:**
[What should happen]

**Actual Result:**
[What actually happens]

### Reproducibility

- [ ] **100% reproducible** - Always happens
- [ ] **Intermittent** - Sometimes happens ([X]% of the time)
- [ ] **Race condition** - Timing-dependent
- [ ] **Environment-specific** - Only in [specific condition]

**Conditions Required:**
- [Condition 1 that must be true]
- [Condition 2]

### Reproduction Test Case

```language
// Test that demonstrates the bug
test("reproduces bug [BUG-ID]", () => {
  // Setup
  const data = setupTestData();

  // Execute
  const result = executeAction(data);

  // Assert (this should fail before fix, pass after)
  expect(result).toBe(expectedValue);
});
```

---

## Root Cause Analysis

### What Went Wrong?

[Detailed technical explanation of the root cause]

**Root Cause Category:**
- [ ] Logic Error
- [ ] Null/Undefined Reference
- [ ] Off-by-one Error
- [ ] Race Condition
- [ ] Memory Leak
- [ ] Resource Exhaustion
- [ ] Configuration Error
- [ ] Integration Issue
- [ ] Data Validation Missing
- [ ] Edge Case Not Handled
- [ ] Incorrect Assumption
- [ ] Other: [Specify]

### Why Did It Happen?

**5 Whys Analysis:**
1. **Why** did the symptom occur?
   - [Answer]
2. **Why** did [answer 1] happen?
   - [Answer]
3. **Why** did [answer 2] happen?
   - [Answer]
4. **Why** did [answer 3] happen?
   - [Answer]
5. **Why** did [answer 4] happen?
   - **ROOT CAUSE:** [Final answer]

### Technical Details

**Code Location:**
- **File:** `path/to/buggy/file.ext`
- **Lines:** [Line numbers]
- **Function/Method:** `buggyFunction()`
- **Commit Introduced:** [git commit hash]

**Problematic Code:**
```language
// BEFORE (Buggy)
function calculateTotal(items) {
  let total = 0;
  for (let i = 0; i <= items.length; i++) {  // Bug: <= should be <
    total += items[i].price;  // Crashes on items.length
  }
  return total;
}
```

**Why This Code Was Written:**
[Explain the original intent and how the bug was introduced]

**When Introduced:**
- **Commit:** [hash]
- **Date:** YYYY-MM-DD
- **Context:** [What was being implemented]
- **PR/Review:** [Link if available]

### System Context

**Data Flow:**
```
User Input → Validation → Processing → Database → Response
                         ^^^^^^^^^^^
                         Bug occurs here
```

**State at Bug Occurrence:**
```
[Description of system state when bug manifests]
```

**Dependencies Involved:**
- [Dependency 1 and its role]
- [Dependency 2 and its role]

---

## The Fix

### Solution Approach

**Fix Strategy:**
- [ ] **Minimal Fix** - Address immediate symptom
- [ ] **Comprehensive Fix** - Address root cause completely
- [ ] **Refactoring** - Restructure to prevent similar bugs

**Rationale:**
[Why this approach was chosen]

### Implementation

**Fixed Code:**
```language
// AFTER (Fixed)
function calculateTotal(items) {
  let total = 0;
  for (let i = 0; i < items.length; i++) {  // Fixed: < instead of <=
    total += items[i].price;
  }
  return total;
}
```

**Changes Made:**
1. [Change 1]
2. [Change 2]
3. [Change 3]

**Files Modified:**
- `path/to/file1.ext` - [What changed]
- `path/to/file2.ext` - [What changed]

**Commit:** [git commit hash]
**PR:** [Link to pull request]

### Testing

**Regression Test:**
```language
// Test to prevent recurrence
test("calculates total correctly", () => {
  const items = [
    { price: 10 },
    { price: 20 },
    { price: 30 }
  ];

  const result = calculateTotal(items);

  expect(result).toBe(60);
});

test("handles empty array", () => {
  const result = calculateTotal([]);
  expect(result).toBe(0);
});

test("handles single item", () => {
  const items = [{ price: 10 }];
  const result = calculateTotal(items);
  expect(result).toBe(10);
});
```

**Test Coverage:**
- [ ] Unit tests added
- [ ] Integration tests added
- [ ] Edge cases covered
- [ ] Regression test added

**Manual Testing:**
- [ ] Tested in development
- [ ] Tested in staging
- [ ] Verified in production

---

## Prevention

### How to Prevent Similar Bugs

**Category Pattern:** [Pattern name - e.g., "Off-by-one errors in loops"]

**Prevention Strategies:**
1. [Strategy 1 - e.g., "Use forEach/map instead of manual iteration"]
2. [Strategy 2 - e.g., "Add linting rule for loop conditions"]
3. [Strategy 3 - e.g., "Include empty array tests"]

### Code Review Checklist Addition

Add these checks for future code reviews:
- [ ] [Check 1]
- [ ] [Check 2]
- [ ] [Check 3]

### Static Analysis Rules

**Linting Rules to Add:**
```javascript
// ESLint rule example
{
  "rules": {
    "no-unsafe-loop-condition": "error"
  }
}
```

**Type Checking Improvements:**
[What type annotations or checks would have caught this?]

### Testing Improvements

**Test Cases Previously Missing:**
- [ ] Empty input test
- [ ] Boundary condition test
- [ ] Null/undefined test
- [ ] Large input test
- [ ] Concurrent access test

### Architectural Considerations

**Design Patterns to Apply:**
- [Pattern that would prevent this]

**Refactoring Opportunities:**
- [Larger structural changes to consider]

---

## Related Bugs

### Similar Bugs in Codebase

- [BUG-YYYY-NNN](./BUG-YYYY-NNN-description.md) - Similar off-by-one error
- [BUG-YYYY-NNN](./BUG-YYYY-NNN-description.md) - Related issue

### Pattern Analysis

**Frequency:** [Xth occurrence of this bug category]

**Common Characteristics:**
- [Characteristic 1]
- [Characteristic 2]

**Systemic Issue?**
[Is this revealing a deeper problem with our architecture, process, or team knowledge?]

---

## Lessons Learned

### Technical Lessons

**What We Learned:**
- [Learning 1]
- [Learning 2]
- [Learning 3]

**Knowledge Gaps Identified:**
- [Gap 1 - What the team didn't know]
- [Gap 2]

### Process Lessons

**Why Wasn't This Caught Earlier?**
- **In Development:** [Why developer didn't catch it]
- **In Code Review:** [Why reviewer didn't catch it]
- **In Testing:** [Why tests didn't catch it]
- **In Staging:** [Why staging didn't catch it]

**Process Improvements:**
- [ ] Update code review checklist
- [ ] Add automated test requirement
- [ ] Improve linting configuration
- [ ] Update documentation
- [ ] Schedule training

### Impact on System Understanding

**New Insights About the System:**
- [What we now understand that we didn't before]

**Documentation Updates Needed:**
- [ ] Update API documentation
- [ ] Update architecture diagram
- [ ] Update developer guide
- [ ] Create ADR for [decision]

---

## Timeline

| Date | Event | Owner |
|------|-------|-------|
| YYYY-MM-DD | Bug introduced in [commit] | [Name] |
| YYYY-MM-DD | Bug deployed to production | [Name] |
| YYYY-MM-DD | Bug reported by [source] | [Reporter] |
| YYYY-MM-DD | Investigation started (Inspector) | [Name] |
| YYYY-MM-DD | Root cause identified | [Name] |
| YYYY-MM-DD | Fix implemented | [Name] |
| YYYY-MM-DD | Fix deployed to production | [Name] |
| YYYY-MM-DD | Fix verified | [Name] |
| YYYY-MM-DD | Retrospective documented | [Name] |

**Time Metrics:**
- **Time in Production:** [X days from deploy to discovery]
- **Time to Identify:** [X hours from report to root cause]
- **Time to Fix:** [X hours from root cause to fix deployed]
- **Total Resolution Time:** [X days from report to verified fix]

---

## Action Items

### Immediate Actions (Done)
- [x] Fix deployed
- [x] Regression test added
- [x] Documentation updated

### Follow-up Actions
- [ ] [Action 1] - Owner: [Name] - Due: YYYY-MM-DD
- [ ] [Action 2] - Owner: [Name] - Due: YYYY-MM-DD
- [ ] [Action 3] - Owner: [Name] - Due: YYYY-MM-DD

### Long-term Improvements
- [ ] [Improvement 1] - Owner: [Name] - Target: Q[N]
- [ ] [Improvement 2] - Owner: [Name] - Target: Q[N]

---

## Related Documents

- **Task Packet:** `.ai/tasks/[bug-id]/`
- **Bug Report:** [Link to original issue]
- **Fix PR:** [Link to pull request]
- **Test Cases:** `path/to/tests/`
- **Architecture:** [Link to relevant architecture docs]
- **ADRs:** [Links to related decisions]
- **Similar Bugs:** [Links to related retrospectives]

---

## References

- **Pattern Document:** `docs/investigations/patterns/[category].md`
- **Code Review Guide:** [Link]
- **Testing Standards:** [Link]
- **Secure Coding Guide:** [Link] (if security-related)

---

**Document Version:** 1.0
**Template Version:** 1.0.0
**Last Updated:** 2026-01-14
**Template Location:** `.ai-pack/templates/investigations/retrospective-template.md`

## Usage Instructions

1. Copy this template to `docs/investigations/BUG-YYYY-NNN-short-description.md`
2. Fill in all sections during and after investigation
3. Conduct thorough root cause analysis
4. Document all prevention measures
5. Update `docs/investigations/README.md` with entry
6. Link to pattern document if applicable
7. Track action items to completion
8. Use as reference for similar bugs in future
