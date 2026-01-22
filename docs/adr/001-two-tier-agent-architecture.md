# ADR 001: Two-Tier Agent Architecture with A2A Protocol Support

**Status:** Proposed
**Date:** 2026-01-22
**Deciders:** AI-Pack Core Team
**Technical Story:** Extend AI-Pack to support both lightweight local agents and A2A protocol-based remote agents

## Context and Problem Statement

AI-Pack currently operates as a single-agent framework where one AI agent (e.g., Claude Code) executes all roles by switching context via prompts (Orchestrator, Engineer, Reviewer, etc.). This approach works well for sequential tasks but has limitations:

1. **No parallelism:** Cannot run Engineer and Tester simultaneously
2. **Context pollution:** All role context lives in one conversation
3. **No true delegation:** Orchestrator can't spawn independent agents
4. **Scaling limits:** Large tasks require single agent to manage all complexity
5. **No external agent integration:** Cannot leverage specialized agents from other systems

The emerging A2A (Agent-to-Agent) protocol from Google/Linux Foundation provides a standardized way for agents to communicate, but introduces complexity and overhead unsuitable for all scenarios.

**Key Question:** How can AI-Pack support both simple local tasks and complex agent-to-agent workflows without forcing all tasks through heavyweight protocols?

## Decision Drivers

- **Incremental adoption:** Must work with existing AI-Pack workflows
- **Cost efficiency:** Avoid protocol overhead for simple tasks
- **True multi-agent:** Enable parallel execution when beneficial
- **Standards compliance:** Leverage A2A for cross-agent interoperability
- **Backward compatibility:** Don't break existing skills and workflows
- **Developer experience:** Simple tasks stay simple
- **Enterprise readiness:** Support complex, long-running agent workflows

## Considered Options

### Option 1: Lightweight Agents Only
Extend Task tool to spawn role-specific agents that return results synchronously.

**Pros:**
- Simple implementation
- Low overhead
- Familiar pattern (extends existing Task tool)
- Cost-effective

**Cons:**
- No standardized external agent communication
- Limited to Claude Code ecosystem
- No support for long-running tasks
- Cannot integrate third-party agents

### Option 2: A2A Protocol Only
Adopt A2A for all agent communication.

**Pros:**
- Standards-based
- Future-proof
- Enterprise-grade features
- Cross-vendor interoperability

**Cons:**
- High overhead for simple tasks
- Requires A2A infrastructure for all operations
- Increases cost significantly
- Complexity overkill for 80% of use cases

### Option 3: Two-Tier Architecture (Recommended)
Support both lightweight local agents and A2A remote agents with automatic selection based on task characteristics.

**Pros:**
- Best of both worlds
- Graceful degradation
- Cost-optimized (lightweight for most tasks)
- Standards-based when needed
- Incremental migration path
- True multi-agent when beneficial

**Cons:**
- More complex implementation
- Need decision logic for tier selection
- Two code paths to maintain

## Decision Outcome

**Chosen option:** Option 3 - Two-Tier Architecture

We will implement a dual-tier agent system:

### Tier 1: Lightweight Agents (Local)
For quick, focused tasks within the same conversation session.

**Implementation:**
- Enhanced Claude Code Task tool
- Role presets from `/roles/*.md`
- Synchronous request/response
- Minimal context overhead
- No protocol overhead

**Use Cases:**
- Quick code reviews (< 5 min)
- Single file implementations
- Test execution
- Simple refactoring
- Bug fixes

**Example:**
```bash
bd spawn engineer "implement login validation" --mode=lightweight
```

### Tier 2: A2A Agents (Remote)
For complex, long-running, or external agent tasks requiring protocol-based communication.

**Implementation:**
- A2A protocol JSON-RPC over HTTP
- Agent Card-based discovery
- Asynchronous task execution
- State persistence via Beads
- Cross-vendor agent support

