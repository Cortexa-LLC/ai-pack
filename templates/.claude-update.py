#!/usr/bin/env python3
"""
AI-Pack Claude Code Integration - UPDATE Script

This script updates existing projects that already have ai-pack to add
the new Claude Code integration features (commands, skills, rules, hooks).

Usage:
    python3 .ai-pack/templates/.claude-update.py

What it does:
- Checks if project already has .claude/ directory
- Merges new templates with existing customizations
- Preserves project-specific commands/skills/rules
- Updates hooks and settings
- Creates backup before updating
"""

import os
import sys
import shutil
from pathlib import Path
from datetime import datetime

# Colors for terminal output
class Colors:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

def print_header(msg):
    print(f"\n{Colors.HEADER}{Colors.BOLD}{'='*60}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{msg}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{'='*60}{Colors.ENDC}\n")

def print_success(msg):
    print(f"{Colors.OKGREEN}✓ {msg}{Colors.ENDC}")

def print_warning(msg):
    print(f"{Colors.WARNING}⚠ {msg}{Colors.ENDC}")

def print_error(msg):
    print(f"{Colors.FAIL}✗ {msg}{Colors.ENDC}")

def print_info(msg):
    print(f"{Colors.OKCYAN}ℹ {msg}{Colors.ENDC}")

def check_prerequisites():
    """Check if we're in the right place."""
    print_header("Checking Prerequisites")

    # Check if .ai-pack exists
    if not os.path.isdir('.ai-pack'):
        print_error("Not in a project with ai-pack submodule")
        print_info("This script should be run from project root with .ai-pack/")
        return False
    print_success("Found .ai-pack/ submodule")

    # Check if templates exist
    template_dir = Path('.ai-pack/templates/.claude')
    if not template_dir.exists():
        print_error("Claude Code templates not found in .ai-pack/")
        print_info("Your ai-pack submodule may need updating:")
        print_info("  git submodule update --remote .ai-pack")
        return False
    print_success("Found Claude Code templates in .ai-pack/")

    return True

def backup_existing():
    """Create backup of existing .claude/ directory."""
    print_header("Creating Backup")

    if not os.path.exists('.claude'):
        print_info("No existing .claude/ directory to backup")
        return None

    timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
    backup_dir = f'.claude.backup.{timestamp}'

    try:
        shutil.copytree('.claude', backup_dir)
        print_success(f"Created backup: {backup_dir}/")
        return backup_dir
    except Exception as e:
        print_error(f"Failed to create backup: {e}")
        return None

def detect_customizations():
    """Detect project-specific customizations in .claude/."""
    print_header("Detecting Customizations")

    if not os.path.exists('.claude'):
        print_info("No existing .claude/ directory")
        return {
            'commands': [],
            'skills': [],
            'rules': [],
            'hooks': [],
            'custom_settings': False
        }

    customizations = {
        'commands': [],
        'skills': [],
        'rules': [],
        'hooks': [],
        'custom_settings': False
    }

    # Framework files (will be overwritten)
    framework_commands = {
        'task-init.md', 'task-status.md', 'orchestrate.md', 'engineer.md',
        'review.md', 'test.md', 'inspect.md', 'architect.md', 'designer.md',
        'pm.md', 'help.md'
    }
    framework_skills = {'orchestrator', 'engineer'}
    framework_rules = {'gates.md', 'task-packets.md', 'workflows.md'}
    framework_hooks = {'task-init.py', 'task-status.py', 'check-task-packet.py'}

    # Check for custom commands
    commands_dir = Path('.claude/commands/ai-pack')
    if commands_dir.exists():
        for cmd in commands_dir.glob('*.md'):
            if cmd.name not in framework_commands:
                customizations['commands'].append(cmd.name)
                print_info(f"Found custom command: {cmd.name}")

    # Check for custom skills
    skills_dir = Path('.claude/skills')
    if skills_dir.exists():
        for skill_dir in skills_dir.iterdir():
            if skill_dir.is_dir() and skill_dir.name not in framework_skills:
                if (skill_dir / 'SKILL.md').exists():
                    customizations['skills'].append(skill_dir.name)
                    print_info(f"Found custom skill: {skill_dir.name}")

    # Check for custom rules
    rules_dir = Path('.claude/rules')
    if rules_dir.exists():
        for rule in rules_dir.glob('*.md'):
            if rule.name not in framework_rules and rule.name != 'README.md':
                customizations['rules'].append(rule.name)
                print_info(f"Found custom rule: {rule.name}")

    # Check for custom hooks
    hooks_dir = Path('.claude/hooks')
    if hooks_dir.exists():
        for hook in hooks_dir.glob('*.py'):
            if hook.name not in framework_hooks:
                customizations['hooks'].append(hook.name)
                print_info(f"Found custom hook: {hook.name}")

    # Check for custom settings
    settings_file = Path('.claude/settings.json')
    if settings_file.exists():
        # Read and check if it differs from template
        try:
            import json
            with open(settings_file, 'r') as f:
                current_settings = json.load(f)

            # Check if there are hooks beyond the standard task-packet check
            hooks = current_settings.get('hooks', {})
            if hooks and len(hooks) > 1 or \
               (hooks.get('UserPromptSubmit') and
                len(hooks['UserPromptSubmit']) > 1):
                customizations['custom_settings'] = True
                print_warning("Found custom settings.json with additional hooks")
        except:
            pass

    if not any([customizations['commands'], customizations['skills'],
                customizations['rules'], customizations['hooks'],
                customizations['custom_settings']]):
        print_success("No customizations detected - safe to fully update")

    return customizations

