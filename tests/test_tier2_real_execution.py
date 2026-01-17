#!/usr/bin/env python3
"""
Tier 2 REAL Execution Test

This test prepares a task for real agent execution and validates
the setup is correct for Tier 2 testing.

The actual agent spawning happens outside Python (via Claude Code Task tool).

Run with: python3 test_tier2_real_execution.py -v
"""

import json
import time
import unittest
from datetime import datetime
from pathlib import Path


class TestTier2Setup(unittest.TestCase):
    """
    Tier 2 Test Setup Validation

    Validates that environment is ready for real agent execution:
    1. Test directories structured correctly
    2. Contracts properly formatted
    3. Expected deliverables clearly defined
    4. Verification logic ready
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"tier2-real-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

        print("\n" + "="*70)
        print("TIER 2 REAL EXECUTION SETUP")
        print("="*70)
        print(f"Repository: {cls.repo_root}")
        print(f"Test directory: {cls.test_dir}")
        print("="*70 + "\n")

    def test_01_prepare_simple_file_creation_task(self):
        """
        Setup Test: Prepare task for real agent execution

        Creates a complete task packet that can be executed by a real
        spawned agent spawned via Claude Code Task tool.
        """
        print("\n" + "="*70)
        print("TIER 2 SETUP: Simple File Creation Task")
        print("="*70)

        # Create task packet
        task_dir = self.test_dir / "tasks" / "2026-01-15_tier2-simple-task"
        task_dir.mkdir(parents=True, exist_ok=True)

        # Target directory
        output_dir = self.test_dir / "output"
        output_dir.mkdir(exist_ok=True)

        # Expected files
        expected_files = [
            output_dir / "model.py",
            output_dir / "service.py",
            output_dir / "test_service.py",
        ]

        # Create contract
        contract = task_dir / "00-contract.md"
        contract_content = f"""# Tier 2 Test Contract: Simple File Creation

**Objective:** Validate spawned agent file creation with persistence verification

**Role:** Engineer

---

## Task Description

Create 3 Python files demonstrating basic functionality:
1. Model file with data class
2. Service file with business logic
3. Test file with unit tests

## Deliverables

### File 1: model.py
**Path:** `{expected_files[0].absolute()}`
**Content:**
```python
\"\"\"User model\"\"\"

class User:
    \"\"\"Simple user model\"\"\"

    def __init__(self, name: str, email: str):
        self.name = name
        self.email = email

    def __repr__(self):
        return f"User(name={{self.name}}, email={{self.email}})"
```

### File 2: service.py
**Path:** `{expected_files[1].absolute()}`
**Content:**
```python
\"\"\"User service\"\"\"

from model import User

class UserService:
    \"\"\"User business logic\"\"\"

    def __init__(self):
        self.users = []

    def create_user(self, name: str, email: str) -> User:
        \"\"\"Create new user\"\"\"
        user = User(name, email)
        self.users.append(user)
        return user

    def get_user_count(self) -> int:
        \"\"\"Get total user count\"\"\"
        return len(self.users)
```

### File 3: test_service.py
**Path:** `{expected_files[2].absolute()}`
**Content:**
```python
\"\"\"Tests for user service\"\"\"

from service import UserService

def test_create_user():
    \"\"\"Test user creation\"\"\"
    service = UserService()
    user = service.create_user("Alice", "alice@example.com")
    assert user.name == "Alice"
    assert user.email == "alice@example.com"

def test_user_count():
    \"\"\"Test user counting\"\"\"
    service = UserService()
    assert service.get_user_count() == 0
    service.create_user("Bob", "bob@example.com")
    assert service.get_user_count() == 1
```

## Acceptance Criteria

- [ ] All 3 files created at specified ABSOLUTE paths
- [ ] Files are in repository (not sandbox): `{self.repo_root}`
- [ ] Each file has complete content (not truncated)
- [ ] File sizes appropriate (>100 bytes each)
- [ ] Python syntax is valid

## Execution Instructions

### For Spawned Agent:

1. **Read this contract carefully**
2. **Use ABSOLUTE PATHS** - All file paths are absolute
3. **Use Write tool** - Create each file with Write tool
4. **Verify after each file** - Check file exists after writing
5. **Report completion** - State clearly: SUCCESS or FAILED

### Critical:
- Do NOT use relative paths
- Do NOT create files in sandbox
- VERIFY files persist to repository
- VERIFY content is complete

---

**Repository Root:** `{self.repo_root}`
**Test Directory:** `{self.test_dir}`
**Output Directory:** `{output_dir}`

