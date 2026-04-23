# Knowledge-First Gate

**Version:** 1.0.0
**Last Updated:** 2026-04-22

## Overview

The Knowledge-First Gate enforces mandatory knowledge system checks before expensive search operations. This gate prevents token waste from cumbersome grep/glob operations when answers already exist in indexed knowledge systems.

## Rationale

### Token Cost Comparison

| Operation | Avg Token Cost | Notes |
|-----------|---------------|-------|
| `kg__search_knowledge` | ~500 tokens | Returns indexed entities with context |
| `upk__search_knowledge` | ~300 tokens | Returns personal learnings and conversations |
| `grep -r "pattern"` | 5,000-50,000 tokens | Reads every matching line across codebase |
| `glob "**/*.go"` + Read files | 10,000+ tokens | Reads all files sequentially |

**Cost ratio:** Knowledge search is 10-100x more efficient than file search.

### Knowledge Systems

1. **Project Knowledge Graph (KG)** - `kg__*` tools
   - Code entities (functions, types, files, modules)
   - Architecture decisions
   - Past bug investigations
   - Component relationships

2. **Personal Knowledge (UPK)** - `upk__*` tools
   - Cross-project learnings
   - Conversation history
   - Design patterns discovered
   - Problem-solving approaches

3. **Organizational Knowledge** - MCP servers (e.g., `compass`, `wiki`, `jira`)
   - Team structures
   - Planning artifacts
   - Documentation
   - Work items

---

## Mandatory Workflow

### Rule 1: Knowledge Before Search

**REQUIRED:** Search appropriate knowledge system BEFORE using Grep, Glob, or file reads.

```
BEFORE Grep OR Glob OR Read THEN
  STEP 1: Determine search scope
    - Project code/architecture? → kg__search_knowledge
    - Cross-project patterns/learnings? → upk__search_knowledge
    - Organizational context? → org MCP tools

  STEP 2: Search knowledge system
    knowledge_result = search_knowledge(query)
  
  STEP 3: Evaluate result
    IF knowledge_result answers question THEN
      use knowledge result
      SKIP file search
    ELSE
      proceed to file search
      RECORD findings back to knowledge
    END IF
END BEFORE
```

### Rule 2: Specific File Context Before Reading

**REQUIRED:** Use `kg__get_file_context` BEFORE reading any code file.

```
BEFORE Read(file.go) THEN
  context = kg__get_file_context({file: "file.go"})
  
  IF context exists THEN
    review functions/types defined
    identify relevant sections
    Read only necessary portions
  ELSE
    Read entire file
    IF significant code THEN
      kg__index_project()  # Index for future
    END IF
  END IF
END BEFORE
```

### Rule 3: Write Back Findings

**REQUIRED:** Record discoveries immediately to knowledge systems.

```
AFTER finding significant information THEN
  IF project-specific (code/architecture) THEN
    kg__add_entity OR kg__add_observation
  ELSE IF cross-project learning THEN
    upk__add_learning
  ELSE IF conversation context THEN
    upk__add_conversation
  END IF
END AFTER
```

---

## Forbidden Operations

### ❌ GATE VIOLATIONS

These operations WITHOUT prior knowledge search are BLOCKED:

```
VIOLATION: Direct Grep without knowledge check
  ❌ grep -r "functionName"
  ✅ kg__search_knowledge({query: "functionName"}) THEN grep

VIOLATION: File reading without context check
  ❌ Read(pkg/auth/handler.go)
  ✅ kg__get_file_context({file: "pkg/auth/handler.go"}) THEN Read

VIOLATION: Codebase exploration without preflight
  ❌ Glob("**/*.go") then Read each
  ✅ kg__get_preflight_context({task: "auth flow"}) THEN targeted reads

VIOLATION: Searching for known patterns
  ❌ grep -r "TODO|FIXME|BUG"
  ✅ kg__search_knowledge({query: "technical debt"})
```

---

## Enforcement Rules

### 🚫 Hard Blocks

**IF** agent uses Grep/Glob/Read **WITHOUT** prior knowledge check **THEN**:
1. Operation is REJECTED
2. Agent MUST search knowledge first
3. Only after knowledge check returns insufficient results → file search allowed

### ⚠️ Soft Warnings

