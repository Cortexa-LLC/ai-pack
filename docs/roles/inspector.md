# Inspector Role — Model Performance

**Role:** `inspector`  
**Description:** Bug root-cause analysis  
**Benchmark Task:** Models asked to rank likely root causes with supporting evidence  
**Last Updated:** 2026-05-28

---

## Model Performance Grades

*Benchmarked: 14 models × 5 prompts each = 70 benchmark runs for this role.*  
*Project: `/Users/bryanw/Projects/Vibe/ai-pack`*

| Model | Tier | Grade | Pass Rate | Avg Latency | Avg Tokens |
|-------|------|-------|-----------|-------------|------------|
| gpt-4.1-nano | minimal | **A** | 100% | 7.0s | 929 |
| gpt-4.1-mini | low | **A** | 100% | 13.9s | 967 |
| claude-haiku-4-5 | minimal | **A** | 100% | 10.8s | 1149 |
| o4-mini | low | **A** | 100% | 63.4s | 8406 |
| gpt-5.1-codex-mini | medium | **A** | 90% | 6.8s | 895 |
| gpt-4.1 | medium | **A** | 100% | 17.5s | 924 |
| claude-sonnet-4-5 | medium | **A** | 100% | 24.3s | 1141 |
| claude-sonnet-4-5-20250929 | medium | **A** | 100% | 24.4s | 1149 |
| claude-sonnet-4-6 | medium | **A** | 100% | 25.7s | 1150 |
| gpt-5.1-codex | high | D ⚠️ | 60% | 47.2s | 1011 |
| gpt-5.2-codex | high | B | 80% | 19.4s | 1061 |
| claude-opus-4-5 | high | **A** | 100% | 22.2s | 1149 |
| claude-opus-4-6 | high | **A** | 100% | 26.5s | 1150 |
| gpt-4o-mini | minimal | **A** | 100% | 11.2s | 658 |


### Notable Exceptions

- **gpt-5.1-codex** (D, 60%): Latency-sensitive model — some prompts exceed 60s timeout under load
- **gpt-5.2-codex** (B, 80%): Latency-sensitive model — some prompts exceed 60s timeout under load

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

Each model ran **5 role-appropriate prompts** designed for `inspector` tasks.  
Evaluation uses keyword + structure heuristics tailored to `inspector` output expectations.  
Grade: A (≥90% pass), B (≥75%), C (≥60%), D (≥40%), F (<40%).

---

*Last updated: 2026-05-28 | Data from `.claude/performance_grades/`*
