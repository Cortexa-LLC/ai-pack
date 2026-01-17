# TC-GT-001: Lean Flow Gate Enforcement

**Category:** Gates
**Priority:** Critical
**Status:** Active
**Last Updated:** 2026-01-15

---

## Objective

Validate that Gate 05 (Lean Flow) correctly enforces small batch sizes, WIP limits, and token budget constraints to prevent production failures.

## Background

**Production Failures from Violating Lean Flow Principles:**
- Harvana WunderGraph: 25 files → 5 token limit failures
- Multiple false success reports from parallel agent chaos
- File persistence failures from complex multi-file tasks

**Root Cause (User Observation):**
> "A lot of what we're seeing is due to trying to do too much all at once."

**Solution:**
Gate 05 enforces Lean Flow principles from "Accelerate" (Gene Kim, Jez Humble, Nicole Forsgren):
- Small batch sizes (≤14 files per task packet)
- WIP limits (≤3 concurrent spawned agents)
- Token budget pre-verification

**Reference:** `principles/LEAN-FLOW.md`, `gates/05-lean-flow.md`

---

## Prerequisites

- Project with ai-pack framework
- Ability to create task packets
- Ability to spawn spawned agents
- Understanding of batch size limits

---

## Test Scenario

### Scenario A: Batch Size Enforcement

**Test A1: Ideal Batch Size (1-5 files)**

1. **Create task packet with 3 files:**
   ```bash
   mkdir -p .ai/tasks/2026-01-15_test-lean-flow-ideal
   cp .ai-pack/templates/task-packet/* .ai/tasks/2026-01-15_test-lean-flow-ideal/
   ```

2. **Fill contract with 3-file estimate:**
   ```markdown
   ## Lean Flow Analysis

   **Estimated Files:** 3 files
   - UserController.cs
   - UserService.cs
   - UserControllerTests.cs

   **Batch Size Evaluation:** ✅ IDEAL (1-5 files)
   ```

3. **Expected: Gate PASSES immediately**
   ```
   ✅ GATE 05: LEAN FLOW - PASSED

   Batch size: 3 files
   Limit: ≤14 files
   Status: IDEAL batch size

   Proceeding with task packet creation...
   ```

**Test A2: Acceptable Batch Size (6-14 files)**

4. **Create task packet with 10 files:**
   ```markdown
   ## Lean Flow Analysis

   **Estimated Files:** 10 files
   - 5 controller files
   - 3 service files
   - 2 test files

   **Batch Size Evaluation:** ⚠️ ACCEPTABLE (6-14 files)

   ### Batch Size Justification

   Files: 10 (within acceptable range)

   **Why not decomposed further:**
   - High cohesion - all files implement single user management feature
   - Already minimal viable batch for user CRUD operations
   - Tightly coupled - splitting would create integration complexity

   **Contingency for token limits:**
   - If token limit hit, will decompose into:
     - Batch 1: Controllers + tests (5 files)
     - Batch 2: Services + tests (5 files)

   **Estimated tokens:** ~10 × 3000 = 30,000 tokens
   **Status:** Within 25K-42K limit? YES (acceptable risk)
   ```

5. **Expected: Gate WARNS but PASSES with plan**
   ```
   ⚠️ GATE 05: LEAN FLOW - WARNING

   Batch size: 10 files
   Recommended limit: ≤14 files

   This batch is larger than ideal. You MUST:
   ✅ Documented decomposition consideration in 00-contract.md
   ✅ Explained why not decomposed further
   ✅ Created contingency plan for token limit issues

   PROCEEDING WITH CAUTION...
   ```

**Test A3: Too Large Batch Size (15-26 files)**

6. **Create task packet with 20 files:**
   ```markdown
   ## Lean Flow Analysis

   **Estimated Files:** 20 files
   - Complete authentication system with all endpoints
   ```

