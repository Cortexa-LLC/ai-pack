#!/usr/bin/env python3
"""
Setup SonarQube Community Edition with sonar-cxx plugin for C++ projects.

Performs full end-to-end setup:
  1. Start CE via sonarqube-switch.py (or docker-compose directly)
  2. Wait for operational status
  3. Change default admin password
  4. Generate analysis token -> sonar.env
  5. Download and install sonar-cxx plugin
  6. Restart SonarQube to load plugin
  7. Create project quality profile with all relevant rules activated
  8. Set quality profile as default for language
  9. Verify setup and print summary

Usage:
    # From project root (where sonarqube-switch.py lives):
    python3 .ai-pack/scripts/setup-sonarqube-ce.py

    # With options:
    python3 .ai-pack/scripts/setup-sonarqube-ce.py --profile "my-project" --language cxx
    python3 .ai-pack/scripts/setup-sonarqube-ce.py --skip-plugin   # plugin already installed
    python3 .ai-pack/scripts/setup-sonarqube-ce.py --skip-start    # SonarQube already running
"""

import argparse
import json
import subprocess
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from base64 import b64encode
from pathlib import Path


SONARQUBE_URL = "http://localhost:9000"
DEFAULT_ADMIN_USER = "admin"
DEFAULT_ADMIN_PASS = "admin"

SONAR_CXX_RELEASE_API = (
    "https://api.github.com/repos/SonarOpenCommunity/sonar-cxx/releases"
)
SONAR_CXX_TAG = "latest-snapshot"  # Use snapshot for SonarQube 25.1+ / 26.x

# Rule repositories to activate per language
LANGUAGE_RULE_REPOS = {
    "cxx": ["clangtidy", "cxx", "clangsa"],
    "py":  ["python"],
    "go":  ["go"],
    "java": ["java", "squid"],
}

# Rules to deactivate after activation — low signal / high noise
RULES_TO_DEACTIVATE = {
    "cxx": [
        "cxx:UndocumentedApi",           # Requires Doxygen on every public API — too noisy
        "cxx:InsufficientCommentDensity", # Comment density is project-style, not a bug
        "cxx:NoSonar",                    # Meta-rule for suppression markers
        "clangtidy:readability-identifier-naming",  # Enforced locally via .clang-tidy
    ],
}


# ---------------------------------------------------------------------------
# HTTP helpers
# ---------------------------------------------------------------------------

def _auth_header(user: str, password: str) -> dict:
    creds = b64encode(f"{user}:{password}".encode()).decode()
    return {"Authorization": f"Basic {creds}"}


def _token_header(token: str) -> dict:
    creds = b64encode(f"{token}:".encode()).decode()
    return {"Authorization": f"Basic {creds}"}


def api_get(path: str, token: str = None, user: str = None, password: str = None) -> dict:
    headers = {}
    if token:
        headers = _token_header(token)
    elif user:
        headers = _auth_header(user, password)
    req = urllib.request.Request(f"{SONARQUBE_URL}{path}", headers=headers)
    with urllib.request.urlopen(req, timeout=10) as r:
        return json.loads(r.read())


def api_post(path: str, data: dict, user: str, password: str) -> tuple[dict, int]:
    headers = _auth_header(user, password)
    headers["Content-Type"] = "application/x-www-form-urlencoded"
    body = urllib.parse.urlencode(data).encode()
    req = urllib.request.Request(
        f"{SONARQUBE_URL}{path}", data=body, headers=headers, method="POST"
    )
    try:
        with urllib.request.urlopen(req, timeout=15) as r:
            text = r.read().decode()
            return json.loads(text) if text.strip() else {}, r.status
    except urllib.error.HTTPError as e:
        text = e.read().decode()
        return json.loads(text) if text.strip() else {}, e.code


# ---------------------------------------------------------------------------
# Step helpers
# ---------------------------------------------------------------------------

def step(msg: str) -> None:
    print(f"\n{'='*60}")
    print(f"  {msg}")
    print(f"{'='*60}")


def ok(msg: str) -> None:
    print(f"  [OK] {msg}")


def warn(msg: str) -> None:
    print(f"  [WARN] {msg}", file=sys.stderr)


def fail(msg: str) -> None:
    print(f"  [FAIL] {msg}", file=sys.stderr)
    sys.exit(1)


