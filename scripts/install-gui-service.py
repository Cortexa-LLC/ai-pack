#!/usr/bin/env python3
"""
AI-Pack GUI Server - Service Installation
Installs the GUI server (Vite dev server) as a user service
"""

import os
import sys
import platform
import subprocess
from pathlib import Path

# Colors
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    BOLD = '\033[1m'
    NC = '\033[0m'

def print_header(text):
    """Print formatted header"""
    print(f"\n{Colors.BLUE}{'=' * 70}{Colors.NC}")
    print(f"{Colors.BLUE}   {text}{Colors.NC}")
    print(f"{Colors.BLUE}{'=' * 70}{Colors.NC}\n")

def detect_os():
    """Detect operating system"""
    system = platform.system().lower()
    if system == 'darwin':
        return 'macos'
    elif system == 'linux':
        return 'linux'
    else:
        return None

def get_project_root():
    """Get project root directory"""
    script_dir = Path(__file__).parent
    return script_dir.parent.absolute()

def check_node():
    """Check if Node.js and npm are available"""
    try:
        result = subprocess.run(['npm', '--version'], capture_output=True, text=True)
        if result.returncode == 0:
            print(f"{Colors.GREEN}✓ npm version: {result.stdout.strip()}{Colors.NC}")
            return True
    except FileNotFoundError:
        pass

    print(f"{Colors.RED}✗ npm not found{Colors.NC}")
    print(f"{Colors.YELLOW}  Install Node.js from: https://nodejs.org/{Colors.NC}")
    return False

def check_gui_dependencies(project_root):
    """Check if GUI dependencies are installed"""
    gui_dir = project_root / 'gui'
    node_modules = gui_dir / 'node_modules'

    if not node_modules.exists():
        print(f"{Colors.YELLOW}⚠ GUI dependencies not installed{Colors.NC}")
        print(f"{Colors.BLUE}Installing dependencies...{Colors.NC}")

        try:
            subprocess.run(['npm', 'install'], cwd=gui_dir, check=True)
            print(f"{Colors.GREEN}✓ Dependencies installed{Colors.NC}\n")
        except subprocess.CalledProcessError as e:
            print(f"{Colors.RED}✗ Failed to install dependencies: {e}{Colors.NC}")
            return False

    return True

def get_npm_path():
    """Get the full path to npm"""
    try:
        result = subprocess.run(['which', 'npm'], capture_output=True, text=True, check=True)
        return result.stdout.strip()
    except subprocess.CalledProcessError:
        # Try common locations
        for path in ['/opt/homebrew/bin/npm', '/usr/local/bin/npm', '/usr/bin/npm']:
            if Path(path).exists():
                return path
        return 'npm'

def install_macos_service(project_root):
    """Install macOS LaunchAgent for GUI server"""
    print(f"{Colors.BLUE}Installing macOS Launch Agent for GUI server...{Colors.NC}\n")

    # Paths
    plist_template = project_root / 'config' / 'com.cortexa.ai-pack.gui-server.plist'
    launch_agents_dir = Path.home() / 'Library' / 'LaunchAgents'
    install_plist = launch_agents_dir / 'com.cortexa.ai-pack.gui-server.plist'

    # Create LaunchAgents directory
    launch_agents_dir.mkdir(parents=True, exist_ok=True)

    # Read template
    with open(plist_template, 'r') as f:
        plist_content = f.read()

    # Replace placeholders
    plist_content = plist_content.replace('{{PROJECT_ROOT}}', str(project_root))
    npm_path = get_npm_path()
    plist_content = plist_content.replace('{{NPM_PATH}}', npm_path)

    # Write to install location
    with open(install_plist, 'w') as f:
        f.write(plist_content)

    print(f"{Colors.GREEN}✓ Plist installed to: {install_plist}{Colors.NC}")

    # Unload existing service
    subprocess.run(['launchctl', 'unload', str(install_plist)],
                   capture_output=True, check=False)

    # Load the service
    result = subprocess.run(['launchctl', 'load', str(install_plist)],
                           capture_output=True)

    if result.returncode == 0:
        print(f"{Colors.GREEN}✓ Launch Agent loaded{Colors.NC}\n")
    else:
        print(f"{Colors.YELLOW}⚠ Failed to load service{Colors.NC}")
        print(f"Error: {result.stderr.decode()}")
        return False

    # Check if running
    import time
    time.sleep(3)

    result = subprocess.run(['launchctl', 'list'], capture_output=True, text=True)
    if 'com.cortexa.ai-pack.gui-server' in result.stdout:
        print(f"{Colors.GREEN}✓ GUI server is running{Colors.NC}")
        print(f"{Colors.GREEN}✓ Access at: http://localhost:3000{Colors.NC}\n")
    else:
        print(f"{Colors.YELLOW}⚠ Service loaded but may not be running{Colors.NC}")
        print(f"{Colors.YELLOW}  Check logs: tail -f /tmp/gui-server.log{Colors.NC}\n")

    # Print management commands
    print(f"{Colors.BLUE}Management commands:{Colors.NC}")
    print(f"  Start:   {Colors.GREEN}launchctl load {install_plist}{Colors.NC}")
    print(f"  Stop:    {Colors.RED}launchctl unload {install_plist}{Colors.NC}")
    print(f"  Restart: {Colors.YELLOW}launchctl unload {install_plist} && launchctl load {install_plist}{Colors.NC}")
    print(f"  Logs:    {Colors.BLUE}tail -f /tmp/gui-server.log{Colors.NC}")

    return True

