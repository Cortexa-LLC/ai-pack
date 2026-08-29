---
name: prd
description: >
  Run a product discovery interview with the product owner, then delegate PRD drafting
  to the product-manager agent. The interview happens in the main session (subagents
  cannot ask the user questions); the finished PRD lands in docs/product/. Use when the
  user wants requirements defined interactively rather than from an existing brief.
  Triggers on: "create a PRD", "product requirements for X", "help me spec this feature",
  "interview me about requirements", "PRD discovery".
  <example>create a PRD for the notification system</example>
  <example>help me spec this feature</example>
  <example>interview me about the requirements for billing export</example>
  <example>product requirements for the mobile onboarding flow</example>
  <example>run PRD discovery for the reporting dashboard</example>
---

# PRD Discovery

Three phases: **interview the product owner → delegate drafting → review loop**.

The division of labor is structural, not stylistic: spawned subagents **cannot** converse
with the user — only the main session can use `AskUserQuestion`. So the interview runs
here, and only the drafting is delegated to the `ai-pack:product-manager` agent.

---

## Phase 1 — Interview

**KG first — don't re-ask what's already decided.** Before the first question round:

```bash
kg__search_knowledge({query: "<feature or product area> decision"})
```

Recorded scope decisions, rejected features, and fixed metrics are settled ground —
open the interview by *confirming* them ("last time X was ruled out because Y — still
true?") rather than re-litigating them as open questions. Interview time goes to what
the KG doesn't know.

**KG availability:** If the `kg__*` tools are not in your tool list, or the first KG
call fails with a server/connection error, the knowledge graph is not installed —
skip every KG step silently (KG-first queries *and* KG checkpointing), rely on file
exploration, and do not mention the absence in your report unless the task is *about*
the KG. Never retry, never attempt a bash `kg` fallback, never treat missing KG as a
blocker or error.

Use the **AskUserQuestion** tool in rounds: **2–4 questions per round**, each with concrete
multiple-choice options plus an "Other" escape so the owner is never boxed in.

### Coverage checklist

Work through these areas across rounds (not one round per area — group what fits):

1. **Problem** — what hurts, who has it, how badly, current workaround
2. **Target users / personas** — who exactly, and which persona is primary
3. **Jobs-to-be-done** — what users are trying to accomplish, not what UI they want
4. **Scope boundaries** — what's in v1, and **explicitly elicit non-goals** ("which of
   these are we deliberately NOT doing?") — an unasked non-goal becomes scope creep
5. **Constraints** — technical, timeline, compliance, budget
6. **Success metrics** — push for measurable: "how would we know this worked?" and then
   "what number, measured how?"
7. **Priorities / sequencing** — what must ship first, what can wait
8. **Risks and failure modes** — what would make this fail or not be adopted

### Be adaptive, not bureaucratic

- **Skip what's answered.** If the user's initial ask already covers the problem and
  users, don't re-ask — start with scope and metrics.
- **Follow the thread.** When an answer reveals something load-bearing (a compliance
  requirement, a second user population, a hard deadline), chase it in the next round
  even if it's off-checklist.
- **Stop when marginal rounds stop changing the PRD.** Typically 2–4 rounds. This is an
  interview, not an interrogation — coverage of every checklist item is not the goal;
  a defensible PRD is. Gaps the owner can't answer yet become Open Questions.

### The WHAT/WHY vs HOW boundary

If the owner starts specifying architecture ("use websockets", "put it in the billing
service"), don't argue and don't embed design in the PRD: capture it verbatim under
**Constraints** and note it will be flagged for the architect. The PM defines WHAT and
WHY; the architect defines HOW.

---

## Phase 2 — Synthesis and Delegation

Compile everything into a **fully self-contained drafting brief** — the agent shares no
memory with this session:

- The full interview transcript (questions asked, answers given, including "Other" text)
- Any briefs, notes, or documents the user provided
- The feature slug to use for the filename
- Explicit gaps: questions the owner couldn't answer, marked as material for the
  PRD's Open Questions section

Spawn the drafter — **`run_in_background: false`**, the user is waiting on this result:

```text
Agent({
  subagent_type: "ai-pack:product-manager",
  run_in_background: false,
  description: "Draft PRD for <feature>",
  prompt: "All context provided. Draft a PRD to docs/product/prd-<slug>.md
           (use templates/product/prd.md if present).

           ## Discovery transcript
           <full Q&A transcript>

           ## Supplied materials
           <briefs / notes, or 'none'>

           ## Known gaps (put in Open Questions, do not invent answers)
           <list>"
})
```

---

## Phase 3 — Review Loop

1. **Present the draft** to the owner: the PRD's summary (problem, scope headline,
   epics, metrics) plus its full Open Questions list. Link the file path.
2. **One AskUserQuestion round** for corrections: wrong scope? missing story? metric
   off? answers to any open questions?
3. **Apply corrections:**
   - Small edits (wording, a metric value, an added acceptance criterion) — edit the
     PRD file inline yourself.
   - Structural changes (new epic, re-scoped feature, changed problem statement) —
     spawn a second `ai-pack:product-manager` pass with the original transcript plus
     the correction round.
4. Repeat the round once more only if the corrections were structural; otherwise done.

**No placeholders may survive.** Every section of the final PRD either has real content
or its gap sits explicitly under "Open Questions — needs product owner input". Bracketed
template text left in a delivered PRD is a defect.

---

## Done

Report to the user: the PRD path (`docs/product/prd-<slug>.md`), a two-line summary,
remaining open questions, and — if any constraints were technical — a note that the
next step is `ai-pack:architect` for design.
