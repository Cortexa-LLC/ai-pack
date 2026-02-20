# Role Updates: Beads Enforcement

**Date:** 2026-01-18
**Version:** 1.0.0
**Status:** Complete

## Summary

Updated Orchestrator and Engineer roles to properly enforce Beads usage throughout all workflows. All task lifecycle operations now REQUIRE Beads commands, not just task packet updates.

---

## Changes Made

### 1. Orchestrator Role (`roles/orchestrator.md`)

**Version:** 1.1.0 → 1.2.0

#### 1.1 Role Overview (New)
Added critical notice at top of role:
```markdown
**⚠️ CRITICAL:** All task operations MUST use Beads commands.
See [Beads Enforcement Gate](../gates/06-beads-enforcement.md) for mandatory requirements.
```

#### 1.2 Task Creation (Section 1)
**Before:**
```
STEP 1: Create task packet directory
STEP 2: Copy templates
STEP 3: Fill out contract
STEP 4: Proceed to planning
```

**After:**
```
STEP 1: MANDATORY - Create Beads task
  task_id=$(bd create "Task description" --priority high --json | jq -r '.id')

STEP 2: Create task packet directory

STEP 3: Copy templates

STEP 4: Link Beads ID in 00-contract.md
  echo "**Beads Task:** ${task_id}" >> 00-contract.md

STEP 5: Fill out contract

STEP 6: Proceed to planning

ENFORCEMENT: Gate blocks if task packet exists without Beads task.
```

**Impact:** Task packets can no longer be created without Beads tasks first.

#### 1.3 Task Decomposition (Section 2)
**Before:**
- Mentioned using `bd create` in examples
- No explicit enforcement

**After:**
```markdown
**CRITICAL: All decomposition MUST use Beads commands.**
See [Beads Enforcement Gate](../gates/06-beads-enforcement.md) Rule 1.

MANDATORY Beads Workflow:
STEP 1: Analyze requirements
STEP 2: MANDATORY - Create Beads tasks (bd create)
STEP 3: MANDATORY - Set dependencies (bd dep add)
STEP 4: THEN create task packets
STEP 5: Verify with bd ready

ENFORCEMENT: Cannot create task packets before Beads tasks.
```

**Impact:** Orchestrators must use `bd create` and `bd dep add` before creating task packets.

#### 1.4 Progress Monitoring (Section 3)
**Before:**
- Listed Beads commands as activities
- No explicit enforcement

**After:**
```markdown
**ENFORCEMENT:** See [Beads Enforcement Gate](../gates/06-beads-enforcement.md) Rule 4.

**CRITICAL:** Progress monitoring MUST use Beads commands, not file inspection.
Task packets are documentation; Beads is state.

Monitoring Activities:
- MANDATORY: Check completion with bd list
- MANDATORY: Identify blockers with bd list --status blocked
- MANDATORY: Find ready work with bd ready
```

**Impact:** Orchestrators cannot check progress by inspecting files only.

#### 1.5 Agent Registration (Section 2.13)
**Before:**
```
EVERY agent spawned MUST have a Beads task.
NO EXCEPTIONS.
```

**After:**
```
**ENFORCEMENT:** See [Beads Enforcement Gate](../gates/06-beads-enforcement.md) Rule 6.

EVERY agent spawned MUST have a Beads task.
NO EXCEPTIONS.
GATE VIOLATION if skipped.
```

**Impact:** Explicit gate reference strengthens enforcement.

---

### 2. Engineer Role (`roles/engineer.md`)

**Version:** 1.1.0 → 1.2.0

#### 2.1 Role Overview (New)
Added critical notice at top of role:
```markdown
**⚠️ CRITICAL:** All task lifecycle operations MUST use Beads commands.
See [Beads Enforcement Gate](../gates/06-beads-enforcement.md) for mandatory requirements.
```

#### 2.2 Task Discovery (Section 0.7)
**Before:**
```bash
# Find tasks ready to work on
bd ready

# Mark task as in-progress
bd start bd-a1b2
```

**After:**
```bash
**ENFORCEMENT:** See [Beads Enforcement Gate](../gates/06-beads-enforcement.md).
All task operations MUST use Beads commands.

**CRITICAL:** Task discovery MUST use bd ready command, not manual selection.
See Rule 3 of Beads Enforcement Gate.

# Step 1: MANDATORY - Find tasks ready to work on
bd ready

# MANDATORY - Mark task as in-progress
bd start bd-a1b2

# GATE ENFORCEMENT: Work cannot begin without bd start command
```