7. **Expected: Gate WARNS, requires decomposition plan**
   ```
   ⚠️ GATE 05: LEAN FLOW - WARNING

   Estimated files: 20 files
   Recommended limit: ≤14 files per task packet

   This batch is larger than ideal. While not blocking, you MUST:
   1. Document decomposition consideration in 10-plan.md
   2. Explain why not decomposed further
   3. Create contingency plan for token limit issues

   Recommended: Decompose into 2 smaller batches for faster feedback.

   Proceeding with caution...
   ```

**Test A4: Critical Batch Size (27+ files)**

8. **Attempt to create task packet with 35 files:**
   ```markdown
   ## Lean Flow Analysis

   **Estimated Files:** 35 files
   - Complete e-commerce checkout system
   ```

9. **Expected: Gate BLOCKS**
   ```
   ❌ GATE 05: LEAN FLOW - BLOCKED

   Estimated files: 35 files
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

   EXAMPLE DECOMPOSITION:
   ❌ WRONG: "Implement complete checkout" (35 files)
   ✅ CORRECT:
     - Task 1: "Shopping cart model + service" (7 files)
     - Task 2: "Checkout controller + validation" (6 files)
     - Task 3: "Payment processing" (7 files)
     - Task 4: "Order fulfillment" (7 files)
     - Task 5: "Integration tests" (5 files)

   Reference: principles/LEAN-FLOW.md
   Cannot proceed until decomposed.
   ```

10. **User decomposes into 3 task packets (12, 11, 12 files)**

11. **Expected: Gate PASSES for each smaller batch**

---

### Scenario B: WIP Limit Enforcement

**Test B1: Ideal WIP (1 agent)**

12. **Orchestrator spawns 1 spawned agent:**
    ```python
    # Check current WIP
    active_agents = count_active_background_agents()
    # Result: 0

    # Spawn agent
    Task(..., )
    ```

13. **Expected: Gate PASSES**
    ```
    ✅ GATE 05: LEAN FLOW - PASSED

    Current WIP: 0 agents
    Spawning: 1 agent
    Maximum allowed: 3 agents

    Status: IDEAL WIP level
    Proceeding...
    ```

**Test B2: Acceptable WIP (2-3 agents)**

14. **Orchestrator spawns 2nd agent (now 2 total):**
    ```python
    active_agents = count_active_background_agents()
    # Result: 1

    Task(..., )
    ```

15. **Expected: Gate PASSES with note**
    ```
    ✅ GATE 05: LEAN FLOW - PASSED

    Current WIP: 1 agent
    Spawning: 2nd agent
    Maximum allowed: 3 agents

    Status: Within WIP limits
    Proceeding...
    ```

16. **Orchestrator spawns 3rd agent (now 3 total):**
    ```python
    active_agents = count_active_background_agents()
    # Result: 2

    Task(..., )
    ```

17. **Expected: Gate WARNS but PASSES**
    ```
    ⚠️ GATE 05: LEAN FLOW - WARNING

    Current WIP: 2 agents
    Spawning: 3rd agent
    Maximum allowed: 3 agents

    You are at the WIP limit. Consider:
    - Waiting for one agent to complete before spawning next
    - Sequential execution may be faster overall
    - Verification overhead increases with agent count

    PROCEEDING but cannot spawn additional agents...
    ```

**Test B3: WIP Limit Exceeded (4+ agents)**

18. **Orchestrator attempts to spawn 4th agent:**
    ```python
    active_agents = count_active_background_agents()
    # Result: 3

    # Attempt to spawn 4th
    Task(..., )
    ```

19. **Expected: Gate BLOCKS**
    ```
    ❌ GATE 05: LEAN FLOW - BLOCKED

    Current spawned agents: 3
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

    Queue Theory (Little's Law):
      Cycle Time = WIP / Throughput

      With 6 agents:
      - WIP = 6 tasks, Throughput = 2/hour
      - Cycle Time = 3 hours per task

      With 2 agents:
      - WIP = 2 tasks, Throughput = 2/hour
      - Cycle Time = 1 hour per task

      Result: 66% faster with LOWER WIP!

    REQUIRED ACTION (choose one):

    Option 1: Wait for completion
      - Monitor existing agents
      - Run verification protocol as they complete
      - Spawn new agent after WIP drops below 3

    Option 2: Sequential execution
      - Remove 
      - Execute tasks one at a time

    Option 3: Further decomposition
      - Break work into even smaller batches

    Cannot spawn 4th agent until WIP reduced.
    ```

