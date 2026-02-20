# Beads Workflow Quick Reference

**Version:** 1.0.0
**Last Updated:** 2026-01-18
**Purpose:** Quick reference for proper Beads usage in AI-Pack workflows

---

## Critical Rule

**Task packets are DOCUMENTATION. Beads is STATE.**

Every task lifecycle operation MUST use Beads commands. Updating task packets alone is NOT sufficient.

---

## Related Documentation

- **[Work Item Patterns](WORK-ITEM-PATTERNS.md)** - Comprehensive guide to Epics, Stories, Tasks, Spikes, and Issues across Beads and GitHub
- **[Beads Enforcement Gate](../gates/06-beads-enforcement.md)** - Mandatory rules for Beads usage
- **[GitHub Integration Usage](GITHUB-INTEGRATION-USAGE.md)** - Sync Beads with GitHub Issues

---

## Orchestrator Workflows

### Task Decomposition

```bash
# STEP 1: Create Beads tasks
design_id=$(bd create "Design auth architecture" --priority high --json | jq -r '.id')
impl_id=$(bd create "Implement auth service" --priority high --json | jq -r '.id')
test_id=$(bd create "Write auth tests" --priority normal --json | jq -r '.id')

# STEP 2: Set dependencies
bd dep add ${impl_id} ${design_id}  # impl depends on design
bd dep add ${test_id} ${impl_id}    # tests depend on impl

# STEP 3: Create task packet
mkdir -p .ai/tasks/ai-pack-4wx-20260118090000-auth
cp .ai-pack/templates/task-packet/* .ai/tasks/ai-pack-4wx-20260118090000-auth/

# STEP 4: Link in contract
cat >> .ai/tasks/ai-pack-4wx-20260118090000-auth/00-contract.md << EOF
**Beads Tasks:**
- ${design_id}: Design auth architecture
- ${impl_id}: Implement auth service
- ${test_id}: Write auth tests

**Dependencies:**
- Implementation depends on design
- Tests depend on implementation
EOF
```

### Monitoring Progress

```bash
# Check what's ready to work on
bd ready

# Check overall status
bd list --status open

# Check specific task
bd show ${task_id}

# Check blocked tasks
bd list --status blocked

# Check completed work
bd list --status closed
```

### Agent Spawning

```bash
# STEP 1: Spawn agent with Task tool
engineer = Task(
  subagent_type="general-purpose",
  description="Implement authentication",
  prompt="Act as Engineer. Task: .ai/tasks/ai-pack-4wx-20260118090000-auth/"
)

# STEP 2: MANDATORY - Create Beads tracking task
agent_task=$(bd create "Agent: Engineer - Implement authentication" \
  --assignee "Engineer-1" \
  --priority high \
  --json | jq -r '.id')

# STEP 3: Mark in-progress
bd start ${agent_task}

# STEP 4: Document
echo "Spawned Engineer-1 (Beads: ${agent_task})" >> .ai/tasks/ai-pack-4wx-20260118090000-auth/20-work-log.md

# STEP 5: Monitor agent
bd show ${agent_task}
bd list --assignee "Engineer-*" --status in_progress
```

---

## Engineer Workflows

### Finding Next Task

```bash
# STEP 1: Find what's ready
bd ready

# STEP 2: Get task details
task_id=$(bd ready --json | jq -r '.[0].id')
bd show ${task_id}

# STEP 3: Verify task packet exists
task_title=$(bd show ${task_id} --json | jq -r '.title')
# Find corresponding task packet in .ai/tasks/
```

### Starting Work

```bash
# STEP 1: MANDATORY - Mark in Beads
bd start ${task_id}

# STEP 2: THEN update work log
cat >> .ai/tasks/*/20-work-log.md << EOF
## Session $(date +%Y-%m-%d_%H:%M)

**Beads Task:** ${task_id}
**Status:** Started
EOF
```

### During Implementation

