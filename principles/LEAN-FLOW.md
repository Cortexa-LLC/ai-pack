# Lean Flow Principles for AI-Pack

**Based on:** "Accelerate" by Gene Kim, Jez Humble, and Nicole Forsgren

**Last Updated:** 2026-01-15

---

## Core Problem

**Root Cause of Production Failures:**
> "A lot of what we're seeing is due to trying to do too much all at once."

**Evidence:**
- Token limit failures: 25-file tasks exceed context budget
- False success reporting: Too many parallel spawned agents to verify
- File persistence issues: Complex task packets with 15+ deliverables
- Orchestrator overwhelm: Tracking too many in-flight work items

---

## Principle 1: Small Batch Sizes

### Definition

**Small Batch:** A unit of work that can be completed, verified, and integrated in a single focused session.

### Guidelines for AI-Pack

**Task Packet Size Limits:**
- ✅ **IDEAL:** 1-5 files per task packet
- ⚠️ **ACCEPTABLE:** 6-14 files per task packet (requires decomposition plan)
- ❌ **TOO LARGE:** 15+ files per task packet (MUST decompose)

**Spawned Agent Delegation:**
- ✅ **IDEAL:** Single-purpose agents (one deliverable each)
- ⚠️ **ACCEPTABLE:** Multi-file agents (3-5 related files)
- ❌ **TOO LARGE:** Complex agents (6+ files or multiple concerns)

**Code Review Batches:**
- ✅ **IDEAL:** Review 1-3 files at a time
- ⚠️ **ACCEPTABLE:** Review 4-9 related files
- ❌ **TOO LARGE:** Review 10+ files (cognitive overload)

### Why Small Batches Work

**Faster Feedback:**
- Small tasks complete quickly
- Issues detected immediately
- Corrections applied before more work done

**Lower Risk:**
- Less context to lose if agent fails
- Easier to verify correctness
- Simpler rollback if needed

**Better Quality:**
- Focused attention on single concern
- Thorough verification possible
- Less chance of missed issues

**Token Budget Management:**
- Small context fits in token limits
- Less prompt verbosity needed
- Complete output fits in response window

### Examples

**❌ WRONG: Large Batch**
```markdown
## Task Packet: Implement Complete Authentication System

**Deliverables:**
1. User model with validation
2. Password hashing service
3. JWT token generation
4. Login endpoint
5. Logout endpoint
6. Password reset endpoint
7. Email verification
8. Session management
9. Role-based access control
10. Audit logging
11. Rate limiting
12. Integration tests for all endpoints
13. Unit tests for all services
14. API documentation
15. Database migrations
16. Password complexity validation
17. Account lockout logic
18. Two-factor authentication
19. OAuth integration
20. Session refresh tokens
21. API rate limiting middleware
22. Security headers middleware
23. CORS configuration
24. Authentication event logging
25. User profile endpoints

**Estimated Files:** 40+
**Token Budget Required:** 80K+ (EXCEEDS LIMITS)
```

**✅ CORRECT: Small Batches**
```markdown
## Task Packet 1: User Model Foundation (Session 1)

**Deliverables:**
1. User model with basic fields
2. Password hashing service
3. Password validation logic
4. User repository
5. Unit tests for all

**Estimated Files:** 5
**Token Budget:** ~15K

---

## Task Packet 2: Login Endpoint (Session 2)
**Prerequisites:** Task Packet 1 complete

**Deliverables:**
1. Login controller
2. JWT token generation service
3. Authentication middleware
4. Session management
5. Integration tests

**Estimated Files:** 5
**Token Budget:** ~15K

---

## Task Packet 3: Password Reset Flow (Session 3)
**Prerequisites:** Task Packet 2 complete

**Deliverables:**
1. Password reset controller
2. Email verification service
3. Token validation
4. Reset token storage
5. Integration tests

**Estimated Files:** 5
**Token Budget:** ~15K
```

---

## Principle 2: Limit Work In Progress (WIP)

### Definition

**Work In Progress:** Tasks that have been started but not yet completed and verified.

### WIP Limits for AI-Pack

**Orchestrator WIP Limits:**
- ✅ **MAXIMUM:** 3 spawned agents running simultaneously
- ⚠️ **PREFERRED:** 2 spawned agents (easier to verify)
- 🎯 **IDEAL:** 1 spawned agent (complete before next)

**Task Packet WIP Limits:**
- ✅ **MAXIMUM:** 2 active task packets in `.ai/tasks/active/`
- 🎯 **IDEAL:** 1 active task packet (finish before starting next)

**Engineer WIP Limits:**
- ✅ **MAXIMUM:** 1 task packet active at a time
- 🎯 **IDEAL:** Complete current file before starting next

