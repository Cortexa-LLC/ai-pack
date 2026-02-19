# a2a-agent Architecture

**Last updated:** 2026-02-19

This document describes the design of the `a2a-agent` server — its components, request flows,
task lifecycle, and streaming subsystem. For the rationale behind specific design decisions, see
the ADRs in `docs/adr/`.

---

## Table of Contents

1. [Component Overview](#component-overview)
2. [HTTP API Surface](#http-api-surface)
3. [Task Lifecycle](#task-lifecycle)
4. [SSE Task Streaming](#sse-task-streaming)
5. [LLM Provider Abstraction](#llm-provider-abstraction)
6. [Concurrency Model](#concurrency-model)
7. [Persistence](#persistence)
8. [Key Data Structures](#key-data-structures)

---

## Component Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        a2a-agent server                         │
│                                                                 │
│  ┌─────────────┐   ┌──────────────┐   ┌────────────────────┐   │
│  │  HTTP layer │   │ AgentServer  │   │  streaming.Service │   │
│  │  (net/http) │──►│  (server.go) │──►│  (factory + run)   │   │
│  └─────────────┘   └──────┬───────┘   └────────┬───────────┘   │
│                           │                    │               │
│              ┌────────────┼──────────┐         │               │
│              ▼            ▼          ▼         ▼               │
│  ┌──────────────┐  ┌──────────┐  ┌──────────────────────────┐  │
│  │ TaskExecution│  │  Beads   │  │  LLM Providers           │  │
│  │ (per task)   │  │  Client  │  │  ┌──────────────────┐    │  │
│  │              │  │  (tasks, │  │  │ AnthropicProvider │    │  │
│  │ streamChan   │  │  status) │  │  │ OpenAIProvider    │    │  │
│  │ cancel()     │  └──────────┘  │  │ ModelSelector     │    │  │
│  └──────────────┘                │  └──────────────────┘    │  │
│                                  └──────────────────────────┘  │
│                                                                 │
│  ┌───────────────────┐   ┌────────────────────────────────────┐ │
│  │  ExecutionLog     │   │  MCP Manager                       │ │
│  │  (SQLite JSONL)   │   │  (tool servers per project)        │ │
│  └───────────────────┘   └────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

**Key packages:**

| Package | Responsibility |
|---------|---------------|
| `internal/server` | HTTP handlers, task orchestration, `AgentServer` |
| `internal/streaming` | `StreamProvider` interface, Anthropic/OpenAI adapters |
| `internal/protocol` | Shared types (`StreamEvent`, `Message`, tool use/result) |
| `internal/monitoring` | Metrics, model selection, logging |
| `internal/beads` | Client for the Beads task-tracking database (`bd`) |
| `internal/execution_log` | Persistent SQLite log of all task events |
| `internal/mcp` | MCP server manager (spawns/tracks MCP processes per project) |
| `internal/graphql` | GraphQL schema, resolvers, generated types |

---

## HTTP API Surface

```
POST /a2a/tasks              — Create and queue a new agent task
GET  /a2a/tasks              — List all known tasks
GET  /a2a/tasks/{taskID}     — Get task detail
POST /a2a/tasks/{taskID}/cancel — Cancel a running task

GET  /a2a/status/{taskID}    — Polling endpoint (used by `agent wait`)
GET  /stream/{taskID}        — SSE stream of task progress events

POST /graphql                — GraphQL endpoint (GUI queries + mutations)
GET  /graphql                — GraphQL playground

GET  /health                 — Health check
GET  /version                — Build version info
```

The `agent` CLI binary connects to these endpoints. The GUI uses GraphQL. The SSE stream is
the primary way to observe a running task in real time.

---

## Task Lifecycle

```
                  POST /a2a/tasks
                        │
                        ▼
               ┌─────────────────┐
               │    "queued"     │ ◄── TaskExecution created, streamChan opened
               └────────┬────────┘     workerPool slot required to advance
                        │
                  worker slot acquired
                        │
                        ▼
               ┌─────────────────┐
               │  "in_progress"  │ ◄── LLM streaming begins, events flow to streamChan
               └────────┬────────┘
                        │
            ┌───────────┴─────────────┐
            │                         │
            ▼                         ▼
  ┌──────────────────┐       ┌─────────────────┐
  │   "completed"    │       │    "failed"     │
  └──────────────────┘       └─────────────────┘
            │                         │
            └────────────┬────────────┘
                         │
                    streamChan closed
                         │
                    task removed from activeTasks
                    log written to .beads/tasks/{id}/execution.log
```

**Status values:** `queued` → `in_progress` → `completed` | `failed` | `cancelled`

A task stuck in `in_progress` for longer than its configured timeout is transitioned to
`failed` by the stuck-agent watchdog that runs independently.

---

## SSE Task Streaming

Every agent task gets a `streamChan chan *protocol.StreamEvent` (buffered, capacity 100).
The task runner writes events to this channel as work proceeds; `GET /stream/{taskID}`
reads from it and forwards to the HTTP client as SSE.

### Two streaming paths

```
GET /stream/{taskID}
        │
        ├── task in activeTasks?
        │       │
        │       YES ─► streamActiveTaskFromChannel()   [channel-based, live]
        │       │
        │       NO  ─► findTaskProjectRoot()
        │                   │
        │           ┌───────┴───────┐
        │           │               │
        │     per-task log      no per-task log
        │     exists?           (old task)
        │           │               │
        │     streamCompleted   streamFromGlobalLog()
        │     TaskLog()         [SQLite execution_log]
        │     [JSONL file]
        │
        └── in all cases: send "connected" event first,
                          send "stream_closed" event last
```

### Active task streaming — heartbeat and inactivity

`streamActiveTaskFromChannel` runs a ticker alongside the event channel select:

```
Server (streamActiveTaskFromChannel)          Client (agent --stream)
  │                                                  │
  │  [30s ticker fires]                              │
  ├─ ": heartbeat\n\n" ─────────────────────────────►│ reset 2m inactivity timer
  │                                                  │
  │  [LLM produces text chunk]                       │
  ├─ "data: {content_block_delta}\n\n" ─────────────►│ reset 2m inactivity timer + display
  │                                                  │
  │  [LLM calls a tool]                              │
  ├─ "data: {tool_use}\n\n" ────────────────────────►│ reset 2m inactivity timer + display
  │                                                  │
  │  [server process crashes / network partition]    │
  ✗                                                  │
                                                     ├─ no bytes for 2 minutes
                                                     └─ ⏰ "Stream inactive" → exit(1)
```

- **Server side:** 30-second SSE comment heartbeat (`": heartbeat\n\n"`) keeps proxies and
  load balancers from dropping idle connections (typical proxy idle timeout: 60–90 s).
- **Client side:** `time.AfterFunc(2m)` cancels the HTTP request context if no bytes arrive.
  The timer resets on **raw bytes** — heartbeat bytes and event bytes are equivalent. The
  2-minute window equals approximately 4 missed heartbeats, which clearly indicates a dead
  connection rather than a legitimately quiet agent between turns.

> **ADR:** See `docs/adr/002-sse-task-stream-heartbeat.md` for the rationale behind the
> specific interval choices.

### Event types (StreamEvent.Type)

| Type | Meaning |
|------|---------|
| `connected` | Stream opened (always first) |
| `content_block_delta` | Text chunk from LLM |
| `tool_use` | Agent is calling a tool |
| `tool_result` | Tool execution result |
| `api_call` | LLM API call started |
| `api_response` | LLM API response received |
| `task_completed` | Task finished successfully |
| `task_failed` | Task failed with error |
| `stream_closed` | Stream closing (always last) |

---

## LLM Provider Abstraction

All LLM interaction goes through the `streaming.Service`, which wraps provider-specific
factories behind a single `StreamProvider` interface:

```
streaming.Service
      │
      ├── AnthropicFactory ──► AnthropicStreamAdapter
      │   (claude-* models)    implements StreamProvider
      │
      └── OpenAIFactory ──────► chatStreamAdapter       (chat completions API)
          (gpt-*, codex-*)      responsesStreamAdapter   (responses API — codex only)
                                both implement StreamProvider
```

**StreamProvider interface** (`internal/streaming/interfaces.go`):
```go
type StreamProvider interface {
    Next() bool          // advance to next event
    Current() StreamEvent
    Err() error
    Close()
    GetMessage() *CompletedMessage  // final token counts, stop reason
    GetModel() string
    GetProvider() string
}
```

**Model routing** is handled by `ModelSelector` (`internal/server/model_selector.go`), which
picks the provider and model based on role tier, configured preferences, and adaptive
performance grades stored in `.claude/performance_grades/`.

**Codex models** (`gpt-5.1-codex`, `gpt-5.1-codex-mini`, `gpt-5.2-codex`) use OpenAI's
Responses API (`/v1/responses`) instead of Chat Completions — they are not supported on
`/v1/chat/completions`. The `isCodexModel()` helper in `openai_adapter.go` routes these.

---

## Concurrency Model

```
New() creates:
  workerPool = make(chan struct{}, maxConcurrent)  // semaphore
  taskQueue  = make(chan *TaskExecution, 100)

  → N goroutines: taskWorker() — each acquires workerPool slot, runs task

Task submission (POST /a2a/tasks):
  1. Create TaskExecution{Status: "queued"}
  2. Register in activeTasks map (under mu write lock)
  3. Send to taskQueue channel

taskWorker goroutine:
  1. Receive from taskQueue
  2. Acquire workerPool slot (blocks if maxConcurrent reached)
  3. Set status → "in_progress"
  4. Run agent loop (LLM + tools)
  5. Set status → "completed" or "failed"
  6. Close streamChan
  7. Release workerPool slot
```

`activeTasks map[string]*TaskExecution` is protected by `sync.RWMutex`. Stream handlers
take a read lock to check `activeTasks`; task runners take a write lock to update status.

---

## Persistence

Three persistence layers:

| Layer | Location | Purpose |
|-------|----------|---------|
| **Beads** | `.beads/tasks/` (per project) | Human-readable task tracking; `bd` CLI integration |
| **Execution log** | SQLite in server data dir | Queryable history for all tasks; powers `GET /stream/{id}` for completed tasks |
| **Per-task event log** | `.beads/tasks/{id}/execution.log` | JSONL of all StreamEvents; played back by `streamCompletedTaskLog()` |

When a client streams a **completed** task, the server reads the JSONL event log and replays
it as SSE at 10 ms/event — providing the same event format as the live stream.

---

## Key Data Structures

### TaskExecution
```go
type TaskExecution struct {
    TaskID      string
    Role        string
    Task        string
    Config      *AgentConfig
    StartTime   time.Time
    Status      string       // "queued" | "in_progress" | "completed" | "failed"
    Result      string
    Error       string
    ProjectRoot string

    streamChan  chan *protocol.StreamEvent  // buffered(100); closed when task ends
    cancel      context.CancelFunc         // cancels the agent's context
}
```

### StreamEvent (protocol package)
```go
type StreamEvent struct {
    Type      string       // event type (see table above)
    TaskID    string
    Timestamp time.Time
    Delta     *DeltaContent // for content_block_delta
    ToolUse   *ToolUse      // for tool_use
    ToolResult *ToolResult  // for tool_result
    Message   *CompletedMessage // for task_completed (token counts, stop reason)
    Error     string        // for task_failed
}
```

### AgentConfig
Loaded from role `.md` files (e.g., `roles/engineer.md`). Specifies model, tier, timeout,
max turns, allowed tools, and quality gates. The server's markdown parser extracts
`**Field:** value` pairs from a `---`-delimited header block at the top of each role file.

---

## Related Docs

- `docs/adr/001-two-tier-agent-architecture.md` — Why local vs A2A agents
- `docs/adr/002-sse-task-stream-heartbeat.md` — Why 30s heartbeat + 2m inactivity timeout
- `a2a-agent/docs/LOG-STREAMING.md` — Log streaming detail
- `a2a-agent/docs/WORKING-DIRECTORY.md` — Project root resolution
- `docs/A2A-PROTOCOL.md` — A2A protocol overview
