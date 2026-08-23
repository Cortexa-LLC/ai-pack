#!/usr/bin/env bash
# uninstall-server.sh — remove the machine-level ai-pack ≤2.x agent-server installation.
#
#   bash scripts/uninstall-server.sh [--dry-run] [--purge]
#
# Default: stops services, removes service definitions, binaries, and the
# agent-mcp MCP registration. Archives ~/.ai-pack/tasks.db first.
# --purge additionally removes ~/.ai-pack and ~/.claude/performance_grades.
# Never touches ~/.claude/metrics (irreplaceable history) or any project's .ai/.

set -uo pipefail

DRY_RUN=0; PURGE=0
for a in "$@"; do
  [ "$a" = "--dry-run" ] && DRY_RUN=1
  [ "$a" = "--purge" ] && PURGE=1
done
[ "$DRY_RUN" = 1 ] && echo "(dry run — nothing will be changed)"

run() { if [ "$DRY_RUN" = 1 ]; then echo "would: $*"; else "$@"; fi; }

echo "== ai-pack 2.x server uninstall =="

# 1. Archive the task database before anything else
if [ -f "$HOME/.ai-pack/tasks.db" ]; then
  ARCH="$HOME/.ai-pack/archive-$(date +%Y%m%d)"
  if [ "$DRY_RUN" = 1 ]; then
    echo "would: archive ~/.ai-pack/tasks.db -> $ARCH/"
  else
    mkdir -p "$ARCH"
    if command -v sqlite3 >/dev/null 2>&1; then
      sqlite3 "$HOME/.ai-pack/tasks.db" "VACUUM INTO '$ARCH/tasks-backup.db'" 2>/dev/null \
        || cp "$HOME/.ai-pack/tasks.db" "$ARCH/tasks-backup.db"
    else
      cp "$HOME/.ai-pack/tasks.db" "$ARCH/tasks-backup.db"
    fi
    echo "archived: task history -> $ARCH/tasks-backup.db"
  fi
fi

# 2. Stop and remove services
case "$(uname -s)" in
  Darwin)
    for label in com.cortexa.ai-pack.agent-server com.cortexa.ai-pack.gui com.cortexa.ai-pack.gui-server; do
      launchctl print "gui/$(id -u)/$label" >/dev/null 2>&1 && run launchctl bootout "gui/$(id -u)/$label"
      plist="$HOME/Library/LaunchAgents/$label.plist"
      [ -f "$plist" ] && run rm "$plist" && [ "$DRY_RUN" = 0 ] && echo "removed: $plist"
    done
    ;;
  Linux)
    for unit in ai-pack-agent-server ai-pack-gui-server; do
      systemctl --user is-enabled "$unit" >/dev/null 2>&1 && {
        run systemctl --user stop "$unit"
        run systemctl --user disable "$unit"
      }
      f="$HOME/.config/systemd/user/$unit.service"
      [ -f "$f" ] && run rm "$f" && [ "$DRY_RUN" = 0 ] && echo "removed: $f"
    done
    [ "$DRY_RUN" = 0 ] && systemctl --user daemon-reload 2>/dev/null || true
    ;;
esac

# 3. Unregister the agent-mcp MCP server
if command -v claude >/dev/null 2>&1; then
  if claude mcp list 2>/dev/null | grep -q "^agent-mcp:"; then
    run claude mcp remove agent-mcp -s user
  fi
fi

# 4. Remove binaries
for b in /usr/local/bin/agent /usr/local/bin/agent-mcp /usr/local/bin/agent-server; do
  [ -f "$b" ] || continue
  if [ "$DRY_RUN" = 1 ]; then
    echo "would remove: $b"
  elif rm "$b" 2>/dev/null; then
    echo "removed: $b"
  else
    echo "needs sudo: sudo rm $b"
  fi
done
# NOTE: /usr/local/bin/kg is NOT removed — the 3.x plugin still uses it.

# 5. Optional purge of remaining state
if [ "$PURGE" = 1 ]; then
  [ -d "$HOME/.ai-pack" ] && run rm -rf "$HOME/.ai-pack" && [ "$DRY_RUN" = 0 ] && echo "purged: ~/.ai-pack (archive included — copy it elsewhere first if wanted)"
  [ -d "$HOME/.claude/performance_grades" ] && run rm -rf "$HOME/.claude/performance_grades" && [ "$DRY_RUN" = 0 ] && echo "purged: ~/.claude/performance_grades"
else
  echo ""
  echo "kept: ~/.ai-pack (config + archived task history) and ~/.claude/performance_grades"
  echo "      re-run with --purge to remove them"
fi

echo ""
echo "kept: ~/.claude/metrics (historical cost data), per-project .ai/ directories, /usr/local/bin/kg"
echo "Done. Install the 3.x plugin with 'make install-plugin' from the ai-pack repo."