```bash
# When discovering subtasks
subtask_id=$(bd create "Add password hashing" --depends-on ${task_id} --json | jq -r '.id')
echo "Created subtask: ${subtask_id}" >> .ai/tasks/*/20-work-log.md

# When getting blocked
bd block ${task_id} "Waiting for API credentials"
echo "BLOCKER: Waiting for API credentials" >> .ai/tasks/*/20-work-log.md

# When unblocked
bd unblock ${task_id}
echo "UNBLOCKED: Credentials received" >> .ai/tasks/*/20-work-log.md

# Regular progress updates
echo "Implemented login endpoint" >> .ai/tasks/*/20-work-log.md
# Note: Beads status stays "in_progress" - no bd command needed for progress notes
```

### Completing Work

```bash
# STEP 1: Verify completion
# - All acceptance criteria met
# - All tests passing
# - Code coverage adequate

# STEP 2: MANDATORY - Close in Beads
bd close ${task_id}

# STEP 3: THEN update task packet
cat >> .ai/tasks/*/40-acceptance.md << EOF
## Completion

**Beads Task:** ${task_id}
**Status:** Closed
**Completed:** $(date)

✅ All acceptance criteria met
✅ All tests passing
✅ Code coverage: 87%
EOF

# STEP 4: Find next work
bd ready
```

---

## Tester Workflows

### Validation Task

```bash
# STEP 1: Create validation task
validation_id=$(bd create "Validate ${feature} TDD compliance" --priority high --json | jq -r '.id')

# STEP 2: Start validation
bd start ${validation_id}

# STEP 3: Run validation
# Check TDD compliance
# Check test coverage
# Check test quality

# STEP 4: If issues found
IF issues_found THEN
  # Block original task
  bd block ${original_task_id} "TDD violations - see review"

  # Create fix tasks
  fix_id=$(bd create "Fix TDD violations in ${feature}" \
    --depends-on ${validation_id} \
    --priority high \
    --json | jq -r '.id')

  # Complete validation (with issues)
  bd close ${validation_id}

  # Document
  cat >> .ai/tasks/*/30-review.md << EOF
  **Tester Verdict:** CHANGES REQUIRED
  **Validation Task:** ${validation_id}
  **Fix Task:** ${fix_id}
  **Original Task Blocked:** ${original_task_id}
  EOF
ELSE
  # Close validation
  bd close ${validation_id}

  # Unblock original if was blocked
  bd unblock ${original_task_id}

  # Document
  echo "**Tester Verdict:** APPROVED" >> .ai/tasks/*/30-review.md
  echo "**Validation Task:** ${validation_id}" >> .ai/tasks/*/30-review.md
END IF
```

---

## Reviewer Workflows

### Code Review Task

```bash
# STEP 1: Create review task
review_id=$(bd create "Review ${feature} code quality" --priority high --json | jq -r '.id')

# STEP 2: Start review
bd start ${review_id}

# STEP 3: Conduct review
# Check code quality
# Check standards compliance
# Check security
# Check documentation

# STEP 4: If changes needed
IF changes_needed THEN
  # Block original task
  bd block ${original_task_id} "Code quality issues - see review"

  # Create remediation tasks
  fix_id=$(bd create "Address code quality issues in ${feature}" \
    --depends-on ${review_id} \
    --priority high \
    --json | jq -r '.id')

  # Complete review
  bd close ${review_id}

  # Document
  cat >> .ai/tasks/*/30-review.md << EOF
  **Reviewer Verdict:** CHANGES REQUESTED
  **Review Task:** ${review_id}
  **Remediation Task:** ${fix_id}
  **Original Task Blocked:** ${original_task_id}
  EOF
ELSE
  # Close review
  bd close ${review_id}

  # Document approval
  echo "**Reviewer Verdict:** APPROVED" >> .ai/tasks/*/30-review.md
  echo "**Review Task:** ${review_id}" >> .ai/tasks/*/30-review.md
END IF
```

---

## Common Patterns

### Epic with Subtasks

