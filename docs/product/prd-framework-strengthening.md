# PRD: AI-Pack Framework Strengthening

**Status:** Accepted — product-owner decisions applied (see Decisions)
**Date:** 2026-08-23

## Problem Statement

- **Problem:** AI-Pack's post-pivot proposition — a coordinated agent team with persistent memory, at zero metered cost — rests on claims the framework cannot currently verify or protect. Investigation found four concrete weaknesses:
  1. **Unverifiable memory quality.** The knowledge graph is well populated (8,299 entities, 8,953 relations, 1,776 observations at time of investigation) and every role reads it first — but it contains obsolete server-era guidance (e.g. records directing agents to use the retired agent-mcp task DB and `agent` CLI) that KG-first agents will faithfully retrieve, and observations carry zero-value timestamps (`0001-01-01`), so staleness cannot even be assessed. No pruning, curation, or health tooling exists. The core differentiator can silently become a liability.
  2. **Unguarded product surface.** The product *is* prompt files. A prompt edit ships if CI's structure/frontmatter lint passes; no regression test exercises agent behavior. A planted-defect validation harness for the reviewer's audit mode was specified but never built.
  3. **Release incoherence.** `plugin/.claude-plugin/plugin.json` says version 2.2.0 while `VERSION` says 3.0.0; no CHANGELOG exists; docs disagree with each other (intro.md says six agents / three skills, README says seven / four). Install is clone-the-repo + local marketplace + submodule + Python script.
  4. **Blind delegation.** The spawn-logging hook shipped so recently its log is a zero-byte file. Nothing reads it. Delegation adherence and parallelism — the behaviors the orchestrate skill exists to produce — are unmeasured, and the deliberately deferred enforcement decision is waiting on exactly this evidence.
- **Current state:** The user trusts that agents behave as their definitions say, hand-checks the KG when suspicious, hand-maintains versions, and reads raw JSONL if delegation questions arise.
- **Opportunity:** The ecosystem is crowded with large agent collections and memory MCPs, but almost none integrate roles + workflows + persistent memory as one disciplined system. Strengthening verifiability of that integration is what keeps "why ai-pack over a 112-agent grab-bag" answerable.

## Target Users

- **Solo senior engineer on Claude Code with a Max subscription (primary, current):** runs multi-agent workflows daily across several consumer projects; wants delegation that provably works, memory that can be trusted, and zero metered API cost.
- **Small engineering teams (secondary, committed):** multi-developer conventions — shared knowledge graph, per-user subscription tokens, team-level delegation visibility — are in scope as Epic 6 (owner decision 2026-08-23).

## Goals and Success Metrics

- **Goal:** Agents never act on retired-era guidance — **Metric:** 0 obsolete server-era directives in top-10 results for a curated benchmark query set (e.g. "task tracking", "orchestrate", "agent spawn") run against the ai-pack project KG; measured by a repeatable check.
- **Goal:** Memory staleness is assessable — **Metric:** 100% of new observations carry a real timestamp; a health report distinguishes fresh vs. legacy knowledge.
- **Goal:** Prompt regressions are caught before ship — **Metric:** every PR touching `plugin/agents/` or `plugin/skills/` runs a behavioral check gate; reviewer audit harness detects ≥80% of a fixed planted-defect set.
- **Goal:** One coherent version story — **Metric:** version string identical across plugin.json, VERSION, and docs; CI fails on mismatch; every release has a CHANGELOG entry.
- **Goal:** Delegation is observable — **Metric:** the user can answer "delegation rate, parallelism rate, role mix, last 30 days" with a single documented command; 100% of Agent spawns in hook-enabled sessions are captured.

## Scope

### In Scope
- KG health: curation of stale knowledge, staleness signals, health reporting
- Behavioral quality gate for agent/skill definition changes, including the specified reviewer planted-defect harness
- Release hygiene: single version source of truth, changelog, docs-consistency checking, repeatable release steps
- A minimal read layer over the existing spawn log (report, not GUI)
- Priority-ordered groundwork only; everything funded by the Max subscription

### Non-Goals (explicitly out of scope)
- **Server revival** (API-driven agent server, GUI dashboard, GraphQL/SSE) — deprecated 2026-08-22 as cost-prohibitive; superseded by Claude Code
- **Multi-provider model routing / performance-grade model selection** — retired with the server; rebuild only on demonstrated need
- **Any metered-API feature** — the zero-marginal-cost proposition is load-bearing
- **Per-specialty engineer skills** — return only if usage data shows need (prior decision; schema-review earmarked for when federated-GraphQL work resumes)
- **Output-style delegation enforcement** — stays deferred until Epic 4 produces the evidence it was deferred for
- **Marketplace publishing before release discipline exists** — publication is committed, but only after Epic 3 lands (owner decision 2026-08-23)

## Epics and User Stories

### Epic 1: Trustworthy Memory (KG Health) — P0
Goal: The knowledge graph remains an asset, not a liability, as it accumulates.
Rationale for priority: KG-first reading is wired into every role today; polluted memory degrades every agent on every task, and pollution compounds with use.

