# Debugging Decision Tree

**Version:** 1.0.0
**Created:** 2026-01-31

## Purpose

This guide helps engineers and orchestrators quickly determine the appropriate debugging approach based on bug complexity and characteristics.

---

## Quick Decision Tree

```
Bug Reported
    ↓
    1. Can you identify the root cause immediately?
    ├─ YES
    │   ↓
    │   2. Is it confined to a single file/module?
    │   ├─ YES → SIMPLE BUG
    │   │   └─ Engineer fixes directly (bugfix workflow Phase 1)
    │   │
    │   └─ NO → COMPLEX BUG
    │       └─ Delegate to Inspector for investigation
    │
    └─ NO
        ↓
        3. Is it a production-only or runtime issue?
        ├─ YES → RUNTIME INVESTIGATION
        │   └─ Delegate to Spelunker for runtime analysis
        │
        └─ NO
            ↓
            4. Does it affect multiple modules or suggest design problem?
            ├─ YES → COMPLEX/ARCHITECTURAL
            │   └─ Delegate to Inspector
            │       └─ Inspector may identify architectural issue
            │           └─ Consider refactoring instead of simple fix
            │
            └─ NO → PROCEED WITH CAUTION
                └─ Engineer attempts fix (bugfix workflow Phase 1)
                    └─ If complexity discovered → Stop and request Inspector
```

---

## Detailed Classification Guide

### Simple Bug (Direct to Engineer)

**Characteristics:**
- ✅ Root cause obvious from error message
- ✅ Single file/module affected
- ✅ Can reproduce in < 5 minutes
- ✅ Clear stack trace points to issue
- ✅ Similar to previously fixed bugs
- ✅ No architectural concerns
- ✅ Fix approach is straightforward

**Examples:**
```
"Login button text shows 'Lgoin' instead of 'Login'"
→ Typo, 1 file, obvious fix
→ Action: Engineer fixes directly

"Null pointer exception in UserService.getProfile line 42"
→ Clear stack trace, single location
→ Action: Engineer fixes with proper null check

"Off-by-one error in pagination (page 1 shows items 0-9 instead of 1-10)"
→ Logic error, clear location, obvious fix
→ Action: Engineer fixes loop bounds

"Form validation message has wrong color (red instead of orange)"
→ CSS issue, single file, straightforward
→ Action: Engineer updates CSS
```

**Workflow:** Engineer uses [Bugfix Workflow](../../workflows/bugfix.md) Phase 1 directly

**Expected Time:** 15-60 minutes including tests

---

### Complex Bug (Inspector Investigation Required)

**Characteristics:**
- ⚠️ Root cause unclear or mysterious
- ⚠️ Multiple modules/files involved
- ⚠️ Intermittent or hard to reproduce
- ⚠️ No clear error message or stack trace
- ⚠️ Potential design/architectural issue
- ⚠️ Similar bugs may exist elsewhere
- ⚠️ Affects multiple features/code paths

**Examples:**
```
"Users randomly logged out after 10-15 minutes"
→ Intermittent, timing issue, unclear cause
→ Action: Inspector investigates session management, identifies race condition

"Data inconsistency: order totals don't match between cart, checkout, and confirmation"
→ Multi-module, unclear where discrepancy originates
→ Action: Inspector traces data flow, identifies calculation logic scattered across modules

"API returns 500 error but only for some users in specific conditions"
→ Root cause unclear, not reproducible locally
→ Action: Inspector investigates with production logs, identifies edge case

"Application slows down after 1000 concurrent users"
→ Performance issue, needs profiling
→ Action: Spelunker investigates runtime behavior (or Inspector for static analysis)

"Search results missing items when query contains special characters"
→ Unclear where filtering breaks, multiple search implementations
→ Action: Inspector investigates all search implementations, identifies inconsistency
```

**Workflow:**
1. Orchestrator delegates to Inspector (or Spelunker for runtime)
2. Inspector/Spelunker investigates and creates retrospective
3. Inspector creates task packet with fix specification
4. Orchestrator delegates to Engineer with task packet
5. Engineer implements fix per specification

**Expected Time:** Investigation 1-3 hours, Fix 1-2 hours

---

### Runtime Investigation (Spelunker Required)

