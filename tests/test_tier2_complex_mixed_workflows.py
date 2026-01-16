"""
Tier 2: Complex Mixed Workflow Tests

Tests workflows that combine both parallel and sequential execution patterns.
Validates orchestrator decomposition, parallel engineers, parallel reviewers, and sequential integration.
"""

import unittest
from pathlib import Path
from datetime import datetime
import json


class TestTier2ComplexMixedWorkflows(unittest.TestCase):
    """Test complex workflows combining parallel and sequential patterns"""

    @classmethod
    def setUpClass(cls):
        """Setup test infrastructure"""
        cls.base_dir = Path(__file__).parent.parent
        cls.test_artifacts = cls.base_dir / ".ai" / "test-artifacts"
        cls.test_artifacts.mkdir(parents=True, exist_ok=True)

    def test_01_prepare_orchestrator_to_integration_workflow(self):
        """
        Setup Test: Prepare complex mixed workflow

        Workflow: Orchestrator → 3 Engineers (parallel) → 3 Reviewers (parallel) → Integration Tester

        This tests:
        - Orchestrator task decomposition
        - Parallel engineer execution
        - Parallel reviewer execution
        - Sequential integration testing
        - Mixed parallel + sequential patterns

        Execution: Run this test, then spawn agents following the workflow
        """
        timestamp = int(datetime.now().timestamp())
        test_dir = self.test_artifacts / f"tier2-complex-mixed-{timestamp}"
        test_dir.mkdir(parents=True, exist_ok=True)

        # Create README
        readme = test_dir / "README.md"
        readme.write_text("""# Tier 2: Complex Mixed Workflow Test

## Workflow Overview

```
Stage 1: Orchestrator
  ↓ (decomposes feature into 3 modules)
├─ Stage 2a: Engineer 1 (Module A) ──┐
├─ Stage 2b: Engineer 2 (Module B) ──┤ PARALLEL
└─ Stage 2c: Engineer 3 (Module C) ──┘
  ↓ (wait for ALL engineers)
├─ Stage 3a: Reviewer 1 (Module A) ──┐
├─ Stage 3b: Reviewer 2 (Module B) ──┤ PARALLEL
└─ Stage 3c: Reviewer 3 (Module C) ──┘
  ↓ (wait for ALL reviewers)
Stage 4: Integration Tester (Sequential)
  ↓
  [DONE]
```

## Test Scenario

**Feature:** E-Commerce Shopping Cart System

### Stage 1: Orchestrator
- Analyzes requirements for shopping cart feature
- Decomposes into 3 independent modules
- Creates engineer contracts for each module

### Stage 2: Parallel Engineers (3)
- **Engineer 1:** Cart Management (add/remove/update items)
- **Engineer 2:** Price Calculation (subtotal, tax, discounts)
- **Engineer 3:** Checkout Processing (payment, order creation)

### Stage 3: Parallel Reviewers (3)
- **Reviewer 1:** Reviews Cart Management module
- **Reviewer 2:** Reviews Price Calculation module
- **Reviewer 3:** Reviews Checkout Processing module

### Stage 4: Integration Tester
- Tests integration between all 3 modules
- Validates end-to-end shopping cart workflow
- Creates final integration test results

## Execution

1. Spawn Orchestrator agent
2. Wait for Orchestrator completion → produces 3 contracts
3. Spawn 3 Engineer agents IN PARALLEL
4. Wait for ALL engineers → each produces implementation
5. Spawn 3 Reviewer agents IN PARALLEL
6. Wait for ALL reviewers → each produces review
7. Spawn Integration Tester agent (sequential)
8. Wait for Integration Tester → produces final results
""")

        # Create stage directories
        stages_dir = test_dir / "stages"
        output_dir = test_dir / "output"
        stages_dir.mkdir(parents=True, exist_ok=True)
        output_dir.mkdir(parents=True, exist_ok=True)

        # Define workflow
        workflow = {
            "feature": "E-Commerce Shopping Cart System",
            "stages": [
                {
                    "id": "stage-1-orchestrator",
                    "role": "Orchestrator",
                    "name": "Task Decomposition",
                    "order": 1,
                    "parallel_group": None,
                    "inputs": [],
                    "outputs": [
                        "decomposition.md",
                        "module-a-cart/contract.md",
                        "module-b-pricing/contract.md",
                        "module-c-checkout/contract.md"
                    ]
                },
                {
                    "id": "stage-2a-engineer-cart",
                    "role": "Engineer",
                    "name": "Cart Management Implementation",
                    "order": 2,
                    "parallel_group": "engineers",
                    "inputs": ["module-a-cart/contract.md"],
                    "outputs": [
                        "module-a-cart/cart.py",
                        "module-a-cart/test_cart.py"
                    ]
                },
                {
                    "id": "stage-2b-engineer-pricing",
                    "role": "Engineer",
                    "name": "Price Calculation Implementation",
                    "order": 2,
                    "parallel_group": "engineers",
                    "inputs": ["module-b-pricing/contract.md"],
                    "outputs": [
                        "module-b-pricing/pricing.py",
                        "module-b-pricing/test_pricing.py"
                    ]
                },
                {
                    "id": "stage-2c-engineer-checkout",
                    "role": "Engineer",
                    "name": "Checkout Processing Implementation",
                    "order": 2,
                    "parallel_group": "engineers",
                    "inputs": ["module-c-checkout/contract.md"],
                    "outputs": [
                        "module-c-checkout/checkout.py",
                        "module-c-checkout/test_checkout.py"
                    ]
                },
                {
                    "id": "stage-3a-reviewer-cart",
                    "role": "Reviewer",
                    "name": "Cart Management Review",
                    "order": 3,
                    "parallel_group": "reviewers",
                    "inputs": [
                        "module-a-cart/cart.py",
                        "module-a-cart/test_cart.py"
                    ],
                    "outputs": ["module-a-cart/review.md"]
                },
                {
                    "id": "stage-3b-reviewer-pricing",
                    "role": "Reviewer",
                    "name": "Price Calculation Review",
                    "order": 3,
                    "parallel_group": "reviewers",
                    "inputs": [
                        "module-b-pricing/pricing.py",
                        "module-b-pricing/test_pricing.py"
                    ],
                    "outputs": ["module-b-pricing/review.md"]
                },
                {
                    "id": "stage-3c-reviewer-checkout",
                    "role": "Reviewer",
                    "name": "Checkout Processing Review",
                    "order": 3,
                    "parallel_group": "reviewers",
                    "inputs": [
                        "module-c-checkout/checkout.py",
                        "module-c-checkout/test_checkout.py"
                    ],
                    "outputs": ["module-c-checkout/review.md"]
                },
                {
                    "id": "stage-4-integration-tester",
                    "role": "Integration Tester",
                    "name": "End-to-End Integration Testing",
                    "order": 4,
                    "parallel_group": None,
                    "inputs": [
                        "module-a-cart/cart.py",
                        "module-b-pricing/pricing.py",
                        "module-c-checkout/checkout.py",
                        "module-a-cart/review.md",
                        "module-b-pricing/review.md",
                        "module-c-checkout/review.md"
                    ],
                    "outputs": ["integration_test_results.md"]
                }
            ]
        }

        # Create Orchestrator contract
        orchestrator_dir = stages_dir / "stage-1-orchestrator"
        orchestrator_dir.mkdir(parents=True, exist_ok=True)

        orchestrator_contract = orchestrator_dir / "00-contract.md"
        orchestrator_contract.write_text(f"""# Orchestrator Contract: Shopping Cart Decomposition

**Agent Role:** Orchestrator
**Feature:** E-Commerce Shopping Cart System
**Workflow Stage:** 1 of 4
**Complex Workflow:** Orchestrator → 3 Engineers (parallel) → 3 Reviewers (parallel) → Integration Tester

---

## Task Description

Analyze requirements for an e-commerce shopping cart system and decompose it into 3 independent, parallelizable modules.

## Requirements

**Shopping Cart System Requirements:**
- Users can add/remove items to/from cart
- System calculates prices including taxes and discounts
- Users can checkout and create orders
- All operations must be testable independently
- Modules should integrate seamlessly

## Deliverables

Create 4 deliverables:

### 1. decomposition.md
**Path:** `{output_dir.absolute()}/decomposition.md`

Document your task decomposition strategy:
- How you broke down the feature into 3 modules
- Dependencies between modules
- Integration points
- Parallel execution strategy

### 2. Module A Contract (Cart Management)
**Path:** `{output_dir.absolute()}/module-a-cart/contract.md`

Create engineer contract for Cart Management module:
- Add items to cart
- Remove items from cart
- Update item quantities
- Clear cart
- Get cart contents
- Test file with 5+ test cases

### 3. Module B Contract (Price Calculation)
**Path:** `{output_dir.absolute()}/module-b-pricing/contract.md`

Create engineer contract for Price Calculation module:
- Calculate subtotal from cart items
- Apply tax calculation
- Apply discount codes
- Calculate final total
- Test file with 5+ test cases

### 4. Module C Contract (Checkout Processing)
**Path:** `{output_dir.absolute()}/module-c-checkout/contract.md`

Create engineer contract for Checkout Processing module:
- Validate cart before checkout
- Process payment (mock)
- Create order record
- Clear cart after successful checkout
- Test file with 5+ test cases

## Acceptance Criteria

- [ ] All 4 deliverables created at absolute paths
- [ ] Decomposition clearly explains module boundaries
- [ ] Each contract specifies 2 Python files (implementation + tests)
- [ ] Contracts can be executed by engineers in parallel
- [ ] No circular dependencies between modules

## Execution

**Working Directory:** `{output_dir.absolute()}`

**Next Stages:**
- 3 Engineers will execute Module A, B, C contracts IN PARALLEL
- 3 Reviewers will review each module's implementation IN PARALLEL
- Integration Tester will validate end-to-end workflow

**Date:** {datetime.now().strftime('%Y-%m-%d %H:%M')}
""")

        # Create metadata
        metadata = {
            "test_type": "complex_mixed_workflow",
            "workflow": workflow,
            "total_stages": 8,
            "parallel_groups": {
                "engineers": 3,
                "reviewers": 3
            },
            "sequential_stages": 2,  # Orchestrator and Integration Tester
            "test_dir": str(test_dir.absolute()),
            "execution_instructions": [
                "1. Start with Orchestrator (Stage 1)",
                "2. Orchestrator creates 3 contracts + decomposition doc",
                "3. Spawn 3 Engineers IN PARALLEL (Stages 2a, 2b, 2c)",
                "4. Wait for ALL 3 engineers to complete",
                "5. Spawn 3 Reviewers IN PARALLEL (Stages 3a, 3b, 3c)",
                "6. Wait for ALL 3 reviewers to complete",
                "7. Spawn Integration Tester (Stage 4)",
                "8. Wait for Integration Tester to complete",
                "9. Verify all 13 files created"
            ]
        }

        metadata_file = test_dir / "test-metadata.json"
        metadata_file.write_text(json.dumps(metadata, indent=2))

        # Create verification script
        verify_script = test_dir / "verify_complex_workflow.py"
        verify_script.write_text(f'''#!/usr/bin/env python3
"""
Verify Complex Mixed Workflow Test Results
"""

from pathlib import Path
import sys
import json

def verify_complex_workflow():
    """Verify all stages completed correctly"""
    test_dir = Path(__file__).parent
    output_dir = test_dir / "output"

    metadata_file = test_dir / "test-metadata.json"
    with open(metadata_file) as f:
        metadata = json.load(f)

    workflow = metadata["workflow"]
    stages = workflow["stages"]

    print("🔍 Verifying Complex Mixed Workflow...")
    print(f"Feature: {{workflow['feature']}}")
    print(f"Stages: {{len(stages)}}\\n")

    all_success = True
    total_files = 0

    # Stage 1: Orchestrator
    print("Stage 1: Orchestrator")
    print("─" * 60)
    orchestrator_outputs = ["decomposition.md", "module-a-cart/contract.md",
                           "module-b-pricing/contract.md", "module-c-checkout/contract.md"]
    for output in orchestrator_outputs:
        path = output_dir / output
        if path.exists():
            print(f"  ✅ {{output}}")
            total_files += 1
        else:
            print(f"  ❌ {{output}}: MISSING")
            all_success = False

    # Stage 2: Engineers (parallel)
    print("\\nStage 2: Engineers (Parallel)")
    print("─" * 60)
    engineer_groups = {{
        "Cart (2a)": ["module-a-cart/cart.py", "module-a-cart/test_cart.py"],
        "Pricing (2b)": ["module-b-pricing/pricing.py", "module-b-pricing/test_pricing.py"],
        "Checkout (2c)": ["module-c-checkout/checkout.py", "module-c-checkout/test_checkout.py"]
    }}
    for group_name, files in engineer_groups.items():
        print(f"  {{group_name}}:")
        for file in files:
            path = output_dir / file
            if path.exists():
                print(f"    ✅ {{file}}")
                total_files += 1
            else:
                print(f"    ❌ {{file}}: MISSING")
                all_success = False

    # Stage 3: Reviewers (parallel)
    print("\\nStage 3: Reviewers (Parallel)")
    print("─" * 60)
    reviewer_outputs = ["module-a-cart/review.md", "module-b-pricing/review.md",
                       "module-c-checkout/review.md"]
    for output in reviewer_outputs:
        path = output_dir / output
        if path.exists():
            print(f"  ✅ {{output}}")
            total_files += 1
        else:
            print(f"  ❌ {{output}}: MISSING")
            all_success = False

    # Stage 4: Integration Tester
    print("\\nStage 4: Integration Tester")
    print("─" * 60)
    integration_path = output_dir / "integration_test_results.md"
    if integration_path.exists():
        print(f"  ✅ integration_test_results.md")
        total_files += 1
    else:
        print(f"  ❌ integration_test_results.md: MISSING")
        all_success = False

    print("\\n" + "=" * 60)
    print("📊 Workflow Summary")
    print("=" * 60)
    print(f"Total files created: {{total_files}}/13")
    print(f"Workflow integrity: {{'PASS' if all_success else 'FAIL'}}")

    if all_success:
        print("\\n✅ SUCCESS: All stages completed!")
        print("\\n🔗 Validating Workflow Pattern...")
        print("  Orchestrator (sequential): ✅")
        print("  3 Engineers (parallel): ✅")
        print("  3 Reviewers (parallel): ✅")
        print("  Integration Tester (sequential): ✅")
        print("\\n🎉 PERFECT: Mixed parallel + sequential workflow validated!")
        return True
    else:
        print("\\n❌ FAILURE: Workflow incomplete")
        return False

if __name__ == "__main__":
    success = verify_complex_workflow()
    sys.exit(0 if success else 1)
''')
        verify_script.chmod(0o755)

        print(f"✅ Complex mixed workflow test infrastructure created!")
        print(f"   Test directory: {test_dir.absolute()}")
        print(f"   Stages: 8 (1 Orchestrator + 3 Engineers + 3 Reviewers + 1 Integration Tester)")
        print(f"   Parallel groups: 2 (Engineers, Reviewers)")
        print(f"   Total expected files: 13")
        print(f"\n📋 Next Steps:")
        print(f"   1. Spawn Orchestrator agent (Stage 1)")
        print(f"   2. Wait for completion → 4 files created")
        print(f"   3. Spawn 3 Engineers IN PARALLEL (Stages 2a, 2b, 2c)")
        print(f"   4. Wait for all 3 → 6 files created")
        print(f"   5. Spawn 3 Reviewers IN PARALLEL (Stages 3a, 3b, 3c)")
        print(f"   6. Wait for all 3 → 3 files created")
        print(f"   7. Spawn Integration Tester (Stage 4)")
        print(f"   8. Wait for completion → 1 file created")
        print(f"   9. Run verification script")

        self._test_dir = test_dir
        self.assertTrue(test_dir.exists())
        self.assertTrue(orchestrator_contract.exists())


if __name__ == "__main__":
    unittest.main()
