# Role Performance Documentation

This directory contains per-role model performance notes — which models work best/worst for each role, and why.

## Quick Reference

| Role | Default Model | Best Performing | Worst Performing | Benchmarks Available |
|------|--------------|-----------------|------------------|----------------------|
| [engineer](./engineer.md) | gpt-4o-mini | ? (no data) | claude-sonnet-4-5 (F, 33%) | ✅ Partial |
| [architect](./architect.md) | ? | ? | claude-sonnet-4-5 (F, 33%) | ✅ Partial |
| [tester](./tester.md) | ? | ? | claude-sonnet-4-5 (F, 33%) | ✅ Partial |
| [reviewer](./reviewer.md) | ? | ? | claude-sonnet-4-5 (F, 33%) | ✅ Partial |
| [orchestrator](./orchestrator.md) | ? | ? | claude-sonnet-4-5 (F, 33%) | ✅ Partial |
| [spelunker](./spelunker.md) | claude-sonnet-4-6 | ? | claude-sonnet-4-5 (F, 33%) | ✅ Partial |
| [cartographer](./cartographer.md) | ? | ? | claude-sonnet-4-5 (F, 25%) | ✅ Partial |
| [inspector](./inspector.md) | ? | ? | ? | ❌ None |
| [archaeologist](./archaeologist.md) | ? | ? | ? | ❌ None |
| [product-manager](./product-manager.md) | gpt-4o-mini | ? | ? | ❌ None |
| [designer](./designer.md) | ? | ? | ? | ❌ None |
| [strategist](./strategist.md) | ? | ? | ? | ❌ None |

## Grade Legend

| Grade | Success Rate | Recommendation |
|-------|-------------|----------------|
| A | >90% | Strongly recommended — default choice |
| B | 75–90% | Recommended — reliable for production |
| C | 60–75% | Acceptable — monitor for edge cases |
| D | 40–60% | Marginal — consider alternatives |
| F | <40% | Failing — do not use |
| ? | No data | Benchmark needed |

## Understanding the Data

### What "Grade" Means

Grades are computed from benchmark runs tracked in `.claude/performance_grades/`. Each grade file covers one model × role × project combination. A grade of **F** does not necessarily mean the model is bad — it may mean:

1. The benchmark project (`xasm++`, a 6502 assembler) is unusually specialized
2. The benchmark sample size is too small (3-12 attempts)
3. The model was tested before role prompts were tuned

### Confidence Scores

Each grade includes a **confidence score** (0.0–1.0) indicating how reliable the grade is based on sample size:
- `0.15` = Very low confidence (only 3 attempts)
- `0.30` = Low confidence (6 attempts)
- `0.60` = Moderate confidence (12 attempts)
- `0.80+` = High confidence (20+ attempts)

All current benchmark data has low confidence due to small sample sizes.

## How to Add New Benchmark Data

1. Run a task with a specific model/role combination
2. The system records results in `.claude/performance_grades/<model>_<role>_<project>.json`
3. Update the relevant role file in this directory with the new grade
4. Update `docs/models/<model>.md` with the cross-reference
5. Update the master matrix in `docs/models/README.md`

See [MAINTAINING-DOCS.md](../MAINTAINING-DOCS.md) for full procedures.
