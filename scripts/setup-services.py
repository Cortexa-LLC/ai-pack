#!/usr/bin/env python3
"""
AI-Pack Service Setup Script
Manages background services via launchd (macOS) or systemd (Linux).
"""

import os
import sys
import shutil
import argparse
import subprocess
from pathlib import Path


class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    CYAN = '\033[0;36m'
    RESET = '\033[0m'


SERVER_LABEL = "com.cortexa.ai-pack.agent-server"
GUI_LABEL = "com.cortexa.ai-pack.gui"
BEADS_DOLT_LABEL = "com.beads.dolt-shared"
SERVER_UNIT = "ai-pack-agent-server"
GUI_UNIT = "ai-pack-gui"

BEADS_DOLT_DIR = Path.home() / '.beads' / 'dolt'


def get_project_root():
    return Path(__file__).parent.parent.absolute()


def get_uid():
    return os.getuid()


# ─── macOS / launchd ──────────────────────────────────────────────────────────

def get_launchd_dir():
    return Path.home() / 'Library' / 'LaunchAgents'


def _npm_path():
    found = shutil.which('npm')
    return found or '/opt/homebrew/bin/npm'


def _go_bin():
    return str(Path.home() / 'go' / 'bin')


def _base_path():
    return f"/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:{_go_bin()}"


def generate_beads_dolt_plist():
    dolt = shutil.which('dolt') or '/opt/homebrew/bin/dolt'
    data_dir = str(BEADS_DOLT_DIR)
    log_file = str(Path.home() / '.beads' / 'dolt-shared.log')
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{BEADS_DOLT_LABEL}</string>

    <key>ProgramArguments</key>
    <array>
        <string>{dolt}</string>
        <string>sql-server</string>
        <string>--host</string>
        <string>127.0.0.1</string>
        <string>--port</string>
        <string>3307</string>
        <string>--data-dir</string>
        <string>{data_dir}</string>
        <string>--loglevel</string>
        <string>warning</string>
    </array>

    <key>WorkingDirectory</key>
    <string>{data_dir}</string>

    <key>StandardOutPath</key>
    <string>{log_file}</string>

    <key>StandardErrorPath</key>
    <string>{log_file}</string>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
"""


def generate_server_plist(project_root):
    log_dir = project_root / 'logs'
    log_dir.mkdir(exist_ok=True)
    server_bin = project_root / 'bin' / 'agent-server'
    home = str(Path.home())
    api_key = os.environ.get('ANTHROPIC_API_KEY', '')
    server_port = os.environ.get('AIPACK_AGENT_SERVER_PORT', '8080')
    env_block = f"""
        <key>HOME</key>
        <string>{home}</string>
        <key>PATH</key>
        <string>{_base_path()}</string>
        <key>AIPACK_AGENT_SERVER_PORT</key>
        <string>{server_port}</string>"""
    if api_key:
        env_block += f"""
        <key>ANTHROPIC_API_KEY</key>
        <string>{api_key}</string>"""

    return f"""<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{SERVER_LABEL}</string>

    <key>ProgramArguments</key>
    <array>
        <string>{server_bin}</string>
        <string>--server</string>
    </array>

    <key>WorkingDirectory</key>
    <string>{project_root}</string>

    <key>EnvironmentVariables</key>
    <dict>{env_block}
    </dict>

    <key>StandardOutPath</key>
    <string>{log_dir}/agent-server.log</string>

    <key>StandardErrorPath</key>
    <string>{log_dir}/agent-server.error.log</string>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
        <key>Crashed</key>
        <true/>
    </dict>

    <key>ThrottleInterval</key>
    <integer>10</integer>

    <key>ProcessType</key>
    <string>Background</string>
