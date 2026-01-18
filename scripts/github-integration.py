#!/usr/bin/env python3
"""
GitHub Integration for AI-Pack with Beads
Version: 1.0.0
Date: 2026-01-18

This script provides optional GitHub integration for AI-Pack projects.
All features are configurable via .github-integration.yml

Usage:
    python3 scripts/github-integration.py <command> [options]

Commands:
    init              Initialize GitHub integration
    sync              Sync Beads tasks with GitHub issues
    import            Import GitHub issues to Beads
    export            Export Beads tasks to GitHub
    monitor           Monitor CI/CD workflows
    check-ci          Check current CI status
    create-epic       Create epic from Beads task
    create-pr         Create PR from current branch
    status            Show integration status

Prerequisites:
    - yq (YAML parser): pip install yq or brew install yq
    - jq (JSON parser): pip install jq or brew install jq
    - gh (GitHub CLI): brew install gh or https://cli.github.com/
    - bd (Beads): installed via Beads installation
"""

import sys
import os
import subprocess
import json
import shutil
import time
from pathlib import Path
from typing import Dict, List, Optional, Any
import argparse

# ANSI color codes
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'  # No Color

# Script paths
SCRIPT_DIR = Path(__file__).parent.resolve()
PROJECT_ROOT = SCRIPT_DIR.parent
CONFIG_FILE = PROJECT_ROOT / '.github-integration.yml'
CONFIG_EXAMPLE = PROJECT_ROOT / '.github-integration.yml.example'

#==============================================================================
# LOGGING
#==============================================================================

def log_info(message: str):
    """Log an informational message."""
    print(f"{Colors.BLUE}[INFO]{Colors.NC} {message}")

def log_success(message: str):
    """Log a success message."""
    print(f"{Colors.GREEN}[SUCCESS]{Colors.NC} {message}")

def log_warning(message: str):
    """Log a warning message."""
    print(f"{Colors.YELLOW}[WARNING]{Colors.NC} {message}")

def log_error(message: str):
    """Log an error message."""
    print(f"{Colors.RED}[ERROR]{Colors.NC} {message}", file=sys.stderr)

#==============================================================================
# UTILITIES
#==============================================================================

def run_command(cmd: List[str], capture_output=True, check=True) -> Optional[subprocess.CompletedProcess]:
    """Run a shell command and return the result."""
    try:
        result = subprocess.run(
            cmd,
            capture_output=capture_output,
            text=True,
            check=check
        )
        return result
    except subprocess.CalledProcessError as e:
        if check:
            log_error(f"Command failed: {' '.join(cmd)}")
            log_error(f"Error: {e.stderr}")
        return None
    except FileNotFoundError:
        log_error(f"Command not found: {cmd[0]}")
        return None

def check_command_exists(cmd: str) -> bool:
    """Check if a command exists in PATH."""
    return shutil.which(cmd) is not None

#==============================================================================
# CONFIGURATION
#==============================================================================

def is_enabled() -> bool:
    """Check if GitHub integration is enabled."""
    if not CONFIG_FILE.exists():
        return False

    result = run_command(['yq', 'eval', '.github.enabled', str(CONFIG_FILE)])
    if result and result.stdout.strip() == 'true':
        return True
    return False

def check_prerequisites() -> bool:
    """Check if all required tools are installed."""
    missing = []

    required_tools = {
        'yq': 'Install: pip install yq or brew install yq',
        'jq': 'Install: pip install jq or brew install jq',
        'gh': 'Install: brew install gh or https://cli.github.com/',
        'bd': 'Install Beads: https://github.com/steveyegge/beads'
    }

    for tool, install_msg in required_tools.items():
        if not check_command_exists(tool):
            log_error(f"{tool} not found. {install_msg}")
            missing.append(tool)

    if missing:
        log_error(f"Missing required tools: {', '.join(missing)}")
        return False

    # Check GitHub CLI authentication
    result = run_command(['gh', 'auth', 'status'], check=False)
    if result and result.returncode != 0:
        log_warning("GitHub CLI not authenticated. Run: gh auth login")

    return True

