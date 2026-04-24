# Claude Code Instructions for AI-Pack

## 🚫 ABSOLUTE PROHIBITION — READ FIRST

**NEVER use the built-in `Agent` tool** (the one in your tool list that spawns a sub-agent inline).

That tool bypasses the ai-pack agent server entirely. It is **FORBIDDEN** in this project.

The ONLY way to spawn agents is via the **`agent` bash command**:
```bash
agent engineer <beads-id> --stream   # ✅ CORRECT
```

Using the `Agent(...)` tool call is ALWAYS wrong here, even if it seems convenient.

---

## ⚠️ CRITICAL SESSION RULES (MANDATORY)

### 1. Orchestrator Role (DEFAULT)
**You are ALWAYS Orchestrator unless explicitly told otherwise.**

As Orchestrator:
- Delegate work to specialized agents via `agent` CLI (ONLY method)
- Monitor progress via Beads task tracking
- Coordinate parallel execution (up to 10+ agents)
- Do NOT do implementation work directly
- Only switch roles when user explicitly says "Work as Engineer", "Act as Reviewer", etc.

**Reference:** [roles/orchestrator.md](roles/orchestrator.md)

### 2. Agent CLI (PRIMARY INTERFACE - MANDATORY)

**Quick Start (sequential):**
```bash
# 1. Create Beads task with working directory and task packet
BID=$(bd create "Task description

Working directory: /Users/bryanw/Projects/Vibe/ai-pack
Task packet: .ai/tasks/<beads-id>-<YYYYMMDDHHMMSS>-<short-desc>/

Details..." --priority P1 --json | jq -r '.id')

# 2. Create timestamped task packet dir
TS=$(date +%Y%m%d%H%M%S)
SLUG="${BID}-${TS}-short-desc"
mkdir -p .ai/tasks/$SLUG
cp templates/task-packet/*.md .ai/tasks/$SLUG/

# 3. Spawn agent — blocks until complete, streams live output
agent engineer $BID --stream

# 4. Close task
bd close $BID -r "Complete"
```

**Parallel execution (multiple workstreams):**
```bash
# Spawn all agents in background (no --stream = non-blocking)
agent engineer ai-pack-task1
agent engineer ai-pack-task2
agent engineer ai-pack-task3

# Attach to each one to get live output and block until done
agent wait ai-pack-task1 --stream
agent wait ai-pack-task2 --stream
agent wait ai-pack-task3 --stream
```

**CRITICAL Rules:**
- ✅ Sequential task: use `agent <role> <id> --stream` (blocks until complete)
- ✅ Parallel tasks: spawn without `--stream`, then `agent wait <id> --stream`
- ✅ Use `agent` bash CLI exclusively — the `Agent` tool in your tool list is FORBIDDEN
- ✅ Beads priority format: P0–P4 (NOT high/medium/low)
- ✅ Beads description MUST contain `Working directory:` and `Task packet:` lines
- ❌ NEVER poll manually for completion (use `--stream` or `agent wait`)
- ❌ NEVER use Task tool with run_in_background (broken)
- ❌ NEVER implement code directly as Orchestrator

**⚠️ Beads task ID vs task packet slug — CRITICAL DISTINCTION:**
```
Beads task ID:        HomeControl-qx7               ← pass this to agent commands
Task packet slug:     HomeControl-qx7-20260424-072021-short-desc  ← directory name only
```
`agent logs`, `agent status`, `agent results`, `agent wait`, `agent diff`, `agent files`
all take the **Beads task ID** (e.g. `HomeControl-qx7`), **NOT** the task packet directory
name. The slug includes the timestamp and short-desc suffix — strip everything after the
shortid (3-char alphanumeric after the last hyphen of the project prefix).

**After any task completes, IMMEDIATELY continue — do NOT ask for permission.**

### 3. Beads Task Management

**Quick Commands:**
```bash
bd ready              # Find next available task
bd show <task-id>     # View task details
bd close <task-id>    # Complete task
bd list --running     # Check active agents
```

