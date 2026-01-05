# Python Language-Specific Rules

> Based on PEP 8, PEP 20 (Zen of Python), and modern Python best practices

## Formatting Standards (Python Specific)

**Indentation:** **4 spaces** (no tabs) - **Mandatory per PEP 8**

Python's PEP 8 style guide mandates 4-space indentation. This is non-negotiable in the Python community.

**Note:** Cortexa LLC uses **language-specific indentation standards**:
- **C++**: 2 spaces (see lang-cpp.md)
- **Python**: 4 spaces (this document - PEP 8)
- **JavaScript/TypeScript**: 2 spaces (see lang-javascript.md)
- **Java**: 4 spaces (see lang-java.md)
- **Kotlin**: 4 spaces (see lang-kotlin.md)

**Example:**
```python
class DataAnalyzer:
    """Analyzes data streams."""

    def __init__(self, source: str) -> None:
        """Initialize analyzer with data source."""
        self.source = source
        self.results: list[dict] = []

    def analyze(self, data: bytes) -> dict:
        """Analyze data and return results."""
        if not data:
            return {}

        # 4-space indentation throughout
        result = {
            'source': self.source,
            'size': len(data),
            'valid': self._validate(data)
        }
        return result

    def _validate(self, data: bytes) -> bool:
        """Private method for validation."""
        return len(data) > 0
```

---

## Overview

This file will contain Python-specific best practices including:
- **PEP 8** - Official Python Style Guide
- **PEP 20** - The Zen of Python
- **PEP 257** - Docstring Conventions
- **Type Hints** - Modern Python type annotations (PEP 484)
- **Testing** - pytest best practices
- **Pythonic Idioms** - Idiomatic Python patterns

---

## Quick Standards Summary

### Formatting
- **Indentation:** 4 spaces (mandatory)
- **Line Length:** 79 characters (code), 72 characters (docstrings/comments)
- **Imports:** Separate stdlib, third-party, local with blank lines
- **Blank Lines:** 2 between top-level definitions, 1 between methods

### Naming
- `module_name` - lowercase with underscores
- `ClassName` - CapWords (PascalCase)
- `function_name` - lowercase with underscores
- `variable_name` - lowercase with underscores
- `CONSTANT_NAME` - uppercase with underscores
- `_private_name` - leading underscore for internal use

### Type Hints (Python 3.10+)
```python
def process_data(items: list[str], count: int = 10) -> dict[str, int]:
    """Process items and return counts."""
    return {item: len(item) for item in items[:count]}
```

### Docstrings (PEP 257)
```python
def complex_function(arg1: str, arg2: int) -> bool:
    """
    One-line summary ending with period.

    More detailed explanation of what the function does,
    including any side effects or important notes.

    Args:
        arg1: Description of first argument
        arg2: Description of second argument

    Returns:
        Description of return value

    Raises:
        ValueError: When arg2 is negative
    """
    if arg2 < 0:
        raise ValueError("arg2 must be non-negative")
    return len(arg1) > arg2
```

---

## TODO: Full Python Guidelines

This file will be expanded to include:
- [ ] Zen of Python principles
- [ ] Comprehensive PEP 8 coverage
- [ ] Type hinting best practices
- [ ] Error handling patterns
- [ ] Context managers and RAII equivalents
- [ ] Generator and iterator patterns
- [ ] Async/await best practices
- [ ] Testing with pytest
- [ ] Common anti-patterns
- [ ] Python idioms and patterns

---

**For now, always use 4-space indentation per PEP 8. Full guidelines coming soon.**
