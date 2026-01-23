"""String utility functions for common string operations."""


class StringUtils:
    """A utility class providing common string manipulation methods."""

    @staticmethod
    def reverse(s: str) -> str:
        """
        Reverse the given string.

        Args:
            s: The string to reverse.

        Returns:
            The reversed string.

        Example:
            >>> StringUtils.reverse("hello")
            'olleh'
        """
        return s[::-1]

    @staticmethod
    def capitalize_words(s: str) -> str:
        """
        Capitalize the first letter of each word in the string.

        Args:
            s: The string to capitalize.

        Returns:
            A new string with each word capitalized.

        Example:
            >>> StringUtils.capitalize_words("hello world")
            'Hello World'
        """
        return s.title()
