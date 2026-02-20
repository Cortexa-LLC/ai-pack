# claude-opus-4-6 Performance Profile

**Model ID:** `claude-opus-4-6`  
**Provider:** Anthropic  
**Tier:** 4 (High Cost — Premium)  
**Cost:** ~$15.00 per 1M input tokens, ~$75.00 per 1M output tokens  
**Context Window:** 200K tokens  
**Last Updated:** 2026-02-19

---

## Overview

Claude Opus 4.6 is the highest-capability model in the escalation ladder. It is the **last resort** for tasks where all other models have failed. The premium cost (~5× claude-sonnet-4-5) means it should be used sparingly — only for tasks requiring deep reasoning, complex architectural decisions, or specialized domain expertise.

---

## Performance Grades by Role

| Role | Grade | Success Rate | Attempts | Project | Last Tested | Notes |
|------|-------|-------------|----------|---------|-------------|-------|
| reviewer | ? | No data | — | — | — | Configured in reviewer-opus.md |
| orchestrator | ? | No data | — | — | — | Top escalation target |
| architect | ? | No data | — | — | — | Premium reasoning for complex design |

> **Note:** No benchmark grade files exist for claude-opus-4-6. Given the cost, benchmarks should be run on targeted scenarios only.

---

## Recommended Use Cases

### ✅ Use claude-opus-4-6 For:
- Architectural decisions with major long-term impact
- Security-critical code review
- Complex debugging that requires deep reasoning
- Tasks where cheaper models have all failed
- Premium code review (see `reviewer-opus.md`)

### ❌ Avoid claude-opus-4-6 For:
- Routine engineering tasks (20× more expensive than gpt-4o-mini)
- Simple edits and pattern following
- High-volume tasks (budget will spike rapidly)
- When claude-sonnet-4-5 or claude-sonnet-4-6 suffices

---

## Cost Warning

| Scenario | gpt-4o-mini | claude-opus-4-6 | Difference |
|----------|-------------|-----------------|------------|
| 1M tokens | $0.15 | $15.00 | 100× more expensive |
| 10M tokens | $1.50 | $150.00 | $148.50 extra |
| Daily at 20M tokens | $3.00 | $300.00 | $297/day extra |

---

## Escalation Path

```
claude-sonnet-4-5 → claude-sonnet-4-6 → claude-opus-4-6
```

claude-opus-4-6 is the **terminus** — there is no further escalation. If opus fails, the task needs human intervention or task decomposition.

---

## Notes for Future Benchmarks

- [ ] Run targeted benchmarks on reviewer-opus role
- [ ] Compare reviewer output quality: opus vs. sonnet
- [ ] Measure cost-per-successful-review to quantify value
- [ ] Test on architectural design tasks to validate premium positioning

---

*Last updated: 2026-02-19 | Data source: Role configuration files (no grade files yet)*
