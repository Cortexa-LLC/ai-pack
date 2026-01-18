# Beads Enforcement Gate

**Version:** 1.0.0
**Last Updated:** 2026-01-18
**Gate Type:** BLOCKING
**Enforcement Level:** MANDATORY

## Overview

This gate enforces the use of Beads task memory system throughout all AI-Pack workflows. Beads provides persistent, git-backed task tracking that survives session boundaries and enables cross-session memory.

**Critical Rule:** All task lifecycle operations MUST use Beads commands. Updating task packets alone is NOT sufficient.

---

## Gate Rules

### Rule 1: Task Creation MUST Use Beads

**REQUIREMENT:** Every non-trivial task MUST be created in Beads before work begins.

**Enforcement:**
```
BEFORE starting ANY task:
  IF task is non-trivial THEN
    REQUIRE: bd create command executed
    REQUIRE: Beads task ID documented in task packet
    BLOCK: Work without Beads task
  END IF
```

**Implementation:**
```bash
# MANDATORY - Create Beads task
task_id=$(bd create "Implement user authentication" --priority high --json | jq -r '.id')

# MANDATORY - Document in task packet
echo "**Beads Task:** ${task_id}" >> .ai/tasks/2026-01-18_auth/00-contract.md

# Now work can begin
```

**Verification:**
- ✅ `bd list` shows the task
- ✅ Task packet references Beads ID
- ✅ `.beads/issues.jsonl` contains task entry

**Violation:**
```bash
# ❌ WRONG - Just creating task packet without Beads
mkdir .ai/tasks/2026-01-18_auth
cp templates/* .ai/tasks/2026-01-18_auth/
# Missing: bd create command!
```

---

### Rule 2: Task Status MUST Use Beads Commands

**REQUIREMENT:** All status changes MUST be reflected in Beads, not just task packets.

**Mandatory Status Updates:**

| Workflow Event | Required Beads Command | Required Packet Update |
|----------------|------------------------|------------------------|
| Starting work | `bd start <id>` | Update 20-work-log.md |
| Completing task | `bd close <id>` | Update 40-acceptance.md |
| Getting blocked | `bd block <id> "reason"` | Update 20-work-log.md |
| Unblocking | `bd unblock <id>` | Update 20-work-log.md |
| Adding subtask | `bd create --depends-on <id>` | Update 10-plan.md |

**Enforcement Pattern:**
```bash
# Starting work
BEFORE implementing code:
  bd start ${task_id}  # MANDATORY
  echo "## Session $(date)" >> .ai/tasks/*/20-work-log.md  # THEN this

# Completing work
AFTER all work complete:
  bd close ${task_id}  # MANDATORY
  echo "✅ Task complete" >> .ai/tasks/*/40-acceptance.md  # THEN this

# Getting blocked
WHEN blocked:
  bd block ${task_id} "Waiting for API credentials"  # MANDATORY
  echo "BLOCKER: [details]" >> .ai/tasks/*/20-work-log.md  # THEN this
```

**Verification:**
```bash
# Check Beads status
bd show ${task_id}
# Status should match current workflow phase

# Check task packet
grep "Beads Task:" .ai/tasks/*/00-contract.md
# Should contain valid Beads ID
```

**Violation Examples:**
```bash
# ❌ WRONG - Only updating task packet
echo "Starting implementation" >> .ai/tasks/*/20-work-log.md
# Missing: bd start command!

# ❌ WRONG - Only updating task packet
echo "Task complete" >> .ai/tasks/*/40-acceptance.md
# Missing: bd close command!
```

---

### Rule 3: Task Discovery MUST Use bd ready

**REQUIREMENT:** Engineers MUST use `bd ready` to find next work, not manual task selection.

**Enforcement:**
```
WHEN finding next task:
  REQUIRE: bd ready command
  BLOCK: "Just pick something" approach
```

**Correct Pattern:**
```bash
# MANDATORY - Use bd ready to find next work
next_task=$(bd ready --json | jq -r '.[0]')
task_id=$(echo $next_task | jq -r '.id')
task_title=$(echo $next_task | jq -r '.title')

echo "Found next task: ${task_id} - ${task_title}"

# Get details
bd show ${task_id}

# Start work
bd start ${task_id}
```

**Why This Matters:**
- ✅ Respects dependencies automatically
- ✅ Follows priority order
- ✅ Prevents working on blocked tasks
- ✅ Coordinates with other agents

