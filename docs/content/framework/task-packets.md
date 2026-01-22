---
sidebar_position: 2
title: Task Packets
---

# Task Packets

Task packets provide structured templates for organizing work through all phases of development.

## Templates

Located in `templates/task-packet/`:

1. **00-contract.md** - Task definition and acceptance criteria
2. **10-plan.md** - Implementation plan
3. **20-work-log.md** - Execution log and progress tracking
4. **30-review.md** - Review findings and feedback
5. **40-acceptance.md** - Sign-off and completion

## Usage

```bash
# Create task directory
TASK_ID=$(date +%Y-%m-%d)_feature-name
mkdir -p .ai/tasks/$TASK_ID

# Copy templates
cp .ai-pack/templates/task-packet/*.md .ai/tasks/$TASK_ID/
```text

## Workflow

1. **Define** - Fill out contract with requirements
2. **Plan** - Create implementation plan
3. **Execute** - Implement while updating work log
4. **Review** - Conduct review, document findings
5. **Accept** - Complete acceptance checklist

## Integration with Beads

Task packets complement Beads for comprehensive task management:
- **Beads** - Persistent cross-session task tracking
- **Task Packets** - Detailed work artifacts and documentation
