#!/usr/bin/env python3
"""
Benchmark test suite across all models in ModelsByTier.

Sends a standardized coding/reasoning prompt to each model (OpenAI + Anthropic),
records latency, token usage, and output quality, then writes results into the
GlobalGradeManager JSON format so the Performance-by-Model tab shows real data.

Usage:
    python3 scripts/benchmark-models.py [--runs N] [--role ROLE] [--dry-run]
"""

import os
import sys
import json
import time
import math
import datetime
import argparse
import re
from pathlib import Path
from typing import Optional

# Suppress deprecation warnings from third-party libraries
import warnings
warnings.filterwarnings("ignore", category=DeprecationWarning)

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

PROJECT_ROOT = Path(__file__).resolve().parent.parent
GRADES_DIR = PROJECT_ROOT / ".claude" / "performance_grades"
PROJECT_ID = str(PROJECT_ROOT)       # matches what the Go agent records

# Roles to benchmark (engineer is primary; add more to broaden dataset)
DEFAULT_ROLES = ["engineer", "tester", "reviewer"]

# Number of benchmark runs per model per role
DEFAULT_RUNS = 5

# --------------------------------------------------------------------------
# Models – mirrors monitoring/model_selector.go ModelsByTier
# --------------------------------------------------------------------------

# Each entry: (id, provider, tier, api_model_name)
# api_model_name is the string sent to the API (may differ from the registry ID)
MODELS = [
    # TierMinimal
    ("gpt-4o-mini",     "openai",     "minimal", "gpt-4o-mini"),
    ("gpt-4.1-nano",    "openai",     "minimal", "gpt-4.1-nano"),
    ("claude-haiku-4-5","anthropic",  "minimal", "claude-haiku-4-5"),

    # TierLow
    ("gpt-4.1-mini",    "openai",     "low",     "gpt-4.1-mini"),
    ("o4-mini",         "openai",     "low",     "o4-mini"),

    # TierMedium
    ("gpt-5.1-codex-mini","openai",   "medium",  "gpt-5.1-codex-mini"),
    ("gpt-4.1",         "openai",     "medium",  "gpt-4.1"),
    ("claude-sonnet-4-6","anthropic", "medium",  "claude-sonnet-4-6"),
    ("claude-sonnet-4-5","anthropic", "medium",  "claude-sonnet-4-5-20250929"),
    ("claude-sonnet-4-5-20250929","anthropic","medium","claude-sonnet-4-5-20250929"),

    # TierHigh
    ("gpt-5.1-codex",   "openai",     "high",    "gpt-5.1-codex"),
    ("gpt-5.2-codex",   "openai",     "high",    "gpt-5.2-codex"),
    ("claude-opus-4-5", "anthropic",  "high",    "claude-opus-4-5-20251101"),
    ("claude-opus-4-6", "anthropic",  "high",    "claude-opus-4-6"),
]

# ---------------------------------------------------------------------------
# Benchmark prompts – varied so repeated calls exercise real reasoning
# ---------------------------------------------------------------------------

BENCHMARK_PROMPTS = [
    {
        "system": (
            "You are a senior software engineer. "
            "Respond with concise, correct code only."
        ),
        "user": (
            "Implement a Python function `sieve_of_eratosthenes(n: int) -> list[int]` "
            "that returns all prime numbers up to n using the Sieve of Eratosthenes. "
            "Include type hints and a brief docstring. Do not include a main block."
        ),
    },
    {
        "system": (
            "You are a senior software engineer. "
            "Respond with concise, correct code only."
        ),
        "user": (
            "Implement a Python function `binary_search(arr: list, target) -> int` "
            "that returns the index of target in sorted arr, or -1 if not found. "
            "Include type hints and a brief docstring."
        ),
    },
    {
        "system": (
            "You are a senior software engineer. "
            "Respond with concise, correct code only."
        ),
        "user": (
            "Implement a Python class `LRUCache` with methods `get(key)` and `put(key, value)` "
            "using O(1) average time complexity. Include type hints and a brief docstring."
        ),
    },
    {
        "system": (
            "You are a senior software engineer. "
            "Respond with concise, correct code only."
        ),
        "user": (
            "Write a Python function `deep_merge(base: dict, override: dict) -> dict` "
            "that recursively merges two dictionaries, with override values taking precedence. "
            "Include type hints and a brief docstring."
        ),
    },
    {
        "system": (
            "You are a senior software engineer. "
            "Respond with concise, correct code only."
        ),
        "user": (
            "Implement a Python function `topological_sort(graph: dict[str, list[str]]) -> list[str]` "
            "that performs a topological sort on a DAG represented as an adjacency list. "
            "Raise ValueError on cycles. Include type hints and a brief docstring."
        ),
    },
]