def load_config() -> Dict[str, Any]:
    """Load configuration from .github-integration.yml."""
    if not CONFIG_FILE.exists():
        log_error(f"Configuration file not found: {CONFIG_FILE}")
        log_info("Run 'init' command to create configuration file")
        sys.exit(1)

    # Read full config as JSON using yq
    result = run_command(['yq', 'eval', '-o=json', '.', str(CONFIG_FILE)])
    if not result:
        log_error("Failed to load configuration")
        sys.exit(1)

    try:
        config = json.loads(result.stdout)
        return config
    except json.JSONDecodeError as e:
        log_error(f"Invalid configuration file: {e}")
        sys.exit(1)

def get_config_value(key: str, default: Any = None) -> Any:
    """Get a specific configuration value."""
    # Use environment variable if set
    env_key = key.upper().replace('.', '_')
    env_value = os.environ.get(env_key)
    if env_value:
        return env_value

    # Otherwise get from config file
    result = run_command(['yq', 'eval', f'.{key}', str(CONFIG_FILE)])
    if result and result.stdout.strip() and result.stdout.strip() != 'null':
        return result.stdout.strip()

    return default

#==============================================================================
# INITIALIZATION
#==============================================================================

def init_github_integration():
    """Initialize GitHub integration by creating config file."""
    log_info("Initializing GitHub integration...")

    # Check if config already exists
    if CONFIG_FILE.exists():
        response = input(f"{CONFIG_FILE} already exists. Overwrite? [y/N]: ")
        if response.lower() != 'y':
            log_info("Initialization cancelled")
            return

    # Copy example config
    if not CONFIG_EXAMPLE.exists():
        log_error(f"Example configuration not found: {CONFIG_EXAMPLE}")
        sys.exit(1)

    shutil.copy(CONFIG_EXAMPLE, CONFIG_FILE)
    log_success(f"Created configuration file: {CONFIG_FILE}")

    # Try to detect repository
    result = run_command(['git', 'config', '--get', 'remote.origin.url'], check=False)
    if result and result.returncode == 0:
        repo_url = result.stdout.strip()
        # Extract owner/repo from URL
        if 'github.com' in repo_url:
            # Handle both HTTPS and SSH URLs
            repo_path = repo_url.split('github.com')[-1].strip(':/')
            repo_path = repo_path.replace('.git', '')
            log_info(f"Detected repository: {repo_path}")

            # Update config with detected repository
            run_command(['yq', 'eval', f'.github.repository = "{repo_path}"', '-i', str(CONFIG_FILE)])
            log_success(f"Set repository to: {repo_path}")

    log_info("")
    log_info("Next steps:")
    log_info("1. Edit .github-integration.yml with your settings")
    log_info("2. Set GITHUB_TOKEN environment variable:")
    log_info("   export GITHUB_TOKEN=\"ghp_your_token_here\"")
    log_info("   Or authenticate with: gh auth login")
    log_info("3. Enable features you want to use")
    log_info("4. Run: ./scripts/github-integration.py status")
    log_info("5. Start syncing: ./scripts/github-integration.py sync")

#==============================================================================
# BEADS <-> GITHUB SYNC
#==============================================================================