- **US-101** — As a solo engineer, I want obsolete server-era knowledge in the KG neutralized (removed, superseded, or demoted) so that KG-first agents stop retrieving instructions for tools that no longer exist.
  - Acceptance criteria:
    - Given the benchmark query set, When any role runs its KG-first search, Then no returned observation directs use of retired components (agent CLI, agent-mcp, port 8082, Beads)
    - Given a piece of knowledge that is historical but true (e.g. the deprecation record itself), Then it is retained and identifiable as history, not guidance
  - Priority: P0
- **US-102** — As a solo engineer, I want a KG health report (counts, growth, stale/legacy share, orphaned entities) so that I can see graph condition without hand-written Cypher.
  - Acceptance criteria:
    - A single documented command produces the report for the current project
    - Report flags observations with zero-value timestamps as unassessable-age legacy
  - Priority: P0
- **US-103** — As a solo engineer, I want observation timestamps usable end-to-end so that future staleness policy is possible. (Investigation nuance: new writes already return real timestamps, but search results surface `0001-01-01` for stored observations — whether this is a storage or retrieval gap sits in the kg binary; architect input needed.)
  - Acceptance criteria:
    - Given any agent writes an observation, When it is later retrieved via search, Then it carries the real creation time
    - Legacy zero-dated observations are distinguishable from timestamped ones in the health report (US-102)
  - Priority: P1

### Epic 2: Agent Definition Quality Assurance — P0
Goal: A prompt edit cannot silently degrade an agent.
Rationale: the product surface is prompts; today's only gate is structural lint. This directly protects the "curated, disciplined roles" differentiator.

- **US-201** — As a maintainer, I want the specified planted-defect harness for the reviewer's audit mode built, so that reviewer prompt changes are validated against known-findable defects.
  - Acceptance criteria:
    - A fixture project with a documented planted-defect set exists; a harness run reports which defects the reviewer found
    - Given a reviewer.md change that removes an audit technique, When the harness runs, Then the score drop is visible before merge
  - Priority: P0
- **US-202** — As a maintainer, I want a lightweight behavioral smoke check for each agent (does it follow its KG-first step, output format, and completion contract on a canned task) so that regressions in any role are caught, not just the reviewer.
  - Acceptance criteria:
    - Each of the seven agents has at least one canned task with machine-checkable output assertions
    - Documented single command runs the suite; a failing agent is named with the violated contract
  - Priority: P1
- **US-203** — As a maintainer, I want prompt-change PRs to surface harness results in CI so that the gate is enforced, not remembered. (Subscription-funded CI execution pattern exists for consumer-project PR review; applicability here — architect input needed.)
  - Acceptance criteria:
    - Given a PR touching plugin/agents/ or plugin/skills/, Then harness status is visible on the PR before merge
  - Priority: P1

### Epic 3: Release & Versioning Discipline — P0 (raised from P1)
Goal: One version, one changelog, docs that match the tree, an install a stranger can complete.
Rationale: raised to P0 by owner decisions — ai-pack is a product for others and will publish to a public marketplace once this epic lands. Public contradictions (2.2.0 vs 3.0.0) are now product defects, not hygiene.

- **US-301** — As a user updating the plugin, I want a single authoritative version so that "what am I running" has one answer.
  - Acceptance criteria:
    - plugin.json, VERSION, and docs report the same version; CI fails on divergence
  - Priority: P1
- **US-302** — As a user, I want a CHANGELOG so that I can see what changed between updates.
  - Acceptance criteria:
    - CHANGELOG exists, covers the 3.x pivot onward; releases add entries (checked in CI or release steps)
  - Priority: P1
- **US-303** — As a user, I want docs counts/claims (agents, skills) checked against the tree so that documentation drift is caught automatically.
  - Acceptance criteria:
    - Given intro/README disagree with plugin/ contents, When CI runs, Then the build fails naming the mismatch
  - Priority: P1 (raised — audience-facing docs are product surface now)
- **US-304** — As an outside engineer, I want to install ai-pack without cloning the repo or hand-running submodule/python steps, so that first contact does not require understanding the repo layout. (Marketplace publication is the distribution goal; the exact packaging is architect input.)
  - Acceptance criteria:
    - A documented install path exists that a first-time user completes in under five minutes with no repo clone
    - The kg dependency installs or degrades gracefully (agents function without KG rather than erroring)
  - Priority: P0 within this epic — blocks marketplace publication

### Epic 4: Delegation Observability — P1
Goal: The user can see whether the framework is doing its job, from data already being collected.
Rationale: the hook shipped (metadata only: ts, session_id, subagent_type, description — no prompt content); the read layer is the missing half. Also the explicit evidence gate for the deferred enforcement decision. P1 not P0 only because the log must accumulate data to be worth reading.

