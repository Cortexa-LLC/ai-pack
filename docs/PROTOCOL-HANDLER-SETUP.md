# Agent Protocol Handler Setup

**Purpose**: Enable `agent://` URL scheme for spawning agents from browsers and other applications.

**Format**: `agent://<role>/<task-description>`

**Example**: `agent://engineer/implement%20user%20authentication`

---

## Overview

The agent protocol handler allows you to spawn AI-Pack agents using URLs. This enables:

- Browser-based agent spawning
- Integration with task management tools
- Command-line URL handling
- Cross-application workflows

---

## URL Format

```
agent://<role>/<task-description>

Components:
  <role>         - Agent role: engineer, tester, or reviewer
  <task-description> - URL-encoded task description
```

**Examples:**

```
agent://engineer/implement%20REST%20API
agent://tester/create%20unit%20tests
agent://reviewer/security%20audit
```

---

## Platform-Specific Setup

### macOS

#### Option 1: Manual Registration (Recommended)

1. **Create Handler App** using Automator:

   a. Open Automator
   b. Create new "Application"
   c. Add "Run Shell Script" action
   d. Set shell to `/bin/bash`
   e. Paste this script:

   ```bash
   AGENT_URL="$1"
   AGENT_TYPE=$(echo "$AGENT_URL" | sed 's|agent://||' | cut -d'/' -f1)
   TASK_DESC=$(echo "$AGENT_URL" | sed 's|agent://||' | cut -d'/' -f2-)
   TASK_DESC=$(python3 -c "import urllib.parse; print(urllib.parse.unquote('$TASK_DESC'))")

   cd /path/to/ai-pack
   ./.ai-pack/bd spawn "$AGENT_TYPE" "$TASK_DESC"
   ```

   f. Save as "AI-Pack Handler.app" in Applications folder

2. **Register Protocol**:

   Create `~/Library/Preferences/com.aipack.handler.plist`:

   ```xml
   <?xml version="1.0" encoding="UTF-8"?>
   <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
   <plist version="1.0">
   <dict>
       <key>CFBundleURLTypes</key>
       <array>
           <dict>
               <key>CFBundleURLName</key>
               <string>AI-Pack Agent Protocol</string>
               <key>CFBundleURLSchemes</key>
               <array>
                   <string>agent</string>
               </array>
           </dict>
       </array>
   </dict>
   </plist>
   ```

3. **Register with Launch Services**:

   ```bash
   /System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister \
     -f "/Applications/AI-Pack Handler.app"
   ```

4. **Test**:

   ```bash
   open "agent://engineer/test%20task"
   ```

#### Option 2: Using Provided Script

The setup script creates `.ai-pack/agent_protocol_handler.sh`. To use:

1. Create app wrapper using Automator (steps 1-2 above)
2. Point to the generated script

#### Verification

```bash
# Test URL handling
open "agent://engineer/implement%20greeting%20function"

# Check if protocol is registered
/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister \
  -dump | grep agent
```

---

### Linux

#### Automatic Registration (via setup.py)

The setup script automatically registers the handler on Linux if `update-desktop-database` is available.

#### Manual Registration

1. **Handler Script** (created by setup.py):

   `.ai-pack/agent_protocol_handler.sh`

2. **Desktop Entry**:

   Create `~/.local/share/applications/ai-pack-handler.desktop`:

   ```ini
   [Desktop Entry]
   Type=Application
   Name=AI-Pack Agent Handler
   Exec=/path/to/ai-pack/.ai-pack/agent_protocol_handler.sh %u
   MimeType=x-scheme-handler/agent
   NoDisplay=true
   ```

3. **Update MIME Database**:

   ```bash
   update-desktop-database ~/.local/share/applications
   ```

4. **Set Default Handler**:

   ```bash
   xdg-mime default ai-pack-handler.desktop x-scheme-handler/agent
   ```

#### Verification

```bash
# Test URL handling
xdg-open "agent://engineer/test%20task"

# Check registration
xdg-mime query default x-scheme-handler/agent
# Should output: ai-pack-handler.desktop
```

---

### Windows

#### Manual Registry Setup

1. **Handler Script** (created by setup.py):

   `.ai-pack\agent_protocol_handler.bat`

