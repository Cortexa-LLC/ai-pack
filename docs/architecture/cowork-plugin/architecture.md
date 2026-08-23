# Architecture: AI-Pack Cowork Plugin

**Date:** 2026-05-28
**Status:** Historical — implemented and since superseded. `cowork-plugin/` was renamed to
`plugin/` (2026-06-09), `roles/claude-code/` was removed (2026-08-22), and `plugin/agents/`
is now the single canonical source of role definitions. Directory paths below have been
updated to the current `plugin/` name; other details reflect the design at time of writing.

---

## Problem

AI-Pack's multi-agent system currently runs exclusively via the API (agent server on port 8082,
spawning Claude Code subprocesses). Every agent turn costs API credits. The user has a Claude Max
subscription which provides unlimited usage through Claude Desktop / Cowork, but there was no
integration path between ai-pack's agent roles and Cowork's native sub-agent infrastructure.

---

## Solution

Package ai-pack's agent role definitions as a **Claude Cowork plugin**. Cowork's native
sub-agent coordination replaces the API-backed agent server for knowledge work tasks,
using the Max subscription instead of API credits.

The two execution paths remain independent and coexist:

| Dimension | API Path (existing) | Cowork Path (new) |
|---|---|---|
| Entry point | `agent engineer <id> --stream` | Cowork desktop app |
| Orchestration | ai-pack agent server (port 8082) | Cowork native |
| Agent execution | Claude Code subprocesses | Cowork sub-agents |
| Role definitions | `roles/*.md` | `plugin/agents/*.md` |
| Cost | API key per turn | Max subscription |
| MCP role | Drives the agent server | Task DB (read/write); GitHub; KG |

The agent-mcp server is wired into the Cowork plugin as a **read/write interface to the
shared task database only** — not as an execution driver. Cowork agents can check for
existing tasks, create task records for work started in Cowork (so `agent list` reflects
both paths), and read status/logs of API-path tasks. Cowork handles sub-agent spawning
natively via the plugin's `agents/` directory. The `mcp__agent-mcp__spawn_agent` tool
must never be called from Cowork — that would bypass the Max subscription and incur API costs.

---

## Prior Work

Two portable role definitions already exist and are ready to drop into the plugin:

- `roles/claude-code/engineer.md` — 180-line distillation of `roles/engineer.md`
- `roles/claude-code/architect.md` — 150-line distillation of `roles/architect.md`

These were written specifically to be server-runtime-free: no `TaskComplete` tool, no Beads
commands, no KG tool calls, no server frontmatter. They need only Cowork agent frontmatter
added (a `description` field with `<example>` trigger blocks).

---

## Plugin Structure

Based on the `anthropics/knowledge-work-plugins` GitHub repository (the canonical reference
for Cowork plugin authoring), the plugin follows this layout:

```
plugin/
├── .claude-plugin/
│   └── plugin.json              # Required manifest
├── agents/
│   ├── engineer.md              # Implementation specialist
│   ├── architect.md             # Technical design specialist
│   ├── reviewer.md              # Code review specialist
│   └── spelunker.md             # Codebase investigation specialist
├── skills/
│   └── orchestrate/
│       └── SKILL.md             # How to decompose work and delegate to agents
└── .mcp.json                    # External tool connections (GitHub only for now)
```

### plugin.json

```json
{
  "name": "ai-pack",
  "version": "0.1.0",
  "description": "Multi-agent engineering system — delegate implementation, architecture, review, and investigation to specialized sub-agents",
  "author": {
    "name": "Bryan Woodruff"
  }
}
```

### Agent file format

Each agent file in `agents/` uses YAML frontmatter + markdown system prompt body:

```markdown
---
name: engineer
description: >
  Implementation specialist for software engineering tasks. Writes code, fixes bugs,
  creates tests. Use when a task requires making changes to files in a codebase.
  <example>implement the authentication feature</example>
  <example>fix the bug in the streaming adapter</example>
  <example>add unit tests for the payment module</example>
  <example>refactor the config loader to support environment overrides</example>
---

[system prompt — content from roles/claude-code/engineer.md]
```

The `description` field is critical: Cowork uses it to decide when to invoke the agent.
Include 3–5 `<example>` blocks with representative natural-language triggers.

### .mcp.json

GitHub for PR/issue access. `kg` for persistent memory and investigation findings.
`agent-mcp` for shared task DB access — check for duplicate work, log Cowork-initiated
tasks, read status/logs of API-path runs. Do NOT use `spawn_agent` from Cowork.

```json
{
  "mcpServers": {
    "github": {
      "type": "http",
      "url": "https://api.githubcopilot.com/mcp/"
    },
    "kg": {
      "command": "kg",
      "args": ["server", "--stdio"]
    },
    "agent-mcp": {
      "command": "/usr/local/bin/agent-mcp",
      "args": [],
      "type": "stdio"
    }
  }
}
```

### Skill format (orchestrate)

The `orchestrate` skill teaches Cowork how to decompose engineering work and delegate to
the right sub-agents. It fires automatically when the user describes a multi-step
engineering task.

SKILL.md frontmatter:

```yaml
---
name: orchestrate
description: >
  Decompose and delegate engineering work to specialized sub-agents. Use when the user
  describes a software task that involves implementation, design, review, or investigation.
  Triggers on: "build this feature", "fix this bug", "design the architecture for",
  "review this code", "investigate why", "implement and test".
---
```

---

## Agents to Build

### 1. engineer (HIGH PRIORITY)

