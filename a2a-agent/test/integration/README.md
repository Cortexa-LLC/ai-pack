# Orchestrator End-to-End Integration Tests

These tests validate the complete orchestrator workflow using the actual agent CLI commands.

## What These Tests Validate

### 1. **Basic Spawn and Wait** (`TestE2EOrchestratorSpawnAndWait`)
Tests the fundamental orchestrator pattern:
- Spawn agent via `agent engineer <task-id>`
- Block until completion via `agent wait <task-id>`
- Verify status via `agent status <task-id>`

**Validates:**
- Agent spawning mechanism
- Completion detection
- Status querying

### 2. **Parallel Agent Execution** (`TestE2EOrchestratorParallelAgents`)
Tests spawning multiple agents concurrently:
- Spawn 3 agents in parallel
- List all running agents via `agent list --server`
- Wait for all to complete

**Validates:**
- Parallel execution capability
- Server tracking of multiple agents
- Concurrent completion detection

### 3. **Stream Mode** (`TestE2EOrchestratorStreamMode`)
Tests real-time monitoring via SSE:
- Spawn with `--stream` flag
- Verify blocking until completion
- Check real-time progress updates

**Validates:**
- SSE streaming functionality
- Real-time orchestrator monitoring
- Immediate completion detection

### 4. **Cross-Project Coordination** (`TestE2EOrchestratorCrossProjectCoordination`)
Tests machine-wide agent coordination:
- Spawn agents in different project directories
- Query status from different project
- List all agents machine-wide

**Validates:**
- Machine-wide server operation
- Cross-project task discovery
- Project root tracking

### 5. **Log Following** (`TestE2EOrchestratorLogsFollowing`)
Tests real-time log monitoring:
- Spawn agent
- Follow logs via `agent logs <task-id> --follow`
- Verify log streaming

**Validates:**
- Log streaming functionality
- Real-time orchestrator monitoring
- Auto-exit on completion

### 6. **Agent Discovery** (`TestE2EOrchestratorAgentDiscovery`)
Tests capability discovery:
- Query server capabilities via `agent discovery`
- Verify max concurrent limit
- List available agent roles

**Validates:**
- Discovery endpoint
- Capability reporting
- Agent role enumeration

## Running the Tests

### Prerequisites

1. **Build the binaries:**
   ```bash
   cd a2a-agent
   go build -o bin/agent-server ./cmd/agent-server
   go build -o bin/agent ./cmd/agent
   ```

2. **Set environment variables:**
   ```bash
   export ANTHROPIC_API_TOKEN=your-token-here
   ```

### Run All Tests

```bash
cd a2a-agent
go test -v ./test/integration
```

### Run Specific Test

```bash
go test -v ./test/integration -run TestE2EOrchestratorSpawnAndWait
```

### Run with Race Detection

```bash
go test -v -race ./test/integration
```

### Skip Long-Running Tests

```bash
go test -v -short ./test/integration
```

## Test Architecture

### Test Structure

Each test follows this pattern:

```go
1. setupTestEnvironment()     // Create temp dir, .beads/, .ai/
2. startAgentServer()          // Start server on port 8888
3. waitForServer()             // Wait for /health to return 200
4. createMockBeadsTask()       // Create task ID and metadata
5. Run CLI commands            // Execute actual agent CLI
6. Verify results              // Check status, logs, results
7. Cleanup                     // Stop server, remove temp dir
```

### Why End-to-End?

These tests use the actual CLI commands that an orchestrator would use:
- `agent engineer <task-id>` - Spawn
- `agent wait <task-id>` - Block until completion
- `agent status <task-id>` - Check status
- `agent list --server` - See all agents
- `agent logs <task-id> --follow` - Monitor progress
- `agent discovery` - Query capabilities

**NOT** testing internal server APIs directly. This validates the actual orchestrator experience.

## Common Issues

### Server binary not found
```
SKIP: Agent server binary not found - run 'go build -o bin/agent-server ./cmd/agent-server' first
```
**Fix:** Build the server binary first.

### Port already in use
```
Failed to start server: address already in use
```
**Fix:** Kill any existing agent-server on port 8888:
```bash
pkill -f "agent-server.*8888"
```

### Timeout waiting for server
```
Server did not become ready in time
```
**Fix:** Check if server binary is working:
```bash
./bin/agent-server --server --port 8888
```

## Test Output

Successful test run should show:
```
=== RUN   TestE2EOrchestratorSpawnAndWait
    orchestrator_e2e_test.go:45: Started agent-server (PID 12345) in /tmp/e2e-orchestrator-123
    orchestrator_e2e_test.go:119: Server is ready
    orchestrator_e2e_test.go:123: Created mock Beads task: e2e-456 - Test orchestrator spawn
    orchestrator_e2e_test.go:39: Agent spawned: Task spawned successfully
    orchestrator_e2e_test.go:52: Agent completed successfully
--- PASS: TestE2EOrchestratorSpawnAndWait (5.23s)
```

## Integration with Orchestrator Role

These tests validate the patterns documented in:
- `.ai-pack/roles/orchestrator.md` - Section 2.13 (Agent Spawning)
- Orchestrator workflow for parallel execution
- Cross-project coordination capabilities
- Real-time monitoring via streaming

## Performance Expectations

- **Basic spawn and wait**: ~5-10 seconds
- **Parallel 3 agents**: ~10-20 seconds
- **Stream mode**: ~5-10 seconds
- **Cross-project**: ~10-15 seconds
- **Full test suite**: ~60-90 seconds

Times depend on:
- Server startup time (~1-2 seconds)
- Mock agent execution (simulated)
- Network latency (localhost)

## Future Enhancements

Potential additions:
- [ ] Test WIP limit enforcement (>3 agents)
- [ ] Test agent failure scenarios
- [ ] Test result verification with expected outputs
- [ ] Test Beads task closing after completion
- [ ] Test agent retry mechanism
- [ ] Test timeout handling
- [ ] Benchmark concurrent agent spawning
