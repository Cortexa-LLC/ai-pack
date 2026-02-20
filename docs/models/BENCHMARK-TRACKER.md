# Model × Role Benchmark Master Tracker

> **Living document** — update after every benchmark run.  
> **Last Updated:** 2026-02-19  
> **Data Source:** `.claude/performance_grades/` and `.claude/performance_grades_backup/`

---

## Master Grade Matrix

> Grades: **A** (>90%) | **B** (75-90%) | **C** (60-75%) | **D** (40-60%) | **F** (<40%) | **?** (no data)

| Role | gpt-4o-mini | gpt-4o | claude-haiku-4-5 | claude-sonnet-4-5 | claude-sonnet-4-6 | claude-opus-4-6 |
|------|:-----------:|:------:|:----------------:|:-----------------:|:-----------------:|:---------------:|
| **engineer** | ? | ? | ? | F ⚠️ | ? | ? |
| **architect** | ? | ? | ? | F ⚠️ | ? | ? |
| **tester** | ? | ? | ? | F ⚠️ | ? | ? |
| **reviewer** | ? | ? | ? | F ⚠️ | ? | ? |
| **orchestrator** | ? | ? | ? | F ⚠️ | ? | ? |
| **spelunker** | ? | ? | ? | F ⚠️ | ? | ? |
| **cartographer** | ? | ? | ? | F ⚠️ | ? | ? |
| **inspector** | ? | ? | ? | ? | ? | ? |
| **archaeologist** | ? | ? | ? | ? | ? | ? |
| **product-manager** | ? | ? | ? | ? | ? | ? |
| **designer** | ? | ? | ? | ? | ? | ? |
| **strategist** | ? | ? | ? | ? | ? | ? |

⚠️ = F grade from existing benchmark data (see detailed records below)

---

## Detailed Benchmark Records

Every benchmark run should produce one row here.

| Date | Model | Role | Project | Success Rate | Attempts | Confidence | Grade | Notes |
|------|-------|------|---------|-------------|----------|------------|-------|-------|
| 2026-02-03 | claude-sonnet-4-5 | orchestrator | xasm++ | 33% (2/6) | 6 | 0.30 | F | 4 failures, 0 retries |
| 2026-02-14 | claude-sonnet-4-5 | architect | xasm++ | 33% (4/12) | 12 | 0.60 | F | Most attempts in dataset |
| 2026-02-14 | claude-sonnet-4-5 | tester | xasm++ | 33% (1/3) | 3 | 0.15 | F | Low confidence |
| 2026-02-14 | claude-sonnet-4-5 | spelunker | xasm++ | 33% (1/3) | 3 | 0.15 | F | Model since replaced by 4-6 |
| 2026-02-14 | claude-sonnet-4-5 | cartographer | xasm++ | 25% (1/4) | 4 | 0.20 | F | Lowest success rate overall |
| 2026-02-14 | claude-sonnet-4-5 | reviewer | xasm++ | 33% (1/3) | 3 | 0.15 | F | Low confidence |
| 2026-02-18 | claude-sonnet-4-5 | engineer | xasm++ | 33% (1/3) | 3 | 0.15 | F | Low confidence |
| 2026-02-19 | claude-sonnet-4-5 | engineer | ai-pack | 33% (1/3) | 3 | 0.15 | F | 67% error rate on ai-pack project |

---

## Statistical Summary

```
Total benchmark runs (rows):  8
Models benchmarked:           1 of 6 (claude-sonnet-4-5 only)
Roles benchmarked:            7 of 12 (missing: inspector, archaeologist, PM, designer, strategist)
Projects benchmarked:         2 (xasm++, ai-pack)
Overall pass rate:            ~32% across all runs
Highest success rate:         33% (tied across most roles)
Lowest success rate:          25% (cartographer)
Most confident grade:         architect (0.60, 12 attempts)
Least confident grades:       many roles at 0.15 (3 attempts)
```

---

## Coverage Gaps Priority List

Prioritized by impact and current knowledge gap:

| Priority | Gap | Why It Matters |
|----------|-----|----------------|
| 🔴 HIGH | gpt-4o-mini on engineer role | It's the configured default — we're flying blind |
| 🔴 HIGH | gpt-4o-mini on tester role | Same — needs baseline data |
| 🔴 HIGH | claude-sonnet-4-6 on spelunker | The designated model has zero benchmarks |
| 🟡 MED | claude-opus-4-6 on architect | Most complex role needs premium model tested |
| 🟡 MED | claude-opus-4-6 on reviewer | reviewer-opus.md exists but no data |
| 🟡 MED | claude-sonnet-4-5 on a simple web project | Validates whether xasm++ grades are outliers |
| 🟢 LOW | inspector, archaeologist, PM, designer, strategist | No benchmarks at all — start with defaults |

---

## How to Update This File

### After a Benchmark Run

1. **Add a row** to the Detailed Benchmark Records table with all fields
2. **Update the Grade Matrix** cell for the model × role intersection
3. **Update the model file** in `docs/models/<model>.md`
4. **Update the role file** in `docs/roles/<role>.md`
5. **Update the Summary Statistics** block
6. **Remove the row** from Coverage Gaps if covered

### Grade Calculation

```
success_rate = successes / total_attempts
grade = A if success_rate > 0.90
      = B if success_rate > 0.75
      = C if success_rate > 0.60
      = D if success_rate > 0.40
      = F otherwise

confidence = min(1.0, total_attempts / 20)
  # 0.15 = 3 attempts, 0.30 = 6, 0.60 = 12, 1.0 = 20+
```

### Data Sources

- Grade files: `.claude/performance_grades/<model>_<role>_<project>.json`
- Backup grades: `.claude/performance_grades_backup/`
- Usage metrics: `.claude/metrics/daily/YYYY-MM-DD.json`

### JSON Grade File Format

```json
{
  "model": "claude-sonnet-4-5-20250929",
  "role": "engineer",
  "project": "ai-pack",
  "total_attempts": 3,
  "successes": 1,
  "failures": 2,
  "errors": 2,
  "success_rate": 0.33,
  "grade": "F",
  "confidence": 0.15,
  "avg_execution_time": 283.4,
  "last_updated": "2026-02-19T...",
  "escalations": 0,
  "downgrades": 0
}
```

---

## Benchmark Session Log

Track when benchmark campaigns were run:

| Session Date | Operator | Models Tested | Roles Tested | Projects | Notes |
|-------------|---------|--------------|--------------|---------|-------|
| ~2026-02-03 | auto | claude-sonnet-4-5 | orchestrator | xasm++ | First known benchmark |
| ~2026-02-14 | auto | claude-sonnet-4-5 | architect, tester, spelunker, cartographer, reviewer | xasm++ | Batch benchmark |
| 2026-02-18 | auto | claude-sonnet-4-5 | engineer | xasm++ | Follow-up run |
| 2026-02-19 | auto | claude-sonnet-4-5 | engineer | ai-pack | New project added |
| _(next)_ | | | | | Run gpt-4o-mini across all roles on a web project |

---

*Maintained by: Engineering team | Update frequency: After each benchmark run | Review: Monthly*
