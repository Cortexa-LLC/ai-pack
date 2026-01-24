"""
Tier 2: Parallel Agent Execution Tests

Tests parallel execution of multiple spawned agents using real Claude Code
Task tool spawning. These are SETUP tests that prepare contracts and spawn agents.
"""

import unittest
from pathlib import Path
from datetime import datetime
import json


class TestTier2ParallelExecution(unittest.TestCase):
    """Test parallel execution of multiple spawned agents"""

    @classmethod
    def setUpClass(cls):
        """Setup test infrastructure"""
        cls.base_dir = Path(__file__).parent.parent
        cls.test_artifacts = cls.base_dir / ".ai" / "test-artifacts"
        cls.test_artifacts.mkdir(parents=True, exist_ok=True)

    def test_01_prepare_parallel_5_engineers_task(self):
        """
        Setup Test: Prepare 5 parallel Engineer tasks

        This test creates contracts for 5 independent features that will be
        developed simultaneously by 5 different Engineer agents.

        Execution: Run this test, then spawn agents using Claude Code Task tool
        """
        # Create test directory with timestamp
        timestamp = int(datetime.now().timestamp())
        test_dir = self.test_artifacts / f"multi-agent-parallel-{timestamp}"
        test_dir.mkdir(parents=True, exist_ok=True)

        # Create README
        readme = test_dir / "README.md"
        readme.write_text("""# Tier 2: Parallel Agent Execution Tests

## Test Scenarios

### Test 1: 5 Parallel Engineers
- **Objective:** Validate concurrent execution without conflicts
- **Agents:** 5 Engineers working on different features
- **Isolation:** Directory-based (each feature in separate subdirectory)
- **Total Files:** 15 (3 per feature)

### Test 2: Sequential Workflow Chain (Future)
- **Objective:** Validate task handoffs through workflow stages
- **Flow:** PRD → Architect → Engineer → Reviewer → Tester

### Test 3: Complex Mixed Flow (Future)
- **Objective:** Combine parallel and sequential execution
- **Flow:** Orchestrator → 3 Engineers (parallel) → 3 Reviewers (parallel) → Tester
""")

        # Define 5 features with contracts
        features = [
            {
                "id": "feature-1-auth",
                "name": "User Authentication System",
                "files": [
                    {
                        "path": "auth/authenticator.py",
                        "description": "User authentication with SHA-256 hashing",
                        "lines": 75
                    },
                    {
                        "path": "auth/session_manager.py",
                        "description": "Session lifecycle management",
                        "lines": 45
                    },
                    {
                        "path": "tests/test_auth.py",
                        "description": "Authentication tests",
                        "lines": 40
                    }
                ]
            },
            {
                "id": "feature-2-api",
                "name": "REST API Endpoints",
                "files": [
                    {
                        "path": "api/product_controller.py",
                        "description": "CRUD endpoints for products",
                        "lines": 86
                    },
                    {
                        "path": "api/response_formatter.py",
                        "description": "API response formatting",
                        "lines": 42
                    },
                    {
                        "path": "tests/test_api.py",
                        "description": "API endpoint tests",
                        "lines": 60
                    }
                ]
            },
            {
                "id": "feature-3-cache",
                "name": "In-Memory Caching System",
                "files": [
                    {
                        "path": "cache/memory_cache.py",
                        "description": "TTL-based memory cache",
                        "lines": 83
                    },
                    {
                        "path": "cache/lru_cache.py",
                        "description": "LRU eviction cache",
                        "lines": 55
                    },
                    {
                        "path": "tests/test_cache.py",
                        "description": "Cache system tests",
                        "lines": 60
                    }
                ]
            },
            {
                "id": "feature-4-validator",
                "name": "Input Validation System",
                "files": [
                    {
                        "path": "validation/validators.py",
                        "description": "Validation utility functions",
                        "lines": 81
                    },
                    {
                        "path": "validation/schema_validator.py",
                        "description": "Schema-based validation",
                        "lines": 58
                    },
                    {
                        "path": "tests/test_validation.py",
                        "description": "Validation tests",
                        "lines": 65
                    }
                ]
            },
            {
                "id": "feature-5-logger",
                "name": "Structured Logging System",
                "files": [
                    {
                        "path": "logging/logger.py",
                        "description": "Multi-level structured logger",
                        "lines": 87
                    },
                    {
                        "path": "logging/formatter.py",
                        "description": "Log formatters (text, JSON, compact)",
                        "lines": 50
                    },
                    {
                        "path": "tests/test_logging.py",
                        "description": "Logging system tests",
                        "lines": 75
                    }
                ]
            }
        ]

        # Create task directory and contracts for each feature
        tasks_dir = test_dir / "tasks"
        output_dir = test_dir / "output"

        for idx, feature in enumerate(features, 1):
            feature_task_dir = tasks_dir / feature["id"]
            feature_task_dir.mkdir(parents=True, exist_ok=True)

            feature_output_dir = output_dir / feature["id"]

            # Create contract
            contract = feature_task_dir / "00-contract.md"

            # Build file deliverables section
            files_section = ""
            for file_idx, file_info in enumerate(feature["files"], 1):
                file_path = feature_output_dir / file_info["path"]
                files_section += f"""
### File {file_idx}: {file_info["path"]}
**Path:** `{file_path.absolute()}`
**Description:** {file_info["description"]}
**Approximate Lines:** {file_info["lines"]}

**Note:** Implement complete, production-ready code with proper:
- Error handling
- Type hints
- Docstrings
- Test coverage
"""

            contract_content = f"""# Engineer Contract: {feature["name"]}

**Agent Role:** Engineer
**Feature:** {feature["name"]}
**Parallel Test:** {idx} of 5 concurrent engineers

---

## Task Description

Implement {feature["name"].lower()} as part of a parallel execution test.
You are one of 5 engineers working simultaneously on different features.

## Deliverables

Create {len(feature["files"])} Python files:
{files_section}

## Acceptance Criteria

- [ ] All {len(feature["files"])} files created at absolute paths
- [ ] Files in {feature["id"]} directory
- [ ] Complete content (not truncated)
- [ ] Valid Python syntax
- [ ] No conflicts with other parallel agents

## Execution

**Working Directory:** `{feature_output_dir.absolute()}`

**Parallel Context:** This agent is one of 5 running simultaneously. Work independently in your feature directory.

**Other Features (DO NOT MODIFY):**
{chr(10).join(f'- {f["id"]}' for f in features if f["id"] != feature["id"])}

**Date:** {datetime.now().strftime('%Y-%m-%d %H:%M')}
"""
            contract.write_text(contract_content)

        # Create metadata file
        metadata = {
            "test_type": "parallel_execution",
            "agents": 5,
            "files_per_agent": 3,
            "total_files": 15,
            "test_dir": str(test_dir.absolute()),
            "features": [f["id"] for f in features],
            "execution_instructions": [
                "1. Review contracts in tasks/feature-*-*/00-contract.md",
                "2. Spawn 5 agents simultaneously using Claude Code Task tool",
                "3. Each agent reads its contract and creates 3 files",
                "4. Monitor agent completion via task notifications",
                "5. Verify all 15 files exist in output/ directory",
                "6. Run verification script to validate deliverables"
            ]
        }

        metadata_file = test_dir / "test-metadata.json"
        metadata_file.write_text(json.dumps(metadata, indent=2))

        # Create verification script
        verify_script = test_dir / "verify_parallel_execution.py"
        verify_script.write_text(f'''#!/usr/bin/env python3
"""
Verify Parallel Execution Test Results

Run this after all 5 agents complete to validate deliverables.
"""

from pathlib import Path
import sys

def verify_parallel_execution():
    """Verify all files created correctly"""
    test_dir = Path(__file__).parent
    output_dir = test_dir / "output"

    features = {json.dumps([f["id"] for f in features])}

    results = {{
        "total_files": 0,
        "total_lines": 0,
        "features": {{}}
    }}

    for feature_id in features:
        feature_dir = output_dir / feature_id
        if not feature_dir.exists():
            print(f"❌ FAIL: Feature directory missing: {{feature_id}}")
            return False

        py_files = list(feature_dir.rglob("*.py"))
        if len(py_files) != 3:
            print(f"❌ FAIL: Expected 3 files for {{feature_id}}, found {{len(py_files)}}")
            return False

        feature_lines = 0
        for py_file in py_files:
            lines = len(py_file.read_text().splitlines())
            feature_lines += lines
            results["total_lines"] += lines

        results["features"][feature_id] = {{
            "files": len(py_files),
            "lines": feature_lines
        }}
        results["total_files"] += len(py_files)

        print(f"✅ {{feature_id}}: {{len(py_files)}} files, {{feature_lines}} lines")

    print(f"\\n📊 Summary:")
    print(f"   Total files: {{results['total_files']}}")
    print(f"   Total lines: {{results['total_lines']}}")
    print(f"   Features: {{len(results['features'])}}")

    if results["total_files"] == 15:
        print("\\n✅ PASS: All parallel agents completed successfully!")
        return True
    else:
        print(f"\\n❌ FAIL: Expected 15 files, found {{results['total_files']}}")
        return False

if __name__ == "__main__":
    success = verify_parallel_execution()
    sys.exit(0 if success else 1)
''')
        verify_script.chmod(0o755)

        print(f"✅ Test infrastructure created at: {test_dir.absolute()}")
        print(f"   - 5 contracts in tasks/ directory")
        print(f"   - Verification script: verify_parallel_execution.py")
        print(f"   - Test metadata: test-metadata.json")
        print(f"\n📋 Next Steps:")
        print(f"   1. Spawn 5 agents using Claude Code Task tool")
        print(f"   2. Wait for all agents to complete")
        print(f"   3. Run: python {verify_script.name}")

        # Store test dir for potential use by next test
        self._test_dir = test_dir

        self.assertTrue(test_dir.exists())
        self.assertEqual(len(list((tasks_dir).glob("feature-*"))), 5)


if __name__ == "__main__":
    unittest.main()