**Use Cases:**
- System architecture design (hours)
- Deep bug investigation
- Multi-step research
- External agent delegation
- Cross-team coordination

**Example:**
```bash
bd spawn architect "design microservices architecture" \
  --protocol=a2a \
  --endpoint=agent://company-architect-service
```

### Automatic Tier Selection

Orchestrator role will automatically choose tier based on:

```python
def choose_agent_tier(task):
    """Auto-select lightweight vs A2A"""

    # Lightweight conditions
    if (
        task.estimated_time < 5min and
        task.context_size < 10_000_tokens and
        task.is_synchronous and
        not task.requires_state_persistence and
        not task.requires_external_agent
    ):
        return "lightweight"

    # A2A conditions
    if (
        task.requires_external_agent or
        task.is_long_running or
        task.needs_cross_vendor or
        task.requires_async or
        task.needs_state_persistence
    ):
        return "a2a"

    return "lightweight"  # Default to simpler tier
```

## Implementation Plan

### Phase 1: Lightweight Agents (MVP)
**Timeline:** Sprint 1-2
**Priority:** High

1. Extend Task tool with role presets
2. Add lightweight agent spawning to Beads
3. Integrate with existing roles (`/roles/*.md`)
4. Update Orchestrator to spawn lightweight agents
5. Add quality gates enforcement for agents
6. Documentation and examples

**Deliverables:**
- `bd spawn` command for lightweight agents
- Agent lifecycle management
- Role-based task delegation
- Simple parent-child task tracking

### Phase 2: A2A Protocol Foundation
**Timeline:** Sprint 3-5
**Priority:** Medium

1. Implement A2A protocol client
2. Agent Card schema and parser
3. JSON-RPC communication layer
4. Task lifecycle management (A2A spec)
5. State persistence with Beads
6. Agent registry

**Deliverables:**
- A2A client SDK integration
- Agent discovery system
- Remote agent communication
- Task state persistence

### Phase 3: Hybrid Orchestration
**Timeline:** Sprint 6-8
**Priority:** Medium

1. Auto-decision logic (lightweight vs A2A)
2. Agent-to-agent delegation
3. Cross-tier context sharing
4. Performance monitoring
5. Error handling and fallbacks
6. Comprehensive testing

**Deliverables:**
- Intelligent tier selection
- Unified agent management
- Cross-agent workflows
- Production-ready system

## Architecture

### Directory Structure

```
.ai-pack/
├── agents/
│   ├── lightweight/
│   │   ├── engineer.yml
│   │   ├── reviewer.yml
│   │   ├── tester.yml
│   │   └── README.md
│   ├── a2a/
│   │   ├── architect.yml
│   │   ├── orchestrator.yml
│   │   ├── inspector.yml
│   │   └── README.md
│   └── registry.jsonl         # A2A agent registry
├── protocols/
│   ├── lightweight.md          # Lightweight protocol spec
│   └── a2a.md                 # A2A integration guide
└── config.yml                 # Agent configuration
```

### Configuration

```yaml
# .ai-pack/config.yml
agents:
  default_tier: lightweight

  lightweight:
    max_concurrent: 3
    timeout: 5min
    tools_allowed: [read, write, edit, grep, bash]
    context_limit: 10000

  a2a:
    enabled: true
    registry: .beads/agents/registry.jsonl
    state_store: .beads/agents/state/
    message_queue: .beads/agents/queue/
    max_concurrent: 2
    discovery:
      auto: true
      endpoints:
        - https://agents.company.com/registry

  orchestrator:
    auto_delegate: true
    decision_threshold: 5min
    fallback_tier: lightweight
```

### Lightweight Agent Definition

```yaml
# .ai-pack/agents/lightweight/engineer.yml
name: engineer
description: Implementation specialist following TDD
tier: lightweight
mode: synchronous

context:
  - role: /roles/engineer.md
  - gates: [tdd-enforcement, code-quality-review]

tools:
  - write
  - edit
  - read
  - bash

limits:
  timeout: 5min
  max_context: 10000

output:
  format: task-packet
  artifacts: [code, tests]
```