</dict>
</plist>
"""


def generate_gui_plist(project_root):
    log_dir = project_root / 'logs'
    log_dir.mkdir(exist_ok=True)
    npm = _npm_path()
    home = str(Path.home())
    server_port = os.environ.get('AIPACK_AGENT_SERVER_PORT', '8080')

    return f"""<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{GUI_LABEL}</string>

    <key>ProgramArguments</key>
    <array>
        <string>{npm}</string>
        <string>run</string>
        <string>dev</string>
    </array>

    <key>WorkingDirectory</key>
    <string>{project_root}/gui</string>

    <key>EnvironmentVariables</key>
    <dict>
        <key>HOME</key>
        <string>{home}</string>
        <key>PATH</key>
        <string>{_base_path()}</string>
        <key>NODE_ENV</key>
        <string>development</string>
        <key>VITE_API_BASE_URL</key>
        <string>http://localhost:{server_port}</string>
    </dict>

    <key>StandardOutPath</key>
    <string>{log_dir}/gui.log</string>

    <key>StandardErrorPath</key>
    <string>{log_dir}/gui.error.log</string>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
        <key>Crashed</key>
        <true/>
    </dict>

    <key>ThrottleInterval</key>
    <integer>10</integer>

    <key>ProcessType</key>
    <string>Background</string>
</dict>
</plist>
"""


def _launchctl(*args, check=False):
    return subprocess.run(['launchctl', *args], capture_output=True, text=True, check=check)


def _domain():
    return f"gui/{get_uid()}"


def install_macos(project_root):
    launchd_dir = get_launchd_dir()
    launchd_dir.mkdir(parents=True, exist_ok=True)

    # Ensure shared Beads Dolt data directory exists before starting the service
    BEADS_DOLT_DIR.mkdir(parents=True, exist_ok=True)
    print(f"{Colors.GREEN}✅ Beads Dolt data dir: {BEADS_DOLT_DIR}{Colors.RESET}")

    plists = {
        BEADS_DOLT_LABEL: (launchd_dir / f"{BEADS_DOLT_LABEL}.plist", generate_beads_dolt_plist()),
        SERVER_LABEL:     (launchd_dir / f"{SERVER_LABEL}.plist",      generate_server_plist(project_root)),
        GUI_LABEL:        (launchd_dir / f"{GUI_LABEL}.plist",         generate_gui_plist(project_root)),
    }

    print(f"{Colors.BLUE}Installing launchd plists...{Colors.RESET}\n")
    for label, (path, content) in plists.items():
        path.write_text(content)
        print(f"{Colors.GREEN}✅ Written: {path}{Colors.RESET}")

        # bootout any existing instance first (ignore errors)
        _launchctl('bootout', _domain(), str(path))

        result = _launchctl('bootstrap', _domain(), str(path))
        if result.returncode == 0:
            print(f"{Colors.GREEN}✅ Bootstrapped: {label}{Colors.RESET}")
        else:
            err = result.stderr.strip()
            if 'already' in err.lower():
                print(f"{Colors.YELLOW}⚠️  Already loaded: {label}{Colors.RESET}")
            else:
                print(f"{Colors.RED}❌ Bootstrap failed: {err}{Colors.RESET}")
        print()

    print(f"{Colors.GREEN}✅ Services installed. Logs: {project_root}/logs/{Colors.RESET}")
    print(f"\nControl:\n  make start-all / stop-all / restart-all / status-services")


def uninstall_macos():
    launchd_dir = get_launchd_dir()
    for label in (BEADS_DOLT_LABEL, SERVER_LABEL, GUI_LABEL):
        plist_path = launchd_dir / f"{label}.plist"
        result = _launchctl('bootout', _domain(), label)
        if result.returncode == 0:
            print(f"{Colors.GREEN}✅ Stopped & unloaded: {label}{Colors.RESET}")
        else:
            print(f"{Colors.YELLOW}⚠️  Not running: {label}{Colors.RESET}")
        if plist_path.exists():
            plist_path.unlink()
            print(f"{Colors.GREEN}✅ Removed: {plist_path}{Colors.RESET}")
    print(f"\n{Colors.GREEN}✅ Services uninstalled{Colors.RESET}")


def start_macos(label):
    r = _launchctl('kickstart', '-k', f"{_domain()}/{label}")
    if r.returncode == 0:
        print(f"{Colors.GREEN}✅ Started: {label}{Colors.RESET}")
    else:
        err = r.stderr.strip()
        if 'not find' in err.lower() or 'no such' in err.lower():
            print(f"{Colors.RED}❌ Service not installed. Run: make setup-services{Colors.RESET}")
        else:
            print(f"{Colors.RED}❌ {err}{Colors.RESET}")
        sys.exit(1)


def stop_macos(label):
    r = _launchctl('kill', 'TERM', f"{_domain()}/{label}")
    if r.returncode == 0:
        print(f"{Colors.GREEN}✅ Stopped: {label}{Colors.RESET}")
    else:
        print(f"{Colors.YELLOW}⚠️  Not running: {label}{Colors.RESET}")


def status_macos(project_root):
    print(f"{Colors.BLUE}AI-Pack Service Status (macOS){Colors.RESET}\n")
    launchd_dir = get_launchd_dir()
    for label in (BEADS_DOLT_LABEL, SERVER_LABEL, GUI_LABEL):
        plist_path = launchd_dir / f"{label}.plist"
        installed = plist_path.exists()
        r = _launchctl('print', f"{_domain()}/{label}")
        running = r.returncode == 0
        pid = ''
        for line in r.stdout.splitlines():
            if 'pid' in line.lower() and '=' in line:
                pid = line.split('=')[-1].strip().rstrip(';')
                break

        status_str = f"{Colors.GREEN}running (PID {pid}){Colors.RESET}" if running else f"{Colors.YELLOW}stopped{Colors.RESET}"
        install_str = f"{Colors.GREEN}installed{Colors.RESET}" if installed else f"{Colors.RED}not installed{Colors.RESET}"
        print(f"  {label}")
        print(f"    Plist:  {install_str}")
        print(f"    Status: {status_str}")
        print()

    log_dir = project_root / 'logs'
    print(f"{Colors.CYAN}Logs: {log_dir}{Colors.RESET}")
    if log_dir.exists():
        for f in sorted(log_dir.glob('*.log')):
            print(f"  • {f.name} ({f.stat().st_size:,} bytes)")
    print()
    print(f"{Colors.CYAN}Commands:{Colors.RESET}")
    print(f"  make start-all / stop-all / restart-all / status-services")


# ─── Linux / systemd ──────────────────────────────────────────────────────────

def get_systemd_dir():
    return Path.home() / '.config' / 'systemd' / 'user'


def generate_server_unit(project_root):
    home = str(Path.home())
    server_bin = project_root / 'bin' / 'agent-server'
    api_key = os.environ.get('ANTHROPIC_API_KEY', '')
    env_line = f"\nEnvironment=ANTHROPIC_API_KEY={api_key}" if api_key else ""
    return f"""[Unit]
