# Extending Roles

**Version:** 1.0.0
**Last Updated:** 2026-02-28
**Related:** [Extending Skills](extending-skills.md), [ADR 006](../adr/006-role-extension-ocp.md)

---

## Overview

The ai-pack role system uses a **three-tier model** to balance immutability of shared
definitions against project-level flexibility:

| Tier | Where | Purpose |
|---|---|---|
| 1 | `skills/*.skill.md` | Composable capability units — open for extension |
| 2 | `roles/<name>.md` | Base role identity — **closed to modification** |
| 3 | `.ai/roles/<name>.md` | Project customisation — full override or partial extension |

This guide covers **Tier 3**: customising a role for your project. For adding or
replacing skills, see [Extending Skills](extending-skills.md).

---

## Decision Tree: Which Mechanism to Use?

```
Do you want to add a new capability (TDD enforcement, KG access, etc.)?
  └─ YES → Add or create a skill (Tier 1). See extending-skills.md.

Do you need to customise the role itself (identity, responsibilities, token budget)?
  └─ YES → Use a .ai/roles/<name>.md file (Tier 3).
       │
       ├─ Small change (swap a skill, raise token budget, adjust one section)?
       │     └─ Use EXTENSION (add **Extends:** header). Inherits everything else.
       │
       └─ Major change (different identity, different workflow, isolate from upstream)?
             └─ Use FULL OVERRIDE (no **Extends:** header). Own the whole file.
```

---

## Tier 3a — Full Override

A file at `.ai/roles/<name>.md` **without** an `**Extends:**` header completely
replaces the base role. The server ignores `roles/<name>.md` entirely for this project.

**Use when:**
- The project role diverges significantly from the base.
- You explicitly do NOT want upstream base-role changes to affect this project.

**Example — completely custom orchestrator:**

`.ai/roles/orchestrator.md`

```markdown
# Orchestrator — Acme Corp

**Version:** 2.0.0
**Skills:** general, kg_reader

## Role Overview

Acme's Orchestrator is responsible for managing the Acme delivery pipeline only.
It does not handle generic feature workflows.

## Primary Responsibilities

1. Receive task from Acme Jira integration
2. Assign to Engineer via `agent update --claim`
3. Monitor SLA using internal tooling
4. Escalate to `#oncall` Slack channel on breach

...
```

**Cost:** You own the full file. Upstream changes to `roles/orchestrator.md`
(bug fixes, new policy sections) must be merged manually.

---

## Tier 3b — Extension (Recommended for Small Changes)

Add `**Extends:** <base-role-name>` as the first structured header in your project
file. The server performs an **overlay merge**: start with the base role, then apply
your overrides field by field. Absent fields are inherited unchanged.

**Syntax:**

```markdown
# Engineer — Acme Corp

**Extends:** engineer
**Skills:** general, kg_reader, bdd, code_review
**Token Budget:** 120000
```

That's the entire file. Three lines of structured fields are all you need to:
- Swap the skill list (replace `tdd` with `bdd`)
- Raise the token budget
- Inherit every other field (Tools, Responsibilities, etc.) from `roles/engineer.md`

---

## Merge Rules

### Structured fields

These headers are merged field by field:

| Field | Behaviour |
|---|---|
| `**Skills:**` | Your list replaces the base list entirely |
| `**Tools:**` | Your list replaces the base list entirely |
| `**Token Budget:**` | Your value replaces the base value |
| `**Extends:**` | Consumed by the merge; never passed to the agent |

### Prose sections

Sections identified by their `##` or `###` heading:

- **Present in your override** → your version replaces the base section.
- **Absent in your override** → base section is inherited unchanged.

Heading text must match exactly (case-sensitive).

### What is never merged

- `**Version:**` and `**Last Updated:**` from your file are passed through as-is
  (or omitted if absent; the base version is not inherited for these metadata fields).

---

## Worked Examples

### Example 1 — Swap one skill (TDD → BDD)

**Goal:** Replace `tdd` with `bdd` in the engineer role. Everything else stays the same.

**Step 1:** Find the base skills list.

```bash
grep '^\*\*Skills:\*\*' roles/engineer.md
```
Output: `**Skills:** general, kg_reader, tdd, code_review`

**Step 2:** Create the extension file.

`.ai/roles/engineer.md`

```markdown
# Engineer — Acme Corp

**Extends:** engineer
**Skills:** general, kg_reader, bdd, code_review
```

Done. The engineer now uses BDD instead of TDD. All other fields (Tools, Token Budget,
Responsibilities, etc.) are inherited from `roles/engineer.md`.

---

### Example 2 — Raise token budget and add custom responsibility section

`.ai/roles/architect.md`

```markdown
# Architect — Acme Corp

**Extends:** architect
**Token Budget:** 150000

## Primary Responsibilities

### 0. Load Acme Design System constraints

Before beginning any architecture work, read `docs/acme-design-system.md`.
All component choices must align with the Acme Design System.

### 1. Technical Feasibility Assessment

[… rest of responsibilities identical to base …]
```

**Effect:**
- Token budget raised to 150 000.
- "Primary Responsibilities" section replaced (your full section text applies).
- All other sections (Capabilities, Deliverables, ADR template, etc.) inherited from base.

---

### Example 3 — Add project-specific tools to the reviewer

`.ai/roles/reviewer.md`

```markdown
# Reviewer — Acme Corp

**Extends:** reviewer
**Tools:** [read, grep, glob, bash, webfetch, jira_fetch]
```

The reviewer now has `jira_fetch` in addition to (or instead of) the base tool list.
Skills, Token Budget, and all prose sections are inherited unchanged.

---

### Example 4 — Disable a skill via `.ai/workflow.md` (not an extension file)

Skills can be disabled at runtime without touching the role file at all. This is the
right approach when you want the base role but need to turn off a gate globally.

`.ai/workflow.md` (excerpt):

```markdown
## Skill Overrides

**Disabled:** tdd-enforcement
```

See [ADR 004](../adr/004-role-skill-ocp.md) and [Extending Skills](extending-skills.md)
for full workflow gate documentation.

---

## Pitfalls to Avoid

### ❌ Extending a project override

```markdown
# Bad — extending another .ai/roles/ file is not supported

**Extends:** .ai/roles/engineer   ← INVALID
```

`Extends:` must reference a base role name (a file in `roles/`). One level only.

### ❌ Partial skills list accidentally dropping required skills

```markdown
**Skills:** bdd
```

This replaces the entire base skills list with only `bdd`. If the base had `general`,
`kg_reader`, and `code_review`, they are all dropped. Copy the full base list and
modify it:

```markdown
**Skills:** general, kg_reader, bdd, code_review
```

### ❌ Heading mismatch breaking section inheritance

```markdown
# Base role has:            # Your override has:
## Primary Responsibilities ## primary responsibilities  ← lowercase, won't match
```

Use the exact heading text from the base role file to trigger section replacement.
Mismatched headings create a new section; both versions will be present.

---

## Quick Reference

```
.ai/roles/<name>.md   (no Extends:)   → full override, own everything
.ai/roles/<name>.md   (+ Extends:)    → overlay merge, inherit unchanged fields
.ai/skills/<name>.skill.md            → add/replace a skill
.ai/workflow.md                        → disable gates, skill runtime policy
```

---

## Related Documents

- [Extending Skills](extending-skills.md) — add or replace skill capability units
- [ADR 006](../adr/006-role-extension-ocp.md) — design rationale for this model
- [ADR 004](../adr/004-role-skill-ocp.md) — skill composition and gate policy
- [docs/architecture/skill-schema.md](../architecture/skill-schema.md) — skill file format
