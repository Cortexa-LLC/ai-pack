# Multi-Agent Workflow Example

This example demonstrates agent-to-agent communication using the A2A protocol.

## Scenario

**Engineer → Tester Workflow**

1. Engineer agent creates a simple function
2. Engineer spawns Tester agent via A2A protocol
3. Tester runs tests on the code
4. Results are coordinated

## Setup

### 1. Start the A2A Server

```bash
# Terminal 1: Start agent-server in server mode
cd a2a-agent
python3 scripts/start-server.py
```

Server will be running on `http://localhost:8080`

### 2. Run the Multi-Agent Script

```bash
# Terminal 2: Execute the workflow
./examples/engineer-tester-workflow.sh
```

## How It Works

### Step 1: Engineer Creates Code

The workflow starts by calling the Engineer agent:

```bash
curl -X POST http://localhost:8080/a2a/execute \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "a2a.execute",
    "params": {
      "role": "engineer",
      "task": "Create a simple add(a, b) function in Go with basic implementation"
    },
    "id": 1
  }'
```

**Returns:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "task_id": "task-engineer-20260124-...",
    "status": "queued",
    "stream_url": "/stream/task-engineer-20260124-..."
  },
  "id": 1
}
```

### Step 2: Wait for Engineer Completion

Poll the status endpoint:

```bash
curl -X POST http://localhost:8080/a2a/status \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "a2a.status",
    "params": {
      "task_id": "task-engineer-20260124-..."
    },
    "id": 2
  }'
```

**Returns:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "task_id": "task-engineer-20260124-...",
    "status": "completed",
    "role": "engineer",
    "created_at": "...",
    "completed_at": "..."
  },
  "id": 2
}
```

### Step 3: Engineer Spawns Tester

Now the Engineer agent (or orchestrator) calls the Tester:

```bash
curl -X POST http://localhost:8080/a2a/execute \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "a2a.execute",
    "params": {
      "role": "tester",
      "task": "Test the add() function that was just created. Run all tests and report coverage."
    },
    "id": 3
  }'
```

### Step 4: Monitor Both Tasks

```bash
# Check tester status
curl -X POST http://localhost:8080/a2a/status \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "a2a.status",
    "params": {
      "task_id": "task-tester-20260124-..."
    },
    "id": 4
  }'
```

### Step 5: Review Results

```bash
# Engineer results
cat .beads/tasks/task-engineer-20260124-.../30-results.md

# Tester results
cat .beads/tasks/task-tester-20260124-.../30-results.md
```

## Agent Communication Pattern

```
┌─────────────────────────────────────────────────────────────┐
│                      Orchestrator                           │
│                    (Shell Script / CLI)                     │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ POST /a2a/execute
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                   Agent Server (A2A)                        │
│                   localhost:8080                            │
└─────────────────────────────────────────────────────────────┘
          │                                   │
          │ Spawn                             │ Spawn
          ▼                                   ▼
┌──────────────────────┐          ┌──────────────────────┐
│  Engineer Agent      │          │  Tester Agent        │
│  - Create code       │          │  - Run tests         │
│  - Save to files     │          │  - Check coverage    │
└──────────────────────┘          └──────────────────────┘
          │                                   │
          │                                   │
          ▼                                   ▼
    .beads/tasks/               .beads/tasks/
    task-engineer-*/            task-tester-*/
```

## Advanced: Programmatic Orchestration

For more complex workflows, agents can call other agents programmatically:

```go
// Example: Agent calling another agent via A2A
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type A2ARequest struct {
    JSONRPC string                 `json:"jsonrpc"`
    Method  string                 `json:"method"`
    Params  map[string]interface{} `json:"params"`
    ID      int                    `json:"id"`
}

func spawnAgent(role, task string) (string, error) {
    req := A2ARequest{
        JSONRPC: "2.0",
        Method:  "a2a.execute",
        Params: map[string]interface{}{
            "role": role,
            "task": task,
        },
        ID: 1,
    }

    body, _ := json.Marshal(req)
    resp, err := http.Post(
        "http://localhost:8080/a2a/execute",
        "application/json",
        bytes.NewBuffer(body),
    )
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    taskID := result["result"].(map[string]interface{})["task_id"].(string)
    return taskID, nil
}
```

## Next Steps

1. Run the example workflow
2. Check the generated code and tests
3. Modify the workflow for your use case
4. Build multi-step agent orchestration
