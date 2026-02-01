# Bugfix Workflow

**Version:** 1.1.0
**Last Updated:** 2026-01-31

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

### Phase 0: Complexity Assessment (FIRST STEP)

**Objective:** Determine whether bug requires investigation before attempting fix.

**CRITICAL:** This assessment prevents engineers from thrashing on complex issues. Jumping to TDD without understanding scope leads to 100+ turn debugging sessions.

**Complexity Indicators:**

**Simple Bug (Proceed to Phase 1 directly):**
```
✅ Root cause obvious from error message
✅ Single file/module affected
✅ Can reproduce in < 5 minutes
✅ Clear stack trace points to issue
✅ Similar to previously fixed bugs
✅ No architectural concerns
✅ Fix approach is straightforward
```

**Complex Bug (Delegate to Inspector or Spelunker first):**
```
⚠️ Root cause unclear or mysterious
⚠️ Multiple modules/files involved
⚠️ Intermittent or hard to reproduce
⚠️ No clear error message or stack trace
⚠️ Potential design/architectural issue
⚠️ Similar bugs may exist elsewhere
⚠️ Affects multiple features/paths
⚠️ Production-only issue (can't reproduce locally)
```

**Architectural Issue (Consider Refactoring instead):**
```
🔴 Multiple implementations of same logic detected
🔴 Logic scattered across 5+ files
🔴 Duplicate code causing inconsistency
🔴 Violation of SOLID principles
🔴 Bug is symptom, not root cause
🔴 Similar bugs fixed before in different locations
🔴 Code smells strongly suggest design problem
```

---

**Decision Tree:**
```
IS bug simple and obvious?
├─ YES → Proceed to Phase 1 (Engineer can handle)
└─ NO
   └─ IS root cause unclear OR multi-module?
      ├─ YES → Delegate to Inspector (static analysis)
      │        OR Spelunker (runtime investigation)
      └─ NO
         └─ ARE there multiple implementations or architectural smells?
            ├─ YES → Consider REFACTORING instead
            │        (Consult Architect, use refactor workflow)
            └─ NO → Proceed to Phase 1 with caution
```

---

**Why This Phase Matters:**

**Without Phase 0 (Anti-Pattern):**
- Engineer: "I'll fix this bug with TDD"
- → 50 turns: trying approach A
- → 50 turns: trying approach B
- → 100 turns: reverting and trying approach C
- → Discovers: issue spans 5 modules, architectural problem
- → Total: 300+ turns wasted, still not fixed

**With Phase 0 (Proper Flow):**
- Orchestrator: "This looks complex - multiple modules affected"
- → Delegates to Inspector for investigation
- → Inspector: root cause analysis, identifies architectural issue
- → Inspector: creates task packet with fix strategy
- → Engineer: implements fix per specification in 20 turns
- → Result: Fixed correctly, root cause addressed

---

**For Engineers:**

IF you are assigned a bug directly (without investigation):
- Perform this complexity assessment yourself
- IF you discover it's more complex than expected:
  - Multiple modules affected
  - Unclear root cause
  - Architectural concerns
  - Taking more than 30 turns without progress

STOP and request:
- Inspector investigation (for complex bugs)
- Orchestrator guidance (for architectural issues)
- Task packet creation (for unclear requirements)

**Anti-Pattern Recognition:**
```
Symptoms of thrashing:
- 50+ turns without progress
- Trying multiple approaches
- Reverting changes repeatedly
- Touching more files than expected
- Discovering new complexity continuously

→ STOP: You need investigation/planning, not more TDD attempts
```

---

**Complexity Assessment Examples:**

**Simple Bugs (Direct to Engineer):**
- "Login button text shows 'Lgoin' instead of 'Login'" → Typo, 1 file, obvious
- "Null pointer exception in UserService.getProfile line 42" → Clear stack trace, single location
- "Off-by-one error in pagination (page 1 shows items 0-9 instead of 1-10)" → Logic error, obvious fix

**Complex Bugs (Inspector/Spelunker First):**
- "Users randomly logged out after 10-15 minutes" → Intermittent, timing issue
- "Data inconsistency: order totals don't match between cart, checkout, and confirmation" → Multi-module
- "API returns 500 error but only for some users" → Root cause unclear, needs investigation
- "Performance degrades after 1000 concurrent users" → Needs profiling, Spelunker investigation

**Architectural Issues (Consider Refactoring):**
- "Validation logic inconsistent: 5 different implementations across modules" → Design problem
- "Same null pointer bug fixed 3 times in 3 different payment processors" → Architectural smell
- "Fix requires changing shared utility used by 20+ modules" → Refactoring candidate
- "Bug reveals violation of single responsibility principle" → Design issue

---

### Phase 1: Bug Triage & Reproduction

**Objective:** Understand, reproduce, and assess the bug.

**NOTE:** If you reached Phase 1, bug is assessed as "simple" with clear scope. For complex bugs, Inspector or Spelunker should investigate first.

**DELEGATION STRATEGY:**

