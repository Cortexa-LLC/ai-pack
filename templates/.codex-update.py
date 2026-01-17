#!/usr/bin/env python3
"""
AI-Pack Codex Integration - UPDATE Script

This script updates existing projects that already have ai-pack to refresh
Codex integration assets (AGENTS.md and optional .codex/ templates).

Usage:
    python3 .ai-pack/templates/.codex-update.py
    python3 .ai-pack/templates/.codex-update.py -y  # Auto-confirm
"""

import argparse
import filecmp
import os
import shutil
import sys
from datetime import datetime
from pathlib import Path


class Colors:
    HEADER = "\033[95m"
    OKBLUE = "\033[94m"
    OKCYAN = "\033[96m"
    OKGREEN = "\033[92m"
    WARNING = "\033[93m"
    FAIL = "\033[91m"
    ENDC = "\033[0m"
    BOLD = "\033[1m"


def print_header(msg):
    print(f"\n{Colors.HEADER}{Colors.BOLD}{'='*60}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{msg}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{'='*60}{Colors.ENDC}\n")


def print_success(msg):
    print(f"{Colors.OKGREEN}OK {msg}{Colors.ENDC}")


def print_warning(msg):
    print(f"{Colors.WARNING}WARN {msg}{Colors.ENDC}")


def print_error(msg):
    print(f"{Colors.FAIL}ERR {msg}{Colors.ENDC}")


def print_info(msg):
    print(f"{Colors.OKCYAN}INFO {msg}{Colors.ENDC}")


def check_prerequisites():
    print_header("Checking Prerequisites")

    if not os.path.isdir(".ai-pack"):
        print_error("Not in a project with ai-pack submodule")
        print_info("Run from project root with .ai-pack/")
        return False
    print_success("Found .ai-pack/ submodule")

    agents_template = Path(".ai-pack/templates/AGENTS.md")
    codex_template_dir = Path(".ai-pack/templates/.codex")
    if not agents_template.exists():
        print_error("AGENTS.md template not found in .ai-pack/")
        print_info("Update submodule: git submodule update --remote .ai-pack")
        return False
    if not codex_template_dir.exists():
        print_error(".codex templates not found in .ai-pack/")
        print_info("Update submodule: git submodule update --remote .ai-pack")
        return False

    print_success("Found Codex templates in .ai-pack/")
    return True


def backup_existing():
    print_header("Creating Backup")

    if not Path("AGENTS.md").exists() and not Path(".codex").exists():
        print_info("No existing AGENTS.md or .codex/ to backup")
        return None

    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    backup_dir = Path(f".codex.backup.{timestamp}")
    backup_dir.mkdir(parents=True, exist_ok=True)

    if Path("AGENTS.md").exists():
        shutil.copy2("AGENTS.md", backup_dir / "AGENTS.md")
        print_success("Backed up AGENTS.md")

    if Path(".codex").exists():
        shutil.copytree(".codex", backup_dir / ".codex")
        print_success("Backed up .codex/")

    print_success(f"Backup created at {backup_dir}/")
    return str(backup_dir)


def detect_customizations():
    print_header("Detecting Customizations")

    customizations = {
        "custom_agents": False,
        "custom_codex_files": [],
        "modified_template_files": [],
    }

    agents_template = Path(".ai-pack/templates/AGENTS.md")
    agents_file = Path("AGENTS.md")
    if agents_file.exists():
        if not filecmp.cmp(agents_template, agents_file, shallow=False):
            customizations["custom_agents"] = True
            print_warning("AGENTS.md differs from template")
        else:
            print_success("AGENTS.md matches template")
    else:
        print_info("No AGENTS.md found")

    template_dir = Path(".ai-pack/templates/.codex")
    target_dir = Path(".codex")

    template_files = {
        path.relative_to(template_dir)
        for path in template_dir.rglob("*")
        if path.is_file()
    }

    if target_dir.exists():
        for path in target_dir.rglob("*"):
            if not path.is_file():
                continue
            rel_path = path.relative_to(target_dir)
            if rel_path not in template_files:
                customizations["custom_codex_files"].append(str(rel_path))
                print_info(f"Custom .codex file: {rel_path}")
            else:
                template_path = template_dir / rel_path
                if not filecmp.cmp(template_path, path, shallow=False):
                    customizations["modified_template_files"].append(str(rel_path))
                    print_warning(f"Modified template file: {rel_path}")
    else:
        print_info("No .codex/ directory found")

    if not any(customizations.values()):
        print_success("No customizations detected")

    return customizations


