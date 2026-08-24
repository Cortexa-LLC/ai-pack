# ADR-010: Five-Minute No-Clone Install — Marketplace Fetch, Self-Bootstrapping kg, Graceful Degradation

**Date:** 2026-08-23
**Status:** Accepted
**Deciders:** Bryan Woodruff

---

## Context

US-304 (PRD "Framework Strengthening", Epic 3, P0 — gates marketplace publication):
an outside engineer installs ai-pack in under five minutes with no repo clone and no
hand-run submodule/python steps, and the kg dependency installs or degrades gracefully.

Current install path (README Quick Start) fails all of that:

1. `git clone` + `make install-plugin` — registers the checkout as a *local*
   marketplace. The clone exists only to serve as marketplace source.
2. `git submodule update --init mcp && python3 mcp/install.py --mcp kg` — builds the
   kg binary **from source**, requiring Go ≥1.24 with CGO (kg embeds KuzuDB via
   `go-kuzu v0.11.3`), and installs to `/usr/local/bin`.
3. `plugin/.mcp.json` launches bare `kg server --stdio` from PATH. If `kg` is absent,
   the kg MCP server fails and every role's mandatory "KG first" step
   (`kg__search_knowledge` before any exploration) plus KG checkpointing has no
   defined behavior — agent files contain **no** missing-KG instruction today, and
   the orchestrate skill actively instructs the clone-era submodule/python setup.

Facts constraining the design (verified against official Claude Code docs and the
binaries/source in this tree):

- **Marketplace fetch needs no clone.** `/plugin marketplace add <owner>/<repo>`
  accepts GitHub shorthand; this repo already has `.claude-plugin/marketplace.json`
  (name `ai-pack`, plugin source `./plugin`). Claude Code fetches the repo itself —
  but does **not** init submodules, so the `mcp/` source tree is unavailable at
  install time regardless.
- **No install-time lifecycle hook exists.** Claude Code plugins have no postinstall
  script. Accepted patterns for binary-backed MCP servers: manual install (the
  official LSP plugins do this) or lazy download at first server start.
- **A missing MCP command degrades, not breaks.** The plugin still loads; the server
  shows "Executable not found" in `/plugin` → Errors; only that server's tools
  disappear. Agents/skills keep working — *if* their instructions tolerate absent
  `kg__*` tools.
- **`${CLAUDE_PLUGIN_ROOT}` is supported** in a plugin's `.mcp.json` command.
- **The kg binary is not self-contained.** It dynamically links
  `@rpath/libkuzu.dylib`, and today's build embeds an rpath pointing into the
  builder's Go module cache
  (`~/go/pkg/mod/github.com/kuzudb/go-kuzu@v0.11.3/lib/dynamic/darwin`) — a copied
  binary breaks on any other machine. go-kuzu ships prebuilt dynamic libs for
  darwin, linux-amd64, linux-arm64, windows; CGO makes cross-compilation
  impractical, but native-runner CI builds are routine.
- **Cortexa-LLC/mcp is public with zero GitHub releases today.**
- Plugin updates are version-gated on `plugin/.claude-plugin/plugin.json` via
  `/plugin marketplace update` (matches the ADR-guarded release discipline in
  `docs/RELEASING.md`).

---

## Decision

Three parts: (1) distribute the plugin through the GitHub-hosted marketplace already
in this repo; (2) obtain kg as a **prebuilt, checksum-pinned release artifact**
fetched by a launcher script the plugin ships — no toolchain, no user step; (3) make
KG absence a **defined, silent degradation** in every role and skill contract.

### 1. Plugin fetch — GitHub marketplace, two slash commands

The repo is already a marketplace (`.claude-plugin/marketplace.json`, plugin source
`./plugin`). No packaging change to the fetch path. The entire user-facing install:

```
/plugin marketplace add Cortexa-LLC/ai-pack
/plugin install ai-pack@ai-pack
```

