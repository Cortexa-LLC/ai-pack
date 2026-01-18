# GitHub MCP Integration for AI-Pack Repository

**Version:** 1.0.0
**Date:** 2026-01-18
**Audience:** AI-Pack Maintainers

## Overview

This document describes how the AI-Pack repository itself uses GitHub MCP for its own development, maintenance, and community management.

**Repository:** `https://github.com/Cortexa-LLC/ai-pack`

---

## Use Cases for AI-Pack Repository

### 1. Issue Triage and Management

**Scenario:** Community members report issues, request features, ask questions

**GitHub MCP Integration:**

```bash
# Daily issue triage
issues=$(mcp github search_issues \
  --repo "Cortexa-LLC/ai-pack" \
  --state "open" \
  --no-label "triaged" \
  --sort "created" \
  --json)

# Categorize each issue
for issue in $(echo $issues | jq -c '.items[]'); do
  issue_id=$(echo $issue | jq -r '.number')
  issue_body=$(echo $issue | jq -r '.body')

  # AI analysis of issue type
  category=$(analyze_issue_category "$issue_body")

  case $category in
    "bug")
      mcp github add_labels --issue $issue_id --labels "bug,needs-investigation"
      mcp github add_to_project --issue $issue_id --project "Bug Triage"
      ;;
    "feature-request")
      mcp github add_labels --issue $issue_id --labels "enhancement,needs-discussion"
      mcp github add_comment --issue $issue_id --body "Thank you for the feature request! We'll discuss this in our next planning meeting."
      ;;
    "question")
      mcp github add_labels --issue $issue_id --labels "question,documentation"
      # Could auto-respond with relevant docs
      ;;
    "duplicate")
      original=$(find_duplicate_issue "$issue_body")
      mcp github add_labels --issue $issue_id --labels "duplicate"
      mcp github add_comment --issue $issue_id --body "Duplicate of #${original}"
      mcp github close_issue --issue $issue_id
      ;;
  esac

  # Mark as triaged
  mcp github add_labels --issue $issue_id --labels "triaged"
done
```

**Benefits:**
- ✅ Automated categorization
- ✅ Consistent labeling
- ✅ Faster response times
- ✅ Reduced maintainer burden

---

### 2. Pull Request Review Automation

**Scenario:** Contributors submit PRs to AI-Pack repository

**GitHub MCP Integration:**

```bash
# Auto-review for common issues
pr_number=$1

# Get PR details
pr=$(mcp github get_pull_request --pr $pr_number --json)
files=$(mcp github get_pr_files --pr $pr_number --json)

# Check for required files
checklist=(
  "Changes include test updates"
  "Documentation updated if needed"
  "Version bumped if breaking change"
  "CHANGELOG.md updated"
)

violations=()

# Check if role files modified but tests not updated
role_files_changed=$(echo $files | jq '[.[] | select(.filename | contains("roles/"))] | length')
test_files_changed=$(echo $files | jq '[.[] | select(.filename | contains("test"))] | length')

if [ $role_files_changed -gt 0 ] && [ $test_files_changed -eq 0 ]; then
  violations+=("Role files modified but no test updates found")
fi

# Check documentation
if [ $(echo $files | jq '[.[] | select(.filename | contains(".md"))] | length') -eq 0 ]; then
  violations+=("No documentation updates - consider updating relevant .md files")
fi

# Add automated review
if [ ${#violations[@]} -gt 0 ]; then
  review_body="## Automated Review Findings

Thank you for your contribution! Please address the following before merge:

$(printf '- [ ] %s\n' "${violations[@]}")

Once addressed, request re-review. See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines."

  mcp github create_review \
    --pr $pr_number \
    --event "COMMENT" \
    --body "$review_body"
else
  mcp github create_review \
    --pr $pr_number \
    --event "APPROVE" \
    --body "✅ Automated checks pass. Awaiting maintainer review."
fi
```

**Benefits:**
- ✅ Consistent review criteria
- ✅ Faster feedback to contributors
- ✅ Maintainer time saved
- ✅ Better PR quality

