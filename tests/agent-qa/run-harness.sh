#!/usr/bin/env bash
# run-harness.sh -- planted-defect harness for the reviewer agent's AUDIT MODE (US-201).
#
# Runs the ai-pack reviewer system prompt (plugin/agents/reviewer.md, frontmatter
# stripped) via `claude -p` against the fixture codebase, captures the review, and
# scores which of the six planted defects the review found.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
REVIEWER_MD="$REPO_ROOT/plugin/agents/reviewer.md"
FIXTURE_DIR="$SCRIPT_DIR/fixture"
DEFECTS="$SCRIPT_DIR/defects.yaml"
OUTPUT_DIR="$SCRIPT_DIR/results"
TASK_PROMPT="Perform a full adversarial audit of this whole project."
DRY_RUN=0
THRESHOLD=""
SUMMARY=""

usage() {
  cat <<'EOF'
Usage: run-harness.sh [--dry-run] [--output <dir>] [--threshold <n>] [--summary <file>]

Runs the ai-pack reviewer agent's AUDIT MODE against the planted-defect
fixture (tests/agent-qa/fixture/) and scores which of the six documented
defects (defects.yaml) the review found.

  --dry-run        Validate prompt extraction, print the command a real run
                   would execute, then generate a synthetic review from the
                   defects.yaml descriptions and score it. No claude
                   invocation, no network. Must score 6/6 — the dry run
                   scores with --require-all, so anything less exits 1.
  --output <dir>   Directory for review-output.txt (default: results/).
  --threshold <n>  Minimum found-count for exit 0, passed to score.py
                   (default: 4). Real runs only; the dry run always
                   requires all defects.
  --summary <file> Also write score.py's machine-readable JSON summary
                   ({"score", "total", "threshold", "per_defect"}) to
                   <file>. Used by CI (ADR-011); exit codes unchanged.
  -h, --help       Show this help.

NOTE: a real run (without --dry-run) invokes the claude CLI, consumes
Claude subscription quota, and typically takes several minutes. It runs
against a throwaway copy of the fixture in a temp directory outside this
repository, so the agent under test cannot read defects.yaml, this
repo's git history, or anything else that would leak the answer key.
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run) DRY_RUN=1 ;;
    --output)
      [ $# -ge 2 ] || { echo "error: --output requires a directory argument" >&2; exit 2; }
      OUTPUT_DIR="$2"; shift ;;
    --threshold)
      [ $# -ge 2 ] || { echo "error: --threshold requires a number argument" >&2; exit 2; }
      THRESHOLD="$2"; shift ;;
    --summary)
      [ $# -ge 2 ] || { echo "error: --summary requires a file argument" >&2; exit 2; }
      SUMMARY="$2"; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "error: unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
  shift
done

[ -f "$REVIEWER_MD" ] || { echo "error: reviewer prompt not found: $REVIEWER_MD" >&2; exit 1; }
[ -f "$DEFECTS" ] || { echo "error: defect manifest not found: $DEFECTS" >&2; exit 1; }
[ -d "$FIXTURE_DIR" ] || { echo "error: fixture directory not found: $FIXTURE_DIR" >&2; exit 1; }

# Strip the YAML frontmatter (everything through the second '---' fence) so
# only the reviewer body is used as the system prompt.
extract_system_prompt() {
  awk 'f==2 {print} /^---$/ && f<2 {f++}' "$REVIEWER_MD"
}

SYSTEM_PROMPT="$(extract_system_prompt)"
if [ -z "$SYSTEM_PROMPT" ]; then
  echo "error: extracted an empty system prompt from $REVIEWER_MD" >&2
  exit 1
fi
if ! printf '%s' "$SYSTEM_PROMPT" | grep -q "AUDIT MODE"; then
  echo "error: extracted prompt does not contain 'AUDIT MODE' -- reviewer.md changed shape?" >&2
  exit 1
fi

# --- CLI flag support probe (issue #32) -------------------------------------
# The harness caps the agent under test with `--max-turns`. That flag stopped
# appearing in `claude --help` as of CLI 2.1.231 but still parses, so it is an
# undocumented dependency: the day the CLI drops it, the turn cap silently
# disappears (or the invocation errors mid-run) with nothing to point at.
#
# The probe pairs the flag under test with a deliberately invalid one. The CLI
# reports the FIRST unrecognized option and the flag under test comes first, so
# "unknown option '--max-turns'" means it is gone. Argument parsing rejects the
# invalid flag before any session starts, so this costs no quota and no network.
#
# There is no supported replacement to switch to yet: as of 2.1.236 the flag list
# has no turn cap at all (`--max-budget-usd` caps spend, not turns). When one
# lands, swap it in here and in the two invocations below.
PROBE_FLAG="--ai-pack-flag-support-probe"

# 0 = accepted, 1 = rejected as unknown, 2 = indeterminate.
cli_flag_status() {
  local flag="$1" value="${2-}" out
  out=$(ANTHROPIC_API_KEY="" claude -p "$flag" ${value:+"$value"} "$PROBE_FLAG" "" 2>&1 || true)
  case "$out" in
    *"unknown option '$flag'"*)       return 1 ;;
    *"unknown option '$PROBE_FLAG'"*) return 0 ;;
    *)                                return 2 ;;
  esac
}