**Violation:**
```bash
# ❌ WRONG - Manual task selection
ls .ai/tasks/
cd .ai/tasks/2026-01-18_something
# Missing: bd ready to find correct task!
```

---

### Rule 4: Progress Monitoring MUST Use Beads

**REQUIREMENT:** Orchestrators MUST use Beads commands to monitor progress, not just file inspection.

**Enforcement:**
```
WHEN checking progress:
  REQUIRE: bd list OR bd show commands
  BLOCK: Checking only task packet files
```

**Correct Pattern:**
```bash
# Check overall progress
bd list --status open

# Check specific task
bd show ${task_id}

# Check what's ready
bd ready

# Check blocked tasks
bd list --status blocked
```

**Why This Matters:**
- ✅ Single source of truth
- ✅ Cross-session visibility
- ✅ Real-time status
- ✅ Dependency awareness

**Violation:**
```bash
# ❌ WRONG - Only checking files
ls .ai/tasks/
cat .ai/tasks/*/20-work-log.md
# Missing: bd list to see actual status!
```

---

### Rule 5: Dependencies MUST Use bd dep add

**REQUIREMENT:** Task dependencies MUST be declared in Beads, not just mentioned in task packets.

**Enforcement:**
```
WHEN creating dependent tasks:
  REQUIRE: bd dep add command
  BLOCK: Text-only dependency documentation
```

**Correct Pattern:**
```bash
# Create tasks
bd create "Design API" --priority high  # Returns bd-a1b2
bd create "Implement API" --priority high  # Returns bd-b2c3
bd create "Test API" --priority normal  # Returns bd-c3d4

# MANDATORY - Declare dependencies
bd dep add bd-b2c3 bd-a1b2  # Implementation depends on design
bd dep add bd-c3d4 bd-b2c3  # Testing depends on implementation

# Verify
bd show bd-c3d4
# Should show dependencies listed
```

**Why This Matters:**
- ✅ `bd ready` respects dependencies
- ✅ Cannot start blocked tasks
- ✅ Automatic dependency tracking
- ✅ Clear task ordering

**Violation:**
```bash
# ❌ WRONG - Only documenting in task packet
echo "Depends on: Design API task" >> .ai/tasks/*/00-contract.md
# Missing: bd dep add command!
```

---

### Rule 6: Agent Coordination MUST Use Beads

**REQUIREMENT:** When Orchestrator spawns agents, MUST create Beads tracking tasks.

**Enforcement:**
```
WHEN spawning agent with Task tool:
  REQUIRE: bd create "Agent: {Role} - {Task}"
  REQUIRE: bd start command for agent task
  BLOCK: Agent tracking without Beads
```

**Correct Pattern:**
```bash
# 1. Spawn agent
agent = Task(
  subagent_type="general-purpose",
  description="Implement login feature",
  prompt="Act as Engineer. Task packet: .ai/tasks/2026-01-18_login/"
)

# 2. MANDATORY - Create Beads task for agent
agent_task_id=$(bd create "Agent: Engineer - Implement login feature" \
  --assignee "Engineer-1" \
  --priority high \
  --description "Task packet: .ai/tasks/2026-01-18_login/" \
  --json | jq -r '.id')

# 3. MANDATORY - Mark as in-progress
bd start ${agent_task_id}

# 4. Document in work log
echo "Spawned Engineer-1 (Beads: ${agent_task_id})" >> .ai/tasks/*/20-work-log.md
```

**Monitoring Pattern:**
```bash
# Check active agents
bd list --assignee "Engineer-*" --status in_progress

# Check agent details
bd show ${agent_task_id}
```

**Why This Matters:**
- ✅ `/ai-pack agents` command works
- ✅ Cross-session agent tracking
- ✅ Coordinator can monitor progress
- ✅ Git-backed audit trail

**Violation:**
```bash
# ❌ WRONG - Spawning agent without Beads tracking
agent = Task(...)
echo "Spawned Engineer-1" >> work-log.md
# Missing: bd create for agent tracking!
```

---

## Workflow Integration

### Standard Workflow with Beads

**Phase 1: Understanding**
```bash
# Orchestrator creates task in Beads
task_id=$(bd create "Understand authentication requirements" --priority high --json | jq -r '.id')

# Create task packet
mkdir -p .ai/tasks/2026-01-18_auth
# Copy templates...

# Link in contract
echo "**Beads Task:** ${task_id}" > .ai/tasks/2026-01-18_auth/00-contract.md

# Start work
bd start ${task_id}

# Work happens...

# Complete
bd close ${task_id}
```