---

### 3. Release Management

**Scenario:** Preparing a new AI-Pack release

**GitHub MCP Integration:**

```bash
#!/bin/bash
# scripts/prepare-release.sh

VERSION=$1  # e.g., "2.1.0"

echo "Preparing release v${VERSION}..."

# 1. Create milestone for release
milestone=$(mcp github create_milestone \
  --title "v${VERSION}" \
  --description "Release ${VERSION}" \
  --json)
milestone_id=$(echo $milestone | jq -r '.number')

# 2. Find all merged PRs since last release
last_release=$(mcp github get_latest_release --json | jq -r '.tag_name')
prs=$(mcp github search_issues \
  --query "is:pr is:merged base:main merged:>${last_release}" \
  --json)

# 3. Categorize PRs
breaking_changes=()
features=()
fixes=()
docs=()

for pr in $(echo $prs | jq -c '.items[]'); do
  pr_number=$(echo $pr | jq -r '.number')
  pr_title=$(echo $pr | jq -r '.title')
  pr_labels=$(echo $pr | jq -r '.labels[].name')

  if echo "$pr_labels" | grep -q "breaking"; then
    breaking_changes+=("$pr_title (#$pr_number)")
  elif echo "$pr_labels" | grep -q "enhancement"; then
    features+=("$pr_title (#$pr_number)")
  elif echo "$pr_labels" | grep -q "bug"; then
    fixes+=("$pr_title (#$pr_number)")
  elif echo "$pr_labels" | grep -q "documentation"; then
    docs+=("$pr_title (#$pr_number)")
  fi

  # Add to milestone
  mcp github update_issue --issue $pr_number --milestone $milestone_id
done

# 4. Generate CHANGELOG entry
changelog_entry="## [${VERSION}] - $(date +%Y-%m-%d)

### Breaking Changes
$(printf '- %s\n' "${breaking_changes[@]}")

### Features
$(printf '- %s\n' "${features[@]}")

### Bug Fixes
$(printf '- %s\n' "${fixes[@]}")

### Documentation
$(printf '- %s\n' "${docs[@]}")

[${VERSION}]: https://github.com/Cortexa-LLC/ai-pack/compare/${last_release}...v${VERSION}
"

# 5. Update CHANGELOG.md
echo "$changelog_entry" | cat - CHANGELOG.md > temp && mv temp CHANGELOG.md

# 6. Update VERSION file
echo "$VERSION" > VERSION

# 7. Create release PR
release_branch="release/v${VERSION}"
git checkout -b $release_branch
git add CHANGELOG.md VERSION
git commit -m "chore: prepare release v${VERSION}"
git push -u origin $release_branch

mcp github create_pull_request \
  --title "Release v${VERSION}" \
  --body "## Release v${VERSION}

This PR prepares the ${VERSION} release.

### Changes
${changelog_entry}

### Checklist
- [x] CHANGELOG.md updated
- [x] VERSION file updated
- [ ] All tests passing
- [ ] Documentation reviewed
- [ ] Release notes drafted

### Post-Merge
After merging, the release will be automatically created via GitHub Actions." \
  --base "main" \
  --head "$release_branch" \
  --labels "release"

echo "✅ Release PR created. Review and merge to publish v${VERSION}"
```

**Benefits:**
- ✅ Automated CHANGELOG generation
- ✅ Consistent release process
- ✅ Traceability (PRs → release)
- ✅ Reduced manual work

---

### 4. Documentation Maintenance

**Scenario:** Ensuring documentation stays up-to-date

**GitHub MCP Integration:**

