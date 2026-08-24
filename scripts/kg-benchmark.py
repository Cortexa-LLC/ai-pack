#!/usr/bin/env python3
"""Read-only US-101 verification benchmark for the ai-pack knowledge graph.

Implements the ADR-009 Decision 2 benchmark (docs/adr/009-kg-health.md): runs the
seven retired-component queries through the exact agent-facing search semantics
(`search_knowledge` -> kglib KeywordSearch: tokenized OR CONTAINS over entity name
and observation content, ORDER BY e.updated_at DESC LIMIT 10, top-3 observations
per entity by created_at DESC), then scans everything an agent would actually see
(entity name + top-3 observation contents) for retired-component patterns.

Two pattern tiers, per the ADR's acceptance metric:

  - FLAGGED (broad mention patterns): reported for before/after comparison with
    the 64% baseline. Mentions are not failures — "Remove stale Beads references"
    is a legitimate current-era record.
  - DIRECTIVE (invocation-shaped patterns — tool names, CLI commands, ports):
    a result FAILS if a directive pattern is surfaced without the [OBSOLETE]
    marker. Target: 0. Exit code is non-zero on any failure.

Requirements: python3, kuzu==0.11.3 (pinned to the kg binary's storage format).
Opens the DB read-only, so it is safe to run alongside kg servers.
"""

import argparse
import os
import re
import subprocess
import sys

QUERIES = [
    "task tracking",
    "agent spawn",
    "orchestrate",
    "agent create",
    "agent-mcp",
    "beads",
    "performance grade",
]

# Broad mention patterns — the ADR-009 baseline measurement regex.
FLAGGED = re.compile(
    r"agent-mcp|mcp__agent|agent create|agent CLI|port 8082|\b8082\b|beads|"
    r"\bbd init\b|\bbd create\b|performance.grade|seed-grades|launchd|agent[- ]server",
    re.IGNORECASE,
)

# Invocation-shaped patterns: content that *directs* use of a retired component.
DIRECTIVE = re.compile(
    r"mcp__agent[-_]?mcp__\w+"
    r"|\bagent (create|list|show|close|update|reviewer|engineer|pr-shepherd|architect|inspector|spelunker|orchestrator)\b"
    r"|\bbd (init|create|update|flush|block|unblock|dep)\b"
    r"|\bport 8082\b|127\.0\.0\.1:8082|localhost:8082"
    r"|seed-grades\.py"
    r"|launchctl|launchd plist"
    r"|/usr/local/bin/agent-mcp",
    re.IGNORECASE,
)

# A directive match is downgraded to a mention when the immediately preceding
# context marks it as something being removed or called stale — the ADR's
# example: a "Remove stale Beads references" completion is acceptable history.
MENTION_CONTEXT = re.compile(r"stale|remov|delet|replac|retir|obsolete", re.IGNORECASE)
MENTION_WINDOW = 60

PROJECT_ID = "ai-pack"


def directive_hit(text):
    """Return the first directive match in text that is not a removal mention."""
    for m in DIRECTIVE.finditer(text or ""):
        before = text[max(0, m.start() - MENTION_WINDOW):m.start()]
        if not MENTION_CONTEXT.search(before):
            return m
    return None


def keyword_search(conn, query, limit=10):
    """Faithful replica of kglib KeywordSearch (mcp/src/kglib/search.go:39-145)."""
    tokens = query.lower().split()
    params = {"project_id": PROJECT_ID}
    name_conds, obs_conds = [], []
    for i, tok in enumerate(tokens):
        params[f"t{i}"] = tok
        name_conds.append(f"lower(e.name) CONTAINS $t{i}")
        obs_conds.append(f"lower(o.content) CONTAINS $t{i}")
    # kglib expresses this as one query with a correlated EXISTS subquery; the
    # Python binder rejects that form, so the same predicate (name match OR any
    # observation match) is computed as two queries unioned before the identical
    # ORDER BY e.updated_at DESC LIMIT.
    by_name = f"""
        MATCH (e:Entity)
        WHERE e.project_id = $project_id AND ({' OR '.join(name_conds)})
        RETURN DISTINCT e.id, e.name, e.type, e.updated_at
    """
    by_obs = f"""
        MATCH (e:Entity)-[:HAS_OBSERVATION]->(o:Observation)
        WHERE e.project_id = $project_id AND ({' OR '.join(obs_conds)})
        RETURN DISTINCT e.id, e.name, e.type, e.updated_at
    """
    matched = {}
    for cypher in (by_name, by_obs):
        result = conn.execute(cypher, params)
        while result.has_next():
            eid, name, etype, updated = result.get_next()
            matched[eid] = (eid, name, etype, updated)
    entities = sorted(matched.values(), key=lambda r: r[3], reverse=True)[:limit]
    entities = [(eid, name, etype) for eid, name, etype, _ in entities]
    out = []
    for eid, name, etype in entities:
        obs_result = conn.execute(
            "MATCH (o:Observation) WHERE o.entity_id = $id "
            "RETURN o.content ORDER BY o.created_at DESC LIMIT 3",
            {"id": eid},
        )
        obs = []
        while obs_result.has_next():
            obs.append(obs_result.get_next()[0])
        out.append((name, etype, obs))
    return out


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--db", default=None, help="path to knowledge.db (default: <git toplevel>/.ai/knowledge.db)")
    parser.add_argument("--verbose", action="store_true", help="print every surfaced result")
    args = parser.parse_args()

    import kuzu  # deferred so --help works without the pinned dep

    db_path = args.db
    if db_path is None:
        root = subprocess.run(
            ["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True
        ).stdout.strip()
        db_path = os.path.join(root, ".ai", "knowledge.db")
    if not os.path.exists(db_path):
        sys.exit(f"ABORT: no knowledge.db at {db_path}")

    db = kuzu.Database(db_path, read_only=True)
    conn = kuzu.Connection(db)

    total_results = total_flagged = 0
    failures = []
    print(f"{'query':<20} {'results':>7} {'flagged':>7} {'directive-unmarked':>18}")
    for query in QUERIES:
        results = keyword_search(conn, query)
        flagged = 0
        query_failures = []
        for name, etype, obs in results:
            surfaced = [name] + obs
            if any(FLAGGED.search(s or "") for s in surfaced):
                flagged += 1
            if directive_hit(name):
                query_failures.append((name, etype, f"entity name: {name!r}"))
            for content in obs:
                if content.startswith("[OBSOLETE"):
                    continue
                m = directive_hit(content)
                if m:
                    query_failures.append(
                        (name, etype, f"obs matches {m.group(0)!r}: {content[:120]!r}")
                    )
            if args.verbose:
                print(f"    [{etype}] {name}")
        print(f"{query:<20} {len(results):>7} {flagged:>7} {len(query_failures):>18}")
        total_results += len(results)
        total_flagged += flagged
        failures.extend((query, *f) for f in query_failures)

    pct = (100 * total_flagged // total_results) if total_results else 0
    print(f"\nTotals: {total_results} results, {total_flagged} flagged mentions ({pct}%), "
          f"{len(failures)} unmarked directive hits (target 0)")
    if failures:
        print("\nFAILURES — retired-component directives surfaced without [OBSOLETE] marker:")
        for query, name, etype, detail in failures:
            print(f"  query {query!r} -> [{etype}] {name}: {detail}")
        sys.exit(1)
    print("PASS: no unmarked retired-component directives surfaced.")


if __name__ == "__main__":
    main()
