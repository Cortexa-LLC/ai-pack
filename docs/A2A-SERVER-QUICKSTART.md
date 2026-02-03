# A2A Server Quick Start Guide

**Version**: 2.1.0
**Last Updated**: 2026-01-24

Get the AI-Pack A2A agent server running in minutes.

---

## Prerequisites

### 1. Go Installation

The A2A server is written in Go. You need Go 1.21 or higher.

**Check if installed:**
```bash
go version
# Should show: go version go1.21 or higher
```

**Install Go:**
- macOS: `brew install go`
- Linux: https://go.dev/doc/install
- Windows: https://go.dev/dl/

### 2. Anthropic API Key

You need an API key to call Claude.

**Option A: Environment Variable**
```bash
export ANTHROPIC_API_KEY="sk-ant-api03-..."
```

**Option B: Claude Code Authentication** (automatic)
- If you're using Claude Code, the server automatically uses your authenticated session
- No manual API key needed

### 3. Beads (Recommended)

For proper task tracking integration:

```bash
# macOS/Linux
curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/scripts/install.sh | bash

# Windows PowerShell
irm https://raw.githubusercontent.com/steveyegge/beads/main/install.ps1 | iex

# Verify installation
bd version
```

---

## Quick Start (3 Steps)

### Step 1: Start the Server

```bash
# From ai-pack root directory
cd a2a-agent
python3 scripts/start-server.py
```

You should see:

```
🚀 AI-Pack Agent Server v2.1.0 starting on localhost:8080

   📡 A2A Protocol Endpoints:
      - GET  /a2a/discovery      (Agent capabilities)
      - POST /a2a/execute        (Execute task - JSON-RPC 2.0)
      - POST /a2a/status         (Task status - JSON-RPC 2.0)

   🔄 Streaming:
      - GET  /stream/:task_id    (Real-time progress - SSE)

   🎯 Features:
      - A2A Protocol Compliance  ✅
      - SSE Streaming            ✅
      - Parallel Execution       ✅ (max 3 concurrent)
      - Beads Integration        ✅

🔗 Registering agent:// protocol handler...
✅ Protocol handler registered
```

**Note**: On first run, the server automatically registers the `agent://` protocol handler for your system. This allows the `agent` CLI to work.

Leave this terminal running.

### Step 2: Create a Task in Beads

In a **new terminal**:

```bash
# Initialize Beads in your project (first time only)
bd init

# Create a task
bd create "Implement user authentication with JWT tokens"
# Returns: bd-a1b2
```

### Step 3: Run an Agent

```bash
# Spawn engineer agent for the task
agent engineer bd-a1b2
```

The agent will:
1. ✅ Validate task exists in Beads
2. ✅ Check dependencies are met
3. ✅ Mark task as started (`bd start bd-a1b2`)
4. ✅ Execute the work
5. ✅ Mark task as complete (`bd close bd-a1b2`)

---

## Alternative: Manual Server Start

If you prefer not to use the Python script:

```bash
cd a2a-agent

# Install Go dependencies
go mod tidy

# Run the server
go run cmd/agent-server/main.go --server
```

---

## Configuration

### Default Configuration

The server uses `a2a-agent/configs/agent-server.json`:

```json
{
  "server": {
    "port": 8080,
    "host": "localhost",
    "max_concurrent_agents": 3,
    "worker_pool_size": 3
  },
  "api": {
    "mode": "direct",
    "max_tokens": 32000,
    "anthropic_model": "claude-sonnet-4-20250514"
  },
  "logging": {
    "level": "info",
    "format": "json"
  }
}
```

### Custom Configuration

Create a custom config file:

```bash
cp configs/agent-server.json configs/my-config.json
# Edit my-config.json
go run cmd/agent-server/main.go --server --config configs/my-config.json
```

### Common Customizations

**Increase concurrent agents:**
```json
{
  "server": {
    "max_concurrent_agents": 5
  }
}
```

**Change port:**
```json
{
  "server": {
    "port": 3000
  }
}
```

**Use proxy (enterprise):**
```json
{
  "api": {
    "mode": "proxy",
    "proxy": {
      "type": "claude-code",
      "base_url": "http://localhost:3333"
    }
  }
}
```

---

## Available Agents

| Agent | Purpose | Typical Use |
|-------|---------|-------------|
| **engineer** | Implementation | Code, features, bug fixes |
| **tester** | Testing | Unit tests, integration tests |
| **reviewer** | Code review | Quality, security, best practices |
| **architect** | System design | Architecture, technical design |
| **product-manager** | Requirements | User stories, acceptance criteria |

### Agent Usage Examples

```bash
# Implementation
bd create "Add password reset endpoint"
agent engineer bd-a1b2

# Testing
bd create "Create tests for password reset"
agent tester bd-x7z9

# Code review
bd create "Review authentication module"
agent reviewer bd-k3m4

# Architecture
bd create "Design microservices architecture"
agent architect bd-p5q6

# Requirements
bd create "Write user stories for checkout flow"
agent product-manager bd-r8s9
```

---

## Workflow Example

Complete feature development workflow:

```bash
# 1. Plan the work
bd create "Implement user profile API"
bd create "Create tests for user profile API"
bd create "Review user profile implementation"

# Set dependencies
bd dep add bd-b2 bd-a1  # Tests depend on implementation
bd dep add bd-c3 bd-a1  # Review depends on implementation

# 2. Check what's ready
bd ready
# Shows: bd-a1 (no dependencies)

# 3. Implement
agent engineer bd-a1

# 4. Check what's ready now
bd ready
# Shows: bd-b2, bd-c3 (dependencies met)

# 5. Run in parallel (both are ready)
agent tester bd-b2 &
agent reviewer bd-c3 &
wait

# 6. View results
bd show bd-a1
bd show bd-b2
bd show bd-c3
```

---

## Troubleshooting

### Server Won't Start

**Problem:** `go: command not found`
```bash
# Install Go first
brew install go  # macOS
# Or download from https://go.dev/dl/
```

**Problem:** `ANTHROPIC_API_KEY not set`
```bash
# Set environment variable
export ANTHROPIC_API_KEY="sk-ant-api03-..."

# Or use Claude Code (automatic)
```

**Problem:** `Port 8080 already in use`
```bash
# Use different port
go run cmd/agent-server/main.go --server --port 3000

# Or find and kill process on port 8080
lsof -i :8080
kill <PID>
```

### Agent Command Fails

**Problem:** `agent: command not found`
```bash
# Build the agent CLI
cd a2a-agent
go build -o ../bin/agent cmd/agent/main.go

# Add to PATH
export PATH="$PATH:$(pwd)/../bin"

# Or use full path
../bin/agent engineer bd-a1b2
```

**Problem:** `Beads task bd-xxxx does not exist`
```bash
# Verify task exists
bd show bd-xxxx

# List all tasks
bd list

# Create the task if missing
bd create "Task description"
```

**Problem:** `task has unmet dependencies: [bd-x, bd-y]`
```bash
# Check which tasks are blocking
bd show bd-x
bd show bd-y

# Complete dependencies first
agent engineer bd-x
agent engineer bd-y

# Then retry original task
agent engineer bd-a1b2
```

### Beads Not Installed

**Problem:** Agent shows warning about Beads
```bash
# Install Beads
curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/scripts/install.sh | bash

# Initialize in project
bd init
```

---

## Performance Tips

### 1. Parallel Execution

Run independent tasks in parallel:

```bash
# Create multiple independent tasks
bd create "Implement feature A"  # bd-a1
bd create "Implement feature B"  # bd-b2
bd create "Implement feature C"  # bd-c3

# Run all in parallel
agent engineer bd-a1 &
agent engineer bd-b2 &
agent engineer bd-c3 &
wait
```

### 2. Increase Concurrency

```bash
# Start server with more concurrent slots
go run cmd/agent-server/main.go --server --max-concurrent 5
```

### 3. Background Execution

```bash
# Run agent in background
agent --async engineer bd-a1b2

# Returns immediately, task runs in background
# Check results later:
bd show bd-a1b2
cat .beads/tasks/task-*/30-results.md
```

---

## Health Checks

### Verify Server is Running

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "version": "2.1.0",
  "server": "ai-pack-agent-server",
  "features": {
    "a2a_protocol": true,
    "sse_streaming": true,
    "parallel_execution": true,
    "monitoring": true
  }
}
```

### Check Metrics

```bash
curl http://localhost:8080/metrics
```

Shows:
- Tasks spawned
- Tasks completed
- API calls made
- Average execution time

---

## Next Steps

1. **Read the A2A Usage Guide**: [docs/content/framework/a2a-usage-guide.md](content/framework/a2a-usage-guide.md)
2. **Understand Beads Integration**: [docs/BEADS-AGENT-INTEGRATION.md](BEADS-AGENT-INTEGRATION.md)
3. **Customize Agent Configs**: [agents/README.md](../agents/README.md)
4. **Set Up Protocol Handler**: [docs/content/framework/protocol-handler.md](content/framework/protocol-handler.md)

---

## Common Commands Reference

```bash
# Server
cd a2a-agent && python3 scripts/start-server.py
go run cmd/agent-server/main.go --server
go run cmd/agent-server/main.go --server --port 3000

# Beads
bd init                                    # Initialize project
bd create "Task description"               # Create task
bd show bd-xxxx                           # View task
bd start bd-xxxx                          # Mark started
bd close bd-xxxx                          # Mark complete
bd ready                                  # Show ready tasks
bd list                                   # List all tasks
bd dep add bd-child bd-parent             # Add dependency

# Agents
agent engineer bd-xxxx                    # Implementation
agent tester bd-xxxx                      # Testing
agent reviewer bd-xxxx                    # Code review
agent architect bd-xxxx                   # Architecture
agent product-manager bd-xxxx                # Requirements
agent --async engineer bd-xxxx            # Background mode

# Health
curl http://localhost:8080/health         # Server health
curl http://localhost:8080/metrics        # Performance metrics
```

---

## Support

- **Documentation**: [docs/content/framework/](content/framework/)
- **Issues**: [GitHub Issues](https://github.com/Cortexa-LLC/ai-pack/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Cortexa-LLC/ai-pack/discussions)

---

**Happy agent automation!** 🚀
