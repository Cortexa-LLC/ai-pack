# Gate 08: Spawned Agent Permission Verification

**Priority:** CRITICAL
**Type:** BLOCKING
**Enforcement:** Pre-Spawn (before spawned agents)
**Last Updated:** 2026-01-15

---

## Purpose

Prevent silent file persistence failures by verifying Write(*) permissions are configured BEFORE spawning spawned agents that need to create or modify files.

**Problem Prevented:**
- Background agents claim success but files not persisted
- PRDs, ADRs, and other artifacts generated but not saved
- Manual extraction required from agent logs

**Production Evidence:**
- consumer-project PRD: 26KB generated, not persisted (permission failure)
- consumer-project ADRs: 6 architecture documents referenced but missing
- Multiple specialist agent failures due to missing Write(*) permission

**Reference:** TC-BA-005 (Permission Pre-Verification test case)

---

## Gate Rules

### Rule 1: Settings File Existence (BLOCKING)

**Requirement:** `.claude/settings.json` MUST exist before spawning spawned agents.

**Verification:**
```bash
if [ ! -f .claude/settings.json ]; then
  BLOCK with error: "settings.json not found"
  REQUIRE: Run setup script OR create manually
  STATUS: BLOCKED
fi
```

**Error Message:**
```
❌ GATE 08: BACKGROUND AGENT PERMISSIONS - BLOCKED

.claude/settings.json not found

PROBLEM:
Background agents need pre-configured Write(*) permissions.
Without settings.json, agents cannot persist files to repository.

SOLUTIONS:

Option 1: Use ai-pack setup script (Recommended)
  python3 .ai-pack/templates/.claude-setup.py

Option 2: Manual setup
  mkdir -p .claude
  cp .ai-pack/templates/.claude/settings.json .claude/
  Verify: python3 tests/tools/verify-background-agent-permissions.py

Option 3: Run agents in FOREGROUND
  Use interactive mode
  Agents will prompt for permissions interactively
  Trade-off: Sequential execution (no parallel)

Cannot spawn spawned agent until fixed.
```

---

### Rule 2: Write(*) Permission (BLOCKING)

**Requirement:** `Write(*)` MUST be in `permissions.allow` array.

**Verification:**
```bash
if ! grep -q '"Write(\*)"' .claude/settings.json; then
  BLOCK with error: "Write(*) not configured"
  REQUIRE: Add to permissions.allow
  STATUS: BLOCKED
fi
```

**Error Message:**
```
❌ GATE 08: BACKGROUND AGENT PERMISSIONS - BLOCKED

Write(*) not in permissions.allow

Current configuration:
{
  "permissions": {
    "allow": [
      "Edit(*)",
      "Read(*)",
      "Bash(git:*)"
      // ❌ Missing: "Write(*)"
    ]
  }
}

PROBLEM:
Background agents need Write(*) to persist files.
Without it:
- Write() calls fail silently
- Files not created
- Agent reports "success" but output missing

REQUIRED FIX:
Add "Write(*)" to permissions.allow array:

{
  "permissions": {
    "allow": [
      "Write(*)",  // ← Add this
      "Edit(*)",
      "Read(*)",
      "Bash(git:*)"
    ],
    "defaultMode": "bypassPermissions"
  }
}

EVIDENCE:
- consumer-project PRD: 26KB generated, not persisted
- Multiple specialist failures from missing permission

Cannot spawn spawned agent until Write(*) configured.
```

---

### Rule 3: defaultMode Configuration (BLOCKING)

**Requirement:** `permissions.defaultMode` MUST be `"bypassPermissions"`.

**Verification:**
```bash
if ! grep -q '"bypassPermissions"' .claude/settings.json; then
  BLOCK with error: "defaultMode not set correctly"
  REQUIRE: Set to bypassPermissions
  STATUS: BLOCKED
fi
```

**Error Message:**
```
❌ GATE 08: BACKGROUND AGENT PERMISSIONS - BLOCKED

defaultMode not set to "bypassPermissions"

Current configuration:
{
  "permissions": {
    "defaultMode": "acceptEdits"  // ❌ WRONG
  }
}

PROBLEM:
Background agents run non-interactively and cannot prompt for permissions.
With wrong defaultMode:
- Permission prompts block execution
- Agent hangs waiting for user input
- Timeout or silent failure

REQUIRED FIX:
Set defaultMode to "bypassPermissions":

{
  "permissions": {
    "allow": [
      "Write(*)",
      "Edit(*)",
      "Read(*)"
    ],
    "defaultMode": "bypassPermissions"  // ← Must be this
  }
}

WHY: Background agents must have pre-approved permissions.

Cannot spawn spawned agent until defaultMode fixed.
```

