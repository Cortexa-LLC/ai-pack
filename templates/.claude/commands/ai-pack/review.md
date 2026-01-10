---
description: Start code review workflow with Reviewer role
---

# /ai-pack review - Code Review

Activates the **Reviewer role** to conduct quality assurance on completed work.

## When to Use This

Use after:
- Engineer has completed implementation
- Tester has validated tests (use `/ai-pack test` first)
- Self-review is complete
- All acceptance criteria believed met

**This is MANDATORY before work can be accepted.**

## Reviewer Responsibilities

### Phase 1: Preparation

1. Read the task packet:
   - `.ai/tasks/*/00-contract.md` - Requirements
   - `.ai/tasks/*/10-plan.md` - Implementation plan
   - `.ai/tasks/*/20-work-log.md` - Execution notes

2. Read the Tester report:
   - `.ai/tasks/*/30-review.md` - Check Tester verdict first
   - If Tester BLOCKED, resolve issues before code review

### Phase 2: Code Review

Review code against these criteria:

#### 🔴 BLOCKING Issues (Must Fix)

- **Security vulnerabilities** (injection, XSS, auth bypass, etc.)
- **Major clean code violations** (god functions, tight coupling)
- **Architecture inconsistencies** (violates patterns)
- **Missing critical tests** (no coverage for critical paths)
- **Broken functionality** (doesn't meet acceptance criteria)

#### 🟡 MAJOR Issues (Should Fix)

- **Moderate clean code violations** (magic numbers, unclear names)
- **Missing documentation** (complex logic unexplained)
- **Performance concerns** (obvious inefficiencies)
- **Incomplete test coverage** (<80% overall)
- **Inconsistent patterns** (doesn't follow project conventions)

#### 🟢 MINOR Issues (Nice to Fix)

- **Style inconsistencies** (formatting, naming preferences)
- **Optimization opportunities** (not performance-critical)
- **Enhanced documentation** (already adequate but could improve)

### Phase 3: Language-Specific Checks

**For C# Projects (MANDATORY):**

```bash
# 1. Check formatting
dotnet csharpier . --check

# 2. Check build with analyzers
dotnet build /warnaserror

# 3. Verify no obsolete packages
grep -r "StyleCop.Analyzers" *.csproj
```

**Required files:**
- `.editorconfig` - Analyzer configuration
- `.csharpierrc.json` - Formatter configuration

See: `.ai-pack/quality/clean-code/csharp-modern-tooling.md`

### Phase 4: Documentation

Document findings in `.ai/tasks/*/30-review.md`:

```markdown
## Reviewer Assessment

**Reviewer:** [Your name]
**Date:** [Date]
**Commit:** [commit hash]

### Verdict: [APPROVED | CHANGES REQUESTED]

### Issues Found

#### Blocking (Must Fix)
- [Issue 1] - [Description] - [Location]

#### Major (Should Fix)
- [Issue 2] - [Description] - [Location]

#### Minor (Nice to Fix)
- [Issue 3] - [Description] - [Location]

### Positive Observations
- [What was done well]

### Next Steps
- [What Engineer should do]
```

### Phase 5: Verdict

**APPROVED:**
- No blocking issues
- ≤2 major issues (documented but acceptable)
- Minor issues only (optional fixes)
- Work meets acceptance criteria

**CHANGES REQUESTED:**
- Any blocking issues present
- >2 major issues
- Acceptance criteria not met

## Reference Documentation

**Primary:** [.ai-pack/roles/reviewer.md](../../.ai-pack/roles/reviewer.md)

**Gates:**
- [.ai-pack/gates/35-code-quality-review.md](../../.ai-pack/gates/35-code-quality-review.md)

**Standards:**
- [.ai-pack/quality/engineering-standards.md](../../.ai-pack/quality/engineering-standards.md)
- [.ai-pack/quality/clean-code/](../../.ai-pack/quality/clean-code/)

## Related Commands

- `/ai-pack test` - Run Tester validation first
- `/ai-pack task-status` - Check task progress
- `/ai-pack engineer` - Return to Engineer role to fix issues
- `/ai-pack help` - Show all commands

## Activation

This command will:
1. Load the Reviewer role definition
2. Verify Tester has validated
3. Guide you through code review checklist
4. Help document findings

Ready to review this code?
