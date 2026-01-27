#!/usr/bin/env python3
"""
Validate Code with SonarQube

Agent-friendly code validation using SonarQube Community Edition.
Runs sonar-scanner, fetches violations via API, enriches with rule metadata.

Usage:
    python3 scripts/validate-with-sonarqube.py <file_or_directory>
    python3 scripts/validate-with-sonarqube.py src/server.go --format json
    python3 scripts/validate-with-sonarqube.py . --project myproject --severity CRITICAL

Examples:
    # Validate single file
    python3 scripts/validate-with-sonarqube.py server.go

    # Validate directory
    python3 scripts/validate-with-sonarqube.py ./src

    # JSON output for agents
    python3 scripts/validate-with-sonarqube.py server.go --format json

    # Filter by severity
    python3 scripts/validate-with-sonarqube.py server.go --severity BLOCKER,CRITICAL
"""

import argparse
import csv
import json
import os
import subprocess
import sys
import time
import urllib.request
import urllib.error
from base64 import b64encode
from pathlib import Path
from typing import Dict, List, Optional, Any


# Configuration
DEFAULT_SONARQUBE_URL = "http://localhost:9000"
DEFAULT_CONFIG_FILE = ".sonarqube-config"
RULES_DIR = "quality/sonarqube-rules"


class SonarQubeValidator:
    """Validate code using SonarQube and enrich with rule metadata."""

    def __init__(self, sonarqube_url: str, token: str):
        """Initialize validator."""
        self.sonarqube_url = sonarqube_url.rstrip("/")
        self.token = token
        self.rules_cache = {}

    def load_rules_metadata(self, language: str) -> Dict[str, Dict]:
        """Load rule metadata from extracted CSV files."""
        if language in self.rules_cache:
            return self.rules_cache[language]

        rules_file = Path(RULES_DIR) / language / "rules.csv"
        if not rules_file.exists():
            return {}

        rules = {}
        try:
            with open(rules_file, "r", encoding="utf-8") as f:
                reader = csv.DictReader(f)
                for row in reader:
                    rules[row["rule_id"]] = row
        except Exception as e:
            print(f"Warning: Could not load rules for {language}: {e}", file=sys.stderr)

        self.rules_cache[language] = rules
        return rules

    def detect_language(self, file_path: str) -> Optional[str]:
        """Detect language from file extension."""
        ext = Path(file_path).suffix.lower()

        language_map = {
            ".go": "go",
            ".py": "python",
            ".java": "java",
            ".kt": "kotlin",
            ".swift": "swift",
            ".js": "javascript",
            ".ts": "javascript",
            ".jsx": "javascript",
            ".tsx": "javascript",
            ".cpp": "cpp",
            ".cc": "cpp",
            ".cxx": "cpp",
            ".c": "cpp",
            ".h": "cpp",
            ".hpp": "cpp",
            ".cs": "csharp",
            ".rb": "ruby",
            ".php": "php",
            ".scala": "scala",
            ".sh": "shell",
            ".bash": "shell",
        }

        return language_map.get(ext)

    def run_sonar_scanner(self, source_path: str, project_key: str) -> bool:
        """Run sonar-scanner on the source."""
        print(f"Running sonar-scanner on {source_path}...")

        # Check if sonar-scanner is installed
        try:
            subprocess.run(
                ["sonar-scanner", "-v"],
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
                check=True
            )
        except (FileNotFoundError, subprocess.CalledProcessError):
            print("Error: sonar-scanner not found", file=sys.stderr)
            print("Install from: https://docs.sonarqube.org/latest/analysis/scan/sonarscanner/", file=sys.stderr)
            return False

        # Prepare scanner arguments
        scanner_args = [
            "sonar-scanner",
            f"-Dsonar.projectKey={project_key}",
            f"-Dsonar.sources={source_path}",
            f"-Dsonar.host.url={self.sonarqube_url}",
            f"-Dsonar.login={self.token}",
        ]

        try:
            result = subprocess.run(
                scanner_args,
                capture_output=True,
                text=True,
                check=False
            )

            if result.returncode != 0:
                print("sonar-scanner output:", file=sys.stderr)
                print(result.stdout, file=sys.stderr)
                print(result.stderr, file=sys.stderr)
                return False

            return True
        except Exception as e:
            print(f"Error running sonar-scanner: {e}", file=sys.stderr)
            return False

    def fetch_violations(self, project_key: str, severity_filter: Optional[List[str]] = None) -> List[Dict]:
        """Fetch violations from SonarQube API."""
        print("Fetching violations from SonarQube...")

        url = f"{self.sonarqube_url}/api/issues/search"
        params = {
            "componentKeys": project_key,
            "resolved": "false",
            "ps": "500"  # Page size
        }

        if severity_filter:
            params["severities"] = ",".join(severity_filter)

        # Build URL with query parameters
        query_string = "&".join(f"{k}={v}" for k, v in params.items())
        full_url = f"{url}?{query_string}"

        # Make authenticated request
        req = urllib.request.Request(full_url)
        credentials = b64encode(f"{self.token}:".encode()).decode()
        req.add_header("Authorization", f"Basic {credentials}")

        try:
            with urllib.request.urlopen(req, timeout=30) as response:
                data = json.loads(response.read().decode())
                return data.get("issues", [])
        except urllib.error.HTTPError as e:
            print(f"HTTP Error {e.code}: {e.reason}", file=sys.stderr)
            return []
        except Exception as e:
            print(f"Error fetching violations: {e}", file=sys.stderr)
            return []

    def enrich_violation(self, violation: Dict, language: Optional[str] = None) -> Dict:
        """Enrich violation with rule metadata."""
        rule_key = violation.get("rule", "")
        rule_id = rule_key.split(":")[-1]  # Extract S#### from language:S####

        enriched = {
            "rule_id": rule_id,
            "rule_key": rule_key,
            "file": violation.get("component", "").split(":")[-1],
            "line": violation.get("line"),
            "message": violation.get("message"),
            "severity": violation.get("severity"),
            "type": violation.get("type"),
            "status": violation.get("status"),
        }

        # Add rule metadata if available
        if language:
            rules_metadata = self.load_rules_metadata(language)
            if rule_id in rules_metadata:
                metadata = rules_metadata[rule_id]
                enriched["rule_title"] = metadata.get("title")
                enriched["impacts"] = metadata.get("impacts")
                enriched["remediation_cost"] = metadata.get("remediation_cost")
                enriched["tags"] = metadata.get("tags", "").split(";")
                enriched["scope"] = metadata.get("scope")

        return enriched

    def validate(self, source_path: str, project_key: str, severity_filter: Optional[List[str]] = None, language: Optional[str] = None) -> Dict[str, Any]:
        """Validate source code and return violations."""
        # Detect language if not provided
        if not language and Path(source_path).is_file():
            language = self.detect_language(source_path)

        # Run scanner
        if not self.run_sonar_scanner(source_path, project_key):
            return {
                "success": False,
                "error": "sonar-scanner failed"
            }

        # Wait a moment for indexing
        time.sleep(2)

        # Fetch violations
        violations = self.fetch_violations(project_key, severity_filter)

        # Enrich violations
        enriched_violations = [
            self.enrich_violation(v, language)
            for v in violations
        ]

        # Calculate summary
        summary = {
            "total": len(enriched_violations),
            "by_severity": {},
            "by_type": {},
        }

        for v in enriched_violations:
            severity = v.get("severity", "UNKNOWN")
            vtype = v.get("type", "UNKNOWN")

            summary["by_severity"][severity] = summary["by_severity"].get(severity, 0) + 1
            summary["by_type"][vtype] = summary["by_type"].get(vtype, 0) + 1

        return {
            "success": True,
            "source": source_path,
            "project_key": project_key,
            "language": language,
            "violations": enriched_violations,
            "summary": summary
        }