**Characteristics:**
- 🔍 Production-only issue (can't reproduce locally)
- 🔍 Performance issue requiring profiling
- 🔍 Intermittent bug (timing, race conditions, Heisenbugs)
- 🔍 Complex distributed system issue
- 🔍 Need to understand actual runtime behavior
- 🔍 Deep call stack mysteries
- 🔍 External integration failures
- 🔍 Unfamiliar live system investigation

**Examples:**
```
"Feature works locally but fails in production"
→ Environment-specific issue
→ Action: Spelunker investigates production environment, identifies missing config

"Performance degrades after 24 hours uptime"
→ Memory leak or resource exhaustion
→ Action: Spelunker profiles runtime, identifies leak

"Occasional database deadlocks under high load"
→ Race condition, timing-dependent
→ Action: Spelunker traces concurrent requests, identifies lock contention

"Integration with payment provider fails randomly"
→ External dependency issue
→ Action: Spelunker investigates runtime behavior, network logs, identifies timeout issue
```

**Workflow:**
1. Orchestrator delegates to Spelunker
2. Spelunker investigates runtime (traces execution, profiles, inspects state)
3. Spelunker creates runtime report
4. Orchestrator delegates to Engineer with runtime findings
5. Engineer implements fix based on runtime analysis

**Expected Time:** Investigation 2-4 hours, Fix 1-3 hours

---

### Architectural Issue (Consider Refactoring)

**Characteristics:**
- 🔴 Multiple implementations of same logic detected
- 🔴 Logic scattered across 5+ files
- 🔴 Duplicate code causing inconsistency
- 🔴 Violation of SOLID principles
- 🔴 Bug is symptom, not root cause
- 🔴 Similar bugs fixed before in different locations
- 🔴 Code smells strongly suggest design problem

**Examples:**
```
"Email validation inconsistent: registration accepts 'user@domain' but checkout rejects it"
→ Validation logic duplicated in 5 places with different implementations
→ Action: Inspector identifies duplication, recommends consolidation
→ Decision: Refactor to single validation utility instead of patching each location

"Same null pointer bug fixed 3 times in 3 different payment processors"
→ Duplicate payment processing code
→ Action: Inspector identifies pattern, recommends refactoring to unified payment abstraction
→ Decision: Refactor to eliminate duplication

"Adding new product field requires updating 12 different files"
→ Shotgun surgery, poor abstraction
→ Action: Architect designs proper data model
→ Decision: Refactor to proper layered architecture

"Authentication logic scattered across controllers, services, and middleware"
→ Responsibility diffusion, SOLID violation
→ Action: Inspector maps all auth code, identifies design issue
→ Decision: Refactor to centralized auth module
```

**Workflow:**
1. Inspector investigates bug
2. Inspector identifies architectural issue (not just a bug)
3. Inspector documents design problem and recommendation
4. Orchestrator escalates to Architect (for design) or considers refactoring
5. If refactoring approved:
   - Architect designs proper structure
   - Engineer implements refactoring (using [Refactor Workflow](../../workflows/refactor.md))
6. If not approved (time constraints):
   - Engineer implements minimal bug fix
   - Create backlog item for future refactoring
   - Document technical debt

**Expected Time:** Investigation 2-3 hours, Refactoring 4-12 hours (vs quick patch 1 hour)

**Trade-off:** Refactoring takes longer initially but prevents recurring similar bugs

---

## Decision Factors

### When to Delegate to Inspector

**Delegate if ANY of these are true:**
- Root cause not immediately obvious
- Affects 3+ files/modules
- Intermittent or unreproducible locally
- Similar bugs reported before
- Investigation will take 2+ hours
- Design concerns suspected
- "I don't know why this is happening"

### When to Delegate to Spelunker

**Delegate if ANY of these are true:**
- Production-only issue
- Performance problem
- Timing-dependent bug (race conditions)
- Need runtime profiling or tracing
- Distributed system issue
- External integration mystery
- "It works on my machine but not in production"

### When Engineer Can Proceed Directly

**Proceed only if ALL of these are true:**
- Root cause is clear
- Affects 1-2 files
- Reproducible quickly
- Fix approach is obvious
- No architectural concerns
- Confident in solution
- "I know exactly what's wrong and how to fix it"

### When to Consider Refactoring

**Consider refactoring if ANY of these are true:**
- Duplicate code causing bug
- Similar bugs fixed before
- Fix requires changing 5+ files
- SOLID principles violated
- Bug is symptom of design issue
- Quick fix will create technical debt
- "This is the third time we've fixed this pattern"

---

## Anti-Patterns to Avoid

### The Thrashing Engineer

**Pattern:**
```
Engineer attempts fix → Fails → Tries different approach → Fails →
Tries another approach → Fails → 100+ turns → Still not fixed
```

**Why it happens:**
- Jumped to TDD without understanding complexity
- Didn't recognize need for investigation
- Assumed simple when actually complex

**Solution:**
- STOP after 30 turns without progress
- Request Inspector investigation
- Get root cause analysis BEFORE continuing

### The Symptom Fixer

**Pattern:**
```
Bug: Users can't checkout
Fix: Add null check in CheckoutController
Result: Bug moves to OrderConfirmationController
Fix: Add null check there too
Result: Bug moves to PaymentController
...endless whack-a-mole
```

**Why it happens:**
- Fixed symptom, not root cause
- Didn't investigate where null originates
- Quick fix without understanding

**Solution:**
- Inspector investigates root cause
- Identifies: User object not properly initialized
- Proper fix: Fix initialization, not add null checks everywhere

### The Hasty Refactorer

**Pattern:**
```
Bug: Validation inconsistent
Engineer: "I'll refactor all validation!"
→ 100 files changed
→ 3 days of work
→ Breaking changes
→ Tests failing everywhere
```

**Why it happens:**
- Correct diagnosis (architectural issue)
- Wrong execution (too big, too fast)
- No incremental plan

**Solution:**
- Inspector identifies architectural issue
- Architect designs incremental refactoring plan
- Engineer implements in small, safe steps
- Use [Refactor Workflow](../../workflows/refactor.md)

---

## Communication Templates

### Engineer Requesting Investigation

```markdown
I've been assigned [BUG-ID] but discovered it's more complex than expected:

**Complexity Indicators:**
- Root cause unclear: [what you found]
- Multiple modules affected: [list files/modules]
- Attempted approaches: [what you tried]

**Request:**
Recommend delegating to Inspector for root cause analysis before fix attempt.

**Progress So Far:**
- Turns spent: [number]
- Files examined: [list]
- Findings: [what you learned]
```

### Inspector Identifying Architectural Issue

```markdown
Investigation of [BUG-ID] complete. Root cause analysis reveals architectural issue.

**Root Cause:**
[Describe the design problem]

**Architectural Smell:**
- Pattern: [Duplication | SOLID violation | etc.]
- Scope: [affected files/modules]
- Impact: [why this matters]

**Options:**

**Option A: Minimal Fix (Quick)**
- Fix immediate bug
- Leave architecture as-is
- Time: 1-2 hours
- Risk: Similar bugs will recur

**Option B: Proper Refactoring (Thorough)**
- Fix architectural issue
- Eliminate bug pattern
- Time: 4-8 hours
- Risk: Upfront time investment

**Recommendation:**
[Your recommendation with rationale]

**Next Steps:**
- If Option A: Create task packet for minimal fix
- If Option B: Engage Architect for refactoring design
```

### Orchestrator Escalating to Architect

```markdown
Bug investigation [BUG-ID] by Inspector revealed architectural issue requiring design input.

**Issue Summary:**
[Brief description]

**Architectural Concern:**
[What design problem exists]

**Request:**
Please design proper architecture to eliminate this bug pattern and prevent recurrence.

**Context:**
- Affected modules: [list]
- Inspector findings: [link to retrospective]
- Business impact: [severity]
```

---

## Success Metrics

### Good Debugging Process

```
✅ Bugs routed to correct role first time
✅ Complex bugs don't cause engineer thrashing
✅ Architectural issues identified early
✅ Root cause fixes, not symptom patches
✅ Refactoring considered when appropriate
✅ No wasted effort on wrong approach
```

### Poor Debugging Process

```
❌ Engineers thrash on complex bugs (100+ turns)
❌ Same bug fixed multiple times in different locations
❌ Symptom fixes that don't address root cause
❌ Architectural issues discovered too late
❌ Refactoring happens reactively, not proactively
❌ Time wasted on wrong debugging approach
```

---

## Quick Reference Card

```
┌─────────────────────────────────────────────────┐
│  Debugging Decision Quick Reference              │
├─────────────────────────────────────────────────┤
│                                                  │
│  Root cause obvious?                            │
│  → YES, single file → Engineer directly         │
│  → NO, unclear → Inspector                      │
│                                                  │
│  Production-only?                               │
│  → YES → Spelunker                              │
│  → NO → Inspector or Engineer                   │
│                                                  │
│  Multiple modules affected?                     │
│  → YES → Inspector                              │
│  → NO → Engineer (with caution)                 │
│                                                  │
│  Duplicate code involved?                       │
│  → YES → Consider refactoring                   │
│  → NO → Standard bugfix                         │
│                                                  │
│  Engineer thrashing (30+ turns)?                │
│  → STOP → Request Inspector                     │
│                                                  │
│  Similar bug fixed before?                      │
│  → YES → Architectural issue → Refactor         │
│  → NO → Standard bugfix                         │
│                                                  │
└─────────────────────────────────────────────────┘
```

---

## Related Documentation

- [Bugfix Workflow](../../workflows/bugfix.md) - Detailed bugfix process
- [Refactor Workflow](../../workflows/refactor.md) - Safe refactoring process
- [Inspector Role](../../roles/inspector.md) - Bug investigation specialist
- [Spelunker Role](../../roles/spelunker.md) - Runtime investigation specialist
- [Engineer Role](../../roles/engineer.md) - Implementation and complexity assessment
- [Orchestrator Role](../../roles/orchestrator.md) - Delegation strategy

---

**Last reviewed:** 2026-01-31
**Next review:** Quarterly or when debugging practices evolve
