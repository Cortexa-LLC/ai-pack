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
□ Check for persisted planning artifacts (see below)
```

**Commands:**
```bash
git log --oneline -20  # recent changes
git diff main           # current changes
```

**Check for Planning Artifacts:**
```
IF task relates to existing feature/architecture THEN
  Check docs/product/[feature-name]/ for:
  - PRD (Product Requirements Document)
  - Epics and user stories
  - Original requirements

  Check docs/architecture/[feature-name]/ for:
  - Architecture documents
  - API specifications
  - Data models

  Check docs/adr/ for:
  - Architecture Decision Records
  - Technical decisions and rationale

  Check docs/investigations/ for:
  - Related bug retrospectives
  - Known issues and patterns

  These documents provide context on WHY decisions were made,
  WHAT requirements exist, and HOW the system is designed.
END IF
```

**Documentation Location Reference:**
```
docs/
├── product/[feature-name]/      - PRDs, requirements, user stories
├── architecture/[feature-name]/ - Technical design, APIs, data models
├── adr/                         - Architecture Decision Records
└── investigations/              - Bug retrospectives, lessons learned
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

#### 2.5 Artifact Persistence Checkpoint (IF SPECIALISTS USED)

**GATE CHECKPOINT:** [Artifact Persistence Gate](../gates/10-persistence.md#11-artifact-repository-persistence)

**TRIGGER:** Planning involved Product Manager, Architect, or Inspector roles.

**REQUIREMENT:** Before proceeding to implementation, verify planning artifacts persisted to repository.

**Verification Procedure:**
```
IF Product Manager OR Architect OR Inspector involved THEN
  STEP 1: Remind specialist to persist artifacts
    Orchestrator: "Planning deliverables must be persisted to docs/
                   before we proceed to implementation."

  STEP 2: Verify artifacts committed
    CHECK docs/product/[feature-name]/ (if PM involved)
    CHECK docs/architecture/[feature-name]/ (if Architect involved)
    CHECK docs/adr/ (if Architect involved)
    CHECK docs/investigations/ (if Inspector involved)

  STEP 3: Verify cross-references present
    Artifacts must reference related documents
    See Orchestrator role section 2.11 for details

  STEP 4: IF verification fails THEN
    BLOCK implementation
    REQUIRE persistence completion
    RE-VERIFY before proceeding
  END IF

  STEP 5: ONLY AFTER verified THEN
    proceed_to_implementation()
  END IF
END IF
```

**Enforcement:**
```
GATE BLOCKS implementation if:
  ❌ Planning artifacts not committed to docs/
  ❌ Cross-references missing
  ❌ Files don't follow naming conventions
  ❌ Specialist hasn't confirmed persistence

GATE PASSES when:
  ✅ All artifacts committed to repository
  ✅ Cross-references present
  ✅ Files properly named and located
  ✅ Specialist confirmed persistence
```

**Why This Checkpoint:**
```
- Planning artifacts document decisions for years
- Engineers need context during implementation
- Future teams need to understand "why"
- Version control tracks evolution
- Single source of truth for specifications
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

**Objective:** Verify quality through mandatory reviews and ensure acceptance criteria met.

### Activities

#### 4.1 Worker Self-Review
```
□ Run full test suite locally
□ Check test coverage
□ Verify build succeeds
□ Run linters
□ Check for security issues
□ Basic quality self-assessment
```

**Verification Commands:**
```bash
npm run build         # verify build
npm run lint          # check code style
npm run test:coverage # verify coverage
```

**Self-Review Checklist:**
- Read your own code critically
- Check against clean-code standards
- Verify SOLID principles applied
- Ensure consistency with existing code

---

#### 4.2 Tester Validation (MANDATORY for Code Changes)

**ENFORCEMENT:** Orchestrator MUST delegate to Tester agent for all work packages involving code changes. Enforced by **[Code Quality Review Gate](../gates/35-code-quality-review.md)**.

**Tester Responsibilities:**
```
□ Validate TDD process compliance (RED-GREEN-REFACTOR)
□ Verify test coverage meets targets (80-90%)
□ Assess test quality and reliability
□ Verify test type coverage (unit/integration/e2e)
□ Check test scenarios (happy/edge/error cases)
□ Provide verdict: APPROVED or CHANGES REQUIRED
```

**Tester Delegation (Orchestrator):**
```
IF work includes code changes THEN
  tester = Task(
    subagent_type="general-purpose",
    prompt="You are the Tester role. Validate TDD compliance and test sufficiency.
            Focus: TDD process, coverage (80-90%), test quality.
            Report in .ai/tasks/${task_id}/30-review.md"
  )

  tester_result = wait_for_completion(tester)

  IF tester_result == "CHANGES REQUIRED" THEN
    coordinate_test_fixes()
    resubmit_to_tester()
  END IF
END IF
```

**Tester Blocking Conditions:**
```
❌ WORK INCOMPLETE if:
- TDD not followed
- Coverage < 80%
- Tests failing
- Critical logic untested (<95%)
- Error handling untested (<90%)
- Integration points untested (<100%)
- Flaky tests present
```

---

#### 4.3 Test Issue Resolution (IF NEEDED)

**Trigger:** Tester verdict == "CHANGES REQUIRED"

**Resolution Process:**
```
□ Worker reviews Tester findings
□ Worker addresses Critical findings (mandatory)
□ Worker addresses Major findings (mandatory)
□ Worker re-runs tests and coverage
□ Worker requests Tester re-validation
□ Repeat until Tester verdict: APPROVED
```

**Status:** Work remains INCOMPLETE until Tester approves

---

#### 4.4 Reviewer Validation (MANDATORY for Code Changes)

**ENFORCEMENT:** Orchestrator MUST delegate to Reviewer agent for all work packages involving code changes. Enforced by **[Code Quality Review Gate](../gates/35-code-quality-review.md)**.

**Reviewer Responsibilities:**
```
□ Review code quality against standards
□ Verify architecture consistency
□ Assess security concerns
□ Evaluate documentation adequacy
□ Check acceptance criteria met
□ Provide verdict: APPROVED or CHANGES REQUESTED
```

**Reviewer Delegation (Orchestrator):**
```
IF work includes code changes THEN
  reviewer = Task(
    subagent_type="general-purpose",
    prompt="You are the Reviewer role. Review code quality and standards.
            Focus: quality, architecture, security, documentation.
            Report in .ai/tasks/${task_id}/30-review.md"
  )

  reviewer_result = wait_for_completion(reviewer)

  IF reviewer_result == "CHANGES REQUESTED" THEN
    coordinate_code_fixes()
    resubmit_to_tester()    // Verify tests still pass
    resubmit_to_reviewer()
  END IF
END IF
```

**Reviewer Blocking Conditions:**
```
❌ WORK INCOMPLETE if:
- Security vulnerabilities
- Major standards violations
- Architecture violations
- Poor error handling
- Missing critical tests
- Acceptance criteria not met
```

---

#### 4.5 Code Issue Resolution (IF NEEDED)

**Trigger:** Reviewer verdict == "CHANGES REQUESTED"

**Resolution Process:**
```
□ Worker reviews Reviewer findings
□ Worker addresses Critical findings (mandatory)
□ Worker addresses Major findings (mandatory)
□ Worker re-runs all tests (verify still passing)
□ Worker requests Tester re-validation (ensure tests still pass)
□ Worker requests Reviewer re-validation
□ Repeat until Reviewer verdict: APPROVED
```

**Status:** Work remains INCOMPLETE until Reviewer approves

---

#### 4.6 User Acceptance

**Prerequisites:**
```
✓ Tester validation: APPROVED (if code changes)
✓ Reviewer validation: APPROVED (if code changes)
✓ All blocking issues resolved
```

**User Acceptance Verification:**
```
□ All acceptance criteria met
□ Requirements satisfied
□ Edge cases handled
□ Error cases handled
□ User expectations met
□ Documentation complete
```

**Acceptance Process:**
```
FOR each acceptance criterion:
  verify implemented
  verify tested
  verify working correctly
END FOR

IF all criteria met THEN
  user_approval = confirm_with_user()
END IF
```

---

### Exit Criteria (Gate: Accepted)

**UPDATED Exit Criteria:**
```
✓ Worker self-review complete
✓ Tester validation: APPROVED (MANDATORY for code changes)
✓ Reviewer validation: APPROVED (MANDATORY for code changes)
✓ All Critical/Major findings resolved
✓ Tests passing (100%)
✓ Coverage meets target (80-90%)
✓ Standards compliance verified
✓ All acceptance criteria met
✓ User approved (if applicable)
✓ Documentation complete (30-review.md, 40-acceptance.md)
✓ Task complete and signed off
```

**Blocking Conditions (Gate Fails):**
```
❌ Cannot proceed to acceptance if:
- Code changes present AND Tester not invoked
- Code changes present AND Reviewer not invoked
- Tester verdict: CHANGES REQUIRED (unresolved)
- Reviewer verdict: CHANGES REQUESTED (unresolved)
- Blocking issues unresolved
- Acceptance criteria not met
```

**Work Status:**
```
IF Tester OR Reviewer has blocking issues THEN
  WORK STATUS = INCOMPLETE
  BLOCK acceptance
  BLOCK sign-off
  REQUIRE fixes and re-validation
END IF
```

---

## Role Assignments

### Orchestrator
- **Phase 1:** Coordinate understanding
- **Phase 2:** Validate plan, determine parallel execution (DEFAULT), and enforce artifact persistence
- **Phase 3:** Monitor progress across parallel workers
- **Phase 4:** Coordinate mandatory reviews (Tester + Reviewer) and verify completion

**Phase 2 Artifact Persistence Enforcement (MANDATORY if specialists used):**
```
When Product Manager, Architect, or Inspector involved:
- Remind specialist to persist deliverables to docs/
- Verify artifacts committed to repository
- Verify cross-references present in artifacts
- Block implementation until persistence verified
- See Orchestrator role sections 2.10 and 2.11 for details
```

**Phase 4 Review Coordination (MANDATORY for code changes):**
```
When code changes present:
- Delegate to Tester agent (TDD and test validation)
- Delegate to Reviewer agent (code quality validation)
- Coordinate issue resolution if either finds blocking issues
- Verify both validations pass before acceptance
- Block completion if either validation fails
```

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
- **Phase 4:** Self-review, address Tester/Reviewer findings

**Phase 4 Responsibilities:**
```
- Complete self-review (tests, coverage, linters)
- Address Tester findings (if any)
- Address Reviewer findings (if any)
- Re-validate until both approve
- Update 30-review.md with resolution
```

**Parallel Worker Responsibilities:**
```
When working in parallel with other workers:
- Own isolated deliverable
- Avoid modifying files owned by other workers
- Communicate dependency needs early
- Complete independently
- Flag integration concerns
```

### Tester
- **Phase 4:** Validate TDD compliance and test sufficiency (MANDATORY for code changes)

**Tester Responsibilities:**
```
For all work packages with code changes:
- Verify TDD process followed (RED-GREEN-REFACTOR)
- Validate test coverage (80-90% overall, 95%+ critical logic)
- Assess test quality (clarity, independence, reliability)
- Verify test type coverage (unit/integration/e2e pyramid)
- Check test scenarios (happy/edge/error cases)
- Document findings in 30-review.md
- Provide verdict: APPROVED or CHANGES REQUIRED
- Re-validate after fixes if changes required
```

### Reviewer
- **Phase 4:** Review code quality and standards (MANDATORY for code changes)

**Reviewer Responsibilities:**
```
For all work packages with code changes:
- Review code against quality standards
- Verify architecture consistency
- Assess security concerns
- Evaluate documentation adequacy
- Check acceptance criteria met
- Document findings in 30-review.md
- Provide verdict: APPROVED or CHANGES REQUESTED
- Re-validate after fixes if changes requested
```

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

### Code Quality Review Gate (Phase 4 - NEW)
- Tester validation mandatory for code changes
- Reviewer validation mandatory for code changes
- Both must approve before acceptance
- Work incomplete if either finds blocking issues
- See [Code Quality Review Gate](../gates/35-code-quality-review.md)

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

## Artifact Persistence

**Principle:** Planning artifacts (requirements, designs, retrospectives) must be persisted to the repository when they transition from planning to implementation/completion.

### When to Persist Artifacts

**After Product Manager Phase:**
- PRDs, epics, user stories → `docs/product/[feature-name]/`
- See [Product Manager Role](../roles/product-manager.md) for details

**After Architect Phase:**
- Architecture docs, API specs, data models → `docs/architecture/[feature-name]/`
- ADRs → `docs/adr/`
- See [Architect Role](../roles/architect.md) for details

**After Bug Fix and Verification:**
- Bug investigation retrospectives → `docs/investigations/`
- See [Inspector Role](../roles/inspector.md) for details

### Why This Matters

**Long-Term Value:**
- Documents capture decisions for years, not just current work
- Engineers reference these during implementation
- Future teams understand context for "why" decisions were made
- Version control tracks evolution of requirements and design
- Single source of truth for product and technical specifications

**Temporary vs Permanent:**
```
.ai/tasks/                    → Temporary work-in-progress
docs/                         → Permanent, committed documentation

.ai/tasks/[task-id]/          → Active work, deleted after completion
docs/product/[feature]/       → Long-lived requirements
docs/architecture/[feature]/  → Long-lived technical design
docs/investigations/          → Long-lived learning and patterns
docs/adr/                     → Long-lived decision records
```

### Repository Structure

```
project-root/
├── docs/
│   ├── product/
│   │   └── [feature-name]/
│   │       ├── prd.md
│   │       ├── epics.md
│   │       └── user-stories.md
│   ├── architecture/
│   │   └── [feature-name]/
│   │       ├── architecture.md
│   │       ├── api-spec.md
│   │       └── data-models.md
│   ├── adr/
│   │   ├── 001-decision-title.md
│   │   ├── 002-decision-title.md
│   │   └── README.md
│   └── investigations/
│       ├── BUG-123-description.md
│       ├── BUG-456-description.md
│       └── README.md
└── .ai/
    └── tasks/ (temporary, not committed)
```

**Note:** Detailed persistence procedures are documented in:
- [Feature Workflow](feature.md) - Phase 0 artifact persistence
- [Bugfix Workflow](bugfix.md) - Retrospective persistence
- Individual role documents (Product Manager, Architect, Inspector)

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