# ---------------------------------------------------------------------------
# Quality evaluator
# ---------------------------------------------------------------------------

def evaluate_quality(output: str, prompt_index: int) -> bool:
    """
    Returns True if the model produced a reasonable, code-containing response.
    Checks for Python function/class definition, minimum length, and absence of
    obvious error indicators.
    """
    if not output or len(output) < 80:
        return False

    # Must contain a function or class definition
    if not re.search(r'\b(def |class )\w+', output):
        return False

    # Must look like actual code (has colon-ending lines typical of Python)
    if output.count(':') < 2:
        return False

    # Reject if the model refused or produced only prose
    refusal_patterns = [
        r"i cannot", r"i'm unable", r"i am unable",
        r"as an ai", r"i don't have", r"i won't",
    ]
    lower = output.lower()
    if any(re.search(p, lower) for p in refusal_patterns):
        return False

    return True


# ---------------------------------------------------------------------------
# API wrappers
# ---------------------------------------------------------------------------

def call_openai(api_model: str, prompt: dict, timeout: int = 60) -> dict:
    """Call OpenAI and return timing + token data.

    Routes to the appropriate API endpoint based on model family:
    - codex models → v1/responses (Responses API)
    - o-series (o1/o3/o4) → chat completions with max_completion_tokens, no system
    - all others → chat completions standard
    """
    import openai

    client = openai.OpenAI(api_key=os.environ["OPENAI_API_KEY"])

    # Codex models only work via the Responses API
    is_codex = "codex" in api_model.lower()
    is_reasoning_model = api_model.startswith(("o1", "o3", "o4"))

    start = time.monotonic()
    try:
        if is_codex:
            user_input = prompt["user"]
            if prompt.get("system"):
                user_input = f"{prompt['system']}\n\n{prompt['user']}"
            resp = client.responses.create(
                model=api_model,
                input=user_input,
                max_output_tokens=1024,
            )
            elapsed_ms = (time.monotonic() - start) * 1000
            text = resp.output_text if hasattr(resp, "output_text") else ""
            tokens_in = resp.usage.input_tokens if resp.usage else 0
            tokens_out = resp.usage.output_tokens if resp.usage else 0
        else:
            messages = []
            if prompt.get("system"):
                messages.append({"role": "system", "content": prompt["system"]})
            messages.append({"role": "user", "content": prompt["user"]})

            if is_reasoning_model:
                messages = [m for m in messages if m["role"] != "system"]
                resp = client.chat.completions.create(
                    model=api_model,
                    messages=messages,
                    max_completion_tokens=1024,
                    timeout=timeout,
                )
            else:
                resp = client.chat.completions.create(
                    model=api_model,
                    messages=messages,
                    max_tokens=1024,
                    timeout=timeout,
                )
            elapsed_ms = (time.monotonic() - start) * 1000
            text = resp.choices[0].message.content or ""
            tokens_in = resp.usage.prompt_tokens if resp.usage else 0
            tokens_out = resp.usage.completion_tokens if resp.usage else 0

        return {
            "success": True,
            "text": text,
            "tokens_in": tokens_in,
            "tokens_out": tokens_out,
            "total_tokens": tokens_in + tokens_out,
            "elapsed_ms": elapsed_ms,
            "error": None,
        }
    except Exception as exc:
        elapsed_ms = (time.monotonic() - start) * 1000
        return {
            "success": False,
            "text": "",
            "tokens_in": 0,
            "tokens_out": 0,
            "total_tokens": 0,
            "elapsed_ms": elapsed_ms,
            "error": str(exc),
        }


