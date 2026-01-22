---
sidebar_position: 0
title: Framework Overview
---

# AI-Pack Framework Overview

AI-Pack provides a comprehensive workflow framework for AI agent-based software development.

## Core Components

### Quality Gates
Quality gates define rules and constraints that govern what actions are permitted. See [Quality Gates](../gates/index) for details.

### Agent Roles
Different agent personas with specific responsibilities. See [Agent Roles](../roles/index) for details.

### Workflows
Structured processes for different types of work. See [Workflows](../workflows/index) for details.

### Task Packets
Structured templates for organizing work through all phases. Located in `templates/task-packet/`.

### Beads Integration
Git-backed task tracking system that provides cross-session memory for AI agents.

## Directory Structure

```text
your-project/
├── .ai-pack/           # Git submodule (read-only shared pack)
├── .beads/             # Beads task memory system
├── .ai/                # Local workspace
└── docs/               # Permanent documentation
```text

See the [Getting Started](../getting-started) guide for setup instructions.