### A2A Agent Definition

```yaml
# .ai-pack/agents/a2a/architect.yml
name: architect
description: System architecture and design specialist
tier: a2a
protocol: json-rpc-2.0

discovery:
  agent_card: https://agents.company.com/architect/card
  endpoint: https://agents.company.com/architect/a2a

authentication:
  scheme: bearer
  credentials_env: ARCHITECT_AGENT_TOKEN

capabilities:
  - architecture-design
  - system-review
  - technical-decisions

task_management:
  async: true
  state_persistence: required
  progress_updates: streaming

communication:
  mode: asynchronous
  transport: https
  streaming: sse

output:
  format: artifacts
  types: [diagrams, documents, decisions]
```

## Beads Integration

### Enhanced Commands

```bash
# Lightweight agent spawn
bd spawn <role> "<task>" [--mode=lightweight] [--timeout=5m]

# A2A agent spawn
bd spawn <role> "<task>" --protocol=a2a [--endpoint=<url>]

# List active agents
bd agents list [--tier=lightweight|a2a]

# Agent status
bd agents status <agent-id>

# Agent logs
bd agents logs <agent-id> [--follow]

# Agent communication (A2A only)
bd agent-send <agent-id> --message="<json>"
bd agent-recv <agent-id>

# Kill agent
bd agent-kill <agent-id>
```

### Task Packet Enhancement

```markdown
# Task Packet Metadata

## Agent Information
- **Tier**: lightweight | a2a
- **Spawned By**: orchestrator (bd-parent-123)
- **Agent ID**: bd-agent-abc
- **Communication**: sync | async
- **Protocol**: lightweight | a2a

## For Lightweight Agents
- **Context Size**: 2.5KB
- **Execution Mode**: synchronous
- **Expected Duration**: 3min
- **Tools Granted**: [read, write, edit]

## For A2A Agents
- **Endpoint**: https://agents.company.com/architect/a2a
- **Task ID**: task-xyz-789
- **State Checkpoint**: .beads/agents/state/task-xyz-789.json
- **Progress Updates**: SSE stream active
- **Artifacts Expected**: [architecture-diagram.png, decisions.md]
```

## Integration with Existing AI-Pack Components

### Roles
All existing roles (`/roles/*.md`) work with both tiers:
- Lightweight: Role content loaded as agent context
- A2A: Role capabilities advertised in Agent Card

### Workflows
Workflows orchestrate agent spawning:
```markdown
# Feature Workflow (Enhanced)

## Phase 1: Design
- Spawn **Architect** (A2A, async) for system design
- Wait for architecture artifacts

## Phase 2: Implementation
- Spawn **Engineer** (lightweight, sync) for each component
- Parallel execution if independent

## Phase 3: Testing
- Spawn **Tester** (lightweight, sync) for validation
- Spawn **Reviewer** (lightweight, sync) for code review

## Phase 4: Integration
- Orchestrator validates all artifacts
- Marks feature complete
```

### Quality Gates
Gates apply to both tiers:
- Lightweight: Enforced in spawned agent context
- A2A: Communicated via Agent Card capabilities

### Task Packets
Enhanced to track agent lineage:
```
Orchestrator (bd-orch-001)
  └─> Architect Agent (bd-arch-002, A2A)
      └─> Design Task (task-design-123)
  └─> Engineer Agent (bd-eng-003, lightweight)
      └─> Implementation Task (task-impl-456)
  └─> Tester Agent (bd-test-004, lightweight)
      └─> Test Task (task-test-789)
```

## Security Considerations

### Lightweight Agents
- Execute in same security context as parent
- Inherit parent's tool permissions
- Access limited by AI-Pack gates
- No network exposure

