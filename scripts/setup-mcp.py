#!/usr/bin/env python3
"""
setup-mcp.py — Register AI-Pack MCP servers in Claude Code settings.

Usage:
    python3 scripts/setup-mcp.py [--local] [--dry-run]

Flags:
    --local         Write to .claude/settings.local.json in the current project
                    instead of the global ~/.claude/settings.json.
    --dry-run       Print what would be written without modifying any files.

MCP servers registered:
    kg   kg server --stdio
         Knowledge graph search/query. Project root and ID are auto-detected
         from the working directory when the server starts (no --project flag
         needed — handled by kg's built-in project root detection).
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


def claude_settings_dir() -> Path:
    """Return the platform-appropriate Claude Code settings directory.

    - macOS / Linux:  ~/.claude/
    - Windows:        %APPDATA%\\Claude\\  (Claude stores settings in AppData)
    """
    if sys.platform == "win32":
        appdata = os.environ.get("APPDATA", "")
        if appdata:
            return Path(appdata) / "Claude"
    return Path.home() / ".claude"


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--local", dest="write_local", action="store_true",
                        help="Write to the project-local .claude/settings.local.json instead of global")
    parser.add_argument("--dry-run", action="store_true",
                        help="Print changes without writing")
    args = parser.parse_args()

    cwd = Path(os.getcwd())

    if args.write_local:
        settings_path = cwd / ".claude" / "settings.local.json"
        scope = "project-local"
    else:
        settings_path = claude_settings_dir() / "settings.json"
        scope = "global"

    print(f"AI-Pack MCP Setup")
    print(f"  Settings file : {settings_path} ({scope})")
    print()

    settings = load_json(settings_path)
    servers = settings.setdefault("mcpServers", {})

    added = []
    updated = []

    # kg MCP server
    new_kg = build_kg_entry()
    if "kg" not in servers:
        servers["kg"] = new_kg
        added.append("kg")
    elif servers["kg"] != new_kg:
        servers["kg"] = new_kg
        updated.append("kg")
    else:
        print("  kg  — already configured (no changes)")

    if added:
        print(f"  Adding   : {', '.join(added)}")
    if updated:
        print(f"  Updating : {', '.join(updated)}")

    if added or updated:
        settings["mcpServers"] = servers
        save_json(settings_path, settings, args.dry_run)
        print()
        print("Next steps:")
        print(f"  1. Run:  kg index                  (indexes this codebase)")
        print(f"  2. Restart Claude Code to pick up the new MCP server")
    else:
        print()
        print("Nothing to do — all MCP servers already configured.")


if __name__ == "__main__":
    main()
