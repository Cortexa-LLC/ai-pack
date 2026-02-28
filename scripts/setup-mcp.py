#!/usr/bin/env python3
"""
setup-mcp.py — Register AI-Pack MCP servers in Claude Code settings.

Usage:
    python3 scripts/setup-mcp.py [--global] [--project-id NAME] [--dry-run]

Flags:
    --global        Write to ~/.claude/settings.json (affects all projects).
                    Default: write to .claude/settings.local.json in the
                    current project (recommended — keeps config project-local
                    and out of version control).
    --project-id    Override the project ID used by kg (default: basename of cwd).
    --dry-run       Print what would be written without modifying any files.

MCP servers registered:
    kg   kg server --stdio --project <project-id>
         Knowledge graph search and query for the current codebase.
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


def build_kg_entry(project_id: str) -> dict:
    return {
        "command": "kg",
        "args": ["server", "--stdio", "--project", project_id],
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
    parser.add_argument("--global", dest="write_global", action="store_true",
                        help="Write to the global Claude settings instead of project-local")
    parser.add_argument("--project-id", default=None,
                        help="Project ID for kg (default: basename of current directory)")
    parser.add_argument("--dry-run", action="store_true",
                        help="Print changes without writing")
    args = parser.parse_args()

    cwd = Path(os.getcwd())
    project_id = args.project_id or cwd.name

    if args.write_global:
        settings_path = claude_settings_dir() / "settings.json"
        scope = "global"
    else:
        settings_path = cwd / ".claude" / "settings.local.json"
        scope = "project-local"

    print(f"AI-Pack MCP Setup")
    print(f"  Settings file : {settings_path} ({scope})")
    print(f"  Project ID    : {project_id}")
    print()

    settings = load_json(settings_path)
    servers = settings.setdefault("mcpServers", {})

    added = []
    updated = []

    # kg MCP server
    new_kg = build_kg_entry(project_id)
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
