#!/usr/bin/env python3
"""
AI-Pack Claude Code Integration Setup Script

This script sets up Claude Code integration for a project using the ai-pack framework.
Run this after adding ai-pack as a git submodule.

Usage:
    python3 .claude-setup.py
"""

import os
import sys
import shutil
import stat
import json
from pathlib import Path


def print_header(text):
    """Print a formatted header."""
    print()
    print("=" * 80)
    print(f"  {text}")
    print("=" * 80)
    print()


def print_step(step_num, total_steps, text):
    """Print a formatted step."""
    print(f"[{step_num}/{total_steps}] {text}")


def check_prerequisites():
    """Check if prerequisites are met."""
    print_header("Checking Prerequisites")

    # Check if we're in the right directory
    cwd = Path.cwd()
    if cwd.name == "ai-pack" or ".ai-pack" in str(cwd):
        print("❌ Error: Running from wrong directory")
        print()
        print("This script should be run from your PROJECT root, not from the ai-pack directory.")
        print()
        print(f"Current directory: {cwd}")
        print()
        print("To fix:")
        print("  1. Navigate to your project root (where you want .ai-pack as a submodule)")
        print("  2. Run: python3 .ai-pack/templates/.claude-setup.py")
        print()
        return False

    # Check if .ai-pack exists
    if not Path(".ai-pack").exists():
        print("❌ Error: .ai-pack/ directory not found")
        print()
        print("This project needs the ai-pack framework as a git submodule.")
        print()
        print("To add it:")
        print("  git submodule add <ai-pack-url> .ai-pack")
        print("  git submodule update --init")
        print()
        return False

    print("✅ .ai-pack/ framework found")

    # Check if templates exist
    template_dir = Path(".ai-pack/templates/.claude")
    if not template_dir.exists():
        print(f"❌ Error: {template_dir} not found")
        print()
        print("Your ai-pack version may be outdated or incomplete.")
        print("Try updating: git submodule update --remote .ai-pack")
        print()
        return False

    print("✅ Claude Code integration templates found")
    print()

    return True


def copy_templates():
    """Copy .claude templates to project root."""
    print_header("Copying Claude Code Integration Templates")

    template_dir = Path(".ai-pack/templates/.claude")
    target_dir = Path(".claude")

    # Check if .claude already exists
    if target_dir.exists():
        print(f"⚠️  .claude/ directory already exists")
        response = input("Overwrite? [y/N]: ").strip().lower()
        if response != 'y':
            print("Skipping template copy")
            print()
            return True
        print("Removing existing .claude/")
        shutil.rmtree(target_dir)

    # Copy templates
    try:
        print(f"Copying {template_dir} → {target_dir}")
        shutil.copytree(template_dir, target_dir)
        print("✅ Templates copied successfully")
        print()

        # List what was copied
        print("Created:")
        for item in sorted(target_dir.rglob("*")):
            if item.is_file():
                # Use relative path from target_dir to avoid path resolution issues
                rel_path = item.relative_to(target_dir.parent)
                print(f"  {rel_path}")
        print()

        return True

    except Exception as e:
        print(f"❌ Error copying templates: {e}")
        print()
        return False


def verify_hook_paths():
    """Verify that hook paths in settings.json use the cd command pattern."""
    print_header("Verifying Hook Path Configuration")

    settings_file = Path(".claude/settings.json")

    if not settings_file.exists():
        print("⚠️  .claude/settings.json not found, skipping")
        print()
        return True

    try:
        with open(settings_file, 'r') as f:
            settings = json.load(f)

        # Check if hooks use the correct pattern: cd $(git rev-parse...) && python3
        issues_found = []
        if "hooks" in settings:
            hooks = settings["hooks"]
            for hook_type in hooks:
                if not isinstance(hooks[hook_type], list):
                    continue

                for hook_group in hooks[hook_type]:
                    if not isinstance(hook_group, dict):
                        continue

                    if "hooks" in hook_group and isinstance(hook_group["hooks"], list):
                        for hook in hook_group["hooks"]:
                            if isinstance(hook, dict) and "command" in hook:
                                command = hook["command"]
                                # Check if it's a Python hook without the cd prefix
                                if "python3 .claude/hooks/" in command and \
                                   not command.startswith("cd $(git rev-parse"):
                                    issues_found.append(f"{hook_type}: {command}")

        if issues_found:
            print("⚠️  Hook paths may not work correctly:")
            for issue in issues_found:
                print(f"  {issue}")
            print()
            print("Run: python3 .ai-pack/templates/.claude-update.py")
            print("Or see: docs/TROUBLESHOOTING.md")
            print()
            return False
        else:
            print("✅ Hook paths configured correctly")
            print()
            return True

    except Exception as e:
        print(f"❌ Error verifying hook paths: {e}")
        print()
        return False


