# AI-Pack

**A comprehensive AI agent workflow framework for software development**

AI-Pack provides structured processes, quality gates, agent roles, and coding standards for AI agent-based software development. It ensures quality, consistency, and proper governance throughout the development lifecycle.

## Overview

AI-Pack is designed as a git submodule that projects include at `.ai-pack/`. It provides:

1. **AI Workflow Framework** - Structured processes for AI agent-based development
   - Quality gates (enforcement rules)
   - Agent roles (orchestrator, engineer, reviewer, etc.)
   - Development workflows (feature, bugfix, refactor, research)
   - Task packet templates (contract, plan, work-log, review, acceptance)

2. **Clean Code Standards** - Industry-leading coding principles and practices
   - Universal design principles (SOLID, DRY, YAGNI, etc.)
   - Language-specific guidelines (C++, Python, JavaScript/TypeScript, Java, Kotlin)
   - Testing best practices and TDD workflow
   - Architecture patterns and refactoring techniques

These components work together seamlessly with AI coding assistants like Claude Code through `.ai-pack` integration.

---

## AI Workflow Framework

The AI Workflow Framework provides structured processes, roles, and templates for AI agent-based software development.

### Framework Components

#### 🚦 Gates - Quality Controls
Quality gates define rules and constraints that govern what actions are permitted. Located in `gates/`:

- **[00-global-gates.md](gates/00-global-gates.md)** - Universal rules (safety, quality, communication)
- **[10-persistence.md](gates/10-persistence.md)** - File operations and state management rules
- **[20-tool-policy.md](gates/20-tool-policy.md)** - Tool usage policies and approvals
- **[25-execution-strategy.md](gates/25-execution-strategy.md)** - **MANDATORY** execution strategy analysis and parallel engineer enforcement
- **[30-verification.md](gates/30-verification.md)** - Verification and validation requirements
- **[35-code-quality-review.md](gates/35-code-quality-review.md)** - **MANDATORY** Tester and Reviewer validation for all code changes

#### 👥 Roles - Agent Personas
Roles define different agent personas with specific responsibilities. Located in `roles/`:

- **[orchestrator.md](roles/orchestrator.md)** - High-level coordinator, delegates work, monitors progress
  - **ENFORCED:** Automatically analyzes and applies parallel execution for 3+ independent subtasks (max 5 concurrent)
  - **MANDATORY:** Must complete execution strategy analysis before delegation (enforced by [Execution Strategy Gate](gates/25-execution-strategy.md))
  - **MANDATORY:** Must delegate to Tester and Reviewer for all code changes (enforced by [Code Quality Review Gate](gates/35-code-quality-review.md))
- **[engineer.md](roles/engineer.md)** - Implementation specialist, writes code, creates tests
  - Executes specific tasks following TDD workflow and established patterns
- **[inspector.md](roles/inspector.md)** - Bug investigation specialist, conducts root cause analysis
  - **Investigates:** Bug reports, reproduces issues, identifies root cause
  - **Delivers:** RCA document, task packet for Engineer, regression test specifications
  - **Optional:** Invoked by Orchestrator for complex bugs or directly by user
- **[product-manager.md](roles/product-manager.md)** - Requirements specialist, creates PRDs and user stories
  - **Defines:** Product requirements, success metrics, epics and user stories (JIRA-style)
  - **Collaborates:** Works with Engineers and Architect on technical feasibility and breakdown
  - **Delivers:** PRD, epics, user stories with acceptance criteria
  - **Optional:** Invoked by Orchestrator for large features or directly by user
- **[designer.md](roles/designer.md)** - UX specialist, creates user flows and wireframes for value stream delivery
  - **Designs:** User workflows, journey maps, wireframes (HTML for web/iOS/Android), design specifications
  - **Collaborates:** Works with Product Manager on requirements, Architect on feasibility
  - **Delivers:** User research, user flows, wireframes, design specs, accessibility requirements
  - **Optional:** Invoked by Orchestrator for user-facing features with significant UI/UX work
- **[architect.md](roles/architect.md)** - Technical design specialist, system architecture and design
  - **Designs:** System architecture, API specifications, data models, technology choices
  - **Collaborates:** Works with Product Manager and Designer on feasibility, Engineers on implementation
  - **Delivers:** Architecture documents, API specs, data models, ADRs
  - **Optional:** Invoked by Orchestrator for complex features requiring architectural design
