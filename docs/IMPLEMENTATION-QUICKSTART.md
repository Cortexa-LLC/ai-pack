# A2A Implementation Quick Start Guide

**Status**: Ready to Begin Phase 1
**Last Updated**: 2026-01-23

## Prerequisites

### Required
- Claude Code CLI installed and configured
- Anthropic API key set in environment
- Git repository access
- task management system (existing)

### Phase 2 Additional Requirements
- Go 1.21+ (for A2A server development)
- Python 3.9+ (for script execution)
- Node.js 18+ (for JavaScript scripts)

## Project Status

✅ **Planning Complete**
- Comprehensive implementation plan documented
- Architecture decisions finalized
- Security model defined
- 10-week roadmap established

📋 **Ready to Start**
- Phase 1: Lightweight Agents
- Target: Weeks 1-2 (2-3 weeks)

## Phase 1: Lightweight Agents

### Week 1: Setup and Configuration

#### Day 1-2: Agent Configuration
**Goal**: Create lightweight agent configurations

**Tasks**:
1. Create agent configuration templates:
   ```bash
   # Create engineer agent config
   .ai-pack/agents/lightweight/engineer.yml

   # Create reviewer agent config
   .ai-pack/agents/lightweight/reviewer.yml

   # Create tester agent config
   .ai-pack/agents/lightweight/tester.yml
   ```

2. Define agent delegation modes and timeouts

3. Map roles to existing `/roles/*.md` files

**Acceptance Criteria**:
- [ ] 3 agent configs created (engineer, reviewer, tester)
- [ ] Each config maps to role file
- [ ] Delegation modes specified
- [ ] Timeout limits defined

#### Day 3-4: Beads Integration
**Goal**: Enhance task packet tracking for agents

**Tasks**:
1. Update task packet template to include agent metadata
2. Create parent-child task relationships
3. Add agent execution tracking

**Acceptance Criteria**:
- [ ] Task packets track agent lineage
- [ ] Parent-child relationships recorded
- [ ] Agent execution logged

#### Day 5: Orchestrator Enhancement
**Goal**: Enable agent spawning from orchestrator

**Tasks**:
1. Create `bd spawn` command wrapper
2. Implement role loading logic
3. Add task packet creation
4. Test single agent spawn

**Acceptance Criteria**:
- [ ] `bd spawn engineer "task"` works
- [ ] Role context loaded from file
- [ ] Task packet created in `.beads/`
- [ ] Agent executes and returns results

### Week 2: Testing and Validation

#### Day 6-7: Parallel Execution
**Goal**: Enable multiple concurrent agents

**Tasks**:
1. Test spawning 2 agents in parallel
2. Verify independent execution
3. Test result aggregation
4. Measure performance

**Test Scenario**:
```bash
# Spawn engineer and tester in parallel
bd spawn engineer "implement login validation"
bd spawn tester "test login validation"
```

**Acceptance Criteria**:
- [ ] 2-3 agents run concurrently
- [ ] No context pollution between agents
- [ ] Results return independently
- [ ] < 5s spawn latency

#### Day 8-9: Real Workflow Testing
**Goal**: Execute complete feature workflow

**Test Scenario**: User registration feature
- Architect: Design approach (or manual for Phase 1)
- Engineer: Implement backend
- Engineer: Implement frontend
- Tester: Write tests
- Reviewer: Code review

**Acceptance Criteria**:
- [ ] Complete workflow executes successfully
- [ ] Task packets track full lineage
- [ ] Parallel agents complete faster than sequential
- [ ] All artifacts properly returned

#### Day 10: Documentation and Handoff
**Goal**: Document Phase 1 results

**Tasks**:
1. Document Phase 1 results
2. Create usage examples
3. Measure success metrics
4. Plan Phase 2 kickoff

**Deliverables**:
- [ ] Phase 1 completion report
- [ ] Success metrics documented
- [ ] Examples and templates created
- [ ] Phase 2 readiness checklist

