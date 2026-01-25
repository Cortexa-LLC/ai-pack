# Building A2A Agent

## Quick Start

```bash
# Build both binaries
make

# Build just the server
make server

# Build just the CLI
make agent

# Clean build artifacts
make clean

# Run tests
make test

# Install to /usr/local/bin
sudo make install
```

## Development

```bash
# Start server in development mode
make dev-server

# Format code
make fmt

# Run linter
make lint
```

## Version Information

```bash
# Show version info
make version
```

Output includes:
- Git tag/commit
- Short commit hash
- Build timestamp

## Binaries

After building:
- `bin/agent-server` - The A2A protocol server
- `bin/agent` - The CLI client for managing agents

## Requirements

- Go 1.21 or later
- Make

## Manual Build

If you prefer not to use Make:

```bash
# Server
go build -o bin/agent-server ./cmd/agent-server

# CLI
go build -o bin/agent ./cmd/agent
```
