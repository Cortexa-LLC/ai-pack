# Coordination Timer System

**Version:** 1.0.0
**Status:** ACTIVE

---

## Overview

The coordination timer system solves the challenge that **Claude cannot trigger time-based events**. Since AI cannot "wake up" after N minutes, we use a **bash background process** to update a checkpoint file that Claude can monitor.

**Key Insight:** Claude can't track time, but it CAN read files updated by bash timers.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ Orchestrator spawns parallel agents                          │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Start coordination-timer.sh in background                    │
│ bash .claude/scripts/coordination-timer.sh 30 1200 &         │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Timer loop (every 30 seconds):                               │
│ 1. Sleep 30 seconds                                           │
│ 2. Increment checkpoint count                                │
│ 3. Update .claude/.coordination-checkpoint file               │
│ 4. Write: count, timestamp, message                          │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Coordinator periodically checks:                             │
│ - Read .claude/.coordination-checkpoint                      │
│ - Compare to last processed checkpoint                       │
│ - If new checkpoint → trigger coordination actions           │
└─────────────────────────────────────────────────────────────┘
```

---

## Components

### 1. coordination-timer.sh

**Location:** `.claude/scripts/coordination-timer.sh`

**Purpose:** Background bash process that updates checkpoint file on schedule

**Usage:**
```bash
bash .claude/scripts/coordination-timer.sh [interval_seconds] [max_checks]

# Default: 30 seconds, 1200 checks (10 hours)
bash .claude/scripts/coordination-timer.sh 30 1200 &
```

**What it does:**
- Runs in background (note the `&`)
- Every N seconds: increments counter, writes checkpoint file
- Continues until max_checks reached or killed
- Creates `.claude/.coordination-checkpoint` with:
  - Line 1: Checkpoint count (integer)
  - Line 2: Timestamp (YYYY-MM-DD HH:MM:SS)
  - Line 3: Message ("⏰ Coordination checkpoint N triggered...")

**Example checkpoint file:**
```
3
2026-01-10 14:45:00
⏰ Coordination checkpoint 3 triggered at 2026-01-10 14:45:00
```

### 2. check-coordination-trigger.py

**Location:** `.claude/scripts/check-coordination-trigger.py`

**Purpose:** Checks if new checkpoint reached and provides actionable guidance

**Usage:**
```bash
python3 .claude/scripts/check-coordination-trigger.py
```

**Exit codes:**
- `0` - No coordination needed (silent success)
- `1` - Coordination required (prints alert)

**Output when triggered:**
```
============================================================
🚨 COORDINATION CHECK-IN REQUIRED
============================================================
Checkpoint: 3
Timestamp: 2026-01-10 14:45:00
Message: ⏰ Coordination checkpoint 3 triggered