Restart Claude Code when prompted. Done — under a minute on a normal connection.
Updates: `/plugin marketplace update ai-pack` (auto-applies when `plugin.json`
version increases). `make install-plugin` stays as the *contributor* path for local
checkouts; the README stops presenting it as the user path.

### 2. kg — prebuilt release artifacts + self-bootstrapping launcher

**Release artifacts (Cortexa-LLC/mcp repo).** A GitHub Actions release workflow,
triggered by tag `kg/vX.Y.Z`, builds on native runners and publishes per-platform
tarballs `kg-<version>-<os>-<arch>.tar.gz` plus a `checksums.txt`:

- Targets v1: `darwin-arm64`, `darwin-x86_64`, `linux-x86_64`, `linux-arm64`
  (runners: `macos-14`, `macos-13`, `ubuntu-24.04`, `ubuntu-24.04-arm`).
  Windows deferred (see Consequences).
- Each tarball contains `kg` **and** `libkuzu.dylib`/`libkuzu.so` side by side, with
  the binary's rpath rewritten to `@loader_path` (macOS, `install_name_tool`) /
  `$ORIGIN` (Linux, `-ldflags` or `patchelf`) so the pair is relocatable. This fixes
  the module-cache rpath defect as a prerequisite.

**Launcher (this repo).** `plugin/.mcp.json` changes to:

```json
{ "mcpServers": { "kg": {
    "type": "stdio",
    "command": "${CLAUDE_PLUGIN_ROOT}/bin/kg-launch.sh",
    "args": ["server", "--stdio"] } } }
```

`plugin/bin/kg-launch.sh` (POSIX sh, no dependencies beyond `curl`/`tar`/`shasum`)
resolves a kg binary and `exec`s it with the given args, in order:

1. `$AI_PACK_KG` if set (explicit override / air-gapped escape hatch).
2. `kg` on PATH (existing source builds and contributor installs keep working
   unchanged).
3. Cached bootstrap: `~/.ai-pack/bin/kg-<version>/kg`.
4. **Bootstrap:** read `${CLAUDE_PLUGIN_ROOT}/kg.lock.json` — a file this repo
   commits, containing the pinned kg version, artifact base URL, and **per-platform
   sha256** — download the platform tarball from the mcp repo's GitHub Release,
   verify the checksum, extract into a temp dir, atomically `mv` into the cache path
   (a `mkdir`-based lock prevents two concurrent sessions racing), then exec.
5. Any failure (unsupported platform, offline, checksum mismatch): print one
   diagnostic line to stderr naming the manual-install fallback and **exit
   non-zero**. Claude Code then lists kg under `/plugin` → Errors and the session
   proceeds without `kg__*` tools — which part 3 makes safe.

A `SessionStart` hook entry (added to the existing `plugin/hooks/hooks.json`) runs
`kg-launch.sh --fetch-only` in the background so that if the first MCP connect ever
races a slow download past the MCP startup timeout, the artifact is cached and the
server comes up cleanly next session. Version bumps are just a `kg.lock.json` edit
released with a normal plugin version bump; the launcher sees an uncached version
and re-bootstraps. Old cache dirs are left in place (few tens of MB; manual removal
documented).

### 3. Graceful-degradation contract (all 7 agents + orchestrate/prd skills)

One normative block, added verbatim to the KG-first section of every file in
`plugin/agents/` and to both KG-referencing skills:

> **KG availability:** If the `kg__*` tools are not in your tool list, or the first
> KG call fails with a server/connection error, the knowledge graph is not installed
> — **skip every KG step silently** (KG-first queries *and* KG checkpointing), rely
> on file exploration, and do not mention the absence in your report unless the task
> is *about* the KG. Never retry, never attempt a bash `kg` fallback, never treat
> missing KG as a blocker or error.

The orchestrate skill's clone-era setup instructions (`git submodule update` +
`python3 mcp/install.py`, SKILL.md lines ~58–66) are deleted and replaced with a
pointer to the degradation contract. This is the contract US-304's second criterion
names: agents function without KG rather than erroring; the KG is an accelerator,
never a dependency.

