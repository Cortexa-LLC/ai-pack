# Standard Workflow

**Version:** 1.0.0
**Last Updated:** 2026-01-07

## Overview

The Standard Workflow is a general-purpose process applicable to any development task. It provides a structured, four-phase approach that ensures quality and completeness.

**Use this workflow for:** Any task that doesn't fit specialized workflows (feature, bugfix, refactor, research).

---

## Workflow Phases

```
Phase 1: Understanding → Phase 2: Planning → Phase 3: Implementation → Phase 4: Review
     ↓                        ↓                    ↓                      ↓
  [GATE: Ready]          [GATE: Planned]     [GATE: Implemented]   [GATE: Accepted]
```

---

## Phase 1: Understanding

**Objective:** Gain complete understanding of requirements and context.

### Activities

#### 1.1 Requirements Clarification
```
□ Read and understand user request
□ Identify success criteria
□ Clarify ambiguities
□ Document assumptions
□ Confirm understanding with user
```

**Tools:**
- AskUserQuestion - for clarifications
- TodoWrite - to track understanding tasks

**Deliverables:**
- Clear task description
- Success criteria documented
- Ambiguities resolved
- Assumptions documented

---

#### 1.2 Codebase Exploration
```
□ Identify relevant files
□ Read existing code
□ Understand architecture
□ Identify patterns to follow
□ Map dependencies
```

**Tools:**
- Task (Explore agent) - for open-ended exploration
- Read - for reading specific files
- Grep - for finding code patterns
- Glob - for finding files

**Exploration Questions:**
- Where is similar functionality?
- What patterns are used?
- How is testing structured?
- What are the dependencies?
- How does this fit the architecture?

---

#### 1.3 Context Gathering
```
□ Review related issues/tasks
□ Check recent changes (git log)
□ Understand user context
□ Identify constraints
□ Note technical debt in area
```

**Commands:**
```bash
git log --oneline -20  # recent changes
git diff main           # current changes
```

---

### Exit Criteria (Gate: Ready)

```
✓ Requirements clear and unambiguous
✓ Success criteria defined
✓ Codebase context understood
✓ Existing patterns identified
✓ Dependencies mapped
✓ Constraints documented
✓ Ready to plan implementation
```

**If not met:** Continue exploration, ask more questions, document gaps.

---

## Phase 2: Planning

**Objective:** Design a sound implementation approach.

### Activities

#### 2.1 Approach Design
```
□ Identify implementation strategy
□ Break down into steps
□ Identify technical challenges
□ Consider alternatives
□ Select best approach
```

**Considerations:**
- Follows existing patterns?
- Minimizes complexity?
- Respects SOLID principles?
- Avoids over-engineering?
- Testable design?

---

#### 2.2 Risk Assessment
```
□ Identify technical risks
□ Assess impact on existing code
□ Consider performance implications
□ Evaluate security concerns
□ Plan mitigation strategies
```

**Risk Categories:**
- Technical (complexity, unknowns)
- Integration (breaking changes)
- Performance (scalability)
- Security (vulnerabilities)
- Maintenance (technical debt)

---

#### 2.3 Resource Estimation
```
□ Estimate effort (not time)
□ Identify required expertise
□ Note dependencies on externals
□ Plan for testing
□ Consider documentation needs
□ Determine parallel execution strategy (DEFAULT)
```

**Effort Levels:**
- Trivial: Single file, < 50 lines, obvious approach
- Small: 2-3 files, straightforward
- Medium: Multiple files, some complexity
- Large: Significant changes, architectural impact
- Very Large: Major refactoring or feature

---

#### 2.4 Execution Strategy Determination (MANDATORY)

**GATE CHECKPOINT:** [Execution Strategy Analysis](../gates/25-execution-strategy.md)

**REQUIREMENT:** Before proceeding to implementation, orchestrator MUST analyze and document execution strategy.

