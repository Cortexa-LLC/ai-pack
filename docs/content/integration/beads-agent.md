---
sidebar_position: 3
title: Beads & Agent Integration
---

# Beads-Agent Integration Gap

**Status**: ⚠️ **NEEDS IMPLEMENTATION**
**Priority**: High
**Effective Date**: 2026-01-24

---

## Problem Statement

The current agent implementation accepts arbitrary task descriptions instead of Beads task IDs. This breaks the integration between the task tracking system (Beads) and agent execution.

### Current (Incorrect) Behavior

```bash
# Create a task in Beads
bd create "Implement user authentication"
# Returns: bd-a1b2

# Agent command doesn't use the Beads task ID
agent engineer "implement user authentication"  # ❌ No connection to Beads
```

**Issues:**
- ❌ No connection between Beads tasks and agent execution
- ❌ Agent doesn't know about task dependencies
- ❌ Agent can't update task status (start, complete)
- ❌ No task context from Beads
- ❌ Duplicate task descriptions (one in Beads, one in agent command)

---

## Expected Behavior

### Workflow

1. **Create task in Beads:**
   ```bash
   bd create "Implement user authentication"
   # Returns: bd-a1b2
   ```

2. **Spawn agent with Beads task ID:**
   ```bash
   agent engineer bd-a1b2  # ✅ Uses Beads task ID
   ```

3. **Agent behavior:**
   - Reads task details: `bd show bd-a1b2`
   - Gets description, context, dependencies, acceptance criteria
   - Checks dependencies are met
   - Marks task as started: `bd start bd-a1b2`
   - Executes the work
   - Updates task with results
   - Marks task complete: `bd close bd-a1b2`

---

## Required Changes

### 1. Update Agent CLI

**File**: `a2a-agent/cmd/agent/main.go`

**Change:**
```go
// Current (incorrect)
func main() {
    role := args[0]
    task := strings.Join(args[1:], " ")  // Free-form description
    // ...
}

// Proposed (correct)
func main() {
    role := args[0]
    taskID := args[1]  // Beads task ID (bd-xxxx)

    // Validate it's a Beads task ID
    if !strings.HasPrefix(taskID, "bd-") {
        fmt.Println("Error: Expected Beads task ID (e.g., bd-a1b2)")
        os.Exit(1)
    }

    // Build agent:// URL with task ID
    agentURL := fmt.Sprintf("agent://%s/%s", role, taskID)
    // ...
}
```

### 2. Update Agent Server

**File**: `a2a-agent/internal/server/a2a_handlers.go`

**Add Beads integration:**

```go
import (
    "os/exec"
    "encoding/json"
)

type BeadsTask struct {
    ID          string   `json:"id"`
    Title       string   `json:"title"`
    Description string   `json:"description"`
    Status      string   `json:"status"`
    Dependencies []string `json:"dependencies"`
}

func (s *AgentServer) validateBeadsTask(taskID string) error {
    if !strings.HasPrefix(taskID, "bd-") || len(taskID) <= 3 {
        return fmt.Errorf("invalid Beads task ID format: %s", taskID)
    }

    // Verify task exists by trying to get it
    _, err := s.getBeadsTask(taskID)
    if err != nil {
        return fmt.Errorf("Beads task %s does not exist: %w", taskID, err)
    }

    return nil
}

func (s *AgentServer) getBeadsTask(taskID string) (*BeadsTask, error) {
    cmd := exec.Command("bd", "show", taskID, "--json")
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get Beads task: %w", err)
    }

    var task BeadsTask
    if err := json.Unmarshal(output, &task); err != nil {
        return nil, fmt.Errorf("failed to parse Beads task: %w", err)
    }

    return &task, nil
}

func (s *AgentServer) startBeadsTask(taskID string) error {
    cmd := exec.Command("bd", "start", taskID)
    return cmd.Run()
}

func (s *AgentServer) completeBeadsTask(taskID string) error {
    cmd := exec.Command("bd", "close", taskID)
    return cmd.Run()
}
```

**Update execute handler:**

```go
func (s *AgentServer) ExecuteTaskSync(role, taskInput string) (*TaskResult, error) {
    // Validate task ID if it looks like a Beads task
    var beadsTaskID string
    var isBeadsTask bool

    if strings.HasPrefix(taskInput, "bd-") && len(taskInput) > 3 {
        // Validate the task exists
        if err := s.validateBeadsTask(taskInput); err != nil {
            return nil, fmt.Errorf("invalid Beads task: %w", err)
        }

        beadsTaskID = taskInput
        isBeadsTask = true

        // Get task details from Beads
        task, err := s.getBeadsTask(beadsTaskID)
        if err != nil {
            return nil, fmt.Errorf("failed to get Beads task: %w", err)
        }

        // Check dependencies are met
        if len(task.Dependencies) > 0 {
            for _, depID := range task.Dependencies {
                dep, err := s.getBeadsTask(depID)
                if err != nil || (dep.Status != "closed" && dep.Status != "done") {
                    return nil, fmt.Errorf("task %s has unmet dependencies: %v", beadsTaskID, task.Dependencies)
                }
            }
        }

        // Mark task as started in Beads
        if err := s.startBeadsTask(beadsTaskID); err != nil {
            return nil, fmt.Errorf("failed to start task: %w", err)
        }

        // Use task description from Beads
        taskInput = task.Title
        if task.Description != "" {
            taskInput = task.Title + "\n\n" + task.Description
        }
    }

    // Execute with task description (from Beads or free-form)
    result, err := s.executeAgentTask(role, taskInput)

    // Mark task as complete in Beads if applicable
    if err == nil && isBeadsTask {
        if err := s.completeBeadsTask(beadsTaskID); err != nil {
            // Log but don't fail
            monitoring.Logger.Warn("failed_to_complete_beads_task", "task_id", beadsTaskID, "error", err)
        }
    }

    return result, err
}
```

