# Changelog

All notable changes to ai-pack are documented in this file. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versions follow
[Semantic Versioning](https://semver.org/).

The plugin manifest (`plugin/.claude-plugin/plugin.json`) is the single source of
truth for the current version. The `VERSION` file is a one-line mirror of it, and
CI (`release-consistency`) fails if they disagree or if this file lacks an entry
for the manifest version.

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