# ---------------------------------------------------------------------------
# 1. Start SonarQube
# ---------------------------------------------------------------------------

def start_sonarqube(skip: bool) -> None:
    step("Start SonarQube Community Edition")
    if skip:
        ok("Skipped (--skip-start)")
        return

    if Path("sonarqube-switch.py").exists():
        print("  Using sonarqube-switch.py ce ...")
        subprocess.run(["python3", "sonarqube-switch.py", "ce"], check=True)
    elif Path(".ai-pack/docker-compose.sonarqube.yml").exists():
        print("  Using docker-compose directly ...")
        subprocess.run([
            "docker-compose",
            "-p", "sonarqube-ce",
            "-f", ".ai-pack/docker-compose.sonarqube.yml",
            "-f", "docker-compose.sonarqube-ce.yml",
            "up", "-d",
        ], check=True)
    else:
        fail("Cannot find sonarqube-switch.py or docker-compose files. Run from project root.")


# ---------------------------------------------------------------------------
# 2. Wait for operational
# ---------------------------------------------------------------------------

def wait_for_sonarqube() -> None:
    step("Wait for SonarQube to be operational")
    print("  Polling /api/system/status ...")
    for attempt in range(60):
        try:
            data = api_get("/api/system/status")
            if data.get("status") == "UP":
                ok(f"SonarQube is UP (version {data.get('version', '?')})")
                return
        except Exception:
            pass
        print(f"  [{attempt+1}/60] Not ready yet, waiting 5s ...", end="\r")
        time.sleep(5)
    fail("SonarQube did not become operational within 5 minutes.")


# ---------------------------------------------------------------------------
# 3. Change admin password
# ---------------------------------------------------------------------------

def verify_admin_credentials(password: str) -> bool:
    """Return True if admin credentials are valid."""
    try:
        data = api_get(
            "/api/authentication/validate",
            user=DEFAULT_ADMIN_USER,
            password=password,
        )
        return data.get("valid", False)
    except Exception:
        return False


def change_admin_password(new_pass: str) -> str:
    """
    Ensure admin credentials work and optionally change the password.

    SonarQube 26.x forces a password change on first web UI login, so the
    default 'admin' password may have already been changed by the user.
    If the provided password doesn't authenticate, we prompt interactively.
    """
    step("Configure admin credentials")

    # First, check if the provided password already works
    if verify_admin_credentials(new_pass):
        ok("Admin credentials verified.")
        return new_pass

    # If not, try the factory default
    if new_pass != DEFAULT_ADMIN_PASS and verify_admin_credentials(DEFAULT_ADMIN_PASS):
        # Factory default works — change it to the requested password
        data, code = api_post(
            "/api/users/change_password",
            {
                "login": DEFAULT_ADMIN_USER,
                "previousPassword": DEFAULT_ADMIN_PASS,
                "password": new_pass,
            },
            user=DEFAULT_ADMIN_USER,
            password=DEFAULT_ADMIN_PASS,
        )
        if code in (200, 204):
            ok(f"Admin password updated.")
            return new_pass
        else:
            warn(f"Could not change password (HTTP {code})")
            return DEFAULT_ADMIN_PASS

    # Neither works — SonarQube 26.x forces a password change on first web login.
    # Prompt interactively.
    print()
    print("  Admin password is not the default. SonarQube 26.x requires a")
    print("  password change on first web UI login.")
    print("  Pass the current password via --admin-password, or enter it now.")
    import getpass
    current = getpass.getpass("  Current admin password: ")
    if verify_admin_credentials(current):
        ok("Admin credentials verified.")
        return current
    else:
        fail("Admin credentials invalid. Run with --admin-password <current-password>.")


# ---------------------------------------------------------------------------
# 4. Generate token -> sonar.env
# ---------------------------------------------------------------------------