### A2A Agents
- Require explicit authentication
- Support multiple auth schemes (bearer, API key, OAuth)
- Agent Cards specify required credentials
- Support credential rotation
- Audit logging for all communications
- Support air-gapped deployments

### Configuration Security
```yaml
# Secure credential management
agents:
  a2a:
    credentials:
      provider: env  # or keychain, vault
      mapping:
        architect: ARCHITECT_AGENT_TOKEN
        inspector: INSPECTOR_AGENT_KEY
```

## Performance Implications

### Lightweight Agents
- **Latency:** ~2-5s (agent spawn + execution)
- **Cost:** ~1.2x single-agent (extra context setup)
- **Concurrency:** 3 parallel agents recommended
- **Memory:** Minimal (shared parent process)

### A2A Agents
- **Latency:** ~10-30s (discovery + task creation)
- **Cost:** Variable (depends on remote agent)
- **Concurrency:** 2 parallel recommended
- **Network:** HTTP/SSE overhead
- **State:** Persistent storage required

### Optimization Strategies
1. Cache Agent Cards (avoid repeated discovery)
2. Reuse lightweight agent contexts
3. Batch related tasks to same agent
4. Use streaming for long-running A2A tasks
5. Implement agent pooling for frequent operations

## Risks and Mitigation

### Risk 1: Complexity
**Mitigation:**
- Phase 1 delivers value with lightweight only
- A2A optional, added later
- Clear documentation and examples

### Risk 2: Cost Increase
**Mitigation:**
- Default to lightweight tier
- Auto-selection optimizes cost
- Usage monitoring and alerts

### Risk 3: A2A Adoption Slow
**Mitigation:**
- Lightweight agents provide immediate value
- A2A enables future capabilities
- Graceful degradation if A2A unavailable

### Risk 4: Agent Coordination Bugs
**Mitigation:**
- Comprehensive testing
- Agent lifecycle state machines
- Beads tracking for debugging
- Rollback to single-agent mode

## Success Metrics

### Phase 1 (Lightweight)
- [ ] 80% of tasks use lightweight agents
- [ ] 30% reduction in orchestrator context size
- [ ] 2x task completion rate (via parallelism)
- [ ] < 5s agent spawn latency

### Phase 2 (A2A)
- [ ] Successfully delegate to external architect agent
- [ ] 24-hour long-running task completion
- [ ] Cross-vendor agent collaboration
- [ ] 99.9% A2A message delivery

### Phase 3 (Hybrid)
- [ ] 95% auto-selection accuracy
- [ ] Zero manual tier selection needed
- [ ] 50% reduction in complex task time
- [ ] Positive developer feedback

## Alternatives Considered

### Alternative 1: MCP-Only
Use Anthropic's Model Context Protocol instead of A2A.

**Analysis:** MCP is for tool/context provisioning, not agent coordination. Complementary to A2A, not alternative.

### Alternative 2: Custom Protocol
Build proprietary agent communication protocol.

**Analysis:** Reinventing wheel; A2A provides standard with vendor support.

### Alternative 3: Prompt Engineering Only
Continue current single-agent context switching.

**Analysis:** Cannot achieve true parallelism or external agent integration.

## Related Decisions

- ADR 002: Beads Integration for Agent State Management (planned)
- ADR 003: Quality Gates for Multi-Agent Workflows (planned)

## References

- [A2A Protocol Specification](https://a2a-protocol.org)
- [AI-Pack A2A Protocol Summary](../A2A-PROTOCOL.md)
- [Google A2A Announcement](https://developers.googleblog.com/en/a2a-a-new-era-of-agent-interoperability/)
- [Model Context Protocol (MCP)](https://modelcontextprotocol.io)

---

**ADR Status:** Proposed
**Next Steps:**
1. Team review and feedback
2. Prototype lightweight agent spawning
3. Validate with real workflows
4. Finalize decision
