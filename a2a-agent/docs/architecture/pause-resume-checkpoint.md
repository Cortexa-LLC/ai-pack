# Pause/Resume Checkpoint Architecture

**Status:** Design  
**Author:** Spelunker Agent  
**Relates to:** `internal/server/server.go`, `internal/server/a2a_handlers.go`, `cmd/agent/main.go`

---

## Overview

When an agent hits its `MaxBudgetTokens` limit the current code does a hard-stop:

```go
// server.go:1634-1637 (current behaviour — to be replaced)
if used >= limit {
    logMsg(fmt.Sprintf("❌ Token budget exhausted: %d/%d tokens used", used, limit))
    return finalResult.String(), fmt.Errorf("token budget exceeded: used %d tokens (limit: %d)", used, limit)
}
```

This loses all conversation context. The goal is to replace that `return` with a **pause checkpoint** that serialises the agentic-loop state to disk, emits a `budget_paused` SSE event, closes the stream, and then allows a subsequent `POST /a2a/tasks/{taskID}/resume` call (with an optional budget top-up) to reload the checkpoint and continue exactly where execution left off.

---

## 1. Task Status State Machine

### Current states (`internal/constants/constants.go`)

```
queued → in_progress → completed
                     → failed
                     → blocked   (soft-coded string in server.go:2042)
```

### New state: `paused`

Add to `internal/constants/constants.go`:

```go
StatusPaused = "paused"
```

Updated state machine:

```
queued → in_progress → completed
                     → failed
                     → blocked
                     → paused → in_progress  (on resume)
                              → failed       (if resume fails / checkpoint corrupt)
```

**Rules:**
- A task can only be resumed from `paused`. Attempting to resume a `completed`, `failed`, or `in_progress` task returns HTTP 409.
- A `paused` task is **not** retried by the worker pool (the existing orphan-recovery path in `handleOrphanedTasks`, ~line 2536, must also skip `paused` tasks).
- Cancellation of a `paused` task transitions it directly to `failed` (same as cancelling `in_progress`).
- The `TaskExecution.Status` field comment at line 102 must be updated:

```go
Status string // "queued", "in_progress", "completed", "failed", "blocked", "paused"
```

---

## 2. Checkpoint Schema

### File location

```
{ProjectRoot}/.beads/tasks/{taskID}/checkpoint.json
```

This sits alongside the existing `task.json` / `status.json` files that `loadTaskStatusFromDisk` already reads.

### Go struct definitions

Place in a new file `internal/server/checkpoint.go`:

