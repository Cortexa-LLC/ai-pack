#!/usr/bin/env python3
"""Score a reviewer AUDIT MODE run against the planted-defect manifest (US-201).

Reads defects.yaml (via a minimal hand-rolled YAML-subset parser -- no pyyaml)
and a review output file. A defect counts as FOUND when any of its
detection_patterns matches the review text (case-insensitive regex).

Exit status: 0 when found-count >= threshold, 1 otherwise.
"""

import argparse
import os
import re
import sys

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))


def _unquote(value):
    value = value.strip()
    if len(value) >= 2 and value[0] == value[-1] and value[0] in ('"', "'"):
        return value[1:-1]
    return value


def parse_defects(path):
    """Parse the constrained YAML subset documented in defects.yaml.

    Supported shape: a top-level `defects:` list of flat mappings whose only
    nested value is the `detection_patterns` string list.
    """
    defects = []
    current = None
    in_patterns = False
    with open(path, encoding="utf-8") as fh:
        for raw in fh:
            stripped = raw.strip()
            if not stripped or stripped.startswith("#") or stripped == "defects:":
                continue
            if stripped.startswith("- id:"):
                current = {
                    "id": _unquote(stripped.split(":", 1)[1]),
                    "detection_patterns": [],
                }
                defects.append(current)
                in_patterns = False
            elif stripped.startswith("- ") and in_patterns and current is not None:
                current["detection_patterns"].append(_unquote(stripped[2:]))
            elif ":" in stripped and current is not None:
                key, _, value = stripped.partition(":")
                key = key.strip()
                if key == "detection_patterns":
                    in_patterns = True
                else:
                    current[key] = _unquote(value)
                    in_patterns = False
    if not defects:
        raise ValueError(f"no defects parsed from {path}")
    for d in defects:
        missing = [k for k in ("id", "file", "technique", "description") if k not in d]
        if missing or not d["detection_patterns"]:
            raise ValueError(f"defect {d.get('id', '?')} incomplete: missing {missing or 'patterns'}")
    return defects


def first_match(defect, text):
    """Return the first detection pattern that matches text, or None."""
    for pattern in defect["detection_patterns"]:
        if re.search(pattern, text, re.IGNORECASE):
            return pattern
    return None


def print_descriptions(defects):
    """Emit a synthetic review built from the manifest's own descriptions.

    Used by run-harness.sh --dry-run to exercise the scoring path without a
    claude invocation. Each description contains at least one of its own
    detection patterns, so this synthetic review must score N/N.
    """
    print("## Review Summary (synthetic dry-run output)")
    for d in defects:
        print(f"- [{d['file']}] {d['description']}")


def main():
    parser = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter
    )
    parser.add_argument(
        "--defects",
        default=os.path.join(SCRIPT_DIR, "defects.yaml"),
        help="path to the defect manifest (default: defects.yaml beside this script)",
    )
    parser.add_argument("--review", help="path to the captured review output")
    parser.add_argument(
        "--threshold",
        type=int,
        default=4,
        help="minimum found-count for exit 0 (default: 4; PRD 80%% target is 5)",
    )
    parser.add_argument(
        "--require-all",
        action="store_true",
        help="require every defect to be found (threshold = defect count); "
        "overrides --threshold; used by the dry-run self-test",
    )
    parser.add_argument(
        "--print-descriptions",
        action="store_true",
        help="print a synthetic review built from the manifest descriptions and exit",
    )
    args = parser.parse_args()

    defects = parse_defects(args.defects)

    threshold = len(defects) if args.require_all else args.threshold
    if not 1 <= threshold <= len(defects):
        parser.error(
            f"--threshold must be between 1 and {len(defects)} "
            f"(the defect count), got {threshold}"
        )

    if args.print_descriptions:
        print_descriptions(defects)
        return 0

    if not args.review:
        parser.error("--review is required unless --print-descriptions is given")
    with open(args.review, encoding="utf-8") as fh:
        text = fh.read()

    found = 0
    rows = []
    for d in defects:
        pattern = first_match(d, text)
        if pattern is not None:
            found += 1
        rows.append((d["id"], d["technique"], "FOUND" if pattern else "MISSED", pattern or "-"))

    id_w = max(len(r[0]) for r in rows + [("ID",)])
    tech_w = max(len(r[1]) for r in rows + [("", "TECHNIQUE")])
    print(f"{'ID':<{id_w}}  {'TECHNIQUE':<{tech_w}}  {'RESULT':<6}  MATCHED PATTERN")
    for rid, tech, result, pattern in rows:
        print(f"{rid:<{id_w}}  {tech:<{tech_w}}  {result:<6}  {pattern}")

    total = len(defects)
    pct = round(100 * found / total)
    print(f"SCORE: {found}/{total} ({pct}%)")
    return 0 if found >= threshold else 1


if __name__ == "__main__":
    sys.exit(main())
