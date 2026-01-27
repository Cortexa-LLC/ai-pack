#!/usr/bin/env python3
"""
Tests for hello.py module.

Following TDD approach - tests written first.
"""

import unittest
import sys
from io import StringIO
from pathlib import Path

# Add current directory to path for import
sys.path.insert(0, str(Path(__file__).parent))

import hello


class TestHelloModule(unittest.TestCase):
    """Test cases for hello module."""

    def test_greet_returns_string(self):
        """Test that greet() returns a string."""
        result = hello.greet()
        self.assertIsInstance(result, str)

    def test_greet_default_message(self):
        """Test greet() with no arguments returns default greeting."""
        result = hello.greet()
        self.assertEqual(result, "Hello, World!")

    def test_greet_with_name(self):
        """Test greet() with custom name."""
        result = hello.greet("Alice")
        self.assertEqual(result, "Hello, Alice!")

    def test_greet_with_empty_name(self):
        """Test greet() handles empty string gracefully."""
        result = hello.greet("")
        self.assertEqual(result, "Hello, World!")

    def test_greet_with_none(self):
        """Test greet() handles None gracefully."""
        result = hello.greet(None)
        self.assertEqual(result, "Hello, World!")

    def test_greet_strips_whitespace(self):
        """Test greet() strips leading/trailing whitespace from name."""
        result = hello.greet("  Bob  ")
        self.assertEqual(result, "Hello, Bob!")

    def test_main_prints_greeting(self):
        """Test main() prints greeting to stdout."""
        captured_output = StringIO()
        sys.stdout = captured_output
        
        hello.main()
        
        sys.stdout = sys.__stdout__
        output = captured_output.getvalue()
        self.assertIn("Hello, World!", output)

    def test_module_has_docstring(self):
        """Test module has docstring."""
        self.assertIsNotNone(hello.__doc__)
        self.assertGreater(len(hello.__doc__), 0)

    def test_greet_has_docstring(self):
        """Test greet function has docstring."""
        self.assertIsNotNone(hello.greet.__doc__)
        self.assertGreater(len(hello.greet.__doc__), 0)


class TestHelloEdgeCases(unittest.TestCase):
    """Edge case tests for hello module."""

    def test_greet_with_special_characters(self):
        """Test greet() handles special characters in name."""
        result = hello.greet("José")
        self.assertEqual(result, "Hello, José!")

    def test_greet_with_numbers(self):
        """Test greet() handles names with numbers."""
        result = hello.greet("User123")
        self.assertEqual(result, "Hello, User123!")

    def test_greet_with_very_long_name(self):
        """Test greet() handles very long names."""
        long_name = "A" * 1000
        result = hello.greet(long_name)
        self.assertEqual(result, f"Hello, {long_name}!")

    def test_greet_type_error_on_invalid_input(self):
        """Test greet() raises TypeError on invalid input type."""
        with self.assertRaises(TypeError):
            hello.greet(123)  # Should raise TypeError for non-string

    def test_greet_type_error_on_list(self):
        """Test greet() raises TypeError on list input."""
        with self.assertRaises(TypeError):
            hello.greet(["Alice", "Bob"])


if __name__ == "__main__":
    unittest.main()
