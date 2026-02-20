# Maintaining Model Performance Documentation

This guide explains how to keep `docs/models/` and `docs/roles/` up to date as new benchmark data comes in.

---

## Overview

Model performance documentation lives in two places:

```
docs/
├── models/
│   ├── README.md                  ← Master overview + grade matrix
│   ├── BENCHMARK-TRACKER.md       ← Detailed benchmark records (PRIMARY update target)
│   ├── gpt-4o-mini.md             ← Per-model capability notes
│   ├── gpt-4o.md
│   ├── claude-haiku-4-5.md
│   ├── claude-sonnet-4-5.md
│   ├── claude-sonnet-4-6.md
│   └── claude-opus-4-6.md
└── roles/
    ├── README.md                  ← Role overview + quick reference table
    ├── engineer.md                ← Per-role model performance notes
    ├── architect.md
    ├── tester.md
    ├── reviewer.md
    ├── orchestrator.md
    ├── spelunker.md
    ├── cartographer.md
    ├── inspector.md
    ├── archaeologist.md
    ├── product-manager.md
    ├── designer.md
    └── strategist.md
```

---

## Update Procedure

### After Each Benchmark Run

**Minimum required (5 minutes):**

1. **Add a row to `docs/models/BENCHMARK-TRACKER.md`**

   ```markdown
   | 2026-03-01 | gpt-4o-mini | engineer | my-project | 80% (8/10) | 10 | 0.50 | B | First gpt-4o-mini benchmark |
   ```

2. **Update the Grade Matrix** in `BENCHMARK-TRACKER.md` — change `?` to the grade letter

**Full update (15 minutes):**

3. **Update the model file** in `docs/models/<model>.md`:
   - Add a row to "Performance Grades by Role" table
   - Add an entry to "Benchmark History" section
   - Update Benchmark Summary Statistics

4. **Update the role file** in `docs/roles/<role>.md`:
   - Add a row to "Model Performance Grades" table
   - Add an entry to "Benchmark History" section
   - Update recommendations if grade changes expectations

5. **Update `docs/models/README.md`**:
   - Update the detailed grade records table

6. **Update `docs/roles/README.md`**:
   - Update Best/Worst Performing columns for the role

---

## Grade Calculation Reference

```python
def calculate_grade(successes, total):
    rate = successes / total
    if rate > 0.90: return 'A'
    if rate > 0.75: return 'B'
    if rate > 0.60: return 'C'
    if rate > 0.40: return 'D'
    return 'F'

def calculate_confidence(total):
    return min(1.0, total / 20)
    # 3 attempts  → 0.15 (very low)
    # 6 attempts  → 0.30 (low)
    # 10 attempts → 0.50 (moderate)
    # 12 attempts → 0.60 (moderate)
    # 20 attempts → 1.00 (high)
```

---

## Reading Grade Files

Benchmark data is stored in `.claude/performance_grades/`:

```bash
# List all grade files
ls .claude/performance_grades/

# Read a specific grade file
cat .claude/performance_grades/claude-sonnet-4-5-20250929_engineer_ai-pack.json
```

Grade file format:
```json
{
  "model": "claude-sonnet-4-5-20250929",
  "role": "engineer",
  "project": "ai-pack",
  "total_attempts": 3,
  "successes": 1,
  "failures": 2,
  "errors": 2,
  "success_rate": 0.3333,
  "grade": "F",
  "confidence": 0.15,
  "avg_execution_time": 283.4,
  "last_updated": "2026-02-19T12:34:56Z",
  "escalations": 0,
  "downgrades": 0
}
```

---

## Commit Convention

```bash
# After updating docs
git add docs/models/ docs/roles/
git commit -m "docs(models): update benchmark grades <model> x <role> <date>

- <model> on <role> @ <project>: <grade> (<success_rate>%, <n> attempts)
- Updated BENCHMARK-TRACKER.md, <model>.md, roles/<role>.md"
```

---

## Monthly Review Checklist

At the start of each month, review:

- [ ] Are there new models to add? (Check Anthropic/OpenAI release notes)
- [ ] Are any grades now outdated? (>6 months old)
- [ ] Are there Coverage Gaps that should be prioritized?
- [ ] Have any role configurations changed? (Check `roles/*.md` git history)
- [ ] Should any "?" cells be promoted to benchmark priority?

---

## Automated Updates (Future)

When the benchmark runner is enhanced, it should automatically:
1. Write to `.claude/performance_grades/<model>_<role>_<project>.json`
2. Trigger a documentation update script: `scripts/update-perf-docs.sh`
3. Open a PR with updated docs if grades change significantly

Until then, manual updates per this guide are required.

---

*Last updated: 2026-02-19*