---

### Rule 4: Local Override Detection (WARNING)

**Check:** Detect if `.claude/settings.local.json` exists (may override settings.json).

**Verification:**
```bash
if [ -f .claude/settings.local.json ]; then
  WARN: "settings.local.json may override settings.json"
  RECOMMEND: Add Write(*) to local file OR remove it
  STATUS: WARNING (not blocking)
fi
```

**Warning Message:**
```
⚠️ GATE 08: BACKGROUND AGENT PERMISSIONS - WARNING

.claude/settings.local.json detected

This file may override settings.json configuration.

RECOMMENDATION:
Ensure Write(*) is also in settings.local.json:
  grep "Write(\*)" .claude/settings.local.json

OR remove settings.local.json:
  rm .claude/settings.local.json

Proceeding but verify permissions work correctly...
```

---

## Enforcement Points

### Point 1: Before Spawning Background Planning Agents

**When:** Before calling `Task(..., spawning agents)` for Cartographer, Architect, Designer, Inspector

**Check:**
```python
# Before spawning background planning agent
run_permission_verification()

def run_permission_verification():
    # Check 1: settings.json exists
    if not Path(".claude/settings.json").exists():
        raise PermissionError("Gate 08 BLOCKED: settings.json not found")

    # Check 2: Load settings
    with open(".claude/settings.json") as f:
        settings = json.load(f)

    # Check 3: Write(*) in allow list
    allow_list = settings.get("permissions", {}).get("allow", [])
    if "Write(*)" not in allow_list:
        raise PermissionError("Gate 08 BLOCKED: Write(*) not configured")

    # Check 4: defaultMode correct
    default_mode = settings.get("permissions", {}).get("defaultMode")
    if default_mode != "bypassPermissions":
        raise PermissionError("Gate 08 BLOCKED: defaultMode incorrect")

    # All checks passed
    return True
```

**Passes:** All checks pass
**Fails:** Any check fails (BLOCKING)

---

### Point 2: Before Spawning Background Worker Agents

**When:** Before calling `Task(..., spawning agents)` for Engineer/Worker roles that create files

**Check:**
```python
# Lighter check for worker agents (may not create as many files)
if will_create_files(task):
    run_permission_verification()
```

**Passes:** Verification passes OR task won't create files
**Fails:** Verification fails and task will create files (BLOCKING)

---

## Verification Utility

**Tool:** `tests/tools/verify-background-agent-permissions.py`

**Usage:**
```bash
# Verify permissions before spawning agents
python3 tests/tools/verify-background-agent-permissions.py

# From specific directory
python3 tests/tools/verify-background-agent-permissions.py --repo-root /path/to/repo
```

**Output:**
```
🔍 Checking spawned agent permissions...

✅ .claude/settings.json exists
✅ settings.json is valid JSON
✅ Write(*) permission configured
✅ Edit(*) permission configured
✅ Read(*) permission configured
✅ defaultMode: bypassPermissions
✅ Git permissions configured

============================================================
PERMISSION VERIFICATION RESULTS
============================================================

✅ ALL CHECKS PASSED

Background agents are properly configured for:
  - Writing files to repository
  - Editing existing files
  - Reading project files
  - Running git commands

Safe to spawn spawned agents with spawning agents

✅ PERMISSION VERIFICATION PASSED

Safe to spawn spawned agents!
```

---

## Required Configuration

**Minimum `.claude/settings.json`:**

```json
{
  "permissions": {
    "allow": [
      "Write(*)",         // REQUIRED for file creation
      "Edit(*)",          // RECOMMENDED for file editing
      "Read(*)",          // RECOMMENDED for reading files
      "Bash(git:*)",      // RECOMMENDED for git operations
      "Bash(npm:install)",
      "Bash(npm:test)",
      "Bash(dotnet:*)"
    ],
    "defaultMode": "bypassPermissions"  // REQUIRED for spawned agents
  }
}
```

**Why Each Permission:**

