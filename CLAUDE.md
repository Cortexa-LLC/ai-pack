# Claude Code Instructions for AI-Pack

## Project Structure

This is the AI-Pack multi-agent system with an agent server (a2a-agent) and GUI.

## Critical: Do NOT Delete Runtime Data

**NEVER delete or modify files in `.claude/` directories** - these contain critical runtime data:

### `.claude/performance_grades/`
- Contains performance grading data for intelligent model selection
- Used by the adaptive performance grading system
- Files are named: `{model}_{role}_{project}.json`
- **Regeneration requires historical task data** - deletion causes loss of optimization data

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
# Regenerate performance grades (requires task history)
python3 a2a-agent/scripts/backfill-performance-grades.py

# Regenerate daily metrics (requires Anthropic CSV exports in ~/Downloads)
python3 a2a-agent/scripts/backfill-metrics.py
```

## General Guidelines

- ✅ DO: Read these files to understand system state
- ✅ DO: Let the server create/update them during normal operation
- ❌ DON'T: Delete `.claude/` directories
- ❌ DON'T: Mark them as "transient" or "temporary" for cleanup
- ❌ DON'T: Remove them when "cleaning up the project"
