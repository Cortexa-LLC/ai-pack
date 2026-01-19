# GitHub Integration Setup Guide

## Installation Paths

AI-Pack can be installed at different locations in your project as a git submodule. This guide uses `${AI_PACK_ROOT}` to represent the installation path.

### Common Installation Paths

```bash
# Option 1: Hidden directory (recommended)
.ai-pack/

# Option 2: Visible directory
ai-pack/

# Option 3: Tools subdirectory
tools/ai-pack/

# Option 4: Custom path (any location works)
vendor/ai-pack/
lib/ai-pack/
external/dependencies/ai-pack/
```

### How Auto-Detection Works

The detection scripts provided below automatically find AI-Pack using:

1. **Git submodule configuration** - Reads `.gitmodules` to find the exact path
2. **Filesystem search** - Searches for directories named `ai-pack` or `.ai-pack`
3. **Directory tree walk** - Walks up parent directories to find AI-Pack

This works for **any installation location**, not just common paths.

## Environment Setup

### Option 1: Shell Alias (Recommended)

Add to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
# Detect AI-Pack location (works for any installation path)
detect_ai_pack() {
  # Method 1: Check git submodules (ai-pack is always a submodule)
  if [ -f .gitmodules ]; then
    local submodule_path=$(git config --file .gitmodules --get-regexp path | \
                          grep -i 'ai-pack' | \
                          awk '{print $2}' | \
                          head -1)
    if [ -n "$submodule_path" ] && [ -d "$submodule_path" ]; then
      echo "$submodule_path"
      return 0
    fi
  fi

  # Method 2: Search for ai-pack directory (handles non-submodule cases)
  local found=$(find . -maxdepth 3 -type d \( -name "ai-pack" -o -name ".ai-pack" \) 2>/dev/null | head -1)
  if [ -n "$found" ]; then
    echo "${found#./}"  # Remove leading ./
    return 0
  fi

  # Method 3: Walk up directory tree
  local dir="$PWD"
  while [ "$dir" != "/" ]; do
    if [ -d "$dir/.ai-pack" ]; then
      # Return relative path from current directory
      realpath --relative-to="$PWD" "$dir/.ai-pack" 2>/dev/null || echo ".ai-pack"
      return 0
    fi
    dir=$(dirname "$dir")
  done

  # Not found
  return 1
}

export AI_PACK_ROOT=$(detect_ai_pack)
if [ -z "$AI_PACK_ROOT" ]; then
  echo "Warning: AI-Pack not found, defaulting to .ai-pack"
  export AI_PACK_ROOT=".ai-pack"
fi

# Convenience aliases
alias gh-init='${AI_PACK_ROOT}/scripts/github-integration.py init'
alias gh-sync='${AI_PACK_ROOT}/scripts/github-integration.py sync'
alias gh-status='${AI_PACK_ROOT}/scripts/github-integration.py status'
alias gh-export='${AI_PACK_ROOT}/scripts/github-integration.py export'
alias gh-import='${AI_PACK_ROOT}/scripts/github-integration.py import'
alias gh-epic='${AI_PACK_ROOT}/scripts/github-integration.py create-epic'
```

### Option 2: Project Script Wrapper

Create `scripts/gh-integration` in your project root:

```bash
#!/bin/bash
# Wrapper for GitHub integration - detects AI-Pack location

# Detect AI-Pack installation (works for any path)
detect_ai_pack() {
  # Method 1: Check git submodules
  if [ -f .gitmodules ]; then
    local submodule_path=$(git config --file .gitmodules --get-regexp path | \
                          grep -i 'ai-pack' | \
                          awk '{print $2}' | \
                          head -1)
    if [ -n "$submodule_path" ] && [ -d "$submodule_path" ]; then
      echo "$submodule_path"
      return 0
    fi
  fi

  # Method 2: Search for ai-pack directory
  local found=$(find . -maxdepth 3 -type d \( -name "ai-pack" -o -name ".ai-pack" \) 2>/dev/null | head -1)
  if [ -n "$found" ]; then
    echo "${found#./}"
    return 0
  fi

  return 1
}

AI_PACK_ROOT=$(detect_ai_pack)
if [ -z "$AI_PACK_ROOT" ]; then
  echo "Error: AI-Pack not found in this repository"
  echo "Expected: Git submodule or directory named 'ai-pack' or '.ai-pack'"
  exit 1
fi

# Execute the integration script
exec "${AI_PACK_ROOT}/scripts/github-integration.py" "$@"
```

Make it executable:
```bash
chmod +x scripts/gh-integration
```

Usage:
```bash
./scripts/gh-integration init
./scripts/gh-integration sync
./scripts/gh-integration create-epic bd-epic-123
```

### Option 3: Environment Variable

Set in your project:

```bash
# .env file or shell
export AI_PACK_ROOT=".ai-pack"

