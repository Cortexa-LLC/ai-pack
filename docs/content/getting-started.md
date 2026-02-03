---
sidebar_position: 2
---

# Getting Started

This guide will help you integrate AI-Pack into your project.

## Prerequisites

- Git
- Python 3.8+ (for Claude Code hooks)
- Node.js 20+ (for Docusaurus documentation)
- [Beads](https://github.com/steveyegge/beads) task tracker (optional but recommended)

## Installation

### 1. Add AI-Pack as a Submodule

```bash
cd your-project
git submodule add https://github.com/Cortexa-LLC/ai-pack .ai-pack
git submodule update --init --recursive
```text

This creates a read-only `.ai-pack/` directory containing the framework.

### 2. Create Local Workspace

```bash
# Create task workspace (temporary)
mkdir -p .ai/tasks

# Create permanent documentation directory
mkdir -p docs/{market,product,design,architecture,adr,investigations,archaeology,incidents}
```text

### 3. Initialize Beads (Recommended)

Beads provides persistent, git-backed task tracking:

```bash
# Install Beads (macOS/Linux)
curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/scripts/install.sh | bash

# Windows PowerShell
irm https://raw.githubusercontent.com/steveyegge/beads/main/install.ps1 | iex

# Initialize in your project
bd init
```text

### 4. Copy Bootstrap Template

```bash
cp .ai-pack/templates/CLAUDE.md ./CLAUDE.md
```text

Edit `CLAUDE.md` with project-specific details:
- Project name
- Technology stack
- Key files and directories
- Special considerations

### 5. Commit Framework Setup

```bash
git add .ai-pack .beads/issues.jsonl CLAUDE.md
git commit -m "Add ai-pack framework and initialize Beads"
```text

## Claude Code Integration (Optional)

For native Claude Code integration with slash commands, skills, and hooks:

### 1. Run Setup Script

```bash
python3 .ai-pack/templates/.claude-setup.py
```text

This creates:
- `.claude/commands/ai-pack/` - Slash commands
- `.claude/skills/` - Auto-triggered roles
- `.claude/rules/` - Modular enforcement rules
- `.claude/hooks/` - Python enforcement scripts
- `.claude/settings.json` - Hook configuration

### 2. Configure Permissions

Background agents need file operation permissions. The setup script creates `.claude/settings.json` with safe defaults:

```json
{
  "description": "AI-Pack Framework Configuration for Claude Code",
  "permissions": {
    "allow": [
      "Write(*)",
      "Edit(*)",
      "Bash(mkdir:*)",
      "Bash(cp:*)",
      "Bash(mv:*)",
      "Bash(ls:*)",
      "Bash(cat:*)",
      "Bash(git pull:*)",
      "Bash(git fetch:*)",
      "Bash(git status:*)",
      "Bash(git diff:*)",
      "Bash(git log:*)",
      "Bash(npm:*)",
      "Bash(dotnet:*)",
      "Bash(pytest:*)"
    ],
    "deny": [
      "Bash(rm -rf /*)",
      "Bash(git push --force*)",
      "Bash(sudo rm*)"
    ],
    "defaultMode": "acceptEdits"
  },
  "hooks": {
    "SessionStart": [
      {
        "description": "Start ai-pack monitoring timers",
        "hooks": [{
          "type": "command",
          "command": "python3 .claude/hooks/session-start.py"
        }]
      }
    ],
    "UserPromptSubmit": [
      {
        "description": "Enforce task packet gate",
        "hooks": [{
          "type": "command",
          "command": "python3 .claude/hooks/check-task-packet.py"
        }]
      }
    ]
  }
}
```text

**Key Changes from Old Config:**
- `defaultMode: "acceptEdits"` (not `"bypassPermissions"`) - More granular control
- Explicit allow list for safe Bash commands
- Explicit deny list for dangerous operations
- Session hooks for monitoring timers

**Security Note:** This configuration balances safety with productivity. Destructive operations still require approval.

See [Claude Code Configuration](./claude-code-configuration.md) for complete details.

### 3. Restart Claude Code

Close and reopen Claude Code for settings to take effect.

## Directory Structure

After setup, your project will have:

```text
your-project/
├── .ai-pack/                    # Git submodule (read-only)
│   ├── quality/                 # Clean code standards
│   ├── gates/                   # Quality gates
│   ├── roles/                   # Agent roles
│   ├── workflows/               # Development workflows
│   └── templates/               # Task-packet templates
│
├── .beads/                      # Beads task memory (committed)
│   ├── issues.jsonl             # Git-tracked task database
│   └── *.db                     # Local cache (git-ignored)
│
├── .ai/                         # Local workspace (temporary)
│   ├── tasks/                   # Active task packets
│   └── repo-overrides.md        # Project-specific rules
│
├── docs/                        # Permanent documentation
│   ├── market/                  # Market requirements (Strategist)
│   ├── product/                 # Product requirements (Product Manager)
│   ├── design/                  # UX design (Designer)
│   ├── architecture/            # Technical design (Architect)
│   ├── adr/                     # Architecture Decision Records
│   ├── investigations/          # Bug retrospectives (Inspector)
│   ├── archaeology/             # Legacy code investigations (Archaeologist)
│   └── incidents/               # Production incidents (Spelunker)
│
├── .claude/                     # Claude Code integration (optional)
│   ├── commands/ai-pack/        # Slash commands
│   ├── skills/                  # Auto-triggered roles
│   ├── rules/                   # Modular rules
│   ├── hooks/                   # Enforcement scripts
│   └── settings.json            # Configuration
│
└── CLAUDE.md                    # Bootstrap instructions
```text

## Basic Usage

### Working with Tasks

```bash
# Create a task
bd create "Implement user authentication" --priority high

# List available tasks (no blocking dependencies)
bd ready

# Start working on a task
bd start bd-a1b2

# View task details
bd show bd-a1b2

# Mark task complete
bd close bd-a1b2

# Add dependency (task Y depends on task X)
bd dep add bd-x bd-y
```text

### Creating Task Packets

For structured task tracking:

```bash
# Create task directory
TASK_ID=$(date +%Y-%m-%d)_feature-name
mkdir -p .ai/tasks/$TASK_ID

# Copy templates
cp .ai-pack/templates/task-packet/00-contract.md .ai/tasks/$TASK_ID/
cp .ai-pack/templates/task-packet/10-plan.md .ai/tasks/$TASK_ID/
cp .ai-pack/templates/task-packet/20-work-log.md .ai/tasks/$TASK_ID/
cp .ai-pack/templates/task-packet/30-review.md .ai/tasks/$TASK_ID/
cp .ai-pack/templates/task-packet/40-acceptance.md .ai/tasks/$TASK_ID/
```text

Or use the Claude Code command:

```text
/ai-pack task-init feature-name
```text

### Following the Workflow

1. **Track** - Mark task in progress: `bd start bd-a1b2`
2. **Define** - Fill out `00-contract.md` with requirements
3. **Plan** - Create implementation plan in `10-plan.md`
4. **Execute** - Implement while updating `20-work-log.md`
5. **Review** - Conduct review, document in `30-review.md`
6. **Accept** - Complete acceptance checklist in `40-acceptance.md`
7. **Complete** - Mark done: `bd close bd-a1b2`
8. **Next** - Find next task: `bd ready`

## Available Slash Commands (Claude Code)

After setup, these commands are available:

- `/ai-pack task-init <name>` - Create task packet
- `/ai-pack task-status` - Check progress
- `/ai-pack orchestrate` - Assume Orchestrator role
- `/ai-pack engineer` - Assume Engineer role
- `/ai-pack test` - Validate tests (MANDATORY)
- `/ai-pack review` - Code review (MANDATORY)
- `/ai-pack inspect` - Bug investigation
- `/ai-pack architect` - Architecture design
- `/ai-pack designer` - UX workflows
- `/ai-pack strategist` - Market analysis
- `/ai-pack product-manager` - Product requirements
- `/ai-pack help` - Show all commands

## Next Steps

- **[Agent-to-Agent (A2A)](./framework/agent-to-agent)** - Learn about the multi-agent workflow system
- **[Workflows](./workflows/overview)** - Learn about development processes
- **[Roles](./roles/overview)** - Understand agent personas
- **[Quality Gates](./gates/overview)** - Review enforcement rules
- **[Clean Code](./quality/clean-code)** - Study coding standards
- **[Beads Integration](./quality/tooling/beads-integration)** - Deep dive into task tracking

## Updating AI-Pack

To get the latest framework updates:

```bash
cd .ai-pack
git pull origin main
cd ..
git add .ai-pack
git commit -m "Update ai-pack framework"
```text

For Claude Code integration updates:

```bash
python3 .ai-pack/templates/.claude-update.py
git add .claude/
git commit -m "Update ai-pack Claude Code integration"
```text

## Troubleshooting

### Background Agents Cannot Write Files

If agents fail with permission errors:

1. Check `.claude/settings.json` has correct permissions
2. Verify `defaultMode: "bypassPermissions"` is set
3. Restart Claude Code

### Task Packet Gate Blocking Work

If the hook blocks implementation:

1. Create task packet: `/ai-pack task-init <name>`
2. Fill out `00-contract.md` at minimum
3. Try again

### Beads Command Not Found

Install Beads:

```bash
# macOS/Linux
curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/scripts/install.sh | bash

# Windows
irm https://raw.githubusercontent.com/steveyegge/beads/main/install.ps1 | iex
```text

## Support

- [GitHub Issues](https://github.com/Cortexa-LLC/ai-pack/issues)
- [GitHub Discussions](https://github.com/Cortexa-LLC/ai-pack/discussions)
- [Claude Code Documentation](https://code.claude.com/docs)
- [Beads Documentation](https://github.com/steveyegge/beads)
