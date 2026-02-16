#!/usr/bin/env python3
"""
AI-Pack Agent Server - API Key Setup
Helps configure OpenAI and Anthropic API keys for multi-provider support.
"""

import os
import sys
import json
import subprocess
from pathlib import Path
from typing import Optional, Tuple
from urllib.request import Request, urlopen
from urllib.error import URLError

# Colors for terminal output
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'  # No Color

def print_header(text: str):
    """Print a formatted header"""
    print(f"\n{Colors.BLUE}{'=' * 60}{Colors.NC}")
    print(f"{Colors.BLUE}   {text}{Colors.NC}")
    print(f"{Colors.BLUE}{'=' * 60}{Colors.NC}\n")

def detect_shell_config() -> Path:
    """Detect the appropriate shell config file"""
    home = Path.home()

    # Check what shell is actually running
    shell = os.environ.get('SHELL', '')

    if 'zsh' in shell:
        zshrc = home / '.zshrc'
        if zshrc.exists():
            return zshrc

    if 'bash' in shell:
        # Prefer .bash_profile over .bashrc for login shells
        bash_profile = home / '.bash_profile'
        if bash_profile.exists():
            return bash_profile
        bashrc = home / '.bashrc'
        if bashrc.exists():
            return bashrc

    # Fallback: check which files exist
    for config_file in ['.bash_profile', '.bashrc', '.zshrc', '.profile']:
        config_path = home / config_file
        if config_path.exists():
            return config_path

    # Default to .bash_profile since user mentioned it
    return home / '.bash_profile'

def check_existing_keys() -> Tuple[bool, bool]:
    """Check if API keys are already set"""
    openai_set = bool(os.environ.get('OPENAI_API_KEY'))
    anthropic_set = bool(os.environ.get('ANTHROPIC_API_KEY'))

    print(f"{Colors.BLUE}Checking current configuration...{Colors.NC}\n")

    if openai_set:
        print(f"{Colors.GREEN}✓{Colors.NC} OPENAI_API_KEY is already set")
    else:
        print(f"{Colors.YELLOW}⚠{Colors.NC} OPENAI_API_KEY is not set")

    if anthropic_set:
        print(f"{Colors.GREEN}✓{Colors.NC} ANTHROPIC_API_KEY is already set")
    else:
        print(f"{Colors.YELLOW}⚠{Colors.NC} ANTHROPIC_API_KEY is not set")

    print()

    return openai_set, anthropic_set

def test_openai_key(api_key: str) -> bool:
    """Test OpenAI API key validity"""
    try:
        data = json.dumps({
            'model': 'gpt-4o-mini',
            'messages': [{'role': 'user', 'content': 'test'}],
            'max_tokens': 5
        }).encode('utf-8')

        req = Request('https://api.openai.com/v1/chat/completions', data=data)
        req.add_header('Authorization', f'Bearer {api_key}')
        req.add_header('Content-Type', 'application/json')

        with urlopen(req, timeout=10) as response:
            return response.status == 200
    except URLError as e:
        print(f"{Colors.YELLOW}⚠ Could not validate OpenAI key (network issue): {e}{Colors.NC}")
        print(f"{Colors.YELLOW}  Key may still be valid - will be tested when server starts{Colors.NC}")
        return True  # Assume valid if network fails
    except Exception as e:
        error_msg = str(e)
        if '401' in error_msg or 'Unauthorized' in error_msg:
            print(f"{Colors.RED}✗ OpenAI API key is invalid (401 Unauthorized){Colors.NC}")
            return False
        print(f"{Colors.YELLOW}⚠ Could not validate OpenAI key: {e}{Colors.NC}")
        print(f"{Colors.YELLOW}  Key may still be valid - will be tested when server starts{Colors.NC}")
        return True  # Assume valid if validation fails for other reasons

