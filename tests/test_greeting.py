#!/usr/bin/env python3
"""
Unit Tests for Greeting Function

Following TDD practices:
- Test-first development
- Edge case coverage
- Input validation
- >80% code coverage target

Status: EXECUTABLE
Priority: HIGH
"""

import unittest
import sys
import os

# Add src directory to path
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..', 'src')))

from greeting import greet


class TestGreetingFunction(unittest.TestCase):
    """Test suite for the greet function"""

    def test_greet_simple_name(self):
        """Test greeting with a simple name"""
        result = greet("Alice")
        self.assertEqual(result, "Hello, Alice!")

    def test_greet_name_with_spaces(self):
        """Test greeting with name containing spaces"""
        result = greet("John Doe")
        self.assertEqual(result, "Hello, John Doe!")

    def test_greet_empty_string(self):
        """Test greeting with empty string"""
        result = greet("")
        self.assertEqual(result, "Hello, !")

    def test_greet_single_character(self):
        """Test greeting with single character name"""
        result = greet("A")
        self.assertEqual(result, "Hello, A!")

    def test_greet_name_with_numbers(self):
        """Test greeting with name containing numbers"""
        result = greet("User123")
        self.assertEqual(result, "Hello, User123!")

    def test_greet_name_with_special_chars(self):
        """Test greeting with name containing special characters"""
        result = greet("O'Brien")
        self.assertEqual(result, "Hello, O'Brien!")

    def test_greet_unicode_name(self):
        """Test greeting with unicode characters"""
        result = greet("José")
        self.assertEqual(result, "Hello, José!")

    def test_greet_long_name(self):
        """Test greeting with a very long name"""
        long_name = "A" * 100
        result = greet(long_name)
        self.assertEqual(result, f"Hello, {long_name}!")

    def test_greet_type_validation(self):
        """Test that function accepts string input"""
        # This should work with string input
        result = greet("Test")
        self.assertIsInstance(result, str)

    def test_greet_return_format(self):
        """Test that return value follows expected format"""
        result = greet("World")
        self.assertTrue(result.startswith("Hello, "))
        self.assertTrue(result.endswith("!"))
        self.assertIn("World", result)


class TestGreetingEdgeCases(unittest.TestCase):
    """Test suite for edge cases and boundary conditions"""

    def test_greet_whitespace_name(self):
        """Test greeting with whitespace-only name"""
        result = greet("   ")
        self.assertEqual(result, "Hello,    !")

    def test_greet_newline_in_name(self):
        """Test greeting with newline character"""
        result = greet("Name\nWithNewline")
        self.assertEqual(result, "Hello, Name\nWithNewline!")

    def test_greet_tab_in_name(self):
        """Test greeting with tab character"""
        result = greet("Name\tWithTab")
        self.assertEqual(result, "Hello, Name\tWithTab!")


if __name__ == '__main__':
    # Run tests with verbose output
    unittest.main(verbosity=2)
