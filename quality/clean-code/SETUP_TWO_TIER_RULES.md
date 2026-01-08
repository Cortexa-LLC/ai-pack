# Two-Tier Rule System Setup Guide

This guide explains how to implement the two-tier rule system (shared + project-specific) in your projects.

## How It Works

### Architecture

```
your-project/
├── .ai-pack/                          # Git submodule (shared standards)
│   ├── 01-design-principles.md          # Tier 1: Shared
│   ├── 02-solid-principles.md           # Tier 1: Shared
│   ├── lang-cpp.md                      # Tier 1: Shared
│   ├── lang-python.md                   # Tier 1: Shared
│   ├── PROJECT-README.md                # Tier 2: Project-specific
│   ├── PROJECT-sourcerer.md             # Tier 2: Project-specific
│   └── PROJECT-architecture.md          # Tier 2: Project-specific
└── .gitmodules                          # Git submodule configuration
```

**Key Design:**
- **Submodule files:** Tracked by `Cortexa-LLC/clean-code` repository
- **PROJECT-*.md files:** Tracked by your project repository (git-ignored by submodule)
- **Both discovered by Claude Code:** Automatically reads all `.md` files in `.ai-pack/`

## How Claude Code Discovers Both Tiers

**Claude Code automatically:**

1. **Scans `.ai-pack/` directory** for all `.md` files
2. **Reads shared standards** from submodule files (no `PROJECT-` prefix)
3. **Reads project-specific rules** from `PROJECT-*.md` files
4. **Applies both during code generation and review**

**No configuration needed!** Just place files in `.ai-pack/` and Claude finds them.

### Example Discovery Process

When Claude Code starts in a project with this structure:

```
.ai-pack/
├── 01-design-principles.md    # ✅ Discovered: Shared
├── 02-solid-principles.md     # ✅ Discovered: Shared
├── lang-cpp.md                # ✅ Discovered: Shared (2-space indentation)
├── PROJECT-README.md          # ✅ Discovered: Project
├── PROJECT-sourcerer.md       # ✅ Discovered: Project (formatting rules)
└── PROJECT-architecture.md    # ✅ Discovered: Project (CPU-agnostic rules)
```

**Claude applies:**
- ✅ SOLID principles (shared)
- ✅ C++ Core Guidelines (shared)
- ✅ 2-space indentation for C++ (shared lang-cpp.md)
- ✅ Sourcerer formatting rules (project-specific)
- ✅ CPU-agnostic architecture (project-specific)

## Setup: Adding to Existing Project

### Step 1: Add Submodule

```bash
cd your-project

# Add clean-code as submodule
git submodule add git@github.com:Cortexa-LLC/clean-code.git .ai-pack

# Initialize submodule
git submodule update --init --recursive
```

### Step 2: Create Project-Specific Rules

Create a directory to store your project-specific rules:

```bash
mkdir -p docs/project-rules
```

Create a `PROJECT-README.md` to explain your setup:

```bash
cat > docs/project-rules/PROJECT-README.md << 'EOF'
# Your Project Coding Standards

This project uses a **two-tier rule system**:

## Tier 1: Shared Standards (Submodule)
All files without `PROJECT-` prefix come from [Cortexa LLC Clean Code Standards](https://github.com/Cortexa-LLC/clean-code).

## Tier 2: Project-Specific Rules
- `PROJECT-yourproject.md` - Project-specific conventions
- `PROJECT-architecture.md` - Architecture patterns

**Both tiers are automatically discovered by Claude Code.**
EOF
```

Create project-specific rules:

```bash
cat > docs/project-rules/PROJECT-yourproject.md << 'EOF'
# Your Project Specific Rules

## Formatting
- Indentation: 2 spaces (for C++) or 4 spaces (for Python)
- Line length: 100 characters
- No trailing whitespace

## Project Conventions
- Use snake_case for database column names
- All API endpoints must use versioning
- GraphQL schemas go in `src/graphql/schemas/`

## Architecture
- Repository pattern for data access
- Service layer for business logic
- Controller layer for HTTP handlers
EOF
```

Copy to `.ai-pack/` for Claude Code to discover:

```bash
cp docs/project-rules/PROJECT-*.md .ai-pack/
```

### Step 3: Commit Project Files

```bash
# Stage submodule and project files
git add .gitmodules .ai-pack docs/project-rules/

# Commit
git commit -m "Add two-tier coding standards

- Add Cortexa LLC clean code standards as submodule
- Add project-specific rules in docs/project-rules/
- Working copies in .ai-pack/ for Claude Code discovery
- Setup two-tier rule system"
```

**Note:** The `PROJECT-*.md` files in `.ai-pack/` are git-ignored by the submodule and not tracked by your project repository. They are working copies for Claude Code to discover. The version-controlled source is in `docs/project-rules/`.

### Step 4: Push

```bash
git push origin main
```

**Done!** Claude Code will now discover and apply both shared and project-specific rules.

## Git Behavior Explanation

### What Gets Tracked Where

**Submodule Repository (`Cortexa-LLC/clean-code`):**
- Tracks: All files **except** `PROJECT-*.md`
- `.gitignore` contains:
  ```
  PROJECT-*.md
  project-*.md
  *-overrides.md
  ```

**Your Project Repository:**
- Tracks: `.gitmodules` (submodule configuration)
- Tracks: `.ai-pack` as a submodule reference (commit hash)
- Tracks: `PROJECT-*.md` files in a separate directory (e.g., `docs/project-rules/`)
- **Important:** Git submodules don't allow tracking files inside the submodule directory from the parent repository. Store `PROJECT-*.md` files elsewhere in your project (like `docs/project-rules/`) and copy them to `.ai-pack/` as working files for Claude Code to discover.