def generate_token(admin_pass: str, env_file: str) -> str:
    step(f"Generate analysis token -> {env_file}")

    # Check existing
    env_path = Path(env_file)
    if env_path.exists():
        for line in env_path.read_text().splitlines():
            if line.startswith("SONAR_TOKEN="):
                token = line.split("=", 1)[1].strip()
                # Verify it works
                try:
                    data = api_get("/api/authentication/validate", token=token)
                    if data.get("valid"):
                        ok(f"Existing token in {env_file} is valid, skipping.")
                        return token
                except Exception:
                    pass

    token_name = f"setup-{int(time.time())}"
    data, code = api_post(
        "/api/user_tokens/generate",
        {"name": token_name, "type": "GLOBAL_ANALYSIS_TOKEN"},
        user=DEFAULT_ADMIN_USER,
        password=admin_pass,
    )
    if code != 200 or "token" not in data:
        fail(f"Failed to generate token (HTTP {code}): {data}")

    token = data["token"]
    env_path.write_text(f"SONAR_HOST_URL={SONARQUBE_URL}\nSONAR_TOKEN={token}\n")
    env_path.chmod(0o600)

    # Ensure gitignored
    gitignore = Path(".gitignore")
    if gitignore.exists() and env_file not in gitignore.read_text():
        with gitignore.open("a") as f:
            f.write(f"\n{env_file}\n")
        ok(f"Added {env_file} to .gitignore")

    ok(f"Token saved to {env_file}")
    return token


# ---------------------------------------------------------------------------
# 5. Install sonar-cxx plugin
# ---------------------------------------------------------------------------

def get_sonar_cxx_jar_url() -> tuple[str, str]:
    """Return (download_url, filename) for the sonar-cxx snapshot JAR."""
    req = urllib.request.Request(
        SONAR_CXX_RELEASE_API,
        headers={"User-Agent": "setup-sonarqube-ce"}
    )
    releases = json.loads(urllib.request.urlopen(req, timeout=15).read())
    for release in releases:
        if release["tag_name"] == SONAR_CXX_TAG:
            for asset in release.get("assets", []):
                if asset["name"].endswith(".jar") and "plugin" in asset["name"]:
                    return asset["browser_download_url"], asset["name"]
    fail(f"Could not find sonar-cxx release tag '{SONAR_CXX_TAG}' on GitHub.")


def install_sonar_cxx_plugin(skip: bool) -> None:
    step("Install sonar-cxx C++ plugin")

    if skip:
        ok("Skipped (--skip-plugin)")
        return

    # Check if already installed
    try:
        data = api_get(
            "/api/plugins/installed",
            user=DEFAULT_ADMIN_USER,
            password=DEFAULT_ADMIN_PASS,
        )
        for plugin in data.get("plugins", []):
            if plugin.get("key") == "cxx":
                ok(f"sonar-cxx already installed (v{plugin.get('version', '?')})")
                return
    except Exception:
        pass

    url, filename = get_sonar_cxx_jar_url()
    jar_path = Path(f"/tmp/{filename}")

    print(f"  Downloading {filename} ...")
    urllib.request.urlretrieve(url, jar_path)
    size_mb = jar_path.stat().st_size / 1_048_576
    ok(f"Downloaded {filename} ({size_mb:.1f} MB)")

    print("  Copying plugin into container ...")
    subprocess.run([
        "docker", "cp",
        str(jar_path),
        f"sonarqube:/opt/sonarqube/extensions/plugins/{filename}",
    ], check=True)
    ok("Plugin copied to container.")

    print("  Restarting SonarQube to load plugin ...")
    subprocess.run(["docker", "restart", "sonarqube"], check=True)
    time.sleep(10)
    wait_for_sonarqube()


# ---------------------------------------------------------------------------
# 6. Create quality profile and activate rules
# ---------------------------------------------------------------------------

