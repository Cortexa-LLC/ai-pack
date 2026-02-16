#!/usr/bin/env python3
"""
Update LLM Provider Costs in Configuration

This script updates the provider cost configuration with current pricing.
Can be run manually or scheduled via cron for automatic updates.

Usage:
    # Interactive mode - prompts for each model
    ./scripts/update-provider-costs.py

    # Update specific model
    ./scripts/update-provider-costs.py --model "anthropic:claude-sonnet-4-5" --input 3.00 --output 15.00

    # Set from environment variables (for automation)
    ANTHROPIC_SONNET_INPUT=3.00 ANTHROPIC_SONNET_OUTPUT=15.00 ./scripts/update-provider-costs.py --auto

    # Show current costs
    ./scripts/update-provider-costs.py --show

Schedule with cron (update every 30 days):
    0 0 1 * * /path/to/update-provider-costs.py --auto
"""

import argparse
import json
import os
import sys
from datetime import datetime
from pathlib import Path
from typing import Dict, Optional

# Default config path
DEFAULT_CONFIG_PATH = Path.home() / ".claude" / "agent-server.json"

# Provider pricing information sources
PRICING_URLS = {
    "anthropic": "https://www.anthropic.com/pricing",
    "openai": "https://openai.com/api/pricing/",
}

# Environment variable mappings for automated updates
ENV_MAPPINGS = {
    "anthropic:claude-sonnet-4-5": ("ANTHROPIC_SONNET_45_INPUT", "ANTHROPIC_SONNET_45_OUTPUT"),
    "anthropic:claude-sonnet-4-5-20250929": ("ANTHROPIC_SONNET_45_INPUT", "ANTHROPIC_SONNET_45_OUTPUT"),
    "anthropic:claude-haiku-4-5": ("ANTHROPIC_HAIKU_45_INPUT", "ANTHROPIC_HAIKU_45_OUTPUT"),
    "openai:gpt-4o": ("OPENAI_GPT4O_INPUT", "OPENAI_GPT4O_OUTPUT"),
    "openai:gpt-4o-mini": ("OPENAI_GPT4O_MINI_INPUT", "OPENAI_GPT4O_MINI_OUTPUT"),
    "openai:gpt-5.2-mini": ("OPENAI_GPT52_MINI_INPUT", "OPENAI_GPT52_MINI_OUTPUT"),
}


def load_config(config_path: Path) -> Dict:
    """Load agent server configuration."""
    if not config_path.exists():
        print(f"❌ Config file not found: {config_path}")
        print(f"   Create it with default values first")
        sys.exit(1)

    with open(config_path) as f:
        return json.load(f)


def save_config(config: Dict, config_path: Path) -> None:
    """Save agent server configuration."""
    # Ensure directory exists
    config_path.parent.mkdir(parents=True, exist_ok=True)

    with open(config_path, 'w') as f:
        json.dump(config, f, indent=2)
        f.write('\n')  # Add trailing newline


def show_current_costs(config: Dict) -> None:
    """Display current provider costs."""
    costs = config.get('provider_costs', {})
    models = costs.get('models', {})
    last_updated = costs.get('last_updated', 'Never')

    print("\n📊 Current Provider Costs (per 1M tokens)")
    print("=" * 80)
    print(f"Last Updated: {last_updated}\n")

    if not models:
        print("No costs configured yet.\n")
        return

    for model_key, model_cost in sorted(models.items()):
        provider = model_cost.get('provider', 'unknown')
        desc = model_cost.get('description', model_key)
        input_cost = model_cost.get('input_cost_per_1m', 0)
        output_cost = model_cost.get('output_cost_per_1m', 0)

        print(f"  {desc}")
        print(f"    Model: {model_key}")
        print(f"    Input:  ${input_cost:.2f} / 1M tokens")
        print(f"    Output: ${output_cost:.2f} / 1M tokens")
        print()

    print(f"Pricing Sources:")
    for provider, url in PRICING_URLS.items():
        print(f"  • {provider.capitalize()}: {url}")
    print()


def update_model_cost(config: Dict, model_key: str, input_cost: float, output_cost: float,
                      description: Optional[str] = None) -> None:
    """Update cost for a specific model."""
    if 'provider_costs' not in config:
        config['provider_costs'] = {'models': {}, 'update_interval': 30}

    if 'models' not in config['provider_costs']:
        config['provider_costs']['models'] = {}

    # Determine provider from model key
    provider = model_key.split(':')[0] if ':' in model_key else 'unknown'

    # Get existing entry or create new one
    model_data = config['provider_costs']['models'].get(model_key, {})

    # Update costs
    model_data['provider'] = provider
    model_data['input_cost_per_1m'] = input_cost
    model_data['output_cost_per_1m'] = output_cost

    if description:
        model_data['description'] = description
    elif 'description' not in model_data:
        model_data['description'] = model_key

    config['provider_costs']['models'][model_key] = model_data
    config['provider_costs']['last_updated'] = datetime.utcnow().isoformat() + 'Z'


