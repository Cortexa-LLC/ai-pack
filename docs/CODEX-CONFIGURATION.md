# Codex Configuration Guide

This guide explains how to wire ai-pack into Codex using `AGENTS.md`.

## Overview

Codex reads a project-level `AGENTS.md` file as its entry point. ai-pack provides
an `AGENTS.md` template that links to shared gates, roles, workflows, and task
packet conventions.

## Setup (Recommended)

After adding ai-pack as a submodule:

```bash
python3 .ai-pack/templates/.codex-setup.py
```

This creates:
- `AGENTS.md` in your project root
- `.ai/tasks/` for task packets
- `.ai/repo-overrides.md` for project-specific rules
- `.codex/` for optional Codex-specific guidance
- `.codex/commands/` for CLI command equivalents

## Manual Setup

If you prefer manual setup:

```bash
cp .ai-pack/templates/AGENTS.md .
mkdir -p .ai/tasks
cp .ai-pack/templates/repo-overrides.md .ai/repo-overrides.md
cp -r .ai-pack/templates/.codex .codex
```

## Customization Checklist

1. Edit `AGENTS.md` with your project name and repo URL.
2. Update `.ai/repo-overrides.md` with project-specific rules.
3. Add any Codex-specific rules in `.codex/rules/`.
4. Keep `AGENTS.md` concise and link to `.ai-pack/` for shared standards.

## CLI Command Equivalents

Codex does not support slash commands. Use these scripts instead:

```bash
python3 .codex/commands/health.py
python3 .codex/commands/agents.py
python3 .codex/commands/orchestrate.py
python3 .codex/commands/task-queue.py
```

## Updating Existing Projects

```bash
python3 .ai-pack/templates/.codex-update.py
```

If you customized `AGENTS.md` or `.codex/` files, the update script will save
new templates alongside your files with a `.new` suffix for manual review.
