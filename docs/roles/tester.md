# Tester Role — Model Performance

**Role:** `tester`  
**Description:** Test suite creation and quality assurance  
**Benchmark Task:** Models asked to write comprehensive test suites  
**Last Updated:** 2026-05-28

---

## Model Performance Grades

*Benchmarked: 14 models × 5 prompts each = 70 benchmark runs for this role.*  
*Project: `/Users/bryanw/Projects/Vibe/ai-pack`*

| Model | Tier | Grade | Pass Rate | Avg Latency | Avg Tokens |
|-------|------|-------|-----------|-------------|------------|
| gpt-4.1-nano | minimal | **A** | 100% | 1.7s | 276 |
| gpt-4.1-mini | low | **A** | 100% | 5.1s | 287 |
| claude-haiku-4-5 | minimal | **A** | 100% | 2.0s | 399 |
| o4-mini | low | **A** | 100% | 28.4s | 3557 |
| gpt-5.1-codex-mini | medium | **A** | 100% | 2.2s | 259 |
| gpt-4.1 | medium | **A** | 100% | 2.5s | 247 |
| claude-sonnet-4-5 | medium | **A** | 100% | 3.9s | 350 |
| claude-sonnet-4-5-20250929 | medium | **A** | 100% | 4.5s | 368 |
| claude-sonnet-4-6 | medium | **A** | 100% | 6.5s | 571 |
| gpt-5.1-codex | high | **A** | 100% | 15.3s | 382 |
| gpt-5.2-codex | high | **A** | 100% | 3.7s | 311 |
| claude-opus-4-5 | high | **A** | 100% | 3.1s | 298 |
| claude-opus-4-6 | high | **A** | 100% | 6.3s | 456 |
| gpt-4o-mini | minimal | **A** | 100% | 3.4s | 238 |

---

## Recommended Model by Use Case

| Use Case | Recommended Model | Rationale |
|----------|------------------|-----------|
| Cost-sensitive | `gpt-4.1-nano` or `claude-haiku-4-5` | A-grade, minimal tier |
| Default workloads | `gpt-4.1` or `claude-sonnet-4-6` | A-grade, medium tier, balanced |
| Complex / high-stakes | `claude-opus-4-6` | A-grade, premium quality |
| High throughput | `gpt-5.1-codex-mini` | Fastest (5.4s avg), A-grade |

---

## Escalation Path

```
gpt-4.1-nano → gpt-4.1-mini → gpt-4.1 → claude-sonnet-4-6 → claude-opus-4-6
```

---

## Benchmark Methodology

Each model ran **5 role-appropriate prompts** designed for `tester` tasks.  
Evaluation uses keyword + structure heuristics tailored to `tester` output expectations.  
Grade: A (≥90% pass), B (≥75%), C (≥60%), D (≥40%), F (<40%).

---

*Last updated: 2026-05-28 | Data from `.claude/performance_grades/`*
