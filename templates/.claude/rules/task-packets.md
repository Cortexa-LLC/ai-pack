---
paths: **/*
---

# Task Packet Requirements

Task packets are the fundamental unit of work tracking in the ai-pack framework.

## When Required

**MANDATORY for all non-trivial tasks:**
- Requires >2 steps
- Involves code changes
- Takes >30 minutes
- Needs verification

**NOT required for:**
- Trivial one-liners
- Reading/exploring code
- Answering questions

## Creating Task Packets

```bash
# Create task packet directory with timestamp
TS=$(date +%Y%m%d%H%M%S)
SLUG="${TID}-${TS}-short-desc"
mkdir -p .ai/tasks/$SLUG

# Two files:
#   task.md   — everything the agent needs to do the work
#   result.md — written by the agent when done
```

The directory layout is:

```
.ai/tasks/<task-id>-<YYYYMMDDHHMMSS>-<short-desc>/
├── task.md      # Requirements, acceptance criteria, files to change, constraints
└── result.md    # Status, summary, findings (written by agent on completion)
```

## task.md — What to Include

**REQUIRED fields:**

- **What to do** — Detailed description of the work
- **Files to change** — Specific paths and what changes are needed
- **Acceptance criteria** — How to verify the work is complete
- **Constraints** — What NOT to change, dependencies, time limits
- **Context** — Background/history (omit if obvious)

**Example:**
```markdown
## What to do

Modify the agent resume functionality to work with tasks that failed due to
timeout. Add --extend flag. Checkpoint must include ResumeReason field.

## Files to change

- internal/server/task_execution.go — write checkpoint on timeout
- internal/server/checkpoint.go — add ResumeReason field
- cmd/agent/commands/server.go — add --extend flag

## Acceptance criteria

- [ ] Tasks that timeout write a checkpoint before failing
- [ ] Resume command accepts failed tasks with "TIMEOUT:" prefix
- [ ] --extend flag allows extending timeout instead of resetting

## Constraints

- Do not touch unrelated server code
- Go build must stay clean
```

## result.md — What to Include

**Written by the agent when the task is complete:**

- Status (DONE / BLOCKED / PARTIAL)
- Summary of what was done
- Findings or decisions made
- Any blockers or follow-up items

**Example:**
```markdown
## Status

DONE

## Summary

Added ResumeReason field to checkpoint, wired timeout handler to write
checkpoint before failing, and added --extend flag to the server command.

## Changes

- internal/server/task_execution.go — timeout writes checkpoint with TIMEOUT: prefix
- internal/server/checkpoint.go — ResumeReason string field added
- cmd/agent/commands/server.go — --extend flag added, default 30m

## Follow-up

None.
```

## Location Rules

**Directory naming convention:**

```
.ai/tasks/<task-id>-<YYYYMMDDHHMMSS>-<short-desc>/
```

**✅ CORRECT:**
```
.ai/tasks/myproj-abc-20260609120000-fix-timeout/
```

**❌ NEVER:**
```
.ai-pack/  (Framework is read-only)
```

## ⚠️ Task Packet Slug ≠ Task ID

The directory name is the **task packet slug** — it is NOT the Task ID.

```
Task packet slug:  HomeControl-qx7-20260424-072021-short-desc   ← directory name
Task ID:           HomeControl-qx7                               ← use with agent CLI
```

Always pass the **Task ID** (the `agent create` output) to `agent` commands:
```bash
agent logs HomeControl-qx7       ✅
agent logs HomeControl-qx7-20260424-072021-short-desc  ❌
```

## Reference

Full documentation: `.ai-pack/gates/00-global-gates.md`