```go
package server

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/cortexa-llc/ai-pack/a2a-agent/internal/streaming"
)

// AgentCheckpoint captures the full mutable state of the agentic loop
// at the moment the token budget was exhausted. All fields are required
// for a successful resume.
type AgentCheckpoint struct {
    // Identity
    TaskID    string    `json:"task_id"`
    CreatedAt time.Time `json:"created_at"`

    // Loop counters (mirrors local variables in runAgentLoop)
    Turn                  int   `json:"turn"`
    TotalInputTokens      int64 `json:"total_input_tokens"`
    TotalOutputTokens     int64 `json:"total_output_tokens"`
    InactiveTurns         int   `json:"inactive_turns"`
    ConsecutiveErrorTurns int   `json:"consecutive_error_turns"`
    LastTextLength        int   `json:"last_text_length"`
    LastToolSignature     string `json:"last_tool_signature"`

    // Token budget snapshot
    BudgetLimit int64 `json:"budget_limit"` // original MaxBudgetTokens
    BudgetUsed  int64 `json:"budget_used"`  // TotalInputTokens + TotalOutputTokens at pause

    // Conversation history
    // streaming.Message must be JSON-serialisable (Role, Content, ToolUses, ToolResults).
    // All existing fields on streaming.Message are already plain Go types, so
    // encoding/json handles this with no custom marshaller required.
    Messages []streaming.Message `json:"messages"`

    // Accumulated result text (finalResult.String() at pause point)
    PartialResult string `json:"partial_result"`

    // Config snapshot — enough to recreate the agent; the full AgentConfig
    // is re-loaded from disk on resume, but MaxBudgetTokens is overridden
    // by the resume call's additionalBudget parameter.
    Role        string `json:"role"`
    ProjectRoot string `json:"project_root"`
    Model       string `json:"model"` // from AgentConfig.Model at pause time
}

const checkpointFileName = "checkpoint.json"

// writeCheckpoint serialises cp to {projectRoot}/.beads/tasks/{taskID}/checkpoint.json.
// It is called from the budget-exhausted branch (server.go ~line 1634) before
// setting status to "paused" and closing the stream.
func writeCheckpoint(projectRoot, taskID string, cp *AgentCheckpoint) error {
    dir := filepath.Join(projectRoot, ".beads", "tasks", taskID)
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return fmt.Errorf("checkpoint mkdir: %w", err)
    }
    data, err := json.MarshalIndent(cp, "", "  ")
    if err != nil {
        return fmt.Errorf("checkpoint marshal: %w", err)
    }
    path := filepath.Join(dir, checkpointFileName)
    // Write atomically: temp file + rename
    tmp := path + ".tmp"
    if err := os.WriteFile(tmp, data, 0o644); err != nil {
        return fmt.Errorf("checkpoint write: %w", err)
    }
    return os.Rename(tmp, path)
}

// loadCheckpoint reads and deserialises the checkpoint for taskID.
func loadCheckpoint(projectRoot, taskID string) (*AgentCheckpoint, error) {
    path := filepath.Join(projectRoot, ".beads", "tasks", taskID, checkpointFileName)
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("checkpoint read: %w", err)
    }
    var cp AgentCheckpoint
    if err := json.Unmarshal(data, &cp); err != nil {
        return nil, fmt.Errorf("checkpoint unmarshal: %w", err)
    }
    return &cp, nil
}

// checkpointPath returns the canonical path for callers that only need
// the path string (e.g. for the SSE event payload).
func checkpointPath(projectRoot, taskID string) string {
    return filepath.Join(projectRoot, ".beads", "tasks", taskID, checkpointFileName)
}
```

---

## 3. Agentic Loop Changes (`internal/server/server.go`)

### 3a. Replace hard-stop (~line 1629) with pause-and-checkpoint

**Before (lines 1632–1637):**

```go
if used >= limit {
    logMsg(fmt.Sprintf("❌ Token budget exhausted: %d/%d tokens used", used, limit))
    return finalResult.String(), fmt.Errorf("token budget exceeded: used %d tokens (limit: %d)", used, limit)
}
```

**After:**

```go
if used >= limit {
    logMsg(fmt.Sprintf("⏸️  Token budget exhausted: %d/%d tokens used — pausing", used, limit))

    cp := &AgentCheckpoint{
        TaskID:                execution.TaskID,
        CreatedAt:             time.Now(),
        Turn:                  turn,
        TotalInputTokens:      totalInputTokens,
        TotalOutputTokens:     totalOutputTokens,
        InactiveTurns:         inactiveTurns,
        ConsecutiveErrorTurns: consecutiveErrorTurns,
        LastTextLength:        lastTextLength,
        LastToolSignature:     lastToolSignature,
        BudgetLimit:           limit,
        BudgetUsed:            used,
        Messages:              messages,
        PartialResult:         finalResult.String(),
        Role:                  execution.Role,
        ProjectRoot:           execution.ProjectRoot,
        Model:                 config.Model,
    }
    if err := writeCheckpoint(execution.ProjectRoot, execution.TaskID, cp); err != nil {
        logMsg(fmt.Sprintf("⚠️  Failed to write checkpoint: %v", err))
        // Fall through to hard-fail if we can't persist state
        return finalResult.String(), fmt.Errorf("token budget exceeded and checkpoint failed: %w", err)
    }

    // Emit SSE pause event BEFORE closing the stream
    s.sendStreamEvent(execution, "budget_paused", map[string]interface{}{
        "used":            used,
        "limit":           limit,
        "turn":            turn,
        "checkpoint_path": checkpointPath(execution.ProjectRoot, execution.TaskID),
    })

    // Transition status to "paused"
    s.mu.Lock()
    execution.Status = constants.StatusPaused
    s.mu.Unlock()
    s.persistTaskStatus(execution) // existing helper that writes status.json

    // Close stream (same as normal completion)
    s.closeStream(execution)

    // Return a sentinel error that the caller (spawnAgentLoop or equivalent)
    // recognises as a clean pause rather than a failure.
    return finalResult.String(), ErrTaskPaused
}
```

