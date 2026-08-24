#!/usr/bin/env python3
"""One-off curation of server-era knowledge in the ai-pack knowledge graph.

Implements ADR-009 Decision 2 (docs/adr/009-kg-health.md, US-101): the graph still
contains guidance for the retired agent-server era (agent CLI, agent-mcp, port 8082,
Beads/bd, performance-grade seeding, launchd). KG-first agents retrieve it as if
current. This script:

  1. ARCHIVES first (non-optional on --apply): `kg export` JSONL to .ai/archives/
     plus a dated copy of the whole knowledge.db directory. Nothing is destroyed
     unrecoverably.
  2. DELETES pure server-era operational noise (no guidance or historical value
     beyond the archive):
       - all `task_outcome` entities
       - all `exec:*` execution-record topics (written by the retired server's
         KG write-back; nothing writes these anymore)
       - server-era task-completion topics (`ai-pack-XXX completion`) whose content
         directs use of retired components
       - dead-code entities: `file` entities whose path no longer exists in the
         tree, plus `function`/`type`/`import` entities related only to such files,
         plus observation-less doc topics contained only in such files, plus an
         explicit dead-code name list (e.g. ensureBeadsForProject)
  3. MARKS genuinely historical records (the deprecation inventory and the
     investigations that explain why the pivot happened) by prefixing each
     observation with the [OBSOLETE] marker. Raw `SET o.content` is used because
     kglib has no update API and delete+re-add would falsify created_at.
     Idempotent: rows already carrying the prefix are skipped.
  4. Keeps everything else unmarked — including the deprecation decision record
     and all current-era knowledge that merely *mentions* retired components.

Defaults to --dry-run (prints every planned mutation). --apply executes.

HAZARD (ADR-009): never use `kg index` / index_project as a cleanup mechanism.
It calls clearProjectData unconditionally and wipes ALL project knowledge,
including agent-written topics and observations (Cortexa-LLC/mcp#2).

Requirements: python3, kuzu==0.11.3 (pinned — matches the kg binary's embedded
Kuzu storage format). Run with no `kg server` process open on this project
(KuzuDB is single-writer); --allow-running-server overrides the check for
environments where the servers only hold the DB open per-request.

Verification: scripts/kg-benchmark.py (read-only) is the repeatable US-101 gate.
"""

import argparse
import datetime
import os
import re
import shutil
import subprocess
import sys

OBSOLETE_PREFIX = (
    "[OBSOLETE — server era, retired 2026-08-22, see tag v2.0-server-final] "
)

# Broad retired-component pattern (same family as the ADR-009 baseline measurement).
# Used only to classify server-era completion topics for deletion.
RETIRED_CONTENT = re.compile(
    r"agent-mcp|mcp__agent|agent create|agent CLI|port 8082|\b8082\b|beads|"
    r"\bbd init\b|\bbd create\b|performance.grade|seed-grades|launchd|agent[- ]server|"
    r"\bagent (pr-shepherd|engineer|reviewer|architect|inspector|spelunker|orchestrator|"
    r"update|list|show|close)\b",
    re.IGNORECASE,
)

# Server-era task-completion topic names (the retired task DB's ID scheme).
COMPLETION_NAME = re.compile(r"^ai-pack-[a-z0-9]+ completion$", re.IGNORECASE)

# Genuinely historical records: every observation gets the [OBSOLETE] prefix.
# The deprecation *decision* record and current-era knowledge stay unmarked.
MARK_ALL_ENTITIES = [
    "agent-server deprecation inventory",
    "GUI server-coupling investigation",
    "plugin layer state assessment investigation",
]

# Entities where only observations matching a pattern are historical.
MARK_OBS_PATTERNS = {
    "kg-standalone-extraction": re.compile(r"agent[- ]server", re.IGNORECASE),
}

# Dead code known by name even if not linked to a dead file entity.
DEAD_CODE_NAMES = ["ensureBeadsForProject"]

PROJECT_ID = "ai-pack"


def rows(conn, cypher, params=None):
    result = conn.execute(cypher, params or {})
    out = []
    while result.has_next():
        out.append(result.get_next())
    return out


