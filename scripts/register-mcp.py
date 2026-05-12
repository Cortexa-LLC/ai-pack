#!/usr/bin/env python3
"""
register-mcp.py — Idempotently register agent-mcp in ~/.claude/settings.json
(global scope) so Claude Code picks it up in every project.

Usage:
    python3 scripts/register-mcp.py [--dry-run] [--install-dir /usr/local/bin]

Flags:
    --dry-run       Print what would be written without modifying any files.
    --install-dir   Directory where agent-mcp binary lives (default: /usr/local/bin).
"""

import argparse
import json
import os
import sys


SETTINGS_JSON = os.path.expanduser("~/.claude/settings.json")
SERVER_NAME = "agent-mcp"


def load_json(path: str) -> dict:
    if os.path.exists(path):
        with open(path) as f:
            try:
                return json.load(f)
            except json.JSONDecodeError:
                print(f"WARNING: {path} is not valid JSON — starting fresh.", file=sys.stderr)
    return {}


def merge_server(config: dict, binary_path: str) -> tuple[dict, bool]:
    """
    Merge the agent-mcp server entry into the config dict.
    Returns (updated_config, changed).
    """
    mcp_servers = config.setdefault("mcpServers", {})

    new_entry = {
        "command": binary_path,
        "args": [],
        "type": "stdio",
    }

    existing = mcp_servers.get(SERVER_NAME)
    if existing == new_entry:
        return config, False  # Already registered, nothing to do

    mcp_servers[SERVER_NAME] = new_entry
    return config, True


def main() -> None:
    parser = argparse.ArgumentParser(description="Register agent-mcp in ~/.claude/settings.json")
    parser.add_argument("--dry-run", action="store_true", help="Print without writing")
    parser.add_argument("--install-dir", default="/usr/local/bin",
                        help="Directory containing the agent-mcp binary")
    args = parser.parse_args()

    binary_path = os.path.join(args.install_dir, "agent-mcp")

    config = load_json(SETTINGS_JSON)
    updated, changed = merge_server(config, binary_path)

    pretty = json.dumps(updated, indent=2) + "\n"

    if not changed:
        print(f"✅  agent-mcp already registered in {SETTINGS_JSON} — no changes needed.")
        return

    if args.dry_run:
        print(f"[dry-run] Would write to {SETTINGS_JSON}:")
        print(pretty)
        return

    os.makedirs(os.path.dirname(SETTINGS_JSON), exist_ok=True)
    with open(SETTINGS_JSON, "w") as f:
        f.write(pretty)

    print(f"✅  Registered '{SERVER_NAME}' in {SETTINGS_JSON}")
    print(f"    command: {binary_path}")
    print()
    print("Restart Claude Code (or run /mcp-restart) for the change to take effect.")


if __name__ == "__main__":
    main()
