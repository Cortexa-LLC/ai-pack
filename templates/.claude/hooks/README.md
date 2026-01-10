## AI-Pack Enforcement Hooks

This directory contains Python scripts that enforce ai-pack gates through Claude Code hooks.

## Hook Scripts

### `task-init.py`
Creates new task packets with templates.
- **Used by:** `/ai-pack task-init` command
- **Purpose:** Initialize task packet directory structure

### `task-status.py`
Displays current task packet status and progress.
- **Used by:** `/ai-pack task-status` command
- **Purpose:** Show progress through task lifecycle

### `check-task-packet.py`
Enforces Task Packet gate before implementation work.
- **Used by:** `UserPromptSubmit` hook
- **Purpose:** Block implementation without task packet
- **Exit codes:**
  - `0` - Allow (gate passed)
  - `1` - Error (technical failure)
  - `2` - Block (gate violation)

## Setup

These hooks are automatically configured when you run the setup script:

```bash
python3 .claude-setup.py
```

This will:
1. Make hook scripts executable
2. Configure Claude Code settings to use hooks
3. Verify setup

## Manual Setup

If you need to set up manually:

```bash
# Make scripts executable
chmod +x .claude/hooks/*.py

# Configure hooks in .claude/settings.json
```

See `settings.json` template for hook configuration.

## Hook Configuration

Hooks are configured in `.claude/settings.json`:

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "python3 .claude/hooks/check-task-packet.py"
          }
        ]
      }
    ]
  }
}
```

## Exit Codes

Hook scripts use standard exit codes:
- **0** - Success, allow operation
- **1** - Technical error
- **2** - Gate violation, block operation

## Testing Hooks

Test hooks manually:

```bash
# Test task-init
python3 .claude/hooks/task-init.py test-task

# Test task-status
python3 .claude/hooks/task-status.py

# Test gate enforcement
echo '{"user_input": "implement login"}' | python3 .claude/hooks/check-task-packet.py
```

## References

- Hooks Documentation: https://code.claude.com/docs/en/hooks.md
- Hooks Guide: https://code.claude.com/docs/en/hooks-guide.md
- AI-Pack Gates: `.ai-pack/gates/`
