# Claude Code Configuration Guide

**Last Updated:** 2026-01-16

> **⚠️ IMPORTANT:** The setup script (`python3 .ai-pack/templates/.claude-setup.py`) now creates the recommended configuration automatically. This document explains the configuration in detail.

This document explains the required Claude Code configuration for ai-pack integration, particularly for spawned agent support.

---

## Critical Configuration: settings.json

The `.claude/settings.json` file controls Claude Code behavior including permissions and hooks.

### Location

```
.claude/settings.json
```

**Important:** This file is typically not committed to git (local configuration). Each developer needs to set it up.

---

## Required Configuration for Spawned Agents

Background agents spawned via the Task tool need file operation permissions. Without proper configuration, agents will fail when trying to create deliverables.

### Recommended settings.json (Auto-created by setup script)

```json
{
  "description": "AI-Pack Framework Configuration for Claude Code",
  "permissions": {
    "allow": [
      "Write(*)",
      "Edit(*)",
      "Bash(mkdir:*)",
      "Bash(cp:*)",
      "Bash(mv:*)",
      "Bash(ls:*)",
      "Bash(cat:*)",
      "Bash(git pull:*)",
      "Bash(git fetch:*)",
      "Bash(git status:*)",
      "Bash(git diff:*)",
      "Bash(git log:*)",
      "Bash(npm:*)",
      "Bash(dotnet:*)",
      "Bash(pytest:*)",
      "Bash(python:*)",
      "Bash(python3:*)"
    ],
    "deny": [
      "Bash(rm -rf /*)",
      "Bash(rm -rf ~/*)",
      "Bash(git push --force*)",
      "Bash(git push -f *)",
      "Bash(sudo rm*)",
      "Bash(sudo chmod*)"
    ],
    "defaultMode": "acceptEdits"
  },
  "hooks": {
    "SessionStart": [
      {
        "description": "Start ai-pack monitoring timers (coordination + watchdog)",
        "hooks": [{
          "type": "command",
          "command": "python3 .claude/hooks/session-start.py"
        }]
      }
    ],
    "SessionEnd": [
      {
        "description": "Stop ai-pack monitoring timers",
        "hooks": [{
          "type": "command",
          "command": "python3 .claude/hooks/session-end.py"
        }]
      }
    ],
    "UserPromptSubmit": [
      {
        "description": "Enforce task packet gate before implementation work",
        "matcher": "",
        "hooks": [{
          "type": "command",
          "command": "python3 .claude/hooks/check-task-packet.py"
        }]
      }
    ]
  }
}
```

**Key Features:**
- ✅ `defaultMode: "acceptEdits"` - Balanced permission model (not the old `"bypassPermissions"`)
- ✅ Explicit allow list for safe operations (file ops, safe git commands, build tools)
- ✅ Explicit deny list for dangerous operations (destructive rm, force push, sudo)
- ✅ Session hooks for monitoring timers
- ✅ Task packet enforcement hook

---

## Configuration Sections

### 1. Permissions (REQUIRED for Spawned Agents)

```json
{
  "permissions": {
    "allow": [
      "Write(*)",
      "Edit(*)",
      "Read(*)"
    ],
    "defaultMode": "bypassPermissions"
  }
}
```

**Why this is needed:**

During Tier 2 validation testing, we discovered spawned agents require explicit permissions to:

1. **Create deliverable files** - Test artifacts, outputs, reports
2. **Modify code files** - Implementation tasks
3. **Write test results** - Validation and acceptance testing
4. **Execute contracts** - Any file operation specified in task contracts

**Permission Breakdown:**

| Permission | Purpose | Example |
|------------|---------|---------|
| `Write(*)` | Create new files anywhere | Creating `output/data.json` from contract |
| `Edit(*)` | Modify existing files | Updating code during implementation |
| `Read(*)` | Read any file | Accessing context, contracts, dependencies |

**defaultMode Options:**

- `"acceptEdits"` - Auto-approve allowed operations, prompt for others (recommended, current default)
- `"bypassPermissions"` - Agents inherit all permissions automatically (deprecated, less secure)
- `"promptUser"` - Ask for permission each time (safest but interrupts agents)

