---
description: Show all ai-pack commands and framework overview
---

# /ai-pack help - Command Reference

## Available Commands

### Task Management

**`/ai-pack task-init <task-name>`**
Create a new task packet with templates.
Example: `/ai-pack task-init add-login`

**`/ai-pack task-status`**
Show current task progress and status.

---

### Role Selection

**`/ai-pack orchestrate`**
Coordinate complex multi-step tasks (3+ subtasks, parallel execution).

**`/ai-pack engineer`**
Implement well-defined tasks with clear requirements.

**`/ai-pack inspect`**
Investigate complex bugs with root cause analysis.

**`/ai-pack architect`**
Design technical architecture and system design.

**`/ai-pack designer`**
Create UX workflows and HTML wireframes.

**`/ai-pack pm`**
Define product requirements and create PRDs.

---

### Quality Assurance

**`/ai-pack test`**
Validate TDD compliance and test coverage (MANDATORY before review).

**`/ai-pack review`**
Conduct code quality review (MANDATORY before completion).

---

### Help

**`/ai-pack help`**
Show this help message.

---

## Quick Start

### 1. Starting a New Task

```bash
# Create task packet
/ai-pack task-init my-feature

# Fill out contract and plan (in .ai/tasks/YYYY-MM-DD_my-feature/)
# - Edit 00-contract.md (requirements)
# - Edit 10-plan.md (approach)

# Choose role based on complexity
/ai-pack engineer        # Simple, straightforward task
/ai-pack orchestrate     # Complex, multi-step task
```

### 2. During Implementation

```bash
# Check progress
/ai-pack task-status

# Update work log regularly
# Edit .ai/tasks/*/20-work-log.md
```

### 3. Quality Assurance

```bash
# First: Validate tests
/ai-pack test

# Second: Review code
/ai-pack review

# Fix any issues found, then re-run
```

---

## When to Use Each Role

| Role | When to Use | Output |
|------|-------------|--------|
| **Orchestrator** | 3+ subtasks, coordination needed | Delegation plan, parallel execution |
| **Engineer** | Clear requirements, ready to code | Implementation, tests |
| **Inspector** | Complex bug, unknown root cause | RCA document, fix plan |
| **Architect** | Architecture decisions needed | Architecture doc, ADRs |
| **Designer** | User-facing UI/UX needed | Wireframes, UX flows |
| **PM** | Large feature, unclear requirements | PRD, user stories |
| **Tester** | Validate test coverage | Test assessment |
| **Reviewer** | Validate code quality | Code review report |

---

## Directory Structure

```
project-root/
├── .ai-pack/              # Framework (git submodule, read-only)
│   ├── gates/             # Quality gates
│   ├── roles/             # Role definitions
│   ├── workflows/         # Process workflows
│   ├── templates/         # Templates
│   └── quality/           # Standards
│
├── .ai/                   # Project workspace
│   ├── tasks/             # Active task packets
│   │   └── YYYY-MM-DD_task-name/
│   │       ├── 00-contract.md
│   │       ├── 10-plan.md
│   │       ├── 20-work-log.md
│   │       ├── 30-review.md
│   │       └── 40-acceptance.md
│   └── repo-overrides.md  # Project-specific rules
│
└── docs/                  # Persisted artifacts
    ├── product/           # PRDs
    ├── architecture/      # Architecture docs
    ├── design/            # UX wireframes
    ├── adr/               # Architecture Decision Records
    └── investigations/    # Bug RCAs
```

---

## Task Packet Lifecycle

```
1. CONTRACT (00-contract.md)
   ↓ Define requirements and acceptance criteria

2. PLAN (10-plan.md)
   ↓ Document implementation approach

3. WORK LOG (20-work-log.md)
   ↓ Track execution progress

4. REVIEW (30-review.md)
   ↓ Quality assurance (Tester + Reviewer)

5. ACCEPTANCE (40-acceptance.md)
   ↓ Sign-off and completion
```

---

## Mandatory Gates

### Code Quality Review Gate
- **Tester validation** - MUST approve tests
- **Reviewer validation** - MUST approve code
- Both required before work complete

### Execution Strategy Gate
- Analyze parallelization opportunities
- Document sequential vs parallel
- Required for 3+ subtasks

### Persistence Gate
- Planning artifacts → `docs/`
- Cross-references between documents
- Committed to repository

See: `.ai-pack/gates/` for all gates

---

## Key Standards

### Test-Driven Development (TDD)
```
RED → Write failing test
GREEN → Write minimal code to pass
REFACTOR → Clean up code
REPEAT → Next requirement
```

Coverage targets: 80-90% overall, 95%+ critical logic

### Clean Code
- No magic numbers
- No god functions (>20 lines)
- No deep nesting (>3 levels)
- Meaningful names
- Single responsibility

See: `.ai-pack/quality/engineering-standards.md`

---

## C# Projects (MANDATORY)

Before committing C# code:

```bash
# Step 1: Format code
dotnet csharpier .

# Step 2: Build with analyzers
dotnet build /warnaserror

# Step 3: Run tests
dotnet test

# All must pass before commit
```

See: `.ai-pack/quality/clean-code/csharp-modern-tooling.md`

---

## Getting Help

**Framework Documentation:**
- `.ai-pack/README.md` - Overview
- `.ai-pack/roles/*.md` - Role details
- `.ai-pack/workflows/*.md` - Workflow guides
- `.ai-pack/gates/*.md` - Quality gates
- `.ai-pack/quality/` - Coding standards

**Project-Specific:**
- `CLAUDE.md` - Project context
- `.ai/repo-overrides.md` - Project rules

---

## Common Workflows

### Feature Development
```bash
/ai-pack task-init new-feature
# Optional: /ai-pack pm (if requirements unclear)
# Optional: /ai-pack architect (if complex design)
# Optional: /ai-pack designer (if UI/UX needed)
/ai-pack engineer           # or /ai-pack orchestrate
/ai-pack test
/ai-pack review
```

### Bug Fix
```bash
/ai-pack task-init fix-bug-name
# Optional: /ai-pack inspect (if root cause unclear)
/ai-pack engineer
/ai-pack test
/ai-pack review
```

### Refactoring
```bash
/ai-pack task-init refactor-component
/ai-pack engineer
/ai-pack test
/ai-pack review
```

---

## Tips

1. **Always create task packet first** - Required for non-trivial tasks
2. **Update work log regularly** - Document decisions and progress
3. **Run Tester before Reviewer** - Tests must validate first
4. **Use parallel execution** - When subtasks are independent
5. **Persist planning artifacts** - Product/Architecture/Investigation docs go in `docs/`
6. **Cross-reference documents** - Link PRD ↔ Architecture ↔ Code

---

## Framework Version

Check version: `cat .ai-pack/VERSION`

Update framework:
```bash
git submodule update --remote .ai-pack
git add .ai-pack
git commit -m "Update ai-pack framework"
```

---

For more detailed information, see the full documentation in `.ai-pack/`