# Use in commands
${AI_PACK_ROOT}/scripts/github-integration.py sync
```

## Configuration File Location

The GitHub integration configuration follows the same pattern:

```
${AI_PACK_ROOT}/.github-integration.yml
```

**Default Locations Checked (in order):**
1. `${AI_PACK_ROOT}/.github-integration.yml` (if AI_PACK_ROOT set)
2. `.ai-pack/.github-integration.yml`
3. `ai-pack/.github-integration.yml`
4. `.github-integration.yml` (project root)

## Quick Start Examples

### Using Shell Alias
```bash
# Initialize
gh-init

# Configure
# Edit ${AI_PACK_ROOT}/.github-integration.yml

# Set token
export GITHUB_TOKEN="ghp_your_token_here"

# Sync
gh-sync
```

### Using Project Wrapper
```bash
# Initialize
./scripts/gh-integration init

# Configure
# Edit ${AI_PACK_ROOT}/.github-integration.yml

# Sync
./scripts/gh-integration sync
```

### Direct Invocation
```bash
# Replace with your actual path
${AI_PACK_ROOT}/scripts/github-integration.py init
${AI_PACK_ROOT}/scripts/github-integration.py sync
```

## Integration with CI/CD

### GitHub Actions

```yaml
name: Beads ↔ GitHub Sync

on:
  schedule:
    - cron: '*/5 * * * *'  # Every 5 minutes
  workflow_dispatch:

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          submodules: true  # Important: fetch AI-Pack submodule

      - name: Detect AI-Pack location
        id: detect
        run: |
          # Method 1: Check git submodules
          if [ -f .gitmodules ]; then
            SUBMODULE_PATH=$(git config --file .gitmodules --get-regexp path | \
                            grep -i 'ai-pack' | \
                            awk '{print $2}' | \
                            head -1)
            if [ -n "$SUBMODULE_PATH" ] && [ -d "$SUBMODULE_PATH" ]; then
              echo "path=$SUBMODULE_PATH" >> $GITHUB_OUTPUT
              exit 0
            fi
          fi

          # Method 2: Search filesystem
          FOUND=$(find . -maxdepth 3 -type d \( -name "ai-pack" -o -name ".ai-pack" \) 2>/dev/null | head -1)
          if [ -n "$FOUND" ]; then
            echo "path=${FOUND#./}" >> $GITHUB_OUTPUT
            exit 0
          fi

          echo "Error: AI-Pack not found"
          exit 1

      - name: Sync Beads with GitHub
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          ${{ steps.detect.outputs.path }}/scripts/github-integration.py sync
```

### Cron Job

```bash
# Detect AI-Pack location in cron (works for any installation path)
*/5 * * * * cd /path/to/project && \
  AI_PACK_ROOT=$(git config --file .gitmodules --get-regexp path | grep -i 'ai-pack' | awk '{print $2}' | head -1) && \
  [ -z "$AI_PACK_ROOT" ] && AI_PACK_ROOT=$(find . -maxdepth 3 -type d \( -name "ai-pack" -o -name ".ai-pack" \) 2>/dev/null | head -1 | sed 's|^\./||') && \
  ${AI_PACK_ROOT}/scripts/github-integration.py sync >> sync.log 2>&1
```

## Documentation Path References

Throughout this documentation, when you see:
- `${AI_PACK_ROOT}/scripts/...` - Replace with your actual AI-Pack path
- `${AI_PACK_ROOT}/.github-integration.yml` - Configuration file location

**Examples:**
```bash
# If installed at .ai-pack/
.ai-pack/scripts/github-integration.py sync
.ai-pack/.github-integration.yml

# If installed at ai-pack/
ai-pack/scripts/github-integration.py sync
ai-pack/.github-integration.yml

# If installed at tools/ai-pack/
tools/ai-pack/scripts/github-integration.py sync
tools/ai-pack/.github-integration.yml
```

## Next Steps

1. Choose your preferred setup method above
2. Initialize integration: `gh-init` or equivalent
3. Configure: Edit `${AI_PACK_ROOT}/.github-integration.yml`
4. Set GitHub token: `export GITHUB_TOKEN="..."`
5. Test: `gh-status` or equivalent
6. Start syncing: `gh-sync` or equivalent

## See Also

- [GitHub Integration Usage](GITHUB-INTEGRATION-USAGE.md) - Complete feature guide
- [Work Item Patterns](WORK-ITEM-PATTERNS.md) - Epic/Story/Task patterns
- [Integration Summary](GITHUB-INTEGRATION-SUMMARY.md) - Quick reference
