#!/usr/bin/env python3
"""
Query SonarQube Rules

Search, filter, and explain SonarQube rules from extracted metadata.

Usage:
    python3 scripts/query-rules.py --language go
    python3 scripts/query-rules.py --language python --severity CRITICAL
    python3 scripts/query-rules.py --rule S1192
    python3 scripts/query-rules.py --language java --type BUG --format json

Examples:
    # List all Go rules
    python3 scripts/query-rules.py --language go

    # Find critical bugs in Python
    python3 scripts/query-rules.py --language python --type BUG --severity CRITICAL

    # Explain a specific rule
    python3 scripts/query-rules.py --rule S1192

    # Find all security-related rules
    python3 scripts/query-rules.py --language go --impact SECURITY

    # Export as JSON
    python3 scripts/query-rules.py --language java --format json > java_bugs.json
"""

import argparse
import csv
import json
import sys
from pathlib import Path
from typing import List, Dict, Optional


RULES_DIR = Path("quality/sonarqube-rules")
SUPPORTED_LANGUAGES = ["cpp", "csharp", "go", "python", "javascript", "java", "kotlin", "swift"]


class RuleQuery:
    """Query and filter SonarQube rules."""

    def __init__(self, rules_dir: Path = RULES_DIR):
        """Initialize rule query."""
        self.rules_dir = rules_dir
        self.rules_cache = {}

    def load_rules(self, language: str) -> List[Dict]:
        """Load rules for a language."""
        if language in self.rules_cache:
            return self.rules_cache[language]

        rules_file = self.rules_dir / language / "rules.csv"
        if not rules_file.exists():
            return []

        rules = []
        try:
            with open(rules_file, "r", encoding="utf-8") as f:
                reader = csv.DictReader(f)
                for row in reader:
                    rules.append(row)
        except Exception as e:
            print(f"Error loading rules for {language}: {e}", file=sys.stderr)
            return []

        self.rules_cache[language] = rules
        return rules

    def find_rule_by_id(self, rule_id: str) -> Optional[Dict]:
        """Find a rule by ID across all languages."""
        for language in SUPPORTED_LANGUAGES:
            rules = self.load_rules(language)
            for rule in rules:
                if rule["rule_id"] == rule_id:
                    return rule
        return None

    def filter_rules(
        self,
        language: Optional[str] = None,
        rule_type: Optional[str] = None,
        severity: Optional[str] = None,
        impact: Optional[str] = None,
        tags: Optional[List[str]] = None,
        status: Optional[str] = None,
        search: Optional[str] = None
    ) -> List[Dict]:
        """Filter rules by criteria."""
        # Determine which languages to search
        if language:
            languages = [language]
        else:
            languages = SUPPORTED_LANGUAGES

        results = []
        for lang in languages:
            rules = self.load_rules(lang)

            for rule in rules:
                # Apply filters
                if rule_type and rule.get("type") != rule_type:
                    continue

                if severity and rule.get("severity") != severity:
                    continue

                if impact:
                    rule_impacts = rule.get("impacts", "")
                    if impact not in rule_impacts:
                        continue

                if tags:
                    rule_tags = rule.get("tags", "").split(";")
                    if not any(tag in rule_tags for tag in tags):
                        continue

                if status and rule.get("status") != status:
                    continue

                if search:
                    search_lower = search.lower()
                    title = rule.get("title", "").lower()
                    rule_id = rule.get("rule_id", "").lower()
                    if search_lower not in title and search_lower not in rule_id:
                        continue

                results.append(rule)

        return results


def format_output_table(rules: List[Dict], limit: Optional[int] = None):
    """Format rules as a table."""
    if not rules:
        print("No rules found")
        return

    # Apply limit
    display_rules = rules[:limit] if limit else rules

    # Calculate column widths
    max_id_len = max(len(r.get("rule_id", "")) for r in display_rules)
    max_title_len = min(max(len(r.get("title", "")) for r in display_rules), 60)

    # Header
    print(f"{'Rule ID':<{max_id_len}}  {'Severity':<10}  {'Type':<20}  {'Title':<{max_title_len}}")
    print("-" * (max_id_len + max_title_len + 36))

    # Rows
    for rule in display_rules:
        rule_id = rule.get("rule_id", "")
        severity = rule.get("severity", "")
        rule_type = rule.get("type", "")
        title = rule.get("title", "")

        # Truncate title if too long
        if len(title) > max_title_len:
            title = title[:max_title_len-3] + "..."

        print(f"{rule_id:<{max_id_len}}  {severity:<10}  {rule_type:<20}  {title}")

    # Summary
    if len(rules) > len(display_rules):
        print(f"\n... and {len(rules) - len(display_rules)} more")

    print(f"\nTotal: {len(rules)} rules")


def format_output_detailed(rule: Dict):
    """Format a single rule with full details."""
    print("=" * 70)
    print(f"Rule: {rule.get('rule_id')}")
    print("=" * 70)
    print()
    print(f"Title: {rule.get('title')}")
    print(f"Type: {rule.get('type')}")
    print(f"Severity: {rule.get('severity')}")
    print(f"Status: {rule.get('status')}")
    print(f"Scope: {rule.get('scope')}")
    print()

    if rule.get("impacts"):
        print(f"Impacts: {rule.get('impacts')}")

    if rule.get("tags"):
        print(f"Tags: {rule.get('tags')}")

    if rule.get("remediation_cost"):
        print(f"Remediation Cost: {rule.get('remediation_cost')}")

    if rule.get("default_quality_profiles"):
        print(f"Quality Profiles: {rule.get('default_quality_profiles')}")

    if rule.get("languages"):
        print(f"Languages: {rule.get('languages')}")

    print()
    print(f"Metadata: {rule.get('metadata_path')}")


