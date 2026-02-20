# Spelunker Role — Model Performance

**Role:** `spelunker`  
**Description:** Runtime investigation specialist for live systems and production issue exploration  
**Configured Model:** `claude-sonnet-4-6` (explicitly set in `roles/spelunker.md`)  
**Max Budget Tokens:** 1,000,000  
**Max Turns:** 250  
**Last Updated:** 2026-02-19

---

## Model Performance Grades

| Model | Grade | Success Rate | Attempts | Confidence | Project | Last Tested | Notes |
|-------|-------|-------------|----------|------------|---------|-------------|-------|
| claude-sonnet-4-5 | **F** | 33% (1/3) | 3 | 0.15 (low) | xasm++ | 2026-02-14 | Old version, replaced by 4-6 |
| claude-sonnet-4-6 | ? | — | 0 | — | — | — | **Designated model**, no benchmarks yet |
| claude-opus-4-6 | ? | — | 0 | — | — | — | Potential upgrade path |

---

## Role-Specific Model Considerations

### Why claude-sonnet-4-6 is Hardcoded
The spelunker role has an **explicit model assignment** (`Model: claude-sonnet-4-6` in the role spec), unlike most roles which use a configurable default. This indicates:
1. The role requires higher reasoning capability than gpt-4o-mini can provide
2. claude-sonnet-4-6 was evaluated as the best fit for runtime investigation
3. Cost/capability trade-off was consciously decided in favor of quality

### Spelunker Task Profile
Runtime investigation requires:
- **Large context understanding** — reading logs, traces, stack frames
- **Deductive reasoning** — piecing together evidence from multiple sources
- **Pattern recognition** — identifying anomalies in runtime behavior
- **Actionable output** — producing specific root cause + fix recommendations

These needs align well with sonnet-tier models (200K context, strong reasoning).

### Why claude-sonnet-4-5 Failed on xasm++
The old grade (F, 33%) for claude-sonnet-4-5 reflects:
- Pre-update model on a specialized domain
- Small sample (3 attempts)
- Since replaced by claude-sonnet-4-6 as the designated model

Do not use this grade to judge the current spelunker configuration.

---

## Configuration Notes

```yaml
# From roles/spelunker.md
Model: claude-sonnet-4-6
MaxBudgetTokens: 1000000  # 1M tokens — generous for thorough investigation
MaxTurns: 250             # High turn count — complex investigations take time
Timeout: 10min
```

The high token budget and turn count reflect the nature of deep investigation tasks.

---

## Task Types and Model Fit

| Task Type | Recommended Model | Notes |
|-----------|------------------|-------|
| Production bug investigation | claude-sonnet-4-6 | Default, designated |
| Log analysis | claude-sonnet-4-6 | Large context needed |
| Performance root cause | claude-sonnet-4-6 | Multi-factor analysis |
| Security incident investigation | claude-opus-4-6 | High stakes, premium reasoning |
| Simple runtime errors | claude-sonnet-4-5 | Could downgrade if costs are concern |

---

## Benchmark History

### 2026-02-14: claude-sonnet-4-5 on xasm++ (Spelunker)
- **Result:** F (33% success rate) — **outdated, model since replaced**
- **Sample:** 3 attempts, 1 success, 2 failures
- **Confidence:** 0.15 (very low)

---

## Gaps in Data

- [ ] **Critical:** No benchmarks for claude-sonnet-4-6 (the actual designated model)
- [ ] No benchmarks on production-scale investigation scenarios
- [ ] Need baseline: what does a "successful spelunker run" look like vs. failure?

---

*Last updated: 2026-02-19 | Grade data from `.claude/performance_grades_backup/`*
