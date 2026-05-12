#!/usr/bin/env python3
"""
register-mcp.py — Idempotently register mcp-agent in ~/.claude.json (global scope)
so Claude Code picks it up in every project.

Usage:
    python3 scripts/register-mcp.py [--dry-run] [--install-dir /usr/local/bin]

Flags:
    --dry-run       Print what would be written without modifying any files.
    --install-dir   Directory where mcp-agent binary lives (default: /usr/local/bin).
"""

import argparse
import json
import os
import sys


CLAUDE_JSON = os.path.expanduser("~/.claude.json")
SERVER_NAME = "mcp-agent"


def load_claude_json(path: str) -> dict:
    if os.path.exists(path):
        with open(path) as f:
            try:
                return json.load(f)
            except json.JSONDecodeError:
                print(f"WARNING: {path} is not valid JSON — starting fresh.", file=sys.stderr)
    return {}


def merge_server(config: dict, binary_path: str) -> tuple[dict, bool]:
    """
    Merge the mcp-agent server entry into the config dict.
    Returns (updated_config, changed).
    """
    # Claude Code uses `mcpServers` at the top level of ~/.claude.json
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
    parser = argparse.ArgumentParser(description="Register mcp-agent in ~/.claude.json")
    parser.add_argument("--dry-run", action="store_true", help="Print without writing")
    parser.add_argument("--install-dir", default="/usr/local/bin",
                        help="Directory containing the mcp-agent binary")
    args = parser.parse_args()

    binary_path = os.path.join(args.install_dir, "mcp-agent")

    config = load_claude_json(CLAUDE_JSON)
    updated, changed = merge_server(config, binary_path)

    pretty = json.dumps(updated, indent=2) + "\n"

    if not changed:
        print(f"✅  mcp-agent already registered in {CLAUDE_JSON} — no changes needed.")
        return

    if args.dry_run:
        print(f"[dry-run] Would write to {CLAUDE_JSON}:")
        print(pretty)
        return

    with open(CLAUDE_JSON, "w") as f:
        f.write(pretty)

    print(f"✅  Registered '{SERVER_NAME}' in {CLAUDE_JSON}")
    print(f"    command: {binary_path}")
    print()
    print("Restart Claude Code (or run /mcp-restart) for the change to take effect.")


if __name__ == "__main__":
    main()