**Trigger:** Any work package with 2+ subtasks

**Mandatory Analysis Procedure:**
```
STEP 1: Count subtasks and assess independence
  - List all subtasks with files they modify
  - Identify which subtasks are independent
  - Document dependencies between subtasks

STEP 2: Evaluate for parallel execution
  IF 3+ independent subtasks THEN
    REQUIRED: Plan parallel execution (not optional)
    Workers: min(independent_count, 5)
    Launch: Single message block with multiple Task() calls
  ELSE IF mix of independent and dependent THEN
    REQUIRED: Plan hybrid execution
    Phase 1: Sequential for dependent chain
    Phase 2: Parallel for independent group
  ELSE
    Sequential execution acceptable
  END IF

STEP 3: Consider shared context constraints
  ✅ SHARED: Source repo, build folders, coverage data
  ❌ FORBIDDEN: Deleting builds, per-worker branches
  ⚠️ COORDINATE: Build ops, coverage merging, migrations

STEP 4: Document strategy decision
  Write analysis to task packet 10-plan.md:
  - Subtask inventory
  - Independence assessment
  - Strategy choice (PARALLEL/SEQUENTIAL/HYBRID)
  - Rationale
  - Worker assignment plan
  - Shared context coordination plan
```

**Parallel Execution Strategy (ENFORCED):**
```
FOR 3+ independent subtasks:
  MANDATORY: Parallel worker execution

  Implementation:
  1. Identify independent work units
  2. Assign one worker per unit (max 5)
  3. Assign different files to each worker
  4. Define shared context coordination
  5. Plan integration testing

  Launch Pattern:
  Single message with multiple Task() calls:
  - Task(worker, "subtask 1 description")
  - Task(worker, "subtask 2 description")
  - Task(worker, "subtask 3 description")

  Benefits:
  - 3-4x faster delivery
  - Independent verification
  - Clear ownership
  - Enforced parallelization

FOR 1-2 subtasks OR strong dependencies:
  Sequential single-worker approach acceptable
```

**Strategy Decision Matrix:**
```
Subtasks | Independence | Dependencies | Strategy    | Required
---------|--------------|--------------|-------------|----------
3+       | All          | None         | PARALLEL    | Mandatory
3+       | Mixed        | Some         | HYBRID      | Mandatory
3+       | None         | Strong       | SEQUENTIAL  | Justified
1-2      | Any          | Any          | SEQUENTIAL  | Default
```

**Gate Compliance Checklist:**
```
□ Subtask count determined
□ Independence assessed for each
□ Dependencies identified
□ Shared context constraints evaluated
□ Strategy determined (PARALLEL/SEQUENTIAL/HYBRID)
□ Rationale documented
□ Worker assignments planned
□ Launch pattern defined

IF all checked THEN proceed to implementation
ELSE complete missing analysis
```

---

### Deliverables

**Implementation Plan:**
```
Approach Summary:
- High-level strategy
- Key technical decisions
- Patterns to use

Step-by-Step Plan:
1. [Step 1 description]
2. [Step 2 description]
...

Critical Files:
- file1.ext - [what changes]
- file2.ext - [what changes]

Testing Strategy:
- Unit tests for [...]
- Integration tests for [...]

Risks and Mitigation:
- [Risk]: [Mitigation plan]
```

---

### Exit Criteria (Gate: Planned)

```
✓ Implementation approach clear
✓ Steps well-defined
✓ Risks identified and mitigated
✓ Testing strategy planned
✓ Execution strategy determined (PARALLEL/SEQUENTIAL/HYBRID)
✓ Shared context constraints evaluated
✓ Worker assignment plan documented
✓ Plan approved (for non-trivial tasks)
```

**For non-trivial tasks:**
- Enter Plan Mode
- Present plan to user
- Get approval before implementing

---

## Phase 3: Implementation

**Objective:** Execute the plan with quality and verification.

### Activities

