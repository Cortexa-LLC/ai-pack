#!/usr/bin/env python3
"""
setup-mcp.py — Register AI-Pack MCP servers in Claude Code settings.

Usage:
    python3 scripts/setup-mcp.py [--local] [--dry-run]

Flags:
    --local         Write config to .claude/settings.local.json in
                    the current project instead of the global ~/.claude.json.
    --dry-run       Print what would be written without modifying any files.

MCP servers registered:
    kg   kg server --stdio
         Knowledge graph search/query/write. Used by Claude Code interactive
         sessions. The agent-server manages its own per-project KG subprocesses
         dynamically — no static entry is needed in agent-server.json.

Target updated:
    ~/.claude.json (or .claude/settings.local.json with --local)
    → mcpServers.kg  for interactive Claude Code sessions
"""

import argparse
import json
import os
import sys
from pathlib import Path


def load_json(path: Path) -> dict:
    if path.exists():
        try:
            return json.loads(path.read_text())
        except json.JSONDecodeError as e:
            print(f"Warning: {path} is not valid JSON ({e}). Starting fresh.", file=sys.stderr)
    return {}


def save_json(path: Path, data: dict, dry_run: bool) -> None:
    text = json.dumps(data, indent=2) + "\n"
    if dry_run:
        print(f"[dry-run] Would write to {path}:")
        print(text)
    else:
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(text)
        print(f"✅ Written: {path}")


def build_kg_entry() -> dict:
    return {
        "command": "kg",
        "args": ["server", "--stdio"],
    }


def claude_global_config_path() -> Path:
    """Return the path to Claude Code's global config file.

    - macOS / Linux:  ~/.claude.json
    - Windows:        %APPDATA%\\Claude\\claude.json
    """
    if sys.platform == "win32":
        appdata = os.environ.get("APPDATA", "")
        if appdata:
            return Path(appdata) / "Claude" / "claude.json"
    return Path.home() / ".claude.json"


def configure_claude_settings(settings_path: Path, dry_run: bool) -> None:
    """Update mcpServers in a Claude Code config file."""
    settings = load_json(settings_path)
    servers = settings.setdefault("mcpServers", {})
    new_kg = build_kg_entry()

    if "kg" not in servers:
        servers["kg"] = new_kg
        print(f"  Adding kg   → {settings_path}")
        settings["mcpServers"] = servers
        save_json(settings_path, settings, dry_run)
    elif servers["kg"] != new_kg:
        servers["kg"] = new_kg
        print(f"  Updating kg → {settings_path}")
        settings["mcpServers"] = servers
        save_json(settings_path, settings, dry_run)
    else:
        print(f"  kg already configured in {settings_path} (no changes)")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--local", dest="write_local", action="store_true",
                        help="Write to the project-local .claude/settings.local.json instead of global")
    parser.add_argument("--dry-run", action="store_true",
                        help="Print changes without writing")
    args = parser.parse_args()

    cwd = Path(os.getcwd())

    print("AI-Pack MCP Setup")
    print()

    if args.write_local:
        settings_path = cwd / ".claude" / "settings.local.json"
    else:
        settings_path = claude_global_config_path()
    configure_claude_settings(settings_path, args.dry_run)

    print()
    print("Next steps:")
    print("  1. Run:  kg index                  (indexes this codebase)")
    print("  2. Restart Claude Code to pick up the new MCP server")
    print()
    print("Note: The agent-server starts per-project KG servers automatically.")
    print("      No manual agent-server configuration is required.")


if __name__ == "__main__":
    main()