```bash
# Check for stale documentation
docs_dir="."
code_files=$(find roles gates workflows -name "*.md")

for doc in $code_files; do
  # Get last modified date
  last_modified=$(git log -1 --format="%ai" -- "$doc")
  days_old=$(( ($(date +%s) - $(date -d "$last_modified" +%s)) / 86400 ))

  # Check if "Last reviewed" date is stale
  last_reviewed=$(grep "Last reviewed:" "$doc" | grep -oP '\d{4}-\d{2}-\d{2}')
  if [ -n "$last_reviewed" ]; then
    review_days_old=$(( ($(date +%s) - $(date -d "$last_reviewed" +%s)) / 86400 ))

    # Flag if not reviewed in 90 days
    if [ $review_days_old -gt 90 ]; then
      # Create issue for doc review
      mcp github create_issue \
        --title "Documentation Review: $(basename $doc)" \
        --body "**File:** \`$doc\`
**Last Reviewed:** $last_reviewed ($review_days_old days ago)
**Last Modified:** $last_modified ($days_old days ago)

This document needs review to ensure it's current.

Checklist:
- [ ] Content still accurate
- [ ] Examples still work
- [ ] Links not broken
- [ ] Update 'Last reviewed' date" \
        --labels "documentation,maintenance" \
        --assignee "@Cortexa-LLC/ai-pack-maintainers"
    fi
  fi
done
```

**Benefits:**
- ✅ Proactive doc maintenance
- ✅ No stale documentation
- ✅ Better user experience
- ✅ Issue tracking for reviews

---

### 5. Community Engagement

**Scenario:** Managing community discussions and feedback

**GitHub MCP Integration:**

```bash
# Weekly community report
start_date=$(date -d "7 days ago" +%Y-%m-%d)

# Get activity metrics
new_issues=$(mcp github search_issues \
  --query "repo:Cortexa-LLC/ai-pack is:issue created:>=$start_date" \
  --json | jq '.total_count')

closed_issues=$(mcp github search_issues \
  --query "repo:Cortexa-LLC/ai-pack is:issue closed:>=$start_date" \
  --json | jq '.total_count')

new_prs=$(mcp github search_issues \
  --query "repo:Cortexa-LLC/ai-pack is:pr created:>=$start_date" \
  --json | jq '.total_count')

merged_prs=$(mcp github search_issues \
  --query "repo:Cortexa-LLC/ai-pack is:pr merged:>=$start_date" \
  --json | jq '.total_count')

# Get top contributors
contributors=$(mcp github get_contributors \
  --since "$start_date" \
  --json | jq -r '.[] | "\(.login): \(.contributions) contributions"')

# Create discussion post
discussion_body="# Weekly Community Report ($(date +%Y-%m-%d))

## Activity This Week
- 🎫 New Issues: $new_issues
- ✅ Closed Issues: $closed_issues
- 🔀 New PRs: $new_prs
- 🎉 Merged PRs: $merged_prs

## Top Contributors
$contributors

## Highlights
[Manually add highlights here]

## Upcoming
- Next release: v2.1.0 (planned)
- Focus areas: [...]

Thank you to our amazing community! 🙏"

# Post to discussions
mcp github create_discussion \
  --title "Weekly Report: $(date +%Y-%m-%d)" \
  --body "$discussion_body" \
  --category "Announcements"
```

**Benefits:**
- ✅ Community visibility
- ✅ Contributor recognition
- ✅ Transparent development
- ✅ Engagement tracking

---

### 6. Dependency Management

**Scenario:** Managing dependencies and security updates

**GitHub MCP Integration:**