20. **Orchestrator waits for agent-1 to complete**

21. **Agent-1 completes, WIP drops to 2**

22. **Orchestrator spawns new agent:**
    ```
    ✅ GATE 05: LEAN FLOW - PASSED

    Current WIP: 2 agents
    Spawning: New agent
    Status: Within limits

    Proceeding...
    ```

---

### Scenario C: Token Budget Enforcement

**Test C1: Safe Token Budget (<20K tokens)**

23. **Estimate task with 5 files:**
    ```markdown
    ## Token Budget Estimation

    Files × Average Tokens = Estimated Total
    5 × 3,000 tokens = 15,000 tokens

    Agent Output Limit: 25K-32K tokens

    Status: 15K tokens → ✅ SAFE
    ```

24. **Expected: Gate PASSES**
    ```
    ✅ GATE 05: LEAN FLOW - PASSED

    Estimated tokens: 15,000
    Agent limit: 25K-32K tokens
    Status: SAFE (well under limit)

    Proceeding...
    ```

**Test C2: Approaching Limit (20-25K tokens)**

25. **Estimate task with 8 files:**
    ```markdown
    ## Token Budget Estimation

    Files × Average Tokens = Estimated Total
    8 × 3,000 tokens = 24,000 tokens

    Status: 24K tokens → ⚠️ APPROACHING LIMIT
    ```

26. **Expected: Gate WARNS but PASSES**
    ```
    ⚠️ GATE 05: LEAN FLOW - WARNING

    Estimated tokens: 24,000
    Agent limit: 25K-32K tokens
    Status: APPROACHING LIMIT

    Risk: May hit token limit during execution
    Recommendation: Consider decomposing into 2 batches of 4 files

    PROCEEDING WITH CAUTION...
    ```

**Test C3: High Risk (25-42K tokens)**

27. **Estimate task with 12 files:**
    ```markdown
    ## Token Budget Estimation

    Files × Average Tokens = Estimated Total
    12 × 3,000 tokens = 36,000 tokens

    Status: 36K tokens → ❌ HIGH RISK
    ```

28. **Expected: Gate WARNS strongly**
    ```
    ⚠️ GATE 05: LEAN FLOW - WARNING

    Estimated files: 12 files
    Estimated tokens: ~36K tokens
    Agent limit: 25K-32K tokens

    This batch is approaching token limits. While not blocking,
    there is significant risk of token limit failure.

    RISK ASSESSMENT:
    - Probability of token limit: ~40%
    - Impact if fails: Agent truncates output, reports success but incomplete
    - Recovery time: 15-30 minutes to detect and retry

    RECOMMENDATION:
    Decompose into 2 smaller batches:
    - Batch 1: 6 files (~18K tokens) ✅ Safe
    - Batch 2: 6 files (~18K tokens) ✅ Safe

    Benefits of decomposition:
    - Faster feedback
    - Lower failure risk (0% vs 40%)
    - Easier verification

    PROCEEDING WITH CAUTION
    Monitor agent output for token limit warnings...
    ```

**Test C4: Guaranteed Failure (>42K tokens)**

29. **Attempt task with 20 files:**
    ```markdown
    ## Token Budget Estimation

    Files × Average Tokens = Estimated Total
    20 × 3,000 tokens = 60,000 tokens

    Status: 60K tokens → ❌ GUARANTEED FAILURE
    ```

30. **Expected: Gate BLOCKS**
    ```
    ❌ GATE 05: LEAN FLOW - BLOCKED

    Estimated files: 20
    Estimated tokens: 60,000
    Agent limit: 25K-32K tokens

    PROBLEM:
    This batch is too large and will hit token limits, as seen in:
    - TC-BA-002: Token Limit Detection
    - Harvana WunderGraph gateway (5 failed attempts)
    - Agent truncates output, reports "success" but incomplete

    REQUIRED ACTION:
    1. Decompose into batches of ≤14 files
    2. Each batch should target <42K tokens
    3. Create multiple task packets if needed

    CALCULATION:
      Current: 20 files × 3K tokens = 60,000 tokens
      Target: ≤14 files × 3K tokens = ≤42,000 tokens

    Cannot proceed with current batch size.
    ```

