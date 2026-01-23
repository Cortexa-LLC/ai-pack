# AI-Pack Runtime Configuration

This directory contains runtime configuration for AI-Pack's two-tier agent architecture.

## Directory Structure

```
.ai-pack/
├── README.md                    # This file
├── config.yml                   # Main configuration (to be created)
├── a2a-server.yml              # A2A server config (Phase 2)
├── agents/
│   ├── lightweight/            # Phase 1 agent configurations
│   ├── a2a/                    # Phase 2 A2A agent configurations
│   └── .approved-scripts.json  # Script approval tracking (generated)
└── scripts/                    # Custom automation scripts
    └── README.md               # Script documentation
```

## Implementation Phases

### Phase 1: Lightweight Agents (Current)
- Using Claude Code Task tool for agent spawning
- Configurations in `agents/lightweight/`
- No additional infrastructure required
- See: `docs/A2A-IMPLEMENTATION-PLAN.md` for details

### Phase 2: A2A Server (Future)
- Go-based A2A protocol server
- Direct Anthropic API integration
- Configurations in `agents/a2a/`
- Requires Go A2A server running
- See: `docs/A2A-IMPLEMENTATION-PLAN.md` for roadmap

## Getting Started

### Phase 1 Setup

1. Review the implementation plan:
   ```bash
   cat docs/A2A-IMPLEMENTATION-PLAN.md
   ```

2. Create lightweight agent configurations (examples coming soon)

3. Use via orchestrator:
   ```bash
   bd spawn engineer "implement feature X"
   ```

### Phase 2 Setup (Future)

Phase 2 implementation will include:
- Go A2A server installation
- Configuration templates
- Script library
- Complete setup guide

## Configuration

Configuration files will be created during implementation:
- `config.yml` - Main AI-Pack configuration
- `a2a-server.yml` - A2A server settings
- `agents/tool-permissions.yml` - Tool security policies

## Documentation

- **Implementation Plan**: `docs/A2A-IMPLEMENTATION-PLAN.md`
- **A2A Protocol**: `docs/A2A-PROTOCOL.md`
- **ADR 001**: `docs/adr/001-two-tier-agent-architecture.md`

## Status

**Current Phase**: Planning Complete ✅
**Next Step**: Phase 1 Implementation (Lightweight Agents)
**Timeline**: See `docs/A2A-IMPLEMENTATION-PLAN.md` Section "Implementation Roadmap"
