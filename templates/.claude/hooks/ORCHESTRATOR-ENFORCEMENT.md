# Orchestrator Role Enforcement Hook

## Purpose

This hook **structurally enforces** Orchestrator role boundaries by **blocking submissions** that violate the delegation pattern.

## Problem It Solves

The Orchestrator role has a fundamental issue: LLMs trained to "be helpful" will execute work themselves instead of delegating, even when explicitly instructed not to. Textual guidance alone is insufficient.

**Pattern observed:**
```
Orchestrator writes: "Task:Tester validates tests IN [prompt]"
Then immediately runs: npm test, analyzes results, writes reviews
```

This is NOT delegation - it's execution with descriptive text.

## How It Works

The hook runs on **every user prompt submission** and:

1. **Detects if Orchestrator role is active** by scanning recent conversation for activation markers
2. **Inspects all tool calls** about to be executed
3. **Blocks execution tools** that violate delegation pattern
4. **Allows only**: Task tool calls and read-only verification

### Allowed Tools for Orchestrator

✅ **Task** - Spawn specialist agents (THE PRIMARY TOOL)
✅ **Read** - Read task packets, logs, settings (verification only)
✅ **Bash (limited)** - Only safe read-only commands:
  - `cat .claude/settings.json`
  - `grep -A`
  - `test -f` / `test -d`
  - `ls`
  - `pwd`

### Blocked Tools for Orchestrator

❌ **Write** - Except `10-plan.md` (planning phase)
❌ **Edit** - Except `10-plan.md` (planning phase)
❌ **Bash (execution)** - Any command that:
  - Runs tests (`dotnet test`, `npm test`, `pytest`, etc.)
  - Runs builds (`dotnet build`, `npm run`, etc.)
  - Writes files (`>`, `>>`, `echo`)

## Violation Example

**Blocked submission:**
```python
# Orchestrator attempts:
Bash(command="npm test -- --coverage")
```

**Hook response:**
```
🛑 ORCHESTRATOR ROLE VIOLATION BLOCKED

Violations detected:
  ❌ Bash execution: npm test -- --coverage

What Orchestrator MUST do instead:
  ✅ Use Task(...) tool to spawn specialist agents
  ✅ Delegate execution to Tester agent

This submission has been BLOCKED.
```

## Correct Pattern

**Orchestrator should do:**
```python
Task(subagent_type="general-purpose",
     description="Validate SDK tests and coverage",
     prompt="Act as Tester. Run tests and validate coverage per .ai-pack/roles/tester.md...",
     run_in_background=true)
```

**NOT:**
```python
# WRONG - Orchestrator doing execution
Bash(command="npm test")
Read(file_path="test-results.json")
Write(file_path=".ai/tasks/.../30-review.md", content="...")
```

## Technical Details

### Hook Lifecycle

1. User hits Enter (submits prompt)
2. Claude Code calls `UserPromptSubmit` hooks
3. `orchestrator-enforcement.py` receives hook context (JSON via stdin)
4. Hook parses conversation + tool calls
5. If violation: Exit code 2 (BLOCK submission)
6. If clean: Exit code 0 (ALLOW submission)

### Exit Codes

- **0** - Allow (no violation or not Orchestrator)
- **1** - Error (technical failure, fail open)
- **2** - Block (Orchestrator using execution tools)

### Hook Input Format

Claude Code passes context as JSON:
```json
{
  "conversation": [
    {"role": "user", "content": "..."},
    {"role": "assistant", "content": "..."}
  ],
  "toolCalls": [
    {"name": "Bash", "input": {"command": "npm test"}},
    {"name": "Write", "input": {"file_path": "..."}}
  ]
}
```

## Limitations

1. **Detection delay** - Hook runs on submission, not during composition
2. **Conversation scanning** - Checks last 10 messages for Orchestrator activation
3. **Pattern matching** - Uses text patterns, not semantic understanding
4. **Fail open** - On technical errors, allows submission (safety)

## Testing the Hook

### Test 1: Block Orchestrator Bash execution
```bash
# Activate Orchestrator role, then try:
Bash(command="dotnet test")
# Expected: BLOCKED
```

### Test 2: Block Orchestrator Write
```bash
# Activate Orchestrator role, then try:
Write(file_path=".ai/tasks/foo/30-review.md", content="test")
# Expected: BLOCKED
```

### Test 3: Allow Orchestrator Task delegation
```bash
# Activate Orchestrator role, then try:
Task(subagent_type="general-purpose", description="test", prompt="...")
# Expected: ALLOWED
```

### Test 4: Allow non-Orchestrator execution
```bash
# As Engineer or directly (no role), try:
Bash(command="dotnet test")
# Expected: ALLOWED (hook only enforces on Orchestrator)
```

## Maintenance

### Adding Forbidden Commands

Edit `orchestrator-enforcement.py`, update `forbidden_bash` list:

```python
forbidden_bash = [
    'dotnet test',
    'npm test',
    'your-new-command',  # Add here
]
```

### Adding Safe Commands

Edit `safe_commands` list:

```python
safe_commands = [
    'cat .claude/settings.json',
    'ls ',
    'your-safe-command',  # Add here
]
```

### Adjusting Orchestrator Detection

Edit `orchestrator_indicators` list:

```python
orchestrator_indicators = [
    'orchestrator role',
    'acting as the **orchestrator**',
    'your-new-pattern',  # Add here
]
```

## Integration with Framework

This hook is part of the ai-pack framework's **layered enforcement**:

1. **Layer 1 (Passive)** - Documentation in `CLAUDE.md`
2. **Layer 2 (Active)** - Skill file instructions
3. **Layer 3 (Enforcement)** - **THIS HOOK** - Blocks violations
4. **Layer 4 (Monitoring)** - Watchdog detects systemic failures

**This is Layer 3 - the enforcement layer that actually prevents violations.**

## Why This Is Necessary

**Textual guidance has failed repeatedly:**
- Orchestrator skill has explicit "DON'T RUN TESTS" instructions
- Multiple checkpoints and warnings
- Clear examples of wrong vs right patterns

**Yet violations still occur because:**
- LLM training prioritizes "helpfulness" over role constraints
- Tool availability creates temptation to "just do it"
- Delegation requires trust that agent will succeed

**Structural enforcement solves this:**
- Removes execution capability from Orchestrator
- Forces delegation pattern mechanically
- Makes violations impossible, not just discouraged

## See Also

- [Orchestrator Skill](./../skills/orchestrator/SKILL.md) - Role definition
- [Task Packet Hook](./check-task-packet.py) - Similar enforcement pattern
- [Hooks README](./README.md) - Hook system overview

---

**Status:** Active enforcement (runs on every user submission when Orchestrator active)
**Severity:** Critical (prevents core role boundary violations)
**Maintenance:** Update forbidden/safe command lists as needed
