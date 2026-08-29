# Releasing ai-pack

`plugin/.claude-plugin/plugin.json` is the **single source of truth** for the
version. The `VERSION` file is a one-line mirror; the `release-consistency` CI
job fails any PR where they disagree, where `CHANGELOG.md` lacks an entry for
the manifest version, or where the docs' spelled-out agent/skill counts drift
from the tree.

## Steps

1. **Bump the version** in `plugin/.claude-plugin/plugin.json` (semver: breaking
   → major, features → minor, fixes → patch).
2. **Mirror it to `VERSION`** — the file is exactly one line, the version string.
3. **Add a `CHANGELOG.md` entry** — `## [<version>] — <date>` at the top, Keep a
   Changelog style (Added / Changed / Removed / Fixed).
4. **Open a PR.** `main` is protected; the `release-consistency` job enforces
   the three checks above alongside plugin validation and the docs build.
5. **After merge**, run `make update-plugin` to resync the locally installed
   plugin.
6. **Tag the release** on `main`: `git tag v<version> && git push origin v<version>`.

## Bumping the pinned kg release

`plugin/kg.lock.json` pins the prebuilt kg artifact that `plugin/bin/kg-launch.sh`
downloads. It is deliberately decoupled from the kg source repo: a release there is
inert until this file names it, which keeps the plugin version the single
user-facing version.

It currently ships with an **empty pin** (`"version": ""`), because no kg release
exists yet. The launcher handles that state by design — it resolves `$AI_PACK_KG`
and `PATH` as usual, and its download path exits cleanly into the documented
graceful degradation. To populate or bump it:

1. Cut a kg release in [Cortexa-LLC/mcp](https://github.com/Cortexa-LLC/mcp),
   tagged `kg/vX.Y.Z`, publishing one `kg-vX.Y.Z-<platform>.tar.gz` per supported
   platform (`darwin-arm64`, `darwin-x86_64`, `linux-x86_64`, `linux-arm64`).
   Each tarball must be relocatable — the binary plus its `libkuzu`, rpath
   rewritten to `@loader_path`/`$ORIGIN` (ADR-010, WB-1).
2. Record the tag in `version`, the release download base in `base_url`, and each
   artifact's `sha256` under its platform key. The launcher rejects a version that
   is not a `vX.Y.Z` tag, a `base_url` that is not `https`, and any digest that is
   not 64 hex characters — so a malformed edit fails closed rather than fetching
   something unverified.
3. Ship it with a normal plugin version bump and CHANGELOG entry (steps 1–4 above).
   Users pick it up on `/plugin marketplace update ai-pack`; the launcher sees an
   uncached version and re-bootstraps.

Old cache directories under `~/.ai-pack/bin/` are left in place across bumps;
removing them is manual and safe at any time.

Run `tests/kg-launch/test-kg-launch.sh` after touching the launcher or the lock
file shape.

## Marketplace publication

Publishing to the Claude Code plugin marketplace (US-304 in
`docs/product/prd-framework-strengthening.md`) is a separate, pending
workstream — committed, but gated on this release discipline being in place.
These steps do not publish anything externally.
