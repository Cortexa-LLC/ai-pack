#!/usr/bin/env python3
"""
AI-Pack Agent Server - Service Installation
Installs the agent server as a user service (macOS LaunchAgent or Linux systemd user service)
"""

import os
import sys
import platform
import subprocess
import shutil
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

def check_binary(project_root):
    """Check if agent-server binary is installed"""
    installed_binary = Path('/usr/local/bin/agent-server')
    local_binary = project_root / 'a2a-agent' / 'bin' / 'agent-server'

    if installed_binary.exists():
        print(f"{Colors.GREEN}✓ Agent server found at /usr/local/bin/agent-server{Colors.NC}\n")
        return True

    print(f"{Colors.YELLOW}⚠ Agent server not found at /usr/local/bin/agent-server{Colors.NC}")
    print(f"{Colors.BLUE}You need to install the binary first:{Colors.NC}")
    print(f"  {Colors.BOLD}cd {project_root}{Colors.NC}")
    print(f"  {Colors.BOLD}make build{Colors.NC}")
    print(f"  {Colors.BOLD}sudo make install{Colors.NC}\n")

    return False

def check_api_keys():
    """Check if required API keys are set"""
    print(f"{Colors.BLUE}Checking API keys...{Colors.NC}")

    # Load from shell config
    home = Path.home()
    for config_file in ['.bash_profile', '.bashrc', '.zshrc']:
        config_path = home / config_file
        if config_path.exists():
            try:
                with open(config_path, 'r') as f:
                    for line in f:
                        if line.startswith('export OPENAI_API_KEY='):
                            key = line.split('=', 1)[1].strip().strip('"').strip("'")
                            if key and not os.environ.get('OPENAI_API_KEY'):
                                os.environ['OPENAI_API_KEY'] = key
                        elif line.startswith('export ANTHROPIC_API_KEY='):
                            key = line.split('=', 1)[1].strip().strip('"').strip("'")
                            if key and not os.environ.get('ANTHROPIC_API_KEY'):
                                os.environ['ANTHROPIC_API_KEY'] = key
            except Exception:
                pass

    openai_key = os.environ.get('OPENAI_API_KEY')
    anthropic_key = os.environ.get('ANTHROPIC_API_KEY')

    if not openai_key:
        print(f"{Colors.YELLOW}⚠ OPENAI_API_KEY not set{Colors.NC}")

    if not anthropic_key:
        print(f"{Colors.RED}✗ ANTHROPIC_API_KEY not set (required){Colors.NC}")
        print(f"{Colors.YELLOW}  Run: python3 ./scripts/setup-api-keys.py{Colors.NC}")
        return False, None, None

    print(f"{Colors.GREEN}✓ API keys configured{Colors.NC}\n")
    return True, openai_key, anthropic_key

def install_macos_service(project_root, openai_key, anthropic_key):
    """Install macOS LaunchAgent (user service)"""
    print(f"{Colors.BLUE}Installing macOS Launch Agent (user service)...{Colors.NC}\n")

    # Paths
    plist_template = project_root / 'config' / 'com.cortexa.ai-pack.agent-server.plist'
    launch_agents_dir = Path.home() / 'Library' / 'LaunchAgents'
    install_plist = launch_agents_dir / 'com.cortexa.ai-pack.agent-server.plist'

    # Create LaunchAgents directory
    launch_agents_dir.mkdir(parents=True, exist_ok=True)

    # Read template
    with open(plist_template, 'r') as f:
        plist_content = f.read()

    # Replace placeholders
    plist_content = plist_content.replace('{{PROJECT_ROOT}}', str(project_root))
    plist_content = plist_content.replace('{{HOME}}', str(Path.home()))

    # Write to install location
    with open(install_plist, 'w') as f:
        f.write(plist_content)

    print(f"{Colors.GREEN}✓ Plist installed to: {install_plist}{Colors.NC}")

    # Add environment variables using PlistBuddy
    if openai_key:
        try:
            subprocess.run([
                '/usr/libexec/PlistBuddy',
                '-c', f'Add :EnvironmentVariables:OPENAI_API_KEY string {openai_key}',
                str(install_plist)
            ], check=False, capture_output=True)
        except Exception:
            subprocess.run([
                '/usr/libexec/PlistBuddy',
                '-c', f'Set :EnvironmentVariables:OPENAI_API_KEY {openai_key}',
                str(install_plist)
            ], check=False, capture_output=True)

    if anthropic_key:
        try:
            subprocess.run([
                '/usr/libexec/PlistBuddy',
                '-c', f'Add :EnvironmentVariables:ANTHROPIC_API_KEY string {anthropic_key}',
                str(install_plist)
            ], check=False, capture_output=True)
        except Exception:
            subprocess.run([
                '/usr/libexec/PlistBuddy',
                '-c', f'Set :EnvironmentVariables:ANTHROPIC_API_KEY {anthropic_key}',
                str(install_plist)
            ], check=False, capture_output=True)

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
    time.sleep(2)

    result = subprocess.run(['launchctl', 'list'], capture_output=True, text=True)
    if 'com.cortexa.ai-pack.agent-server' in result.stdout:
        print(f"{Colors.GREEN}✓ Service is running{Colors.NC}\n")
    else:
        print(f"{Colors.YELLOW}⚠ Service loaded but may not be running{Colors.NC}")
        print(f"{Colors.YELLOW}  Check logs: tail -f /tmp/agent-server.log{Colors.NC}\n")

    # Print management commands
    print(f"{Colors.BLUE}Management commands:{Colors.NC}")
    print(f"  Start:   {Colors.GREEN}launchctl load {install_plist}{Colors.NC}")
    print(f"  Stop:    {Colors.RED}launchctl unload {install_plist}{Colors.NC}")
    print(f"  Restart: {Colors.YELLOW}launchctl unload {install_plist} && launchctl load {install_plist}{Colors.NC}")
    print(f"  Logs:    {Colors.BLUE}tail -f /tmp/agent-server.log{Colors.NC}")

    return True

