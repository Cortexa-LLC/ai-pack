# Skill Composition Algorithm

**Version:** 1.0.0
**Status:** Accepted (ADR 004)
**Last Updated:** 2026-02-26

---

## Overview

When an agent is spawned the server assembles its `AgentConfig` by loading the role file,
resolving every skill it lists, and composing the final configuration. This document
specifies that algorithm precisely so that implementations are interoperable.

---

## Input → Output

**Input:**
- Role name (e.g., `engineer`)
- Project root path
- `.ai/workflow.md` (optional gate overrides)

**Output:** `AgentConfig` with:
- `Tools` — union of role tools + all skill tools (deduped)
- `Context.Gates` — union of all active skill gates after workflow overrides
- `Context.RoleContent` — assembled system prompt (role body + injected fragments)
- `SkillsLoaded []string` — names of skills successfully composed
- `MaxBudgetTokens` — role budget + sum of skill MaxExtraTokens (capped at model max)

---

## Algorithm

```
FUNCTION ComposeAgentConfig(roleName, projectRoot):

  1. LOAD role file
     path = resolveRolePath(roleName, projectRoot)     # .ai/roles/ → roles/ → etc.
     roleConfig, roleBody = parseRoleMarkdown(path)
     skillNames = roleConfig.Skills                    # may be empty → ["general"]
     IF skillNames IS EMPTY:
       skillNames = ["general"]

  2. RESOLVE skills
     skills = []
     FOR name IN skillNames:
       path = resolveSkillPath(name, projectRoot)      # .ai/skills/ → skills/
       IF path EXISTS:
         skill = parseSkillMarkdown(path)
         skills.APPEND(skill)
       ELSE:
         LOG WARNING "skill not found: {name}, skipping"
         // unknown skill → skip, do not fail hard

  3. APPLY workflow gate overrides
     overrides = parseWorkflowGateConfig(projectRoot) // .ai/workflow.md
     // overrides.disable[roleName] = [gate1, gate2, ...]

     FOR skill IN skills:
       IF skill.Optional:
         disabledGates = overrides.disable[roleName] OR []
         skill.Gates = skill.Gates - disabledGates
       // Non-optional skill gates are never removed

  4. MERGE tools
     toolSet = SET(roleConfig.Tools)
     FOR skill IN skills:
       toolSet.ADD_ALL(skill.Tools)
     agentConfig.Tools = SORTED_LIST(toolSet)           // deterministic order

  5. MERGE gates
     gateSet = SET()
     FOR skill IN skills:
       gateSet.ADD_ALL(skill.Gates)
     agentConfig.Context.Gates = SORTED_LIST(gateSet)

  6. MERGE budget
     extra = SUM(skill.MaxExtraTokens FOR skill IN skills)
     agentConfig.MaxBudgetTokens = MIN(roleConfig.MaxBudgetTokens + extra, MODEL_MAX)

  7. ASSEMBLE system prompt
     preamble = []
     role_context = []
     postamble = []

     FOR skill IN skills SORTED BY (InjectAt, Slot):
       SWITCH skill.InjectAt:
         CASE "preamble":   preamble.APPEND(skill.PromptFragment)
         CASE "role_context": role_context.APPEND(skill.PromptFragment)
         CASE "postamble":  postamble.APPEND(skill.PromptFragment)

     agentConfig.Context.RoleContent = JOIN([
       JOIN(preamble,      separator="\n\n---\n\n"),
       roleBody,                                        // the role's own body, unmodified
       JOIN(role_context,  separator="\n\n---\n\n"),
       JOIN(postamble,     separator="\n\n---\n\n"),
     ], separator="\n\n")

  8. RETURN agentConfig

END FUNCTION
```

---

## Skill Path Resolution

```
FUNCTION resolveSkillPath(name, projectRoot):
  candidates = [
    projectRoot + "/.ai/skills/" + name + ".skill.md",      // project override
    projectRoot + "/.ai-pack/skills/" + name + ".skill.md", // submodule override
    projectRoot + "/skills/" + name + ".skill.md",           // framework default
    "../skills/" + name + ".skill.md",                       // dev: parent dir
  ]
  RETURN first existing path in candidates
```

---

## Gate Override Format (`.ai/workflow.md`)

Projects can suppress or add gates per role without modifying role or skill files.

```markdown
# Workflow Configuration

## Skill Gate Overrides

| Role | Action | Gates |
|------|--------|-------|
| engineer | disable | tdd-enforcement |
| reviewer | disable | arch-review |
| * | disable | (none) |
```

Or equivalently via YAML front-matter:

```yaml
---
skill_gates:
  engineer:
    disable:
      - tdd-enforcement
  reviewer:
    disable:
      - arch-review
---
```

Rules:
- Only `Optional: true` skill gates can be disabled
- `Optional: false` skill gates ignore the override and emit a warning
- `*` wildcard applies to all roles
- Role-specific entries take precedence over wildcard

---

## Prompt Assembly — Visual