- **[tester.md](roles/tester.md)** - Testing specialist, validates TDD compliance and test sufficiency
  - **ENFORCED:** Mandatory validation for all code changes
  - **Validates:** TDD process, coverage (80-90%), test quality, test scenarios
- **[reviewer.md](roles/reviewer.md)** - Quality assurance, code review, standards compliance
  - **ENFORCED:** Mandatory validation for all code changes
  - **Reviews:** Code quality, architecture, security, documentation

**Configuration:** See **[PARALLEL-ENGINEERS-CONFIG.md](PARALLEL-ENGINEERS-CONFIG.md)** for enforced parallel execution details

#### 🔄 Workflows - Development Processes
Workflows define structured processes for different types of work. Located in `workflows/`:

- **[standard.md](workflows/standard.md)** - General workflow for any task
- **[feature.md](workflows/feature.md)** - Adding new functionality
- **[bugfix.md](workflows/bugfix.md)** - Fixing defects
- **[refactor.md](workflows/refactor.md)** - Improving code structure
- **[research.md](workflows/research.md)** - Investigating and understanding code

#### 📋 Task-Packet Templates
Structured templates for organizing work through all phases. Located in `templates/task-packet/`:

- **[00-contract.md](templates/task-packet/00-contract.md)** - Task definition and acceptance criteria
- **[10-plan.md](templates/task-packet/10-plan.md)** - Implementation plan
- **[20-work-log.md](templates/task-packet/20-work-log.md)** - Execution log and progress tracking
- **[30-review.md](templates/task-packet/30-review.md)** - Review findings and feedback
- **[40-acceptance.md](templates/task-packet/40-acceptance.md)** - Sign-off and completion

### Deployment Model

The ai-pack framework is designed for the following structure in your projects:

```
your-project/
├── .ai-pack/                        # Git submodule (read-only shared pack)
│   ├── quality/                     # Clean code standards
│   ├── gates/                       # Quality gates
│   ├── roles/                       # Agent roles
│   ├── workflows/                   # Development workflows
│   └── templates/                   # Task-packet templates
│
├── .ai/                             # Local workspace (your project)
│   ├── tasks/                       # Active task packets (temporary)
│   │   └── 2026-01-07_feature-x/   # Example task
│   │       ├── 00-contract.md      # From template
│   │       ├── 10-plan.md          # From template
│   │       ├── 20-work-log.md      # From template
│   │       ├── 30-review.md        # From template
│   │       └── 40-acceptance.md    # From template
│   └── repo-overrides.md           # Optional project-specific deltas
│
├── docs/                            # Permanent documentation (committed)
│   ├── product/                     # Product requirements
│   │   └── [feature-name]/
│   │       ├── prd.md               # Product Requirements Document
│   │       ├── epics.md             # Epic definitions
│   │       └── user-stories.md      # User stories with acceptance criteria
│   ├── design/                      # UX design and wireframes
│   │   └── [feature-name]/
│   │       ├── user-research.md     # User research and insights
│   │       ├── user-flows.md        # User flows and journey maps
│   │       ├── design-specs.md      # Design specifications
│   │       └── wireframes/          # HTML wireframes (viewable in browser)
│   │           ├── wireframe-web.html
│   │           ├── wireframe-ios.html
│   │           └── wireframe-android.html
│   ├── architecture/                # Technical design
│   │   └── [feature-name]/
│   │       ├── architecture.md      # System architecture
│   │       ├── api-spec.md          # API specifications
│   │       └── data-models.md       # Data models and schemas
│   ├── adr/                         # Architecture Decision Records
│   │   ├── 001-decision-title.md    # Sequentially numbered
│   │   ├── 002-decision-title.md
│   │   └── README.md                # Index of all ADRs
│   └── investigations/              # Bug retrospectives
│       ├── BUG-123-description.md
│       └── README.md                # Index by root cause category
│
└── CLAUDE.md                        # Bootstrap instructions for AI
```

