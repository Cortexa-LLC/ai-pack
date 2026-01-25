# A2A Agent Server

**Agent-to-Agent (A2A) Protocol Server for AI-Pack**

This directory contains the Go-based agent server infrastructure that implements the A2A protocol for spawning and managing AI agents.

---

## Overview

The A2A agent server provides:
- **A2A Protocol Compliance** (JSON-RPC 2.0)
- **SSE Streaming** for real-time progress updates
- **Parallel Execution** with configurable concurrency
- **Structured Logging & Metrics**

---

## Directory Structure

```
a2a-agent/
├── cmd/                    # Application entry points
│   ├── agent/             # Agent CLI
│   └── agent-server/      # Agent server
├── internal/              # Private implementation (Go enforced)
│   ├── auth/             # API key authentication
│   ├── config/           # Configuration loading
│   ├── monitoring/       # Logging and metrics
│   ├── protocol/         # A2A protocol implementation
│   ├── proxy/            # Proxy transport layer
│   └── server/           # HTTP server and handlers
├── configs/               # Configuration files
│   ├── agent-server.json        # Default config
│   └── agent-server-direct.json # Direct mode config
├── scripts/               # Helper scripts
│   ├── start-server.py   # Server startup (cross-platform)
│   └── setup.py          # Installation script
├── go.mod                 # Go module definition
├── go.sum                 # Go dependencies
└── README.md              # This file
```

---

## Quick Start

### Prerequisites

- **Go 1.21+** installed
- **Make** (optional, for build convenience)
- **ANTHROPIC_API_KEY** environment variable or Claude Code authentication

### Building

```bash
cd a2a-agent

# Build both server and CLI
make

# Or build individually
make server  # Builds bin/agent-server
make agent   # Builds bin/agent
```

See [BUILD.md](BUILD.md) for all build options.

### Starting the Server

```bash
# Run the server
./bin/agent-server --server

# Or use make
make dev-server
```

---

## Agent CLI Usage

The `agent` CLI provides commands for spawning and managing agent tasks using Beads task IDs.

### Spawning Agents

```bash
# Spawn agent (fire and forget)
agent engineer <beads-task-id>

# Spawn and wait for completion (polling)
agent engineer <beads-task-id> --wait

# Spawn and stream real-time progress (SSE)
agent engineer <beads-task-id> --stream
```

**Example:**
```bash
# Create a Beads task first
bd create "Implement user authentication" --priority high

# Spawn engineer agent with Beads task ID
agent engineer xasm++-vp5

# Or stream progress in real-time
agent engineer xasm++-vp5 --stream
```

### Monitoring Tasks

```bash
# Check task status
agent status <beads-task-id>

# View task results
agent results <beads-task-id>

# View execution logs
agent logs <beads-task-id>

# List all active agents
agent list

# List only running agents
agent list --running
```

### Metrics and Utilities

```bash
# Show server metrics
agent metrics

# Show modified files (git)
agent files <beads-task-id>

# Show git diff
agent diff <beads-task-id>

# Wait for task completion
agent wait <beads-task-id>
```

### How Task IDs Work

**You use Beads task IDs everywhere.** The CLI handles internal task ID conversion automatically:

1. You provide Beads task ID (e.g., `xasm++-vp5`)
2. CLI finds corresponding internal task ID (e.g., `task-engineer-20260125-150405`)
3. CLI uses internal ID for API calls
4. You never need to know about internal IDs

**Task metadata location:**
```
.beads/tasks/task-engineer-20260125-150405/
├── 00-metadata.json      # Task metadata (includes beads_task_id)
├── 10-plan.md           # Agent's plan
├── 20-work-log.md       # Progress tracking
├── 30-results.md        # Final results
├── agent-prompt.txt     # Prompt sent to agent
└── execution.log        # Execution log
```

### Streaming vs Polling

**Streaming (`--stream`):**
- Real-time SSE updates
- Lower latency
- Better for monitoring multiple agents
- Recommended for orchestrators

**Polling (`--wait`):**
- Checks status every 5 seconds
- Simpler implementation
- Good for simple scripts

---

## Configuration

Server configuration is in `configs/agent-server.json`:

```json
{
  "server": {
    "port": 8080,
    "host": "localhost"
  },
  "agent": {
    "maxConcurrent": 3,
    "maxTokens": 32000,
    "model": "claude-sonnet-4"
  }
}
```

### Configuration Modes

- **Default** (`agent-server.json`): Standard configuration
- **Direct** (`agent-server-direct.json`): Direct API calls (no proxy)
- **Proxy** (see .gitignore): Routes through proxy (enterprise environments)

---

## Development

### Running Tests

```bash
cd a2a-agent
go test ./...
```

### Building

```bash
cd a2a-agent
go build -o bin/agent-server cmd/agent-server/main.go
```

### Code Organization

- **`cmd/`**: Executable entry points (main.go files)
- **`internal/`**: Private packages (enforced by Go compiler)
  - Can only be imported by code in `a2a-agent/`
  - Not accessible to external projects

---

## API Endpoints

### A2A Protocol Endpoints

- `POST /api/a2a/invoke` - Invoke an agent task
- `GET /api/a2a/stream/:taskId` - Stream task progress (SSE)
- `GET /api/a2a/status/:taskId` - Get task status
- `POST /api/a2a/cancel/:taskId` - Cancel running task

### Health & Metrics

- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics

---

## Architecture

### Request Flow

```
Client Request
    ↓
HTTP Server (server.go)
    ↓
A2A Handlers (a2a_handlers.go)
    ↓
Protocol Layer (protocol/a2a.go)
    ↓
Anthropic API (via proxy or direct)
    ↓
SSE Stream (stream_handler.go)
    ↓
Client Response
```

### Concurrency Model

- Task queue with configurable worker pool
- Semaphore-based concurrency limiting
- Per-task progress tracking
- Graceful shutdown on SIGTERM/SIGINT

---

## Related Documentation

- **Consumer Framework**: `../README.md` (main AI-Pack docs)
- **Roles & Workflows**: `../roles/`, `../workflows/`
- **A2A Usage Guide**: `../docs/content/framework/a2a-usage-guide.md`

---

## License

See LICENSE file in project root.