def format_output_json(rules: List[Dict]):
    """Format rules as JSON."""
    print(json.dumps(rules, indent=2))


def format_output_markdown(rules: List[Dict]):
    """Format rules as Markdown."""
    if not rules:
        print("No rules found")
        return

    print("# SonarQube Rules")
    print()
    print(f"Total: {len(rules)} rules")
    print()

    # Group by type
    by_type = {}
    for rule in rules:
        rule_type = rule.get("type", "UNKNOWN")
        if rule_type not in by_type:
            by_type[rule_type] = []
        by_type[rule_type].append(rule)

    for rule_type in sorted(by_type.keys()):
        print(f"## {rule_type}")
        print()
        print("| Rule ID | Severity | Title |")
        print("|---------|----------|-------|")

        for rule in sorted(by_type[rule_type], key=lambda r: r.get("rule_id", "")):
            rule_id = rule.get("rule_id", "")
            severity = rule.get("severity", "")
            title = rule.get("title", "").replace("|", "\\|")
            print(f"| {rule_id} | {severity} | {title} |")

        print()


def format_output_summary(rules: List[Dict]):
    """Format rules as summary statistics."""
    if not rules:
        print("No rules found")
        return

    print(f"Total Rules: {len(rules)}")
    print()

    # By severity
    by_severity = {}
    for rule in rules:
        severity = rule.get("severity", "UNKNOWN")
        by_severity[severity] = by_severity.get(severity, 0) + 1

    print("By Severity:")
    for severity in ["BLOCKER", "CRITICAL", "MAJOR", "MINOR", "INFO", "UNKNOWN"]:
        count = by_severity.get(severity, 0)
        if count > 0:
            print(f"  {severity}: {count}")

    print()

    # By type
    by_type = {}
    for rule in rules:
        rule_type = rule.get("type", "UNKNOWN")
        by_type[rule_type] = by_type.get(rule_type, 0) + 1

    print("By Type:")
    for rule_type, count in sorted(by_type.items()):
        print(f"  {rule_type}: {count}")

    print()

    # By impact
    impacts_found = set()
    for rule in rules:
        impacts_str = rule.get("impacts", "")
        if impacts_str:
            for impact in impacts_str.split(";"):
                if ":" in impact:
                    impacts_found.add(impact)

    if impacts_found:
        print("Impacts:")
        for impact in sorted(impacts_found):
            count = sum(1 for r in rules if impact in r.get("impacts", ""))
            print(f"  {impact}: {count}")


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="Query and filter SonarQube rules",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # List all Go rules
  %(prog)s --language go

  # Find critical bugs
  %(prog)s --language python --type BUG --severity CRITICAL

  # Explain a specific rule
  %(prog)s --rule S1192

  # Find security rules with high impact
  %(prog)s --language go --impact SECURITY:HIGH

  # Search by keyword
  %(prog)s --language java --search "string literal"

  # Export as JSON
  %(prog)s --language go --format json > go_rules.json
        """
    )

    parser.add_argument(
        "--language",
        choices=SUPPORTED_LANGUAGES,
        help="Filter by language"
    )

    parser.add_argument(
        "--rule",
        help="Get details for specific rule ID (e.g., S1192)"
    )

    parser.add_argument(
        "--type",
        choices=["CODE_SMELL", "BUG", "VULNERABILITY", "SECURITY_HOTSPOT"],
        help="Filter by rule type"
    )

    parser.add_argument(
        "--severity",
        choices=["BLOCKER", "CRITICAL", "MAJOR", "MINOR", "INFO"],
        help="Filter by severity"
    )

    parser.add_argument(
        "--impact",
        help="Filter by impact (e.g., SECURITY:HIGH, MAINTAINABILITY:MEDIUM)"
    )

    parser.add_argument(
        "--tags",
        help="Filter by tags (comma-separated)"
    )

    parser.add_argument(
        "--status",
        choices=["ready", "deprecated", "closed"],
        help="Filter by status"
    )

    parser.add_argument(
        "--search",
        help="Search in title and rule ID"
    )

    parser.add_argument(
        "--format",
        choices=["table", "json", "markdown", "summary"],
        default="table",
        help="Output format (default: table)"
    )

    parser.add_argument(
        "--limit",
        type=int,
        help="Limit number of results (table format only)"
    )

    args = parser.parse_args()

    query = RuleQuery()

    # Handle specific rule lookup
    if args.rule:
        rule = query.find_rule_by_id(args.rule)
        if rule:
            if args.format == "json":
                format_output_json([rule])
            else:
                format_output_detailed(rule)
        else:
            print(f"Rule not found: {args.rule}", file=sys.stderr)
            sys.exit(1)
        return

    # Parse tags
    tags = None
    if args.tags:
        tags = [t.strip() for t in args.tags.split(",")]

    # Filter rules
    rules = query.filter_rules(
        language=args.language,
        rule_type=args.type,
        severity=args.severity,
        impact=args.impact,
        tags=tags,
        status=args.status,
        search=args.search
    )

    # Output results
    if args.format == "table":
        format_output_table(rules, limit=args.limit)
    elif args.format == "json":
        format_output_json(rules)
    elif args.format == "markdown":
        format_output_markdown(rules)
    elif args.format == "summary":
        format_output_summary(rules)


if __name__ == "__main__":
    main()
