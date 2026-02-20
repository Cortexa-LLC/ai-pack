# Beads Integration Guide

**Version:** 1.0
**Last Updated:** 2026-01-11
**Purpose:** Guide for using Steve Yegge's Beads task memory system with ai-pack workflows

---

## Overview

**Beads** is a coding agent memory system that provides persistent, git-backed task tracking. It replaces session-based TodoWrite with cross-session memory that survives conversation boundaries.

**Key Benefits:**
- ✅ Tasks persist across AI sessions (solves "50 First Dates" problem)
- ✅ Git-backed JSONL storage in `.beads/issues.jsonl`
- ✅ Dependency tracking with task graphs
- ✅ Hash-based IDs prevent merge collisions in multi-agent workflows
- ✅ Cross-platform (Windows, macOS, Linux, FreeBSD)
- ✅ Agent-native JSON format
- ✅ Automatic sync via git pull/push

---

## Installation

### Quick Install

**macOS/Linux/FreeBSD:**
```bash
curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/scripts/install.sh | bash
```

**Windows:**
```powershell
irm https://raw.githubusercontent.com/steveyegge/beads/main/install.ps1 | iex
```

**Homebrew (macOS/Linux):**
```bash
brew tap steveyegge/beads
brew install bd
```

**Go Install:**
```bash
go install github.com/steveyegge/beads/cmd/bd@latest
```

### Verify Installation
```bash
bd version
bd help
```

---

## Project Setup

### Initialize Beads (Simple Single-Branch Mode)

By default, Beads commits directly to your current branch (typically `main`):

```bash
cd /path/to/project
bd init
```

This creates:
- `.beads/issues.jsonl` - Git-tracked task database (JSONL format)
- `.beads/*.db` - Local SQLite cache (git-ignored)
- `.beads/bd.sock` - Per-project daemon socket

**What gets committed to git:**
- ✅ `.beads/issues.jsonl` - The task database (source of truth)

**What's git-ignored:**
- ❌ `.beads/*.db` - SQLite cache (regenerated from JSONL)
- ❌ `.beads/bd.sock` - Daemon socket

### Automatic Sync Behavior

After `bd init`, Beads automatically:
1. **Export** - Changes debounce to JSONL after 5 seconds
2. **Import** - Auto-detects when `.beads/issues.jsonl` is newer (after `git pull`)
3. **Commit** - Commits `.beads/issues.jsonl` to your current branch

**No additional branches are created.** All work happens on the branch you're already using.

---

## Core Commands for AI Agents

### Task Creation

```bash
# Create a task with priority
bd create "Implement user authentication" --priority high

# Create with description
bd create "Add dark mode toggle" --description "Support system theme preference"

# Create with dependencies (reference existing task)
bd create "Write authentication tests" --depends-on bd-a1b2
```

**Priorities:** `critical`, `high`, `normal`, `low`

### Finding Work (Most Important Command)

```bash
# Show all tasks ready to work on (no blocking dependencies)
bd ready

# Show ready tasks with high priority
bd ready --priority high

# JSON output (for agent parsing)
bd ready --json
```

**This replaces:** Looking at a todo list to see what's next.

### Task Status Updates

```bash
# Start working on a task
bd start bd-a1b2

# Mark task as complete
bd close bd-a1b2

# Mark as blocked
bd block bd-a1b2 "Waiting for API documentation"

# Unblock a task
bd unblock bd-a1b2
```

### Task Information

```bash
# Show full task details
bd show bd-a1b2

# Show task with change history
bd show bd-a1b2 --history

# List all tasks
bd list

# List tasks by status
bd list --status open
bd list --status closed
```

### Dependency Management

```bash
# Add dependency (bd-a1b2 must complete before bd-f14c can start)
bd dep add bd-f14c bd-a1b2

# Remove dependency
bd dep rm bd-f14c bd-a1b2

# Show task dependencies
bd show bd-f14c
```

### Search and Filter

```bash
# Search task content
bd search "authentication"

# List by assignee (for multi-agent workflows)
bd list --assignee "Engineer"

# Filter by priority
bd list --priority high
```

---

## Workflow Integration

### Orchestrator Workflow

**Phase 1: Task Decomposition**
```bash
# Break down user request into tasks
bd create "Phase 1: Requirements analysis" --priority high
bd create "Phase 2: Design API endpoints" --priority high
bd create "Phase 3: Implement authentication service" --priority normal
bd create "Phase 4: Write integration tests" --priority normal

# Set up dependencies
bd dep add bd-b2c3 bd-a1b2  # Design depends on requirements
bd dep add bd-c3d4 bd-b2c3  # Implementation depends on design
bd dep add bd-d4e5 bd-c3d4  # Tests depend on implementation
```

