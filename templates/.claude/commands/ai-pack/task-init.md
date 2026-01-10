---
description: Create a new task packet with ai-pack templates
---

# /ai-pack task-init - Initialize Task Packet

Creates a new task packet directory with all required templates from the ai-pack framework.

## Usage

```
/ai-pack task-init <task-name>
```

**Example:**
```
/ai-pack task-init add-login-feature
```

## What This Does

1. Creates `.ai/tasks/YYYY-MM-DD_<task-name>/` directory
2. Copies all templates from `.ai-pack/templates/task-packet/`:
   - `00-contract.md` - Requirements and acceptance criteria
   - `10-plan.md` - Implementation approach
   - `20-work-log.md` - Progress tracking
   - `30-review.md` - Quality assurance
   - `40-acceptance.md` - Sign-off and completion
3. Initializes with timestamp and task name

## Arguments

- `$1` (required): Task name (will be slugified: lowercase, hyphens, no spaces)

## Task Packet Structure

After running this command, you'll have:

```
.ai/tasks/2026-01-10_add-login-feature/
├── 00-contract.md      # Start here: Define requirements
├── 10-plan.md          # Then: Plan implementation
├── 20-work-log.md      # During: Track progress
├── 30-review.md        # After: Quality review
└── 40-acceptance.md    # Finally: Sign-off
```

## Next Steps

1. **Fill out the contract** - Edit `00-contract.md`:
   - Define what needs to be done
   - List acceptance criteria
   - Identify stakeholders

2. **Create the plan** - Edit `10-plan.md`:
   - Choose appropriate workflow (feature, bugfix, refactor, etc.)
   - Break down implementation steps
   - Identify risks and dependencies

3. **Select your role** - Based on task complexity:
   - Simple task → `/ai-pack engineer` (implement directly)
   - Complex task → `/ai-pack orchestrate` (coordinate subtasks)
   - Bug investigation → `/ai-pack inspect` (root cause analysis)

4. **Begin implementation** - Follow the workflow phases

## Related Commands

- `/ai-pack task-status` - Check current task progress
- `/ai-pack orchestrate` - Start orchestrator workflow
- `/ai-pack engineer` - Start engineer workflow
- `/ai-pack help` - Show all ai-pack commands

## Implementation

```bash
!python3 /Users/bryanw/Projects/Claude/ai-pack/templates/.claude/hooks/task-init.py "$1"
```
