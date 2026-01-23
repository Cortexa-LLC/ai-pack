"""Simple Calculator class with basic arithmetic operations."""


class Calculator:
    """A simple calculator that performs basic arithmetic operations."""

    def add(self, a: float, b: float) -> float:
        """Add two numbers together.

        Args:
            a: The first number
            b: The second number

        Returns:
            The sum of a and b
        """
        return a + b

    def subtract(self, a: float, b: float) -> float:
        """Subtract the second number from the first.

        Args:
            a: The first number
            b: The second number

        Returns:
            The difference of a and b (a - b)
        """
        return a - b
