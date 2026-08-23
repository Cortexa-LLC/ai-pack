---
sidebar_position: 1
slug: /
---

# AI-Pack Documentation

**AI-Pack is a Claude Code plugin that turns a single Claude Code session into a coordinated engineering team.**

## What is AI-Pack?

AI-Pack ships three things:

- **Six specialized subagents** — architect, engineer, inspector, pr-shepherd, reviewer, and spelunker. Each is a self-contained role definition with its own tool discipline, quality gates, and reporting format.
- **Three workflow skills** — orchestrate (decompose and delegate multi-step work), pre-push (review-and-fix loop on local commits), and shepherd-pr (drive a GitHub PR to a green, approved state).
- **A knowledge-graph MCP server (`kg`)** — persistent, per-project memory that lets agents accumulate and recall findings across sessions.

## Who is it for?

Anyone using [Claude Code](https://docs.anthropic.com/en/docs/claude-code) on real software projects who wants more than a single generalist assistant: parallel specialists for implementation and review, a repeatable path from "fix this bug" to a merged PR, and project knowledge that survives session boundaries.

## How it works

Claude Code provides the execution loop — the tools, the permission system, and the native `Agent` tool for spawning subagents. AI-Pack provides what runs on top of it:

- **Roles.** Each subagent definition constrains a spawned agent to one job (design, implement, investigate, review, shepherd) with clear quality gates and a structured completion report.
- **Coordination.** The skills package proven multi-agent patterns: the orchestrate skill decomposes work and delegates it to the right roles, with parallel spawns where tasks are independent; pre-push and shepherd-pr run bounded fix-and-verify loops.
- **Memory.** Subagents share no context with each other or with your main session — every brief must be self-contained. The `kg` knowledge graph fills the gap: agents checkpoint findings into it and pull prior context out of it, so knowledge persists across agents and across sessions.

## Documentation structure

- **[Getting Started](./getting-started.md)** — install the plugin and verify it works
- **[Agents](./agents.md)** — the seven subagents and when to use each
- **[Skills](./skills.md)** — the four workflow skills and what triggers them
- **[Knowledge Graph](./knowledge-graph.md)** — persistent project memory via the `kg` MCP server
- **[Task Packets](./task-packets.md)** — an optional convention for briefs that outlive a session
- **Workflows** — general engineering process guides: [bugfix](./workflows/bugfix.md), [feature](./workflows/feature.md), [refactor](./workflows/refactor.md), [research](./workflows/research.md), [standard](./workflows/standard.md)
- **[Clean Code](./quality/clean-code/00-general-rules.md)** — coding standards and best practices agents are held to

## Support

- [GitHub Issues](https://github.com/Cortexa-LLC/ai-pack/issues)
- [GitHub Discussions](https://github.com/Cortexa-LLC/ai-pack/discussions)

## History

AI-Pack 1.x/2.0 was an API-driven agent server: a Go server running coding agents against the Claude API, with an `agent` CLI, an `agent-mcp` MCP server, and a React GUI. That architecture was deprecated on 2026-08-22 in favor of the plugin model — Claude Code natively provides the execution loop the server used to implement. The server-era code is preserved at tag [`v2.0-server-final`](https://github.com/Cortexa-LLC/ai-pack/releases/tag/v2.0-server-final).
