---
sidebar_position: 5
title: "A2A Usage Guide"
---

# A2A Usage Guide

**Version**: 1.0.0 (Phase 1)
**Status**: Production Ready

---

## Overview

AI-Pack Phase 1 provides a lightweight agent spawning system that enables you to delegate tasks to specialized AI agents. Each agent operates autonomously within its defined role and tool permissions.

**Key Features:**
- 🤖 3 specialized agent roles (Engineer, Tester, Reviewer)
- 📦 Automatic task tracking via Beads
- 🛠️ Full tool access (file operations, web, bash, MCP servers)
- ⚡ Fast spawn times (~0.06s average)
- 🔒 Role-based permissions and quality gates

---

## Quick Start

### Basic Usage

```bash
# Spawn an engineer to implement a feature
.ai-pack/bd spawn engineer "implement a user authentication function"

# Spawn a tester to create tests
.ai-pack/bd spawn tester "create tests for the auth function"

# Spawn a reviewer to review code
.ai-pack/bd spawn reviewer "review the authentication implementation"
```

### What Happens When You Spawn

1. **Configuration Loading**: Agent config loaded from `.ai-pack/agents/lightweight/<role>.yml`
2. **Role Context Injection**: Role-specific instructions loaded from `roles/<role>.md`
3. **Task Packet Creation**: Task tracking created in `.beads/tasks/task-<role>-<timestamp>`
4. **Task ID Return**: Unique task ID returned for tracking
5. **Manual Execution**: Orchestrator (you) executes the agent via Task tool

**Important**: In Phase 1, agents execute sequentially (one after another), not in parallel.

---

## Agent Roles

### Engineer Agent

**Purpose**: Implementation specialist focused on writing clean, tested code

**Capabilities:**
- Write, read, and edit files
- Execute bash commands
- Search codebase (grep, glob)
- Access all MCP servers

**Quality Gates:**
- TDD enforcement (tests first)
- Code quality review (clean code principles)

**Best For:**
- Feature implementation
- Bug fixes
- Refactoring
- API development

**Example:**
```bash
.ai-pack/bd spawn engineer "implement a REST API endpoint for user registration with validation"
```

### Tester Agent

**Purpose**: Testing specialist focused on comprehensive test coverage

**Capabilities:**
- Write and edit test files
- Execute test suites
- Search codebase
- Access testing tools

**Quality Gates:**
- TDD enforcement
- >80% coverage target
- Edge case verification

**Best For:**
- Unit test creation
- Integration test creation
- Test coverage improvement
- Bug reproduction tests

**Example:**
```bash
.ai-pack/bd spawn tester "create comprehensive tests for the UserRegistration class with >80% coverage"
```

### Reviewer Agent

**Purpose**: Code review specialist focused on quality and security

**Capabilities:**
- Read files
- Search codebase
- Execute linters (via bash)
- Generate review reports

**Quality Gates:**
- Code quality standards
- Security verification
- Performance review

**Best For:**
- Code review
- Security audits
- Architecture review
- Best practices enforcement

**Example:**
```bash
.ai-pack/bd spawn reviewer "review the user registration implementation for security issues"
```

---

## Common Usage Patterns

### Pattern 1: Feature Development Workflow

```bash
# Step 1: Implement the feature
.ai-pack/bd spawn engineer "implement a password reset feature with email verification"

# Step 2: Create tests
.ai-pack/bd spawn tester "create tests for the password reset feature"

# Step 3: Review the implementation
.ai-pack/bd spawn reviewer "review password reset implementation for security"
```

**Result**: Complete feature with implementation, tests, and security review.

### Pattern 2: Bug Fix Workflow

```bash
# Step 1: Engineer reproduces and fixes
.ai-pack/bd spawn engineer "fix the authentication timeout bug in src/auth.py:45"

# Step 2: Tester creates regression tests
.ai-pack/bd spawn tester "create tests to prevent the auth timeout bug from recurring"
```

**Result**: Bug fixed with regression tests.

### Pattern 3: Refactoring Workflow

```bash
# Step 1: Review current code
.ai-pack/bd spawn reviewer "analyze src/database.py for refactoring opportunities"

# Step 2: Implement refactoring
.ai-pack/bd spawn engineer "refactor database.py based on reviewer recommendations"

# Step 3: Verify tests still pass
.ai-pack/bd spawn tester "update and verify all tests after database refactoring"
```

**Result**: Clean refactoring with maintained test coverage.

### Pattern 4: Documentation Workflow

```bash
# Generate API documentation
.ai-pack/bd spawn engineer "add comprehensive docstrings to all public methods in src/api/"

# Review documentation quality
.ai-pack/bd spawn reviewer "review API documentation for completeness and clarity"
```

**Result**: Well-documented codebase.

### Pattern 5: Security Audit

