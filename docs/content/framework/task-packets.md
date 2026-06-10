---
sidebar_position: 2
title: Task Packets
---

# Task Packets

Task packets provide structured templates for organizing work through all phases of development.

## Templates

Located in `templates/task-packet/`:

1. **task.md** - Task brief: description, acceptance criteria, files to change, constraints
2. **result.md** - Completion record written by the agent: status, summary, findings

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
