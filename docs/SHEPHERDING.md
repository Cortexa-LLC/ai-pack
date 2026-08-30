# Shepherding a PR: three variants, and which to use

ai-pack ships three ways to drive a pull request to merge-ready. They are not
redundant — each solves a different problem, and two of them travel while one does
not. This document says which to reach for, and records the distribution constraint
that shaped the split (issue #29).

---

## Quick answer

| You want to… | Use | Ships to consumers? |
|---|---|---|
| Shepherd one PR from your own session, without blocking it | **`ai-pack:shepherd-pr` skill** | Yes |
| Shepherd several PRs at once, or mix shepherding with other delegated work | **`ai-pack:pr-shepherd` agent** | Yes |
| Guarantee the loop runs to completion, in this repo | **`shepherd-pr` workflow** | **No — repo-local** |

Default to the skill. Reach for the agent when shepherding is one workstream among
several. Reach for the workflow when you want the round budget enforced by code
rather than by model discipline — and only inside this repository.

---

## The three variants

### Skill — `ai-pack:shepherd-pr` (`plugin/skills/shepherd-pr/SKILL.md`)

A **non-blocking state machine**. One pass per invocation: check state → fix findings
→ push → reply → schedule a wakeup. `ScheduleWakeup` re-fires the skill instead of
holding the main loop, so your session stays responsive while CI runs.

This is the preferred entry point for a single PR. It is also the variant that
carries the routing rules in the most detail — Routes A through G, the convergence
guard, and the escalation table.

**Trade-off:** completion depends on wakeups firing and on the model following the
route table each pass. Nothing in the runtime forces the next round to happen.

### Agent — `ai-pack:pr-shepherd` (`plugin/agents/pr-shepherd.md`)

A **delegated worker**, spawned via the `Agent` tool. It runs the same state machine
synchronously inside its own context, so the orchestrator can spawn several in
parallel — one per PR — or interleave shepherding with unrelated delegated work.

**Trade-off:** it must run to completion in-turn, which is why its definition carries
the explicit run-to-completion rule and the phantom-watcher anti-pattern. A model that
ends its turn believing a background watcher will resume it simply stops.

### Workflow — `shepherd-pr` (`.claude/workflows/shepherd-pr.js`)

A **deterministic loop in script code**. The round counter, the minimum-rounds policy,
the non-convergence check, and every exit condition live in JavaScript; a subagent is
invoked once per round to fix findings and return structured state. The script cannot
forget to loop, cannot stop at round 3, and cannot invent a watcher.

```
Workflow({ name: "shepherd-pr", args: { pr: 46, branch: "fix/…" } })
Workflow({ name: "shepherd-pr", args: { pr: 46, branch: "fix/…", repo: "owner/name" } })
```

`repo` is optional — omitted, the round agent resolves it from the checkout with
`gh repo view --json nameWithOwner`. Pass it explicitly when driving a PR in a
different repository than the one you are standing in.

**If you copy this file, keep it portable.** It must contain no absolute filesystem
paths and no hardcoded `owner/name`: a copied-and-unedited script with a baked-in slug
sends every review fetch to the original repository, and an absolute path tells each
round agent it is somewhere it is not.

Exits, as returned in `outcome`: `merge-ready` (clean verdict) · `blocked-on-owner` ·
`stuck` (no progress for two rounds, past `MIN_ROUNDS`) · `cap-reached` (hard
backstop) · `error` (two consecutive round agents failed). `MIN_ROUNDS` floors only
the `stuck` exit — a clean verdict or an owner blocker still returns at round 1.

**Trade-off:** it is repo-local. See below.

---

## Why the workflow does not ship

**Plugins cannot distribute workflows.** Verified empirically, not inferred:

1. A `workflows/` directory was added to `plugin/`, installed with
   `make update-plugin`, and confirmed present in the installed copy under
   `~/.claude/plugins/cache/ai-pack/ai-pack/<version>/workflows/`.
2. `claude plugin validate` passed without acknowledging it either way.
3. Invoking it by name failed:
   `Workflow "plugin-workflow-probe" not found. Available: deep-research, shepherd-pr`
   — `deep-research` is built in, `shepherd-pr` resolves from `.claude/workflows/`.

So the file ships but never registers. Corroborating: across the 207 plugins in the
marketplace catalog the only component types that appear are `agents`, `commands`,
`hooks`, `lspServers`, `mcpServers`, and `skills`. There is no `workflows` component,
and no installed plugin ships such a directory.

**Consequence.** Workflow-backed orchestration is repo-local by necessity. A consumer
project that wants the deterministic loop must copy
`.claude/workflows/shepherd-pr.js` into its own `.claude/workflows/`. The agent and
skill variants are the ones that travel with the plugin, which is why both are kept
rather than collapsed into the workflow.

Re-test this if Claude Code adds a workflow component type; the probe above takes
about a minute and the finding is worth exactly one command to re-confirm.

---

## What the workflow variant has and has not proven

Validated on PR #46 (run `wf_34283f6c-f2a`, 1 agent, ~8.5 min): drove a
`CHANGES_REQUESTED` head with 1 Critical, 1 Major, and 2 Minors to `APPROVED` with all
checks green and zero unresolved threads, in one round. It blocked synchronously on
CI rather than declaring a phantom watcher, verified its fix against the live check
payload rather than by inspection, and swept a flagged defect into a second file the
reviewer had not cited.

**Not yet proven:** it exited at round 1, so `MIN_ROUNDS`, the non-convergence check,
and the hard cap are all unexercised, and no `stuck` or `blocked-on-owner` outcome has
been observed. The claim that the loop cannot be abandoned is sound by construction,
but a PR needing three or more rounds is what would demonstrate the budget machinery.

---

## Keeping the variants honest

`plugin/skills/shepherd-pr/SKILL.md` and `plugin/agents/pr-shepherd.md` are **parallel
implementations of one state machine**. Every behavioral change to one must be applied
to the other and verified per item — never asserted. This has gone wrong twice, both
times as a CHANGELOG claiming parity that did not hold:

- Route G's terminal-verdict exit landed in the skill but not the agent, so the agent
  kept looping on a concluded review until its wait budget escalated it as "stuck".
- The review-gate exclusion landed in the skill's `CI_FAILING` but not the agent's
  `ALL_OK`, which made Route G's own guard unreachable for `CHANGES_REQUESTED` — the
  one verdict it existed to handle.

Where a rule enumerates states, prefer the **complement** over the list. Matching
"not `APPROVED` and not `PENDING`" cannot drift out of sync the way
"`COMMENTED` or `CHANGES_REQUESTED`" did when `DISMISSED` appeared.
