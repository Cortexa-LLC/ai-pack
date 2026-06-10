---
description: Show active agents, their roles, and task assignments
---

# /ai-pack agents - Agent Status

Display information about active agents (spawned workers) by querying the task tracking system.

## What This Shows

When the Orchestrator spawns agents for parallel execution, it creates corresponding tasks. This command queries those tasks to show you:

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

### Agent Tasks (from task database)

For each agent:
- **Task ID** - task ID (e.g., `ai-pack-a1b2`)
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
STATUS      TASK-ID           DESCRIPTION
----------  ----------------  -----------
RUNNING     ai-pack-a1b2      Agent: Engineer - Implement login API
RUNNING     ai-pack-b2c3      Agent: Engineer - User profile API
COMPLETED   ai-pack-c3d4      Agent: Reviewer - Review authentication
```

### Verbose Format (--verbose / -v)

```
AI-Pack Agent Status (Verbose)
==============================

Active Agents: 2 / 5 maximum

1. Task ID: ai-pack-a1b2
   Assignee: Engineer-1
   Task:     Agent: Engineer - Implement login API
   Status:   in_progress
   Started:  2026-01-14 14:23:45

2. Task ID: ai-pack-b2c3
   Assignee: Engineer-2
   Task:     Agent: Engineer - User profile API
   Status:   in_progress
   Started:  2026-01-14 14:23:45

Completed: 1
  - ai-pack-c3d4: Agent: Reviewer - Review authentication (Reviewer-1)

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
- Agent didn't register in task database
- Too many agents spawned
- Coordination issues between agents

**Capacity planning:**
- Check available slots before spawning more agents
- Verify you haven't hit the 5-agent limit

**Quick status checks:**
- Default compact format shows what's running at a glance
- Use `--verbose` when debugging or need full details (assignee, timestamps, etc.)

## Implementation

This command queries the `agent` CLI for tasks that match the agent naming pattern.

### Prerequisites

1. **Agents registered:** Orchestrator creates tasks when spawning
2. **Naming convention:** Agent tasks titled `"Agent: {Role} - {Description}"`

### Query Logic

```bash
# Query all agent tasks (filter by naming pattern)
agent list --json | jq '.[] | select(.title | startswith("Agent:"))'

# Or filter by assignee pattern
agent list --json | jq '.[] | select(.assignee | test("Engineer-|Tester-|Reviewer-"))'
```

### Status Mapping

| Agent CLI Status | Agent State | Meaning |
|------------------|-------------|---------|
| `in_progress` | Active | Agent currently working |
| `closed` | Completed | Agent finished successfully |
| `blocked` | Blocked | Agent stuck on dependency |
| `open` | Not Started | Task created but not claimed |

## How to Execute This Command

**Implementation using agent CLI:**