def sync_beads_to_github():
    """Export Beads tasks to GitHub issues."""
    log_info("Syncing Beads tasks to GitHub...")

    config = load_config()

    # Check if feature is enabled
    if not config.get('features', {}).get('issue_sync', {}).get('enabled', False):
        log_warning("Issue sync is not enabled in configuration")
        return

    # Get repository
    repository = get_config_value('github.repository')
    if not repository:
        log_error("Repository not configured")
        sys.exit(1)

    # Get Beads tasks
    result = run_command(['bd', 'list', '--json'])
    if not result:
        log_error("Failed to get Beads tasks")
        return

    try:
        tasks = json.loads(result.stdout)
    except json.JSONDecodeError:
        log_error("Failed to parse Beads tasks")
        return

    if not tasks:
        log_info("No Beads tasks to sync")
        return

    # Get sync rules
    sync_rules = config.get('sync_rules', {}).get('beads_to_github', {})
    allowed_statuses = sync_rules.get('statuses', ['open', 'in_progress', 'blocked'])
    allowed_priorities = sync_rules.get('priorities', ['critical', 'high', 'normal'])
    exclude_patterns = sync_rules.get('exclude_patterns', [])

    synced_count = 0
    skipped_count = 0

    for task in tasks:
        task_id = task.get('id', '')
        title = task.get('title', '')
        status = task.get('status', '')
        priority = task.get('priority', 'normal')

        # Check if task should be synced
        if status not in allowed_statuses:
            skipped_count += 1
            continue

        if priority not in allowed_priorities:
            skipped_count += 1
            continue

        # Check exclude patterns
        should_exclude = False
        for pattern in exclude_patterns:
            if pattern.startswith('^'):
                import re
                if re.match(pattern, title):
                    should_exclude = True
                    break
            elif pattern in title:
                should_exclude = True
                break

        if should_exclude:
            skipped_count += 1
            continue

        # Check if issue already exists for this task
        # Look for Beads ID in issue body
        search_result = run_command([
            'gh', 'issue', 'list',
            '--repo', repository,
            '--search', f'"{task_id}" in:body',
            '--json', 'number,title',
            '--limit', '1'
        ], check=False)

        if search_result and search_result.stdout.strip() != '[]':
            # Issue already exists
            skipped_count += 1
            continue

        # Create GitHub issue
        labels = ['ai-pack', 'beads-synced']
        priority_label = config.get('labels', {}).get('priority_mapping', {}).get(priority)
        if priority_label:
            labels.append(priority_label)

        # Build issue body
        body = f"""**Beads Task:** {task_id}

{task.get('description', '')}

---
Synced from Beads task system
"""

        # Create issue
        create_result = run_command([
            'gh', 'issue', 'create',
            '--repo', repository,
            '--title', title,
            '--body', body,
            '--label', ','.join(labels)
        ], check=False)

        if create_result and create_result.returncode == 0:
            issue_number = create_result.stdout.strip().split('/')[-1]
            log_success(f"Created issue #{issue_number} for task {task_id}: {title}")

            # Add comment to Beads task with GitHub issue link
            run_command([
                'bd', 'comment', task_id,
                f"GitHub Issue: #{issue_number}"
            ], check=False)

            synced_count += 1
        else:
            log_error(f"Failed to create issue for task {task_id}")

    log_success(f"Synced {synced_count} tasks to GitHub")
    if skipped_count > 0:
        log_info(f"Skipped {skipped_count} tasks (already synced or excluded)")

def import_github_to_beads():
    """Import GitHub issues to Beads tasks."""
    log_info("Importing GitHub issues to Beads...")

    config = load_config()

    # Check if feature is enabled
    if not config.get('features', {}).get('issue_sync', {}).get('enabled', False):
        log_warning("Issue sync is not enabled in configuration")
        return

    # Get repository
    repository = get_config_value('github.repository')
    if not repository:
        log_error("Repository not configured")
        sys.exit(1)

    # Get sync rules
    sync_rules = config.get('sync_rules', {}).get('github_to_beads', {})
    required_labels = sync_rules.get('required_labels', ['ai-pack'])
    exclude_labels = sync_rules.get('exclude_labels', ['wontfix', 'duplicate'])
    only_open = sync_rules.get('only_open', True)

    # Build label filter
    label_filter = ','.join(required_labels)

    # Get GitHub issues
    state_filter = 'open' if only_open else 'all'
    result = run_command([
        'gh', 'issue', 'list',
        '--repo', repository,
        '--label', label_filter,
        '--state', state_filter,
        '--json', 'number,title,body,labels,state',
        '--limit', '100'
    ])

    if not result:
        log_error("Failed to get GitHub issues")
        return

    try:
        issues = json.loads(result.stdout)
    except json.JSONDecodeError:
        log_error("Failed to parse GitHub issues")
        return

    if not issues:
        log_info("No GitHub issues to import")
        return

    imported_count = 0
    skipped_count = 0

    for issue in issues:
        issue_number = issue.get('number')
        title = issue.get('title', '')
        body = issue.get('body', '')
        labels = [l.get('name', '') for l in issue.get('labels', [])]

        # Check exclude labels
        if any(label in exclude_labels for label in labels):
            skipped_count += 1
            continue

        # Check if Beads task already exists
        # Look for issue number in Beads comments
        existing_tasks = run_command(['bd', 'list', '--json'])
        if existing_tasks:
            try:
                tasks = json.loads(existing_tasks.stdout)
                issue_exists = False
                for task in tasks:
                    task_id = task.get('id', '')
                    # Check if this task references this issue
                    comments_result = run_command(['bd', 'show', task_id], check=False)
                    if comments_result and f"#{issue_number}" in comments_result.stdout:
                        issue_exists = True
                        break

                if issue_exists:
                    skipped_count += 1
                    continue
            except json.JSONDecodeError:
                pass

        # Determine priority from labels
        priority = 'normal'
        priority_mapping = config.get('labels', {}).get('priority_mapping', {})
        for p, label in priority_mapping.items():
            if label in labels:
                priority = p
                break

        # Create Beads task
        create_result = run_command([
            'bd', 'create', title,
            '--priority', priority,
            '--json'
        ])

        if create_result:
            try:
                task_data = json.loads(create_result.stdout)
                task_id = task_data.get('id', '')

                # Add comment with GitHub issue link
                run_command([
                    'bd', 'comment', task_id,
                    f"GitHub Issue: #{issue_number}\n{body}"
                ], check=False)

                log_success(f"Created Beads task {task_id} for issue #{issue_number}: {title}")

                # Optionally create task packet
                if config.get('task_packets', {}).get('auto_create_packets', False):
                    create_task_packet_for_import(task_id, title, issue_number, body)

                imported_count += 1
            except json.JSONDecodeError:
                log_error(f"Failed to parse Beads task creation result for issue #{issue_number}")
        else:
            log_error(f"Failed to create Beads task for issue #{issue_number}")

    log_success(f"Imported {imported_count} issues to Beads")
    if skipped_count > 0:
        log_info(f"Skipped {skipped_count} issues (already imported or excluded)")

