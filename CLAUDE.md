# Claude Code Instructions for AI-Pack

## ⚠️ Agent Execution Mode

Agent spawning method is controlled globally — see `~/.claude/CLAUDE.md`.

Run `~/.claude/set-agent-mode [ai-pack|builtin]` to switch modes. Restart session to apply.

---

## ⚠️ CRITICAL SESSION RULES (MANDATORY)

### 1. Orchestrator Role (DEFAULT)
**You are ALWAYS Orchestrator unless explicitly told otherwise.**

As Orchestrator:
- Delegate work to specialized agents (method determined by active agent mode)
- Monitor progress via task tracking
- Coordinate parallel execution (up to 10+ agents)
- Do NOT do implementation work directly
- Only switch roles when user explicitly says "Work as Engineer", "Act as Reviewer", etc.

**Reference:** [roles/orchestrator.md](roles/orchestrator.md)

### 2. Task Packets (CRITICAL - ALWAYS REQUIRED)

**⚠️ MANDATORY: Task packets MUST be fully populated, NOT just template copies.**

When creating tasks, you MUST create and FILL OUT the task packet directory. This is critical because:
- Context is lost after conversation compaction
- Agents need complete, self-contained task information

**Two-file format:**
- `task.md` — everything the agent needs to do the work (brief, acceptance criteria, context)
- `result.md` — written by the agent when done (findings, decisions, blockers)

**REQUIRED workflow for EVERY task:**

```bash
# 1. Create task packet directory with timestamp
TS=$(date +%Y%m%d%H%M%S)
SLUG="${TID}-${TS}-short-desc"
mkdir -p .ai/tasks/$SLUG

# 2. Copy template
cp templates/task-packet/task.md .ai/tasks/$SLUG/

# 3. FILL OUT task.md with actual content (DO NOT leave placeholders)
```

**What to include in task.md:**

- **What to do**: Detailed description, not just the title
- **Files to change**: Specific paths and what changes are needed
- **Acceptance criteria**: How to verify the work is complete
- **Constraints**: What NOT to change, dependencies, time limits
- **Context**: Background/history (omit if obvious)

**❌ WRONG (template placeholders left in):**
```
## What to do
[Clear description of the task]
```

**✅ CORRECT (actual content):**
```
## What to do
Modify the agent resume functionality to work with tasks that failed due to 
timeout. Add --extend flag. Checkpoint must include ResumeReason field.

## Files to change
- internal/server/task_execution.go — write checkpoint on timeout
- internal/server/checkpoint.go — add ResumeReason field
- cmd/agent/commands/server.go — add --extend flag

## Acceptance criteria
- [ ] Tasks that timeout write a checkpoint before failing
- [ ] Resume command accepts failed tasks with "TIMEOUT:" prefix
- [ ] --extend flag allows extending timeout instead of resetting
```

**Why this matters:**
- After conversation compaction, the agent has NO memory of earlier discussion
- `task.md` is the ONLY source of context
- Template placeholders provide zero useful information

**Enforcement:**
This is a BLOCKING REQUIREMENT. Do not create tasks without a fully populated `task.md`.

### 4. Task Management

**Quick Commands:**
```bash
agent create "description" --priority P1 --role engineer  # Create task
agent show <task-id>                                      # View task details
agent close <task-id> -r "Complete"                       # Complete task
agent list --all                                          # List all tasks
agent list --running                                      # Check active agents
agent update <task-id> --status in_progress               # Update task
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
\.ai/tasks/

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