**Key Concepts:**
- **`.ai-pack/`** - Git submodule containing shared standards and framework (this repository) - READ-ONLY
- **`.ai/`** - Local workspace in your project for task state and overrides - PROJECT-SPECIFIC, TEMPORARY
- **`docs/`** - Permanent documentation repository - PROJECT-SPECIFIC, COMMITTED
- **`CLAUDE.md`** - Bootstrap instructions at project root (copy from `templates/CLAUDE.md`)
- **Task packets** - Instances of templates created in `.ai/tasks/` for each task
- **Repo overrides** - Project-specific customizations to shared standards

**Critical Invariants:**
- ✅ Task packets go in `.ai/tasks/` (never in `.ai-pack/`)
- ✅ `.ai-pack/` is read-only shared framework
- ✅ `.ai/tasks/` preserved during framework updates
- ✅ Framework improvements happen in ai-pack repo (not ad hoc in projects)
- ✅ Planning artifacts persisted to `docs/` when transitioning to implementation
- ✅ `.ai/tasks/` is temporary, `docs/` is permanent

### Artifact Persistence Pattern

AI-Pack enforces a **two-tier documentation system**:

**Temporary: `.ai/tasks/`** (Work-in-Progress)
- Active task packets during development
- Draft plans, work logs, review notes
- Cleaned up after task completion
- NOT committed to long-term repository

**Permanent: `docs/`** (Long-Lived Documentation)
- Product requirements (PRDs, epics, user stories)
- Architecture designs (system docs, API specs, data models)
- Architecture Decision Records (ADRs)
- Bug investigation retrospectives
- COMMITTED to repository for long-term reference

**Persistence Triggers:**

When planning phases complete and work transitions to implementation, artifacts MUST be persisted:

```
Product Manager Phase Complete:
  .ai/tasks/[id]/prd.md          → docs/product/[feature-name]/prd.md
  .ai/tasks/[id]/epics.md        → docs/product/[feature-name]/epics.md
  .ai/tasks/[id]/user-stories.md → docs/product/[feature-name]/user-stories.md

Architect Phase Complete:
  .ai/tasks/[id]/architecture.md → docs/architecture/[feature-name]/architecture.md
  .ai/tasks/[id]/api-spec.md     → docs/architecture/[feature-name]/api-spec.md
  .ai/tasks/[id]/data-models.md  → docs/architecture/[feature-name]/data-models.md
  .ai/tasks/[id]/adrs/adr-*.md   → docs/adr/adr-NNN-*.md

Bug Fix Verified:
  .ai/tasks/[id]/retrospective.md → docs/investigations/BUG-ID-description.md
```

**Why This Matters:**
- **Long-term knowledge**: PRDs and architecture docs referenced for years
- **Team onboarding**: New developers understand "why" behind decisions
- **Traceability**: Clear chain from requirements → design → implementation
- **Organizational learning**: Bug patterns inform systemic improvements
- **Version control**: Track evolution of requirements and designs

**Enforcement:** See [10-persistence.md](gates/10-persistence.md) - Section 11: "Artifact Repository Persistence"

### Quick Start

#### 1. Add Framework to Your Project

```bash
# Add ai-pack as submodule
cd your-project
git submodule add https://github.com/Cortexa-LLC/ai-pack .ai-pack
git submodule update --init --recursive

# Create local workspace
mkdir -p .ai/tasks

# Copy bootstrap template to project root
cp .ai-pack/templates/CLAUDE.md ./CLAUDE.md

# Customize CLAUDE.md with project-specific details
# (Edit project name, tech stack, key files, etc.)
```

#### 2. Create a Task Packet

When starting a new task, create a task packet from templates:

```bash
# Create task directory
TASK_ID=$(date +%Y-%m-%d)_feature-name
mkdir -p .ai/tasks/$TASK_ID

# Copy templates
cp .ai-pack/templates/task-packet/00-contract.md .ai/tasks/$TASK_ID/
cp .ai-pack/templates/task-packet/10-plan.md .ai/tasks/$TASK_ID/
cp .ai-pack/templates/task-packet/20-work-log.md .ai/tasks/$TASK_ID/
cp .ai-pack/templates/task-packet/30-review.md .ai/tasks/$TASK_ID/
cp .ai-pack/templates/task-packet/40-acceptance.md .ai/tasks/$TASK_ID/
```

#### 3. Follow the Workflow

