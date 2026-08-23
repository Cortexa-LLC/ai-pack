---
sidebar_position: 5
---

# Knowledge Graph

The `kg` knowledge graph gives agents **persistent, per-project memory**. Subagents share no context with each other or with your main session — without kg, everything an agent learns dies with its session. With kg, findings are checkpointed as entities and observations that any later agent (or later session) can query.

## What it is

`kg` is a standalone MCP server that maintains a knowledge graph for each project:

- **Entities** — the things worth remembering: components, modules, bugs, decisions, investigations.
- **Observations** — facts attached to entities: "the retry logic in `client.go` swallows context cancellation", "PR #42 introduced the two-counter system".
- **Links** — relationships between entities, so agents can traverse from a component to its known issues and related decisions.

Data is stored per-project in **`.ai/knowledge.db`**. This file is the project's accumulated memory — **never delete it**. It is typically git-ignored (local memory), so deleting it loses everything agents have learned about the project.

## How the plugin wires it

The plugin's `plugin/.mcp.json` declares the server:

```json
{
  "mcpServers": {
    "kg": {
      "type": "stdio",
      "command": "kg",
      "args": ["server", "--stdio"]
    }
  }
}
```

Claude Code launches the `kg` binary over stdio when a session starts, and its tools become available to your session and to every spawned subagent (as `kg__*` / `mcp__kg__*` tools).

## How agents use it

The agent role definitions build kg into their workflow:

- **Preflight context.** Before starting work, agents can call `kg__get_preflight_context` or `kg__search_knowledge` for the component they are about to touch — surfacing prior decisions, known issues, and architectural context in one call instead of re-reading docs.
- **Checkpointing findings.** Investigative agents (inspector, spelunker) persist root causes, traced flows, and dead ends as they go via `kg__add_entity` / `kg__add_observation`. If a session is interrupted, the next agent resumes from the checkpoint instead of starting over.
- **File context.** `kg__get_file_context` returns what the graph knows about a specific file — which entities reference it and what has been observed about it.

The practical effect: the second time any agent touches a subsystem, it starts with everything the first agent learned.

## Install

`kg` is provided by the `mcp` git submodule of the ai-pack repo:

```bash
git submodule update --init mcp
python3 mcp/install.py --mcp kg
```

This installs the `kg` binary onto your PATH, where the plugin's `.mcp.json` expects to find it.

## Verify

```bash
# The kg server should appear in Claude Code's MCP list
claude mcp list

# Or run the bundled verification script from the ai-pack repo
scripts/verify-kg.sh
```

Inside a session, asking Claude to "search the knowledge graph for `<component>`" should invoke `kg__search_knowledge` without errors.

## Where data lives

| Path | What it is |
|------|-----------|
| `.ai/knowledge.db` | The project's knowledge graph. Accumulated memory — never delete. |

If `kg` is missing from `claude mcp list`, re-run the install steps above and restart Claude Code. The plugin degrades gracefully without it — agents still work, they just lose cross-session memory.