**Priority format:** P0 (critical) → P4 (low)

---

## Project Structure

This is the AI-Pack multi-agent system with an agent server (a2a-agent) and GUI.

## Critical: Do NOT Delete Runtime Data

**NEVER delete or modify files in `.claude/` directories** - these contain critical runtime data:

### `.claude/performance_grades/`
- Contains performance grading data for intelligent model selection
- Used by the adaptive performance grading system
- Files are named: `{model}_{role}_{project}.json`
- **Regeneration is fast** — run `python3 scripts/seed-grades.py` (fetches live LiveBench scores, no API key needed)

### `.claude/metrics/daily/`
- Contains daily token usage and cost tracking per project
- Used for cost analysis and budget monitoring
- Files are named: `YYYY-MM-DD.json`
- **Regeneration requires CSV exports from Anthropic** - deletion causes permanent data loss

### Why These Are Gitignored
These directories are in `.gitignore` because they contain project-specific runtime data that:
- Changes frequently during server operation
- Should not be committed to version control
- Can be regenerated from historical data (but it's expensive)

## If You Accidentally Deleted Them

If these files were deleted, they can be regenerated:

```bash
# Regenerate performance grades from LiveBench (no API key needed)
python3 scripts/seed-grades.py

# Regenerate daily metrics (requires Anthropic CSV exports in ~/Downloads)
python3 a2a-agent/scripts/backfill-metrics.py
```

## Performance Grades — Setup & Maintenance

Grade files in `.claude/performance_grades/` seed the model selector so it picks
the best cost/quality model for each role. They are populated from
[LiveBench](https://livebench.ai) coding scores (contamination-free, externally maintained).

### Initial setup / after wiping grades
```bash
python3 scripts/seed-grades.py
```
Fetches two LiveBench CSV releases and writes 192 grade files (16 models × 12 roles).
No API keys required.

### When LiveBench publishes a new release
1. Check `https://livebench.ai/table_YYYY_MM_DD.csv` for the new date
2. Add an entry to `LIVEBENCH_SOURCES` in `scripts/seed-grades.py` with the URL and its coding column names
3. Re-run `python3 scripts/seed-grades.py`

### Grade thresholds (LiveBench coding score out of 100)
| Grade | Score | Effect |
|-------|-------|--------|
| A | ≥ 60 | Preferred — selector picks cheapest A first |
| B | ≥ 45 | Acceptable — selected if no A in tier |
| C | ≥ 30 | Avoided — only used as last-resort fallback |
| D | < 30 | Excluded from selection |

### Current grades (Jan 2026 data)
Models without LiveBench data are left ungraded; real task runs will populate them.
The selector falls back to cheapest-in-tier for ungraded models.

## .claudeignore - Context Management

The project uses `.claudeignore` files (similar to `.gitignore`) to prevent agents from reading files that would bloat context:

### How It Works
- Place `.claudeignore` in the project root to define global ignore patterns
- Override patterns in subdirectories with additional `.claudeignore` files
- Patterns use glob syntax: `*`, `**`, `!` for negation
- Applied to Read, Glob, and Grep tools automatically

### Example Patterns
```
# Node dependencies
**/node_modules/
**/package-lock.json

# Build artifacts
**/build/
**/dist/

# Task logs
.beads/tasks/

# Large files
**/*.log
**/*.bin
```

### Why This Matters
Agents have limited context windows. Reading large files like `package-lock.json` (737KB) or `node_modules` can quickly exhaust context, causing agents to run out of memory and fail tasks.

## General Guidelines

- ✅ DO: Read these files to understand system state
- ✅ DO: Let the server create/update them during normal operation
- ✅ DO: Use `.claudeignore` to exclude large/irrelevant files from agent context
- ❌ DON'T: Delete `.claude/` directories
- ❌ DON'T: Mark them as "transient" or "temporary" for cleanup
- ❌ DON'T: Remove them when "cleaning up the project"