### How Updates Work

**Updating Shared Standards:**
```bash
# Pull latest shared standards
git submodule update --remote .ai-pack

# Commit the submodule update
git add .ai-pack
git commit -m "Update to latest clean code standards"
git push
```

**Updating Project Rules:**
```bash
# Edit working files in .ai-pack/
vim .ai-pack/PROJECT-yourproject.md

# Copy updated files back to version-controlled location
cp .ai-pack/PROJECT-*.md docs/project-rules/

# Commit changes
git add docs/project-rules/PROJECT-*.md
git commit -m "Update project-specific rules"
git push
```

**Note:** The `PROJECT-*.md` files in `.ai-pack/` are working copies that Claude Code discovers. The version-controlled copies live in `docs/project-rules/` (or similar directory in your project).

### Verification Commands

```bash
# Check submodule status
git submodule status

# See what's tracked by project
git status .ai-pack/

# See what's tracked by submodule
cd .ai-pack
git status
cd ..

# List all files Claude Code will see
ls -la .ai-pack/*.md
```

## Example: Sourcerer Project

The Sourcerer project uses this pattern:

**Tier 1 - Shared (from submodule):**
- C++ Core Guidelines (lang-cpp.md)
- SOLID Principles (02-solid-principles.md)
- Refactoring patterns (03-refactoring.md)

**Tier 2 - Project-Specific:**
- `PROJECT-sourcerer.md` - Formatting (2-space, Google C++ style)
- `PROJECT-architecture.md` - CPU-agnostic design rules

Claude Code automatically applies:
- ✅ C++ smart pointers (shared rule)
- ✅ 2-space indentation (shared lang-cpp.md)
- ✅ No tabs (project rule)
- ✅ CPU-agnostic analyzer (project rule)
- ✅ SOLID principles (shared rule)

## Language-Specific Indentation

The shared standards define language-specific indentation:

| Language | Indentation | Source |
|----------|-------------|--------|
| C++ | 2 spaces | lang-cpp.md (Google C++ Style Guide) |
| Python | 4 spaces | lang-python.md (PEP 8 mandatory) |
| JavaScript/TypeScript | 2 spaces | lang-javascript.md (Airbnb/Google) |
| Java | 4 spaces | lang-java.md (Oracle Conventions) |
| Kotlin | 4 spaces | lang-kotlin.md (Kotlin Conventions) |

**Projects inherit these automatically** unless overridden in `PROJECT-*.md` files.

## Troubleshooting

### Claude Not Discovering Project Rules

**Problem:** Claude only applies shared rules, not project-specific ones.

**Solution:**
1. Ensure files are named `PROJECT-*.md`
2. Ensure files are in `.ai-pack/` directory (working copies)
3. Verify files are plain Markdown (`.md` extension)
4. If missing, copy from version-controlled location:
   ```bash
   cp docs/project-rules/PROJECT-*.md .ai-pack/
   ```
5. Restart Claude Code

### Submodule Not Updating

**Problem:** `git submodule update` doesn't pull latest changes.

**Solution:**
```bash
# Update submodule to latest
git submodule update --remote .ai-pack

# Or manually:
cd .ai-pack
git pull origin main
cd ..
git add .ai-pack
git commit -m "Update submodule"
```

### Project Files Tracked by Submodule

**Problem:** `PROJECT-*.md` files show up in submodule's `git status`.

**Solution:**
1. Ensure submodule's `.gitignore` contains `PROJECT-*.md`
2. Remove from submodule tracking:
   ```bash
   cd .ai-pack
   git rm --cached PROJECT-*.md
   cd ..
   ```

### Conflicts During Submodule Update

**Problem:** Merge conflicts in submodule after update.

**Solution:**
```bash
# Reset submodule to remote state
cd .ai-pack
git fetch origin
git reset --hard origin/main
cd ..
git add .ai-pack
git commit -m "Reset submodule to latest"
```

## Advanced Usage

### Multiple Projects Sharing Same Standards

All Cortexa LLC projects can share the same submodule:

```bash
# Project A
cd project-a
git submodule add git@github.com:Cortexa-LLC/clean-code.git .ai-pack
mkdir -p docs/project-rules
cat > docs/project-rules/PROJECT-project-a.md << 'EOF'
# Project A specific rules
EOF
cp docs/project-rules/PROJECT-*.md .ai-pack/

# Project B
cd ../project-b
git submodule add git@github.com:Cortexa-LLC/clean-code.git .ai-pack
mkdir -p docs/project-rules
cat > docs/project-rules/PROJECT-project-b.md << 'EOF'
# Project B specific rules
EOF
cp docs/project-rules/PROJECT-*.md .ai-pack/
```

Both projects:
- ✅ Share the same core standards
- ✅ Have independent project rules
- ✅ Update standards independently

### Testing Rules Locally

```bash
# Clone just the standards to test
git clone git@github.com:Cortexa-LLC/clean-code.git /tmp/test-standards

# Add your project files
cat > /tmp/test-standards/.ai-pack/PROJECT-test.md << 'EOF'
# Test rules
EOF

# Open in Claude Code to test
cd /tmp/test-standards
# Claude will discover both shared and test rules
```

---

**The two-tier system provides:**
- ✅ **Consistency:** Shared standards across all projects
- ✅ **Flexibility:** Project-specific customization
- ✅ **Simplicity:** Claude discovers both automatically
- ✅ **Maintainability:** Update shared standards in one place
- ✅ **No Conflicts:** Clean separation via naming convention
