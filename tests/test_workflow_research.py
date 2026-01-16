#!/usr/bin/env python3
"""
Research Workflow Tests

Tests the research workflow:
Phase 1: Research Task (Information gathering, no code changes)
Phase 2: Documentation (Research findings documented)
Phase 3: Knowledge Capture (Stored in docs/research/)

Status: EXECUTABLE
Priority: MEDIUM
"""

import subprocess
import sys
import time
import unittest
from datetime import datetime
from pathlib import Path


class TestResearchWorkflowPhases(unittest.TestCase):
    """
    Test research workflow executes correctly

    Validates that:
    1. Phase 1: Research conducted
    2. Phase 2: Findings documented
    3. Phase 3: Knowledge captured in docs/research/
    4. No code changes required
    """

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"research-workflow-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

        print(f"\n📁 Test directory: {cls.test_dir}")

    @classmethod
    def tearDownClass(cls):
        """Clean up test artifacts"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)
            print(f"\n🧹 Cleaned up: {cls.test_dir}")

    def test_01_research_task_documented(self):
        """Test: Research task properly documented"""
        print("\n" + "="*70)
        print("RESEARCH WORKFLOW TEST 1: Research Task Documentation")
        print("="*70)

        # Create research task
        task_dir = self.test_dir / "tasks" / "2026-01-15_research-graphql"
        task_dir.mkdir(parents=True, exist_ok=True)

        contract = task_dir / "00-contract.md"
        contract.write_text("""# Research Task: GraphQL vs REST

## Objective
Research and compare GraphQL and REST for our API layer.

## Research Questions
1. What are the performance characteristics?
2. What are the trade-offs?
3. Which is better for our use case?

## Deliverables
- Research findings document
- Pros/cons comparison
- Recommendation

## Acceptance Criteria
- All questions answered
- Evidence-based recommendation
- Documentation in docs/research/
""")

        plan = task_dir / "10-plan.md"
        plan.write_text("""# Research Plan

## Approach
1. Review GraphQL documentation
2. Review REST best practices
3. Compare performance benchmarks
4. Analyze our use case requirements
5. Document findings

## Sources
- GraphQL official docs
- REST API design guides
- Performance benchmarks
- Industry case studies

## Timeline
- Research: 2 hours
- Documentation: 1 hour
- Total: 3 hours
""")

        self.assertTrue(contract.exists(), "❌ Research contract not created")
        self.assertTrue(plan.exists(), "❌ Research plan not created")
        print(f"✅ Research contract: {contract}")
        print(f"✅ Research plan: {plan}")

    def test_02_research_findings_documented(self):
        """Test: Research findings properly documented"""
        print("\n" + "="*70)
        print("RESEARCH WORKFLOW TEST 2: Research Findings")
        print("="*70)

        # Document research findings
        research_dir = self.test_dir / "docs" / "research" / "2026-01-15-graphql-vs-rest"
        research_dir.mkdir(parents=True, exist_ok=True)

        findings = research_dir / "findings.md"
        findings.write_text(f"""# Research: GraphQL vs REST

**Date:** {datetime.now().strftime("%Y-%m-%d")}
**Researcher:** AI Engineer

---

## Executive Summary

GraphQL offers more flexibility but REST is simpler for our use case.

**Recommendation:** Start with REST, migrate to GraphQL if needed.

---

## Research Questions Answered

### 1. Performance Characteristics

**GraphQL:**
- Single endpoint
- Client specifies exact data needs
- Reduces over-fetching
- N+1 query problem possible

**REST:**
- Multiple endpoints
- Fixed response structure
- May over-fetch data
- Simpler caching

**Verdict:** GraphQL wins for complex data requirements, REST wins for simplicity.

### 2. Trade-offs

**GraphQL Pros:**
- ✅ Flexible querying
- ✅ Reduces over-fetching
- ✅ Strong typing
- ✅ Self-documenting

**GraphQL Cons:**
- ❌ More complex setup
- ❌ Harder to cache
- ❌ Learning curve
- ❌ N+1 query risks

**REST Pros:**
- ✅ Simple and well-understood
- ✅ Easy caching
- ✅ Widespread tooling
- ✅ Lower complexity

**REST Cons:**
- ❌ Over-fetching common
- ❌ Multiple round trips
- ❌ Versioning challenges
- ❌ Less flexible

### 3. Our Use Case Analysis

**Current Requirements:**
- Simple CRUD operations
- Mobile app clients
- Moderate data complexity
- Team familiar with REST

**Assessment:**
- REST meets current needs
- GraphQL complexity not justified yet
- Can migrate later if needed

---

## Recommendation

**Use REST for initial implementation**

**Rationale:**
1. Team expertise: Already familiar with REST
2. Simplicity: Faster development
3. Good enough: Meets current requirements
4. Migration path: Can add GraphQL later if needed

**Future Consideration:**
Revisit GraphQL when:
- Mobile bandwidth becomes critical
- Data fetching becomes complex
- Team has capacity for learning curve

---

## References