def call_anthropic(api_model: str, prompt: dict, timeout: int = 60) -> dict:
    """Call Anthropic messages API and return timing + token data."""
    import anthropic

    client = anthropic.Anthropic(api_key=os.environ["ANTHROPIC_API_KEY"])

    messages = [{"role": "user", "content": prompt["user"]}]
    kwargs = dict(
        model=api_model,
        messages=messages,
        max_tokens=1024,
    )
    if prompt.get("system"):
        kwargs["system"] = prompt["system"]

    start = time.monotonic()
    try:
        resp = client.messages.create(**kwargs)  # type: ignore[arg-type]
        elapsed_ms = (time.monotonic() - start) * 1000
        text = resp.content[0].text if resp.content else ""
        tokens_in = resp.usage.input_tokens if resp.usage else 0
        tokens_out = resp.usage.output_tokens if resp.usage else 0
        return {
            "success": True,
            "text": text,
            "tokens_in": tokens_in,
            "tokens_out": tokens_out,
            "total_tokens": tokens_in + tokens_out,
            "elapsed_ms": elapsed_ms,
            "error": None,
        }
    except Exception as exc:
        elapsed_ms = (time.monotonic() - start) * 1000
        return {
            "success": False,
            "text": "",
            "tokens_in": 0,
            "tokens_out": 0,
            "total_tokens": 0,
            "elapsed_ms": elapsed_ms,
            "error": str(exc),
        }


# ---------------------------------------------------------------------------
# Grade file helpers (mirrors Go performance_grading.go logic)
# ---------------------------------------------------------------------------

def sanitize_filename(s: str) -> str:
    for ch in r'/\:*?"<>|':
        s = s.replace(ch, "_")
    return s


def grade_filename(model_id: str, role_id: str, project_id: str) -> Path:
    name = (
        f"{sanitize_filename(model_id)}_"
        f"{sanitize_filename(role_id)}_"
        f"{sanitize_filename(project_id)}.json"
    )
    return GRADES_DIR / name


def calculate_grade(success_rate: float, retry_rate: float, error_rate: float) -> str:
    """
    Mirrors PerformanceGradeManager.recalculateGrade() thresholds from config.go:
      A: success >= 0.90, retry <= 0.05
      B: success >= 0.80, retry <= 0.10
      C: success >= 0.70, retry <= 0.20
      D: success >= 0.60, retry <= 0.30
      F: otherwise
    """
    if success_rate >= 0.90 and retry_rate <= 0.05:
        return "A"
    if success_rate >= 0.80 and retry_rate <= 0.10:
        return "B"
    if success_rate >= 0.70 and retry_rate <= 0.20:
        return "C"
    if success_rate >= 0.60 and retry_rate <= 0.30:
        return "D"
    return "F"


def calculate_confidence(total_attempts: int) -> float:
    """Simple confidence score: tapers toward 1.0 as attempts → 50."""
    if total_attempts == 0:
        return 0.0
    return min(1.0, total_attempts / 50.0)


def load_or_create_grade(model_id: str, role_id: str, project_id: str) -> dict:
    path = grade_filename(model_id, role_id, project_id)
    if path.exists():
        with open(path) as f:
            return json.load(f)

    now = datetime.datetime.now(datetime.timezone.utc).isoformat().replace("+00:00", "Z")
    return {
        "model_id": model_id,
        "role_id": role_id,
        "project_id": project_id,
        "total_attempts": 0,
        "successes": 0,
        "failures": 0,
        "retries": 0,
        "total_tokens_used": 0,
        "total_execution_time_ms": 0,
        "average_tokens": 0,
        "average_execution_time": 0.0,
        "error_rate": 0.0,
        "retry_rate": 0.0,
        "success_rate": 0.0,
        "escalation_count": 0,
        "downgrade_count": 0,
        "grade": "F",
        "confidence_score": 0.0,
        "last_updated": now,
        "sample_size": 0,
    }


