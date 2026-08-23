#!/usr/bin/env bash
# uninstall-project.sh — remove server-era (ai-pack ≤2.x) integration from a consumer project.
#
# Run from the ROOT of the project that integrated ai-pack:
#   bash /path/to/ai-pack/scripts/uninstall-project.sh [--dry-run]
#
# Removes only files the ai-pack 2.x installer placed (.ai-pack submodule,
# .claude/commands/ai-pack, task hooks, rules, template skills/scripts).
# PRESERVES .ai/ (knowledge graph + task history) and never touches your own
# .claude content beyond the known ai-pack file list.
#
# The 3.x plugin needs NO per-project integration — install it once with
# `make install-plugin` from the ai-pack repo and it works in every project.

set -uo pipefail

DRY_RUN=0
[ "${1:-}" = "--dry-run" ] && DRY_RUN=1 && echo "(dry run — nothing will be removed)"

if [ ! -d .git ] && ! git rev-parse --git-dir >/dev/null 2>&1; then
  echo "ERROR: run this from the root of the project you want to clean." >&2
  exit 1
fi
if [ -d plugin/agents ] && [ -f .claude-plugin/marketplace.json ]; then
  echo "ERROR: this looks like the ai-pack repo itself, not a consumer project." >&2
  exit 1
fi

removed=0
drop() {
  local path="$1"
  [ -e "$path" ] || return 0
  if [ "$DRY_RUN" = 1 ]; then
    echo "would remove: $path"
  else
    git rm -r -q --cached "$path" 2>/dev/null || true
    rm -rf "$path"
    echo "removed: $path"
  fi
  removed=$((removed+1))
}

echo "== ai-pack 2.x project integration cleanup =="

# 1. The .ai-pack submodule
if [ -e .ai-pack ]; then
  if [ "$DRY_RUN" = 1 ]; then
    echo "would remove: .ai-pack submodule (+ .gitmodules entry, .git/modules/.ai-pack)"
  else
    git submodule deinit -f .ai-pack 2>/dev/null || true
    git rm -f -q .ai-pack 2>/dev/null || rm -rf .ai-pack
    rm -rf "$(git rev-parse --git-dir)/modules/.ai-pack" 2>/dev/null || true
    echo "removed: .ai-pack submodule"
  fi
  removed=$((removed+1))
fi

# 2. Slash commands
drop .claude/commands/ai-pack

# 3. Hooks
drop .claude/hooks/task-init.py
drop .claude/hooks/task-status.py

# 4. Rules
for f in gates.md knowledge-first.md orchestrator-enforcement.md task-packets.md workflows.md; do
  drop ".claude/rules/$f"
done

# 5. Template skills
for d in coordinator engineer orchestrator pr-feedback-loop reviewer watchdog; do
  drop ".claude/skills/$d"
done

# 6. Template scripts
for f in check-coordination-trigger.py coordination-timer.py coordinator-check.py \
         github-integration-timer.py rotate-work-log.py watchdog-check.py watchdog-timer.py; do
  drop ".claude/scripts/$f"
done

# 7. Template agent stubs
drop .claude/agents/Bash.md
drop .claude/agents/general-purpose.md

# 8. Hook registrations in settings.json (backs up first)
if [ -f .claude/settings.json ] && grep -qE "task-(init|status)\.py|coordination-timer|watchdog-(check|timer)|coordinator-check|github-integration-timer" .claude/settings.json; then
  if [ "$DRY_RUN" = 1 ]; then
    echo "would clean: ai-pack hook registrations in .claude/settings.json"
  else
    cp .claude/settings.json .claude/settings.json.ai-pack-backup
    python3 - <<'PYEOF'
import json
p = '.claude/settings.json'
MARKERS = ('task-init.py', 'task-status.py', 'coordination-timer', 'watchdog-check',
           'watchdog-timer', 'coordinator-check', 'github-integration-timer',
           'check-coordination-trigger', 'rotate-work-log')
d = json.load(open(p))
hooks = d.get('hooks', {})
for event in list(hooks):
    kept = []
    for matcher in hooks[event]:
        inner = [h for h in matcher.get('hooks', [])
                 if not any(m in h.get('command', '') for m in MARKERS)]
        if inner:
            matcher['hooks'] = inner
            kept.append(matcher)
    if kept:
        hooks[event] = kept
    else:
        del hooks[event]
if not hooks and 'hooks' in d:
    del d['hooks']
json.dump(d, open(p, 'w'), indent=2)
print('cleaned: .claude/settings.json (backup at .claude/settings.json.ai-pack-backup)')
PYEOF
  fi
  removed=$((removed+1))
fi

# 9. Things we deliberately do NOT touch
echo ""
echo "== preserved (never removed) =="
[ -d .ai ] && echo "  .ai/            — knowledge graph + task history (keep this)"
if [ -f CLAUDE.md ] && grep -qE "agent create|CRITICAL SESSION RULES|\.ai-pack" CLAUDE.md; then
  echo "  CLAUDE.md       — contains ai-pack 2.x workflow text; review and edit manually"
fi
[ -f .claudeignore ] && echo "  .claudeignore   — generic; review manually if unwanted"

echo ""
if [ "$removed" = 0 ]; then
  echo "Nothing to clean — no ai-pack 2.x integration found."
else
  echo "Done ($removed item(s)). Commit the removals when ready."
fi
echo "The 3.x plugin needs no per-project setup: run 'make install-plugin' once in the ai-pack repo."