def setup_quality_profile(profile_name: str, language: str, admin_pass: str) -> None:
    step(f"Configure quality profile '{profile_name}' for language '{language}'")

    # Check if profile already exists
    data = api_get(
        f"/api/qualityprofiles/search?language={language}",
        user=DEFAULT_ADMIN_USER,
        password=admin_pass,
    )
    profile_key = None
    for p in data.get("profiles", []):
        if p["name"] == profile_name:
            profile_key = p["key"]
            ok(f"Profile '{profile_name}' already exists (key={profile_key})")
            break

    if not profile_key:
        data, code = api_post(
            "/api/qualityprofiles/create",
            {"name": profile_name, "language": language},
            user=DEFAULT_ADMIN_USER,
            password=admin_pass,
        )
        if code != 200 or "profile" not in data:
            fail(f"Failed to create quality profile (HTTP {code}): {data}")
        profile_key = data["profile"]["key"]
        ok(f"Created profile '{profile_name}' (key={profile_key})")

    # Activate rules for each repository
    repos = LANGUAGE_RULE_REPOS.get(language, [])
    if not repos:
        warn(f"No rule repositories configured for language '{language}'.")
        return

    total_activated = 0
    for repo in repos:
        data, code = api_post(
            "/api/qualityprofiles/activate_rules",
            {"targetKey": profile_key, "repositories": repo},
            user=DEFAULT_ADMIN_USER,
            password=admin_pass,
        )
        succeeded = data.get("succeeded", 0)
        failed = data.get("failed", 0)
        total_activated += succeeded
        ok(f"  [{repo}] activated={succeeded} failed={failed}")

    ok(f"Total rules activated: {total_activated}")

    # Deactivate noisy low-signal rules
    noisy = RULES_TO_DEACTIVATE.get(language, [])
    for rule in noisy:
        data, code = api_post(
            "/api/qualityprofiles/deactivate_rule",
            {"key": profile_key, "rule": rule},
            user=DEFAULT_ADMIN_USER,
            password=admin_pass,
        )
        if code in (200, 204):
            ok(f"  Deactivated: {rule}")
        else:
            warn(f"  Could not deactivate {rule} (HTTP {code})")

    # Set as default
    data, code = api_post(
        "/api/qualityprofiles/set_default",
        {"qualityProfile": profile_name, "language": language},
        user=DEFAULT_ADMIN_USER,
        password=admin_pass,
    )
    if code in (200, 204):
        ok(f"Profile '{profile_name}' set as default for language '{language}'")
    else:
        warn(f"Could not set profile as default (HTTP {code}): {data}")


# ---------------------------------------------------------------------------
# 7. Verify and summarise
# ---------------------------------------------------------------------------

CUSTOM_GATE_NAME = "ai-pack"


def configure_quality_gate(admin_pass: str) -> None:
    """
    Create a custom quality gate that checks both new AND overall violations.

    The built-in 'Sonar way' gate is immutable and only checks new_violations,
    so pre-existing issues never surface as gate failures. We create a custom
    gate 'ai-pack' with both conditions and set it as the project default.
    """
    step(f"Configure quality gate '{CUSTOM_GATE_NAME}'")

    # Check if our custom gate already exists
    data = api_get("/api/qualitygates/list", user=DEFAULT_ADMIN_USER, password=admin_pass)
    existing = next(
        (g for g in data.get("qualitygates", []) if g["name"] == CUSTOM_GATE_NAME), None
    )

    if existing:
        ok(f"Quality gate '{CUSTOM_GATE_NAME}' already exists.")
        _ensure_gate_default(admin_pass)
        return

    # Create the gate
    data, code = api_post(
        "/api/qualitygates/create",
        {"name": CUSTOM_GATE_NAME},
        user=DEFAULT_ADMIN_USER,
        password=admin_pass,
    )
    if code not in (200, 201):
        warn(f"Could not create quality gate (HTTP {code}): {data}")
        return
    ok(f"Created quality gate '{CUSTOM_GATE_NAME}'")

    # Add conditions
    conditions = [
        # Fail if any new violations introduced
        {"metric": "new_violations", "op": "GT", "error": "0"},
        # Fail if overall violation count exceeds threshold (catches legacy debt accumulation)
        {"metric": "violations", "op": "GT", "error": "0"},
    ]
    for cond in conditions:
        cond["gateName"] = CUSTOM_GATE_NAME
        data, code = api_post(
            "/api/qualitygates/create_condition",
            cond,
            user=DEFAULT_ADMIN_USER,
            password=admin_pass,
        )
        label = f"{cond['metric']} {cond['op']} {cond['error']}"
        if code in (200, 201):
            ok(f"  Added condition: {label}")
        else:
            warn(f"  Could not add condition '{label}' (HTTP {code}): {data}")

    _ensure_gate_default(admin_pass)


def _ensure_gate_default(admin_pass: str) -> None:
    """Set the custom gate as the default."""
    data, code = api_post(
        "/api/qualitygates/set_as_default",
        {"name": CUSTOM_GATE_NAME},
        user=DEFAULT_ADMIN_USER,
        password=admin_pass,
    )
    if code in (200, 204):
        ok(f"Quality gate '{CUSTOM_GATE_NAME}' set as default.")
    else:
        warn(f"Could not set gate as default (HTTP {code}): {data}")