**Impact:** Engineers cannot manually select tasks or start work without `bd start`.

#### 2.3 During Implementation (Section 0.7)
**Before:**
```bash
# If you get blocked
bd block bd-a1b2 "Waiting for API key"

# When task complete
bd close bd-a1b2
```

**After:**
```bash
# If you get blocked - MANDATORY use bd block
bd block bd-a1b2 "Waiting for API key"
# THEN update work log
echo "BLOCKER: Waiting for API key" >> .ai/tasks/*/20-work-log.md

# When unblocked - MANDATORY use bd unblock
bd unblock bd-a1b2
# THEN update work log
echo "UNBLOCKED: API key received" >> .ai/tasks/*/20-work-log.md

# When task complete
# MANDATORY - Close in Beads FIRST
bd close bd-a1b2

# THEN update task packet
echo "✅ Task complete" >> .ai/tasks/*/40-acceptance.md
```

**Impact:** Engineers must use Beads commands BEFORE updating task packets.

#### 2.4 Implementation Cycle (Section 1)
**Before:**
```
1. Understand requirements
2. Read existing code
3. MANDATORY TDD Cycle
```

**After:**
```
1. Understand requirements
2. Read existing code
3. MANDATORY - Start Beads task
   bd start <task-id>
   # Task must be in "in_progress" before implementing
4. MANDATORY TDD Cycle
```

**Impact:** Cannot begin TDD without marking task in-progress in Beads.

#### 2.5 During Work (Work Acceptance Criteria section)
**Before:**
```
WHILE working:
  update progress
  IF stuck THEN
    document blocker
  END IF
```

**After:**
```
WHILE working:
  update progress (work log only - Beads stays "in_progress")

  IF stuck THEN
    # MANDATORY - Block in Beads FIRST
    bd block <task-id> "Reason"
    # THEN document in work log
    echo "BLOCKER: [reason]" >> work-log.md

    # When unblocked
    bd unblock <task-id>
    echo "UNBLOCKED: [resolution]" >> work-log.md
  END IF
```

**Impact:** Blocking/unblocking must use Beads commands first.

#### 2.6 Completion Checklist (Before Completion section)
**Before:**
```
✓ Work log updated
✓ Commit messages clear
✓ Ready for review
```

**After:**
```
✓ Work log updated
✓ Commit messages clear
✓ Beads task closed with bd close <task-id> (MANDATORY - BLOCKING)
✓ Ready for review
```

Plus new section:
```markdown
**⚠️ CRITICAL: Beads Task Closure (MANDATORY)**

STEP 1: Verify all work complete
STEP 2: MANDATORY - Close in Beads FIRST (bd close)
STEP 3: THEN update acceptance document
STEP 4: Find next work (bd ready)

IF task not closed in Beads THEN
  GATE VIOLATION - Work incomplete
  BLOCK acceptance
  REQUIRE: bd close command
END IF
```

**Impact:** Task cannot be marked complete without `bd close` command.

---

## Enforcement Summary

### Key Principle Enforced

**"Task packets are DOCUMENTATION. Beads is STATE."**

| Component | Purpose | Enforcement |
|-----------|---------|-------------|
| **Beads** | Source of truth for task STATE | MANDATORY - all lifecycle ops |
| **Task Packets** | Documentation of WHAT and WHY | Created AFTER Beads task |

### Order of Operations (ENFORCED)

**Orchestrator:**
1. `bd create` → Create Beads task
2. `bd dep add` → Set dependencies
3. THEN create task packet
4. Link Beads ID in contract
5. Monitor with `bd list`, `bd ready`, `bd show`

**Engineer:**
1. `bd ready` → Find next task (not manual selection)
2. `bd start` → Mark in-progress
3. Implement (TDD)
4. If blocked: `bd block` → THEN update work log
5. If unblocked: `bd unblock` → THEN update work log
6. `bd close` → THEN update acceptance
7. `bd ready` → Find next task

### Violations Now Blocked

