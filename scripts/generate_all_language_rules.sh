#!/usr/bin/env bash
#
# Generate SonarQube rule sets for all supported languages
#
# This script extracts rules from the SonarQube rspec repository for all
# languages defined in our clean code standards and outputs them in multiple
# formats (CSV, JSON, Markdown).
#
# Usage:
#   ./generate_all_language_rules.sh [rspec_root] [output_dir]
#
# Examples:
#   ./generate_all_language_rules.sh ~/Projects/Vibe/rspec ./quality/sonarqube-rules
#   ./generate_all_language_rules.sh  # Uses defaults
#

set -uo pipefail

# Default paths
RSPEC_ROOT="${1:-$HOME/Projects/Vibe/rspec}"
OUTPUT_DIR="${2:-$(pwd)/quality/sonarqube-rules}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EXTRACTOR="${SCRIPT_DIR}/rspec_rule_extractor.py"

# Languages we support (based on quality/clean-code/lang-*.md files)
# Format: "display_name:rspec_code"
LANGUAGES=(
	"C++:cpp"
	"C#:csharp"
	"Go:go"
	"Python:python"
	"JavaScript:javascript"
	"Java:java"
	"Kotlin:kotlin"
	"Swift:swift"
)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
	echo -e "${BLUE}ℹ${NC} $*"
}

log_success() {
	echo -e "${GREEN}✓${NC} $*"
}

log_error() {
	echo -e "${RED}✗${NC} $*" >&2
}

log_warn() {
	echo -e "${YELLOW}⚠${NC} $*" >&2
}

# Validate inputs
if [[ ! -d "$RSPEC_ROOT" ]]; then
	log_error "rspec repository not found at: $RSPEC_ROOT"
	log_info "Please provide the correct path:"
	log_info "  $0 /path/to/rspec [output_dir]"
	exit 1
fi

if [[ ! -f "$EXTRACTOR" ]]; then
	log_error "Extractor script not found at: $EXTRACTOR"
	exit 1
fi

if [[ ! -x "$EXTRACTOR" ]]; then
	log_info "Making extractor script executable..."
	chmod +x "$EXTRACTOR"
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"
log_info "Output directory: $OUTPUT_DIR"
log_info "rspec repository: $RSPEC_ROOT"
echo

