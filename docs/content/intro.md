---
sidebar_position: 1
slug: /
---

# AI-Pack Documentation

Welcome to **AI-Pack**, a comprehensive AI agent workflow framework for software development.

## What is AI-Pack?

AI-Pack provides structured processes, quality gates, agent roles, and coding standards for AI agent-based software development. It ensures quality, consistency, and proper governance throughout the development lifecycle.

## Key Features

### 🚦 Quality Gates
Enforcement rules that govern permitted actions:
- **Global Gates** - Universal safety and quality rules
- **TDD Enforcement** - Mandatory test-driven development
- **Persistence Rules** - File operations and state management
- **Code Quality Review** - Mandatory validation gates
- **Architectural Review** - System change oversight

### 👥 Agent Roles
Specialized agent personas with specific responsibilities:
- **Orchestrator** - High-level coordination and delegation
- **Engineer** - Implementation specialist following TDD
- **Inspector** - Bug investigation and root cause analysis
- **Strategist** - Market analysis and business strategy
- **Product Manager** - Product requirements and user stories
- **Designer** - UX workflows and wireframes
- **Architect** - Technical design and system architecture
- **Tester** - Testing validation and TDD compliance
- **Reviewer** - Code quality and standards compliance

### 🔄 Development Workflows
Structured processes for different types of work:
- Feature development
- Bug fixing
- Code refactoring
- Research and investigation

### 🤖 Agent-to-Agent (A2A) Workflow
Production-grade agent spawning system for autonomous task delegation:
- Specialized agent roles (Engineer, Tester, Reviewer)
- Fast spawn times (~0.06s average)
- Automatic task tracking via Beads
- Full tool access with quality gates
- Protocol handler support (`agent://` URLs)
- **Parallel execution** via Go-based A2A server ✅
- **Real-time streaming** with SSE progress updates ✅
- **Production infrastructure** with structured logging and metrics ✅

### 📋 Task Memory System
Persistent, git-backed task tracking using **[Beads](https://github.com/steveyegge/beads)**:
- Cross-session memory that survives AI conversation boundaries
- Git-backed storage in `.beads/issues.jsonl`
- Full dependency tracking and task graphs
- Multi-agent coordination with hash-based IDs

### 🎯 Clean Code Standards
Industry-leading coding principles and practices:
- Universal design principles (SOLID, DRY, YAGNI)
- Language-specific guidelines (C++, Python, JavaScript/TypeScript, Java, Kotlin, Swift)
- Testing best practices and TDD workflow
- Architecture patterns and refactoring techniques

## Quick Start

### 1. Add Framework to Your Project

```bash
# Add ai-pack as submodule
cd your-project
git submodule add https://github.com/Cortexa-LLC/ai-pack .ai-pack
git submodule update --init --recursive

# Create local workspace
mkdir -p .ai/tasks

# Initialize Beads for task tracking
bd init

# Copy bootstrap template
cp .ai-pack/templates/CLAUDE.md ./CLAUDE.md

# Commit framework setup
git add .ai-pack .beads/issues.jsonl CLAUDE.md
git commit -m "Add ai-pack framework and initialize Beads"
```text

### 2. Claude Code Integration

For native Claude Code integration with slash commands, skills, and hooks:

```bash
# Run automated setup
python3 .ai-pack/templates/.claude-setup.py

# Configure permissions for background agents
# See docs/CLAUDE-CODE-CONFIGURATION.md for details
```text

### 3. Start Using Tasks

```bash
# Create a task in Beads
agent create "Implement user authentication" --priority high

# View available tasks
agent list --status queued

# Start working on a task
bd start bd-a1b2

# Mark task complete
agent close bd-a1b2
```text

## Documentation Structure

- **[Getting Started](/docs/getting-started)** - Installation and setup guide
- **[Agent-to-Agent (A2A)](/docs/framework/agent-to-agent)** - Multi-agent workflow system
- **[Workflows](/docs/workflows/overview)** - Development process guides
- **[Roles](/docs/roles/overview)** - Agent persona specifications
- **[Quality Gates](/docs/gates/overview)** - Enforcement rules and controls
- **[Clean Code](/docs/quality/clean-code)** - Coding standards and best practices
- **[Task Memory](/docs/beads)** - Beads integration and task tracking

## Why AI-Pack?

**Structured AI Development:**
- Clear processes and workflows
- Quality gates that prevent mistakes
- Consistent patterns across projects

**Cross-Session Memory:**
- Tasks persist across AI conversations
- Git-backed task database
- Team coordination and dependency tracking

**Quality Governance:**
- Mandatory code review and testing
- TDD enforcement with blocking gates
- Architecture oversight for significant changes

**Knowledge Capture:**
- Complete history in task packets
- Decisions and rationale preserved
- Long-term documentation in `docs/`

## Support

For questions or discussions:
- [GitHub Issues](https://github.com/Cortexa-LLC/ai-pack/issues)
- [GitHub Discussions](https://github.com/Cortexa-LLC/ai-pack/discussions)

## License

MIT License - Copyright (c) 2025 Cortexa LLC

---

**Ready to get started?** Head to the [Getting Started](/docs/getting-started) guide.
