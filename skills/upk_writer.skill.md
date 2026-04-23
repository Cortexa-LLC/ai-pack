# Personal Knowledge Writer (UPK)
<!-- skills/upk_writer.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 16
**Tools:** mcp__upk__add_learning, mcp__upk__add_conversation
**Gates:** knowledge-first
**MaxExtraTokens:** 5000
**Optional:** true

---

## Personal Knowledge — Record What You Learn

When you discover something meaningful during a task, **record it to personal knowledge immediately**. Future agents (and future you) across ALL projects will benefit from these learnings.

### ⚠️ MANDATORY: Write Back Findings

**REQUIRED operations after learning:**
- ✅ MUST record cross-project insights with `upk__add_learning`
- ✅ MUST record important conversations with `upk__add_conversation`
- ✅ MUST write incrementally (don't wait until task completes)

This is enforced by the **[Knowledge-First Gate](../gates/15-knowledge-first.md)**.

---

## What to Record

### Cross-Project Learnings

| Discovery | How to Record | Example |
|-----------|---------------|---------|
| Design pattern that worked well | `upk__add_learning` | "Circuit breaker pattern in Go: use hystrix-go, set timeout to 5s" |
| Performance optimization discovered | `upk__add_learning` | "Redis connection pooling: 20 connections optimal for 1000 req/s load" |
| Common pitfall avoided | `upk__add_learning` | "GraphQL N+1 queries: use DataLoader, batch within 10ms window" |
| Tool/library evaluation | `upk__add_learning` | "Tried zap vs logrus: zap 3x faster, but logrus better structured output" |
| Debugging technique that worked | `upk__add_learning` | "Memory leaks in goroutines: use pprof with GODEBUG=gctrace=1" |

### Important Conversations

| Conversation Type | How to Record | Example |
|-------------------|---------------|---------|
| Design discussion | `upk__add_conversation` | User explained why sync vs async matters for auth flow |
| Requirements clarification | `upk__add_conversation` | Product manager defined SLA requirements: p95 < 200ms |
| Architecture decision | `upk__add_conversation` | Team decided on event sourcing for audit trail |
| Feedback on approach | `upk__add_conversation` | User preferred monolithic deployment over microservices |

---

## Tools

- **`upk__add_learning`** `{content: string, source?: string}` — Record a learning or insight. Use for cross-project patterns, techniques, tool comparisons, and solutions to recurring problems.

- **`upk__add_conversation`** `{title: string, summary: string}` — Record important conversations. Use for design discussions, requirements clarifications, and decision rationale.

---

## When to Write

### Immediate Write-Back (MANDATORY)

```
DURING task execution:
  
  IF discovered new pattern OR technique THEN
    upk__add_learning({
      content: "<what you learned, how it worked, context>",
      source: "<project or conversation reference>"
    })
  END IF
  
  IF important conversation occurred THEN
    upk__add_conversation({
      title: "<brief topic>",
      summary: "<key points, decisions made, rationale>"
    })
  END IF

END DURING
```

**Write incrementally — do NOT wait until task completes.**

The task may time out or be interrupted. Every learning written to UPK is preserved across sessions and available to future agents. A timed-out task with UPK notes is recoverable; one without is wasted learning.

---

## Learning Quality Guidelines

### ✅ Good Learnings

```
upk__add_learning({
  content: "Rate limiting in Express.js: used express-rate-limit middleware. \
  Set windowMs to 15min, max 100 requests. Key by IP address. \
  For authenticated endpoints, key by user ID to prevent multi-IP bypass. \
  Store in Redis for distributed rate limiting across servers.",
  source: "ai-pack API server implementation"
})

upk__add_learning({
  content: "GraphQL Federation: composition requires @key directive on each \
  type to define entity resolution. Use @external + @requires for field \
  dependencies across services. Learned: keep entities small, move complex \
  logic to separate types to avoid circular dependencies.",
  source: "federation migration project"
})
```

**Why good:**
- Specific implementation details
- Context about why/when to use
- Pitfalls avoided
- Enough detail to reproduce

### ❌ Poor Learnings

```
upk__add_learning({
  content: "Rate limiting is good",
  source: "project"
})

upk__add_learning({
  content: "Fixed bug in auth.go line 42",
  source: "ai-pack"
})
```

**Why poor:**
- Too vague ("is good" - what specifically?)
- Not cross-project applicable (line-specific fix)
- Missing implementation details
- Can't reproduce from this info

---

## Conversation Quality Guidelines

### ✅ Good Conversations

```
upk__add_conversation({
  title: "Auth flow requirements - sync vs async token validation",
  summary: "User explained auth flow must be synchronous because client needs \
  immediate access token to proceed. Async validation would require polling or \
  webhooks, adding complexity. Decision: use synchronous JWT validation with \
  short-lived tokens (15min) and refresh token pattern. Rationale: simplicity \
  for client integration, security via short expiry."
})

upk__add_conversation({
  title: "Database choice - Postgres vs DynamoDB for metrics",
  summary: "Discussed tradeoffs for time-series metrics storage. User prefers \
  Postgres with TimescaleDB extension over DynamoDB. Reasons: (1) team has \
  Postgres expertise, (2) easier to query for dashboards, (3) lower cost at \
  current scale (<1M writes/day). DynamoDB considered for future if scale \
  increases 10x. Key insight: optimize for team skills + current scale, not \
  hypothetical future."
})
```

**Why good:**
- Captures decision + rationale
- Includes tradeoffs considered
- Specific enough to inform future similar decisions
- Records "why" not just "what"

### ❌ Poor Conversations

```
upk__add_conversation({
  title: "Talked about auth",
  summary: "User wants auth"
})
```

**Why poor:**
- No decision captured
- No rationale
- Too vague to be useful later

---

## UPK vs Project KG

**When to use UPK Writer:**
- ✅ Cross-project applicable learning
- ✅ General design patterns
- ✅ Tool/library comparisons
- ✅ User conversation about preferences
- ✅ Techniques that transfer between projects

**When to use KG Writer (see kg_writer skill):**
- ✅ Project-specific code entities
- ✅ This codebase's architecture
- ✅ Bug investigations in current project
- ✅ Component relationships here
- ✅ Project-specific design decisions

**Both when:**
- Learning has both cross-project value AND project-specific application
- Record the general pattern to UPK
- Record the specific implementation to project KG

---

## Required Workflow

### After External Search

```
AFTER WebSearch OR research THEN
  IF found valuable information THEN
    upk__add_learning({
      content: "<what you learned, how to apply, gotchas>",
      source: "<where you learned it>"
    })
  END IF
END AFTER
```

### After Important Conversation

```
AFTER user explains requirements OR preferences THEN
  upk__add_conversation({
    title: "<topic discussed>",
    summary: "<decisions made, rationale, tradeoffs>"
  })
END AFTER
```

### After Solving Novel Problem

```
AFTER solving problem not in knowledge base THEN
  upk__add_learning({
    content: "<problem description, solution approach, why it worked>",
    source: "<project context>"
  })
END AFTER
```

---

## Incremental Write Pattern

**CRITICAL:** Write as you learn, not at the end.

```
INVESTIGATION TASK: "Why is API slow?"

[Hypothesis 1: Database queries]
  test: run query analyzer
  result: queries under 10ms ✓
  → WRITE: upk__add_learning({
      content: "[INVESTIGATION] API slowness: database queries not the issue. \
      Verified with pg_stat_statements, all queries <10ms.",
      source: "ai-pack performance investigation"
    })

[Hypothesis 2: Network latency]
  test: check request/response times
  result: 2s latency in external API call ✓ ROOT CAUSE
  → WRITE: upk__add_learning({
      content: "[SOLUTION] API slowness caused by synchronous external API \
      call with 2s latency. Fixed by: (1) added timeout to 500ms, (2) added \
      circuit breaker to fail fast, (3) cached responses for 5min. Result: \
      p95 latency dropped from 2.1s to 150ms.",
      source: "ai-pack performance fix"
    })

[Implementation]
  code: add circuit breaker, caching
  → WRITE: upk__add_learning({
      content: "Circuit breaker in Go: used sony/gobreaker. Config: \
      maxRequests=100, timeout=500ms, interval=30s. Pairs well with Redis \
      cache for degraded mode operation.",
      source: "ai-pack external API integration"
    })
```

**Result:** Even if task times out after hypothesis 2, both learnings are preserved and available to the agent that continues the work.

---

## Compliance

**Write-back requirement:**
- Every external search that yields valuable info MUST be recorded
- Every important user conversation MUST be recorded
- Every novel solution MUST be recorded
- Write incrementally during task, not at completion

**Quality standard:**
- Learnings must be specific and reproducible
- Conversations must capture decision rationale
- Include enough context for future applicability

---

## See Also

- **[UPK Reader](upk_reader.skill.md)** - Searching personal knowledge
- **[KG Writer](kg_writer.skill.md)** - Recording project-specific knowledge
- **[Knowledge-First Gate](../gates/15-knowledge-first.md)** - Enforcement policy
