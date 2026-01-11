---
description: Check and report ai-pack system health across all components
---

# /ai-pack health - System Health Check

Comprehensive health check of the entire ai-pack system including infrastructure, roles, and agents.

## What This Checks

This command performs a full system diagnostic across:
1. **Infrastructure** - Scripts, timers, checkpoints
2. **Configuration** - Permissions, settings, hooks
3. **Roles** - Orchestrator, Coordinator, agents
4. **Workflows** - Task packets, work logs, gates

## Usage

```bash
/ai-pack health
```

Optional flags:
- `/ai-pack health --quick` - Fast check (30 seconds)
- `/ai-pack health --deep` - Comprehensive check (2-3 minutes)
- `/ai-pack health --fix` - Attempt auto-fix of issues

## Health Check Categories

### 1. Infrastructure Health

**Checks:**
- ✅ Coordination timer running
- ✅ Checkpoint file being updated
- ✅ Agent status tracker functional
- ✅ Scripts executable and present

**Verifies:**
```bash
# Timer process
ps aux | grep coordination-timer

# Checkpoint file
test -f .claude/.coordination-checkpoint && echo "OK" || echo "MISSING"

# Last update time
stat .claude/.coordination-checkpoint

# Status tracker
python3 .claude/scripts/agent-status-tracker.py report
```

**Status:**
- 🟢 GREEN - All infrastructure operational
- 🟡 YELLOW - Some components not running (non-critical)
- 🔴 RED - Critical infrastructure missing

---

### 2. Configuration Health

**Checks:**
- ✅ Permissions in settings.json
- ✅ Hooks configured
- ✅ Claude Code integration active
- ✅ Git repository initialized

**Verifies:**
```bash
# Permissions check
cat .claude/settings.json | grep -A 10 permissions

# Required permissions present
grep "Write(\*)" .claude/settings.json
grep "Edit(\*)" .claude/settings.json
grep "defaultMode" .claude/settings.json

# Hooks configured
cat .claude/settings.json | grep -A 5 hooks
```

**Status:**
- 🟢 GREEN - All configuration correct
- 🟡 YELLOW - Missing optional configuration
- 🔴 RED - Missing critical permissions

---

### 3. Role Health

**Checks:**
- ✅ Orchestrator following workflow
- ✅ Coordinator monitoring agents
- ✅ Engineers completing work
- ✅ Reviewers conducting reviews

**Verifies:**
```bash
# Check for task packets
ls -la .ai/tasks/

# Check for work logs
find .ai/tasks -name "20-work-log.md"

# Check for plan files
find .ai/tasks -name "10-plan.md"

# Check role assignments
grep -r "Orchestrator\|Coordinator\|Engineer" .ai/tasks/*/20-work-log.md
```

**Status:**
- 🟢 GREEN - All roles functioning correctly
- 🟡 YELLOW - Some roles not following workflow
- 🔴 RED - Major workflow violations

---

### 4. Agent Health

**Checks:**
- ✅ Agents making progress
- ✅ No agents stuck >5 minutes
- ✅ Work logs being updated
- ✅ Commits being made

**Verifies:**
```bash
# Agent status
python3 .claude/scripts/agent-status-tracker.py report

# Recent commits
git log --since="30 minutes ago" --oneline | wc -l

# Work log freshness
find .ai/tasks -name "20-work-log.md" -mmin -10

# Blocked agents
python3 .claude/scripts/agent-status-tracker.py report | grep "blocked"
```

**Status:**
- 🟢 GREEN - All agents healthy and progressing
- 🟡 YELLOW - Some agents slow but not blocked
- 🔴 RED - Agents blocked or stuck

---

### 5. Workflow Health

**Checks:**
- ✅ Gates being enforced
- ✅ TDD workflow followed
- ✅ Reviews happening
- ✅ Tests passing

**Verifies:**
```bash
# Check for test files created
git log --since="1 hour ago" --name-only | grep -E "test|spec"

# Check for review documents
find .ai/tasks -name "30-review.md"

# Check gate compliance
grep "GATE:" .ai/tasks/*/20-work-log.md
```

**Status:**
- 🟢 GREEN - All gates enforced, workflow followed
- 🟡 YELLOW - Some workflow shortcuts
- 🔴 RED - Critical gates skipped

---

## Health Report Format

```
============================================================
AI-PACK SYSTEM HEALTH REPORT
============================================================
Generated: 2026-01-10 15:30:00
Duration: 45 seconds

OVERALL HEALTH: 🟢 GREEN (4/5 systems healthy)

────────────────────────────────────────────────────────────
1. INFRASTRUCTURE HEALTH: 🟢 GREEN
────────────────────────────────────────────────────────────
✅ Coordination timer running (PID 12345)
✅ Checkpoint file updating (last: 30s ago)
✅ Agent status tracker functional
✅ All scripts present and executable

────────────────────────────────────────────────────────────
2. CONFIGURATION HEALTH: 🟢 GREEN
────────────────────────────────────────────────────────────
✅ Permissions configured (Write, Edit, Bash)
✅ defaultMode: acceptEdits
✅ Hooks configured
✅ Git repository initialized

────────────────────────────────────────────────────────────
3. ROLE HEALTH: 🟢 GREEN
────────────────────────────────────────────────────────────
✅ Orchestrator delegating (not implementing)
✅ Coordinator monitoring (3 check-ins in last 5 min)
✅ Engineers following TDD workflow
✅ Task packets present and maintained

────────────────────────────────────────────────────────────
4. AGENT HEALTH: 🟡 YELLOW
────────────────────────────────────────────────────────────
✅ 2/3 agents healthy and progressing
⚠️ Agent C slow (4 commits in 30 min, expected 6+)
✅ No agents blocked
✅ Work logs updated recently

Action: Monitor Agent C, may need guidance

────────────────────────────────────────────────────────────
5. WORKFLOW HEALTH: 🟢 GREEN
────────────────────────────────────────────────────────────
✅ TDD workflow followed (tests before implementation)
✅ Code quality gate enforced
✅ Review process active
✅ All tests passing

────────────────────────────────────────────────────────────
ISSUES DETECTED: 1
────────────────────────────────────────────────────────────

⚠️ YELLOW - Agent C Slow Progress
   Location: .ai/tasks/2026-01-10_api-endpoints/agent-c/
   Impact: Minor delay, not blocking
   Recommendation: Coordinator should check in with Agent C
   Auto-fix: No

────────────────────────────────────────────────────────────
RECOMMENDATIONS
────────────────────────────────────────────────────────────

1. Check in with Agent C (slow progress)
2. Continue monitoring - system healthy overall

────────────────────────────────────────────────────────────
SYSTEM STATUS: ✅ OPERATIONAL
────────────────────────────────────────────────────────────

All critical systems functioning. One agent running slow but
not blocked. Continue normal operation.

============================================================
```