def install_linux_service(project_root):
    """Install Linux systemd user service for GUI server"""
    print(f"{Colors.BLUE}Installing systemd user service for GUI server...{Colors.NC}\n")

    # Paths
    service_template = project_root / 'config' / 'ai-pack-gui-server.service'
    systemd_user_dir = Path.home() / '.config' / 'systemd' / 'user'
    install_service = systemd_user_dir / 'ai-pack-gui-server.service'

    # Create systemd user directory
    systemd_user_dir.mkdir(parents=True, exist_ok=True)

    # Read template
    with open(service_template, 'r') as f:
        service_content = f.read()

    # Replace placeholders
    service_content = service_content.replace('{{PROJECT_ROOT}}', str(project_root))
    service_content = service_content.replace('{{USER}}', os.getenv("USER"))
    npm_path = get_npm_path()
    service_content = service_content.replace('{{NPM_PATH}}', npm_path)

    # Write service file
    with open(install_service, 'w') as f:
        f.write(service_content)

    print(f"{Colors.GREEN}✓ Service file installed to: {install_service}{Colors.NC}")

    # Reload systemd
    subprocess.run(['systemctl', '--user', 'daemon-reload'], check=True)

    # Enable and start service
    subprocess.run(['systemctl', '--user', 'enable', 'ai-pack-gui-server.service'], check=True)
    subprocess.run(['systemctl', '--user', 'start', 'ai-pack-gui-server.service'], check=True)

    print(f"{Colors.GREEN}✓ Service enabled and started{Colors.NC}\n")

    # Check status
    result = subprocess.run(['systemctl', '--user', 'is-active', 'ai-pack-gui-server.service'],
                           capture_output=True, text=True)

    if result.stdout.strip() == 'active':
        print(f"{Colors.GREEN}✓ GUI server is running{Colors.NC}")
        print(f"{Colors.GREEN}✓ Access at: http://localhost:3000{Colors.NC}\n")
    else:
        print(f"{Colors.YELLOW}⚠ Service may not be running{Colors.NC}")
        subprocess.run(['systemctl', '--user', 'status', 'ai-pack-gui-server.service', '--no-pager'])

    # Print management commands
    print(f"\n{Colors.BLUE}Management commands:{Colors.NC}")
    print(f"  Start:   {Colors.GREEN}systemctl --user start ai-pack-gui-server{Colors.NC}")
    print(f"  Stop:    {Colors.RED}systemctl --user stop ai-pack-gui-server{Colors.NC}")
    print(f"  Restart: {Colors.YELLOW}systemctl --user restart ai-pack-gui-server{Colors.NC}")
    print(f"  Status:  {Colors.BLUE}systemctl --user status ai-pack-gui-server{Colors.NC}")
    print(f"  Logs:    {Colors.BLUE}journalctl --user -u ai-pack-gui-server -f{Colors.NC}")
    print(f"  Disable: {Colors.RED}systemctl --user disable ai-pack-gui-server{Colors.NC}")

    return True

def main():
    """Main installation flow"""
    print_header("AI-Pack GUI Server - Service Installation")

    # Detect OS
    os_type = detect_os()
    if not os_type:
        print(f"{Colors.RED}✗ Unsupported operating system{Colors.NC}")
        return 1

    print(f"{Colors.BLUE}Detected OS: {os_type.upper()}{Colors.NC}\n")

    # Get project root
    project_root = get_project_root()
    print(f"{Colors.BLUE}Project root: {project_root}{Colors.NC}\n")

    # Check Node.js/npm
    print(f"{Colors.BLUE}Checking Node.js...{Colors.NC}")
    if not check_node():
        return 1
    print()

    # Check GUI dependencies
    if not check_gui_dependencies(project_root):
        return 1

    # Install service based on OS
    if os_type == 'macos':
        success = install_macos_service(project_root)
    else:  # linux
        success = install_linux_service(project_root)

    if not success:
        return 1

    print(f"\n{Colors.BLUE}{'=' * 70}{Colors.NC}")
    print(f"{Colors.GREEN}✓ GUI server installed as a user service{Colors.NC}")
    print(f"{Colors.BLUE}{'=' * 70}{Colors.NC}\n")
    print(f"{Colors.GREEN}Access the GUI at: http://localhost:3000{Colors.NC}")
    print(f"{Colors.YELLOW}The GUI server will start automatically on login{Colors.NC}\n")

    return 0

if __name__ == '__main__':
    try:
        sys.exit(main())
    except KeyboardInterrupt:
        print(f"\n{Colors.YELLOW}Installation cancelled by user{Colors.NC}")
        sys.exit(1)
    except Exception as e:
        print(f"{Colors.RED}✗ Error: {e}{Colors.NC}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
