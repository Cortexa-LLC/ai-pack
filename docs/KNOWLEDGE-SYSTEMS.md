# Knowledge Systems Architecture

**Version:** 1.0
**Last Updated:** 2026-04-22

## Overview

AI-Pack uses a **three-tier knowledge architecture** to minimize token consumption and maximize agent efficiency:

```
┌─────────────────────────────────────────────────────────────┐
│                    KNOWLEDGE SYSTEMS                         │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Project    │  │   Personal   │  │    Org       │      │
│  │   KG (kg)    │  │   KG (upk)   │  │   Systems    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│        ↓                 ↓                  ↓                │
│   This project      All projects      Organization         │
│   code/arch         learnings          data/planning       │
│                                                               │
└─────────────────────────────────────────────────────────────┘
                           ↓
              ┌────────────────────────────┐
              │   Knowledge-First Gate     │
              │   (15-knowledge-first.md)  │
              └────────────────────────────┘
                           ↓
              Search knowledge BEFORE file ops
              Record findings BACK to knowledge
```

---

## Three-Tier Architecture

### 1. Project Knowledge Graph (KG)

**Scope:** Current project only  
**Tools:** `kg__*` (search, get_file_context, query_graph, add_entity, add_observation)  
**Skills:** kg_reader.skill.md, kg_writer.skill.md  

**Contains:**
- Code entities (functions, types, files, modules) in THIS codebase
- Architecture decisions for THIS project
- Component relationships HERE
- Past bug investigations in THIS project
- Call graphs and dependencies

**Use for:**
- Finding where code is located in THIS project
- Understanding THIS project's architecture
- Tracing dependencies within THIS codebase
- Reviewing past investigations HERE

**Example queries:**
```
kg__search_knowledge({query: "authentication middleware"})
  → "pkg/auth/middleware.go - JWT validation for all API routes"

kg__get_file_context({file: "pkg/auth/middleware.go"})
  → [ValidateJWT, RefreshToken, RevokeToken functions]

kg__query_graph({
  cypher: "MATCH (f:function {name: 'ValidateJWT'})-[:CALLS]->(g) RETURN g"
})
  → [parseToken, verifySignature, checkExpiry]
```

---

### 2. Personal Knowledge (UPK)

**Scope:** All projects, all conversations  
**Tools:** `upk__*` (search_knowledge, add_learning, add_conversation)  
**Skills:** upk_reader.skill.md, upk_writer.skill.md  

**Contains:**
- Cross-project learnings and insights
- Design patterns discovered across work
- Tool/library evaluations
- User conversations and preferences
- Solutions to recurring problems
- General techniques and approaches

**Use for:**
- "How have I solved X before?"
- "What patterns work for Y?"
- "What did I learn about Z?"
- "What did user prefer for A?"

**Example queries:**
```
upk__search_knowledge({query: "rate limiting implementation"})
  → "Learned: express-rate-limit middleware with Redis backend.
     Set windowMs=15min, max=100. Key by user ID for authenticated."

upk__search_knowledge({query: "conversation about database choice"})
  → "User explained preference for Postgres over DynamoDB because
     team has Postgres expertise and current scale doesn't justify
     DynamoDB complexity."
```

---

### 3. Organizational Knowledge

**Scope:** Company/team-specific  
**Tools:** MCP servers (compass, wiki, jira, etc.) - varies by organization  
**Skills:** None (org-specific, not included in open-source ai-pack)  

**Contains:**
- Team structures and org charts
- Planning artifacts (initiatives, milestones, projects)
- Work items (Jira issues, GitHub issues)
- Documentation (Confluence, wikis)
- Code repositories and applications
- Domain models and API specs

**Use for:**
- "Who owns service X?"
- "What are the Q2 initiatives?"
- "Find work items for team Y"
- "What's the status of project Z?"

**Example queries (organization-specific):**
```
compass__executeGremlin({
  query: "g.V().hasLabel('DedicatedTeam')
           .has('name', 'Seller Experience Engineering')
           .out('HAS_PROJECT').valueMap()"
})
  → List of projects owned by team

wiki__search_wiki({cql: "space = 'ARCH' AND text ~ 'auth'"})
  → Architectural docs about authentication
```

---

## Decision Tree: Which Knowledge System?

```
QUESTION from user or task:

├─ About THIS project's code/architecture?
│  ├─ YES → Use Project KG (kg__*)
│  │       Examples:
│  │       - "Where is auth handled?"
│  │       - "What does handleRequest function do?"
│  │       - "Why did we choose this pattern?"
│  │
│  └─ NO ↓

├─ About how I've solved similar problems before?
│  ├─ YES → Use Personal Knowledge (upk__*)
│  │       Examples:
│  │       - "How have I implemented rate limiting?"
│  │       - "What did I learn about GraphQL federation?"
│  │       - "What did user prefer for database choice?"
│  │
│  └─ NO ↓

├─ About organization/team/planning?
│  ├─ YES → Use Organizational MCP tools
│  │       Examples:
│  │       - "Who owns the auth service?"
│  │       - "What are our Q2 initiatives?"
│  │       - "Find Jira tickets for feature X"
│  │
│  └─ NO → External search (WebSearch, docs)
```

