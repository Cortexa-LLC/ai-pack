# Knowledge Graph Writer
<!-- skills/kg_writer.skill.md -->

**Version:** 1.1
**InjectAt:** role_context
**Slot:** 25
**Tools:** kg__add_entity, kg__add_observation, kg__link_entities, kg__index_project
**Gates:** (none)
**MaxExtraTokens:** 10000
**Optional:** true

---

## Knowledge Graph — Write Back What You Learn

When you discover something meaningful during a task, record it in the knowledge graph. Future agents (and future you) will benefit.

### What to record

| Discovery | How to record |
|-----------|--------------|
| New function, type, or file created | `kg__add_entity` (type: function/type/file) |
| Why a design decision was made | `kg__add_observation` on the relevant entity |
| A bug's root cause and fix | `kg__add_observation` on the affected component |
| A new dependency between components | `kg__link_entities` (relation: DEPENDS_ON, CALLS, IMPORTS) |
| Codebase significantly changed | `kg__index_project` to re-index |

### Tools

- **`kg__add_entity`** `{name: string, type: string}` — Create or upsert an entity. Types: `function`, `type`, `file`, `module`, `topic`, `package`, `import`. Returns the entity ID.
- **`kg__add_observation`** `{entity_id: string, content: string}` — Attach a note to an entity: bug found, design decision, caveat, performance characteristic. Prefer observations over new entities for incremental findings.
- **`kg__link_entities`** `{from_id: string, relation: string, to_id: string}` — Create a directed relation. Relations: `CONTAINS`, `IMPORTS`, `CALLS`, `IMPLEMENTS`, `BELONGS_TO`, `DEPENDS_ON`, `RELATES_TO`.
- **`kg__index_project`** `{}` — Re-index the entire project. Use after making significant structural changes (new packages, major refactors).

### Required workflow

**Write incrementally — do not wait until the task is complete.**

The task may time out. Every finding written to the KG is preserved across retries. A timed-out task with KG notes is recoverable; one without is wasted work.

1. **At each significant discovery** — root cause confirmed, hypothesis ruled out, new lead found — write it immediately with `kg__add_observation`. Prefix investigation notes with `[INVESTIGATION]`.
2. Add observations to any entities you modified or learned something new about.
3. If you created new components, add them as entities and link them to their parents.
4. If root causes or design decisions became clear during the task, record them as observations — they're the highest-value knowledge for future agents.

### KG after reasoning

When structured/sequential thinking concludes with a validated finding, decision, or ruled-out hypothesis — **write it back immediately**.

```
AFTER reasoning concludes:
  kg__add_entity({name: "<topic>", type: "topic"})  ← get entity ID
  kg__add_observation({entity_id: "<id>", content:
    "[REASONING] <conclusion, what was validated or eliminated, confidence>"})
```

Future agents working on the same topic start from your conclusion, not from scratch. Reasoning that lives only in a transient response is lost.