**Phase 2: Work Coordination**
```bash
# Find next available task
bd ready

# Delegate to Engineer
# Agent sees bd-a1b2 is ready and starts it
bd start bd-a1b2
```

**Phase 3: Progress Tracking**
```bash
# Check overall progress
bd list --status open

# Verify dependencies satisfied
bd show bd-c3d4  # Check if dependencies are closed
```

### Engineer Workflow

**Step 1: Find Next Task**
```bash
# What can I work on?
bd ready

# Get task details
bd show bd-a1b2
```

**Step 2: Start Work**
```bash
# Mark task as in-progress
bd start bd-a1b2
```

**Step 3: During Implementation**
```bash
# Discover subtasks during work
bd create "Add password hashing utility" --depends-on bd-a1b2
bd create "Configure JWT secrets" --depends-on bd-a1b2

# Block on external dependency
bd block bd-a1b2 "Waiting for security audit approval"
```

**Step 4: Complete Work**
```bash
# Mark task complete
bd close bd-a1b2

# Find next task
bd ready
```

### Parallel Worker Coordination

When Orchestrator spawns multiple parallel workers:

```bash
# Orchestrator creates independent tasks
bd create "Implement user service" --priority high --assignee "Worker-1"
bd create "Implement product service" --priority high --assignee "Worker-2"
bd create "Implement order service" --priority high --assignee "Worker-3"

# Workers find their tasks
bd ready --assignee "Worker-1"  # Each worker filters by assignee

# Workers start and complete independently
# Hash-based IDs prevent collisions even if workers create tasks simultaneously
```

---

## Integration with ai-pack Roles

### Orchestrator Role

**Uses Beads for:**
- Task decomposition (replaces TodoWrite)
- Work delegation (assign to Engineers)
- Progress monitoring (bd list, bd ready)
- Dependency management (bd dep add)

**Key Commands:**
- `bd create` - Create tasks during planning
- `bd dep add` - Define task dependencies
- `bd ready` - Find next available work
- `bd list --status open` - Monitor progress

### Engineer Role

**Uses Beads for:**
- Finding next task (bd ready)
- Marking progress (bd start, bd close)
- Creating subtasks during implementation
- Blocking on external dependencies

**Key Commands:**
- `bd ready` - "What should I work on?"
- `bd start <id>` - "Starting this task"
- `bd close <id>` - "Task complete"
- `bd show <id>` - "What's this task about?"

### Tester Role

**Uses Beads for:**
- Tracking test validation tasks
- Marking test pass/fail status
- Creating test remediation tasks

**Key Commands:**
- `bd create` - Create test validation tasks
- `bd close` - Mark tests as passing
- `bd create --depends-on` - Create fix tasks for test failures

### Reviewer Role

**Uses Beads for:**
- Tracking code review tasks
- Creating tasks for requested changes
- Marking review complete

**Key Commands:**
- `bd create` - Create review tasks
- `bd block` - Block on changes requested
- `bd close` - Mark review approved

---

## Agent Coordination with Beads

**Use Case:** When Orchestrator spawns parallel spawned agents for execution.

### Orchestrator Pattern for Agent Tracking

When spawning agents with the Task tool, create corresponding Beads tasks:

```bash
# 1. Spawn agent with Task tool
Task(subagent_type="general-purpose",
     description="Implement login feature",
     prompt="Act as Engineer. Implement login per task packet...",
     )

# 2. Create Beads task IMMEDIATELY
bd create "Agent: Engineer - Implement login feature" \
  --assignee "Engineer-1" \
  --priority high \
  --description "Task packet: .ai/tasks/ai-pack-4ef-20260114090000-login/"

# Returns task ID, e.g., bd-a1b2

# 3. Mark as in-progress
bd start bd-a1b2

# 4. Document in work log
echo "Spawned Engineer-1 (Beads ID: bd-a1b2)" >> .ai/tasks/*/20-work-log.md
```

### Agent Naming Convention

**Pattern:** `"Agent: {Role} - {Task Description}"`

**Examples:**
```bash
bd create "Agent: Engineer - Implement user profile API" --assignee "Engineer-1"
bd create "Agent: Tester - Validate authentication tests" --assignee "Tester-1"
bd create "Agent: Reviewer - Review login implementation" --assignee "Reviewer-1"
```

### Monitoring Active Agents

**Check agent status:**

```bash
# List all active agents
bd list --status in_progress --assignee "Engineer-*"
bd list --status in_progress --assignee "Tester-*"
bd list --status in_progress --assignee "Reviewer-*"

# Or use the /ai-pack agents command
/ai-pack agents
```

**Detailed status with jq:**

