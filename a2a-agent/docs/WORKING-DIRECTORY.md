# Working Directory Requirements

## Critical: Server Must Run from Project Root

The A2A agent server **must be started from your project root directory**, not from the `a2a-agent` subdirectory.

### ✅ Correct: Start from Project Root

```bash
cd /path/to/your/project
.ai-pack/a2a-agent/bin/agent-server -server
```

**Result:** Agents will execute in `/path/to/your/project` and can access your project files.

### ❌ Incorrect: Starting from a2a-agent Directory

```bash
cd /path/to/your/project/.ai-pack/a2a-agent
./bin/agent-server -server
```

**Result:** Agents will execute in `/path/to/your/project/.ai-pack/a2a-agent` and **cannot find your project files**.

## How It Works

1. The server determines the working directory using `os.Getwd()` at startup
2. This directory is stored as `rootDir` in the server
3. All spawned agents receive this `rootDir` in their prompt
4. Agents are instructed to perform all file operations relative to this directory

## Verification

When the server starts, check the startup logs:

```
📂 Working Directory: /Users/yourname/Projects/your-project
   ⚠️  All spawned agents will execute in this directory
```

**Verify this path is your project root**, not a subdirectory!

## Common Mistake

If agents report errors like:
- "File not found"
- "Cannot read file at path X"
- "Directory does not exist"

Check that:
1. The server was started from the **project root**
2. The working directory in the logs is correct
3. The paths the agent is trying to access are relative to that directory

## Startup Scripts

If using a startup script, ensure it changes to the project root first:

### Shell Script Example

```bash
#!/bin/bash
# Good: Change to project root first
cd /path/to/your/project
.ai-pack/a2a-agent/bin/agent-server -server
```

### Python Script Example

```python
import os
import subprocess

# Change to project root
project_root = "/path/to/your/project"
os.chdir(project_root)

# Start server
subprocess.run([".ai-pack/a2a-agent/bin/agent-server", "-server"])
```

## Agent Perspective

When an agent is spawned, it receives a prompt that includes:

```
**Working Directory:**
/path/to/your/project

**IMPORTANT:** All file operations (Read, Write, Edit, Glob, Grep, Bash)
must be performed relative to the working directory above.
```

The agent is explicitly instructed to use this directory for all operations.

## Troubleshooting

### Problem: Agent can't find files

**Symptom:** Agent reports file not found errors even though files exist.

**Solution:**
1. Stop the server
2. `cd` to your project root
3. Start the server again from there
4. Verify the "Working Directory" in logs

### Problem: Agent creates files in wrong location

**Symptom:** Agent creates files in `.ai-pack/a2a-agent/` instead of project root.

**Solution:** Same as above - restart server from project root.

## Best Practice

Add a helper script at your project root:

**`start-agent-server.sh`:**
```bash
#!/bin/bash
# Ensures server always starts from project root
cd "$(dirname "$0")"  # Change to script directory (project root)
.ai-pack/a2a-agent/bin/agent-server -server "$@"
```

Usage:
```bash
./start-agent-server.sh
```

This guarantees the server always runs from the correct directory.
