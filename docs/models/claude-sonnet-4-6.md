# claude-sonnet-4-6 Performance Profile

**Model ID:** `claude-sonnet-4-6`
**Provider:** Anthropic
**Tier:** 3 (Medium Cost)
**Cost:** ~$3.00 per 1M input tokens
**Context Window:** 200K tokens
**Last Updated:** 2026-02-28

---

## Overview

Claude Sonnet 4.6 is the successor to claude-sonnet-4-5, featuring improvements in reasoning, instruction following, and task completion. It is the designated model for the **spelunker** role and is the default medium-tier choice in most escalation paths.

---

## Observed Strengths

- **Improved over 4-5:** Better reasoning and code understanding
- **Designated for spelunker role:** Explicitly selected in role configuration
- **Large context window:** 200K tokens for large codebase analysis
- **Recent model:** More up-to-date training data

## Observed Weaknesses

- **Same cost tier as 4-5:** No cost advantage over predecessor

---

## Recommended Use Cases

### ✅ Use claude-sonnet-4-6 For:
- **Spelunker role** (explicitly recommended in role spec)
- Runtime investigation and debugging
- Tasks where claude-sonnet-4-5 fails
- Large codebase analysis

### ❌ Avoid For:
- High-volume, simple tasks (use gpt-4o-mini)
- When claude-sonnet-4-5 works fine (no cost benefit)

---

## Escalation Path

```
claude-sonnet-4-5 → claude-sonnet-4-6 → claude-opus-4-6
```

---

*Last updated: 2026-02-28*
