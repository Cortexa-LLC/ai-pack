---
sidebar_position: 2
---

# Getting Started

Install the AI-Pack plugin into Claude Code, set up the knowledge-graph server, and run your first multi-agent workflow.

## Prerequisites

- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) installed and working
- Git (to clone the repo and initialize the `mcp` submodule)
- Python 3 (used by the `kg` installer)

## Installation

### 1. Clone the repo

```bash
git clone https://github.com/Cortexa-LLC/ai-pack.git
cd ai-pack
```

### 2. Install the plugin

```bash
make install-plugin
```

This registers the local marketplace with Claude Code and installs `ai-pack@ai-pack`. It is a one-time step; the plugin then works in any project you open with Claude Code.

### 3. Set up the `kg` knowledge-graph server

The plugin's `.mcp.json` launches a `kg` MCP server that gives agents persistent memory. `kg` is a standalone binary provided by the `mcp` git submodule:

```bash
git submodule update --init mcp
python3 mcp/install.py --mcp kg
```

See [Knowledge Graph](./knowledge-graph.md) for what it does and how agents use it.

### 4. Restart Claude Code

Close and reopen Claude Code so the plugin and MCP server load.

## Verify the installation

**Agents.** In a Claude Code session, the six subagents should be available to the `Agent` tool as `ai-pack:*` subagent types: `ai-pack:architect`, `ai-pack:engineer`, `ai-pack:inspector`, `ai-pack:pr-shepherd`, `ai-pack:reviewer`, `ai-pack:spelunker`. Ask Claude to "list the available subagent types" if you want to confirm.

**Knowledge graph.** From a terminal:

```bash
claude mcp list
```

The output should include the `kg` server. You can also run `scripts/verify-kg.sh` from the ai-pack repo.

**Skills.** The three skills appear as `/ai-pack:orchestrate`, `/ai-pack:pre-push`, and `/ai-pack:shepherd-pr`, and also trigger automatically on matching requests.

## First use

Open Claude Code in any project and describe work in plain language — the skills and agents pick it up:

**Drive a PR to merge-ready:**

```text
shepherd PR #42 to merge-ready
```

This triggers the shepherd-pr skill, which spawns the pr-shepherd agent to watch CI, fix failures, and address reviewer threads until the PR is green and approved.

**Multi-step engineering work:**

```text
investigate why the login flow intermittently fails, then fix it
```

This triggers the orchestrate skill, which decomposes the request — an inspector agent finds the root cause, then an engineer agent implements the fix with tests.

**Check your commits before pushing:**

```text
review my commits before I push
```

This triggers the pre-push skill: a reviewer agent examines your local diff, and if issues are found, an engineer agent fixes them and amends the commit, looping until the review passes.

## Updating

After pulling changes to the ai-pack repo:

```bash
make update-plugin
```

## Next steps

- **[Agents](./agents.md)** — what each subagent does and when to use it
- **[Skills](./skills.md)** — the three workflow skills in detail
- **[Knowledge Graph](./knowledge-graph.md)** — persistent project memory
- **[Task Packets](./task-packets.md)** — optional convention for multi-session briefs

## Support

- [GitHub Issues](https://github.com/Cortexa-LLC/ai-pack/issues)
- [GitHub Discussions](https://github.com/Cortexa-LLC/ai-pack/discussions)
- [Claude Code Documentation](https://docs.anthropic.com/en/docs/claude-code)