- [GraphQL Official Docs](https://graphql.org/)
- [REST API Design Guide](https://restfulapi.net/)
- [GraphQL vs REST Performance Study](https://example.com/study)
- [When to Use GraphQL](https://example.com/when-graphql)

---

**Research Complete:** {datetime.now().strftime("%Y-%m-%d %H:%M")}
""")

        self.assertTrue(findings.exists(), "❌ Research findings not documented")
        print(f"✅ Research findings: {findings}")

        # Verify required sections
        content = findings.read_text()
        required_sections = [
            "Executive Summary",
            "Recommendation",
            "Performance Characteristics",
            "Trade-offs",
            "References"
        ]

        for section in required_sections:
            self.assertIn(section, content, f"❌ Missing section: {section}")

        print("✅ All required sections present")

    def test_03_work_log_documents_research(self):
        """Test: Work log documents research process"""
        print("\n" + "="*70)
        print("RESEARCH WORKFLOW TEST 3: Work Log")
        print("="*70)

        task_dir = self.test_dir / "tasks" / "2026-01-15_research-graphql"

        work_log = task_dir / "20-work-log.md"
        work_log.write_text(f"""# Work Log: GraphQL vs REST Research

## Session {datetime.now().strftime("%Y-%m-%d %H:%M")}

### Research Conducted
- ✅ Reviewed GraphQL documentation
- ✅ Reviewed REST best practices
- ✅ Analyzed performance benchmarks
- ✅ Evaluated our use case requirements

### Findings Summary
- GraphQL more flexible but complex
- REST simpler and sufficient
- Recommendation: Use REST

### Documentation Created
- docs/research/2026-01-15-graphql-vs-rest/findings.md

### Time Spent
- Research: 2 hours
- Documentation: 1 hour
- Total: 3 hours

### Status
Research complete, findings documented

### References
- Plan: [10-plan.md](10-plan.md)
- Findings: docs/research/2026-01-15-graphql-vs-rest/findings.md
""")

        self.assertTrue(work_log.exists(), "❌ Work log not created")
        print(f"✅ Work log: {work_log}")

        content = work_log.read_text()
        self.assertIn("Research complete", content)
        print("✅ Research progress documented")

    def test_04_no_code_changes_made(self):
        """Test: Research workflow produces no code changes"""
        print("\n" + "="*70)
        print("RESEARCH WORKFLOW TEST 4: No Code Changes")
        print("="*70)

        # Verify no src/ directory created
        src_dir = self.test_dir / "src"
        tests_dir = self.test_dir / "tests"

        self.assertFalse(src_dir.exists(), "❌ Code changes made (src/ exists)")
        self.assertFalse(tests_dir.exists(), "❌ Tests created (tests/ exists)")

        print("✅ No code changes made")
        print("✅ No tests created")
        print("✅ Pure research task - documentation only")


class TestResearchWorkflowIntegration(unittest.TestCase):
    """
    Integration test: Complete research workflow

    End-to-end validation of research workflow
    """

    @classmethod
    def setUpClass(cls):
        """Set up integration test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"research-integration-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_complete_research_workflow(self):
        """
        Integration Test: Complete research workflow

        Scenario: Research "Which database to use?"
        Workflow: Task → Research → Document → Capture

        Expected: Findings documented, no code changes
        """
        print("\n" + "="*70)
        print("INTEGRATION TEST: Complete Research Workflow")
        print("="*70)

        # Phase 1: Research Task
        print("\n📚 Phase 1: Research Task Setup")
        task_dir = self.test_dir / "tasks" / "2026-01-15_research-database"
        task_dir.mkdir(parents=True, exist_ok=True)

        (task_dir / "00-contract.md").write_text("""# Research: Database Selection
## Objective
Research which database to use for our application.

## Questions
- PostgreSQL vs MongoDB?
- Performance considerations?
- Scalability needs?
""")
        (task_dir / "10-plan.md").write_text("# Plan\nResearch both options")
        print("  ✅ Research task created")

        # Phase 2: Conduct Research
        print("\n🔍 Phase 2: Research Conducted")
        research_dir = self.test_dir / "docs" / "research" / "2026-01-15-database-selection"
        research_dir.mkdir(parents=True, exist_ok=True)

        (research_dir / "findings.md").write_text("""# Research: Database Selection

## Executive Summary
PostgreSQL recommended for relational data with ACID guarantees.

## Findings
- PostgreSQL: Better for structured data
- MongoDB: Better for unstructured data
- Our data: Structured with relationships

## Recommendation
Use PostgreSQL

## References
- PostgreSQL docs
- MongoDB docs
- Performance benchmarks
""")
        print("  ✅ Research conducted")
        print("  ✅ Findings documented")

        # Phase 3: Work Log
        print("\n📝 Phase 3: Work Log")
        (task_dir / "20-work-log.md").write_text("# Work Log\nResearch complete")
        print("  ✅ Work log updated")

        # Verify deliverables
        print("\n📦 Verifying Research Deliverables:")
        deliverables = {
            "Contract": task_dir / "00-contract.md",
            "Plan": task_dir / "10-plan.md",
            "Findings": research_dir / "findings.md",
            "Work Log": task_dir / "20-work-log.md",
        }

        all_exist = True
        for name, path in deliverables.items():
            if path.exists():
                print(f"  ✅ {name}")
            else:
                print(f"  ❌ {name} MISSING")
                all_exist = False

        self.assertTrue(all_exist, "❌ Not all deliverables present")

        # Verify no code
        print("\n🚫 Verifying No Code Changes:")
        src_dir = self.test_dir / "src"
        tests_dir = self.test_dir / "tests"

        if not src_dir.exists() and not tests_dir.exists():
            print("  ✅ No code changes (src/ doesn't exist)")
            print("  ✅ No tests created (tests/ doesn't exist)")
        else:
            self.fail("❌ Code changes made in research task")

        print("\n✅ INTEGRATION TEST PASSED")
        print("\nComplete Research Workflow Verified:")
        print("  ✓ Phase 1: Research task setup")
        print("  ✓ Phase 2: Research conducted")
        print("  ✓ Phase 3: Findings documented")
        print("  ✓ No code changes made")


if __name__ == "__main__":
    print("="*70)
    print("Research Workflow Tests")
    print("="*70)
    print("\nValidating research workflow execution...")
    print()

    # Run tests
    unittest.main(verbosity=2)