---

## Knowledge-First Enforcement

All three knowledge tiers are protected by the **[Knowledge-First Gate](../gates/15-knowledge-first.md)**, which enforces:

### ⚠️ MANDATORY Rules

1. **Search knowledge BEFORE file operations**
   - ❌ BLOCKED: `grep -r "pattern"` without prior `kg__search_knowledge`
   - ❌ BLOCKED: `Read(file)` without prior `kg__get_file_context`
   - ✅ ALLOWED: File search ONLY AFTER knowledge search returns empty

2. **Write findings BACK to knowledge**
   - After grep finds answer → `kg__add_observation` or `upk__add_learning`
   - After user conversation → `upk__add_conversation`
   - After solving novel problem → record to appropriate knowledge tier

3. **Use appropriate knowledge tier**
   - Project-specific → Project KG (kg)
   - Cross-project → Personal Knowledge (upk)
   - Organization → Org MCP tools

---

## Token Efficiency Gains

### Example: "Find handleRequest function"

**Without knowledge-first (old way):**
```bash
# Step 1: Grep everywhere
grep -r "handleRequest" .
# Result: 50,000 tokens (reads all matches across codebase)

# Step 2: Read multiple files
Read pkg/server/handler.go    # 10,000 tokens
Read pkg/api/handler.go       # 8,000 tokens
Read pkg/core/handler.go      # 7,000 tokens
# Total: 75,000 tokens
```

**With knowledge-first (new way):**
```javascript
// Step 1: Search project KG
kg__search_knowledge({query: "handleRequest"})
// Result: 500 tokens
// Answer: "handleRequest in pkg/server/handler.go:42"

// Step 2: Get file context
kg__get_file_context({file: "pkg/server/handler.go"})
// Result: 300 tokens
// Answer: [handleRequest, validateRequest, sendResponse functions]

// Step 3: Read targeted section
Read pkg/server/handler.go (offset: 40, limit: 20)
// Result: 2,000 tokens

// Total: 2,800 tokens
```

**Savings: 96% token reduction** (75,000 → 2,800)

---

## Recording Strategy: Dual Recording

Some discoveries have value at multiple tiers. Record to BOTH:

### Example: Circuit Breaker Implementation

```javascript
// 1. Project KG (where it's located in THIS project)
entity_id = kg__add_entity({
  name: "CircuitBreaker",
  type: "type"
})
kg__add_observation({
  entity_id: entity_id,
  content: "Circuit breaker for external API calls. Located in \
  pkg/resilience/circuit_breaker.go. Config: maxRequests=100, \
  timeout=500ms, interval=30s."
})

// 2. Personal Knowledge (reusable pattern for ANY project)
upk__add_learning({
  content: "Circuit breaker in Go: used sony/gobreaker. Config: \
  maxRequests=100, timeout=500ms, interval=30s. Pairs well with \
  Redis cache for degraded mode. Pattern: wrap external calls, \
  fail fast on timeout, recover gradually.",
  source: "ai-pack circuit breaker implementation"
})
```

**Result:**
- Future agents on THIS project know where circuit breaker is: Project KG
- Future agents on OTHER projects learn the pattern: Personal Knowledge

---

## Workflow Examples

### Workflow 1: Finding Code in Current Project

```
USER: "Where is authentication handled?"

AGENT WORKFLOW:

1. Search project KG (MANDATORY)
   kg__search_knowledge({query: "authentication"})
   → Result: "auth handled in pkg/auth/middleware.go, JWT validation"

2. Get file context
   kg__get_file_context({file: "pkg/auth/middleware.go"})
   → [ValidateJWT, RefreshToken, RevokeToken]

3. Read targeted function
   Read pkg/auth/middleware.go (find ValidateJWT function)
   → Read only that function, not entire file

4. Answer user
   "Authentication is handled in pkg/auth/middleware.go via the \
   ValidateJWT function. It validates JWT tokens for all API routes."

TOKENS USED: ~4,000 (vs ~50,000 without knowledge-first)
```

### Workflow 2: Implementing New Feature Using Past Learnings