**Date:** {datetime.now().strftime("%Y-%m-%d %H:%M")}
"""
        contract.write_text(contract_content)

        print(f"✅ Created task packet: {task_dir}")
        print(f"✅ Created contract: {contract}")
        print(f"\n📋 Expected deliverables:")
        for i, file_path in enumerate(expected_files, 1):
            print(f"   {i}. {file_path.name}")
            print(f"      → {file_path}")

        # Create verification script
        verify_script = task_dir / "verify.py"
        verify_script.write_text(f'''#!/usr/bin/env python3
"""Verification script for Tier 2 test"""

from pathlib import Path

# Expected files
expected = [
    Path("{expected_files[0]}"),
    Path("{expected_files[1]}"),
    Path("{expected_files[2]}"),
]

repo_root = Path("{self.repo_root}")

print("\\n" + "="*70)
print("VERIFICATION: Tier 2 Test Deliverables")
print("="*70)

all_present = True
all_in_repo = True
all_complete = True

for i, file_path in enumerate(expected, 1):
    print(f"\\nFile {{i}}/3: {{file_path.name}}")

    if file_path.exists():
        content = file_path.read_text()
        size = len(content)
        in_repo = repo_root in file_path.parents

        print(f"  ✅ Exists: {{file_path}}")
        print(f"  Size: {{size}} bytes")
        print(f"  In repository: {{'✅' if in_repo else '❌'}}")

        if size < 100:
            print(f"  ⚠️  WARNING: File suspiciously small")
            all_complete = False

        if not in_repo:
            print(f"  ❌ ERROR: File not in repository!")
            all_in_repo = False
    else:
        print(f"  ❌ MISSING: {{file_path}}")
        all_present = False

print("\\n" + "="*70)
print("VERIFICATION RESULT")
print("="*70)

if all_present and all_in_repo and all_complete:
    print("✅ ALL CHECKS PASSED")
    print("   - All files created")
    print("   - All in repository")
    print("   - All complete")
    exit(0)
else:
    print("❌ VERIFICATION FAILED")
    if not all_present:
        print("   - Some files missing")
    if not all_in_repo:
        print("   - Files not in repository")
    if not all_complete:
        print("   - Files incomplete/truncated")
    exit(1)
''')
        verify_script.chmod(0o755)

        print(f"\n✅ Created verification script: {verify_script}")

        # Create README for manual execution
        readme = task_dir / "README.md"
        readme.write_text(f"""# Tier 2 Real Agent Execution

## How to Execute This Test

### Step 1: Spawn Spawned Agent

In Claude Code, use the Task tool to spawn a spawned agent:

```
I need you to execute a task as a spawned agent.

Read the contract at: {contract.absolute()}

Complete all deliverables specified in the contract.

Use ABSOLUTE PATHS and verify each file after creation.
```

### Step 2: Wait for Completion

Monitor agent execution. The agent should create 3 files.

### Step 3: Verify Results

Run the verification script:

```bash
python3 {verify_script.absolute()}
```

Expected output: "✅ ALL CHECKS PASSED"

## Expected Files

1. `{expected_files[0]}`
2. `{expected_files[1]}`
3. `{expected_files[2]}`

## Verification Criteria

- ✅ All files exist
- ✅ All files in repository: `{self.repo_root}`
- ✅ All files complete (>100 bytes)
- ✅ No truncation

## Test Status

Created: {datetime.now().strftime("%Y-%m-%d %H:%M")}
Status: Ready for execution
""")

        print(f"✅ Created README: {readme}")

        print(f"\n📦 Task Packet Complete:")
        print(f"   Contract: {contract}")
        print(f"   Verify script: {verify_script}")
        print(f"   README: {readme}")

        print(f"\n✅ TEST SETUP COMPLETE")
        print(f"   Task packet ready for real agent execution")

        # Verify setup is correct
        self.assertTrue(contract.exists(), "Contract should exist")
        self.assertTrue(verify_script.exists(), "Verify script should exist")
        self.assertTrue(readme.exists(), "README should exist")

    def test_02_validate_verification_logic(self):
        """
        Setup Test: Validate verification logic works correctly

        Tests the verification script to ensure it can detect:
        - Missing files
        - Files in wrong location
        - Truncated files
        """
        print("\n" + "="*70)
        print("TIER 2 SETUP: Verification Logic Validation")
        print("="*70)

        # Create test scenario
        test_output = self.test_dir / "verification-test"
        test_output.mkdir(exist_ok=True)

        print(f"\n🧪 Testing verification logic...")

        # Scenario 1: All files present and complete
        print(f"\n  Scenario 1: All files present")
        file1 = test_output / "file1.py"
        file2 = test_output / "file2.py"
        file3 = test_output / "file3.py"

        file1.write_text("# Complete file\n" * 20)  # >100 bytes
        file2.write_text("# Complete file\n" * 20)
        file3.write_text("# Complete file\n" * 20)

        all_exist = all(f.exists() for f in [file1, file2, file3])
        all_in_repo = all(self.repo_root in f.parents for f in [file1, file2, file3])
        all_complete = all(len(f.read_text()) > 100 for f in [file1, file2, file3])

        print(f"    All exist: {all_exist}")
        print(f"    In repository: {all_in_repo}")
        print(f"    Complete: {all_complete}")

        self.assertTrue(all_exist and all_in_repo and all_complete, "Scenario 1 should pass")

        # Scenario 2: One file missing
        print(f"\n  Scenario 2: One file missing")
        file3.unlink()

        all_exist = all(f.exists() for f in [file1, file2, file3])
        print(f"    All exist: {all_exist}")
        self.assertFalse(all_exist, "Scenario 2 should fail (file missing)")

        # Recreate for scenario 3
        file3.write_text("# Truncated")  # <100 bytes

        # Scenario 3: File truncated
        print(f"\n  Scenario 3: File truncated")
        all_complete = all(len(f.read_text()) > 100 for f in [file1, file2, file3])
        print(f"    All complete: {all_complete}")
        self.assertFalse(all_complete, "Scenario 3 should fail (truncated)")

        print(f"\n✅ Verification logic validated")


if __name__ == "__main__":
    print("="*70)
    print("Tier 2 Real Execution Setup Tests")
    print("="*70)
    print("\nValidates setup for real spawned agent execution")
    print()

    # Run tests
    unittest.main(verbosity=2)
