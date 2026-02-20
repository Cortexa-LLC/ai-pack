# Cartographer Role — Model Performance

**Role:** `cartographer`  
**Description:** Architecture mapping and documentation  
**Benchmark Task:** Models asked to describe architecture from file/component structure  
**Last Updated:** 2026-05-28

---

## Model Performance Grades

*Benchmarked: 14 models × 5 prompts each = 70 benchmark runs for this role.*  
*Project: `/Users/bryanw/Projects/Vibe/ai-pack`*

| Model | Tier | Grade | Pass Rate | Avg Latency | Avg Tokens |
|-------|------|-------|-----------|-------------|------------|
| gpt-4.1-nano | minimal | **A** | 100% | 7.4s | 1158 |
| gpt-4.1-mini | low | **A** | 100% | 12.0s | 1110 |
| claude-haiku-4-5 | minimal | **A** | 100% | 8.6s | 1255 |
| o4-mini | low | **A** | 100% | 40.2s | 5437 |
| gpt-5.1-codex-mini | medium | **A** | 100% | 4.2s | 858 |
| gpt-4.1 | medium | **A** | 100% | 20.0s | 1232 |
| claude-sonnet-4-5 | medium | **A** | 100% | 20.6s | 1258 |
| claude-sonnet-4-5-20250929 | medium | **A** | 100% | 21.5s | 1258 |
| claude-sonnet-4-6 | medium | **A** | 100% | 20.9s | 1216 |
| gpt-5.1-codex | high | B | 80% | 21.9s | 658 |
| gpt-5.2-codex | high | **A** | 100% | 15.8s | 1012 |
| claude-opus-4-5 | high | **A** | 100% | 15.9s | 1258 |
| claude-opus-4-6 | high | **A** | 100% | 21.8s | 1259 |
| gpt-4o-mini | minimal | **A** | 100% | 17.5s | 1080 |


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

Each model ran **5 role-appropriate prompts** designed for `cartographer` tasks.  
Evaluation uses keyword + structure heuristics tailored to `cartographer` output expectations.  
Grade: A (≥90% pass), B (≥75%), C (≥60%), D (≥40%), F (<40%).

---

*Last updated: 2026-05-28 | Data from `.claude/performance_grades/`*
