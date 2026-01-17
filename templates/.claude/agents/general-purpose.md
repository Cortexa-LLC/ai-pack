---
name: general-purpose
description: General-purpose agent for researching complex questions, searching for code, and executing multi-step tasks. Handles file operations, code exploration, and implementation work.
tools: All tools
model: Sonnet
permissionMode: bypassPermissions
---

# General Purpose Spawned Agent

This agent is used by the Orchestrator for parallel execution of independent tasks.

## Permission Mode

**Mode:** `bypassPermissions`

This agent has full Write/Edit permissions without user approval, enabling it to:
- Create and modify files
- Execute build commands
- Run tests
- Perform git operations (excluding destructive ones blocked by deny list)

**Important:** This enables fast parallel execution for engineer tasks. The parent conversation's
deny list (in settings.json) still applies to prevent destructive operations like:
- `rm -rf` commands
- Force push to git
- Destructive file operations

## Usage

This agent is automatically used when Orchestrator spawns parallel engineers via Task tool:

```bash
# Orchestrator spawns parallel engineers
Task(subagent_type="general-purpose", prompt="Implement feature X")
Task(subagent_type="general-purpose", prompt="Write tests for Y")
Task(subagent_type="general-purpose", prompt="Create documentation for Z")
```

All three run in parallel with full file creation permissions.

## Safety

Even with `bypassPermissions`, destructive operations are blocked by the deny list in
`.claude/settings.json`:
- No `rm -rf /` or similar destructive commands
- No force push to git repositories
- No sudo operations
- No destructive file system operations

Engineers can safely create/modify files for implementation work.