#### 3.1 Incremental Development
```
□ Follow TDD cycle (Red-Green-Refactor)
□ Implement one step at a time
□ Run tests after each change
□ Commit at logical checkpoints
□ Update work log
```

**TDD Cycle:**
```
FOR each feature/change:
  1. Write failing test (RED)
  2. Write minimal code to pass (GREEN)
  3. Refactor for quality (REFACTOR)
  4. Verify all tests still pass
  5. Commit if appropriate
END FOR
```

---

#### 3.2 Continuous Testing
```
□ Run tests frequently
□ Fix failures immediately
□ Monitor coverage
□ Add tests for edge cases
□ Verify no regressions
```

**Testing Commands:**
```bash
npm test              # run all tests
npm test -- --coverage  # with coverage
pytest tests/         # Python tests
cargo test           # Rust tests
```

---

#### 3.3 Progress Tracking
```
□ Update work log regularly
□ Mark completed subtasks
□ Document decisions made
□ Note issues encountered
□ Report blockers immediately
```

**Work Log Updates:**
```
## Session [timestamp]

Completed:
- [✓] Implemented feature X
- [✓] Added tests for happy path

In Progress:
- [ ] Adding error handling tests

Decisions:
- Chose approach A over B because [reason]

Blockers:
- None currently
```

---

### Exit Criteria (Gate: Implemented)

```
✓ All planned steps completed
✓ All tests passing (100%)
✓ Code coverage 80-90%
✓ Follows coding standards
✓ No TODO/FIXME unaddressed
✓ Work log complete
✓ Ready for review
```

**If not met:** Continue implementation, fix issues, complete work.

---

## Phase 4: Review

**Objective:** Verify quality and ensure acceptance criteria met.

### Activities

#### 4.1 Quality Verification
```
□ Run full test suite
□ Check test coverage
□ Verify build succeeds
□ Run linters
□ Check for security issues
```

**Verification Commands:**
```bash
npm run build         # verify build
npm run lint          # check code style
npm run test:coverage # verify coverage
```

---

#### 4.2 Standards Compliance
```
□ Code follows style guide
□ Design principles applied
□ Patterns consistent
□ Documentation adequate
□ No code smells
```

**Self-Review Checklist:**
- Read your own code critically
- Check against clean-code standards
- Verify SOLID principles applied
- Ensure consistency with existing code

---

#### 4.3 User Acceptance
```
□ All acceptance criteria met
□ Requirements satisfied
□ Edge cases handled
□ Error cases handled
□ User expectations met
```

**Acceptance Verification:**
```
FOR each acceptance criterion:
  verify implemented
  verify tested
  verify working correctly
END FOR
```

---

### Exit Criteria (Gate: Accepted)

```
✓ Quality verification complete
✓ Standards compliance verified
✓ All acceptance criteria met
✓ User approved (if applicable)
✓ Documentation complete
✓ Task complete
```

---

## Role Assignments

### Orchestrator
- **Phase 1:** Coordinate understanding
- **Phase 2:** Validate plan and determine parallel execution (DEFAULT)
- **Phase 3:** Monitor progress across parallel workers
- **Phase 4:** Verify completion and coordinate integration

**Parallel Coordination Responsibilities:**
```
When using parallel workers (DEFAULT for 3+ subtasks):
- Spawn workers in single message block
- Track progress independently per worker
- Coordinate integration points
- Resolve conflicts between parallel streams
- Ensure overall coherence
- Verify cross-cutting concerns
```

### Worker
- **Phase 1:** Explore and understand
- **Phase 2:** Develop plan
- **Phase 3:** Implement solution (potentially in parallel with others)
- **Phase 4:** Self-review

**Parallel Worker Responsibilities:**
```
When working in parallel with other workers:
- Own isolated deliverable
- Avoid modifying files owned by other workers
- Communicate dependency needs early
- Complete independently
- Flag integration concerns
```

