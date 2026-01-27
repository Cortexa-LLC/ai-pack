# SonarQube Rule Suppressions Guide

**Purpose:** Document standards conflicts where project conventions differ from SonarQube rules.

---

## Philosophy

**We ONLY suppress rules when:**
1. ✅ Our project standard conflicts with SonarQube's rule
2. ✅ The project standard is based on official language conventions (e.g., gofmt, PEP 8)
3. ✅ The suppression is well-documented with justification

**We NEVER suppress:**
1. ❌ Legitimate code smells or bugs
2. ❌ Security vulnerabilities
3. ❌ Complexity issues
4. ❌ Rules just because they're inconvenient

---

## Current Suppressions

### Go: Test Function Naming (S100)

**Rule:** go:S100 - Function names should comply with naming convention
**SonarQube Expects:** camelCase function names
**Go Convention:** Test functions use underscores for readability

**Example:**
```go
// ✅ Go Convention (Idiomatic)
func TestHandleLogsStream_SSEHeaders(t *testing.T) { }
func TestHandleLogsStream_ConnectedEvent(t *testing.T) { }

// ❌ SonarQube Wants (Not idiomatic in Go)
func TestHandleLogsStreamSSEHeaders(t *testing.T) { }
func TestHandleLogsStreamConnectedEvent(t *testing.T) { }
```

**Justification:**
- Go community convention uses underscores in test names for better readability
- Separates test name from scenario: `TestFunction_Scenario`
- Widely adopted pattern in Go ecosystem (Kubernetes, Docker, etc.)
- Complies with `go test` requirements

**References:**
- [Effective Go - Names](https://go.dev/doc/effective_go#names)
- [Go Wiki - Test Names](https://github.com/golang/go/wiki/TestComments)

**Configuration:**
```properties
sonar.issue.ignore.multicriteria.go1.ruleKey=go:S100
sonar.issue.ignore.multicriteria.go1.resourceKey=**/*_test.go
```

**Applied to:** `**/*_test.go` (test files only)

---

## Adding New Suppressions

### Process

1. **Identify Conflict**
   - Verify rule violation is due to project standard, not legitimate issue
   - Confirm project standard is based on official language convention
   - Document both SonarQube's expectation and project's standard

2. **Document Justification**
   - Add entry to this document
   - Include code examples showing both approaches
   - Provide references to official language docs
   - Explain why project standard takes precedence

3. **Update Configuration**
   - Add suppression to `sonar-project.properties`
   - Use specific file patterns (limit scope)
   - Use descriptive key names (e.g., `go1`, `go2`, `csharp1`)

4. **Test Suppression**
   - Run SonarQube analysis
   - Verify rule is suppressed in target files
   - Verify rule still applies to non-target files

5. **Commit Changes**
   - Update `.sonarqube-suppressions.properties`
   - Update `sonar-project.properties`
   - Update this documentation
   - Include justification in commit message

### Template

```markdown
### Language: Rule Name (RuleID)

**Rule:** language:RuleID - Description
**SonarQube Expects:** [what SonarQube wants]
**Project Standard:** [what we do instead]

**Example:**
```language
// ✅ Project Standard
[code example]

// ❌ SonarQube Wants
[code example]
```

**Justification:**
- [Reason 1]
- [Reason 2]
- [Reference to official docs]

**References:**
- [Link to language spec/guide]

**Configuration:**
```properties
sonar.issue.ignore.multicriteria.xxx.ruleKey=language:RuleID
sonar.issue.ignore.multicriteria.xxx.resourceKey=**/*pattern*
```

**Applied to:** [file patterns]
```

---

## Quarterly Review Checklist

Suppressions should be reviewed every quarter to ensure they remain valid:

- [ ] Verify SonarQube rule hasn't changed
- [ ] Verify language convention hasn't changed
- [ ] Check if new SonarQube version addresses conflict
- [ ] Confirm suppression is still necessary
- [ ] Update documentation if needed
- [ ] Remove obsolete suppressions

**Next Review:** 2026-04-27

---

## Testing Suppressions

### Verify Suppression Works

```bash
# Run SonarQube analysis
python3 scripts/validate-with-sonarqube.py <path> --format json

# Check specific rule is suppressed
python3 scripts/validate-with-sonarqube.py <path> --format json | \
  jq '.violations[] | select(.rule_key == "go:S100")'

# Should return empty if suppression works
```

### Verify Suppression Scope

```bash
# Rule should still apply to non-test files
# Create temporary non-test file with violation
# Verify SonarQube flags it
```

---

## Suppression Categories

### ✅ Valid Suppressions

1. **Language Convention Conflicts**
   - Official language style guides (gofmt, PEP 8, etc.)
   - Community-established idioms
   - Tool-enforced standards (formatters, linters)

2. **Framework Requirements**
   - Framework-mandated patterns
   - Required by testing frameworks
   - Architectural patterns enforced by framework

### ❌ Invalid Suppressions

1. **Convenience**
   - "It's too hard to fix"
   - "We always do it this way"
   - "It works fine"

2. **Legitimate Issues**
   - Code smells
   - Security vulnerabilities
   - Complexity issues
   - Duplicate code

3. **Team Preference**
   - Personal style choices
   - "I don't like this rule"
   - Not based on official standards

---

## Examples of Rejected Suppressions

### ❌ Example 1: Cognitive Complexity

**Proposed:** Suppress S3776 (cognitive complexity) for complex business logic

**Rejected Because:**
- Complexity is a legitimate code smell
- Should refactor instead of suppress
- No conflict with language standards
- Project benefits from reduced complexity

**Correct Action:** Refactor function into smaller helpers

### ❌ Example 2: Duplicated Strings

**Proposed:** Suppress S1192 (duplicated strings) for test data

**Rejected Because:**
- Duplication is a legitimate maintainability issue
- Should extract constants
- No conflict with language standards
- Makes tests more maintainable

**Correct Action:** Extract test constants

### ❌ Example 3: Magic Numbers

**Proposed:** Suppress magic number warnings for well-known constants (e.g., 80 for HTTP port)

**Rejected Because:**
- Magic numbers reduce readability
- Should use named constants
- No conflict with language standards
- Self-documenting code is better

**Correct Action:** Define constants like `const DefaultHTTPPort = 80`

---

## FAQ

**Q: Can I suppress a rule for a single line?**
A: Only if it's a standards conflict. Use `//nolint:RuleID` with justification comment.

**Q: How do I know if it's a standards conflict?**
A: Check official language documentation. If the language/tool enforces it, it's a standards conflict. If it's just preference, it's not.

**Q: What if multiple tools disagree?**
A: Prefer official language standards first, then community consensus, then team decision (documented).

**Q: Can I suppress security rules?**
A: **Never.** Security rules should never be suppressed. Fix the vulnerability.

**Q: What about deprecated language features?**
A: Not a suppression case. Migrate to modern features or document why legacy code must remain.

---

## Maintenance

**Configuration Files:**
- `.sonarqube-suppressions.properties` - Suppression definitions (deprecated - use sonar-project.properties)
- `sonar-project.properties` - Active SonarQube configuration
- `docs/SONARQUBE-SUPPRESSIONS.md` - This documentation

**Ownership:**
- Technical Lead: Approve new suppressions
- Team: Propose suppressions with justification
- Everyone: Question inappropriate suppressions

**Process:**
1. Developer proposes suppression
2. Team reviews justification
3. Technical Lead approves/rejects
4. Update documentation and configuration
5. Quarterly review to remove obsolete suppressions

---

**Last Updated:** 2026-01-27
**Next Review:** 2026-04-27
**Owner:** Technical Lead
