# ADR 006: Three-Tier OCP Model for Role Extension and Inheritance

**Status:** Accepted
**Date:** 2026-02-28
**Deciders:** AI-Pack Core Team
**Supersedes:** Partial guidance in ADR 004 (skill composition only)
**Related:** [ADR 004 — Role/Skill OCP](004-role-skill-ocp.md), [Skill Schema](../architecture/skill-schema.md), [Skill Composition](../architecture/skill-composition.md)

---

## Context and Problem Statement

ADR 004 established that skills can be composed into roles (Tier 1 extension). It left
two gaps unresolved:

1. **Role identity cannot be extended, only replaced.** The current `.ai/roles/<name>.md`
   mechanism is full substitution. A project that needs to change one field (e.g., swap
   `tdd` for `bdd`) must copy the entire 60-line role file. Every subsequent upstream
   change to the base role must be manually merged.

2. **No decision framework.** Users have three overlapping mechanisms (skills, full
   override, extension) with no guidance on which to reach for. The result is
   inconsistent usage and maintenance debt.

The gap matters because:

- Maintenance cost of full overrides scales with role file size × number of projects.
- Skill composition (ADR 004) solves capability-level extension but not role-level
  customisation (identity text, default skills list, allowed tools, token budget).
- Without a coherent model, future contributors make ad-hoc decisions that fragment
  the extension story.

---

## Decision

Formalise a **three-tier extension model** with explicit rules for when each tier applies.

```
Tier 1  skills/*.skill.md           OPEN to extension
Tier 2  roles/<name>.md             CLOSED to modification (base identity)
Tier 3  .ai/roles/<name>.md         Project customisation (two sub-modes)
          a) Full override (no Extends:)   — complete replacement
          b) Extension   (**Extends:**)    — inherits base, overrides specific fields
Policy  .ai/workflow.md             Gates overrides and skill disabling (ADR 004)
```

### Tier 1 — Skills (open for extension)

Any project may drop a `.skill.md` file in `.ai/skills/`. The composition algorithm
(ADR 004) resolves project skills before framework skills, so a project skill with
the same name shadows the framework one. New skill names are additive.

**Extension point:** Add or replace skills. Never edit `skills/` directly.

### Tier 2 — Base Roles (closed to modification)

Files in `roles/` are the canonical identity definition for each role. They are
immutable from the project's perspective (treated as read-only even when the project
owns the repository). Modification requires a framework-level PR and version bump.

**Why closed?** The base role is the shared contract across every project using
ai-pack. Editing it in place breaks other projects. Skills (Tier 1) handle capability
extension without touching the role file.

### Tier 3a — Full Override (`.ai/roles/<name>.md`, no `Extends:` header)

The project supplies a complete role definition. The server uses this file in place of
the base role. No merging occurs. Appropriate when:

- The project's role identity diverges significantly from the base.
- Upstream base-role changes should NOT automatically apply.

**Cost:** The project is responsible for tracking upstream changes manually.

### Tier 3b — Role Extension (`.ai/roles/<name>.md` with `**Extends:** <name>`)

The project file declares `**Extends:** <base-role-name>`. The server applies an
**overlay merge**: start with the base role, then apply overrides from the project
file field-by-field. Fields absent in the project file are inherited unchanged.

**Mergeable fields:**

| Field | Merge strategy |
|---|---|
| `**Skills:**` | Project list replaces base list entirely |
| `**Tools:**` | Project list replaces base list entirely |
| `**Token Budget:**` | Project value replaces base value |
| Free-text sections (Overview, Responsibilities, etc.) | Project sections replace matching base sections; absent sections are inherited |
| `**Extends:**` | Consumed by merge; not passed to the agent |

**Inheritance does not chain.** Only one level of `Extends:` is supported. A project
override may not extend another project override.

### Policy — `.ai/workflow.md` (gate and skill overrides)

Gate overrides and skill disabling remain in `.ai/workflow.md` per ADR 004. This file
governs runtime policy; Tier 3 files govern static role identity. They are orthogonal
and both apply.

---

## Rationale

### Why overlay merge rather than deep inheritance chains?

Chains create distance between definition and effective behaviour. An agent's
configuration becomes the product of N files; debugging requires tracing every link.
A single extension level keeps the effective role reconstructible by reading two files
(base + override). Full overrides are always self-contained. The tracing problem is
bounded.

ADR 004 explicitly rejected multi-level inheritance for skill composition for the same
reason (diamond problem, opaque hierarchies). This ADR applies the same principle to
role files.

### Why field-level merge rather than section-level text merge?

Text merging (e.g., git-style) is fragile against whitespace and formatting changes.
Structured field merging on well-defined headers is deterministic and testable.
Sections without explicit field semantics use presence-based replacement: if the
project file contains a matching heading, that section replaces the base; otherwise
the base section is preserved.

### Why keep full override at all?

Some projects diverge enough that inheritance provides no value and introduces
confusion ("why doesn't this field work?"). A clear escape hatch — omit `Extends:`
and own the whole file — avoids forcing projects into a merge model that doesn't fit.

---

## Consequences

### Positive

- **Maintenance reduction.** A project changing one skill or token budget writes a
  5-line extension file rather than copying the full role.
- **Upstream-safe.** Extension files inherit upstream changes automatically (for
  unoverridden fields). Full overrides opt out explicitly.
- **Decision clarity.** The tier model answers "which mechanism should I use?" with
  rules rather than judgment.
- **Deterministic merging.** Field-level overlay is reproducible without git tooling.

### Negative / Trade-offs

- **One more concept.** Users must understand `Extends:` vs. full override vs. skill
  addition. Mitigated by user-facing docs (this ADR's companion deliverables).
- **No multi-level inheritance.** A project cannot extend another project's override.
  Enforced at load time with a clear error.
- **Section matching is heading-based.** Heading text must match exactly for section
  replacement to work. Cosmetic heading edits in the base role break existing
  extensions. Mitigated by stable heading conventions in base roles.

### Neutral

- Existing full-override files (`.ai/roles/<name>.md` without `Extends:`) continue
  to work unchanged. No migration required.
- `.ai/workflow.md` gate policy (ADR 004) is unaffected.

---

## Alternatives Considered

### A. YAML overlay sidecar (`.ai/roles/engineer.yaml`)

Separate YAML file defines field overrides; base `.md` remains the prose source.

**Why not chosen:** Two file formats per role. The `.md` format is already structured
(headers as fields); a YAML sidecar adds format complexity with no benefit over
in-file `Extends:` declaration.

### B. Deep inheritance chains

Allow extension files to themselves be extended (`A extends B extends C`).

**Why not chosen:** Bounded tracing (max two files) is a core goal. Chains defeat
this and reintroduce the diamond problem. One level is sufficient for all known
use cases; further flexibility can be added in a future ADR if needed.

### C. Merge only at the tools/skills/gates level (no text-section merge)

Extension file overrides only structured fields; prose sections are always inherited.

**Why not chosen:** Projects frequently need to change the instructions section
(different responsibilities, different step ordering) without writing a full override.
Section-level replacement is needed for real-world customisation.

---

## Related Documents

- [ADR 004 — Role/Skill OCP](004-role-skill-ocp.md)
- [docs/architecture/skill-schema.md](../architecture/skill-schema.md)
- [docs/architecture/skill-composition.md](../architecture/skill-composition.md)
- [docs/guides/extending-roles.md](../guides/extending-roles.md) *(companion guide)*
- [docs/guides/extending-skills.md](../guides/extending-skills.md) *(companion guide)*
