# ADR 002: SSE Task Stream Heartbeat and Client Inactivity Detection

**Status:** Accepted
**Date:** 2026-02-19
**Deciders:** AI-Pack Core Team

## Context and Problem Statement

The task progress stream (`GET /stream/{taskID}`) sends SSE events only when an agent makes
progress — API calls, tool uses, status updates. Between agent turns (while waiting for an LLM
response or long-running tool), the stream is completely silent. This caused two failure modes:

1. **`agent --stream` hangs forever** when the server dies silently (network partition, process
   crash). The client's `resp.Body.Read` blocks indefinitely with no data arriving.

2. **Proxies and load balancers drop idle connections** after their configured timeout (typically
   60–90 seconds), silently severing the stream mid-task.

Additionally, `agent wait` (the polling-based variant) was found orphaning processes after server
restarts: 6 orphaned `agent wait e2e-*` processes were polling the server at 1-second intervals
indefinitely because 404 responses were not treated as terminal.

## Decision

### Server: 30-second SSE heartbeat comment

`streamActiveTaskFromChannel` sends an SSE comment every 30 seconds between real events:

```
: heartbeat

```

SSE comments (lines starting with `:`) are part of the SSE spec and are ignored by clients, but
reset TCP keepalive timers and proxy idle timeouts. The interval (30s) was chosen to be well under
the typical proxy idle timeout (60–90s) while not adding meaningful overhead.

### Client: 2-minute inactivity timeout tied to raw bytes

`streamTaskProgressWithInactivity` uses `context.WithCancel` + `time.AfterFunc`:

```
Server                          Client (agent --stream)
  │                                  │
  ├─ ": heartbeat\n\n" ──────────────► reset 2m timer
  ├─ ": heartbeat\n\n" ──────────────► reset 2m timer
  ├─ "data: {api_call}\n\n" ─────────► reset 2m timer + display event
  ├─ ": heartbeat\n\n" ──────────────► reset 2m timer
  │                                  │
  ✗ server dies silently             │
  │                                  ├─ no bytes for 2 minutes
  │                                  └─ ⏰ "Stream inactive" → exit(1)
```

The timer resets on **raw bytes received** (before SSE parsing), so heartbeat bytes and real event
bytes are treated equally. The 2-minute window ≈ 4 missed heartbeats — clearly indicative of a
dead connection rather than a legitimately quiet agent.

### Client: `agent wait` exits on 3 consecutive 404s

Polling-based `agent wait` now exits with an error after receiving 3 consecutive 404 responses from
`/a2a/status/{taskID}`. A task that no longer exists on the server (after a restart) is treated as
terminal rather than retried indefinitely.

### Client: `agent wait` hard timeout (default 4h)

`agent wait --timeout <duration>` provides a wall-clock upper bound. The 4-hour default covers
even the longest expected engineering tasks.

## Consequences

**Positive:**
- No more orphaned `agent wait` or `agent --stream` processes after server restarts.
- Proxies no longer drop idle streams mid-task.
- Dead connections detected within 2 minutes (down from infinity).
- The connection lifecycle is explicit and documented rather than relying on OS TCP teardown.

**Negative / Trade-offs:**
- Each active task stream now has an additional goroutine (the ticker) and a 30-second timer.
  This is negligible at the concurrency levels this server operates at.
- The 2-minute inactivity window means a truly stuck server (not dead, just frozen) takes
  2 minutes to detect. This is acceptable given the server's own stuck-agent detection fires
  independently and will eventually send a `task_failed` event.

## Related

- `a2a-agent/internal/server/stream_handler.go` — heartbeat ticker in `streamActiveTaskFromChannel`
- `a2a-agent/cmd/agent/main.go` — `streamTaskProgressWithInactivity`, `waitForTaskCompletion`
- `a2a-agent/internal/server/orchestrator_session.go` — orchestrator SSE uses same 30s heartbeat pattern
- `docs/adr/001-two-tier-agent-architecture.md`

## Note on Missing Architecture Overview

`docs/architecture/` is currently empty. A comprehensive a2a-agent architecture document
(component diagram, request flows, task lifecycle) would be a valuable addition. The SSE streaming
design above is one piece of that larger picture.