def copy_claude_md():
    """Copy CLAUDE.md template to project root."""
    print_header("Copying CLAUDE.md Template")

    source = Path(".ai-pack/templates/CLAUDE.md")
    target = Path("CLAUDE.md")

    if not source.exists():
        print(f"⚠️  {source} not found, skipping")
        print()
        return True

    # Check if CLAUDE.md already exists
    if target.exists():
        print(f"⚠️  CLAUDE.md already exists")
        response = input("Overwrite? [y/N]: ").strip().lower()
        if response != 'y':
            print("Skipping CLAUDE.md copy")
            print()
            return True
        print("Overwriting existing CLAUDE.md")

    # Copy the template
    try:
        shutil.copy2(source, target)
        print(f"✅ Copied {source} → {target}")
        print()
        print("⚠️  IMPORTANT: Edit CLAUDE.md to customize for your project:")
        print("   - Project name and repository URL")
        print("   - Technology stack (language, framework, versions)")
        print("   - Key architectural patterns")
        print("   - Critical files and their purposes")
        print("   - Testing strategy")
        print("   - Build and deploy commands")
        print()
        return True

    except Exception as e:
        print(f"❌ Error copying CLAUDE.md: {e}")
        print()
        return False


def make_hooks_executable():
    """Make hook scripts executable."""
    print_header("Configuring Hook Scripts")

    hooks_dir = Path(".claude/hooks")

    if not hooks_dir.exists():
        print("⚠️  .claude/hooks/ not found, skipping")
        print()
        return True

    # Make all .py files executable
    made_executable = []
    for script in hooks_dir.glob("*.py"):
        try:
            # Get current permissions
            current = script.stat().st_mode
            # Add execute permission for user, group, other
            script.chmod(current | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)
            made_executable.append(script.name)
        except Exception as e:
            print(f"⚠️  Could not make {script.name} executable: {e}")

    if made_executable:
        print("✅ Made hook scripts executable:")
        for name in made_executable:
            print(f"  {name}")
    else:
        print("⚠️  No hook scripts found")

    print()
    return True


def create_ai_directory():
    """Create .ai/ directory structure."""
    print_header("Creating .ai/ Directory Structure")

    ai_dir = Path(".ai")
    tasks_dir = ai_dir / "tasks"

    # Create directories
    tasks_dir.mkdir(parents=True, exist_ok=True)
    print(f"✅ Created {tasks_dir}/")

    # Create .gitignore for .ai/
    gitignore = ai_dir / ".gitignore"
    if not gitignore.exists():
        gitignore.write_text("# AI-Pack workspace\n# Task packets are tracked\n")
        print(f"✅ Created {gitignore}")

    # Create repo-overrides.md if it doesn't exist
    overrides = ai_dir / "repo-overrides.md"
    if not overrides.exists():
        overrides.write_text("""# Project-Specific Overrides

This file contains project-specific rules that override or extend the ai-pack framework defaults.

## Language/Technology

- Language: [e.g., Python, C#, JavaScript]
- Framework: [e.g., Django, .NET, React]

## Coding Standards

[Any project-specific coding standards that differ from .ai-pack/quality/]

## Testing Requirements

[Any project-specific testing requirements]

## Build/Deploy

[Project-specific build or deployment considerations]

## Notes

[Any other project-specific guidance for AI assistants]
""")
        print(f"✅ Created {overrides}")

    print()
    return True