| Permission | Required? | Purpose |
|------------|-----------|---------|
| `Write(*)` | **REQUIRED** | Create new files (PRDs, ADRs, code, tests) |
| `Edit(*)` | Recommended | Modify existing files |
| `Read(*)` | Recommended | Read project files for context |
| `Bash(git:*)` | Recommended | Commit changes, check status |
| `defaultMode: bypassPermissions` | **REQUIRED** | Non-interactive permission approval |

---

## Bypass Conditions

**This gate CANNOT be bypassed.**

**Alternatives if permissions cannot be configured:**

1. **Run agents interactively:**
   ```python
   Task(
       subagent_type="general-purpose",
       description="Create PRD",
       prompt="..."
   )
   ```
   - Can prompt for permissions interactively
   - Sequential execution only
   - No configuration needed

2. **Manual artifact creation:**
   - Don't spawn agent
   - Create artifacts manually
   - Update task packet directly

**Rationale:** Spawned agents MUST have Write(*) permission. No workaround exists for spawned execution without permissions.

---

## Recovery Procedures

### Scenario 1: Agent Completed But Files Missing

**Symptoms:**
- Agent reports success
- `ls <expected-file>` → "No such file or directory"
- Agent output contains file content

**Root Cause:** Permission failure (Write(*) not configured)

**Recovery:**
```bash
# Step 1: Verify permissions
python3 tests/tools/verify-background-agent-permissions.py

# Step 2: Extract content from agent output
grep -A 1000 "Write(" <agent-output-file>

# Step 3: Manually create missing files
# (Copy content from agent output)

# Step 4: Fix permissions
# Add Write(*) to .claude/settings.json

# Step 5: Verify fix
python3 tests/tools/verify-background-agent-permissions.py
```

---

### Scenario 2: Agent Hangs or Times Out

**Symptoms:**
- Agent doesn't complete
- No output after long time
- Process appears stuck

**Root Cause:** Wrong defaultMode (agent waiting for permission prompt)

**Recovery:**
```bash
# Step 1: Kill hung agent
# (Use task manager or kill command)

# Step 2: Fix defaultMode
# Set to "bypassPermissions" in .claude/settings.json

# Step 3: Verify fix
python3 tests/tools/verify-background-agent-permissions.py

# Step 4: Re-spawn agent
# Try again with correct configuration
```

---

## Integration with Other Gates

**Related Gates:**
- **Gate 10 (Persistence):** Verifies files exist after agent completes
- **Gate 05 (Lean Flow):** Prevents overwhelming verification with too many agents

**Test Cases:**
- **TC-BA-005:** Permission Pre-Verification (validates this gate)
- **TC-BA-001:** File Persistence Verification (detects failures from missing permissions)
- **TC-BA-003:** Working Directory Context (ensures files written to correct location)

**Flow:**
```
1. Gate 08 (this gate): Verify permissions BEFORE spawning
   ↓ PASS
2. Spawn spawned agent
   ↓
3. Agent executes with Write(*) permission
   ↓
4. Gate 10: Verify files persist after completion
   ↓ PASS
5. Task complete
```

---

## Metrics

**Effectiveness Tracking:**

```markdown
## Permission Gate Metrics

**Before Gate 08:**
- Permission failures: ~40% of spawned agents
- Files not persisted: ~40%
- Manual extraction required: ~40%
- Time to discover: 10-30 minutes

**After Gate 08:**
- Permission failures: 0% (blocked before spawn)
- Files not persisted: 0% (permissions verified)
- Manual extraction required: 0%
- Time to discover: Immediate (pre-flight check)

**Improvement:**
- 100% prevention of permission failures
- 100% reduction in manual extraction
- Immediate feedback vs delayed discovery
```

---

## Examples

### Example 1: Missing Write(*) Permission

**Scenario:** Orchestrator attempts to spawn Cartographer for PRD

```python
# Orchestrator code
Task(
    subagent_type="general-purpose",
    description="Create PRD",
    prompt="""Cartographer role from .ai-pack/roles/cartographer.md

    Create Product Requirements Document for user authentication feature.
    Persist to: docs/product/2026-01-15-auth/prd.md
    """,
    spawning agents  # Background agent
)
```

**Gate 08 Response:**
```
❌ GATE 08: BACKGROUND AGENT PERMISSIONS - BLOCKED

Write(*) not in permissions.allow

Current .claude/settings.json:
{
  "permissions": {
    "allow": [
      "Edit(*)",
      "Read(*)"
    ],
    "defaultMode": "bypassPermissions"
  }
}

PROBLEM:
Cartographer needs Write(*) to create PRD.
Without it, agent will fail silently and PRD won't be created.

REQUIRED FIX:
Add "Write(*)" to permissions.allow:

{
  "permissions": {
    "allow": [
      "Write(*)",  // ← Add this
      "Edit(*)",
      "Read(*)"
    ],
    "defaultMode": "bypassPermissions"
  }
}

Save changes and retry.

Cannot spawn Cartographer until Write(*) configured.
```

