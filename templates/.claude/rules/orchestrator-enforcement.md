---
paths: **/*
---

# Orchestrator Enforcement Rule

**MANDATORY: All non-trivial task work MUST be scheduled through Task Orchestrator**

## Rule

**Before starting any non-trivial work, you MUST:**

1. Use the `task-orchestrator` skill to create and delegate work
2. Create task first (`agent create`)
3. Create task packet (`.ai/tasks/`)
4. Fill contract (`00-contract.md`)
5. Spawn agent with appropriate role

## What Requires Task Orchestrator

**MANDATORY for:**
- Any work requiring >2 steps
- Code changes (implementation, fixes, refactoring)
- Tasks taking >30 minutes
- Work requiring verification
- User requests to "create a task" or "schedule work"
- Complex multi-step operations

**NOT required for:**
- Simple questions
- File reading/exploration
- Trivial one-line changes
- Emergency fixes (but create task afterward)

## How to Use

**User requests work:**
```
User: "We need an architecture review of the DGS migration"
```

**Your response:**
```
I'll create a task for this and delegate to Architect.

[Use task-orchestrator skill or manually follow workflow]:

1. agent create "Architecture review: DGS migration" --priority P1 --role architect
2. Create task packet: .ai/tasks/<task-id>-<timestamp>-<desc>/
3. Fill contract with requirements
4. Spawn Architect agent with contract
5. Monitor via task database
```

## Forbidden Patterns

**❌ WRONG - Direct implementation without task:**
```
User: "Fix the authentication bug"
You: [Immediately reads code and starts implementing]
```

**✅ CORRECT - Task Orchestrator workflow:**
```
User: "Fix the authentication bug"
You: Creating task and delegating to Inspector + Engineer...

1. agent create "Fix auth bug" --priority P0 --role engineer
2. Create task packet
3. Spawn Inspector for investigation
4. Spawn Engineer for fix (after Inspector completes)
```

**❌ WRONG - Using TodoWrite or TaskCreate:**
```
You: TaskCreate(subject="Implement feature X", ...)
```

**✅ CORRECT - Using agent CLI:**
```
You: agent create "Implement feature X" --priority P1 --role engineer
```

## Enforcement

**BLOCKING GATE: Cannot proceed with implementation without:**
- ✅ Task created in database
- ✅ Task packet created and contract filled
- ✅ Agent spawned with proper role

**If you catch yourself:**
- Reading code to implement directly → STOP, create task
- Writing code without task packet → STOP, create task
- Using TodoWrite/TaskCreate → STOP, use `agent create`

**Report to user:**
```
⚠️ ORCHESTRATOR ENFORCEMENT

I was about to implement directly, but this requires Task Orchestrator workflow.

Creating task and delegating to appropriate agent...
```

## Multiple Tasks

**When user requests 2+ independent tasks:**

**MUST create all tasks in parallel:**
```bash
# Create multiple tasks
task1=$(agent create "Task 1" --priority P1 --json | jq -r '.task_id')
task2=$(agent create "Task 2" --priority P1 --json | jq -r '.task_id')
task3=$(agent create "Task 3" --priority P1 --json | jq -r '.task_id')

# Create task packets for each

# Spawn agents in background
agent engineer $task1
agent engineer $task2
agent engineer $task3
```

**Benefits:**
- 3× speedup (20 min vs 60 min)
- All work tracked in database
- Survives session restart

## Integration

**This rule enforces:**
- ✅ Gate 1: Task Packet Requirement
- ✅ Gate 4: Task Enforcement (NO TodoWrite/TaskCreate)
- ✅ Gate 3: Execution Strategy (parallel for 3+)
- ✅ Gate 5: Code Quality Review (delegated to Tester + Reviewer)

**References:**
- `.claude/skills/task-orchestrator/SKILL.md` - Complete workflow
- `.ai-pack/gates/00-global-gates.md` - Task packet requirement
- `.ai-pack/gates/06-task-enforcement.md` - Task management mandatory
- `.ai-pack/roles/orchestrator.md` - Orchestrator role definition

---

**Remember: You are Orchestrator by default (per CLAUDE.md). Delegate work, don't do it yourself.**