2. **Registry Entries**:

   Create `agent-protocol.reg`:

   ```reg
   Windows Registry Editor Version 5.00

   [HKEY_CLASSES_ROOT\agent]
   @="URL:AI-Pack Agent Protocol"
   "URL Protocol"=""

   [HKEY_CLASSES_ROOT\agent\DefaultIcon]
   @="C:\\path\\to\\ai-pack\\.ai-pack\\icon.ico,1"

   [HKEY_CLASSES_ROOT\agent\shell]

   [HKEY_CLASSES_ROOT\agent\shell\open]

   [HKEY_CLASSES_ROOT\agent\shell\open\command]
   @="\"C:\\path\\to\\ai-pack\\.ai-pack\\agent_protocol_handler.bat\" \"%1\""
   ```

   **Important**: Replace `C:\\path\\to\\ai-pack` with your actual installation path.

3. **Import Registry**:

   ```cmd
   regedit /s agent-protocol.reg
   ```

#### PowerShell Script (Alternative)

```powershell
# Register agent:// protocol
$handlerPath = "C:\path\to\ai-pack\.ai-pack\agent_protocol_handler.bat"

New-Item -Path "HKCU:\Software\Classes\agent" -Force
Set-ItemProperty -Path "HKCU:\Software\Classes\agent" -Name "(Default)" -Value "URL:AI-Pack Agent Protocol"
Set-ItemProperty -Path "HKCU:\Software\Classes\agent" -Name "URL Protocol" -Value ""

New-Item -Path "HKCU:\Software\Classes\agent\shell\open\command" -Force
Set-ItemProperty -Path "HKCU:\Software\Classes\agent\shell\open\command" -Name "(Default)" -Value "`"$handlerPath`" `"%1`""
```

#### Verification

```cmd
REM Test URL handling
start agent://engineer/test%20task

REM Check registration
reg query HKEY_CLASSES_ROOT\agent
```

---

## Usage Examples

### Browser Integration

#### HTML Links

```html
<a href="agent://engineer/implement%20login%20API">
  Spawn Engineer: Implement Login API
</a>

<a href="agent://tester/create%20unit%20tests%20for%20auth">
  Spawn Tester: Create Auth Tests
</a>

<a href="agent://reviewer/security%20review%20of%20auth">
  Spawn Reviewer: Security Review
</a>
```

#### JavaScript

```javascript
// Spawn agent from JavaScript
function spawnAgent(role, task) {
  const encodedTask = encodeURIComponent(task);
  const url = `agent://${role}/${encodedTask}`;
  window.location.href = url;
}

// Usage
spawnAgent('engineer', 'implement user authentication');
spawnAgent('tester', 'create tests for auth module');
spawnAgent('reviewer', 'review authentication code');
```

### Command Line

#### macOS/Linux

```bash
# Spawn engineer
open "agent://engineer/implement%20REST%20API"     # macOS
xdg-open "agent://engineer/implement%20REST%20API" # Linux

# With dynamic task
TASK="implement user registration"
ENCODED=$(python3 -c "import urllib.parse; print(urllib.parse.unquote('$TASK'))")
open "agent://engineer/$ENCODED"
```

#### Windows

```cmd
REM Spawn engineer
start agent://engineer/implement%20REST%20API

REM With dynamic task
set TASK=implement user registration
start agent://engineer/%TASK: =%%20%
```

### Task Management Integration

#### JIRA Integration

Use custom fields or automation rules:

```
Trigger: Issue status changed to "Ready for Development"
Action: POST to webhook with agent:// URL
URL: agent://engineer/implement%20{{issue.summary}}
```

#### Notion Integration

Create database with formula field:

```
concat("agent://engineer/", replaceAll(prop("Task"), " ", "%20"))
```

#### Slack Integration

Slash command:

```
/spawn engineer implement user auth
→ Generates: agent://engineer/implement%20user%20auth
```

---

## URL Encoding

### Common Encodings

| Character | Encoded | Example |
|-----------|---------|---------|
| Space | `%20` | `user auth` → `user%20auth` |
| Slash | `%2F` | `src/auth.py` → `src%2Fauth.py` |
| Quote | `%22` | `"test"` → `%22test%22` |
| Ampersand | `%26` | `A & B` → `A%20%26%20B` |

### Python Encoding

```python
import urllib.parse

