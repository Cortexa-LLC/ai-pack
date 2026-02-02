#!/usr/bin/env python3
"""
AI-Pack launchd Setup Script
Manages macOS launchd plists for auto-start on login
"""

import os
import sys
import argparse
import subprocess
from pathlib import Path


class Colors:
    """ANSI color codes for terminal output."""
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    CYAN = '\033[0;36m'
    RESET = '\033[0m'


# launchd service labels
SERVER_LABEL = "com.cortexa.ai-pack.agent-server"
GUI_LABEL = "com.cortexa.ai-pack.gui"


def get_project_root():
    """Get the project root directory."""
    return Path(__file__).parent.parent.absolute()


def get_launchd_dir():
    """Get the LaunchAgents directory."""
    return Path.home() / 'Library' / 'LaunchAgents'


def get_plist_paths():
    """Get paths to plist files."""
    launchd_dir = get_launchd_dir()
    return {
        'server': launchd_dir / f"{SERVER_LABEL}.plist",
        'gui': launchd_dir / f"{GUI_LABEL}.plist"
    }


def generate_server_plist(project_root):
    """Generate plist content for agent server."""
    python_path = sys.executable
    start_script = project_root / 'scripts' / 'start-all.py'
    log_dir = project_root / 'logs'
    log_dir.mkdir(exist_ok=True)

    return f"""<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{SERVER_LABEL}</string>

    <key>ProgramArguments</key>
    <array>
        <string>{python_path}</string>
        <string>{start_script}</string>
        <string>--server-only</string>
    </array>

    <key>WorkingDirectory</key>
    <string>{project_root}</string>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <true/>

    <key>StandardOutPath</key>
    <string>{log_dir}/agent-server.log</string>

    <key>StandardErrorPath</key>
    <string>{log_dir}/agent-server.error.log</string>

    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:{Path.home() / 'go' / 'bin'}:{Path.home() / '.local' / 'bin'}</string>
        <key>HOME</key>
        <string>{Path.home()}</string>
    </dict>

    <key>ProcessType</key>
    <string>Background</string>

    <key>Nice</key>
    <integer>0</integer>
</dict>
</plist>
"""


def generate_gui_plist(project_root):
    """Generate plist content for GUI."""
    python_path = sys.executable
    start_script = project_root / 'scripts' / 'start-all.py'
    log_dir = project_root / 'logs'
    log_dir.mkdir(exist_ok=True)

    return f"""<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{GUI_LABEL}</string>

    <key>ProgramArguments</key>
    <array>
        <string>{python_path}</string>
        <string>{start_script}</string>
        <string>--gui-only</string>
    </array>

    <key>WorkingDirectory</key>
    <string>{project_root}</string>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <false/>

    <key>StandardOutPath</key>
    <string>{log_dir}/gui.log</string>

    <key>StandardErrorPath</key>
    <string>{log_dir}/gui.error.log</string>

    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:{Path.home() / '.nvm' / 'versions' / 'node'}:{Path.home() / '.local' / 'bin'}</string>
        <key>HOME</key>
        <string>{Path.home()}</string>
    </dict>

    <key>ProcessType</key>
    <string>Background</string>

    <key>Nice</key>
    <integer>0</integer>
</dict>
</plist>
"""


def install_plists(project_root):
    """Install launchd plists."""
    launchd_dir = get_launchd_dir()
    launchd_dir.mkdir(parents=True, exist_ok=True)

    plist_paths = get_plist_paths()

    print(f"{Colors.BLUE}Installing launchd plists...{Colors.RESET}")
    print()

    # Generate and write server plist
    print(f"{Colors.CYAN}Creating server plist...{Colors.RESET}")
    server_plist = generate_server_plist(project_root)
    with open(plist_paths['server'], 'w') as f:
        f.write(server_plist)
    print(f"{Colors.GREEN}✅ {plist_paths['server']}{Colors.RESET}")

    # Generate and write GUI plist
    print(f"{Colors.CYAN}Creating GUI plist...{Colors.RESET}")
    gui_plist = generate_gui_plist(project_root)
    with open(plist_paths['gui'], 'w') as f:
        f.write(gui_plist)
    print(f"{Colors.GREEN}✅ {plist_paths['gui']}{Colors.RESET}")

    print()

    # Load the plists
    print(f"{Colors.CYAN}Loading services...{Colors.RESET}")

    for service, plist_path in plist_paths.items():
        try:
            subprocess.run(['launchctl', 'load', str(plist_path)], check=True, capture_output=True)
            print(f"{Colors.GREEN}✅ {service} loaded{Colors.RESET}")
        except subprocess.CalledProcessError as e:
            # Check if already loaded
            if b'already loaded' in e.stderr:
                print(f"{Colors.YELLOW}⚠️  {service} already loaded, reloading...{Colors.RESET}")
                subprocess.run(['launchctl', 'unload', str(plist_path)], capture_output=True)
                subprocess.run(['launchctl', 'load', str(plist_path)], check=True, capture_output=True)
                print(f"{Colors.GREEN}✅ {service} reloaded{Colors.RESET}")
            else:
                print(f"{Colors.RED}❌ Failed to load {service}: {e.stderr.decode()}{Colors.RESET}")

    print()
    print(f"{Colors.GREEN}✅ launchd setup complete{Colors.RESET}")
    print()
    print("Logs location:")
    print(f"  {project_root}/logs/")
    print()
    print("Services will start automatically on next login.")
    print("To start now:")
    print(f"  launchctl start {SERVER_LABEL}")
    print(f"  launchctl start {GUI_LABEL}")


