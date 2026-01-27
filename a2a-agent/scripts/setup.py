#!/usr/bin/env python3
"""
AI-Pack Setup Script
Cross-platform installation for AI-Pack Phase 1
"""

import os
import sys
import platform
import subprocess
import shutil
from pathlib import Path


class AIPackInstaller:
    """Cross-platform AI-Pack installer."""

    def __init__(self):
        self.platform = platform.system()
        self.home = Path.home()
        self.install_dir = Path(__file__).parent.absolute()
        self.ai_pack_dir = self.install_dir / ".ai-pack"
        self.errors = []
        self.warnings = []

    def print_banner(self):
        """Print installation banner."""
        print("""
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║                    AI-Pack Installation                          ║
║                        Phase 1 v1.0.0                            ║
║                                                                  ║
║  Cross-platform agent spawning system for Claude Code           ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
        """)
        print(f"Platform: {self.platform}")
        print(f"Python: {sys.version.split()[0]}")
        print(f"Install Directory: {self.install_dir}")
        print()

    def check_prerequisites(self):
        """Check system prerequisites."""
        print("📋 Checking prerequisites...")

        # Check Python version
        if sys.version_info < (3, 8):
            self.errors.append("Python 3.8+ required")
            return False

        print(f"  ✅ Python {sys.version.split()[0]}")

        # Check for Claude Code CLI
        claude_code = shutil.which("claude")
        if claude_code:
            print(f"  ✅ Claude Code CLI found: {claude_code}")
        else:
            self.warnings.append("Claude Code CLI not found - required for agent execution")
            print("  ⚠️  Claude Code CLI not found")

        # Check for git
        git = shutil.which("git")
        if git:
            print("  ✅ Git found")
        else:
            self.warnings.append("Git not found - recommended for version control")

        print()
        return len(self.errors) == 0

    def create_directories(self):
        """Create necessary directories."""
        print("📁 Creating directory structure...")

        dirs = [
            self.ai_pack_dir,
            self.ai_pack_dir / "agents" / "lightweight",
            self.install_dir / "roles",
            self.install_dir / ".beads" / "tasks",
            self.install_dir / "tests",
        ]

        for dir_path in dirs:
            dir_path.mkdir(parents=True, exist_ok=True)
            print(f"  ✅ {dir_path.relative_to(self.install_dir)}")

        print()

    def make_executable(self, file_path):
        """Make a file executable (Unix-like systems)."""
        if self.platform in ["Darwin", "Linux"]:
            os.chmod(file_path, 0o755)

    def setup_bd_command(self):
        """Setup bd spawn command."""
        print("🔧 Setting up bd spawn command...")

        bd_script = self.ai_pack_dir / "bd"
        bd_spawn = self.ai_pack_dir / "bd_spawn.py"

        # Verify files exist
        if not bd_spawn.exists():
            self.errors.append(f"bd_spawn.py not found at {bd_spawn}")
            return False

        if not bd_script.exists():
            self.errors.append(f"bd script not found at {bd_script}")
            return False

        # Make executable
        self.make_executable(bd_script)
        self.make_executable(bd_spawn)

        print("  ✅ bd script configured")
        print("  ✅ bd_spawn.py configured")
        print()
        return True

    def verify_agent_configs(self):
        """Verify agent configuration files."""
        print("⚙️  Verifying agent configurations...")

        agents_dir = self.ai_pack_dir / "agents" / "lightweight"
        expected_agents = ["engineer.yml", "tester.yml", "reviewer.yml"]

        all_found = True
        for agent in expected_agents:
            agent_path = agents_dir / agent
            if agent_path.exists():
                print(f"  ✅ {agent}")
            else:
                self.errors.append(f"Missing agent config: {agent}")
                print(f"  ❌ {agent}")
                all_found = False

        print()
        return all_found

    def verify_role_files(self):
        """Verify role definition files."""
        print("📝 Verifying role definitions...")

        roles_dir = self.install_dir / "roles"
        expected_roles = ["engineer.md", "tester.md", "reviewer.md"]

        all_found = True
        for role in expected_roles:
            role_path = roles_dir / role
            if role_path.exists():
                print(f"  ✅ {role}")
            else:
                self.warnings.append(f"Missing role file: {role}")
                print(f"  ⚠️  {role}")
                all_found = False

        print()
        return all_found

    def setup_gitignore(self):
        """Setup .gitignore for runtime files."""
        print("🔒 Configuring .gitignore...")

        gitignore_path = self.install_dir / ".gitignore"

        # Read existing gitignore
        existing_entries = set()
        if gitignore_path.exists():
            with open(gitignore_path) as f:
                existing_entries = {line.strip() for line in f if line.strip() and not line.startswith('#')}

        # Entries to add
        required_entries = {
            ".beads/",
            "__pycache__/",
            "*.pyc",
            ".DS_Store",
            "*.swp",
            ".vscode/",
            ".idea/",
        }

        # Add missing entries
        new_entries = required_entries - existing_entries
        if new_entries:
            with open(gitignore_path, 'a') as f:
                if existing_entries:
                    f.write("\n# AI-Pack runtime files\n")
                for entry in sorted(new_entries):
                    f.write(f"{entry}\n")
            print(f"  ✅ Added {len(new_entries)} entries to .gitignore")
        else:
            print("  ✅ .gitignore already configured")

        print()

    def register_protocol_handler(self):
        """Register agent:// protocol handler (platform-specific)."""
        print("🔗 Setting up agent:// protocol handler...")

        if self.platform == "Darwin":  # macOS
            self._register_macos_protocol()
        elif self.platform == "Linux":
            self._register_linux_protocol()
        elif self.platform == "Windows":
            self._register_windows_protocol()
        else:
            self.warnings.append(f"Protocol handler not supported on {self.platform}")
            print(f"  ⚠️  Protocol handler not supported on {self.platform}")

        print()

    def _register_macos_protocol(self):
        """Register protocol handler on macOS."""
        # Create handler script
        handler_script = self.ai_pack_dir / "agent_protocol_handler.sh"
        handler_content = f"""#!/bin/bash
# AI-Pack agent:// protocol handler for macOS

AGENT_URL="$1"

# Extract agent type and task from URL
# Format: agent://engineer/task-description
AGENT_TYPE=$(echo "$AGENT_URL" | sed 's|agent://||' | cut -d'/' -f1)
TASK_DESC=$(echo "$AGENT_URL" | sed 's|agent://||' | cut -d'/' -f2-)

# Decode URL encoding
TASK_DESC=$(python3 -c "import urllib.parse; print(urllib.parse.unquote('$TASK_DESC'))")

# Execute bd spawn
cd "{self.install_dir}"
./.ai-pack/bd spawn "$AGENT_TYPE" "$TASK_DESC"
"""
        with open(handler_script, 'w') as f:
            f.write(handler_content)

        self.make_executable(handler_script)
        print("  ✅ Protocol handler script created")
        print("  ℹ️  Manual registration required - see docs/PROTOCOL-HANDLER-SETUP.md")

    def _register_linux_protocol(self):
        """Register protocol handler on Linux."""
        # Create desktop entry
        desktop_entry = self.home / ".local" / "share" / "applications" / "ai-pack-handler.desktop"
        desktop_entry.parent.mkdir(parents=True, exist_ok=True)

        handler_script = self.ai_pack_dir / "agent_protocol_handler.sh"
        handler_content = f"""#!/bin/bash
# AI-Pack agent:// protocol handler for Linux

AGENT_URL="$1"
AGENT_TYPE=$(echo "$AGENT_URL" | sed 's|agent://||' | cut -d'/' -f1)
TASK_DESC=$(echo "$AGENT_URL" | sed 's|agent://||' | cut -d'/' -f2-)
TASK_DESC=$(python3 -c "import urllib.parse; print(urllib.parse.unquote('$TASK_DESC'))")

cd "{self.install_dir}"
./.ai-pack/bd spawn "$AGENT_TYPE" "$TASK_DESC"
"""
        with open(handler_script, 'w') as f:
            f.write(handler_content)
        self.make_executable(handler_script)

        desktop_content = f"""[Desktop Entry]
Type=Application
Name=AI-Pack Agent Handler
Exec={handler_script} %u
MimeType=x-scheme-handler/agent
NoDisplay=true
"""
        with open(desktop_entry, 'w') as f:
            f.write(desktop_content)

        # Update MIME database
        try:
            subprocess.run(["update-desktop-database",
                          str(self.home / ".local" / "share" / "applications")],
                          check=False)
            print("  ✅ Protocol handler registered")
        except Exception as e:
            self.warnings.append(f"Could not update desktop database: {e}")
            print("  ⚠️  Manual MIME database update may be required")

    def _register_windows_protocol(self):
        """Register protocol handler on Windows."""
        handler_script = self.ai_pack_dir / "agent_protocol_handler.bat"
        handler_content = f"""@echo off
REM AI-Pack agent:// protocol handler for Windows

set AGENT_URL=%1
set AGENT_URL=%AGENT_URL:agent://=%

for /f "delims=/ tokens=1" %%a in ("%AGENT_URL%") do set AGENT_TYPE=%%a
for /f "delims=/ tokens=2*" %%a in ("%AGENT_URL%") do set TASK_DESC=%%a

cd /d "{self.install_dir}"
python .ai-pack\\bd_spawn.py %AGENT_TYPE% "%TASK_DESC%"
"""
        with open(handler_script, 'w') as f:
            f.write(handler_content)

        print("  ✅ Protocol handler script created")
        print("  ℹ️  Registry update required - see docs/PROTOCOL-HANDLER-SETUP.md")

    def run_tests(self):
        """Run basic verification tests."""
        print("🧪 Running verification tests...")

        # Test bd spawn
        try:
            result = subprocess.run(
                [sys.executable, str(self.ai_pack_dir / "bd_spawn.py"), "--help"],
                capture_output=True,
                text=True,
                timeout=5
            )
            if result.returncode == 0 or "usage" in result.stdout.lower():
                print("  ✅ bd_spawn.py executable")
            else:
                self.warnings.append("bd_spawn.py may not be working correctly")
        except Exception as e:
            self.warnings.append(f"Could not test bd_spawn.py: {e}")

        print()

    def print_summary(self):
        """Print installation summary."""
        print("\n" + "="*70)

        if self.errors:
            print("\n❌ Installation completed with ERRORS:\n")
            for error in self.errors:
                print(f"  • {error}")
            print("\nPlease fix these errors before using AI-Pack.")
            return False

        if self.warnings:
            print("\n⚠️  Installation completed with warnings:\n")
            for warning in self.warnings:
                print(f"  • {warning}")
            print()

        print("✅ AI-Pack Phase 1 installation COMPLETE!\n")
        print("Next steps:")
        print("  1. Test installation: .ai-pack/bd spawn engineer \"test task\"")
        print("  2. Read usage guide: docs/USAGE-GUIDE.md")
        print("  3. Review examples: tests/")
        print("  4. For protocol handler: docs/PROTOCOL-HANDLER-SETUP.md")
        print()
        print("="*70)
        return True

    def install(self):
        """Run complete installation."""
        self.print_banner()

        if not self.check_prerequisites():
            print("\n❌ Prerequisites check failed")
            self.print_summary()
            return False

        self.create_directories()

        if not self.setup_bd_command():
            print("\n❌ bd command setup failed")
            self.print_summary()
            return False

        self.verify_agent_configs()
        self.verify_role_files()
        self.setup_gitignore()
        self.register_protocol_handler()
        self.run_tests()

        return self.print_summary()


def main():
    """Main entry point."""
    installer = AIPackInstaller()
    success = installer.install()
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
