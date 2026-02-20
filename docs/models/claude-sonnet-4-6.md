# claude-sonnet-4-6 Performance Profile

**Model ID:** `claude-sonnet-4-6`  
**Provider:** Anthropic  
**Tier:** 3 (Medium Cost)  
**Cost:** ~$3.00 per 1M input tokens  
**Context Window:** 200K tokens  
**Last Updated:** 2026-02-19

---

## Overview

Claude Sonnet 4.6 is the successor to claude-sonnet-4-5, featuring improvements in reasoning, instruction following, and task completion. It is the designated model for the **spelunker** role (explicitly set in `roles/spelunker.md`) and is increasingly used alongside claude-sonnet-4-5 in production.

---

## Performance Grades by Role

| Role | Grade | Success Rate | Attempts | Project | Last Tested | Notes |
|------|-------|-------------|----------|---------|-------------|-------|
| spelunker | ? | No data | — | — | — | Designated model, benchmarks pending |
| engineer | ? | No data | — | — | — | Used in escalation path |
| orchestrator | ? | No data | — | — | — | Used in escalation path |

> **Note:** claude-sonnet-4-6 began appearing in usage metrics on 2026-02-18 (327 calls) and 2026-02-19 (161 calls), but no performance grades have been recorded yet in `.claude/performance_grades/`.

---

## Usage Metrics (from daily metrics)

| Date | Calls | Input Tokens | Output Tokens | Cost (est.) |
|------|-------|-------------|---------------|-------------|
| 2026-02-18 | 327 | — | — | — |
| 2026-02-19 | 161 | — | — | — |

---

## Observed Strengths

- **Designated for spelunker role:** Explicitly selected in role configuration as best-fit
- **Improved over 4-5:** Anthropic reports improvements in reasoning and code understanding
- **Large context window:** 200K tokens for large codebase analysis
- **Recent model:** More up-to-date training data than 4-5 variant

## Observed Weaknesses

- **No benchmark data yet:** Cannot make evidence-based claims
- **Same cost tier as 4-5:** No cost advantage over predecessor
- **Limited track record:** Newer model with less production history in this system

---

## Recommended Use Cases

### ✅ Use claude-sonnet-4-6 For:
- **Spelunker role** (explicitly recommended in role spec)
- Runtime investigation and debugging
- Tasks where claude-sonnet-4-5 fails
- Large codebase analysis

### ❌ Avoid For:
- High-volume, simple tasks (use gpt-4o-mini)
- When claude-sonnet-4-5 works fine (no cost benefit)
- Until benchmark data confirms improvement over 4-5

---

## Escalation Path

```
claude-sonnet-4-5 → claude-sonnet-4-6 → claude-opus-4-6
```

---

## Notes for Future Benchmarks

- [ ] Run full benchmark suite across all roles
- [ ] Compare directly against claude-sonnet-4-5 on same tasks
- [ ] Focus initial benchmarks on spelunker role (designated use case)
- [ ] Track cost differences vs. 4-5 variant

---

*Last updated: 2026-02-19 | Data source: `.claude/metrics/daily/` (no grade files yet)*
