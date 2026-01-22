---
sidebar_position: 2
title: "Lean Flow Enforcement"
---

# Gate 05: Lean Flow Enforcement

**Priority:** CRITICAL
**Type:** BLOCKING
**Enforcement:** Pre-Delegation
**Last Updated:** 2026-01-15

---

## Purpose

Enforce Lean Flow principles to prevent:
- Token limit failures from oversized batches
- Verification chaos from too many parallel agents
- Context overload from excessive WIP
- Slow cycle times from large batch sizes

**Based on:** "Accelerate" by Gene Kim, Jez Humble, Nicole Forsgren (see `principles/LEAN-FLOW.md`)

---

## Gate Rules

### Rule 1: Small Batch Size Enforcement (BLOCKING)

**Requirement:** All task packets and agent delegations MUST respect batch size limits.

**Batch Size Limits:**
```text
Task Packet Files:
├─ 1-5 files   → ✅ PASS: Ideal batch size
├─ 6-14 files  → ⚠️ CONDITIONAL PASS: Requires decomposition plan
├─ 15-26 files → ❌ FAIL: MUST decompose into 2-3 task packets
└─ 27+ files   → ❌ FAIL: MUST decompose into 3+ task packets

Agent Delegation:
├─ 1-5 files   → ✅ PASS: Ideal delegation
├─ 6-14 files  → ⚠️ CONDITIONAL PASS: Document why not decomposed
├─ 15+ files   → ❌ FAIL: MUST decompose into multiple agents
└─ 27+ files   → ❌ CRITICAL FAIL: Guaranteed token limit failure
```text

**Verification:**
```bash
# Before creating task packet
ESTIMATE file_count for task

IF file_count > 26 THEN
  BLOCK with error: "CRITICAL: Task too large (${file_count} files)"
  REQUIRE: Decomposition into ≤14 file batches
  STATUS: BLOCKED until decomposed

ELSE IF file_count > 14 THEN
  WARN: "Task large (${file_count} files)"
  REQUIRE: Document decomposition plan in 10-plan.md
  STATUS: CONDITIONAL PASS with plan

ELSE
  PASS: Batch size within limits
END IF
```text

**Error Messages:**

**Critical Failure (27+ files):**
```text
❌ LEAN FLOW GATE BLOCKED

Estimated files: 27+ files
Batch size limit: ≤14 files per task packet

PROBLEM:
Large batches cause token limit failures, as seen in:
- Harvana WunderGraph gateway (25 files → 5 token limit failures)
- Multiple production failures documented in TC-BA-002

REQUIRED ACTION:
1. Decompose task into smaller batches (≤14 files each)
2. Create separate task packets for each batch
3. Document dependencies in 10-plan.md
4. Sequence work appropriately

EXAMPLE:
❌ WRONG: "Implement complete auth system" (35 files)
✅ CORRECT:
  - Task 1: "User model + password hashing" (7 files)
  - Task 2: "Login endpoint + JWT" (6 files)
  - Task 3: "Registration + validation" (7 files)
  - Task 4: "Session management" (6 files)
  - Task 5: "Integration tests" (5 files)

Reference: principles/LEAN-FLOW.md
Cannot proceed until decomposed.
```text

**Warning (15-26 files):**
```text
⚠️ LEAN FLOW GATE WARNING

Estimated files: ${file_count} files
Recommended limit: ≤14 files per task packet

This batch is larger than ideal. While not blocking, you MUST:
1. Document decomposition consideration in 10-plan.md
2. Explain why not decomposed further
3. Create contingency plan for token limit issues

Recommended: Decompose into 2 smaller batches for faster feedback.

Proceeding with caution...
```text

---

### Rule 2: Work In Progress (WIP) Limit Enforcement (BLOCKING)

**Requirement:** Orchestrators MUST NOT exceed WIP limits for spawned agents.

**WIP Limits:**
```text
Spawned Agents:
├─ 0-1 agents  → ✅ PASS: Ideal (complete before next)
├─ 2-3 agents  → ⚠️ ACCEPTABLE: Within limits
├─ 4-5 agents  → ❌ FAIL: Exceeds limit, blocks spawning
└─ 6+ agents   → ❌ CRITICAL FAIL: Verification chaos

Active Task Packets:
├─ 1 packet    → ✅ PASS: Ideal focus
├─ 2 packets   → ⚠️ ACCEPTABLE: Within limits
└─ 3+ packets  → ❌ FAIL: Too much context switching
```text

