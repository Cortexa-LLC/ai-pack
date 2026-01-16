#!/usr/bin/env python3
"""
Agent Output Verification Utility
Implements the 5-step verification protocol for background agents
Cross-platform Python implementation
"""

import sys
import re
import argparse
from pathlib import Path
from typing import List, Tuple, Dict
from dataclasses import dataclass
from enum import Enum

class VerificationStatus(Enum):
    """Verification result status"""
    PASSED = "passed"
    FAILED = "failed"
    WARNING = "warning"

@dataclass
class VerificationResult:
    """Result of a verification step"""
    step: int
    name: str
    status: VerificationStatus
    message: str
    details: List[str]

class AgentOutputVerifier:
    """Verifies background agent output according to orchestrator protocol"""

    # Error patterns to detect
    ERROR_PATTERNS = {
        'token_limit': [
            r'exceeded.*token.*maximum',
            r'output.*limit.*reached',
            r'token.*limit.*exceeded',
            r'maximum.*tokens.*exceeded',
        ],
        'api_error': [
            r'API Error',
            r'rate limit',
            r'timeout',
            r'connection.*error',
        ],
        'permission': [
            r'permission denied',
            r'access denied',
            r'forbidden',
            r'not authorized',
        ],
        'tool_failure': [
            r'tool.*failed',
            r'command.*failed',
            r'operation.*failed',
        ]
    }

    def __init__(self, agent_output_file: Path, working_dir: Path):
        self.agent_output_file = agent_output_file
        self.working_dir = working_dir
        self.agent_output = ""
        self.results: List[VerificationResult] = []

    def load_agent_output(self) -> bool:
        """Load agent output file"""
        try:
            with open(self.agent_output_file, 'r', encoding='utf-8', errors='ignore') as f:
                self.agent_output = f.read()
            return True
        except Exception as e:
            print(f"Error reading agent output: {e}")
            return False

    def step1_check_errors(self) -> VerificationResult:
        """Step 1: Check for error patterns (BLOCKING)"""
        errors_found = []

        for error_type, patterns in self.ERROR_PATTERNS.items():
            for pattern in patterns:
                matches = re.finditer(pattern, self.agent_output, re.IGNORECASE)
                for match in matches:
                    errors_found.append(f"{error_type}: {match.group()}")

        if errors_found:
            return VerificationResult(
                step=1,
                name="Error Pattern Check",
                status=VerificationStatus.FAILED,
                message=f"⚠️ ERROR DETECTED: {len(errors_found)} error(s) found",
                details=errors_found
            )
        else:
            return VerificationResult(
                step=1,
                name="Error Pattern Check",
                status=VerificationStatus.PASSED,
                message="✓ No error patterns found",
                details=[]
            )

    def step2_check_write_calls(self) -> VerificationResult:
        """Step 2: Verify Write() call count"""
        # Count Write() tool calls
        write_pattern = r'Write\('
        write_matches = re.findall(write_pattern, self.agent_output)
        write_count = len(write_matches)

        if write_count == 0:
            return VerificationResult(
                step=2,
                name="Write() Call Analysis",
                status=VerificationStatus.WARNING,
                message=f"⚠️ WARNING: No Write() calls detected",
                details=[
                    "Agent may have failed before reaching implementation",
                    "Or agent only performed read-only operations"
                ]
            )
        else:
            return VerificationResult(
                step=2,
                name="Write() Call Analysis",
                status=VerificationStatus.PASSED,
                message=f"✓ Write() calls detected: {write_count}",
                details=[]
            )

    def step3_extract_claimed_files(self) -> VerificationResult:
        """Step 3: Extract claimed files from agent output"""
        # Patterns for file creation claims
        file_patterns = [
            r'Created:?\s+(.+)',
            r'Wrote to:?\s+(.+)',
            r'Generated:?\s+(.+)',
            r'Saved:?\s+(.+)',
            r'Created file:?\s+(.+)',
        ]

        claimed_files = set()
        for pattern in file_patterns:
            matches = re.finditer(pattern, self.agent_output, re.IGNORECASE)
            for match in matches:
                file_path = match.group(1).strip()
                # Clean up common artifacts
                file_path = file_path.split()[0]  # Take first word
                file_path = file_path.rstrip('.,;:')  # Remove trailing punctuation
                if file_path and not file_path.startswith('*'):
                    claimed_files.add(file_path)

        if not claimed_files:
            return VerificationResult(
                step=3,
                name="Claimed Files Extraction",
                status=VerificationStatus.WARNING,
                message="⚠️ WARNING: No files claimed",
                details=["Agent may have failed before creating files"]
            )
        else:
            return VerificationResult(
                step=3,
                name="Claimed Files Extraction",
                status=VerificationStatus.PASSED,
                message=f"✓ Files claimed: {len(claimed_files)}",
                details=list(claimed_files)
            )

    def step4_verify_files_exist(self, claimed_files: List[str]) -> VerificationResult:
        """Step 4: Verify each claimed file actually exists"""
        if not claimed_files:
            return VerificationResult(
                step=4,
                name="File Existence Verification",
                status=VerificationStatus.WARNING,
                message="⏭️ Skipped (no files claimed)",
                details=[]
            )

        missing_files = []
        empty_files = []
        existing_files = []

        for file_path in claimed_files:
            # Try both absolute and relative to working directory
            full_path = Path(file_path)
            if not full_path.is_absolute():
                full_path = self.working_dir / file_path

            if not full_path.exists():
                missing_files.append(str(file_path))
            elif full_path.stat().st_size == 0:
                empty_files.append(str(file_path))
            else:
                size = full_path.stat().st_size
                existing_files.append(f"{file_path} ({size:,} bytes)")

        details = []
        if existing_files:
            details.extend([f"✓ {f}" for f in existing_files])
        if empty_files:
            details.extend([f"⚠️ EMPTY: {f}" for f in empty_files])
        if missing_files:
            details.extend([f"❌ MISSING: {f}" for f in missing_files])

        if missing_files:
            return VerificationResult(
                step=4,
                name="File Existence Verification",
                status=VerificationStatus.FAILED,
                message=f"❌ MISSING: {len(missing_files)} of {len(claimed_files)} files",
                details=details
            )
        elif empty_files:
            return VerificationResult(
                step=4,
                name="File Existence Verification",
                status=VerificationStatus.WARNING,
                message=f"⚠️ WARNING: {len(empty_files)} empty file(s)",
                details=details
            )
        else:
            return VerificationResult(
                step=4,
                name="File Existence Verification",
                status=VerificationStatus.PASSED,
                message=f"✓ All {len(claimed_files)} file(s) exist and not empty",
                details=details
            )

    def step5_overall_status(self, results: List[VerificationResult]) -> VerificationResult:
        """Step 5: Determine overall status"""
        # Check for blocking failures
        has_failure = any(r.status == VerificationStatus.FAILED for r in results)
        has_warning = any(r.status == VerificationStatus.WARNING for r in results)

        if has_failure:
            # Analyze root cause
            root_causes = []
            for result in results:
                if result.status == VerificationStatus.FAILED:
                    if "token" in result.message.lower():
                        root_causes.append("Token limit exceeded - reduce prompt verbosity")
                    elif "permission" in result.message.lower():
                        root_causes.append("Permission denied - check .claude/settings.json")
                    elif "MISSING" in result.message:
                        root_causes.append("Files not persisted - possible sandbox isolation")
                    else:
                        root_causes.append(f"Failure in {result.name}")

            return VerificationResult(
                step=5,
                name="Overall Status",
                status=VerificationStatus.FAILED,
                message="❌ FAILED - Agent verification failed",
                details=[
                    "Root causes:",
                    *[f"  - {cause}" for cause in root_causes],
                    "",
                    "Required actions:",
                    "  1. Review failure details above",
                    "  2. Fix identified root cause(s)",
                    "  3. Re-spawn agent",
                    "  4. Re-verify completion"
                ]
            )
        elif has_warning:
            return VerificationResult(
                step=5,
                name="Overall Status",
                status=VerificationStatus.WARNING,
                message="⚠️ WARNING - Agent completed with warnings",
                details=[
                    "Review warnings above to ensure expected behavior"
                ]
            )
        else:
            return VerificationResult(
                step=5,
                name="Overall Status",
                status=VerificationStatus.PASSED,
                message="✅ VERIFIED - Agent completed successfully",
                details=[
                    "All verification steps passed",
                    "Files persisted successfully"
                ]
            )

    def verify(self) -> Tuple[bool, List[VerificationResult]]:
        """Execute complete 5-step verification protocol"""
        if not self.load_agent_output():
            return False, []

        # Step 1: Check for errors (BLOCKING)
        step1 = self.step1_check_errors()
        self.results.append(step1)

        if step1.status == VerificationStatus.FAILED:
            # Immediate failure - don't continue
            step5 = self.step5_overall_status(self.results)
            self.results.append(step5)
            return False, self.results

        # Step 2: Check Write() calls
        step2 = self.step2_check_write_calls()
        self.results.append(step2)

        # Step 3: Extract claimed files
        step3 = self.step3_extract_claimed_files()
        self.results.append(step3)

        # Step 4: Verify file existence
        claimed_files = step3.details if step3.status != VerificationStatus.WARNING else []
        step4 = self.step4_verify_files_exist(claimed_files)
        self.results.append(step4)

        # Step 5: Overall status
        step5 = self.step5_overall_status(self.results)
        self.results.append(step5)

        success = step5.status == VerificationStatus.PASSED
        return success, self.results

    def format_report(self, results: List[VerificationResult]) -> str:
        """Format verification results as markdown report"""
        report = "## Agent Completion Verification\n\n"

        for result in results:
            report += f"### Step {result.step}: {result.name}\n\n"
            report += f"**Result:** {result.message}\n\n"

            if result.details:
                report += "**Details:**\n"
                for detail in result.details:
                    if detail:  # Skip empty strings
                        report += f"- {detail}\n"
                report += "\n"

            report += "---\n\n"

        return report