## Phase 2: A2A Server (Weeks 3-10)

### Week 3-4: Go Server Foundation

#### Setup Development Environment
```bash
# Create Go project
mkdir -p a2a-server
cd a2a-server
go mod init github.com/yourorg/ai-pack-a2a-server

# Install dependencies
go get github.com/anthropics/anthropic-sdk-go/v2
go get github.com/a2aproject/a2a-go
```

#### Initial Structure
```
a2a-server/
├── cmd/a2a-server/main.go
├── internal/
│   ├── agent/runtime.go
│   ├── protocol/jsonrpc.go
│   └── tools/registry.go
├── go.mod
└── go.sum
```

#### Tasks
- [ ] Implement JSON-RPC 2.0 handler
- [ ] Integrate Anthropic SDK
- [ ] Create basic file tools
- [ ] Agent Card schema
- [ ] Simple architect agent test

### Week 5-6: Tool Ecosystem
- [ ] Complete file/directory operations
- [ ] Web fetch and search
- [ ] Script execution engine
- [ ] Bash constraints
- [ ] Tool permission system
- [ ] Script approval mechanism

### Week 7-8: Production Features
- [ ] SSE streaming
- [ ] Beads state persistence
- [ ] Checkpoint/resume
- [ ] agent:// protocol handler
- [ ] Error handling
- [ ] Logging and monitoring

### Week 9-10: Integration and Polish
- [ ] End-to-end testing
- [ ] Performance optimization
- [ ] Token efficiency validation
- [ ] Security audit
- [ ] Documentation
- [ ] User guide

## Success Metrics

### Phase 1 Targets
- [ ] 80% of tasks use lightweight agents successfully
- [ ] 30% reduction in orchestrator context size
- [ ] 2x task completion rate (parallelism)
- [ ] < 5s agent spawn latency
- [ ] 95% task completion success rate

### Phase 2 Targets
- [ ] 30-40% token reduction vs Claude Code wrapper
- [ ] 5 concurrent agents supported
- [ ] < 10s task creation latency
- [ ] SSE updates every 5s
- [ ] 100% checkpoint recovery success

## Resources

### Documentation
- [A2A Implementation Plan](./A2A-IMPLEMENTATION-PLAN.md) - Complete 900+ line plan
- [A2A Protocol Overview](./A2A-PROTOCOL.md) - Protocol specification
- [ADR 001](./adr/001-two-tier-agent-architecture.md) - Architecture decision

### Configuration
- [.ai-pack/README.md](../.ai-pack/README.md) - Runtime configuration
- [.ai-pack/scripts/README.md](../.ai-pack/scripts/README.md) - Script guidelines

### Code Examples
See `docs/A2A-IMPLEMENTATION-PLAN.md` for:
- Go code examples (runtime, tools, protocol)
- Configuration YAML templates
- Script examples (Python, Node.js)
- Workflow scenarios

## Getting Help

### Issues and Questions
- Review implementation plan thoroughly first
- Check ADR 001 for architectural decisions
- Consult A2A protocol documentation

### Testing and Validation
- Start with simple single-agent scenarios
- Gradually increase complexity
- Use task packets for debugging
- Monitor token usage and performance

## Next Steps

### Immediate (This Week)
1. Review complete implementation plan
2. Set up development environment
3. Create Phase 1 agent configurations
4. Test first lightweight agent spawn

### Short Term (Weeks 2-4)
1. Complete Phase 1 implementation
2. Validate parallel execution
3. Document results
4. Begin Phase 2 Go server development

### Medium Term (Weeks 5-10)
1. Build complete A2A server
2. Implement all tool sets
3. Add script execution
4. Production hardening
5. Documentation and launch

---

**Ready to begin?** Start with Phase 1, Day 1 tasks above.

**Questions?** Review `docs/A2A-IMPLEMENTATION-PLAN.md` for comprehensive details.

**Tracking**: Use task packets to track implementation progress.
