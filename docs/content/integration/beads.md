---
sidebar_position: 2
title: Beads Integration
---

# Beads Integration

AI-Pack uses [Beads](https://github.com/steveyegge/beads) for persistent, git-backed task tracking.

## Key Benefits

- ✅ **Cross-session memory** - Tasks persist across conversations
- ✅ **Git-backed storage** - Tasks in `.beads/issues.jsonl`, versioned with code
- ✅ **Dependency tracking** - Full task graphs with automatic "ready" detection
- ✅ **Multi-agent coordination** - Hash-based IDs prevent merge collisions

## Installation

```bash
# macOS/Linux
curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/scripts/install.sh | bash

# Windows PowerShell
irm https://raw.githubusercontent.com/steveyegge/beads/main/install.ps1 | iex
```text

## Core Workflow

```bash
bd ready              # Find next available task
bd show bd-a1b2       # View task details
bd start bd-a1b2      # Begin work
bd close bd-a1b2      # Mark complete
bd create "Task"      # Create new task
bd dep add bd-x bd-y  # Add dependency
```text

## Integration with AI-Pack

- **Orchestrator** uses Beads for task decomposition
- **Engineer** uses Beads to find next work
- **Task persistence** across sessions and machines

## Enforcement

The [Beads Enforcement Gate](../gates/06-beads-enforcement) makes Beads usage **MANDATORY, BLOCKING** for all task operations.
