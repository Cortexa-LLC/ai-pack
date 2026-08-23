# Claude Code Instructions for AI-Pack

## What This Repo Is

This repo IS the ai-pack Claude Code plugin. `plugin/agents/` is the **single canonical
source** of the role definitions (architect, engineer, inspector, pr-shepherd, reviewer,
spelunker). There are deliberately no duplicate copies anywhere else in the tree — edit
the files in `plugin/agents/` and `plugin/skills/` in place.

After changing plugin files, refresh the installed plugin with `make update-plugin`.

## Orchestrator Convention

Sessions in this repo delegate implementation work to the plugin's subagents via the
`Agent` tool (`subagent_type: "ai-pack:<role>"`) rather than doing it directly.

- Every subagent prompt must be **fully self-contained** — agents share no memory with
  your session or with each other. Include file paths, acceptance criteria, and context.
- For parallel work, send multiple `Agent` tool calls in a single message.

## Task Packets

For durable briefs that survive conversation compaction, create a task packet under
`.ai/tasks/<slug>/`:

- `task.md` — what to do, files to change, acceptance criteria, constraints, context
- `result.md` — written by the agent when done

Task packets MUST be fully populated — never leave template placeholders.

**WRONG** (placeholder left in):
```
## What to do
[Clear description of the task]
```

**CORRECT** (actual content):
```
## What to do
Add --extend flag to the resume command. Checkpoint must include ResumeReason field.

## Acceptance criteria
- [ ] Resume accepts failed tasks with "TIMEOUT:" prefix
```

Templates: `templates/task-packet/`.

## Runtime Data — Do NOT Delete

Never delete or "clean up" these runtime directories:

- `.claude/` — session runtime data
- `data/` — knowledge-graph and other persistent stores
- `logs/` — execution logs
- `.ai/` — task packets and agent output

## Server Era Ended

ai-pack 1.x/2.0 was an API-driven agent server. It is gone from this tree:

- No `agent` CLI, no agent server, no port 8082, no GUI
- No performance-grade seeding or LiveBench scripts
- History preserved at tag `v2.0-server-final`

If instructions elsewhere reference `agent create`, `agent list`, or server ports,
they are stale — ignore them.