Add the sentinel error near the top of `server.go` (alongside other package-level errors):

```go
// ErrTaskPaused is returned by the agentic loop when a token-budget pause
// checkpoint has been written successfully. The task status is already "paused"
// when this error surfaces; callers must NOT mark the task as "failed".
var ErrTaskPaused = errors.New("task paused: token budget exhausted")
```

### 3b. Caller of the agentic loop

Wherever `runAgentLoop` (or the equivalent function containing the loop) is called, add a branch for `ErrTaskPaused`:

```go
result, err := s.runAgentLoop(ctx, execution, config, initialPrompt)
if err != nil {
    if errors.Is(err, ErrTaskPaused) {
        // Status already set to "paused", stream already closed.
        // Nothing more to do — do NOT call updateTaskCompletion or closeStream again.
        logMsg(fmt.Sprintf("⏸️  Task %s paused at checkpoint", execution.TaskID))
        return
    }
    // ... existing failure handling ...
}
```

---

## 4. Stream Behaviour on Pause

| Phase | Action |
|---|---|
| Normal turns | `sendStreamEvent` works as today |
| Budget-exhausted detection | Emit `budget_paused` event (see §3a above) |
| After `budget_paused` event | Call `closeStream(execution)` immediately |
| Reconnection while paused | Client receives HTTP 404 or a static `{"status":"paused"}` — **do not reopen** the channel |

**Why close the stream?**  
Keeping the channel open requires the goroutine to stay alive and consume memory indefinitely. Because the checkpoint is durable on disk, the client can poll `/a2a/status/{taskID}` to discover `paused` state without a live SSE connection.

**SSE event format** (`type = "budget_paused"`):

```
event: budget_paused
data: {"used":850000,"limit":1000000,"turn":42,"checkpoint_path":"/project/.beads/tasks/abc-xyz/checkpoint.json"}
```

The existing `sendStreamEvent` function at ~line 2393 already writes to both the live channel and the per-task log file (`writeStreamEventToFile`), so historical clients can replay this event from the log.

---

## 5. Server API

### Route Registration

Add to `cmd/agent-server/main.go` alongside existing routes:

```go
mux.HandleFunc("/a2a/tasks/", s.HandleTasksRouter) // new router for /a2a/tasks/{id}/resume
```

Or, if the mux supports path parameters, register directly:

```go
mux.HandleFunc("/a2a/resume/", s.HandleResumeTask)  // consistent with /a2a/retry/ and /a2a/cancel/
```

**Recommendation:** Use `/a2a/resume/{taskID}` to match the existing `/a2a/retry/{taskID}` and `/a2a/cancel/{taskID}` pattern that the mux already uses.

### Handler

Add to `internal/server/a2a_handlers.go`:

```go
// HandleResumeTask handles POST /a2a/resume/{taskID}
//
// Optional JSON body:
//   {"additional_budget": 500000}
//
// If additional_budget is omitted or 0 the task resumes with 2× the original
// limit (i.e. original limit + original limit).
func (s *AgentServer) HandleResumeTask(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
        return
    }

    taskID := strings.TrimPrefix(r.URL.Path, "/a2a/resume/")
    if taskID == "" {
        http.Error(w, "task ID required", http.StatusBadRequest)
        return
    }

    // Parse optional body
    var body struct {
        AdditionalBudget int `json:"additional_budget"`
    }
    if r.ContentLength > 0 {
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
            http.Error(w, "invalid JSON body", http.StatusBadRequest)
            return
        }
    }

    // Load task status to verify it is "paused"
    taskInfo, err := s.loadTaskStatusFromDisk(taskID)
    if err != nil {
        http.Error(w, fmt.Sprintf("task not found: %v", err), http.StatusNotFound)
        return
    }
    if taskInfo["status"] != constants.StatusPaused {
        http.Error(w,
            fmt.Sprintf("task is %s, not paused", taskInfo["status"]),
            http.StatusConflict) // 409
        return
    }

    // Locate project root (same lookup used by HandleRetryTask)
    projectRoot := s.findProjectRootForTask(taskID)
    if projectRoot == "" {
        http.Error(w, "cannot locate project root for task", http.StatusInternalServerError)
        return
    }

    // Load checkpoint
    cp, err := loadCheckpoint(projectRoot, taskID)
    if err != nil {
        http.Error(w, fmt.Sprintf("checkpoint load failed: %v", err), http.StatusInternalServerError)
        return
    }

    // Determine new budget
    newBudget := int(cp.BudgetLimit) + body.AdditionalBudget
    if body.AdditionalBudget == 0 {
        newBudget = int(cp.BudgetLimit) * 2 // default: double the original
    }

    // Dispatch resume — runs in worker pool, same as HandleRetryTask
    go s.resumeFromCheckpoint(taskID, projectRoot, cp, newBudget)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusAccepted) // 202
    json.NewEncoder(w).Encode(map[string]interface{}{
        "task_id":       taskID,
        "status":        "resuming",
        "budget_before": cp.BudgetUsed,
        "budget_new":    newBudget,
        "checkpoint":    checkpointPath(projectRoot, taskID),
    })
}
```

---

## 6. Resume Execution (`internal/server/server.go`)

Add `resumeFromCheckpoint` to `server.go`:

```go
// resumeFromCheckpoint creates a new TaskExecution for a previously paused task,
// restores the conversation state from cp, and continues the agentic loop.
func (s *AgentServer) resumeFromCheckpoint(taskID, projectRoot string, cp *AgentCheckpoint, newBudget int) {
    // Re-load config so any role-file changes are picked up
    config, err := s.loadAgentConfig(cp.Role, projectRoot)
    if err != nil {
        monitoring.Logger.Error("resume_config_load_failed", "task_id", taskID, "error", err)
        s.markTaskFailed(taskID, projectRoot, fmt.Sprintf("resume: config load failed: %v", err))
        return
    }

    // Override budget with the new (extended) limit
    config.Delegation.MaxBudgetTokens = newBudget

    // Create a fresh TaskExecution — new stream channel, new cancel context
    ctx, cancel := context.WithCancel(context.Background())
    execution := &TaskExecution{
        TaskID:      taskID,
        Role:        cp.Role,
        Task:        taskID, // task description is in the messages history
        Config:      config,
        StartTime:   time.Now(),
        Status:      constants.StatusInProgress,
        ProjectRoot: projectRoot,
        streamChan:  make(chan *protocol.StreamEvent, 100),
        streamOpen:  true,
        cancel:      cancel,
    }

    // Register execution so SSE subscribers can attach
    s.mu.Lock()
    s.executions[taskID] = execution
    s.mu.Unlock()

    defer cancel()

    // Emit a resume event so live subscribers know the loop restarted
    s.sendStreamEvent(execution, "status_update", map[string]interface{}{
        "status":       constants.StatusInProgress,
        "resumed_from": checkpointPath(projectRoot, taskID),
        "budget_new":   newBudget,
        "turn_start":   cp.Turn,
    })

    // Continue the loop, passing restored state.
    // The agentic loop function signature needs a new overload (or additional
    // parameters) to accept pre-existing messages + counters:
    result, err := s.continueAgentLoop(ctx, execution, config, cp)

    // Handle result identically to a normal loop exit
    if err != nil && !errors.Is(err, ErrTaskPaused) {
        s.handleTaskFailure(execution, result, err)
        return
    }
    if errors.Is(err, ErrTaskPaused) {
        // Paused again — a new checkpoint has already been written inside continueAgentLoop
        return
    }
    s.handleTaskCompletion(execution, result)
}
```