```bash
#!/bin/bash

# Parse arguments
VERBOSE_FORMAT=false
if [ "$1" == "--verbose" ] || [ "$1" == "-v" ]; then
  VERBOSE_FORMAT=true
fi

# Get agent tasks
ACTIVE=$(agent list --status in_progress --json 2>/dev/null | jq '.[] | select(.title | startswith("Agent:"))' 2>/dev/null)
ACTIVE_COUNT=$(echo "$ACTIVE" | jq -s 'length' 2>/dev/null || echo "0")

COMPLETED=$(agent list --status closed --json 2>/dev/null | jq '.[] | select(.title | startswith("Agent:"))' 2>/dev/null)
COMPLETED_COUNT=$(echo "$COMPLETED" | jq -s 'length' 2>/dev/null || echo "0")

BLOCKED=$(agent list --status blocked --json 2>/dev/null | jq '.[] | select(.title | startswith("Agent:"))' 2>/dev/null)
BLOCKED_COUNT=$(echo "$BLOCKED" | jq -s 'length' 2>/dev/null || echo "0")

AVAILABLE=$((5 - ACTIVE_COUNT))

# Compact format output (default)
if [ "$VERBOSE_FORMAT" != true ]; then
  echo "STATUS      TASK-ID           DESCRIPTION"
  echo "----------  ----------------  -----------"

  # Show running agents
  if [ "$ACTIVE_COUNT" -gt 0 ]; then
    echo "$ACTIVE" | jq -r '"RUNNING     \(.id)  \(.title[:50])"'
  fi

  # Show completed agents
  if [ "$COMPLETED_COUNT" -gt 0 ]; then
    echo "$COMPLETED" | jq -r '"COMPLETED   \(.id)  \(.title[:50])"'
  fi

  # Show blocked agents
  if [ "$BLOCKED_COUNT" -gt 0 ]; then
    echo "$BLOCKED" | jq -r '"BLOCKED     \(.id)  \(.title[:50])"'
  fi

  # Show message if no agents
  if [ "$((ACTIVE_COUNT + COMPLETED_COUNT + BLOCKED_COUNT))" -eq 0 ]; then
    echo "No agents found"
  fi

  exit 0
fi

# Verbose format output
echo "AI-Pack Agent Status"
echo "=================================="
echo ""

echo "Active Agents: $ACTIVE_COUNT / 5 maximum"
echo ""

# Display active agents
if [ "$ACTIVE_COUNT" -gt 0 ]; then
  echo "$ACTIVE" | jq -r '
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
  echo "$BLOCKED" | jq -r '  "  - \(.id): \(.title)"'
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
agent list --json | jq '.[] | select(.title | startswith("Agent:"))'
```

**Show active Engineers only:**

```bash
agent list --status in_progress --assignee "Engineer-*"
```

**Quick compact format (status + IDs + title):**

```bash
agent list --json | jq -r '.[] | select(.title | startswith("Agent:")) | "\(.status | ascii_upcase)  \(.id)  \(.title)"'
```

**Show with formatted output:**

```bash
agent list --assignee "Engineer-*" --json | jq -r '
  "Active agents:",
  (.[] | select(.status == "in_progress") | "  \(.assignee): \(.title)"),
  "",
  "Progress: \([ .[] | select(.status == "closed") ] | length)/\(length) completed"
'
```

**Check specific agent:**

```bash
agent show ai-pack-a1b2
```

## Related Commands

- `/ai-pack agents --verbose` - Full details with role and timestamps
- `/ai-pack task-status` - Overall task progress
- `/ai-pack orchestrate` - Spawn agents for complex tasks
- `agent list --assignee "Engineer-*"` - Direct query
- `agent show <task-id>` - View agent details

## Troubleshooting

**"No agent tasks found"**
- Orchestrator didn't create tasks when spawning
- Check work logs for spawn records: `grep -i "spawned" .ai/tasks/*/result.md`
- Agent naming convention not followed

**"Agents showing but not actually running"**
- Agents completed but Orchestrator didn't close tasks
- Run `agent close <task-id>` manually
- Check work logs to verify actual completion status

**Task exists but agent never started:**
- Task created but agent spawn failed
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

# 2. Create task immediately
task_id=$(agent create "Agent: Engineer - Implement feature" \
  --assignee "Engineer-1" \
  --priority high --json | jq -r '.id')

# 3. Mark as in-progress
agent update --claim $task_id

# 4. Document in work log
echo "Spawned Engineer-1 (Task ID: $task_id)" >> .ai/tasks/*/result.md
```

See: [Orchestrator Role - Section 2.13](../../.ai-pack/roles/orchestrator.md#213-agent-registration-protocol-mandatory)

## References

- **Orchestrator Role:** [.ai-pack/roles/orchestrator.md](../../.ai-pack/roles/orchestrator.md)
- **Execution Strategy Gate:** [.ai-pack/gates/25-execution-strategy.md](../../.ai-pack/gates/25-execution-strategy.md)

---

**Note:** This command provides visibility into agent-based orchestration. If working directly (not via Orchestrator), there won't be agent tasks to display.
