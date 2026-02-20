# Strategist Role — Model Performance

**Role:** `strategist`  
**Description:** Business strategy and planning  
**Benchmark Task:** Models asked to outline strategies with options and trade-offs  
**Last Updated:** 2026-05-28

---

## Model Performance Grades

*Benchmarked: 14 models × 5 prompts each = 70 benchmark runs for this role.*  
*Project: `/Users/bryanw/Projects/Vibe/ai-pack`*

| Model | Tier | Grade | Pass Rate | Avg Latency | Avg Tokens |
|-------|------|-------|-----------|-------------|------------|
| gpt-4.1-nano | minimal | **A** | 100% | 8.7s | 1135 |
| gpt-4.1-mini | low | **A** | 100% | 16.9s | 1110 |
| claude-haiku-4-5 | minimal | **A** | 100% | 12.6s | 1153 |
| o4-mini | low | **A** | 100% | 10.9s | 1531 |
| gpt-5.1-codex-mini | medium | **A** | 100% | 6.4s | 1142 |
| gpt-4.1 | medium | **A** | 100% | 18.3s | 1070 |
| claude-sonnet-4-5 | medium | **A** | 100% | 27.5s | 1153 |
| claude-sonnet-4-5-20250929 | medium | **A** | 100% | 26.8s | 1153 |
| claude-sonnet-4-6 | medium | **A** | 100% | 27.1s | 1154 |
| gpt-5.1-codex | high | **A** | 100% | 46.1s | 1082 |
| gpt-5.2-codex | high | **A** | 100% | 19.5s | 1142 |
| claude-opus-4-5 | high | **A** | 100% | 25.0s | 1153 |
| claude-opus-4-6 | high | **A** | 100% | 28.7s | 1154 |
| gpt-4o-mini | minimal | **A** | 100% | 15.3s | 945 |

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

Each model ran **5 role-appropriate prompts** designed for `strategist` tasks.  
Evaluation uses keyword + structure heuristics tailored to `strategist` output expectations.  
Grade: A (≥90% pass), B (≥75%), C (≥60%), D (≥40%), F (<40%).

---

*Last updated: 2026-05-28 | Data from `.claude/performance_grades/`*