---

## Expected Behavior

### Batch Size Gate

**1-5 files:**
```
✅ PASS immediately
Status: IDEAL batch size
```

**6-14 files:**
```
⚠️ WARN but PASS
Requires: Justification in 00-contract.md
Status: ACCEPTABLE with plan
```

**15-26 files:**
```
⚠️ WARN strongly
Requires: Decomposition plan in 10-plan.md
Status: CONDITIONAL PASS
```

**27+ files:**
```
❌ BLOCK immediately
Requires: Decompose into ≤14 file batches
Status: BLOCKED until decomposed
```

---

### WIP Limit Gate

**1 agent:**
```
✅ PASS immediately
Status: IDEAL WIP
```

**2-3 agents:**
```
✅ PASS with note
Status: Within limits
```

**4+ agents:**
```
❌ BLOCK immediately
Requires: Wait for completion OR sequential execution
Status: BLOCKED at WIP limit
```

---

### Token Budget Gate

**<20K tokens:**
```
✅ PASS immediately
Status: SAFE
```

**20-25K tokens:**
```
⚠️ WARN but PASS
Status: APPROACHING LIMIT
```

**25-42K tokens:**
```
⚠️ WARN strongly
Risk: 40% failure probability
Recommendation: Decompose
Status: HIGH RISK
```

**>42K tokens:**
```
❌ BLOCK immediately
Requires: Decompose to <42K tokens
Status: GUARANTEED FAILURE
```

---

## Actual Behavior (Execution Record)

**Test Run:** [Date]

### Scenario A: Batch Size

**Test A1 (3 files):**
- Gate decision: [PASS/WARN/BLOCK]
- Status message: [Captured]
- Allowed to proceed: [Yes/No]

**Test A2 (10 files):**
- Gate decision: [PASS/WARN/BLOCK]
- Required justification: [Yes/No]
- Justification provided: [Yes/No]
- Allowed to proceed: [Yes/No]

**Test A3 (20 files):**
- Gate decision: [PASS/WARN/BLOCK]
- Required plan: [Yes/No]
- Plan provided: [Yes/No]
- Allowed to proceed: [Yes/No]

**Test A4 (35 files):**
- Gate decision: [PASS/WARN/BLOCK]
- Blocked spawning: [Yes/No]
- Decomposition required: [Yes/No]
- After decomposition: [PASS/FAIL]

### Scenario B: WIP Limits

**Test B1 (1 agent):**
- Gate decision: [PASS/WARN/BLOCK]
- Agent spawned: [Yes/No]

**Test B2 (2-3 agents):**
- Gate decision for 2nd: [PASS/WARN/BLOCK]
- Gate decision for 3rd: [PASS/WARN/BLOCK]
- Warning shown: [Yes/No]

**Test B3 (4th agent):**
- Gate decision: [PASS/WARN/BLOCK]
- Blocked spawning: [Yes/No]
- Required wait: [Yes/No]
- After completion: [PASS/FAIL]

### Scenario C: Token Budget

**Test C1 (15K tokens):**
- Gate decision: [PASS/WARN/BLOCK]

**Test C2 (24K tokens):**
- Gate decision: [PASS/WARN/BLOCK]
- Warning shown: [Yes/No]

**Test C3 (36K tokens):**
- Gate decision: [PASS/WARN/BLOCK]
- Risk assessment shown: [Yes/No]
- Recommendation given: [Yes/No]

**Test C4 (60K tokens):**
- Gate decision: [PASS/WARN/BLOCK]
- Blocked execution: [Yes/No]
- Required decomposition: [Yes/No]

**Deviations:**
[Any differences from expected]

---

## Pass/Fail Criteria

### PASS Criteria