**Verification:**
```python
# Before spawning agent
active_agents = count_active_agents()

IF active_agents >= 3 THEN
  BLOCK with error: "WIP limit reached (${active_agents}/3)"
  REQUIRE: Wait for agent completion
  STATUS: BLOCKED until WIP reduced

ELSE IF active_agents == 2 THEN
  WARN: "WIP approaching limit (2/3)"
  RECOMMEND: Consider sequential execution
  STATUS: PASS with warning

ELSE
  PASS: WIP within ideal limits
END IF
```text

**Error Messages:**

**WIP Limit Exceeded (4+ agents):**
```text
❌ WIP LIMIT EXCEEDED

Current agents: ${active_agents}
Maximum allowed: 3 concurrent agents

PROBLEM:
Too many parallel agents causes:
- Verification protocol overwhelm
- Context dilution across agents
- Coordination overhead exceeding benefit
- Slower overall completion (queue theory)

Queue Theory (Little's Law):
  Cycle Time = WIP / Throughput

  Example with 6 agents:
  - WIP = 6 tasks
  - Throughput = 2 tasks/hour
  - Cycle Time = 3 hours per task

  Example with 2 agents:
  - WIP = 2 tasks
  - Throughput = 2 tasks/hour
  - Cycle Time = 1 hour per task

  Result: 66% faster with LOWER WIP!

REQUIRED ACTION (choose one):

Option 1: Wait for completion
  - Monitor existing agents
  - Run verification protocol as they complete
  - Spawn new agent after WIP drops below 3

Option 2: Further decomposition
  - Break work into even smaller batches
  - Complete and verify before next

Reference: principles/LEAN-FLOW.md Section 2
Cannot spawn agent until WIP reduced.
```text

**WIP Warning (3 agents):**
```text
⚠️ WIP APPROACHING LIMIT

Current agents: 3
Maximum allowed: 3 concurrent agents

You are at the WIP limit. Consider:
- Waiting for one agent to complete before spawning next
- Sequential execution may be faster overall
- Verification overhead increases with agent count

Proceeding but cannot spawn additional agents...
```text

---

### Rule 3: Token Budget Pre-Verification (BLOCKING)

**Requirement:** Orchestrators MUST estimate token budget before delegation.

**Token Budget Guidelines:**
```text
File Count → Estimated Tokens:
├─ 1-5 files   → ~3K-15K tokens  → ✅ SAFE
├─ 6-14 files  → ~18K-42K tokens → ⚠️ APPROACHING LIMIT
├─ 15-26 files → ~45K-78K tokens → ❌ LIKELY TO FAIL
└─ 27+ files   → ~81K+ tokens    → ❌ GUARANTEED FAILURE

Agent Output Limit: 25K-32K tokens (varies by model)
```text

**Verification:**
```python
# Before spawning agent
file_count = estimate_file_count(task)
estimated_tokens = file_count * 3000  # Conservative estimate

IF estimated_tokens > 32000 THEN
  BLOCK with error: "Token budget exceeded"
  REQUIRE: Decompose into smaller batches
  STATUS: BLOCKED

ELSE IF estimated_tokens > 25000 THEN
  WARN: "Token budget approaching limit"
  RECOMMEND: Consider decomposition
  STATUS: PASS with warning

ELSE
  PASS: Token budget within safe limits
END IF
```text

**Error Messages:**

**Token Budget Exceeded:**
```text
❌ TOKEN BUDGET EXCEEDED

Estimated files: ${file_count}
Estimated tokens: ${estimated_tokens}
Agent limit: 25K-32K tokens

PROBLEM:
This batch is too large and will hit token limits, as seen in:
- TC-BA-002: Token Limit Detection
- Harvana WunderGraph gateway (5 failed attempts)
- Agent truncates output, reports "success" but incomplete

REQUIRED ACTION:
1. Decompose into batches of ≤8 files
2. Each batch should target <20K tokens
3. Create multiple task packets if needed

CALCULATION:
  Current: ${file_count} files × 3K tokens = ${estimated_tokens} tokens
  Target: ≤8 files × 3K tokens = ≤24K tokens

Cannot proceed with current batch size.
```text

---

### Rule 4: Flow Optimization (ADVISORY)

**Requirement:** Optimize for flow, not utilization.

**Flow Metrics:**
```text
Cycle Time: Time from task start to completion
Lead Time: Time from user request to deployment
Deployment Frequency: How often we ship
Change Fail Rate: Percentage requiring rework
```text

**Optimization Principles:**
```text
❌ WRONG: Maximize Utilization
- Keep all agents busy
- Start many tasks in parallel
- High WIP
- Result: Long cycle times, frequent failures

✅ CORRECT: Maximize Flow
- Complete work before starting new
- Low WIP (1-3 active items)
- Small batches
- Result: Fast cycle times, low failure rates
```text

