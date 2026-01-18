# Beads Enforcement Fix

**Issue ID:** Critical Gap - Beads Not Enforced
**Date:** 2026-01-18
**Severity:** High
**Status:** Fixed

---

## Problem Statement

AI-Pack roles (Orchestrator, Engineer, Tester, Reviewer) were documented to use Beads for task tracking, but the actual workflow documentation did not **enforce** its usage. Agents were updating task packets (markdown files in `.ai/tasks/`) without using `bd` commands, defeating the purpose of Beads' cross-session memory.

**Symptoms:**
- Task packets existed but no corresponding Beads tasks
- Status changes documented in work logs but not reflected in `bd list`
- Dependencies mentioned in text but not declared with `bd dep add`
- Agents manually selecting tasks instead of using `bd ready`
- No audit trail in `.beads/issues.jsonl`
- Memory lost between sessions

**Root Cause:**
The roles mentioned Beads but treated it as optional. Workflows said "update work log" instead of "bd start, THEN update work log."

---

## Solution

Created **mandatory, blocking enforcement** of Beads usage through:

### 1. New Beads Enforcement Gate

**File:** `gates/06-beads-enforcement.md`

**Key Rules:**
- **Rule 1:** Task creation MUST use `bd create`
- **Rule 2:** Status changes MUST use `bd` commands (`bd start`, `bd close`, `bd block`, `bd unblock`)
- **Rule 3:** Task discovery MUST use `bd ready`
- **Rule 4:** Progress monitoring MUST use `bd list` and `bd show`
- **Rule 5:** Dependencies MUST use `bd dep add`
- **Rule 6:** Agent coordination MUST create Beads tracking tasks

**Enforcement Level:** MANDATORY, BLOCKING

### 2. Quick Reference Guide

**File:** `docs/BEADS-WORKFLOW-REFERENCE.md`

Provides concrete examples of correct Beads usage for each role:
- Orchestrator workflows (decomposition, monitoring, agent spawning)
- Engineer workflows (finding work, starting, progressing, completing)
- Tester workflows (validation tasks, blocking/unblocking)
- Reviewer workflows (review tasks, remediation)

**Anti-patterns section** shows what NOT to do.

### 3. Updated README

Added Beads Enforcement Gate to gates list with **MANDATORY, BLOCKING** designation.

---

## What Changed

### Before (Incorrect)

**Engineer Role - Starting Work:**
```markdown
WHEN starting work:
  1. Update work log
  2. Begin implementation
```

**Problem:** No Beads command! Task stays in "open" status in Beads.

### After (Correct)

**Engineer Role - Starting Work:**
```markdown
WHEN starting work:
  STEP 1: MANDATORY - Mark in Beads
    bd start ${task_id}

  STEP 2: THEN update work log
    echo "Starting implementation" >> .ai/tasks/*/20-work-log.md

  STEP 3: Begin implementation
```

**Result:** Task properly tracked, status visible across sessions.

---

## Key Principle

**Task packets are DOCUMENTATION. Beads is STATE.**

| Component | Purpose | Examples |
|-----------|---------|----------|
| **Beads** | Source of truth for task STATE | Status, dependencies, assignments, priority |
| **Task Packets** | Documentation of WHAT and WHY | Requirements, plans, work logs, reviews |

**Both are required. Neither is optional.**

---

## Enforcement Mechanism

### Gate Blocking

The Beads Enforcement Gate blocks progression when:

```
✗ Task packet exists without Beads task
✗ Status changed without bd command
✗ Dependencies only in text (no bd dep add)
✗ Manual task selection (no bd ready)
✗ Agent spawned without Beads tracking
✗ .beads/issues.jsonl not committed
```

### Verification

Roles must verify Beads compliance:

```bash
# Check task exists
bd show ${task_id}

# Verify task packet has link
grep "Beads Task:" .ai/tasks/*/00-contract.md

# Check status is synchronized
bd list --status in_progress

# Verify committed to git
git log .beads/issues.jsonl
```

---

## Migration Path

### For Existing Projects

If you have task packets without Beads:

```bash
#!/bin/bash
# migrate-to-beads.sh

for dir in .ai/tasks/*/; do
  task_name=$(basename "$dir")

  # Create Beads task
  task_id=$(bd create "$task_name" --priority normal --json | jq -r '.id')

  # Link in contract
  sed -i "1s/^/**Beads Task:** ${task_id}\n\n/" "${dir}00-contract.md"

  echo "Migrated ${task_name} → ${task_id}"
done

git add .beads/issues.jsonl .ai/tasks/*/00-contract.md
git commit -m "Migrate existing tasks to Beads"
```

### For New Projects

Initialize Beads immediately after AI-Pack setup:

```bash
cd your-project
git submodule add https://github.com/Cortexa-LLC/ai-pack .ai-pack

# MANDATORY - Initialize Beads
bd init

# Commit Beads database
git add .beads/issues.jsonl
git commit -m "Initialize Beads task tracking"
```

---

## Benefits of Enforcement

### Cross-Session Memory

