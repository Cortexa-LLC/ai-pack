# AI-Pack Scripts

Support scripts for the ai-pack plugin repo.

## Available Scripts

### `check-usage.py`

Check API usage and costs for Claude and OpenAI. Compares current costs vs projected
costs with a multi-provider setup.

```bash
python3 scripts/check-usage.py
```

### `check-api-usage.py`

Parse local agent logs to show which models and providers were used. Dates from the
server era; useful for analyzing historical logs.

```bash
python3 scripts/check-api-usage.py
```

### `verify-kg.sh`

Verify the `kg` knowledge-graph indexer builds and responds correctly. Expects a `kg`
binary (see the `mcp` submodule: `git submodule update --init mcp && python3 mcp/install.py --mcp kg`).

```bash
./scripts/verify-kg.sh
```

### `reset-submodule.py`

Safely remove and re-add the `.ai-pack` submodule in a consuming project, cleaning git
cache/config to fix "git directory is found locally" errors.

```bash
# From your project root (the repo that contains .ai-pack as a submodule)
python3 .ai-pack/scripts/reset-submodule.py
```

---

**Note:** Prefer Python (`.py`) over Bash (`.sh`) for new scripts (cross-platform).
Document any new script in this README.

## Migration from 2.x

- **`uninstall-server.sh`** — machine-level removal of the 2.x agent-server: stops launchd/systemd services, removes service definitions, `/usr/local/bin/{agent,agent-mcp,agent-server}`, and the `agent-mcp` MCP registration. Archives `~/.ai-pack/tasks.db` first. `--dry-run` previews; `--purge` also removes `~/.ai-pack` and `~/.claude/performance_grades`. Never touches `~/.claude/metrics`, per-project `.ai/`, or the `kg` binary.
- **`uninstall-project.sh`** — run from a consumer project's root to remove 2.x integration: the `.ai-pack` submodule, `.claude/commands/ai-pack/`, task hooks, rules, template skills/scripts, and their `settings.json` hook registrations (backed up first). Preserves `.ai/` and flags — but never edits — `CLAUDE.md`. `--dry-run` previews.
