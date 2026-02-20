# Architect Role — Model Performance

**Role:** `architect`  
**Description:** Technical design and system architecture specialist  
**Configured Default:** (not specified — role spec lacks explicit model recommendation)  
**Escalation Path:** (not documented in role spec)  
**Last Updated:** 2026-02-19

---

## Model Performance Grades

| Model | Grade | Success Rate | Attempts | Confidence | Project | Last Tested | Notes |
|-------|-------|-------------|----------|------------|---------|-------------|-------|
| gpt-4o-mini | ? | — | 0 | — | — | — | Not in current escalation path |
| claude-sonnet-4-5 | **F** | 33% (4/12) | 12 | 0.60 (moderate) | xasm++ | 2026-02-14 | Most attempts, moderate confidence |
| claude-opus-4-6 | ? | — | 0 | — | — | — | Natural fit for role complexity |

> **Notable:** The architect role has the **most benchmark data** (12 attempts) and a **moderate confidence score** (0.60). This makes the F grade more meaningful than the F grades from 3-attempt samples.

---

## Analysis: claude-sonnet-4-5 as Architect

With 12 attempts and 4 successes (33% success rate), the architect role benchmarks are the most statistically significant data in the system:

- **12 attempts** vs. 3-6 for other roles
- **0.60 confidence** vs. 0.15 for most others
- Consistent with the hypothesis that claude-sonnet-4-5 struggles on the `xasm++` project specifically

However, a 33% success rate on architectural tasks is concerning even for a challenging project. Architecture requires:
- Deep understanding of existing system structure
- Long-horizon reasoning about trade-offs
- Clear documentation of decisions and rationale

These are areas where claude-sonnet-4-5 may genuinely underperform compared to larger models.

---

## Role-Specific Model Considerations

### Architect Task Complexity
Architectural tasks are inherently more complex than engineering tasks:
- Require understanding the *whole* system, not just one feature
- Must balance competing concerns (performance, maintainability, cost)
- Decisions have long-term consequences
- Output must be clear enough for other agents to implement

This suggests the **architect role benefits more from premium models** than most other roles.

### Recommended Model Hierarchy for Architect

```
claude-sonnet-4-5 → claude-sonnet-4-6 → claude-opus-4-6
```

Skip gpt-4o-mini and gpt-4o for architect tasks — the complexity justifies starting at a higher tier.

---

## Task Types and Model Fit

| Task Type | Recommended Model | Reasoning |
|-----------|------------------|-----------|
| API design | claude-sonnet-4-5 | Needs good reasoning, sonnet should suffice |
| System architecture | claude-sonnet-4-6 | Complex reasoning, newer model preferred |
| ADR (Architecture Decision Records) | claude-sonnet-4-5 | Documentation-heavy, sonnet capable |
| Security architecture | claude-opus-4-6 | High stakes, premium reasoning needed |
| Refactoring strategy | claude-sonnet-4-5 | Moderate complexity |
| Microservices design | claude-opus-4-6 | Complex trade-offs, worth the premium |

---

## Benchmark History

### 2026-02-03 to 2026-02-14: claude-sonnet-4-5 on xasm++ (Architect)
- **Result:** F (33% success rate)
- **Sample:** 12 attempts, 4 successes, 2 explicit failures (6 unclear)
- **Confidence:** 0.60 (moderate — most reliable data in system)
- **Context:** xasm++ is a 6502 assembler requiring niche domain knowledge
- **Escalation activity:** 0 escalations, 0 downgrades recorded
- **Avg execution time:** ~283 seconds/attempt

---

## Gaps in Data

- [ ] No benchmarks on gpt-4o-mini (likely too weak for this role)
- [ ] No benchmarks on claude-opus-4-6 (likely the best fit)
- [ ] No benchmarks on claude-sonnet-4-6 (successor, may perform better)
- [ ] No benchmarks on typical software projects (only xasm++)
- [ ] Need to capture specific failure modes (incomplete diagrams? wrong patterns? timeout?)

---

## Action Items

1. **Add explicit model recommendation** to `roles/architect.md`
2. **Benchmark claude-opus-4-6** for architect role — likely strongest fit
3. **Run benchmarks on a web app project** to complement xasm++ data
4. **Document failure modes** from the 8 failed architect attempts

---

*Last updated: 2026-02-19 | Grade data from `.claude/performance_grades_backup/`*
