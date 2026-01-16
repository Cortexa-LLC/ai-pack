"""
Tier 2: Sequential Workflow Tests

Tests multi-stage workflows with task handoffs between different agent roles.
Validates that agents can read previous outputs and pass artifacts through gates.
"""

import unittest
from pathlib import Path
from datetime import datetime
import json


class TestTier2SequentialWorkflows(unittest.TestCase):
    """Test sequential workflow execution with role-based handoffs"""

    @classmethod
    def setUpClass(cls):
        """Setup test infrastructure"""
        cls.base_dir = Path(__file__).parent.parent
        cls.test_artifacts = cls.base_dir / ".ai" / "test-artifacts"
        cls.test_artifacts.mkdir(parents=True, exist_ok=True)

    def test_01_prepare_prd_to_tester_workflow(self):
        """
        Setup Test: Prepare 5-stage sequential workflow

        Workflow: PRD → Architect → Engineer → Reviewer → Tester

        Each stage reads the previous stage's output and produces its own
        deliverables. This tests proper artifact handoffs and gate enforcement.

        Execution: Run this test, then spawn agents SEQUENTIALLY using Claude Code
        """
        # Create test directory with timestamp
        timestamp = int(datetime.now().timestamp())
        test_dir = self.test_artifacts / f"tier2-sequential-{timestamp}"
        test_dir.mkdir(parents=True, exist_ok=True)

        # Create README
        readme = test_dir / "README.md"
        readme.write_text("""# Tier 2: Sequential Workflow Test

## Workflow Overview

```
Stage 1: PRD Agent
  ↓ (produces requirements.md)
Stage 2: Architect Agent
  ↓ (reads requirements.md, produces architecture.md)
Stage 3: Engineer Agent
  ↓ (reads architecture.md, produces implementation files)
Stage 4: Reviewer Agent
  ↓ (reads implementation, produces review.md)
Stage 5: Tester Agent
  ↓ (reads all previous outputs, produces test_results.md)
```

## Test Scenario

**Feature:** User Profile Management System

**Workflow Stages:**
1. **PRD Agent**: Defines requirements for user profile feature
2. **Architect Agent**: Designs system architecture
3. **Engineer Agent**: Implements the feature (3 Python files)
4. **Reviewer Agent**: Reviews code quality and compliance
5. **Tester Agent**: Validates implementation against requirements

## Execution

Agents must be spawned SEQUENTIALLY (one at a time):
1. Spawn PRD agent → Wait for completion
2. Spawn Architect agent → Wait for completion
3. Spawn Engineer agent → Wait for completion
4. Spawn Reviewer agent → Wait for completion
5. Spawn Tester agent → Wait for completion

Each agent reads previous outputs before creating new deliverables.
""")

        # Create base directories
        stages_dir = test_dir / "stages"
        output_dir = test_dir / "output"
        stages_dir.mkdir(parents=True, exist_ok=True)
        output_dir.mkdir(parents=True, exist_ok=True)

        # Define workflow stages
        workflow = {
            "feature": "User Profile Management System",
            "stages": [
                {
                    "id": "stage-1-prd",
                    "role": "PRD",
                    "name": "Product Requirements Document",
                    "order": 1,
                    "inputs": [],
                    "outputs": ["requirements.md"],
                    "description": "Define product requirements for user profile management"
                },
                {
                    "id": "stage-2-architect",
                    "role": "Architect",
                    "name": "System Architecture Design",
                    "order": 2,
                    "inputs": ["requirements.md"],
                    "outputs": ["architecture.md"],
                    "description": "Design system architecture based on requirements"
                },
                {
                    "id": "stage-3-engineer",
                    "role": "Engineer",
                    "name": "Implementation",
                    "order": 3,
                    "inputs": ["requirements.md", "architecture.md"],
                    "outputs": [
                        "profile/user_profile.py",
                        "profile/profile_service.py",
                        "tests/test_profile.py"
                    ],
                    "description": "Implement user profile system according to architecture"
                },
                {
                    "id": "stage-4-reviewer",
                    "role": "Reviewer",
                    "name": "Code Review",
                    "order": 4,
                    "inputs": [
                        "requirements.md",
                        "architecture.md",
                        "profile/user_profile.py",
                        "profile/profile_service.py",
                        "tests/test_profile.py"
                    ],
                    "outputs": ["review.md"],
                    "description": "Review implementation for quality and compliance"
                },
                {
                    "id": "stage-5-tester",
                    "role": "Tester",
                    "name": "Testing & Validation",
                    "order": 5,
                    "inputs": [
                        "requirements.md",
                        "architecture.md",
                        "profile/user_profile.py",
                        "profile/profile_service.py",
                        "tests/test_profile.py",
                        "review.md"
                    ],
                    "outputs": ["test_results.md"],
                    "description": "Validate implementation against requirements"
                }
            ]
        }

        # Create contracts for each stage
        for stage in workflow["stages"]:
            stage_dir = stages_dir / stage["id"]
            stage_dir.mkdir(parents=True, exist_ok=True)

            contract = stage_dir / "00-contract.md"

            # Build inputs section
            inputs_section = ""
            if stage["inputs"]:
                inputs_section = "## Input Artifacts\n\n"
                inputs_section += "Read and analyze the following artifacts from previous stages:\n\n"
                for input_file in stage["inputs"]:
                    input_path = output_dir / input_file
                    inputs_section += f"- **{input_file}**: `{input_path.absolute()}`\n"
                inputs_section += "\n**CRITICAL:** You MUST read all input artifacts before creating your deliverables.\n"
            else:
                inputs_section = "## Input Artifacts\n\nNone (first stage in workflow)\n"

            # Build outputs section
            outputs_section = "## Deliverables\n\n"
            for output_file in stage["outputs"]:
                output_path = output_dir / output_file
                outputs_section += f"### {output_file}\n"
                outputs_section += f"**Path:** `{output_path.absolute()}`\n\n"

            # Stage-specific content guidance
            content_guidance = self._get_stage_content_guidance(stage["role"], workflow["feature"])

            # Build acceptance criteria
            acceptance_criteria = f"""## Acceptance Criteria

- [ ] All {len(stage["outputs"])} deliverable(s) created at absolute paths
- [ ] Files in correct output directory
- [ ] Complete content (not truncated)
"""
            if stage["inputs"]:
                acceptance_criteria += f"- [ ] All {len(stage['inputs'])} input artifact(s) read and analyzed\n"
                acceptance_criteria += "- [ ] Deliverables align with previous stages\n"

            if stage["role"] == "Engineer":
                acceptance_criteria += "- [ ] Valid Python syntax\n"

            # Create contract
            contract_content = f"""# {stage["role"]} Contract: {stage["name"]}

**Agent Role:** {stage["role"]}
**Feature:** {workflow["feature"]}
**Workflow Stage:** {stage["order"]} of {len(workflow["stages"])}
**Sequential Workflow:** {" → ".join([s["role"] for s in workflow["stages"]])}

---

## Task Description

{stage["description"]}

{inputs_section}

{outputs_section}

{content_guidance}

{acceptance_criteria}

## Execution

**Working Directory:** `{output_dir.absolute()}`

**Sequential Context:**
- This is stage {stage["order"]} of a {len(workflow["stages"])}-stage workflow
- You must wait for previous stages to complete before starting
- Your deliverables will be used by subsequent stages

**Previous Stages:**
{chr(10).join(f"- {s['role']}: {', '.join(s['outputs'])}" for s in workflow['stages'][:stage['order']-1]) if stage['order'] > 1 else "None (first stage)"}

**Subsequent Stages:**
{chr(10).join(f"- {s['role']}: {s['name']}" for s in workflow['stages'][stage['order']:]) if stage['order'] < len(workflow['stages']) else "None (final stage)"}

**Date:** {datetime.now().strftime('%Y-%m-%d %H:%M')}
"""
            contract.write_text(contract_content)

        # Create metadata
        metadata = {
            "test_type": "sequential_workflow",
            "workflow": workflow,
            "total_stages": len(workflow["stages"]),
            "total_outputs": sum(len(s["outputs"]) for s in workflow["stages"]),
            "test_dir": str(test_dir.absolute()),
            "execution_order": [s["id"] for s in workflow["stages"]],
            "execution_instructions": [
                "1. Review workflow in README.md",
                "2. Start with stage-1-prd contract",
                "3. Spawn PRD agent via Claude Code Task tool",
                "4. Wait for PRD agent completion",
                "5. Spawn Architect agent (reads PRD output)",
                "6. Wait for Architect completion",
                "7. Spawn Engineer agent (reads PRD + Architecture)",
                "8. Wait for Engineer completion",
                "9. Spawn Reviewer agent (reads all previous outputs)",
                "10. Wait for Reviewer completion",
                "11. Spawn Tester agent (reads all previous outputs)",
                "12. Wait for Tester completion",
                "13. Run verification script"
            ]
        }

        metadata_file = test_dir / "test-metadata.json"
        metadata_file.write_text(json.dumps(metadata, indent=2))

        # Create verification script
        verify_script = test_dir / "verify_sequential_workflow.py"
        verify_script.write_text(f'''#!/usr/bin/env python3
"""
Verify Sequential Workflow Test Results

Run this after all 5 stages complete to validate deliverables and handoffs.
"""

from pathlib import Path
import sys
import json

def verify_sequential_workflow():
    """Verify all stages completed correctly"""
    test_dir = Path(__file__).parent
    output_dir = test_dir / "output"

    # Load workflow definition
    metadata_file = test_dir / "test-metadata.json"
    with open(metadata_file) as f:
        metadata = json.load(f)

    workflow = metadata["workflow"]
    stages = workflow["stages"]

    print("🔍 Verifying Sequential Workflow...")
    print(f"Feature: {{workflow['feature']}}")
    print(f"Stages: {{len(stages)}}\\n")

    all_success = True
    total_files = 0

    for stage in stages:
        stage_name = f"Stage {{stage['order']}}: {{stage['role']}}"
        print(f"\\n{{stage_name}}")
        print("─" * 60)

        # Check outputs
        missing_outputs = []
        for output_file in stage["outputs"]:
            output_path = output_dir / output_file
            if output_path.exists():
                size = output_path.stat().st_size
                lines = len(output_path.read_text().splitlines()) if output_path.suffix in ['.md', '.py'] else 0
                print(f"  ✅ {{output_file}}: {{size}} bytes, {{lines}} lines")
                total_files += 1
            else:
                print(f"  ❌ {{output_file}}: MISSING")
                missing_outputs.append(output_file)
                all_success = False

        # Check input dependencies (for stages > 1)
        if stage["order"] > 1:
            print(f"\\n  Dependencies:")
            for input_file in stage["inputs"]:
                input_path = output_dir / input_file
                if input_path.exists():
                    print(f"    ✅ {{input_file}}")
                else:
                    print(f"    ❌ {{input_file}}: MISSING (workflow broken!)")
                    all_success = False

        if missing_outputs:
            print(f"\\n  ⚠️  Stage incomplete: missing {{len(missing_outputs)}} file(s)")

    # Summary
    print("\\n" + "=" * 60)
    print("📊 Workflow Summary")
    print("=" * 60)
    print(f"Total files created: {{total_files}}/{{{metadata['total_outputs']}}}")
    print(f"Workflow integrity: {{'PASS' if all_success else 'FAIL'}}")

    if all_success:
        print("\\n✅ SUCCESS: All stages completed, workflow intact!")

        # Validate handoffs
        print("\\n🔗 Validating Handoffs...")
        for i, stage in enumerate(stages[1:], 1):
            prev_stage = stages[i-1]
            print(f"  {{prev_stage['role']}} → {{stage['role']}}: ", end="")

            # Check that current stage's inputs match previous outputs
            handoff_valid = all((output_dir / out).exists() for out in prev_stage["outputs"])
            if handoff_valid:
                print("✅ Clean handoff")
            else:
                print("❌ Broken handoff")
                all_success = False

        if all_success:
            print("\\n🎉 PERFECT: All handoffs validated!")

        return True
    else:
        print("\\n❌ FAILURE: Workflow incomplete or broken")
        return False

if __name__ == "__main__":
    success = verify_sequential_workflow()
    sys.exit(0 if success else 1)
''')
        verify_script.chmod(0o755)

        # Create execution guide
        execution_guide = test_dir / "EXECUTION-GUIDE.md"
        execution_guide.write_text(f"""# Sequential Workflow Execution Guide

## Overview

This test validates a 5-stage sequential workflow where each agent reads previous outputs and produces new artifacts that are passed to the next stage.

## Workflow Diagram

```
┌─────────────────┐
│   PRD Agent     │ Creates requirements.md
└────────┬────────┘
         ↓
┌─────────────────┐
│ Architect Agent │ Reads requirements.md
└────────┬────────┘ Creates architecture.md
         ↓
┌─────────────────┐
│ Engineer Agent  │ Reads requirements.md + architecture.md
└────────┬────────┘ Creates 3 Python files
         ↓
┌─────────────────┐
│ Reviewer Agent  │ Reads all previous outputs
└────────┬────────┘ Creates review.md
         ↓
┌─────────────────┐
│  Tester Agent   │ Reads all previous outputs
└────────┬────────┘ Creates test_results.md
         ↓
      [DONE]
```

## Execution Steps

### Stage 1: PRD Agent

**Contract:** `stages/stage-1-prd/00-contract.md`

**In Claude Code:**
```
Spawn a PRD agent to create requirements.md for the User Profile Management System.
Use contract at: {test_dir.absolute()}/stages/stage-1-prd/00-contract.md
```

**Wait for completion notification**, then verify:
```bash
ls -lh output/requirements.md
```

---

### Stage 2: Architect Agent

**Contract:** `stages/stage-2-architect/00-contract.md`

**IMPORTANT:** Only start after Stage 1 completes!

**In Claude Code:**
```
Spawn an Architect agent to create architecture.md based on requirements.md.
Use contract at: {test_dir.absolute()}/stages/stage-2-architect/00-contract.md
```

**Wait for completion notification**, then verify:
```bash
ls -lh output/architecture.md
```

---

### Stage 3: Engineer Agent

**Contract:** `stages/stage-3-engineer/00-contract.md`

**IMPORTANT:** Only start after Stage 2 completes!

**In Claude Code:**
```
Spawn an Engineer agent to implement the user profile system.
Use contract at: {test_dir.absolute()}/stages/stage-3-engineer/00-contract.md
```

**Wait for completion notification**, then verify:
```bash
ls -lh output/profile/
ls -lh output/tests/
```

---

### Stage 4: Reviewer Agent

**Contract:** `stages/stage-4-reviewer/00-contract.md`

**IMPORTANT:** Only start after Stage 3 completes!

**In Claude Code:**
```
Spawn a Reviewer agent to review the implementation.
Use contract at: {test_dir.absolute()}/stages/stage-4-reviewer/00-contract.md
```

**Wait for completion notification**, then verify:
```bash
ls -lh output/review.md
```

---

### Stage 5: Tester Agent

**Contract:** `stages/stage-5-tester/00-contract.md`

**IMPORTANT:** Only start after Stage 4 completes!

**In Claude Code:**
```
Spawn a Tester agent to validate the complete implementation.
Use contract at: {test_dir.absolute()}/stages/stage-5-tester/00-contract.md
```

**Wait for completion notification**, then verify:
```bash
ls -lh output/test_results.md
```

---

## Verification

After all 5 stages complete, run:

```bash
cd {test_dir.absolute()}
python verify_sequential_workflow.py
```

Expected output:
```
✅ SUCCESS: All stages completed, workflow intact!
🔗 Validating Handoffs...
  PRD → Architect: ✅ Clean handoff
  Architect → Engineer: ✅ Clean handoff
  Engineer → Reviewer: ✅ Clean handoff
  Reviewer → Tester: ✅ Clean handoff
🎉 PERFECT: All handoffs validated!
```

## Expected Files

```
output/
├── requirements.md         (Stage 1: PRD)
├── architecture.md         (Stage 2: Architect)
├── profile/
│   ├── user_profile.py     (Stage 3: Engineer)
│   └── profile_service.py  (Stage 3: Engineer)
├── tests/
│   └── test_profile.py     (Stage 3: Engineer)
├── review.md               (Stage 4: Reviewer)
└── test_results.md         (Stage 5: Tester)
```

Total: 7 files across 5 stages

## Key Validation Points

1. **Sequential Execution:** Each stage waits for previous completion
2. **Artifact Reading:** Each stage reads previous outputs
3. **Handoff Integrity:** All required inputs exist when needed
4. **Content Quality:** Deliverables are complete and well-structured
5. **Workflow Integration:** Final output integrates all stages
""")

        print(f"✅ Sequential workflow test infrastructure created!")
        print(f"   Test directory: {test_dir.absolute()}")
        print(f"   Stages: {len(workflow['stages'])}")
        print(f"   Total expected files: {sum(len(s['outputs']) for s in workflow['stages'])}")
        print(f"\n📋 Next Steps:")
        print(f"   1. Read: {execution_guide.name}")
        print(f"   2. Start with Stage 1 (PRD agent)")
        print(f"   3. Execute stages SEQUENTIALLY (wait for each to complete)")
        print(f"   4. Run verification script when all stages finish")

        self._test_dir = test_dir
        self.assertTrue(test_dir.exists())
        self.assertEqual(len(list(stages_dir.glob("stage-*"))), 5)

    def _get_stage_content_guidance(self, role: str, feature: str) -> str:
        """Get role-specific content guidance for contracts"""

        guidance = {
            "PRD": f"""## Content Guidance

Create a comprehensive Product Requirements Document including:

### 1. Executive Summary
- Feature overview
- Business objectives
- Target users

### 2. Functional Requirements
List specific requirements:
- **FR-1:** User can create a profile with username, email, bio
- **FR-2:** User can update profile information
- **FR-3:** User can view their profile
- **FR-4:** User can delete their profile
- **FR-5:** System validates email format
- **FR-6:** System ensures username uniqueness

### 3. Non-Functional Requirements
- **NFR-1:** Performance (response time < 200ms)
- **NFR-2:** Security (password hashing required)
- **NFR-3:** Data validation on all inputs

### 4. User Stories
Write 3-5 user stories in format:
"As a [user], I want to [action] so that [benefit]"

### 5. Acceptance Criteria
For each functional requirement, define acceptance criteria

### 6. Out of Scope
Explicitly state what is NOT included in this version""",

            "Architect": f"""## Content Guidance

Create a comprehensive System Architecture Document including:

### 1. Architecture Overview
- High-level system design
- Component diagram (text-based)
- Data flow

### 2. Component Design

**Profile Module:**
- UserProfile class (data model)
- ProfileService class (business logic)
- Responsibilities and interfaces

### 3. Data Model

**UserProfile:**
```
- id: str
- username: str
- email: str
- bio: str
- created_at: datetime
- updated_at: datetime
```

### 4. API Design (if applicable)
- Methods and signatures
- Input/output specifications
- Error handling strategy

### 5. File Structure
```
profile/
  user_profile.py    # Data model
  profile_service.py # Business logic
tests/
  test_profile.py    # Unit tests
```

### 6. Technology Decisions
- Why Python?
- Design patterns used
- Testing approach

### 7. Validation Against Requirements
Map each functional requirement (FR-1, FR-2, etc.) to architectural components""",

            "Engineer": f"""## Content Guidance

Implement the user profile system according to the architecture with 3 files:

### File 1: profile/user_profile.py

Implement UserProfile data model:
- Class with all required fields (id, username, email, bio, timestamps)
- Validation methods (email format, username format)
- Serialization methods (to_dict, from_dict)
- Type hints on all methods
- Docstrings

### File 2: profile/profile_service.py

Implement ProfileService business logic:
- create_profile(username, email, bio) -> UserProfile
- get_profile(profile_id) -> UserProfile
- update_profile(profile_id, **updates) -> UserProfile
- delete_profile(profile_id) -> bool
- validate_username(username) -> bool
- validate_email(email) -> bool
- In-memory storage (dictionary)
- Error handling for duplicates, not found, etc.

### File 3: tests/test_profile.py

Implement comprehensive unit tests:
- test_create_profile_success()
- test_create_profile_duplicate_username()
- test_update_profile()
- test_delete_profile()
- test_email_validation()
- test_username_validation()
- Use Python's unittest or pytest
- Cover happy paths and error cases

**Code Quality Requirements:**
- Valid Python 3.8+ syntax
- Type hints throughout
- Docstrings for all classes and methods
- Error handling with appropriate exceptions
- Clean, readable code""",

            "Reviewer": f"""## Content Guidance

Create a comprehensive Code Review Document including:

### 1. Review Summary
- Overall assessment (APPROVED / NEEDS CHANGES / REJECTED)
- Key findings summary

### 2. Requirements Compliance
Review each functional requirement from requirements.md:
- FR-1: ✅/❌ Implementation assessment
- FR-2: ✅/❌ Implementation assessment
- (etc. for all FRs)

### 3. Architecture Compliance
Verify implementation follows architecture.md:
- File structure matches design: ✅/❌
- Classes implement specified interfaces: ✅/❌
- Data model matches specification: ✅/❌

### 4. Code Quality Review

**Strengths:**
- List positive aspects (good practices, clean code, etc.)

**Issues Found:**
- **Critical:** Issues that must be fixed
- **Major:** Issues that should be fixed
- **Minor:** Nice-to-have improvements

Example issue format:
```
**Issue:** Missing email validation in UserProfile
**Severity:** Critical
**Location:** profile/user_profile.py, line 25
**Recommendation:** Add regex validation for email format
```

### 5. Test Coverage Review
- Are all requirements tested?
- Are error cases covered?
- Test quality assessment

### 6. Security Review
- Password/data handling
- Input validation
- Potential vulnerabilities

### 7. Recommendations
Prioritized list of recommended changes

### 8. Approval Decision
Final decision with justification""",

            "Tester": f"""## Content Guidance

Create a comprehensive Test Results Document including:

### 1. Test Execution Summary
- Total tests run
- Tests passed / failed
- Overall status (PASS / FAIL)

### 2. Requirements Validation

Test each functional requirement:

**FR-1: User can create profile**
- Test Case: Create profile with valid data
- Status: PASS / FAIL
- Evidence: (what was tested, results)

**FR-2: User can update profile**
- Test Case: Update existing profile
- Status: PASS / FAIL
- Evidence: (what was tested, results)

(Continue for all FRs)

### 3. Test Execution Details

For each test file (tests/test_profile.py):
```
Test: test_create_profile_success
Status: PASS
Duration: 0.01s
Output: Profile created successfully with id='123'

Test: test_create_profile_duplicate_username
Status: PASS
Duration: 0.01s
Output: Correctly raised DuplicateUsernameError
```

### 4. Edge Cases & Error Handling
Test results for error scenarios:
- Invalid email format
- Duplicate username
- Missing required fields
- Profile not found

### 5. Code Quality Validation
- Python syntax: ✅ No errors
- Type hints: ✅ Complete
- Docstrings: ✅ Present
- Linting: List any warnings

### 6. Integration with Requirements & Architecture
- Requirements coverage: X/Y requirements tested
- Architecture compliance: ✅/❌
- Review findings addressed: ✅/❌

### 7. Defects Found
List any bugs or issues discovered during testing

### 8. Final Verdict
- **Recommendation:** APPROVE FOR RELEASE / NEEDS FIXES / REJECT
- **Justification:** (why?)
- **Confidence Level:** HIGH / MEDIUM / LOW"""
        }

        return guidance.get(role, "## Content Guidance\n\nCreate deliverables according to role responsibilities.\n")


if __name__ == "__main__":
    unittest.main()
