# AI-Pack Hook Scripts

This directory contains utility scripts used by ai-pack commands.

## Scripts

### `task-init.py`
Creates new task packets with templates.
- **Used by:** `/ai-pack task-init` command
- **Purpose:** Initialize task packet directory structure
- **Usage:** `python3 .claude/hooks/task-init.py <task-name>`

### `task-status.py`
Displays current task packet status and progress.
- **Used by:** `/ai-pack task-status` command  
- **Purpose:** Show progress through task lifecycle
- **Usage:** `python3 .claude/hooks/task-status.py`

## Setup

These scripts are installed when you copy the `.claude/` template directory:

```bash
# Copy from ai-pack template
cp -r .ai-pack/templates/.claude .claude/

# Make scripts executable
chmod +x .claude/hooks/*.py
```

## Note on Hooks

**AI-pack no longer uses Claude Code hooks for enforcement.** The agent server framework handles all monitoring and workflow orchestration directly.

The scripts in this directory are **utility scripts for commands**, not enforcement hooks.

## References

- AI-Pack Documentation: `.ai-pack/README.md`
- AI-Pack Gates: `.ai-pack/gates/`
