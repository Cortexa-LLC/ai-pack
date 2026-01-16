#!/usr/bin/env python3
"""
AI-Pack Automated Test Runner

Runs all executable Python tests and generates reports.

Usage:
    python3 run_tests.py                    # Run all tests
    python3 run_tests.py --unit             # Run only unit tests
    python3 run_tests.py --integration      # Run only integration tests
    python3 run_tests.py --quick            # Skip slow integration tests
"""

import argparse
import sys
import unittest
from pathlib import Path
from datetime import datetime


def discover_tests(tests_dir: Path, pattern: str = "test_*.py"):
    """Discover all test files matching pattern"""
    loader = unittest.TestLoader()
    suite = loader.discover(
        start_dir=str(tests_dir),
        pattern=pattern,
        top_level_dir=None  # Let unittest figure it out
    )
    return suite


def run_test_suite(suite, verbosity=2):
    """Run test suite and return results"""
    runner = unittest.TextTestRunner(verbosity=verbosity)
    result = runner.run(suite)
    return result


def generate_report(result, report_file: Path):
    """Generate markdown test report"""
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")

    report = f"""# AI-Pack Test Execution Report

**Generated:** {timestamp}
**Tests Run:** {result.testsRun}

---

## Summary

- **Total Tests:** {result.testsRun}
- **Passed:** {result.testsRun - len(result.failures) - len(result.errors) - len(result.skipped)}
- **Failed:** {len(result.failures)}
- **Errors:** {len(result.errors)}
- **Skipped:** {len(result.skipped)}
- **Success Rate:** {(result.testsRun - len(result.failures) - len(result.errors)) * 100 // result.testsRun if result.testsRun > 0 else 0}%

---

## Status

"""

    if result.wasSuccessful():
        report += """✅ **ALL TESTS PASSED**

The framework is working correctly. All automated validations passed.

Safe to proceed with workflow changes.
"""
    else:
        report += f"""❌ **TESTS FAILED**

{len(result.failures)} test(s) failed, {len(result.errors)} error(s) occurred.

**Required Actions:**
1. Review failures below
2. Fix identified issues
3. Re-run tests
4. Verify all pass before deployment

**DO NOT deploy changes until all tests pass.**
"""

    # Add failures
    if result.failures:
        report += "\n\n---\n\n## Failures\n\n"
        for test, traceback in result.failures:
            report += f"### {test}\n\n```\n{traceback}\n```\n\n"

    # Add errors
    if result.errors:
        report += "\n\n---\n\n## Errors\n\n"
        for test, traceback in result.errors:
            report += f"### {test}\n\n```\n{traceback}\n```\n\n"

    # Add skipped
    if result.skipped:
        report += "\n\n---\n\n## Skipped Tests\n\n"
        for test, reason in result.skipped:
            report += f"### {test}\n\n**Reason:** {reason}\n\n"

    report += f"""
---

**Report Location:** {report_file}
**Generated:** {timestamp}
"""

    # Write report
    report_file.parent.mkdir(parents=True, exist_ok=True)
    report_file.write_text(report)

    return report


def main():
    parser = argparse.ArgumentParser(
        description="Run AI-Pack automated tests"
    )
    parser.add_argument(
        "--unit",
        action="store_true",
        help="Run only unit tests"
    )
    parser.add_argument(
        "--integration",
        action="store_true",
        help="Run only integration tests"
    )
    parser.add_argument(
        "--quick",
        action="store_true",
        help="Skip slow integration tests"
    )
    parser.add_argument(
        "-v", "--verbose",
        action="store_true",
        help="Verbose output"
    )

    args = parser.parse_args()

    # Find tests directory
    tests_dir = Path(__file__).parent

    print("="*70)
    print("AI-Pack Automated Test Suite")
    print("="*70)
    print()

    # Determine which tests to run
    if args.unit:
        pattern = "test_[!integration]*.py"
        print("Running: Unit tests only")
    elif args.integration:
        pattern = "test_integration*.py"
        print("Running: Integration tests only")
    elif args.quick:
        pattern = "test_[!integration]*.py"
        print("Running: Quick tests (no integration)")
    else:
        pattern = "test_*.py"
        print("Running: All tests")

    print(f"Tests directory: {tests_dir}")
    print()

    # Discover and run tests
    suite = discover_tests(tests_dir, pattern)
    verbosity = 2 if args.verbose else 1
    result = run_test_suite(suite, verbosity)

    # Generate report
    report_file = tests_dir / "reports" / f"{datetime.now().strftime('%Y-%m-%d-%H%M%S')}-test-run.md"
    report = generate_report(result, report_file)

    print()
    print("="*70)
    print("TEST REPORT")
    print("="*70)
    print()
    print(report)

    # Exit with appropriate code
    if result.wasSuccessful():
        print(f"\n✅ Test report saved to: {report_file}")
        return 0
    else:
        print(f"\n❌ Test report saved to: {report_file}")
        return 1


if __name__ == "__main__":
    sys.exit(main())
