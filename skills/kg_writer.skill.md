# Knowledge Graph Writer
<!-- skills/kg_writer.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 25
**Tools:** mcp__kg__create_entities, mcp__kg__create_relations, mcp__kg__add_observations, mcp__kg__delete_entities, mcp__kg__delete_relations, mcp__kg__delete_observations
**Gates:** (none)
**MaxExtraTokens:** 10000
**Optional:** true

---

## Knowledge Graph Access (Write)

You have write access to the project knowledge graph. After completing
significant work, update the graph to reflect new components, decisions,
and relationships:

- `create_entities` — record new components, services, or concepts
- `create_relations` — link entities (dependencies, ownership, data flow)
- `add_observations` — annotate existing entities with new findings
- `delete_entities` / `delete_relations` / `delete_observations` — remove stale entries

Keep graph updates focused and accurate. Prefer observations over new entities
for incremental discoveries on existing components.
