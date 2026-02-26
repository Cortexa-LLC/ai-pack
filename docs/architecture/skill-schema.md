# Skill File Schema

**Version:** 1.0.0
**Status:** Accepted (ADR 004)
**Last Updated:** 2026-02-26

---

## Overview

A skill is a discrete, reusable unit of agent capability. Skills are markdown files with
structured headers followed by a free-form prompt fragment. The filename convention is
`<name>.skill.md`.

---

## Directory Layout

```
project-root/
├── skills/                   # Framework-level skills (shared across projects)
│   ├── general.skill.md
│   ├── tdd.skill.md
│   ├── kg_reader.skill.md
│   ├── kg_writer.skill.md
│   ├── code_review.skill.md
│   └── arch_review.skill.md
├── roles/
│   └── engineer.md           # CLOSED — lists Skills, no gates
└── .ai/
    ├── skills/               # Project-level overrides and additions
    │   └── my_custom.skill.md
    └── workflow.md           # Gate overrides per role
```

Resolution order (first match wins):
1. `.ai/skills/<name>.skill.md` — project override
2. `.ai-pack/skills/<name>.skill.md` — submodule override (if mounted)
3. `skills/<name>.skill.md` — framework default

---

## Skill File Format

```markdown
# <Human Readable Name>
<!-- skills/<name>.skill.md -->

**Version:** 1.0
**InjectAt:** preamble | role_context | postamble
**Slot:** <integer>
**Tools:** <comma-separated tool names> | (none)
**Gates:** <comma-separated gate names> | (none)
**MaxExtraTokens:** <integer>
**Optional:** true | false

---

<prompt fragment — injected verbatim into system prompt at InjectAt position>
```

---

## Header Fields

### `**Version:**` *(required)*

Semantic version of the skill. Used for compatibility checks and diagnostics.

```
**Version:** 1.0
```

---

### `**InjectAt:**` *(required)*

Where in the assembled system prompt this skill's fragment is injected.

| Value | Position |
|-------|----------|
| `preamble` | Before the role body (e.g., universal constraints, agent policy) |
| `role_context` | After the role body (e.g., capability-specific instructions) |
| `postamble` | At the very end (e.g., output formatting, task completion reminder) |

Multiple skills at the same `InjectAt` position are ordered by their `**Slot:**` value.

---

### `**Slot:**` *(required)*

Integer ordering within the same `InjectAt` group. Lower values inject first.

```
**Slot:** 10     # inject early in this group
**Slot:** 90     # inject late in this group
```

Recommended convention:
- `10–19` — universal foundation skills (e.g., `general`)
- `20–49` — tooling/access skills (e.g., `kg_reader`, `kg_writer`)
- `50–79` — workflow enforcement skills (e.g., `tdd`, `code_review`)
- `80–99` — postamble / completion reminders

---

### `**Tools:**` *(optional)*

Comma-separated list of tool names this skill adds to the agent's allowed tool set.
The server merges these with the role's `**Tools:**` header (union, deduplication).

```
**Tools:** Read, Write, Edit, Bash, Grep, Glob
**Tools:** (none)
```

If omitted, the skill adds no new tools.

---

### `**Gates:**` *(optional)*

Comma-separated list of gate names this skill enforces when active.
The server merges these with any other active skill gates (union).

```
**Gates:** tdd-enforcement, code-quality-review
**Gates:** (none)
```

If omitted, the skill enforces no gates. Gate enforcement can be overridden
per-role in `.ai/workflow.md` (see [Composition Algorithm](skill-composition.md#gate-overrides)).

---

### `**MaxExtraTokens:**` *(optional)*

Maximum additional budget tokens this skill requests from the session budget.
Additive across all composed skills; the server caps at the model's hard limit.

```
**MaxExtraTokens:** 50000
**MaxExtraTokens:** 0
```

Defaults to `0` if omitted.

---

### `**Optional:**` *(optional)*

Whether this skill can be omitted by a project-level disable configuration.

```
**Optional:** true   # can be listed in .ai/workflow.md skills.disable
**Optional:** false  # always active when listed in a role's **Skills:**
```

Defaults to `true` if omitted. Skills marked `false` are treated as required
regardless of workflow configuration.

---

## Prompt Fragment

The text after the `---` separator is injected verbatim (with a single blank line
separator) into the assembled system prompt at the `InjectAt` position.

Guidelines:
- Write the fragment as if it were part of the system prompt already
- Do NOT include preamble text like "As a skill, you should..." 
- Use second-person ("You have access to...", "When writing tests...")
- Keep it focused — one skill, one concern
- Fragments from multiple skills at the same `InjectAt` are joined with `\n\n---\n\n`

---

## Example: `general` skill

```markdown
# General
<!-- skills/general.skill.md -->

**Version:** 1.0
**InjectAt:** preamble
**Slot:** 10
**Tools:** Read, Write, Edit, Bash, Grep, Glob
**Gates:** (none)
**MaxExtraTokens:** 0
**Optional:** false

---

You are a capable software engineering agent.
Complete the task in the working directory.
Verify your work before finishing.
```

---

## Example: `tdd` skill

```markdown
# Test-Driven Development
<!-- skills/tdd.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 50
**Tools:** (none)
**Gates:** tdd-enforcement
**MaxExtraTokens:** 0
**Optional:** true

---

## TDD Discipline

You follow strict Test-Driven Development:

1. Write a failing test first
2. Write the minimal code to make it pass
3. Refactor with tests green

Never skip the red phase. All new behaviour must be covered by tests before
the implementation is written. Run tests after every change.
```

---

## Example: `kg_reader` skill

```markdown
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

## Knowledge Graph Access

You have read access to the project knowledge graph via MCP tools:
- `search_nodes` — find entities by name, type, or content
- `open_nodes` — retrieve specific entities by name
- `read_graph` — read the full graph (use sparingly)

Consult the knowledge graph when you need architectural context, component
relationships, or design decision history before starting implementation work.
```

---

## Role File — Minimal Change

To adopt skills, a role file needs only two changes:

1. **Add `**Skills:**` header** listing the ordered skill names
2. **Remove `**Gates:**` header** (gates are now owned by skills)

```markdown
# Engineer Role

**Agent:** engineer
**Tier:** medium
**MaxTurns:** 450
**MaxBudgetTokens:** 3000000
**Skills:** general, kg_reader, kg_writer, tdd, code_review

---

(engineer personality and domain-specific instructions here)
```

The role body (below `---`) is unchanged. The role remains closed to modification.

---

## Related Documents

- ADR: [docs/adr/004-role-skill-ocp.md](../adr/004-role-skill-ocp.md)
- Composition: [docs/architecture/skill-composition.md](skill-composition.md)
