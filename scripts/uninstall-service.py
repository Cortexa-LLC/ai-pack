#!/usr/bin/env python3
"""
AI-Pack Agent Server - Service Uninstallation
Removes the agent server user service
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

def uninstall_macos_service():
    """Uninstall macOS LaunchAgent"""
    print(f"{Colors.BLUE}Uninstalling macOS Launch Agent...{Colors.NC}\n")

    plist_path = Path.home() / 'Library' / 'LaunchAgents' / 'com.cortexa.ai-pack.agent-server.plist'

    if not plist_path.exists():
        print(f"{Colors.YELLOW}⚠ Service not installed{Colors.NC}")
        return True

    # Unload service
    result = subprocess.run(['launchctl', 'unload', str(plist_path)],
                           capture_output=True)

    if result.returncode == 0:
        print(f"{Colors.GREEN}✓ Service unloaded{Colors.NC}")
    else:
        print(f"{Colors.YELLOW}⚠ Service may not have been running{Colors.NC}")

    # Remove plist file
    plist_path.unlink()
    print(f"{Colors.GREEN}✓ Service file removed{Colors.NC}")

    return True

def uninstall_linux_service():
    """Uninstall Linux systemd user service"""
    print(f"{Colors.BLUE}Uninstalling systemd user service...{Colors.NC}\n")

    service_path = Path.home() / '.config' / 'systemd' / 'user' / 'ai-pack-agent-server.service'

    if not service_path.exists():
        print(f"{Colors.YELLOW}⚠ Service not installed{Colors.NC}")
        return True

    # Stop service
    subprocess.run(['systemctl', '--user', 'stop', 'ai-pack-agent-server.service'],
                  capture_output=True)
    print(f"{Colors.GREEN}✓ Service stopped{Colors.NC}")

    # Disable service
    subprocess.run(['systemctl', '--user', 'disable', 'ai-pack-agent-server.service'],
                  capture_output=True)
    print(f"{Colors.GREEN}✓ Service disabled{Colors.NC}")

    # Remove service file
    service_path.unlink()
    print(f"{Colors.GREEN}✓ Service file removed{Colors.NC}")

    # Reload systemd
    subprocess.run(['systemctl', '--user', 'daemon-reload'], check=True)
    print(f"{Colors.GREEN}✓ Systemd reloaded{Colors.NC}")

    return True

def main():
    """Main uninstallation flow"""
    print_header("AI-Pack Agent Server - Service Uninstallation")

    # Detect OS
    os_type = detect_os()
    if not os_type:
        print(f"{Colors.RED}✗ Unsupported operating system{Colors.NC}")
        return 1

    print(f"{Colors.BLUE}Detected OS: {os_type.upper()}{Colors.NC}\n")

    # Confirm uninstallation
    response = input(f"{Colors.YELLOW}Are you sure you want to uninstall the service? (y/n): {Colors.NC}").strip().lower()
    if response != 'y':
        print(f"{Colors.YELLOW}Uninstallation cancelled{Colors.NC}")
        return 0

    print()

    # Uninstall service based on OS
    if os_type == 'macos':
        success = uninstall_macos_service()
    else:  # linux
        success = uninstall_linux_service()

    if not success:
        return 1

    print(f"\n{Colors.BLUE}{'=' * 70}{Colors.NC}")
    print(f"{Colors.GREEN}✓ Agent server service uninstalled{Colors.NC}")
    print(f"{Colors.BLUE}{'=' * 70}{Colors.NC}\n")
    print(f"{Colors.YELLOW}You can manually start the server with:{Colors.NC}")
    print(f"  {Colors.BLUE}agent-server --server{Colors.NC}")
    print(f"{Colors.YELLOW}(If installed to /usr/local/bin, which is in your PATH){Colors.NC}\n")
    print(f"{Colors.YELLOW}To uninstall the binary itself:{Colors.NC}")
    print(f"  {Colors.BLUE}sudo make uninstall{Colors.NC}\n")

    return 0

if __name__ == '__main__':
    try:
        sys.exit(main())
    except KeyboardInterrupt:
        print(f"\n{Colors.YELLOW}Uninstallation cancelled by user{Colors.NC}")
        sys.exit(1)
    except Exception as e:
        print(f"{Colors.RED}✗ Error: {e}{Colors.NC}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