def create_task_packet_for_import(task_id: str, title: str, issue_number: int, body: str):
    """Create task packet for imported GitHub issue."""
    # Get current date
    from datetime import datetime
    date_str = datetime.now().strftime('%Y-%m-%d')

    # Create safe filename from title
    safe_title = ''.join(c if c.isalnum() or c in ('-', '_') else '_' for c in title.lower())
    safe_title = safe_title[:50]  # Limit length

    # Create task packet directory
    task_dir = PROJECT_ROOT / '.ai' / 'tasks' / f'{date_str}_{safe_title}'
    task_dir.mkdir(parents=True, exist_ok=True)

    # Create contract file
    contract_file = task_dir / '00-contract.md'
    contract_content = f"""# Task Contract

**Task ID:** {task_id}
**GitHub Issue:** #{issue_number}
**Created:** {date_str}

## Requirements

{body}

## Acceptance Criteria

- [ ] Implementation matches GitHub issue requirements
- [ ] Tests added and passing
- [ ] Code reviewed
- [ ] GitHub issue updated with completion status

## Stakeholders

- GitHub Issue Reporter
"""

    contract_file.write_text(contract_content)

    # Copy other templates
    templates_dir = PROJECT_ROOT / '.ai-pack' / 'templates' / 'task-packet'
    if templates_dir.exists():
        for template in ['10-plan.md', '20-work-log.md', '30-review.md', '40-acceptance.md']:
            src = templates_dir / template
            if src.exists():
                shutil.copy(src, task_dir / template)

    log_info(f"Created task packet: {task_dir}")

#==============================================================================
# CI/CD MONITORING
#==============================================================================

def check_ci_status():
    """Check current CI/CD workflow status."""
    log_info("Checking CI/CD status...")

    config = load_config()
    repository = get_config_value('github.repository')

    if not repository:
        log_error("Repository not configured")
        sys.exit(1)

    # Get recent workflow runs
    result = run_command([
        'gh', 'run', 'list',
        '--repo', repository,
        '--limit', '5',
        '--json', 'name,status,conclusion,createdAt,headBranch'
    ])

    if not result:
        log_error("Failed to get workflow runs")
        return

    try:
        runs = json.loads(result.stdout)
    except json.JSONDecodeError:
        log_error("Failed to parse workflow runs")
        return

    if not runs:
        log_info("No recent workflow runs found")
        return

    log_info("")
    log_info("Recent Workflows:")
    log_info("-" * 80)

    for run in runs:
        name = run.get('name', 'Unknown')
        status = run.get('status', 'unknown')
        conclusion = run.get('conclusion', 'none')
        branch = run.get('headBranch', 'unknown')

        # Color code based on status/conclusion
        if status == 'completed':
            if conclusion == 'success':
                status_str = f"{Colors.GREEN}{status} - {conclusion}{Colors.NC}"
            elif conclusion == 'failure':
                status_str = f"{Colors.RED}{status} - {conclusion}{Colors.NC}"
            else:
                status_str = f"{Colors.YELLOW}{status} - {conclusion}{Colors.NC}"
        else:
            status_str = f"{Colors.BLUE}{status}{Colors.NC}"

        print(f"  {name} ({branch}): {status_str}")

    log_info("-" * 80)