1. **Define** - Fill out `00-contract.md` with requirements
2. **Plan** - Create implementation plan in `10-plan.md`
3. **Execute** - Implement while updating `20-work-log.md`
4. **Review** - Conduct review, document in `30-review.md`
5. **Accept** - Complete acceptance checklist in `40-acceptance.md`

### Use Cases

**Multi-Agent Development:**
- Orchestrator agent coordinates complex features
- Engineer agents implement specific components
- Reviewer agents ensure quality and compliance

**Structured Task Management:**
- Clear contracts define expectations upfront
- Plans document approach before implementation
- Work logs track progress and decisions
- Reviews ensure quality
- Acceptance provides formal sign-off

**Quality Governance:**
- Gates enforce safety and quality rules
- Workflows ensure consistent processes
- Templates provide structured documentation
- Standards guide implementation

**Knowledge Capture:**
- Task packets document complete history
- Decisions and rationale preserved
- Lessons learned captured
- Future reference enabled

---

## Clean Code Standards

The `quality/clean-code/` directory contains comprehensive coding standards and best practices based on industry-leading sources including Martin Fowler's design principles, SOLID principles, and modern language-specific best practices.

### Universal Standards

- **[00-general-rules.md](quality/clean-code/00-general-rules.md)** - Universal rules: No tabs (spaces only), TDD workflow, all tests must pass (zero tolerance for failures), 80-90% code coverage target

### Core Design Principles

- **[01-design-principles.md](quality/clean-code/01-design-principles.md)** - Beck's Four Rules of Simple Design, Tell Don't Ask, Dependency Injection, and Seams for Testability
- **[02-solid-principles.md](quality/clean-code/02-solid-principles.md)** - Complete SOLID principles with practical examples and code smells
- **[03-refactoring.md](quality/clean-code/03-refactoring.md)** - Code smells catalog and refactoring techniques
- **[04-testing.md](quality/clean-code/04-testing.md)** - Test Pyramid, test doubles, and testing best practices
- **[05-architecture.md](quality/clean-code/05-architecture.md)** - Bounded Contexts and architectural patterns
- **[06-code-review-checklist.md](quality/clean-code/06-code-review-checklist.md)** - Comprehensive code review guidelines

### Development Practices

- **[07-development-practices.md](quality/clean-code/07-development-practices.md)** - YAGNI, Frequency Reduces Difficulty, Continuous Integration, Technical Debt, Refactoring
- **[08-deployment-patterns.md](quality/clean-code/08-deployment-patterns.md)** - Feature Toggles, Blue-Green Deployment, Canary Release, Parallel Change
- **[09-system-evolution.md](quality/clean-code/09-system-evolution.md)** - Strangler Fig, Sacrificial Architecture, MonolithFirst, Semantic Diffusion

### API and Interface Design

- **[10-api-design.md](quality/clean-code/10-api-design.md)** - Command Query Separation, Naming conventions, API design principles

### Documentation Standards

- **[11-documentation-standards.md](quality/clean-code/11-documentation-standards.md)** - Inline documentation, Markdown+Mermaid diagrams, and Architecture Decision Records (ADRs)

### Language-Specific Guidelines

Each language follows its community's established indentation standards:

| Language | Indentation | File |
|----------|-------------|------|
| C++ | 2 spaces | [lang-cpp.md](quality/clean-code/lang-cpp-basics.md) |
| C# | 4 spaces | [lang-csharp.md](quality/clean-code/lang-csharp.md) |
| Python | 4 spaces | [lang-python.md](quality/clean-code/lang-python.md) |
| JavaScript/TypeScript | 2 spaces | [lang-javascript.md](quality/clean-code/lang-javascript.md) |
| Java | 2 spaces | [lang-java.md](quality/clean-code/lang-java.md) |
| Kotlin | 4 spaces | [lang-kotlin.md](quality/clean-code/lang-kotlin.md) |

**[C++ Guidelines](quality/clean-code/lang-cpp-basics.md)** - Comprehensive C++ guidelines:
- All 55 items from Scott Meyers' *Effective C++*
- C++ Core Guidelines (P, F, I, C, R, ES, E, CP, Enum, Con, T, Per, SF, SL sections)
- Modern C++17/20 best practices
- 2-space indentation (Google C++ Style Guide)
- See also: [lang-cpp-design.md](quality/clean-code/lang-cpp-design.md), [lang-cpp-advanced.md](quality/clean-code/lang-cpp-advanced.md), [lang-cpp-modern.md](quality/clean-code/lang-cpp-modern.md), [lang-cpp-guidelines.md](quality/clean-code/lang-cpp-guidelines.md), [lang-cpp-reference.md](quality/clean-code/lang-cpp-reference.md)