**Option A: Delegate to Inspector (Static Code Analysis for Complex Bugs)**
```
IF bug is complex AND root cause unclear AND reproducible locally THEN
  Orchestrator delegates to Inspector
  Inspector conducts static code investigation and retrospective
  Inspector creates task packet for Engineer
  Orchestrator delegates to Engineer with task packet
END IF
```

**Option B: Delegate to Spelunker (Runtime Investigation for Production/Live System Issues)**
```
IF production-only issue OR runtime behavior investigation needed THEN
  Orchestrator delegates to Spelunker
  Spelunker investigates runtime behavior (traces execution, inspects state)
  Spelunker creates runtime report for Engineer
  Orchestrator delegates to Engineer with runtime findings
END IF
```

**Option C: Hybrid Approach (Both Inspector and Spelunker)**
```
IF complex issue requires both runtime and static analysis THEN
  Orchestrator delegates to Spelunker (runtime investigation)
  Orchestrator delegates to Inspector (static code analysis)
  Wait for combined findings
  Orchestrator delegates to Engineer with full context
END IF
```

**Option D: Engineer Self-Investigation (For Simple Bugs)**
```
IF bug is simple OR root cause obvious THEN
  Orchestrator delegates to Engineer
  Engineer follows bugfix workflow phases 1-4
END IF
```

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

## When Bug Indicates Architectural Problem

**CRITICAL:** Some bugs are symptoms of deeper architectural issues. Fixing the immediate bug without addressing the root design problem leads to recurring similar bugs.

### Red Flags (Architectural Smells)

```
🔴 Multiple implementations of same logic (DRY violation)
   Example: Validation logic duplicated in 5 controllers

🔴 Logic scattered across many files
   Example: Payment processing spread across 8 modules

🔴 Fix requires changing 5+ files
   Example: Adding field requires updates in 7 places

🔴 Similar bugs fixed before in different locations
   Example: Third time fixing "null user" bug in different module

🔴 Code violates SOLID principles
   Example: God class with 3000 lines doing everything

🔴 Copy-paste code causing inconsistency
   Example: Same algorithm with slight variations in 4 places

🔴 Shotgun surgery required
   Example: Simple change needs touching 10+ files
```

### What to Do When Architectural Issue Detected

```
WHEN architectural smell detected:

STEP 1: Document the architectural concern
  - Describe the design problem
  - List affected modules/files
  - Explain why it's a design issue not just a bug

STEP 2: Assess severity and urgency
  IF bug is CRITICAL and blocking users THEN
    implement minimal fix to unblock
    document technical debt created
    create follow-up refactoring task
  ELSE
    escalate to Orchestrator for refactoring consideration
  END IF

STEP 3: Create refactoring proposal
  - Problem description
  - Proposed architecture
  - Migration strategy
  - Risk assessment

STEP 4: Orchestrator decides
  Option A: Refactor now (if time permits)
    → Delegate to Architect for design
    → Then Engineer implements refactoring
  Option B: Minimal fix + refactor later
    → Engineer fixes immediate bug
    → Create backlog item for refactoring
  Option C: Accept technical debt
    → Document reasoning
    → Track for future review
```

### Example: Validation Duplication

**Bug Report:**
"User registration fails validation but checkout succeeds with same invalid email"

**Investigation:**
- Email validation duplicated in 5 places
- Each implementation slightly different
- Some allow "user@domain" (no TLD), others reject
- No single source of truth

**Diagnosis:** Architectural issue, not simple bug

**Options:**

**Option A: Minimal Fix (Quick)**
```javascript
// Fix the immediate bug: make all validations consistent
// Copy the strictest validation to all 5 locations
// Document: "TODO: Consolidate validation logic"
```

**Option B: Proper Fix (Takes longer but correct)**
```javascript
// 1. Create shared validation utility
class EmailValidator {
  static isValid(email) {
    // Single implementation
  }
}

// 2. Replace all 5 duplicates with calls to utility
// 3. Add comprehensive tests
// 4. Verify all modules use consistent validation
```

**Decision Criteria:**
- If CRITICAL: Option A (fix now, refactor later)
- If MAJOR: Option B (fix it right)
- If MINOR: Option B (no excuse not to do it right)

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

## Post-Fix: Retrospective Artifact Persistence

**CRITICAL:** After bug is fixed and verified, bug investigation retrospective MUST be persisted to repository for organizational learning.

### When to Persist Retrospective

```
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
```

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
```
project-root/
├── docs/
│   ├── investigations/
│   │   ├── BUG-123-null-pointer-in-payment.md
│   │   ├── BUG-145-race-condition-order-processing.md
│   │   └── README.md (index by root cause category)
│   └── ...
└── .ai/
    └── tasks/ (temporary work-in-progress)
```

**Communication Pattern:**
```
"Bug fix complete and retrospective committed to repository.

Location: docs/investigations/[bug-id]-[description].md

Root Cause: [Brief explanation]
Category: [Pattern category]
Lessons Learned: [Summary]

This retrospective is now part of the organizational knowledge base."
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
✓ Retrospective persisted to docs/investigations/ (for non-trivial bugs)
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