def setup_github_extensions():
    """Optionally set up GitHub integration role extensions."""
    print_header("GitHub Integration (Optional)")

    print("AI-Pack includes optional GitHub integration for:")
    print("  • Auto-sync Beads tasks → GitHub Issues")
    print("  • Agent-triggered actions (Orchestrator creates epic, Security creates SEC issue, etc.)")
    print("  • Epic/Story management")
    print("  • CI/CD monitoring")
    print()

    response = input("Enable GitHub integration? [y/N]: ").strip().lower()

    if response != 'y':
        print("Skipping GitHub integration setup")
        print()
        return True

    print()
    print("Setting up GitHub integration...")
    print()

    # Create .ai/roles/ directory
    roles_dir = Path(".ai/roles")
    roles_dir.mkdir(parents=True, exist_ok=True)
    print(f"✅ Created {roles_dir}/")

    # Create role extensions for key roles
    extensions = {
        "orchestrator-github-extension.md": """# Orchestrator GitHub Extension

**Base Role:** `.ai-pack/roles/orchestrator.md` (immutable, managed by ai-pack)
**Extension Type:** GitHub Integration

## Overview

This extension adds GitHub integration capabilities to the Orchestrator role,
enabling automatic synchronization between Beads tasks and GitHub Issues.

## GitHub Integration Features

### Automatic Epic Creation

When you create an epic in Beads with the Orchestrator role:

```bash
bd create "Epic: User Authentication System" --assignee Orchestrator
```

**Auto-triggers** (if enabled in configuration):
- Creates GitHub Epic Issue with label `epic`
- Creates Story Issues for all dependent tasks
- Links stories in epic checklist
- Bidirectional references (Beads ↔ GitHub)

### Work Breakdown Sync

When breaking down epics into stories:
- Story creation automatically syncs to GitHub
- Issues labeled with `story`
- Linked to parent epic
- Task packets created if configured

## Configuration

Enable in `${AI_PACK_ROOT}/.github-integration.yml`:

```yaml
features:
  agent_triggers:
    enabled: true
    orchestrator:
      epic_creation: true      # Auto-create epics
      work_breakdown: true     # Auto-sync stories
```

## Usage

### Manual Sync

```bash
# Initialize integration
${AI_PACK_ROOT}/scripts/github-integration.py init

# Create epic (manual sync)
${AI_PACK_ROOT}/scripts/github-integration.py create-epic <beads-task-id>

# Full sync
${AI_PACK_ROOT}/scripts/github-integration.py sync
```

### Automatic Sync (Recommended)

With agent triggers enabled, GitHub updates happen automatically when
you create epics or break down work. No manual commands needed.

## References

- [GitHub Integration Setup](../.ai-pack/docs/GITHUB-INTEGRATION-SETUP.md)
- [GitHub Agent Triggers](../.ai-pack/docs/GITHUB-AGENT-TRIGGERS.md)
- [Work Item Patterns](../.ai-pack/docs/WORK-ITEM-PATTERNS.md)
- [Base Orchestrator Role](../.ai-pack/roles/orchestrator.md)
""",
        "engineer-github-extension.md": """# Engineer GitHub Extension

**Base Role:** `.ai-pack/roles/engineer.md` (immutable, managed by ai-pack)
**Extension Type:** GitHub Integration

## Overview

This extension adds GitHub integration capabilities to the Engineer role,
enabling automatic status updates when working on tasks.

## GitHub Integration Features

### Task Lifecycle Sync

When you start or complete tasks:

```bash
bd start bd-story-123    # → GitHub issue labeled "in-progress"
bd complete bd-story-123 # → GitHub issue marked complete
```

**Auto-triggers** (if enabled):
- Updates issue labels
- Adds status comments
- Updates epic checklists
- Moves cards on Project boards

### Pull Request Integration

Optional automatic draft PR creation when pushing feature branches.

## Configuration

Enable in `${AI_PACK_ROOT}/.github-integration.yml`:

```yaml
features:
  agent_triggers:
    enabled: true
    engineer:
      task_start: true           # Auto-update on bd start
      task_complete: true        # Auto-update on bd complete
      auto_draft_pr: false       # Optional: auto-create draft PRs
```

## Usage

### Typical Engineer Workflow

```bash
# 1. Start task (syncs to GitHub)
bd start bd-story-456

# 2. Implement solution
# ... work on code ...

# 3. Complete task (syncs to GitHub)
bd complete bd-story-456

# GitHub issue automatically updated at each step
```

### Manual Operations

```bash
# Export specific task
${AI_PACK_ROOT}/scripts/github-integration.py export

# Create PR manually
${AI_PACK_ROOT}/scripts/github-integration.py create-pr
```

## References

- [GitHub Integration Setup](../.ai-pack/docs/GITHUB-INTEGRATION-SETUP.md)
- [GitHub Agent Triggers](../.ai-pack/docs/GITHUB-AGENT-TRIGGERS.md)
- [Base Engineer Role](../.ai-pack/roles/engineer.md)
""",
        "security-github-extension.md": """# Security GitHub Extension

**Base Role:** `.ai-pack/roles/security.md` (not yet defined in base)
**Extension Type:** GitHub Integration

## Overview

This extension defines GitHub integration for Security role, enabling
automatic creation of private security issues for investigations.

## GitHub Integration Features

### SEC Issue Creation

When Security role creates investigation tasks:

```bash
bd create "SEC: SQL injection in user search" --assignee Security
```

**Auto-triggers** (if enabled):
- Creates private GitHub issue (org repos only)
- Labels: `security`, `needs-review`
- Assigns to security team
- Tracks investigation progress

### Security Investigation Workflow

1. Security discovers vulnerability
2. Creates SEC task in Beads
3. GitHub issue auto-created (private)
4. Investigation tracked in both systems
5. Resolution synced back

## Configuration

Enable in `${AI_PACK_ROOT}/.github-integration.yml`:

```yaml
features:
  agent_triggers:
    enabled: true
    security:
      sec_issue_creation: true    # Auto-create SEC issues
      sec_labels:
        - "security"
        - "needs-review"
        - "vulnerability"
      sec_private: true             # Private visibility (org only)
      sec_assignees:
        - "security-team"
```

## Usage

### Creating Security Issue

```bash
# In Beads
bd create "SEC: Investigation - XSS in comment form" \\
  --assignee Security \\
  --priority critical

# Automatically creates private GitHub issue
```

### Manual Operations

```bash
# Create security issue manually
${AI_PACK_ROOT}/scripts/github-integration.py create-security-issue <task-id>
```

## Security Considerations

- **Private issues**: Only available in organization repositories
- **Access control**: Limit who can see security issues
- **Sync patterns**: Optionally exclude SEC tasks from public sync
- **Audit trail**: Maintain investigation history

## References

- [GitHub Integration Setup](../.ai-pack/docs/GITHUB-INTEGRATION-SETUP.md)
- [GitHub Agent Triggers](../.ai-pack/docs/GITHUB-AGENT-TRIGGERS.md)
""",
    }

    # Write role extensions
    created_files = []
    for filename, content in extensions.items():
        filepath = roles_dir / filename
        filepath.write_text(content)
        created_files.append(f".ai/roles/{filename}")
        print(f"✅ Created {filepath}")

    print()

    # Update repo-overrides.md with GitHub extensions
    overrides = Path(".ai/repo-overrides.md")
    if overrides.exists():
        with open(overrides, 'a') as f:
            f.write("""

## GitHub Integration

This project uses GitHub integration for automated Beads ↔ GitHub synchronization.

### Role Extensions

This project extends the following roles with GitHub integration:
- **Orchestrator**: See [.ai/roles/orchestrator-github-extension.md](.ai/roles/orchestrator-github-extension.md)
- **Engineer**: See [.ai/roles/engineer-github-extension.md](.ai/roles/engineer-github-extension.md)
- **Security**: See [.ai/roles/security-github-extension.md](.ai/roles/security-github-extension.md)

### Setup Required

To complete GitHub integration setup:

1. **Initialize integration:**
   ```bash
   ${AI_PACK_ROOT}/scripts/github-integration.py init
   ```

2. **Configure** `${AI_PACK_ROOT}/.github-integration.yml`:
   - Set repository: `your-org/your-repo`
   - Enable agent triggers
   - Configure role-specific options

3. **Authenticate:**
   ```bash
   gh auth login
   # OR
   export GITHUB_TOKEN="ghp_your_token_here"
   ```

4. **Test:**
   ```bash
   ${AI_PACK_ROOT}/scripts/github-integration.py status
   ```

### Documentation

- [Setup Guide](../.ai-pack/docs/GITHUB-INTEGRATION-SETUP.md)
- [Agent Triggers](../.ai-pack/docs/GITHUB-AGENT-TRIGGERS.md)
- [Work Item Patterns](../.ai-pack/docs/WORK-ITEM-PATTERNS.md)
""")
        print(f"✅ Updated {overrides} with GitHub integration section")

    print()
    print("✅ GitHub integration role extensions created")
    print()
    print("Next steps:")
    print("  1. Run: ${AI_PACK_ROOT}/scripts/github-integration.py init")
    print("  2. Configure: ${AI_PACK_ROOT}/.github-integration.yml")
    print("  3. Authenticate: gh auth login")
    print()

    return True


