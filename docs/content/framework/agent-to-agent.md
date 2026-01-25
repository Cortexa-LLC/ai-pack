---
sidebar_position: 3
---

# Agent-to-Agent (A2A) Workflow

AI-Pack implements a lightweight agent spawning system that enables autonomous task delegation and multi-agent coordination. This system allows you to spawn specialized AI agents that work independently while maintaining quality gates and task tracking.

## Overview

The A2A workflow enables:
- 🤖 **Specialized agent roles** (Engineer, Tester, Reviewer)
- 📦 **Automatic task tracking** via Beads
- 🛠️ **Full tool access** (file operations, web, bash, MCP servers)
- ⚡ **Fast spawn times** (~0.06s average)
- 🔒 **Role-based permissions** and quality gates
- 🚀 **Parallel execution** via Go-based A2A server (Phase 2)
- 📊 **Real-time streaming** with SSE progress updates (Phase 2)

## Quick Start

### Basic Agent Spawning

```bash
# Spawn an engineer to implement a feature
.ai-pack/bd spawn engineer "implement a user authentication function"

# Spawn a tester to create tests
.ai-pack/bd spawn tester "create tests for the auth function"

# Spawn a reviewer to review code
.ai-pack/bd spawn reviewer "review the authentication implementation"
```

### Multi-Agent Workflow Example

```bash
# Complete feature development workflow
.ai-pack/bd spawn engineer "implement user registration API"
.ai-pack/bd spawn tester "create comprehensive test suite for user registration"
.ai-pack/bd spawn reviewer "review user registration implementation for security"
```

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

## Common Workflow Patterns

### Pattern 1: Feature Development

```bash
# Step 1: Implement the feature
.ai-pack/bd spawn engineer "implement password reset with email verification"

# Step 2: Create tests
.ai-pack/bd spawn tester "create tests for password reset feature"

# Step 3: Review the implementation
.ai-pack/bd spawn reviewer "review password reset for security issues"
```

**Result**: Complete feature with implementation, tests, and security review.

### Pattern 2: Bug Fix

```bash
# Step 1: Engineer reproduces and fixes
.ai-pack/bd spawn engineer "fix authentication timeout bug in src/auth.py:45"

# Step 2: Tester creates regression tests
.ai-pack/bd spawn tester "create tests to prevent auth timeout bug from recurring"
```

**Result**: Bug fixed with regression tests.

### Pattern 3: Refactoring

```bash
# Step 1: Review current code
.ai-pack/bd spawn reviewer "analyze src/database.py for refactoring opportunities"

# Step 2: Implement refactoring
.ai-pack/bd spawn engineer "refactor database.py based on reviewer recommendations"

# Step 3: Verify tests still pass
.ai-pack/bd spawn tester "update and verify all tests after database refactoring"
```

**Result**: Clean refactoring with maintained test coverage.

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

## Agent Configuration

### Configuration Files

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

## Protocol Handler Integration

Enable `agent://` URL scheme for spawning agents from browsers and other applications.

### URL Format

```
agent://<role>/<task-description>

Examples:
agent://engineer/implement%20REST%20API
agent://tester/create%20unit%20tests
agent://reviewer/security%20audit
```

### Browser Integration

```html
<a href="agent://engineer/implement%20login%20API">
  Spawn Engineer: Implement Login API
</a>
```

See the [Protocol Handler Setup Guide](../PROTOCOL-HANDLER-SETUP) for complete installation instructions.

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

- **Engineer**: Writing new code, fixing bugs, implementing features, refactoring
- **Tester**: Creating test suites, improving coverage, writing test cases
- **Reviewer**: Code quality assessment, security audits, architecture review

### 3. Sequential Workflows

In Phase 1, agents run sequentially. Structure your workflow accordingly:

```bash
# Good: Clear sequence
.ai-pack/bd spawn engineer "task A"
# Wait for completion, then:
.ai-pack/bd spawn engineer "task B that depends on task A"
```

**Phase 2 Note**: Future versions will enable true parallel execution.

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

**Phase 2**: Parallel execution will reduce total time to ~3 minutes.

## Phase 1: Sequential Execution (Legacy)

The initial implementation used sequential execution:
- Agents run one after another (not concurrently)
- Spawn overhead minimal (~0.06s)
- Execution time is additive
- Stable and reliable

## Phase 2: Production Features (Current) ✅

Phase 2 has delivered advanced capabilities:

### ✅ Parallel Execution
- Multiple agents running concurrently
- 2x+ speedup for multi-agent workflows
- Independent goroutines per agent
- Configurable concurrency limits

### ✅ Direct API Integration
- Anthropic API client in Go
- Better control over API calls
- Proxy support for enterprise environments

### ✅ Real-Time Streaming
- SSE streaming for progress updates
- Live agent status monitoring
- Background execution support
- Task status endpoints

### ✅ A2A Protocol Compliance
- JSON-RPC 2.0 implementation
- Discovery endpoint (`/a2a/discovery`)
- Task execution endpoint (`/a2a/execute`)
- Results aggregation
- Status monitoring (`/a2a/status`)

### ✅ Production Infrastructure
- Structured logging (JSON format)
- Performance metrics collection
- Health check endpoints
- Rate limiting support

## Resources

**Documentation:**
- [Usage Guide](../USAGE-GUIDE) - Complete A2A usage guide
- [Phase 2 Handoff](../PHASE2-HANDOFF) - Future roadmap
- [Protocol Handler Setup](../PROTOCOL-HANDLER-SETUP) - URL scheme integration

**Test Examples:**
- Single Agent: `tests/run_agent_integration_tests.py`
- Parallel Spawn: `tests/parallel_execution_test.py`
- Multi-Agent Workflow: `tests/workflow_test_user_registration.py`

**Configuration:**
- Agent Configs: `.ai-pack/agents/lightweight/*.yml`
- Role Definitions: `roles/*.md`

## Support

For questions or issues with the A2A workflow:
- [GitHub Issues](https://github.com/Cortexa-LLC/ai-pack/issues)
- [GitHub Discussions](https://github.com/Cortexa-LLC/ai-pack/discussions)

---

**Status**: Phase 2 Production Ready ✅
**Version**: 2.0.0
**Last Updated**: 2026-01-24

**Phase 1**: ✅ Sequential execution, task tracking, tool access
**Phase 2**: ✅ Parallel execution, A2A protocol, SSE streaming, production infrastructure