def build_plan(conn, repo_root):
    """Return (deletes, marks): entity delete list and observation mark list."""
    deletes = []  # (entity_id, name, type, reason, obs_count)
    marks = []  # (obs_id, entity_name, old_content)
    seen = set()

    obs_counts = dict(
        rows(
            conn,
            "MATCH (e:Entity)-[:HAS_OBSERVATION]->(o:Observation) "
            "RETURN e.id, count(o)",
        )
    )

    def add_delete(eid, name, etype, reason):
        if eid in seen:
            return
        seen.add(eid)
        deletes.append((eid, name, etype, reason, obs_counts.get(eid, 0)))

    # 1. task_outcome entities (server-era operational noise).
    for eid, name, etype in rows(
        conn, "MATCH (e:Entity) WHERE e.type = 'task_outcome' RETURN e.id, e.name, e.type"
    ):
        add_delete(eid, name, etype, "task_outcome noise")

    # 2. exec:* execution-record topics.
    for eid, name, etype in rows(
        conn,
        "MATCH (e:Entity) WHERE e.name STARTS WITH 'exec:' RETURN e.id, e.name, e.type",
    ):
        add_delete(eid, name, etype, "exec:* server execution record")

    # 3. Server-era completion topics whose content directs retired components.
    for eid, name, content in rows(
        conn,
        "MATCH (e:Entity)-[:HAS_OBSERVATION]->(o:Observation) "
        "WHERE e.type = 'topic' RETURN e.id, e.name, o.content",
    ):
        if COMPLETION_NAME.match(name or "") and RETIRED_CONTENT.search(content or ""):
            add_delete(eid, name, "topic", "server-era completion directing retired components")

    # 4. Dead file entities: source no longer exists in the tree.
    files = rows(conn, "MATCH (e:Entity) WHERE e.type = 'file' RETURN e.id, e.name")
    missing = {
        eid for eid, name in files if not os.path.exists(os.path.join(repo_root, name))
    }
    for eid, name in files:
        if eid in missing:
            add_delete(eid, name, "file", "dead file (not in tree)")

    # 5. Symbols related only to dead files.
    linked_missing, linked_present = {}, set()
    for fid, xid in rows(
        conn,
        "MATCH (f:Entity)-[r]-(x:Entity) WHERE f.type = 'file' "
        "AND x.type IN ['function', 'type', 'import'] RETURN f.id, x.id",
    ):
        if fid in missing:
            linked_missing.setdefault(xid, True)
        else:
            linked_present.add(xid)
    dead_syms = set(linked_missing) - linked_present
    for eid, name, etype in rows(
        conn,
        "MATCH (e:Entity) WHERE e.type IN ['function', 'type', 'import'] "
        "RETURN e.id, e.name, e.type",
    ):
        if eid in dead_syms:
            add_delete(eid, name, etype, "dead symbol (source file not in tree)")

    # 6. Observation-less doc topics contained only in dead files.
    topic_missing, topic_present = set(), set()
    for fid, tid in rows(
        conn,
        "MATCH (f:Entity)-[r]-(t:Entity) WHERE f.type = 'file' AND t.type = 'topic' "
        "RETURN f.id, t.id",
    ):
        if fid in missing:
            topic_missing.add(tid)
        else:
            topic_present.add(tid)
    dead_topics = topic_missing - topic_present
    for eid, name in rows(
        conn, "MATCH (e:Entity) WHERE e.type = 'topic' RETURN e.id, e.name"
    ):
        if eid in dead_topics and obs_counts.get(eid, 0) == 0:
            add_delete(eid, name, "topic", "doc topic of dead file (no observations)")

    # 7. Explicit dead-code names.
    for name in DEAD_CODE_NAMES:
        for eid, etype in rows(
            conn,
            "MATCH (e:Entity) WHERE e.name = $n RETURN e.id, e.type",
            {"n": name},
        ):
            add_delete(eid, name, etype, "dead code (explicit list)")

    # 8. Marks: historical records get the [OBSOLETE] prefix (idempotent).
    deleted_ids = {d[0] for d in deletes}
    for ent_name in MARK_ALL_ENTITIES:
        for oid, content, eid in rows(
            conn,
            "MATCH (e:Entity {name: $n})-[:HAS_OBSERVATION]->(o:Observation) "
            "RETURN o.id, o.content, e.id",
            {"n": ent_name},
        ):
            if eid not in deleted_ids and not content.startswith("[OBSOLETE"):
                marks.append((oid, ent_name, content))
    for ent_name, pattern in MARK_OBS_PATTERNS.items():
        for oid, content, eid in rows(
            conn,
            "MATCH (e:Entity {name: $n})-[:HAS_OBSERVATION]->(o:Observation) "
            "RETURN o.id, o.content, e.id",
            {"n": ent_name},
        ):
            if (
                eid not in deleted_ids
                and pattern.search(content or "")
                and not content.startswith("[OBSOLETE")
            ):
                marks.append((oid, ent_name, content))

    return deletes, marks


