# gpt-4o-mini Performance Profile

**Model ID:** `gpt-4o-mini`  
**Provider:** OpenAI  
**Tier:** 1 (Minimal Cost — Default)  
**Cost:** ~$0.15 per 1M input tokens, ~$0.60 per 1M output tokens  
**Context Window:** 128K tokens  
**Last Updated:** 2026-02-19

---

## Overview

GPT-4o-mini is the **default recommended model** for most roles in this system. It represents the best cost-efficiency trade-off for pattern-following, simple edits, test writing, and refactoring. The philosophy is "start cheap, escalate only when necessary" — gpt-4o-mini is the first model tried for the majority of tasks.

---

## Performance Grades by Role

| Role | Grade | Success Rate | Attempts | Project | Last Tested | Notes |
|------|-------|-------------|----------|---------|-------------|-------|
| engineer | ? | No data | — | — | — | Default model, benchmarks pending |
| tester | ? | No data | — | — | — | Default model, benchmarks pending |
| reviewer | ? | No data | — | — | — | Used in reviewer-codex variants |
| product-manager | ? | No data | — | — | — | Default model |
| designer | ? | No data | — | — | — | Default model |
| strategist | ? | No data | — | — | — | Default model |

> **Note:** Despite being the configured default, no benchmark grade files have been generated for gpt-4o-mini yet. This is a gap in our observability — the system may be defaulting to claude-sonnet-4-5 in practice.

---

## Recommended Use Cases

### ✅ Use gpt-4o-mini For:
- Pattern-following tasks with clear templates
- Simple code edits and refactoring
- Test writing from existing test patterns
- Documentation generation
- Straightforward feature implementations
- Tasks estimated to complete in < 30 minutes

### ❌ Avoid gpt-4o-mini For:
- Complex multi-step reasoning
- Large codebase analysis (>50K context)
- Architecture design decisions
- Security-critical code review
- Tasks requiring deep domain expertise

---

## Cost Advantage

At ~$0.15/1M input tokens vs. ~$3.00/1M for claude-sonnet-4-5, gpt-4o-mini is **20× cheaper**. For high-volume routine tasks, this is a significant budget consideration.

| Scenario | gpt-4o-mini | claude-sonnet-4-5 | Savings |
|----------|-------------|-------------------|---------|
| 1M tokens/day | $0.15 | $3.00 | $2.85/day |
| 10M tokens/day | $1.50 | $30.00 | $28.50/day |
| 100M tokens/day | $15.00 | $300.00 | $285/day |

---

## Escalation Path

```
gpt-4o-mini → gpt-4o → claude-sonnet-4-5 → claude-opus-4-6
```

**When to escalate FROM gpt-4o-mini:**
- Task fails after 1-2 attempts
- Output quality consistently poor
- Task requires reasoning beyond pattern matching
- Context window exceeded (>128K tokens)

---

## Notes for Future Benchmarks

- [ ] Run benchmark suite for all roles configured to default to gpt-4o-mini
- [ ] Measure success rate on simple vs. complex tasks
- [ ] Compare escalation frequency vs. expected
- [ ] Track cost savings vs. if claude-sonnet-4-5 used by default

---

*Last updated: 2026-02-19 | Data source: Role configuration files in `roles/` (no grade files yet)*
