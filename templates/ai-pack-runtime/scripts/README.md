# AI-Pack Scripts

This directory contains custom automation scripts that can be executed by AI agents.

## Purpose

Scripts provide reusable, cross-platform automation for complex operations that are:
- Too complex for simple tool calls
- Require specialized libraries (Python packages, npm modules)
- Need to be version-controlled and tested
- Shared across multiple agents

## Supported Script Types

### Python Scripts (.py)
- **Interpreter**: `python3`
- **Use Cases**: Data analysis, dependency management, code refactoring
- **Common Libraries**: pandas, requests, pathlib, ast, json

### Node.js Scripts (.js)
- **Interpreter**: `node`
- **Use Cases**: API validation, package management, build automation
- **Common Libraries**: axios, fs, openapi-validator

### Ruby Scripts (.rb)
- **Interpreter**: `ruby`
- **Use Cases**: Text processing, file manipulation
- **Status**: Supported but less common

## Script Requirements

### Shebang Line
All scripts should include a shebang line:

```python
#!/usr/bin/env python3
```

```javascript
#!/usr/bin/env node
```

### Documentation
Each script should include:
- Purpose description
- Usage instructions
- Input/output format
- Example usage

### Example Script Template

```python
#!/usr/bin/env python3
"""
Script Name: analyze_dependencies.py
Purpose: Analyze Python project dependencies and find unused packages
Usage: python analyze_dependencies.py [--format json|text]
Output: JSON or text report of unused/missing packages
"""

import sys
import json
import argparse

def main():
    parser = argparse.ArgumentParser(description='Analyze dependencies')
    parser.add_argument('--format', choices=['json', 'text'], default='text')
    args = parser.parse_args()

    # Script logic here
    result = {"unused": [], "missing": []}

    if args.format == 'json':
        print(json.dumps(result, indent=2))
    else:
        print(f"Unused packages: {', '.join(result['unused'])}")

if __name__ == "__main__":
    main()
```

## Security

### Approval Process
1. First execution of any script requires user approval
2. Approval is tracked in `../.approved-scripts.json`
3. Changes to script require re-approval (SHA256 hash verification)

### Script Restrictions
- Must be located in this directory (`.ai-pack/scripts/`)
- Cannot escape to parent directories
- Subject to timeout limits (5 min default)
- Output size limits (10MB default)
- Network access allowed but rate-limited

### Role-Based Access
Different agent roles have access to different scripts:
- **Architect**: analysis, validation scripts
- **Engineer**: generation, testing scripts
- **Refactor**: refactoring, migration scripts
- See `tool-permissions.yml` for complete mappings

## Example Scripts (Coming in Phase 2)

The following scripts will be provided as examples:

### analyze_dependencies.py
Analyze project dependencies and find unused packages.

### refactor_imports.py
Bulk update import statements across codebase.

### migrate_database.py
Database migration helper script.

### generate_mocks.py
Generate test mocks from interfaces/types.

### validate_api_contracts.js
Validate API responses against OpenAPI spec.

## Creating New Scripts

1. **Create script file** in this directory:
   ```bash
   touch .ai-pack/scripts/my_script.py
   chmod +x .ai-pack/scripts/my_script.py
   ```

2. **Add shebang and documentation**:
   ```python
   #!/usr/bin/env python3
   """
   Description and usage here
   """
   ```

3. **Implement logic** with proper error handling

4. **Test manually**:
   ```bash
   python3 .ai-pack/scripts/my_script.py --help
   ```

5. **Use in agent workflow** - first execution will prompt for approval

## Agent Usage

Agents execute scripts using the `execute_script` tool:

```python
# Example: Agent executing dependency analysis
result = execute_script(
    script="analyze_dependencies.py",
    args=["--format", "json"]
)
```

## Best Practices

1. **Output Format**: Support both JSON and human-readable output
2. **Error Handling**: Exit with non-zero code on errors
3. **Idempotency**: Scripts should be safe to run multiple times
4. **Documentation**: Include help text and examples
5. **Dependencies**: Document any required packages
6. **Cross-Platform**: Use pathlib, avoid platform-specific commands

## Debugging

Scripts write to stdout/stderr which is captured by the agent runtime:

```python
import sys

# Normal output
print("Result data")

# Errors and warnings
print("Warning: something happened", file=sys.stderr)

# Exit codes
sys.exit(0)  # Success
sys.exit(1)  # Error
```

## Resources

- **Implementation Plan**: `docs/A2A-IMPLEMENTATION-PLAN.md` (Section 2.4)
- **Tool Permissions**: `.ai-pack/config.yml` (scripts section)
- **Security Model**: `docs/A2A-IMPLEMENTATION-PLAN.md` (Security Model section)

## Status

**Current**: Directory structure ready
**Next**: Phase 2 implementation will add example scripts
**Timeline**: Week 5-6 of implementation roadmap
