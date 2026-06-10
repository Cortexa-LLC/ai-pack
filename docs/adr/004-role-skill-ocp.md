# ADR 004: OCP-Based Skill Composition for Roles

**Status:** Accepted
**Date:** 2026-02-26
**Deciders:** AI-Pack Core Team
**Related:** [Skill Schema](../architecture/skill-schema.md), [Composition Algorithm](../architecture/skill-composition.md)

---

## Context and Problem Statement

The current role system has three Open/Closed Principle violations that make the framework
brittle to extend:

1. **Adding a capability requires editing a role file.** Tools, gates, and prompt sections
   live inside the role `.md`. Every new feature (TDD enforcement, KG access, graph querying)
   requires opening and modifying the role definition. Roles should be *closed* to
   modification once deployed.

2. **Gates are baked into the role.** `**Gates:** tdd-enforcement, code-quality-review`
   in `roles/engineer.md` conflates *who the agent is* with *what workflow policy governs it*.
   A team that does not use TDD must edit the role file to remove the gate — a direct OCP
   violation. Gates are workflow policy; roles are identity.

3. **No reuse of capabilities across roles.** KG access is copy-pasted into every role that
   needs it. If the KG connection protocol changes, every role file must be updated.

The framework needs a mechanism where roles are stable and immutable, but their capabilities
can be extended by composing discrete units of functionality called **skills**.

---

## Decision

Introduce **skills** as the fundamental unit of composable capability.

### What a skill is

A skill is a markdown file (`skills/<name>.skill.md`) that declares:
- **Tools** it adds to the agent's tool set
- **Gates** it enforces when active
- **A prompt fragment** injected at a defined position in the system prompt
- **Metadata** controlling injection order and optionality

### What a role becomes

A role file is *closed* — it declares only:
- Identity and model parameters (tier, model, context, delegation, tools, timeouts)
- The ordered list of skills it accepts: `**Skills:** general, tdd, kg_reader, kg_writer`

The role body describes *who the agent is*, not *what it can do*. Capabilities come
from skills attached at spawn time.

### How gates are separated

`**Gates:**` is removed from role files entirely. Gates are declared by skills (an agent
gains gate enforcement when it gains the skill) or configured in `.ai/workflow.md`
(project-level override). This makes workflow policy independent of role identity.

### Composition at spawn time

When an agent is spawned with role `engineer` the server:
1. Loads `roles/engineer.md` → parses `**Skills:**` list
2. For each skill name, loads `skills/<name>.skill.md` (project `.ai/skills/` first)
3. Merges: tools union, gates union, ordered prompt fragments
4. Applies `.ai/workflow.md` gate overrides (allow-list / deny-list per role)
5. Assembles the final `AgentConfig` with composed system prompt

The role file is never modified. New capabilities are added by creating new skill files
and listing them in the role's `**Skills:**` header.

---

## Consequences

### Positive

- **Roles are immutable.** Deploying a new capability never requires touching role files.
- **Gates are portable.** TDD enforcement lives in `skills/tdd.skill.md`; any role can
  gain or lose it by listing or omitting it in `**Skills:**`.
- **Skills are reusable.** `kg_reader` is defined once; both `engineer` and `reviewer`
  list it without copying its prompt or gate.
- **Project customization is clean.** Projects drop a file in `.ai/skills/` to override
  or add a skill without touching the shared role definition.
- **Backward compatible.** Roles without a `**Skills:**` line are treated as having only
  `general`; existing role files keep working unchanged.

### Negative / Trade-offs

- **New indirection.** Understanding an agent's full capability now requires reading the
  role file *and* every skill file it references.
- **Ordering matters.** Skill prompt fragments inject at positions (`preamble`, `role_context`,
  `postamble`); wrong ordering produces unexpected system prompts. The schema mandates
  explicit `**Slot:**` values to make order deterministic.
- **Migration effort.** Existing gates in role files must be extracted into skill files.
  This is a one-time migration tracked in the implementation plan.

### Neutral

- Existing `.md` role format is preserved; only the `**Gates:**` header moves out.
- The `general` skill becomes the implicit base every agent receives even when no role
  is specified, replacing the current "no-role" fallback.

---

## Alternatives Considered

### A. YAML overlay files per role

Each role gets a sidecar `roles/engineer.yaml` with tools and gates. Skills remain
implicit; configuration is role-centric.

**Why not chosen:** Doesn't solve reuse — every role YAML still duplicates KG config,
TDD gates, etc. Two file formats (`.md` + `.yaml`) per role adds complexity with no
composability gain.

### B. Plugin / hook system (runtime code injection)

Skills are Go plugins or registered hook functions that the server loads dynamically.

**Why not chosen:** Over-engineered for the current use case. Markdown + structured
headers are already the established format across the framework. Go plugin compilation
adds toolchain complexity. The markdown-based approach delivers the same OCP benefit
with no new build dependencies.

### C. Inherit from a base role

Roles explicitly extend parent roles (`**Extends:** base-engineer`). Capabilities flow
via inheritance chain.

**Why not chosen:** Inheritance creates deep, opaque hierarchies. The diamond problem
arises when two parents declare conflicting gates. Composition (the chosen approach)
is more explicit and avoids these coupling issues.

---

## Implementation Scope

This ADR covers the **design specification only**. Implementation tasks:

1. Write skill parser in `internal/server/server_helpers.go` (extend `parseMarkdownConfig`)
2. Extend `AgentConfig` with `Skills []string`, composed `SkillTools`, `SkillGates`, `SkillPromptFragments`
3. Implement composition in `loadAgentConfig` — skill resolution, merge, gate overlay
4. Create `skills/` directory with initial skills: `general`, `tdd`, `kg_reader`, `kg_writer`, `code_review`, `arch_review`
5. Migrate gates out of existing role files into corresponding skills
6. Parse `.ai/workflow.md` gate overrides

Tracked in: `.ai/tasks/ai-pack-aew-20260225192348-role-skill-ocp/task.md`

---

## Related Documents

- Architecture: [docs/architecture/skill-schema.md](../architecture/skill-schema.md)
- Composition: [docs/architecture/skill-composition.md](../architecture/skill-composition.md)
- Role Extension Guide: [ROLE-EXTENSION-GUIDE.md](../../ROLE-EXTENSION-GUIDE.md)
