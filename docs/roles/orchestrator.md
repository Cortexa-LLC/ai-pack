# Orchestrator Role — Model Performance

**Role:** `orchestrator`  
**Description:** Multi-agent task coordination and delegation  
**Configured Default:** (not specified — uses system default)  
**Last Updated:** 2026-02-19

---

## Model Performance Grades

| Model | Grade | Success Rate | Attempts | Confidence | Project | Last Tested | Notes |
|-------|-------|-------------|----------|------------|---------|-------------|-------|
| claude-sonnet-4-5 | **F** | 33% (2/6) | 6 | 0.30 (low) | xasm++ | 2026-02-03 | 4 failures, 0 retries |
| claude-opus-4-6 | ? | — | 0 | — | — | — | Premium fit for complex orchestration |
| gpt-4o | ? | — | 0 | — | — | — | Potentially viable for simpler orchestration |

---

## Role-Specific Model Considerations

### Orchestrator Task Complexity
The orchestrator role is one of the **most complex roles** in the system:
- Must understand the full project context
- Decomposes tasks intelligently across multiple agents
- Manages dependencies and sequencing
- Handles failures and re-routing
- Maintains coherence across long multi-turn sessions

This argues for **higher-tier models** as the default. The F grade for claude-sonnet-4-5 with 6 attempts (moderate sample) is more concerning than the 3-attempt F grades.

### Confidence Analysis
- **0.30 confidence** with 6 attempts is the second-highest in the dataset
- 2 successes out of 6 attempts represents a real pattern, not statistical noise
- The 4 failures on xasm++ orchestration likely reflect domain complexity

### Recommended Model for Orchestrator
Based on role complexity and available data:
1. **Primary:** claude-sonnet-4-6 or claude-opus-4-6
2. **Budget alternative:** claude-sonnet-4-5 (monitor closely)
3. **Avoid:** gpt-4o-mini (orchestration requires deep reasoning)

---

## Escalation Path for Orchestrator

```
claude-sonnet-4-5 → claude-sonnet-4-6 → claude-opus-4-6
```

Unlike engineering roles, **skip the cheap models** for orchestration. The cost of a failed orchestration run (wasted sub-agent work) often exceeds the savings from using a cheaper orchestrator.

---

## Benchmark History

### 2026-02-03: claude-sonnet-4-5 on xasm++ (Orchestrator)
- **Result:** F (33% success rate)
- **Sample:** 6 attempts, 2 successes, 4 failures
- **Confidence:** 0.30 (low-moderate)
- **Context:** xasm++ project — niche 6502 assembler domain
- **Execution time:** ~283 seconds average
- **Escalations/Downgrades:** None recorded

---

## Gaps in Data

- [ ] No benchmarks for claude-opus-4-6 (most appropriate model)
- [ ] No benchmarks on typical software orchestration scenarios
- [ ] Need to understand failure modes: did orchestration produce wrong sub-tasks? Failed to coordinate? Timed out?
- [ ] No benchmarks across different orchestration complexity levels (simple 2-agent vs. complex 5-agent pipelines)

---

*Last updated: 2026-02-19 | Grade data from `.claude/performance_grades_backup/`*