| Violation | Gate Blocks | Fix Required |
|-----------|-------------|--------------|
| Task packet without Beads task | ✅ Yes | Create Beads task first |
| Work started without `bd start` | ✅ Yes | Execute `bd start` |
| Status changed without `bd` command | ✅ Yes | Use proper `bd` command |
| Blocked without `bd block` | ✅ Yes | Execute `bd block` |
| Completed without `bd close` | ✅ Yes | Execute `bd close` |
| Manual task selection (no `bd ready`) | ✅ Yes | Use `bd ready` |

---

## Benefits of Updates

### 1. Cross-Session Memory
**Before:** Tasks lost between sessions
**After:** Tasks persist in `.beads/issues.jsonl`, git-backed

### 2. Dependency Management
**Before:** Dependencies in text only, not enforced
**After:** `bd dep add` enforces dependencies, `bd ready` respects them

### 3. Multi-Agent Coordination
**Before:** Race conditions, duplicate work
**After:** Beads prevents conflicts via status tracking

### 4. Audit Trail
**Before:** No history of task state changes
**After:** Complete git history in `.beads/issues.jsonl`

### 5. Real-Time Progress
**Before:** Must inspect files to check progress
**After:** `bd list` shows real-time status

---

## Verification

### Check Beads Enforcement in Roles

```bash
# Verify Orchestrator enforces Beads
grep -c "MANDATORY.*bd" roles/orchestrator.md
# Should show multiple occurrences

# Verify Engineer enforces Beads
grep -c "MANDATORY.*bd" roles/engineer.md
# Should show multiple occurrences

# Check for gate references
grep "Beads Enforcement Gate" roles/orchestrator.md roles/engineer.md
# Should show multiple references
```

### Test Workflow Compliance

```bash
# Orchestrator test: Create task
task_id=$(bd create "Test feature" --priority high --json | jq -r '.id')
mkdir -p .ai/tasks/ai-pack-4wx-20260118090000-test/
echo "**Beads Task:** ${task_id}" >> .ai/tasks/ai-pack-4wx-20260118090000-test/00-contract.md

# Engineer test: Start work
bd start ${task_id}
# Implement...
bd close ${task_id}

# Verify
bd show ${task_id}
# Should show status: closed
```

---

## Migration Path

### For Existing AI-Pack Projects

If roles were using old pattern (task packets without Beads):

1. **Audit existing tasks:**
```bash
for dir in .ai/tasks/*/; do
  if ! grep -q "Beads Task:" "${dir}00-contract.md"; then
    echo "Missing Beads: ${dir}"
  fi
done
```

2. **Migrate to Beads:**
```bash
for dir in .ai/tasks/*/; do
  task_name=$(basename "$dir")
  task_id=$(bd create "$task_name" --priority normal --json | jq -r '.id')
  echo "**Beads Task:** ${task_id}" >> "${dir}00-contract.md"
done
```

3. **Commit:**
```bash
git add .beads/issues.jsonl .ai/tasks/*/00-contract.md
git commit -m "Migrate existing tasks to Beads enforcement"
```

---

## Related Documentation

- **[Beads Enforcement Gate](../gates/06-beads-enforcement.md)** - Full enforcement rules
- **[Beads Workflow Reference](BEADS-WORKFLOW-REFERENCE.md)** - Quick reference guide
- **[Beads Integration Guide](../quality/tooling/beads-integration.md)** - Complete Beads documentation
- **[Beads Enforcement Fix](BEADS-ENFORCEMENT-FIX.md)** - Problem analysis and solution

---

## Next Steps

1. ✅ **COMPLETE:** Orchestrator role updated
2. ✅ **COMPLETE:** Engineer role updated
3. **TODO:** Update Tester role to enforce Beads for validation tasks
4. **TODO:** Update Reviewer role to enforce Beads for review tasks
5. **TODO:** Update task packet templates to include Beads ID field
6. **TODO:** Create Claude Code skill to auto-enforce Beads usage

---

## Summary

**Status:** Beads enforcement now properly integrated into Orchestrator and Engineer roles.

**Impact:**
- ✅ Task packets can no longer be created without Beads tasks
- ✅ Work cannot start without `bd start`
- ✅ Progress must be tracked with Beads commands
- ✅ Tasks cannot complete without `bd close`
- ✅ Cross-session memory fully functional
- ✅ Multi-agent coordination enforced

**Enforcement:** All violations now blocked by Beads Enforcement Gate.

---

**Last Updated:** 2026-01-18
**Authors:** AI-Pack Maintainers