def update_grade(grade: dict, elapsed_ms: float, tokens: int, success: bool) -> dict:
    """Add one benchmark result into an existing grade dict and recalculate."""
    grade["total_attempts"] += 1
    grade["total_execution_time_ms"] += elapsed_ms
    grade["total_tokens_used"] += tokens

    if success:
        grade["successes"] += 1
    else:
        grade["failures"] += 1

    n = grade["total_attempts"]
    grade["sample_size"] = n
    grade["success_rate"] = grade["successes"] / n
    grade["error_rate"] = grade["failures"] / n
    grade["retry_rate"] = grade["retries"] / n
    grade["average_tokens"] = int(grade["total_tokens_used"] / n)
    grade["average_execution_time"] = grade["total_execution_time_ms"] / n
    grade["grade"] = calculate_grade(
        grade["success_rate"], grade["retry_rate"], grade["error_rate"]
    )
    grade["confidence_score"] = calculate_confidence(n)
    grade["last_updated"] = datetime.datetime.now(datetime.timezone.utc).isoformat().replace("+00:00", "Z")
    return grade


def save_grade(grade: dict) -> None:
    path = grade_filename(grade["model_id"], grade["role_id"], grade["project_id"])
    GRADES_DIR.mkdir(parents=True, exist_ok=True)
    with open(path, "w") as f:
        json.dump(grade, f, indent=2)
        f.write("\n")


# ---------------------------------------------------------------------------
# Main benchmark runner
# ---------------------------------------------------------------------------

