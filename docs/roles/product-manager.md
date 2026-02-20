# Product Manager Role — Model Performance

**Role:** `product-manager`  
**Description:** Requirements specification and user story creation  
**Benchmark Task:** Models asked to write structured requirements from user needs  
**Last Updated:** 2026-05-28

---

## Model Performance Grades

*Benchmarked: 14 models × 5 prompts each = 70 benchmark runs for this role.*  
*Project: `/Users/bryanw/Projects/Vibe/ai-pack`*

| Model | Tier | Grade | Pass Rate | Avg Latency | Avg Tokens |
|-------|------|-------|-----------|-------------|------------|
| gpt-4.1-nano | minimal | **A** | 100% | 6.3s | 776 |
| gpt-4.1-mini | low | **A** | 100% | 11.4s | 916 |
| claude-haiku-4-5 | minimal | **A** | 100% | 11.2s | 1138 |
| o4-mini | low | **A** | 100% | 10.4s | 1356 |
| gpt-5.1-codex-mini | medium | **A** | 100% | 6.7s | 1091 |
| gpt-4.1 | medium | **A** | 100% | 13.8s | 797 |
| claude-sonnet-4-5 | medium | **A** | 100% | 23.7s | 1138 |
| claude-sonnet-4-5-20250929 | medium | **A** | 100% | 24.5s | 1138 |
| claude-sonnet-4-6 | medium | **A** | 100% | 23.3s | 1139 |
| gpt-5.1-codex | high | **A** | 100% | 23.3s | 942 |
| gpt-5.2-codex | high | **A** | 100% | 16.7s | 972 |
| claude-opus-4-5 | high | **A** | 100% | 21.6s | 1138 |
| claude-opus-4-6 | high | **A** | 100% | 24.7s | 1139 |
| gpt-4o-mini | minimal | **A** | 100% | 10.7s | 793 |

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

Each model ran **5 role-appropriate prompts** designed for `product-manager` tasks.  
Evaluation uses keyword + structure heuristics tailored to `product-manager` output expectations.  
Grade: A (≥90% pass), B (≥75%), C (≥60%), D (≥40%), F (<40%).

---

*Last updated: 2026-05-28 | Data from `.claude/performance_grades/`*