# Generate rules for each language
TOTAL_LANGUAGES=${#LANGUAGES[@]}
SUCCESSFUL=0
FAILED=0

for lang_spec in "${LANGUAGES[@]}"; do
	# Parse "Display Name:code" format
	IFS=':' read -r display_name lang_code <<< "$lang_spec"

	log_info "Processing $display_name ($lang_code)..."

	# Create language-specific directory
	lang_dir="${OUTPUT_DIR}/${lang_code}"
	mkdir -p "$lang_dir"

	# Generate CSV
	csv_file="${lang_dir}/rules.csv"
	if python3 "$EXTRACTOR" "$RSPEC_ROOT" "$lang_code" --format csv --output "$csv_file" 2>/dev/null; then
		rule_count=$(tail -n +2 "$csv_file" | wc -l | tr -d ' ')
		log_success "CSV: $rule_count rules → $csv_file"
	else
		log_error "Failed to generate CSV for $display_name"
		((FAILED++))
		continue
	fi

	# Generate JSON
	json_file="${lang_dir}/rules.json"
	if python3 "$EXTRACTOR" "$RSPEC_ROOT" "$lang_code" --format json --output "$json_file" 2>/dev/null; then
		log_success "JSON: → $json_file"
	else
		log_error "Failed to generate JSON for $display_name"
		((FAILED++))
		continue
	fi

	# Generate Markdown (optional - warning only if fails)
	md_file="${lang_dir}/rules.md"
	if python3 "$EXTRACTOR" "$RSPEC_ROOT" "$lang_code" --format markdown --output "$md_file" 2>/dev/null; then
		log_success "Markdown: → $md_file"
	else
		log_warn "Failed to generate Markdown for $display_name"
	fi

	((SUCCESSFUL++))
	echo
done

# Summary
echo "======================================================================"
echo
if [[ $FAILED -eq 0 ]]; then
	log_success "Successfully generated rules for all $SUCCESSFUL languages"
else
	log_warn "Generated rules for $SUCCESSFUL/$TOTAL_LANGUAGES languages ($FAILED failed)"
fi

# Generate index file
INDEX_FILE="${OUTPUT_DIR}/README.md"
cat > "$INDEX_FILE" <<EOF
# SonarQube Rules Index

Generated: $(date -u +"%Y-%m-%d %H:%M:%S UTC")

This directory contains SonarQube rule sets extracted from the [rspec repository](https://github.com/SonarSource/rspec) for all languages supported by our clean code standards.

## Supported Languages

EOF

for lang_spec in "${LANGUAGES[@]}"; do
	IFS=':' read -r display_name lang_code <<< "$lang_spec"
	csv_file="${OUTPUT_DIR}/${lang_code}/rules.csv"

	if [[ -f "$csv_file" ]]; then
		rule_count=$(tail -n +2 "$csv_file" | wc -l | tr -d ' ')
		cat >> "$INDEX_FILE" <<EOF
### $display_name (\`$lang_code\`)
- **Rules**: $rule_count
- **Files**:
  - [CSV](./${lang_code}/rules.csv) - Machine-readable format for shell scripts, CI/CD
  - [JSON](./${lang_code}/rules.json) - Structured data format for programmatic integration
  - [Markdown](./${lang_code}/rules.md) - Human-readable documentation (optional)

EOF
	fi
done

cat >> "$INDEX_FILE" <<EOF

## File Formats

### CSV Format (Primary)
Columns: rule_id, title, type, severity, status, scope, languages, tags, impacts, remediation_cost, default_quality_profiles, language_specific_docs, metadata_path

**Use cases:**
- Shell scripts and command-line tools
- CI/CD pipeline integration (grep, awk, cut)
- Spreadsheet import for analysis
- Fast rule lookups and filtering

### JSON Format (Primary)
Array of rule objects with full metadata including impacts, tags, and language-specific documentation.

**Use cases:**
- Language-specific tooling (Go, Python, JavaScript, etc.)
- Web-based dashboards and reporting
- Complex rule analysis and cross-referencing
- API integration

### Markdown Format (Optional)
Human-readable tables grouped by rule type (CODE_SMELL, BUG, VULNERABILITY, SECURITY_HOTSPOT).

**Use cases:**
- Documentation and reference
- Quick browsing in GitHub/GitLab
- Code review guidelines

## Updating Rules

To regenerate all rule sets:

\`\`\`bash
cd /path/to/ai-pack
./scripts/generate_all_language_rules.sh ~/Projects/Vibe/rspec ./quality/sonarqube-rules
\`\`\`

To generate rules for a specific language:

\`\`\`bash
python3 ./scripts/rspec_rule_extractor.py ~/Projects/Vibe/rspec go --format csv > go_rules.csv
\`\`\`

## Integration with CI/CD

These rule sets can be integrated into your CI/CD pipeline to enforce code quality standards:

1. **Static Analysis**: Use the CSV/JSON files to configure SonarQube, SonarLint, or golangci-lint
2. **Pre-commit Hooks**: Check code against specific rule sets before commits
3. **PR Review**: Automatically comment on PRs that violate critical rules
4. **Quality Gates**: Block merges that introduce new violations

## Rule Interpretation

- **Type**: CODE_SMELL (maintainability), BUG (reliability), VULNERABILITY (security), SECURITY_HOTSPOT (security review)
- **Severity**: BLOCKER, CRITICAL, MAJOR, MINOR, INFO
- **Status**: ready, deprecated, closed
- **Impacts**: MAINTAINABILITY, RELIABILITY, SECURITY (with levels: LOW, MEDIUM, HIGH)

## References

- [SonarQube Rules](https://rules.sonarsource.com/)
- [rspec Repository](https://github.com/SonarSource/rspec)
- [Our Clean Code Standards](../clean-code/)
EOF

log_success "Generated index file: $INDEX_FILE"
echo
log_info "Done! Rule sets are available in: $OUTPUT_DIR"