def run_benchmark(
    runs: int = DEFAULT_RUNS,
    roles: list = None,
    dry_run: bool = False,
    model_filter: str = None,
    reset: bool = False,
) -> None:
    if roles is None:
        roles = DEFAULT_ROLES

    openai_key = os.environ.get("OPENAI_API_KEY", "")
    anthropic_key = os.environ.get("ANTHROPIC_API_KEY", "")

    if not openai_key:
        print("⚠️  OPENAI_API_KEY not set – OpenAI models will be skipped")
    if not anthropic_key:
        print("⚠️  ANTHROPIC_API_KEY not set – Anthropic models will be skipped")

    total_models = len(MODELS)
    results_summary = []  # (model_id, role, grade, success_rate, avg_latency)

    for model_idx, (model_id, provider, tier, api_model) in enumerate(MODELS, 1):
        # Optional filter
        if model_filter and model_filter.lower() not in model_id.lower():
            continue

        # Skip if API key missing
        if provider == "openai" and not openai_key:
            print(f"\n[{model_idx}/{total_models}] ⏭  Skipping {model_id} (no OpenAI key)")
            continue
        if provider == "anthropic" and not anthropic_key:
            print(f"\n[{model_idx}/{total_models}] ⏭  Skipping {model_id} (no Anthropic key)")
            continue

        print(f"\n{'='*70}")
        print(f"[{model_idx}/{total_models}] Model: {model_id}  (api: {api_model}, tier: {tier})")
        print(f"{'='*70}")

        for role in roles:
            print(f"\n  Role: {role}")
            grade = load_or_create_grade(model_id, role, PROJECT_ID)
            if reset and not dry_run:
                # Clear accumulated data so the benchmark starts fresh
                path = grade_filename(model_id, role, PROJECT_ID)
                if path.exists():
                    os.remove(path)
                    print(f"    ↩  Reset existing grade file")
                grade = load_or_create_grade(model_id, role, PROJECT_ID)
            run_successes = 0
            run_failures = 0

            for run_num in range(1, runs + 1):
                prompt = BENCHMARK_PROMPTS[(run_num - 1) % len(BENCHMARK_PROMPTS)]
                print(f"    Run {run_num}/{runs} – prompt: {prompt['user'][:60]}...", end="", flush=True)

                if dry_run:
                    print(" [DRY RUN]")
                    continue

                # Make the API call
                if provider == "openai":
                    result = call_openai(api_model, prompt, timeout=120)
                else:
                    result = call_anthropic(api_model, prompt, timeout=120)

                if not result["success"]:
                    print(f" ❌ API error: {result['error']}")
                    grade = update_grade(grade, result["elapsed_ms"], 0, False)
                    run_failures += 1
                    save_grade(grade)
                    continue

                quality_ok = evaluate_quality(result["text"], run_num - 1)
                status = "✅" if quality_ok else "⚠️ "
                print(
                    f" {status} {result['elapsed_ms']:.0f}ms "
                    f"{result['total_tokens']} tokens"
                )
                if not quality_ok:
                    print(f"       Quality check failed – output preview: "
                          f"{result['text'][:80]!r}")

                grade = update_grade(grade, result["elapsed_ms"], result["total_tokens"], quality_ok)
                if quality_ok:
                    run_successes += 1
                else:
                    run_failures += 1
                save_grade(grade)

                # Brief pause to avoid rate limiting
                time.sleep(0.5)

            if not dry_run:
                print(f"\n  → Grade: {grade['grade']}  "
                      f"success={grade['success_rate']:.0%}  "
                      f"avg_latency={grade['average_execution_time']:.0f}ms  "
                      f"avg_tokens={grade['average_tokens']}")
                results_summary.append((model_id, role, grade["grade"],
                                        grade["success_rate"],
                                        grade["average_execution_time"]))

    # ---------------------------------------------------------------------------
    # Final summary table
    # ---------------------------------------------------------------------------
    if not dry_run and results_summary:
        print(f"\n\n{'='*70}")
        print("BENCHMARK SUMMARY")
        print(f"{'='*70}")
        print(f"{'Model':<30} {'Role':<12} {'Grade':<6} {'SuccessRate':<12} {'Avg Latency'}")
        print(f"{'-'*70}")
        for model_id, role, grade, sr, avg_lat in sorted(results_summary, key=lambda x: x[0]):
            print(f"{model_id:<30} {role:<12} {grade:<6} {sr:<12.0%} {avg_lat:.0f}ms")
        print(f"\nGrade files written to: {GRADES_DIR}")


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Benchmark all models in ModelsByTier and populate performance grade data."
    )
    parser.add_argument(
        "--runs", type=int, default=DEFAULT_RUNS,
        help=f"Number of benchmark runs per model/role (default: {DEFAULT_RUNS})"
    )
    parser.add_argument(
        "--roles", nargs="+", default=DEFAULT_ROLES,
        help=f"Roles to benchmark (default: {' '.join(DEFAULT_ROLES)})"
    )
    parser.add_argument(
        "--model", type=str, default=None,
        help="Only benchmark models whose ID contains this substring"
    )
    parser.add_argument(
        "--reset", action="store_true",
        help="Delete existing grade files for matched models before benchmarking (start fresh)"
    )
    parser.add_argument(
        "--dry-run", action="store_true",
        help="Print what would be done without making API calls"
    )

    args = parser.parse_args()

    print("🚀 ai-pack Model Benchmark Suite")
    print(f"   Runs per model/role : {args.runs}")
    print(f"   Roles               : {', '.join(args.roles)}")
    print(f"   Model filter        : {args.model or '(all)'}")
    print(f"   Project ID          : {PROJECT_ID}")
    print(f"   Grades dir          : {GRADES_DIR}")
    print(f"   Dry run             : {args.dry_run}")

    run_benchmark(
        runs=args.runs,
        roles=args.roles,
        dry_run=args.dry_run,
        model_filter=args.model,
        reset=args.reset,
    )
