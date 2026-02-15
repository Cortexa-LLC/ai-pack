#!/usr/bin/env python3
"""
Check API usage for Claude and OpenAI
Compare current costs vs projected costs with multi-provider setup
"""

import os
import sys
import json
from datetime import datetime, timedelta
from typing import Dict, Optional
from urllib.request import Request, urlopen
from urllib.error import URLError

# Colors
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    BOLD = '\033[1m'
    NC = '\033[0m'

def get_anthropic_usage(api_key: str, days: int = 30) -> Optional[Dict]:
    """
    Get Anthropic usage data
    Note: Anthropic doesn't have a direct usage API endpoint yet,
    so we'll estimate from the console data if available
    """
    print(f"{Colors.BLUE}Fetching Anthropic (Claude) usage for last {days} days...{Colors.NC}")
    print(f"{Colors.YELLOW}Note: Anthropic doesn't provide programmatic usage API yet{Colors.NC}")
    print(f"{Colors.YELLOW}Please check manually at: https://console.anthropic.com/settings/billing{Colors.NC}\n")

    # Placeholder for when API becomes available
    return None

def get_openai_usage(api_key: str, days: int = 30) -> Optional[Dict]:
    """Get OpenAI usage data"""
    print(f"{Colors.BLUE}Fetching OpenAI usage for last {days} days...{Colors.NC}")

    start_date = (datetime.now() - timedelta(days=days)).strftime('%Y-%m-%d')
    end_date = datetime.now().strftime('%Y-%m-%d')

    try:
        # OpenAI usage endpoint
        url = f'https://api.openai.com/v1/usage?start_date={start_date}&end_date={end_date}'
        req = Request(url)
        req.add_header('Authorization', f'Bearer {api_key}')

        with urlopen(req, timeout=10) as response:
            if response.status == 200:
                data = json.loads(response.read().decode())
                return data
            else:
                print(f"{Colors.YELLOW}⚠ OpenAI usage API returned {response.status}{Colors.NC}")
                print(f"{Colors.YELLOW}Check manually at: https://platform.openai.com/usage{Colors.NC}\n")
                return None
    except URLError as e:
        print(f"{Colors.RED}✗ Error fetching OpenAI usage: {e}{Colors.NC}")
        return None
    except Exception as e:
        print(f"{Colors.RED}✗ Error: {e}{Colors.NC}")
        return None

def estimate_current_claude_usage():
    """Estimate current Claude-only usage and project savings"""
    print(f"\n{Colors.BLUE}{'=' * 70}{Colors.NC}")
    print(f"{Colors.BOLD}   Cost Comparison: Current vs Multi-Provider{Colors.NC}")
    print(f"{Colors.BLUE}{'=' * 70}{Colors.NC}\n")

    print(f"{Colors.YELLOW}Enter your current Claude usage (from console.anthropic.com/settings/billing):{Colors.NC}\n")

    # Get user input
    try:
        monthly_cost = float(input("Current monthly cost (in USD, e.g., 150.50): $").strip())
    except (ValueError, EOFError):
        print(f"\n{Colors.YELLOW}Using estimated values...{Colors.NC}")
        monthly_cost = 150.0  # Default estimate

    print(f"\n{Colors.BLUE}Current Setup (Claude Only):{Colors.NC}")
    print(f"  Monthly Cost: ${monthly_cost:.2f}")
    print(f"  Primary Model: Claude Sonnet ($3/$15 per 1M tokens)")
    print(f"  Usage: 100% Claude\n")

    # Calculate projected costs with multi-provider
    # CORRECTED: Use cheaper models as primary, GPT-5.2 for quality when needed
    print(f"{Colors.GREEN}Projected Setup (Multi-Provider - Cost Optimized):{Colors.NC}")

    # Optimal breakdown favoring cheaper models
    # Assume average of $9/1M for Claude Sonnet (weighted avg of input/output)
    claude_avg = 9.00
    gpt_5_2_mini_avg = 1.50  # Avg of $0.60/$2.40
    gpt_4o_mini_avg = 0.375  # Avg of $0.15/$0.60
    gpt_5_2_avg = 10.00  # Avg of $5/$15
    claude_sonnet_avg = 9.00  # Avg of $3/$15

    # NEW strategy: Favor cheapest models
    gpt4o_mini_cost = monthly_cost * 0.70 * (gpt_4o_mini_avg / claude_avg)  # 70% on cheapest
    gpt5_2_mini_cost = monthly_cost * 0.20 * (gpt_5_2_mini_avg / claude_avg)  # 20% on mini
    claude_cost = monthly_cost * 0.10  # 10% on Claude for critical work

    projected_total = gpt4o_mini_cost + gpt5_2_mini_cost + claude_cost
    savings = monthly_cost - projected_total
    savings_percent = (savings / monthly_cost) * 100

    print(f"  70% of work → GPT-4o-mini:    ${gpt4o_mini_cost:.2f}/mo  (simple tasks)")
    print(f"  20% of work → GPT-5.2-mini:   ${gpt5_2_mini_cost:.2f}/mo  (moderate tasks)")
    print(f"  10% of work → Claude Sonnet:  ${claude_cost:.2f}/mo  (critical only)")
    print(f"  {Colors.BOLD}Total Projected Cost: ${projected_total:.2f}/mo{Colors.NC}\n")

    print(f"{Colors.GREEN}{Colors.BOLD}Monthly Savings: ${savings:.2f} ({savings_percent:.1f}%){Colors.NC}\n")

    # Annual projection
    annual_savings = savings * 12
    print(f"{Colors.YELLOW}Annual Savings: ${annual_savings:.2f}{Colors.NC}\n")

    # Recommendations
    print(f"{Colors.BLUE}Optimization Strategy (Corrected for Maximum Savings):{Colors.NC}")
    print(f"  ✓ Use GPT-4o-mini for most tasks (70%) - 96% cheaper!")
    print(f"  ✓ Use GPT-5.2-mini for quality work (20%) - 83% cheaper!")
    print(f"  ✓ Reserve Claude for critical only (10%)")
    print(f"  ✓ Note: GPT-5.2 is actually MORE expensive than Claude ($10 vs $9 avg)")
    print(f"  ✓ For best savings, favor mini models over premium models\n")

    return {
        'current': monthly_cost,
        'projected': projected_total,
        'savings': savings,
        'savings_percent': savings_percent
    }