**Phase 2: Planning**
```bash
# Decompose into subtasks
design_id=$(bd create "Design auth architecture" --priority high --json | jq -r '.id')
impl_id=$(bd create "Implement auth service" --priority high --json | jq -r '.id')
test_id=$(bd create "Write auth tests" --priority normal --json | jq -r '.id')

# Set dependencies
bd dep add ${impl_id} ${design_id}
bd dep add ${test_id} ${impl_id}

# Document in plan
cat >> .ai/tasks/2026-01-18_auth/10-plan.md << EOF
## Subtasks
- ${design_id}: Design auth architecture
- ${impl_id}: Implement auth service (depends on design)
- ${test_id}: Write auth tests (depends on implementation)
EOF
```

**Phase 3: Implementation**
```bash
# Find next task
next=$(bd ready --json | jq -r '.[0].id')

# Start work
bd start ${next}

# During work - update both
echo "Implementing login endpoint" >> .ai/tasks/*/20-work-log.md

# If blocked
bd block ${next} "Waiting for security review"
echo "BLOCKER: Waiting for security review" >> .ai/tasks/*/20-work-log.md

# When unblocked
bd unblock ${next}

# When complete
bd close ${next}
echo "✅ Implementation complete" >> .ai/tasks/*/20-work-log.md
```

**Phase 4: Review**
```bash
# Tester creates validation task
test_task=$(bd create "Validate auth implementation" --priority high --json | jq -r '.id')

bd start ${test_task}
# Run tests...

IF tests_pass THEN
  bd close ${test_task}
ELSE
  bd block ${impl_id} "Tests failing - see test report"
  # Create fix tasks...
END IF
```

---

## Enforcement Checklist

### For Orchestrators

**Before delegation:**
```
□ All tasks created with bd create
□ Dependencies set with bd dep add
□ Priorities assigned
□ Task IDs documented in task packets
□ Agent tasks created for spawned workers
```

**During coordination:**
```
□ Using bd list to monitor progress
□ Using bd ready to find next work
□ Using bd show to check task details
□ Updating Beads when status changes
```

### For Engineers

**Before starting work:**
```
□ Used bd ready to find next task
□ Verified Beads task exists
□ Executed bd start command
□ Documented Beads ID in work log
```

**During implementation:**
```
□ Using bd block when stuck
□ Using bd unblock when resolved
□ Creating subtasks with bd create --depends-on
□ Keeping Beads status synchronized
```

**Before completion:**
```
□ All acceptance criteria met
□ Executed bd close command
□ Updated task packet
□ Used bd ready to find next work
```

### For Reviewers/Testers

**During validation:**
```
□ Created validation task with bd create
□ Marked in-progress with bd start
□ Blocked work with bd block if issues found
□ Closed with bd close when approved
```

---

## Verification Commands

### Check Beads Usage

```bash
# Verify task exists
bd show ${task_id}

# Check status matches workflow phase
bd list --status in_progress

# Verify dependencies
bd show ${task_id} | grep -A 10 "Dependencies"

# Check for agent tracking
bd list --assignee "Engineer-*"

# Verify no orphaned task packets
for dir in .ai/tasks/*/; do
  if ! grep -q "Beads Task:" "${dir}00-contract.md"; then
    echo "WARNING: ${dir} missing Beads task link"
  fi
done
```

### Audit Beads Compliance

```bash
# Check .beads/issues.jsonl is committed
git log --oneline .beads/issues.jsonl

# Verify task count matches work
task_count=$(bd list --json | jq 'length')
packet_count=$(ls -d .ai/tasks/*/ | wc -l)
echo "Beads tasks: ${task_count}, Task packets: ${packet_count}"

# Should match or Beads > packets (for agent tracking)
```

---

## Common Violations

### Violation 1: Task Packet Without Beads

**Symptom:**
```bash
$ ls .ai/tasks/
2026-01-18_feature-x/

$ grep "Beads Task:" .ai/tasks/2026-01-18_feature-x/00-contract.md
# No output - missing Beads reference!
```

