#!/usr/bin/env python3
"""
AI-Pack Agent Server Startup Script
A2A Protocol + Streaming + Parallel Execution
Cross-platform Python implementation
"""

import os
import sys
import json
import subprocess
import shutil
from pathlib import Path


class Colors:
    """ANSI color codes for terminal output (cross-platform)."""
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    RESET = '\033[0m'

    @staticmethod
    def disable_on_windows():
        """Disable colors on Windows if ANSI not supported."""
        if sys.platform == 'win32':
            # Windows 10+ supports ANSI, but check just in case
            try:
                import colorama
                colorama.init()
            except ImportError:
                # Disable colors if colorama not available
                Colors.RED = Colors.GREEN = Colors.YELLOW = Colors.BLUE = Colors.RESET = ''


def print_banner():
    """Print server startup banner."""
    Colors.disable_on_windows()
    print()
    print(f"{Colors.BLUE}╔══════════════════════════════════════════════════════════════════╗{Colors.RESET}")
    print(f"{Colors.BLUE}║                                                                  ║{Colors.RESET}")
    print(f"{Colors.BLUE}║                      AI-Pack Agent Server                        ║{Colors.RESET}")
    print(f"{Colors.BLUE}║                                                                  ║{Colors.RESET}")
    print(f"{Colors.BLUE}║   Features: A2A Protocol + SSE Streaming + Parallel Execution   ║{Colors.RESET}")
    print(f"{Colors.BLUE}║                                                                  ║{Colors.RESET}")
    print(f"{Colors.BLUE}╚══════════════════════════════════════════════════════════════════╝{Colors.RESET}")
    print()


def check_go_installation():
    """Check if Go is installed and get version."""
    if not shutil.which('go'):
        print(f"{Colors.RED}❌ Error: Go is not installed{Colors.RESET}")
        print()
        print("Please install Go 1.21+ first:")
        print()
        print("  macOS:   brew install go")
        print("  Linux:   https://go.dev/dl/")
        print("  Windows: https://go.dev/dl/")
        print()
        print("Then run this script again.")
        sys.exit(1)

    # Get Go version
    try:
        result = subprocess.run(['go', 'version'], capture_output=True, text=True, check=True)
        version = result.stdout.split()[2].replace('go', '')
        print(f"{Colors.GREEN}✅ Go {version} installed{Colors.RESET}")
        return True
    except subprocess.CalledProcessError:
        print(f"{Colors.RED}❌ Error: Could not determine Go version{Colors.RESET}")
        sys.exit(1)


def check_api_key():
    """Check for ANTHROPIC_API_KEY or Claude Code authentication."""
    api_key = os.environ.get('ANTHROPIC_API_KEY')

    if not api_key:
        # Check if Claude Code is configured
        home = Path.home()
        claude_settings = home / '.claude' / 'settings.json'

        if claude_settings.exists():
            try:
                with open(claude_settings, 'r') as f:
                    settings = json.load(f)
                    if 'apiKeyHelper' in settings:
                        print(f"{Colors.GREEN}✅ Using Claude Code authentication{Colors.RESET}")
                        return
            except (json.JSONDecodeError, IOError):
                pass

        print(f"{Colors.YELLOW}⚠️  Warning: ANTHROPIC_API_KEY not set and Claude Code not configured{Colors.RESET}")
        print()
        print("Option 1 - Set API key manually:")
        print("  export ANTHROPIC_API_KEY=\"your-key-here\"")
        print()
        print("Option 2 - Use Claude Code login (if you're already logged in):")
        print("  claude login")
        print()
        sys.exit(1)
    else:
        print(f"{Colors.GREEN}✅ ANTHROPIC_API_KEY configured{Colors.RESET}")

    print()


def install_dependencies(a2a_agent_dir):
    """Install Go dependencies."""
    print("📦 Installing Go dependencies...")

    try:
        subprocess.run(['go', 'mod', 'tidy'], cwd=a2a_agent_dir, check=True)
        print(f"{Colors.GREEN}✅ Dependencies installed{Colors.RESET}")
        print()
    except subprocess.CalledProcessError as e:
        print(f"{Colors.RED}❌ Error installing dependencies: {e}{Colors.RESET}")
        sys.exit(1)


def print_features():
    """Display server features."""
    print(f"{Colors.BLUE}🎯 Features:{Colors.RESET}")
    print("   - A2A Protocol Compliance (JSON-RPC 2.0)  ✅")
    print("   - SSE Streaming (Real-time progress)       ✅")
    print("   - Parallel Execution (configurable)        ✅")
    print("   - Structured Logging & Metrics             ✅")
    print()


def start_server(a2a_agent_dir):
    """Start the agent server."""
    print(f"{Colors.GREEN}🔥 Starting AI-Pack Agent Server...{Colors.RESET}")
    print()

    try:
        # Run the server (this will block until server stops)
        subprocess.run(
            ['go', 'run', 'cmd/agent-server/main.go', '--server'],
            cwd=a2a_agent_dir,
            check=True
        )
    except subprocess.CalledProcessError as e:
        print(f"{Colors.RED}❌ Server exited with error: {e}{Colors.RESET}")
        sys.exit(1)
    except KeyboardInterrupt:
        print()
        print(f"{Colors.YELLOW}Server stopped by user{Colors.RESET}")
        sys.exit(0)


def main():
    """Main entry point."""
    # Determine paths
    script_dir = Path(__file__).parent.absolute()
    a2a_agent_dir = script_dir.parent

    # Change to a2a-agent directory
    os.chdir(a2a_agent_dir)

    # Run startup checks
    print_banner()
    check_go_installation()
    check_api_key()
    install_dependencies(a2a_agent_dir)
    print_features()
    start_server(a2a_agent_dir)


if __name__ == '__main__':
    main()
