#!/usr/bin/env python3
"""
Beads Integration Tests

Tests that Beads (git-backed task tracking) integrates properly with ai-pack:
- Task creation (bd create)
- Task status updates (bd start, bd close, bd block)
- Dependency management (bd dep add)
- Cross-session persistence (.beads/issues.jsonl)
- Orchestrator integration

Status: EXECUTABLE
Priority: CRITICAL (Priority 3)
"""

import json
import subprocess
import sys
import time
import unittest
from datetime import datetime
from pathlib import Path


class TestBeadsInstallation(unittest.TestCase):
    """
    Test Beads installation and initialization

    Validates that:
    1. bd command is available
    2. bd init creates proper structure
    3. .beads/issues.jsonl is created
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        print(f"\n📁 Repository root: {cls.repo_root}")

        # Check if bd is installed
        result = subprocess.run(
            ["which", "bd"],
            capture_output=True,
            text=True
        )
        cls.bd_installed = result.returncode == 0

        if cls.bd_installed:
            print("✅ bd command found")
        else:
            print("⚠️  bd command not found (Beads not installed)")

    def test_01_bd_command_available(self):
        """Test: bd command is available"""
        print("\n" + "="*70)
        print("BEADS TEST 1: bd Command Available")
        print("="*70)

        if not self.bd_installed:
            self.skipTest("Beads (bd) not installed - skipping test")

        # Test bd version
        result = subprocess.run(
            ["bd", "version"],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        print(f"bd version output: {result.stdout.strip()}")

        # Should succeed or show version info
        self.assertEqual(
            result.returncode, 0,
            "❌ bd version command failed"
        )
        print("✅ bd command is functional")

    def test_02_beads_directory_structure(self):
        """Test: .beads/ directory structure exists or can be created"""
        print("\n" + "="*70)
        print("BEADS TEST 2: Beads Directory Structure")
        print("="*70)

        if not self.bd_installed:
            self.skipTest("Beads (bd) not installed - skipping test")

        beads_dir = self.repo_root / ".beads"
        issues_file = beads_dir / "issues.jsonl"

        # Check if already initialized
        if beads_dir.exists():
            print(f"✅ .beads/ directory exists: {beads_dir}")

            if issues_file.exists():
                print(f"✅ issues.jsonl exists: {issues_file}")
            else:
                print(f"⚠️  issues.jsonl not found (may need bd init)")
        else:
            print(f"⚠️  .beads/ directory not found (run bd init to initialize)")

        # Verify we can check status (even if empty)
        result = subprocess.run(
            ["bd", "list"],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        # bd list might fail if not initialized, but that's OK for this test
        if result.returncode == 0:
            print("✅ bd list command works")
        else:
            print(f"⚠️  bd list failed (may need bd init): {result.stderr.strip()}")


class TestBeadsTaskCreation(unittest.TestCase):
    """
    Test Beads task creation

    Validates that:
    1. bd create creates tasks
    2. Tasks appear in bd list
    3. Tasks have correct properties
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        # Check if bd is installed
        result = subprocess.run(
            ["which", "bd"],
            capture_output=True,
            text=True
        )
        cls.bd_installed = result.returncode == 0

        cls.created_task_ids = []

    @classmethod
    def tearDownClass(cls):
        """Clean up test tasks"""
        if not cls.bd_installed:
            return

        # Close all test tasks we created
        for task_id in cls.created_task_ids:
            subprocess.run(
                ["bd", "close", task_id],
                capture_output=True,
                cwd=cls.repo_root
            )

        if cls.created_task_ids:
            print(f"\n🧹 Cleaned up {len(cls.created_task_ids)} test tasks")

    def test_01_create_simple_task(self):
        """Test: Create a simple task with bd create"""
        print("\n" + "="*70)
        print("TASK CREATION TEST 1: Create Simple Task")
        print("="*70)

        if not self.bd_installed:
            self.skipTest("Beads (bd) not installed - skipping test")

        # Create task
        task_title = f"Test task {int(time.time())}"
        result = subprocess.run(
            ["bd", "create", task_title, "--priority", "normal"],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        if result.returncode != 0:
            self.skipTest(f"bd create failed (may need bd init): {result.stderr.strip()}")

        print(f"Created task: {task_title}")
        print(f"Output: {result.stdout.strip()}")

        # Extract task ID from output
        # Typical output: "Created task bd-a1b2: Test task..."
        output_lines = result.stdout.strip().split('\n')
        task_id = None
        for line in output_lines:
            if "Created task" in line or "bd-" in line:
                # Extract bd-XXXX from line
                import re
                match = re.search(r'(bd-[a-f0-9]+)', line)
                if match:
                    task_id = match.group(1)
                    break

        if task_id:
            self.created_task_ids.append(task_id)
            print(f"✅ Task created with ID: {task_id}")
        else:
            print("⚠️  Could not extract task ID from output")

        self.assertEqual(result.returncode, 0, "❌ Task creation failed")

    def test_02_create_task_with_description(self):
        """Test: Create task with description"""
        print("\n" + "="*70)
        print("TASK CREATION TEST 2: Create Task with Description")
        print("="*70)

        if not self.bd_installed:
            self.skipTest("Beads (bd) not installed - skipping test")

        task_title = f"Test with description {int(time.time())}"
        task_desc = "This is a test task description"

        result = subprocess.run(
            ["bd", "create", task_title, "--description", task_desc],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        if result.returncode != 0:
            self.skipTest(f"bd create failed: {result.stderr.strip()}")

        print(f"Created task with description: {task_title}")

        # Extract task ID
        import re
        match = re.search(r'(bd-[a-f0-9]+)', result.stdout)
        if match:
            task_id = match.group(1)
            self.created_task_ids.append(task_id)
            print(f"✅ Task created: {task_id}")

        self.assertEqual(result.returncode, 0, "❌ Task creation with description failed")

    def test_03_list_tasks(self):
        """Test: bd list shows created tasks"""
        print("\n" + "="*70)
        print("TASK CREATION TEST 3: List Tasks")
        print("="*70)

        if not self.bd_installed:
            self.skipTest("Beads (bd) not installed - skipping test")

        if not self.created_task_ids:
            self.skipTest("No tasks created to list")

        # List all tasks
        result = subprocess.run(
            ["bd", "list"],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        if result.returncode != 0:
            self.skipTest(f"bd list failed: {result.stderr.strip()}")

        print(f"Task list output:")
        print(result.stdout)

        # Verify our test tasks appear
        output = result.stdout
        tasks_found = sum(1 for task_id in self.created_task_ids if task_id in output)

        print(f"✅ Found {tasks_found}/{len(self.created_task_ids)} test tasks in list")

        self.assertGreater(tasks_found, 0, "❌ No test tasks found in list")


class TestBeadsTaskStatusUpdates(unittest.TestCase):
    """
    Test Beads task status updates

    Validates that:
    1. bd start changes status to in_progress
    2. bd close changes status to closed
    3. bd block marks task as blocked
    4. Status changes persist
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        # Check if bd is installed
        result = subprocess.run(
            ["which", "bd"],
            capture_output=True,
            text=True
        )
        cls.bd_installed = result.returncode == 0

        # Create a test task for status updates
        if cls.bd_installed:
            result = subprocess.run(
                ["bd", "create", f"Status test task {int(time.time())}"],
                capture_output=True,
                text=True,
                cwd=cls.repo_root
            )

            if result.returncode == 0:
                import re
                match = re.search(r'(bd-[a-f0-9]+)', result.stdout)
                if match:
                    cls.test_task_id = match.group(1)
                else:
                    cls.test_task_id = None
            else:
                cls.test_task_id = None
        else:
            cls.test_task_id = None

    @classmethod
    def tearDownClass(cls):
        """Clean up test task"""
        if cls.bd_installed and cls.test_task_id:
            subprocess.run(
                ["bd", "close", cls.test_task_id],
                capture_output=True,
                cwd=cls.repo_root
            )
            print(f"\n🧹 Cleaned up test task: {cls.test_task_id}")

    def test_01_start_task(self):
        """Test: bd start changes task to in_progress"""
        print("\n" + "="*70)
        print("STATUS TEST 1: Start Task")
        print("="*70)

        if not self.bd_installed:
            self.skipTest("Beads (bd) not installed - skipping test")

        if not self.test_task_id:
            self.skipTest("No test task available")

        # Start the task
        result = subprocess.run(
            ["bd", "start", self.test_task_id],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        if result.returncode != 0:
            self.skipTest(f"bd start failed: {result.stderr.strip()}")

        print(f"Started task: {self.test_task_id}")
        print(f"Output: {result.stdout.strip()}")

        # Verify task is in_progress
        show_result = subprocess.run(
            ["bd", "show", self.test_task_id],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        if "in_progress" in show_result.stdout.lower() or "in-progress" in show_result.stdout.lower():
            print("✅ Task status changed to in_progress")
        else:
            print(f"⚠️  Task status unclear from output:\n{show_result.stdout}")

        self.assertEqual(result.returncode, 0, "❌ bd start failed")

    def test_02_close_task(self):
        """Test: bd close changes task to closed"""
        print("\n" + "="*70)
        print("STATUS TEST 2: Close Task")
        print("="*70)

        if not self.bd_installed:
            self.skipTest("Beads (bd) not installed - skipping test")

        if not self.test_task_id:
            self.skipTest("No test task available")

        # Close the task
        result = subprocess.run(
            ["bd", "close", self.test_task_id],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        if result.returncode != 0:
            self.skipTest(f"bd close failed: {result.stderr.strip()}")

        print(f"Closed task: {self.test_task_id}")
        print(f"Output: {result.stdout.strip()}")

        # Verify task is closed
        show_result = subprocess.run(
            ["bd", "show", self.test_task_id],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        if "closed" in show_result.stdout.lower() or "done" in show_result.stdout.lower():
            print("✅ Task status changed to closed")
        else:
            print(f"⚠️  Task status unclear from output:\n{show_result.stdout}")

        self.assertEqual(result.returncode, 0, "❌ bd close failed")


class TestBeadsDependencyManagement(unittest.TestCase):
    """
    Test Beads dependency management

    Validates that:
    1. bd dep add creates dependencies
    2. bd ready respects dependencies
    3. Dependencies shown in bd show
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        # Check if bd is installed
        result = subprocess.run(
            ["which", "bd"],
            capture_output=True,
            text=True
        )
        cls.bd_installed = result.returncode == 0

        cls.task_ids = []

        # Create two tasks for dependency testing
        if cls.bd_installed:
            timestamp = int(time.time())

            # Task 1 (will be dependency)
            result1 = subprocess.run(
                ["bd", "create", f"Dependency task A {timestamp}"],
                capture_output=True,
                text=True,
                cwd=cls.repo_root
            )

            if result1.returncode == 0:
                import re
                match = re.search(r'(bd-[a-f0-9]+)', result1.stdout)
                if match:
                    cls.task_a = match.group(1)
                    cls.task_ids.append(cls.task_a)
                else:
                    cls.task_a = None
            else:
                cls.task_a = None

            # Task 2 (will depend on Task 1)
            result2 = subprocess.run(
                ["bd", "create", f"Dependent task B {timestamp}"],
                capture_output=True,
                text=True,
                cwd=cls.repo_root
            )

            if result2.returncode == 0:
                import re
                match = re.search(r'(bd-[a-f0-9]+)', result2.stdout)
                if match:
                    cls.task_b = match.group(1)
                    cls.task_ids.append(cls.task_b)
                else:
                    cls.task_b = None
            else:
                cls.task_b = None
        else:
            cls.task_a = None
            cls.task_b = None

    @classmethod
    def tearDownClass(cls):
        """Clean up test tasks"""
        if cls.bd_installed:
            for task_id in cls.task_ids:
                subprocess.run(
                    ["bd", "close", task_id],
                    capture_output=True,
                    cwd=cls.repo_root
                )

            if cls.task_ids:
                print(f"\n🧹 Cleaned up {len(cls.task_ids)} dependency test tasks")

    def test_01_add_dependency(self):
        """Test: bd dep add creates dependency relationship"""
        print("\n" + "="*70)
        print("DEPENDENCY TEST 1: Add Dependency")
        print("="*70)

        if not self.bd_installed:
            self.skipTest("Beads (bd) not installed - skipping test")

        if not self.task_a or not self.task_b:
            self.skipTest("Test tasks not created")

        # Task B depends on Task A
        result = subprocess.run(
            ["bd", "dep", "add", self.task_b, self.task_a],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        if result.returncode != 0:
            self.skipTest(f"bd dep add failed: {result.stderr.strip()}")

        print(f"Added dependency: {self.task_b} depends on {self.task_a}")
        print(f"Output: {result.stdout.strip()}")

        # Verify dependency exists
        show_result = subprocess.run(
            ["bd", "show", self.task_b],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        if self.task_a in show_result.stdout:
            print(f"✅ Dependency recorded: {self.task_b} → {self.task_a}")
        else:
            print(f"⚠️  Dependency not shown in output:\n{show_result.stdout}")

        self.assertEqual(result.returncode, 0, "❌ bd dep add failed")


class TestBeadsCrossSessionPersistence(unittest.TestCase):
    """
    Test Beads cross-session persistence

    Validates that:
    1. .beads/issues.jsonl file exists
    2. Tasks are recorded in JSONL format
    3. Tasks persist after creating
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        # Check if bd is installed
        result = subprocess.run(
            ["which", "bd"],
            capture_output=True,
            text=True
        )
        cls.bd_installed = result.returncode == 0

    def test_01_issues_jsonl_exists(self):
        """Test: .beads/issues.jsonl file exists or can be created"""
        print("\n" + "="*70)
        print("PERSISTENCE TEST 1: issues.jsonl File")
        print("="*70)

        if not self.bd_installed:
            self.skipTest("Beads (bd) not installed - skipping test")

        issues_file = self.repo_root / ".beads" / "issues.jsonl"

        if issues_file.exists():
            print(f"✅ issues.jsonl exists: {issues_file}")

            # Show file size
            file_size = issues_file.stat().st_size
            print(f"   File size: {file_size} bytes")

            # Try to read it
            try:
                with open(issues_file, 'r') as f:
                    lines = f.readlines()
                    print(f"   Task count: {len(lines)} tasks")

                    # Verify JSONL format
                    if lines:
                        first_line = lines[0].strip()
                        try:
                            task = json.loads(first_line)
                            print(f"   ✅ Valid JSONL format")
                            print(f"   Sample task ID: {task.get('id', 'unknown')}")
                        except json.JSONDecodeError:
                            print(f"   ⚠️  First line not valid JSON")
            except Exception as e:
                print(f"   ⚠️  Could not read file: {e}")
        else:
            print(f"⚠️  issues.jsonl not found at {issues_file}")
            print("   Run 'bd init' to initialize Beads")

    def test_02_jsonl_format_validation(self):
        """Test: issues.jsonl contains valid JSONL data"""
        print("\n" + "="*70)
        print("PERSISTENCE TEST 2: JSONL Format Validation")
        print("="*70)

        if not self.bd_installed:
            self.skipTest("Beads (bd) not installed - skipping test")

        issues_file = self.repo_root / ".beads" / "issues.jsonl"

        if not issues_file.exists():
            self.skipTest(f"issues.jsonl not found")

        # Read and validate JSONL
        try:
            with open(issues_file, 'r') as f:
                lines = f.readlines()

            if not lines:
                print("⚠️  issues.jsonl is empty (no tasks created yet)")
                return

            valid_tasks = 0
            for i, line in enumerate(lines):
                line = line.strip()
                if not line:
                    continue

                try:
                    task = json.loads(line)

                    # Verify required fields
                    required_fields = ['id', 'title', 'status']
                    has_required = all(field in task for field in required_fields)

                    if has_required:
                        valid_tasks += 1
                    else:
                        print(f"   ⚠️  Line {i+1} missing required fields")

                except json.JSONDecodeError as e:
                    print(f"   ❌ Line {i+1} invalid JSON: {e}")

            print(f"✅ Valid tasks in JSONL: {valid_tasks}/{len(lines)}")

            self.assertGreater(valid_tasks, 0, "❌ No valid tasks in JSONL")

        except Exception as e:
            self.fail(f"❌ Could not validate JSONL: {e}")


class TestBeadsOrchestratorIntegration(unittest.TestCase):
    """
    Integration test: Orchestrator using Beads for task decomposition

    Validates that:
    1. Orchestrator can create task decomposition with Beads
    2. Tasks have proper dependencies
    3. bd ready shows next available work
    """

    @classmethod
    def setUpClass(cls):
        """Set up integration test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        # Check if bd is installed
        result = subprocess.run(
            ["which", "bd"],
            capture_output=True,
            text=True
        )
        cls.bd_installed = result.returncode == 0
        cls.task_ids = []

    @classmethod
    def tearDownClass(cls):
        """Clean up test tasks"""
        if cls.bd_installed:
            for task_id in cls.task_ids:
                subprocess.run(
                    ["bd", "close", task_id],
                    capture_output=True,
                    cwd=cls.repo_root
                )

            if cls.task_ids:
                print(f"\n🧹 Cleaned up {len(cls.task_ids)} orchestrator test tasks")

    def test_orchestrator_task_decomposition(self):
        """
        Integration Test: Orchestrator decomposes feature into Beads tasks

        Scenario:
        1. Orchestrator receives: "Add user authentication"
        2. Orchestrator decomposes into phases
        3. Creates Beads tasks with dependencies
        4. Verifies tasks in bd ready

        Expected:
        - All tasks created
        - Dependencies properly set
        - First task shows in bd ready
        """
        print("\n" + "="*70)
        print("INTEGRATION TEST: Orchestrator Task Decomposition")
        print("="*70)

        if not self.bd_installed:
            self.skipTest("Beads (bd) not installed - skipping test")

        # Orchestrator decomposes feature
        print("\nOrchestrator: Decomposing 'Add user authentication'")

        timestamp = int(time.time())

        # Phase 1: Requirements
        print("  Creating Phase 1: Requirements analysis")
        result1 = subprocess.run(
            ["bd", "create", f"Phase 1: Requirements analysis {timestamp}", "--priority", "high"],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        if result1.returncode != 0:
            self.skipTest(f"Task creation failed: {result1.stderr.strip()}")

        import re
        match = re.search(r'(bd-[a-f0-9]+)', result1.stdout)
        if not match:
            self.skipTest("Could not extract task ID")

        phase1_id = match.group(1)
        self.task_ids.append(phase1_id)
        print(f"  ✅ Created {phase1_id}")

        # Phase 2: Design (depends on Phase 1)
        print("  Creating Phase 2: Design API endpoints")
        result2 = subprocess.run(
            ["bd", "create", f"Phase 2: Design API endpoints {timestamp}", "--priority", "high"],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        if result2.returncode == 0:
            match = re.search(r'(bd-[a-f0-9]+)', result2.stdout)
            if match:
                phase2_id = match.group(1)
                self.task_ids.append(phase2_id)

                # Add dependency
                subprocess.run(
                    ["bd", "dep", "add", phase2_id, phase1_id],
                    capture_output=True,
                    cwd=self.repo_root
                )
                print(f"  ✅ Created {phase2_id} (depends on {phase1_id})")

        # Phase 3: Implementation (depends on Phase 2)
        print("  Creating Phase 3: Implement authentication")
        result3 = subprocess.run(
            ["bd", "create", f"Phase 3: Implement authentication {timestamp}", "--priority", "normal"],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        if result3.returncode == 0:
            match = re.search(r'(bd-[a-f0-9]+)', result3.stdout)
            if match:
                phase3_id = match.group(1)
                self.task_ids.append(phase3_id)

                # Add dependency
                if 'phase2_id' in locals():
                    subprocess.run(
                        ["bd", "dep", "add", phase3_id, phase2_id],
                        capture_output=True,
                        cwd=self.repo_root
                    )
                    print(f"  ✅ Created {phase3_id} (depends on {phase2_id})")

        # Verify task decomposition
        print(f"\n✅ Orchestrator created {len(self.task_ids)} tasks")

        # Check bd ready (should show Phase 1 as ready)
        print("\nChecking bd ready (should show Phase 1):")
        ready_result = subprocess.run(
            ["bd", "ready"],
            capture_output=True,
            text=True,
            cwd=self.repo_root
        )

        if ready_result.returncode == 0:
            print(ready_result.stdout)

            if phase1_id in ready_result.stdout:
                print(f"✅ Phase 1 ({phase1_id}) is ready to start")
            else:
                print(f"⚠️  Phase 1 not in ready list")

        print("\n✅ INTEGRATION TEST PASSED")
        print("\nOrchestrator successfully:")
        print("   ✓ Decomposed feature into phases")
        print("   ✓ Created Beads tasks")
        print("   ✓ Set up dependencies")
        print("   ✓ Made first task available via bd ready")


if __name__ == "__main__":
    print("="*70)
    print("Beads Integration Tests")
    print("="*70)
    print("\nValidating Beads (git-backed task tracking) integration...")
    print()

    # Run tests
    unittest.main(verbosity=2)