def update_integration(customizations, backup_dir):
    """Update Claude Code integration."""
    print_header("Updating Claude Code Integration")

    template_dir = Path('.ai-pack/templates/.claude')
    target_dir = Path('.claude')

    # Strategy:
    # 1. Copy all framework files (overwrite)
    # 2. Preserve custom files
    # 3. Merge settings.json if custom

    # Create target directories
    target_dir.mkdir(exist_ok=True)
    (target_dir / 'commands' / 'ai-pack').mkdir(parents=True, exist_ok=True)
    (target_dir / 'skills').mkdir(exist_ok=True)
    (target_dir / 'rules').mkdir(exist_ok=True)
    (target_dir / 'hooks').mkdir(exist_ok=True)

    # Copy framework commands (always update)
    print_info("Updating framework commands...")
    src_commands = template_dir / 'commands' / 'ai-pack'
    dst_commands = target_dir / 'commands' / 'ai-pack'
    for cmd in src_commands.glob('*.md'):
        shutil.copy2(cmd, dst_commands / cmd.name)
    print_success(f"Updated {len(list(src_commands.glob('*.md')))} commands")

    # Copy framework skills (always update)
    print_info("Updating framework skills...")
    for skill in ['orchestrator', 'engineer']:
        src_skill = template_dir / 'skills' / skill
        dst_skill = target_dir / 'skills' / skill
        if src_skill.exists():
            dst_skill.mkdir(exist_ok=True)
            shutil.copy2(src_skill / 'SKILL.md', dst_skill / 'SKILL.md')
    print_success("Updated framework skills")

    # Copy framework rules (always update)
    print_info("Updating framework rules...")
    src_rules = template_dir / 'rules'
    dst_rules = target_dir / 'rules'
    for rule in src_rules.glob('*.md'):
        shutil.copy2(rule, dst_rules / rule.name)
    print_success(f"Updated {len(list(src_rules.glob('*.md')))} rules")

    # Copy framework hooks (always update)
    print_info("Updating framework hooks...")
    src_hooks = template_dir / 'hooks'
    dst_hooks = target_dir / 'hooks'
    for hook in src_hooks.glob('*.py'):
        shutil.copy2(hook, dst_hooks / hook.name)
        # Make executable
        os.chmod(dst_hooks / hook.name, 0o755)
    shutil.copy2(src_hooks / 'README.md', dst_hooks / 'README.md')
    print_success(f"Updated {len(list(src_hooks.glob('*.py')))} hooks")

    # Handle settings.json
    if customizations['custom_settings'] and backup_dir:
        print_warning("Custom settings.json detected")
        print_info("New template saved as: .claude/settings.json.new")
        print_info("Review and merge manually:")
        print_info(f"  diff {backup_dir}/.claude/settings.json .claude/settings.json.new")
        shutil.copy2(template_dir / 'settings.json',
                     target_dir / 'settings.json.new')
    else:
        print_info("Updating settings.json...")
        shutil.copy2(template_dir / 'settings.json',
                     target_dir / 'settings.json')
        print_success("Updated settings.json")

    # Copy main README
    shutil.copy2(template_dir / 'README.md', target_dir / 'README.md')
    print_success("Updated main README.md")

    # Report on preserved customizations
    if customizations['commands']:
        print_success(f"Preserved {len(customizations['commands'])} custom commands")
    if customizations['skills']:
        print_success(f"Preserved {len(customizations['skills'])} custom skills")
    if customizations['rules']:
        print_success(f"Preserved {len(customizations['rules'])} custom rules")
    if customizations['hooks']:
        print_success(f"Preserved {len(customizations['hooks'])} custom hooks")