def monitor_ci():
    """Continuously monitor CI/CD workflows."""
    log_info("Starting CI/CD monitoring...")
    log_info("Press Ctrl+C to stop")

    config = load_config()
    check_interval = config.get('features', {}).get('ci_monitoring', {}).get('check_interval', 60)

    try:
        while True:
            check_ci_status()

            # Check for failures
            repository = get_config_value('github.repository')
            result = run_command([
                'gh', 'run', 'list',
                '--repo', repository,
                '--limit', '1',
                '--json', 'name,status,conclusion,databaseId'
            ], check=False)

            if result and result.returncode == 0:
                try:
                    runs = json.loads(result.stdout)
                    if runs:
                        latest_run = runs[0]
                        if latest_run.get('status') == 'completed' and latest_run.get('conclusion') == 'failure':
                            handle_ci_failure(latest_run)
                except json.JSONDecodeError:
                    pass

            time.sleep(check_interval)
    except KeyboardInterrupt:
        log_info("\nMonitoring stopped")

def handle_ci_failure(run: Dict[str, Any]):
    """Handle CI/CD workflow failure."""
    config = load_config()
    ci_config = config.get('features', {}).get('ci_monitoring', {})

    workflow_name = run.get('name', 'Unknown')
    run_id = run.get('databaseId', '')

    log_warning(f"CI failure detected: {workflow_name}")

    # Create GitHub issue if enabled
    if ci_config.get('auto_create_failure_issues', False):
        create_ci_failure_issue(workflow_name, run_id)

    # Create Beads task if enabled
    if ci_config.get('auto_create_failure_tasks', False):
        create_ci_failure_task(workflow_name, run_id)

def create_ci_failure_issue(workflow_name: str, run_id: str):
    """Create GitHub issue for CI failure."""
    config = load_config()
    repository = get_config_value('github.repository')

    title = f"CI Failure: {workflow_name}"
    body = f"""Automated CI failure report.

**Workflow:** {workflow_name}
**Run ID:** {run_id}
**Run URL:** https://github.com/{repository}/actions/runs/{run_id}

## Action Required

1. Review the workflow logs
2. Identify the root cause
3. Fix the issue
4. Verify the fix with a new run

---
Auto-generated by ai-pack GitHub integration
"""

    result = run_command([
        'gh', 'issue', 'create',
        '--repo', repository,
        '--title', title,
        '--body', body,
        '--label', 'ci-failure,priority-critical'
    ], check=False)

    if result and result.returncode == 0:
        issue_number = result.stdout.strip().split('/')[-1]
        log_success(f"Created CI failure issue: #{issue_number}")
    else:
        log_error("Failed to create CI failure issue")

def create_ci_failure_task(workflow_name: str, run_id: str):
    """Create Beads task for CI failure."""
    config = load_config()
    repository = get_config_value('github.repository')

    ci_config = config.get('ci_config', {})
    assignee = ci_config.get('on_failure', {}).get('assignee', 'Engineer-CI')
    priority = ci_config.get('on_failure', {}).get('priority', 'critical')

    title = f"Fix CI failure: {workflow_name}"

    result = run_command([
        'bd', 'create', title,
        '--priority', priority,
        '--json'
    ])

    if result:
        try:
            task_data = json.loads(result.stdout)
            task_id = task_data.get('id', '')

            # Add comment with failure details
            run_command([
                'bd', 'comment', task_id,
                f"CI Failure Details:\n"
                f"Workflow: {workflow_name}\n"
                f"Run ID: {run_id}\n"
                f"URL: https://github.com/{repository}/actions/runs/{run_id}"
            ], check=False)

            log_success(f"Created Beads task {task_id} for CI failure")
        except json.JSONDecodeError:
            log_error("Failed to parse Beads task creation result")
    else:
        log_error("Failed to create Beads task for CI failure")

#==============================================================================
# EPIC MANAGEMENT
#==============================================================================

