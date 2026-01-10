---
description: Assume Orchestrator role for complex multi-step tasks
---

# /ai-pack orchestrate - Orchestrator Role

Activates the **Orchestrator role** for coordinating complex, multi-step tasks that require delegation and coordination.

## When to Use This Role

Use the Orchestrator role when:
- Task requires 3+ independent subtasks
- Multiple roles needed (Engineer + Reviewer + Tester)
- Complex coordination required
- Parallel execution opportunities exist
- Task involves multiple phases or specialists

**Don't use for:** Simple, single-step tasks (use `/ai-pack engineer` instead)

## Orchestrator Responsibilities

### Phase 1: Task Analysis & Planning
1. Read the task packet contract (`.ai/tasks/*/00-contract.md`)
2. Analyze for parallelization opportunities
3. Break down into subtasks
4. Determine which specialists needed:
   - **Inspector** - Bug investigation
   - **Product Manager** - Requirements definition
   - **Architect** - Technical design
   - **Designer** - UX workflows and wireframes
   - **Engineer** - Implementation
   - **Tester** - Test validation
   - **Reviewer** - Code quality review

### Phase 2: Delegation Strategy
1. Create delegation plan in `.ai/tasks/*/10-plan.md`
2. Document execution strategy (sequential vs parallel)
3. Define acceptance criteria for each subtask
4. Coordinate shared context (build folders, databases, coverage)

### Phase 3: Execution Coordination
1. Spawn parallel workers when possible (Task tool)
2. Monitor progress via work logs
3. Resolve blockers
4. Coordinate handoffs between roles

### Phase 4: Quality Assurance
1. Trigger Tester validation (`/ai-pack test`)
2. Trigger Reviewer validation (`/ai-pack review`)
3. Ensure all acceptance criteria met
4. Update acceptance document

### Phase 5: Completion
1. Verify all subtasks complete
2. Document outcomes
3. Archive task packet

## Key Gates to Follow

**MANDATORY:**
1. **Execution Strategy Gate** - Document parallel vs sequential before execution
2. **Persistence Gate** - Ensure artifacts saved to `docs/` for planning specialists
3. **Code Quality Review Gate** - Both Tester and Reviewer must approve

See: `.ai-pack/gates/` for all gates

## Reference Documentation

**Primary:** [.ai-pack/roles/orchestrator.md](../../.ai-pack/roles/orchestrator.md)

**Related Gates:**
- [.ai-pack/gates/25-execution-strategy.md](../../.ai-pack/gates/25-execution-strategy.md) - Parallel execution
- [.ai-pack/gates/35-code-quality-review.md](../../.ai-pack/gates/35-code-quality-review.md) - Quality gates

**Related Workflows:**
- [.ai-pack/workflows/standard.md](../../.ai-pack/workflows/standard.md) - Standard workflow
- [.ai-pack/workflows/feature.md](../../.ai-pack/workflows/feature.md) - Feature development

## Related Commands

- `/ai-pack task-init` - Create task packet first
- `/ai-pack task-status` - Check progress
- `/ai-pack engineer` - Delegate to Engineer
- `/ai-pack review` - Trigger review phase
- `/ai-pack help` - Show all commands

## Activation

This command will:
1. Load the Orchestrator role definition
2. Verify task packet exists
3. Guide you through orchestration workflow
4. Enable access to delegation patterns

Ready to orchestrate this task?
