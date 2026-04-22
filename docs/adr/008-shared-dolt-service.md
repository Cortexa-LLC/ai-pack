# ADR-008: Shared Dolt Service for Multi-Project Task Tracking

**Date:** 2026-04-22
**Status:** Accepted
**Deciders:** Bryan Woodruff

---

## Context

Beads uses Dolt as its SQL backend. Dolt is a MySQL-compatible database with git-style
version control built in. Early in AI-Pack development, each project ran its own embedded
Dolt process (spawned inline by `bd`). This created several problems as the number of
projects grew:

- Multiple Dolt processes competing for system resources
- Port conflicts when projects happened to use the same default port
- No shared visibility — the agent-server could not query tasks across projects from a
  single connection
- Service management complexity: N launchd plists for N projects

---

## Decision

Run a **single, shared Dolt SQL server** for all projects on the developer's machine.

| Parameter | Value |
|-----------|-------|
| Data directory | `~/.beads/dolt/` |
| Listen address | `127.0.0.1:3307` |
| Service label (macOS) | `com.beads.dolt-shared` |
| Default MySQL port | 3306 (avoided to prevent conflicts) |

Each project is provisioned as an **isolated database** within that single server.
Dolt natively supports this: when `dolt sql-server --data-dir <path>` is used, every
subdirectory of `<path>` that contains a Dolt repository is exposed as a separate MySQL
database. Projects cannot see each other's data unless they explicitly `USE <other_db>`.

---

## Alternatives Considered

### Option A — Embedded Dolt per project (previous default)

`bd init` without flags starts an in-process Dolt that listens on a random port and
exits when `bd` exits. Fine for single-project use, but:

- Cannot share a persistent connection across agent-server sessions
- Port is ephemeral — the agent-server cannot hold a stable connection
- No way to run `bd` commands from a different shell without restarting Dolt

**Rejected:** Does not scale beyond one project.

### Option B — One persistent Dolt process per project

Each project gets its own `dolt sql-server` on a unique port, managed by a dedicated
launchd plist.

- Full isolation between projects
- Simple per-project configuration

**Rejected:** Resource overhead grows linearly with projects. Managing N plists is
friction. Cross-project visibility requires N connections.

### Option C — Shared Dolt server, single `data-dir` (chosen)

One process, one port, all projects as separate databases.

- O(1) resource overhead regardless of project count
- Single launchd plist (`com.beads.dolt-shared`) manages the service
- Projects are isolated at the MySQL `USE <db>` level
- Beads and the agent-server connect once and stay connected

---

## Consequences

**Positive:**
- Adding a new project is `bd init --server-host 127.0.0.1 --server-port 3307` — no new
  service or port required.
- The agent-server auto-provisions new projects (see `ensureBeadsForProject`).
- Single log file (`~/.beads/dolt-shared.log`) for all Dolt activity.
- Dolt's git-style history applies per-database — each project retains full task history.

**Negative / Tradeoffs:**
- The shared server is a single point of failure for all projects. If it crashes, all
  projects lose Beads connectivity until it restarts (launchd `KeepAlive: true` restarts
  it automatically).
- The `~/.beads/dolt/` data directory must be backed up to preserve all project task
  history. Per-project `.beads/` directories contain config only, not the DB files.
- Port 3307 is non-standard; it must not conflict with another MySQL instance on the
  machine. Override by editing the launchd plist and re-running `bd init` with the
  new port.

---

## Implementation

The shared Dolt service is installed by `make setup-services` (via `scripts/setup-services.py`).
The generated launchd plist is written to:

```
~/Library/LaunchAgents/com.beads.dolt-shared.plist
```

Key plist parameters:
```xml
<string>dolt sql-server --host 127.0.0.1 --port 3307 --data-dir ~/.beads/dolt --loglevel warning</string>
```

New projects are provisioned at first use:
1. Manually: `bd init --server-host 127.0.0.1 --server-port 3307`
2. Automatically: `AgentServer.ensureBeadsForProject()` in `internal/server/server_core.go:999`

---

## Related Documents

- [`docs/guides/shared-dolt-service.md`](../guides/shared-dolt-service.md) — Setup and adoption guide
- [`docs/MULTI_PROJECT_SUPPORT.md`](../MULTI_PROJECT_SUPPORT.md) — Multi-project agent-server design
- [`INSTALL.md`](../../INSTALL.md) — Full installation guide
