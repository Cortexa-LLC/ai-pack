# Architect Role — Model Performance

**Role:** `architect`  
**Description:** System design and architecture decisions  
**Benchmark Task:** Models asked to design systems with explicit trade-offs  
**Last Updated:** 2026-05-28

---

## Model Performance Grades

*Benchmarked: 14 models × 5 prompts each = 70 benchmark runs for this role.*  
*Project: `/Users/bryanw/Projects/Vibe/ai-pack`*

| Model | Tier | Grade | Pass Rate | Avg Latency | Avg Tokens |
|-------|------|-------|-----------|-------------|------------|
| gpt-4.1-nano | minimal | **A** | 100% | 8.7s | 1122 |
| gpt-4.1-mini | low | **A** | 100% | 17.0s | 1122 |
| claude-haiku-4-5 | minimal | **A** | 100% | 7.9s | 1126 |
| o4-mini | low | **A** | 100% | 49.6s | 6646 |
| gpt-5.1-codex-mini | medium | **A** | 100% | 7.8s | 1117 |
| gpt-4.1 | medium | **A** | 100% | 22.4s | 1122 |
| claude-sonnet-4-5 | medium | **A** | 100% | 18.5s | 1126 |
| claude-sonnet-4-5-20250929 | medium | **A** | 100% | 19.5s | 1126 |
| claude-sonnet-4-6 | medium | **A** | 100% | 16.7s | 1127 |
| gpt-5.1-codex | high | **A** | 100% | 26.1s | 1065 |
| gpt-5.2-codex | high | **A** | 100% | 18.4s | 1067 |
| claude-opus-4-5 | high | **A** | 100% | 14.0s | 1126 |
| claude-opus-4-6 | high | **A** | 100% | 18.5s | 1127 |
| gpt-4o-mini | minimal | **A** | 100% | 14.9s | 1028 |

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

Each model ran **5 role-appropriate prompts** designed for `architect` tasks.  
Evaluation uses keyword + structure heuristics tailored to `architect` output expectations.  
Grade: A (≥90% pass), B (≥75%), C (≥60%), D (≥40%), F (<40%).

---

*Last updated: 2026-05-28 | Data from `.claude/performance_grades/`*
