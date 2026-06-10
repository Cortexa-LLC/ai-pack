# Task Management Enforcement Gate

**Version:** 2.2.0
**Last Updated:** 2026-04-28
**Gate Type:** BLOCKING
**Enforcement Level:** MANDATORY

## Overview

This gate enforces the use of SQLite-backed task management throughout all AI-Pack workflows. The task database provides persistent, centralized task tracking that survives session boundaries and enables cross-session memory.

**Critical Rule:** All task lifecycle operations MUST use `agent` CLI commands. Updating task packets alone is NOT sufficient.

---

## Gate Rules

### Rule 1: Task Creation MUST Use Agent CLI

**REQUIREMENT:** Every non-trivial task MUST be created in the task database before work begins.

**Enforcement:**
```
BEFORE starting ANY task:
  IF task is non-trivial THEN
    REQUIRE: agent create command executed
    REQUIRE: Task ID documented in task packet
    BLOCK: Work without database task
  END IF
```

**Implementation:**
```bash
# MANDATORY - Create task in database
task_id=$(agent create "Implement user authentication" --priority P1 --role engineer --json | jq -r '.task_id')

# Create task packet and fill out task.md
mkdir -p .ai/tasks/${task_id}-$(date +%Y%m%d%H%M%S)-auth/
cp .ai-pack/templates/task-packet/task.md .ai/tasks/${task_id}-$(date +%Y%m%d%H%M%S)-auth/
# ... fill out task.md with actual requirements ...

# Now work can begin
```

**Verification:**
- ✅ `agent list --all` shows the task
- ✅ Task packet has populated task.md
- ✅ `~/.ai-pack/tasks.db` contains task entry

---

### Rule 2: Task Status MUST Use Agent Commands

**REQUIREMENT:** All status changes MUST be reflected in the task database, not just task packets.

**Mandatory Status Updates:**

| Workflow Event | Required Agent Command | Required Packet Update |
|----------------|------------------------|------------------------|
| Starting work | `agent update --claim <id>` | Begin work on task.md brief |
| Completing task | `agent close <id>` | Write result.md |
| Partial update | `agent update <id> --result "progress"` | Note in result.md |

---

### Rule 3: Progress Monitoring MUST Use Agent Commands

**REQUIREMENT:** Orchestrators MUST use agent commands to monitor progress.

**Correct Pattern:**
```bash
# Check overall progress
agent list --all

# Check specific task
agent show ${task_id}

# Check running agents
agent list --running
```

---

## Summary

**Key Principle:** Task packets are DOCUMENTATION. Database is STATE.

**Simple Rule:**
```
BEFORE any task operation:
  Ask: "Did I run the agent command?"
  IF no THEN gate violation
```

---

**Last Updated:** 2026-04-28
**Enforcement Level:** MANDATORY, BLOCKING