```bash
# Check Dependabot alerts
alerts=$(mcp github get_dependabot_alerts \
  --repo "Cortexa-LLC/ai-pack" \
  --state "open" \
  --json)

critical_count=$(echo $alerts | jq '[.[] | select(.security_vulnerability.severity == "critical")] | length')
high_count=$(echo $alerts | jq '[.[] | select(.security_vulnerability.severity == "high")] | length')

if [ $critical_count -gt 0 ] || [ $high_count -gt 0 ]; then
  # Create urgent issue
  mcp github create_issue \
    --title "🚨 Security: $critical_count critical, $high_count high severity alerts" \
    --body "Dependabot has detected security vulnerabilities:

**Critical:** $critical_count
**High:** $high_count

View all alerts: https://github.com/Cortexa-LLC/ai-pack/security/dependabot

Action required: Review and update dependencies." \
    --labels "security,priority-critical" \
    --assignee "@Cortexa-LLC/ai-pack-maintainers"
fi

# Auto-merge low-risk dependency updates
dependabot_prs=$(mcp github search_issues \
  --query "repo:Cortexa-LLC/ai-pack is:pr is:open author:app/dependabot" \
  --json)

for pr in $(echo $dependabot_prs | jq -c '.items[]'); do
  pr_number=$(echo $pr | jq -r '.number')
  pr_title=$(echo $pr | jq -r '.title')

  # Check if it's a minor/patch update (not major)
  if echo "$pr_title" | grep -qE "bump.*from.*[0-9]+\.[0-9]+\.[0-9]+.*to.*[0-9]+\.[0-9]+\.[0-9]+"; then
    # Check CI status
    ci_status=$(mcp github get_ci_status --pr $pr_number --json)
    if [ "$(echo $ci_status | jq -r '.state')" == "success" ]; then
      # Auto-approve and merge
      mcp github create_review --pr $pr_number --event "APPROVE" --body "✅ Automated approval for dependency update"
      mcp github merge_pull_request --pr $pr_number --merge-method "squash"
      echo "Auto-merged: $pr_title"
    fi
  fi
done
```

**Benefits:**
- ✅ Proactive security monitoring
- ✅ Automated low-risk updates
- ✅ Reduced maintenance burden
- ✅ Faster dependency updates

---

### 7. CI/CD Pipeline Management

**Scenario:** Monitoring and managing GitHub Actions workflows

**GitHub MCP Integration:**

```bash
# Monitor workflow health
workflows=$(mcp github list_workflows --json)

for workflow in $(echo $workflows | jq -c '.workflows[]'); do
  workflow_id=$(echo $workflow | jq -r '.id')
  workflow_name=$(echo $workflow | jq -r '.name')

  # Get recent runs
  runs=$(mcp github get_workflow_runs \
    --workflow $workflow_id \
    --per-page 20 \
    --json)

  # Calculate success rate
  total=$(echo $runs | jq '.workflow_runs | length')
  success=$(echo $runs | jq '[.workflow_runs[] | select(.conclusion == "success")] | length')
  success_rate=$((success * 100 / total))

  # Alert if success rate drops below 80%
  if [ $success_rate -lt 80 ]; then
    mcp github create_issue \
      --title "⚠️ CI Health: $workflow_name success rate at ${success_rate}%" \
      --body "**Workflow:** $workflow_name
**Success Rate:** $success_rate% (last 20 runs)
**Failures:** $((total - success))

Recent failures may indicate:
- Flaky tests
- Infrastructure issues
- Code quality regression

Action required: Investigate and stabilize workflow." \
      --labels "ci,infrastructure" \
      --assignee "@Cortexa-LLC/ai-pack-maintainers"
  fi
done
```

**Benefits:**
- ✅ Proactive CI monitoring
- ✅ Early detection of issues
- ✅ Improved reliability
- ✅ Better developer experience

---

### 8. Contributor Onboarding

**Scenario:** New contributors submitting first PRs

**GitHub MCP Integration:**