def load_config() -> Dict[str, str]:
    """Load SonarQube configuration."""
    config_file = Path(DEFAULT_CONFIG_FILE)

    if not config_file.exists():
        return {}

    config = {}
    with open(config_file, "r") as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith("#"):
                if "=" in line:
                    key, value = line.split("=", 1)
                    config[key.strip()] = value.strip()

    return config


def format_output_text(result: Dict[str, Any]):
    """Format result as human-readable text."""
    if not result["success"]:
        print(f"Error: {result.get('error')}")
        return

    print(f"\nValidation Results for: {result['source']}")
    print(f"Language: {result.get('language', 'Unknown')}")
    print(f"Project: {result['project_key']}")
    print("\n" + "=" * 70)

    summary = result["summary"]
    print(f"\nSummary: {summary['total']} violations found")

    if summary["by_severity"]:
        print("\nBy Severity:")
        for severity in ["BLOCKER", "CRITICAL", "MAJOR", "MINOR", "INFO"]:
            count = summary["by_severity"].get(severity, 0)
            if count > 0:
                print(f"  {severity}: {count}")

    if summary["by_type"]:
        print("\nBy Type:")
        for vtype, count in summary["by_type"].items():
            print(f"  {vtype}: {count}")

    if result["violations"]:
        print("\n" + "=" * 70)
        print("\nViolations:")
        for i, v in enumerate(result["violations"], 1):
            print(f"\n{i}. [{v['severity']}] {v['rule_id']}")
            if v.get("rule_title"):
                print(f"   {v['rule_title']}")
            print(f"   File: {v['file']}" + (f", Line: {v['line']}" if v.get('line') else ""))
            print(f"   {v['message']}")
            if v.get("remediation_cost"):
                print(f"   Remediation: {v['remediation_cost']}")