```bash
bd list --assignee "Engineer-*" --json | jq -r '
  "Active agents:",
  (.[] | select(.status == "in_progress") | "  \(.assignee): \(.title)"),
  "",
  "Progress: \([ .[] | select(.status == "closed") ] | length)/\(length) completed"
'
```

### Worker Pattern for Beads Updates

**Engineers/Testers/Reviewers spawned by Orchestrator should:**

```bash
# Find your assigned Beads task (documented in work log)
grep "Beads ID:" .ai/tasks/*/20-work-log.md

# Example output: "Spawned Engineer-1 (Beads ID: bd-a1b2)"

# Update when blocked
bd block bd-a1b2 "Waiting for API credentials"

# Unblock when resolved
bd unblock bd-a1b2

# Mark complete when finished
bd close bd-a1b2
```

### Benefits of Beads for Agent Coordination

1. **Cross-session persistence** - Agent tasks survive session boundaries
2. **Git-backed audit trail** - All agent activity tracked in `.beads/issues.jsonl`
3. **Real-time status** - Orchestrator monitors via `bd list` and `/ai-pack agents`
4. **Collision-free IDs** - Hash-based IDs prevent conflicts in parallel execution
5. **Dependency tracking** - Can link agent tasks to feature tasks with `bd dep add`

### Agent Status Mapping

| Beads Status | Agent State | Meaning |
|--------------|-------------|---------|
| `in_progress` | Active | Agent currently working |
| `closed` | Completed | Agent finished successfully |
| `blocked` | Blocked | Agent stuck on dependency/issue |
| `open` | Not Started | Task created but agent not spawned yet |

### Migration from agent-status-tracker.py

**Legacy system (DEPRECATED):**
```bash
python3 .claude/scripts/agent-status-tracker.py register "engineer-1" "Engineer" "Task" "orchestrator-123"
python3 .claude/scripts/agent-status-tracker.py report
```

**New Beads system:**
```bash
bd create "Agent: Engineer - Task" --assignee "Engineer-1" --priority high
bd start bd-a1b2
bd list --assignee "Engineer-*" --status in_progress
```

**Automated migration:** Use `scripts/migrate-agent-status-to-beads.py`

---

## Task Lifecycle

```
┌─────────────────────────────────────────────────────────┐
│                    TASK LIFECYCLE                        │
└─────────────────────────────────────────────────────────┘

  bd create
      │
      ▼
  [OPEN] ──────────────┐
      │                │
      │ bd start       │ bd block
      ▼                ▼
  [IN_PROGRESS]    [BLOCKED]
      │                │
      │                │ bd unblock
      │                ▼
      │            [OPEN]
      │                │
      │ bd close       │
      ▼                │
  [CLOSED] ◄───────────┘
```

**Status Meanings:**
- **OPEN** - Task created, ready to start (if no dependencies)
- **IN_PROGRESS** - Someone is actively working on it
- **BLOCKED** - Cannot proceed due to external dependency
- **CLOSED** - Task completed successfully

---

## Multi-Session Continuity

### The "50 First Dates" Problem (SOLVED)

**Before Beads:**
- AI agent: "What was I working on yesterday?"
- User: "You were implementing authentication..."
- Agent starts fresh every session

**With Beads:**
```bash
# New AI session, agent asks "What should I do?"
bd ready
# Output: bd-a1b2 "Implement user authentication" [IN_PROGRESS]

bd show bd-a1b2
# Output: Full context including description, dependencies, history
```

The agent **remembers** because task state persists in git.

### Cross-Machine Sync

Work on multiple machines:

```bash
# Machine A
bd create "Add feature X"
git add .beads/issues.jsonl
git commit -m "Add feature X task"
git push

# Machine B
git pull  # Beads auto-imports new tasks
bd ready  # See tasks created on Machine A
```

---

## Advanced Features (For Later Use)

### Protected Branch Workflow (NOT CURRENTLY USED)

**When to use:** If your `main` branch has GitHub/GitLab protection rules requiring PRs.

**Setup:**
```bash
bd init --branch beads-sync
```

This commits Beads metadata to a separate `beads-sync` branch using git worktrees, keeping `main` clean.

**Configuration:**
```bash
# Switch to sync branch model
bd config set sync.branch beads-sync

# Revert to direct commits
bd config set sync.branch ""
```

**Documentation:** See Beads docs/PROTECTED_BRANCHES.md for details.

**NOTE:** ai-pack currently uses **single-branch mode** (direct commits to `main`). Protected branch workflow is documented here for future reference but is not actively used.

---

## Best Practices

### 1. Task Granularity

**Good task size:**
```bash
bd create "Implement login endpoint"
bd create "Add JWT token generation"
bd create "Write authentication tests"
```

**Too large:**
```bash
bd create "Build entire authentication system"
```

