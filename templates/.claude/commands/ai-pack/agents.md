---
description: Show active agents, their roles, and task assignments via Beads
---

# /ai-pack agents - Agent Status

Display information about active agents (spawned workers) by querying the Beads task tracking system.

## What This Shows

When the Orchestrator spawns agents for parallel execution, it creates corresponding Beads tasks. This command queries those tasks to show you:

1. **Active agents** - Workers currently in progress
2. **Their roles** - Engineer, Tester, Reviewer, etc.
3. **Their tasks** - What each agent is working on
4. **Progress** - Completed vs in-progress agents
5. **Blockers** - Any blocked agents

## Usage

```bash
/ai-pack agents           # Compact view with status icons (default)
/ai-pack agents --verbose # Full details with role and timestamps
/ai-pack agents -v        # Verbose form alias
```

**Options:**
- `--verbose` / `-v` - Show full details including role, assignee, and timestamps

## What Gets Reported

### Agent Tasks (from Beads)

For each agent:
- **Task ID** - Beads task ID (e.g., `bd-a1b2`)
- **Assignee** - Role and ID (e.g., `Engineer-1`)
- **Task** - What the agent is working on
- **Status** - `in_progress`, `closed`, or `blocked`
- **Started** - When the task began

### Agent Limits

- **Maximum concurrent**: 5 agents (framework limit)
- **Current active**: Agents with `in_progress` status
- **Available slots**: 5 minus active count

### Shared Context

Agents share:
- Source repository (no per-agent branches)
- Build folders and test coverage
- Database connections
- Coordination required per [Execution Strategy Gate](../../.ai-pack/gates/25-execution-strategy.md)

## Example Output

### Compact Format (Default)

```
STATUS      BEADS-ID      INTERNAL-ID                             DESCRIPTION
----------  ------------  --------------------------------------  -----------
RUNNING     bd-a1b2       task-engineer-20260129-143045-000000    Agent: Engineer - Implement login API
RUNNING     bd-b2c3       task-engineer-20260129-143048-000000    Agent: Engineer - User profile API
COMPLETED   bd-c3d4       task-reviewer-20260129-143050-000000    Agent: Reviewer - Review authentication
```

### Verbose Format (--verbose / -v)

```
AI-Pack Agent Status (Verbose)
==============================

Active Agents: 2 / 5 maximum

1. Task ID: bd-a1b2
   Assignee: Engineer-1
   Task:     Agent: Engineer - Implement login API
   Status:   in_progress
   Started:  2026-01-14 14:23:45

2. Task ID: bd-b2c3
   Assignee: Engineer-2
   Task:     Agent: Engineer - User profile API
   Status:   in_progress
   Started:  2026-01-14 14:23:45

Completed: 1
  - bd-c3d4: Agent: Reviewer - Review authentication (Reviewer-1)

Blocked: 0

Available capacity: 3 slots

Shared Context Reminder:
- All agents share the same source repository
- Coordinate builds and test runs
- No per-agent git branches
- See: .ai-pack/gates/25-execution-strategy.md
```

## When to Use This

**During orchestration:**
- After spawning parallel workers
- To verify agents started correctly
- To monitor progress
- To check for blockers

**Debugging:**
- Agent didn't register in Beads
- Too many agents spawned
- Coordination issues between agents

**Capacity planning:**
- Check available slots before spawning more agents
- Verify you haven't hit the 5-agent limit

**Quick status checks:**
- Default compact format shows what's running at a glance
- Use `--verbose` when debugging or need full details (assignee, timestamps, etc.)

## Implementation

This command queries Beads for tasks that match the agent naming pattern.

### Prerequisites

1. **Beads initialized:** `.beads/issues.jsonl` exists
2. **Agents registered:** Orchestrator creates Beads tasks when spawning
3. **Naming convention:** Agent tasks titled `"Agent: {Role} - {Description}"`

### Query Logic

```bash
# Check Beads initialized
test -f .beads/issues.jsonl || echo "Beads not initialized - run 'bd init'"

# Query all agent tasks (filter by naming pattern)
bd list --json | jq '.[] | select(.title | startswith("Agent:"))'

# Or filter by assignee pattern
bd list --json | jq '.[] | select(.assignee | test("Engineer-|Tester-|Reviewer-"))'
```

