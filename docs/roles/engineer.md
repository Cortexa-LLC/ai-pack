# Engineer Role — Model Performance

**Role:** `engineer`  
**Description:** Implementation specialist for writing code, tests, bug fixes  
**Configured Default:** gpt-4o-mini  
**Last Updated:** 2026-02-19

---

## Model Performance Grades

| Model | Grade | Success Rate | Attempts | Confidence | Project | Last Tested | Notes |
|-------|-------|-------------|----------|------------|---------|-------------|-------|
| gpt-4o-mini | ? | — | 0 | — | — | — | Default model, no benchmarks yet |
| gpt-4o | ? | — | 0 | — | — | — | First escalation target |
| claude-haiku-4-5 | ? | — | 0 | — | — | — | Alternative cheap model |
| claude-sonnet-4-5 | **F** | 33% (1/3) | 3 | 0.15 (low) | ai-pack | 2026-02-19 | 2 failures, 67% error rate |
| claude-sonnet-4-5 | **F** | 33% (1/3) | 3 | 0.15 (low) | xasm++ | 2026-02-18 | Same grade, different project |
| claude-sonnet-4-6 | ? | — | 0 | — | — | — | No benchmarks yet |
| claude-opus-4-6 | ? | — | 0 | — | — | — | Top escalation target |

---

## Escalation Path

```
gpt-4o-mini → gpt-4o → claude-sonnet-4-5 → claude-opus-4-6
```

The engineer role is configured to start with the cheapest model and escalate based on failure. Current benchmark data only covers claude-sonnet-4-5, which shows poor performance — but this may reflect the complexity of benchmark projects rather than the model's general capability.

---

## Role-Specific Model Considerations

### Why gpt-4o-mini is the Default
- Engineer tasks are often well-defined, pattern-following work
- Cheap model handles 80%+ of simple implementation tasks
- Cost savings are significant at scale (20× cheaper than claude-sonnet-4-5)

### When claude-sonnet-4-5 Fails as Engineer
Observed failure patterns from benchmark data:
- **Error rate 67% on ai-pack** — complex project with tight integration requirements
- **Low confidence (0.15)** — sample size of only 3 attempts is insufficient
- Specific failure modes not documented yet (tool call errors? incomplete output? wrong logic?)

### Recommendations
1. **Start with gpt-4o-mini** — it's the configured default for good reason
2. **Escalate to gpt-4o** if gpt-4o-mini fails (before trying claude-sonnet-4-5)
3. **Document failure modes** when escalating — helps tune the escalation policy
4. **Do not default to claude-sonnet-4-5** based on current benchmark data

---

## Task Types and Model Fit

| Task Type | Recommended Model | Reasoning |
|-----------|------------------|-----------|
| Simple edits, refactoring | gpt-4o-mini | Pattern following, no deep reasoning needed |
| Test writing | gpt-4o-mini | Template-driven, coverage-focused |
| Bug fixes (clear root cause) | gpt-4o-mini | Well-defined problem, bounded solution |
| Feature implementation | gpt-4o-mini → gpt-4o | Escalate if initial attempt fails |
| Complex debugging | gpt-4o → claude-sonnet-4-5 | Multi-file analysis, reasoning needed |
| Security-critical code | claude-sonnet-4-5 → claude-opus-4-6 | High stakes, needs careful review |
| Architecture-heavy features | Delegate to architect role | Not engineer's domain |

---

## Benchmark History

### 2026-02-19: claude-sonnet-4-5 on ai-pack (Engineer)
- **Result:** F (33% success rate)
- **Sample:** 3 attempts, 1 success, 2 failures
- **Error rate:** 67%
- **Confidence:** 0.15 (very low — 3 attempts insufficient)
- **Context:** ai-pack is the agent framework project itself — high complexity
- **Implication:** Do not use claude-sonnet-4-5 as default engineer on complex self-referential projects

### 2026-02-18: claude-sonnet-4-5 on xasm++ (Engineer)
- **Result:** F (33% success rate)
- **Sample:** 3 attempts, 1 success, 2 failures
- **Context:** xasm++ is a 6502 assembler — niche domain requiring specialized knowledge
- **Implication:** Model struggled with domain-specific constraints

---

## Gaps in Data

- [ ] No gpt-4o-mini benchmarks (despite being the default model)
- [ ] No gpt-4o benchmarks
- [ ] No claude-sonnet-4-6 benchmarks
- [ ] No benchmarks on typical/simple engineering tasks
- [ ] No benchmarks across diverse project types

---

*Last updated: 2026-02-19 | Grade data from `.claude/performance_grades/`*