Description=AI-Pack Agent Server
After=network.target

[Service]
Type=simple
ExecStart={server_bin} --server
WorkingDirectory={project_root}
Restart=on-failure
RestartSec=5
Environment=HOME={home}
Environment=PATH=/usr/local/bin:/usr/bin:/bin:{home}/go/bin{env_line}

[Install]
WantedBy=default.target
"""


def generate_gui_unit(project_root):
    home = str(Path.home())
    npm = _npm_path()
    return f"""[Unit]
Description=AI-Pack GUI Dev Server
After=network.target

[Service]
Type=simple
ExecStart={npm} run dev
WorkingDirectory={project_root}/gui
Restart=on-failure
RestartSec=5
Environment=HOME={home}
Environment=PATH=/usr/local/bin:/usr/bin:/bin:{home}/.local/bin
Environment=NODE_ENV=development

[Install]
WantedBy=default.target
"""


def _systemctl(*args, check=False):
    return subprocess.run(['systemctl', '--user', *args], capture_output=True, text=True, check=check)


def install_linux(project_root):
    unit_dir = get_systemd_dir()
    unit_dir.mkdir(parents=True, exist_ok=True)

    log_dir = project_root / 'logs'
    log_dir.mkdir(exist_ok=True)

    units = {
        f"{SERVER_UNIT}.service": generate_server_unit(project_root),
        f"{GUI_UNIT}.service":    generate_gui_unit(project_root),
    }

    print(f"{Colors.BLUE}Installing systemd user units...{Colors.RESET}\n")
    for name, content in units.items():
        path = unit_dir / name
        path.write_text(content)
        print(f"{Colors.GREEN}✅ Written: {path}{Colors.RESET}")

    _systemctl('daemon-reload')

    for name in units:
        unit = name  # e.g. ai-pack-agent-server.service
        r = _systemctl('enable', '--now', unit)
        if r.returncode == 0:
            print(f"{Colors.GREEN}✅ Enabled & started: {unit}{Colors.RESET}")
        else:
            print(f"{Colors.RED}❌ Failed: {r.stderr.strip()}{Colors.RESET}")
        print()

    print(f"{Colors.GREEN}✅ Services installed{Colors.RESET}")
    print(f"\nControl:\n  make start-all / stop-all / restart-all / status-services")


def uninstall_linux():
    for unit in (f"{SERVER_UNIT}.service", f"{GUI_UNIT}.service"):
        _systemctl('disable', '--now', unit)
        path = get_systemd_dir() / unit
        if path.exists():
            path.unlink()
            print(f"{Colors.GREEN}✅ Removed: {path}{Colors.RESET}")
    _systemctl('daemon-reload')
    print(f"\n{Colors.GREEN}✅ Services uninstalled{Colors.RESET}")


def start_linux(unit):
    r = _systemctl('start', f"{unit}.service")
    if r.returncode == 0:
        print(f"{Colors.GREEN}✅ Started: {unit}{Colors.RESET}")
    else:
        err = r.stderr.strip()
        if 'not found' in err.lower() or 'no such' in err.lower():
            print(f"{Colors.RED}❌ Service not installed. Run: make setup-services{Colors.RESET}")
        else:
            print(f"{Colors.RED}❌ {err}{Colors.RESET}")
        sys.exit(1)


def stop_linux(unit):
    r = _systemctl('stop', f"{unit}.service")
    if r.returncode == 0:
        print(f"{Colors.GREEN}✅ Stopped: {unit}{Colors.RESET}")
    else:
        print(f"{Colors.YELLOW}⚠️  Not running: {unit}{Colors.RESET}")


def status_linux(project_root):
    print(f"{Colors.BLUE}AI-Pack Service Status (Linux){Colors.RESET}\n")
    for unit in (f"{SERVER_UNIT}.service", f"{GUI_UNIT}.service"):
        r = _systemctl('status', unit)
        active = 'active (running)' in r.stdout
        status_str = f"{Colors.GREEN}running{Colors.RESET}" if active else f"{Colors.YELLOW}stopped{Colors.RESET}"
        print(f"  {unit}: {status_str}")
        print()
    print(f"  make start-all / stop-all / restart-all / status-services")


# ─── Dispatch ─────────────────────────────────────────────────────────────────

def get_platform():
    if sys.platform == 'darwin':
        return 'macos'
    if sys.platform.startswith('linux'):
        return 'linux'
    return 'unsupported'


def main():
    parser = argparse.ArgumentParser(description='AI-Pack service management')
    parser.add_argument('action', choices=['install', 'uninstall', 'start-server', 'start-gui',
                                           'stop-server', 'stop-gui', 'status'])
    args = parser.parse_args()

    platform = get_platform()
    if platform == 'unsupported':
        print(f"{Colors.RED}❌ Unsupported platform: {sys.platform}{Colors.RESET}")
        print("Supported: macOS (launchd), Linux (systemd)")
        sys.exit(1)

    project_root = get_project_root()

    if args.action == 'install':
        if platform == 'macos':
            install_macos(project_root)
        else:
            install_linux(project_root)

    elif args.action == 'uninstall':
        if platform == 'macos':
            uninstall_macos()
        else:
            uninstall_linux()

    elif args.action == 'start-server':
        if platform == 'macos':
            start_macos(SERVER_LABEL)
        else:
            start_linux(SERVER_UNIT)

    elif args.action == 'start-gui':
        if platform == 'macos':
            start_macos(GUI_LABEL)
        else:
            start_linux(GUI_UNIT)

    elif args.action == 'stop-server':
        if platform == 'macos':
            stop_macos(SERVER_LABEL)
        else:
            stop_linux(SERVER_UNIT)

    elif args.action == 'stop-gui':
        if platform == 'macos':
            stop_macos(GUI_LABEL)
        else:
            stop_linux(GUI_UNIT)

    elif args.action == 'status':
        if platform == 'macos':
            status_macos(project_root)
        else:
            status_linux(project_root)


if __name__ == '__main__':
    main()
