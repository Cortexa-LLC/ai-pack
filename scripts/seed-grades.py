#!/usr/bin/env python3
"""
Seed performance grades from LiveBench coding scores.

Fetches the latest LiveBench CSV, computes a coding score per model
(avg of code_completion, code_generation, python, typescript, javascript),
and writes grade files compatible with the Go GlobalGradeManager format.

Grade thresholds (livebench coding score out of 100):
  A  >=  60   — highly capable
  B  >=  45   — capable
  C  >=  30   — limited
  D   <  30   — not recommended

Models absent from LiveBench receive hardcoded estimates marked with
lower confidence so real task data overrides them quickly.

Usage:
    python3 scripts/seed-grades.py [--wipe] [--dry-run] [--grades-dir DIR]
    python3 scripts/seed-grades.py --wipe        # wipe + re-seed (default behavior)

Data source: https://livebench.ai/table_2026_01_08.csv
"""

import argparse
import csv
import io
import json
import os
import sys
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

# Multiple releases are needed because different models appear in different snapshots.
# Each entry: (url, [coding columns])
LIVEBENCH_SOURCES = [
    # Jan 2026: covers claude-opus-4-6, claude-sonnet-4-6, gpt-5.x-codex, gemini-2.5-*
    (
        "https://livebench.ai/table_2026_01_08.csv",
        ["code_completion", "code_generation", "python", "typescript", "javascript"],
    ),
    # Apr 2025: covers gpt-4.1, gpt-4.1-mini, gpt-4.1-nano, gpt-4o-mini, o4-mini
    (
        "https://livebench.ai/table_2025_04_02.csv",
        ["LCB_generation", "coding_completion"],
    ),
]

_DATA_DIR = Path(os.environ.get("AGENT_DATA_DIR", Path.home() / ".claude"))
DEFAULT_GRADES_DIR = _DATA_DIR / "performance_grades"

PROJECT_ROOT = Path(__file__).resolve().parent.parent
PROJECT_ID = str(PROJECT_ROOT)

# Roles that get grade files (mirrors DEFAULT_ROLES in old benchmark script)
ROLES = [
    "engineer", "tester", "reviewer", "orchestrator", "architect",
    "inspector", "spelunker", "cartographer", "archaeologist",
    "product-manager", "designer", "strategist",
]

# ---------------------------------------------------------------------------
# LiveBench model name → our model ID
# Pick base/medium-effort variants as most representative of default usage.
# ---------------------------------------------------------------------------
LIVEBENCH_TO_OURS = {
    # Jan 2026 release
    "claude-haiku-4-5-20251001":                        "claude-haiku-4-5",
    "gemini-2.5-flash-lite-preview-09-2025-highthinking": "gemini-2.5-flash-lite",
    "gemini-2.5-flash-preview-09-2025-highthinking":    "gemini-2.5-flash",
    "gpt-5.1-codex-mini":                               "gpt-5.1-codex-mini",
    "claude-sonnet-4-5-20250929":                       "claude-sonnet-4-5",
    "claude-sonnet-4-6-thinking-auto-medium-effort":    "claude-sonnet-4-6",
    "gemini-2.5-pro-06-05-highthinking":                "gemini-2.5-pro",
    "gpt-5.1-codex":                                    "gpt-5.1-codex",
    "gpt-5.2-codex":                                    "gpt-5.2-codex",
    "claude-opus-4-5-20251101-medium-effort":           "claude-opus-4-5",
    "claude-opus-4-6-thinking-auto-high-effort":        "claude-opus-4-6",
    # Apr 2025 release (models not present in Jan 2026)
    "gpt-4.1-2025-04-14":                              "gpt-4.1",
    "gpt-4.1-mini-2025-04-14":                         "gpt-4.1-mini",
    "gpt-4.1-nano-2025-04-14":                         "gpt-4.1-nano",
    "gpt-4o-mini-2024-07-18":                          "gpt-4o-mini",
    "o4-mini-2025-04-16-medium":                       "o4-mini",
}

# Models absent from LiveBench Jan 2026 (gpt-4o-mini, gpt-4.1-nano, gpt-4.1-mini,
# o4-mini, gpt-4.1) are intentionally left ungraded. The selector falls back to
# cheapest-in-tier for ungraded models, which is preferable to estimates that may
# not reflect real agentic editing capability.
ESTIMATES: dict[str, float] = {}

# ---------------------------------------------------------------------------
# Grade thresholds
# ---------------------------------------------------------------------------
def score_to_grade(score: float) -> str:
    if score >= 60.0:
        return "A"
    if score >= 45.0:
        return "B"
    if score >= 30.0:
        return "C"
    return "D"