### `continueAgentLoop` signature

Extract the agentic loop into a reusable form that accepts restored state. The simplest approach is to add optional "seed" parameters; the cleanest approach is a new internal struct:

```go
type agentLoopState struct {
    Messages              []streaming.Message
    Turn                  int
    TotalInputTokens      int64
    TotalOutputTokens     int64
    InactiveTurns         int
    ConsecutiveErrorTurns int
    LastTextLength        int
    LastToolSignature     string
    PartialResult         string
}

// continueAgentLoop resumes from an existing checkpoint state.
// For a fresh run, callers pass a state with Turn=1 and an empty Messages
// slice (the current runAgentLoop behaviour).
func (s *AgentServer) continueAgentLoop(
    ctx context.Context,
    execution *TaskExecution,
    config *AgentConfig,
    cp *AgentCheckpoint,   // nil = fresh start
) (string, error) {
    // Restore state from checkpoint, or initialise fresh
    var state agentLoopState
    if cp != nil {
        state = agentLoopState{
            Messages:              cp.Messages,
            Turn:                  cp.Turn,
            TotalInputTokens:      cp.TotalInputTokens,
            TotalOutputTokens:     cp.TotalOutputTokens,
            InactiveTurns:         cp.InactiveTurns,
            ConsecutiveErrorTurns: cp.ConsecutiveErrorTurns,
            LastTextLength:        cp.LastTextLength,
            LastToolSignature:     cp.LastToolSignature,
            PartialResult:         cp.PartialResult,
        }
    } else {
        state.Turn = 1
        // Messages initialised by caller before this function
    }

    // ... rest of loop logic, reading/writing `state.*` instead of bare locals ...
}
```

> **Engineering note:** The current `runAgentLoop` uses bare local variables (`turn`, `totalInputTokens`, etc.). Refactoring to use `agentLoopState` is a mechanical search-and-replace within that function. The loop body is otherwise unchanged.

---

## 7. CLI: `agent resume`

Add to `cmd/agent/main.go`:

### Switch case

```go
case "resume":
    handleResume(os.Args[2:])
```

### Handler function

```go
func handleResume(args []string) {
    fs := flag.NewFlagSet("resume", flag.ExitOnError)
    budgetFlag := fs.Int("budget", 0,
        "Additional tokens to add to the budget (e.g. 500000). "+
            "If 0, the server defaults to 2× the original limit.")
    fs.Parse(args)

    if fs.NArg() < 1 {
        fmt.Fprintln(os.Stderr, "Usage: agent resume <task-id> [--budget +500000]")
        os.Exit(1)
    }
    taskID := fs.Arg(0)

    body := map[string]interface{}{
        "additional_budget": *budgetFlag,
    }
    bodyJSON, _ := json.Marshal(body)

    url := fmt.Sprintf("%s/a2a/resume/%s", ServerURL, taskID)
    resp, err := http.Post(url, "application/json", strings.NewReader(string(bodyJSON)))
    if err != nil {
        fmt.Printf("❌ Failed to resume task: %v\n", err)
        os.Exit(1)
    }
    defer resp.Body.Close()
    respBody, _ := io.ReadAll(resp.Body)

    if resp.StatusCode != http.StatusAccepted {
        fmt.Printf("❌ Server rejected resume (%d): %s\n", resp.StatusCode, string(respBody))
        os.Exit(1)
    }

    var result map[string]interface{}
    json.Unmarshal(respBody, &result)
    fmt.Printf("⏯️  Task %s resuming (new budget: %v tokens)\n",
        taskID, result["budget_new"])
    fmt.Printf("   Run `agent logs %s` to follow progress\n", taskID)
}
```

