# Extending Skills

**Version:** 1.0.0
**Last Updated:** 2026-02-28
**Related:** [Extending Roles](extending-roles.md), [ADR 004](../adr/004-role-skill-ocp.md), [Skill Schema](../architecture/skill-schema.md)

---

## Overview

A **skill** is a discrete, reusable capability unit: a structured `.skill.md` file that
declares tools, gates, prompt fragments, and token cost. Roles compose skills; skills
themselves are independent of any role.

There are two ways to extend at the skill tier:

| Goal | Mechanism |
|---|---|
| Add a brand-new capability | Create `.ai/skills/<name>.skill.md` |
| Replace a framework skill | Create `.ai/skills/<name>.skill.md` with the same name |

Project skills (`.ai/skills/`) are resolved **before** framework skills (`skills/`).
A project skill with the same name as a framework skill shadows it entirely.

---

## Skill File Format

```markdown
# <Human-readable title>
<!-- .ai/skills/<name>.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 60
**Tools:** (none)
**Gates:** <gate-name>
**MaxExtraTokens:** 500
**Optional:** true

---

<Prompt fragment — plain prose injected into the agent's system prompt>
```

### Header fields

| Field | Required | Description |
|---|---|---|
| `**Version:**` | Yes | Semver string |
| `**InjectAt:**` | Yes | Where the prompt fragment lands: `preamble`, `role_context`, or `postamble` |
| `**Slot:**` | Yes | Integer; controls order within the same `InjectAt` position. Lower = earlier. |
| `**Tools:**` | Yes | Comma-separated tool names this skill enables, or `(none)` |
| `**Gates:**` | Yes | Gate name this skill enforces, or `(none)` |
| `**MaxExtraTokens:**` | Yes | Maximum additional tokens this skill may consume |
| `**Optional:**` | Yes | `true` = can be disabled via `.ai/workflow.md`; `false` = always active |

See [docs/architecture/skill-schema.md](../architecture/skill-schema.md) for full schema
and the list of reserved slot numbers.

---

## Worked Examples

### Example 1 — New skill: BDD enforcement

**Goal:** Add a BDD discipline skill (analogous to the framework `tdd` skill) so the
engineer uses Gherkin scenarios instead of unit tests first.

`.ai/skills/bdd.skill.md`

```markdown
# Behaviour-Driven Development
<!-- .ai/skills/bdd.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 50
**Tools:** (none)
**Gates:** bdd-enforcement
**MaxExtraTokens:** 0
**Optional:** true

---

## BDD Discipline

You follow strict Behaviour-Driven Development:

1. **Scenario first** — write a failing Gherkin scenario that captures the intended
   behaviour before writing any implementation code.
2. **Step definitions** — implement the minimal step code to make the scenario pass.
3. **Refactor** — clean up with all scenarios green.

Never skip the scenario-first step. All new behaviour must be expressed in Gherkin
before the code that satisfies it exists.
```

**Wire it up** — reference the skill from your role extension:

`.ai/roles/engineer.md`

```markdown
# Engineer — Acme Corp

**Extends:** engineer
**Skills:** general, kg_reader, bdd, code_review
```

The `tdd` skill (and its `tdd-enforcement` gate) is dropped; `bdd` and its gate take
its place.

---

### Example 2 — Shadow a framework skill to lower its token cost

The framework `code_review` skill uses 800 extra tokens. Your project wants a shorter
review prompt.

`.ai/skills/code_review.skill.md`

```markdown
# Code Review (Acme)
<!-- .ai/skills/code_review.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 70
**Tools:** (none)
**Gates:** (none)
**MaxExtraTokens:** 200
**Optional:** true

---

## Code Review Checklist

- Does the change follow the Acme style guide?
- Are all new public functions documented?
- Are error paths tested?
```

Because `.ai/skills/code_review.skill.md` exists, the composition algorithm uses it
instead of `skills/code_review.skill.md`. No changes to any role file are required.

