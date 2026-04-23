---
description: Create and orchestrate tasks via AI-Pack framework. Automatically creates Beads tasks, task packets, spawns agents, and monitors progress. MANDATORY for all non-trivial work.
---

# Task Orchestrator - Automated Task Management

You are using the **Task Orchestrator** workflow from the ai-pack framework.

## Purpose

Automate the complete task lifecycle:
1. Create Beads task for persistence
2. Create task packet for documentation
3. Spawn appropriate agent role
4. Monitor progress via Beads
5. Coordinate quality gates
6. Ensure completion

## When This Skill Is Required

**MANDATORY for:**
- Any non-trivial task (>2 steps, code changes, >30 minutes)
- User requests "create a task for..."
- User asks to "schedule work" or "queue work"
- Complex multi-step operations
- Work requiring multiple agents

**NOT required for:**
- Simple questions
- File reading/exploration
- Trivial one-line changes

## Workflow

### Step 1: Create Beads Task (MANDATORY FIRST)

```bash
# Create Beads task with proper metadata
bd create \
  --title="<concise-title>" \
  --description="<detailed-description>" \
  --type=<task|bug|feature|epic> \
  --priority=<0-4 or P0-P4> \
  --json

# Capture task ID from JSON output
# Example: listingsgql-9fj
```

**Priority levels:**
- P0/0 = Critical (production down)
- P1/1 = High (blocking work)
- P2/2 = Medium (normal priority) **← DEFAULT**
- P3/3 = Low (nice to have)
- P4/4 = Backlog (someday/maybe)

**Task types:**
- `task` - General work item
- `bug` - Defect or error
- `feature` - New functionality
- `epic` - Large multi-task initiative

### Step 2: Create Task Packet

```bash
# Create task packet directory
TASK_ID=<beads-task-id>  # From Step 1 (e.g., listingsgql-9fj)
TIMESTAMP=$(date +%Y%m%d%H%M%S)
SHORT_DESC=<short-description>  # Lowercase, hyphens (e.g., dgs-architecture-review)

TASK_DIR=".ai/tasks/${TASK_ID}-${TIMESTAMP}-${SHORT_DESC}"
mkdir -p "$TASK_DIR"

# Copy all templates
cp .ai-pack/templates/task-packet/*.md "$TASK_DIR/"

# Verify templates copied
ls -la "$TASK_DIR"
# Should see: 00-contract.md, 10-plan.md, 20-work-log.md, 30-review.md, 40-acceptance.md
```

**Task packet format:**
```
.ai/tasks/<beads-id>-<YYYYMMDDHHMMSS>-<short-desc>/
├── 00-contract.md    # Requirements and acceptance criteria
├── 10-plan.md        # Implementation approach
├── 20-work-log.md    # Progress tracking
├── 30-review.md      # Quality assurance
└── 40-acceptance.md  # Final sign-off
```

### Step 3: Fill Contract (00-contract.md)

**Minimum required fields:**

```markdown
**Task ID:** <beads-id>-<YYYYMMDDHHMMSS>-<short-desc>
**Beads Task:** <beads-id>
**Created:** <YYYY-MM-DD>
**Requestor:** <User Name>
**Assigned Role:** <Architect|Engineer|Tester|Reviewer|Inspector>
**Workflow:** <Standard|Feature|Bugfix|Refactor|Research>

## Task Description
<What needs to be done>

## Success Criteria
✓ <Measurable outcome 1>
✓ <Measurable outcome 2>

## Acceptance Criteria
### Functional Requirements
□ <Requirement 1>
□ <Requirement 2>

### Quality Requirements
□ All tests passing
□ Code coverage ≥ 80%
□ No linting errors
```

**Use Write tool to fill out contract with actual task details.**

### Step 4: Determine Agent Role

**Role selection based on task type:**

| Task Type | Role | When to Use |
|-----------|------|-------------|
| Architecture review | **Architect** | Design decisions, API design, system architecture |
| Requirements gathering | **Product Manager** | Large features, unclear requirements, PRD creation |
| Code implementation | **Engineer** | Writing code, implementing features, bug fixes |
| Bug investigation | **Inspector** | Complex bugs, root cause unknown, pattern analysis |
| Test validation | **Tester** | Test coverage verification, TDD compliance |
| Code review | **Reviewer** | Code quality, standards compliance, security |
| UI/UX design | **Designer** | User-facing workflows, wireframes, mockups |

