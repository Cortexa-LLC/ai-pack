# Multi-Project Beads Support

## Problem

The agent-server previously could only support a single project's `.beads/` directory. When running agents from different project directories, the Beads validation would fail because the server was looking for `.beads/` in its own working directory rather than the project's directory.

## Solution

Implemented multi-project support by:

1. **Client-side detection**: The `agent` CLI now detects the project root where it's invoked
2. **Protocol extension**: Added `project_root` field to `ExecuteTaskRequest`
3. **Server-side support**: Server uses the project root when executing Beads commands
4. **Beads client enhancement**: Added `*FromDir()` variants of all Beads operations

## Changes Made

### Protocol (`internal/protocol/a2a.go`)
- Added `ProjectRoot string` field to `ExecuteTaskRequest`

### Beads Client (`internal/beads/beads.go`)
Added working directory support to all operations:
- `GetTaskFromDir(taskID, workingDir)`
- `StartTaskFromDir(taskID, workingDir)`
- `CompleteTaskFromDir(taskID, workingDir)`
- `CheckDependenciesFromDir(taskID, workingDir)`
- `ValidateTaskIDFromDir(taskID, workingDir)`
- `GetTaskDescriptionFromDir(input, projectRoot)`

### Agent CLI (`cmd/agent/main.go`)
- Added `detectProjectRoot()` function that uses:
  1. `git rev-parse --show-superproject-working-tree` (for submodules)
  2. `git rev-parse --show-toplevel` (for regular repos)
  3. Current working directory (fallback)
- Updated `handleSpawn()` to detect and send project root

### Server (`internal/server/server.go`)
- Updated `spawnAgentTask()` signature to accept `projectRoot` parameter
- Uses `projectRoot` for all Beads operations
- Stores `project_root` in task metadata
- Updated `completeBeadsTask()` to use project root from metadata

### A2A Handlers (`internal/server/a2a_handlers.go`)
- Updated to pass `project_root` from request to `spawnAgentTask()`

### Public API (`internal/server/public_api.go`)
- Updated legacy methods to pass empty string for `projectRoot` (uses server root as fallback)

## Testing

### Before Fix
```bash
# Server running from ai-pack/a2a-agent/
cd /path/to/project
agent engineer xasm++-asp --stream
# ❌ Error: task xasm++-asp does not exist or is not accessible
```

### After Fix
```bash
# Server can run from anywhere
cd /path/to/project
agent engineer xasm++-asp --stream
# ✅ Works! Uses project's .beads/ directory
```

## How It Works

1. User runs `agent engineer xasm++-asp` from `/path/to/project`
2. CLI detects project root: `/path/to/project`
3. CLI sends request with `project_root: "/path/to/project"`
4. Server receives request and uses project root for all Beads commands:
   - `agent show xasm++-asp` (with working dir = `/path/to/project`)
   - `agent update --claim xasm++-asp` (with working dir = `/path/to/project`)
   - `agent close xasm++-asp` (with working dir = `/path/to/project`)

## Task Storage Fix (Critical Update)

**Issue Discovered:** Initial implementation stored task metadata in server's `.beads/tasks/` directory, polluting the ai-pack project.

**Solution:** Task metadata is now stored in the **project's** `.beads/tasks/` directory where the work is being done.

**Changes:**
- Added `ProjectRoot` field to `TaskExecution` struct
- Updated all file operations to use `execution.ProjectRoot` instead of `s.rootDir`
- Task packets created in: `{projectRoot}/.beads/tasks/{taskID}/`

See [PROJECT_ROOT_STORAGE.md](PROJECT_ROOT_STORAGE.md) for complete details.

## Benefits

- ✅ Single agent-server supports multiple projects
- ✅ Server can run from any directory
- ✅ Each project's `.beads/` directory is isolated
- ✅ Task metadata stored in project, not server directory
- ✅ No cross-project pollution
- ✅ Works with git submodules (detects superproject)
- ✅ Backward compatible (defaults to server root if no project_root specified)

## Deployment

1. Rebuild binaries: `cd a2a-agent && make`
2. Restart agent-server: `pkill agent-server && agent-server --server &`
3. No configuration changes needed - automatic detection

## Future Enhancements

- Consider caching project root detection for performance
- Add explicit `--project-root` flag override to CLI
- Support non-git project root detection (.beads/ ancestor search)
