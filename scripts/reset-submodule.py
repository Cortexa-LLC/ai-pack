#!/usr/bin/env python3
"""
AI-Pack Submodule Reset Script

Automatically resets the .ai-pack Git submodule after v2.0.0 infrastructure changes.
Effective Date: 2026-01-24

This script safely removes and re-adds the .ai-pack submodule, cleaning all Git cache
and configuration to avoid "git directory is found locally" errors.

Usage:
    # From your project root (the repo that contains .ai-pack as a submodule)
    python3 .ai-pack/scripts/reset-submodule.py

    # Or if ai-pack is elsewhere
    python3 /path/to/ai-pack/scripts/reset-submodule.py
"""

import argparse
import os
import shutil
import subprocess
import sys
from pathlib import Path


class Color:
    """ANSI color codes for terminal output"""
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'


def print_header(msg):
    """Print a header message"""
    print(f"\n{Color.HEADER}{Color.BOLD}=== {msg} ==={Color.ENDC}")


def print_success(msg):
    """Print a success message"""
    print(f"{Color.OKGREEN}✓ {msg}{Color.ENDC}")


def print_warning(msg):
    """Print a warning message"""
    print(f"{Color.WARNING}⚠ {msg}{Color.ENDC}")


def print_error(msg):
    """Print an error message"""
    print(f"{Color.FAIL}✗ {msg}{Color.ENDC}")


def print_info(msg):
    """Print an info message"""
    print(f"{Color.OKCYAN}ℹ {msg}{Color.ENDC}")


def run_command(cmd, check=True, capture_output=False, suppress_errors=False):
    """Run a shell command and handle errors"""
    try:
        if capture_output:
            result = subprocess.run(
                cmd,
                shell=True,
                check=check,
                capture_output=True,
                text=True
            )
            return result.stdout.strip() if result.returncode == 0 else ""
        else:
            subprocess.run(cmd, shell=True, check=check)
            return True
    except subprocess.CalledProcessError as e:
        if not suppress_errors:
            print_warning(f"Command failed (may be expected): {cmd}")
        return False if not capture_output else ""


def check_git_repo():
    """Check if we're in a Git repository"""
    if not Path(".git").exists():
        print_error("Not a Git repository!")
        print_info("Run this script from the root of your parent repository")
        sys.exit(1)
    print_success("Git repository detected")


def check_submodule_exists(submodule_path):
    """Check if the submodule exists"""
    if not Path(submodule_path).exists() and not Path(f".git/modules/{submodule_path}").exists():
        print_warning(f"Submodule '{submodule_path}' not found")
        print_info("This may be normal if it's already been removed")
        return False
    return True


def show_current_state(submodule_path):
    """Show current submodule state"""
    print_header("Current Submodule State")

    print_info("Submodule status:")
    run_command("git submodule status || true", suppress_errors=True)

    print_info("\nSubmodule configuration:")
    run_command("git config -f .gitmodules --get-regexp '^submodule\\.' || true", suppress_errors=True)

    print_info("\nCached submodule repositories:")
    if Path(".git/modules").exists():
        try:
            modules = list(Path(".git/modules").iterdir())
            if modules:
                for mod in modules:
                    print(f"  - {mod.name}")
            else:
                print("  (none)")
        except Exception as e:
            print_warning(f"Could not list modules: {e}")
    else:
        print("  (no .git/modules directory)")


def deinitialize_submodule(submodule_path):
    """Deinitialize the submodule"""
    print_header("Step 1: Deinitializing Submodule")
    run_command(f"git submodule deinit -f {submodule_path} || true", suppress_errors=True)
    print_success("Submodule deinitialized")


def remove_from_index(submodule_path):
    """Remove submodule from Git index and working tree"""
    print_header("Step 2: Removing from Index")
    run_command(f"git rm -f {submodule_path} || true", suppress_errors=True)
    print_success("Submodule removed from index")


def remove_cached_repo(submodule_path):
    """Remove cached submodule repository"""
    print_header("Step 3: Removing Cached Repository")
    cache_path = Path(f".git/modules/{submodule_path}")

    if cache_path.exists():
        try:
            shutil.rmtree(cache_path)
            print_success(f"Removed cached repo: {cache_path}")
        except Exception as e:
            print_error(f"Failed to remove cached repo: {e}")
            print_info("Try running: chmod -R u+w .git/modules && rm -rf .git/modules/.ai-pack")
            sys.exit(1)
    else:
        print_info("No cached repository found (already clean)")


def remove_config_entries(submodule_path):
    """Remove submodule configuration entries"""
    print_header("Step 4: Removing Configuration Entries")

    # Remove from .gitmodules
    run_command(
        f"git config -f .gitmodules --remove-section submodule.{submodule_path} 2>/dev/null || true",
        suppress_errors=True
    )

    # Remove from .git/config
    run_command(
        f"git config --remove-section submodule.{submodule_path} 2>/dev/null || true",
        suppress_errors=True
    )

    print_success("Configuration entries removed")


def remove_working_directory(submodule_path):
    """Remove submodule working directory"""
    print_header("Step 5: Removing Working Directory")
    submodule_dir = Path(submodule_path)

    if submodule_dir.exists():
        try:
            # Make sure files are writable
            for root, dirs, files in os.walk(submodule_dir):
                for d in dirs:
                    os.chmod(os.path.join(root, d), 0o755)
                for f in files:
                    os.chmod(os.path.join(root, f), 0o644)

            shutil.rmtree(submodule_dir)
            print_success(f"Removed working directory: {submodule_dir}")
        except Exception as e:
            print_error(f"Failed to remove working directory: {e}")
            print_info(f"Try running: chmod -R u+w {submodule_path} && rm -rf {submodule_path}")
            sys.exit(1)
    else:
        print_info("Working directory already removed")


