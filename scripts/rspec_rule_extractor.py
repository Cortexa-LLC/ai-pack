#!/usr/bin/env python3
"""
SonarQube RSpec Rule Extractor

Extracts rules for specific programming languages from the SonarQube rspec repository.
Parses metadata.json files and generates machine-readable output (CSV, JSON, or Markdown).

Usage:
    python3 rspec_rule_extractor.py /path/to/rspec <language> [--format csv|json|markdown]

Examples:
    python3 rspec_rule_extractor.py ~/Projects/Vibe/rspec go --format csv > go_rules.csv
    python3 rspec_rule_extractor.py ~/Projects/Vibe/rspec python --format json > python_rules.json
    python3 rspec_rule_extractor.py ~/Projects/Vibe/rspec java --format markdown > java_rules.md
"""

import json
import csv
import sys
import argparse
from pathlib import Path
from typing import List, Dict, Any, Optional
from dataclasses import dataclass, asdict


@dataclass
class RuleInfo:
	"""Information about a single SonarQube rule."""
	rule_id: str
	title: str
	type: str
	severity: str
	status: str
	scope: str
	languages: List[str]
	tags: List[str]
	impacts: Dict[str, str]
	remediation_cost: str
	default_quality_profiles: List[str]
	metadata_path: str
	language_specific_docs: List[str]


class RSpecExtractor:
	"""Extract rules from SonarQube rspec repository."""

	# Language code mapping: clean-code name -> rspec directory name
	LANGUAGE_MAP = {
		'cpp': 'cfamily',
		'c++': 'cfamily',
		'c': 'cfamily',
		'csharp': 'csharp',
		'c#': 'csharp',
		'go': 'go',
		'python': 'python',
		'py': 'python',
		'javascript': 'javascript',
		'js': 'javascript',
		'typescript': 'javascript',
		'ts': 'javascript',
		'java': 'java',
		'kotlin': 'kotlin',
		'kt': 'kotlin',
		'swift': 'swift',
		'ruby': 'ruby',
		'rb': 'ruby',
		'php': 'php',
		'scala': 'scala',
		'shell': 'shell',
		'bash': 'shell',
		'sh': 'shell',
	}

	def __init__(self, rspec_root: Path):
		"""Initialize extractor with path to rspec repository."""
		self.rspec_root = rspec_root
		self.rules_dir = rspec_root / "rules"

		if not self.rules_dir.exists():
			raise ValueError(f"Rules directory not found: {self.rules_dir}")

	def normalize_language(self, language: str) -> str:
		"""Normalize language name to rspec directory name."""
		lang_lower = language.lower()
		return self.LANGUAGE_MAP.get(lang_lower, lang_lower)

	def find_rules_for_language(self, language: str) -> List[RuleInfo]:
		"""
		Find all rules that apply to the specified language.

		Args:
			language: Language code (e.g., 'go', 'python', 'java')

		Returns:
			List of RuleInfo objects for the language
		"""
		normalized_lang = self.normalize_language(language)
		rules = []

		for rule_dir in sorted(self.rules_dir.iterdir()):
			if not rule_dir.is_dir():
				continue

			meta_path = rule_dir / "metadata.json"
			if not meta_path.exists():
				continue

			try:
				rule_info = self._parse_rule(rule_dir, meta_path, normalized_lang)
				if rule_info:
					rules.append(rule_info)
			except Exception as e:
				print(f"Warning: Could not parse {meta_path}: {e}", file=sys.stderr)
				continue

		return rules

	def _parse_rule(self, rule_dir: Path, meta_path: Path, language: str) -> Optional[RuleInfo]:
		"""Parse a single rule's metadata and check if it applies to the language."""
		try:
			meta = json.loads(meta_path.read_text(encoding="utf-8"))
		except Exception as e:
			raise ValueError(f"Failed to parse JSON: {e}")

		# Check if language-specific directory exists
		lang_dir = rule_dir / language
		has_lang_specific = lang_dir.exists() and lang_dir.is_dir()

		# If no language-specific directory, this rule doesn't apply
		if not has_lang_specific:
			return None

		# Get language-specific docs
		lang_docs = self._get_language_docs(lang_dir)

		# Extract impacts
		impacts = {}
		if "code" in meta and "impacts" in meta["code"]:
			impacts = meta["code"]["impacts"]

		# Extract remediation cost
		remediation_cost = "Unknown"
		if "remediation" in meta:
			rem = meta["remediation"]
			func = rem.get("func", "")
			const_cost = rem.get("constantCost", "")
			if const_cost:
				remediation_cost = f"{func}: {const_cost}"
			else:
				remediation_cost = func

		# Determine applicable languages (all languages with subdirectories)
		languages = [d.name for d in rule_dir.iterdir() if d.is_dir() and d.name != "metadata.json"]

		return RuleInfo(
			rule_id=rule_dir.name,
			title=meta.get("title", ""),
			type=meta.get("type", ""),
			severity=meta.get("defaultSeverity", ""),
			status=meta.get("status", ""),
			scope=meta.get("scope", ""),
			languages=languages,
			tags=meta.get("tags", []),
			impacts=impacts,
			remediation_cost=remediation_cost,
			default_quality_profiles=meta.get("defaultQualityProfiles", []),
			metadata_path=str(meta_path),
			language_specific_docs=lang_docs,
		)

	def _get_language_docs(self, lang_dir: Path) -> List[str]:
		"""Get list of language-specific documentation files."""
		docs = []
		for file in lang_dir.iterdir():
			if file.suffix in ['.adoc', '.md']:
				docs.append(file.name)
		return sorted(docs)


