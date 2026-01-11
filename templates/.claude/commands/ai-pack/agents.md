---
description: Show active agents, their roles, and task assignments
---

# /ai-pack agents - Agent Status

Display information about active agents (spawned workers) and their current roles/tasks.

## What This Shows

When you use the Orchestrator role to delegate work, it spawns agents using the Task tool. This command helps you:

1. **See what agents are running** - How many concurrent workers
2. **Understand their roles** - Engineer, Reviewer, Inspector, etc.
3. **Track their tasks** - What each agent is working on
4. **Monitor progress** - Which agents completed, which are active

## Usage

```bash
/ai-pack agents
```

No arguments needed - it scans for active agent processes.

## What Gets Reported

### Active Agents

For each running agent:
- **Agent ID** - Unique identifier
- **Role** - Engineer, Reviewer, Tester, Inspector, Architect, Designer, PM
- **Task** - Brief description of what they're working on
- **Status** - Running, Completed, Failed, Blocked
- **Started** - When the agent was spawned

### Agent Limits

- **Maximum concurrent**: 5 agents (framework limit)
- **Current active**: Shows count
- **Available slots**: Remaining capacity

### Shared Context

Reminds you that agents share:
- Source repository (no per-agent branches)
- Build folders
- Test coverage data
- Database (coordination required)

## Example Output

```
AI-Pack Agent Status
====================

Active Agents: 3 / 5 maximum

1. Agent: agent_abc123
   Role:    Engineer
   Task:    Implement login API endpoint
   Status:  Running
   Started: 2026-01-10 14:23:45

2. Agent: agent_def456
   Role:    Engineer
   Task:    Implement user profile API endpoint
   Status:  Running
   Started: 2026-01-10 14:23:45

3. Agent: agent_ghi789
   Role:    Reviewer
   Task:    Review authentication implementation
   Status:  Completed
   Started: 2026-01-10 14:30:12
   Ended:   2026-01-10 14:35:48

Available Slots: 2

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

**Debugging:**
- Agent didn't start as expected
- Too many agents spawned
- Coordination issues between agents

**Capacity planning:**
- Check if you can spawn more agents
- See if you've hit the 5-agent limit

## Implementation Notes

This command checks for:

1. **Background Task processes** - Agents run as background tasks
2. **Agent IDs in logs** - Searches for agent spawn records
3. **Task tool invocations** - Recent Task() calls
4. **Work log entries** - Notes in `.ai/tasks/*/20-work-log.md`

**Note:** Agent tracking depends on:
- Orchestrator documenting spawned agents in work log
- Background tasks being properly tracked
- Agent processes still being active

## Related Commands

- `/ai-pack task-status` - Overall task progress
- `/ai-pack orchestrate` - Spawn agents for complex tasks
- `/ai-pack help` - Show all commands

## Technical Details

### How Agents Work

When Orchestrator delegates:

```python
# Orchestrator spawns Engineer agent
Task(subagent_type="general-purpose",
     description="Implement login feature",
     prompt="Act as Engineer role. Implement login with TDD...")
```

This creates:
- A subprocess/background task
- An agent ID (for tracking)
- A dedicated execution context
- Access to all tools (Read, Write, Edit, Bash, etc.)

### Parallel Execution

For 3+ independent subtasks:

```python
# Spawn multiple agents in SINGLE message
Task(...) + Task(...) + Task(...)
```

All agents:
- Run concurrently
- Share the same codebase
- Must coordinate builds/tests
- Report back to Orchestrator

See: `.ai-pack/gates/25-execution-strategy.md`

## Coordination Best Practices

When multiple agents are active:

1. **Avoid concurrent file edits** - Agents shouldn't modify the same file simultaneously
2. **Coordinate builds** - Run builds sequentially, not in parallel
3. **Merge coverage** - Combine test coverage from all agents
4. **Use work log** - Document what each agent completed

## Troubleshooting

**No agents showing:**
- Agents may have already completed
- Orchestrator didn't spawn any (direct execution)
- Work log not updated with agent info

**Too many agents:**
- Maximum is 5 concurrent
- Wait for some to complete
- Consider sequential execution

**Agent coordination issues:**
- Check shared context constraints
- Review execution strategy gate
- Verify agents aren't conflicting

## References

- **Orchestrator Role:** [.ai-pack/roles/orchestrator.md](../../.ai-pack/roles/orchestrator.md)
- **Execution Strategy Gate:** [.ai-pack/gates/25-execution-strategy.md](../../.ai-pack/gates/25-execution-strategy.md)
- **Parallel Workers Config:** [.ai-pack/PARALLEL-ENGINEERS-CONFIG.md](../../.ai-pack/PARALLEL-ENGINEERS-CONFIG.md)

## How to Execute This Command

**CRITICAL: Check infrastructure exists first!**

This command requires the ai-pack infrastructure to be deployed. Before attempting to show agent status, you MUST:

1. **Check if agent status tracker exists:**
   ```bash
   test -f .claude/scripts/agent-status-tracker.py && echo "OK" || echo "NOT DEPLOYED"
   ```

2. **Check if work logs directory exists:**
   ```bash
   test -d .ai/tasks && echo "OK" || echo "NO TASKS"
   ```

3. **If infrastructure missing:**
   - Report: "No active agents - Orchestrator has not spawned any parallel workers yet"
   - OR: "ai-pack infrastructure not deployed - run update script first"
   - DO NOT attempt to run commands that will fail

4. **If infrastructure present:**
   - Run agent status tracker: `python3 .claude/scripts/agent-status-tracker.py report`
   - Check work logs: `find .ai/tasks -name "20-work-log.md" -exec tail -20 {} \;`
   - Look for spawn records: `grep -i "spawned\|delegated" .ai/tasks/*/20-work-log.md 2>/dev/null || echo "No spawn records"`

## Alternative: Check Via Work Log

If agent status tracker not available, check manually:

```bash
# Only if .ai/tasks exists
if [ -d .ai/tasks ]; then
  # Read work log
  find .ai/tasks -name "20-work-log.md" -exec tail -20 {} \;

  # Look for agent spawn records
  grep -i "spawned\|agent\|delegated" .ai/tasks/*/20-work-log.md 2>/dev/null
else
  echo "No task directories found - no agents have been spawned"
fi
```

Orchestrator should document:
- When agents were spawned
- What role each agent assumed
- What task each agent received
- When agents completed

---

**Note:** This command helps monitor agent-based orchestration. If you're working directly (not via Orchestrator), you won't have spawned agents to track. This command will gracefully report "No active agents" rather than failing with errors.
