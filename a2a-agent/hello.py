#!/usr/bin/env python3
"""
Hello World Module

A simple greeting module demonstrating clean Python code with:
- Type hints
- Comprehensive docstrings
- Proper error handling
- Test-driven development
"""

from typing import Optional


def greet(name: Optional[str] = None) -> str:
    """
    Generate a greeting message.

    Args:
        name: The name to greet. If None, empty, or whitespace-only,
              defaults to "World".

    Returns:
        A greeting string in the format "Hello, {name}!"

    Raises:
        TypeError: If name is not a string, None, or cannot be processed as text.

    Examples:
        >>> greet()
        'Hello, World!'
        >>> greet("Alice")
        'Hello, Alice!'
        >>> greet("  Bob  ")
        'Hello, Bob!'
    """
    # Validate input type
    if name is not None and not isinstance(name, str):
        raise TypeError(
            f"name must be a string or None, got {type(name).__name__}"
        )

    # Handle None, empty string, or whitespace-only strings
    if not name or not name.strip():
        name = "World"
    else:
        # Strip leading/trailing whitespace
        name = name.strip()

    return f"Hello, {name}!"


def main() -> None:
    """
    Main entry point for the module.

    Prints a greeting to stdout.
    """
    print(greet())


if __name__ == "__main__":
    main()
