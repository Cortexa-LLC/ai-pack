# Claude Code Bootstrap Instructions

**Project:** [Your Project Name]
**Repository:** [repo-url]

---

## Framework Integration

This project uses the **ai-pack framework** for structured AI-assisted development.

### Directory Structure

```
project-root/
├── .ai-pack/           # Git submodule (read-only shared framework)
├── .ai/                # Local workspace (project-specific)
│   ├── tasks/          # Active task packets
│   └── repo-overrides.md  # Optional project-specific rules
└── CLAUDE.md           # This file
```

---

## Required Reading: Gates and Standards

Before any task, read these foundational documents:

### Quality Gates (Must Follow)
1. **[.ai-pack/gates/00-global-gates.md](.ai-pack/gates/00-global-gates.md)** - Universal rules (safety, quality, communication)
2. **[.ai-pack/gates/10-persistence.md](.ai-pack/gates/10-persistence.md)** - File operations and state management
3. **[.ai-pack/gates/20-tool-policy.md](.ai-pack/gates/20-tool-policy.md)** - Tool usage policies
4. **[.ai-pack/gates/30-verification.md](.ai-pack/gates/30-verification.md)** - Verification requirements

### Engineering Standards
- **[.ai-pack/quality/engineering-standards.md](.ai-pack/quality/engineering-standards.md)** - Clean code standards index
- **[.ai-pack/quality/clean-code/](.ai-pack/quality/clean-code/)** - Detailed standards by topic

---

## Task Management Protocol

### For Non-Trivial Tasks

**1. Create Task Packet**

Every non-trivial task requires a task packet in `.ai/tasks/`:

```bash
# Create task directory
TASK_ID=$(date +%Y-%m-%d)_task-name
mkdir -p .ai/tasks/$TASK_ID

# Copy templates from .ai-pack
cp .ai-pack/templates/task-packet/00-contract.md .ai/tasks/$TASK_ID/
cp .ai-pack/templates/task-packet/10-plan.md .ai/tasks/$TASK_ID/
cp .ai-pack/templates/task-packet/20-work-log.md .ai/tasks/$TASK_ID/
cp .ai-pack/templates/task-packet/30-review.md .ai/tasks/$TASK_ID/
cp .ai-pack/templates/task-packet/40-acceptance.md .ai/tasks/$TASK_ID/
```

**2. Follow Task Lifecycle**

All task packets go through these phases:

1. **Contract** (`00-contract.md`) - Define requirements and acceptance criteria
2. **Plan** (`10-plan.md`) - Document implementation approach
3. **Work Log** (`20-work-log.md`) - Track execution progress
4. **Review** (`30-review.md`) - Quality assurance
5. **Acceptance** (`40-acceptance.md`) - Sign-off and completion

**3. CRITICAL: Task Packet Location**

✅ **Correct:** `.ai/tasks/YYYY-MM-DD_task-name/`
❌ **NEVER:** `.ai-pack/` (this is shared framework, not for task state)

---

## Role Enforcement

Choose your role based on the task:

### Orchestrator Role
**Use when:** Complex multi-step tasks requiring coordination

**Responsibilities:**
- Break down work into subtasks
- Delegate to worker agents
- Monitor progress
- Coordinate reviews

**Reference:** [.ai-pack/roles/orchestrator.md](.ai-pack/roles/orchestrator.md)

---

### Worker Role
**Use when:** Implementing specific, well-defined tasks

**Responsibilities:**
- Write code and tests
- Follow established patterns
- Update work log
- Report progress and blockers

**Reference:** [.ai-pack/roles/worker.md](.ai-pack/roles/worker.md)

---

### Reviewer Role
**Use when:** Conducting quality assurance

**Responsibilities:**
- Review code against standards
- Verify test coverage
- Check architecture consistency
- Document findings

**Reference:** [.ai-pack/roles/reviewer.md](.ai-pack/roles/reviewer.md)

---

## Workflow Selection

Choose appropriate workflow for the task type:

| Task Type | Workflow | When to Use |
|-----------|----------|-------------|
| General | [standard.md](.ai-pack/workflows/standard.md) | Any task not fitting specialized workflows |
| New Feature | [feature.md](.ai-pack/workflows/feature.md) | Adding new functionality |
| Bug Fix | [bugfix.md](.ai-pack/workflows/bugfix.md) | Fixing defects |
| Refactoring | [refactor.md](.ai-pack/workflows/refactor.md) | Improving code structure |
| Investigation | [research.md](.ai-pack/workflows/research.md) | Understanding code/architecture |

---

## Project-Specific Rules

### Override Location
If this project has specific rules beyond the shared standards:
- **[.ai/repo-overrides.md](.ai/repo-overrides.md)** - Project-specific deltas

### Important Project Context

[Add project-specific information here:]

**Technology Stack:**
- [Language]: [Version]
- [Framework]: [Version]
- [Build Tool]: [Version]

**Key Architectural Patterns:**
- [Pattern 1]
- [Pattern 2]

**Critical Files:**
- [File 1] - [Purpose]
- [File 2] - [Purpose]

**Testing Strategy:**
- Test Framework: [Name]
- Coverage Target: [X]%
- Test Commands: `[command]`

**Build and Deploy:**
- Build: `[command]`
- Test: `[command]`
- Deploy: `[command]`

---

## Common Operations

### Starting a New Task

1. Read gates and standards (see above)
2. Create task packet in `.ai/tasks/`
3. Fill out `00-contract.md`
4. Select appropriate workflow
5. Assume appropriate role
6. Execute workflow phases

### Working on Existing Task

1. Read task packet in `.ai/tasks/YYYY-MM-DD_task-name/`
2. Review current phase
3. Continue from where left off
4. Update work log regularly

### Updating Framework

```bash
# Update shared framework (preserves .ai/tasks/)
git submodule update --remote .ai-pack
git add .ai-pack
git commit -m "Update ai-pack framework"
```

---

## Invariants (Critical)

### ✅ DO
- Create task packets in `.ai/tasks/`
- Follow gates and workflows
- Update work logs regularly
- Reference standards when making decisions
- Ask questions when uncertain

### ❌ NEVER
- Put task packets in `.ai-pack/`
- Edit `.ai-pack/` files directly (contribute to ai-pack repo instead)
- Overwrite `.ai/tasks/` during updates
- Skip gate checkpoints
- Proceed with failing tests

---

## Quick Reference

**Gates:** `.ai-pack/gates/`
**Roles:** `.ai-pack/roles/`
**Workflows:** `.ai-pack/workflows/`
**Templates:** `.ai-pack/templates/`
**Standards:** `.ai-pack/quality/`

**Task Packets:** `.ai/tasks/YYYY-MM-DD_task-name/`
**Overrides:** `.ai/repo-overrides.md` (optional)

---

## Getting Help

- **Framework Documentation:** See `.ai-pack/README.md`
- **Standards Index:** See `.ai-pack/quality/engineering-standards.md`
- **Workflow Guides:** See `.ai-pack/workflows/*.md`
- **Role Definitions:** See `.ai-pack/roles/*.md`

---

**Last Updated:** [Date]
**Framework Version:** [Version from .ai-pack/VERSION]
