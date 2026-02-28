# Model Performance Documentation

This directory contains per-model capability notes, observed strengths/weaknesses across roles, and recommended use cases.

## Available Models

| Model | Provider | Tier | Cost (per 1M tokens) | Documentation |
|-------|----------|------|----------------------|---------------|
| gpt-4o-mini | OpenAI | Tier 1 (Minimal) | ~$0.15 | [gpt-4o-mini.md](./gpt-4o-mini.md) |
| claude-haiku-4-5 | Anthropic | Tier 1 (Minimal) | ~$1.25 | [claude-haiku-4-5.md](./claude-haiku-4-5.md) |
| gpt-4o | OpenAI | Tier 2 (Low) | ~$2.50 | [gpt-4o.md](./gpt-4o.md) |
| claude-sonnet-4-5 | Anthropic | Tier 3 (Medium) | ~$3.00 | [claude-sonnet-4-5.md](./claude-sonnet-4-5.md) |
| claude-sonnet-4-6 | Anthropic | Tier 3 (Medium) | ~$3.00 | [claude-sonnet-4-6.md](./claude-sonnet-4-6.md) |
| claude-opus-4-6 | Anthropic | Tier 4 (High) | ~$15.00 | [claude-opus-4-6.md](./claude-opus-4-6.md) |

## Grade Legend

| Grade | Success Rate | Meaning |
|-------|-------------|---------|
| A | >90% | Excellent — deploy confidently |
| B | 75–90% | Good — solid production choice |
| C | 60–75% | Fair — works but watch for edge cases |
| D | 40–60% | Poor — frequent failures, use only when needed |
| F | <40% | Failing — do not use for this role |

## How Grades Are Assigned

Grades come from two sources:

1. **Production data** — recorded automatically from live task executions. Each task completion updates `.claude/performance_grades/<model>_<role>_<project>.json`.
2. **Cold-start seeds** — new models start at Grade C (`source: "cold_start"`). See `docs/adr/005-grade-seeding-redesign.md` for the seeding policy.

The grade selector picks the **cheapest model with Grade A or B** for each role. Run `python3 scripts/seed-grades.py` after adding a new model to initialize its cold-start grade files.
