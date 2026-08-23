---
sidebar_position: 3
---

# Agents

AI-Pack ships six specialized subagents in `plugin/agents/`. Each is a self-contained role definition — its own tool discipline, quality gates, and reporting format — spawned via Claude Code's native `Agent` tool as `ai-pack:<name>`.

## How agents are spawned

Subagents share **no memory** with your main session or with each other. Every prompt you pass to the `Agent` tool must be fully self-contained: what to do, which files are involved, acceptance criteria, and any context the agent needs.

- **One agent:** use the `Agent` tool with `subagent_type: "ai-pack:engineer"` (or another role) and a complete brief.
- **Parallel agents:** send multiple `Agent` tool calls in a single message — independent tasks run concurrently.
- **Coordination:** for multi-step work, the [orchestrate skill](./skills.md#orchestrate) handles decomposition and delegation for you.

Agents can persist findings to the [knowledge graph](./knowledge-graph.md), so what one agent learns can be recalled by later agents and later sessions.

## architect

**What it does:** Technical design — system architecture, module boundaries, API contracts, data models, and Architecture Decision Records (ADRs). The architect defines *how*; engineers implement the detailed solution.

**When to use:** A task requires design decisions before implementation — new integrations, module boundaries, data models, or feasibility evaluation.

**Examples:**
- "design the architecture for the new notification system"
- "evaluate whether to use WebSockets or SSE for real-time updates"
- "create an ADR for the database migration strategy"

## engineer

**What it does:** Implementation — writes code, fixes bugs, creates tests. Applies test-driven development where it fits, runs the project's build and lint gates, and reports exactly which files changed with a suggested commit message.

**When to use:** A task requires making changes to files in a codebase.

**Examples:**
- "implement the authentication feature"
- "fix the bug in the streaming adapter"
- "refactor the config loader to support environment overrides"

## inspector

**What it does:** Root-cause analysis for complex bugs. Produces a retrospective and a precise fix specification that an engineer can execute. The inspector *investigates*; the engineer *fixes*.

**When to use:** The bug is complex, the root cause is unclear, multiple modules are involved, or the failure is intermittent. Do **not** use it for simple obvious bugs — hand those directly to the engineer.

**Examples:**
- "investigate why the payment processor occasionally returns 500"
- "root cause analysis for the memory leak in the worker pool"
- "investigate this race condition and document the fix spec"

## pr-shepherd

**What it does:** Iterative PR driver. Watches CI, fixes failures, addresses reviewer threads, and loops until the PR is merge-ready — all checks green, all threads resolved, review verdict APPROVED. Usually invoked through the [shepherd-pr skill](./skills.md#shepherd-pr).

**When to use:** A PR has failing CI checks or open reviewer threads and you want it driven to an approved state without manual intervention.

**Examples:**
- "shepherd PR #42 to merge-ready"
- "drive PR #15 until it's green and all threads are resolved"
- "fix the CI failures on my PR and address the reviewer comments"

## reviewer

**What it does:** Code review focused on quality, security, and best practices. Produces structured findings with severity levels, ending in an APPROVED or CHANGES REQUIRED verdict.

**When to use:** You want a second-opinion review of code changes, a PR review, or a security audit of a module. The [pre-push skill](./skills.md#pre-push) uses it automatically against local commits.

**Examples:**
- "review the authentication handler for security issues"
- "review this PR before I merge it"
- "review these files before I push"

## spelunker

**What it does:** Codebase investigation — explores unfamiliar code, traces execution paths, maps dependencies, and answers "how does X work" questions. Reports findings with file:line references.

**When to use:** Before engineering when the implementation path is unclear, or when debugging requires understanding how a system behaves before writing a fix.

**Examples:**
- "how does the authentication flow work end to end"
- "trace the execution path for this failing test"
- "map the dependencies of the notification module"

## Choosing an agent

| Situation | Agent |
|-----------|-------|
| Design decision needed before code | architect |
| Files need to change | engineer |
| Complex bug, unclear root cause | inspector |
| PR blocked on CI or review threads | pr-shepherd |
| Second opinion on code quality/security | reviewer |
| "How does X work?" / unfamiliar code | spelunker |