# ---------------------------------------------------------------------------
# Fetch + parse LiveBench CSV
# ---------------------------------------------------------------------------
def fetch_livebench_scores() -> dict[str, float]:
    """Fetches multiple LiveBench CSV releases and returns {our_model_id: coding_score}.
    Later sources do not overwrite earlier ones (Jan 2026 scores take priority)."""
    scores: dict[str, float] = {}
    for url, coding_cols in LIVEBENCH_SOURCES:
        print(f"Fetching {url} ...")
        with urllib.request.urlopen(url, timeout=30) as resp:
            data = resp.read().decode("utf-8")
        reader = csv.DictReader(io.StringIO(data))
        for row in reader:
            lb_name = row["model"]
            if lb_name not in LIVEBENCH_TO_OURS:
                continue
            our_id = LIVEBENCH_TO_OURS[lb_name]
            if our_id in scores:
                continue  # earlier (newer) source takes priority
            vals = [float(row[c]) for c in coding_cols if row.get(c, "").strip()]
            if not vals:
                continue
            scores[our_id] = round(sum(vals) / len(vals), 1)
    return scores


# ---------------------------------------------------------------------------
# Write grade file
# ---------------------------------------------------------------------------
def write_grade(grades_dir: Path, model_id: str, role: str,
                score: float, source: str, confidence: float, dry_run: bool):
    grade = score_to_grade(score)
    filename = f"{model_id}_{role}_{PROJECT_ID.replace('/', '_')}.json"
    path = grades_dir / filename

    payload = {
        "model_id": model_id,
        "role_id": role,
        "project_id": PROJECT_ID,
        "total_attempts": 0,
        "successes": 0,
        "failures": 0,
        "retries": 0,
        "total_tokens_used": 0,
        "total_execution_time_ms": 0,
        "average_tokens": 0,
        "average_execution_time": 0,
        "error_rate": 0,
        "retry_rate": 0,
        "success_rate": 0,
        "escalation_count": 0,
        "downgrade_count": 0,
        "grade": grade,
        "confidence_score": confidence,
        "last_updated": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
        "sample_size": 0,
        "source": source,
        "livebench_coding_score": score,
    }

    if dry_run:
        print(f"  [dry-run] {model_id:<30} {role:<20} score={score:>5.1f}  grade={grade}")
        return

    path.write_text(json.dumps(payload, indent=4))


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
def main():
    parser = argparse.ArgumentParser(description=__doc__,
                                     formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--wipe", action="store_true", default=True,
                        help="Delete existing grade files before seeding (default: on)")
    parser.add_argument("--no-wipe", dest="wipe", action="store_false",
                        help="Preserve existing grade files (only add missing)")
    parser.add_argument("--dry-run", action="store_true",
                        help="Print what would be written without touching disk")
    parser.add_argument("--grades-dir", default=str(DEFAULT_GRADES_DIR),
                        help=f"Grade directory (default: {DEFAULT_GRADES_DIR})")
    args = parser.parse_args()

    grades_dir = Path(args.grades_dir)

    if not args.dry_run:
        grades_dir.mkdir(parents=True, exist_ok=True)

    # Wipe existing files
    if args.wipe and not args.dry_run:
        existing = list(grades_dir.glob("*.json"))
        for f in existing:
            f.unlink()
        print(f"Wiped {len(existing)} existing grade files from {grades_dir}")

    # Fetch live scores
    try:
        lb_scores = fetch_livebench_scores()
    except Exception as e:
        print(f"ERROR: Could not fetch LiveBench data: {e}", file=sys.stderr)
        sys.exit(1)

    print(f"\nLiveBench coding scores (from {len(LIVEBENCH_SOURCES)} releases):")
    for model_id, score in sorted(lb_scores.items(), key=lambda x: -x[1]):
        print(f"  {model_id:<30} {score:>5.1f}  grade={score_to_grade(score)}")

    # Combine: LiveBench scores take precedence
    all_scores: dict[str, tuple[float, str, float]] = {}  # model_id → (score, source, confidence)
    for model_id, score in lb_scores.items():
        all_scores[model_id] = (score, "livebench_2026_01_08", 0.8)
    for model_id, score in ESTIMATES.items():
        if model_id not in all_scores:
            all_scores[model_id] = (score, "livebench_estimate", 0.3)

    # Write grade files
    total = 0
    print(f"\nWriting grade files to {grades_dir} ...")
    for model_id, (score, source, confidence) in sorted(all_scores.items()):
        for role in ROLES:
            write_grade(grades_dir, model_id, role, score, source, confidence, args.dry_run)
            total += 1

    print(f"\n{'[dry-run] Would write' if args.dry_run else 'Wrote'} {total} grade files "
          f"({len(all_scores)} models × {len(ROLES)} roles)")


if __name__ == "__main__":
    main()
