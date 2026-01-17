# AI-Pack Agent Configurations

This directory contains agent configurations that enable background workers to operate with appropriate permissions for parallel execution.

## Why Agent Configurations?

When the Orchestrator spawns parallel engineers via the Task tool, those spawned agents need permission to create/edit files. By default, spawned agents **do not** inherit the main conversation's `bypassPermissions` setting.

**Solution:** Agent-specific `permissionMode: bypassPermissions` in configuration files.

## Agent Files

### general-purpose.md

**Purpose:** Main agent for parallel engineer tasks

**Permissions:** Full Write/Edit permissions via `bypassPermissions`

**Used for:**
- Parallel implementation tasks
- File creation and modification
- Code exploration and refactoring
- Test writing

**Example:**
```bash
# Orchestrator spawns 3 parallel engineers
Task(subagent_type="general-purpose", prompt="Implement feature A")
Task(subagent_type="general-purpose", prompt="Write tests for B")
Task(subagent_type="general-purpose", prompt="Create docs for C")
```

All three agents can create files without user approval.

### Bash.md

**Purpose:** Command execution specialist

**Permissions:** Bash commands via `bypassPermissions`

**Used for:**
- Build operations (npm, dotnet, cargo, etc.)
- Test execution
- Git operations (safe ones)
- Package management

**Example:**
```bash
Task(subagent_type="Bash", prompt="Run build and tests")
Task(subagent_type="Bash", prompt="Install dependencies")
```

## Safety Guarantees

Even with `bypassPermissions`, destructive operations are **blocked by deny list** in `.claude/settings.json`:

```json
{
  "deny": [
    "Bash(rm -rf /*)",
    "Bash(rm -rf ~/*)",
    "Bash(git push --force*)",
    "Bash(sudo rm*)",
    ...
  ]
}
```

**Engineers can safely:**
- ✅ Create and modify source files
- ✅ Run build commands
- ✅ Execute tests
- ✅ Commit changes (with approval)

**Engineers cannot:**
- ❌ Delete system files (`rm -rf /`)
- ❌ Force push to git
- ❌ Run destructive sudo commands
- ❌ Modify system-level configurations

## Permission Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `bypassPermissions` | Auto-approve all operations (subject to deny list) | Background engineers needing file creation |
| `requireApproval` | Ask user for every operation | Interactive work |
| `autoDeny` | Block all operations | Read-only agents |

## Customization

Projects can override agent configurations by creating files in `.claude/agents/` that match the agent name.

**Example:** Custom engineer with restricted permissions:

```yaml
---
name: general-purpose
description: Custom restricted engineer
tools: ["Read", "Grep", "Glob"]
model: Sonnet
permissionMode: requireApproval
---
```

Place in project's `.claude/agents/general-purpose.md` to override template.

## Testing Agent Permissions

To verify spawned agents have correct permissions:

```bash
# Spawn a test spawned agent
claude code --task "Create a test file: echo 'test' > test.txt"

# Should succeed without permission prompt if bypassPermissions is set
# Should prompt if requireApproval
# Should fail if autoDeny
```

## Reference

- Claude Code Docs: [Sub-agents Permission Modes](https://code.claude.com/docs/en/sub-agents.md#permission-modes)
- Claude Code Docs: [IAM System](https://code.claude.com/docs/en/iam.md)
- AI-Pack: [Parallel Execution Config](../../PARALLEL-ENGINEERS-CONFIG.md)

---

**Created:** 2026-01-12
**AI-Pack Version:** v1.1.0+
