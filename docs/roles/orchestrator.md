# Orchestrator Role — Model Performance

**Role:** `orchestrator`  
**Description:** Multi-agent task coordination and delegation  
**Benchmark Task:** Models asked to produce delegation plans with dependency ordering  
**Last Updated:** 2026-05-28

---

## Model Performance Grades

*Benchmarked: 14 models × 5 prompts each = 70 benchmark runs for this role.*  
*Project: `/Users/bryanw/Projects/Vibe/ai-pack`*

| Model | Tier | Grade | Pass Rate | Avg Latency | Avg Tokens |
|-------|------|-------|-----------|-------------|------------|
| gpt-4.1-nano | minimal | **A** | 100% | 5.2s | 729 |
| gpt-4.1-mini | low | **A** | 100% | 9.3s | 632 |
| claude-haiku-4-5 | minimal | **A** | 100% | 10.5s | 1140 |
| o4-mini | low | **A** | 100% | 74.1s | 9916 |
| gpt-5.1-codex-mini | medium | **A** | 100% | 5.5s | 822 |
| gpt-4.1 | medium | **A** | 100% | 13.1s | 761 |
| claude-sonnet-4-5 | medium | **A** | 100% | 22.6s | 1144 |
| claude-sonnet-4-5-20250929 | medium | **A** | 100% | 23.3s | 1139 |
| claude-sonnet-4-6 | medium | **A** | 100% | 22.0s | 1146 |
| gpt-5.1-codex | high | **A** | 100% | 14.8s | 662 |
| gpt-5.2-codex | high | **A** | 100% | 15.4s | 885 |
| claude-opus-4-5 | high | **A** | 100% | 20.1s | 1145 |
| claude-opus-4-6 | high | **A** | 100% | 25.6s | 1145 |
| gpt-4o-mini | minimal | **A** | 100% | 13.9s | 811 |

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

Each model ran **5 role-appropriate prompts** designed for `orchestrator` tasks.  
Evaluation uses keyword + structure heuristics tailored to `orchestrator` output expectations.  
Grade: A (≥90% pass), B (≥75%), C (≥60%), D (≥40%), F (<40%).

---

*Last updated: 2026-05-28 | Data from `.claude/performance_grades/`*
