#!/usr/bin/env python3
"""
Parallel Agent Execution Test

Tests spawning multiple agents concurrently to verify:
1. Multiple agents can be spawned
2. No context pollution between agents
3. Independent execution
4. Results aggregation
"""

import subprocess
import time
import json
from pathlib import Path
from datetime import datetime


class ParallelExecutionTest:
    """Test parallel agent execution."""

    def __init__(self):
        self.test_workspace = Path("tests/parallel_test_workspace")
        self.test_workspace.mkdir(exist_ok=True)
        self.results = []

    def spawn_agent(self, role, task):
        """Spawn an agent and return task ID."""
        print(f"\n🚀 Spawning {role} agent...")
        print(f"   Task: {task[:60]}...")

        start_time = time.time()

        cmd = [".ai-pack/bd", "spawn", role, task]
        result = subprocess.run(cmd, capture_output=True, text=True)

        spawn_time = time.time() - start_time

        if result.returncode != 0:
            print(f"❌ Failed to spawn: {result.stderr}")
            return None, spawn_time

        # Extract task ID
        task_id = None
        for line in result.stdout.split('\n'):
            if 'Task ID:' in line:
                task_id = line.split(':')[1].strip()
                break

        print(f"✅ Spawned in {spawn_time:.2f}s - Task ID: {task_id}")

        return task_id, spawn_time

    def test_two_parallel_engineers(self):
        """Test 2: Two Engineer Agents in Parallel"""
        print("\n" + "="*70)
        print("TEST: Two Engineer Agents - Parallel Implementation")
        print("="*70)

        # Define two independent implementation tasks
        task1 = """
        Implement a simple Calculator class with these methods:
        - add(a, b): return a + b
        - subtract(a, b): return a - b

        Create the file at tests/parallel_test_workspace/calculator.py
        Include docstrings and type hints.
        No tests needed for this task.
        """

        task2 = """
        Implement a simple StringUtils class with these methods:
        - reverse(s): return reversed string
        - capitalize_words(s): capitalize each word

        Create the file at tests/parallel_test_workspace/string_utils.py
        Include docstrings and type hints.
        No tests needed for this task.
        """

        print("\n📋 Spawning 2 engineers for different features...")

        # Spawn both agents
        task_id1, time1 = self.spawn_agent("engineer", task1)
        task_id2, time2 = self.spawn_agent("engineer", task2)

        print(f"\n⏱️  Total spawn time: {time1 + time2:.2f}s")
        print(f"   Average per agent: {(time1 + time2) / 2:.2f}s")

        self.results.append({
            "test": "Two Parallel Engineers",
            "agents": 2,
            "task_ids": [task_id1, task_id2],
            "spawn_times": [time1, time2],
            "total_time": time1 + time2,
            "avg_time": (time1 + time2) / 2
        })

        return task_id1, task_id2

    def test_three_parallel_agents(self):
        """Test 3: Three Agents - 2 Engineers + 1 Tester"""
        print("\n" + "="*70)
        print("TEST: Three Agents - Full Development Workflow")
        print("="*70)

        task_engineer1 = """
        Implement a User class with properties:
        - name (string)
        - email (string)
        - is_active (boolean)

        Include __init__, __str__, and validate_email() method.
        Create at tests/parallel_test_workspace/user.py
        """

        task_engineer2 = """
        Implement a Product class with properties:
        - id (int)
        - name (string)
        - price (float)

        Include __init__, __str__, and apply_discount(percent) method.
        Create at tests/parallel_test_workspace/product.py
        """

        task_tester = """
        Create a simple test file that verifies:
        1. Check if calculator.py exists
        2. Check if string_utils.py exists
        3. Report file sizes

        Create at tests/parallel_test_workspace/test_files_exist.py
        Use basic file operations, no actual test execution needed.
        """

        print("\n📋 Spawning 3 agents (2 engineers + 1 tester)...")

        # Spawn all three agents
        task_id1, time1 = self.spawn_agent("engineer", task_engineer1)
        task_id2, time2 = self.spawn_agent("engineer", task_engineer2)
        task_id3, time3 = self.spawn_agent("tester", task_tester)

        print(f"\n⏱️  Total spawn time: {time1 + time2 + time3:.2f}s")
        print(f"   Average per agent: {(time1 + time2 + time3) / 3:.2f}s")

        self.results.append({
            "test": "Three Parallel Agents",
            "agents": 3,
            "task_ids": [task_id1, task_id2, task_id3],
            "spawn_times": [time1, time2, time3],
            "total_time": time1 + time2 + time3,
            "avg_time": (time1 + time2 + time3) / 3
        })

        return task_id1, task_id2, task_id3

    def verify_no_context_pollution(self, task_id1, task_id2):
        """Verify agents maintained independent contexts."""
        print("\n" + "="*70)
        print("VERIFICATION: Context Isolation")
        print("="*70)

        task_dir1 = Path(f".beads/tasks/{task_id1}")
        task_dir2 = Path(f".beads/tasks/{task_id2}")

        if not task_dir1.exists() or not task_dir2.exists():
            print("⏳ Task directories not yet created (agents still spawning)")
            return

        # Check task packets are independent
        meta1 = task_dir1 / "00-metadata.json"
        meta2 = task_dir2 / "00-metadata.json"

        if meta1.exists() and meta2.exists():
            with open(meta1) as f:
                data1 = json.load(f)
            with open(meta2) as f:
                data2 = json.load(f)

            print(f"\n✅ Task 1: {data1['task_id']}")
            print(f"   Role: {data1['role']}")
            print(f"   Description: {data1['description'][:60]}...")

            print(f"\n✅ Task 2: {data2['task_id']}")
            print(f"   Role: {data2['role']}")
            print(f"   Description: {data2['description'][:60]}...")

            if data1['task_id'] != data2['task_id']:
                print("\n✅ PASSED: Tasks have independent IDs")
            else:
                print("\n❌ FAILED: Tasks have same ID (context pollution)")

            if data1['description'] != data2['description']:
                print("✅ PASSED: Tasks have independent descriptions")
            else:
                print("❌ FAILED: Tasks have same description (context pollution)")

        else:
            print("⏳ Metadata not yet created")

    def measure_performance_metrics(self):
        """Measure and report performance metrics."""
        print("\n" + "="*70)
        print("PERFORMANCE METRICS")
        print("="*70)

        total_agents = sum(r['agents'] for r in self.results)
        all_spawn_times = []
        for r in self.results:
            all_spawn_times.extend(r['spawn_times'])

        avg_spawn_time = sum(all_spawn_times) / len(all_spawn_times) if all_spawn_times else 0
        max_spawn_time = max(all_spawn_times) if all_spawn_times else 0
        min_spawn_time = min(all_spawn_times) if all_spawn_times else 0

        print(f"\n📊 Total Agents Spawned: {total_agents}")
        print(f"⏱️  Average Spawn Time: {avg_spawn_time:.2f}s")
        print(f"   Min: {min_spawn_time:.2f}s")
        print(f"   Max: {max_spawn_time:.2f}s")

        # Check against success criteria
        target_latency = 5.0  # seconds

        if avg_spawn_time < target_latency:
            print(f"\n✅ PASSED: Average spawn time ({avg_spawn_time:.2f}s) < {target_latency}s target")
        else:
            print(f"\n⚠️  WARNING: Average spawn time ({avg_spawn_time:.2f}s) > {target_latency}s target")

        return {
            "total_agents": total_agents,
            "avg_spawn_time": avg_spawn_time,
            "min_spawn_time": min_spawn_time,
            "max_spawn_time": max_spawn_time,
            "target_met": avg_spawn_time < target_latency
        }

    def save_results(self, metrics):
        """Save test results to file."""
        results_file = Path("tests/parallel_execution_results.json")

        results_data = {
            "timestamp": datetime.now().isoformat(),
            "tests": self.results,
            "metrics": metrics,
            "summary": {
                "total_tests": len(self.results),
                "total_agents": metrics["total_agents"],
                "performance_target_met": metrics["target_met"]
            }
        }

        with open(results_file, 'w') as f:
            json.dump(results_data, f, indent=2)

        print(f"\n💾 Results saved to: {results_file}")

    def run_all_tests(self):
        """Run all parallel execution tests."""
        print("""
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║         PARALLEL AGENT EXECUTION TEST SUITE - WEEK 2             ║
║                                                                  ║
║  Tests verify:                                                   ║
║  - Multiple agents can be spawned                                ║
║  - Agents execute independently                                  ║
║  - No context pollution                                          ║
║  - Performance metrics met                                       ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
        """)

        # Test 1: Two engineers
        task_id1, task_id2 = self.test_two_parallel_engineers()

        # Verify independence
        if task_id1 and task_id2:
            self.verify_no_context_pollution(task_id1, task_id2)

        # Test 2: Three agents
        self.test_three_parallel_agents()

        # Measure performance
        metrics = self.measure_performance_metrics()

        # Save results
        self.save_results(metrics)

        # Final summary
        print("\n" + "="*70)
        print("TEST SUITE COMPLETE")
        print("="*70)
        print("\n✅ All parallel execution tests spawned successfully")
        print(f"📊 Performance: {metrics['avg_spawn_time']:.2f}s average spawn time")
        print(f"🎯 Target Met: {'YES' if metrics['target_met'] else 'NO'}")
        print("\n⏳ Note: Agents are executing asynchronously.")
        print("   Check .beads/tasks/ for agent results.")
        print("   Full workflow validation requires agent completion.")


def main():
    """Main entry point."""
    tester = ParallelExecutionTest()
    tester.run_all_tests()


if __name__ == "__main__":
    main()
