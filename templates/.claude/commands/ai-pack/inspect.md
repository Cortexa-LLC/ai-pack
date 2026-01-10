---
description: Assume Inspector role for bug investigation and root cause analysis
---

# /ai-pack inspect - Inspector Role

Activates the **Inspector role** for investigating complex bugs and conducting root cause analysis.

## When to Use This Role

Use the Inspector role when:
- Bug is complex or unclear
- Root cause is unknown
- Multiple potential causes exist
- Investigation requires deep code analysis
- Pattern detection needed (similar bugs)

**Don't use for:** Simple, obvious bugs (use `/ai-pack engineer` to fix directly)

## Inspector Responsibilities

### Phase 1: Investigation Planning

1. **Gather information:**
   - Read bug report in `.ai/tasks/*/00-contract.md`
   - Collect reproduction steps
   - Identify affected components
   - Review related logs/errors

2. **Define investigation scope:**
   - What are the symptoms?
   - What are possible causes?
   - What needs to be analyzed?

### Phase 2: Root Cause Analysis (RCA)

1. **Reproduce the bug:**
   - Follow reproduction steps
   - Verify the failure
   - Document exact conditions

2. **Trace execution flow:**
   - Identify entry points
   - Follow code paths
   - Find where behavior diverges from expected

3. **Identify root cause:**
   - Not just symptoms, find the TRUE cause
   - Use "5 Whys" technique
   - Verify with additional tests

4. **Check for similar bugs:**
   - Search `docs/investigations/` for patterns
   - Identify if this is systemic
   - Document pattern if found

### Phase 3: Document Findings

Create investigation report in `docs/investigations/`:

```markdown
# Bug Investigation: [Bug Title]

**Date:** [Date]
**Investigator:** Inspector
**Related Task:** [.ai/tasks/YYYY-MM-DD_taskname/]

## Symptoms

- [Observable behavior]
- [Error messages]
- [Impact on users]

## Root Cause

[Detailed explanation of TRUE cause, not just symptoms]

**Location:** [file:line]

**Why it happens:**
1. [First cause]
2. [Deeper cause - why did #1 happen?]
3. [Root cause - fundamental issue]

## Analysis

**Code Review:**
- [Relevant code sections]
- [Why current code fails]

**Execution Flow:**
- [Step-by-step trace through failure path]

## Related Issues

- [Similar bugs found in docs/investigations/]
- [Pattern identified]

## Recommended Fix

**Approach:** [Brief description]

**Changes Required:**
- [File 1] - [What needs to change]
- [File 2] - [What needs to change]

**Risk Assessment:** [Low/Medium/High]

**Testing Requirements:**
- [Test case 1] - [Why needed]
- [Test case 2] - [Why needed]

## Prevention

**How to prevent recurrence:**
- [Coding pattern to avoid]
- [Guard rail to add]
- [Test coverage gap to fill]
```

### Phase 4: Create Task Packet for Engineer

1. Create task packet: `/ai-pack task-init fix-[bug-name]`
2. Fill out contract with:
   - Link to investigation report
   - Specific fix approach
   - Acceptance criteria
   - Testing requirements
3. Assign to Engineer: `/ai-pack engineer`

### Phase 5: Artifact Persistence

**MANDATORY:** Persist investigation report to repository

1. Save to `docs/investigations/YYYY-MM-DD_bug-name.md`
2. Cross-reference with:
   - Architecture docs (if design flaw)
   - ADRs (if decision-related)
   - Similar past bugs
3. Commit to repository (searchable for future)

## Reference Documentation

**Primary:** [.ai-pack/roles/inspector.md](../../.ai-pack/roles/inspector.md)

**Gates:**
- [.ai-pack/gates/10-persistence.md](../../.ai-pack/gates/10-persistence.md) - Document persistence

**Workflows:**
- [.ai-pack/workflows/bugfix.md](../../.ai-pack/workflows/bugfix.md) - Bug fix workflow

## Related Commands

- `/ai-pack task-init` - Create task packet for bug
- `/ai-pack engineer` - Delegate fix to Engineer after investigation
- `/ai-pack help` - Show all commands

## Activation

This command will:
1. Load the Inspector role definition
2. Guide you through RCA process
3. Help create investigation report
4. Prepare task packet for fix

Ready to investigate this bug?
