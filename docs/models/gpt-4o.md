# gpt-4o Performance Profile

**Model ID:** `gpt-4o`  
**Provider:** OpenAI  
**Tier:** 2 (Low Cost)  
**Cost:** ~$2.50 per 1M input tokens, ~$10.00 per 1M output tokens  
**Context Window:** 128K tokens  
**Last Updated:** 2026-02-19

---

## Overview

GPT-4o is the mid-tier OpenAI model used as the first escalation target from gpt-4o-mini. It offers significantly better reasoning and code understanding than gpt-4o-mini at a moderate cost increase. It sits between the cheap defaults and the premium Anthropic models in the escalation ladder.

---

## Performance Grades by Role

| Role | Grade | Success Rate | Attempts | Project | Last Tested | Notes |
|------|-------|-------------|----------|---------|-------------|-------|
| engineer | ? | No data | — | — | — | Second escalation target |
| tester | ? | No data | — | — | — | Escalation target |
| reviewer | ? | No data | — | — | — | Used in reviewer-codex variants |

> **Note:** No benchmark grade files exist for gpt-4o yet. Benchmarks needed.

---

## Recommended Use Cases

### ✅ Use gpt-4o For:
- Tasks that gpt-4o-mini fails (first escalation)
- Moderate complexity code review
- Feature implementation with non-trivial logic
- Debugging with clear error messages and stack traces
- Integration testing design

### ❌ Avoid gpt-4o For:
- Simple pattern-following (use gpt-4o-mini)
- Very large context requirements (>128K, use Claude)
- Tasks that need premium reasoning (use claude-sonnet-4-5+)

---

## Escalation Path

```
gpt-4o-mini → gpt-4o → claude-sonnet-4-5 → claude-opus-4-6
```

---

## Notes for Future Benchmarks

- [ ] Run benchmarks for engineer and tester roles
- [ ] Measure improvement vs. gpt-4o-mini success rate
- [ ] Compare cost-per-successful-task vs. claude-sonnet-4-5

---

*Last updated: 2026-02-19 | Data source: Role configuration files (no grade files yet)*