def check_agent_server_logs():
    """Check agent server logs for actual usage"""
    log_file = '/tmp/agent-server.log'

    if not os.path.exists(log_file):
        print(f"{Colors.YELLOW}⚠ Agent server logs not found at {log_file}{Colors.NC}")
        return

    print(f"\n{Colors.BLUE}Checking agent server logs...{Colors.NC}\n")

    # Count model usage from logs
    try:
        with open(log_file, 'r') as f:
            content = f.read()

            # Count model selections
            gpt_count = content.count('provider","openai')
            claude_count = content.count('provider","anthropic')

            if gpt_count > 0 or claude_count > 0:
                total = gpt_count + claude_count
                gpt_percent = (gpt_count / total * 100) if total > 0 else 0
                claude_percent = (claude_count / total * 100) if total > 0 else 0

                print(f"{Colors.GREEN}Model Usage (from logs):{Colors.NC}")
                print(f"  OpenAI requests:    {gpt_count} ({gpt_percent:.1f}%)")
                print(f"  Anthropic requests: {claude_count} ({claude_percent:.1f}%)")
                print(f"  Total requests:     {total}\n")
            else:
                print(f"{Colors.YELLOW}No model usage found in logs yet{Colors.NC}")
                print(f"{Colors.YELLOW}Start using the multi-provider setup to see usage stats{Colors.NC}\n")
    except Exception as e:
        print(f"{Colors.RED}✗ Error reading logs: {e}{Colors.NC}\n")

def main():
    """Main usage checker"""
    print(f"\n{Colors.BLUE}{'=' * 70}{Colors.NC}")
    print(f"{Colors.BLUE}   AI-Pack Agent Server - Usage & Cost Analysis{Colors.NC}")
    print(f"{Colors.BLUE}{'=' * 70}{Colors.NC}\n")

    # Check for API keys
    anthropic_key = os.environ.get('ANTHROPIC_API_KEY')
    openai_key = os.environ.get('OPENAI_API_KEY')

    if not anthropic_key:
        print(f"{Colors.YELLOW}⚠ ANTHROPIC_API_KEY not set{Colors.NC}")

    if not openai_key:
        print(f"{Colors.YELLOW}⚠ OPENAI_API_KEY not set{Colors.NC}")

    print()

    # Get usage data
    if anthropic_key:
        get_anthropic_usage(anthropic_key)

    if openai_key:
        get_openai_usage(openai_key)

    # Estimate savings
    estimate_current_claude_usage()

    # Check actual usage from logs
    check_agent_server_logs()

    print(f"{Colors.BLUE}{'=' * 70}{Colors.NC}")
    print(f"{Colors.GREEN}Tip: Run this script monthly to track actual savings!{Colors.NC}")
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