### Usage examples

```bash
# Resume with default budget (2× original)
agent resume abc-xyz

# Resume with explicit extra 500k tokens
agent resume abc-xyz --budget 500000

# Resume then tail logs
agent resume abc-xyz --budget 500000 && agent logs abc-xyz
```

---

## 8. SSE Event Reference

| Event type | Trigger | Data fields |
|---|---|---|
| `budget_paused` | Budget exhausted, checkpoint written | `used`, `limit`, `turn`, `checkpoint_path` |
| `status_update` (resuming) | `resumeFromCheckpoint` starts | `status="in_progress"`, `resumed_from`, `budget_new`, `turn_start` |
| Existing events | Unchanged | — |

---

## 9. Orphan-Task Recovery

`handleOrphanedTasks` (~line 2536) currently restarts tasks that are `in_progress` but have no live execution. It must **skip** `paused` tasks — they are intentionally dormant and have a checkpoint on disk:

```go
// In handleOrphanedTasks, add guard:
if status == constants.StatusPaused {
    continue // do not restart paused tasks
}
```

---

## 10. Implementation Checklist

| # | File | Change |
|---|---|---|
| 1 | `internal/constants/constants.go` | Add `StatusPaused = "paused"` |
| 2 | `internal/server/checkpoint.go` | **New file** — `AgentCheckpoint`, `writeCheckpoint`, `loadCheckpoint`, `checkpointPath` |
| 3 | `internal/server/server.go:102` | Update `Status` comment to include `"paused"` |
| 4 | `internal/server/server.go` | Add `var ErrTaskPaused = errors.New(...)` |
| 5 | `internal/server/server.go:1632` | Replace hard-stop `return` with checkpoint+pause flow (§3a) |
| 6 | `internal/server/server.go` | Refactor loop locals into `agentLoopState`; add `continueAgentLoop` (§6) |
| 7 | `internal/server/server.go` | Add `resumeFromCheckpoint` method (§6) |
| 8 | `internal/server/server.go:2536` | Skip `paused` tasks in `handleOrphanedTasks` (§9) |
| 9 | `internal/server/a2a_handlers.go` | Add `HandleResumeTask` (§5) |
| 10 | `cmd/agent-server/main.go` | Register `/a2a/resume/` route |
| 11 | `cmd/agent/main.go` | Add `case "resume"` + `handleResume` (§7) |

---

## 11. Edge Cases and Failure Modes

| Scenario | Handling |
|---|---|
| `writeCheckpoint` fails (disk full, permission error) | Log warning, fall through to existing hard-fail `return` (no data loss beyond current behaviour) |
| Checkpoint file missing on resume | `HandleResumeTask` returns HTTP 500 with descriptive error |
| Task resumed twice concurrently | Second `POST /a2a/resume` sees status `in_progress` → returns HTTP 409 |
| Budget exhausted again during resumed run | Same path: new `budget_paused` event, checkpoint file **overwritten**, status stays `paused` |
| Server restart while task is `paused` | Checkpoint persists on disk; status persists in `status.json`; task is resumable after restart |
| Agent context window overflow during resume | Existing truncation logic (~line 1466) already applies; no special handling needed |
| `streaming.Message` contains non-JSON-serialisable fields added in future | `AgentCheckpoint` will fail marshal; add unit test to catch regressions |

---

## 12. Testing Notes

- **Unit test** `writeCheckpoint` / `loadCheckpoint` round-trip with a realistic `streaming.Message` slice (including `ToolUses` and `ToolResults`).
- **Integration test** mock agentic loop that exhausts budget after 3 turns; assert: (a) `checkpoint.json` exists, (b) status is `"paused"`, (c) SSE log contains `budget_paused` event, (d) resume continues from turn 4 with restored token counters.
- **CLI test** `agent resume <id> --budget 0` triggers default-double behaviour (inspect server log for new budget value).