### 3. Update Protocol Handler

**File**: `a2a-agent/cmd/agent-server/main.go`

**Update URL handling:**

```go
func handleProtocolURL(agentURL, configPath string) {
    // Parse agent://role/bd-xxxx
    parsedURL, err := url.Parse(agentURL)
    // ...

    role := parsedURL.Host
    taskID := strings.TrimPrefix(parsedURL.Path, "/")

    // Validate Beads task ID
    if !strings.HasPrefix(taskID, "bd-") {
        fmt.Printf("❌ Invalid task ID: %s\n", taskID)
        fmt.Println("   Expected Beads task ID (e.g., bd-a1b2)")
        os.Exit(1)
    }

    // Rest of implementation...
}
```

---

## Migration Path

### Phase 1: Add Beads Support (Backward Compatible)

Support both task IDs and descriptions:

```go
// Accept either:
agent engineer bd-a1b2           // Beads task ID (preferred)
agent engineer "task description" // Free-form (deprecated)
```

### Phase 2: Require Beads Tasks

After migration period, require Beads task IDs:

```bash
agent engineer bd-a1b2  # Only this format accepted
```

---

## Documentation Updates Needed

1. **README.md** - Update quick start examples
2. **docs/content/framework/a2a-usage-guide.md** - Show Beads integration
3. **docs/content/framework/agent-to-agent.md** - Document Beads workflow
4. **a2a-agent/README.md** - Add Beads prerequisites

### Example Usage Section

```markdown
## Agent Usage with Beads

### 1. Create a task in Beads

```bash
bd create "Implement user authentication with JWT"
# Returns: bd-a1b2
```

### 2. Add task details (optional)

```bash
bd edit bd-a1b2
# Add description, acceptance criteria, notes
```

### 3. Spawn agent for the task

```bash
agent engineer bd-a1b2
```

### 4. Agent automatically:
- Reads task description from Beads
- Marks task as started (`bd start bd-a1b2`)
- Executes the work
- Marks task as complete (`bd close bd-a1b2`)

### 5. View task status

```bash
bd show bd-a1b2
bd list --status closed
```
```

---

## Benefits

### With Beads Integration ✅

- ✅ Single source of truth for tasks (Beads)
- ✅ Task status automatically tracked
- ✅ Dependency management works
- ✅ Task history preserved
- ✅ Cross-session memory
- ✅ Multi-agent coordination
- ✅ Git-backed persistence

### Without Integration ❌

- ❌ Duplicate task descriptions
- ❌ No status tracking
- ❌ No dependency management
- ❌ Lost context between sessions
- ❌ No traceability

---

## Testing Plan

### Unit Tests

```go
func TestGetBeadsTask(t *testing.T) {
    task, err := server.getBeadsTask("bd-a1b2")
    assert.NoError(t, err)
    assert.Equal(t, "bd-a1b2", task.ID)
    assert.NotEmpty(t, task.Title)
}

func TestStartBeadsTask(t *testing.T) {
    err := server.startBeadsTask("bd-a1b2")
    assert.NoError(t, err)

    // Verify status changed
    task, _ := server.getBeadsTask("bd-a1b2")
    assert.Equal(t, "started", task.Status)
}
```

### Integration Tests

```bash
# Create test task
bd create "Test task for agent integration"
TASK_ID=$(bd list --format json | jq -r '.[-1].id')

# Run agent
agent engineer $TASK_ID

# Verify task is complete
bd show $TASK_ID | grep "Status: closed"
```

---

## Implementation Checklist

- [ ] Update `cmd/agent/main.go` to accept task IDs
- [ ] Add Beads integration to `internal/server/server.go`
- [ ] Add `bd show --json` support check
- [ ] Add `bd start` integration
- [ ] Add `bd close` integration
- [ ] Add dependency checking
- [ ] Update protocol handler URL parsing
- [ ] Add unit tests for Beads integration
- [ ] Add integration tests
- [ ] Update documentation (README, usage guides)
- [ ] Add migration guide for existing users
- [ ] Update examples in all docs

---

## Related Issues

- Beads integration documented in: `quality/tooling/beads-integration.md`
- Task packet templates: `templates/task-packet/`
- Beads enforcement gate: `gates/06-beads-enforcement.md`

---

**Status**: 🔴 **NOT IMPLEMENTED**
**Next Steps**: Implement Beads integration in a2a-agent server
**Owner**: TBD
**Target Version**: 2.1.0

---

**Last Updated**: 2026-01-24
