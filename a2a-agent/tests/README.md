# A2A Agent Tests

This directory contains tests for the AI-Pack Agent Server.

## Directory Structure

### `e2e/` - End-to-End Tests
End-to-end tests that verify complete workflows, including:
- Multi-project support
- Client-server integration
- Beads task lifecycle
- Real agent execution scenarios

**Example:**
- `test_multi_project.sh` - Tests multi-project Beads support

### Running E2E Tests

```bash
# Run specific e2e test
cd a2a-agent/tests/e2e
./test_multi_project.sh

# Or from project root
cd a2a-agent
./tests/e2e/test_multi_project.sh
```

### Prerequisites

E2E tests require:
- Agent server running (`agent-server --server`)
- Beads installed (`bd` command available)
- Valid test projects with `.beads/` directories

## Unit Tests

Go unit tests are located alongside the code:
- `internal/*/.../*_test.go` - Standard Go unit tests

Run unit tests:
```bash
cd a2a-agent
go test ./...
```

## Test Coverage

For Go test coverage:
```bash
go test -cover ./...
```

## Adding New Tests

### E2E Tests
1. Create test script in `tests/e2e/`
2. Make it executable: `chmod +x tests/e2e/test_name.sh`
3. Document prerequisites and expected environment

### Unit Tests
1. Create `*_test.go` file alongside the code being tested
2. Follow Go testing conventions
3. Use table-driven tests for multiple scenarios

## CI/CD

E2E tests should be run in CI pipeline after:
- Building binaries
- Starting test agent-server
- Setting up test projects

Unit tests run on every commit.
