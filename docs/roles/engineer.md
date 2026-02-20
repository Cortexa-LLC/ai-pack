# Engineer Role — Model Performance

**Role:** `engineer`  
**Description:** Code implementation, TDD, bug fixing  
**Benchmark Task:** Models asked to write working code with tests  
**Last Updated:** 2026-05-28

---

## Model Performance Grades

*Benchmarked: 14 models × 5 prompts each = 70 benchmark runs for this role.*  
*Project: `/Users/bryanw/Projects/Vibe/ai-pack`*

| Model | Tier | Grade | Pass Rate | Avg Latency | Avg Tokens |
|-------|------|-------|-----------|-------------|------------|
| gpt-4.1-nano | minimal | **A** | 100% | 1.8s | 297 |
| gpt-4.1-mini | low | **A** | 100% | 3.3s | 286 |
| claude-haiku-4-5 | minimal | **A** | 100% | 2.3s | 419 |
| o4-mini | low | **A** | 100% | 22.6s | 2744 |
| gpt-5.1-codex-mini | medium | **A** | 100% | 2.9s | 371 |
| gpt-4.1 | medium | **A** | 100% | 2.8s | 234 |
| claude-sonnet-4-5 | medium | **A** | 100% | 4.2s | 355 |
| claude-sonnet-4-5-20250929 | medium | B | 83% | 0.4s | 0 |
| claude-sonnet-4-6 | medium | **A** | 100% | 6.1s | 554 |
| gpt-5.1-codex | high | **A** | 100% | 8.6s | 295 |
| gpt-5.2-codex | high | **A** | 100% | 3.7s | 325 |
| claude-opus-4-5 | high | **A** | 100% | 3.3s | 307 |
| claude-opus-4-6 | high | **A** | 100% | 5.4s | 394 |
| gpt-4o-mini | minimal | **A** | 100% | 4.1s | 238 |


### Notable Exceptions

- **claude-sonnet-4-5-20250929** (B, 83%): Legacy snapshot (Sep 2025) — prefer `claude-sonnet-4-5` or newer

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

Each model ran **5 role-appropriate prompts** designed for `engineer` tasks.  
Evaluation uses keyword + structure heuristics tailored to `engineer` output expectations.  
Grade: A (≥90% pass), B (≥75%), C (≥60%), D (≥40%), F (<40%).

---

*Last updated: 2026-05-28 | Data from `.claude/performance_grades/`*
