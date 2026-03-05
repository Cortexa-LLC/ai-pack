#!/usr/bin/env python3
"""
AI-Pack Unified Startup Script
Starts agent server and/or GUI with proper coordination
"""

import os
import sys
import argparse
import subprocess
import signal
import time
from pathlib import Path


class Colors:
    """ANSI color codes for terminal output."""
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    MAGENTA = '\033[0;35m'
    CYAN = '\033[0;36m'
    RESET = '\033[0m'


def print_banner(mode="all"):
    """Print startup banner."""
    print()
    print(f"{Colors.BLUE}╔══════════════════════════════════════════════════════════════════╗{Colors.RESET}")
    print(f"{Colors.BLUE}║                                                                  ║{Colors.RESET}")
    print(f"{Colors.BLUE}║                         AI-Pack Startup                          ║{Colors.RESET}")
    print(f"{Colors.BLUE}║                                                                  ║{Colors.RESET}")

    if mode == "all":
        print(f"{Colors.BLUE}║              Starting: Agent Server + GUI                        ║{Colors.RESET}")
    elif mode == "server":
        print(f"{Colors.BLUE}║              Starting: Agent Server Only                         ║{Colors.RESET}")
    elif mode == "gui":
        print(f"{Colors.BLUE}║              Starting: GUI Only                                  ║{Colors.RESET}")

    print(f"{Colors.BLUE}║                                                                  ║{Colors.RESET}")
    print(f"{Colors.BLUE}╚══════════════════════════════════════════════════════════════════╝{Colors.RESET}")
    print()


def check_dependencies(need_go=True, need_node=True):
    """Check if required dependencies are installed."""
    issues = []

    if need_go:
        import shutil
        if not shutil.which('go'):
            issues.append("Go is not installed. Install from: https://go.dev/dl/")

    if need_node:
        import shutil
        if not shutil.which('node'):
            issues.append("Node.js is not installed. Install from: https://nodejs.org/")
        if not shutil.which('npm'):
            issues.append("npm is not installed. Install Node.js from: https://nodejs.org/")

    if issues:
        print(f"{Colors.RED}❌ Missing dependencies:{Colors.RESET}")
        for issue in issues:
            print(f"   {issue}")
        print()
        return False

    return True


def check_api_key():
    """Check for ANTHROPIC_API_KEY."""
    api_key = os.environ.get('ANTHROPIC_API_KEY')

    if not api_key:
        # Check Claude Code settings
        home = Path.home()
        claude_settings = home / '.claude' / 'settings.json'

        if claude_settings.exists():
            import json
            try:
                with open(claude_settings, 'r') as f:
                    settings = json.load(f)
                    if 'apiKeyHelper' in settings:
                        print(f"{Colors.GREEN}✅ Using Claude Code authentication{Colors.RESET}")
                        return True
            except (json.JSONDecodeError, IOError):
                pass

        print(f"{Colors.YELLOW}⚠️  Warning: ANTHROPIC_API_KEY not set{Colors.RESET}")
        print()
        print("Set API key:")
        print("  export ANTHROPIC_API_KEY=\"your-key-here\"")
        print()
        print("Or use Claude Code login:")
        print("  claude login")
        print()
        return False

    print(f"{Colors.GREEN}✅ ANTHROPIC_API_KEY configured{Colors.RESET}")
    return True


def install_gui_deps(gui_dir):
    """Install GUI dependencies if needed."""
    node_modules = gui_dir / 'node_modules'

    if not node_modules.exists():
        print(f"{Colors.CYAN}📦 Installing GUI dependencies...{Colors.RESET}")
        try:
            subprocess.run(['npm', 'install'], cwd=gui_dir, check=True)
            print(f"{Colors.GREEN}✅ GUI dependencies installed{Colors.RESET}")
        except subprocess.CalledProcessError as e:
            print(f"{Colors.RED}❌ Failed to install GUI dependencies: {e}{Colors.RESET}")
            return False

    return True


def start_server(project_root):
    """Start the agent server in a subprocess."""
    print(f"{Colors.GREEN}🚀 Starting Agent Server...{Colors.RESET}")

    import shutil
    server_bin = project_root / 'bin' / 'agent-server'
    if not server_bin.exists():
        found = shutil.which('agent-server')
        if not found:
            print(f"{Colors.RED}❌ agent-server binary not found. Run: make build-server{Colors.RESET}")
            return None
        server_bin = Path(found)

    try:
        # Start server as subprocess
        proc = subprocess.Popen(
            [str(server_bin), '--server'],
            cwd=project_root,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            bufsize=1
        )

        # Wait a bit and check if it started
        time.sleep(2)

        if proc.poll() is not None:
            print(f"{Colors.RED}❌ Server failed to start{Colors.RESET}")
            return None

        print(f"{Colors.GREEN}✅ Agent Server started (PID: {proc.pid}){Colors.RESET}")
        print(f"{Colors.CYAN}   http://localhost:8080{Colors.RESET}")
        print()

        return proc

    except Exception as e:
        print(f"{Colors.RED}❌ Failed to start server: {e}{Colors.RESET}")
        return None


