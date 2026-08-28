#!/usr/bin/env python3
"""SessionStart hook (all sources): tell the main session to consult the KG first.

The KG-first rule is written into every plugin/agents/*.md role, so it reaches
SUBAGENTS only. A main orchestrator session that does work inline — rather than
delegating it — never sees the rule, and the orchestrator is precisely the role
that reads the most raw file content. This hook closes that gap by injecting the
directive at session start.

Claude Code adds a SessionStart hook's plain-text stdout to the session context
(one of the three events where stdout is context, not just debug log). No matcher
is set in hooks.json, so this fires for every source — startup, resume, clear,
compact, fork — which matters most for `compact`: the directive must survive
compaction or it silently stops applying mid-session.

The text below is deliberately short. It is prepended to every session, so its
own token cost is charged on each one, and it only pays for itself by displacing
file reads that cost more.

Rule, as with the spawn logger: a hook must never break a session — exit 0 no
matter what, and print nothing on failure.
"""
import os
import sys

DIRECTIVE = """\
## ai-pack: consult the knowledge graph first

This project ships a knowledge graph (the `kg` MCP server). Before reading source
files to answer "how does X work", "why is this the way it is", or "what changed
here recently", call `kg__search_knowledge({query: "<component>"})`. It holds prior
review findings, empirical gotchas, and per-PR implementation records — one call
routinely replaces several file reads.

This applies to THIS session, not only to subagents you spawn.

- Use short, specific queries (1-3 words); tokens are OR-matched.
- Observations prefixed `[OBSOLETE]` are history, not guidance — they may explain
  why something changed, but never follow them.
- Entries are timestamped; when two conflict, prefer the newer one, and verify
  against the tree before acting on anything load-bearing.
- If a query returns nothing useful, fall back to reading files and move on — do
  not retry variations.
- Closing the loop is the other half: after landing meaningful work, record what
  changed and the gotchas found via `kg__add_entity` / `kg__add_observation`. The
  graph is only worth reading because previous sessions wrote to it."""

try:
    sys.stdout.write(DIRECTIVE + "\n")
    # Flush INSIDE the guard: a broken pipe surfaces here, where it is caught,
    # rather than during interpreter shutdown where it would print a traceback
    # to stderr. Losing the directive on a broken pipe is the correct outcome;
    # emitting noise into the session is not.
    sys.stdout.flush()
except Exception:
    pass  # never break the session over a directive

# _exit rather than sys.exit: skips atexit's flush of a possibly-broken stdout,
# which is the traceback path above. Safe only because the flush above is
# explicit — a bare os._exit() here would drop the buffered directive entirely.
os._exit(0)