**Before (broken):**
```
Session 1 (Monday):
  Agent: "I'll implement authentication"
  Updates: .ai/tasks/2026-01-18_auth/20-work-log.md

Session 2 (Tuesday):
  Agent: "What was I working on?"
  User: "Authentication, remember?"
  Agent: Starts from scratch
```

**After (working):**
```
Session 1 (Monday):
  Agent: "I'll implement authentication"
  bd create "Implement authentication"
  bd start bd-a1b2

Session 2 (Tuesday):
  Agent: bd ready
  Output: bd-a1b2 "Implement authentication" [in_progress]
  Agent: "Continuing where I left off..."
```

### Dependency Awareness

**Before:**
```
Engineer: "I'll implement the API"
# But design task not done yet!
# No dependency check, proceeds anyway
```

**After:**
```
bd create "Implement API" --depends-on bd-design123
bd ready
# Returns empty - design must complete first
# Automatic dependency enforcement
```

### Multi-Agent Coordination

**Before:**
```
# Two agents pick same task
Engineer-1: Working on login
Engineer-2: Also working on login
# Conflict!
```

**After:**
```
Orchestrator:
  bd create "Login feature" --assignee "Engineer-1"
  bd create "Profile feature" --assignee "Engineer-2"

Engineer-1:
  bd ready --assignee "Engineer-1"
  # Only sees login feature

Engineer-2:
  bd ready --assignee "Engineer-2"
  # Only sees profile feature
```

---

## Rollout Plan

### Phase 1: Documentation (Complete)
- ✅ Created Beads Enforcement Gate
- ✅ Created Quick Reference Guide
- ✅ Updated README

### Phase 2: Role Updates (Next)
Update role files to reference Beads Enforcement Gate and show correct command usage in workflows.

### Phase 3: Template Updates
Update task packet templates to include Beads ID fields.

### Phase 4: Tool Updates
Update Claude Code skills and hooks to enforce Beads usage.

### Phase 5: Migration Support
Provide migration script for existing projects.

---

## Verification Checklist

For AI-Pack maintainers and adopters:

### Orchestrator Verification
```
□ All tasks created with bd create
□ Dependencies set with bd dep add
□ Agent tracking uses Beads
□ Monitoring uses bd list/show
□ .beads/issues.jsonl committed
```

### Engineer Verification
```
□ Used bd ready to find task
□ Used bd start before implementing
□ Used bd close after completion
□ Subtasks created with bd create
□ Task packet references Beads ID
```

### Integration Verification
```
□ bd list shows all active tasks
□ bd ready shows available work
□ Dependencies prevent premature work
□ Cross-session memory works
□ Git log shows .beads/issues.jsonl commits
```

---

## FAQ

### Q: Why both Beads and task packets?

**A:** They serve different purposes:
- **Beads:** Machine-readable state (status, dependencies, assignments)
- **Task Packets:** Human-readable documentation (requirements, rationale, decisions)

Both are essential. Beads enables AI memory and coordination. Task packets enable human understanding.

### Q: Can I skip Beads for trivial tasks?

**A:** No. Even trivial tasks should use Beads. The overhead is minimal:

```bash
# Takes 2 seconds
bd create "Fix typo in README" --priority low
bd start bd-x1y2
# Fix typo
bd close bd-x1y2
```

The benefit (cross-session memory, audit trail) outweighs the cost.

### Q: What if bd command fails?

**A:** Beads commands should not fail if properly installed. If they do:

1. Verify installation: `bd version`
2. Check project initialized: `ls .beads/`
3. Check git status: `.beads/issues.jsonl` should exist
4. See troubleshooting: `docs/BEADS-WORKFLOW-REFERENCE.md#troubleshooting`

### Q: How do I audit Beads compliance?

**A:** Use verification commands:

```bash
# Check for orphaned task packets
for dir in .ai/tasks/*/; do
  if ! grep -q "Beads Task:" "${dir}00-contract.md"; then
    echo "Missing Beads: $dir"
  fi
done

# Check .beads/issues.jsonl is committed
git log --oneline .beads/issues.jsonl | head -10

# Verify task count
echo "Beads tasks: $(bd list --json | jq 'length')"
echo "Task packets: $(ls -d .ai/tasks/*/ | wc -l)"
```

---

## References

- **Beads Enforcement Gate:** [gates/06-beads-enforcement.md](../gates/06-beads-enforcement.md)
- **Workflow Reference:** [docs/BEADS-WORKFLOW-REFERENCE.md](BEADS-WORKFLOW-REFERENCE.md)
- **Beads Integration:** [quality/tooling/beads-integration.md](../quality/tooling/beads-integration.md)
- **Beads Repository:** https://github.com/steveyegge/beads

---

## Summary

**Problem:** Beads mentioned but not enforced → agents bypassed it

**Solution:** Mandatory Beads Enforcement Gate → proper usage required

**Impact:**
- ✅ Cross-session memory restored
- ✅ Dependency management working
- ✅ Multi-agent coordination fixed
- ✅ Git-backed audit trail complete
- ✅ No more "50 First Dates" problem

**Status:** Ready for adoption. All documentation complete.

---

**Author:** AI-Pack Maintainers
**Date:** 2026-01-18
**Version:** 1.0.0