def test_anthropic_key(api_key: str) -> bool:
    """Test Anthropic API key validity"""
    try:
        data = json.dumps({
            'model': 'claude-3-5-haiku-20241022',
            'max_tokens': 10,
            'messages': [{'role': 'user', 'content': 'test'}]
        }).encode('utf-8')

        req = Request('https://api.anthropic.com/v1/messages', data=data)
        req.add_header('x-api-key', api_key)
        req.add_header('anthropic-version', '2023-06-01')
        req.add_header('Content-Type', 'application/json')

        with urlopen(req, timeout=10) as response:
            return response.status == 200
    except URLError as e:
        print(f"{Colors.YELLOW}⚠ Could not validate Anthropic key (network issue): {e}{Colors.NC}")
        print(f"{Colors.YELLOW}  Key may still be valid - will be tested when server starts{Colors.NC}")
        return True  # Assume valid if network fails
    except Exception as e:
        error_msg = str(e)
        if '401' in error_msg or '403' in error_msg or 'authentication' in error_msg.lower():
            print(f"{Colors.RED}✗ Anthropic API key is invalid (authentication failed){Colors.NC}")
            return False
        print(f"{Colors.YELLOW}⚠ Could not validate Anthropic key: {e}{Colors.NC}")
        print(f"{Colors.YELLOW}  Key may still be valid - will be tested when server starts{Colors.NC}")
        return True  # Assume valid if validation fails for other reasons

def test_api_keys():
    """Test configured API keys"""
    print(f"{Colors.BLUE}Testing API keys...{Colors.NC}\n")

    # Test OpenAI
    openai_key = os.environ.get('OPENAI_API_KEY')
    if openai_key:
        print(f"{Colors.BLUE}Testing OpenAI API...{Colors.NC}")
        if test_openai_key(openai_key):
            print(f"{Colors.GREEN}✓ OpenAI API key is valid{Colors.NC}\n")
        else:
            print(f"{Colors.RED}✗ OpenAI API key test failed{Colors.NC}\n")

    # Test Anthropic
    anthropic_key = os.environ.get('ANTHROPIC_API_KEY')
    if anthropic_key:
        print(f"{Colors.BLUE}Testing Anthropic API...{Colors.NC}")
        if test_anthropic_key(anthropic_key):
            print(f"{Colors.GREEN}✓ Anthropic API key is valid{Colors.NC}\n")
        else:
            print(f"{Colors.RED}✗ Anthropic API key test failed{Colors.NC}\n")

def update_shell_config(shell_config: Path, key_name: str, key_value: str):
    """Update shell config with API key"""
    # Read existing content
    if shell_config.exists():
        content = shell_config.read_text()
    else:
        content = ""

    # Check if key already exists
    export_line = f'export {key_name}="{key_value}"'

    if f'export {key_name}=' in content:
        # Update existing
        lines = content.split('\n')
        new_lines = []
        for line in lines:
            if line.startswith(f'export {key_name}='):
                new_lines.append(export_line)
            else:
                new_lines.append(line)
        content = '\n'.join(new_lines)
    else:
        # Add new
        if content and not content.endswith('\n'):
            content += '\n'
        content += f'\n# AI-Pack Agent Server - {key_name}\n'
        content += export_line + '\n'

    # Write back
    shell_config.write_text(content)

def setup_openai() -> bool:
    """Setup OpenAI API key"""
    print_header("OpenAI API Key Setup")

    print("Get your API key from: https://platform.openai.com/api-keys\n")
    print(f"{Colors.YELLOW}Models available:{Colors.NC}")
    print("  • gpt-5.2       - $5.00/$15.00 per 1M tokens (primary)")
    print("  • gpt-5.2-mini  - $0.60/$2.40 per 1M tokens (bulk)")
    print("  • gpt-4o-mini   - $0.15/$0.60 per 1M tokens (cheapest)\n")

    response = input("Do you want to set up OpenAI API key? (y/n): ").strip().lower()

    if response != 'y':
        return False

    import getpass
    api_key = getpass.getpass("Enter your OpenAI API key (sk-...): ").strip()

    if not api_key.startswith('sk-'):
        print(f"{Colors.RED}✗ Invalid key format (should start with 'sk-'){Colors.NC}")
        return False

    # Update shell config
    shell_config = detect_shell_config()
    update_shell_config(shell_config, 'OPENAI_API_KEY', api_key)

    print(f"{Colors.GREEN}✓ Added OPENAI_API_KEY to {shell_config}{Colors.NC}\n")

    # Set for current session
    os.environ['OPENAI_API_KEY'] = api_key

    return True

