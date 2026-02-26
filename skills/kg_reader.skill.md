# Knowledge Graph Reader
<!-- skills/kg_reader.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 20
**Tools:** mcp__kg__search_nodes, mcp__kg__open_nodes, mcp__kg__read_graph
**Gates:** (none)
**MaxExtraTokens:** 10000
**Optional:** true

---

## Knowledge Graph Access (Read)

You have read access to the project knowledge graph via MCP tools:

- `search_nodes` — find entities by name, type, or observation content
- `open_nodes` — retrieve specific entities by exact name
- `read_graph` — read the entire graph (use sparingly; prefer search)

Consult the knowledge graph at the start of complex tasks to understand
architectural context, component relationships, and prior design decisions
before writing code or making recommendations.
