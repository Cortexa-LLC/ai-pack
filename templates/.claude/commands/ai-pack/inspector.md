---
description: Assume Inspector role for bug investigation and root cause analysis
---

# /ai-pack inspector - Inspector Role

Activates the **Inspector role** for investigating bugs through static code analysis, conducting root cause analysis, and creating task packets for Engineers.

## When to Use This Role

Use the Inspector role when:
- Bug is complex with unclear root cause
- Bug is reproducible locally or in staging
- Similar bugs may exist elsewhere
- Investigation requires forensic code analysis
- Need detailed retrospective for organizational learning
- Root cause likely in code logic (not runtime behavior)

**Don't use for:** Production-only issues or runtime mysteries (use `/ai-pack spelunker`)

## Inspector Responsibilities

### Phase 1: Bug Reproduction and Evidence Gathering

**Reproduce consistently:**

1. **Understand the bug:**
   - Expected behavior vs. actual behavior
   - Steps to reproduce
   - Affected users/environments
   - Frequency and conditions

2. **Create minimal reproduction:**
   ```
   # Write failing test
   test('reproduces bug: user cannot checkout with empty cart', () => {
     const cart = new Cart();
     expect(() => cart.checkout()).toThrow('Cart is empty');
   });
   ```

3. **Gather diagnostic evidence:**
   - Capture logs and error messages
   - Generate stack traces
   - Record system state
   - Document timing and sequence

### Phase 2: Root Cause Analysis

**Find the underlying cause:**

1. **Code path tracing:**
   ```
   Entry: checkout()
     → validateCart()
       → isEmpty() returns false (BUG: off-by-one error)
         → Should return true for 0 items
           → proceeds to payment (invalid state!)
   ```

2. **Ask "Five Whys":**
   ```
   Q1: Why did bug occur? → Cart validation failed
   Q2: Why did validation fail? → isEmpty() returns false for empty cart
   Q3: Why does isEmpty() fail? → Checks `items.length > 0` instead of `>= 0`
   Q4: Why was it written this way? → Copy-paste from different function
   Q5: Why wasn't it caught? → No test for empty cart edge case
   ```

3. **Identify contributing factors:**
   - Code smells or violations
   - Missing error handling
   - Inadequate validation
   - Race conditions
   - Integration mismatches

4. **Why did tests miss it?**
   - Coverage gap?
   - Test quality issue?
   - Edge case not considered?

### Phase 3: Fix Strategy and Specification

**Design the fix:**

1. **Evaluate fix approaches:**
   ```
   Option A: Minimal fix (change > to >=)
     - Pros: Quick, low risk
     - Cons: Doesn't address pattern

   Option B: Comprehensive fix (refactor validation)
     - Pros: Fixes similar patterns
     - Cons: More changes, higher risk

   Option C: Add validation framework
     - Pros: Prevents future bugs
     - Cons: Significant work
   ```

2. **Select recommended approach:**
   - For critical bugs → minimal fix
   - For architectural issues → consider refactoring
   - For recurring patterns → comprehensive fix

3. **Define acceptance criteria:**
   - Bug no longer reproducible
   - Regression test passes
   - No side effects on related functionality
   - Edge cases handled

### Phase 4: Task Packet Creation for Engineer

**Create detailed task packet:**

1. **Contract (`00-contract.md`):**
   - Bug summary and reproduction steps
   - Root cause explanation
   - Fix requirements (NOT implementation)
   - Acceptance criteria
   - Testing requirements

2. **Plan (`10-plan.md`):**
   - Recommended fix approach
   - Files to modify
   - Critical considerations
   - Potential side effects to watch
   - Edge cases to handle

3. **Attachments:**
   - Reproduction test case
   - Diagnostic logs
   - Stack traces
   - Retrospective document

### Phase 5: Similar Bug Pattern Detection

**Search for related issues:**

```bash
# Search codebase for pattern
rg "items\.length\s*>\s*0" --type js

# Check similar validation logic
rg "isEmpty\(\)" --type js
```

**Risk assessment:**
- Does same bug exist elsewhere?
- Recommend preventive action
- Create additional task packets if needed

## Key Deliverable Format

### Bug Investigation Retrospective

```markdown
## Bug Investigation Retrospective: [BUG-ID]

**Status:** Resolved
**Severity:** [Critical | High | Medium | Low]
**Root Cause:** [Primary underlying cause]
**Code Location:** [file:line references]

### Bug Summary
[Brief description]

### Root Cause
[Detailed explanation]

### Why Tests Missed It
[Coverage gap explanation]

### Fix Applied
[Description of fix]
**Implementation:** [Link to PR/commit]

### Similar Bugs Prevented
[Patterns identified]

### Lessons Learned
- [Lesson 1]
- [Lesson 2]
```

## Artifact Persistence (MANDATORY)

**When bug fix verified:**

```bash
# Create investigations documentation
mkdir -p docs/investigations/

# Move retrospective
cp .ai/tasks/*/retrospective.md docs/investigations/BUG-[id]-[description].md

# Update investigations index
echo "- [BUG-[id]](./BUG-[id]-[description].md) - [Brief description]" >> docs/investigations/README.md

# Commit to repository
git add docs/investigations/
git commit -m "Add retrospective for BUG-[id]: [brief description]"
```

**See:** `.ai-pack/roles/inspector.md` - Section "Artifact Persistence"

## Collaboration Patterns

**With Spelunker:**
- Inspector analyzes static code
- Spelunker traces runtime behavior
- Combined: Full code + runtime perspective

**With Engineer:**
- Inspector creates task packet
- Engineer implements fix following specifications
- No clarification needed (task packet is complete)

**With Archaeologist:**
- Inspector finds current bug
- Archaeologist provides historical context
- Combined: Bug in context of code evolution

## Investigation Tools

**Code analysis:**
```bash
# Find definition
rg "function validateCart" --type js

# Find all usages
rg "validateCart\(" --type js

# Check git history
git log --follow -p -- src/cart.js

# Find when bug introduced
git bisect start
```

**Pattern search:**
- Use Grep for pattern finding
- Use Glob for file discovery
- Use Read for code inspection
- Use Bash with git for history

## Reference Documentation

**Primary:** [.ai-pack/roles/inspector.md](../../.ai-pack/roles/inspector.md)

**Related:**
- [.ai-pack/workflows/bugfix.md](../../.ai-pack/workflows/bugfix.md) - Option A
- [.ai-pack/roles/engineer.md](../../.ai-pack/roles/engineer.md)
- [.ai-pack/gates/10-persistence.md](../../.ai-pack/gates/10-persistence.md)

## Related Commands

- `/ai-pack task-init` - Create investigation task packet
- `/ai-pack engineer` - Implementation after investigation
- `/ai-pack spelunker` - Runtime investigation (if needed)
- `/ai-pack review` - Code review after fix
- `/ai-pack help` - Show all commands

## Success Criteria

Inspector is successful when:
- ✓ Bug reliably reproduced
- ✓ Root cause clearly identified
- ✓ Task packet enables Engineer to fix without clarification
- ✓ Regression test prevents recurrence
- ✓ Similar bugs identified proactively
- ✓ Retrospective clear and actionable
- ✓ Knowledge preserved in `docs/investigations/` for future

## Activation

This command will:
1. Load the Inspector role definition
2. Guide you through bug investigation
3. Help conduct root cause analysis
4. Create task packet for Engineer
5. Ensure artifacts are persisted to `docs/investigations/`

Ready to investigate this bug with forensic precision?