def verify_setup():
    """Verify the setup is complete."""
    print_header("Verifying Setup")

    checks = [
        (".claude/commands/ai-pack/", "Slash commands"),
        (".claude/skills/", "Auto-triggered skills"),
        (".claude/rules/", "Modular rules"),
        (".claude/hooks/", "Enforcement hooks"),
        (".claude/agents/", "Agent configurations"),
        (".claude/settings.json", "Hook configuration"),
        (".ai/tasks/", "Task packet directory"),
    ]

    all_good = True
    for path, description in checks:
        if Path(path).exists():
            print(f"✅ {description:30} {path}")
        else:
            print(f"❌ {description:30} {path} (MISSING)")
            all_good = False

    print()

    if all_good:
        print("✅ Setup complete!")
    else:
        print("⚠️  Setup incomplete - some components missing")

    print()
    return all_good


def print_next_steps():
    """Print next steps for the user."""
    print_header("Next Steps")

    print("1. Customize CLAUDE.md:")
    print("   vim CLAUDE.md")
    print("   # Edit with project-specific details:")
    print("   #   - Project name and repo URL")
    print("   #   - Technology stack")
    print("   #   - Key files and patterns")
    print("   #   - Build/test commands")
    print()

    print("2. Customize .ai/repo-overrides.md:")
    print("   # Add project-specific rules and standards")
    print()

    print("3. Commit the integration:")
    print("   git add .claude/ .ai/ CLAUDE.md")
    print("   git commit -m 'Add ai-pack Claude Code integration'")
    print()

    print("4. Start using ai-pack commands:")
    print("   /ai-pack help              # See all commands")
    print("   /ai-pack task-init <name>  # Start a new task")
    print()

    print("5. Optional: Test the setup:")
    print("   python3 .claude/hooks/task-status.py")
    print()


def main():
    """Main setup flow."""
    print()
    print("╔════════════════════════════════════════════════════════════════════════════╗")
    print("║                                                                            ║")
    print("║                   AI-Pack Claude Code Integration Setup                   ║")
    print("║                                                                            ║")
    print("╚════════════════════════════════════════════════════════════════════════════╝")

    # Run setup steps
    steps = [
        ("Checking prerequisites", check_prerequisites),
        ("Copying templates", copy_templates),
        ("Copying CLAUDE.md", copy_claude_md),
        ("Verifying hook paths", verify_hook_paths),
        ("Making hooks executable", make_hooks_executable),
        ("Creating .ai/ structure", create_ai_directory),
        ("GitHub integration setup", setup_github_extensions),
        ("Verifying setup", verify_setup),
    ]

    for step_num, (description, func) in enumerate(steps, 1):
        if not func():
            print()
            print(f"❌ Setup failed at: {description}")
            print()
            return 1

    # Print next steps
    print_next_steps()

    print("Setup complete! 🎉")
    print()

    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except KeyboardInterrupt:
        print()
        print("Setup cancelled by user")
        sys.exit(1)
    except Exception as e:
        print()
        print(f"❌ Unexpected error: {e}")
        sys.exit(1)
