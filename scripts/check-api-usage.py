#!/usr/bin/env python3
"""
Check current API usage from agent server logs
Shows which models/providers are being used
"""

import re
import sys
from pathlib import Path
from collections import defaultdict
from datetime import datetime

# Colors
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    BOLD = '\033[1m'
    NC = '\033[0m'

def parse_log_file(log_path):
    """Parse agent server logs for API usage"""

    api_calls = []
    model_usage = defaultdict(int)
    provider_usage = defaultdict(int)

    try:
        with open(log_path, 'r') as f:
            for line in f:
                # Look for API timing logs: "API: XXXms | in:XXXX out:XXX"
                api_match = re.search(r'API: (\d+)ms \| in:(\d+) out:(\d+)', line)
                if api_match:
                    duration_ms = int(api_match.group(1))
                    input_tokens = int(api_match.group(2))
                    output_tokens = int(api_match.group(3))

                    # Extract timestamp
                    time_match = re.search(r'time=(\S+)', line)
                    timestamp = time_match.group(1) if time_match else 'unknown'

                    # Extract task_id
                    task_match = re.search(r'task_id=(\S+)', line)
                    task_id = task_match.group(1) if task_match else 'unknown'

                    api_calls.append({
                        'timestamp': timestamp,
                        'task_id': task_id,
                        'duration_ms': duration_ms,
                        'input_tokens': input_tokens,
                        'output_tokens': output_tokens,
                        'provider': 'anthropic',  # Currently all calls go to Anthropic
                        'model': 'claude-sonnet-4-5-20250929'  # Default model
                    })

                # Look for model selection logs (when implemented)
                if 'model_selected' in line:
                    model_match = re.search(r'model_selected.*requested=(\S+).*provider=(\S+)', line)
                    if model_match:
                        model = model_match.group(1)
                        provider = model_match.group(2)
                        model_usage[model] += 1
                        provider_usage[provider] += 1

                # Look for provider initialization
                if 'openai_client_initialized' in line:
                    provider_usage['openai'] = 0  # Mark as available

    except FileNotFoundError:
        print(f"{Colors.RED}✗ Log file not found: {log_path}{Colors.NC}")
        return None, None, None

    return api_calls, model_usage, provider_usage

def calculate_costs(api_calls):
    """Calculate costs for API calls"""

    total_cost = 0
    provider_costs = defaultdict(float)

    # Pricing per 1M tokens
    claude_sonnet_input = 3.00
    claude_sonnet_output = 15.00

    for call in api_calls:
        input_cost = (call['input_tokens'] / 1_000_000) * claude_sonnet_input
        output_cost = (call['output_tokens'] / 1_000_000) * claude_sonnet_output

        call_cost = input_cost + output_cost
        total_cost += call_cost
        provider_costs[call['provider']] += call_cost

    return total_cost, provider_costs

def main():
    """Main analysis"""
    print(f"\n{Colors.BLUE}{'=' * 70}{Colors.NC}")
    print(f"{Colors.BOLD}   API Usage Analysis - Agent Server{Colors.NC}")
    print(f"{Colors.BLUE}{'=' * 70}{Colors.NC}\n")

    log_path = Path('/tmp/agent-server.log')

    print(f"{Colors.BLUE}Analyzing logs: {log_path}{Colors.NC}\n")

    api_calls, model_usage, provider_usage = parse_log_file(log_path)

    if api_calls is None:
        return 1

    if not api_calls:
        print(f"{Colors.YELLOW}No API calls found in logs{Colors.NC}")
        print(f"{Colors.YELLOW}The server may not have processed any tasks yet{Colors.NC}\n")
        return 0

    # Calculate totals
    total_input_tokens = sum(c['input_tokens'] for c in api_calls)
    total_output_tokens = sum(c['output_tokens'] for c in api_calls)
    total_cost, provider_costs = calculate_costs(api_calls)

    # Print summary
    print(f"{Colors.BOLD}Current Session Summary:{Colors.NC}")
    print(f"  Total API calls: {len(api_calls)}")
    print(f"  Input tokens:  {total_input_tokens:,} ({total_input_tokens/1_000_000:.3f}M)")
    print(f"  Output tokens: {total_output_tokens:,} ({total_output_tokens/1_000_000:.3f}M)")
    print(f"  Total cost:    ${total_cost:.4f}\n")

    # Provider breakdown
    print(f"{Colors.BOLD}Provider Usage:{Colors.NC}")

    # Check if OpenAI is available but not used
    if 'openai' in provider_usage:
        print(f"  {Colors.GREEN}✓ OpenAI: Available (not yet wired up){Colors.NC}")

    if provider_costs:
        for provider, cost in provider_costs.items():
            calls = sum(1 for c in api_calls if c['provider'] == provider)
            print(f"  {Colors.YELLOW}• {provider.capitalize()}: {calls} calls, ${cost:.4f}{Colors.NC}")

    print()

    # Model breakdown (currently all Claude Sonnet)
    print(f"{Colors.BOLD}Model Usage:{Colors.NC}")
    print(f"  {Colors.YELLOW}• claude-sonnet-4-5-20250929: {len(api_calls)} calls (100%){Colors.NC}")
    print()

    # Status
    print(f"{Colors.BLUE}{'=' * 70}{Colors.NC}")
    print(f"{Colors.YELLOW}⚠ Status: Multi-provider infrastructure ready but not wired up{Colors.NC}")
    print(f"{Colors.BLUE}{'=' * 70}{Colors.NC}\n")

    print(f"{Colors.YELLOW}Current behavior:{Colors.NC}")
    print(f"  • All API calls go to Claude Sonnet (${Colors.RED}$3/$15 per 1M tokens{Colors.NC})")
    print(f"  • OpenAI client initialized but not used")
    print(f"  • Agent configs with 'model:' field are ignored\n")

    print(f"{Colors.GREEN}Next step:{Colors.NC}")
    print(f"  Wire up the model selector to route requests based on agent config")
    print(f"  This will enable automatic cost optimization:\n")
    print(f"    • Simple tasks → {Colors.GREEN}GPT-4o-mini ($0.15/$0.60){Colors.NC}")
    print(f"    • Moderate tasks → {Colors.GREEN}GPT-5.2-mini ($0.60/$2.40){Colors.NC}")
    print(f"    • Critical tasks → {Colors.YELLOW}Claude Sonnet ($3/$15){Colors.NC}\n")

    print(f"{Colors.BLUE}{'=' * 70}{Colors.NC}\n")

    return 0

if __name__ == '__main__':
    try:
        sys.exit(main())
    except KeyboardInterrupt:
        print(f"\n{Colors.YELLOW}Cancelled by user{Colors.NC}")
        sys.exit(1)
    except Exception as e:
        print(f"{Colors.RED}✗ Error: {e}{Colors.NC}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