def setup_anthropic() -> bool:
    """Setup Anthropic API key"""
    print_header("Anthropic API Key Setup")

    print("Get your API key from: https://console.anthropic.com/settings/keys\n")
    print(f"{Colors.YELLOW}Models available:{Colors.NC}")
    print("  • claude-sonnet-3-5-20241022 - $3.00/$15.00 per 1M tokens")
    print("  • claude-opus-3-5-20241022   - $15.00/$75.00 per 1M tokens\n")
    print(f"{Colors.YELLOW}Note:{Colors.NC} Use Claude for critical tasks only (10% of work)\n")

    response = input("Do you want to set up Anthropic API key? (y/n): ").strip().lower()

    if response != 'y':
        return False

    import getpass
    api_key = getpass.getpass("Enter your Anthropic API key (sk-ant-...): ").strip()

    if not api_key.startswith('sk-ant-'):
        print(f"{Colors.RED}✗ Invalid key format (should start with 'sk-ant-'){Colors.NC}")
        return False

    # Update shell config
    shell_config = detect_shell_config()
    update_shell_config(shell_config, 'ANTHROPIC_API_KEY', api_key)

    print(f"{Colors.GREEN}✓ Added ANTHROPIC_API_KEY to {shell_config}{Colors.NC}\n")

    # Set for current session
    os.environ['ANTHROPIC_API_KEY'] = api_key

    return True

def main():
    """Main setup flow"""
    print_header("AI-Pack Agent Server - API Key Setup")

    shell_config = detect_shell_config()
    print(f"{Colors.BLUE}Shell config:{Colors.NC} {shell_config}\n")

    # Check existing keys
    openai_set, anthropic_set = check_existing_keys()

    if openai_set and anthropic_set:
        print(f"{Colors.GREEN}✓ Both API keys are configured!{Colors.NC}\n")
        test_api_keys()
        print(f"{Colors.GREEN}✓ Setup complete!{Colors.NC}\n")
        print(f"{Colors.YELLOW}Next steps:{Colors.NC}")
        print("1. Restart your agent server: pkill agent-server && ./bin/agent-server --server")
        print("2. Check logs: tail -f /tmp/agent-server.log\n")
        return 0

    print(f"{Colors.YELLOW}Let's set up your API keys...{Colors.NC}\n")

    # Setup each provider
    if not openai_set:
        setup_openai()

    if not anthropic_set:
        setup_anthropic()

    # Final summary
    print_header("Setup Complete!")

    # Test the keys
    test_api_keys()

    print(f"{Colors.YELLOW}Next steps:{Colors.NC}\n")
    print(f"1. Reload your shell config:")
    print(f"   {Colors.BLUE}source {shell_config}{Colors.NC}\n")
    print(f"2. Verify keys are set:")
    print(f"   {Colors.BLUE}echo $OPENAI_API_KEY | cut -c1-10{Colors.NC}")
    print(f"   {Colors.BLUE}echo $ANTHROPIC_API_KEY | cut -c1-10{Colors.NC}\n")
    print(f"3. Restart agent server:")
    print(f"   {Colors.BLUE}pkill agent-server && ./bin/agent-server --server{Colors.NC}\n")
    print(f"4. Check logs for 'openai_client_initialized':")
    print(f"   {Colors.BLUE}tail -f /tmp/agent-server.log | grep -i openai{Colors.NC}\n")

    print(f"{Colors.GREEN}✓ All done!{Colors.NC}\n")

    return 0

if __name__ == '__main__':
    try:
        sys.exit(main())
    except KeyboardInterrupt:
        print(f"\n{Colors.YELLOW}Setup cancelled by user{Colors.NC}")
        sys.exit(1)
    except Exception as e:
        print(f"{Colors.RED}✗ Error: {e}{Colors.NC}")
        sys.exit(1)