def uninstall_plists():
    """Uninstall launchd plists."""
    plist_paths = get_plist_paths()

    print(f"{Colors.YELLOW}Uninstalling launchd plists...{Colors.RESET}")
    print()

    for service, plist_path in plist_paths.items():
        if not plist_path.exists():
            print(f"{Colors.YELLOW}⚠️  {service} plist not found{Colors.RESET}")
            continue

        # Unload the service
        try:
            subprocess.run(['launchctl', 'unload', str(plist_path)], check=True, capture_output=True)
            print(f"{Colors.GREEN}✅ {service} unloaded{Colors.RESET}")
        except subprocess.CalledProcessError as e:
            if b'Could not find specified service' not in e.stderr:
                print(f"{Colors.RED}❌ Failed to unload {service}: {e.stderr.decode()}{Colors.RESET}")
            else:
                print(f"{Colors.YELLOW}⚠️  {service} not running{Colors.RESET}")

        # Remove the plist file
        try:
            plist_path.unlink()
            print(f"{Colors.GREEN}✅ {service} plist removed{Colors.RESET}")
        except Exception as e:
            print(f"{Colors.RED}❌ Failed to remove {service} plist: {e}{Colors.RESET}")

    print()
    print(f"{Colors.GREEN}✅ launchd uninstall complete{Colors.RESET}")


def show_status():
    """Show status of launchd services."""
    print(f"{Colors.BLUE}AI-Pack Service Status{Colors.RESET}")
    print()

    plist_paths = get_plist_paths()

    for service, plist_path in plist_paths.items():
        print(f"{Colors.CYAN}{service.upper()}:{Colors.RESET}")

        # Check if plist exists
        if not plist_path.exists():
            print(f"  Status: {Colors.RED}Not installed{Colors.RESET}")
            print(f"  Install: make setup-launchd")
            print()
            continue

        print(f"  Plist: {Colors.GREEN}Installed{Colors.RESET}")
        print(f"  Path: {plist_path}")

        # Check if service is loaded/running
        try:
            label = SERVER_LABEL if service == 'server' else GUI_LABEL
            result = subprocess.run(
                ['launchctl', 'list', label],
                capture_output=True,
                text=True,
                check=False
            )

            if result.returncode == 0:
                print(f"  Status: {Colors.GREEN}Running{Colors.RESET}")

                # Parse PID from output
                for line in result.stdout.split('\n'):
                    if '"PID"' in line:
                        pid = line.split('=')[-1].strip().rstrip(';')
                        print(f"  PID: {pid}")
                        break
            else:
                print(f"  Status: {Colors.YELLOW}Not running{Colors.RESET}")
                print(f"  Start: launchctl start {label}")

        except Exception as e:
            print(f"  Status: {Colors.RED}Error: {e}{Colors.RESET}")

        print()

    # Show log locations
    project_root = get_project_root()
    log_dir = project_root / 'logs'
    print(f"{Colors.CYAN}LOGS:{Colors.RESET}")
    print(f"  Directory: {log_dir}")
    if log_dir.exists():
        print(f"  Files:")
        for log_file in sorted(log_dir.glob('*.log')):
            size = log_file.stat().st_size
            print(f"    • {log_file.name} ({size:,} bytes)")
    else:
        print(f"  {Colors.YELLOW}No logs yet{Colors.RESET}")
    print()

    # Show control commands
    print(f"{Colors.CYAN}CONTROL:{Colors.RESET}")
    print(f"  Start:   launchctl start {SERVER_LABEL}")
    print(f"           launchctl start {GUI_LABEL}")
    print(f"  Stop:    launchctl stop {SERVER_LABEL}")
    print(f"           launchctl stop {GUI_LABEL}")
    print(f"  Restart: launchctl kickstart -k gui/{os.getuid()}/{SERVER_LABEL}")
    print(f"           launchctl kickstart -k gui/{os.getuid()}/{GUI_LABEL}")
    print()


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description='Setup launchd for AI-Pack')
    parser.add_argument('action', choices=['install', 'uninstall', 'status'],
                        help='Action to perform')
    args = parser.parse_args()

    # Check if macOS
    if sys.platform != 'darwin':
        print(f"{Colors.RED}❌ This script only works on macOS{Colors.RESET}")
        print()
        print("For other platforms, consider:")
        print("  • Linux: systemd user services")
        print("  • Windows: Task Scheduler or NSSM")
        sys.exit(1)

    project_root = get_project_root()

    if args.action == 'install':
        install_plists(project_root)
    elif args.action == 'uninstall':
        uninstall_plists()
    elif args.action == 'status':
        show_status()


if __name__ == '__main__':
    main()
