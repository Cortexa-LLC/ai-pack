# Architecture Decision: Skills Are Context, Not Execution

## Date
2026-01-10

## Status
**IMPLEMENTED** - Critical fix for orchestration pattern

## Context

### Problem Identified
The Orchestrator role has a fundamental failure mode:

1. Orchestrator activates (via skill or command)
2. It calls `Skill("ai-pack:test")` to "delegate" to Tester
3. Tester skill loads and **executes directly in the same session**
4. Tester runs tests, analyzes results, writes reviews
5. This bypasses all orchestration and defeats delegation pattern

**Root cause:** Skills contain both:
- **Context** (role definition, guidance, instructions)
- **Execution capability** (direct tool access)

When a skill activates, it can execute immediately, bypassing the Task tool delegation pattern.

### Why Enforcement Hook Failed
The `orchestrator-enforcement.py` hook blocks execution tools when **Orchestrator role is active**, but:

1. Orchestrator calls `Skill("test")` - This is allowed (delegation mechanism)
2. Tester skill activates - **New context**, Orchestrator no longer "active"
3. Tester executes with full tool access - Hook doesn't block (not Orchestrator anymore)

The hook can't detect skill transitions that change the active role.

## Decision

**Skills MUST be context-only, not execution.**

### New Architecture

```
┌─────────────────────────────────────────────────────┐
│ USER REQUEST                                        │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
         ┌───────────────────────┐
         │ Orchestrator Skill    │ ← Context only (guidance)
         │ (Read-only)           │
         └───────────┬───────────┘
                     │
                     ▼
         Uses Task tool to spawn:
                     │
         ┌───────────┴───────────┐
         │                       │
         ▼                       ▼
    ┌─────────┐           ┌──────────┐
    │ Agent A │           │ Agent B  │ ← Background tasks
    │ (Tester)│           │ (Reviewer)│ ← Full tool access
    └─────────┘           └──────────┘ ← Autonomous execution
         │                       │
         └───────────┬───────────┘
                     │
                     ▼
              Work completes
```

**Key principle:** Skills provide context, Task tool provides execution.

### Implementation

**1. Skills Load Context Only**
- Skills contain role definition and guidance
- Skills DO NOT execute tools directly
- Skills DO NOT call Task tool themselves
- Skills are pure documentation/context

**2. Orchestrator Delegates via Task Tool**
- Orchestrator (in main session) calls Task tool
- Task tool spawns **separate subprocess** with Tester context
- Subprocess has full tool access and executes autonomously
- Main session (Orchestrator) waits and monitors

**3. No Skill-to-Skill Transitions**
- Never `Skill("test")` from Orchestrator
- Only `Task(subagent_type="general-purpose", prompt="Act as Tester...")`
- Skills don't trigger other skills

## Consequences

### Positive
- ✅ Clear separation: context vs execution
- ✅ Orchestrator cannot bypass delegation
- ✅ Enforcement hook works (no role transitions)
- ✅ Skills become pure documentation
- ✅ Delegation is the only execution path

### Negative
- ❌ Cannot use Skill tool for role switching
- ❌ Skills must be loaded via Task tool prompts
- ❌ User cannot directly invoke skills for execution
- ❌ More verbose (Task tool requires full prompt)

### Neutral
- Skills become "role templates" not "role executors"
- Commands (like `/ai-pack test`) must spawn Task agents, not activate skills
- Skills are context references for Task tool prompts

## Migration Path

### Phase 1: Update Orchestrator Skill ✅ DONE
- Remove Skill tool usage
- Only allow Task tool calls
- Block via hook

### Phase 2: Update Command Files ← **NEXT**
- Commands should spawn Task agents
- Commands should NOT activate skills directly
- Example: `/ai-pack test` → Task(prompt="Act as Tester...")

### Phase 3: Deprecate Skill Tool in Skills
- Skills document their role
- Skills do NOT execute
- Skills loaded by Task tool via prompt context

### Phase 4: Update Hook to Block Skill Tool
- Add Skill tool to forbidden list for Orchestrator
- Force Task tool as only delegation mechanism

## Examples

### WRONG (Current broken pattern)
```python
# Orchestrator tries to delegate:
Skill("ai-pack:test")  # ❌ Skill activates and executes directly
```

### RIGHT (Correct delegation pattern)
```python
# Orchestrator delegates via Task tool:
Task(subagent_type="general-purpose",
     description="Validate SDK tests",
     prompt="""
     Act as Tester role from .ai-pack/roles/tester.md

     Validate C# SDK tests:
     - Run: dotnet test --filter HarvanaFeatureFlagProviderTests
     - Check coverage: 90%+ target
     - Document in 30-review.md

     Follow Tester role guidelines completely.
     """,
     run_in_background=true)  # Spawns separate agent
```

### Command Implementation

**WRONG:**
```markdown
# /ai-pack test command
Activates Tester skill and executes validation.
```

**RIGHT:**
```markdown
# /ai-pack test command
Creates a todo to spawn Tester agent via Task tool.
User or Orchestrator then makes Task call.
```

## Related Decisions

- **orchestrator-enforcement.py hook** - Blocks execution tools
- **Task tool delegation pattern** - Only valid delegation mechanism
- **Background agents** - True parallel execution with autonomy

## References

- [Orchestrator Skill](templates/.claude/skills/orchestrator/SKILL.md)
- [Enforcement Hook](templates/.claude/hooks/ORCHESTRATOR-ENFORCEMENT.md)
- [Task Tool Documentation](https://docs.claude.ai/agent-sdk)

## Status of Skills

| Skill | Current State | Target State |
|-------|---------------|--------------|
| orchestrator | Execution capable | Context only ✅ (blocked by hook) |
| engineer | Execution capable | Context only (needs update) |
| tester | Execution capable | Context only (needs update) |
| reviewer | Execution capable | Context only (needs update) |
| coordinator | Monitoring capable | Context only (needs update) |

## Next Actions

1. ✅ Orchestrator enforcement hook created
2. ⏳ **Update `/ai-pack test` command to spawn Task agent**
3. ⏳ **Update `/ai-pack review` command to spawn Task agent**
4. ⏳ **Add Skill tool to forbidden list in hook**
5. ⏳ **Document this pattern in framework README**

---

**Decision:** Skills are context providers, not executors. Task tool is the only execution mechanism.