def archive(db_path, repo_root):
    stamp = datetime.date.today().strftime("%Y%m%d")
    archives = os.path.join(os.path.dirname(db_path), "archives")
    os.makedirs(archives, exist_ok=True)
    export_path = os.path.join(archives, f"kg-export-pre-curation-{stamp}.jsonl")
    if not os.path.exists(export_path):
        try:
            subprocess.run(
                ["kg", "export", "-o", export_path], check=True, cwd=repo_root
            )
        except (OSError, subprocess.CalledProcessError) as exc:
            sys.exit(f"ABORT: mandatory `kg export` archive failed: {exc}")
    backup = f"{db_path}.bak-{stamp}"
    if not os.path.exists(backup):
        if os.path.isdir(db_path):
            shutil.copytree(db_path, backup)
        else:
            shutil.copy2(db_path, backup)
        wal = db_path + ".wal"
        if os.path.isfile(wal):
            shutil.copy2(wal, backup + ".wal")
    print(f"Archived: {export_path}\nBackup:   {backup}")


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--db", default=None, help="path to knowledge.db (default: <repo-root>/.ai/knowledge.db)")
    parser.add_argument("--repo-root", default=None, help="tree to check file existence against (default: git toplevel)")
    parser.add_argument("--apply", action="store_true", help="execute mutations (default: dry run)")
    parser.add_argument("--allow-running-server", action="store_true",
                        help="proceed despite a running `kg server` process")
    args = parser.parse_args()

    import kuzu  # deferred so --help works without the pinned dep

    repo_root = args.repo_root or subprocess.run(
        ["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True
    ).stdout.strip()
    db_path = args.db or os.path.join(repo_root, ".ai", "knowledge.db")
    if not os.path.exists(db_path):
        sys.exit(f"ABORT: no knowledge.db at {db_path}")

    check = subprocess.run(["pgrep", "-f", "kg server"], capture_output=True)
    if check.returncode == 0 and args.apply and not args.allow_running_server:
        sys.exit(
            "ABORT: `kg server` processes are running (KuzuDB is single-writer).\n"
            "Stop them, or pass --allow-running-server if they only open the DB "
            "per-request.\nPIDs: " + ", ".join(check.stdout.decode().split())
        )

    if args.apply:
        archive(db_path, repo_root)

    db = kuzu.Database(db_path, read_only=not args.apply)
    conn = kuzu.Connection(db)
    deletes, marks = build_plan(conn, repo_root)

    print(f"\n=== Curation plan ({'APPLY' if args.apply else 'DRY RUN'}) ===")
    by_reason = {}
    for eid, name, etype, reason, nobs in deletes:
        by_reason.setdefault(reason, []).append((name, etype, nobs))
    for reason, items in sorted(by_reason.items()):
        print(f"\nDELETE — {reason}: {len(items)} entities "
              f"({sum(n for _, _, n in items)} observations)")
        for name, etype, nobs in sorted(items)[:1000]:
            print(f"  - [{etype}] {name}" + (f" ({nobs} obs)" if nobs else ""))
    print(f"\nMARK [OBSOLETE] — {len(marks)} observations:")
    for oid, ent_name, content in marks:
        print(f"  - {ent_name}: {content[:100]!r}...")
    print(f"\nTotals: delete {len(deletes)} entities "
          f"({sum(d[4] for d in deletes)} observations), mark {len(marks)} observations")

    if not args.apply:
        print("\nDry run only. Re-run with --apply to execute.")
        return

    for eid, name, etype, reason, nobs in deletes:
        conn.execute(
            "MATCH (e:Entity {id: $id})-[:HAS_OBSERVATION]->(o:Observation) "
            "DETACH DELETE o",
            {"id": eid},
        )
        conn.execute("MATCH (e:Entity {id: $id}) DETACH DELETE e", {"id": eid})
    for oid, ent_name, content in marks:
        conn.execute(
            "MATCH (o:Observation {id: $id}) SET o.content = $content",
            {"id": oid, "content": OBSOLETE_PREFIX + content},
        )
    db.close()
    print("\nApplied. Verify with scripts/kg-benchmark.py.")


if __name__ == "__main__":
    main()