- **US-401** — As a solo engineer, I want a delegation report over agent-spawns.jsonl (spawn counts by role, parallel-batch rate, sessions with zero spawns, per-project rollup) so that I can verify delegation and parallelism actually happen.
  - Acceptance criteria:
    - Single documented command; given an empty or missing log, Then it reports "no data yet" rather than erroring
    - Parallel batches (timestamp bursts) are identified as such
  - Priority: P1
- **US-402** — As a maintainer, I want the enforcement question re-opened with data so that the deferred output-style decision is made on evidence.
  - Acceptance criteria:
    - After ≥30 days of log data, a written recommendation (enforce / don't / refine) cites measured delegation rates against the decided floor: ≥70% of multi-step sessions delegating (owner decision 2026-08-23)
  - Priority: P2

### Epic 5: Durable Multi-Session Work — P2 (conditional)
Goal: Multi-session efforts have a discoverable record.
Rationale: task packets are optional and unindexed; `.ai/tasks/` holds dozens of stale server-era directories indistinguishable from live work. But demand post-pivot is unproven — do not build ahead of evidence.

- **US-501** — As a solo engineer, I want to list task packets with status (live vs. done vs. pre-pivot legacy) so that resuming multi-session work doesn't mean spelunking directories.
  - Acceptance criteria:
    - Single command lists packets with slug, dates, and result.md presence; legacy server-era packets are distinguishable
  - Priority: P2 — parked pending usage evidence (owner decision 2026-08-23: keep conditional; decide after ~a month of post-pivot packet usage data)

### Epic 6: Team Enablement — P1 (added by owner decision 2026-08-23)
Goal: Two or more engineers can share one ai-pack installation's benefits without stepping on each other.
Rationale: owner committed to the team story now. WHAT-level scope only; every mechanism here needs architect input.

- **US-601** — As a team member, I want a defined convention for what knowledge is shared vs. personal in a project's KG, so that one engineer's session benefits from — and cannot pollute — the team's accumulated knowledge.
  - Acceptance criteria:
    - A documented convention states what agents write to shared knowledge vs. keep session-local
    - Given two engineers work the same project, When one records a decision, Then the other's KG-first agents retrieve it
  - Priority: P1
- **US-602** — As a team member, I want the subscription-funded execution patterns (CI review tokens, agent runs) to work per-user, so that one person's account is not the team's single point of failure or cost.
  - Acceptance criteria:
    - Documented pattern for per-user OAuth tokens in shared repos; no shared personal credentials
  - Priority: P1
- **US-603** — As a team lead, I want delegation observability (Epic 4) to roll up across developers, so that adherence and parallelism are visible per person and per project.
  - Acceptance criteria:
    - The Epic 4 report accepts multiple developers' logs and attributes spawns per user
  - Priority: P2

## Constraints

- All execution funded by the Claude Max subscription; no metered API usage (product constraint, non-negotiable)
- Plugin-native only: agents, skills, hooks, MCP — no resident services (prior decision, 2026-08-22)
- Timestamp support (US-103) and any KG schema change land in the standalone kg binary/submodule — architect input needed
- Behavioral harness runs consume subscription quota; harness design must budget for this — architect input needed
- Deliverables must not reference internal consumer-project names

## Decisions — product owner, 2026-08-23

The six open questions were answered in a /prd review round:

1. **Product identity: a product for others.** Epic 3 raised to P0; install friction (US-304) added; docs are audience-facing product surface.
2. **Team story: now.** Epic 6 (Team Enablement) added at P1 — shared-KG conventions, per-user token patterns, team-level observability.
3. **Marketplace: yes, after Epic 3.** Publication is committed and gated on release discipline; US-304 blocks it.
4. **Multi-session demand: keep conditional.** Epic 5 stays parked as written; decide on ~a month of real post-pivot packet usage.
5. **Enforcement floor: ≥70% of multi-step sessions delegating.** US-402's recommendation measures against this.
6. **Agent Teams: watch, integrate when stable.** orchestrate stays as-is; when native Agent Teams stabilizes, adapt ai-pack roles to ride on it — differentiation remains memory + role quality + verified workflows.

## Risks

- **Claude Code natively absorbs orchestration (Agent Teams):** erodes the coordination pillar — mitigation: differentiate on memory + role quality + verified workflows; revisit positioning when Agent Teams stabilizes (Q6)
- **KG curation deletes knowledge that was still useful:** agents lose context — mitigation: supersede/demote before delete; retain history as identifiable history (US-101)
- **Harness flakiness (LLM nondeterminism) makes the QA gate noisy:** gate gets ignored — mitigation: score thresholds over pass/fail per defect; treat persistent flake as a design smell in the harness, not the agent
- **Subscription policy changes break the zero-cost or CI-review pattern:** proposition weakens — accept; monitor terms
- **Single-maintainer bus factor:** all discipline is one person's habit — partial mitigation is exactly Epics 2–3 (automated gates over memory); Epic 6 reduces it further
- **Public-product support expectations:** publishing raises issue/support load on a solo maintainer — mitigation: Epic 3's discipline (changelog, coherent versions, clean install) is precisely what keeps support load survivable; set expectations in README
