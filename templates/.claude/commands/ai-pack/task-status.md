---
description: Show current task packet status and progress
---

# /ai-pack task-status - Task Status Report

Displays the current status of active task packets in `.ai/tasks/`.

## Usage

```
/ai-pack task-status
```

## What This Shows

1. **Active Tasks** - All task packets in `.ai/tasks/`
2. **Current Phase** - Which documents have been created/updated
3. **Completion Status** - Progress indicators for each phase:
   - ✅ Contract defined
   - ✅ Plan created
   - 🔄 Work in progress
   - ⏸️ Review pending
   - ⏸️ Acceptance pending

4. **Next Steps** - What to do next based on current phase

## Example Output

```
📋 Active Task Packets
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Task: add-login-feature
Path: .ai/tasks/2026-01-10_add-login-feature/
Status: Work in Progress

Progress:
  ✅ 00-contract.md   (Requirements defined)
  ✅ 10-plan.md       (Implementation planned)
  🔄 20-work-log.md   (Currently executing)
  ⏸️ 30-review.md     (Not started)
  ⏸️ 40-acceptance.md (Not started)

Next Steps:
  - Continue implementation
  - Update work log regularly
  - Run `/ai-pack review` when ready for quality check
```

## No Active Tasks

If no task packets exist:

```
⚠️ No Active Task Packets

Before starting work, create a task packet:
  /ai-pack task-init <task-name>

This is MANDATORY for all non-trivial tasks.
```

## Related Commands

- `/ai-pack task-init` - Create new task packet
- `/ai-pack review` - Start review workflow
- `/ai-pack help` - Show all ai-pack commands

## Implementation

```bash
!python3 /Users/bryanw/Projects/Claude/ai-pack/templates/.claude/hooks/task-status.py
```