---

### Example 2: Wrong defaultMode

**Scenario:** Background agent for architecture design

```python
Task(
    subagent_type="general-purpose",
    description="Create architecture design",
    prompt="Architect role. Create ADRs...",
    spawning agents
)
```

**Gate 08 Response:**
```
❌ GATE 08: BACKGROUND AGENT PERMISSIONS - BLOCKED

defaultMode not set to "bypassPermissions"

Current: "defaultMode": "acceptEdits"

PROBLEM:
Background agents run non-interactively.
With "acceptEdits" mode:
- Agent will prompt for Write permission
- No user to respond to prompt
- Agent hangs or times out

REQUIRED FIX:
{
  "permissions": {
    "allow": ["Write(*)", "Edit(*)", "Read(*)"],
    "defaultMode": "bypassPermissions"  // ← Change this
  }
}

Cannot spawn Architect until defaultMode fixed.
```

---

### Example 3: All Checks Pass

**Scenario:** Properly configured permissions

```bash
# settings.json correctly configured
{
  "permissions": {
    "allow": [
      "Write(*)",
      "Edit(*)",
      "Read(*)",
      "Bash(git:*)"
    ],
    "defaultMode": "bypassPermissions"
  }
}
```

**Gate 08 Response:**
```
✅ GATE 08: BACKGROUND AGENT PERMISSIONS - PASSED

All permission checks passed:
✅ .claude/settings.json exists
✅ Write(*) configured
✅ Edit(*) configured
✅ Read(*) configured
✅ defaultMode: bypassPermissions

Safe to spawn spawned agents!

Proceeding with agent spawn...
```

---

## Troubleshooting

### Issue 1: Permission Check Script Not Found

**Error:** `python3 tests/tools/verify-background-agent-permissions.py: No such file`

**Solution:**
```bash
# Verify ai-pack submodule is initialized
git submodule update --init --recursive

# Or download script directly
curl -O https://raw.githubusercontent.com/.../verify-background-agent-permissions.py
python3 verify-background-agent-permissions.py
```

---

### Issue 2: JSON Syntax Error in settings.json

**Error:** `settings.json is invalid JSON`

**Solution:**
```bash
# Validate JSON
python3 -m json.tool .claude/settings.json

# Common issues:
# - Trailing commas
# - Missing quotes
# - Unescaped backslashes

# Fix and retry
```

---

### Issue 3: Permissions Work Locally But Fail in CI/CD

**Problem:** Different settings in CI environment

**Solution:**
```bash
# CI/CD must also have .claude/settings.json
# Option 1: Copy template in CI script
cp .ai-pack/templates/.claude/settings.json .claude/

# Option 2: Use environment variable
export CLAUDE_BYPASS_PERMISSIONS=true
```

---

## Success Criteria

**This gate is successful when:**

✅ Zero permission failures in spawned agents
✅ 100% of artifacts persist when agents claim success
✅ No manual extraction required
✅ Immediate feedback when permissions missing
✅ Clear remediation guidance provided

**Leading Indicators:**
- Permission verification runs before every background spawn
- Developers proactively configure settings.json
- Permission-related failures drop to zero
- No production failures from missing Write(*) permission

---

## References

**Test Cases:**
- TC-BA-005: Permission Pre-Verification (validates this gate)
- TC-BA-001: File Persistence Verification (detects failures)

**Tools:**
- `tests/tools/verify-background-agent-permissions.py`

**Documentation:**
- `roles/orchestrator.md` - Section 2.14 (Spawned Agent Permission Verification)
- `templates/.claude/settings.json` - Template configuration
- `templates/.claude/PERMISSIONS.md` - Permission documentation

**Production Failures:**
- consumer-project PRD: 26KB generated, not persisted
- consumer-project ADRs: 6 documents missing
- Multiple specialist agent failures

**Related Gates:**
- Gate 10: Persistence Gate (artifact verification)
- Gate 05: Lean Flow (WIP limits)

---

**Version:** 1.0.0
**Status:** Active - MANDATORY
**Enforcement:** BLOCKING on violations