```bash
# Comprehensive security review
.ai-pack/bd spawn reviewer "perform security audit of user authentication system"

# Fix identified issues
.ai-pack/bd spawn engineer "implement password hashing and fix security issues from review"

# Verify fixes
.ai-pack/bd spawn tester "create security-focused tests for authentication vulnerabilities"
```

**Result**: Hardened security with tests.

---

## Task Tracking

### Task Packet Structure

Each spawned agent creates a task packet in `.beads/tasks/task-<role>-<timestamp>/`:

```
task-engineer-20260123-131336-839360/
├── 00-metadata.json    # Task metadata (role, status, timestamps)
├── 10-plan.md          # Agent's execution plan
├── agent-prompt.txt    # Full prompt sent to agent
└── 30-results.md       # Agent's results and deliverables
```

### Viewing Task Results

```bash
# List all tasks
ls .beads/tasks/

# View task metadata
cat .beads/tasks/task-engineer-*/00-metadata.json

# View agent results
cat .beads/tasks/task-engineer-*/30-results.md
```

### Task Metadata Fields

```json
{
  "task_id": "task-engineer-20260123-131336-839360",
  "role": "engineer",
  "spawned_by": "orchestrator",
  "spawned_at": "2026-01-23T13:13:36",
  "status": "completed",
  "description": "implement user registration API",
  "config": {
    "timeout": "10min",
    "delegation": "delegate",
    "tools": ["read", "write", "edit", "bash", "grep", "glob"]
  }
}
```

---

## Best Practices

### 1. Clear Task Descriptions

**Good:**
```bash
.ai-pack/bd spawn engineer "implement UserRegistration class with validate_email, validate_password, and register_user methods. Include type hints and docstrings."
```

**Bad:**
```bash
.ai-pack/bd spawn engineer "do registration"
```

### 2. Appropriate Role Selection

**Choose Engineer For:**
- Writing new code
- Fixing bugs
- Implementing features
- Refactoring

**Choose Tester For:**
- Creating test suites
- Improving coverage
- Writing test cases

**Choose Reviewer For:**
- Code quality assessment
- Security audits
- Architecture review
- Best practices verification

### 3. Sequential Workflows

In Phase 1, agents run sequentially. Structure your workflow accordingly:

```bash
# Good: Clear sequence
.ai-pack/bd spawn engineer "task A"
# Wait for completion, then:
.ai-pack/bd spawn engineer "task B that depends on task A"

# Phase 1 Note: Both agents will run one after another
# Phase 2 will enable true parallel execution
```

### 4. Scope Control

Keep tasks focused and well-scoped:

**Good:**
```bash
.ai-pack/bd spawn engineer "add email validation to UserRegistration.register_user method"
```

**Too Broad:**
```bash
.ai-pack/bd spawn engineer "build entire user management system"
```

### 5. Quality Gates

Trust the quality gates - they ensure:
- Engineers follow TDD practices
- Testers achieve >80% coverage
- Reviewers check security and best practices

---

## Configuration

### Agent Configuration Files

Located in `.ai-pack/agents/lightweight/`:

```yaml
# engineer.yml
name: engineer
tier: lightweight
delegation:
  mode: delegate        # Autonomous execution
  timeout: 10min        # Max execution time
tools:
  - read               # File reading
  - write              # File creation
  - edit               # File editing
  - bash               # Command execution
  - grep               # Content search
  - glob               # File search
context:
  role_file: roles/engineer.md
  gates:
    - tdd-enforcement
    - code-quality-review
success_criteria:
  - Clean, working implementation
  - Proper error handling
  - Type hints included
  - Docstrings complete
```

### Customizing Agents

To modify agent behavior, edit the YAML configuration:

1. **Adjust Timeout**: Change `timeout` field (format: `10min`, `1h`)
2. **Modify Tools**: Add/remove from `tools` list
3. **Update Quality Gates**: Modify `gates` list
4. **Change Success Criteria**: Update `success_criteria` list

**Example - Add Web Access:**
```yaml
tools:
  - read
  - write
  - edit
  - bash
  - grep
  - glob
  - webfetch  # Added web access
```

---

## Troubleshooting

### Issue: Task ID Not Generated

**Symptom**: No task ID returned after spawn

**Solution**: Check bd_spawn.py output for errors
```bash
python3 .ai-pack/bd_spawn.py engineer "test task"
```

### Issue: Agent Lacks Tool Access

**Symptom**: Agent reports missing tool permissions

**Solution**: Verify tool is listed in agent's YAML config
```bash
cat .ai-pack/agents/lightweight/engineer.yml | grep -A 10 "tools:"
```

### Issue: Task Packet Not Created

**Symptom**: No directory in `.beads/tasks/`

**Solution**: Ensure `.beads/` directory exists and has write permissions
```bash
mkdir -p .beads/tasks
chmod 755 .beads/tasks
```

### Issue: Slow Spawn Times