def create_epic(task_id: str):
    """Create GitHub epic from Beads task hierarchy."""
    log_info(f"Creating epic for Beads task: {task_id}")

    config = load_config()
    repository = get_config_value('github.repository')

    if not repository:
        log_error("Repository not configured")
        sys.exit(1)

    # Get Beads task details
    result = run_command(['bd', 'show', task_id, '--json'], check=False)
    if not result:
        log_error(f"Failed to get Beads task: {task_id}")
        return

    try:
        task = json.loads(result.stdout)
    except json.JSONDecodeError:
        log_error("Failed to parse Beads task")
        return

    epic_title = task.get('title', '')
    epic_description = task.get('description', '')

    # Get dependent tasks (stories)
    deps_result = run_command(['bd', 'list', '--json'])
    if not deps_result:
        log_error("Failed to get Beads tasks")
        return

    try:
        all_tasks = json.loads(deps_result.stdout)
    except json.JSONDecodeError:
        log_error("Failed to parse Beads tasks")
        return

    # Find tasks that depend on this epic
    story_tasks = [t for t in all_tasks if task_id in t.get('depends_on', [])]

    # Create epic issue
    epic_labels = config.get('labels', {}).get('epic_label', 'epic')

    # Build epic body with checklist
    epic_body = f"""**Beads Task:** {task_id}

{epic_description}

## Stories

"""

    for story in story_tasks:
        epic_body += f"- [ ] {story.get('title', '')}\n"

    epic_body += f"""
---
Epic managed by ai-pack
"""

    # Create epic issue
    epic_result = run_command([
        'gh', 'issue', 'create',
        '--repo', repository,
        '--title', f"Epic: {epic_title}",
        '--body', epic_body,
        '--label', f'{epic_labels},ai-pack'
    ])

    if not epic_result:
        log_error("Failed to create epic issue")
        return

    epic_number = epic_result.stdout.strip().split('/')[-1]
    log_success(f"Created epic issue: #{epic_number}")

    # Create story issues
    story_label = config.get('labels', {}).get('story_label', 'story')

    for story in story_tasks:
        story_title = story.get('title', '')
        story_description = story.get('description', '')
        story_id = story.get('id', '')

        story_body = f"""**Epic:** #{epic_number}
**Beads Task:** {story_id}

{story_description}

---
Part of Epic #{epic_number}
Managed by ai-pack
"""

        story_result = run_command([
            'gh', 'issue', 'create',
            '--repo', repository,
            '--title', story_title,
            '--body', story_body,
            '--label', f'{story_label},ai-pack'
        ], check=False)

        if story_result and story_result.returncode == 0:
            story_number = story_result.stdout.strip().split('/')[-1]
            log_success(f"Created story issue: #{story_number} - {story_title}")

            # Add comment to Beads task
            run_command([
                'bd', 'comment', story_id,
                f"GitHub Issue: #{story_number} (Epic: #{epic_number})"
            ], check=False)
        else:
            log_error(f"Failed to create story issue for: {story_title}")

    # Add epic reference to Beads task
    run_command([
        'bd', 'comment', task_id,
        f"GitHub Epic: #{epic_number}"
    ], check=False)

    log_success(f"Created epic #{epic_number} with {len(story_tasks)} stories")

#==============================================================================
# STATUS
#==============================================================================

def show_status():
    """Show GitHub integration status."""
    log_info("GitHub Integration Status")
    log_info("=" * 80)

    # Check if config exists
    if not CONFIG_FILE.exists():
        log_error("Configuration file not found")
        log_info("Run 'init' command to create configuration")
        return

    # Check if enabled
    if not is_enabled():
        log_warning("GitHub integration is DISABLED")
        log_info("Edit .github-integration.yml and set github.enabled = true")
        return

    log_success("GitHub integration is ENABLED")
    log_info("")

    # Load config
    config = load_config()

    # Show repository
    repository = get_config_value('github.repository')
    if repository:
        log_info(f"Repository: {repository}")
    else:
        log_warning("Repository not configured")

    log_info("")

    # Show feature status
    log_info("Features:")
    features = config.get('features', {})

    feature_names = {
        'issue_sync': 'Issue Sync',
        'epic_management': 'Epic Management',
        'ci_monitoring': 'CI/CD Monitoring',
        'pr_management': 'PR Management'
    }

    for key, name in feature_names.items():
        enabled = features.get(key, {}).get('enabled', False)
        status = f"{Colors.GREEN}ENABLED{Colors.NC}" if enabled else f"{Colors.YELLOW}DISABLED{Colors.NC}"
        print(f"  {name}: {status}")

    log_info("")

    # Check prerequisites
    log_info("Prerequisites:")
    tools = ['yq', 'jq', 'gh', 'bd']
    for tool in tools:
        exists = check_command_exists(tool)
        status = f"{Colors.GREEN}✓{Colors.NC}" if exists else f"{Colors.RED}✗{Colors.NC}"
        print(f"  {tool}: {status}")

    log_info("")
    log_info("=" * 80)