**Advisory Guidance:**
```text
IF cycle_time > 2_hours FOR small_batch THEN
  INVESTIGATE:
    - Is batch size too large?
    - Is WIP too high?
    - Are there unnecessary handoffs?
    - Is verification taking too long?

  OPTIMIZE:
    - Reduce batch size further
    - Lower WIP limits
    - Minimize handoffs
    - Automate verification

ELSE
  Flow is healthy, continue current practices
END IF
```text

---

## Enforcement Points

### Point 1: Task Packet Creation

**When:** Before creating task packet in `.ai/tasks/`

**Check:**
```bash
# Estimate file count
file_count=$(estimate_files_for_task "$task_description")

# Run batch size check
if [ $file_count -gt 15 ]; then
  echo "❌ BLOCKED: Batch too large ($file_count files)"
  echo "REQUIRED: Decompose into ≤8 file batches"
  exit 1
elif [ $file_count -gt 8 ]; then
  echo "⚠️ WARNING: Large batch ($file_count files)"
  echo "RECOMMENDED: Document decomposition plan"
fi
```text

**Passes:** Batch size ≤14 files (or documented plan for 15-26)
**Fails:** Batch size >26 files (BLOCKING)

---

### Point 2: Agent Spawning

**When:** Before calling `Task(...)`

**Check:**
```python
# Count active agents
active_count = len(list_active_agents())

# Enforce WIP limit
if active_count >= 3:
    raise WIPLimitExceeded(
        f"Cannot spawn agent: WIP limit reached ({active_count}/3)"
    )
elif active_count == 2:
    warn(f"WIP approaching limit ({active_count}/3)")

# Estimate token budget
file_count = estimate_file_count(task)
tokens = file_count * 3000

if tokens > 32000:
    raise TokenBudgetExceeded(
        f"Cannot spawn agent: Token budget exceeded ({tokens} tokens)"
    )
elif tokens > 25000:
    warn(f"Token budget approaching limit ({tokens} tokens)")
```text

**Passes:** WIP ≤3 agents AND token budget ≤32K
**Fails:** WIP ≥4 agents OR token budget >32K (BLOCKING)

---

### Point 3: Agent Delegation

**When:** Before delegating work to any agent

**Check:**
```python
# Verify batch size
if task.file_count > 15:
    block("Batch too large, decompose required")
elif task.file_count > 8:
    warn("Large batch, decomposition recommended")

# Verify WIP limits
if current_wip >= 3:
    block("WIP limit reached")

# Verify token budget
if estimated_tokens > 32000:
    block("Token budget exceeded")
```text

**Passes:** All checks pass
**Fails:** Any check fails (BLOCKING on critical, WARNING on threshold)

---

## Bypass Conditions

**This gate CANNOT be bypassed.**

Lean Flow principles are fundamental to preventing:
- Token limit failures (production critical)
- Verification chaos (quality critical)
- Slow cycle times (productivity critical)

**Rationale:**
Every bypass of this gate in production has resulted in failures:
- Harvana WunderGraph: 25 files → 5 token limit failures
- Harvana tests: Complex task → files not persisted
- Multiple PRD/ADR failures: Parallel agent chaos → missing artifacts

---

## Metrics

**Track these metrics to validate gate effectiveness:**

```markdown
## Lean Flow Metrics

**Batch Size:**
- Average files per task packet: ${avg_batch_size}
- Target: <5 files
- Status: ${batch_size_status}

**Work In Progress:**
- Average concurrent agents: ${avg_wip}
- Maximum concurrent agents: ${max_wip}
- Target: ≤3 agents
- Status: ${wip_status}

**Token Budget:**
- Average token usage: ${avg_tokens}
- Token limit failures: ${token_failures}
- Target: 0 failures
- Status: ${token_status}

**Cycle Time:**
- Average task completion: ${avg_cycle_time}
- Target: <2 hours for small batch
- Status: ${cycle_time_status}

**Quality:**
- Verification failure rate: ${verification_failures}%
- Target: <15%
- Status: ${quality_status}
```text

---

## Examples

### Example 1: Batch Size Enforcement

**Scenario:** Orchestrator about to create large task packet

```markdown
## Task: Implement Complete E-Commerce Checkout

Deliverables:
1. Shopping cart model
2. Cart service
3. Checkout controller
4. Payment processing
5. Order confirmation
6. Email notifications
7. Receipt generation
8. Inventory updates
9. Tax calculation
10. Shipping integration
11. Discount codes
12. Gift cards
13. Refund processing
14. Analytics tracking
15. Integration tests
16. Unit tests
17. API documentation

Estimated files: 25+
```text

