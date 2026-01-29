#!/usr/bin/env python3
"""
Tests for hello.py module.

Following TDD approach - tests written first.
Tests verify:
- Correct greeting message "Hello from AI-Pack"
- Timestamp is included
- Timestamp format is valid ISO 8601
- Output includes both greeting and timestamp
"""

import unittest
import sys
from io import StringIO
from pathlib import Path
from datetime import datetime
import re

# Add current directory to path for import
sys.path.insert(0, str(Path(__file__).parent))

import hello


class TestHelloModule(unittest.TestCase):
    """Test cases for hello module."""

    def test_get_greeting_returns_string(self):
        """Test that get_greeting() returns a string."""
        result = hello.get_greeting()
        self.assertIsInstance(result, str)

    def test_get_greeting_contains_ai_pack(self):
        """Test get_greeting() contains 'Hello from AI-Pack'."""
        result = hello.get_greeting()
        self.assertIn("Hello from AI-Pack", result)

    def test_get_timestamp_returns_string(self):
        """Test that get_timestamp() returns a string."""
        result = hello.get_timestamp()
        self.assertIsInstance(result, str)

    def test_get_timestamp_is_valid_iso_format(self):
        """Test get_timestamp() returns valid ISO 8601 format."""
        result = hello.get_timestamp()
        # Should be parseable as datetime
        try:
            parsed = datetime.fromisoformat(result.replace('Z', '+00:00'))
            self.assertIsInstance(parsed, datetime)
        except ValueError:
            self.fail(f"Timestamp '{result}' is not valid ISO 8601 format")

    def test_format_message_combines_greeting_and_timestamp(self):
        """Test format_message() combines greeting and timestamp."""
        greeting = "Hello from AI-Pack"
        timestamp = "2026-01-27T10:00:00"
        result = hello.format_message(greeting, timestamp)
        
        self.assertIn(greeting, result)
        self.assertIn(timestamp, result)

    def test_format_message_has_newline_separator(self):
        """Test format_message() uses newline to separate components."""
        greeting = "Hello from AI-Pack"
        timestamp = "2026-01-27T10:00:00"
        result = hello.format_message(greeting, timestamp)
        
        self.assertIn("\n", result)

    def test_main_prints_complete_message(self):
        """Test main() prints greeting and timestamp."""
        captured_output = StringIO()
        sys.stdout = captured_output
        
        hello.main()
        
        sys.stdout = sys.__stdout__
        output = captured_output.getvalue()
        
        self.assertIn("Hello from AI-Pack", output)
        # Check for timestamp pattern (ISO 8601)
        iso_pattern = r'\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}'
        self.assertIsNotNone(re.search(iso_pattern, output))

    def test_module_has_docstring(self):
        """Test module has docstring."""
        self.assertIsNotNone(hello.__doc__)
        self.assertGreater(len(hello.__doc__), 0)

    def test_get_greeting_has_docstring(self):
        """Test get_greeting function has docstring."""
        self.assertIsNotNone(hello.get_greeting.__doc__)
        self.assertGreater(len(hello.get_greeting.__doc__), 0)

    def test_get_timestamp_has_docstring(self):
        """Test get_timestamp function has docstring."""
        self.assertIsNotNone(hello.get_timestamp.__doc__)
        self.assertGreater(len(hello.get_timestamp.__doc__), 0)

    def test_format_message_has_docstring(self):
        """Test format_message function has docstring."""
        self.assertIsNotNone(hello.format_message.__doc__)
        self.assertGreater(len(hello.format_message.__doc__), 0)


class TestHelloEdgeCases(unittest.TestCase):
    """Edge case tests for hello module."""

    def test_get_timestamp_is_current(self):
        """Test get_timestamp() returns current time (within 5 seconds)."""
        from datetime import timezone
        before = datetime.now(timezone.utc)
        timestamp_str = hello.get_timestamp()
        after = datetime.now(timezone.utc)
        
        # Parse the timestamp
        timestamp = datetime.fromisoformat(timestamp_str.replace('Z', '+00:00'))
        
        # Should be between before and after (with small margin)
        self.assertLessEqual(before, timestamp)
        self.assertGreaterEqual(after, timestamp)

    def test_format_message_handles_empty_greeting(self):
        """Test format_message() handles empty greeting gracefully."""
        result = hello.format_message("", "2026-01-27T10:00:00")
        self.assertIsInstance(result, str)
        self.assertIn("2026-01-27T10:00:00", result)

    def test_format_message_handles_empty_timestamp(self):
        """Test format_message() handles empty timestamp gracefully."""
        result = hello.format_message("Hello from AI-Pack", "")
        self.assertIsInstance(result, str)
        self.assertIn("Hello from AI-Pack", result)

    def test_get_greeting_is_consistent(self):
        """Test get_greeting() returns same value on multiple calls."""
        result1 = hello.get_greeting()
        result2 = hello.get_greeting()
        self.assertEqual(result1, result2)


if __name__ == "__main__":
    unittest.main()
