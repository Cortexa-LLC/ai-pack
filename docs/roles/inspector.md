# Inspector Role

**Role:** `inspector`
**Description:** Bug root-cause analysis
**Last Updated:** 2026-02-28

> **Note:** `gpt-5.1-codex` and `gpt-5.2-codex` use the OpenAI Responses API with a 60s timeout. Long reasoning responses on complex root-cause tasks may exceed this timeout — prefer Claude models for latency-sensitive inspector workloads.

---

## Escalation Path

```
gpt-4.1-nano → gpt-4.1-mini → gpt-4.1 → claude-sonnet-4-6 → claude-opus-4-6
```

---

*Last updated: 2026-02-28*
