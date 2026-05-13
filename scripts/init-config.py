#!/usr/bin/env python3
"""
Initialize AI-Pack configuration file at ~/.ai-pack/config.json

Creates default configuration if it doesn't exist, preserving user customizations.
"""

import json
import os
import sys
from pathlib import Path


DEFAULT_CONFIG = {
    "server": {
        "host": os.environ.get("SERVER_HOST", "localhost"),
        "port": int(os.environ.get("AIPACK_AGENT_SERVER_PORT", 8080)),
        "max_concurrent_agents": 10
    },
    "api": {
        "anthropic_model": "claude-sonnet-4-6",
        "max_tokens": 24000,
        "timeout_seconds": 600,
        "mode": "direct",
        "adaptive_model_selection": True
    },
    "logging": {
        "level": "info",
        "format": "json"
    },
    "metrics": {
        "enabled": True
    }
}


def get_config_path():
    """Get the config file path: ~/.ai-pack/config.json"""
    home = Path.home()
    config_dir = home / '.ai-pack'
    return config_dir / 'config.json'


def _apply_env_to_existing(config_path):
    """Update server.host/port in an existing config if env vars are explicitly set.
    Returns True if any changes were written, False otherwise.
    """
    env_port = os.environ.get("AIPACK_AGENT_SERVER_PORT")
    env_host = os.environ.get("SERVER_HOST")
    if not env_port and not env_host:
        return False

    with open(config_path) as f:
        config = json.load(f)

    changed = False
    server = config.setdefault("server", {})
    if env_port and str(server.get("port")) != env_port:
        server["port"] = int(env_port)
        changed = True
    if env_host and server.get("host") != env_host:
        server["host"] = env_host
        changed = True

    if changed:
        with open(config_path, 'w') as f:
            json.dump(config, f, indent=2)

    return changed


def init_config(force=False):
    """
    Initialize config file with defaults.

    Args:
        force: If True, overwrite existing config. Otherwise, preserve it.

    Returns:
        Path to config file
    """
    config_path = get_config_path()
    config_dir = config_path.parent

    # Create directory if needed
    config_dir.mkdir(parents=True, exist_ok=True)

    # Check if config exists
    if config_path.exists() and not force:
        # Surgically update server.host/port if env vars are explicitly set
        updated = _apply_env_to_existing(config_path)
        if updated:
            print(f"✅ Config updated: {config_path}")
        else:
            print(f"✅ Config already exists: {config_path}")
        return config_path

    # Write default config
    with open(config_path, 'w') as f:
        json.dump(DEFAULT_CONFIG, f, indent=2)

    if force:
        print(f"✅ Config reset to defaults: {config_path}")
    else:
        print(f"✅ Config created: {config_path}")

    print(f"\nDefault configuration:")
    print(f"  Server:  http://{DEFAULT_CONFIG['server']['host']}:{DEFAULT_CONFIG['server']['port']}")
    print(f"  Model:   {DEFAULT_CONFIG['api']['anthropic_model']}")
    print(f"  Logging: {DEFAULT_CONFIG['logging']['level']}")
    print(f"\nOverride with environment variables:")
    print(f"  AIPACK_AGENT_SERVER_PORT=8080")
    print(f"  ANTHROPIC_MODEL=claude-opus-4-7")
    print(f"  AIPACK_AGENT_SERVER_LOG_LEVEL=debug")

    return config_path


def show_config():
    """Display current configuration"""
    config_path = get_config_path()

    if not config_path.exists():
        print(f"❌ No config file found at {config_path}")
        print(f"   Run: python3 scripts/init-config.py")
        return 1

    with open(config_path) as f:
        config = json.load(f)

    print(f"Current configuration: {config_path}\n")
    print(json.dumps(config, indent=2))
    return 0


def main():
    import argparse

    parser = argparse.ArgumentParser(description='Initialize AI-Pack configuration')
    parser.add_argument('--force', action='store_true',
                       help='Overwrite existing config with defaults')
    parser.add_argument('--show', action='store_true',
                       help='Show current configuration')

    args = parser.parse_args()

    if args.show:
        return show_config()

    init_config(force=args.force)
    return 0


if __name__ == '__main__':
    sys.exit(main())
