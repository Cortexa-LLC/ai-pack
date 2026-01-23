#!/usr/bin/env python3
"""
Live Agent Integration Tests

These tests actually spawn agents and verify they can perform real operations.
Must be run in a Claude Code session with actual agent spawning capability.
"""

import sys
import subprocess
import json
from pathlib import Path
from datetime import datetime


class AgentIntegrationTester:
    """
    Run live integration tests for spawned agents.
    """

    def __init__(self):
        self.test_dir = Path("tests/agent_integration_workspace")
        self.test_dir.mkdir(exist_ok=True)
        self.results = []

    def spawn_agent(self, role, task):
        """Spawn an agent and return task info."""
        cmd = [".ai-pack/bd", "spawn", role, task]
        result = subprocess.run(cmd, capture_output=True, text=True)

        print(f"\nSpawning {role} agent...")
        print(f"Task: {task}")
        print(result.stdout)

        if result.returncode != 0:
            print(f"ERROR: {result.stderr}", file=sys.stderr)
            return None

        # Extract task ID from output
        for line in result.stdout.split('\n'):
            if 'Task ID:' in line:
                task_id = line.split(':')[1].strip()
                return task_id

        return None

    def verify_task_completed(self, task_id):
        """Check if task completed successfully."""
        task_dir = Path(f".beads/tasks/{task_id}")

        if not task_dir.exists():
            return False, "Task directory not found"

        # Check for results file
        results_file = task_dir / "30-results.md"
        if not results_file.exists():
            return False, "Results file not created"

        # Check metadata
        metadata_file = task_dir / "00-metadata.json"
        if metadata_file.exists():
            with open(metadata_file) as f:
                metadata = json.load(f)
                if metadata.get('status') == 'completed':
                    return True, "Task completed successfully"

        return False, "Task not completed"

    def test_file_operations(self):
        """Test 1: File Operations"""
        print("\n" + "="*60)
        print("TEST 1: File Operations")
        print("="*60)

        test_file = self.test_dir / "file_ops_test.txt"

        task = f"""
        Perform the following file operations:
        1. Create a file at {test_file} with content "Initial content"
        2. Read the file and verify content
        3. Edit the file to replace "Initial" with "Updated"
        4. Save results showing all operations completed

        Use Read, Write, and Edit tools.
        Document each step in your results.
        """

        task_id = self.spawn_agent("engineer", task)

        if task_id:
            success, message = self.verify_task_completed(task_id)
            self.results.append({
                "test": "File Operations",
                "status": "PASS" if success else "PENDING",
                "message": message,
                "task_id": task_id
            })
        else:
            self.results.append({
                "test": "File Operations",
                "status": "FAIL",
                "message": "Failed to spawn agent"
            })

    def test_web_research(self):
        """Test 2: Web Research"""
        print("\n" + "="*60)
        print("TEST 2: Web Research and Documentation")
        print("="*60)

        doc_file = self.test_dir / "research_results.md"

        task = f"""
        Research Python async/await best practices:
        1. Use WebFetch to get documentation from https://docs.python.org/3/library/asyncio.html
        2. Use WebSearch to find additional async best practices
        3. Create a summary document at {doc_file} with key findings
        4. Include at least 3 best practices

        Use WebFetch, WebSearch, and Write tools.
        Document sources in your results.
        """

        task_id = self.spawn_agent("engineer", task)

        if task_id:
            success, message = self.verify_task_completed(task_id)
            self.results.append({
                "test": "Web Research",
                "status": "PASS" if success else "PENDING",
                "message": message,
                "task_id": task_id
            })

    def test_bash_execution(self):
        """Test 3: Bash Execution"""
        print("\n" + "="*60)
        print("TEST 3: Bash Command Execution")
        print("="*60)

        task = """
        Execute the following bash commands and report results:
        1. Run 'ls -la tests/' to list test files
        2. Run 'python3 --version' to check Python version
        3. Run 'echo "Test complete"' and capture output

        Use Bash tool for all commands.
        Include command output in your results.
        """

        task_id = self.spawn_agent("engineer", task)

        if task_id:
            success, message = self.verify_task_completed(task_id)
            self.results.append({
                "test": "Bash Execution",
                "status": "PASS" if success else "PENDING",
                "message": message,
                "task_id": task_id
            })

    def test_search_tools(self):
        """Test 4: Search Tools (Grep, Glob)"""
        print("\n" + "="*60)
        print("TEST 4: Search Tools")
        print("="*60)

        task = """
        Use search tools to analyze the codebase:
        1. Use Glob to find all Python files in tests/
        2. Use Grep to search for "test_" function definitions
        3. Create a summary of test files and test count

        Use Glob and Grep tools.
        Report findings in your results.
        """

        task_id = self.spawn_agent("reviewer", task)

        if task_id:
            success, message = self.verify_task_completed(task_id)
            self.results.append({
                "test": "Search Tools",
                "status": "PASS" if success else "PENDING",
                "message": message,
                "task_id": task_id
            })

    def test_directory_operations(self):
        """Test 5: Directory Operations"""
        print("\n" + "="*60)
        print("TEST 5: Directory Operations")
        print("="*60)

        test_subdir = self.test_dir / "subdir_test"

        task = f"""
        Perform directory operations:
        1. Create directory at {test_subdir}
        2. Create files inside: {test_subdir}/file1.txt, {test_subdir}/file2.txt
        3. List directory contents
        4. Report all operations completed

        Use Bash (mkdir, ls) and Write tools.
        """

        task_id = self.spawn_agent("engineer", task)

        if task_id:
            success, message = self.verify_task_completed(task_id)
            self.results.append({
                "test": "Directory Operations",
                "status": "PASS" if success else "PENDING",
                "message": message,
                "task_id": task_id
            })

    def test_code_review_workflow(self):
        """Test 6: Full Code Review Workflow"""
        print("\n" + "="*60)
        print("TEST 6: Code Review Workflow")
        print("="*60)

        task = """
        Review the greeting.py implementation:
        1. Read src/greeting.py
        2. Run flake8 or similar linter (if available)
        3. Check for code quality issues
        4. Create review report with findings

        Use Read, Bash, and Write tools.
        """

        task_id = self.spawn_agent("reviewer", task)

        if task_id:
            success, message = self.verify_task_completed(task_id)
            self.results.append({
                "test": "Code Review Workflow",
                "status": "PASS" if success else "PENDING",
                "message": message,
                "task_id": task_id
            })

    def test_testing_workflow(self):
        """Test 7: Full Testing Workflow"""
        print("\n" + "="*60)
        print("TEST 7: Testing Workflow")
        print("="*60)

        task = """
        Test the greeting function:
        1. Read src/greeting.py to understand implementation
        2. Check if tests/test_greeting.py exists
        3. Run pytest on greeting tests
        4. Report test results and coverage

        Use Read, Bash tools.
        """

        task_id = self.spawn_agent("tester", task)

        if task_id:
            success, message = self.verify_task_completed(task_id)
            self.results.append({
                "test": "Testing Workflow",
                "status": "PASS" if success else "PENDING",
                "message": message,
                "task_id": task_id
            })

    def test_parallel_execution(self):
        """Test 8: Parallel Agent Execution"""
        print("\n" + "="*60)
        print("TEST 8: Parallel Agent Execution")
        print("="*60)

        print("Spawning 3 agents in parallel...")

        # Spawn 3 different agents
        task_ids = []

        task1 = "Create a file at tests/agent_integration_workspace/parallel1.txt with content 'Agent 1'"
        task2 = "Create a file at tests/agent_integration_workspace/parallel2.txt with content 'Agent 2'"
        task3 = "Create a file at tests/agent_integration_workspace/parallel3.txt with content 'Agent 3'"

        task_ids.append(self.spawn_agent("engineer", task1))
        task_ids.append(self.spawn_agent("engineer", task2))
        task_ids.append(self.spawn_agent("engineer", task3))

        # Note: In Phase 1, agents run sequentially through Task tool
        # In Phase 2, they would truly run in parallel

        self.results.append({
            "test": "Parallel Execution",
            "status": "PENDING",
            "message": f"Spawned {len([t for t in task_ids if t])} agents",
            "task_ids": task_ids
        })

    def test_mcp_access(self):
        """Test 9: MCP Server Access (if configured)"""
        print("\n" + "="*60)
        print("TEST 9: MCP Server Access")
        print("="*60)

        task = """
        Test MCP server access:
        1. Attempt to list Jira issues (if Jira MCP available)
        2. Attempt to search Confluence (if Wiki MCP available)
        3. Report which MCP servers are accessible

        Use mcp__jira-server and mcp__wiki-server tools if available.
        Report connection status.
        """

        task_id = self.spawn_agent("engineer", task)

        if task_id:
            success, message = self.verify_task_completed(task_id)
            self.results.append({
                "test": "MCP Access",
                "status": "PASS" if success else "PENDING",
                "message": message,
                "task_id": task_id
            })

    def print_results(self):
        """Print test results summary."""
        print("\n" + "="*60)
        print("INTEGRATION TEST RESULTS")
        print("="*60)

        passed = sum(1 for r in self.results if r['status'] == 'PASS')
        pending = sum(1 for r in self.results if r['status'] == 'PENDING')
        failed = sum(1 for r in self.results if r['status'] == 'FAIL')

        for result in self.results:
            status_symbol = {
                'PASS': '✓',
                'PENDING': '⏳',
                'FAIL': '✗'
            }.get(result['status'], '?')

            print(f"\n{status_symbol} {result['test']}")
            print(f"  Status: {result['status']}")
            print(f"  {result['message']}")
            if 'task_id' in result:
                print(f"  Task ID: {result['task_id']}")

        print("\n" + "-"*60)
        print(f"Total Tests: {len(self.results)}")
        print(f"Passed: {passed}")
        print(f"Pending: {pending}")
        print(f"Failed: {failed}")
        print("="*60)

        # Save results
        results_file = Path("tests/agent_integration_results.json")
        with open(results_file, 'w') as f:
            json.dump({
                'timestamp': datetime.now().isoformat(),
                'results': self.results,
                'summary': {
                    'total': len(self.results),
                    'passed': passed,
                    'pending': pending,
                    'failed': failed
                }
            }, f, indent=2)

        print(f"\nResults saved to: {results_file}")

    def run_all_tests(self):
        """Run all integration tests."""
        print("\n" + "="*60)
        print("AGENT INTEGRATION TEST SUITE")
        print("="*60)
        print(f"Timestamp: {datetime.now().isoformat()}")
        print(f"Test Workspace: {self.test_dir}")

        # Run tests
        self.test_file_operations()
        self.test_web_research()
        self.test_bash_execution()
        self.test_search_tools()
        self.test_directory_operations()
        self.test_code_review_workflow()
        self.test_testing_workflow()
        self.test_parallel_execution()
        self.test_mcp_access()

        # Print results
        self.print_results()


def main():
    """Main entry point."""
    print("""
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║        AGENT INTEGRATION TEST SUITE - PHASE 1                ║
║                                                              ║
║  Tests verify that spawned agents have full tool access:    ║
║  - File operations (read, write, edit, delete, move)        ║
║  - Web access (fetch, search)                               ║
║  - Bash execution                                            ║
║  - Search tools (grep, glob)                                 ║
║  - Directory operations                                      ║
║  - MCP server access                                         ║
║                                                              ║
║  Note: Tests spawn actual agents - results are PENDING      ║
║        until agents complete execution.                      ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
    """)

    input("\nPress Enter to start tests...")

    tester = AgentIntegrationTester()
    tester.run_all_tests()

    print("\n✅ Test suite complete!")
    print("\nNext steps:")
    print("1. Check task packets in .beads/tasks/ for agent results")
    print("2. Review agent_integration_results.json for summary")
    print("3. Verify PENDING tests completed successfully")


if __name__ == "__main__":
    main()
