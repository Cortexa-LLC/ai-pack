#!/usr/bin/env python3
"""
AI-Pack Test Validation Runner
Executes test cases and generates reports (cross-platform)
"""

import os
import sys
import re
import argparse
from pathlib import Path
from datetime import datetime
from typing import List, Dict, Optional, Tuple
from enum import Enum
import subprocess

# ANSI color codes (with Windows support)
class Colors:
    """Cross-platform color support"""

    def __init__(self):
        # Enable ANSI colors on Windows 10+
        if sys.platform == 'win32':
            try:
                import ctypes
                kernel32 = ctypes.windll.kernel32
                kernel32.SetConsoleMode(kernel32.GetStdHandle(-11), 7)
            except:
                pass

        self.RED = '\033[0;31m'
        self.GREEN = '\033[0;32m'
        self.YELLOW = '\033[1;33m'
        self.BLUE = '\033[0;34m'
        self.NC = '\033[0m'  # No Color

    def red(self, text: str) -> str:
        return f"{self.RED}{text}{self.NC}"

    def green(self, text: str) -> str:
        return f"{self.GREEN}{text}{self.NC}"

    def yellow(self, text: str) -> str:
        return f"{self.YELLOW}{text}{self.NC}"

    def blue(self, text: str) -> str:
        return f"{self.BLUE}{text}{self.NC}"

colors = Colors()

class TestStatus(Enum):
    """Test execution status"""
    PASSED = "passed"
    FAILED = "failed"
    SKIPPED = "skipped"
    DEFERRED = "deferred"

class TestCase:
    """Represents a single test case"""

    def __init__(self, file_path: Path):
        self.file_path = file_path
        self.test_id = self._extract_test_id()
        self.category = file_path.parent.name
        self.priority = self._extract_metadata("Priority")
        self.status = self._extract_metadata("Status")
        self.title = self._extract_title()

    def _extract_test_id(self) -> str:
        """Extract test ID from filename (e.g., TC-BA-001)"""
        match = re.match(r'(TC-[A-Z]+-\d+)', self.file_path.stem)
        return match.group(1) if match else self.file_path.stem

    def _extract_metadata(self, field: str) -> str:
        """Extract metadata field from test case markdown"""
        try:
            with open(self.file_path, 'r', encoding='utf-8') as f:
                content = f.read()
                match = re.search(rf'\*\*{field}:\*\*\s+(.+)', content)
                return match.group(1).strip() if match else "Unknown"
        except Exception as e:
            print(f"Warning: Could not read {self.file_path}: {e}")
            return "Unknown"

    def _extract_title(self) -> str:
        """Extract test case title"""
        try:
            with open(self.file_path, 'r', encoding='utf-8') as f:
                for line in f:
                    if line.startswith('# TC-'):
                        return line.strip('# \n')
            return self.test_id
        except:
            return self.test_id

    def __str__(self) -> str:
        return f"[{self.priority}] {self.title}"