### The documented install path, exactly as it will exist (README Quick Start)

```
Prerequisites: Claude Code installed and working. macOS or Linux.

1. In any Claude Code session:
     /plugin marketplace add Cortexa-LLC/ai-pack
     /plugin install ai-pack@ai-pack
2. Restart Claude Code.

That's it. The knowledge-graph server (kg) downloads itself on first launch
(prebuilt, checksum-verified — no toolchain needed). If it can't (offline,
unsupported platform), everything still works; agents simply run without
persistent memory. Check /plugin → Errors to see kg's status, or /mcp.

Offline / air-gapped: build kg from source (github.com/Cortexa-LLC/mcp) and
either put it on PATH or set AI_PACK_KG=/path/to/kg.
```

Two commands, one restart, zero terminal steps — comfortably under five minutes.

---

## Alternatives Rejected

- **Build kg from source at first launch** (launcher runs `go build`): requires
  Go 1.24 + CGO toolchain on the user's machine — exactly the hand-run
  toolchain burden US-304 removes.
- **Commit binaries into the plugin repo** (or git-LFS): four platforms × ~tens of
  MB shipped to every user on every marketplace fetch/update, and git history bloat;
  git is not a binary store.
- **Manual install step, LSP-plugin style** ("brew install / curl-pipe-sh, then
  install the plugin"): honest but reintroduces a hand-run step and a failure mode
  *before* first value; survives only as the documented air-gapped fallback (PATH /
  `AI_PACK_KG` are honored first by the launcher).
- **npm-wrapped binary** (`npx` in `.mcp.json`): adds a Node runtime dependency and
  a second publish surface for no gain over a GitHub Release.
- **Marketplace `command` plugin source running a setup script**: install-time
  arbitrary script execution is opaque to users, re-runs per session, and couples
  plugin fetch to network setup; the launcher keeps setup inside the one component
  that needs it.
- **Ship without kg / make KG opt-in only**: fails "installs **or** degrades" — the
  PRD wants the default install to include memory; degradation is the fallback, not
  the product.
- **Static-link kuzu to avoid bundling the dylib**: go-kuzu ships only dynamic
  libs; producing static kuzu builds is upstream work with no owner. Bundling +
  rpath rewrite achieves relocatability now.

---

## Implementation Work Breakdown

Ordered; WB-2 depends on WB-1's artifact URLs/checksums. WB-3/WB-4 are parallel.

**WB-1 — mcp repo: kg release pipeline** (`github.com/Cortexa-LLC/mcp`)
- Fix relocatability: build kg with rpath `@loader_path` (macOS) / `$ORIGIN`
  (Linux); copy the matching `libkuzu` from
  `go-kuzu@v0.11.3/lib/dynamic/<platform>/` next to it. Verify with
  `otool -l` / `readelf -d` in CI (assert no absolute Go-module-cache rpath).
- `.github/workflows/release-kg.yml`: on tag `kg/v*`, matrix over
  `macos-14`, `macos-13`, `ubuntu-24.04`, `ubuntu-24.04-arm`; build, smoke-test
  (`./kg --version`, `./kg server --stdio` handshake), tar as
  `kg-<version>-<os>-<arch>.tar.gz`, generate `checksums.txt`, attach all to a
  GitHub Release via `gh release create`.
- Acceptance: on a machine with no Go and no module cache, downloading + extracting
  a tarball yields a `kg` that runs `kg server --stdio`.

**WB-2 — ai-pack: launcher + lock + wiring**
- Add `plugin/bin/kg-launch.sh` implementing resolution order
  `AI_PACK_KG → PATH → cache → bootstrap` as specified in Decision 2, including
  checksum verification, temp-dir + atomic `mv`, `mkdir` lock, `--fetch-only` mode,
  and single-line stderr diagnostic on failure. POSIX sh; no bash-isms.
- Add `plugin/kg.lock.json`: `{ "version", "base_url", "sha256": { "<os>-<arch>":
  "<hex>" } }` populated from WB-1's first release.
- Change `plugin/.mcp.json` command to
  `${CLAUDE_PLUGIN_ROOT}/bin/kg-launch.sh` (args unchanged).
- Add `SessionStart` hook to `plugin/hooks/hooks.json` running
  `"${CLAUDE_PLUGIN_ROOT}/bin/kg-launch.sh" --fetch-only` in the background.
- Acceptance: fresh machine (no `kg` anywhere) → install plugin → within one session
  restart, `/mcp` shows kg connected; with network blocked, kg shows failed in
  `/plugin` Errors and sessions run normally.

**WB-3 — ai-pack: degradation contract in role/skill definitions**
- Insert the Decision-3 block (verbatim) into the KG-first section of all 7 files in
  `plugin/agents/` and into `plugin/skills/orchestrate/SKILL.md` and
  `plugin/skills/prd/SKILL.md`.
- Delete the submodule/python setup instructions from
  `plugin/skills/orchestrate/SKILL.md` (~lines 58–66).
- Acceptance: with the kg server disabled, spawning each role on a trivial task
  produces no `kg`-related errors, retries, or report noise.

**WB-4 — ai-pack: docs**
- README Quick Start replaced with the exact block in this ADR; move
  clone/`make install-plugin`/submodule instructions to a "Contributing /
  developing ai-pack" section; update the Knowledge Graph section (bootstrap,
  cache location `~/.ai-pack/bin/`, `AI_PACK_KG`).