**[C# Guidelines](quality/clean-code/lang-csharp.md)** - C# and .NET best practices:
- Microsoft C# Coding Conventions
- StyleCop Analyzers (mandatory)
- Modern C# 12 features
- Async/await patterns
- 4-space indentation (Microsoft standard)

**[Python Guidelines](quality/clean-code/lang-python.md)** - Python best practices:
- PEP 8, PEP 20 (Zen of Python)
- Type hints and modern Python features
- 4-space indentation (PEP 8 mandatory)

**[JavaScript/TypeScript Guidelines](quality/clean-code/lang-javascript.md)** - JavaScript/TypeScript:
- Microsoft TypeScript Coding Guidelines
- Double quotes, prefer undefined over null, no I prefix
- 2-space indentation (JavaScript ecosystem standard)

**[Java Guidelines](quality/clean-code/lang-java.md)** - Java guidelines:
- Google Java Style Guide (with Cortexa LLC override for indentation)
- Effective Java (Joshua Bloch)
- Spring Framework patterns
- SonarQube default Java rules (mandatory)
- 2-space indentation (Cortexa LLC override)

**[Kotlin Guidelines](quality/clean-code/lang-kotlin.md)** - Kotlin conventions:
- Kotlin Coding Conventions (JetBrains)
- SonarQube default Kotlin rules (mandatory)
- Coroutines and Flow patterns
- Android best practices
- 4-space indentation (JetBrains standard)

### Integration with AI Workflow Framework

The Clean Code Standards and AI Workflow Framework work together:

- **Gates** enforce the standards defined in `quality/clean-code/`
- **Workflows** reference standards for implementation guidance
- **Reviewer role** validates compliance with standards
- **Task packets** document adherence to standards

**Quick Access:**
- Full standards: `quality/clean-code/` directory
- Quick reference: [quality/clean-code/RULES_REFERENCE.md](quality/clean-code/RULES_REFERENCE.md)

---

## Usage

### Option 1: Git Submodule (Recommended for Teams)

Add these standards to your project as a submodule:

```bash
cd your-project
git submodule add https://github.com/Cortexa-LLC/ai-pack .ai-pack
git submodule update --init --recursive
```

Update standards in your project:

```bash
git submodule update --remote
git add .ai-pack
git commit -m "Update ai-pack framework"
```

### Option 2: Symbolic Link (For Local Development)

Create a symbolic link to a single shared copy:

```bash
cd your-project
ln -s /path/to/ai-pack .ai-pack
```

### Option 3: Direct Copy (For Standalone Projects)

Copy the standards directly into your project:

```bash
cp -r /path/to/ai-pack .ai-pack
```

## Integration with AI Assistants

AI-Pack is designed to work with Claude Code and other AI assistants that support `.ai-pack`:

1. Add this repository as a submodule to `.ai-pack/` in your project
2. The framework files will be automatically discovered by Claude Code
3. AI assistants will apply these standards and workflows during development

## Project-Specific Customization

### Two-Tier Rule System

AI-Pack supports a **two-tier approach** for managing shared and project-specific rules:

**Tier 1: Shared Standards** (from this submodule)
- Core design principles
- SOLID principles
- Language-specific guidelines
- Universal best practices
- AI workflow framework

**Tier 2: Project-Specific Rules** (in your project)
- Project conventions
- Team preferences
- Technology stack specifics
- Workflow and tooling

### How to Add Project-Specific Rules

When you add this repository as a submodule to `.ai-pack/`, you can also add project-specific rule files directly to the same directory. These files are git-ignored by the submodule but tracked in your project repository.

**Naming Convention for Project Files:**
- `PROJECT-*.md` - Project-specific rules (e.g., `PROJECT-sourcerer.md`)
- `PROJECT-README.md` - Overview of your project's rule structure

**Example Setup:**

```bash
# Add submodule
git submodule add https://github.com/Cortexa-LLC/ai-pack .ai-pack

# Add project-specific rules to the same directory
cat > .ai-pack/PROJECT-README.md << 'EOF'
# My Project Coding Standards

This project uses a **two-tier rule system**:

## Tier 1: Shared Standards (Submodule)
All files without `PROJECT-` prefix come from the Cortexa ai-pack.

## Tier 2: Project-Specific Rules
- `PROJECT-myproject.md` - Project-specific conventions
- `PROJECT-architecture.md` - Architecture rules

**Both tiers are automatically discovered by Claude Code and other AI assistants.**
EOF

# Add your project rules
cat > .ai-pack/PROJECT-myproject.md << 'EOF'
# Project-Specific Rules

## Formatting
- NO TABS - Use 2-space indentation
- Line length: 100 chars (soft), 120 chars (hard)

## Architecture
- Follow microservices pattern
- Use event-driven communication
EOF

# Commit project files (submodule files are not committed to parent)
git add .ai-pack/PROJECT-*.md
git commit -m "Add project-specific coding rules"
```

### How AI Assistants Discover Both Tiers

**Claude Code and similar tools automatically:**
1. ✅ Read all `.md` files in `.ai-pack/` directory
2. ✅ Include both submodule files (shared standards)
3. ✅ Include project-specific files (PROJECT-*.md pattern)
4. ✅ Apply both sets of rules during code generation and review

**No additional configuration needed!** Just place your project files in `.ai-pack/` with the `PROJECT-` prefix.

### Example Directory Structure

After setup, your project's `.ai-pack/` contains both shared and project files:

```
.ai-pack/                              # Git submodule + project files
├── gates/                                # Shared (from submodule)
├── roles/                                # Shared (from submodule)
├── workflows/                            # Shared (from submodule)
├── templates/                            # Shared (from submodule)
├── quality/                              # Shared (from submodule)
│   └── clean-code/                       # Clean code standards
├── README.md                             # Shared (from submodule)
├── PROJECT-README.md                     # Project-specific (your file)
├── PROJECT-myproject.md                  # Project-specific (your file)
└── PROJECT-architecture.md               # Project-specific (your file)
```

**Git Behavior:**
- Submodule tracks: All files except `PROJECT-*`
- Parent project tracks: Only `PROJECT-*` files
- Updates to submodule don't conflict with your project files

---

## Repository Structure

```
ai-pack/
├── README.md                          # This file
├── LICENSE                            # MIT License
├── .gitignore                         # Git ignore rules
├── VERSION                            # Version information
├── PARALLEL-ENGINEERS-CONFIG.md       # Parallel execution configuration
├── GITHUB_SETUP.md                    # GitHub integration guide
│
├── gates/                             # Quality control rules
│   ├── 00-global-gates.md             # Universal rules
│   ├── 10-persistence.md              # File operations rules
│   ├── 20-tool-policy.md              # Tool usage policies
│   ├── 25-execution-strategy.md       # Execution strategy enforcement
│   ├── 30-verification.md             # Verification requirements
│   └── 35-code-quality-review.md      # Code quality review gate
│
├── roles/                             # Agent personas
│   ├── orchestrator.md                # Coordinator role
│   ├── engineer.md                    # Implementation specialist
│   ├── inspector.md                   # Bug investigation specialist
│   ├── product-manager.md             # Requirements specialist
│   ├── architect.md                   # Technical design specialist
│   ├── tester.md                      # Testing specialist
│   └── reviewer.md                    # Quality assurance
│
├── workflows/                         # Development processes
│   ├── standard.md                    # General workflow
│   ├── feature.md                     # Feature development
│   ├── bugfix.md                      # Bug fixing
│   ├── refactor.md                    # Code refactoring
│   └── research.md                    # Code investigation
│
├── templates/                         # Reusable templates
│   ├── CLAUDE.md                      # Bootstrap template
│   └── task-packet/                   # Task packet templates
│       ├── 00-contract.md             # Task definition
│       ├── 10-plan.md                 # Implementation plan
│       ├── 20-work-log.md             # Execution log
│       ├── 30-review.md               # Review findings
│       └── 40-acceptance.md           # Completion sign-off
│
└── quality/                           # Quality standards
    └── clean-code/                    # Clean code standards
        ├── 00-general-rules.md        # Universal standards
        ├── 01-design-principles.md    # Core design principles
        ├── 02-solid-principles.md     # SOLID principles
        ├── 03-refactoring.md          # Refactoring guidelines
        ├── 04-testing.md              # Testing standards
        ├── 05-architecture.md         # Architecture patterns
        ├── 06-code-review-checklist.md # Review checklist
        ├── 07-development-practices.md # Development workflow
        ├── 08-deployment-patterns.md  # Deployment strategies
        ├── 09-system-evolution.md     # System evolution
        ├── 10-api-design.md           # API design principles
        ├── 11-documentation-standards.md # Documentation standards
        ├── lang-cpp-basics.md         # C++ basics
        ├── lang-cpp-design.md         # C++ design patterns
        ├── lang-cpp-advanced.md       # C++ advanced topics
        ├── lang-cpp-modern.md         # Modern C++ features
        ├── lang-cpp-guidelines.md     # C++ Core Guidelines
        ├── lang-cpp-reference.md      # C++ quick reference
        ├── lang-csharp.md             # C# guidelines
        ├── lang-python.md             # Python guidelines
        ├── lang-javascript.md         # JavaScript/TypeScript
        ├── lang-java.md               # Java guidelines
        ├── lang-kotlin.md             # Kotlin guidelines
        ├── RULES_REFERENCE.md         # Quick reference
        └── CHANGELOG.md               # Version history
```

---

## Using AI-Pack in Your Projects

### Code Reviews

Reference specific guidelines during code reviews:

```
This violates the Single Responsibility Principle (quality/clean-code/02-solid-principles.md).
Consider extracting the database logic into a separate repository class.
```

### CI/CD Integration

Add automated checks based on these standards:

```yaml
# .github/workflows/code-quality.yml
- name: Check code compliance
  run: |
    # Run linters configured per these standards
    clang-tidy --config-file=.ai-pack/quality/clean-code/clang-tidy-config
```

### Team Onboarding

Use as onboarding material for new team members:

1. **Week 1:** Read AI Workflow Framework (gates, roles, workflows)
2. **Week 2:** Study design principles (quality/clean-code/01-06)
3. **Week 3:** Review development practices (quality/clean-code/07-09)
4. **Week 4:** Study language-specific guidelines (quality/clean-code/lang-*)

---

## Sources and Attribution

This repository synthesizes best practices from:

- **Martin Fowler** - [martinfowler.com](https://martinfowler.com)
  - Design Patterns, Refactoring, Deployment Patterns
- **Kent Beck** - Four Rules of Simple Design
- **Robert C. Martin (Uncle Bob)** - SOLID Principles
- **Scott Meyers** - *Effective C++*
- **ISO C++ Standards Committee** - C++ Core Guidelines
- **Google** - Industry best practices

---

## Contributing

To suggest improvements or additions:

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/new-guideline`)
3. Make your changes with clear documentation
4. Submit a pull request

### Contribution Guidelines

- Provide concrete code examples for each principle
- Include both "good" and "bad" examples
- Cite authoritative sources
- Keep language neutral in core principles (language-specific details go in `quality/clean-code/lang-*` files)

---

## Versioning

This repository uses [Semantic Versioning](https://semver.org/):

- **Major version** (X.0.0): Breaking changes to structure or significant rewrites
- **Minor version** (0.X.0): New guidelines or sections added
- **Patch version** (0.0.X): Clarifications, typo fixes, minor improvements

Current version: See [VERSION](VERSION) file

---

## License

MIT License - See [LICENSE](LICENSE) file for details.

Copyright (c) 2025 Cortexa LLC

---

## Related Projects

- **[Sourcerer](https://github.com/Cortexa-LLC/sourcerer)** - A project following these standards

---

## Support

For questions or discussions:

- Open an [issue](https://github.com/Cortexa-LLC/ai-pack/issues)
- Start a [discussion](https://github.com/Cortexa-LLC/ai-pack/discussions)

---

## Changelog

See [quality/clean-code/CHANGELOG.md](quality/clean-code/CHANGELOG.md) for version history.

---

*Building better software through better standards and structured AI workflows*

Copyright (c) 2025 Cortexa LLC