```bash
# Detect first-time contributors
pr_number=$1

pr=$(mcp github get_pull_request --pr $pr_number --json)
author=$(echo $pr | jq -r '.user.login')

# Check if first-time contributor
pr_count=$(mcp github search_issues \
  --query "repo:Cortexa-LLC/ai-pack is:pr author:$author" \
  --json | jq '.total_count')

if [ $pr_count -eq 1 ]; then
  # Welcome message
  welcome_message="# Welcome to AI-Pack! 🎉

Thank you @${author} for your first contribution to AI-Pack! We're excited to have you in our community.

## What Happens Next?

1. **Automated Checks:** Our CI will run tests and checks on your PR
2. **Review:** A maintainer will review your changes (usually within 48 hours)
3. **Feedback:** We may request changes or ask questions
4. **Merge:** Once approved, we'll merge your contribution!

## Resources for Contributors

- 📖 [Contributing Guide](https://github.com/Cortexa-LLC/ai-pack/blob/main/CONTRIBUTING.md)
- 💬 [Discussion Forum](https://github.com/Cortexa-LLC/ai-pack/discussions)
- 🐛 [Report Issues](https://github.com/Cortexa-LLC/ai-pack/issues/new/choose)

## Need Help?

Feel free to ask questions in the PR comments or join our [Discussions](https://github.com/Cortexa-LLC/ai-pack/discussions).

Thank you for contributing! 🙏"

  mcp github add_comment \
    --issue $pr_number \
    --body "$welcome_message"

  mcp github add_labels \
    --issue $pr_number \
    --labels "first-time-contributor"
fi
```

**Benefits:**
- ✅ Welcoming community
- ✅ Clear expectations
- ✅ Better retention
- ✅ Easier onboarding

---

## Implementation for AI-Pack Repository

### Setup Script

**File:** `scripts/setup-github-mcp.sh`

```bash
#!/bin/bash
# Setup GitHub MCP for AI-Pack repository maintenance

set -e

echo "🔧 Setting up GitHub MCP for AI-Pack repository..."

# 1. Check prerequisites
command -v docker >/dev/null 2>&1 || { echo "❌ Docker required"; exit 1; }
[ -n "$GITHUB_TOKEN" ] || { echo "❌ GITHUB_TOKEN not set"; exit 1; }

# 2. Create configuration directory
mkdir -p .github-mcp

# 3. Create environment file
cat > .github-mcp/.env << EOF
GITHUB_TOKEN=${GITHUB_TOKEN}
GITHUB_REPOSITORY=Cortexa-LLC/ai-pack
GITHUB_TOOLSETS=issues,pull_requests,workflows,repositories,security
LOG_LEVEL=info
EOF

# 4. Pull and start GitHub MCP server
docker pull ghcr.io/github/github-mcp-server:latest
docker run -d \
  --name ai-pack-github-mcp \
  --env-file .github-mcp/.env \
  --restart unless-stopped \
  -p 3000:3000 \
  ghcr.io/github/github-mcp-server:latest

echo "✅ GitHub MCP server started"

# 5. Create maintenance scripts
cat > scripts/daily-maintenance.sh << 'SCRIPT'
#!/bin/bash
# Daily maintenance tasks

echo "Running daily maintenance..."

# Issue triage
./scripts/triage-issues.sh

# Check documentation staleness
./scripts/check-docs.sh

# Monitor CI health
./scripts/monitor-ci.sh

echo "✅ Daily maintenance complete"
SCRIPT

chmod +x scripts/daily-maintenance.sh

# 6. Setup cron job (optional)
echo "To run daily maintenance automatically, add to crontab:"
echo "0 9 * * * cd $(pwd) && ./scripts/daily-maintenance.sh >> logs/maintenance.log 2>&1"

echo ""
echo "✅ Setup complete!"
echo ""
echo "Test with: docker logs ai-pack-github-mcp"
echo "Usage: See docs/GITHUB-MCP-SELF-HOSTING.md"
```

### Maintenance Automation

**File:** `.github/workflows/maintenance.yml`

```yaml
name: Repository Maintenance

on:
  schedule:
    - cron: '0 9 * * *'  # Daily at 9 AM UTC
  workflow_dispatch:  # Manual trigger

jobs:
  triage-issues:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup GitHub MCP
        run: |
          docker pull ghcr.io/github/github-mcp-server:latest
          docker run -d \
            --name github-mcp \
            -e GITHUB_TOKEN=${{ secrets.GITHUB_TOKEN }} \
            -p 3000:3000 \
            ghcr.io/github/github-mcp-server:latest

      - name: Triage New Issues
        run: ./scripts/triage-issues.sh

  check-documentation:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Need history for last modified dates

      - name: Check Stale Docs
        run: ./scripts/check-docs.sh

  monitor-ci-health:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Analyze Workflow Health
        run: ./scripts/monitor-ci.sh

  community-report:
    runs-on: ubuntu-latest
    if: github.event.schedule == '0 9 * * 1'  # Weekly on Monday
    steps:
      - uses: actions/checkout@v4

      - name: Generate Community Report
        run: ./scripts/community-report.sh
```

