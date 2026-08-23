# PRD: AI-Pack Framework Strengthening

**Status:** Draft
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
- **Small teams (explicitly deferred):** multi-developer conventions (shared KG, shared logs, per-user tokens) are a separate product decision — see Open Questions. Nothing here may preclude it, but nothing targets it.

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
- **Team/multi-developer features** — separate product decision (Open Questions)
- **Per-specialty engineer skills** — return only if usage data shows need (prior decision; schema-review earmarked for when federated-GraphQL work resumes)
- **Output-style delegation enforcement** — stays deferred until Epic 4 produces the evidence it was deferred for
- **Marketplace publishing** — pending the product-identity decision (Open Questions)

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

### Epic 3: Release & Versioning Discipline — P1
Goal: One version, one changelog, docs that match the tree.
Rationale: cheap to fix, currently contradictory in public (2.2.0 vs 3.0.0, six-vs-seven agents), and prerequisite to any future distribution decision.

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
  - Priority: P2

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
    - After ≥30 days of log data, a written recommendation (enforce / don't / refine) cites measured delegation rates
  - Priority: P2

### Epic 5: Durable Multi-Session Work — P2 (conditional)
Goal: Multi-session efforts have a discoverable record.
Rationale: task packets are optional and unindexed; `.ai/tasks/` holds dozens of stale server-era directories indistinguishable from live work. But demand post-pivot is unproven — do not build ahead of evidence.

- **US-501** — As a solo engineer, I want to list task packets with status (live vs. done vs. pre-pivot legacy) so that resuming multi-session work doesn't mean spelunking directories.
  - Acceptance criteria:
    - Single command lists packets with slug, dates, and result.md presence; legacy server-era packets are distinguishable
  - Priority: P2 — proceed only if Open Question 4 answers "yes"

## Constraints

- All execution funded by the Claude Max subscription; no metered API usage (product constraint, non-negotiable)
- Plugin-native only: agents, skills, hooks, MCP — no resident services (prior decision, 2026-08-22)
- Timestamp support (US-103) and any KG schema change land in the standalone kg binary/submodule — architect input needed
- Behavioral harness runs consume subscription quota; harness design must budget for this — architect input needed
- Deliverables must not reference internal consumer-project names

## Open Questions — needs product owner input

1. **Is ai-pack a product for others or personal infrastructure?** The repo is public with docs and CI, but install assumes cloning and a single user. This decision gates how much of Epic 3 matters, whether US-303-style polish is worth it, and all of Q3. — blocks: Scope emphasis, Epic 3 depth
2. **Team story: now or later?** Shared KG, shared spawn logs, per-user OAuth tokens are all unaddressed. Deliberately excluded here; needs an explicit yes/no with timing. — blocks: Target Users, Non-Goals
3. **Publish to a public plugin marketplace?** Depends on Q1; would force release discipline (Epic 3) to P0 and raise support expectations. — blocks: Epic 3 priority, distribution scope
4. **Is there real multi-session demand post-pivot?** Task-packet usage since the pivot is the evidence; if none, Epic 5 should be dropped rather than built speculatively. — blocks: Epic 5
5. **What delegation-adherence threshold triggers enforcement?** US-402 will produce numbers; the acceptable floor (e.g. "≥70% of multi-step sessions delegate") is a product call. — blocks: US-402 recommendation
6. **Native Agent Teams posture.** Claude Code's experimental Agent Teams overlaps the orchestrate skill's territory. Watch, adopt, or integrate when stable? — blocks: long-term positioning of the orchestrate skill

## Risks

- **Claude Code natively absorbs orchestration (Agent Teams):** erodes the coordination pillar — mitigation: differentiate on memory + role quality + verified workflows; revisit positioning when Agent Teams stabilizes (Q6)
- **KG curation deletes knowledge that was still useful:** agents lose context — mitigation: supersede/demote before delete; retain history as identifiable history (US-101)
- **Harness flakiness (LLM nondeterminism) makes the QA gate noisy:** gate gets ignored — mitigation: score thresholds over pass/fail per defect; treat persistent flake as a design smell in the harness, not the agent
- **Subscription policy changes break the zero-cost or CI-review pattern:** proposition weakens — accept; monitor terms
- **Single-maintainer bus factor:** all discipline is one person's habit — partial mitigation is exactly Epics 2–3 (automated gates over memory)
