# Cold Start Protocol for AI-Pack Submodule Adoption

**Status:** Draft
**Date:** 2026-02-19
**Author:** Captured from live failure analysis on ai-pack itself

---

## Why This Document Exists

On 2026-02-19, working in the ai-pack project's own codebase, we hit every cold start failure
mode back-to-back:

1. **Engineer agent burned 600 turns (~$40) doing discovery** on a task that should have taken
   20 turns — because the task brief was vague and didn't include the code to write.
2. **Architect agent destroyed `roles/engineer.md`** (1,514 lines → 21 lines) and closed the
   task as "complete" — because there was no task packet and no verifiable acceptance criteria.
3. **PM agent produced a skeleton PRD** with placeholder dates and no concrete decisions —
   because the acceptance criterion was prose ("document exists") not verifiable output.
4. **Spelunker hit a 1M token budget** at turn 65 still reading files — because the brief
   didn't include the specific code it needed to investigate.

These failures happened in the project that *built* the framework. A new adopting project
will have all the same gaps and none of the institutional knowledge to recover.

---

## The Cold Start Problem

When a project adds `.ai-pack` as a git submodule, agents start with **zero project context**:

- No knowledge of the codebase structure, key files, or conventions
- No knowledge of project-specific acceptance patterns
- No established task packet infrastructure
- No codebase map to guide orchestrators writing task briefs

The framework assumes this context exists. When it doesn't, every failure mode above becomes
the default outcome.

---

## Gap Analysis: What the Current Templates Are Missing

### Gap 1: No task packet infrastructure requirement before first agent spawn

**Current state:** `templates/CLAUDE.md` says "create a task packet before non-trivial tasks"
but gives no checklist for what must exist *before the first task ever runs*.

**Failure evidence:** Our architect agent was spawned with only a task (`agent create`) —
no `.ai/tasks/` directory, no `task.md`, no `task.md`. The engineer role's mandatory
pre-implementation check (section 0) had nothing to find. The agent skipped straight to
writing files with no plan.

**Required fix:** The adoption checklist must require `.ai/tasks/` infrastructure to exist
(or the first task to set it up) before any engineering agent is spawned.

### Gap 2: Acceptance criteria are prose, not verifiable commands

**Current state:** `agent create --acceptance "..."` stores prose that is never machine-verified.
Agents mark tasks closed without running the acceptance check.

**Failure evidence:** Architect task acceptance was: *"Proposed changes to engineer.md
documented. roles/shared/orchestrator-engineer-handoff.md created."* Both files existed
(one was a 21-line overwrite, one was a 29-line stub). Task marked complete. No gate caught
the regression.

xasm++ tasks that succeed use acceptance criteria like:
```
CRITICAL: ALL 704 tests must pass. NO REGRESSIONS ANYWHERE.
```
which agents verify by running `./build/xasm++ --test` and checking exit code.

**Required fix:** Acceptance criteria must be expressed as shell commands with expected
output or exit codes. Template must enforce this format.

### Gap 3: No `result.md` enforcement when task packets are missing

**Current state:** The engineer role requires a `result.md` sign-off, but only checks
`.ai/tasks/*/result.md`. If no task packet exists, this gate silently doesn't fire.

**Failure evidence:** None of the three agent tasks today produced a `result.md`
because none had task packets. All were marked "complete" anyway.

**Required fix:** Either (a) require task packets for all agent-spawned work, or (b) add
a fallback acceptance check to the agent loop that doesn't depend on task packet existence.

### Gap 4: No "Agent Orientation" section in CLAUDE.md template

**Current state:** `templates/CLAUDE.md` covers task packets and the orchestrator role, but
has no section for orienting agents to the project's specific codebase.

**Failure evidence:** Agents reading `CLAUDE.md` on a new project learn the framework rules
but nothing about *this project* — its key files, its build commands, its conventions, its
directory structure.

**Required fix:** A mandatory "Agent Orientation" section in CLAUDE.md (see below).

### Gap 5: No "first write" deadline enforced in agent loop

**Current state:** Agents can run indefinitely in read-only mode. The inactive turn counter
only fires when tool signatures repeat — reading 600 different files looks like "progress".

**Failure evidence:** Engineer agent ai-pack-8x0 spent all 600 turns reading. The spelunker
spent 65 turns at 1M tokens still reading. No gate fired.

**Required fix:** `MaxTurnsBeforeFirstWrite` field in role config (to be implemented).

---

## The Codebase Map Decision

**Question:** Should ai-pack auto-generate a codebase map on adoption?

**Decision: No auto-generation. Human-curated "Agent Orientation" section in CLAUDE.md.**

**Rationale:**

