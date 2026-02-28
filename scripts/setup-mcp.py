#!/usr/bin/env python3
"""
setup-mcp.py — Register AI-Pack MCP servers in Claude Code and agent-server settings.

Usage:
    python3 scripts/setup-mcp.py [--local] [--dry-run]

Flags:
    --local         Write Claude Code config to .claude/settings.local.json in
                    the current project instead of the global ~/.claude/settings.json.
    --dry-run       Print what would be written without modifying any files.

MCP servers registered:
    kg   kg server --stdio
         Knowledge graph search/query/write. Project root and ID are auto-detected
         from the working directory when the server starts.

Targets updated:
    1. ~/.claude.json (or .claude/settings.local.json with --local)
       → mcpServers.kg  for interactive Claude Code sessions
    2. ~/.claude/agent-server.json
       → mcp.servers.kg + mcp.enabled_servers  for agent sessions
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


def claude_settings_dir() -> Path:
    """Return the platform-appropriate Claude Code settings directory (for local settings)."""
    if sys.platform == "win32":
        appdata = os.environ.get("APPDATA", "")
        if appdata:
            return Path(appdata) / "Claude"
    return Path.home() / ".claude"


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


def configure_agent_server(dry_run: bool) -> None:
    """Update mcp.servers and mcp.enabled_servers in ~/.claude/agent-server.json."""
    agent_server_path = claude_settings_dir() / "agent-server.json"
    config = load_json(agent_server_path)

    if not config:
        print(f"  ~/.claude/agent-server.json not found — skipping agent-server config")
        print(f"  (Run the agent-server once to create it, then re-run this script)")
        return

    mcp_section = config.setdefault("mcp", {})
    mcp_section["enabled"] = True

    # Add kg to servers
    servers = mcp_section.setdefault("servers", {})
    new_kg = build_kg_entry()
    server_changed = servers.get("kg") != new_kg
    if server_changed:
        servers["kg"] = new_kg

    # Add kg to enabled_servers (preserve existing entries)
    enabled = mcp_section.setdefault("enabled_servers", [])
    enabled_changed = "kg" not in enabled
    if enabled_changed:
        enabled.append("kg")

    if server_changed or enabled_changed:
        print(f"  Adding kg   → {agent_server_path}")
        save_json(agent_server_path, config, dry_run)
    else:
        print(f"  kg already configured in {agent_server_path} (no changes)")


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

    # 1. Claude Code settings (interactive sessions)
    if args.write_local:
        settings_path = cwd / ".claude" / "settings.local.json"
    else:
        settings_path = claude_global_config_path()
    configure_claude_settings(settings_path, args.dry_run)

    # 2. Agent-server config (agent sessions)
    configure_agent_server(args.dry_run)

    print()
    print("Next steps:")
    print("  1. Run:  kg index                  (indexes this codebase)")
    print("  2. Restart Claude Code / agent-server to pick up the new MCP server")


if __name__ == "__main__":
    main()