```bash
# Create epic
epic_id=$(bd create "User Authentication System" --priority high --json | jq -r '.id')

# Create subtasks depending on epic
bd create "Design auth API" --depends-on ${epic_id} --priority high
bd create "Implement JWT tokens" --depends-on ${epic_id} --priority high
bd create "Add password hashing" --depends-on ${epic_id} --priority high
bd create "Write auth tests" --depends-on ${epic_id} --priority normal

# When ready to start, epic dependencies must be closed first
# Then subtasks become available via bd ready
```

### Parallel Tasks with Integration

```bash
# Create independent parallel tasks
frontend_id=$(bd create "Build frontend component" --priority high --json | jq -r '.id')
backend_id=$(bd create "Build backend API" --priority high --json | jq -r '.id')
db_id=$(bd create "Build database layer" --priority high --json | jq -r '.id')

# Create integration task that depends on all
integration_id=$(bd create "Integration testing" --priority normal --json | jq -r '.id')
bd dep add ${integration_id} ${frontend_id}
bd dep add ${integration_id} ${backend_id}
bd dep add ${integration_id} ${db_id}

# All three parallel tasks must close before integration shows in bd ready
```

### Sequential Workflow

```bash
# Create phase-based tasks
phase1=$(bd create "Phase 1: Research" --priority high --json | jq -r '.id')
phase2=$(bd create "Phase 2: Design" --priority high --json | jq -r '.id')
phase3=$(bd create "Phase 3: Implement" --priority high --json | jq -r '.id')
phase4=$(bd create "Phase 4: Test" --priority normal --json | jq -r '.id')

# Set dependencies (each phase depends on previous)
bd dep add ${phase2} ${phase1}
bd dep add ${phase3} ${phase2}
bd dep add ${phase4} ${phase3}

# bd ready will only show phase1 initially
# After closing phase1, phase2 becomes available, etc.
```

---

## Verification

### Check Beads Compliance

```bash
# Verify task exists in Beads
bd show ${task_id}

# Verify task packet has Beads link
grep "Beads Task:" .ai/tasks/*/00-contract.md

# Check for orphaned task packets (packets without Beads tasks)
for dir in .ai/tasks/*/; do
  if ! grep -q "Beads Task:" "${dir}00-contract.md"; then
    echo "WARNING: ${dir} missing Beads reference"
  fi
done

# Verify .beads/issues.jsonl is committed
git log --oneline .beads/issues.jsonl | head -5
```

### Check Status Sync

```bash
# Get Beads status
beads_status=$(bd show ${task_id} --json | jq -r '.status')

# Compare with work log
echo "Beads status: ${beads_status}"
tail -10 .ai/tasks/*/20-work-log.md

# Should match current workflow phase
```

---

## Anti-Patterns (DON'T DO THIS)

### ❌ Creating Task Packet Without Beads

```bash
# WRONG
mkdir .ai/tasks/ai-pack-4wx-20260118090000-feature
cp templates/* .ai/tasks/ai-pack-4wx-20260118090000-feature/
# Missing: bd create command!

# CORRECT
task_id=$(bd create "Feature implementation" --priority high --json | jq -r '.id')
mkdir .ai/tasks/ai-pack-4wx-20260118090000-feature
cp templates/* .ai/tasks/ai-pack-4wx-20260118090000-feature/
echo "**Beads Task:** ${task_id}" >> .ai/tasks/ai-pack-4wx-20260118090000-feature/00-contract.md
```

### ❌ Status Update Without Beads

```bash
# WRONG
echo "Starting implementation" >> .ai/tasks/*/20-work-log.md
# Missing: bd start command!

# CORRECT
bd start ${task_id}
echo "Starting implementation" >> .ai/tasks/*/20-work-log.md
```

### ❌ Manual Task Selection

```bash
# WRONG
cd .ai/tasks/ai-pack-4wx-20260118090000-something
# Just picking a task arbitrarily

# CORRECT
next_task=$(bd ready --json | jq -r '.[0].id')
bd show ${next_task}
# Then find corresponding task packet
```

### ❌ Dependencies Only in Text

