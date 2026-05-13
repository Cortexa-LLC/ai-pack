# Orchestrate Skill

**Use MCP agent tools — NOT the built-in Agent() — when delegating multi-step tasks.**

This skill tells you exactly how to spawn and coordinate ai-pack agents.

---

## The Golden Rule

Claude Code has a built-in `Agent()` tool (shown as "Agent" in tool listings).
**Do not use it for ai-pack orchestration.** It does not track tasks, does not
write to the beads task system, and leaves no audit trail.

Instead, use the three MCP tools provided by the `agent-mcp` server:
`create_task` → `spawn_agent` → `get_task_status` / `get_task_logs`.

---

## Tool Reference

### `create_task`

Create a task in the ai-pack task system. Always do this first.

```json
{
  "description": "Refactor the authentication module to use JWT",
  "role": "engineer",
  "priority": "P1"
}
```

Returns `{ "task_id": "ai-pack-a1b2" }`.  Save this ID — you'll need it.

**Roles:** `engineer` | `architect` | `reviewer` | `spelunker`
**Priority:** `P0` (critical) → `P4` (low). Omit for default (P2).

---

### `spawn_agent`

Start an agent to work the task.

```json
{
  "task_id": "ai-pack-a1b2",
  "role": "engineer",
  "stream": true
}
```

**`stream: true` (default)** — blocks until the agent completes and returns all
output. Use this for sequential pipelines where each step depends on the previous.

**`stream: false`** — fires the agent and returns immediately with a confirmation
message. Use this for parallel tasks that can run concurrently.

---

### `get_task_status`

Check whether a non-streaming task has finished.

```json
{ "task_id": "ai-pack-a1b2" }
```

---

### `get_task_logs`

Retrieve the tail of a task's log output.

```json
{ "task_id": "ai-pack-a1b2", "lines": 100 }
```

---

### `list_tasks`

Survey what's running or recently completed.

```json
{ "status": "running" }
```

Valid statuses: `all` | `running` | `completed` | `failed` | `open`.

---

## Standard Orchestration Patterns

### Sequential Pipeline

Use when each step depends on the previous step's output.

```
1. create_task("Write failing tests for the auth module", role="engineer")
   → task_id = "ai-pack-aa11"

2. spawn_agent(task_id="ai-pack-aa11", role="engineer", stream=true)
   (blocks until tests are written)

3. create_task("Implement auth module to pass the tests", role="engineer")
   → task_id = "ai-pack-bb22"

4. spawn_agent(task_id="ai-pack-bb22", role="engineer", stream=true)
   (blocks until implementation is done)

5. create_task("Review the auth implementation", role="reviewer")
   → task_id = "ai-pack-cc33"

6. spawn_agent(task_id="ai-pack-cc33", role="reviewer", stream=true)
```

### Parallel Fan-Out

Use when tasks are independent and can run concurrently.

```
1. create_task("Write tests for module A", role="engineer") → id_a
2. create_task("Write tests for module B", role="engineer") → id_b
3. create_task("Write tests for module C", role="engineer") → id_c

4. spawn_agent(task_id=id_a, role="engineer", stream=false)  ← fire
5. spawn_agent(task_id=id_b, role="engineer", stream=false)  ← fire
6. spawn_agent(task_id=id_c, role="engineer", stream=false)  ← fire

7. [wait / poll get_task_status for each]

8. Once all done → create and spawn a review task
```

### Investigate-then-Fix

Use the spelunker role to understand a system before engineering.

```
1. create_task("Investigate why login times out on mobile", role="spelunker")
   → task_id = "ai-pack-dd44"

2. spawn_agent(task_id="ai-pack-dd44", role="spelunker", stream=true)
   (returns investigation report)

3. create_task("Fix login timeout based on investigation", role="engineer")
   → task_id = "ai-pack-ee55"

4. spawn_agent(task_id="ai-pack-ee55", role="engineer", stream=true)
```

---

## Decision Guide: stream=true vs stream=false

| Scenario | stream |
|---|---|
| Step B depends on Step A's output | `true` |
| Multiple independent tasks | `false` |
| You need the final output in this conversation | `true` |
| You just want to kick off background work | `false` |
| Reviewer must see completed code | `true` (after engineer) |

---

## Anti-Patterns to Avoid

❌ **Do not use the built-in Agent() tool** for ai-pack task delegation.

❌ **Do not skip create_task** — spawn_agent requires a task_id.

❌ **Do not run the agent CLI directly in Bash** when MCP tools are available.
   Use `create_task` + `spawn_agent` so tasks appear in the task system.

❌ **Do not spawn agents in a tight loop without polling** — parallel tasks
   can overwhelm the system. Use `list_tasks` to check capacity.

---

## Quick Reference

```
create_task(description, role, priority?)  → { task_id }
spawn_agent(task_id, role, stream?)        → output string
get_task_status(task_id)                   → status string
get_task_logs(task_id, lines?)             → log tail string
list_tasks(status?)                        → task list string
```