def format_output_json(result: Dict[str, Any]):
    """Format result as JSON."""
    print(json.dumps(result, indent=2))


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="Validate code with SonarQube Community Edition",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Validate single file
  %(prog)s server.go

  # Validate directory with JSON output
  %(prog)s ./src --format json

  # Filter by severity
  %(prog)s server.go --severity BLOCKER,CRITICAL

  # Specify language explicitly
  %(prog)s script --language python
        """
    )

    parser.add_argument(
        "source",
        help="File or directory to validate"
    )

    parser.add_argument(
        "--project",
        default=None,
        help="Project key (default: generated from source path)"
    )

    parser.add_argument(
        "--severity",
        default=None,
        help="Filter by severity (comma-separated: BLOCKER,CRITICAL,MAJOR,MINOR,INFO)"
    )

    parser.add_argument(
        "--language",
        default=None,
        help="Language hint (go, python, java, etc.)"
    )

    parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format (default: text)"
    )

    parser.add_argument(
        "--url",
        default=None,
        help=f"SonarQube URL (default: from config or {DEFAULT_SONARQUBE_URL})"
    )

    parser.add_argument(
        "--token",
        default=None,
        help="API token (default: from config)"
    )

    args = parser.parse_args()

    # Load configuration
    config = load_config()

    # Get SonarQube URL and token
    sonarqube_url = args.url or config.get("SONARQUBE_URL", DEFAULT_SONARQUBE_URL)
    token = args.token or config.get("SONARQUBE_TOKEN")

    if not token:
        print("Error: No API token provided", file=sys.stderr)
        print(f"Run: python3 scripts/setup-sonarqube.py", file=sys.stderr)
        print(f"Or provide --token argument", file=sys.stderr)
        sys.exit(1)

    # Validate source exists
    source_path = Path(args.source)
    if not source_path.exists():
        print(f"Error: Source not found: {args.source}", file=sys.stderr)
        sys.exit(1)

    # Generate project key
    project_key = args.project
    if not project_key:
        # Use source path as project key (sanitized)
        project_key = str(source_path.resolve()).replace("/", "_").replace("\\", "_")
        project_key = "aipack_" + project_key.replace(".", "_")

    # Parse severity filter
    severity_filter = None
    if args.severity:
        severity_filter = [s.strip().upper() for s in args.severity.split(",")]

    # Create validator and run
    validator = SonarQubeValidator(sonarqube_url, token)
    result = validator.validate(
        str(source_path),
        project_key,
        severity_filter=severity_filter,
        language=args.language
    )

    # Output results
    if args.format == "json":
        format_output_json(result)
    else:
        format_output_text(result)

    # Exit with error code if violations found
    if result["success"] and result["summary"]["total"] > 0:
        sys.exit(1)


if __name__ == "__main__":
    main()