```
USER: "Implement rate limiting for the API"

AGENT WORKFLOW:

1. Search personal knowledge (MANDATORY)
   upk__search_knowledge({query: "rate limiting implementation"})
   → Result: "I've used express-rate-limit with Redis. Config: \
             windowMs=15min, max=100, key by user ID."

2. Search project KG for existing rate limiting
   kg__search_knowledge({query: "rate limiting"})
   → Result: none (not implemented yet)

3. Implement based on past learning
   [create pkg/middleware/rate_limiter.go]
   [use learned pattern: Redis backend, 15min window, 100 req limit]

4. Record to project KG
   entity_id = kg__add_entity({name: "RateLimiter", type: "type"})
   kg__add_observation({
     entity_id: entity_id,
     content: "Rate limiter middleware. Redis backend, 15min window, \
     100 requests max per user. Located in pkg/middleware/rate_limiter.go"
   })

5. Record Go-specific learning to UPK
   upk__add_learning({
     content: "Rate limiting in Go: used go-redis/rate. Pattern: \
     create limiter per user ID, store in Redis with 15min TTL. \
     Integration: wrap HTTP handler with rate check middleware.",
     source: "ai-pack rate limiter implementation"
   })

RESULT:
- Reused past learning (no need to research rate limiting from scratch)
- Recorded Go implementation for future Go projects
- Recorded location in THIS project for future maintenance
```

### Workflow 3: Finding Team Context for Feature

```
USER: "Who should review the auth changes?"

AGENT WORKFLOW:

1. Search organizational knowledge (org MCP)
   compass__executeGremlin({
     query: "g.V().has('name', 'auth-service').in('OWNS').valueMap()"
   })
   → Result: "auth-service owned by 'Platform Security' team"

2. Get team members
   compass__executeGremlin({
     query: "g.V().hasLabel('DedicatedTeam')
             .has('name', 'Platform Security')
             .out('HAS_MEMBER').valueMap()"
   })
   → Result: [Alice (tech lead), Bob (senior eng), Carol (eng)]

3. Answer user
   "Auth changes should be reviewed by the Platform Security team. \
   Suggest: Alice (tech lead) or Bob (senior engineer)."

4. Record conversation to UPK
   upk__add_conversation({
     title: "Auth review - Platform Security team",
     summary: "Learned Platform Security team owns auth-service. \
     Alice is tech lead, Bob is senior engineer. Both suitable for \
     auth code reviews."
   })

RESULT:
- Found reviewer via org knowledge (not guessing)
- Recorded for future auth work
```

---

## Skills Reference

| Skill | Slot | Purpose | Key Tools |
|-------|------|---------|-----------|
| upk_reader | 15 | Search personal knowledge | upk__search_knowledge |
| upk_writer | 16 | Record learnings/conversations | upk__add_learning, upk__add_conversation |
| kg_reader | 20 | Search project code/arch | kg__search_knowledge, kg__get_file_context |
| kg_writer | 25 | Record project discoveries | kg__add_entity, kg__add_observation |

**Load order:** UPK → KG → other skills  
**Rationale:** Personal knowledge (cross-project) loaded first, project knowledge second, general tools last

---

## Gates Reference

| Gate | File | Purpose |
|------|------|---------|
| knowledge-first | gates/15-knowledge-first.md | Enforce knowledge check before file ops |

---

## Migration Guide

### For Existing Agents

**Before (no knowledge-first):**
```javascript
// Old way - immediate grep
grep -r "handleRequest"
// Reads 50,000 tokens
```

**After (knowledge-first):**
```javascript
// New way - check knowledge first
kg__search_knowledge({query: "handleRequest"})
// IF not found THEN grep
// Typically 500 tokens, only 50,000 if KG empty
```

### Rollout Phases

**Phase 1 (Current): Opt-in**
- Skills mark tools as `Optional: true`
- Agents can skip knowledge check
- Compliance tracked but not enforced

**Phase 2 (Target): Enforced**
- Gate blocks file ops without knowledge check
- Agents MUST search knowledge first
- Exceptions logged for review

**Phase 3 (Future): Intelligent**
- System auto-selects knowledge source
- Hybrid queries (knowledge + file search)
- Automatic knowledge indexing on writes

---

## Compliance Metrics

Track knowledge-first adoption:

```
Knowledge-First Ratio = knowledge_searches / total_searches

Targets:
- Phase 1 (opt-in): ≥ 40% knowledge-first
- Phase 2 (enforced): ≥ 80% knowledge-first
- Phase 3 (intelligent): ≥ 95% knowledge-first

Token Savings:
- Baseline: 50,000 tokens/query average (grep + read)
- With knowledge: 3,000 tokens/query average
- Target savings: 90%+ reduction
```

---

## See Also

- **[Knowledge-First Gate](../gates/15-knowledge-first.md)** - Enforcement policy
- **[UPK Reader Skill](../skills/upk_reader.skill.md)** - Personal knowledge search
- **[UPK Writer Skill](../skills/upk_writer.skill.md)** - Recording learnings
- **[KG Reader Skill](../skills/kg_reader.skill.md)** - Project knowledge search
- **[KG Writer Skill](../skills/kg_writer.skill.md)** - Recording project discoveries

---

**Last reviewed:** 2026-04-22  
**Next review:** After 30 days of enforcement data
