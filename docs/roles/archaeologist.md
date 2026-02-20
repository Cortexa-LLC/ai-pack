# Archaeologist Role — Model Performance

**Role:** `archaeologist`  
**Description:** Legacy code understanding and rationale recovery  
**Benchmark Task:** Models asked to explain rationale behind legacy code decisions  
**Last Updated:** 2026-05-28

---

## Model Performance Grades

*Benchmarked: 14 models × 5 prompts each = 70 benchmark runs for this role.*  
*Project: `/Users/bryanw/Projects/Vibe/ai-pack`*

| Model | Tier | Grade | Pass Rate | Avg Latency | Avg Tokens |
|-------|------|-------|-----------|-------------|------------|
| gpt-4.1-nano | minimal | **A** | 100% | 9.5s | 1185 |
| gpt-4.1-mini | low | **A** | 100% | 13.4s | 1131 |
| claude-haiku-4-5 | minimal | **A** | 100% | 11.3s | 1215 |
| o4-mini | low | **A** | 100% | 39.2s | 4606 |
| gpt-5.1-codex-mini | medium | **A** | 100% | 6.9s | 1192 |
| gpt-4.1 | medium | **A** | 100% | 21.6s | 1201 |
| claude-sonnet-4-5 | medium | **A** | 100% | 27.0s | 1215 |
| claude-sonnet-4-5-20250929 | medium | **A** | 100% | 27.7s | 1217 |
| claude-sonnet-4-6 | medium | **A** | 100% | 25.9s | 1216 |
| gpt-5.1-codex | high | B | 80% | 33.4s | 1050 |
| gpt-5.2-codex | high | **A** | 100% | 19.4s | 1089 |
| claude-opus-4-5 | high | **A** | 100% | 23.6s | 1215 |
| claude-opus-4-6 | high | **A** | 100% | 24.6s | 1216 |
| gpt-4o-mini | minimal | **A** | 100% | 16.7s | 1034 |


### Notable Exceptions

- **gpt-5.1-codex** (B, 80%): Latency-sensitive model — some prompts exceed 60s timeout under load

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

Each model ran **5 role-appropriate prompts** designed for `archaeologist` tasks.  
Evaluation uses keyword + structure heuristics tailored to `archaeologist` output expectations.  
Grade: A (≥90% pass), B (≥75%), C (≥60%), D (≥40%), F (<40%).

---

*Last updated: 2026-05-28 | Data from `.claude/performance_grades/`*
