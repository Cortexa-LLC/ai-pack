.PHONY: all build test clean

# Build binaries required for tests
build:
	@echo "Building agent binaries..."
	@cd ../.. && go build -o bin/agent-server ./cmd/agent-server
	@cd ../.. && go build -o bin/agent ./cmd/agent
	@echo "✓ Binaries built: bin/agent-server, bin/agent"

# Run all integration tests
test: build clean
	@echo "Running end-to-end orchestrator tests..."
	@go test -v -timeout 15m .

# Run specific test
test-spawn: build clean
	@go test -v -timeout 5m -run TestE2EOrchestratorSpawnAndWait .

test-parallel: build clean
	@go test -v -timeout 12m -run TestE2EOrchestratorParallelAgents .

test-stream: build clean
	@go test -v -timeout 5m -run TestE2EOrchestratorStreamMode .

test-cross-project: build clean
	@go test -v -timeout 8m -run TestE2EOrchestratorCrossProjectCoordination .

test-logs: build clean
	@go test -v -timeout 5m -run TestE2EOrchestratorLogsFollowing .

test-discovery: build clean
	@go test -v -timeout 1m -run TestE2EOrchestratorAgentDiscovery .

# Run with race detection
test-race: build clean
	@echo "Running tests with race detector..."
	@go test -v -race -timeout 15m .

# Clean up test artifacts
clean:
	@echo "Cleaning up..."
	@pkill -f "agent-server.*8888" 2>/dev/null || true
	@rm -rf /tmp/e2e-orchestrator-*
	@echo "✓ Cleaned up test artifacts"

# Check prerequisites
check:
	@echo "Checking prerequisites..."
	@command -v go >/dev/null 2>&1 || { echo "✗ Go not found"; exit 1; }
	@command -v curl >/dev/null 2>&1 || { echo "✗ curl not found"; exit 1; }
	@test -n "$$ANTHROPIC_API_KEY" -o -n "$$ANTHROPIC_API_TOKEN" || { echo "⚠️  ANTHROPIC_API_KEY or ANTHROPIC_API_TOKEN not set (tests may fail)"; }
	@echo "✓ Prerequisites OK"

# Show help
help:
	@echo "Orchestrator E2E Integration Tests"
	@echo ""
	@echo "Commands:"
	@echo "  make build           Build agent binaries"
	@echo "  make test            Run all tests"
	@echo "  make test-spawn      Test basic spawn and wait"
	@echo "  make test-parallel   Test parallel execution"
	@echo "  make test-stream     Test stream mode"
	@echo "  make test-cross-project  Test cross-project coordination"
	@echo "  make test-logs       Test log following"
	@echo "  make test-discovery  Test agent discovery"
	@echo "  make test-race       Run with race detector"
	@echo "  make clean           Clean up test artifacts"
	@echo "  make check           Check prerequisites"
	@echo ""
	@echo "Environment:"
	@echo "  ANTHROPIC_API_TOKEN=your-token-here"

all: check build test