**Reference role definitions:** `.ai-pack/roles/<role>.md`

### Step 5: Spawn Agent with Contract

```python
# CRITICAL: Get working directory first
PROJECT_ROOT=$(pwd)

# Spawn agent with proper role and contract
Agent(
    subagent_type="general-purpose",
    description="<Brief task summary>",
    prompt=f"""You are working as <Role> following .ai-pack/roles/<role>.md

CRITICAL WORKING DIRECTORY CONTEXT:
- Repository root: {PROJECT_ROOT}
- Verify location with: pwd
- Use absolute paths for ALL file operations
- Example: Write(file_path="{PROJECT_ROOT}/path/to/file", content="...")

TASK CONTRACT: {PROJECT_ROOT}/.ai/tasks/<task-packet-dir>/00-contract.md

BEADS TASK: <beads-task-id>

Read the contract at 00-contract.md for complete requirements, acceptance criteria, and context.

DELIVERABLES:
- Update 20-work-log.md with progress
- Create all artifacts specified in contract
- Update Beads task status as you progress
- Report absolute paths of all files created

START: Read contract, fill out plan in 10-plan.md, then execute per your role.""",
    run_in_background=True  # For parallel execution
)
```

**Agent responsibilities:**
1. Read contract (00-contract.md)
2. Fill out plan (10-plan.md)
3. Execute work per role definition
4. Update work log (20-work-log.md)
5. Update Beads task status
6. Report completion

### Step 6: Monitor Progress

**Via Beads:**

```bash
# Check agent's Beads task status
bd show <beads-task-id>

# List all active agent work
bd list --status=in_progress

# Check for blocked work
bd list --status=blocked
```

**Via Work Log:**

```bash
# Read agent's progress updates
tail -f .ai/tasks/<task-packet-dir>/20-work-log.md
```

**Via Agent Output:**

```bash
# Check background agent output (if run_in_background=True)
# Agent tool returns output file path in result
tail -f <agent-output-file>
```

### Step 7: Quality Gates (For Code Changes)

**MANDATORY validations:**

1. **Tester validation** (if code/tests changed):
   ```bash
   # Spawn Tester agent
   Agent(
       subagent_type="general-purpose",
       description="Validate tests and coverage",
       prompt=f"""You are working as Tester following .ai-pack/roles/tester.md
   
   WORKING DIRECTORY: {PROJECT_ROOT}
   TASK: Validate tests for task {BEADS_TASK_ID}
   CONTRACT: {PROJECT_ROOT}/.ai/tasks/{TASK_DIR}/00-contract.md
   
   Verify:
   - TDD process followed
   - Test coverage ≥ 80%
   - All tests passing
   
   Write verdict to 30-review.md: APPROVED or CHANGES REQUIRED""",
       run_in_background=True
   )
   ```

2. **Reviewer validation** (if code changed):
   ```bash
   # Spawn Reviewer agent
   Agent(
       subagent_type="general-purpose",
       description="Review code quality",
       prompt=f"""You are working as Reviewer following .ai-pack/roles/reviewer.md
   
   WORKING DIRECTORY: {PROJECT_ROOT}
   TASK: Review code for task {BEADS_TASK_ID}
   CONTRACT: {PROJECT_ROOT}/.ai/tasks/{TASK_DIR}/00-contract.md
   
   Verify:
   - Code quality standards met
   - Security concerns addressed
   - Best practices followed
   
   Write verdict to 30-review.md: APPROVED or CHANGES REQUESTED""",
       run_in_background=True
   )
   ```

**Both must approve before closing Beads task.**

### Step 8: Completion

```bash
# Verify all acceptance criteria met (read 00-contract.md)
# Verify quality gates passed (read 30-review.md)
# Update acceptance document (40-acceptance.md)

# Close Beads task
bd close <beads-task-id>

# Verify closure
bd show <beads-task-id>
# Status should be: closed
```

## Enforcement Rule

**ALL task work MUST go through this workflow.**

**FORBIDDEN:**
- ❌ Creating tasks without Beads tracking
- ❌ Working on tasks without task packets
- ❌ Skipping quality gates for code changes
- ❌ Using TodoWrite or TaskCreate tools (use Beads instead)
- ❌ Working directly without creating task first

