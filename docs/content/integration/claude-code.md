---
sidebar_position: 1
title: Claude Code Integration
---

# Claude Code Integration

AI-Pack includes native Claude Code integration with commands, skills, rules, and hooks.

## Quick Setup

```bash
# After adding ai-pack as submodule
python3 .ai-pack/templates/.claude-setup.py
```text

This creates:
- `.claude/commands/ai-pack/` - Slash commands
- `.claude/skills/` - Auto-triggered roles
- `.claude/rules/` - Modular rules
- `.claude/hooks/` - Enforcement scripts
- `.claude/settings.json` - Configuration

## Available Commands

- `/ai-pack task-init <name>` - Create task packets
- `/ai-pack orchestrate` - Assume Orchestrator role
- `/ai-pack engineer` - Assume Engineer role
- `/ai-pack test` - Validate tests (MANDATORY)
- `/ai-pack review` - Code review (MANDATORY)
- `/ai-pack help` - Show all commands

## Configuration

See [Claude Code Configuration](../claude-code-configuration) for detailed setup.

## Enforcement Layers

1. **Passive Documentation** - `CLAUDE.md` in project root
2. **Active Rules** - `.claude/rules/*.md` auto-loaded
3. **Auto-Triggered Skills** - Activate on keywords
4. **Manual Commands** - Explicit invocation
5. **Hook Enforcement** - Blocks gate violations
