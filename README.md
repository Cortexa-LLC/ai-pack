# Clean Code Standards

**Comprehensive coding standards and best practices for software development**

A curated collection of clean code principles, design patterns, and language-specific guidelines based on industry-leading sources including Martin Fowler's design principles, SOLID principles, and modern language-specific best practices.

## Overview

This repository serves as a centralized, reusable set of coding standards AND a comprehensive AI agent workflow framework. It provides both:

1. **Clean Code Standards** - Industry-leading coding principles and practices
2. **AI Workflow Framework** - Structured processes for AI agent-based development

These standards are designed to work seamlessly with AI coding assistants like Claude Code through `.ai-pack` integration.

---

## AI Workflow Framework

The AI Workflow Framework provides structured processes, roles, and templates for AI agent-based software development. It ensures quality, consistency, and proper governance throughout the development lifecycle.

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
- **[architect.md](roles/architect.md)** - Technical design specialist, system architecture and design
  - **Designs:** System architecture, API specifications, data models, technology choices
  - **Collaborates:** Works with Product Manager on feasibility, Engineers on implementation
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
│   ├── tasks/                       # Active task packets
│   │   └── 2026-01-07_feature-x/   # Example task
│   │       ├── 00-contract.md      # From template
│   │       ├── 10-plan.md          # From template
│   │       ├── 20-work-log.md      # From template
│   │       ├── 30-review.md        # From template
│   │       └── 40-acceptance.md    # From template
│   └── repo-overrides.md           # Optional project-specific deltas
│
└── CLAUDE.md                        # Bootstrap instructions for AI
```

**Key Concepts:**
- **`.ai-pack/`** - Git submodule containing shared standards and framework (this repository) - READ-ONLY
- **`.ai/`** - Local workspace in your project for task state and overrides - PROJECT-SPECIFIC
- **`CLAUDE.md`** - Bootstrap instructions at project root (copy from `templates/CLAUDE.md`)
- **Task packets** - Instances of templates created in `.ai/tasks/` for each task
- **Repo overrides** - Project-specific customizations to shared standards

**Critical Invariants:**
- ✅ Task packets go in `.ai/tasks/` (never in `.ai-pack/`)
- ✅ `.ai-pack/` is read-only shared framework
- ✅ `.ai/tasks/` preserved during framework updates
- ✅ Framework improvements happen in ai-pack repo (not ad hoc in projects)

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

### Integration with Clean Code Standards

The AI Workflow Framework and Clean Code Standards work together:

- **Gates** enforce the standards defined in `quality/clean-code/`
- **Workflows** reference standards for implementation guidance
- **Reviewer role** validates compliance with standards
- **Task packets** document adherence to standards

**Navigation:**
- Clean code standards: `quality/clean-code/` directory
- Workflow framework: `gates/`, `roles/`, `workflows/`, `templates/` directories
- Quick access to standards: `quality/engineering-standards.md`

---

## Contents

### Universal Standards

- **[00-general-rules.md](00-general-rules.md)** - Universal rules: No tabs (spaces only), TDD workflow, all tests must pass (zero tolerance for failures), 80-90% code coverage target

### Core Design Principles

- **[01-design-principles.md](01-design-principles.md)** - Beck's Four Rules of Simple Design, Tell Don't Ask, Dependency Injection, and Seams for Testability
- **[02-solid-principles.md](02-solid-principles.md)** - Complete SOLID principles with practical examples and code smells
- **[03-refactoring.md](03-refactoring.md)** - Code smells catalog and refactoring techniques
- **[04-testing.md](04-testing.md)** - Test Pyramid, test doubles, and testing best practices
- **[05-architecture.md](05-architecture.md)** - Bounded Contexts and architectural patterns
- **[06-code-review-checklist.md](06-code-review-checklist.md)** - Comprehensive code review guidelines

### Development Practices

- **[07-development-practices.md](07-development-practices.md)** - YAGNI, Frequency Reduces Difficulty, Continuous Integration, Technical Debt, Refactoring
- **[08-deployment-patterns.md](08-deployment-patterns.md)** - Feature Toggles, Blue-Green Deployment, Canary Release, Parallel Change
- **[09-system-evolution.md](09-system-evolution.md)** - Strangler Fig, Sacrificial Architecture, MonolithFirst, Semantic Diffusion

### API and Interface Design

- **[10-api-design.md](10-api-design.md)** - Command Query Separation, Naming conventions, API design principles

### Language-Specific Guidelines

Each language follows its community's established indentation standards:

| Language | Indentation | File |
|----------|-------------|------|
| C++ | 2 spaces | [lang-cpp.md](lang-cpp.md) |
| Python | 4 spaces | [lang-python.md](lang-python.md) |
| JavaScript/TypeScript | 2 spaces | [lang-javascript.md](lang-javascript.md) |
| Java | 4 spaces | [lang-java.md](lang-java.md) |
| Kotlin | 4 spaces | [lang-kotlin.md](lang-kotlin.md) |

**[lang-cpp.md](lang-cpp.md)** - Comprehensive C++ guidelines:
- All 55 items from Scott Meyers' *Effective C++*
- C++ Core Guidelines (P, F, I, C, R, ES, E, CP, Enum, Con, T, Per, SF, SL sections)
- Modern C++17/20 best practices
- 2-space indentation (Google C++ Style Guide)

**[lang-python.md](lang-python.md)** - Python best practices:
- PEP 8, PEP 20 (Zen of Python)
- Type hints and modern Python features
- 4-space indentation (PEP 8 mandatory)

**[lang-javascript.md](lang-javascript.md)** - JavaScript/TypeScript:
- Microsoft TypeScript Coding Guidelines
- Double quotes, prefer undefined over null, no I prefix
- 2-space indentation (JavaScript ecosystem standard)

**[lang-java.md](lang-java.md)** - Java guidelines:
- Oracle Java Code Conventions
- Effective Java (Joshua Bloch)
- Spring Framework patterns
- 4-space indentation (Oracle standard)

**[lang-kotlin.md](lang-kotlin.md)** - Kotlin conventions:
- Kotlin Coding Conventions (JetBrains)
- Coroutines and Flow patterns
- Android best practices
- 4-space indentation (Kotlin standard)

## Usage

### Option 1: Git Submodule (Recommended for Teams)

Add these standards to your project as a submodule:

```bash
cd your-project
git submodule add https://github.com/Cortexa-LLC/clean-code-standards .ai-pack
git submodule update --init --recursive
```

Update standards in your project:

```bash
git submodule update --remote
git add .ai-pack
git commit -m "Update clean code standards"
```

### Option 2: Symbolic Link (For Local Development)

Create a symbolic link to a single shared copy:

```bash
cd your-project
ln -s /path/to/clean-code-standards/.ai-pack .ai-pack
```

### Option 3: Direct Copy (For Standalone Projects)

Copy the standards directly into your project:

```bash
cp -r /path/to/clean-code-standards/.ai-pack your-project/.ai-pack
```

## Integration with AI Assistants

These standards are designed to work with Claude Code and other AI assistants that support `.ai-pack`:

1. Add this repository as a submodule to `.ai-pack/` in your project
2. The rule files will be automatically discovered by Claude Code
3. AI assistants will apply these standards during code generation and review

## Project-Specific Customization

### Two-Tier Rule System

This repository supports a **two-tier approach** for managing shared and project-specific rules:

**Tier 1: Shared Standards** (from this submodule)
- Core design principles
- SOLID principles
- Language-specific guidelines
- Universal best practices

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
git submodule add git@github.com:Cortexa-LLC/clean-code.git .ai-pack

# Add project-specific rules to the same directory
cat > .ai-pack/PROJECT-README.md << 'EOF'
# Sourcerer Coding Standards

This project uses a **two-tier rule system**:

## Tier 1: Shared Standards (Submodule)
All files without `PROJECT-` prefix come from the Cortexa clean-code standards.

## Tier 2: Project-Specific Rules
- `PROJECT-sourcerer.md` - Sourcerer-specific conventions
- `PROJECT-architecture.md` - Plugin architecture rules

**Both tiers are automatically discovered by Claude Code and other AI assistants.**
EOF

# Add your project rules
cat > .ai-pack/PROJECT-sourcerer.md << 'EOF'
# Sourcerer-Specific Rules

## Formatting
- NO TABS - Use 2-space indentation
- Line length: 100 chars (soft), 120 chars (hard)

## Architecture
- CodeAnalyzer must be CPU-agnostic
- Use plugin pattern for extensibility
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
├── 00-general-rules.md                  # Shared (from submodule)
├── 01-design-principles.md              # Shared (from submodule)
├── 02-solid-principles.md               # Shared (from submodule)
├── lang-cpp-basics.md                   # Shared (from submodule)
├── RULES_REFERENCE.md                   # Shared (from submodule)
├── PROJECT-README.md                    # Project-specific (your file)
├── PROJECT-sourcerer.md                 # Project-specific (your file)
└── PROJECT-architecture.md              # Project-specific (your file)
```