## Quick Health Check Format

For `/ai-pack health --quick`:

```
AI-PACK HEALTH: 🟢 GREEN

Infrastructure: 🟢 Timer running, checkpoint OK
Configuration:  🟢 Permissions OK
Roles:          🟢 All following workflow
Agents:         🟡 1 agent slow (not critical)
Workflow:       🟢 Gates enforced

Status: Operational (1 minor issue)
```

## Auto-Fix Capability

With `/ai-pack health --fix`, the command will attempt to automatically fix common issues:

**Can auto-fix:**
- ✅ Start coordination timer if not running
- ✅ Create missing checkpoint file
- ✅ Initialize agent status tracker
- ✅ Set script permissions

**Cannot auto-fix (requires manual intervention):**
- ❌ Missing permissions in settings.json (security risk)
- ❌ Agent blockers (requires diagnosis)
- ❌ Workflow violations (requires role guidance)

## When to Run Health Check

**Regular Schedule:**
- Start of each work session
- After deploying framework updates
- Before spawning parallel agents
- When something feels "off"

**Triggered by:**
- Multiple agent failures
- Unexpected system behavior
- User request for status
- Watchdog detecting patterns

**Automatically (future):**
- Every 15 minutes during active work
- After each work package completion
- On startup of Claude Code session

## Integration with Watchdog

The health check is the primary tool for the Watchdog role:

```
User: "/ai-pack health"
  ↓
Health command runs checks
  ↓
If issues detected → Watchdog activates
  ↓
Watchdog diagnoses root cause
  ↓
Watchdog proposes framework fixes
```

**Health detects symptoms, Watchdog diagnoses diseases.**

## Troubleshooting

### Health Check Fails to Run

**Symptom:** Command errors or times out

**Check:**
```bash
# Scripts present?
ls -la .claude/scripts/

# Python3 available?
python3 --version

# Permissions correct?
ls -la .claude/scripts/*.py
```

**Fix:**
```bash
chmod +x .claude/scripts/*.py
```

---

### False Positives

**Symptom:** Reports issues that don't exist

**Cause:** Cached state or stale files

**Fix:**
```bash
# Clear cache
rm .claude/.agent-status.json
rm .claude/.coordination-last-check

# Re-run health check
/ai-pack health
```

---

### Health Check Too Slow

**Symptom:** Takes >5 minutes

**Cause:** Deep scanning of large repository

**Fix:**
- Use `--quick` flag
- Exclude large directories in .gitignore
- Run health check from specific task directory

## Related Commands

- `/ai-pack watchdog` - Activate Watchdog for deep diagnosis
- `/ai-pack agents` - Check agent status specifically
- `/ai-pack coordinate` - Manual coordination check
- `/ai-pack help` - All commands

## Implementation

The health check runs as a background task to avoid blocking the main session:

```
/ai-pack health
  ↓
Spawn background agent (Task tool, run_in_background=true)
  ↓
Agent performs Watchdog diagnostics
  ↓
Agent generates health report in .ai/reports/health-TIMESTAMP.md
  ↓
Main session notified when complete
```

**Why background execution:**
- Health checks can take 30-180 seconds depending on depth
- User should not be blocked during diagnostics
- Multiple health checks can run in parallel if needed
- Results saved to file for later review

**Implementation pattern:**
```python
Task(subagent_type="general-purpose",
     description="Run system health check",
     prompt="""Act as Watchdog role from ai-pack. Perform comprehensive health check following the diagnostic checklist in .claude/skills/watchdog/SKILL.md.

Generate report covering:
1. Infrastructure health (timers, checkpoints, scripts)
2. Configuration health (permissions, hooks, settings)
3. Role health (Orchestrator, Coordinator, agents)
4. Agent health (progress, blockers, work logs)
5. Workflow health (gates, TDD, reviews, tests)

Save report to .ai/reports/health-{timestamp}.md and provide summary.
""",
     run_in_background=true)
```

## References

- **Watchdog Skill:** [.claude/skills/watchdog/SKILL.md](../../skills/watchdog/SKILL.md)
- **Coordination Timer:** [.ai-pack/COORDINATION-TIMER.md](../../../.ai-pack/COORDINATION-TIMER.md)
- **Parallel Config:** [.ai-pack/PARALLEL-ENGINEERS-CONFIG.md](../../../.ai-pack/PARALLEL-ENGINEERS-CONFIG.md)

---

**Use this command regularly to ensure your ai-pack system is healthy and operating correctly.**
