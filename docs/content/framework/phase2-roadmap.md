---
sidebar_position: 4
title: "Phase 2 Roadmap"
---

# Phase 2 Roadmap

**Phase 1 Status**: ✅ COMPLETE
**Phase 2 Status**: Ready to Begin

---

## Executive Summary

Phase 1 has successfully validated the AI-Pack agent spawning infrastructure. All core capabilities are working:

- ✅ Agent spawning and configuration system
- ✅ Task tracking via Beads integration
- ✅ Full tool access for agents (file ops, web, bash, MCP)
- ✅ Multi-agent workflow coordination
- ✅ Sequential execution pattern (avoiding bug #13890)

**Total Agents Executed**: 9 (5 parallel test + 4 workflow test)
**Success Rate**: 100%
**Average Spawn Time**: 0.06s

Phase 2 will build on this foundation to enable true parallel execution, direct API control, and production-grade performance.

---

## Phase 1 Accomplishments

### Infrastructure Delivered

#### 1. Agent Configuration System
- **Location**: `.ai-pack/agents/lightweight/*.yml`
- **Agents**: engineer, tester, reviewer
- **Format**: YAML with role, tools, delegation, timeout, success criteria
- **Status**: Production-ready ✅

#### 2. Task Packet System
- **Location**: `.beads/tasks/`
- **Structure**: metadata.json, plan.md, agent-prompt.txt, results.md
- **Integration**: Beads tracking system
- **Status**: Working perfectly ✅

#### 3. Spawning Infrastructure
- **Command**: `.ai-pack/bd spawn <role> <task>`
- **Implementation**: Python-based (bd_spawn.py)
- **Performance**: 0.05-0.10s spawn time
- **Status**: Optimized ✅

#### 4. Role Definitions
- **Location**: `roles/*.md`
- **Roles**: engineer (TDD focus), tester (>80% coverage), reviewer (security & quality)
- **Format**: Markdown with instructions and guidelines
- **Status**: Comprehensive ✅

### Testing Completed

#### 1. Tool Access Verification
**Test**: `tests/test_agent_tool_access.py`
**Results**:
- File operations: PASSED
- Web access: PASSED
- Bash execution: PASSED
- Search tools (grep, glob): PASSED
- MCP access (7 servers): PASSED
- Directory operations: PASSED

**Conclusion**: 100% tool access verified ✅

#### 2. Parallel Spawn Test
**Test**: `tests/parallel_execution_test.py`
**Results**:
- 5 agents spawned successfully
- Unique task IDs generated
- No context pollution
- 0.29s total spawn time (0.06s average)

**Conclusion**: Spawn infrastructure scales ✅

#### 3. Multi-Agent Workflow
**Test**: `tests/workflow_test_user_registration.py`
**Results**:
- 4-agent workflow executed (backend, frontend, tester, reviewer)
- 57KB code generated (~1,680 lines)
- 104 tests created with ~93% coverage
- Comprehensive code review with security analysis

**Conclusion**: Multi-agent coordination works ✅

### Documentation Created

1. **Implementation Plan**: `docs/A2A-IMPLEMENTATION-PLAN.md` (900+ lines)
2. **Architecture Notes**: `docs/PHASE1-ARCHITECTURE-NOTES.md`
3. **Usage Guide**: `docs/USAGE-GUIDE.md`
4. **Protocol Handler**: `docs/PROTOCOL-HANDLER-SETUP.md`
5. **Progress Tracking**: `PHASE1-PROGRESS.md`
6. **Workflow Summary**: `tests/WORKFLOW_EXECUTION_SUMMARY.md`
7. **Tool Access Report**: `tests/agent_integration_workspace/TOOL_ACCESS_REPORT.md`

### Installation System

- **Setup Script**: `setup.py` (cross-platform)
- **Platform Support**: macOS, Linux, Windows
- **Protocol Handler**: `agent://` URL scheme support
- **Prerequisites Check**: Python 3.8+, Claude Code CLI
- **Status**: Ready for distribution ✅

---

## Phase 1 Limitations (By Design)

### Sequential Execution

**Current Behavior**:
```
Agent 1: spawn (0.06s) + execute (3min) = 3min
Agent 2: spawn (0.06s) + execute (3min) = 3min
Agent 3: spawn (0.06s) + execute (2min) = 2min

Total: ~8 minutes (sequential)
```

**Why**: Avoids Claude Code bug #13890 by using foreground execution

**Phase 2 Goal**: True parallel execution
```
Agent 1, 2, 3: spawn concurrently + execute in parallel

Total: ~3 minutes (parallel)
```

### Claude Code Dependency

**Current**: Agents execute via Claude Code Task tool
- ✅ Reliable and stable
- ✅ Full tool access
- ❌ No direct control over token usage
- ❌ Sequential execution only
- ❌ Dependent on Claude Code CLI availability

**Phase 2 Goal**: Direct Anthropic API
- ✅ Full control over API calls
- ✅ Token usage optimization (30-40% reduction target)
- ✅ Concurrent execution via goroutines
- ✅ No external CLI dependency

### No Real-Time Progress

**Current**: Results only available after completion

**Phase 2 Goal**: SSE streaming for real-time progress updates

---

## Phase 2 Requirements

### Core Objectives

1. **Enable Parallel Execution**
   - Multiple agents running concurrently
   - Independent goroutines per agent
   - Resource management and coordination
   - **Target**: 2x speedup for multi-agent workflows

2. **Direct API Integration**
   - Anthropic API client in Go
   - Token usage optimization
   - Request batching where possible
   - **Target**: 30-40% token reduction

3. **Production Readiness**
   - Error handling and recovery
   - Logging and monitoring
   - Rate limit management
   - **Target**: 99.9% uptime

4. **A2A Protocol Compliance**
   - JSON-RPC 2.0 implementation
   - Discovery endpoint
   - Task execution endpoint
   - Results aggregation

### Technical Stack

#### Go A2A Server

**Purpose**: Standalone server for agent execution

**Components**:
```
cmd/
  a2a-server/
    main.go                 # Entry point

internal/
  server/
    server.go              # HTTP server
    discovery.go           # A2A discovery
    execution.go           # Task execution

  agents/
    spawner.go             # Agent lifecycle
    executor.go            # Anthropic API integration
    config.go              # Config loading

  protocol/
    jsonrpc.go             # JSON-RPC 2.0
    a2a.go                 # A2A protocol

  tracking/
    beads.go               # Task tracking
```

**Dependencies**:
- `github.com/anthropics/anthropic-sdk-go` - Anthropic API
- `github.com/gorilla/mux` - HTTP routing
- `gopkg.in/yaml.v3` - YAML parsing
- Standard library for JSON-RPC

#### Key Features

1. **Concurrent Execution**
```go
func (s *Server) ExecuteTaskConcurrent(ctx context.Context, tasks []Task) []Result {
    results := make(chan Result, len(tasks))

    for _, task := range tasks {
        go func(t Task) {
            result := s.executeTask(ctx, t)
            results <- result
        }(task)
    }

    // Collect results
    var collected []Result
    for i := 0; i < len(tasks); i++ {
        collected = append(collected, <-results)
    }

    return collected
}
```

2. **Streaming Progress**
```go
func (s *Server) StreamTaskProgress(ctx context.Context, taskID string) <-chan ProgressUpdate {
    updates := make(chan ProgressUpdate)

    go func() {
        defer close(updates)

        stream := s.anthropic.Messages.NewStreaming(ctx, params)
        for event := range stream.Events() {
            updates <- ProgressUpdate{
                TaskID: taskID,
                Event: event,
                Timestamp: time.Now(),
            }
        }
    }()

    return updates
}
```

3. **Token Optimization**
```go
type TokenOptimizer struct {
    cache map[string]CachedContext
}

func (o *TokenOptimizer) OptimizePrompt(ctx Context, task Task) OptimizedPrompt {
    // Remove redundant context
    // Cache common role definitions
    // Compress verbose instructions
    // Return optimized prompt with 30-40% fewer tokens
}
```

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    User / Orchestrator                   │
└─────────────────────┬───────────────────────────────────┘
                      │
                      │ HTTP/JSON-RPC
                      ▼
┌─────────────────────────────────────────────────────────┐
│                 Go A2A Server (Port 8080)                │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Discovery Endpoint: /a2a/discovery              │  │
│  │  Execution Endpoint: /a2a/execute                │  │
│  │  Streaming Endpoint: /a2a/stream/:task_id        │  │
│  └──────────────────────────────────────────────────┘  │
└───┬─────────────────┬─────────────────┬───────────────┘
    │                 │                 │
    │ Concurrent      │ Concurrent      │ Concurrent
    ▼                 ▼                 ▼
┌──────────┐    ┌──────────┐    ┌──────────┐
│ Agent 1  │    │ Agent 2  │    │ Agent 3  │
│ Goroutine│    │ Goroutine│    │ Goroutine│
└────┬─────┘    └────┬─────┘    └────┬─────┘
     │               │               │
     │ Direct API    │ Direct API    │ Direct API
     ▼               ▼               ▼
┌─────────────────────────────────────────────────────────┐
│              Anthropic API (api.anthropic.com)          │
└─────────────────────────────────────────────────────────┘
```

---

## Migration Strategy

### Phase 2.0: Foundation (Weeks 1-2)

**Goals**:
- Go project structure
- Anthropic API integration
- Single agent execution

**Deliverables**:
- [ ] Go module initialization
- [ ] Anthropic SDK integration
- [ ] Config loader (reuse Phase 1 YAML)
- [ ] Single agent executor
- [ ] Basic HTTP server

**Test**: Execute one agent via Go server

### Phase 2.1: Concurrent Execution (Weeks 3-4)

**Goals**:
- Goroutine-based parallelism
- Resource management
- Error handling

**Deliverables**:
- [ ] Concurrent task executor
- [ ] Goroutine pool management
- [ ] Context cancellation
- [ ] Error aggregation

**Test**: Execute 3 agents in parallel, verify speedup

### Phase 2.2: A2A Protocol (Weeks 5-6)

**Goals**:
- Full A2A compliance
- JSON-RPC 2.0
- Discovery and execution endpoints

**Deliverables**:
- [ ] JSON-RPC server
- [ ] A2A discovery endpoint
- [ ] A2A execution endpoint
- [ ] Results aggregation

**Test**: A2A protocol compliance test suite

### Phase 2.3: Streaming & Production (Weeks 7-8)

**Goals**:
- SSE streaming
- Production hardening
- Monitoring and logging

**Deliverables**:
- [ ] SSE streaming endpoint
- [ ] Structured logging
- [ ] Metrics collection
- [ ] Health checks
- [ ] Rate limiting

**Test**: Load testing, failure recovery, streaming verification

---

## What Carries Forward from Phase 1

### 100% Reusable

1. **Agent Configurations** (`.ai-pack/agents/lightweight/*.yml`)
   - Same YAML format
   - Same role definitions
   - Same tool specifications
   - Go server will load these directly

2. **Role Files** (`roles/*.md`)
   - No changes needed
   - Go server injects into prompts
   - Same role-specific instructions

3. **Task Packet Structure** (`.beads/tasks/`)
   - Same metadata format
   - Same tracking approach
   - Go server writes to same location

4. **bd spawn CLI** (`.ai-pack/bd`)
   - Same user interface
   - Under the hood: HTTP POST to Go server instead of Claude Code Task tool
   - Zero user-facing changes

### Requires Modification

1. **bd_spawn.py** → **HTTP Client**
```python
# Phase 1
task = Task(prompt, mode="delegate")

# Phase 2
import requests
response = requests.post("http://localhost:8080/a2a/execute", json={
    "role": "engineer",
    "task": "implement feature"
})
```

2. **Execution Model**: Foreground → Background with streaming

3. **Progress Visibility**: Post-completion → Real-time via SSE

---

## Success Metrics (Phase 2)

| Metric | Phase 1 Baseline | Phase 2 Target |
|--------|------------------|----------------|
| Parallel speedup | N/A (sequential) | 2x for 2-3 agents |
| Token usage | 100% (via Claude Code) | 60-70% (30-40% reduction) |
| Spawn latency | 0.06s | &lt;0.05s |
| Max concurrent agents | 1 | 5-10 |
| API rate limit handling | N/A | Automatic backoff |
| Streaming progress | No | Yes (SSE) |
| Uptime | N/A | 99.9% |

---

## Risk Assessment

### Technical Risks

1. **Anthropic API Rate Limits**
   - **Risk**: Hitting rate limits with concurrent requests
   - **Mitigation**: Implement backoff, queue management, rate limiter
   - **Severity**: Medium

2. **Goroutine Resource Exhaustion**
   - **Risk**: Too many concurrent agents consuming memory
   - **Mitigation**: Goroutine pool with max concurrency limit
   - **Severity**: Low

3. **Streaming Reliability**
   - **Risk**: SSE connections dropping mid-stream
   - **Mitigation**: Reconnection logic, checkpoint/resume
   - **Severity**: Medium

4. **Token Optimization Effectiveness**
   - **Risk**: May not achieve 30-40% reduction
   - **Mitigation**: Incremental optimization, measure each change
   - **Severity**: Low (nice to have, not critical)

### Integration Risks

1. **Breaking bd spawn Interface**
   - **Risk**: Users have to change workflows
   - **Mitigation**: Keep exact same CLI interface, change only backend
   - **Severity**: High (avoid at all costs)

2. **Task Packet Compatibility**
   - **Risk**: Go server writes incompatible task packets
   - **Mitigation**: Use same JSON schema, validate compatibility
   - **Severity**: Medium

3. **MCP Server Access**
   - **Risk**: Go agents can't access MCP servers
   - **Mitigation**: Implement MCP client in Go or proxy through Claude Code
   - **Severity**: High (critical feature)

---

## Development Timeline

### 8-Week Plan

**Weeks 1-2**: Foundation
- Go project setup
- Anthropic API integration
- Single agent execution
- HTTP server basics

**Weeks 3-4**: Concurrency
- Goroutine-based execution
- Parallel agent spawning
- Resource management
- Error handling

**Weeks 5-6**: A2A Protocol
- JSON-RPC 2.0
- Discovery endpoint
- Execution endpoint
- Results aggregation

**Weeks 7-8**: Production
- SSE streaming
- Monitoring and logging
- Load testing
- Documentation

**Week 9**: Buffer & Polish

**Week 10**: Launch 🚀

---

## Testing Strategy

### Unit Tests (Go)

```go
func TestAgentExecutor_Execute(t *testing.T) {
    executor := NewAgentExecutor(config)
    result := executor.Execute(ctx, task)

    assert.NoError(t, result.Error)
    assert.NotEmpty(t, result.Output)
    assert.True(t, result.Success)
}

func TestConcurrentExecution(t *testing.T) {
    tasks := []Task{task1, task2, task3}
    results := server.ExecuteConcurrent(ctx, tasks)

    assert.Len(t, results, 3)
    // Verify all succeeded
    // Verify total time < sum of individual times
}
```

### Integration Tests

1. **A2A Protocol Compliance**
   - Discovery endpoint returns correct schema
   - Execution endpoint accepts JSON-RPC
   - Results match expected format

2. **Backward Compatibility**
   - Phase 1 bd spawn commands work unchanged
   - Task packets readable by Phase 1 tools
   - Configuration files compatible

3. **Performance Benchmarks**
   - Parallel speedup measurement
   - Token usage comparison
   - Latency under load

---

## Open Questions

1. **MCP Server Access from Go**
   - How will Go agents access MCP servers?
   - Options: Native Go MCP client, proxy through Claude Code, HTTP bridge
   - **Decision Needed**: Week 1 of Phase 2

2. **Agent Tool Execution**
   - How to handle file operations, bash commands in Go?
   - Options: Shell out, native Go implementations, hybrid approach
   - **Decision Needed**: Week 2 of Phase 2

3. **State Persistence**
   - Should server persist agent state for long-running tasks?
   - Options: In-memory only, SQLite, external DB
   - **Decision Needed**: Week 3 of Phase 2

4. **External Agent Integration**
   - When to enable external A2A agents?
   - Phase 2.4 or Phase 3?
   - **Decision Needed**: After Phase 2.3 complete

---

## Resources for Phase 2

### Documentation to Review

1. **Anthropic API Docs**: https://docs.anthropic.com/
2. **A2A Protocol Spec**: `docs/A2A-PROTOCOL.md`
3. **Go Best Practices**: Concurrency patterns, error handling
4. **JSON-RPC 2.0**: https://www.jsonrpc.org/specification

### Code References

1. **Phase 1 Implementation**: All files in `.ai-pack/` and `tests/`
2. **Agent Configurations**: `.ai-pack/agents/lightweight/*.yml`
3. **Task Packet Examples**: `.beads/tasks/task-*`

### Tools & Libraries

1. **Anthropic SDK**: `github.com/anthropics/anthropic-sdk-go`
2. **HTTP Router**: `github.com/gorilla/mux`
3. **YAML Parser**: `gopkg.in/yaml.v3`
4. **Testing**: Standard `testing` package, `testify` for assertions
5. **Logging**: `log/slog` (structured logging)

---

## Handoff Checklist

### Phase 1 Completion

- [x] Agent spawning infrastructure working
- [x] Task tracking via Beads integrated
- [x] Full tool access verified (100% pass)
- [x] Multi-agent workflows tested
- [x] Documentation complete
- [x] Usage guide created
- [x] Installation system created
- [x] Protocol handler documented
- [x] Test suite comprehensive

### Phase 2 Prerequisites

- [ ] Go development environment setup
- [ ] Anthropic API key configured
- [ ] A2A server repository created
- [ ] Development timeline confirmed
- [ ] Team roles assigned
- [ ] Weekly milestones defined

### Knowledge Transfer

- [x] Phase 1 architecture documented
- [x] Limitations and design decisions explained
- [x] Success metrics defined
- [x] Risk assessment complete
- [x] Migration strategy outlined
- [x] Testing approach defined

---

## Conclusion

Phase 1 has successfully validated the core AI-Pack concept:

- ✅ Agent spawning works reliably
- ✅ Task tracking is comprehensive
- ✅ Tool access is complete
- ✅ Multi-agent coordination functions correctly

The foundation is solid and production-ready within its design constraints (sequential execution, Claude Code dependency).

Phase 2 will unlock the full potential:
- 🚀 True parallel execution (2x+ speedup)
- 💰 Token optimization (30-40% reduction)
- 📊 Real-time progress streaming
- 🏗️ Production-grade infrastructure

**Phase 1 Status**: ✅ COMPLETE
**Phase 2 Status**: 🟢 READY TO BEGIN

---

**Prepared by**: AI-Pack Team
**Date**: 2026-01-23
**Version**: 1.0.0
**Next Review**: Phase 2 Week 2 (Concurrency Implementation)