### 2. Hooks (Optional but Recommended)

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [{
          "type": "command",
          "command": "python3 .claude/hooks/check-task-packet.py"
        }]
      }
    ]
  }
}
```

**Purpose:** Enforces task packet requirement before implementation begins.

---

## Security Considerations

### Development Environment (Current Configuration)

```json
{
  "permissions": {
    "allow": [
      "Write(*)",
      "Edit(*)",
      "Read(*)"
    ],
    "defaultMode": "bypassPermissions"
  }
}
```

**Use when:**
- Local development
- Trusted codebase
- Full control over repository
- Running validation tests

**Risk:** Broad file access - agents can modify any file

### Restricted Environment (Production/Shared)

```json
{
  "permissions": {
    "allow": [
      "Write(.ai/tasks/**/*)",
      "Write(.ai/test-artifacts/**/*)",
      "Write(output/**/*)",
      "Edit(src/**/*)",
      "Read(**/*)"
    ],
    "defaultMode": "promptUser"
  }
}
```

**Use when:**
- Shared development environment
- Production codebase
- Security requirements
- Compliance needs

**Trade-off:** Safer but may interrupt agent workflows with permission prompts

### Path Pattern Examples

```json
{
  "permissions": {
    "allow": [
      "Write(.ai/tasks/**/*)",           // Task packets only
      "Write(.ai/test-artifacts/**/*)",  // Test artifacts only
      "Write(tests/**/*)",               // Test files only
      "Edit(src/**/*.py)",               // Python source only
      "Edit(src/**/*.ts)",               // TypeScript source only
      "Read(**/*)"                       // Read anything (usually safe)
    ]
  }
}
```

---

## Validation Testing Requirements

For running the ai-pack validation test suite (Tier 2 tests), you **MUST** use the broad permissions:

```json
{
  "permissions": {
    "allow": [
      "Write(*)",
      "Edit(*)",
      "Read(*)"
    ],
    "defaultMode": "bypassPermissions"
  }
}
```

**Why:**
- Tests spawn spawned agents that create artifacts in `.ai/test-artifacts/`
- Agents need to create multiple directories and files
- Test contracts specify diverse output locations
- Validation requires agents to operate autonomously

**Test failures without proper permissions:**
- Background agents blocked when creating output files
- Contract execution fails with permission errors
- Test artifacts incomplete
- False negative test results

---

## Setup Instructions

### First-Time Setup

1. **Create settings.json:**
   ```bash
   cat > .claude/settings.json <<'EOF'
   {
     "permissions": {
       "allow": [
         "Write(*)",
         "Edit(*)",
         "Read(*)"
       ],
       "defaultMode": "bypassPermissions"
     },
     "hooks": {
       "UserPromptSubmit": [
         {
           "hooks": [{
             "type": "command",
             "command": "python3 .claude/hooks/check-task-packet.py"
           }]
         }
       ]
     }
   }
   EOF
   ```

2. **Verify configuration:**
   ```bash
   cat .claude/settings.json
   ```

3. **Restart Claude Code:**
   - Close and reopen Claude Code session
   - Configuration takes effect on restart

### Verification

Test that spawned agents have write access:

```bash
# This should succeed without permission prompts
python3 tests/multi-agent-tests/test_real_execution.py
```

If agents fail with permission errors, check:
1. `settings.json` exists in `.claude/` directory
2. Permissions section is present and correct
3. Claude Code has been restarted after configuration change

---

## Common Issues

### Issue 1: Spawned Agents Cannot Write Files

**Symptoms:**
- Agents fail with "permission denied" errors
- File creation blocked
- Task execution incomplete

**Solution:**
```bash
# Add permissions to settings.json
cat .claude/settings.json | grep -A5 permissions

# If missing, add:
{
  "permissions": {
    "allow": ["Write(*)", "Edit(*)", "Read(*)"],
    "defaultMode": "bypassPermissions"
  }
}

# Restart Claude Code
```

### Issue 2: Permission Prompts During Agent Execution

**Symptoms:**
- Agent pauses waiting for permission
- Multiple permission prompts
- Agent workflow interrupted

**Solution:**

Change `defaultMode` from `"promptUser"` to `"bypassPermissions"`:

```json
{
  "permissions": {
    "allow": ["Write(*)", "Edit(*)", "Read(*)"],
    "defaultMode": "bypassPermissions"  // <-- Changed
  }
}
```

### Issue 3: Configuration Not Taking Effect

**Symptoms:**
- Changes to settings.json ignored
- Agents still blocked

**Solution:**
1. Save settings.json
2. Close Claude Code completely
3. Reopen Claude Code
4. Try again

---

## Reference

### Claude Code Documentation

- **Settings:** https://code.claude.com/docs/en/settings.md
- **Permissions:** https://code.claude.com/docs/en/permissions.md
- **Hooks:** https://code.claude.com/docs/en/hooks.md

### AI-Pack Documentation

- **Integration Guide:** `templates/.claude/README.md`
- **Setup Script:** `templates/.claude-setup.py`
- **Validation Tests:** `tests/README.md`

---

## History

- **2026-01-16:** Updated to reflect improved security model
  - Changed from `defaultMode: "bypassPermissions"` to `"acceptEdits"`
  - Added explicit allow/deny lists for Bash commands
  - Added SessionStart/SessionEnd hooks for monitoring
  - Setup script now auto-creates recommended configuration
- **2026-01-15:** Initial documentation based on Tier 2 validation findings
  - Identified permission requirements for spawned agents
  - Validated 14-file batch limit
  - Original config used `defaultMode: "bypassPermissions"` (now deprecated)