**Symptom**: Spawn takes >1 second

**Solution**: Check for:
- Large role files (should be less than 50KB)
- Complex YAML parsing
- File system performance

**Normal**: 0.05-0.10s spawn time

---

## Advanced Usage

### Multi-Agent Workflows

For complex features, chain multiple agents:

```bash
# Complete feature development
.ai-pack/bd spawn engineer "implement backend API"
.ai-pack/bd spawn engineer "implement frontend form"
.ai-pack/bd spawn tester "create test suite"
.ai-pack/bd spawn reviewer "review all implementations"
```

See `tests/workflow_test_user_registration.py` for a complete example.

### Custom Task Descriptions

Include specific requirements:

```bash
.ai-pack/bd spawn engineer "implement UserAuth class:
- validate_credentials(username, password) -> bool
- hash_password(password) -> str (use bcrypt)
- create_session(user_id) -> str (return JWT token)
Include type hints, docstrings, and error handling"
```

### Accessing MCP Servers

Agents have access to 7 MCP servers:

1. **GitHub** (git-server) - Repository operations
2. **Airtable** (airtable-server) - Database operations
3. **JIRA** (jira-server) - Issue tracking
4. **Confluence** (wiki-server) - Documentation
5. **Jenkins** (jenkins-server) - CI/CD
6. **MarkItDown** (markitdown-server) - Document conversion
7. **PowerPoint** (pptx-server) - Presentation generation

**Example:**
```bash
.ai-pack/bd spawn engineer "create JIRA ticket for bug found in auth.py using MCP"
```

---

## Performance Characteristics

### Spawn Performance

| Metric | Average | Range |
|--------|---------|-------|
| Spawn Time | 0.06s | 0.05-0.10s |
| Config Load | 0.01s | 0.01-0.02s |
| Task Packet Creation | 0.02s | 0.01-0.03s |
| Total Overhead | 0.09s | 0.07-0.15s |

### Agent Execution (Sequential)

```
Agent 1: spawn (0.06s) + execute (3min) = ~3min
Agent 2: spawn (0.06s) + execute (3min) = ~3min
Agent 3: spawn (0.06s) + execute (2min) = ~2min

Total: ~8 minutes (sequential)
```

**Phase 2 Note**: Parallel execution will reduce total time to ~3 minutes.

---

## Limitations (Phase 1)

### Sequential Execution

Agents run one after another, not concurrently:
- Spawn overhead is minimal (~0.06s)
- Execution time is additive
- No parallel performance gains

**Phase 2** will enable true concurrent execution via Go A2A server.

### Foreground Execution

Agents execute in foreground (synchronous):
- Avoids Claude Code bug #13890
- Stable and reliable
- Blocks until completion

**Phase 2** will use background execution with direct Anthropic API.

### Tool Access

Limited to tools defined in YAML config:
- Cannot dynamically add tools
- Permissions set at spawn time
- MCP access controlled by orchestrator

---

## Migration to Phase 2

Phase 1 infrastructure will carry forward to Phase 2:
- ✅ Agent configurations (compatible)
- ✅ Task packet structure (unchanged)
- ✅ bd spawn CLI (same interface)
- ✅ Beads integration (maintained)

**Changes in Phase 2:**
- Parallel execution (concurrent agents)
- Direct Anthropic API (better token efficiency)
- SSE streaming (real-time progress)
- Background execution (long-running tasks)

---

## Examples Repository

Complete working examples in `tests/`:

1. **Single Agent**: `tests/run_agent_integration_tests.py`
2. **Parallel Spawn**: `tests/parallel_execution_test.py`
3. **Multi-Agent Workflow**: `tests/workflow_test_user_registration.py`

---

## Support and Resources

**Documentation:**
- Architecture Plan: `docs/A2A-IMPLEMENTATION-PLAN.md`
- Phase 1 Notes: `docs/PHASE1-ARCHITECTURE-NOTES.md`
- Progress Tracking: `PHASE1-PROGRESS.md`

**Test Results:**
- Tool Access Report: `tests/agent_integration_workspace/TOOL_ACCESS_REPORT.md`
- Workflow Summary: `tests/WORKFLOW_EXECUTION_SUMMARY.md`

**Configuration:**
- Agent Configs: `.ai-pack/agents/lightweight/*.yml`
- Role Definitions: `roles/*.md`

---

## Quick Reference

```bash
# Spawn agents
.ai-pack/bd spawn engineer "task description"
.ai-pack/bd spawn tester "task description"
.ai-pack/bd spawn reviewer "task description"

# View tasks
ls .beads/tasks/

# Check results
cat .beads/tasks/task-*/30-results.md

# View metadata
cat .beads/tasks/task-*/00-metadata.json
```

---

**Version**: 1.0.0 (Phase 1)
**Last Updated**: 2026-01-23
**Status**: Production Ready ✅
