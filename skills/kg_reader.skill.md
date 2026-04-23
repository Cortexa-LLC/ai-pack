# Project Knowledge Graph Reader
<!-- skills/kg_reader.skill.md -->

**Version:** 1.2
**InjectAt:** role_context
**Slot:** 20
**Tools:** kg__search_knowledge, kg__get_file_context, kg__query_graph, kg__get_preflight_context
**Gates:** knowledge-first
**MaxExtraTokens:** 10000
**Optional:** true

---

## Project Knowledge Graph — Read Before You Grep

The **project knowledge graph** is indexed from **this codebase only**. **Search it before reaching for `grep` or `glob`.** The KG answers architectural questions faster and with more context than raw file search.

### ⚠️ MANDATORY: This gate is ENFORCED

**FORBIDDEN operations without KG check:**
- ❌ Grep before kg__search_knowledge (for finding code)
- ❌ Read files before kg__get_file_context (for understanding file contents)
- ❌ Starting task without kg__get_preflight_context (for gathering context)

**REQUIRED workflow:**
1. MUST search project KG before file searches
2. ONLY IF KG returns nothing → use grep/glob
3. MUST record project findings back to KG

This is enforced by the **[Knowledge-First Gate](../gates/15-knowledge-first.md)**.

### Scope: Project-Specific Only

**This project's KG contains:**
- ✅ Code entities (functions, types, files, modules) in THIS codebase
- ✅ Architecture decisions for THIS project
- ✅ Component relationships HERE
- ✅ Past bug investigations in THIS project

**This project's KG does NOT contain:**
- ❌ Cross-project learnings (use UPK - see upk_reader skill)
- ❌ User conversations (use UPK - see upk_reader skill)
- ❌ Organizational data (use org MCP tools)
- ❌ Code from other projects

### When to use each tool

| Task | Use this | Instead of |
|------|----------|------------|
| Find a function, type, or component in THIS project | `kg__search_knowledge` | `grep -r functionName` |
| Understand what a file contains | `kg__get_file_context` | `cat file.go` |
| Trace call chains / dependencies in THIS codebase | `kg__query_graph` (Cypher) | reading every file in a package |
| Understand prior decisions on THIS project | `kg__search_knowledge` | grepping for comments |

### Tools

- **`kg__search_knowledge`** `{query: string, limit?: number}` — Hybrid keyword+vector search across all entities and observations. Start here for any "where is X" or "how does Y work" question.
- **`kg__get_file_context`** `{file: string}` — Returns all functions, types, and imports defined in a file path. Use before reading the whole file.
- **`kg__query_graph`** `{cypher: string}` — Read-only Cypher query for precise graph traversal (dependencies, call chains, entity counts). Example: `MATCH (f:function)-[:CALLS]->(g:function {name:"parseConfig"}) RETURN f.name`
- **`kg__get_preflight_context`** `{task: string}` — Returns the most relevant entities for a task description. Use at the start of a new sub-task if you need fresh context.

### Required Workflow (MANDATORY)

**BEFORE any file operation:**

```
BEFORE Grep OR Glob OR Read THEN
  STEP 1: Search project KG
    result = kg__search_knowledge({query: "<component or concept>"})
  
  STEP 2: Evaluate result
    IF result answers question THEN
      use KG result
      SKIP file search
    ELSE
      log "KG returned no results - proceeding to file search"
      grep OR glob
      RECORD findings with kg__add_entity/observation (see kg_writer skill)
    END IF
END BEFORE

BEFORE Read(file.go) THEN
  context = kg__get_file_context({file: "path/to/file.go"})
  
  IF context exists THEN
    review functions/types in file
    Read only relevant sections
  ELSE
    Read entire file
    consider kg__index_project() if major gap
  END IF
END BEFORE

BEFORE starting task THEN
  kg__get_preflight_context({task: "<task description>"})
  → surfaces relevant entities for this work
END BEFORE
```

### KG Before Reasoning (MANDATORY)

When about to reason through a question involving existing code, past decisions, or known system state — **search the KG first**. Avoid re-deriving facts that are already stored. Ground your reasoning in known facts before expanding hypotheses.

```
BEFORE structured/sequential thinking on any project-specific topic:
  kg__search_knowledge({query: "<topic>"})
  → if the answer is already there, skip the reasoning and act on what's known
```

### Token Efficiency

```
Cost comparison for "find handleRequest function":

Without KG:
  grep -r "handleRequest" → 5,000-50,000 tokens (reads all matches)
  Read 10+ files → 30,000+ tokens
  Total: ~50,000 tokens

With KG:
  kg__search_knowledge({query: "handleRequest"}) → 500 tokens
  → Result: "handleRequest in pkg/server/handler.go:42"
  Read 1 file, 1 function → 3,000 tokens
  Total: ~3,500 tokens

SAVINGS: 93% token reduction
```

---

## See Also

- **[KG Writer](kg_writer.skill.md)** - Recording project knowledge
- **[UPK Reader](upk_reader.skill.md)** - Personal cross-project knowledge
- **[Knowledge-First Gate](../gates/15-knowledge-first.md)** - Enforcement policy
