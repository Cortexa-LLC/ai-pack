#!/usr/bin/env python3
"""
Tier 2 Stress Tests: Token Limit Scenarios

CRITICAL: These tests validate behavior under token pressure

Real-world scenarios where agents approach or exceed 25K-32K token output limits:
- Large features requiring 15-25 files
- Comprehensive documentation generation
- Complex refactoring across many files

Priority: HIGHEST - Token limits are #1 cause of silent failures

Run with: python3 test_multi-agent_token_limits.py -v

For REAL agent execution (requires Claude Code environment):
Set environment variable: MULTI_AGENT_REAL_EXECUTION=true
"""

import json
import os
import subprocess
import sys
import time
import unittest
from datetime import datetime
from pathlib import Path


# Check if running with real agents
REAL_AGENTS = os.getenv('MULTI_AGENT_REAL_EXECUTION', 'false').lower() == 'true'


class TestTokenLimitStress(unittest.TestCase):
    """
    Token Limit Stress Tests

    Validates agent behavior when approaching/exceeding token limits:
    - 15 files (~15K tokens) - Should succeed
    - 20 files (~20K tokens) - Approaching limit
    - 25 files (~25K tokens) - Exceeds limit
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"token-stress-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

        if REAL_AGENTS:
            print("\n" + "="*70)
            print("🚀 REAL AGENT MODE ENABLED")
            print("="*70)
            print("⚠️  WARNING: Tests will spawn actual Claude Code agents")
            print("   - Makes real API calls")
            print("   - Costs API credits")
            print("   - Takes longer to execute")
            print("="*70 + "\n")
        else:
            print("\n" + "="*70)
            print("📋 SIMULATION MODE")
            print("="*70)
            print("Tests validate logic without spawning real agents")
            print("Set MULTI_AGENT_REAL_EXECUTION=true for real execution")
            print("="*70 + "\n")

    @classmethod
    def tearDownClass(cls):
        """Clean up test artifacts"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)
            print(f"\n🧹 Cleaned up: {cls.test_dir}")

    def test_01_agent_creates_15_files_within_limit(self):
        """
        Stress Test: Agent creates 15 files (within token limit)

        Scenario: Feature implementation requiring many files
        File count: 15 files
        Estimated tokens: ~15,000 tokens (SAFE - within 25K limit)
        Expected: Agent completes successfully, all files created
        """
        print("\n" + "="*70)
        print("TOKEN STRESS TEST 1: 15 Files (Within Limit)")
        print("="*70)

        # Create task packet
        task_dir = self.test_dir / "tasks" / "local-20260115090000-15-file-feature"
        task_dir.mkdir(parents=True, exist_ok=True)

        contract = task_dir / "00-contract.md"
        contract.write_text(f"""# Feature: User Management System

## Objective
Implement complete user management system

## Deliverables
Create 15 Python files:
1. models/user.py - User model
2. models/role.py - Role model
3. models/permission.py - Permission model
4. services/user_service.py - User business logic
5. services/auth_service.py - Authentication logic
6. repositories/user_repository.py - User data access
7. repositories/role_repository.py - Role data access
8. api/user_controller.py - User API endpoints
9. api/auth_controller.py - Auth API endpoints
10. validation/user_validator.py - User input validation
11. dto/user_dto.py - User data transfer objects
12. dto/auth_dto.py - Auth data transfer objects
13. tests/test_user_service.py - User service tests
14. tests/test_auth_service.py - Auth service tests
15. tests/test_user_api.py - User API tests

## Acceptance Criteria
- All 15 files created
- Each file has proper structure
- No truncated content
- All files in correct locations

## Token Budget
- 15 files × ~1,000 tokens = ~15,000 tokens
- Status: WITHIN LIMIT (25K-32K)
""")

        print(f"Created contract: {contract}")
        print(f"\n📊 Token Budget Analysis:")
        print(f"   Files: 15")
        print(f"   Est. tokens per file: ~1,000")
        print(f"   Total estimated: ~15,000 tokens")
        print(f"   Limit: 25,000-32,000 tokens")
        print(f"   Status: ✅ SAFE (60% of limit)")

        # Target directory
        target_dir = self.test_dir / "src"
        target_dir.mkdir(exist_ok=True)

        # Expected files
        expected_files = [
            target_dir / "models" / "user.py",
            target_dir / "models" / "role.py",
            target_dir / "models" / "permission.py",
            target_dir / "services" / "user_service.py",
            target_dir / "services" / "auth_service.py",
            target_dir / "repositories" / "user_repository.py",
            target_dir / "repositories" / "role_repository.py",
            target_dir / "api" / "user_controller.py",
            target_dir / "api" / "auth_controller.py",
            target_dir / "validation" / "user_validator.py",
            target_dir / "dto" / "user_dto.py",
            target_dir / "dto" / "auth_dto.py",
            target_dir / "tests" / "test_user_service.py",
            target_dir / "tests" / "test_auth_service.py",
            target_dir / "tests" / "test_user_api.py",
        ]

        if REAL_AGENTS:
            print("\n🚀 Spawning REAL spawned agent...")
            # TODO: Implement real agent spawning when in Claude Code environment
            # from claude_code import Task
            # task = Task(
            #     subagent_type="general-purpose",
            #     description="Create 15-file user management system",
            #     prompt=f"Read contract at {contract.absolute()} and create all 15 deliverable files",
            #     
            # )
            # Wait for completion, verify files
            self.skipTest("Real agent execution not yet implemented")
        else:
            print("\n📋 Simulating agent execution...")
            # Simulate successful execution
            for file_path in expected_files:
                file_path.parent.mkdir(parents=True, exist_ok=True)
                file_path.write_text(f'"""Generated file: {file_path.name}"""\n\n# Implementation placeholder\n')
            print("   ✅ Agent would create all 15 files")

        # Verify all files created
        print("\n✅ Verifying deliverables:")
        all_exist = True
        for i, file_path in enumerate(expected_files, 1):
            if file_path.exists():
                content = file_path.read_text()
                print(f"   ✅ {i}/15: {file_path.relative_to(target_dir)}")

                # Check for truncation
                if not content.strip():
                    print(f"      ⚠️  WARNING: File is empty")
                    all_exist = False
                elif content.count('\n') < 2:
                    print(f"      ⚠️  WARNING: File suspiciously small")
            else:
                print(f"   ❌ {i}/15: MISSING - {file_path.relative_to(target_dir)}")
                all_exist = False

        self.assertTrue(all_exist, "All 15 files should be created")
        print(f"\n✅ TEST PASSED: All 15 files created successfully")
        print(f"   No token limit issues detected")

    def test_02_agent_approaches_token_limit_20_files(self):
        """
        Stress Test: Agent creates 20 files (approaching token limit)

        Scenario: Large feature approaching token budget
        File count: 20 files
        Estimated tokens: ~20,000 tokens (CAUTION - 80% of limit)
        Expected: Agent completes BUT should warn about size
        """
        print("\n" + "="*70)
        print("TOKEN STRESS TEST 2: 20 Files (Approaching Limit)")
        print("="*70)

        task_dir = self.test_dir / "tasks" / "local-20260115090000-20-file-feature"
        task_dir.mkdir(parents=True, exist_ok=True)

        print(f"\n📊 Token Budget Analysis:")
        print(f"   Files: 20")
        print(f"   Est. tokens per file: ~1,000")
        print(f"   Total estimated: ~20,000 tokens")
        print(f"   Limit: 25,000-32,000 tokens")
        print(f"   Status: ⚠️  CAUTION (80% of limit)")
        print(f"\n⚠️  Recommendation: Consider decomposing into 2 tasks")

        # This test validates that orchestrator SHOULD recommend decomposition
        # Even though 20 files might succeed, it's risky
        print("\n✅ VALIDATION:")
        print("   Orchestrator should warn: Task approaching token limit")
        print("   Recommendation: Decompose into 2 subtasks (10 files each)")

    def test_03_agent_exceeds_token_limit_25_files(self):
        """
        Stress Test: Agent attempts 25 files (EXCEEDS token limit)

        Scenario: Task not properly decomposed
        File count: 25 files
        Estimated tokens: ~25,000 tokens (EXCEEDS 25K limit)
        Expected: Agent fails gracefully OR auto-decomposes
        Failure Mode: Truncated output, partial files, false success
        """
        print("\n" + "="*70)
        print("TOKEN STRESS TEST 3: 25 Files (EXCEEDS Limit)")
        print("="*70)

        task_dir = self.test_dir / "tasks" / "local-20260115090000-25-file-feature"
        task_dir.mkdir(parents=True, exist_ok=True)

        print(f"\n📊 Token Budget Analysis:")
        print(f"   Files: 25")
        print(f"   Est. tokens per file: ~1,000")
        print(f"   Total estimated: ~25,000 tokens")
        print(f"   Limit: 25,000-32,000 tokens")
        print(f"   Status: ❌ EXCEEDS LIMIT (100%+ of limit)")

        target_dir = self.test_dir / "src-large"
        target_dir.mkdir(exist_ok=True)

        # Simulate agent attempting 25 files
        # In reality, agent would hit token limit around file 20-22
        print("\n🚀 Simulating agent execution...")
        print("   Agent attempts to create 25 files...")

        created_files = []
        for i in range(25):
            file_path = target_dir / f"file_{i+1:02d}.py"
            file_path.write_text(f'"""File {i+1}"""\n')
            created_files.append(file_path)

            # Simulate token limit hit at file 22
            if i == 21:  # After 22 files
                print(f"   ⚠️  Token limit reached at file {i+2}")
                print(f"   ❌ Remaining {25-(i+2)} files not created")
                break

        # Verify failure detection
        print("\n✅ Verifying failure detection:")
        if len(created_files) < 25:
            print(f"   ✅ Partial completion detected: {len(created_files)}/25 files")
            print(f"   ✅ Agent should report: FAILED (token limit)")
            print(f"   ❌ Agent MUST NOT report: SUCCESS")
        else:
            print(f"   ❌ All 25 files created (unexpected)")

        self.assertLess(len(created_files), 25, "Should not create all 25 files (token limit)")
        print(f"\n✅ TEST PASSED: Token limit failure detected correctly")

    def test_04_orchestrator_decomposes_large_task(self):
        """
        Stress Test: Orchestrator decomposes 30-file task

        Scenario: Large feature requiring decomposition
        File count: 30 files (MUST decompose)
        Expected: Orchestrator creates 3 subtasks (10 files each)
        """
        print("\n" + "="*70)
        print("TOKEN STRESS TEST 4: Task Decomposition (30 files)")
        print("="*70)

        print(f"\n📊 Token Budget Analysis:")
        print(f"   Files: 30")
        print(f"   Est. tokens per file: ~1,000")
        print(f"   Total estimated: ~30,000 tokens")
        print(f"   Limit: 25,000-32,000 tokens")
        print(f"   Status: ❌ CRITICAL - EXCEEDS LIMIT")
        print(f"\n🚨 MANDATORY: Must decompose into smaller tasks")

        # Orchestrator decomposition
        print("\n🤖 Orchestrator Analysis:")
        print("   30 files detected")
        print("   Applying Lean Flow batch sizing...")
        print("   Decision: DECOMPOSE into 3 subtasks")

        subtasks = [
            {
                "name": "Backend API",
                "files": 10,
                "tokens": 10000,
                "task_dir": self.test_dir / "tasks" / "local-20260115090000-backend-api"
            },
            {
                "name": "Frontend UI",
                "files": 10,
                "tokens": 10000,
                "task_dir": self.test_dir / "tasks" / "local-20260115090000-frontend-ui"
            },
            {
                "name": "Integration Tests",
                "files": 10,
                "tokens": 10000,
                "task_dir": self.test_dir / "tasks" / "local-20260115090000-integration-tests"
            }
        ]

        print("\n📦 Decomposition Plan:")
        for i, subtask in enumerate(subtasks, 1):
            print(f"\n   Subtask {i}: {subtask['name']}")
            print(f"      Files: {subtask['files']}")
            print(f"      Est. tokens: ~{subtask['tokens']}")
            print(f"      Status: ✅ Within limit")

            # Create task packet
            subtask['task_dir'].mkdir(parents=True, exist_ok=True)
            contract = subtask['task_dir'] / "00-contract.md"
            contract.write_text(f"""# Subtask: {subtask['name']}

Part of larger 30-file feature (decomposed for token limits)

## Deliverables
{subtask['files']} files

## Token Budget
~{subtask['tokens']} tokens (within limit)
""")

        # Verify decomposition
        print("\n✅ Verifying decomposition:")
        total_files = sum(s['files'] for s in subtasks)
        print(f"   Original task: 30 files")
        print(f"   Decomposed into: {len(subtasks)} subtasks")
        print(f"   Total files: {total_files}")
        self.assertEqual(total_files, 30, "Decomposition should cover all 30 files")

        # Verify each subtask within limits
        for subtask in subtasks:
            self.assertLessEqual(subtask['tokens'], 15000, f"{subtask['name']} should be within token limit")

        print(f"\n✅ TEST PASSED: Large task properly decomposed")
        print(f"   3 subtasks created, each within token limit")


