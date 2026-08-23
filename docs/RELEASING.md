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

## Marketplace publication

Publishing to the Claude Code plugin marketplace (US-304 in
`docs/product/prd-framework-strengthening.md`) is a separate, pending
workstream — committed, but gated on this release discipline being in place.
These steps do not publish anything externally.