def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(
        description='Verify background agent output according to orchestrator protocol',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
This tool implements the 5-step verification protocol:
  Step 1: Check for error patterns (token limit, API errors, permissions)
  Step 2: Verify Write() call count
  Step 3: Extract claimed files from agent output
  Step 4: Verify each file exists and is not empty
  Step 5: Determine overall status

Examples:
  %(prog)s agent-output.txt
  %(prog)s agent-output.txt --working-dir /path/to/repo
  %(prog)s agent-output.txt --markdown
        """
    )

    parser.add_argument('agent_output_file',
                       type=Path,
                       help='Path to agent output file')
    parser.add_argument('--working-dir',
                       type=Path,
                       default=Path.cwd(),
                       help='Working directory for relative file paths (default: current dir)')
    parser.add_argument('--markdown',
                       action='store_true',
                       help='Output markdown report format')

    args = parser.parse_args()

    # Verify agent output file exists
    if not args.agent_output_file.exists():
        print(f"Error: Agent output file not found: {args.agent_output_file}")
        return 1

    # Create verifier
    verifier = AgentOutputVerifier(args.agent_output_file, args.working_dir)

    # Run verification
    print("Running agent output verification...")
    print()

    success, results = verifier.verify()

    # Output results
    if args.markdown:
        print(verifier.format_report(results))
    else:
        for result in results:
            print(f"Step {result.step}: {result.name}")
            print(f"  {result.message}")
            if result.details:
                for detail in result.details:
                    if detail:
                        print(f"    {detail}")
            print()

    # Exit code
    return 0 if success else 1

if __name__ == '__main__':
    sys.exit(main())
