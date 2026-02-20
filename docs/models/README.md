# Model Performance Documentation

This directory contains per-model capability notes, observed strengths/weaknesses across roles, benchmark results, and recommended use cases.

## Available Models

| Model | Provider | Tier | Cost (per 1M tokens) | Documentation |
|-------|----------|------|----------------------|---------------|
| gpt-4o-mini | OpenAI | Tier 1 (Minimal) | ~$0.15 | [gpt-4o-mini.md](./gpt-4o-mini.md) |
| claude-haiku-4-5 | Anthropic | Tier 1 (Minimal) | ~$1.25 | [claude-haiku-4-5.md](./claude-haiku-4-5.md) |
| gpt-4o | OpenAI | Tier 2 (Low) | ~$2.50 | [gpt-4o.md](./gpt-4o.md) |
| claude-sonnet-4-5 | Anthropic | Tier 3 (Medium) | ~$3.00 | [claude-sonnet-4-5.md](./claude-sonnet-4-5.md) |
| claude-sonnet-4-6 | Anthropic | Tier 3 (Medium) | ~$3.00 | [claude-sonnet-4-6.md](./claude-sonnet-4-6.md) |
| claude-opus-4-6 | Anthropic | Tier 4 (High) | ~$15.00 | [claude-opus-4-6.md](./claude-opus-4-6.md) |
| o1-mini | OpenAI | Tier 3 (Medium) | ~$3.00 | [o1-mini.md](./o1-mini.md) |
| o3-mini | OpenAI | Tier 3 (Medium) | ~$3.00 | [o3-mini.md](./o3-mini.md) |

## Master Performance Matrix

> **Last Updated:** 2026-02-19  
> **Data Source:** `.claude/performance_grades/` and `.claude/performance_grades_backup/`

### Grade Legend
| Grade | Success Rate | Meaning |
|-------|-------------|---------|
| A | >90% | Excellent — deploy confidently |
| B | 75–90% | Good — solid production choice |
| C | 60–75% | Fair — works but watch for edge cases |
| D | 40–60% | Poor — frequent failures, use only when needed |
| F | <40% | Failing — do not use for this role |
| ? | No data | Not yet benchmarked |

### Model × Role Performance Table

| Role | gpt-4o-mini | claude-haiku-4-5 | gpt-4o | claude-sonnet-4-5 | claude-sonnet-4-6 | claude-opus-4-6 |
|------|-------------|-------------------|--------|-------------------|-------------------|-----------------|
| **engineer** | ? | ? | ? | F (33%) | ? | ? |
| **architect** | ? | ? | ? | F (33%) | ? | ? |
| **tester** | ? | ? | ? | F (33%) | ? | ? |
| **reviewer** | ? | ? | ? | ? | ? | ? |
| **orchestrator** | ? | ? | ? | F (33%) | ? | ? |
| **spelunker** | ? | ? | ? | F (33%) | ? | ? |
| **cartographer** | ? | ? | ? | F (25%) | ? | ? |
| **inspector** | ? | ? | ? | ? | ? | ? |
| **archaeologist** | ? | ? | ? | ? | ? | ? |
| **product-manager** | ? | ? | ? | ? | ? | ? |
| **designer** | ? | ? | ? | ? | ? | ? |
| **strategist** | ? | ? | ? | ? | ? | ? |

> **Note:** Most cells are `?` (no data) because benchmark runs are predominantly on `claude-sonnet-4-5-20250929` against a single project (`xasm++`). Expand coverage by running `ai-pack benchmark` across all roles and models.

### Detailed Grade Records (from `.claude/performance_grades/`)

| Model | Role | Project | Grade | Success Rate | Attempts | Last Tested |
|-------|------|---------|-------|-------------|----------|-------------|
| claude-sonnet-4-5-20250929 | engineer | ai-pack | F | 33% | 3 | 2026-02-19 |
| claude-sonnet-4-5-20250929 | engineer | xasm++ | F | 33% | 3 | 2026-02-18 |
| claude-sonnet-4-5-20250929 | architect | xasm++ | F | 33% | 12 | 2026-02-14 |
| claude-sonnet-4-5-20250929 | tester | xasm++ | F | 33% | 3 | 2026-02-14 |
| claude-sonnet-4-5-20250929 | spelunker | xasm++ | F | 33% | 3 | 2026-02-14 |
| claude-sonnet-4-5-20250929 | cartographer | xasm++ | F | 25% | 4 | 2026-02-14 |
| claude-sonnet-4-5-20250929 | orchestrator | xasm++ | F | 33% | 6 | 2026-02-03 |
| claude-sonnet-4-5-20250929 | reviewer | xasm++ | F | 33% | 3 | 2026-02-14 |

> ⚠️ **Context:** These grades reflect early benchmark runs on a single project (`xasm++`, a 6502 assembler). The low grades for `claude-sonnet-4-5` may be due to project-specific complexity, not general model capability. More diverse benchmarks needed.

## How to Update This Documentation

1. Run a benchmark session against a role/model combination
2. Update the grade record in the table above
3. Update the model's individual file in this directory
4. Update the role's file in `docs/roles/`
5. Commit with message: `docs(models): update benchmark grades <model> x <role> <date>`

See [MAINTAINING-DOCS.md](../MAINTAINING-DOCS.md) for full update procedures.
