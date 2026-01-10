---
paths: **/*
---

# AI-Pack Mandatory Gates

All work must follow these quality gates from the ai-pack framework.

## Gate 1: Task Packet Requirement (BLOCKING)

**BEFORE starting ANY non-trivial task:**

```bash
# Create task packet
/ai-pack task-init <task-name>

# Fill out contract and plan
# Edit .ai/tasks/YYYY-MM-DD_task-name/00-contract.md
# Edit .ai/tasks/YYYY-MM-DD_task-name/10-plan.md
```

**Non-trivial = Any task that:**
- Requires >2 steps
- Involves code changes
- Takes >30 minutes
- Needs verification

**Enforcement:** Implementation is BLOCKED without task packet.

**Reference:** `.ai-pack/gates/00-global-gates.md`

---

## Gate 2: Execution Strategy (For 3+ Subtasks)

**For tasks with 3+ independent subtasks:**

1. **Analyze parallelization opportunities** BEFORE execution
2. **Document strategy** in `.ai/tasks/*/10-plan.md`:
   - Which subtasks can run in parallel
   - Which must be sequential
   - Shared resource coordination (builds, databases, coverage)
3. **Execute accordingly** - Use Task tool for parallel work

**Enforcement:** Cannot execute without documented strategy.

**Reference:** `.ai-pack/gates/25-execution-strategy.md`

---

## Gate 3: Code Quality Review (BLOCKING)

**For ALL code changes, MANDATORY validations:**

### Step 1: Tester Validation
```bash
/ai-pack test
```
- TDD process verification
- Coverage validation (80-90% overall, 95%+ critical)
- Test quality assessment
- **Must receive APPROVED verdict**

### Step 2: Reviewer Validation
```bash
/ai-pack review
```
- Code quality verification
- Standards compliance
- Security review
- **Must receive APPROVED verdict**

**Enforcement:** Work status = INCOMPLETE until both approve.

**Reference:** `.ai-pack/gates/35-code-quality-review.md`

---

## Gate 4: Artifact Persistence (For Planning Specialists)

**When specialists used (PM, Architect, Designer, Inspector):**

**MUST persist artifacts to `docs/`:**
- Product Manager → `docs/product/*.md`
- Architect → `docs/architecture/*.md` + `docs/adr/*.md`
- Designer → `docs/design/[feature]/`
- Inspector → `docs/investigations/*.md`

**MUST cross-reference documents:**
- PRD ↔ Architecture
- Architecture ↔ ADRs
- Design ↔ PRD
- Investigation ↔ Architecture

**MUST commit to repository.**

**Enforcement:** Orchestrator BLOCKS implementation until persistence complete.

**Reference:** `.ai-pack/gates/10-persistence.md`

---

## C# Projects: Modern Tooling Gate (BLOCKING)

**Before committing C# code, run:**

```bash
# Step 1: Format
dotnet csharpier .

# Step 2: Build with analyzers
dotnet build /warnaserror

# Step 3: Test
dotnet test

# ALL must pass
```

**Required configuration:**
- `.editorconfig` - Analyzer rules
- `.csharpierrc.json` - Formatter settings
- Modern tooling packages (CSharpier, Roslynator)

**Enforcement:** Reviewer BLOCKS if not followed.

**Reference:** `.ai-pack/quality/clean-code/csharp-modern-tooling.md`

---

## Quick Gate Checklist

Before marking work complete:

- [ ] Task packet created and updated
- [ ] Execution strategy documented (if 3+ subtasks)
- [ ] TDD process followed (RED-GREEN-REFACTOR)
- [ ] Tester validated (APPROVED)
- [ ] Reviewer validated (APPROVED)
- [ ] Planning artifacts persisted (if specialists used)
- [ ] C# tooling passed (if C# project)

**All gates must pass before work is accepted.**
