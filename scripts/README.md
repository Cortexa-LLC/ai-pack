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