---

## Best Practices for AI-Pack Repository

### 1. Label Strategy

```
Type:
- bug
- enhancement
- documentation
- question

Priority:
- priority-critical
- priority-high
- priority-normal
- priority-low

Status:
- triaged
- needs-investigation
- needs-discussion
- blocked
- ready

Component:
- roles
- gates
- workflows
- templates
- quality

Meta:
- first-time-contributor
- good-first-issue
- help-wanted
- wontfix
- duplicate
```

### 2. Milestone Strategy

```
- v2.1.0 (Current Release)
- v2.2.0 (Next Release)
- v3.0.0 (Major Breaking Changes)
- Future (No timeline yet)
```

### 3. Project Boards

```
- Bug Triage (for bug management)
- Feature Roadmap (for enhancement planning)
- Community Contributions (for external PR tracking)
- Documentation (for doc improvements)
```

---

## Security Considerations

### Token Management
- Use GitHub App instead of PAT for production
- Rotate tokens regularly (90 days)
- Minimal required permissions
- Store in GitHub Secrets (Actions) or secure vault

### Audit Logging
```bash
# Log all GitHub MCP operations
export GITHUB_MCP_AUDIT_LOG=".github-mcp/audit.log"

# Every operation logs:
# [timestamp] [user] [action] [resource] [result]
```

### Rate Limiting
```bash
# Monitor rate limit usage
mcp github get_rate_limit

# Implement backoff strategy
if rate_limit_remaining < 100; then
  wait_until rate_limit_reset
fi
```

---

## Monitoring and Metrics

### Key Metrics to Track

```bash
# Issue metrics
- Average time to first response
- Average time to close
- Open issue count trend
- Issue categorization distribution

# PR metrics
- Average time to merge
- Review cycle count
- First-time contributor count
- Merge rate by contributor type

# CI metrics
- Workflow success rate
- Average build time
- Flaky test detection
- Infrastructure uptime

# Community metrics
- Active contributors (monthly)
- Discussion engagement
- Documentation views
- Star growth rate
```

### Dashboard

Create a GitHub Actions dashboard workflow:

**File:** `.github/workflows/metrics-dashboard.yml`

```yaml
name: Update Metrics Dashboard

on:
  schedule:
    - cron: '0 */6 * * *'  # Every 6 hours

jobs:
  update-dashboard:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Collect Metrics
        run: |
          # Use GitHub MCP to collect metrics
          ./scripts/collect-metrics.sh > metrics.json

      - name: Generate Dashboard
        run: |
          # Generate markdown dashboard
          ./scripts/generate-dashboard.sh < metrics.json > METRICS.md

      - name: Commit Dashboard
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add METRICS.md
          git commit -m "Update metrics dashboard [skip ci]" || true
          git push
```

---

## Conclusion

GitHub MCP integration provides powerful automation for AI-Pack repository maintenance:

**✅ Benefits:**
1. Automated issue triage and categorization
2. Consistent PR review process
3. Streamlined release management
4. Proactive documentation maintenance
5. Enhanced community engagement
6. Security and dependency monitoring
7. CI/CD health tracking
8. Better contributor experience

**🎯 Next Steps:**
1. Run `scripts/setup-github-mcp.sh`
2. Test with manual issue triage
3. Deploy maintenance workflow
4. Monitor metrics for 2 weeks
5. Refine automation based on results

This transforms AI-Pack from a manually-maintained repository into a **self-managing, community-friendly open source project**.

---

**Maintainer Contact:** @Cortexa-LLC/ai-pack-maintainers