```
┌─────────────────────────────────────────────────┐
│  PREAMBLE (InjectAt: preamble, Slot: 10..19)    │
│  general.skill.md fragment                      │
├─────────────────────────────────────────────────┤
│  ROLE BODY  (the role's own markdown below ---)  │
│  (immutable, never modified)                    │
├─────────────────────────────────────────────────┤
│  ROLE_CONTEXT (InjectAt: role_context)          │
│  Slot 20: kg_reader fragment                    │
│  ─────────────────────────────────────────────  │
│  Slot 25: kg_writer fragment                    │
│  ─────────────────────────────────────────────  │
│  Slot 50: tdd fragment                          │
├─────────────────────────────────────────────────┤
│  POSTAMBLE (InjectAt: postamble)                │
│  Slot 80: task completion reminder              │
└─────────────────────────────────────────────────┘
```

---

## Data Model Changes (`AgentConfig`)

The existing `AgentConfig` struct (in `internal/server/server_core.go`) needs these
additions to support skill composition:

```go
type AgentConfig struct {
    // --- existing fields unchanged ---
    Name        string
    Description string
    Tier        string
    Model       string
    Context     struct {
        RoleFile    string
        RoleContent string   // now = assembled from role body + skill fragments
        Gates       []string // now = union of skill gates (after overrides)
        AdditionalInstructions string
    }
    Tools       []string    // now = union of role tools + skill tools

    // --- new fields ---
    Skills      []string    // skill names declared by the role
    SkillsLoaded []string   // skills successfully resolved and composed
}

type SkillConfig struct {
    Name              string
    Version           string
    InjectAt          string   // "preamble" | "role_context" | "postamble"
    Slot              int
    Tools             []string
    Gates             []string
    MaxExtraTokens    int64
    Optional          bool
    PromptFragment    string
}
```

---

## Error Handling

| Condition | Behaviour |
|-----------|-----------|
| Skill file not found | Log warning, skip skill, continue composition |
| Skill file malformed | Log error, skip skill, continue composition |
| Unknown `InjectAt` value | Default to `role_context` |
| Missing `Slot` | Default to `50` |
| Required skill (`Optional: false`) not found | Log error, fail agent spawn |
| Gate override targets `Optional: false` gate | Log warning, ignore override |

---

## Backward Compatibility

Roles without `**Skills:**` header receive only the `general` skill implicitly.
This means:
- All existing role files continue to work without modification
- Tools and system prompt are unchanged for un-migrated roles
- Gates declared in old-style `**Gates:**` role headers are still parsed and respected
  (the old `**Gates:**` parser is kept for backward compatibility during migration)
- Once a role is migrated (Skills header added, Gates header removed), the old path
  is no longer used for that role

---

## Migration Sequence

The migration from flat role files to OCP skill composition is a three-phase process:

### Phase 1 — Create Skill Files (no role changes)

Create `skills/*.skill.md` for each capability currently embedded in roles:

| Skill | Gates Extracted | Prompt Extracted From |
|-------|-----------------|-----------------------|
| `general` | (none) | Fallback / new |
| `tdd` | `tdd-enforcement` | `roles/engineer.md` |
| `kg_reader` | (none) | `roles/shared/agent-policy.md` KG section |
| `kg_writer` | (none) | `roles/shared/agent-policy.md` KG section |
| `code_review` | `code-quality-review` | `gates/35-code-quality-review.md` header |
| `arch_review` | `architectural-review` | `roles/architect.md` |

### Phase 2 — Add Skills header to role files

For each role, add:
```
**Skills:** general, <skill1>, <skill2>, ...
```

Keep the `**Gates:**` header during transition (backward-compat parser reads both).

### Phase 3 — Remove Gates from role files

Once all roles have `**Skills:**` headers and the composition runtime is verified,
remove `**Gates:**` headers from role files. The gates are now owned by skills.

---

## Implementation Checklist

| Task | File | Notes |
|------|------|-------|
| `parseSkillMarkdown()` | `internal/server/server_helpers.go` | New function |
| `resolveSkillPath()` | `internal/server/server_helpers.go` | New function |
| `composeSkills()` | `internal/server/server_helpers.go` | New function |
| Extend `AgentConfig` | `internal/server/server_core.go` | Add `Skills`, `SkillsLoaded` |
| Extend `SkillConfig` | `internal/server/server_core.go` | New struct |
| Call `composeSkills` in `loadAgentConfig` | `internal/server/server_helpers.go` | After role parse |
| Parse `.ai/workflow.md` gate overrides | `internal/server/server_helpers.go` | New function |
| Create `skills/general.skill.md` | `skills/` | Foundation skill |
| Create `skills/tdd.skill.md` | `skills/` | Phase 1 migration |
| Create `skills/kg_reader.skill.md` | `skills/` | Phase 1 migration |
| Create `skills/kg_writer.skill.md` | `skills/` | Phase 1 migration |
| Create `skills/code_review.skill.md` | `skills/` | Phase 1 migration |
| Create `skills/arch_review.skill.md` | `skills/` | Phase 1 migration |

---

## Related Documents

- ADR: [docs/adr/004-role-skill-ocp.md](../adr/004-role-skill-ocp.md)
- Schema: [docs/architecture/skill-schema.md](skill-schema.md)
