---
sidebar_position: 4
---

# Skills

AI-Pack ships three workflow skills in `plugin/skills/`. Each packages a multi-agent workflow: it decides which [agents](./agents.md) to spawn, in what order, and when to stop. Skills trigger automatically when your request matches, or explicitly via `/ai-pack:<skill>`.

## orchestrate

**What it does:** Decomposes engineering work and delegates it to the specialized subagents. It is the entry point for anything that takes more than one role: design-then-implement, investigate-then-fix, implement-then-review.

**What triggers it:** Multi-step engineering requests —

- "build the user authentication feature end to end"
- "fix the race condition in the job queue and add a regression test"
- "design and implement the new notification system"
- "investigate why requests are timing out then fix the root cause"

**What it looks like when running:** The session breaks your request into tasks, picks the right agent for each (spelunker or inspector to understand, architect to design, engineer to build, reviewer to check), and spawns them via the `Agent` tool with fully self-contained briefs. Independent tasks are spawned in parallel in a single message; dependent tasks wait for their inputs. Results flow back to the orchestrating session, which stitches them into the next brief or the final summary.

## pre-push

**What it does:** Runs a review-and-fix loop against the local commits that have not yet been pushed to origin, so problems are caught before they reach the remote.

**What triggers it:**

- "review my commits before I push"
- "pre-push check"
- "review and fix my local commits then tell me when it's safe to push"

**What it looks like when running:** A reviewer agent is spawned against the local diff. If it finds issues, an engineer agent is spawned to fix them, the fixes are amended into the commit, and the reviewer runs again. The loop halts when the reviewer approves or after 5 iterations, whichever comes first — you get a final verdict on whether it is safe to push.

## shepherd-pr

**What it does:** Drives an open GitHub PR to a green, approved, mergeable state by spawning the [pr-shepherd agent](./agents.md#pr-shepherd). Handles CI failures and reviewer threads until every check passes and the review verdict is APPROVED.

**What triggers it:**

- "shepherd PR #42 to merge-ready"
- "drive my PR to green"
- "fix the CI failures and reviewer comments on PR #15"

**What it looks like when running:** The shepherd runs a state machine over the `gh` CLI: **check PR state → fix reviewer threads → fix CI failures → push → reply to and resolve threads → schedule a wakeup → repeat**. Each pass does exactly one cycle; when CI is still running after a push, the shepherd schedules a wakeup instead of busy-waiting, and picks up where it left off when checks complete. It stops when the PR is approved with zero open conversations — or reports precisely what still blocks the merge.

## Skills vs. agents

Agents are roles; skills are workflows over those roles. If you want a single job done ("review this file"), spawning an agent directly is enough. If the request has phases, loops, or verdict-driven retries, a skill coordinates the agents for you.