### Why WIP Limits Work

**Reduced Context Switching:**
- Focus on completing work, not starting work
- Less mental overhead tracking multiple tasks
- Better quality from sustained attention

**Faster Cycle Time:**
- Paradoxically, limiting WIP speeds up completion
- Queue theory: WIP ∝ Cycle Time
- Finish tasks faster when you start fewer

**Easier Verification:**
- Fewer items to verify at completion
- Complete verification possible
- Less chance of missed failures

**Lower Token Budget:**
- Fewer agents to track in orchestrator context
- Simpler coordination prompts
- Less verification overhead

### Queue Theory Application

**Little's Law:**
```
Cycle Time = WIP / Throughput

Examples:
- WIP=6 tasks, Throughput=2/day → Cycle Time=3 days
- WIP=2 tasks, Throughput=2/day → Cycle Time=1 day

Reducing WIP from 6→2 cuts cycle time by 66%!
```

**Orchestrator Example:**
```
❌ WRONG: High WIP
- 5 spawned agents spawned
- Orchestrator tracking 5 outputs
- 5 verification protocols to run
- Context consumed by coordination
- Cycle time: 45 minutes

✅ CORRECT: Low WIP
- 1 spawned agent spawned
- Orchestrator tracks 1 output
- 1 verification protocol
- Minimal coordination overhead
- Cycle time: 8 minutes

Result: 5× faster completion per task!
```

### WIP Limit Enforcement

**Gate: Maximum Concurrent Spawned Agents**
```python
# In orchestrator skill

def spawn_background_agent(self, task):
    # Check current WIP
    active_agents = self.count_active_background_agents()

    if active_agents >= 3:
        # WIP LIMIT EXCEEDED
        return """
        ⚠️ WIP LIMIT REACHED

        Current spawned agents: {active_agents}
        Maximum allowed: 3

        MUST complete and verify existing agents before spawning new ones.

        Options:
        1. Wait for agent completion
        2. Decompose task further
        3. Run agent in sequential execution
        """

    # WIP within limits, proceed
    spawn_agent(task, spawning agents)
```

---

## Principle 3: Optimize for Flow

### Definition

**Flow:** The smooth progression of work from start to completion without delays, handoffs, or rework.

### Flow Optimization for AI-Pack

**Minimize Handoffs:**
- ✅ Engineer completes full feature (no handoff to separate tester)
- ✅ Single agent handles related files (no handoff between agents)
- ❌ Avoid: Cartographer → Architect → Designer → Engineer (4 handoffs!)

**Minimize Wait Time:**
- ✅ Background agents run in parallel (when WIP limits allow)
- ✅ Quick verification provides immediate feedback
- ❌ Avoid: Waiting hours for manual review before continuing

**Minimize Rework:**
- ✅ TDD prevents defects (write tests first)
- ✅ Small batches catch issues early
- ❌ Avoid: Large batches requiring extensive rework

### Flow Metrics

**Cycle Time:**
- Time from task packet creation to completion
- Target: < 2 hours for small batch
- Target: < 1 day for medium batch

**Lead Time:**
- Time from user request to deployment
- Target: Same-day for bug fixes
- Target: < 1 week for features

**Deployment Frequency:**
- How often we integrate changes
- Target: Multiple times per day

**Change Fail Rate:**
- Percentage of changes requiring rework
- Target: < 15%

---

## Principle 4: Build Quality In

### Definition

**Quality Built In:** Prevent defects rather than detect and fix them later.

### Quality Gates for AI-Pack

**Shift Left:**
- ✅ TDD: Write tests BEFORE implementation (prevents defects)
- ✅ Small batches: Catch issues early (cheap to fix)
- ✅ Immediate verification: Don't proceed with failures
- ❌ Avoid: Large batch → manual QA → rework cycle

**Automated Verification:**
- ✅ 5-step verification protocol (automatic)
- ✅ File existence checks (automatic)
- ✅ Test execution (automatic)
- ❌ Avoid: Manual verification only

