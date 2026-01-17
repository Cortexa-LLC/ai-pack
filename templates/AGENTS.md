# Codex Bootstrap Instructions

**Project:** [Your Project Name]
**Repository:** [repo-url]

---

## CRITICAL: Task Packet Requirement

BEFORE starting any non-trivial task, you MUST create a task packet.

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

Then fill out:
- `00-contract.md` - Requirements and acceptance criteria
- `10-plan.md` - Implementation approach

ONLY THEN begin implementation.

Non-trivial = any task that:
- Requires more than 2 steps
- Involves code changes
- Takes more than 30 minutes
- Needs verification

---

## Framework Integration

This project uses the ai-pack framework for structured AI-assisted development.

### Directory Structure

```
project-root/
├── .ai-pack/           # Git submodule (read-only shared framework)
│   ├── gates/          # Quality gates
│   ├── roles/          # Agent roles
│   ├── workflows/      # Development workflows
│   ├── templates/      # Task-packet templates
│   └── quality/        # Clean code standards
├── .ai/                # Local workspace (project-specific)
│   ├── tasks/          # Active task packets
│   └── repo-overrides.md  # Project-specific rules
├── .codex/             # Optional Codex-specific guidance
│   └── rules/          # Extra rules referenced by AGENTS.md
└── AGENTS.md           # This file
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

### MANDATORY: Task Packet Creation

Every non-trivial task MUST have a task packet in `.ai/tasks/` created BEFORE implementation begins.

### Task Lifecycle

All task packets go through these phases:

1. Contract (`00-contract.md`) - Define requirements and acceptance criteria
2. Plan (`10-plan.md`) - Document implementation approach
3. Work Log (`20-work-log.md`) - Track execution progress
4. Review (`30-review.md`) - Quality assurance
5. Acceptance (`40-acceptance.md`) - Sign-off and completion

### CRITICAL: Task Packet Location

Correct: `.ai/tasks/YYYY-MM-DD_task-name/`
Never: `.ai-pack/` (shared framework, not for task state)

---

## Role Selection

Choose your role based on the task:

### Orchestrator Role
Use when: Complex multi-step tasks requiring coordination

Responsibilities:
- Break down work into subtasks
- Delegate to worker agents
- Monitor progress
- Coordinate reviews

Reference: [.ai-pack/roles/orchestrator.md](.ai-pack/roles/orchestrator.md)

---

### Worker Role
Use when: Implementing specific, well-defined tasks

Responsibilities:
- Write code and tests
- Follow established patterns
- Update work log
- Report progress and blockers

Reference: [.ai-pack/roles/worker.md](.ai-pack/roles/worker.md)

---

### Reviewer Role
Use when: Conducting quality assurance

Responsibilities:
- Review code against standards
- Verify test coverage
- Check architecture consistency
- Document findings

Reference: [.ai-pack/roles/reviewer.md](.ai-pack/roles/reviewer.md)

---

## Execution Preferences

Follow the execution strategy defined in the roles, especially `.ai-pack/roles/orchestrator.md`:

- Orchestrator runs in the foreground and delegates work to role-specific agents.
- Perform mandatory execution strategy analysis for 2+ subtasks; document PARALLEL/SEQUENTIAL/HYBRID.
- For 3+ independent subtasks, parallel execution is required; launch workers together.
- WIP limits for background agents: max 3, preferred 2, ideal 1.
- Background agents that write files require write permissions; otherwise run them in foreground.
- For every spawned agent, create a corresponding Beads task and verify artifacts persist.

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

Override location:
- **[.ai/repo-overrides.md](.ai/repo-overrides.md)** - Project-specific deltas
- **[.codex/rules/](.codex/rules/)** - Optional Codex-specific rules

Important project context:

[Add project-specific information here:]

Technology stack:
- [Language]: [Version]
- [Framework]: [Version]
- [Build Tool]: [Version]
