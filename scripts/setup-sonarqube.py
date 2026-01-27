#!/usr/bin/env python3
"""
Setup SonarQube Community Edition for AI-Pack

This script:
1. Starts SonarQube CE via Docker Compose
2. Waits for it to be ready
3. Creates an admin token for API access
4. Saves configuration for agent use

Usage:
    python3 scripts/setup-sonarqube.py
"""

import os
import sys
import time
import json
import subprocess
import urllib.request
import urllib.error
import secrets
import string
from pathlib import Path
from base64 import b64encode


def generate_strong_password():
    """
    Generate a strong password in Apple-style format: xxxxx-xxxxx-xxxxx

    Uses cryptographic random function (secrets) to generate a 15-character
    password in 3 groups of 5, including uppercase, lowercase, numbers, and symbols.

    Returns:
        str: Strong password in format xxxxx-xxxxx-xxxxx
    """
    # Define character sets
    uppercase = string.ascii_uppercase
    lowercase = string.ascii_lowercase
    digits = string.digits
    symbols = "$*!@#()&"
    all_chars = uppercase + lowercase + digits + symbols

    def generate_group():
        """Generate a 5-character group with guaranteed complexity."""
        # Ensure at least one character from each type
        group = [
            secrets.choice(uppercase),
            secrets.choice(lowercase),
            secrets.choice(digits),
            secrets.choice(symbols),
            secrets.choice(all_chars)  # Fifth character can be any type
        ]
        # Shuffle to avoid predictable patterns
        secrets.SystemRandom().shuffle(group)
        return ''.join(group)

    # Generate 3 groups
    groups = [generate_group() for _ in range(3)]
    return '-'.join(groups)


# Configuration
SONARQUBE_URL = "http://localhost:9000"
ADMIN_USER = "admin"
ADMIN_PASS = "admin"
NEW_ADMIN_PASS = generate_strong_password()  # Strong cryptographic password
CONFIG_FILE = ".sonarqube-config"
COMPOSE_FILE = "docker-compose.sonarqube.yml"


# ANSI Colors
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'


def log_info(msg):
    """Log info message."""
    print(f"{Colors.BLUE}ℹ{Colors.NC} {msg}")


def log_success(msg):
    """Log success message."""
    print(f"{Colors.GREEN}✓{Colors.NC} {msg}")


def log_error(msg):
    """Log error message."""
    print(f"{Colors.RED}✗{Colors.NC} {msg}", file=sys.stderr)


def log_warn(msg):
    """Log warning message."""
    print(f"{Colors.YELLOW}⚠{Colors.NC} {msg}", file=sys.stderr)


def command_exists(cmd):
    """Check if a command exists."""
    try:
        subprocess.run(
            [cmd, "--version"],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            check=False
        )
        return True
    except FileNotFoundError:
        return False


def install_sonar_scanner():
    """Install sonar-scanner if not present."""
    if command_exists("sonar-scanner"):
        log_success("sonar-scanner already installed")
        return True

    log_info("sonar-scanner not found - attempting to install...")

    # Detect platform
    import platform
    system = platform.system()

    if system == "Darwin":  # macOS
        if command_exists("brew"):
            log_info("Installing sonar-scanner via Homebrew...")
            try:
                subprocess.run(
                    ["brew", "install", "sonar-scanner"],
                    check=True,
                    stdout=subprocess.DEVNULL,
                    stderr=subprocess.DEVNULL
                )
                log_success("sonar-scanner installed via Homebrew")
                return True
            except subprocess.CalledProcessError:
                log_error("Failed to install sonar-scanner via Homebrew")
                return False
        else:
            log_warn("Homebrew not found - cannot auto-install sonar-scanner")
            log_info("Install Homebrew: https://brew.sh")
            log_info("Then run: brew install sonar-scanner")
            return False

    elif system == "Linux":
        log_warn("Auto-install not supported on Linux")
        log_info("Install sonar-scanner:")
        log_info("  https://docs.sonarqube.org/latest/analysis/scan/sonarscanner/")
        return False

    elif system == "Windows":
        log_warn("Auto-install not supported on Windows")
        log_info("Download sonar-scanner:")
        log_info("  https://docs.sonarqube.org/latest/analysis/scan/sonarscanner/")
        log_info("  Extract and add to PATH")
        return False

    return False


def check_prerequisites():
    """Check that required tools are installed."""
    log_info("Checking prerequisites...")

    if not command_exists("docker"):
        log_error("Docker is not installed")
        log_info("Install Docker Desktop: https://www.docker.com/products/docker-desktop")
        sys.exit(1)

    # Check for docker compose (v2) or docker-compose (v1)
    has_compose = command_exists("docker-compose")
    try:
        result = subprocess.run(
            ["docker", "compose", "version"],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            check=False
        )
        has_compose = has_compose or result.returncode == 0
    except Exception:
        pass

    if not has_compose:
        log_error("Docker Compose is not installed")
        sys.exit(1)

    log_success("Prerequisites OK")

    # Check/install sonar-scanner
    if not install_sonar_scanner():
        log_warn("sonar-scanner not installed - manual installation required")
        log_info("You can still use SonarQube web UI, but validation script needs sonar-scanner")


def start_sonarqube():
    """Start SonarQube via Docker Compose."""
    log_info("Starting SonarQube Community Edition...")

    if not Path(COMPOSE_FILE).exists():
        log_error(f"Docker Compose file not found: {COMPOSE_FILE}")
        sys.exit(1)

    # Try docker compose (v2) first, fall back to docker-compose (v1)
    try:
        subprocess.run(
            ["docker", "compose", "-f", COMPOSE_FILE, "up", "-d"],
            check=True,
            capture_output=True
        )
    except subprocess.CalledProcessError:
        try:
            subprocess.run(
                ["docker-compose", "-f", COMPOSE_FILE, "up", "-d"],
                check=True,
                capture_output=True
            )
        except subprocess.CalledProcessError as e:
            log_error(f"Failed to start SonarQube: {e}")
            sys.exit(1)

    log_success("SonarQube containers started")


