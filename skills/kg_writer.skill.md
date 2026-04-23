# Project Knowledge Graph Writer
<!-- skills/kg_writer.skill.md -->

**Version:** 1.2
**InjectAt:** role_context
**Slot:** 25
**Tools:** kg__add_entity, kg__add_observation, kg__link_entities, kg__index_project
**Gates:** knowledge-first
**MaxExtraTokens:** 10000
**Optional:** true

---

## Project Knowledge Graph — Write Back What You Learn

When you discover something meaningful about **THIS project** during a task, record it in the project knowledge graph. Future agents (and future you) working on THIS project will benefit.

### ⚠️ MANDATORY: Write Back Findings

**REQUIRED operations after learning:**
- ✅ MUST record project-specific discoveries with `kg__add_entity` or `kg__add_observation`
- ✅ MUST write incrementally (don't wait until task completes)
- ✅ MUST link related entities with `kg__link_entities`

This is enforced by the **[Knowledge-First Gate](../gates/15-knowledge-first.md)**.

### Scope: Project-Specific Only

**Record to THIS project's KG:**
- ✅ Code entities in THIS codebase (functions, types, files)
- ✅ Architecture decisions for THIS project
- ✅ Bug investigations in THIS project
- ✅ Component relationships HERE
- ✅ Design rationale specific to THIS codebase

**Do NOT record to project KG:**
- ❌ Cross-project learnings → use UPK (see upk_writer skill)
- ❌ General design patterns → use UPK (see upk_writer skill)
- ❌ User conversations → use UPK (see upk_writer skill)
- ❌ Tool comparisons → use UPK (see upk_writer skill)

### What to Record (Project-Specific)

| Discovery | How to Record |
|-----------|--------------|
| New function, type, or file created in THIS project | `kg__add_entity` (type: function/type/file) |
| Why a design decision was made for THIS codebase | `kg__add_observation` on the relevant entity |
| A bug's root cause and fix in THIS project | `kg__add_observation` on the affected component |
| A new dependency between components HERE | `kg__link_entities` (relation: DEPENDS_ON, CALLS, IMPORTS) |
| THIS codebase significantly changed | `kg__index_project` to re-index |

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

### KG After Reasoning (MANDATORY)

When structured/sequential thinking concludes with a validated finding, decision, or ruled-out hypothesis about THIS project — **write it back immediately**.

```
AFTER reasoning concludes about project-specific topic:
  kg__add_entity({name: "<topic>", type: "topic"})  ← get entity ID
  kg__add_observation({entity_id: "<id>", content:
    "[REASONING] <conclusion, what was validated or eliminated, confidence>"})
```

Future agents working on the same topic in THIS project start from your conclusion, not from scratch. Reasoning that lives only in a transient response is lost.

### Dual Recording (Project + Personal)

Some discoveries have both project-specific AND cross-project value:

```
EXAMPLE: Implemented circuit breaker pattern in THIS project

STEP 1: Record project-specific (KG)
  entity_id = kg__add_entity({
    name: "CircuitBreaker",
    type: "type"
  })
  kg__add_observation({
    entity_id: entity_id,
    content: "Circuit breaker for external API calls. Config: \
    maxRequests=100, timeout=500ms, interval=30s. Located in \
    pkg/resilience/circuit_breaker.go"
  })

STEP 2: Record cross-project learning (UPK)
  upk__add_learning({
    content: "Circuit breaker in Go: used sony/gobreaker. Config: \
    maxRequests=100, timeout=500ms, interval=30s. Pairs well with Redis \
    cache for degraded mode. Pattern applicable to any external dependency.",
    source: "ai-pack circuit breaker implementation"
  })
```

**Result:** Project KG has THIS implementation's location and config. Personal knowledge has the reusable pattern for future projects.

---

## See Also

- **[KG Reader](kg_reader.skill.md)** - Searching project knowledge
- **[UPK Writer](upk_writer.skill.md)** - Recording cross-project learnings
- **[Knowledge-First Gate](../gates/15-knowledge-first.md)** - Enforcement policy