### Status Mapping

| Beads Status | Agent State | Meaning |
|--------------|-------------|---------|
| `in_progress` | Active | Agent currently working |
| `closed` | Completed | Agent finished successfully |
| `blocked` | Blocked | Agent stuck on dependency |
| `open` | Not Started | Task created but not claimed |

## How to Execute This Command

**Implementation using Beads:**

```bash
#!/bin/bash

# Parse arguments
VERBOSE_FORMAT=false
if [ "$1" == "--verbose" ] || [ "$1" == "-v" ]; then
  VERBOSE_FORMAT=true
fi

# Check Beads initialized
if [ ! -f .beads/issues.jsonl ]; then
  echo "No Beads database found - Orchestrator hasn't spawned agents yet"
  echo "Or run 'bd init' to initialize Beads"
  exit 0
fi

# Get agent tasks
ACTIVE=$(bd list --status in_progress --json 2>/dev/null | jq '.[] | select(.title | startswith("Agent:"))' 2>/dev/null)
ACTIVE_COUNT=$(echo "$ACTIVE" | jq -s 'length' 2>/dev/null || echo "0")

COMPLETED=$(bd list --status closed --json 2>/dev/null | jq '.[] | select(.title | startswith("Agent:"))' 2>/dev/null)
COMPLETED_COUNT=$(echo "$COMPLETED" | jq -s 'length' 2>/dev/null || echo "0")

BLOCKED=$(bd list --status blocked --json 2>/dev/null | jq '.[] | select(.title | startswith("Agent:"))' 2>/dev/null)
BLOCKED_COUNT=$(echo "$BLOCKED" | jq -s 'length' 2>/dev/null || echo "0")

AVAILABLE=$((5 - ACTIVE_COUNT))

# Compact format output (default)
if [ "$VERBOSE_FORMAT" != true ]; then
  echo "STATUS      BEADS-ID      INTERNAL-ID                             DESCRIPTION"
  echo "----------  ------------  --------------------------------------  -----------"

  # Show running agents
  if [ "$ACTIVE_COUNT" -gt 0 ]; then
    echo "$ACTIVE" | jq -r '"RUNNING     \(.id)            \(.assignee // "none")                                \(.title[:50])"'
  fi

  # Show completed agents
  if [ "$COMPLETED_COUNT" -gt 0 ]; then
    echo "$COMPLETED" | jq -r '"COMPLETED   \(.id)            \(.assignee // "none")                                \(.title[:50])"'
  fi

  # Show blocked agents
  if [ "$BLOCKED_COUNT" -gt 0 ]; then
    echo "$BLOCKED" | jq -r '"BLOCKED     \(.id)            \(.assignee // "none")                                \(.title[:50])"'
  fi

  # Show message if no agents
  if [ "$((ACTIVE_COUNT + COMPLETED_COUNT + BLOCKED_COUNT))" -eq 0 ]; then
    echo "No agents found"
  fi

  exit 0
fi

# Verbose format output
echo "AI-Pack Agent Status (via Beads)"
echo "=================================="
echo ""

echo "Active Agents: $ACTIVE_COUNT / 5 maximum"
echo ""

# Display active agents
if [ "$ACTIVE_COUNT" -gt 0 ]; then
  echo "$ACTIVE" | jq -r '
    . as $agent |
    "\(.id):\n  Assignee: \(.assignee)\n  Task:     \(.title)\n  Status:   \(.status)\n  Started:  \(.created_at)\n"
  '
fi

# Show completed agents
if [ "$COMPLETED_COUNT" -gt 0 ]; then
  echo "Completed: $COMPLETED_COUNT"
  echo "$COMPLETED" | jq -r '  "  - \(.id): \(.title) (\(.assignee))"'
  echo ""
fi

# Show blocked agents
echo "Blocked: $BLOCKED_COUNT"
if [ "$BLOCKED_COUNT" -gt 0 ]; then
  echo "$BLOCKED" | jq -r '  "  - \(.id): \(.title) - Reason: \(.blocked_reason)"'
fi
echo ""

# Calculate available capacity
echo "Available capacity: $AVAILABLE slots"
echo ""
echo "Shared Context Reminder:"
echo "- All agents share the same source repository"
echo "- Coordinate builds and test runs"
echo "- No per-agent git branches"
echo "- See: .ai-pack/gates/25-execution-strategy.md"
```