| Approach | Problem |
|----------|---------|
| Auto-generated at adoption | Stale within weeks. Agents trust stale maps more than no maps — actively harmful. |
| Auto-regenerated on commit | Expensive, noisy. Every commit triggers a spelunker survey. |
| Auto-regenerated on demand | Better, but who decides when? Still goes stale silently. |
| Human-curated in CLAUDE.md | Maintained alongside code by developers who know what changed. Intentional. |

The right answer is a **structured "Agent Orientation" section** in `CLAUDE.md` that
project teams fill in during adoption and update when significant structure changes.
This is the same model as good README hygiene — humans maintain it because they know
what matters.

A git hook that detects structural changes (new top-level directories, renamed key files)
and prints a warning — *"Agent orientation in CLAUDE.md may be stale — consider updating"*
— is more useful than auto-regeneration.

---

## Required: Agent Orientation Section in CLAUDE.md

Every adopting project's `CLAUDE.md` must include this section, filled in at adoption:

```markdown
## Agent Orientation

### What this project does
[One paragraph. What does this codebase build/do?]

### Key directories
| Path | Purpose |
|------|---------|
| `src/` | [what's here] |
| `tests/` | [what's here] |
| [etc.] | |

### How to build
```bash
[exact build command]  # expected: [what success looks like]
```

### How to run tests
```bash
[exact test command]  # expected: [pass/fail signal]
```

### Key files for agents
| File | Why agents need to know about it |
|------|----------------------------------|
| [path] | [what it does] |

### Conventions
- [naming conventions, patterns, things agents should follow]

### What agents must NOT do
- [destructive operations, files to never touch, etc.]
```

This section is **the first thing an orchestrator reads** before writing any task brief.
With it, the orchestrator can write pre-cooked briefs with exact file paths on day one.
Without it, every engineer agent rediscovers the codebase from scratch.

---

## Adoption Checklist

This checklist defines "project is ready for agents." It must be completed before
spawning any engineering agent.

### Step 1: Submodule and template setup
- [ ] `.ai-pack` submodule added at project root
- [ ] `CLAUDE.md` copied from `templates/CLAUDE.md` and customized
- [ ] **Agent Orientation section filled in** (build commands, key dirs, key files)
- [ ] `.claude/rules/` copied from `templates/.claude/rules/`

### Step 2: Task infrastructure
- [ ] `.ai/` directory exists at project root
- [ ] First task packet created: `mkdir -p .ai/tasks/$(date +%Y-%m-%d)_orientation`
- [ ] `bd init` run to configure Beads for this project

### Step 3: Verify before first agent spawn
Run this check before spawning any agent:
```bash
# Must all pass
test -f CLAUDE.md && grep -q "Agent Orientation" CLAUDE.md || echo "FAIL: No Agent Orientation"
test -d .ai/tasks || echo "FAIL: No task packet infrastructure"
test -d .ai-pack || echo "FAIL: No ai-pack submodule"
```

### Step 4: Acceptance criteria format
Every task created with `agent create` must have `--acceptance` expressed as a shell command:
```bash
# Good — verifiable
--acceptance "go build ./... exits 0; go test ./... exits 0"

# Bad — unverifiable prose
--acceptance "files are created and documented"
```

---

## Orchestrator Responsibilities at Cold Start

The orchestrator (human or agent) is responsible for ensuring the Agent Orientation
section is read before writing any task brief. The brief must include:

1. **Exact file paths** — not "the streaming file" but
   `a2a-agent/internal/streaming/openai_adapter.go`
2. **Exact changes** — not "fix multi-turn tool use" but the specific structs, methods,
   and code patterns to add/change
3. **The signal `All context provided`** — triggers execution mode bypass in engineer role,
   skipping discovery sections
4. **Acceptance criteria as commands** — `go build ./... && go test ./internal/streaming/...`

Without these four elements, the task brief is incomplete and should not be submitted to
an engineering agent.

---

## What Success Looks Like

A project with a complete cold start setup should achieve:

- First agent task runs in <50 turns (not 600)
- No agent destroys files (acceptance criteria catches regressions before close)
- Orchestrators write precise briefs on day one (Agent Orientation provides the map)
- Token cost per engineering task <$2 (not $40)

---

## Related

- `templates/CLAUDE.md` — template to update with Agent Orientation section
- `roles/engineer.md` — execution mode bypass (to be added)
- `roles/shared/orchestrator-engineer-handoff.md` — handoff protocol (to be added properly)
- `docs/adr/002-sse-task-stream-heartbeat.md` — example of properly scoped ADR
- `docs/architecture/a2a-agent.md` — example of Agent Orientation content for this project