**IF** knowledge search returns empty **BUT** expected term should exist **THEN**:
1. Log: "KG may not be indexed - proceeding to file search"
2. After finding answer: "Recording to KG for future queries"
3. Call `kg__index_project()` if major gap detected

### 📊 Compliance Tracking

Track knowledge-first compliance:
```
Knowledge-First Ratio = knowledge_searches / (knowledge_searches + file_searches)

Target: ≥ 80% knowledge-first
Warning: < 60% knowledge-first
Critical: < 40% knowledge-first
```

---

## Knowledge System Selection

### Decision Tree

```
QUESTION: "Where is X implemented?"
  → SCOPE: Project code
  → USE: kg__search_knowledge({query: "X implementation"})
  → FALLBACK: grep "X" if KG empty

QUESTION: "How have I solved Y before?"
  → SCOPE: Cross-project
  → USE: upk__search_knowledge({query: "Y solution"})
  → FALLBACK: none (unique to personal knowledge)

QUESTION: "Who owns service Z?"
  → SCOPE: Organizational
  → USE: org MCP tools (compass, wiki)
  → FALLBACK: none (not in code)

QUESTION: "What does function F do?"
  → SCOPE: Project code
  → USE: kg__get_file_context({file: "path/to/file"})
  → THEN: Read specific function
```

---

## Knowledge Query Best Practices

### Effective Knowledge Queries

```
✅ GOOD queries:
- "authentication middleware"          → concept-based
- "rate limiting implementation"       → behavior-based
- "database connection pool"           → component-based
- "JWT token validation bug"           → prior investigation

❌ POOR queries:
- "func"                              → too generic
- "auth.go"                           → filename (use kg__get_file_context)
- "line 42"                           → too specific
```

### Query Refinement

```
IF initial query returns no results THEN
  STEP 1: Broaden query
    "parseUserConfig" → "user config" → "config"
  
  STEP 2: Try related terms
    "JWT" → "authentication" → "auth"
  
  STEP 3: Search by component
    kg__query_graph("MATCH (f:function)-[:BELONGS_TO]->(m:module {name: 'auth'}) RETURN f")
  
  STEP 4: Fall back to file search
    grep -r "parseUserConfig"
  
  STEP 5: Record findings
    kg__add_entity + kg__add_observation
END IF
```

---

## Integration with Other Gates

This gate works with:

- **[Tool Policy Gate](20-tool-policy.md)** - Enforces knowledge-first before Grep/Glob
- **[Lean Flow Gate](05-lean-flow.md)** - Reduces WIP by avoiding unnecessary file reads
- **[Execution Strategy Gate](25-execution-strategy.md)** - Parallel knowledge searches before parallel work

---

## Exceptions

### When Knowledge Check Can Be Skipped

1. **User provides explicit file path**: "Read src/main.go" → direct Read allowed
2. **Following stack trace**: Error points to file:line → direct Read allowed
3. **Iterating known file list**: Already have file context → skip re-check
4. **Knowledge system unavailable**: KG server down → fall back to file search

### Emergency Override

```
IF knowledge system timeout OR unavailable THEN
  log "Knowledge system unavailable - falling back to file search"
  proceed with Grep/Glob
  queue findings for knowledge write-back when system recovers
END IF
```

---

## Success Metrics

### Efficiency Gains

```
Baseline (no knowledge):
  Query: "where is auth handled?"
  → grep -r "auth" (30,000 tokens)
  → Read 15 files (50,000 tokens)
  → Total: 80,000 tokens

With knowledge-first:
  Query: "where is auth handled?"
  → kg__search_knowledge({query: "auth"}) (500 tokens)
  → Result: "auth handled in pkg/auth/middleware.go"
  → Read 1 file (3,000 tokens)
  → Total: 3,500 tokens
  
Savings: 96% token reduction
```

---

## Rollout Strategy

### Phase 1: Opt-in (Current)
- Skills mark knowledge tools as `Optional: true`
- Agents can skip if needed
- Track compliance metrics

### Phase 2: Enforced (Target)
- Gate blocks non-compliant operations
- Knowledge check required before file search
- Exceptions logged and reviewed

### Phase 3: Intelligent (Future)
- System auto-selects knowledge source
- Hybrid queries (knowledge + file search)
- Automatic knowledge indexing

---

**Last reviewed:** 2026-04-22
**Next review:** After 30 days of enforcement data