assert_max_turns_supported() {
  local rc=0
  cli_flag_status --max-turns 60 || rc=$?
  case "$rc" in
    0) ;;
    1)
      echo "error: this claude CLI ($(claude --version 2>/dev/null || echo 'version unknown')) no longer accepts --max-turns." >&2
      echo "       The harness relied on it to cap the agent under test at 60 turns; without a cap a" >&2
      echo "       runaway review would burn quota unbounded, so the harness refuses to run." >&2
      echo "       Fix: replace --max-turns in tests/agent-qa/run-harness.sh with the CLI's current" >&2
      echo "       turn-cap flag (check 'claude --help'), then update this probe. See issue #32." >&2
      exit 1 ;;
    *)
      echo "warning: could not determine whether this claude CLI still accepts --max-turns" >&2
      echo "         (the probe produced neither expected error). Proceeding; see issue #32." >&2 ;;
  esac
}

if command -v claude >/dev/null 2>&1; then
  assert_max_turns_supported
else
  echo "note: claude CLI not on PATH -- skipping the --max-turns support probe (issue #32)."
fi

mkdir -p "$OUTPUT_DIR"
REVIEW_OUT="$OUTPUT_DIR/review-output.txt"

if [ "$DRY_RUN" -eq 1 ]; then
  echo "== dry run: no claude invocation =="
  echo "reviewer prompt: $(printf '%s\n' "$SYSTEM_PROMPT" | wc -l | tr -d ' ') lines extracted from $REVIEWER_MD"
  echo "would run (cwd: a temp-dir copy of $FIXTURE_DIR, outside this repo):"
  printf '  claude -p %q \\\n' "$TASK_PROMPT"
  printf '    --system-prompt "<reviewer body, %s chars>" \\\n' "${#SYSTEM_PROMPT}"
  printf '    --allowedTools "Bash,Read,Glob,Grep" --max-turns 60 --output-format text\n'
  python3 "$SCRIPT_DIR/score.py" --defects "$DEFECTS" --print-descriptions > "$REVIEW_OUT"
  echo "synthetic review written to $REVIEW_OUT"
  # The synthetic review is built from the manifest's own descriptions, so it
  # must find every defect: score with --require-all, not the default gate.
  python3 "$SCRIPT_DIR/score.py" --defects "$DEFECTS" --review "$REVIEW_OUT" --require-all \
    ${SUMMARY:+--summary "$SUMMARY"}
else
  command -v claude >/dev/null 2>&1 || { echo "error: claude CLI not found on PATH" >&2; exit 1; }
  echo "== real run: this consumes Claude subscription quota and takes several minutes =="

  # Run against a throwaway copy of the fixture OUTSIDE this repository.
  # Running in-tree would leak the answer key: with Bash allowed, the agent
  # under test could read ../defects.yaml, and git commands in the fixture
  # cwd would operate on the ai-pack repo (whose history describes the
  # planted defects). A fresh single-commit repo in the copy keeps git
  # commands working without any telltale history.
  WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/recstore.XXXXXX")"
  trap 'rm -rf "$WORK_DIR"' EXIT
  RUN_DIR="$WORK_DIR/fixture"
  mkdir -p "$RUN_DIR"
  cp -R "$FIXTURE_DIR/." "$RUN_DIR/"
  if command -v git >/dev/null 2>&1; then
    git -C "$RUN_DIR" init -q
    git -C "$RUN_DIR" -c user.name=dev -c user.email=dev@localhost add -A
    git -C "$RUN_DIR" -c user.name=dev -c user.email=dev@localhost \
      commit -q -m "record store service" --no-gpg-sign
  fi

  (
    cd "$RUN_DIR"
    ANTHROPIC_API_KEY="" claude -p "$TASK_PROMPT" \
      --system-prompt "$SYSTEM_PROMPT" \
      --allowedTools "Bash,Read,Glob,Grep" \
      --max-turns 60 \
      --output-format text
  ) > "$REVIEW_OUT"
  echo "review written to $REVIEW_OUT"
  python3 "$SCRIPT_DIR/score.py" --defects "$DEFECTS" --review "$REVIEW_OUT" \
    ${THRESHOLD:+--threshold "$THRESHOLD"} ${SUMMARY:+--summary "$SUMMARY"}
fi
