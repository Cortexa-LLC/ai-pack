---
sidebar_position: 6
---

# Task Packets

Task packets are an **optional** convention for writing agent briefs down as files, so they survive beyond a single session.

## The default: briefs in prompts

Normally, no files are needed. Agent briefs are passed directly in the `Agent` tool prompt — fully self-contained: what to do, files to change, acceptance criteria, constraints, context. The [orchestrate skill](./skills.md#orchestrate) composes these briefs for you, and results come back as the agent's report. For most work, that is the whole story.

## When a packet is worth it

Use a task packet when a brief must **outlive any one session** — multi-session epics, work that will be picked up days later, or long-running efforts where conversation history will be compacted away. A file on disk is the only brief a future session is guaranteed to see.

## The two-file convention

Packets live under `.ai/tasks/<slug>/` in the consuming project:

- **`task.md`** — the brief: what to do, files to change, acceptance criteria, constraints, context. Written by whoever is coordinating the work, fully populated — never left as template placeholders.
- **`result.md`** — the output: findings, decisions, blockers. Written by the agent when done.

```bash
mkdir -p .ai/tasks/2026-08-22-auth-feature
cp templates/task-packet/task.md .ai/tasks/2026-08-22-auth-feature/
# fill out task.md, then point the agent's prompt at the packet
```

Templates live in `templates/task-packet/` in the ai-pack repo (alongside ADR, incident, investigation, and security templates in `templates/`).

## Using a packet with an agent

The packet does not replace the self-contained prompt — it feeds it. Point the spawned agent at the packet path in its brief ("read `.ai/tasks/<slug>/task.md` for the full brief; write your findings to `result.md` in the same directory"), and the packet becomes the durable record of both sides of the exchange.
