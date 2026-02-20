# Reviewer Role — Model Performance

**Role:** `reviewer`  
**Description:** Code review and improvement suggestions  
**Benchmark Task:** Models asked to identify code issues and suggest improvements  
**Last Updated:** 2026-05-28

---

## Model Performance Grades

*Benchmarked: 14 models × 5 prompts each = 70 benchmark runs for this role.*  
*Project: `/Users/bryanw/Projects/Vibe/ai-pack`*

| Model | Tier | Grade | Pass Rate | Avg Latency | Avg Tokens |
|-------|------|-------|-----------|-------------|------------|
| gpt-4.1-nano | minimal | **A** | 100% | 1.4s | 213 |
| gpt-4.1-mini | low | **A** | 100% | 3.0s | 281 |
| claude-haiku-4-5 | minimal | **A** | 100% | 2.7s | 467 |
| o4-mini | low | **A** | 100% | 29.1s | 3394 |
| gpt-5.1-codex-mini | medium | **A** | 100% | 2.4s | 301 |
| gpt-4.1 | medium | **A** | 100% | 2.7s | 254 |
| claude-sonnet-4-5 | medium | **A** | 100% | 4.4s | 388 |
| claude-sonnet-4-5-20250929 | medium | **A** | 100% | 4.3s | 370 |
| claude-sonnet-4-6 | medium | **A** | 100% | 6.0s | 564 |
| gpt-5.1-codex | high | **A** | 100% | 8.4s | 323 |
| gpt-5.2-codex | high | **A** | 100% | 3.9s | 297 |
| claude-opus-4-5 | high | **A** | 100% | 3.2s | 300 |
| claude-opus-4-6 | high | **A** | 100% | 5.3s | 393 |
| gpt-4o-mini | minimal | **A** | 100% | 3.9s | 252 |

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

Each model ran **5 role-appropriate prompts** designed for `reviewer` tasks.  
Evaluation uses keyword + structure heuristics tailored to `reviewer` output expectations.  
Grade: A (≥90% pass), B (≥75%), C (≥60%), D (≥40%), F (<40%).

---

*Last updated: 2026-05-28 | Data from `.claude/performance_grades/`*
