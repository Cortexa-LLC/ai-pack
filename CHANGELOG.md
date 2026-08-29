# Changelog

All notable changes to ai-pack are documented in this file. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versions follow
[Semantic Versioning](https://semver.org/).

The plugin manifest (`plugin/.claude-plugin/plugin.json`) is the single source of
truth for the current version. The `VERSION` file is a one-line mirror of it, and
CI (`release-consistency`) fails if they disagree or if this file lacks an entry
for the manifest version.

## [3.4.0] — 2026-08-29

### Fixed

- **The shepherd could not see findings from a body-posting reviewer** (issue #41,
  item 3). The state machine keyed entirely on `THREAD_COUNT` from `reviewThreads`,
  but a reviewer posting via `gh pr review --body-file` submits a single top-level
  review body and *cannot* attach line comments. `THREAD_COUNT` was therefore always
  0, the fix route never fired, and the shepherd waited out its polling budget beside
  a review full of findings. Step 2 now also reads the review body from the same REST
  payload, and routing keys on `ACTIONABLE` — threads, a review body, or both.
- **A by-design verdict failure was reported as broken CI** (issue #41, item 4). The
  review gate check fails deliberately when the reviewer requests changes, but it
  landed in `CI_FAILING`, firing Route D's "CI is failing — investigate CI failures"
  on a PR whose builds were all green and whose real signal was a review finding.
  `CI_FAILING` now excludes the gate check by name (`REVIEW_GATE_CHECK`, overridable
  for consumer repos). While the gate is running it counts toward `CI_PENDING`; once
  it reaches `FAILURE` it falls out of both counts and its signal is carried entirely
  by `VERDICT`. Genuine build failures alongside the gate are still caught.
- **A concluded review was polled as if it might change** (issue #41, item 5). Route A
  terminated only on `APPROVED`, so a `COMMENTED` head fell through to the wait route
  and polled 30 × 90s ≈ 45 minutes before escalating as "reviewer appears stuck" —
  though nothing was stuck and the findings were sitting in the review body. Route E
  now waits only on `PENDING`; `COMMENTED` and `CHANGES_REQUESTED` are treated as
  terminal, routing to the fix path when they carry findings and to a new Route G
  ("Reviewed, Not Approved") when they do not.

### Changed

- Step 5 handles both finding shapes; Steps 7–8 no longer assume every finding lives
  in a resolvable thread — a body-posting reviewer gets one disposition comment per
  round, since it offers nothing to reply to or resolve.
- `plugin/agents/pr-shepherd.md` carries all three corrections, including an explicit
  terminal-verdict exit mirroring Route G — without it the agent variant kept looping
  on a concluded review until `MAX_WAIT_ITER`, reproducing item 5 in the half of the
  stack the first pass left behind. Its Done Check also re-derives the verdict and
  body from one fresh snapshot rather than reusing pre-push state.
- The agent excludes `$REVIEW_GATE_CHECK` from **both** of its CI-health queries, not
  just the skill's. Step 2's `FAILURES` had been diagnosing the gate's by-design
  failure as broken CI, and — worse — Step 7's `ALL_OK` counted it too, pinning
  `ALL_OK=false` on every changes-requested head. Since the terminal-verdict exit is
  guarded on `ALL_OK=true`, that exit could never fire for `CHANGES_REQUESTED`, the
  primary verdict it was written for, and the agent fell back to the loop it was
  meant to replace.
- Both variants test the terminal verdict as the complement of `APPROVED` and
  `PENDING` rather than enumerating `COMMENTED`/`CHANGES_REQUESTED`, so a `DISMISSED`
  or otherwise unrecognised verdict reports and stops instead of matching no route
  and exiting silently.

## [3.3.1] — 2026-08-29

### Fixed

- **The kg download URL would have 404'd on the first real release.** The launcher
  builds `<base_url>/<version>/kg-<version>-<platform>.tar.gz`, and GitHub renders a
  slash-containing tag as path separators — tag `kg/v1.2.3` serves its assets from
  `.../releases/download/kg/v1.2.3/`. The shipped `base_url` omitted the `kg`
  segment, so it would have requested `.../releases/download/v1.2.3/...`. The tag is
  now split correctly across the two lock fields, and `docs/RELEASING.md` step 2
  spells out both halves. No user-visible change: the pin is still empty.
- **`scripts/verify-kg.sh` overclaimed parity with the launcher.** It said it mirrors
  the resolution order, but the launcher opens the exact version pinned in
  `kg.lock.json` while the script takes the first cached version it finds. The
  comment now states the difference.
- **The KG-availability block is formatted consistently** across all seven agents.
  `inspector.md` and `spelunker.md` used an inline arrow form; they now use the same
  standalone paragraph as the others. Wording was already identical.

  These three shipped in 3.3.0's tree without a version bump, so `3.3.0` briefly
  denoted two different plugin trees and a version-gated `/plugin marketplace update`
  could not deliver them. This release restores the invariant; `release-consistency`
  now fails any PR that changes `plugin/` without bumping the manifest.

### Changed

- **`release-consistency` gained a fourth check**: a PR touching `plugin/` must also
  change `plugin/.claude-plugin/plugin.json`. Plugin updates are version-gated, so an
  unbumped change to `plugin/` can never reach an installed user — previously nothing
  caught that.

## [3.3.0] — 2026-08-29

### Added

- **No-clone install path** (issue #18, ADR-010). The documented install is now two
  slash commands — `/plugin marketplace add Cortexa-LLC/ai-pack` then
  `/plugin install ai-pack@ai-pack` — with no clone, no submodule, and no toolchain.
  The clone-based flow moves to a "Developing ai-pack itself" section.
- **Self-resolving kg launcher** (`plugin/bin/kg-launch.sh`). `plugin/.mcp.json` now
  launches the plugin's own script instead of a bare `kg`. It resolves, in order:
  `$AI_PACK_KG`, `kg` on `PATH`, a cached download under `~/.ai-pack/bin/`, then the
  checksum-pinned release named in `plugin/kg.lock.json`. A `SessionStart` hook warms
  the cache in the background via `--fetch-only`. Lock values are validated before
  use — a version that is not a `vX.Y.Z` tag, a non-`https` base URL, or a malformed
  digest all fail closed rather than fetching something unverified.
- **`plugin/kg.lock.json`** — the kg pin. It ships **empty** because no kg release
  exists yet, which the launcher handles by design: resolution falls back to
  `$AI_PACK_KG` and `PATH`, and the download path exits cleanly into graceful
  degradation. `docs/RELEASING.md` documents populating it.
- **Launcher test suite** (`tests/kg-launch/test-kg-launch.sh`) — 16 behavioral tests
  over resolution order, lock validation, checksum verification, cache writes,
  `--fetch-only`, and the offline path. The bootstrap tests build a real tarball and
  serve it over `file://`, so the suite needs no network.

### Changed

- **KG absence is now a defined contract, not undefined behavior.** All seven agents
  and the orchestrate/prd skills carry a verbatim "KG availability" block: if the
  `kg__*` tools are missing or the first call fails, skip every KG step silently,
  never retry, never shell out to `kg`, and never report the absence as a problem.
  Previously no agent file said anything about a missing KG.
- The orchestrate skill's clone-era setup instructions (`git submodule update` +
  `python3 mcp/install.py`) are gone — they broke under a marketplace install, which
  never fetches submodules.
- `scripts/verify-kg.sh` mirrors the launcher's resolution order and its not-found
  message describes the launcher rather than the retired submodule step.

## [3.2.1] — 2026-08-28

### Fixed

- **shepherd-pr: the review verdict never matched an app-posted review** (issue #31).
  The skill read the verdict from `gh pr view --json reviews`, whose GraphQL-backed
  `author.login` reports app logins **without** the `[bot]` suffix
  (`cortexa-llc-reviewer`, not `cortexa-llc-reviewer[bot]`) and carries no type
  field — so the `endswith("[bot]")` filter matched nothing and the verdict was
  always `PENDING`. Route A ("done") could therefore never fire: an approved PR
  kept scheduling wakeups until `MAX_WAIT_ITER` escalated it as "reviewer stuck".
  The verdict now comes from the REST reviews endpoint filtered on
  `.user.type == "Bot"` — the same source and predicate the A2 convergence check
  already used. Verified against a real approved PR: the old filter returned
  `PENDING`, the new one returns `APPROVED`.
- **pr-shepherd agent: same predicate hardened.** Both of its verdict queries were
  already on REST (where the `[bot]` suffix does exist, so they worked), but they
  now match on `.user.type` too, so the two code paths cannot drift apart again.

## [3.2.0] — 2026-08-27

### Fixed

- **Plugin failed to load entirely** — `plugin.json` declared
  `"hooks": "./hooks/hooks.json"`, but the loader already loads that standard
  path automatically. The duplicate registration was rejected outright:
  `claude plugin list` reported `✘ failed to load` with "Duplicate hooks file
  detected". Removing the key restores `✔ enabled`. `manifest.hooks` is only
  for *additional* hook files beyond the standard path — do not re-add it.

### Added

- **KG-first directive for the main session** (`plugin/hooks/kg-first-directive.py`,
  wired as a `SessionStart` hook). The KG-first rule was written into every
  `plugin/agents/*.md` role, so it reached subagents only — a session doing work
  inline never saw it, despite being the role that reads the most raw file content.
  The hook injects the directive at session start for every source (startup,
  resume, clear, compact, fork), so it also survives compaction. It covers the
  write half too: record what changed after landing work, since the graph is only
  worth reading because earlier sessions wrote to it.

## [3.1.1] — 2026-08-27

### Fixed

- **pr-shepherd: the "phantom watcher" termination bug** — the agent could end its
  turn after pushing a fix round, claiming a background watcher would re-invoke it.
  Nothing re-invokes a completed agent. All waiting is now synchronous and inline
  (foreground `gh run watch` / `gh pr checks --watch` plus the bounded sleep loop),
  and the anti-pattern is named and documented.

### Changed

- **Convergence-gated iteration budget** for the pr-shepherd role and the
  `shepherd-pr` skill — a 5-round minimum before "budget exhausted" is a valid stop,
  continued looping while rounds converge, and a distinct **stuck** report when the
  reviewer re-raises the same findings on two consecutive passes.
- **`[OBSOLETE]` KG observations are history, not guidance** — architect, engineer,
  inspector, pr-shepherd, product-manager, reviewer, and spelunker now treat them as
  historical record rather than instructions to follow.
- **reviewer**: `gh pr review --approve` / `--request-changes` is no longer forbidden
  outright. The `gh api …/reviews --method POST` form stays preferred because it
  carries inline comments, with a documented exception for restricted environments
  (the CI review workflow scopes Bash to `gh pr review`) — post one consolidated
  review there instead.

## [3.1.0] — 2026-08-23

### Added

- **product-manager role restored plugin-native** — a product-manager agent, the
  `/prd` interview skill, and a PRD template.
- **Agent-spawn logging hook** — event-driven logging of every `Agent` tool spawn.
- **2.x uninstall scripts** for cleaning the retired server era off machines and
  out of consumer projects.
- **Framework-strengthening PRD accepted**
  (`docs/product/prd-framework-strengthening.md`) — ai-pack is a product for
  others; marketplace publication is committed, gated on release discipline.

### Changed

- **reviewer**: senior-IC adversarial stance by default, plus a whole-project
  AUDIT MODE for reviewing an entire codebase rather than a diff.
- **KG-read-first parity** — all creation roles and planning skills consult the
  knowledge graph before acting, not only after.
- **pr-shepherd hardening** — must run to completion (no stop-and-wait), with a
  settle check and a treadmill guard learned from the first delegated runs.
- `make update-plugin` now forces a resync of the installed plugin.
- Docs site rebuilt plugin-first; server-era content pruned.

## [3.0.0] — 2026-08-22

The pivot: ai-pack is a Claude Code plugin, not an agent server. Claude Code
natively provides the agent execution loop, so the API-driven server stack was
removed from the tree.

### Removed

- **BREAKING**: the Go agent server (port 8082), the `agent` CLI, the
  `agent-mcp` MCP server, and the React GUI. The server era is preserved at tag
  `v2.0-server-final`.

### Changed

- Repo restructured plugin-first; `plugin/agents/` became the single canonical
  source of role definitions (no duplicate copies anywhere in the tree).
- The orchestrate and pre-push skills moved off the deprecated `agent` CLI onto
  the built-in `Agent` tool.

### Added

- CI validation of plugin structure and the docs build, plus branch protection
  on `main`.

## 2.x and earlier

The pre-plugin (agent-server) era is preserved in full at tag
`v2.0-server-final`. Version history carried over from the old `VERSION` file:

- 2.0.0 (2026-01-14): Agent tracking consolidation on Beads (BREAKING)
- 1.1.0 (2026-01-12): Beads integration for task tracking
- 1.0.0 (2026-01-07): Initial release with AI workflow framework
