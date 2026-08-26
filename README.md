# AI-Pack

<p align="center">
  <img src="assets/Banner.png" alt="AI-Pack Banner" width="800" />
</p>

**A Claude Code plugin providing a coordinated team of specialized subagents.**

AI-Pack turns a Claude Code session into an engineering team. It ships seven specialized subagents (architect, engineer, inspector, pr-shepherd, product-manager, reviewer, spelunker), four workflow skills (orchestrate, prd, pre-push, shepherd-pr), and a knowledge-graph MCP server (`kg`) that gives agents persistent memory of your codebase. Claude Code provides the execution loop; AI-Pack provides the roles, the coordination patterns, and the memory.

---

## Quick Start

**Prerequisites:** [Claude Code](https://docs.anthropic.com/en/docs/claude-code) installed and working.

```bash
# 1. Clone the repo
git clone https://github.com/Cortexa-LLC/ai-pack.git
cd ai-pack

# 2. Install the plugin into Claude Code
make install-plugin      # registers the local marketplace, installs ai-pack@ai-pack

# 3. Set up the kg knowledge-graph server (used by plugin/.mcp.json)
git submodule update --init mcp
python3 mcp/install.py --mcp kg

# Later: pull changes, then refresh the installed plugin
make update-plugin
```

After installation, restart Claude Code. The subagents are available via the `Agent` tool as `ai-pack:<name>`, and the skills trigger automatically or via `/ai-pack:<skill>`.

---

## The Agents

Seven subagents live in `plugin/agents/`. Each is a self-contained role definition with its own tool discipline, quality gates, and reporting format.

| Agent | What it does | Example invocation |
|-------|--------------|--------------------|
| **architect** | Technical design: module boundaries, API contracts, data models, ADRs | "design the architecture for the notification system" |
| **engineer** | Implementation: writes code, fixes bugs, creates tests | "implement the authentication feature" |
| **inspector** | Root-cause analysis for complex bugs; produces a fix specification | "investigate why the payment processor occasionally returns 500" |
| **pr-shepherd** | Drives a GitHub PR to merge-ready: watches CI, fixes failures, answers reviewer threads | "shepherd PR #42 to merge-ready" |
| **product-manager** | Product requirements: PRDs with measurable goals, non-goals, epics, and user stories with testable acceptance criteria | "write a PRD for the notification system from this discovery transcript" |
| **reviewer** | Code review: quality, security, best practices, structured findings with severities | "review the auth handler for security issues" |
| **spelunker** | Codebase investigation: traces execution paths, maps dependencies, answers "how does X work" | "trace the execution path for this failing test" |

Spawn them from any Claude Code session:

```
Use the Agent tool with subagent_type "ai-pack:engineer" and a fully
self-contained prompt — subagents share no memory with your session.
```

---

## The Skills

Four skills in `plugin/skills/` package multi-agent workflows:

- **orchestrate** — decomposes engineering work and delegates it to the subagents. Triggers on multi-step requests like "build this feature end to end" or "investigate why X is broken then fix it".
- **prd** — product discovery interview in the main session (subagents cannot question the user), then delegates PRD drafting to the product-manager agent. The finished PRD lands in `docs/product/`. Triggers on "create a PRD" or "help me spec this feature".
- **pre-push** — review-and-fix loop on local commits before pushing. Spawns a reviewer against the local diff; if issues are found, spawns an engineer to fix them, amends, and re-reviews until approved. Triggers on "review my commits before I push".
- **shepherd-pr** — drives an open GitHub PR to a green, approved, mergeable state by spawning the pr-shepherd agent. Triggers on "shepherd PR #42" or "drive my PR to green".

---

## Task Packets (optional)

Agent briefs are passed directly in Agent-tool prompts. For multi-session work where a brief must outlive any single session, an optional two-file convention under `.ai/tasks/<slug>/` is available:

- `task.md` — the orchestrator's brief: what to do, files to change, acceptance criteria, constraints, context. Fully populated, never template placeholders.
- `result.md` — written by the agent when done: findings, decisions, blockers.

Templates live in `templates/task-packet/`. Additional document templates (ADRs, incident reports, investigations, security docs, PRDs) are in `templates/`.

---

## Knowledge Graph

`plugin/.mcp.json` launches the `kg` MCP server (`kg server --stdio`), which indexes your codebase into a persistent knowledge graph. Agents query it for file context, entity relationships, and accumulated observations across sessions. `kg` is a standalone binary installed from the [`mcp` submodule](https://github.com/Cortexa-LLC/mcp) (see Quick Start). Verify the installation with `scripts/verify-kg.sh`.

---

## Repository Layout

```
plugin/            The product: agents/, skills/, .mcp.json
.claude-plugin/    Local marketplace definition (marketplace.json)
templates/         ADR, incident, investigation, security, task-packet, PRD templates
scripts/           Billing checkers, kg verification, submodule reset
docs/              Docusaurus documentation site (docs/website)
assets/            Logos and banners
mcp/               Git submodule providing the kg binary
```

---

## Automated PR Review

Once provisioned, every same-repo, non-draft pull request gets an advisory review
from Claude (`.github/workflows/claude-pr-review.yml`). The workflow installs the
Claude Code CLI on the runner and runs `claude -p` directly — no GitHub App
installation is required. The system prompt is the ai-pack reviewer role
(`plugin/agents/reviewer.md`), loaded from the base branch so a PR cannot tamper
with the prompt that reviews it; existing review threads are fed back in so
repeated pushes don't re-raise the same findings. Claude posts one review per run
with severity-graded findings and a verdict: approve (zero Critical and zero
Major findings), comment, or request changes. The verdict participates in branch
protection — the bot's approval satisfies the required review, and its
request-changes blocks merge — while the workflow's status check itself stays
advisory (a runtime failure never turns the PR red). Merging always remains a
human action; GitHub auto-merge must never be enabled on this repository.

**Provisioning** (one-time, subscription-funded; no metered API key):

1. Mint and store the OAuth token:

```bash
claude setup-token                        # mint a long-lived OAuth token locally
gh secret set CLAUDE_CODE_OAUTH_TOKEN     # store it as a repo Actions secret
```

2. Optional — proper bot identity: set `CORTEXA_LLC_REVIEWER_APP_ID` and
   `CORTEXA_LLC_REVIEWER_PRIVATE_KEY` (credentials of the org's reviewer GitHub
   App) as repo Actions secrets. With them, reviews post as the reviewer bot;
   without them, the workflow degrades gracefully and reviews post as
   `github-actions[bot]`.

Until the OAuth secret exists, the review job skips cleanly (never a red check).
Rotate the token by re-running step 1.

**Fork posture:** the workflow triggers on `pull_request` only (never
`pull_request_target`) and runs only for same-repo, non-draft PRs, so the OAuth
token is never exposed to fork PRs. Fork contributions are reviewed by humans or by
a maintainer pushing the branch into the repo.

---

## Migrating from 2.x (server era)

The 3.x plugin needs **no per-project integration** — no submodule, no hooks, no copied commands. Install it once (`make install-plugin`) and it works in every project; the knowledge graph creates `.ai/knowledge.db` per project automatically on first use.

If you previously integrated the 2.x server, two cleanup scripts remove exactly what the old installer placed (both support `--dry-run`):

```bash
# Once per machine: stop services, remove binaries and MCP registration
# (archives your task history first; --purge also removes ~/.ai-pack)
bash scripts/uninstall-server.sh

# Once per integrated project, from that project's root: remove the
# .ai-pack submodule, ai-pack slash commands, hooks, rules, and template
# skills — preserving .ai/ (knowledge graph + task history) and your CLAUDE.md
bash /path/to/ai-pack/scripts/uninstall-project.sh
```

After cleanup, review the project's `CLAUDE.md` by hand if it was copied from the 2.x template — the scripts flag it but never edit it.

## History

AI-Pack 1.x/2.0 was an API-driven agent server: a Go server on port 8082 running coding agents against the Claude API, with an `agent` CLI, an `agent-mcp` MCP server, and a React GUI. That architecture was deprecated on 2026-08-22 in favor of the plugin model — Claude Code natively provides the execution loop the server used to implement. The server-era code is preserved at tag [`v2.0-server-final`](https://github.com/Cortexa-LLC/ai-pack/releases/tag/v2.0-server-final).
