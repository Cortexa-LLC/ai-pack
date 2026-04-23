# Knowledge Systems — Quick Reference

**🚨 MANDATORY: Search knowledge BEFORE file operations**

---

## Which Knowledge System?

```
┌─────────────────────────────────────────────────────────┐
│ QUESTION TYPE          │ KNOWLEDGE TIER │ TOOLS        │
├─────────────────────────────────────────────────────────┤
│ THIS project's code    │ Project KG     │ kg__*        │
│ Cross-project learning │ Personal       │ upk__*       │
│ Org/team/planning      │ Organizational │ MCP servers  │
└─────────────────────────────────────────────────────────┘
```

---

## Project Knowledge (kg)

**Scope:** THIS project only

### Read
```javascript
// Find code in THIS project
kg__search_knowledge({query: "authentication"})

// Get file contents summary
kg__get_file_context({file: "pkg/auth/handler.go"})

// Get context before starting task
kg__get_preflight_context({task: "implement rate limiting"})

// Query relationships
kg__query_graph({cypher: "MATCH (f:function)-[:CALLS]->(g) RETURN g"})
```

### Write
```javascript
// Add new entity
entity_id = kg__add_entity({name: "CircuitBreaker", type: "type"})

// Record observation/finding
kg__add_observation({
  entity_id: entity_id,
  content: "[BUG FIX] Fixed race condition in connection pool..."
})

// Link entities
kg__link_entities({
  from_id: entity1_id,
  relation: "CALLS",
  to_id: entity2_id
})

// Re-index after major changes
kg__index_project()
```

---

## Personal Knowledge (upk)

**Scope:** ALL projects, ALL conversations

### Read
```javascript
// Find cross-project learnings
upk__search_knowledge({query: "rate limiting patterns"})
```

### Write
```javascript
// Record learning
upk__add_learning({
  content: "Circuit breaker in Go: use sony/gobreaker, config...",
  source: "ai-pack implementation"
})

// Record conversation
upk__add_conversation({
  title: "Database choice - Postgres vs DynamoDB",
  summary: "User prefers Postgres because team has expertise..."
})
```

---

## Mandatory Workflow

### Before ANY file operation:

```
BEFORE Grep OR Glob OR Read:
  ✅ kg__search_knowledge({query: "..."})
  ✅ IF not found THEN grep/glob/read
  ✅ AFTER finding → kg__add_observation

BEFORE Read(file.go):
  ✅ kg__get_file_context({file: "file.go"})
  ✅ Read only relevant sections

BEFORE starting task:
  ✅ kg__get_preflight_context({task: "..."})
  ✅ upk__search_knowledge({query: "..."}) if cross-project
```

### After learning something:

```
AFTER grep finds answer:
  ✅ Record to kg OR upk (depending on scope)

AFTER user conversation:
  ✅ upk__add_conversation

AFTER solving novel problem:
  ✅ Record to both kg (project location) AND upk (reusable pattern)
```

---

## Token Savings

```
WITHOUT knowledge-first:
  grep -r "pattern" → 50,000 tokens
  Read 10 files → 80,000 tokens
  Total: 130,000 tokens

WITH knowledge-first:
  kg__search_knowledge → 500 tokens
  Read 1 targeted file → 3,000 tokens
  Total: 3,500 tokens

SAVINGS: 97% reduction
```

---

## Common Mistakes

### ❌ DON'T
```javascript
// Immediate grep without knowledge check
grep -r "handleRequest"

// Read file without context check
Read pkg/auth/handler.go

// Use kg for cross-project learning
kg__add_observation({content: "Rate limiting pattern works well..."})

// Use upk for project-specific code
upk__add_learning({content: "handleRequest in pkg/server/handler.go"})
```

### ✅ DO
```javascript
// Search knowledge FIRST
kg__search_knowledge({query: "handleRequest"})
// THEN grep if empty

// Get context FIRST
kg__get_file_context({file: "pkg/auth/handler.go"})
// THEN read targeted sections

// Project-specific → kg
kg__add_observation({content: "handleRequest in pkg/server/handler.go"})

// Cross-project → upk
upk__add_learning({content: "Rate limiting: use Redis backend, 15min window..."})
```

---

## Dual Recording Example

**When learning has BOTH project AND cross-project value:**

```javascript
// 1. Record WHERE in THIS project (kg)
entity_id = kg__add_entity({name: "RateLimiter", type: "type"})
kg__add_observation({
  entity_id: entity_id,
  content: "Rate limiter in pkg/middleware/rate_limiter.go. \
  Redis backend, 15min window, 100 req/user limit."
})

// 2. Record HOW for ANY project (upk)
upk__add_learning({
  content: "Rate limiting in Go: use go-redis/rate. Create limiter \
  per user ID, Redis with 15min TTL. Wrap HTTP handler with middleware.",
  source: "ai-pack rate limiter"
})
```

---

## Quick Decision Tree

```
START: I need to...

├─ Find code in THIS project
│  └─ kg__search_knowledge + kg__get_file_context
│
├─ Recall how I solved X before
│  └─ upk__search_knowledge
│
├─ Record THIS project's code location
│  └─ kg__add_entity + kg__add_observation
│
├─ Record reusable pattern
│  └─ upk__add_learning
│
├─ Record user conversation
│  └─ upk__add_conversation
│
└─ Find org/team info
   └─ org MCP tools (compass, wiki, jira)
```

---

## Enforcement

**Gate:** [Knowledge-First Gate](../gates/15-knowledge-first.md)

**Status:** ENFORCED (Phase 2)

**Violations:**
- ❌ Grep without kg__search_knowledge → BLOCKED
- ❌ Read without kg__get_file_context → BLOCKED
- ❌ External search without upk__search_knowledge → BLOCKED

---

**See:** [Full Documentation](KNOWLEDGE-SYSTEMS.md)