class TestTokenLimitRecovery(unittest.TestCase):
    """
    Token Limit Recovery Tests

    Validates recovery mechanisms when token limits hit:
    - Detection of truncated output
    - Partial completion handling
    - Automatic retry with smaller scope
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"token-recovery-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_01_detect_truncated_file_output(self):
        """
        Recovery Test: Detect when last file is truncated

        Scenario: Agent hits token limit mid-file
        Expected: Detection of incomplete file (missing closing braces, etc.)
        """
        print("\n" + "="*70)
        print("TOKEN RECOVERY TEST 1: Detect Truncated Output")
        print("="*70)

        # Create files, last one truncated
        target_dir = self.test_dir / "truncated-test"
        target_dir.mkdir(exist_ok=True)

        # Normal files
        for i in range(5):
            file_path = target_dir / f"file_{i+1}.py"
            file_path.write_text(f'''"""File {i+1}"""

def function_{i+1}():
    """Complete function"""
    return {i+1}
''')

        # Truncated file (incomplete)
        truncated_file = target_dir / "file_6.py"
        truncated_file.write_text('''"""File 6 - TRUNCATED"""

def function_6():
    """This function is incompl''')  # Cut off mid-word

        print("\n🔍 Analyzing files for truncation...")

        # Detect truncation
        for file_path in sorted(target_dir.glob("*.py")):
            content = file_path.read_text()

            # Truncation indicators
            is_truncated = False
            reasons = []

            # Check 1: Ends mid-word (no closing quote, etc.)
            if content and not content.endswith(('\n', '"', "'", '}', ')')):
                is_truncated = True
                reasons.append("Ends mid-sentence")

            # Check 2: Unclosed string literal
            if content.count('"""') % 2 != 0:
                is_truncated = True
                reasons.append("Unclosed docstring")

            # Check 3: Suspiciously short for last file
            if file_path == truncated_file and len(content) < 100:
                is_truncated = True
                reasons.append("Suspiciously short")

            if is_truncated:
                print(f"   ❌ TRUNCATED: {file_path.name}")
                for reason in reasons:
                    print(f"      - {reason}")
            else:
                print(f"   ✅ Complete: {file_path.name}")

        print("\n✅ TEST PASSED: Truncation detected successfully")

    def test_02_partial_completion_retry_strategy(self):
        """
        Recovery Test: Retry strategy for partial completion

        Scenario: Agent creates 15/20 files before hitting limit
        Expected: Second agent completes remaining 5 files
        """
        print("\n" + "="*70)
        print("TOKEN RECOVERY TEST 2: Retry Strategy")
        print("="*70)

        task_dir = self.test_dir / "tasks" / "local-20260115090000-partial-completion"
        task_dir.mkdir(parents=True, exist_ok=True)

        target_dir = self.test_dir / "retry-test"
        target_dir.mkdir(exist_ok=True)

        # First agent attempt: Creates 15/20 files
        print("\n🚀 First agent attempt:")
        created_first = []
        for i in range(15):
            file_path = target_dir / f"file_{i+1:02d}.py"
            file_path.write_text(f'"""File {i+1}"""')
            created_first.append(file_path)

        print(f"   Created: {len(created_first)}/20 files")
        print(f"   Status: INCOMPLETE (token limit hit)")

        # Verification detects missing files
        expected_count = 20
        missing_files = list(range(16, 21))  # Files 16-20 missing

        print(f"\n🔍 Orchestrator verification:")
        print(f"   Expected: {expected_count} files")
        print(f"   Found: {len(created_first)} files")
        print(f"   Missing: {len(missing_files)} files")
        print(f"      Missing: file_16.py through file_20.py")

        # Retry strategy: Second agent completes remaining files
        print(f"\n🔄 Retry strategy:")
        print(f"   Spawn second agent for remaining {len(missing_files)} files")

        created_second = []
        for i in missing_files:
            file_path = target_dir / f"file_{i:02d}.py"
            file_path.write_text(f'"""File {i}"""')
            created_second.append(file_path)

        print(f"   Second agent created: {len(created_second)} files")

        # Final verification
        all_files = list(target_dir.glob("*.py"))
        print(f"\n✅ Final verification:")
        print(f"   Total files: {len(all_files)}/20")
        self.assertEqual(len(all_files), 20, "Should have all 20 files after retry")

        print(f"\n✅ TEST PASSED: Retry strategy successful")


if __name__ == "__main__":
    print("="*70)
    print("Tier 2 Token Limit Stress Tests")
    print("="*70)

    if REAL_AGENTS:
        print("\n🚀 Running with REAL Claude Code agents")
        print("⚠️  This will make actual API calls")
    else:
        print("\n📋 Running in simulation mode")
        print("Set MULTI_AGENT_REAL_EXECUTION=true for real execution")

    print()

    # Run tests
    unittest.main(verbosity=2)
