# claude-sonnet-4-5 Performance Profile

**Model ID:** `claude-sonnet-4-5-20250929`
**Provider:** Anthropic
**Tier:** 3 (Medium Cost)
**Cost:** ~$3.00 per 1M input tokens, ~$15.00 per 1M output tokens
**Context Window:** 200K tokens
**Last Updated:** 2026-02-28

---

## Overview

Claude Sonnet 4.5 is Anthropic's mid-tier model, positioned as an escalation target for roles needing moderate reasoning above cheaper models (gpt-4o-mini, claude-haiku). It balances cost and capability.

---

## Observed Strengths

- **Large context window (200K):** Handles large codebases without truncation
- **Code generation:** Produces syntactically correct code in most cases
- **Instruction following:** Generally adheres to role-specific instructions
- **Multi-step reasoning:** Can break down complex tasks into steps

## Observed Weaknesses

- **Domain-specific knowledge:** May struggle with niche or legacy tech
- **High variance:** Success rates vary across task types and project complexity

---

## Recommended Use Cases

### ✅ Use claude-sonnet-4-5 For:
- Complex reasoning tasks that cheaper models fail
- Large-context analysis (>50K token inputs)
- Tasks escalated from gpt-4o-mini or claude-haiku
- Architecture review, security analysis, complex debugging
- Orchestration of multi-agent workflows

### ❌ Avoid claude-sonnet-4-5 For:
- Simple, pattern-following tasks (use gpt-4o-mini instead)
- High-volume repetitive work (cost inefficient)

---

## Escalation Path

```
gpt-4o-mini → gpt-4o → claude-sonnet-4-5 → claude-opus-4-6
```

**When to escalate TO claude-sonnet-4-5:**
- gpt-4o fails or produces poor quality output
- Task requires >32K context
- Complex multi-step reasoning needed

**When to escalate FROM claude-sonnet-4-5:**
- Consecutive failures
- Task requires deeper specialized reasoning
- Budget allows for opus-level quality

---

*Last updated: 2026-02-28*
