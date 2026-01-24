#!/usr/bin/env python3
"""
Real Multi-Agent Workflow Test: User Registration Feature

This test executes a complete feature development workflow using multiple agents:
1. Engineer Agent: Backend API implementation
2. Engineer Agent: Frontend form implementation
3. Tester Agent: Comprehensive test suite
4. Reviewer Agent: Code review

Tests validate:
- Complete workflow execution
- Task packet lineage tracking
- Results aggregation
- Sequential timing (Phase 1)
"""

import subprocess
import time
import json
from pathlib import Path
from datetime import datetime


class UserRegistrationWorkflow:
    """Execute complete user registration feature workflow."""

    def __init__(self):
        self.workflow_dir = Path("tests/workflow_test_workspace")
        self.workflow_dir.mkdir(exist_ok=True)
        self.task_ids = []
        self.timings = []
        self.start_time = None

    def spawn_agent(self, role, task, step_name):
        """Spawn an agent and track timing."""
        print(f"\n{'='*70}")
        print(f"STEP: {step_name}")
        print(f"{'='*70}")
        print(f"Role: {role}")
        print(f"Task: {task[:100]}...")

        start = time.time()

        cmd = [".ai-pack/bd", "spawn", role, task]
        result = subprocess.run(cmd, capture_output=True, text=True)

        spawn_time = time.time() - start

        if result.returncode != 0:
            print(f"❌ Error: {result.stderr}")
            return None, spawn_time

        # Extract task ID
        task_id = None
        for line in result.stdout.split('\n'):
            if 'Task ID:' in line:
                task_id = line.split(':')[1].strip()
                break

        print(f"✅ Spawned: {task_id}")
        print(f"⏱️  Spawn time: {spawn_time:.2f}s")

        self.task_ids.append({
            "step": step_name,
            "role": role,
            "task_id": task_id,
            "spawn_time": spawn_time
        })

        self.timings.append({
            "step": step_name,
            "spawn_time": spawn_time,
            "timestamp": datetime.now().isoformat()
        })

        return task_id, spawn_time

    def run_complete_workflow(self):
        """Execute the complete user registration workflow."""
        print("""
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║        MULTI-AGENT WORKFLOW TEST: User Registration             ║
║                                                                  ║
║  Complete feature development workflow with 4 agents:           ║
║  1. Engineer: Backend API                                        ║
║  2. Engineer: Frontend Form                                      ║
║  3. Tester: Test Suite                                           ║
║  4. Reviewer: Code Review                                        ║
║                                                                  ║
║  Validates: Sequential execution, task tracking, aggregation    ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
        """)

        self.start_time = time.time()

        # Step 1: Backend Implementation
        backend_task = f"""
Implement a user registration backend API:

Create file: {self.workflow_dir}/registration_api.py

Requirements:
1. UserRegistration class with:
   - validate_email(email: str) -> bool
   - validate_password(password: str) -> bool (min 8 chars, 1 digit)
   - register_user(email: str, password: str, name: str) -> dict

2. Return structure:
   - Success: {{"success": True, "user_id": "...", "message": "..."}}
   - Failure: {{"success": False, "error": "..."}}

3. Include docstrings and type hints
4. Handle basic validation errors

Focus on implementation quality, not tests (tester will handle that).
"""

        task_id1, _ = self.spawn_agent("engineer", backend_task, "Step 1: Backend API")

        # Step 2: Frontend Implementation
        frontend_task = f"""
Implement a user registration frontend form handler:

Create file: {self.workflow_dir}/registration_form.py

Requirements:
1. RegistrationForm class with:
   - validate_form_data(data: dict) -> tuple[bool, list[str]]
   - sanitize_input(text: str) -> str
   - prepare_submission(email: str, password: str, name: str) -> dict

2. Validation checks:
   - Required fields present
   - Email format
   - Password strength
   - Name not empty

3. Include docstrings and type hints
4. Return validation errors as list

Focus on implementation quality, not tests.
"""

        task_id2, _ = self.spawn_agent("engineer", frontend_task, "Step 2: Frontend Form")

        # Step 3: Comprehensive Testing
        test_task = f"""
Create comprehensive test suite for user registration:

Create file: {self.workflow_dir}/test_registration.py

Requirements:
1. Test registration_api.py:
   - Valid email validation
   - Invalid email validation
   - Valid password validation
   - Invalid password validation
   - Successful registration
   - Failed registration

2. Test registration_form.py:
   - Valid form data
   - Missing required fields
   - Invalid email format
   - Weak password
   - Input sanitization

3. Use pytest format with clear test names
4. Include edge cases and error conditions
5. Aim for >80% coverage

Create a complete, runnable test suite.
"""

        task_id3, _ = self.spawn_agent("tester", test_task, "Step 3: Test Suite")

        # Step 4: Code Review
        review_task = f"""
Perform code review of user registration implementation:

Files to review:
- {self.workflow_dir}/registration_api.py (backend)
- {self.workflow_dir}/registration_form.py (frontend)
- {self.workflow_dir}/test_registration.py (tests)

Review for:
1. Code quality and clean code principles
2. Security issues (input validation, password handling)
3. Error handling completeness
4. Type hints and documentation
5. Test coverage adequacy

Create review report: {self.workflow_dir}/REVIEW_REPORT.md

Structure:
- Summary
- Findings by file
- Security concerns
- Recommendations
- Approval status (Approved / Changes Requested)

Be thorough but constructive.
"""

        task_id4, _ = self.spawn_agent("reviewer", review_task, "Step 4: Code Review")

        total_time = time.time() - self.start_time

        # Summary
        print(f"\n{'='*70}")
        print(f"WORKFLOW COMPLETE")
        print(f"{'='*70}")
        print(f"\n📊 Workflow Summary:")
        print(f"   Total Steps: 4")
        print(f"   Total Agents: 4 (2 engineers, 1 tester, 1 reviewer)")
        print(f"   Total Spawn Time: {total_time:.2f}s")
        print(f"   Average per Agent: {total_time/4:.2f}s")

        return {
            "total_steps": 4,
            "total_agents": 4,
            "task_ids": self.task_ids,
            "total_time": total_time,
            "avg_spawn_time": total_time / 4
        }

    def verify_task_lineage(self):
        """Verify task packet lineage tracking."""
        print(f"\n{'='*70}")
        print(f"TASK LINEAGE VERIFICATION")
        print(f"{'='*70}")

        for task in self.task_ids:
            task_id = task['task_id']
            if not task_id:
                continue

            task_dir = Path(f".beads/tasks/{task_id}")
            if task_dir.exists():
                metadata_file = task_dir / "00-metadata.json"
                if metadata_file.exists():
                    with open(metadata_file) as f:
                        metadata = json.load(f)

                    print(f"\n✅ {task['step']}")
                    print(f"   Task ID: {task_id}")
                    print(f"   Role: {metadata.get('role')}")
                    print(f"   Spawned by: {metadata.get('spawned_by')}")
                    print(f"   Status: {metadata.get('status')}")

    def verify_deliverables(self):
        """Check if expected files were created."""
        print(f"\n{'='*70}")
        print(f"DELIVERABLES VERIFICATION")
        print(f"{'='*70}")

        expected_files = [
            "registration_api.py",
            "registration_form.py",
            "test_registration.py",
            "REVIEW_REPORT.md"
        ]

        print(f"\nExpected files in {self.workflow_dir}:")
        for filename in expected_files:
            filepath = self.workflow_dir / filename
            if filepath.exists():
                size = filepath.stat().st_size
                print(f"   ✅ {filename} ({size} bytes)")
            else:
                print(f"   ⏳ {filename} (pending agent completion)")

    def save_results(self, workflow_summary):
        """Save workflow results."""
        results_file = Path("tests/workflow_test_results.json")

        results = {
            "timestamp": datetime.now().isoformat(),
            "workflow": "User Registration Feature",
            "summary": workflow_summary,
            "task_lineage": self.task_ids,
            "timings": self.timings,
            "notes": [
                "Phase 1: Sequential execution (agents run one after another)",
                "Phase 2 will enable true parallel execution",
                "This test validates infrastructure and patterns"
            ]
        }

        with open(results_file, 'w') as f:
            json.dump(results, f, indent=2)

        print(f"\n💾 Results saved to: {results_file}")

    def print_next_steps(self):
        """Print next steps for verification."""
        print(f"\n{'='*70}")
        print(f"NEXT STEPS")
        print(f"{'='*70}")
        print(f"""
Agents are executing (sequentially in Phase 1).

To verify completion:
1. Check task packets in .beads/tasks/
2. Review agent results in 30-results.md files
3. Check deliverables in {self.workflow_dir}/
4. Run tests: python {self.workflow_dir}/test_registration.py (when complete)

Task tracking:
- 4 task packets created
- Each has metadata, plan, and prompt
- Results will be in 30-results.md

Phase 1 Note:
This demonstrates sequential workflow execution.
Phase 2 will enable agents to run concurrently for faster completion.
        """)


def main():
    """Main entry point."""
    workflow = UserRegistrationWorkflow()

    # Execute workflow
    summary = workflow.run_complete_workflow()

    # Verify tracking
    workflow.verify_task_lineage()

    # Check deliverables
    workflow.verify_deliverables()

    # Save results
    workflow.save_results(summary)

    # Next steps
    workflow.print_next_steps()

    print("\n✅ Workflow test complete!")


if __name__ == "__main__":
    main()