class OutputFormatter:
	"""Format rule information for different output types."""

	@staticmethod
	def to_csv(rules: List[RuleInfo], output=sys.stdout):
		"""Output rules as CSV."""
		if not rules:
			print("No rules found", file=sys.stderr)
			return

		fieldnames = [
			"rule_id", "title", "type", "severity", "status", "scope",
			"languages", "tags", "impacts", "remediation_cost",
			"default_quality_profiles", "language_specific_docs", "metadata_path"
		]

		writer = csv.DictWriter(output, fieldnames=fieldnames)
		writer.writeheader()

		for rule in rules:
			row = asdict(rule)
			# Convert lists/dicts to strings
			row["languages"] = ";".join(row["languages"])
			row["tags"] = ";".join(row["tags"])
			row["impacts"] = ";".join(f"{k}:{v}" for k, v in row["impacts"].items())
			row["default_quality_profiles"] = ";".join(row["default_quality_profiles"])
			row["language_specific_docs"] = ";".join(row["language_specific_docs"])
			writer.writerow(row)

	@staticmethod
	def to_json(rules: List[RuleInfo], output=sys.stdout, indent: int = 2):
		"""Output rules as JSON."""
		rules_dict = [asdict(rule) for rule in rules]
		json.dump(rules_dict, output, indent=indent)
		output.write("\n")

	@staticmethod
	def to_markdown(rules: List[RuleInfo], language: str, output=sys.stdout):
		"""Output rules as Markdown table."""
		output.write(f"# SonarQube Rules for {language.title()}\n\n")
		output.write(f"Total rules: {len(rules)}\n\n")

		# Group by type
		by_type = {}
		for rule in rules:
			rule_type = rule.type or "UNKNOWN"
			if rule_type not in by_type:
				by_type[rule_type] = []
			by_type[rule_type].append(rule)

		for rule_type in sorted(by_type.keys()):
			output.write(f"## {rule_type}\n\n")
			output.write("| Rule ID | Title | Severity | Tags | Impacts |\n")
			output.write("|---------|-------|----------|------|----------|\n")

			for rule in sorted(by_type[rule_type], key=lambda r: r.rule_id):
				tags = ", ".join(rule.tags[:3])  # First 3 tags
				if len(rule.tags) > 3:
					tags += "..."
				impacts = ", ".join(f"{k}:{v}" for k, v in rule.impacts.items())

				output.write(f"| {rule.rule_id} | {rule.title} | {rule.severity} | {tags} | {impacts} |\n")

			output.write("\n")


def main():
	"""Main entry point for the script."""
	parser = argparse.ArgumentParser(
		description="Extract SonarQube rules for a specific language",
		epilog="""
Examples:
  %(prog)s ~/Projects/Vibe/rspec go --format csv > go_rules.csv
  %(prog)s ~/Projects/Vibe/rspec python --format json > python_rules.json
  %(prog)s ~/Projects/Vibe/rspec java --format markdown > java_rules.md
		""",
		formatter_class=argparse.RawDescriptionHelpFormatter
	)

	parser.add_argument(
		"rspec_root",
		type=Path,
		help="Path to the rspec repository root"
	)

	parser.add_argument(
		"language",
		type=str,
		help="Language code (e.g., go, python, java, cpp, csharp)"
	)

	parser.add_argument(
		"--format",
		choices=["csv", "json", "markdown"],
		default="csv",
		help="Output format (default: csv)"
	)

	parser.add_argument(
		"--output",
		type=argparse.FileType('w'),
		default=sys.stdout,
		help="Output file (default: stdout)"
	)

	args = parser.parse_args()

	# Validate rspec root
	if not args.rspec_root.exists():
		parser.error(f"rspec_root does not exist: {args.rspec_root}")

	try:
		extractor = RSpecExtractor(args.rspec_root)
		rules = extractor.find_rules_for_language(args.language)

		if not rules:
			print(f"No rules found for language: {args.language}", file=sys.stderr)
			print(f"Normalized to: {extractor.normalize_language(args.language)}", file=sys.stderr)
			sys.exit(1)

		print(f"Found {len(rules)} rules for {args.language}", file=sys.stderr)

		formatter = OutputFormatter()
		if args.format == "csv":
			formatter.to_csv(rules, args.output)
		elif args.format == "json":
			formatter.to_json(rules, args.output)
		elif args.format == "markdown":
			formatter.to_markdown(rules, args.language, args.output)

	except Exception as e:
		print(f"Error: {e}", file=sys.stderr)
		sys.exit(1)


if __name__ == "__main__":
	main()
