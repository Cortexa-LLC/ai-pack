#!/usr/bin/env python3
"""
Seed performance grade files for known models at Grade C (cold-start default).

LiveBench coding scores are NOT used. Historical evidence showed that LiveBench
benchmark scores do not predict agentic editing capability: models with LiveBench
Grade A corrupted files on multi-file Go refactoring tasks.

Grade policy:
  A  — earned through real task outcomes (≥5 successes, error_rate < 0.1)
  B  — not used for cold-start; reserved for models with partial real data
  C  — cold-start default for all unseeded/unproven models (selector avoids)
  D  — model has demonstrated failures; excluded from selection
  F  — model has catastrophic failure history; permanently excluded

This script seeds Grade C for any model × role combination that has no existing
grade file. It never overwrites a file that has real task data (total_attempts > 0)
and never changes a D or F grade.

Run this when adding a new model to the known model list below.

Usage:
    python3 scripts/seed-grades.py [--dry-run] [--grades-dir DIR]
"""

import argparse
import json
import os
from datetime import datetime, timezone
from pathlib import Path

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

_DATA_DIR = Path(os.environ.get("AGENT_DATA_DIR", Path.home() / ".claude"))
DEFAULT_GRADES_DIR = _DATA_DIR / "performance_grades"

PROJECT_ROOT = Path(__file__).resolve().parent.parent
PROJECT_ID = str(PROJECT_ROOT)

# Roles that get grade files
ROLES = [
    "engineer", "tester", "reviewer", "orchestrator", "architect",
    "inspector", "spelunker", "cartographer", "archaeologist",
    "product-manager", "designer", "strategist",
]

# Known model IDs. Add new models here when they become available.
# All models start at Grade C until proven through real task outcomes.
KNOWN_MODELS = [
    "claude-haiku-4-5",
    "claude-sonnet-4-5",
    "claude-sonnet-4-6",
    "claude-opus-4-5",
    "claude-opus-4-6",
    "gemini-2.5-flash-lite",
    "gemini-2.5-flash",
    "gemini-2.5-pro",
    "gpt-4.1-nano",
    "gpt-4.1-mini",
    "gpt-4.1",
    "gpt-4o-mini",
    "gpt-5.1-codex-mini",
    "gpt-5.1-codex",
    "gpt-5.2-codex",
    "o4-mini",
]

# Grades that represent real failure history — never overwrite these.
PROTECTED_GRADES = {"D", "F"}

# Cold-start values — real task outcomes will override quickly.
COLD_START_GRADE = "C"
COLD_START_CONFIDENCE = 0.05


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
def main():
    parser = argparse.ArgumentParser(
        description=__doc__,
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument("--dry-run", action="store_true",
                        help="Print what would be written without touching disk")
    parser.add_argument("--grades-dir", default=str(DEFAULT_GRADES_DIR),
                        help=f"Grade directory (default: {DEFAULT_GRADES_DIR})")
    args = parser.parse_args()

    grades_dir = Path(args.grades_dir)
    if not args.dry_run:
        grades_dir.mkdir(parents=True, exist_ok=True)

    now = datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")
    written = skipped_real = skipped_protected = 0

    for model_id in sorted(KNOWN_MODELS):
        for role in ROLES:
            filename = f"{model_id}_{role}_{PROJECT_ID.replace('/', '_')}.json"
            path = grades_dir / filename

            # Check existing file
            if path.exists():
                try:
                    existing = json.loads(path.read_text())
                except Exception:
                    existing = {}

                grade = existing.get("grade", "")
                attempts = existing.get("total_attempts", 0)

                if grade in PROTECTED_GRADES:
                    skipped_protected += 1
                    continue  # never overwrite D/F

                if attempts > 0:
                    skipped_real += 1
                    continue  # never overwrite real task data

            # Write cold-start Grade C
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
                "grade": COLD_START_GRADE,
                "confidence_score": COLD_START_CONFIDENCE,
                "last_updated": now,
                "sample_size": 0,
                "source": "cold_start",
            }

            if args.dry_run:
                print(f"  [dry-run] {model_id:<30} {role:<20} grade=C")
            else:
                path.write_text(json.dumps(payload, indent=4))
            written += 1

    action = "[dry-run] Would write" if args.dry_run else "Wrote"
    print(f"{action} {written} grade files (Grade C cold-start)")
    print(f"Skipped {skipped_real} with real task data, {skipped_protected} with D/F grades")


if __name__ == "__main__":
    main()
