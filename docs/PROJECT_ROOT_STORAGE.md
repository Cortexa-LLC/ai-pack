# Project-Root-Based Task Storage

## Problem Identified
The server was storing task metadata in its own `.beads/tasks/` directory, polluting the ai-pack project with execution metadata from completely unrelated projects.

## Solution Implemented
Task metadata is now stored in the **project's** `.beads/tasks/` directory, not the server's directory.

## Changes Made

### 1. TaskExecution Structure
- Added `ProjectRoot string` field to track which project the task belongs to

### 2. Task Storage Location
**Before:**
```
Server at: /Users/bryanw/Projects/Vibe/ai-pack/a2a-agent/
Tasks at:  /Users/bryanw/Projects/Vibe/ai-pack/a2a-agent/.beads/tasks/
```

**After:**
```
Server at:    /Users/bryanw/Projects/Vibe/ai-pack/a2a-agent/ (doesn't matter!)
Project at:   /Users/bryanw/Projects/Vintage/tools/xasm++/
Tasks stored: /Users/bryanw/Projects/Vintage/tools/xasm++/.beads/tasks/
```

### 3. Updated Methods
- `createTaskPacketInProject(taskID, role, task, config, projectRoot)`
  Creates task directory in project's .beads/tasks/

- `updateTaskPacketMetadataInProject(taskID, metadata, projectRoot)`
  Updates metadata in project's task directory

- `setupExecutionLogger(execution)`
  Creates execution.log in project's task directory

- `saveTaskResults(execution, result, logMsg)`
  Saves results in project's task directory

- `failTask(execution, errorMsg)`
  Logs failures to project's task directory

### 4. Client Updates
- Spawn response now returns internal task ID directly
- Client uses internal task ID for streaming/status instead of searching metadata

## Remaining Issues

### ⚠️ Status Query Problem
When querying status for a completed task (not in memory), the server doesn't know which project directory to search:

```go
func (s *AgentServer) getTaskStatus(taskID string) (*protocol.TaskStatusResponse, error) {
    // If task not in activeTasks memory...
    // How do we know which project's .beads/tasks/ to search?
    return s.loadTaskStatusFromDisk(taskID)  // Where to look?
}
```

### Possible Solutions

**Option A: Task Registry File**
Store a mapping file in server's directory:
```json
{
  "task-engineer-20260128-062017": "/Users/bryanw/Projects/Vintage/tools/xasm++",
  "task-engineer-20260128-063421": "/Users/bryanw/Projects/OtherProject"
}
```

**Option B: Search Strategy**
1. Check activeTasks first (has ProjectRoot)
2. If not found, return "task not found" (completed tasks require project context)
3. Client should track which project a task belongs to

**Option C: Include in Metadata Response**
When task is spawned, client receives project_root in response and must pass it back for status queries.

**Recommendation:** Option A (Task Registry) is most robust for multi-project scenarios.

## Testing

```bash
# 1. Start server (from anywhere)
cd /Users/bryanw/Projects/Vibe/ai-pack/a2a-agent
./bin/agent-server --server &

# 2. Run agent from project directory
cd /Users/bryanw/Projects/Vintage/tools/xasm++
agent engineer xasm++-asp --stream

# 3. Verify task storage location
ls -la /Users/bryanw/Projects/Vintage/tools/xasm++/.beads/tasks/
# Should see: task-engineer-YYYYMMDD-HHMMSS-NNNNNN/

# 4. Verify server directory is NOT polluted
ls -la /Users/bryanw/Projects/Vibe/ai-pack/a2a-agent/.beads/tasks/
# Should be empty or only contain old tasks
```

## Benefits

- ✅ Server can run from any directory
- ✅ Each project's tasks are isolated in that project's `.beads/` directory
- ✅ No cross-project pollution
- ✅ Task metadata travels with the project in git
- ✅ Multiple projects can use same server instance
- ✅ Server location is irrelevant to task storage

## Migration

Old tasks in server's `.beads/tasks/` directory will still work but won't be accessible via new project-based storage. Clean up manually if needed.