**Batch Size Enforcement:**
✅ Ideal batches (1-5 files) pass immediately
✅ Acceptable batches (6-14 files) warn but pass with justification
✅ Large batches (15-26 files) warn and require plan
✅ Critical batches (27+ files) block immediately
✅ Clear error messages explain limits
✅ Decomposition guidance provided

**WIP Limit Enforcement:**
✅ 1-3 agents allowed to spawn
✅ 4th agent blocked with clear message
✅ Queue theory rationale explained
✅ Options provided (wait, sequential, decompose)
✅ Gate unblocks when WIP drops below 3

**Token Budget Enforcement:**
✅ Safe budgets (<20K) pass immediately
✅ Approaching limit (20-25K) warns
✅ High risk (25-42K) warns strongly with probabilities
✅ Guaranteed failure (>42K) blocks
✅ Decomposition targets provided

**Overall:**
✅ Gates enforce Lean Flow principles
✅ Violations prevented before execution
✅ Clear guidance for remediation
✅ No false positives (valid batches allowed)
✅ No false negatives (violations blocked)

---

### FAIL Criteria

**Batch Size:**
❌ Critical batches (27+ files) not blocked
❌ No guidance on decomposition
❌ Unclear error messages
❌ Valid batches incorrectly blocked

**WIP Limits:**
❌ 4+ agents allowed to spawn
❌ No blocking at WIP limit
❌ Unclear rationale for limits
❌ No remediation options provided

**Token Budget:**
❌ >42K token tasks not blocked
❌ No risk assessment shown
❌ Token limit failures not prevented
❌ Safe tasks incorrectly blocked

---

## Metrics

**Effectiveness Metrics:**
```
Before Gate 05:
- Token limit failures: ~40% of large batches
- Average batch size: 18 files
- WIP violations: Common (5+ agents)
- Cycle time: 3+ hours

After Gate 05:
- Token limit failures: 0% (blocked before execution)
- Average batch size: 8 files
- WIP violations: 0% (enforced)
- Cycle time: <2 hours
```

**Gate Performance:**
```
True Positives (correctly blocked violations): [X]
True Negatives (correctly allowed valid batches): [X]
False Positives (incorrectly blocked valid): [X] (target: 0)
False Negatives (incorrectly allowed violations): [X] (target: 0)

Accuracy: (TP + TN) / (TP + TN + FP + FN) × 100%
Target: >95%
```

---

## Known Issues

**Issue 1: File Count Estimation Accuracy**
- Estimating file count before implementation can be inaccurate
- **Mitigation:** Update estimate during planning phase
- **Contingency:** If actual > estimated, gate checks again before spawning

**Issue 2: Token Budget Variance**
- Actual tokens vary by file complexity (1K-5K per file)
- Conservative 3K average may over-estimate for simple files
- **Mitigation:** Adjust estimate based on file types in justification

**Issue 3: Dynamic WIP Tracking**
- Need real-time tracking of active spawned agents
- **Implementation:** Orchestrator maintains agent registry
- **Verification:** Count agents before each spawn

---

## Integration with Other Tests

**Related Test Cases:**
- **TC-BA-002:** Token Limit Detection (validates token budget enforcement)
- **TC-OR-005:** Task Decomposition Size Limits (validates batch size guidance)
- **TC-OR-001:** Completion Verification Protocol (validates WIP management)

**Test Sequence:**
1. TC-GT-001 (this test) - Validates gate enforcement
2. TC-OR-005 - Validates orchestrator applies limits
3. TC-BA-002 - Validates agent handles small batches

---

## References

**Principles:**
- `principles/LEAN-FLOW.md` - Complete Lean Flow documentation
- "Accelerate" by Gene Kim, Jez Humble, Nicole Forsgren

**Gates:**
- `gates/05-lean-flow.md` - Gate implementation

**Production Evidence:**
- Harvana WunderGraph: 25 files → 5 token limit failures
- User observation: "Trying to do too much all at once"

**Related Documentation:**
- `roles/orchestrator.md` - Section on task decomposition
- `templates/task-packet/00-contract.md` - Lean Flow Analysis section

---

**Version:** 1.0.0
**Last Validated:** [Date]
**Status:** Active - CRITICAL