#==============================================================================
# HELP
#==============================================================================

def show_help():
    """Show help message."""
    help_text = """
GitHub Integration for AI-Pack with Beads

Usage:
    python3 scripts/github-integration.py <command> [options]

Commands:
    init              Initialize GitHub integration (creates config file)
    sync              Bidirectional sync between Beads and GitHub
    import            Import GitHub issues to Beads tasks
    export            Export Beads tasks to GitHub issues
    monitor           Continuously monitor CI/CD workflows
    check-ci          Check current CI/CD status
    create-epic       Create GitHub epic from Beads task hierarchy
    status            Show integration status
    help              Show this help message

Examples:
    # Initialize integration
    python3 scripts/github-integration.py init

    # Check status
    python3 scripts/github-integration.py status

    # Sync tasks and issues
    python3 scripts/github-integration.py sync

    # Import GitHub issues
    python3 scripts/github-integration.py import

    # Export Beads tasks
    python3 scripts/github-integration.py export

    # Create epic from Beads task
    python3 scripts/github-integration.py create-epic bd-a1b2

    # Monitor CI/CD
    python3 scripts/github-integration.py monitor

Prerequisites:
    - yq (YAML parser): pip install yq or brew install yq
    - jq (JSON parser): pip install jq or brew install jq
    - gh (GitHub CLI): brew install gh or https://cli.github.com/
    - bd (Beads): https://github.com/steveyegge/beads

Configuration:
    Edit .github-integration.yml to configure:
    - Repository settings
    - Feature toggles
    - Sync rules
    - Labels and priorities
    - CI/CD monitoring settings

Environment Variables:
    GITHUB_TOKEN       GitHub Personal Access Token
    GITHUB_REPOSITORY  Override repository setting

Documentation:
    - Usage Guide: docs/GITHUB-INTEGRATION-USAGE.md
    - Configuration: .github-integration.yml.example
    - Scripts: scripts/README.md
"""
    print(help_text)

#==============================================================================
# MAIN
#==============================================================================

def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description='GitHub Integration for AI-Pack with Beads',
        add_help=False
    )
    parser.add_argument('command', nargs='?', default='help',
                       choices=['init', 'sync', 'import', 'export', 'monitor',
                               'check-ci', 'create-epic', 'status', 'help'])
    parser.add_argument('args', nargs='*', help='Additional arguments')

    args = parser.parse_args()

    # Handle commands
    if args.command == 'init':
        check_prerequisites()
        init_github_integration()

    elif args.command == 'sync':
        if not check_prerequisites():
            sys.exit(1)
        sync_beads_to_github()
        import_github_to_beads()

    elif args.command == 'import':
        if not check_prerequisites():
            sys.exit(1)
        import_github_to_beads()

    elif args.command == 'export':
        if not check_prerequisites():
            sys.exit(1)
        sync_beads_to_github()

    elif args.command == 'monitor':
        if not check_prerequisites():
            sys.exit(1)
        monitor_ci()

    elif args.command == 'check-ci':
        if not check_prerequisites():
            sys.exit(1)
        check_ci_status()

    elif args.command == 'create-epic':
        if not check_prerequisites():
            sys.exit(1)
        if not args.args:
            log_error("Task ID required. Usage: create-epic <task-id>")
            sys.exit(1)
        create_epic(args.args[0])

    elif args.command == 'status':
        show_status()

    elif args.command == 'help':
        show_help()

    else:
        log_error(f"Unknown command: {args.command}")
        show_help()
        sys.exit(1)

if __name__ == '__main__':
    main()