task = "implement user authentication with OAuth"
encoded = urllib.parse.quote(task)
url = f"agent://engineer/{encoded}"
```

### JavaScript Encoding

```javascript
const task = "implement user authentication with OAuth";
const encoded = encodeURIComponent(task);
const url = `agent://engineer/${encoded}`;
```

### Bash Encoding

```bash
TASK="implement user authentication with OAuth"
ENCODED=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$TASK'))")
URL="agent://engineer/$ENCODED"
```

---

## Security Considerations

### Input Validation

The protocol handler should validate:
- Role is valid (engineer, tester, reviewer)
- Task description is reasonable length (<10KB)
- No command injection attempts

### Sandboxing

Consider running agents in restricted environments:
- Limited file system access
- Network restrictions
- Timeout enforcement

### Audit Logging

Log all agent spawns via protocol handler:
- Timestamp
- Role
- Task description
- Source (if available)
- User context

---

## Troubleshooting

### Handler Not Registered

**macOS:**
```bash
# Check registration
/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister \
  -dump | grep -i agent

# Re-register
/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister \
  -f "/Applications/AI-Pack Handler.app"
```

**Linux:**
```bash
# Check MIME association
xdg-mime query default x-scheme-handler/agent

# Re-register
update-desktop-database ~/.local/share/applications
xdg-mime default ai-pack-handler.desktop x-scheme-handler/agent
```

**Windows:**
```cmd
REM Check registration
reg query HKEY_CLASSES_ROOT\agent

REM Re-import
regedit /s agent-protocol.reg
```

### URL Not Opening

1. **Check handler script permissions** (Unix):
   ```bash
   chmod +x .ai-pack/agent_protocol_handler.sh
   ```

2. **Verify paths in handler** - ensure absolute paths are correct

3. **Test handler directly**:
   ```bash
   .ai-pack/agent_protocol_handler.sh "agent://engineer/test"
   ```

### Task Not Spawning

1. **Check bd spawn works**:
   ```bash
   .ai-pack/bd spawn engineer "test task"
   ```

2. **Check task packet creation**:
   ```bash
   ls .beads/tasks/
   ```

3. **Review handler logs** - add logging to handler script

---

## Advanced Configuration

### Custom Protocol Schemes

Register additional schemes for different environments:

- `agent-dev://` - Development environment
- `agent-staging://` - Staging environment
- `agent-prod://` - Production environment

### Multi-Project Support

Include project identifier in URL:

```
agent://engineer/project-name/task-description
```

Handler extracts project and switches context:

```bash
PROJECT=$(echo "$AGENT_URL" | cut -d'/' -f2)
TASK=$(echo "$AGENT_URL" | cut -d'/' -f3-)

cd "/path/to/$PROJECT"
./.ai-pack/bd spawn engineer "$TASK"
```

### Remote Agent Spawning

For Phase 2, support remote agent spawning:

```
agent://engineer/task?remote=true&server=https://a2a-server.example.com
```

---

## Testing

### Basic Test

```bash
# macOS
open "agent://engineer/implement%20greeting%20function"

# Linux
xdg-open "agent://engineer/implement%20greeting%20function"

# Windows
start agent://engineer/implement%20greeting%20function
```

### Verify Task Creation

```bash
# Check for new task packet
ls -lt .beads/tasks/ | head -5

# View task metadata
cat .beads/tasks/task-engineer-*/00-metadata.json
```

### Integration Test

1. Create HTML test file:

```html
<!DOCTYPE html>
<html>
<head><title>Agent Protocol Test</title></head>
<body>
  <h1>AI-Pack Protocol Handler Test</h1>
  <ul>
    <li><a href="agent://engineer/implement%20test%20function">Test Engineer</a></li>
    <li><a href="agent://tester/create%20test%20suite">Test Tester</a></li>
    <li><a href="agent://reviewer/review%20code%20quality">Test Reviewer</a></li>
  </ul>
</body>
</html>
```

2. Open in browser
3. Click links
4. Verify agents spawn

---

## Future Enhancements (Phase 2)

### A2A Protocol Integration

```
agent://engineer/task?protocol=a2a&timeout=30m
```

### Streaming Progress

```
agent://engineer/task?stream=true&callback=https://example.com/progress
```

### Agent Chaining

```
agent://workflow/feature-development?steps=engineer,tester,reviewer
```

---

## Reference

**Documentation:**
- Setup Script: `setup.py`
- Usage Guide: `docs/USAGE-GUIDE.md`
- Architecture: `docs/A2A-IMPLEMENTATION-PLAN.md`

**Generated Scripts:**
- macOS/Linux: `.ai-pack/agent_protocol_handler.sh`
- Windows: `.ai-pack\agent_protocol_handler.bat`

---

**Last Updated**: 2026-01-23
**Version**: 1.0.0 (Phase 1)