**Too small:**
```bash
bd create "Add import statement"
bd create "Rename variable"
```

### 2. Use Dependencies Wisely

```bash
# Clear dependency chain
bd create "Design database schema" --priority high
bd create "Implement data models" --depends-on bd-a1b2
bd create "Write migration scripts" --depends-on bd-b2c3
```

### 3. Keep Tasks Updated

```bash
# If you discover a blocker, mark it
bd block bd-a1b2 "Waiting for API key from DevOps"

# When unblocked, update immediately
bd unblock bd-a1b2
```

### 4. Use Priorities

```bash
# Critical path items
bd create "Fix production bug" --priority critical

# Nice-to-haves
bd create "Add dark mode" --priority low
```

### 5. Search is Your Friend

```bash
# Find related tasks
bd search "authentication"

# Check what's already done
bd list --status closed | grep "auth"
```

---

## Common Patterns

### Pattern: Epic with Subtasks

```bash
# Create epic
bd create "User Authentication System" --priority high
EPIC_ID=$(bd list --json | jq -r '.[0].id')

# Create subtasks
bd create "Design auth API" --depends-on $EPIC_ID
bd create "Implement JWT tokens" --depends-on $EPIC_ID
bd create "Add password hashing" --depends-on $EPIC_ID
bd create "Write auth tests" --depends-on $EPIC_ID
```

### Pattern: Sequential Workflow

```bash
# Phase-based dependencies
bd create "Phase 1: Research" --priority high
bd create "Phase 2: Design" --depends-on bd-a1b2
bd create "Phase 3: Implement" --depends-on bd-b2c3
bd create "Phase 4: Test" --depends-on bd-c3d4
bd create "Phase 5: Deploy" --depends-on bd-d4e5
```

### Pattern: Parallel Work with Merge

```bash
# Independent parallel tasks
bd create "Build frontend component"
bd create "Build backend API"
bd create "Build database layer"

# Final integration task depends on all
bd create "Integration testing" --depends-on bd-a1b2 --depends-on bd-b2c3 --depends-on bd-c3d4
```

---

## Comparison with TodoWrite (Legacy)

| Feature | TodoWrite (Old) | Beads (New) |
|---------|----------------|-------------|
| **Persistence** | Session-only | Git-backed, permanent |
| **Cross-session** | ❌ Lost on new conversation | ✅ Survives sessions |
| **Dependencies** | ❌ Not supported | ✅ Full dependency graphs |
| **Multi-agent** | ⚠️ Shared context only | ✅ Hash IDs prevent collisions |
| **Search** | ❌ No search | ✅ Full-text search |
| **History** | ❌ No history | ✅ Full change history |
| **Format** | Internal tool | JSON/JSONL standard |
| **Git integration** | ❌ Not versioned | ✅ Versioned with code |
| **Priority** | ✅ Supported | ✅ Supported |
| **Status** | ✅ Supported | ✅ Enhanced (open/in_progress/blocked/closed) |

---

## Troubleshooting

### "bd: command not found"

**Solution:**
```bash
# Add Go bin to PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Or reinstall
curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/scripts/install.sh | bash
```

### Tasks not syncing across machines

**Solution:**
```bash
# Ensure .beads/issues.jsonl is committed
git add .beads/issues.jsonl
git commit -m "Update task status"
git push

# On other machine
git pull  # Auto-imports tasks
```

### "Database out of sync"

**Solution:**
```bash
# Regenerate database from JSONL
bd import
```

### Merge conflicts in .beads/issues.jsonl

**Solution:**
```bash
# Keep the line with the newer `updated_at` timestamp
# Or keep both lines if they're different tasks
# After resolving, run:
bd import
```

---

## References

- **Beads GitHub:** https://github.com/steveyegge/beads
- **Documentation:** https://github.com/steveyegge/beads/tree/main/docs
- **Installation:** https://github.com/steveyegge/beads/blob/main/docs/INSTALLING.md
- **FAQ:** https://github.com/steveyegge/beads/blob/main/docs/FAQ.md
- **Protected Branches:** https://github.com/steveyegge/beads/blob/main/docs/PROTECTED_BRANCHES.md

---

## Summary

**For Orchestrators:**
- Use `bd create` to decompose tasks
- Use `bd dep add` to define dependencies
- Use `bd ready` to find next work
- Use `bd list` to monitor progress

**For Engineers:**
- Use `bd ready` to find next task
- Use `bd start` to begin work
- Use `bd close` to complete work
- Use `bd show` to get task details

**Key Advantage:**
Tasks persist across AI sessions, machines, and team members through git-backed storage.

---

**Last Updated:** 2026-01-11
**Next Review:** When protected branch workflow is needed