**Stop the Line:**
- ✅ Gate blocks on test failures (don't proceed)
- ✅ Orchestrator blocks on verification failures
- ✅ Build failures halt feature work
- ❌ Avoid: Continuing with known failures

---

## Principle 5: Continuous Improvement

### Definition

**Continuous Improvement:** Regular reflection and adjustment based on data.

### Improvement Practices for AI-Pack

**Measure Everything:**
- Task packet completion times
- Token budget utilization
- Verification failure rates
- Rework percentages
- WIP levels

**Retrospectives:**
- After each major feature
- After each production failure
- Weekly team reviews
- Document lessons learned

**Experiment:**
- Try smaller batch sizes
- Test different WIP limits
- Compare sequential vs parallel
- A/B test approaches

---

## Implementation Guidelines

### For Orchestrators

**BEFORE Spawning Spawned Agents:**
1. Check WIP limit (max 3 agents)
2. Verify batch size (max 8 files)
3. Confirm token budget (< 25K)
4. Ensure permissions configured

**DURING Agent Execution:**
1. Track agent count
2. Monitor completion
3. Don't spawn new agents if at limit

**AFTER Agent Completion:**
1. Run 5-step verification immediately
2. Update WIP count (decrement)
3. Only then spawn next agent

### For Engineers

**BEFORE Starting Work:**
1. Verify only 1 active task packet
2. Check batch size is small
3. Confirm clear acceptance criteria

**DURING Implementation:**
1. Complete one file before next
2. Run tests after each file
3. Commit frequently (small batches)

**AFTER Completion:**
1. Verify all acceptance criteria
2. Update work log
3. Mark task complete in Beads

### For Reviewers

**BEFORE Review:**
1. Verify batch size is reviewable (< 5 files)
2. Check tests pass
3. Confirm TDD followed

**DURING Review:**
1. Review one file at a time
2. Block on quality issues (stop the line)
3. Provide immediate feedback

---

## Anti-Patterns to Avoid

### ❌ Anti-Pattern 1: The "Everything at Once" Task Packet

**Problem:**
```markdown
## Task: Build Complete Feature

- 15+ deliverables
- Multiple concerns
- Exceeds token budget
- Takes days to complete
```

**Fix:**
```markdown
## Task 1: Core Model (Small Batch)
- 3 deliverables
- Single concern
- Fits in token budget
- Completes in 1 hour

## Task 2: API Endpoints (Next Batch)
## Task 3: Tests (Final Batch)
```

### ❌ Anti-Pattern 2: The "Parallel Agent Swarm"

**Problem:**
```python
# Spawn 6 agents simultaneously
for task in tasks:
    Task(..., spawning agents)

# Result: Overwhelmed orchestrator, verification chaos
```

**Fix:**
```python
# Respect WIP limits
active_agents = 0
for task in tasks:
    if active_agents < 3:
        Task(..., spawning agents)
        active_agents += 1
    else:
        # Wait for completion or run sequentially
```

### ❌ Anti-Pattern 3: The "Continue Despite Failures"

**Problem:**
```
Agent 1: Token limit failure
Orchestrator: "Continuing to next task..."
Agent 2: Token limit failure
Orchestrator: "Continuing to next task..."
```

**Fix:**
```
Agent 1: Token limit failure
Orchestrator: "STOP. Verification failed. Must fix before proceeding."
- Decompose task smaller
- Fix token budget issue
- Retry with corrected approach
```

---

## Metrics Dashboard (Proposed)

```markdown
## AI-Pack Flow Metrics

**Batch Size:**
- Average files per task packet: 4.2
- Target: < 5
- Status: ✅ Within target

**Work In Progress:**
- Current active task packets: 2
- Current spawned agents: 1
- Target: ≤ 3 agents
- Status: ✅ Within limits

**Cycle Time:**
- Average task packet completion: 1.5 hours
- Target: < 2 hours
- Status: ✅ Meeting target

**Quality:**
- Verification failure rate: 8%
- Target: < 15%
- Status: ✅ Good quality

**Flow:**
- Deployment frequency: 3/day
- Change fail rate: 12%
- Status: ✅ Healthy flow
```

---

## References

**Books:**
- "Accelerate" by Gene Kim, Jez Humble, Nicole Forsgren
- "The DevOps Handbook" by Gene Kim et al.
- "The Phoenix Project" by Gene Kim
- "The Goal" by Eliyahu Goldratt (queue theory)

**Research:**
- State of DevOps Reports (2014-2023)
- Queue Theory applications to software development
- Lean manufacturing principles

**AI-Pack Documentation:**
- `tests/validation/orchestrator/TC-OR-005-task-decomposition.md`
- `tests/validation/background-agents/TC-BA-002-token-limit-detection.md`
- `roles/orchestrator.md` (Section 2.15: Verification Protocol)

---

## Success Criteria

**This framework is successful when:**

✅ Average batch size < 5 files
✅ WIP consistently ≤ 3 spawned agents
✅ Cycle time < 2 hours for small batches
✅ Verification failure rate < 15%
✅ No token limit failures (prevented by small batches)
✅ No file persistence failures (prevented by verification)
✅ Team velocity increases over time

**Leading Indicators:**
- Smaller task packets being created
- Fewer parallel agents spawned
- Faster completion times
- Lower rework rates

---

**Version:** 1.0.0
**Status:** Active
**Maintained By:** Bryan Woodruff