```bash
# WRONG
echo "Depends on: Design task" >> .ai/tasks/*/00-contract.md
# Missing: bd dep add command!

# CORRECT
bd dep add ${impl_task_id} ${design_task_id}
echo "Depends on: Design task (${design_task_id})" >> .ai/tasks/*/00-contract.md
```

### ❌ Completing Without Closing

```bash
# WRONG
echo "✅ Task complete" >> .ai/tasks/*/40-acceptance.md
bd ready  # Still shows this task!

# CORRECT
bd close ${task_id}
echo "✅ Task complete" >> .ai/tasks/*/40-acceptance.md
bd ready  # Task no longer appears
```

---

## Troubleshooting

### "Can't find task for this task packet"

**Problem:** Task packet exists but no Beads task

**Solution:**
```bash
# Create Beads task retroactively
task_name=$(basename "$(dirname "$PWD")")
task_id=$(bd create "$task_name" --priority normal --json | jq -r '.id')

# Link in contract
echo "**Beads Task:** ${task_id}" >> 00-contract.md

# Sync status
bd start ${task_id}
```

### "bd ready shows nothing but there's work to do"

**Problem:** All tasks either in-progress or have unmet dependencies

**Solution:**
```bash
# Check what's in-progress
bd list --status in_progress

# Check blocked tasks
bd list --status blocked

# Check dependencies
for task in $(bd list --status open --json | jq -r '.[].id'); do
  echo "Task: $task"
  bd show $task | grep -A 5 "Dependencies"
done
```

### "Task packet and Beads status don't match"

**Problem:** Status changed in one but not the other

**Solution:**
```bash
# Beads is source of truth - sync packet to Beads
beads_status=$(bd show ${task_id} --json | jq -r '.status')

case $beads_status in
  "open")
    echo "Status: Open - ready to start" >> .ai/tasks/*/20-work-log.md
    ;;
  "in_progress")
    echo "Status: In progress" >> .ai/tasks/*/20-work-log.md
    ;;
  "blocked")
    reason=$(bd show ${task_id} --json | jq -r '.blocked_reason')
    echo "Status: Blocked - ${reason}" >> .ai/tasks/*/20-work-log.md
    ;;
  "closed")
    echo "Status: Closed" >> .ai/tasks/*/40-acceptance.md
    ;;
esac
```

---

## Quick Command Reference

| Operation | Beads Command | Task Packet Update |
|-----------|---------------|-------------------|
| Create task | `bd create "Task name" --priority high` | Document ID in 00-contract.md |
| Start work | `bd start <id>` | Add session to 20-work-log.md |
| Get blocked | `bd block <id> "reason"` | Document blocker in 20-work-log.md |
| Unblock | `bd unblock <id>` | Document resolution in 20-work-log.md |
| Complete | `bd close <id>` | Update 40-acceptance.md |
| Find next | `bd ready` | N/A |
| Check status | `bd show <id>` | N/A |
| Add dependency | `bd dep add <child> <parent>` | Document in 00-contract.md |
| Create subtask | `bd create --depends-on <parent>` | Document in 10-plan.md |

---

## Summary

**Golden Rule:** Every task state change MUST go through Beads first.

**Workflow:**
1. **Create:** `bd create` → then create task packet
2. **Start:** `bd start` → then update work log
3. **Work:** Update work log as needed (no bd command for progress notes)
4. **Block:** `bd block` → then document reason
5. **Unblock:** `bd unblock` → then document resolution
6. **Complete:** `bd close` → then update acceptance

**Why This Matters:**
- ✅ Cross-session memory
- ✅ Dependency management
- ✅ Multi-agent coordination
- ✅ Git-backed audit trail
- ✅ Real-time status tracking

**Enforcement:** See [Beads Enforcement Gate](../gates/06-beads-enforcement.md)

---

**Last Updated:** 2026-01-18
**See Also:**
- [Beads Integration Guide](../quality/tooling/beads-integration.md)
- [Beads Enforcement Gate](../gates/06-beads-enforcement.md)
- [Orchestrator Role](../roles/orchestrator.md)
- [Engineer Role](../roles/engineer.md)