def wait_for_sonarqube():
    """Wait for SonarQube to be ready."""
    log_info("Waiting for SonarQube to be ready...")

    max_attempts = 60
    attempt = 0

    while attempt < max_attempts:
        try:
            req = urllib.request.Request(f"{SONARQUBE_URL}/api/system/status")
            with urllib.request.urlopen(req, timeout=5) as response:
                data = json.loads(response.read().decode())
                if data.get("status") == "UP":
                    print()  # New line after dots
                    log_success("SonarQube is ready!")
                    return
        except (urllib.error.URLError, Exception):
            pass

        print(".", end="", flush=True)
        time.sleep(5)
        attempt += 1

    print()
    log_error("SonarQube failed to start within 5 minutes")
    log_info("Check logs with: docker logs sonarqube")
    sys.exit(1)


def make_authenticated_request(url, username, password, method="GET", data=None):
    """Make an authenticated HTTP request to SonarQube API."""
    credentials = b64encode(f"{username}:{password}".encode()).decode()
    headers = {
        "Authorization": f"Basic {credentials}",
        "Content-Type": "application/x-www-form-urlencoded"
    }

    if data:
        data = urllib.parse.urlencode(data).encode()

    req = urllib.request.Request(url, headers=headers, data=data, method=method)

    try:
        with urllib.request.urlopen(req, timeout=10) as response:
            return response.read().decode(), response.status
    except urllib.error.HTTPError as e:
        return e.read().decode(), e.code


def check_password_changed():
    """Check if admin password has been changed."""
    try:
        req = urllib.request.Request(f"{SONARQUBE_URL}/api/authentication/validate")
        credentials = b64encode(f"{ADMIN_USER}:{NEW_ADMIN_PASS}".encode()).decode()
        req.add_header("Authorization", f"Basic {credentials}")

        with urllib.request.urlopen(req, timeout=5) as response:
            data = json.loads(response.read().decode())
            return data.get("valid", False)
    except Exception:
        return False


def change_admin_password():
    """Change default admin password."""
    log_info("Changing default admin password...")

    if check_password_changed():
        log_success("Admin password already changed")
        return True

    # Change password
    url = f"{SONARQUBE_URL}/api/users/change_password"
    data = {
        "login": ADMIN_USER,
        "previousPassword": ADMIN_PASS,
        "password": NEW_ADMIN_PASS
    }

    response, status_code = make_authenticated_request(
        url, ADMIN_USER, ADMIN_PASS, method="POST", data=data
    )

    if status_code == 204:
        log_success("Admin password changed")
        log_warn(f"New password: {NEW_ADMIN_PASS}")
        log_warn("Please change this in production!")
        return True
    else:
        log_error(f"Failed to change admin password (HTTP {status_code})")
        return False


def generate_token():
    """Generate API token for agents."""
    log_info("Generating API token for agents...")

    token_name = f"ai-pack-agent-{time.strftime('%Y%m%d')}"

    url = f"{SONARQUBE_URL}/api/user_tokens/generate"
    data = {"name": token_name}

    response, status_code = make_authenticated_request(
        url, ADMIN_USER, NEW_ADMIN_PASS, method="POST", data=data
    )

    if status_code == 200:
        try:
            token_data = json.loads(response)
            token = token_data.get("token")

            if token:
                # Save configuration
                config_content = f"""# SonarQube Configuration for AI-Pack
# Generated: {time.strftime('%Y-%m-%d %H:%M:%S')}

SONARQUBE_URL={SONARQUBE_URL}
SONARQUBE_TOKEN={token}
"""
                with open(CONFIG_FILE, "w") as f:
                    f.write(config_content)

                # Set restrictive permissions (Unix-like systems)
                try:
                    os.chmod(CONFIG_FILE, 0o600)
                except Exception:
                    pass

                log_success(f"API token generated and saved to {CONFIG_FILE}")
                log_warn("Keep this token secure!")
                print()
                log_info(f"Token: {token}")
                return True
        except json.JSONDecodeError:
            pass

    log_error("Failed to generate token")
    print(f"Response: {response}")
    return False


def show_next_steps():
    """Display next steps for the user."""
    print()
    print("=" * 70)
    print()
    log_success("SonarQube Community Edition is ready!")
    print()
    log_info(f"Web UI: {SONARQUBE_URL}")
    log_info(f"Username: {ADMIN_USER}")
    log_info(f"Password: {NEW_ADMIN_PASS}")
    print()
    log_info(f"Configuration saved to: {CONFIG_FILE}")
    print()
    log_info("Next steps:")
    print("  1. Validate code:")
    print("     python3 scripts/validate-with-sonarqube.py <file>")
    print()
    print("  2. Query rules:")
    print("     python3 scripts/query-rules.py --language go --severity CRITICAL")
    print()
    print("  3. Stop SonarQube:")
    print(f"     docker compose -f {COMPOSE_FILE} down")
    print()
    print("  4. View logs:")
    print("     docker logs sonarqube")
    print()
    print("=" * 70)


def main():
    """Main execution."""
    print("=" * 70)
    print("  SonarQube Community Edition Setup for AI-Pack")
    print("=" * 70)
    print()

    check_prerequisites()
    start_sonarqube()
    wait_for_sonarqube()
    change_admin_password()
    generate_token()
    show_next_steps()


if __name__ == "__main__":
    main()
