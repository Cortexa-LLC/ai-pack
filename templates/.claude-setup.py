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
                rel_path = item.relative_to(Path.cwd())
                print(f"  {rel_path}")
        print()

        return True

    except Exception as e:
        print(f"❌ Error copying templates: {e}")
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


def verify_setup():
    """Verify the setup is complete."""
    print_header("Verifying Setup")

    checks = [
        (".claude/commands/ai-pack/", "Slash commands"),
        (".claude/skills/", "Auto-triggered skills"),
        (".claude/rules/", "Modular rules"),
        (".claude/hooks/", "Enforcement hooks"),
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

    print("1. Copy and customize CLAUDE.md:")
    print("   cp .ai-pack/templates/CLAUDE.md .")
    print("   # Edit CLAUDE.md with project-specific context")
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
        ("Making hooks executable", make_hooks_executable),
        ("Creating .ai/ structure", create_ai_directory),
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