**Fix:**
```bash
# Create Beads task retroactively
task_id=$(bd create "Feature X implementation" --priority high --json | jq -r '.id')

# Link in contract
echo "**Beads Task:** ${task_id}" >> .ai/tasks/2026-01-18_feature-x/00-contract.md
```

### Violation 2: Status Updated Only in Task Packet

**Symptom:**
```bash
$ cat .ai/tasks/*/20-work-log.md
## Status: In Progress

$ bd list
# Shows task as "open" not "in_progress"
```

**Fix:**
```bash
# Sync Beads status
bd start ${task_id}
```

### Violation 3: Dependencies Only in Text

**Symptom:**
```bash
$ cat .ai/tasks/*/00-contract.md
Dependencies: Must complete design task first

$ bd show ${task_id} | grep Dependencies
# No dependencies listed
```

**Fix:**
```bash
# Add proper dependency
design_id="bd-a1b2"  # Find from bd list
bd dep add ${task_id} ${design_id}
```

---

## Gate Compliance

### Passing Criteria

**Gate PASSES when:**
```
✓ All tasks created with bd create
✓ All status changes use bd commands
✓ Task discovery uses bd ready
✓ Dependencies declared with bd dep add
✓ Agent coordination uses Beads
✓ .beads/issues.jsonl committed to git
✓ Task packets reference Beads IDs
```

### Failing Criteria

**Gate FAILS when:**
```
✗ Task packet exists without Beads task
✗ Status changed without bd command
✗ Dependencies only in text
✗ Manual task selection (no bd ready)
✗ Agent spawned without Beads tracking
✗ .beads/issues.jsonl not committed
```

---

## Benefits of Enforcement

### Cross-Session Memory
```
Session 1 (Monday):
  bd create "Implement feature X"
  bd start bd-a1b2
  # Work begins...

Session 2 (Tuesday):
  bd ready
  # Shows: bd-a1b2 still in_progress
  bd show bd-a1b2
  # Full context preserved
```

### Dependency Awareness
```
bd dep add bd-b2c3 bd-a1b2

bd ready
# Will NOT show bd-b2c3 until bd-a1b2 is closed
```

### Multi-Agent Coordination
```
# Orchestrator spawns 3 engineers
bd create "Agent: Engineer-1 - Feature A" --assignee "Engineer-1"
bd create "Agent: Engineer-2 - Feature B" --assignee "Engineer-2"
bd create "Agent: Engineer-3 - Feature C" --assignee "Engineer-3"

# Each engineer queries their work
bd ready --assignee "Engineer-1"
```

### Git-Backed Audit Trail
```bash
# See all task changes
git log .beads/issues.jsonl

# See when task was created
git log --all --grep="${task_id}"
```

---

## Migration from Non-Beads Workflows

### For Existing Projects

If you have existing task packets without Beads:

```bash
#!/bin/bash
# migrate-to-beads.sh

for dir in .ai/tasks/*/; do
  task_name=$(basename "$dir")

  # Extract task description from contract
  description=$(grep "^## Task Description" "${dir}00-contract.md" -A 5 | tail -n 4)

  # Create Beads task
  task_id=$(bd create "$task_name" --description "$description" --json | jq -r '.id')

  # Link in contract
  sed -i "1s/^/**Beads Task:** ${task_id}\n\n/" "${dir}00-contract.md"

  echo "Migrated ${task_name} → ${task_id}"
done

# Commit
git add .beads/issues.jsonl
git add .ai/tasks/*/00-contract.md
git commit -m "Migrate existing tasks to Beads"
```

---

## References

- [Beads Integration Guide](../quality/tooling/beads-integration.md) - Complete Beads documentation
- [Orchestrator Role](../roles/orchestrator.md) - Beads usage for coordination
- [Engineer Role](../roles/engineer.md) - Beads usage for implementation
- [Standard Workflow](../workflows/standard.md) - Beads integration in workflows

---

## Summary

**Key Principle:** Task packets are DOCUMENTATION. Beads is STATE.

**Enforcement:**
- Task packets document WHAT and WHY
- Beads tracks WHO, WHEN, STATUS
- Both are required, neither is optional
- Beads is the source of truth for task state

**Simple Rule:**
```
BEFORE any task operation:
  Ask: "Did I run the bd command?"
  IF no THEN gate violation
```

---

**Last Updated:** 2026-01-18
**Enforcement Level:** MANDATORY, BLOCKING
**Next Review:** When Beads v2 releases
