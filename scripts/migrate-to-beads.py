#!/usr/bin/env python3
"""
AI-Pack Beads Migration Script

Migrates existing ai-pack projects from v1.0.0 (TodoWrite) to v1.1.0 (Beads).

Usage:
    python migrate-to-beads.py [--dry-run] [--yes]

Options:
    --dry-run   Show what would be done without making changes
    --yes       Skip confirmation prompts (auto-approve)
"""

import os
import sys
import subprocess
import argparse
import shutil
from pathlib import Path
from typing import Tuple, Optional


class Colors:
    """ANSI color codes for terminal output"""
    RESET = "\033[0m"
    BOLD = "\033[1m"
    RED = "\033[91m"
    GREEN = "\033[92m"
    YELLOW = "\033[93m"
    BLUE = "\033[94m"
    MAGENTA = "\033[95m"
    CYAN = "\033[96m"


class MigrationTool:
    """Handles migration from ai-pack v1.0.0 to v1.1.0 with Beads integration"""

    def __init__(self, dry_run: bool = False, auto_yes: bool = False):
        self.dry_run = dry_run
        self.auto_yes = auto_yes
        self.project_root = Path.cwd()
        self.ai_pack_dir = self.project_root / ".ai-pack"
        self.beads_dir = self.project_root / ".beads"
        self.gitignore_path = self.project_root / ".gitignore"

    def print_header(self, text: str):
        """Print section header"""
        print(f"\n{Colors.BOLD}{Colors.CYAN}{'=' * 70}{Colors.RESET}")
        print(f"{Colors.BOLD}{Colors.CYAN}{text:^70}{Colors.RESET}")
        print(f"{Colors.BOLD}{Colors.CYAN}{'=' * 70}{Colors.RESET}\n")

    def print_success(self, text: str):
        """Print success message"""
        print(f"{Colors.GREEN}✓{Colors.RESET} {text}")

    def print_error(self, text: str):
        """Print error message"""
        print(f"{Colors.RED}✗{Colors.RESET} {text}")

    def print_warning(self, text: str):
        """Print warning message"""
        print(f"{Colors.YELLOW}⚠{Colors.RESET}  {text}")

    def print_info(self, text: str):
        """Print info message"""
        print(f"{Colors.BLUE}ℹ{Colors.RESET}  {text}")

    def run_command(self, cmd: list, cwd: Optional[Path] = None, check: bool = True) -> Tuple[int, str, str]:
        """Run shell command and return (returncode, stdout, stderr)"""
        if self.dry_run:
            self.print_info(f"[DRY RUN] Would run: {' '.join(cmd)}")
            return (0, "", "")

        try:
            result = subprocess.run(
                cmd,
                cwd=cwd or self.project_root,
                capture_output=True,
                text=True,
                check=check
            )
            return (result.returncode, result.stdout, result.stderr)
        except subprocess.CalledProcessError as e:
            return (e.returncode, e.stdout, e.stderr)

    def ask_confirmation(self, question: str) -> bool:
        """Ask user for yes/no confirmation"""
        if self.auto_yes:
            return True

        while True:
            response = input(f"{question} [y/N]: ").lower().strip()
            if response in ('y', 'yes'):
                return True
            elif response in ('n', 'no', ''):
                return False
            else:
                print("Please answer 'y' or 'n'")

    def check_prerequisites(self) -> bool:
        """Check if prerequisites are met"""
        self.print_header("Checking Prerequisites")

        all_good = True

        # Check if in project root with .ai-pack/
        if not self.ai_pack_dir.exists():
            self.print_error(f"Not in an ai-pack project root. Expected .ai-pack/ at: {self.ai_pack_dir}")
            self.print_info("Run this script from your project root directory (where .ai-pack/ exists)")
            all_good = False
        else:
            self.print_success(f"Found .ai-pack/ directory")

        # Check if git repository
        returncode, _, _ = self.run_command(["git", "rev-parse", "--is-inside-work-tree"], check=False)
        if returncode != 0:
            self.print_error("Not in a git repository. Beads requires git.")
            self.print_info("Initialize git first: git init")
            all_good = False
        else:
            self.print_success("Git repository detected")

        # Check Python version
        python_version = f"{sys.version_info.major}.{sys.version_info.minor}"
        if sys.version_info < (3, 7):
            self.print_error(f"Python 3.7+ required (found {python_version})")
            all_good = False
        else:
            self.print_success(f"Python {python_version}")

        # Check if Beads is installed
        returncode, stdout, _ = self.run_command(["bd", "--version"], check=False)
        if returncode != 0:
            self.print_warning("Beads not installed (will be installed during migration)")
        else:
            version = stdout.strip()
            self.print_success(f"Beads already installed: {version}")

        # Check if already migrated
        if self.beads_dir.exists() and (self.beads_dir / "issues.jsonl").exists():
            self.print_warning("Beads already initialized (.beads/issues.jsonl exists)")
            self.print_info("Migration may have already been completed")

        return all_good

    def install_beads(self) -> bool:
        """Install Beads CLI tool"""
        self.print_header("Installing Beads")

        # Check if already installed
        returncode, _, _ = self.run_command(["bd", "--version"], check=False)
        if returncode == 0:
            self.print_success("Beads already installed, skipping installation")
            return True

        system = sys.platform
        self.print_info(f"Detected platform: {system}")

        if system == "darwin":  # macOS
            self.print_info("Installing via Homebrew...")
            returncode, stdout, stderr = self.run_command(
                ["brew", "install", "steveyegge/beads/beads"],
                check=False
            )
            if returncode != 0:
                self.print_error("Homebrew installation failed")
                self.print_info("Try manual installation: curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/install.sh | bash")
                return False

        elif system.startswith("linux") or system == "freebsd":
            self.print_info("Installing via curl...")
            returncode, _, _ = self.run_command(
                ["bash", "-c", "curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/install.sh | bash"],
                check=False
            )
            if returncode != 0:
                self.print_error("Installation failed")
                return False

        elif system == "win32":
            self.print_error("Windows installation requires PowerShell")
            self.print_info("Run in PowerShell: irm https://raw.githubusercontent.com/steveyegge/beads/main/install.ps1 | iex")
            return False

        else:
            self.print_error(f"Unsupported platform: {system}")
            self.print_info("See https://github.com/steveyegge/beads for manual installation")
            return False

        # Verify installation
        returncode, stdout, _ = self.run_command(["bd", "--version"], check=False)
        if returncode == 0:
            self.print_success(f"Beads installed successfully: {stdout.strip()}")
            return True
        else:
            self.print_error("Beads installation verification failed")
            return False

    def update_ai_pack_submodule(self) -> bool:
        """Update .ai-pack/ submodule to latest"""
        self.print_header("Updating AI-Pack Submodule")

        if not self.ask_confirmation("Update .ai-pack/ submodule to v1.1.0+?"):
            self.print_info("Skipping submodule update")
            return True

        # Check current commit
        returncode, stdout, _ = self.run_command(
            ["git", "rev-parse", "HEAD"],
            cwd=self.ai_pack_dir,
            check=False
        )
        if returncode == 0:
            current_commit = stdout.strip()[:7]
            self.print_info(f"Current ai-pack commit: {current_commit}")

        # Fetch and checkout main
        self.print_info("Fetching latest from origin...")
        returncode, _, stderr = self.run_command(
            ["git", "fetch", "origin"],
            cwd=self.ai_pack_dir,
            check=False
        )
        if returncode != 0:
            self.print_error(f"Failed to fetch: {stderr}")
            return False

        self.print_info("Checking out main branch...")
        returncode, _, stderr = self.run_command(
            ["git", "checkout", "main"],
            cwd=self.ai_pack_dir,
            check=False
        )
        if returncode != 0:
            self.print_error(f"Failed to checkout main: {stderr}")
            return False

        self.print_info("Pulling latest changes...")
        returncode, _, stderr = self.run_command(
            ["git", "pull", "origin", "main"],
            cwd=self.ai_pack_dir,
            check=False
        )
        if returncode != 0:
            self.print_error(f"Failed to pull: {stderr}")
            return False

        # Get new commit
        returncode, stdout, _ = self.run_command(
            ["git", "rev-parse", "HEAD"],
            cwd=self.ai_pack_dir,
            check=False
        )
        if returncode == 0:
            new_commit = stdout.strip()[:7]
            self.print_success(f"Updated ai-pack to commit: {new_commit}")

        # Stage submodule change
        returncode, _, _ = self.run_command(["git", "add", ".ai-pack"], check=False)
        if returncode == 0:
            self.print_success("Staged .ai-pack/ submodule update")

        return True

    def initialize_beads(self) -> bool:
        """Initialize Beads in project"""
        self.print_header("Initializing Beads")

        # Check if already initialized
        if self.beads_dir.exists() and (self.beads_dir / "issues.jsonl").exists():
            self.print_warning("Beads already initialized")
            if not self.ask_confirmation("Reinitialize Beads (will preserve existing tasks)?"):
                self.print_info("Skipping Beads initialization")
                return True

        self.print_info("Running: bd init")
        returncode, stdout, stderr = self.run_command(["bd", "init"], check=False)

        if returncode != 0:
            self.print_error(f"Failed to initialize Beads: {stderr}")
            return False

        if not self.dry_run and not self.beads_dir.exists():
            self.print_error(".beads/ directory not created")
            return False

        self.print_success("Beads initialized successfully")

        # Verify files
        issues_jsonl = self.beads_dir / "issues.jsonl"
        if not self.dry_run and issues_jsonl.exists():
            self.print_success(f"Created: {issues_jsonl}")
        elif not self.dry_run:
            self.print_warning(f"Expected file not found: {issues_jsonl}")

        return True

    def update_gitignore(self) -> bool:
        """Update .gitignore to exclude .beads/*.db"""
        self.print_header("Updating .gitignore")

        beads_db_pattern = ".beads/*.db"

        # Check if .gitignore exists
        if not self.gitignore_path.exists():
            self.print_warning(".gitignore not found, creating new file")
            if self.dry_run:
                self.print_info(f"[DRY RUN] Would create .gitignore with: {beads_db_pattern}")
                return True

            with open(self.gitignore_path, 'w') as f:
                f.write(f"# Beads SQLite cache (JSONL is committed)\n{beads_db_pattern}\n")
            self.print_success("Created .gitignore with Beads exclusion")
            return True

        # Check if pattern already exists
        with open(self.gitignore_path, 'r') as f:
            content = f.read()

        if beads_db_pattern in content or ".beads/*.db" in content:
            self.print_success(".gitignore already contains Beads exclusion")
            return True

        # Add pattern
        self.print_info(f"Adding '{beads_db_pattern}' to .gitignore")

        if self.dry_run:
            self.print_info(f"[DRY RUN] Would append to .gitignore: {beads_db_pattern}")
            return True

        with open(self.gitignore_path, 'a') as f:
            # Ensure newline before adding
            if not content.endswith('\n'):
                f.write('\n')
            f.write(f"\n# Beads SQLite cache (JSONL is committed)\n{beads_db_pattern}\n")

        self.print_success("Updated .gitignore")
        return True

    def commit_changes(self) -> bool:
        """Commit migration changes"""
        self.print_header("Committing Changes")

        # Check if there are changes to commit
        returncode, stdout, _ = self.run_command(["git", "status", "--porcelain"], check=False)
        if returncode != 0 or not stdout.strip():
            self.print_info("No changes to commit")
            return True

        self.print_info("Changes to commit:")
        print(stdout)

        if not self.ask_confirmation("Commit these changes?"):
            self.print_info("Skipping commit. You can commit manually later.")
            return True

        # Stage files
        files_to_stage = [".ai-pack", ".beads/issues.jsonl", ".gitignore"]
        for file in files_to_stage:
            file_path = self.project_root / file
            if file_path.exists():
                returncode, _, _ = self.run_command(["git", "add", file], check=False)
                if returncode == 0:
                    self.print_success(f"Staged: {file}")

        # Commit
        commit_message = """Migrate to ai-pack v1.1.0 with Beads integration

- Update .ai-pack/ submodule to v1.1.0
- Initialize Beads task tracking
- Add .beads/issues.jsonl (git-backed task database)
- Update .gitignore for .beads/*.db (local cache)

Enables persistent task memory across AI sessions."""

        returncode, _, stderr = self.run_command(
            ["git", "commit", "-m", commit_message],
            check=False
        )

        if returncode != 0:
            self.print_error(f"Commit failed: {stderr}")
            return False

        self.print_success("Changes committed successfully")
        return True

    def verify_migration(self) -> bool:
        """Verify migration was successful"""
        self.print_header("Verifying Migration")

        all_good = True

        # Check Beads is installed
        returncode, stdout, _ = self.run_command(["bd", "--version"], check=False)
        if returncode == 0:
            self.print_success(f"Beads installed: {stdout.strip()}")
        else:
            self.print_error("Beads not found")
            all_good = False

        # Check .beads/ directory
        if self.beads_dir.exists():
            self.print_success(f".beads/ directory exists")
        else:
            self.print_error(".beads/ directory not found")
            all_good = False

        # Check issues.jsonl exists
        issues_jsonl = self.beads_dir / "issues.jsonl"
        if issues_jsonl.exists():
            self.print_success(f".beads/issues.jsonl exists")
        else:
            self.print_error(".beads/issues.jsonl not found")
            all_good = False

        # Check issues.jsonl is tracked by git
        if not self.dry_run:
            returncode, stdout, _ = self.run_command(
                ["git", "ls-files", ".beads/issues.jsonl"],
                check=False
            )
            if returncode == 0 and stdout.strip():
                self.print_success(".beads/issues.jsonl is tracked by git")
            else:
                self.print_warning(".beads/issues.jsonl not tracked by git (stage and commit it)")

        # Check .gitignore contains exclusion
        if self.gitignore_path.exists():
            with open(self.gitignore_path, 'r') as f:
                if ".beads/*.db" in f.read():
                    self.print_success(".gitignore excludes .beads/*.db")
                else:
                    self.print_warning(".gitignore doesn't exclude .beads/*.db")
        else:
            self.print_warning(".gitignore not found")

        # Test Beads commands
        returncode, _, _ = self.run_command(["bd", "list"], check=False)
        if returncode == 0:
            self.print_success("bd list command works")
        else:
            self.print_error("bd list command failed")
            all_good = False

        return all_good

    def print_next_steps(self):
        """Print next steps after migration"""
        self.print_header("Migration Complete!")

        print(f"{Colors.GREEN}✓ Your project has been migrated to ai-pack v1.1.0 with Beads integration{Colors.RESET}\n")

        print(f"{Colors.BOLD}Next Steps:{Colors.RESET}\n")
        print("1. Test Beads integration:")
        print(f"   {Colors.CYAN}bd create 'Test task'{Colors.RESET}")
        print(f"   {Colors.CYAN}bd list{Colors.RESET}")
        print(f"   {Colors.CYAN}bd close <task-id>{Colors.RESET}\n")

        print("2. Push changes to remote:")
        print(f"   {Colors.CYAN}git push{Colors.RESET}\n")

        print("3. Read the Beads integration guide:")
        print(f"   {Colors.CYAN}cat .ai-pack/quality/tooling/beads-integration.md{Colors.RESET}\n")

        print("4. Start using Beads with AI agents:")
        print("   - Orchestrator will use 'bd create' for task breakdown")
        print("   - Engineer will use 'bd ready' to find next work")
        print("   - Tasks persist across AI sessions\n")

        print(f"{Colors.BOLD}Documentation:{Colors.RESET}")
        print(f"   - Migration guide: {Colors.CYAN}.ai-pack/MIGRATION.md{Colors.RESET}")
        print(f"   - Beads integration: {Colors.CYAN}.ai-pack/quality/tooling/beads-integration.md{Colors.RESET}")
        print(f"   - Beads docs: {Colors.CYAN}https://github.com/steveyegge/beads{Colors.RESET}\n")

    def run(self) -> bool:
        """Run full migration process"""
        print(f"\n{Colors.BOLD}{Colors.MAGENTA}AI-Pack Beads Migration Tool{Colors.RESET}")
        print(f"{Colors.MAGENTA}Migrating from v1.0.0 (TodoWrite) to v1.1.0 (Beads){Colors.RESET}")

        if self.dry_run:
            print(f"\n{Colors.YELLOW}Running in DRY RUN mode - no changes will be made{Colors.RESET}")

        # Step 1: Prerequisites
        if not self.check_prerequisites():
            self.print_error("\nPrerequisites not met. Please fix issues above and try again.")
            return False

        # Step 2: Install Beads
        if not self.install_beads():
            self.print_error("\nFailed to install Beads. Please install manually and retry.")
            return False

        # Step 3: Update submodule
        if not self.update_ai_pack_submodule():
            self.print_error("\nFailed to update ai-pack submodule.")
            return False

        # Step 4: Initialize Beads
        if not self.initialize_beads():
            self.print_error("\nFailed to initialize Beads.")
            return False

        # Step 5: Update .gitignore
        if not self.update_gitignore():
            self.print_error("\nFailed to update .gitignore.")
            return False

        # Step 6: Commit changes
        if not self.dry_run:
            self.commit_changes()  # Non-critical if fails

        # Step 7: Verify
        if not self.verify_migration():
            self.print_warning("\nMigration completed with warnings. Check verification output above.")

        # Step 8: Next steps
        self.print_next_steps()

        return True


def main():
    parser = argparse.ArgumentParser(
        description="Migrate ai-pack project from v1.0.0 to v1.1.0 with Beads integration",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  python migrate-to-beads.py              # Interactive migration
  python migrate-to-beads.py --dry-run    # See what would be done
  python migrate-to-beads.py --yes        # Auto-approve all prompts

For more information, see: .ai-pack/MIGRATION.md
        """
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show what would be done without making changes"
    )
    parser.add_argument(
        "--yes", "-y",
        action="store_true",
        help="Skip confirmation prompts (auto-approve)"
    )

    args = parser.parse_args()

    tool = MigrationTool(dry_run=args.dry_run, auto_yes=args.yes)

    try:
        success = tool.run()
        sys.exit(0 if success else 1)
    except KeyboardInterrupt:
        print(f"\n\n{Colors.YELLOW}Migration cancelled by user{Colors.RESET}")
        sys.exit(130)
    except Exception as e:
        print(f"\n{Colors.RED}Unexpected error: {e}{Colors.RESET}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    main()