### Reviewer
- **Phase 4:** Conduct formal review (may review parallel streams independently)

---

## Gate Checkpoints

### Global Gates (All Phases)
- Safety first (no destructive operations)
- Quality baseline (tests must pass)
- Communication protocol (ask when uncertain)
- Error handling (graceful recovery)
- Incremental progress (small steps)

### Persistence Gates
- Read before write
- Atomic operations
- Idempotency where possible
- Version control integration

### Tool Policy Gates
- Use specialized tools
- Parallelize independent operations
- Follow approval requirements

### Verification Gates
- Pre-implementation checks (Phase 1)
- Mid-implementation validation (Phase 3)
- Post-implementation verification (Phase 4)

---

## Task Packet Management

### Throughout Workflow

**Initialize (.ai/tasks/YYYY-MM-DD_task-name/):**
```
00-contract.md    - Define at Phase 1 start
10-plan.md        - Create at Phase 2
20-work-log.md    - Update during Phase 3
30-review.md      - Complete at Phase 4
40-acceptance.md  - Finalize at Phase 4 end
```

**Update Pattern:**
```
Phase 1 (Understanding):
  - Initialize 00-contract.md
  - Document requirements and acceptance criteria

Phase 2 (Planning):
  - Create 10-plan.md
  - Document approach and steps

Phase 3 (Implementation):
  - Update 20-work-log.md regularly
  - Document progress, decisions, issues

Phase 4 (Review):
  - Complete 30-review.md
  - Finalize 40-acceptance.md
  - Sign off on completion
```

---

## Workflow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│ PHASE 1: UNDERSTANDING                                      │
│ ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│ │Requirements │→ │  Codebase   │→ │   Context   │         │
│ │Clarification│  │ Exploration │  │  Gathering  │         │
│ └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                   ↓         │
│                                           [GATE: Ready]     │
└─────────────────────────────────────────────────────────────┘
                                                   ↓
┌─────────────────────────────────────────────────────────────┐
│ PHASE 2: PLANNING                                           │
│ ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│ │  Approach   │→ │    Risk     │→ │  Resource   │         │
│ │   Design    │  │ Assessment  │  │ Estimation  │         │
│ └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                   ↓         │
│                                         [GATE: Planned]     │
└─────────────────────────────────────────────────────────────┘
                                                   ↓
┌─────────────────────────────────────────────────────────────┐
│ PHASE 3: IMPLEMENTATION                                     │
│ ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│ │Incremental  │→ │ Continuous  │→ │  Progress   │         │
│ │Development  │  │   Testing   │  │  Tracking   │         │
│ └─────────────┘  └─────────────┘  └─────────────┘         │
│         ↑                                         ↓         │
│         └─────────[TDD: Red→Green→Refactor]──────┘         │
│                                                   ↓         │
│                                      [GATE: Implemented]    │
└─────────────────────────────────────────────────────────────┘
                                                   ↓
┌─────────────────────────────────────────────────────────────┐
│ PHASE 4: REVIEW                                             │
│ ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│ │   Quality   │→ │  Standards  │→ │    User     │         │
│ │Verification │  │ Compliance  │  │ Acceptance  │         │
│ └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                   ↓         │
│                                        [GATE: Accepted]     │
└─────────────────────────────────────────────────────────────┘
                                                   ↓
                                            [TASK COMPLETE]
```

---

## References

- [Global Gates](../gates/00-global-gates.md)
- [Persistence Gates](../gates/10-persistence.md)
- [Tool Policy](../gates/20-tool-policy.md)
- [Verification Gates](../gates/30-verification.md)
- [Orchestrator Role](../roles/orchestrator.md)
- [Worker Role](../roles/worker.md)
- [Reviewer Role](../roles/reviewer.md)
- [Task Packet Templates](../templates/task-packet/)

---

**Last reviewed:** 2026-01-07
**Next review:** Quarterly or when workflow needs refinement
