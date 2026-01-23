#!/usr/bin/env python3
"""
Greeting Function Demo

Demonstrates the usage of the greeting function with various examples.
"""

from greeting import greet


def main():
    """Run demonstration of greeting function"""

    print("\n" + "="*60)
    print(" GREETING FUNCTION DEMONSTRATION")
    print("="*60 + "\n")

    # Example 1: Simple greeting
    print("Example 1: Simple greeting")
    print(f"  Input: 'Alice'")
    print(f"  Output: {greet('Alice')}")
    print()

    # Example 2: Name with spaces
    print("Example 2: Name with spaces")
    print(f"  Input: 'John Doe'")
    print(f"  Output: {greet('John Doe')}")
    print()

    # Example 3: Unicode name
    print("Example 3: Unicode character support")
    print(f"  Input: 'José García'")
    print(f"  Output: {greet('José García')}")
    print()

    # Example 4: Batch processing
    print("Example 4: Batch processing")
    users = ["Alice", "Bob", "Charlie", "Diana"]
    print(f"  Input: {users}")
    greetings = [greet(user) for user in users]
    print("  Output:")
    for greeting in greetings:
        print(f"    - {greeting}")
    print()

    # Example 5: Integration scenario
    print("Example 5: API-like usage")
    def create_api_response(name: str) -> dict:
        return {
            "status": "success",
            "message": greet(name),
            "timestamp": "2026-01-23T12:44:43Z"
        }

    response = create_api_response("World")
    print(f"  Input: 'World'")
    print(f"  Output (JSON):")
    import json
    print(f"    {json.dumps(response, indent=4)}")
    print()

    print("="*60)
    print(" DEMONSTRATION COMPLETE")
    print("="*60 + "\n")


if __name__ == '__main__':
    main()
