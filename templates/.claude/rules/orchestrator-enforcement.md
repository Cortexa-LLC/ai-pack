---
paths: **/*
---

# Orchestrator Enforcement Rule

**MANDATORY: All non-trivial task work MUST be scheduled through Task Orchestrator**

## Rule

**Before starting any non-trivial work, you MUST:**

1. Use the `task-orchestrator` skill to create and delegate work
2. Create Beads task first (`bd create`)
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

1. bd create --title="Architecture review: DGS migration" --description="..." --type=task --priority=2
2. Create task packet: .ai/tasks/<beads-id>-<timestamp>-<desc>/
3. Fill contract with requirements
4. Spawn Architect agent with contract
5. Monitor via Beads
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

1. bd create --title="Fix auth bug" --type=bug --priority=1
2. Create task packet
3. Spawn Inspector for investigation
4. Spawn Engineer for fix (after Inspector completes)
```

**❌ WRONG - Using TodoWrite or TaskCreate:**
```
You: TaskCreate(subject="Implement feature X", ...)
```

**✅ CORRECT - Using Beads:**
```
You: bd create --title="Implement feature X" --type=feature --priority=2
```

## Enforcement

**BLOCKING GATE: Cannot proceed with implementation without:**
- ✅ Beads task created
- ✅ Task packet created and contract filled
- ✅ Agent spawned with proper role

**If you catch yourself:**
- Reading code to implement directly → STOP, create task
- Writing code without task packet → STOP, create task
- Using TodoWrite/TaskCreate → STOP, use `bd create`

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
# Create multiple Beads tasks
task1=$(bd create --title="Task 1" --json | jq -r '.id')
task2=$(bd create --title="Task 2" --json | jq -r '.id')
task3=$(bd create --title="Task 3" --json | jq -r '.id')

# Create task packets for each

# Spawn agents in parallel (single response)
Agent(...)  # Task 1
Agent(...)  # Task 2
Agent(...)  # Task 3
```

**Benefits:**
- 3× speedup (20 min vs 60 min)
- All work tracked in Beads
- Survives session restart

## Integration

**This rule enforces:**
- ✅ Gate 1: Task Packet Requirement
- ✅ Gate 4: Beads Enforcement (NO TodoWrite/TaskCreate)
- ✅ Gate 3: Execution Strategy (parallel for 3+)
- ✅ Gate 5: Code Quality Review (delegated to Tester + Reviewer)

**References:**
- `.claude/skills/task-orchestrator/SKILL.md` - Complete workflow
- `.ai-pack/gates/00-global-gates.md` - Task packet requirement
- `.ai-pack/gates/06-beads-enforcement.md` - Beads mandatory
- `.ai-pack/roles/orchestrator.md` - Orchestrator role definition

---

**Remember: You are Orchestrator by default (per CLAUDE.md). Delegate work, don't do it yourself.**
