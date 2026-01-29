#!/usr/bin/env python3
"""
Hello from AI-Pack Module

A simple greeting module that prints "Hello from AI-Pack" with current timestamp.

Features:
- Type hints for all functions
- Comprehensive docstrings
- Proper error handling
- ISO 8601 timestamp format
- Test-driven development
"""

from datetime import datetime, timezone


def get_greeting() -> str:
    """
    Get the greeting message.

    Returns:
        str: The greeting message "Hello from AI-Pack"

    Examples:
        >>> get_greeting()
        'Hello from AI-Pack'
    """
    return "Hello from AI-Pack"


def get_timestamp() -> str:
    """
    Get the current timestamp in ISO 8601 format with UTC timezone.

    Returns:
        str: Current timestamp in ISO 8601 format (e.g., "2026-01-27T10:30:45.123456+00:00")

    Examples:
        >>> timestamp = get_timestamp()
        >>> isinstance(timestamp, str)
        True
        >>> 'T' in timestamp  # ISO format has 'T' separator
        True
    """
    return datetime.now(timezone.utc).isoformat()


def format_message(greeting: str, timestamp: str) -> str:
    """
    Format the complete message combining greeting and timestamp.

    Args:
        greeting: The greeting message to display
        timestamp: The timestamp string to display

    Returns:
        str: Formatted message with greeting and timestamp on separate lines

    Examples:
        >>> format_message("Hello from AI-Pack", "2026-01-27T10:00:00")
        'Hello from AI-Pack\\n2026-01-27T10:00:00'
    """
    return f"{greeting}\n{timestamp}"


def main() -> None:
    """
    Main entry point for the module.

    Prints the greeting message and current timestamp to stdout.
    """
    greeting = get_greeting()
    timestamp = get_timestamp()
    message = format_message(greeting, timestamp)
    print(message)


if __name__ == "__main__":
    main()
