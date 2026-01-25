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