def verify_setup(token: str, profile_name: str, language: str, admin_pass: str) -> None:
    step("Verify setup")

    # Server version
    try:
        version = urllib.request.urlopen(
            f"{SONARQUBE_URL}/api/server/version", timeout=5
        ).read().decode()
        ok(f"SonarQube version: {version}")
    except Exception as e:
        warn(f"Could not fetch version: {e}")

    # sonar-cxx plugin
    try:
        data = api_get("/api/plugins/installed", user=DEFAULT_ADMIN_USER, password=admin_pass)
        cxx = next((p for p in data.get("plugins", []) if p.get("key") == "cxx"), None)
        if cxx:
            ok(f"sonar-cxx plugin: v{cxx.get('version', '?')}")
        else:
            warn("sonar-cxx plugin not found in installed plugins list.")
    except Exception as e:
        warn(f"Could not check plugins: {e}")

    # Quality profile
    try:
        data = api_get(
            f"/api/qualityprofiles/search?language={language}",
            user=DEFAULT_ADMIN_USER, password=admin_pass,
        )
        for p in data.get("profiles", []):
            if p["name"] == profile_name:
                ok(f"Quality profile '{profile_name}': {p['activeRuleCount']} active rules, default={p['isDefault']}")
    except Exception as e:
        warn(f"Could not check quality profile: {e}")

    print()
    print("=" * 60)
    print("  Setup complete.")
    print()
    print(f"  Dashboard:  {SONARQUBE_URL}")
    print(f"  To scan:    source sonar.env && sonar-scanner")
    print("=" * 60)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def detect_language() -> str:
    """Auto-detect primary language from project files."""
    if Path("CMakeLists.txt").exists() or list(Path(".").glob("*.cmake")):
        return "cxx"
    if Path("go.mod").exists():
        return "go"
    if Path("pyproject.toml").exists() or Path("setup.py").exists():
        return "py"
    if Path("pom.xml").exists() or Path("build.gradle").exists():
        return "java"
    return "cxx"  # fallback


def detect_profile_name() -> str:
    """Read sonar.projectKey from sonar-project.properties, else use directory name."""
    props = Path("sonar-project.properties")
    if props.exists():
        for line in props.read_text().splitlines():
            if line.startswith("sonar.projectKey="):
                return line.split("=", 1)[1].strip()
    return Path.cwd().name


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Setup SonarQube CE with sonar-cxx plugin and quality profile."
    )
    parser.add_argument(
        "--profile",
        default=None,
        help="Quality profile name (default: auto-detected from sonar-project.properties)",
    )
    parser.add_argument(
        "--language",
        default=None,
        choices=list(LANGUAGE_RULE_REPOS.keys()),
        help="Primary language (default: auto-detected from project files)",
    )
    parser.add_argument(
        "--admin-password",
        default=DEFAULT_ADMIN_PASS,
        help="New admin password (default: keep as 'admin' for local dev)",
    )
    parser.add_argument(
        "--env-file",
        default="sonar.env",
        help="File to write SONAR_TOKEN into (default: sonar.env)",
    )
    parser.add_argument(
        "--skip-start",
        action="store_true",
        help="Skip starting SonarQube (already running)",
    )
    parser.add_argument(
        "--skip-plugin",
        action="store_true",
        help="Skip sonar-cxx plugin download/install",
    )
    args = parser.parse_args()

    profile = args.profile or detect_profile_name()
    language = args.language or detect_language()

    # Read admin password from sonar.env if not overridden on CLI
    if args.admin_password == DEFAULT_ADMIN_PASS and os.path.exists(args.env_file):
        for line in open(args.env_file):
            k, _, v = line.strip().partition("=")
            if k == "SONAR_ADMIN_PASS" and v:
                args.admin_password = v
                break

    print()
    print("=" * 60)
    print("  SonarQube CE Setup")
    print(f"  Profile:  {profile}  Language: {language}")
    print("=" * 60)

    start_sonarqube(args.skip_start)
    wait_for_sonarqube()
    admin_pass = change_admin_password(args.admin_password)
    token = generate_token(admin_pass, args.env_file)
    install_sonar_cxx_plugin(args.skip_plugin)
    setup_quality_profile(profile, language, admin_pass)
    configure_quality_gate(admin_pass)
    verify_setup(token, profile, language, admin_pass)


if __name__ == "__main__":
    main()