---

### Example 3 — New skill that enables extra tools

**Goal:** Give the reviewer the ability to call an internal `jira_fetch` tool.

`.ai/skills/jira_review.skill.md`

```markdown
# Jira Context for Review
<!-- .ai/skills/jira_review.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 65
**Tools:** jira_fetch
**Gates:** (none)
**MaxExtraTokens:** 300
**Optional:** true

---

## Jira Integration

When reviewing a change, fetch the linked Jira ticket with `jira_fetch <ticket-id>`
and verify that the implementation matches the acceptance criteria on the ticket.
Reference the ticket ID in your review summary.
```

`.ai/roles/reviewer.md`

```markdown
# Reviewer — Acme Corp

**Extends:** reviewer
**Skills:** general, kg_reader, code_review, jira_review
```

The reviewer now has access to `jira_fetch` and will be instructed to use it.

---

### Example 4 — Disable a skill at runtime (no file required)

You want the base engineer role but without TDD enforcement for a specific project phase.

`.ai/workflow.md` (excerpt):

```markdown
## Skill Overrides

**Disabled:** tdd-enforcement
```

Skills marked `**Optional:** true` respect this gate. The skill is still listed in the
role's `**Skills:**`, but its gate and prompt fragment are suppressed. This does not
require a `.ai/roles/` file or a `.ai/skills/` file.

---

## Resolution Order

When the server resolves a skill named `foo`:

1. Check `.ai/skills/foo.skill.md` → use if present (project skill).
2. Check `skills/foo.skill.md` → use if present (framework skill).
3. Fail with an error if neither exists.

Project skills always win. There is no merging — the entire file is used.

---

## Slot Numbering Conventions

| Range | Purpose |
|---|---|
| 0–19 | Preamble (agent identity, core policy) |
| 20–49 | Reserved for framework base skills |
| 50–69 | Development discipline skills (tdd, bdd, …) |
| 70–89 | Review and quality skills |
| 90–99 | Postamble (closing instructions) |

Choose a slot that puts your skill in the right position relative to existing skills.
Two skills at the same slot number and inject position are ordered alphabetically by
name (deterministic but arbitrary — avoid collisions by using distinct slot values).

---

## Pitfalls to Avoid

### ❌ Forgetting to add the skill to a role's Skills: list

A skill file in `.ai/skills/` has no effect unless a role lists it in `**Skills:**`.
Creating the file is step 1; wiring it into a role (via a Tier 3 extension or the
base role) is step 2.

### ❌ Using slot 0–49 for project skills

These ranges are reserved for framework skills. A project skill in slot 30 will
interleave with framework preamble content in potentially unexpected ways. Start
project skills at slot 50 or higher.

### ❌ Marking a security gate skill as Optional: true

If a skill enforces a compliance or security gate it must be `**Optional:** false` so
it cannot be suppressed via `.ai/workflow.md`. Optional should only be `true` for
capability skills that users legitimately need to disable.

### ❌ Setting MaxExtraTokens too high for Optional skills

Optional skills may be disabled; their token cost should be small. If a skill needs
> 1 000 extra tokens reconsider whether it should be split into a required base section
and an optional extension section.

---

## Quick Reference

```
.ai/skills/<name>.skill.md          → new skill or shadow framework skill
.ai/roles/<name>.md (Extends:)      → wire skill into a role extension
.ai/workflow.md (Disabled:)         → suppress Optional gates at runtime
```

---

## Related Documents

- [Extending Roles](extending-roles.md) — customise role identity and defaults
- [ADR 004](../adr/004-role-skill-ocp.md) — skill composition algorithm and gate policy
- [ADR 006](../adr/006-role-extension-ocp.md) — three-tier extension model
- [docs/architecture/skill-schema.md](../architecture/skill-schema.md) — full skill file schema
- [docs/architecture/skill-composition.md](../architecture/skill-composition.md) — composition algorithm