============================================================
ACTION REQUIRED:
1. Check agent status: /ai-pack agents
2. Review work logs: Read .ai/tasks/*/20-work-log.md
3. Check git activity: git log -5 --oneline
4. Identify any blocked/stuck agents
5. Intervene if needed
6. Report status
============================================================
```

**How it works:**
- Reads `.claude/.coordination-checkpoint` (current count)
- Reads `.claude/.coordination-last-check` (last processed count)
- If current > last: New checkpoint reached
- Updates last-check file
- Prints alert and exits with code 1

### 3. Integration with Orchestrator

**When:** Orchestrator spawns 2+ parallel agents

**Action:** Start coordination timer

**Code in Orchestrator skill:**
```bash
# After spawning parallel agents
bash .claude/scripts/coordination-timer.sh 30 1200 &
```

**Effect:**
- Timer starts in background
- Checkpoint file created immediately
- Updates every 30 seconds
- Orchestrator can now delegate coordination to Coordinator role

### 4. Integration with Coordinator

**When:** Coordinator role is active (monitoring parallel agents)

**Action:** Check coordination trigger periodically

**Code in Coordinator skill:**
```bash
# Check if coordination checkpoint reached
python3 .claude/scripts/check-coordination-trigger.py
```

**If triggered:**
1. Run coordination check-in process
2. Check agent status
3. Read work logs
4. Check git activity
5. Identify blockers
6. Intervene if needed
7. Report status

---

## Configuration

### Default Settings

```bash
INTERVAL_SECONDS=30      # Check every 30 seconds
MAX_CHECKS=1200          # Run for 10 hours (1200 * 30s = 36000s = 10h)
```

### Custom Intervals

For different monitoring frequencies:

```bash
# Ultra-responsive (every 10 seconds, 6 hours)
bash .claude/scripts/coordination-timer.sh 10 2160 &

# Standard (every 30 seconds, 10 hours) - DEFAULT
bash .claude/scripts/coordination-timer.sh 30 1200 &

# Conservative (every 60 seconds, 12 hours)
bash .claude/scripts/coordination-timer.sh 60 720 &

# Long-running (every 2 minutes, 24 hours)
bash .claude/scripts/coordination-timer.sh 120 720 &
```

### Stopping the Timer

```bash
# Find the timer process
ps aux | grep coordination-timer

# Kill it
pkill -f coordination-timer.sh

# Or kill by PID
kill <PID>
```

### Cleanup

```bash
# Remove checkpoint files
rm .claude/.coordination-checkpoint
rm .claude/.coordination-last-check
```

---

## Workflow

### Scenario: Orchestrator Spawns 3 Engineers

**T+0: Orchestration Phase**
```bash
Orchestrator: I'm delegating to 3 Engineers for parallel implementation

# Spawn agents
Task(...) # Engineer A
Task(...) # Engineer B
Task(...) # Engineer C

# Start coordination timer
bash .claude/scripts/coordination-timer.sh 30 1200 &

Timer started: 30-second checkpoints for 10 hours
```

**T+30s: First Checkpoint**
```
Timer updates: .claude/.coordination-checkpoint
Count: 1
Timestamp: 14:30:30

Coordinator checks: python3 .claude/scripts/check-coordination-trigger.py
Result: ✅ Checkpoint 1 reached

Coordinator actions:
1. Check agent status
2. All 3 agents active, making progress
3. No intervention needed
4. Report: "All agents healthy"
```

**T+60s: Second Checkpoint**
```
Timer updates: Count 2, Timestamp 14:31:00

Coordinator checks: Checkpoint 2 reached

Coordinator actions:
1. Check git log: 2 recent commits (Agents A and B)
2. Check work logs: All agents progressing
3. Agent C: No commits yet (started 60s ago, still analyzing)
4. Assessment: Normal, no intervention needed
```

**T+90s: Third Checkpoint**
```
Timer updates: Count 3, Timestamp 14:31:30

Coordinator checks: Checkpoint 3 reached

Coordinator actions:
1. Git log: 4 commits total (good progress)
2. Work logs: Agents A and B ahead, Agent C catching up
3. Assessment: All healthy
```

**T+180s: Sixth Checkpoint**
```
Timer updates: Count 6, Timestamp 14:33:00

Coordinator checks: Checkpoint 6 reached

Coordinator actions:
1. Git log: 6 commits from A and B, still 0 from C
2. Work logs: Agent C stuck on permission error
3. 🚨 INTERVENTION NEEDED

Coordinator: Agent C, I see you're blocked on permission denied.
Checking settings.json... permissions missing!
Adding permissions now...
Please retry your operations.
```

**T+210s: Seventh Checkpoint**
```
Timer updates: Count 7, Timestamp 14:33:30

Coordinator checks: Checkpoint 7 reached

Coordinator actions:
1. Git log: Agent C now has 2 commits (unblocked!)
2. Work logs: All 3 agents progressing normally
3. Assessment: Issue resolved, all healthy
```

**T+600s: Twentieth Checkpoint (10 minutes elapsed)**
```
Timer updates: Count 20, Timestamp 14:40:00

Coordinator checks: Checkpoint 20 reached

Coordinator actions:
1. Git log: All agents completed work (12 commits total)
2. Work logs: All acceptance criteria met
3. Tests: All passing
4. Assessment: ✅ Work package complete

Coordinator: All agents completed successfully.
Stopping coordination timer...
pkill -f coordination-timer.sh
```

---

## Monitoring Thresholds

### Response Times

Based on 30-second checkpoints:

| Silence Duration | Checkpoints | Assessment | Action |
|-----------------|-------------|------------|--------|
| 0-60s | 1-2 | Normal | Monitor only |
| 60-120s | 2-4 | Acceptable | Check work log |
| 120-180s | 4-6 | Concerning | Investigate |
| 180-300s | 6-10 | Stuck | Intervene |
| 300s+ | 10+ | Blocked | Immediate action |

### Coordinator Responsiveness

```
Every 30 seconds:
→ Check coordination-checkpoint file
→ If new checkpoint: Run full coordination check

Every 2-4 checkpoints (1-2 minutes):
→ Quick status review
→ Look for obvious blockers

Every 6-10 checkpoints (3-5 minutes):
→ Deep investigation if no progress
→ Proactive intervention
```

---

## Benefits

### 1. Real-Time Monitoring

**Before (manual checks):**
- Coordinator had to remember to check
- No schedule, inconsistent monitoring
- Agents could be stuck for 10+ minutes unnoticed

**After (timer system):**
- Automatic 30-second checkpoints
- Consistent, reliable monitoring
- Issues detected within 1-2 minutes

### 2. Non-Intrusive

- Background process uses minimal resources
- Doesn't interrupt agent work
- Silent until checkpoint reached

### 3. Flexible Configuration

- Adjust interval: 10s to 2min
- Adjust duration: 1 hour to 24+ hours
- Stop/restart anytime

### 4. Audit Trail

Checkpoint file provides:
- Timestamp of each check
- Count of coordination events
- History of when checks occurred

---

## Troubleshooting

### Timer Not Starting

**Symptom:** No `.claude/.coordination-checkpoint` file

**Check:**
```bash
ps aux | grep coordination-timer
```

**Fix:**
```bash
# Start manually
bash .claude/scripts/coordination-timer.sh 30 1200 &

# Verify file created
ls -la .claude/.coordination-checkpoint
```

### Checkpoint Not Updating

**Symptom:** File exists but count not incrementing

**Check:**
```bash
# Is timer process still running?
ps aux | grep coordination-timer

# Check file modification time
stat .claude/.coordination-checkpoint
```

**Fix:**
```bash
# Kill old timer
pkill -f coordination-timer

# Start fresh
bash .claude/scripts/coordination-timer.sh 30 1200 &
```

### Check Script Not Detecting New Checkpoints

**Symptom:** File updates but script says "no new checkpoint"

**Check:**
```bash
# Current checkpoint
cat .claude/.coordination-checkpoint

# Last processed
cat .claude/.coordination-last-check
```

**Fix:**
```bash
# Reset last-check counter
echo "0" > .claude/.coordination-last-check

# Next check will trigger
python3 .claude/scripts/check-coordination-trigger.py
```

---

## Limitations

### 1. Requires Bash

- Won't work on systems without bash
- Windows: Use WSL or Git Bash

### 2. Background Process Management

- Process continues after Claude session ends
- Must manually kill if needed
- Could accumulate orphaned processes

**Mitigation:**
```bash
# Add to cleanup script
pkill -f coordination-timer
```

### 3. File I/O Overhead

- Every 30 seconds: Write checkpoint file
- Minimal but not zero overhead
- Not suitable for extremely resource-constrained environments

### 4. No Automatic Stop

- Timer runs for max_checks regardless
- Doesn't auto-stop when agents complete
- Coordinator should kill timer when done

---

## Future Enhancements

### Possible Improvements

1. **Auto-stop on completion**
   - Detect when all agents finished
   - Automatically kill timer process

2. **Adaptive intervals**
   - Start fast (30s), slow down if agents stable
   - Speed up if issues detected

3. **Rich checkpoint data**
   - Include agent status in checkpoint file
   - Add git commit count
   - Add work log excerpts

4. **Multiple timers**
   - Different intervals for different checks
   - Fast timer (30s) for agent health
   - Slow timer (5min) for quality checks

5. **Hook integration**
   - Trigger hook on checkpoint
   - Automatic coordination activation

---

## References

- **Orchestrator Skill:** [templates/.claude/skills/orchestrator/SKILL.md](../templates/.claude/skills/orchestrator/SKILL.md)
- **Coordinator Skill:** [templates/.claude/skills/coordinator/SKILL.md](../templates/.claude/skills/coordinator/SKILL.md)
- **Parallel Engineers Config:** [PARALLEL-ENGINEERS-CONFIG.md](PARALLEL-ENGINEERS-CONFIG.md)

---

**Made with ❤️ by Cortexa LLC**

*Enabling time-based coordination for AI agents*