**Git Behavior:**
- Submodule tracks: All files except `PROJECT-*`
- Parent project tracks: Only `PROJECT-*` files
- Updates to submodule don't conflict with your project files

## Sources and Attribution

This repository synthesizes best practices from:

- **Martin Fowler** - [martinfowler.com](https://martinfowler.com)
  - Design Patterns, Refactoring, Deployment Patterns
- **Kent Beck** - Four Rules of Simple Design
- **Robert C. Martin (Uncle Bob)** - SOLID Principles
- **Scott Meyers** - *Effective C++*
- **ISO C++ Standards Committee** - C++ Core Guidelines
- **Google** - Industry best practices

## Using These Standards in Your Projects

### Code Reviews

Reference specific guidelines during code reviews:

```
This violates the Single Responsibility Principle (02-solid-principles.md).
Consider extracting the database logic into a separate repository class.
```

### CI/CD Integration

Add automated checks based on these standards:

```yaml
# .github/workflows/code-quality.yml
- name: Check code compliance
  run: |
    # Run linters configured per these standards
    clang-tidy --config-file=.ai-pack/clang-tidy-config
```

### Team Onboarding

Use as onboarding material for new team members:

1. **Week 1:** Read design principles (01-06)
2. **Week 2:** Study development practices (07-09)
3. **Week 3:** Review language-specific guidelines (lang-*)

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
- Keep language neutral in core principles (language-specific details go in `lang-*` files)

## Repository Structure

```
clean-code-standards/
├── README.md                          # This file
├── LICENSE                            # MIT License
├── .gitignore                         # Git ignore rules
├── RULES_REFERENCE.md                 # Quick reference guide
├── 00-general-rules.md                # Universal standards
├── 01-design-principles.md            # Core design principles
├── 02-solid-principles.md             # SOLID principles
├── 03-refactoring.md                  # Refactoring guidelines
├── 04-testing.md                      # Testing standards
├── 05-architecture.md                 # Architecture patterns
├── 06-code-review-checklist.md        # Review checklist
├── 07-development-practices.md        # Development workflow
├── 08-deployment-patterns.md          # Deployment strategies
├── 09-system-evolution.md             # System evolution
├── 10-api-design.md                   # API design principles
├── lang-cpp-basics.md                 # C++ basics
├── lang-cpp-design.md                 # C++ design patterns
├── lang-cpp-advanced.md               # C++ advanced topics
├── lang-cpp-modern.md                 # Modern C++ features
├── lang-cpp-guidelines.md             # C++ Core Guidelines
├── lang-cpp-reference.md              # C++ quick reference
├── lang-python.md                     # Python guidelines
├── lang-javascript.md                 # JavaScript/TypeScript
├── lang-java.md                       # Java guidelines
└── lang-kotlin.md                     # Kotlin guidelines
```

## Versioning

This repository uses [Semantic Versioning](https://semver.org/):

- **Major version** (X.0.0): Breaking changes to structure or significant rewrites
- **Minor version** (0.X.0): New guidelines or sections added
- **Patch version** (0.0.X): Clarifications, typo fixes, minor improvements

Current version: **1.0.0**

## License

MIT License - See [LICENSE](LICENSE) file for details.

Copyright (c) 2025 Cortexa LLC

## Related Projects

- **[Sourcerer](https://github.com/Cortexa-LLC/sourcerer)** - A project following these standards

## Support

For questions or discussions:

- Open an [issue](https://github.com/Cortexa-LLC/clean-code-standards/issues)
- Start a [discussion](https://github.com/Cortexa-LLC/clean-code-standards/discussions)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history.

---

**Made with ❤️ by Cortexa LLC**

*Building better software through better standards*
