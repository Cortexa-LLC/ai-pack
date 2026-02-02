# AI-Pack Startup Guide

Complete guide for building, installing, and running AI-Pack services.

## Table of Contents

- [Quick Start](#quick-start)
- [Build and Install](#build-and-install)
- [Starting Services](#starting-services)
- [launchd Auto-Start (macOS)](#launchd-auto-start-macos)
- [Troubleshooting](#troubleshooting)
- [Advanced Usage](#advanced-usage)

---

## Quick Start

```bash
# Build and install everything
make build install

# Start server and GUI
make start-all

# Setup auto-start (macOS)
make setup-launchd
```

---

## Build and Install

### Build Targets

```bash
# Build agent server binaries
make build

# Build GUI for production
make build-gui

# Build everything
make build-all
```

### Install Targets

```bash
# Install agent binaries to /usr/local/bin (requires sudo)
sudo make install

# Or install just the agent
sudo make install-agent

# Uninstall
sudo make uninstall
```

**What gets installed:**
- `agent-server` → `/usr/local/bin/agent-server`
- `agent` → `/usr/local/bin/agent`

---

## Starting Services

### Option 1: Makefile Targets (Recommended)

```bash
# Start both server and GUI
make start-all

# Start only server
make start-server

# Start only GUI
make start-gui

# Stop all services
make stop-all
```

### Option 2: Direct Script Usage

```bash
# Start both
python3 scripts/start-all.py

# Start only server
python3 scripts/start-all.py --server-only

# Start only GUI
python3 scripts/start-all.py --gui-only
```

### What Happens When You Start

**Agent Server:**
- Checks Go installation
- Verifies ANTHROPIC_API_KEY or Claude Code auth
- Installs Go dependencies
- Starts server on `http://localhost:8080`

**GUI:**
- Checks Node.js installation
- Installs npm dependencies (if needed)
- Starts Vite dev server on `http://localhost:3000`

**Both:**
- Run in foreground
- Press `Ctrl+C` to stop both services gracefully

---

## launchd Auto-Start (macOS)

### Setup

```bash
# Install and start services
make setup-launchd
```

This creates two launchd agents:
- `com.cortexa.ai-pack.agent-server` - Agent server (always running)
- `com.cortexa.ai-pack.gui` - GUI dev server (optional)

**What it does:**
- Creates plist files in `~/Library/LaunchAgents/`
- Loads services into launchd
- Services start automatically on login
- Services restart automatically if they crash

### Check Status

```bash
# Show status of all services
make status-launchd
```

Output shows:
- Installation status
- Running/stopped state
- Process IDs
- Log file locations
- Control commands

### Control Services

```bash
# Start services manually
launchctl start com.cortexa.ai-pack.agent-server
launchctl start com.cortexa.ai-pack.gui

# Stop services
launchctl stop com.cortexa.ai-pack.agent-server
launchctl stop com.cortexa.ai-pack.gui

# Restart services (kill and restart)
launchctl kickstart -k gui/$(id -u)/com.cortexa.ai-pack.agent-server
launchctl kickstart -k gui/$(id -u)/com.cortexa.ai-pack.gui
```

### Logs

Logs are written to `logs/` in the project root:

```bash
# View server logs
tail -f logs/agent-server.log

# View server errors
tail -f logs/agent-server.error.log

# View GUI logs
tail -f logs/gui.log

# View GUI errors
tail -f logs/gui.error.log
```

### Uninstall launchd

```bash
# Remove launchd services
make uninstall-launchd
```

This will:
- Stop all services
- Unload from launchd
- Remove plist files
- Services will NOT start on next login

---

## Troubleshooting

### Server Won't Start

**Check dependencies:**
```bash
# Verify Go is installed
go version

# Should show Go 1.21 or later
```

**Check API key:**
```bash
# Verify API key is set
echo $ANTHROPIC_API_KEY

# Or check Claude Code auth
cat ~/.claude/settings.json
```

**Check if port is in use:**
```bash
# Check if something is using port 8080
lsof -i :8080
```

### GUI Won't Start

**Check dependencies:**
```bash
# Verify Node.js is installed
node --version
npm --version

# Should show Node.js 18+ and npm 9+
```

**Check if port is in use:**
```bash
# Check if something is using port 3000
lsof -i :3000
```

**Reinstall dependencies:**
```bash
cd gui
rm -rf node_modules package-lock.json
npm install
```

### launchd Issues

**Service not loading:**
```bash
# Check for errors
launchctl load ~/Library/LaunchAgents/com.cortexa.ai-pack.agent-server.plist

# If already loaded, unload first
launchctl unload ~/Library/LaunchAgents/com.cortexa.ai-pack.agent-server.plist
launchctl load ~/Library/LaunchAgents/com.cortexa.ai-pack.agent-server.plist
```

**Service crashes immediately:**
```bash
# Check error logs
cat logs/agent-server.error.log

# Common issues:
# - Missing API key (set ANTHROPIC_API_KEY in plist)
# - Go not in PATH (update PATH in plist)
# - Missing dependencies (run: cd a2a-agent && go mod tidy)
```

**Finding process:**
```bash
# Find all AI-Pack processes
ps aux | grep -E "(agent-server|vite.*gui)"

# Kill manually if needed
pkill -f agent-server
pkill -f "vite.*gui"
```

---

## Advanced Usage

### Custom Ports

**Agent Server:**
Edit `a2a-agent/cmd/agent-server/main.go` to change default port (8080).

**GUI:**
Edit `gui/vite.config.ts`:
```typescript
export default defineConfig({
  server: {
    port: 3000, // Already configured
  },
})
```

### Running in Production

**Agent Server:**
```bash
# Build binary
make build

# Run directly
./a2a-agent/bin/agent-server --server

# Or install and run
sudo make install
agent-server --server
```

**GUI:**
```bash
# Build for production
make build-gui

# Serve with any static server
cd gui/dist
python3 -m http.server 8000
```

### Environment Variables

**Agent Server:**
- `ANTHROPIC_API_KEY` - Anthropic API key (required)
- `PORT` - Server port (default: 8080)
- `LOG_LEVEL` - Log level (debug, info, warn, error)

**GUI:**
- `VITE_API_BASE_URL` - Override API base URL

### Adding to PATH

After `make install`, binaries are in `/usr/local/bin/` which is usually in PATH.

Verify:
```bash
which agent-server
# Should output: /usr/local/bin/agent-server

which agent
# Should output: /usr/local/bin/agent
```

---

## Makefile Reference

### Build Targets
- `make build` - Build agent binaries
- `make build-gui` - Build GUI for production
- `make build-all` - Build everything

### Install Targets
- `sudo make install` - Install agent binaries
- `sudo make uninstall` - Uninstall agent binaries

### Start/Stop Targets
- `make start-server` - Start agent server (foreground)
- `make start-gui` - Start GUI dev server (foreground)
- `make start-all` - Start both (foreground)
- `make stop-all` - Stop all services

### launchd Targets (macOS)
- `make setup-launchd` - Install and start launchd services
- `make uninstall-launchd` - Remove launchd services
- `make status-launchd` - Show service status

### Test Targets
- `make test` - Run Go tests
- `make test-short` - Run quick tests
- `make test-coverage` - Generate coverage report
- `make test-gui` - Run GUI tests

### Clean Targets
- `make clean` - Clean agent build artifacts
- `make clean-gui` - Clean GUI artifacts
- `make clean-all` - Clean everything

### Code Quality Targets
- `make lint` - Run Go linters
- `make lint-gui` - Run GUI linters
- `make fmt` - Format Go code
- `make sonarqube` - Run SonarQube analysis

---

## Platform Support

### macOS ✅
- Full support for all features
- launchd for auto-start
- Homebrew for dependencies

### Linux 🚧
- Manual start supported
- Consider systemd for auto-start
- Package managers for dependencies

### Windows 🚧
- Manual start supported
- Consider Task Scheduler or NSSM for auto-start
- Windows package managers for dependencies

---

## Next Steps

After getting services running:

1. **Verify Server:** Visit `http://localhost:8080/health`
2. **Open GUI:** Visit `http://localhost:3000`
3. **Read A2A Docs:** See `a2a-agent/README.md`
4. **Configure Claude Code:** See `docs/CLAUDE-CODE-CONFIGURATION.md`

---

## Getting Help

- **Issues:** Open an issue on GitHub
- **Logs:** Check `logs/` directory
- **Status:** Run `make status-launchd`
- **Documentation:** See `README.md` and `a2a-agent/README.md`
