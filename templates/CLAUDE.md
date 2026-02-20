# Claude Code Bootstrap Instructions

**Project:** [Your Project Name]
**Repository:** [repo-url]

---

## Agent Orientation

> ⚠️ **FILL THIS IN BEFORE SPAWNING ANY AGENT.** This section is what lets orchestrators
> write precise task briefs on day one. Without it, every agent rediscovers the codebase
> from scratch — burning 600+ turns and $40+ per task. See `docs/product/cold-start-protocol.md`.

### What this project does
[One paragraph: what does this codebase build or do?]

### Key directories
| Path | Purpose |
|------|---------|
| `src/` | [what's here] |
| `tests/` | [what's here] |

### How to build
```bash
[exact build command]
# Expected success: [what success output looks like]
```

### How to run tests
```bash
[exact test command]
# Expected: [pass/fail signal, e.g. "exit 0, all N tests passing"]
```

### Key files for agents
| File | Why agents need to know |
|------|------------------------|
| [path/to/file] | [what it does, when agents will touch it] |

### Conventions agents must follow
- [e.g. "All Go code in a2a-agent/internal/ — no changes outside that path"]
- [e.g. "Run go build ./... after every change to verify compilation"]

### What agents must NOT do
- [e.g. "Never delete files without explicit instruction"]
- [e.g. "Never modify roles/*.md wholesale — append only"]

---

## ⚠️ CRITICAL: Task Packet Requirement

**BEFORE starting ANY non-trivial task, you MUST:**

```bash
# Use the ai-pack command
/ai-pack task-init <task-name>
```

This will:
1. Create task packet directory: `.ai/tasks/<beads-id>-<YYYYMMDDHHMMSS>-<short-desc>/`
2. Copy ALL templates from `.ai-pack/templates/task-packet/`
3. Set up contract, plan, work log, review, and acceptance documents

**Then fill out:**
- `00-contract.md` - Requirements and acceptance criteria
- `10-plan.md` - Implementation approach

**ONLY THEN begin implementation.**

**Non-Trivial = Any task that:**
- Requires >2 steps
- Involves code changes
- Takes >30 minutes
- Needs verification

**This is MANDATORY and enforced by hooks.**

### Acceptance Criteria Must Be Verifiable Commands

When creating tasks via `bd create --acceptance "..."`, acceptance criteria **must be
expressed as shell commands**, not prose descriptions:

```bash
# ✅ Good — agent can run this and verify exit code
--acceptance "go build ./... exits 0 and go test ./... exits 0"

# ✅ Good — specific and measurable
--acceptance "all 704 tests pass: ./build/xasm++ --test exits 0 with no FAIL lines"

# ❌ Bad — agent marks done without checking
--acceptance "files are created and documented"

# ❌ Bad — unverifiable
--acceptance "implementation is correct"
```

Prose acceptance criteria are never enforced. Agents close tasks as "complete" when they
have satisfied the letter of the criterion — a file existing satisfies "file is created"
even if the file is empty or wrong.

---

## 🎯 Default Role: Orchestrator

**UNLESS EXPLICITLY TOLD OTHERWISE, YOU ARE OPERATING AS ORCHESTRATOR.**

This project uses **Orchestrator as the default role** for all interactions. You should:

- **Always assume** you are in Orchestrator role when starting any task
- **Break down complex tasks** and delegate to specialized agents
- **Monitor and coordinate** multiple subtasks
- **Only exit Orchestrator mode** when the user explicitly instructs you to work as a different role

**To work as a different role, the user must explicitly say:**
- "Work as Engineer on this task"
- "Act as Reviewer for this code"
- "Switch to Inspector role"
- etc.

**By default: You are Orchestrator** - coordinate, delegate, and oversee work.

**Orchestrator Reference:** [.ai-pack/roles/orchestrator.md](.ai-pack/roles/orchestrator.md)

---

## ⚠️ CRITICAL: Beads Task Management (MANDATORY)

**ALL task operations MUST use Beads commands (`bd`).**

This is **MANDATORY and ENFORCED** by the [Beads Enforcement Gate](.ai-pack/gates/06-beads-enforcement.md).

### Why Beads?

Beads is a git-backed task memory system that persists task state across AI sessions.
Unlike task packets (which are documentation), **Beads is the source of truth for task state**.

### Required Beads Commands

```bash
# Create tasks (ALWAYS FIRST STEP)
bd create "Task description" --priority high

# View tasks
bd list                    # All tasks
bd list --status open      # Open tasks
bd ready                   # Tasks ready to work on

# Work on tasks
bd start <task-id>         # Start working
bd close <task-id>         # Complete task
bd block <task-id> "reason"  # Mark blocked

# Dependencies
bd dep add <child-id> <parent-id>  # Add dependency

# Task details
bd show <task-id>          # View full task info
```

### ⚠️ CRITICAL: Beads Task Description Requirement

**EVERY `bd create` command MUST include a multi-line description.**

**Required Format:**
```bash
bd create "Title

Working directory: /absolute/path/to/project
Task packet: .ai/tasks/<beads-id>-<YYYYMMDDHHMMSS>-<short-desc>/

Detailed description..." --priority high
```

**NEVER create tasks with just a title** - this causes warnings:
```bash
# ❌ WRONG - Missing description
bd create "Task title" --priority high

# ⚠️ Creates issue without description.
#    Issues without descriptions lack context for future work.
```

---

### Orchestrator MUST Use Beads

As Orchestrator (your default role), you MUST:

1. **Create Beads tasks BEFORE task packets**
   ```bash
   # Step 1: Create Beads task with FULL description (MANDATORY)
   # ⚠️ ALWAYS include Working directory, Task packet, and detailed description
   task_id=$(bd create "Implement user authentication

Working directory: $(pwd)
Task packet: .ai/tasks/$(date +%Y-%m-%d)_user-auth/

Create login/logout endpoints with JWT token validation and session management.
Include password hashing, email verification, and account lockout after failed attempts." \
     --priority high --json | jq -r '.id')

   # Step 2: THEN create task packet
   /ai-pack task-init user-authentication

   # Step 3: Link in contract
   echo "**Beads Task:** ${task_id}" >> .ai/tasks/*/00-contract.md
   ```

2. **Track all spawned agents with Beads**
   ```bash
   # When spawning Engineer agent - ALWAYS include full description
   bd create "Agent: Engineer - Implement login API

Working directory: $(pwd)
Task packet: .ai/tasks/ai-pack-4la-20260124090000-login-api/

Create RESTful login endpoint with JWT generation, password validation, and rate limiting." \
     --assignee "Engineer-1" --priority high
   ```

3. **Monitor progress with Beads**
   ```bash
   bd list --status in_progress   # See active work
   bd list --status blocked        # See blockers
   bd ready                        # Find next available work
   ```

4. **Manage dependencies with Beads**
   ```bash
   bd dep add <child-task> <parent-task>
   ```

### Enforcement

**BLOCKING GATE:** Cannot proceed without Beads tasks.

- Task packets without Beads tasks → BLOCKED
- Agent spawns without Beads tasks → BLOCKED
- Progress monitoring via file inspection → BLOCKED (use `bd list`)

**Reference:** [Beads Enforcement Gate](.ai-pack/gates/06-beads-enforcement.md)

---

## Framework Integration

This project uses the **ai-pack framework** for structured AI-assisted development.

### Directory Structure

```
project-root/
├── .ai-pack/           # Git submodule (read-only shared framework)
│   ├── gates/          # Quality gates (including Beads enforcement)
│   ├── roles/          # Agent roles
│   ├── workflows/      # Development workflows
│   ├── templates/      # Task-packet templates
│   └── quality/        # Clean code standards
├── .beads/             # Beads task database (git-backed, persistent state)
│   ├── beads.db        # Task state database
│   ├── issues.jsonl    # Task history
│   └── config.yaml     # Beads configuration
├── .ai/                # Local workspace (project-specific)
│   ├── tasks/          # Active task packets (documentation)
│   └── repo-overrides.md  # Project-specific rules
├── .claude/            # Claude Code integration (auto-loaded)
│   ├── commands/ai-pack/  # Slash commands
│   ├── skills/         # Auto-triggered roles
│   ├── rules/          # Modular rules
│   ├── hooks/          # Enforcement scripts
│   └── settings.json   # Hook configuration
└── CLAUDE.md           # This file
```

**Key Distinction:**
- **`.beads/`** = Source of truth for task STATE (open, closed, blocked, dependencies)
- **`.ai/tasks/`** = Documentation of task IMPLEMENTATION (contracts, plans, work logs)

---

## Claude Code Integration

This project includes **Claude Code integration** with commands, skills, and hooks that enforce ai-pack standards.

### Available Commands

Type `/ai-pack` to see all commands:

```bash
/ai-pack help              # Show all commands
/ai-pack task-init <name>  # Create task packet
/ai-pack task-status       # Check progress
/ai-pack orchestrate       # Complex coordination
/ai-pack engineer          # Direct implementation
/ai-pack test              # Validate tests
/ai-pack review            # Code review
/ai-pack inspect           # Bug investigation
/ai-pack architect         # Architecture design
/ai-pack designer          # UX workflows
/ai-pack pm                # Product requirements
```

**Automatic Enforcement:**
- Task packet gate enforced via hooks
- Rules auto-loaded for all files
- Skills auto-trigger based on keywords

See: [.claude/README.md](.claude/README.md) for details

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

**CRITICAL REQUIREMENT:** Every non-trivial task MUST have a task packet in `.ai/tasks/` created BEFORE implementation begins.

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

✅ **Correct:** `.ai/tasks/<beads-id>-<YYYYMMDDHHMMSS>-<short-desc>/`
❌ **NEVER:** `.ai-pack/` (this is shared framework, not for task state)

---

## Role Enforcement

**Default Role: Orchestrator** (unless user explicitly specifies otherwise)

### Orchestrator Role (DEFAULT)
**Use when:** All tasks by default, especially complex multi-step work requiring coordination

**Responsibilities:**
- Break down work into subtasks
- Delegate to specialized agents (Engineer, Tester, Reviewer, etc.)
- Monitor progress via Beads
- Coordinate reviews
- Ensure quality gates passed

**You are ALWAYS in this role unless user says otherwise.**

**Reference:** [.ai-pack/roles/orchestrator.md](.ai-pack/roles/orchestrator.md)

---

### Engineer Role
**Use when:** User explicitly requests "Work as Engineer" or "/ai-pack engineer"

**Responsibilities:**
- Direct implementation of specific, well-defined tasks
- Write code and tests
- Follow established patterns
- Update work log

**Reference:** [.ai-pack/roles/engineer.md](.ai-pack/roles/engineer.md)

---

### Reviewer Role
**Use when:** User explicitly requests "Work as Reviewer" or "/ai-pack review"

**Responsibilities:**
- Review code against standards
- Verify test coverage
- Check architecture consistency
- Document findings

**Reference:** [.ai-pack/roles/reviewer.md](.ai-pack/roles/reviewer.md)

---

### Other Specialized Roles

Available via explicit commands:
- `/ai-pack test` - Tester role
- `/ai-pack inspect` - Inspector role (bug investigation)
- `/ai-pack architect` - Architect role (architecture design)
- `/ai-pack designer` - Designer role (UX workflows)
- `/ai-pack product-manager` - Product Manager role (product requirements)

**Unless instructed otherwise: Stay in Orchestrator role.**

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

1. Read task packet in `.ai/tasks/<beads-id>-<YYYYMMDDHHMMSS>-<short-desc>/`
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

### 🔒 Immutability Rule (CRITICAL)

**`.ai-pack/` is IMMUTABLE:**
```
❌ NEVER edit files in .ai-pack/
   - It's a git submodule managed externally
   - Changes will be lost on submodule update
   - Breaks other projects using ai-pack
   - Violates framework contract

❌ NEVER add files to .ai-pack/
   - Not your territory
   - Will cause merge conflicts
   - Breaks submodule integrity

✅ DO read and reference .ai-pack/
   - Use as documentation
   - Follow its patterns
   - Reference in your code

✅ DO update via git submodule:
   git submodule update --remote .ai-pack
```

### 📝 Extension Pattern

**To extend a role or add project-specific behavior:**

1. **Create extension in `.ai/roles/`:**
   ```bash
   mkdir -p .ai/roles/
   vim .ai/roles/<role-name>-extension.md
   ```

2. **Reference base role from `.ai-pack/roles/`:**
   ```markdown
   # <Role Name> Extension - [Project Name]

   **Base Role:** `.ai-pack/roles/<role-name>.md` (immutable, managed by ai-pack)
   **Extension Type:** Project-specific additions
   ```

3. **Document extension in `.ai/repo-overrides.md`:**
   ```markdown
   ## Role Extensions

   ### <Role Name> Extension
   **Extension Location:** `.ai/roles/<role-name>-extension.md`
   **Base Role:** `.ai-pack/roles/<role-name>.md`
   **Extension Summary:** [Brief description]
   ```

4. **Reference in CLAUDE.md** (if commonly used):
   ```markdown
   ## Role Extensions

   This project extends the following ai-pack roles:
   - **<Role Name>**: See [.ai/roles/<role-name>-extension.md](.ai/roles/<role-name>-extension.md)
   ```

**See:** [.ai-pack/ROLE-EXTENSION-GUIDE.md](.ai-pack/ROLE-EXTENSION-GUIDE.md) for complete guide

### ✅ DO
- Create task packets in `.ai/tasks/`
- Create role extensions in `.ai/roles/`
- Follow gates and workflows
- Update work logs regularly
- Reference standards when making decisions
- Document extensions in `.ai/repo-overrides.md`
- Ask questions when uncertain

### ❌ NEVER
- Edit files in `.ai-pack/` (immutable!)
- Add files to `.ai-pack/` (use `.ai/` instead)
- Put task packets in `.ai-pack/`
- Put role extensions in `.claude/` (use `.ai/roles/`)
- Overwrite `.ai/tasks/` during updates
- Skip gate checkpoints
- Proceed with failing tests
- Leave extensions undocumented

---

## Quick Reference

**Gates:** `.ai-pack/gates/`
**Roles:** `.ai-pack/roles/`
**Workflows:** `.ai-pack/workflows/`
**Templates:** `.ai-pack/templates/`
**Standards:** `.ai-pack/quality/`

**Task Packets:** `.ai/tasks/<beads-id>-<YYYYMMDDHHMMSS>-<short-desc>/`
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
