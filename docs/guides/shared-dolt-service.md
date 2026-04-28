# Shared Dolt Service — Multi-Project Setup

AI-Pack uses a single Dolt SQL server to back task tracking for **all projects**
on the developer's machine. Each project gets its own isolated database within that
server — no separate process or port per project.

This guide explains how the pattern works and how to adopt it in a new project.

---

## How It Works

```
~/.beads/dolt/               ← shared Dolt data directory
├── ai_pack/                 ← database for the ai-pack project
│   └── ...dolt internals...
├── my-project/              ← database for my-project
│   └── ...dolt internals...
└── another-project/         ← database for another-project
    └── ...dolt internals...
```

One `dolt sql-server` process serves all of these databases over MySQL protocol on
`127.0.0.1:3307`. Dolt's `--data-dir` flag makes every subdirectory that contains a
Dolt repo visible as a separate MySQL database. Projects are isolated — querying one
database does not expose another's data.

The server runs as a persistent background service managed by launchd (macOS) or
systemd (Linux). It starts at login and restarts automatically if it crashes.

---

## Prerequisites

- [Dolt](https://docs.dolthub.com/introduction/installation) installed and on `PATH`
- [Beads](https://github.com/steveyegge/beads) (`bd`) installed
- AI-Pack installed (`make setup-services` not yet run)

Verify:
```bash
which dolt   # e.g. /opt/homebrew/bin/dolt
which bd
```

---

## Initial Setup (First Time)

Run once from the AI-Pack repository root:

```bash
make setup-services
```

This installs and starts three services:

| Service label | Purpose |
|---------------|---------|
| `com.beads.dolt-shared` | Shared Dolt SQL server — **must be running before `bd init`** |
| `com.cortexa.ai-pack.agent-server` | AI-Pack agent server |
| `com.cortexa.ai-pack.gui` | GUI dev server |

Verify the Dolt service is up:

```bash
make status-services
# or directly:
mysql -h 127.0.0.1 -P 3307 -u root --protocol=tcp -e "SHOW DATABASES;"
```

Expected output includes one row per initialized project plus Dolt system databases.

---

## Adding a New Project

### Option 1 — Manual initialization

Run this once in the new project's root directory:

```bash
cd /path/to/your-project
bd init --server-host 127.0.0.1 --server-port 3307
```

Beads creates a new database in `~/.beads/dolt/` named after the directory, and writes
a `.beads/` config directory in the project root that points at the shared server.

> **Important:** Always use `--server-host` and `--server-port` flags. Running plain
> `bd init` without flags starts an embedded in-process Dolt instead of connecting to
> the shared server, which conflicts with the service.

Verify the project is registered:

```bash
mysql -h 127.0.0.1 -P 3307 -u root --protocol=tcp -e "SHOW DATABASES;"
# your-project should now appear in the list
```

### Option 2 — Automatic (via agent-server)

When the agent-server processes a task for a project it has never seen before, it
automatically calls:

```go
// internal/server/server_core.go:999
func (s *AgentServer) ensureBeadsForProject(projectRoot string) {
    cmd := exec.Command("bd", "init", "--server-host", "127.0.0.1", "--server-port", "3307")
    cmd.Dir = projectRoot
    ...
}
```

This means running `agent engineer <task-id>` from a new project for the first time
is sufficient — no manual `bd init` needed.

---

## Database Isolation

Each project's database is a fully independent Dolt repository. You can:

- Run `bd` commands in one project without affecting another
- Inspect a project's full task history with Dolt's git tooling:
  ```bash
  cd ~/.beads/dolt/your-project
  dolt log
  dolt diff HEAD~1
  ```
- Roll back task state to a previous point in time via `dolt checkout`

The Dolt server enforces database-level isolation using MySQL's `USE <db>` semantics.
A client connected to `my-project` cannot read `another-project` unless it explicitly
switches databases.

---

## Service Management

```bash
# macOS — control all services
make start-all       # start Dolt + agent-server + GUI
make stop-all        # stop all three
make restart-all     # restart all three
make status-services # show running status + PIDs

# Start/stop individual services
launchctl kickstart -k gui/$(id -u)/com.beads.dolt-shared
launchctl kill TERM  gui/$(id -u)/com.beads.dolt-shared
```

Service configuration lives at:
```
~/Library/LaunchAgents/com.beads.dolt-shared.plist
```

Log output:
```bash
tail -f ~/.beads/dolt-shared.log
```

### Linux (systemd)

The shared Dolt service is **not** installed automatically on Linux. Start it manually:

```bash
dolt sql-server \
  --host 127.0.0.1 \
  --port 3307 \
  --data-dir ~/.beads/dolt \
  --loglevel warning &
```

For a persistent service, add a systemd user unit. The agent-server and GUI units
are installed by `make setup-services` on Linux, but not the Dolt unit (PRs welcome).

---

## Configuration

The Dolt server reads `~/.beads/dolt/config.yaml` at startup. The default config that
ships with AI-Pack sets the listen address and port:

```yaml
listener:
  host: 127.0.0.1
  port: 3307
```

Uncomment and edit fields as needed. See the
[Dolt configuration reference](https://docs.dolthub.com/sql-reference/server/configuration)
for all available options.

### Using a different port

If port 3307 conflicts with another service:

1. Edit `~/.beads/dolt/config.yaml` — change `port: 3307` to your preferred port.
2. Edit `~/Library/LaunchAgents/com.beads.dolt-shared.plist` — update the `--port` argument.
3. Re-run `bd init --server-host 127.0.0.1 --server-port <new-port>` in each project.
4. Update `ensureBeadsForProject` in `internal/server/server_core.go` if you want
   the agent-server auto-provisioning to use the new port.

---

## Troubleshooting

### `bd` commands fail with "connection refused"

The shared Dolt server is not running. Start it:

```bash
launchctl kickstart -k gui/$(id -u)/com.beads.dolt-shared
# or
make start-all
```

### New project not visible after `bd init`

Confirm the init ran against the shared server (not embedded):

```bash
cat /path/to/your-project/.beads/config.json | grep -E "host|port"
# should show: "host": "127.0.0.1", "port": 3307
```

If it shows an embedded config, re-run:

```bash
rm -rf /path/to/your-project/.beads
cd /path/to/your-project
bd init --server-host 127.0.0.1 --server-port 3307
```

### Port 3307 in use by another process

```bash
lsof -i :3307
```

Identify the process. If it is not the shared Dolt service, see
[Using a different port](#using-a-different-port) above.

### Dolt crashes on startup (bad `config.yaml`)

Check the log:
```bash
tail -50 ~/.beads/dolt-shared.log
```

Dolt validates `config.yaml` at startup and will refuse to start if it contains
unknown or malformed keys. Compare against the reference copy at
`.beads/dolt/config.yaml` in this repository.

---

## Backup

All project task history lives in `~/.beads/dolt/`. Back it up like any other data:

```bash
# Simple rsync snapshot
rsync -av ~/.beads/dolt/ ~/backups/beads-dolt-$(date +%Y%m%d)/

# Or use Dolt's built-in remote push (requires a DoltHub account or self-hosted remote)
cd ~/.beads/dolt/my-project
dolt remote add origin https://doltremoteapi.dolthub.com/<org>/<repo>
dolt push origin main
```

---

## Related Documents

- [ADR-008: Shared Dolt Service](../adr/008-shared-dolt-service.md) — Design rationale
- [MULTI_PROJECT_SUPPORT.md](../MULTI_PROJECT_SUPPORT.md) — How the agent-server handles multiple projects
- [INSTALL.md](../../INSTALL.md) — Full installation guide including `make setup-services`
