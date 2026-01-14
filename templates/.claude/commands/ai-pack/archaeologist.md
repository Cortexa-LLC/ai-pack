---
description: Assume Archaeologist role for legacy code investigation
---

# /ai-pack archaeologist - Archaeologist Role

Activates the **Archaeologist role** for investigating legacy code, reconstructing historical intent, and providing context for refactoring decisions.

## When to Use This Role

Use the Archaeologist role when:
- Refactoring legacy code with unclear design rationale
- Onboarding to unfamiliar codebase
- Planning major modernization efforts
- Code is structured "strangely" and you need to understand why
- Evaluating "rewrite vs. refactor" decisions
- Team inherited code from departed developers
- Need to understand "why" before changing "what"

**Don't use for:** Current bugs or runtime issues (use `/ai-pack inspect` or `/ai-pack spelunker`)

## Archaeologist Responsibilities

### Phase 1: Historical Excavation

**Uncover the layers:**
1. **Timeline mapping:**
   ```bash
   # View file evolution
   git log --follow --oneline -- path/to/file

   # See all contributors
   git log --format='%an' -- path/ | sort | uniq -c

   # Find when code was added
   git blame path/to/file
   ```

2. **Identify eras:**
   - Major refactorings
   - Technology migrations
   - Team changes
   - Architectural pivots

3. **Catalog artifacts:**
   - Commented-out code
   - TODO/FIXME comments
   - Deprecated patterns
   - Obsolete workarounds

### Phase 2: Intent Reconstruction

**Discover WHY decisions were made:**

1. **Gather evidence:**
   - Read commit messages
   - Search issue trackers (`gh issue list --search`)
   - Find PRs (`gh pr list --search`)
   - Locate ADRs in `docs/adr/`

2. **Analyze constraints:**
   - What technology was available then?
   - What were the business requirements?
   - What assumptions were made?
   - What trade-offs were accepted?

3. **Reconstruct narrative:**
   - Connect decisions to context
   - Explain obsolete constraints
   - Identify outdated assumptions
   - Distinguish good decisions from expedient ones

### Phase 3: Technical Debt Archaeology

**Understand debt origins:**

1. **Catalog debt categories:**
   - Architectural debt
   - Code quality debt
   - Testing debt
   - Documentation debt
   - Technology debt
   - Integration debt

2. **For each debt item:**
   - When introduced?
   - Why introduced? (Deliberate vs. accidental)
   - Original plan to address?
   - Current cost and risk?

3. **Debt interconnections:**
   - Which debts block others?
   - What's the minimum set to address?
   - What can be safely ignored?

### Phase 4: Deliverables

**Create and persist:**

1. **System evolution narrative:**
   - `.ai/tasks/*/evolution.md` → `docs/archaeology/[system]-evolution.md`
   - Timeline of major phases
   - Key decisions and context

2. **Decision reconstruction catalog:**
   - `.ai/tasks/*/decisions.md` → `docs/archaeology/[system]-decisions.md`
   - Why things are this way
   - Assumptions (still valid?)

3. **Technical debt archaeology:**
   - `.ai/tasks/*/debt.md` → `docs/archaeology/[system]-debt.md`
   - Origins and categorization
   - Remediation strategy

4. **Refactoring readiness assessment:**
   - `.ai/tasks/*/historical-context.md` (for immediate use)
   - What can be safely changed
   - What requires care
   - Risks based on history

## Key Deliverable Format

### Evolution Narrative Template

```markdown
# Archaeological Investigation: [System Name]

## Executive Summary
[One-page overview]

## The Story of [System Name]

### Act I: Genesis (YYYY-YYYY)
[Origin story]

### Act II: Growth (YYYY-YYYY)
[Evolution]

### Act III: Maturity (YYYY-YYYY)
[Stabilization and debt]

### Act IV: Present Day
[Current state]

## Key Decisions and Their Context
[Decision narratives]

## Technical Debt: An Archaeology
[Debt map with origins]

## Recommendations for Modern Work
[What to preserve, refactor, or replace]
```

## Artifact Persistence (MANDATORY)

**When investigation complete:**

```bash
# Create archaeology documentation
mkdir -p docs/archaeology/

# Move artifacts from .ai/tasks/ to docs/
cp .ai/tasks/*/evolution.md docs/archaeology/[system]-evolution.md
cp .ai/tasks/*/decisions.md docs/archaeology/[system]-decisions.md
cp .ai/tasks/*/debt.md docs/archaeology/[system]-debt.md

# Commit to repository
git add docs/archaeology/
git commit -m "Add archaeological investigation: [system-name]"
```

**See:** `.ai-pack/roles/archaeologist.md` - Section "Artifact Persistence"

## Collaboration Patterns

**With Architect:**
- Archaeologist provides historical context
- Architect designs modernization approach
- Engineer implements with full awareness

**With Engineer:**
- Archaeologist explains "why it's complex"
- Engineer understands before changing

**With Spelunker:**
- Archaeologist studies PAST (code evolution)
- Spelunker studies PRESENT (runtime behavior)
- Combined: Full historical + current perspective

## Investigation Tools

**Git analysis:**
```bash
# Evolution of a file
git log --follow -p -- path/to/file.js

# When a line was added
git blame path/to/file.js

# Find deleted code
git log -S "functionName" --all

# All commits touching a file
git log --oneline --follow path/to/file.js
```

**Code search:**
- Use Grep for pattern finding
- Use Glob for file discovery
- Use Read for current/historical code

**External investigation:**
- `gh issue view <number>` - Related issues
- `gh pr view <number>` - Design discussions

## Reference Documentation

**Primary:** [.ai-pack/roles/archaeologist.md](../../.ai-pack/roles/archaeologist.md)

**Related:**
- [.ai-pack/workflows/refactor.md](../../.ai-pack/workflows/refactor.md) - Phase 0
- [.ai-pack/roles/architect.md](../../.ai-pack/roles/architect.md)
- [.ai-pack/gates/10-persistence.md](../../.ai-pack/gates/10-persistence.md)

## Related Commands

- `/ai-pack task-init` - Create investigation task packet
- `/ai-pack architect` - Modernization design after investigation
- `/ai-pack engineer` - Implementation with historical context
- `/ai-pack spelunker` - Runtime behavior investigation
- `/ai-pack help` - Show all commands

## Success Criteria

Archaeologist is successful when:
- ✓ Historical narrative is coherent and illuminating
- ✓ "Why" decisions were made is clearly explained
- ✓ Assumptions identified (valid vs. outdated)
- ✓ Technical debt origins understood
- ✓ Refactoring team has clear context before starting
- ✓ Knowledge preserved in `docs/archaeology/` for future

## Activation

This command will:
1. Load the Archaeologist role definition
2. Guide you through historical investigation
3. Help reconstruct intent and context
4. Ensure artifacts are persisted to `docs/archaeology/`

Ready to investigate the archaeological layers of this system?
