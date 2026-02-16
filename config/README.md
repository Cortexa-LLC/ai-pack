# Service Configuration Templates

This directory contains service configuration templates for running AI-Pack servers as background services.

## Templates

### Agent Server
- **macOS**: `com.cortexa.ai-pack.agent-server.plist` (LaunchAgent)
- **Linux**: `ai-pack-agent-server.service` (systemd user service)

### GUI Server
- **macOS**: `com.cortexa.ai-pack.gui-server.plist` (LaunchAgent)
- **Linux**: `ai-pack-gui-server.service` (systemd user service)

## Template Placeholders

These files contain placeholders that are replaced during installation:

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `{{PROJECT_ROOT}}` | Git repository location | `/Users/user/Projects/ai-pack` |
| `{{USER}}` | Current username | `bryanw` |
| `{{NPM_PATH}}` | npm binary location | `/opt/homebrew/bin/npm` |

## Binary Locations

**Agent Server Binary**: `/usr/local/bin/agent-server`
- Installed via `sudo make install`
- Service references this fixed location
- No hardcoded project paths

**GUI Server**: Runs from git repository
- Uses `npm run dev` from `{{PROJECT_ROOT}}/gui`
- Requires source files, so uses dynamic path
- npm path detected automatically

## Installation Flow

### 1. Build and Install Binaries
```bash
cd /path/to/ai-pack
make build          # Build binaries
sudo make install   # Install to /usr/local/bin
```

### 2. Install Services
```bash
# Agent server service
python3 scripts/install-service.py

# GUI server service
python3 scripts/install-gui-service.py
```

The install scripts:
1. Detect your operating system
2. Find the git repository location dynamically
3. Replace template placeholders with actual values
4. Install service files to:
   - macOS: `~/Library/LaunchAgents/`
   - Linux: `~/.config/systemd/user/`
5. Start and enable services

### 3. Uninstall Services
```bash
# Uninstall services
python3 scripts/uninstall-service.py

# Uninstall binaries
sudo make uninstall
```

## Key Design Principles

1. **No hardcoded paths** - Templates use placeholders
2. **Standard installation location** - Binaries go to `/usr/local/bin`
3. **Dynamic git location** - Works regardless of where repo is cloned
4. **Cross-platform** - Same templates work on macOS and Linux
5. **Separation of concerns** - Binary installation (make) vs service setup (Python scripts)

## Environment Variables

The services automatically set API keys from environment variables:
- `OPENAI_API_KEY` (optional)
- `ANTHROPIC_API_KEY` (required)

Run `python3 scripts/setup-api-keys.py` to configure API keys in your shell profile.