class TestReport:
    """Manages test execution report"""

    def __init__(self, reports_dir: Path, filter_type: str, filter_value: str):
        self.reports_dir = reports_dir
        self.reports_dir.mkdir(parents=True, exist_ok=True)

        timestamp = datetime.now().strftime('%Y-%m-%d-%H%M%S')
        self.report_file = reports_dir / f"{timestamp}-test-run.md"

        self.filter_type = filter_type
        self.filter_value = filter_value
        self.results: List[Dict] = []

    def initialize(self):
        """Create report header"""
        now = datetime.now()

        content = f"""# AI-Pack Test Execution Report

**Date:** {now.strftime('%Y-%m-%d')}
**Time:** {now.strftime('%H:%M:%S')}
**Executor:** {os.getenv('USER', os.getenv('USERNAME', 'Unknown'))}

---

## Test Configuration

**Filter:** {self.filter_type}
**Criteria:** {self.filter_value}

---

## Test Results Summary

"""
        self.report_file.write_text(content, encoding='utf-8')

    def add_result(self, test_case: TestCase, status: TestStatus, notes: str = ""):
        """Add test result"""
        self.results.append({
            'test_id': test_case.test_id,
            'title': test_case.title,
            'status': status,
            'notes': notes
        })

        # Append to report file
        result_content = f"""### {test_case.test_id}: {test_case.title}

**Status:** {"✅ PASSED" if status == TestStatus.PASSED else "❌ FAILED" if status == TestStatus.FAILED else "⏭️ SKIPPED"}
**Notes:** {notes}

---

"""
        with open(self.report_file, 'a', encoding='utf-8') as f:
            f.write(result_content)

    def finalize(self):
        """Finalize report with summary"""
        passed = sum(1 for r in self.results if r['status'] == TestStatus.PASSED)
        failed = sum(1 for r in self.results if r['status'] == TestStatus.FAILED)
        total = passed + failed
        pass_rate = (passed * 100 // total) if total > 0 else 0

        summary = f"""

## Summary Statistics

- **Total Tests:** {total}
- **Passed:** {passed}
- **Failed:** {failed}
- **Pass Rate:** {pass_rate}%

## Recommendations

"""

        if failed > 0:
            summary += f"""⚠️ **CRITICAL:** {failed} test(s) failed.

**Required Actions:**
1. Review failed test cases above
2. Identify root causes
3. Fix identified issues
4. Re-run test suite
5. Verify all tests pass before deployment

**DO NOT deploy workflow changes until all critical tests pass.**
"""
        else:
            summary += """✅ **SUCCESS:** All tests passed.

Workflow changes are validated and ready for deployment.
"""

        summary += f"""

---

**Report Location:** {self.report_file}
**Generated:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
"""

        with open(self.report_file, 'a', encoding='utf-8') as f:
            f.write(summary)

        return passed, failed

class TestRunner:
    """Main test runner"""

    def __init__(self, tests_dir: Path):
        self.tests_dir = tests_dir
        self.validation_dir = tests_dir / 'validation'
        self.reports_dir = tests_dir / 'reports'

        self.categories = [
            'background-agents',
            'orchestrator',
            'gates',
            'workflows',
            'integration'
        ]

    def find_test_cases(self, category: Optional[str] = None,
                       priority: Optional[str] = None,
                       test_id: Optional[str] = None) -> List[TestCase]:
        """Find test cases matching filters"""
        test_cases = []

        # Search categories
        categories_to_search = [category] if category else self.categories

        for cat in categories_to_search:
            cat_dir = self.validation_dir / cat
            if not cat_dir.exists():
                continue

            # Find all TC-*.md files
            for test_file in cat_dir.glob('TC-*.md'):
                test_case = TestCase(test_file)

                # Filter by test_id
                if test_id and not test_case.test_id.startswith(test_id):
                    continue

                # Filter by priority
                if priority and test_case.priority != priority:
                    continue

                test_cases.append(test_case)

        return sorted(test_cases, key=lambda tc: tc.test_id)

    def list_tests(self):
        """List all available test cases"""
        print("Available Test Cases:")
        print()

        for category in self.categories:
            cat_dir = self.validation_dir / category
            if not cat_dir.exists():
                continue

            test_cases = self.find_test_cases(category=category)
            if test_cases:
                print(f"Category: {category}")
                for tc in test_cases:
                    print(f"  {tc}")
                print()

    def open_editor(self, file_path: Path) -> bool:
        """Open file in appropriate editor (cross-platform)"""
        try:
            if sys.platform == 'win32':
                # Windows - try VS Code, then notepad
                if self._try_command(['code', str(file_path)]):
                    return True
                os.startfile(str(file_path))
                return True
            elif sys.platform == 'darwin':
                # macOS - try VS Code, then default app
                if self._try_command(['code', str(file_path)]):
                    return True
                subprocess.run(['open', str(file_path)], check=True)
                return True
            else:
                # Linux - try VS Code, then xdg-open
                if self._try_command(['code', str(file_path)]):
                    return True
                subprocess.run(['xdg-open', str(file_path)], check=True)
                return True
        except Exception as e:
            print(f"  Warning: Could not open editor: {e}")
            return False

    def _try_command(self, cmd: List[str]) -> bool:
        """Try to run a command, return True if successful"""
        try:
            subprocess.run(cmd, check=True, stdout=subprocess.DEVNULL,
                         stderr=subprocess.DEVNULL, timeout=2)
            return True
        except:
            return False

    def run_test(self, test_case: TestCase) -> Tuple[TestStatus, str]:
        """Execute a single test case (manual execution)"""
        print()
        print(colors.yellow(f"Running: {test_case.test_id}"))

        print(f"  Priority: {test_case.priority}")
        print(f"  Status: {test_case.status}")
        print(f"  Title: {test_case.title}")

        # Check if test is active
        if test_case.status != "Active":
            print(colors.yellow(f"  Skipped: Test status is {test_case.status}"))
            return TestStatus.SKIPPED, f"Test status is {test_case.status}"

        # Display test case
        print()
        print(f"  Test case: {test_case.file_path}")
        print("  Follow manual execution steps in test case document")
        print()

        # Manual execution prompt
        execute = input("  Execute test now? (y/n/s=skip): ").strip().lower()

        if execute in ['y', 'yes']:
            # Open test case for manual execution
            print("  Opening test case...")
            self.open_editor(test_case.file_path)

            print()
            result = input("  Test result? (p=pass/f=fail): ").strip().lower()

            if result in ['p', 'pass']:
                print(colors.green("  ✓ PASSED"))
                return TestStatus.PASSED, "Manual execution successful"
            else:
                print(colors.red("  ✗ FAILED"))
                notes = input("  Failure notes: ").strip()
                return TestStatus.FAILED, notes or "Manual execution failed"

        elif execute in ['s', 'skip']:
            print(colors.yellow("  Skipped by user"))
            return TestStatus.SKIPPED, "Skipped by user"

        else:
            print(colors.yellow("  Deferred"))
            return TestStatus.DEFERRED, "Deferred"

    def run_tests(self, category: Optional[str] = None,
                  priority: Optional[str] = None,
                  test_id: Optional[str] = None) -> Tuple[int, int]:
        """Run filtered test cases"""
        # Determine filter description
        if test_id:
            filter_type = "test_id"
            filter_value = test_id
        elif category:
            filter_type = "category"
            filter_value = category
        elif priority:
            filter_type = "priority"
            filter_value = priority
        else:
            filter_type = "all"
            filter_value = "all tests"

        # Create report
        report = TestReport(self.reports_dir, filter_type, filter_value)
        report.initialize()

        print(colors.green("AI-Pack Test Validation Runner"))
        print("=" * 40)
        print()

        # Find test cases
        test_cases = self.find_test_cases(category, priority, test_id)

        if not test_cases:
            print(colors.yellow("No test cases found matching filters"))
            return 0, 0

        print(f"Found {len(test_cases)} test case(s)")

        # Execute tests by category
        current_category = None
        for test_case in test_cases:
            # Print category header
            if test_case.category != current_category:
                current_category = test_case.category
                print()
                print(colors.green(f"Category: {current_category}"))
                print("-" * 40)

            # Run test
            status, notes = self.run_test(test_case)

            # Record result (only if not deferred)
            if status != TestStatus.DEFERRED:
                report.add_result(test_case, status, notes)

        # Finalize report
        passed, failed = report.finalize()

        # Display summary
        print()
        print(colors.green("=" * 40))
        print(colors.green("Test Execution Complete"))
        print(colors.green("=" * 40))
        print()
        print("Results:")
        print(f"  Passed:  {passed}")
        print(f"  Failed:  {failed}")
        print()
        print(f"Report: {report.report_file}")
        print()

        if failed > 0:
            print(colors.red("⚠️  CRITICAL: Some tests failed"))
            print(colors.red("Do NOT deploy until all critical tests pass"))
        else:
            print(colors.green("✅ All tests passed"))

        return passed, failed

def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(
        description='AI-Pack Test Validation Runner',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s --critical                       # Run critical tests
  %(prog)s --category background-agents     # Run background agent tests
  %(prog)s --priority Critical              # Run all critical tests
  %(prog)s --test TC-BA-001                 # Run specific test
  %(prog)s --list                           # List all tests
        """
    )

    parser.add_argument('--category',
                       help='Run tests in specific category')
    parser.add_argument('--priority',
                       help='Run tests with specific priority')
    parser.add_argument('--test',
                       help='Run specific test case by ID')
    parser.add_argument('--critical',
                       action='store_true',
                       help='Run only critical test cases')
    parser.add_argument('--list',
                       action='store_true',
                       help='List all available test cases')

    args = parser.parse_args()

    # Determine tests directory
    script_dir = Path(__file__).parent
    tests_dir = script_dir.parent

    # Create runner
    runner = TestRunner(tests_dir)

    # List tests
    if args.list:
        runner.list_tests()
        return 0

    # Determine filters
    category = args.category
    priority = args.priority
    test_id = args.test

    if args.critical:
        priority = "Critical"

    # Run tests
    passed, failed = runner.run_tests(category, priority, test_id)

    # Exit with appropriate code
    return 1 if failed > 0 else 0

if __name__ == '__main__':
    sys.exit(main())
