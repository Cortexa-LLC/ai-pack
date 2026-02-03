---
paths: **/*
---

# AI-Pack Workflows

Choose the appropriate workflow based on task type.

## Workflow Selection

| Task Type | Workflow | When to Use |
|-----------|----------|-------------|
| **General** | `standard.md` | Default for tasks not fitting specialized workflows |
| **New Feature** | `feature.md` | Adding new functionality |
| **Bug Fix** | `bugfix.md` | Fixing defects or issues |
| **Refactoring** | `refactor.md` | Improving code structure without changing behavior |
| **Investigation** | `research.md` | Understanding code, architecture, or systems |

## Standard Workflow

**Path:** `.ai-pack/workflows/standard.md`

**Phases:**
1. Setup - Create task packet
2. Planning - Define approach
3. Execution - Implement solution
4. Quality Assurance - Tester + Reviewer validation
5. Completion - Sign-off

**Use for:** Most tasks

## Feature Workflow

**Path:** `.ai-pack/workflows/feature.md`

**Phases:**
0. **Product Definition & Architecture** (Optional for large features)
   - Step 0.1: Product Manager → PRD
   - Step 0.2: Architect → Architecture design
   - Step 0.3: Designer → UX wireframes
1. Setup
2. Planning
3. Implementation
4. Testing
5. Review
6. Completion

**Use for:** Adding new capabilities, enhancements

**Optional Planning Specialists:**
- **Product Manager** - Large/unclear features
- **Architect** - Complex technical design
- **Designer** - User-facing UI/UX

## Bugfix Workflow

**Path:** `.ai-pack/workflows/bugfix.md`

**Phases:**
0. **Investigation** (Optional for complex bugs)
   - Option A: Inspector → Root cause analysis
   - Option B: Engineer self-investigation
1. Setup
2. Fix Implementation
3. Regression Testing
4. Review
5. Completion

**Use for:** Defects, errors, incorrect behavior

**When to use Inspector:**
- Root cause unknown
- Complex system interaction
- Pattern detection needed

## Refactor Workflow

**Path:** `.ai-pack/workflows/refactor.md`

**Phases:**
1. Setup
2. Analysis - Identify improvement opportunities
3. Planning - Design refactoring approach
4. Implementation - Apply refactorings incrementally
5. Verification - Ensure behavior unchanged
6. Review
7. Completion

**Use for:** Code quality improvements, architecture changes

**Key principle:** Behavior stays the same

## Research Workflow

**Path:** `.ai-pack/workflows/research.md`

**Phases:**
1. Setup
2. Exploration - Understand codebase/system
3. Documentation - Record findings
4. Recommendations - Propose actions
5. Completion

**Use for:** Understanding code, investigating architecture, learning systems

**Output:** Documentation in `docs/investigations/`

## Workflow Components

### Common to All Workflows

**Quality Gates:**
- Task packet requirement
- Tester validation (for code changes)
- Reviewer validation (for code changes)

**Roles:**
- Orchestrator - Coordinates complex tasks
- Engineer - Implements solutions
- Tester - Validates tests
- Reviewer - Validates code quality

**Documentation:**
- Task packets in `.ai/tasks/`
- Planning artifacts in `docs/`

### Specialist Roles (Optional)

**Product Manager**
- When: Large features, unclear requirements
- Output: PRD in `docs/product/`
- Workflow: Feature (Phase 0)

**Architect**
- When: Architecture decisions, API design
- Output: Architecture docs in `docs/architecture/`, ADRs in `docs/adr/`
- Workflow: Feature (Phase 0)

**Designer**
- When: User-facing UI/UX needed
- Output: Wireframes and UX flows in `docs/design/`
- Workflow: Feature (Phase 0)

**Inspector**
- When: Complex bugs, unknown root cause
- Output: RCA in `docs/investigations/`
- Workflow: Bugfix (Phase 0)

## Choosing Your Workflow

**Ask yourself:**

1. **Is this a new feature?**
   → Use **feature.md**
   → Consider Product Manager for large features
   → Consider Architect for complex design
   → Consider Designer for UI/UX

2. **Is this fixing a bug?**
   → Use **bugfix.md**
   → Consider Inspector for complex investigation

3. **Is this improving code without changing behavior?**
   → Use **refactor.md**

4. **Is this understanding/exploring code?**
   → Use **research.md**

5. **None of the above?**
   → Use **standard.md**

## Reference

Full workflows: `.ai-pack/workflows/*.md`
