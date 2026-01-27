# SonarQube Rules Index

Generated: 2026-01-27 06:32:59 UTC

This directory contains SonarQube rule sets extracted from the [rspec repository](https://github.com/SonarSource/rspec) for all languages supported by our clean code standards.

## Supported Languages

### C++ (`cpp`)
- **Rules**: 916
- **Files**:
  - [CSV](./cpp/rules.csv) - Machine-readable format for shell scripts, CI/CD
  - [JSON](./cpp/rules.json) - Structured data format for programmatic integration
  - [Markdown](./cpp/rules.md) - Human-readable documentation (optional)

### C# (`csharp`)
- **Rules**: 598
- **Files**:
  - [CSV](./csharp/rules.csv) - Machine-readable format for shell scripts, CI/CD
  - [JSON](./csharp/rules.json) - Structured data format for programmatic integration
  - [Markdown](./csharp/rules.md) - Human-readable documentation (optional)

### Go (`go`)
- **Rules**: 108
- **Files**:
  - [CSV](./go/rules.csv) - Machine-readable format for shell scripts, CI/CD
  - [JSON](./go/rules.json) - Structured data format for programmatic integration
  - [Markdown](./go/rules.md) - Human-readable documentation (optional)

### Python (`python`)
- **Rules**: 493
- **Files**:
  - [CSV](./python/rules.csv) - Machine-readable format for shell scripts, CI/CD
  - [JSON](./python/rules.json) - Structured data format for programmatic integration
  - [Markdown](./python/rules.md) - Human-readable documentation (optional)

### JavaScript (`javascript`)
- **Rules**: 677
- **Files**:
  - [CSV](./javascript/rules.csv) - Machine-readable format for shell scripts, CI/CD
  - [JSON](./javascript/rules.json) - Structured data format for programmatic integration
  - [Markdown](./javascript/rules.md) - Human-readable documentation (optional)

### Java (`java`)
- **Rules**: 986
- **Files**:
  - [CSV](./java/rules.csv) - Machine-readable format for shell scripts, CI/CD
  - [JSON](./java/rules.json) - Structured data format for programmatic integration
  - [Markdown](./java/rules.md) - Human-readable documentation (optional)

### Kotlin (`kotlin`)
- **Rules**: 187
- **Files**:
  - [CSV](./kotlin/rules.csv) - Machine-readable format for shell scripts, CI/CD
  - [JSON](./kotlin/rules.json) - Structured data format for programmatic integration
  - [Markdown](./kotlin/rules.md) - Human-readable documentation (optional)

### Swift (`swift`)
- **Rules**: 166
- **Files**:
  - [CSV](./swift/rules.csv) - Machine-readable format for shell scripts, CI/CD
  - [JSON](./swift/rules.json) - Structured data format for programmatic integration
  - [Markdown](./swift/rules.md) - Human-readable documentation (optional)


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

```bash
cd /path/to/ai-pack
./scripts/generate_all_language_rules.sh ~/Projects/Vibe/rspec ./quality/sonarqube-rules
```

To generate rules for a specific language:

```bash
python3 ./scripts/rspec_rule_extractor.py ~/Projects/Vibe/rspec go --format csv > go_rules.csv
```

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