**Gate Response:**
```text
❌ GATE 05: LEAN FLOW - BLOCKED

Batch size: 25+ files
Limit: ≤8 files per task packet

This batch is CRITICALLY oversized and will cause:
- Token limit failures (guaranteed at 25 files)
- Verification complexity
- High failure risk

REQUIRED: Decompose into smaller batches

SUGGESTED DECOMPOSITION:

Task Packet 1: Shopping Cart Foundation
- Cart model, service, tests (4 files)

Task Packet 2: Basic Checkout Flow
- Checkout controller, order model, tests (4 files)

Task Packet 3: Payment Processing
- Payment service, integration, tests (4 files)

Task Packet 4: Order Fulfillment
- Email service, receipt generation, tests (4 files)

Task Packet 5: Inventory & Tax
- Inventory updates, tax calculation, tests (4 files)

Task Packet 6: Advanced Features
- Discounts, gift cards, refunds, tests (6 files)

Task Packet 7: Integration & Docs
- Integration tests, documentation (3 files)

Result: 7 batches of 3-6 files each (all under limit)

Cannot proceed until decomposed.
```text

---

### Example 2: WIP Limit Enforcement

**Scenario:** Orchestrator attempting to spawn 4th agent

```python
# Current state
active_agents = [
    "agent-1: Implementing user authentication (in progress)",
    "agent-2: Creating API documentation (in progress)",
    "agent-3: Setting up database migrations (in progress)"
]

# Orchestrator attempts to spawn 4th agent
Task(
    subagent_type="general-purpose",
    description="Add email notifications"
)
```text

**Gate Response:**
```text
❌ GATE 05: LEAN FLOW - BLOCKED

Current WIP: 3 agents
Attempting to spawn: 4th agent
Maximum allowed: 3 concurrent agents

ACTIVE AGENTS:
1. agent-1: Implementing user authentication
2. agent-2: Creating API documentation
3. agent-3: Setting up database migrations

PROBLEM:
Exceeding WIP limits causes:
- Verification protocol overwhelm
- Longer cycle times (queue theory)
- Coordination complexity
- Higher failure rates

REQUIRED ACTION (choose one):

Option 1: Wait for completion
  Wait for one of the 3 active agents to complete
  Run verification protocol
  Then spawn email notifications agent

Option 2: Defer work
  Add to task queue
  Process after current WIP completes

Cannot spawn 4th agent while at WIP limit.
```text

---

### Example 3: Token Budget Warning

**Scenario:** Orchestrator delegating moderate-sized batch

```python
# Task estimation
task_files = 9  # Just over recommended limit
estimated_tokens = 9 * 3000 = 27,000 tokens
```text

**Gate Response:**
```text
⚠️ GATE 05: LEAN FLOW - WARNING

Estimated files: 9 files
Estimated tokens: ~27K tokens
Agent limit: 25K-32K tokens

This batch is approaching token limits. While not blocking,
there is significant risk of token limit failure.

RISK ASSESSMENT:
- Probability of token limit: ~40%
- Impact if fails: Agent truncates output, reports success but incomplete
- Recovery time: 15-30 minutes to detect and retry

RECOMMENDATION:
Decompose into 2 smaller batches:
- Batch 1: 5 files (~15K tokens) ✅ Safe
- Batch 2: 4 files (~12K tokens) ✅ Safe

Benefits of decomposition:
- Faster feedback (complete first batch in 1 hour vs waiting 2+ hours)
- Lower failure risk (0% vs 40%)
- Easier verification (fewer files to check)

PROCEEDING WITH CAUTION
Monitor agent output for token limit warnings...
```text

---

## Success Criteria

**This gate is successful when:**

✅ Average batch size < 5 files
✅ No task packets > 8 files without documented plan
✅ WIP consistently ≤ 3 agents
✅ Zero token limit failures
✅ Cycle time < 2 hours for small batches
✅ Verification failure rate < 15%

**Leading Indicators:**
- Task packets being decomposed proactively
- WIP limits respected without exception
- Token budget analysis performed before delegation
- Faster overall completion despite "slower" starts

---

## References

**Principles:**
- `principles/LEAN-FLOW.md` - Complete Lean Flow documentation
- "Accelerate" by Gene Kim, Jez Humble, Nicole Forsgren
- Queue Theory (Little's Law)

**Test Cases:**
- `tests/validation/background-agents/TC-BA-002-token-limit-detection.md`
- `tests/validation/orchestrator/TC-OR-005-task-decomposition.md`

**Production Failures:**
- Harvana WunderGraph: 25 files → 5 token limit failures
- Multiple PRD/ADR failures from parallel agent chaos
- File persistence failures from complex multi-file tasks

**Related Gates:**
- Gate 10: Persistence Gate (artifact verification)
- Gate 25: Execution Strategy Gate (parallelization decisions)

---

**Version:** 1.0.0
**Status:** Active - MANDATORY
**Enforcement:** BLOCKING on violations
