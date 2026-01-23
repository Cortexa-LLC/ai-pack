#!/usr/bin/env python3
"""
Integration Tests for Greeting Function

Tests the greeting function in realistic usage scenarios:
- Command-line usage
- Module import scenarios
- Integration with other components
- Real-world data inputs

Status: EXECUTABLE
Priority: HIGH
"""

import unittest
import sys
import os
from io import StringIO

# Add src directory to path
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..', 'src')))

from greeting import greet


class TestGreetingIntegration(unittest.TestCase):
    """Integration tests for greeting function"""

    def test_batch_greeting_multiple_users(self):
        """Test greeting multiple users in sequence"""
        users = ["Alice", "Bob", "Charlie", "Diana"]
        results = [greet(user) for user in users]

        expected = [
            "Hello, Alice!",
            "Hello, Bob!",
            "Hello, Charlie!",
            "Hello, Diana!"
        ]

        self.assertEqual(results, expected)

    def test_greeting_from_file_input(self):
        """Simulate reading names from a file and greeting them"""
        # Simulate file content
        names_input = "John\nJane\nJack\nJill"
        names = names_input.strip().split('\n')

        results = [greet(name) for name in names]

        self.assertEqual(len(results), 4)
        self.assertEqual(results[0], "Hello, John!")
        self.assertEqual(results[-1], "Hello, Jill!")

    def test_greeting_with_user_database(self):
        """Simulate integration with a user database"""
        # Mock user database
        user_db = [
            {"id": 1, "name": "Alice Smith"},
            {"id": 2, "name": "Bob Jones"},
            {"id": 3, "name": "Carol White"}
        ]

        greetings = {user["id"]: greet(user["name"]) for user in user_db}

        self.assertEqual(greetings[1], "Hello, Alice Smith!")
        self.assertEqual(greetings[2], "Hello, Bob Jones!")
        self.assertEqual(greetings[3], "Hello, Carol White!")

    def test_greeting_internationalization(self):
        """Test greeting with international names"""
        international_names = [
            "José García",      # Spanish
            "François Dubois",  # French
            "李明",             # Chinese
            "Müller Schmidt",   # German
            "Андрей Иванов"    # Russian
        ]

        results = [greet(name) for name in international_names]

        # All should be greeted correctly
        self.assertEqual(len(results), 5)
        for result in results:
            self.assertTrue(result.startswith("Hello, "))
            self.assertTrue(result.endswith("!"))

    def test_greeting_output_to_stream(self):
        """Test writing greetings to an output stream"""
        output = StringIO()

        names = ["Alice", "Bob", "Charlie"]
        for name in names:
            greeting = greet(name)
            output.write(greeting + "\n")

        output.seek(0)
        lines = output.readlines()

        self.assertEqual(len(lines), 3)
        self.assertEqual(lines[0].strip(), "Hello, Alice!")
        self.assertEqual(lines[1].strip(), "Hello, Bob!")
        self.assertEqual(lines[2].strip(), "Hello, Charlie!")

    def test_greeting_api_response_format(self):
        """Test greeting function in API response context"""
        # Simulate API endpoint returning greeting
        def api_greet_endpoint(name: str) -> dict:
            return {
                "status": "success",
                "message": greet(name),
                "timestamp": "2026-01-23T12:44:43Z"
            }

        response = api_greet_endpoint("TestUser")

        self.assertEqual(response["status"], "success")
        self.assertEqual(response["message"], "Hello, TestUser!")
        self.assertIn("timestamp", response)

    def test_greeting_error_handling_context(self):
        """Test greeting function maintains stability with edge inputs"""
        # Even with unusual inputs, function should not crash
        edge_cases = [
            "",
            " ",
            "\n",
            "\t",
            "a" * 1000,  # Very long string
            "Name with @#$%",
            "123456789"
        ]

        for test_input in edge_cases:
            try:
                result = greet(test_input)
                self.assertIsInstance(result, str)
                self.assertTrue(result.startswith("Hello, "))
                self.assertTrue(result.endswith("!"))
            except Exception as e:
                self.fail(f"greet() raised {type(e).__name__} for input '{test_input}': {e}")


class TestGreetingPerformance(unittest.TestCase):
    """Performance and stress tests"""

    def test_greeting_performance_bulk_operations(self):
        """Test performance with bulk operations"""
        import time

        # Generate 10000 greetings
        start_time = time.time()

        for i in range(10000):
            greet(f"User{i}")

        end_time = time.time()
        elapsed = end_time - start_time

        # Should complete in reasonable time (< 1 second for 10k operations)
        self.assertLess(elapsed, 1.0, f"Bulk operations took {elapsed:.3f}s, expected < 1.0s")

    def test_greeting_memory_efficiency(self):
        """Test that function doesn't leak memory"""
        results = []

        # Create 1000 greetings
        for i in range(1000):
            result = greet(f"User{i}")
            results.append(result)

        # All results should be unique and correctly formatted
        self.assertEqual(len(results), 1000)
        self.assertEqual(len(set(results)), 1000)  # All unique


if __name__ == '__main__':
    # Run integration tests with verbose output
    unittest.main(verbosity=2)