def start_gui(gui_dir):
    """Start the GUI dev server in a subprocess."""
    print(f"{Colors.GREEN}🎨 Starting GUI...{Colors.RESET}")

    try:
        # Start GUI dev server
        proc = subprocess.Popen(
            ['npm', 'run', 'dev'],
            cwd=gui_dir,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            bufsize=1
        )

        # Wait a bit and check if it started
        time.sleep(2)

        if proc.poll() is not None:
            print(f"{Colors.RED}❌ GUI failed to start{Colors.RESET}")
            return None

        print(f"{Colors.GREEN}✅ GUI started (PID: {proc.pid}){Colors.RESET}")
        print(f"{Colors.CYAN}   http://localhost:3000{Colors.RESET}")
        print()

        return proc

    except Exception as e:
        print(f"{Colors.RED}❌ Failed to start GUI: {e}{Colors.RESET}")
        return None


def stream_output(proc, label, color):
    """Stream output from subprocess with label."""
    try:
        for line in iter(proc.stdout.readline, ''):
            if line:
                print(f"{color}[{label}]{Colors.RESET} {line.rstrip()}")
    except Exception:
        pass


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description='Start AI-Pack services')
    parser.add_argument('--server-only', action='store_true', help='Start only the agent server')
    parser.add_argument('--gui-only', action='store_true', help='Start only the GUI')
    args = parser.parse_args()

    # Determine what to start
    start_server_flag = not args.gui_only
    start_gui_flag = not args.server_only

    mode = "all"
    if args.server_only:
        mode = "server"
    elif args.gui_only:
        mode = "gui"

    # Setup paths
    project_root = Path(__file__).parent.parent.absolute()
    gui_dir = project_root / 'gui'

    # Print banner
    print_banner(mode)

    # Check dependencies
    if not check_dependencies(need_go=start_server_flag, need_node=start_gui_flag):
        sys.exit(1)

    # Check API key (only for server)
    if start_server_flag:
        if not check_api_key():
            sys.exit(1)
        print()

    # Install GUI deps if needed
    if start_gui_flag:
        if not install_gui_deps(gui_dir):
            sys.exit(1)
        print()

    # Start services
    processes = []

    try:
        if start_server_flag:
            server_proc = start_server(project_root)
            if server_proc:
                processes.append(('Server', server_proc, Colors.GREEN))

        if start_gui_flag:
            gui_proc = start_gui(gui_dir)
            if gui_proc:
                processes.append(('GUI', gui_proc, Colors.MAGENTA))

        if not processes:
            print(f"{Colors.RED}❌ No services started{Colors.RESET}")
            sys.exit(1)

        # Print status
        print(f"{Colors.CYAN}{'='*66}{Colors.RESET}")
        print(f"{Colors.GREEN}✅ All services running{Colors.RESET}")
        print()
        print("Services:")
        for label, proc, _ in processes:
            print(f"  • {label}: PID {proc.pid}")
        print()
        print(f"{Colors.YELLOW}Press Ctrl+C to stop all services{Colors.RESET}")
        print(f"{Colors.CYAN}{'='*66}{Colors.RESET}")
        print()

        # Wait for processes
        while True:
            time.sleep(1)

            # Check if any process died
            for label, proc, _ in processes:
                if proc.poll() is not None:
                    print(f"{Colors.RED}❌ {label} stopped unexpectedly{Colors.RESET}")
                    raise KeyboardInterrupt

    except KeyboardInterrupt:
        print()
        print(f"{Colors.YELLOW}Stopping services...{Colors.RESET}")

        # Terminate all processes
        for label, proc, _ in processes:
            try:
                proc.terminate()
                proc.wait(timeout=5)
                print(f"{Colors.GREEN}✅ {label} stopped{Colors.RESET}")
            except subprocess.TimeoutExpired:
                proc.kill()
                print(f"{Colors.YELLOW}⚠️  {label} force killed{Colors.RESET}")
            except Exception as e:
                print(f"{Colors.RED}❌ Error stopping {label}: {e}{Colors.RESET}")

        print()
        print(f"{Colors.GREEN}All services stopped{Colors.RESET}")
        sys.exit(0)


if __name__ == '__main__':
    main()