**REQUIRED:**
- ✅ Create Beads task BEFORE task packet
- ✅ Create task packet BEFORE implementation
- ✅ Fill contract BEFORE spawning agent
- ✅ Spawn agent with proper role definition
- ✅ Monitor via Beads task status
- ✅ Quality gates for all code changes
- ✅ Close Beads task when complete

## Parallel Execution (Multiple Tasks)

**When creating 2+ independent tasks:**

```bash
# Create multiple Beads tasks
task1=$(bd create --title="Task 1" --description="..." --priority=2 --json | jq -r '.id')
task2=$(bd create --title="Task 2" --description="..." --priority=2 --json | jq -r '.id')
task3=$(bd create --title="Task 3" --description="..." --priority=2 --json | jq -r '.id')

# Create multiple task packets
# (Create task packet for each task ID)

# Spawn multiple agents in SINGLE response
Agent(...)  # Task 1
Agent(...)  # Task 2
Agent(...)  # Task 3

# All agents run in parallel
# Monitor all via: bd list --status=in_progress
```

**Benefits:**
- 3 tasks × 20 minutes = 20 minutes total (not 60!)
- N-fold speedup for N independent tasks
- Beads tracks all work persistently

## Dependencies Between Tasks

```bash
# Create parent task
parent_id=$(bd create --title="Parent task" --json | jq -r '.id')

# Create child task
child_id=$(bd create --title="Child task" --json | jq -r '.id')

# Add dependency: child depends on parent
bd dep add $child_id $parent_id

# Check blocked tasks
bd blocked
# Will show child task blocked by parent

# When parent closes, child becomes ready
bd close $parent_id
bd ready  # Shows child task now available
```

## Task Recovery (Session Continuity)

**Tasks persist across sessions via Beads:**

```bash
# Next session - find ready work
bd ready

# Continue task
bd show <task-id>  # Read details
bd update <task-id> --claim  # Claim it

# Spawn agent to continue
Agent(
    description="Continue task <task-id>",
    prompt=f"""Continue working on task <task-id>
    
CONTRACT: {PROJECT_ROOT}/.ai/tasks/<task-dir>/00-contract.md
WORK LOG: {PROJECT_ROOT}/.ai/tasks/<task-dir>/20-work-log.md

Read work log to see what's been completed.
Continue from where previous agent left off.
Update work log with your progress."""
)
```

## Quick Reference

**Complete workflow:**
```bash
# 1. Create Beads task
task_id=$(bd create --title="..." --description="..." --type=task --priority=2 --json | jq -r '.id')

# 2. Create task packet
timestamp=$(date +%Y%m%d%H%M%S)
task_dir=".ai/tasks/${task_id}-${timestamp}-short-desc"
mkdir -p "$task_dir"
cp .ai-pack/templates/task-packet/*.md "$task_dir/"

# 3. Fill contract (use Write tool)

# 4. Spawn agent (use Agent tool with proper prompt)

# 5. Monitor
bd show $task_id
bd list --status=in_progress

# 6. Quality gates (spawn Tester + Reviewer if code changed)

# 7. Close
bd close $task_id
```

## Integration with CLAUDE.md

**This workflow enforces CLAUDE.md requirements:**

✅ **Task Packet Requirement** (Gate 1)
- Creates task packet before implementation

✅ **Beads Enforcement** (Gate 4)
- Uses `bd` for ALL task tracking
- NO TodoWrite or TaskCreate tools

✅ **Knowledge-First** (Gate 2)
- Contract includes KG Orientation section
- Agents search knowledge before file ops

✅ **Execution Strategy** (Gate 3)
- Documents parallel vs sequential in plan
- Spawns multiple agents for 3+ tasks

✅ **Code Quality Review** (Gate 5)
- Mandatory Tester + Reviewer validation
- Both must approve before completion

## References

- **Orchestrator Role:** `.ai-pack/roles/orchestrator.md`
- **Gates:** `.ai-pack/gates/*.md`
- **Workflows:** `.ai-pack/workflows/*.md`
- **Beads Docs:** Run `bd --help` or `bd prime`

---

Now proceed with creating and orchestrating tasks!