def update_claude_md():
    """Update CLAUDE.md if it exists and is outdated."""
    print_header("Checking CLAUDE.md")

    if not os.path.exists('CLAUDE.md'):
        print_info("No CLAUDE.md found - skipping")
        return

    with open('CLAUDE.md', 'r') as f:
        content = f.read()

    # Check if it already has Claude Code integration section
    if 'Claude Code Integration' in content:
        print_success("CLAUDE.md already has integration section")
        return

    # Check if it's based on ai-pack template
    if '.ai-pack/gates/' not in content:
        print_info("CLAUDE.md doesn't appear to be from ai-pack template - skipping")
        return

    print_warning("CLAUDE.md missing Claude Code integration section")
    print_info("New template available at: .ai-pack/templates/CLAUDE.md")
    print_info("Review and merge the 'Claude Code Integration' section manually")

def show_summary(backup_dir, customizations):
    """Show summary of what was done."""
    print_header("Update Complete!")

    print(f"{Colors.BOLD}What was updated:{Colors.ENDC}")
    print("  ✓ 11 slash commands (/ai-pack)")
    print("  ✓ 2 auto-triggered skills")
    print("  ✓ 3 modular rules")
    print("  ✓ 3 Python enforcement hooks")
    print("  ✓ Documentation (READMEs)")
    if not customizations['custom_settings']:
        print("  ✓ settings.json")

    if any([customizations['commands'], customizations['skills'],
            customizations['rules'], customizations['hooks']]):
        print(f"\n{Colors.BOLD}What was preserved:{Colors.ENDC}")
        if customizations['commands']:
            print(f"  ✓ {len(customizations['commands'])} custom commands")
        if customizations['skills']:
            print(f"  ✓ {len(customizations['skills'])} custom skills")
        if customizations['rules']:
            print(f"  ✓ {len(customizations['rules'])} custom rules")
        if customizations['hooks']:
            print(f"  ✓ {len(customizations['hooks'])} custom hooks")

    if backup_dir:
        print(f"\n{Colors.BOLD}Backup location:{Colors.ENDC}")
        print(f"  {backup_dir}/")

    if customizations['custom_settings']:
        print(f"\n{Colors.WARNING}{Colors.BOLD}Manual action required:{Colors.ENDC}")
        print(f"{Colors.WARNING}  Review and merge: .claude/settings.json.new{Colors.ENDC}")
        print(f"{Colors.WARNING}  Compare with: {backup_dir}/.claude/settings.json{Colors.ENDC}")

    print(f"\n{Colors.BOLD}Next steps:{Colors.ENDC}")
    print("  1. Test the integration:")
    print("     /ai-pack help")
    print("  2. Commit the updates:")
    print("     git add .claude/")
    print('     git commit -m "Update ai-pack Claude Code integration"')

    print(f"\n{Colors.OKGREEN}{Colors.BOLD}All done! 🎉{Colors.ENDC}\n")

def main():
    print_header("AI-Pack Claude Code Integration - UPDATE")
    print("This script updates existing projects with new integration features.\n")

    # Check prerequisites
    if not check_prerequisites():
        sys.exit(1)

    # Create backup
    backup_dir = backup_existing()

    # Detect customizations
    customizations = detect_customizations()

    # Confirm update
    if customizations['commands'] or customizations['skills'] or \
       customizations['rules'] or customizations['hooks'] or \
       customizations['custom_settings']:
        print(f"\n{Colors.WARNING}Customizations detected.{Colors.ENDC}")
        print("Framework files will be updated, custom files will be preserved.")
        response = input(f"\n{Colors.BOLD}Continue with update? [y/N]: {Colors.ENDC}")
        if response.lower() not in ['y', 'yes']:
            print_info("Update cancelled")
            sys.exit(0)

    # Update integration
    try:
        update_integration(customizations, backup_dir)
        update_claude_md()
        show_summary(backup_dir, customizations)
    except Exception as e:
        print_error(f"Update failed: {e}")
        if backup_dir:
            print_info(f"Restore from backup: {backup_dir}/")
        sys.exit(1)

if __name__ == '__main__':
    main()