## Detailed Query Examples

**Show all agent tasks:**

```bash
bd list --json | jq '.[] | select(.title | startswith("Agent:"))'
```

**Show active Engineers only:**

```bash
bd list --status in_progress --assignee "Engineer-*"
```

**Quick compact format (status + IDs + title):**

```bash
bd list --json | jq -r '.[] | select(.title | startswith("Agent:")) | "\(.status | ascii_upcase)  \(.id)  \(.title)"'
```

**Show with formatted output:**

```bash
bd list --assignee "Engineer-*" --json | jq -r '
  "Active agents:",
  (.[] | select(.status == "in_progress") | "  \(.assignee): \(.title)"),
  "",
  "Progress: \([ .[] | select(.status == "closed") ] | length)/\(length) completed"
'
```

**Check specific agent:**

```bash
bd show bd-a1b2
```

## Related Commands

- `/ai-pack agents --verbose` - Full details with role and timestamps
- `/ai-pack task-status` - Overall task progress
- `/ai-pack orchestrate` - Spawn agents for complex tasks
- `bd list --assignee "Engineer-*"` - Direct Beads query
- `bd show <task-id>` - View agent details

## Troubleshooting

**"No Beads database found"**
- Beads not initialized: Run `bd init`
- No agents spawned yet: Orchestrator hasn't used parallel execution

**"No agent tasks found"**
- Orchestrator didn't create Beads tasks when spawning
- Check work logs for spawn records: `grep -i "spawned" .ai/tasks/*/20-work-log.md`
- Agent naming convention not followed

**"Agents showing but not actually running"**
- Agents completed but Orchestrator didn't close Beads tasks
- Run `bd close <task-id>` manually
- Check work logs to verify actual completion status

**Beads task exists but agent never started:**
- Beads task created but agent spawn failed
- Check Task tool errors in Orchestrator output
- Verify spawned agent permissions configured

## Agent Registration Protocol

Orchestrators MUST follow this protocol when spawning agents:

```bash
# 1. Spawn agent with Task tool
Task(subagent_type="general-purpose",
     description="Implement feature",
     prompt="...",
     )

# 2. Create Beads task immediately
bd create "Agent: Engineer - Implement feature" \
  --assignee "Engineer-1" \
  --priority high

# 3. Mark as in-progress
bd start bd-a1b2

# 4. Document in work log
echo "Spawned Engineer-1 (Beads ID: bd-a1b2)" >> .ai/tasks/*/20-work-log.md
```

See: [Orchestrator Role - Section 2.13](../../.ai-pack/roles/orchestrator.md#213-agent-registration-protocol-mandatory)

## Migration Note

**Legacy:** This command previously used `.claude/.agent-status.json` managed by `agent-status-tracker.py`.

**Current:** Now queries Beads (`.beads/issues.jsonl`) for git-backed, cross-session persistence.

**Backward Compatibility:** The old tracker is deprecated but still functional. See migration guide: `scripts/migrate-agent-status-to-beads.py`

## References

- **Orchestrator Role:** [.ai-pack/roles/orchestrator.md](../../.ai-pack/roles/orchestrator.md)
- **Beads Integration:** [.ai-pack/quality/tooling/beads-integration.md](../../.ai-pack/quality/tooling/beads-integration.md)
- **Execution Strategy Gate:** [.ai-pack/gates/25-execution-strategy.md](../../.ai-pack/gates/25-execution-strategy.md)
- **Beads Documentation:** https://github.com/steveyegge/beads

---

**Note:** This command provides visibility into agent-based orchestration. If working directly (not via Orchestrator), there won't be agent tasks to display.