def commit_removal(submodule_path):
    """Commit the submodule removal"""
    print_header("Step 6: Committing Removal")

    # Stage .gitmodules if it exists
    run_command("git add .gitmodules 2>/dev/null || true", suppress_errors=True)

    # Try to commit
    commit_msg = f"Remove {submodule_path} submodule for v2.0.0 reset"
    result = run_command(
        f'git commit -m "{commit_msg}" || true',
        suppress_errors=True
    )

    if result:
        print_success("Removal committed")
    else:
        print_info("No changes to commit (already clean)")


def readd_submodule(submodule_path, submodule_url):
    """Re-add the submodule"""
    print_header("Step 7: Re-adding Submodule")

    result = run_command(
        f"git submodule add {submodule_url} {submodule_path}",
        check=False
    )

    if not result:
        print_error("Failed to add submodule")
        print_info("This may mean the submodule already exists")
        print_info(f"Try: git submodule add --force {submodule_url} {submodule_path}")
        sys.exit(1)

    print_success("Submodule added")


def update_submodule(submodule_path):
    """Initialize and update the submodule"""
    print_header("Step 8: Initializing and Updating Submodule")

    result = run_command("git submodule update --init --recursive", check=False)

    if not result:
        print_error("Failed to initialize/update submodule")
        sys.exit(1)

    print_success("Submodule initialized and updated")


def commit_addition(submodule_path):
    """Commit the submodule addition"""
    print_header("Step 9: Committing Addition")

    commit_msg = f"Re-add {submodule_path} submodule (v2.0.0 structure)"
    result = run_command(
        f'git commit -m "{commit_msg}"',
        check=False
    )

    if result:
        print_success("Addition committed")
    else:
        print_warning("No changes to commit (may already be added)")


def verify_reset(submodule_path):
    """Verify the reset was successful"""
    print_header("Verification")

    # Check submodule status
    print_info("Submodule status:")
    status = run_command("git submodule status", capture_output=True)
    print(status if status else "  (no submodules)")

    # Check latest commits
    if Path(submodule_path).exists():
        print_info(f"\nLatest commits in {submodule_path}:")
        original_dir = os.getcwd()
        try:
            os.chdir(submodule_path)
            run_command("git log --oneline -5 || true", suppress_errors=True)
        finally:
            os.chdir(original_dir)

    print_success("\nSubmodule reset complete!")
    print_info(f"The {submodule_path} submodule has been reset with v2.0.0 structure")


def main():
    parser = argparse.ArgumentParser(
        description="Reset the .ai-pack Git submodule for v2.0.0 infrastructure",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Reset .ai-pack submodule (default)
  python3 reset-submodule.py

  # Use SSH instead of HTTPS
  python3 reset-submodule.py --ssh

  # Reset a different submodule path
  python3 reset-submodule.py --path .ai --url https://github.com/example/repo.git

For more information:
  https://github.com/Cortexa-LLC/ai-pack/blob/main/docs/SUBMODULE-RESET.md
        """
    )

    parser.add_argument(
        "--path",
        default=".ai-pack",
        help="Submodule path (default: .ai-pack)"
    )

    parser.add_argument(
        "--url",
        default="https://github.com/Cortexa-LLC/ai-pack.git",
        help="Submodule URL (default: https://github.com/Cortexa-LLC/ai-pack.git)"
    )

    parser.add_argument(
        "--ssh",
        action="store_true",
        help="Use SSH URL instead of HTTPS"
    )

    parser.add_argument(
        "--skip-verify",
        action="store_true",
        help="Skip final verification step"
    )

    parser.add_argument(
        "--yes",
        "-y",
        action="store_true",
        help="Skip confirmation prompt"
    )

    args = parser.parse_args()

    # Override URL if SSH requested
    if args.ssh:
        args.url = "git@github.com:Cortexa-LLC/ai-pack.git"

    print_header("AI-Pack Submodule Reset Script")
    print_info(f"Effective Date: 2026-01-24")
    print_info(f"Submodule Path: {args.path}")
    print_info(f"Submodule URL: {args.url}")

    # Check we're in a Git repo
    check_git_repo()

    # Show current state
    show_current_state(args.path)

    # Confirm
    if not args.yes:
        print_warning("\nThis will completely remove and re-add the submodule.")
        response = input(f"\n{Color.BOLD}Proceed with reset? [y/N]: {Color.ENDC}")
        if response.lower() not in ['y', 'yes']:
            print_info("Reset cancelled")
            sys.exit(0)

    # Execute reset
    try:
        deinitialize_submodule(args.path)
        remove_from_index(args.path)
        remove_cached_repo(args.path)
        remove_config_entries(args.path)
        remove_working_directory(args.path)
        commit_removal(args.path)
        readd_submodule(args.path, args.url)
        update_submodule(args.path)
        commit_addition(args.path)

        if not args.skip_verify:
            verify_reset(args.path)

        print_header("Success")
        print_success("Submodule reset completed successfully!")
        print_info("\nNext steps:")
        print("  1. Verify the submodule is working: cd .ai-pack && git status")
        print("  2. Push your changes: git push origin main")

    except KeyboardInterrupt:
        print_error("\n\nReset interrupted by user")
        sys.exit(1)
    except Exception as e:
        print_error(f"\n\nUnexpected error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
