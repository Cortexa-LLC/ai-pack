---
name: orchestrate
description: >
  Decompose and delegate engineering work to specialized sub-agents. Use when the user
  describes a task that involves implementation, design, review, or codebase investigation.
  Triggers on multi-step work: "build this feature", "fix this bug and add tests",
  "design the architecture for X then implement it", "investigate why X is broken then fix it".
  <example>build the user authentication feature end to end</example>
  <example>fix the race condition in the job queue and add a regression test</example>
  <example>design and implement the new notification system</example>
  <example>investigate why requests are timing out then fix the root cause</example>
  <example>review the PR changes and fix any blocking issues</example>
---

## Agent Roster

Choose the right agent for each job:

| Agent | Use when |
|-------|----------|
| **engineer** | Writing code, fixing bugs, creating tests — any task that modifies files |
| **architect** | Design decisions, new integrations, ADRs, evaluating feasibility — BEFORE engineering |
| **reviewer** | Code review, security checks, quality gates — AFTER engineering, before merging |
| **spelunker** | Investigation when root cause or implementation path is unclear — BEFORE engineering |
| **inspector** | Complex bug root cause analysis when the bug is non-obvious or multi-module |
| **product-manager** | Requirements/PRD work — turning supplied briefs or transcripts into PRDs, epics, user stories — BEFORE architect (works from supplied context; the /prd skill is the interactive path) |

---

## Agent Spawning

**Always set `run_in_background: true`** when spawning agents via the Agent tool.

Never block the conversation waiting for an agent to finish. Spawn, then continue responding to the user. You will be notified automatically when the agent completes.

```text
Agent({ ..., run_in_background: true })  ✅
Agent({ ... })                           ❌  blocks the conversation
```

**Exception:** when sequencing dependent agents (e.g. `spelunker → engineer`, `architect → engineer`), wait for the first agent's output before spawning the next — `run_in_background: true` applies to independent parallel agents, not sequential chains.

### Permissions

Agents that write files (engineer, pr-shepherd) require `Edit` and `Write` permissions. These must be pre-approved — without them the agent will stall waiting for user approval on every file change.

Ensure `~/.claude/settings.json` includes:
```json
{
  "permissions": {
    "allow": ["Edit", "Write"]
  }
}
```

Without these permissions in the global settings, engineer and pr-shepherd agents cannot operate autonomously in the background.

Ensure `kg` is on your PATH. The `mcp/` directory is a git submodule and is empty
on fresh clones — initialize it first:

```bash
git submodule update --init mcp && python3 mcp/install.py --mcp kg
```

The install must leave `kg` on PATH — the plugin's `.mcp.json` launches
`kg server --stdio`.

---

## Sequencing Rules

Always sequence agents so each one's output feeds the next:

**When approach is uncertain:**
```
spelunker → engineer
```

**When architecture is needed first:**
```
architect → engineer
```

**When code must be reviewed before merge:**
```
engineer → reviewer
```

**When root cause is known:**
```
engineer (directly)
```

**Full feature delivery:**
```
architect (design) → engineer (implement) → reviewer (review)
```

**Unknown production issue:**
```
spelunker (trace) → inspector (root cause) → engineer (fix)
```

**Never parallelize** agents whose outputs depend on each other. Architect output feeds engineer input — they cannot run in parallel.

---

## Task Tracking

Use the native **TaskCreate/TaskUpdate** tools to track multi-step work within the
current session — one task per delegated unit of work, updated as agents complete.

The agent's brief lives in its Agent-tool prompt — always fully self-contained.
Task packets (`.ai/tasks/<slug>/` with `task.md` + `result.md`) are **optional**:
reach for one only on multi-session work where a brief must outlive any single
session. If you write one, fill it out completely — never leave template
placeholders.

---

## How to Write a Dense Brief

**The difference between a good brief and a bad brief is cost.**

A vague brief → the agent spends all its context reading files to understand the codebase. A dense brief → the agent writes code immediately.

**Evidence:** Task `ai-pack-8x0` received a 4-bullet description of what to fix. The engineer spent all 600 turns reading files — zero code written, ~$40 burned. The retry task `ai-pack-ms7` had complete code inline. The engineer wrote all changes in 12 turns.

### Required elements in every engineer brief

#### 1. Exact absolute file paths

❌ Bad: "Fix the streaming adapter"
✅ Good: "TARGET FILE: `/path/to/project/internal/streaming/openai_adapter.go`"

#### 2. Specific code or exact changes

❌ Bad: "Fix the tool call handling to emit proper events"
✅ Good:
```go
// Replace Next() with:
func (a *OpenAIChatStreamAdapter) Next() bool {
    if a.done { return false }
    ...
}
```

#### 3. The phrase "All context provided"

This exact phrase triggers Fast Path in the engineer agent — the engineer skips all
planning phases and goes directly to writing code.

Without this phrase, the engineer runs full pre-flight checks (reads ADRs, task packets,
architecture docs) before touching any code.

#### 4. Acceptance criteria as shell commands

❌ Bad: "Verify it works"
✅ Good:
```bash
go build ./...               # must exit 0
go test ./internal/...       # all pass
```

#### 5. One sentence of context (the WHY)

Engineers make judgment calls. Without context, they make wrong ones.

"This fixes the streaming adapter that was emitting empty events for tool calls — the root cause
was that `Next()` was not consuming the accumulator."

---

## What NOT to Do

**❌ Don't give vague briefs**

"Fix the auth middleware" without file paths = agent reads the entire codebase before writing a single line. If you cannot write the exact file path, use spelunker first to find it.

**❌ Don't delegate research you haven't done**

If you don't know what the code currently looks like, spawn a spelunker first. Incorporate the spelunker's findings into the engineer brief. Do not ask the engineer to "figure out the current implementation."

**❌ Don't parallelize dependent agents**

If engineer output is architect input (or vice versa), they must be sequential. Parallel execution only works for truly independent tasks (e.g., reviewing two unrelated PRs).

**❌ Don't skip the spelunker step for complex bugs**

If the root cause is unclear, spawn spelunker/inspector first. A fix built on a wrong root cause assumption creates more bugs.

---

## Brief Template

Use this template for engineer briefs:

```markdown
**All context provided**

## Context
[One sentence: what is the problem and why this fix?]

## Task
[One sentence: what must the engineer produce?]

## Files to change
- `ABSOLUTE_PATH/to/file1.go` — what changes (specific description)
- `ABSOLUTE_PATH/to/file2_test.go` — what tests to add

## Exact changes
[Paste code, exact line numbers, or specific instructions per file]

## Acceptance criteria
- [ ] `<shell command>` — must exit 0
- [ ] `<shell command>` — must produce output X
```
