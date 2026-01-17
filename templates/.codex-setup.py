#!/usr/bin/env python3
"""
AI-Pack Codex Integration Setup Script

This script sets up Codex integration for a project using the ai-pack framework.
Run this after adding ai-pack as a git submodule.

Usage:
    python3 .codex-setup.py
"""

import shutil
import sys
from pathlib import Path


def print_header(text):
    """Print a formatted header."""
    print()
    print("=" * 80)
    print(f"  {text}")
    print("=" * 80)
    print()


def check_prerequisites():
    """Check if prerequisites are met."""
    print_header("Checking Prerequisites")

    if not Path(".ai-pack").exists():
        print("Error: .ai-pack/ directory not found")
        print()
        print("This project needs the ai-pack framework as a git submodule.")
        print()
        print("To add it:")
        print("  git submodule add <ai-pack-url> .ai-pack")
        print("  git submodule update --init")
        print()
        return False

    print("OK: .ai-pack/ framework found")

    template_file = Path(".ai-pack/templates/AGENTS.md")
    codex_templates = Path(".ai-pack/templates/.codex")
    if not template_file.exists():
        print(f"Error: {template_file} not found")
        print()
        print("Your ai-pack version may be outdated or incomplete.")
        print("Try updating: git submodule update --remote .ai-pack")
        print()
        return False

    print("OK: Codex integration template found")

    if not codex_templates.exists():
        print(f"Error: {codex_templates} not found")
        print()
        print("Your ai-pack version may be outdated or incomplete.")
        print("Try updating: git submodule update --remote .ai-pack")
        print()
        return False

    print("OK: Codex optional templates found")
    print()

    return True


def copy_agents_file():
    """Copy AGENTS.md to project root."""
    print_header("Copying Codex Integration Template")

    template_file = Path(".ai-pack/templates/AGENTS.md")
    target_file = Path("AGENTS.md")

    if target_file.exists():
        print("AGENTS.md already exists")
        response = input("Overwrite? [y/N]: ").strip().lower()
        if response != "y":
            print("Skipping AGENTS.md copy")
            print()
            return True
        target_file.unlink()

    try:
        print(f"Copying {template_file} -> {target_file}")
        shutil.copy2(template_file, target_file)
        print("OK: AGENTS.md copied")
        print()
        return True
    except Exception as e:
        print(f"Error copying AGENTS.md: {e}")
        print()
        return False


def copy_codex_templates():
    """Copy .codex templates to project root."""
    print_header("Copying Codex Optional Assets")

    template_dir = Path(".ai-pack/templates/.codex")
    target_dir = Path(".codex")

    if target_dir.exists():
        print(".codex/ already exists")
        response = input("Overwrite? [y/N]: ").strip().lower()
        if response != "y":
            print("Skipping .codex/ copy")
            print()
            return True
        shutil.rmtree(target_dir)

    try:
        print(f"Copying {template_dir} -> {target_dir}")
        shutil.copytree(template_dir, target_dir)
        print("OK: .codex/ copied")
        print()
        return True
    except Exception as e:
        print(f"Error copying .codex/: {e}")
        print()
        return False


def create_ai_directory():
    """Create .ai/ directory structure."""
    print_header("Creating .ai/ Directory Structure")

    ai_dir = Path(".ai")
    tasks_dir = ai_dir / "tasks"

    tasks_dir.mkdir(parents=True, exist_ok=True)
    print(f"OK: Created {tasks_dir}/")

    gitignore = ai_dir / ".gitignore"
    if not gitignore.exists():
        gitignore.write_text("# AI-Pack workspace\n# Task packets are tracked\n")
        print(f"OK: Created {gitignore}")

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
        print(f"OK: Created {overrides}")

    print()
    return True


def verify_setup():
    """Verify the setup is complete."""
    print_header("Verifying Setup")

    checks = [
        ("AGENTS.md", "Codex instructions"),
        (".ai/tasks/", "Task packet directory"),
        (".codex/rules/", "Codex optional rules"),
    ]

    all_good = True
    for path, description in checks:
        if Path(path).exists():
            print(f"OK: {description:24} {path}")
        else:
            print(f"Missing: {description:24} {path}")
            all_good = False

    print()
    if all_good:
        print("OK: Setup complete")
    else:
        print("Warning: Setup incomplete - some components missing")
    print()
    return all_good


def print_next_steps():
    """Print next steps for the user."""
    print_header("Next Steps")

    print("1. Customize AGENTS.md:")
    print("   # Edit AGENTS.md with project-specific context")
    print()

    print("2. Customize .ai/repo-overrides.md:")
    print("   # Add project-specific rules and standards")
    print()

    print("3. Optional: Add Codex-specific rules in .codex/rules/")
    print("   # Link them from AGENTS.md if needed")
    print()

    print("4. Commit the integration:")
    print("   git add .ai/ .codex/ AGENTS.md")
    print("   git commit -m 'Add ai-pack Codex integration'")
    print()


def main():
    """Main setup flow."""
    print()
    print("AI-Pack Codex Integration Setup")
    print()

    steps = [
        ("Checking prerequisites", check_prerequisites),
        ("Copying templates", copy_agents_file),
        ("Copying optional assets", copy_codex_templates),
        ("Creating .ai/ structure", create_ai_directory),
        ("Verifying setup", verify_setup),
    ]

    for description, func in steps:
        if not func():
            print()
            print(f"Setup failed at: {description}")
            print()
            return 1

    print_next_steps()
    print("Setup complete.")
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
        print(f"Unexpected error: {e}")
        sys.exit(1)
