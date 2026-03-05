# Knowledge Graph Reader
<!-- skills/kg_reader.skill.md -->

**Version:** 1.1
**InjectAt:** role_context
**Slot:** 20
**Tools:** mcp__kg__search_knowledge, mcp__kg__get_file_context, mcp__kg__query_graph, mcp__kg__get_preflight_context
**Gates:** (none)
**MaxExtraTokens:** 10000
**Optional:** true

---

## Knowledge Graph — Read Before You Grep

The project knowledge graph is indexed from the codebase. **Search it before reaching for `grep` or `glob`.** The KG answers architectural questions faster and with more context than raw file search.

### When to use each tool

| Task | Use this | Instead of |
|------|----------|------------|
| Find a function, type, or component | `mcp__kg__search_knowledge` | `grep -r functionName` |
| Understand what a file contains | `mcp__kg__get_file_context` | `cat file.go` |
| Trace call chains / dependencies | `mcp__kg__query_graph` (Cypher) | reading every file in a package |
| Understand prior decisions on a topic | `mcp__kg__search_knowledge` | grepping for comments |

### Tools

- **`mcp__kg__search_knowledge`** `{query: string, limit?: number}` — Hybrid keyword+vector search across all entities and observations. Start here for any "where is X" or "how does Y work" question.
- **`mcp__kg__get_file_context`** `{file: string}` — Returns all functions, types, and imports defined in a file path. Use before reading the whole file.
- **`mcp__kg__query_graph`** `{cypher: string}` — Read-only Cypher query for precise graph traversal (dependencies, call chains, entity counts). Example: `MATCH (f:function)-[:CALLS]->(g:function {name:"parseConfig"}) RETURN f.name`
- **`mcp__kg__get_preflight_context`** `{task: string}` — Returns the most relevant entities for a task description. Use at the start of a new sub-task if you need fresh context.

### Required workflow

1. **Before any code exploration** — call `mcp__kg__search_knowledge` with the component or concept you're looking for.
2. **Before reading a file** — call `mcp__kg__get_file_context` to know which functions/types are worth reading.
3. **Only fall back to `grep`/`glob`** if the KG search returns no useful results. If it returns nothing for a term that should exist, the KG may not be indexed yet — note this and proceed with file search.