- `docs/RELEASING.md`: add the kg pin-bump procedure (tag `kg/vX.Y.Z` in mcp repo →
  update `plugin/kg.lock.json` checksums → normal plugin version bump/CHANGELOG).
- `scripts/verify-kg.sh`: update the not-found message (submodule instructions →
  launcher/cache/`AI_PACK_KG`).
- Acceptance: `claude plugin validate ./plugin` passes; docs contain no user-facing
  clone/submodule/python step.

**WB-5 — verification harness**
- Launcher unit tests (bats or plain sh): PATH-hit exec, cache-hit exec, checksum
  mismatch → non-zero + no cache write, offline → non-zero, concurrent bootstrap →
  one download, `--fetch-only` populates cache without exec.
- End-to-end smoke doc/script: temp `$HOME`, marketplace add → install → kg tools
  present; repeat with `AI_PACK_KG=/nonexistent` → agents still complete (WB-3 gate).

---

## Consequences

- **Marketplace publication unblocks:** both US-304 criteria are met by WB-1…WB-4;
  the community-marketplace submission (version/tag/CHANGELOG consistency already
  enforced by `release-consistency` CI) can follow as a separate release step.
- **Two-repo release coupling:** kg versions are pinned by `plugin/kg.lock.json`;
  an mcp-repo release is inert until this repo bumps the lock — deliberate, keeps
  plugin versioning the single user-facing version (per `docs/RELEASING.md`).
- **First-launch network dependency:** one ~tens-of-MB download per kg version per
  machine, checksum-pinned against the tag's release artifacts. Failure is
  non-fatal by construction (Decision 3). A very slow first download may push kg
  availability to the next session; the `SessionStart` `--fetch-only` hook
  self-heals this.
- **Windows deferred:** go-kuzu ships windows libs, but the launcher is POSIX sh
  and no Windows runner smoke exists; Windows users today get documented
  degradation (or manual build + `AI_PACK_KG`). Revisit on demand — the artifact
  naming scheme already reserves `windows-x86_64`.
- **Contributor path unchanged:** PATH resolution before cache means source builds
  (`python3 mcp/install.py --mcp kg`) transparently override the pinned artifact on
  developer machines.
- The ADR-009 note stands: kg-repo changes (timestamp scan fix, `kg health`) ride
  the same new release pipeline — WB-1 is the distribution channel ADR-009's specs
  were missing.
