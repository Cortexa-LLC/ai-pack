"""
Simple Greeting Module

Provides a function to greet users by name.

This module follows clean code principles:
- Single Responsibility: One function, one purpose
- Pure function: No side effects
- Simple and readable
- Well documented

Public API:
    greet(name: str) -> str: Returns a greeting message

Example:
    >>> from greeting import greet
    >>> greet("Alice")
    'Hello, Alice!'
"""


def greet(name: str) -> str:
    """
    Generate a greeting message for the given name.

    This function takes a name as input and returns a personalized
    greeting message in the format "Hello, {name}!".

    Args:
        name (str): The name to include in the greeting.
                   Can be any string value including empty strings,
                   special characters, or unicode characters.

    Returns:
        str: A greeting message in the format "Hello, {name}!"

    Examples:
        >>> greet("Alice")
        'Hello, Alice!'

        >>> greet("John Doe")
        'Hello, John Doe!'

        >>> greet("")
        'Hello, !'

        >>> greet("José")
        'Hello, José!'

    Notes:
        - The function accepts any string input without validation
        - Empty strings are valid input
        - Special characters and unicode are preserved
        - No trimming or modification of the input name is performed
    """
    return f"Hello, {name}!"


# Module metadata
__version__ = "1.0.0"
__author__ = "AI Pack Team"
__all__ = ["greet"]
