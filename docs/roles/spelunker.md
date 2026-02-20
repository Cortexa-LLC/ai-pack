# Spelunker Role — Model Performance

**Role:** `spelunker`  
**Description:** Codebase navigation and exploration  
**Benchmark Task:** Models asked to reason through where to look in a codebase  
**Last Updated:** 2026-05-28

---

## Model Performance Grades

*Benchmarked: 14 models × 5 prompts each = 70 benchmark runs for this role.*  
*Project: `/Users/bryanw/Projects/Vibe/ai-pack`*

| Model | Tier | Grade | Pass Rate | Avg Latency | Avg Tokens |
|-------|------|-------|-----------|-------------|------------|
| gpt-4.1-nano | minimal | **A** | 100% | 8.6s | 1092 |
| gpt-4.1-mini | low | **A** | 100% | 16.3s | 1054 |
| claude-haiku-4-5 | minimal | **A** | 100% | 8.5s | 1147 |
| o4-mini | low | **A** | 100% | 59.8s | 7919 |
| gpt-5.1-codex-mini | medium | **A** | 100% | 6.6s | 1060 |
| gpt-4.1 | medium | **A** | 100% | 18.4s | 1118 |
| claude-sonnet-4-5 | medium | **A** | 100% | 19.0s | 1147 |
| claude-sonnet-4-5-20250929 | medium | **A** | 100% | 17.8s | 1147 |
| claude-sonnet-4-6 | medium | **A** | 100% | 19.5s | 1148 |
| gpt-5.1-codex | high | **A** | 100% | 32.3s | 749 |
| gpt-5.2-codex | high | B | 80% | 21.0s | 1076 |
| claude-opus-4-5 | high | **A** | 100% | 17.6s | 1147 |
| claude-opus-4-6 | high | **A** | 100% | 19.9s | 1148 |
| gpt-4o-mini | minimal | **A** | 100% | 13.2s | 862 |


### Notable Exceptions

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

Each model ran **5 role-appropriate prompts** designed for `spelunker` tasks.  
Evaluation uses keyword + structure heuristics tailored to `spelunker` output expectations.  
Grade: A (≥90% pass), B (≥75%), C (≥60%), D (≥40%), F (<40%).

---

*Last updated: 2026-05-28 | Data from `.claude/performance_grades/`*