def install_linux_service(project_root, openai_key, anthropic_key):
    """Install Linux systemd user service"""
    print(f"{Colors.BLUE}Installing systemd user service...{Colors.NC}\n")

    # Paths
    service_template = project_root / 'config' / 'ai-pack-agent-server.service'
    systemd_user_dir = Path.home() / '.config' / 'systemd' / 'user'
    install_service = systemd_user_dir / 'ai-pack-agent-server.service'

    # Create systemd user directory
    systemd_user_dir.mkdir(parents=True, exist_ok=True)

    # Read template
    with open(service_template, 'r') as f:
        service_content = f.read()

    # Replace placeholders
    service_content = service_content.replace('{{PROJECT_ROOT}}', str(project_root))
    service_content = service_content.replace('{{USER}}', os.getenv("USER"))

    # Add environment variables
    env_lines = []
    if openai_key:
        env_lines.append(f'Environment="OPENAI_API_KEY={openai_key}"')
    if anthropic_key:
        env_lines.append(f'Environment="ANTHROPIC_API_KEY={anthropic_key}"')

    if env_lines:
        # Insert environment variables after the Environment="PATH=..." line
        lines = service_content.split('\n')
        insert_idx = None
        for i, line in enumerate(lines):
            if line.startswith('Environment="PATH='):
                insert_idx = i + 1
                break

        if insert_idx:
            lines = lines[:insert_idx] + env_lines + lines[insert_idx:]
            service_content = '\n'.join(lines)

    # Write service file
    with open(install_service, 'w') as f:
        f.write(service_content)

    print(f"{Colors.GREEN}✓ Service file installed to: {install_service}{Colors.NC}")

    # Reload systemd
    subprocess.run(['systemctl', '--user', 'daemon-reload'], check=True)

    # Enable and start service
    subprocess.run(['systemctl', '--user', 'enable', 'ai-pack-agent-server.service'], check=True)
    subprocess.run(['systemctl', '--user', 'start', 'ai-pack-agent-server.service'], check=True)

    print(f"{Colors.GREEN}✓ Service enabled and started{Colors.NC}\n")

    # Check status
    result = subprocess.run(['systemctl', '--user', 'is-active', 'ai-pack-agent-server.service'],
                           capture_output=True, text=True)

    if result.stdout.strip() == 'active':
        print(f"{Colors.GREEN}✓ Service is running{Colors.NC}\n")
    else:
        print(f"{Colors.YELLOW}⚠ Service may not be running{Colors.NC}")
        subprocess.run(['systemctl', '--user', 'status', 'ai-pack-agent-server.service', '--no-pager'])

    # Print management commands
    print(f"\n{Colors.BLUE}Management commands:{Colors.NC}")
    print(f"  Start:   {Colors.GREEN}systemctl --user start ai-pack-agent-server{Colors.NC}")
    print(f"  Stop:    {Colors.RED}systemctl --user stop ai-pack-agent-server{Colors.NC}")
    print(f"  Restart: {Colors.YELLOW}systemctl --user restart ai-pack-agent-server{Colors.NC}")
    print(f"  Status:  {Colors.BLUE}systemctl --user status ai-pack-agent-server{Colors.NC}")
    print(f"  Logs:    {Colors.BLUE}journalctl --user -u ai-pack-agent-server -f{Colors.NC}")
    print(f"  Disable: {Colors.RED}systemctl --user disable ai-pack-agent-server{Colors.NC}")

    return True

def main():
    """Main installation flow"""
    print_header("AI-Pack Agent Server - Service Installation")

    # Detect OS
    os_type = detect_os()
    if not os_type:
        print(f"{Colors.RED}✗ Unsupported operating system{Colors.NC}")
        return 1

    print(f"{Colors.BLUE}Detected OS: {os_type.upper()}{Colors.NC}\n")

    # Get project root
    project_root = get_project_root()
    print(f"{Colors.BLUE}Project root: {project_root}{Colors.NC}\n")

    # Check binary
    if not check_binary(project_root):
        return 1

    # Check API keys
    keys_ok, openai_key, anthropic_key = check_api_keys()
    if not keys_ok:
        return 1

    # Install service based on OS
    if os_type == 'macos':
        success = install_macos_service(project_root, openai_key, anthropic_key)
    else:  # linux
        success = install_linux_service(project_root, openai_key, anthropic_key)

    if not success:
        return 1

    print(f"\n{Colors.BLUE}{'=' * 70}{Colors.NC}")
    print(f"{Colors.GREEN}✓ Agent server installed as a user service{Colors.NC}")
    print(f"{Colors.BLUE}{'=' * 70}{Colors.NC}\n")
    print(f"{Colors.YELLOW}The agent server will now start automatically on login{Colors.NC}\n")

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