def copy_template_file(src, dst, modified_templates):
    if dst.exists() and str(dst) in modified_templates:
        new_path = dst.with_name(f"{dst.name}.new")
        shutil.copy2(src, new_path)
        print_warning(f"Template saved as {new_path}")
        return
    dst.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(src, dst)


def update_integration(customizations):
    print_header("Updating Codex Integration")

    template_agents = Path(".ai-pack/templates/AGENTS.md")
    agents_file = Path("AGENTS.md")

    if customizations["custom_agents"] and agents_file.exists():
        new_path = Path("AGENTS.md.new")
        shutil.copy2(template_agents, new_path)
        print_warning("AGENTS.md modified - saved new template as AGENTS.md.new")
    else:
        shutil.copy2(template_agents, agents_file)
        print_success("Updated AGENTS.md")

    template_dir = Path(".ai-pack/templates/.codex")
    target_dir = Path(".codex")
    target_dir.mkdir(parents=True, exist_ok=True)

    modified_templates = {f".codex/{path}" for path in customizations["modified_template_files"]}

    for template_file in template_dir.rglob("*"):
        if not template_file.is_file():
            continue
        rel_path = template_file.relative_to(template_dir)
        destination = target_dir / rel_path
        copy_template_file(template_file, destination, modified_templates)

    print_success("Updated .codex/ templates")

    if customizations["custom_codex_files"]:
        print_success(f"Preserved {len(customizations['custom_codex_files'])} custom .codex files")


def show_summary(backup_dir, customizations):
    print_header("Update Complete")

    print(f"{Colors.BOLD}Updated:{Colors.ENDC}")
    print("  - AGENTS.md template")
    print("  - .codex optional assets")

    if customizations["custom_agents"] or customizations["modified_template_files"]:
        print(f"\n{Colors.WARNING}{Colors.BOLD}Manual review recommended:{Colors.ENDC}")
        if customizations["custom_agents"]:
            print("  - Review AGENTS.md.new and merge changes")
        if customizations["modified_template_files"]:
            print("  - Review any *.new files in .codex/")

    if backup_dir:
        print(f"\n{Colors.BOLD}Backup location:{Colors.ENDC}")
        print(f"  {backup_dir}/")

    print(f"\n{Colors.BOLD}Next steps:{Colors.ENDC}")
    print("  1. Review AGENTS.md and .codex rules")
    print("  2. Commit updates:")
    print("     git add AGENTS.md .codex/")

    print(f"\n{Colors.OKGREEN}{Colors.BOLD}All done.{Colors.ENDC}\n")


def main():
    print_header("AI-Pack Codex Integration - UPDATE")
    print("This script refreshes Codex integration assets.\n")

    if not check_prerequisites():
        sys.exit(1)

    customizations = detect_customizations()

    parser = argparse.ArgumentParser(description="Update ai-pack Codex integration")
    parser.add_argument("-y", "--yes", action="store_true", help="Auto-confirm update")
    args = parser.parse_args()

    if any(customizations.values()) and not args.yes:
        response = input("\nContinue with update? [y/N]: ").strip().lower()
        if response not in ["y", "yes"]:
            print_info("Update cancelled")
            sys.exit(0)

    backup_dir = backup_existing()

    try:
        update_integration(customizations)
        show_summary(backup_dir, customizations)
    except Exception as e:
        print_error(f"Update failed: {e}")
        if backup_dir:
            print_info(f"Restore from backup: {backup_dir}/")
        sys.exit(1)


if __name__ == "__main__":
    main()
