#!/usr/bin/env python3
"""
AI-Pack Orchestrator Role Enforcement Hook

Prevents Orchestrator from using execution tools (Bash, Write, Edit).
Forces delegation pattern by blocking direct execution.

This hook inspects the conversation context and tool calls to detect:
1. If Orchestrator role is active
2. If execution tools are being used (not delegation)
3. Blocks the submission if violation detected

Exit codes:
  0 - Allow (no violation)
  1 - Error (technical failure)
  2 - Block (Orchestrator using execution tools)
"""

import os
import sys
import json
from pathlib import Path


def is_orchestrator_active(conversation_context):
    """Check if Orchestrator role is currently active."""
    if not conversation_context:
        return False

    # Check recent messages for orchestrator activation
    recent_text = ""
    for msg in conversation_context[-10:]:  # Last 10 messages
        if isinstance(msg, dict) and 'content' in msg:
            if isinstance(msg['content'], str):
                recent_text += msg['content'].lower()
            elif isinstance(msg['content'], list):
                for item in msg['content']:
                    if isinstance(item, dict) and item.get('type') == 'text':
                        recent_text += item.get('text', '').lower()

    # Look for orchestrator activation patterns
    orchestrator_indicators = [
        'orchestrator role',
        'acting as the **orchestrator**',
        '/ai-pack orchestrate',
        'you are now acting as the **orchestrator**',
        'orchestrator - auto-activated'
    ]

    return any(indicator in recent_text for indicator in orchestrator_indicators)


def check_tool_calls(tool_calls):
    """Check if tool calls contain forbidden execution patterns."""
    if not tool_calls:
        return None  # No tool calls to check

    forbidden_patterns = []

    for tool_call in tool_calls:
        tool_name = tool_call.get('name', '')
        tool_input = tool_call.get('input', {})

        # Check Skill tool usage (FORBIDDEN - breaks delegation pattern)
        if tool_name == 'Skill':
            skill = tool_input.get('skill', 'unknown')
            forbidden_patterns.append(
                f"Skill({skill}) - Skills are context only, use Task tool to spawn agents"
            )

        # Check Write tool usage
        if tool_name == 'Write':
            file_path = tool_input.get('file_path', '')
            # Allow only plan documentation
            if '10-plan.md' not in file_path:
                forbidden_patterns.append(f"Write to {file_path} (only 10-plan.md allowed)")

        # Check Edit tool usage
        if tool_name == 'Edit':
            file_path = tool_input.get('file_path', '')
            # Allow only plan documentation
            if '10-plan.md' not in file_path:
                forbidden_patterns.append(f"Edit {file_path} (only 10-plan.md allowed)")

        # Check Bash execution (not allowed except safe commands)
        if tool_name == 'Bash':
            command = tool_input.get('command', '')

            # Allowed: Read-only verification commands
            safe_commands = [
                'cat .claude/settings.json',
                'grep -A',
                'test -f',
                'test -d',
                'ls ',
                'ls\n',
                'pwd',
                'which '
            ]

            # Check if command is safe
            is_safe = any(command.strip().startswith(safe) for safe in safe_commands)

            # Forbidden execution patterns
            forbidden_bash = [
                'dotnet test',
                'dotnet build',
                'npm test',
                'npm run',
                'pytest',
                'cargo test',
                'cargo build',
                'go test',
                'mvn test',
                'gradle test',
                '>',  # Output redirection (writing files)
                '>>',  # Append redirection
                'echo' # Usually used to write
            ]

            has_forbidden = any(pattern in command for pattern in forbidden_bash)

            if not is_safe and has_forbidden:
                forbidden_patterns.append(f"Bash execution: {command[:100]}")

    return forbidden_patterns if forbidden_patterns else None


def main():
    """Hook entry point - called by Claude Code."""
    try:
        # Read stdin (Claude Code passes hook context as JSON)
        hook_input = sys.stdin.read()

        if not hook_input:
            # No input provided, allow by default
            sys.exit(0)

        # Parse hook input
        try:
            hook_data = json.loads(hook_input)
        except json.JSONDecodeError:
            # Invalid JSON, allow by default (fail open)
            sys.exit(0)

        # Extract conversation context and tool calls
        conversation = hook_data.get('conversation', [])
        tool_calls = hook_data.get('toolCalls', [])

        # Check if Orchestrator is active
        if not is_orchestrator_active(conversation):
            # Not orchestrator, allow
            sys.exit(0)

        # Orchestrator IS active - check tool calls
        violations = check_tool_calls(tool_calls)

        if violations:
            # BLOCK - Orchestrator using execution tools
            print("\n" + "="*60)
            print("🛑 ORCHESTRATOR ROLE VIOLATION BLOCKED")
            print("="*60)
            print("\nThe Orchestrator role is attempting to use EXECUTION tools.")
            print("This violates the delegation pattern.")
            print("\n**Violations detected:**")
            for violation in violations:
                print(f"  ❌ {violation}")
            print("\n**What Orchestrator MUST do instead:**")
            print("  ✅ Use Task(...) tool to spawn specialist agents")
            print("  ✅ Delegate execution to Engineer/Tester/Reviewer")
            print("  ✅ Only use Read tool for task packets and logs")
            print("\n**Orchestrator is NOT allowed to:**")
            print("  ❌ Run tests (dotnet test, npm test, etc.)")
            print("  ❌ Run builds (dotnet build, npm run, etc.)")
            print("  ❌ Write or Edit files (except 10-plan.md)")
            print("  ❌ Execute implementation work")
            print("\n**This submission has been BLOCKED.**")
            print("Spawn agents using Task tool instead.")
            print("="*60 + "\n")

            # Exit with code 2 to block
            sys.exit(2)

        # No violations, allow
        sys.exit(0)

    except Exception as e:
        # Technical error - log and fail open (allow)
        print(f"⚠️  Hook error: {e}", file=sys.stderr)
        sys.exit(0)


if __name__ == '__main__':
    main()