**Source:** `roles/claude-code/engineer.md` (already written)
**Changes needed:** Add YAML frontmatter with `name`, `description`, and `<example>` blocks
**Role:** Writes code, fixes bugs, creates tests. The workhorse agent.

### 2. architect (HIGH PRIORITY)

**Source:** `roles/claude-code/architect.md` (already written)
**Changes needed:** Add YAML frontmatter with `name`, `description`, and `<example>` blocks
**Role:** Produces design documents, ADRs, API specs, feasibility assessments.

### 3. reviewer (MEDIUM PRIORITY)

**Source:** `roles/reviewer.md` — needs portable distillation first (same process used for
engineer and architect: strip server machinery, keep behavioral core)
**Role:** Reviews code for correctness, security, style, test coverage.

### 4. spelunker (MEDIUM PRIORITY)

**Source:** `roles/spelunker.md` — needs portable distillation first
**Role:** Investigates unfamiliar codebases, traces execution paths, maps dependencies.

---

## Skills to Build

### orchestrate (HIGH PRIORITY)

Teaches Cowork how to:
1. Assess whether a task needs one agent or multiple
2. Determine which agent(s) to use (engineer for impl, architect for design, etc.)
3. Write dense, context-rich briefs so agents don't burn turns exploring
4. Sequence or parallelize agents appropriately

Key content to include from `roles/shared/orchestrator-engineer-handoff.md`:
- The "bad brief vs good brief" principle (ai-pack-8x0 vs ai-pack-ms7)
- Providing exact file paths + specific code in briefs
- The "All context provided" fast-path signal
- When to spelunk before engineering

---

## Implementation Steps

### Step 1 — Scaffold the plugin directory

```bash
PROJECT_ROOT=$(git rev-parse --show-toplevel)
mkdir -p "$PROJECT_ROOT/plugin/.claude-plugin"
mkdir -p "$PROJECT_ROOT/plugin/agents"
mkdir -p "$PROJECT_ROOT/plugin/skills/orchestrate"
```

### Step 2 — Write plugin.json

Write `.claude-plugin/plugin.json` per the manifest format above.

### Step 3 — Write agent files

For `engineer` and `architect`:
- Copy content from `roles/claude-code/engineer.md` and `roles/claude-code/architect.md`
- Prepend YAML frontmatter block with `name`, `description`, and `<example>` blocks
- Write to `agents/engineer.md` and `agents/architect.md`

For `reviewer` and `spelunker`:
- Read `roles/reviewer.md` and `roles/spelunker.md`
- Distill to portable format (strip: TaskComplete, Beads commands, KG tools, server frontmatter)
- Keep: behavioral core, quality standards, completion format, escalation rules
- Add YAML frontmatter
- Write to `agents/reviewer.md` and `agents/spelunker.md`

### Step 4 — Write .mcp.json

Write the GitHub-only MCP config per the format above.

### Step 5 — Write the orchestrate skill

Write `skills/orchestrate/SKILL.md` covering:
- When to use each agent
- How to write a dense brief (exact paths, specific code, acceptance criteria as commands)
- Parallelism guidance (which agents can run concurrently vs must sequence)
- The fast-path signal ("All context provided")

### Step 6 — Validate

```bash
claude plugin validate plugin/.claude-plugin/plugin.json
```

If the CLI is unavailable, manually verify:
- `.claude-plugin/plugin.json` exists and has valid JSON with `name` field in kebab-case
- Each `agents/*.md` file has valid YAML frontmatter with `name` and `description`
- Each `skills/*/SKILL.md` file has valid YAML frontmatter with `name` and `description`
- `.mcp.json` is valid JSON

### Step 7 — Package and install

```bash
cd /Users/bryanw/Projects/Vibe/ai-pack
zip -r /tmp/ai-pack.plugin plugin/ -x "*.DS_Store"
cp /tmp/ai-pack.plugin /Users/bryanw/Claude/ai-pack.plugin
```

Then install via CLI:
```bash
claude plugin install /Users/bryanw/Claude/ai-pack.plugin
```

Or drop the `.plugin` file into Cowork's chat to install via the UI.

---

## Acceptance Criteria

- [ ] `plugin/` directory exists at project root with correct structure
- [ ] `plugin validate` passes (or manual check confirms all files present and valid)
- [ ] Engineer agent fires when asked to implement/fix/test something in Cowork
- [ ] Architect agent fires when asked to design or assess feasibility
- [ ] Orchestrate skill surfaces when describing multi-step engineering work
- [ ] Agents produce output in the completion format defined in their role files
- [ ] No reference to TaskComplete, Beads commands, or KG tools in any agent file
- [ ] `.mcp.json` does not reference the ai-pack agent server

---

## What This Does NOT Change

- The existing agent server (port 8082) and all API-path tooling remain untouched
- `roles/*.md` (server-format roles) are not modified
- `roles/claude-code/*.md` (portable roles) are the source; `plugin/agents/` are copies with frontmatter added
- The `agent` CLI, task packets, and `.ai/tasks/` workflow are unchanged

---

## Key Reference Files

| File | Purpose |
|---|---|
| `roles/claude-code/engineer.md` | Portable engineer role — source for agent file |
| `roles/claude-code/architect.md` | Portable architect role — source for agent file |
| `roles/reviewer.md` | Reviewer role — needs distillation |
| `roles/spelunker.md` | Spelunker role — needs distillation |
| `roles/shared/orchestrator-engineer-handoff.md` | Brief-writing principles for orchestrate skill |
| `roles/shared/agent-policy.md` | Cross-cutting agent rules |
| `anthropics/knowledge-work-plugins` (GitHub) | Canonical plugin format reference |