def update_from_env(config: Dict) -> int:
    """Update costs from environment variables."""
    updated = 0

    for model_key, (input_env, output_env) in ENV_MAPPINGS.items():
        input_val = os.getenv(input_env)
        output_val = os.getenv(output_env)

        if input_val and output_val:
            try:
                input_cost = float(input_val)
                output_cost = float(output_val)
                update_model_cost(config, model_key, input_cost, output_cost)
                print(f"✓ Updated {model_key}: ${input_cost:.2f} / ${output_cost:.2f}")
                updated += 1
            except ValueError as e:
                print(f"⚠️  Invalid value for {model_key}: {e}")

    return updated


def interactive_update(config: Dict) -> int:
    """Interactive mode to update costs."""
    costs = config.get('provider_costs', {}).get('models', {})
    updated = 0

    print("\n📝 Interactive Cost Update")
    print("=" * 80)
    print("Press Enter to keep current value, or enter new cost")
    print("Type 'skip' to skip a model\n")

    for model_key, model_cost in sorted(costs.items()):
        desc = model_cost.get('description', model_key)
        current_input = model_cost.get('input_cost_per_1m', 0)
        current_output = model_cost.get('output_cost_per_1m', 0)

        print(f"\n{desc} ({model_key})")
        print(f"  Current: Input ${current_input:.2f} / Output ${current_output:.2f}")

        # Get input cost
        input_str = input(f"  New input cost [${current_input:.2f}]: ").strip()
        if input_str.lower() == 'skip':
            continue

        input_cost = current_input
        if input_str:
            try:
                input_cost = float(input_str)
            except ValueError:
                print(f"  ⚠️  Invalid input, keeping ${current_input:.2f}")

        # Get output cost
        output_str = input(f"  New output cost [${current_output:.2f}]: ").strip()
        output_cost = current_output
        if output_str:
            try:
                output_cost = float(output_str)
            except ValueError:
                print(f"  ⚠️  Invalid output, keeping ${current_output:.2f}")

        # Update if changed
        if input_cost != current_input or output_cost != current_output:
            update_model_cost(config, model_key, input_cost, output_cost)
            print(f"  ✓ Updated to: ${input_cost:.2f} / ${output_cost:.2f}")
            updated += 1

    return updated


def main():
    parser = argparse.ArgumentParser(
        description='Update LLM provider costs in configuration',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__
    )
    parser.add_argument('--config', type=Path, default=DEFAULT_CONFIG_PATH,
                       help=f'Path to config file (default: {DEFAULT_CONFIG_PATH})')
    parser.add_argument('--show', action='store_true',
                       help='Show current costs and exit')
    parser.add_argument('--model', type=str,
                       help='Model key to update (e.g., "anthropic:claude-sonnet-4-5")')
    parser.add_argument('--input', type=float,
                       help='Input cost per 1M tokens')
    parser.add_argument('--output', type=float,
                       help='Output cost per 1M tokens')
    parser.add_argument('--description', type=str,
                       help='Model description')
    parser.add_argument('--auto', action='store_true',
                       help='Update from environment variables (for automation)')

    args = parser.parse_args()

    # Load configuration
    config = load_config(args.config)

    # Show current costs
    if args.show:
        show_current_costs(config)
        return

    # Specific model update
    if args.model:
        if not args.input or not args.output:
            print("❌ Both --input and --output are required with --model")
            sys.exit(1)

        update_model_cost(config, args.model, args.input, args.output, args.description)
        save_config(config, args.config)
        print(f"✅ Updated {args.model}")
        print(f"   Input:  ${args.input:.2f} / 1M tokens")
        print(f"   Output: ${args.output:.2f} / 1M tokens")
        return

    # Automated update from environment
    if args.auto:
        updated = update_from_env(config)
        if updated > 0:
            save_config(config, args.config)
            print(f"\n✅ Updated {updated} model(s) from environment variables")
        else:
            print("\n⚠️  No updates from environment variables")
            print("   Set environment variables like: ANTHROPIC_SONNET_45_INPUT=3.00")
        return

    # Interactive mode
    show_current_costs(config)
    updated = interactive_update(config)

    if updated > 0:
        save_config(config, args.config)
        print(f"\n✅ Updated {updated} model(s)")
        print(f"   Config saved to: {args.config}")
    else:
        print("\n   No changes made")

    print("\n💡 Pricing Sources:")
    for provider, url in PRICING_URLS.items():
        print(f"   • {provider.capitalize()}: {url}")
    print()


if __name__ == '__main__':
    main()
