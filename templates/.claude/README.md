# AI-Pack Claude Code Integration

This directory contains Claude Code integration for the ai-pack framework, providing commands, skills, rules, and hooks that enforce framework standards.

## Directory Structure

```
.claude/
├── commands/
│   └── ai-pack/           # Slash commands (/ai-pack <command>)
│       ├── task-init.md
│       ├── task-status.md
│       ├── orchestrate.md
│       ├── engineer.md
│       ├── review.md
│       ├── test.md
│       ├── inspect.md
│       ├── architect.md
│       ├── designer.md
│       ├── pm.md
│       └── help.md
│
├── skills/                # Auto-triggered role behaviors
│   ├── orchestrator/SKILL.md
│   ├── engineer/SKILL.md
│   └── README.md
│
├── rules/                 # Modular rules (auto-loaded)
│   ├── gates.md
│   ├── task-packets.md
│   ├── workflows.md
│   └── README.md
│
├── hooks/                 # Enforcement scripts
│   ├── task-init.py
│   ├── task-status.py
│   ├── check-task-packet.py
│   └── README.md
│
├── settings.json          # Hook configuration
└── README.md              # This file
```

## Components

### 1. Slash Commands (`commands/ai-pack/`)

Manual commands you invoke explicitly:

```bash
/ai-pack help              # Show all commands
/ai-pack task-init <name>  # Create task packet
/ai-pack task-status       # Check progress
/ai-pack orchestrate       # Assume Orchestrator role
/ai-pack engineer          # Assume Engineer role
/ai-pack review            # Start code review
/ai-pack test              # Validate tests
/ai-pack inspect           # Bug investigation
/ai-pack architect         # Architecture design
/ai-pack designer          # UX workflows
/ai-pack pm                # Product requirements
```

**Centralized namespace:** All commands under `/ai-pack` prefix.

### 2. Skills (`skills/`)

Auto-triggered behaviors based on keywords in your requests:

- **Orchestrator Skill** - Activates on "orchestrate", "coordinate", "delegate"
- **Engineer Skill** - Activates on "implement", "code", "build", "write"

Skills provide role-specific guidance automatically when Claude detects relevant patterns.

### 3. Rules (`rules/`)

Automatically loaded instructions for all files:

- **gates.md** - Mandatory quality gates (task packets, reviews, etc.)
- **task-packets.md** - Task packet requirements and lifecycle
- **workflows.md** - Workflow selection guide

Rules are context-specific and always active.

### 4. Hooks (`hooks/`)

Python scripts that enforce framework gates:

- **task-init.py** - Creates task packets
- **task-status.py** - Shows task progress
- **check-task-packet.py** - Blocks implementation without task packet

Hooks run automatically via Claude Code's hook system (configured in `settings.json`).

### 5. Settings (`settings.json`)

Configures hook behavior:

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [{
          "type": "command",
          "command": "python3 .claude/hooks/check-task-packet.py"
        }]
      }
    ]
  }
}
```

## How It Works

### Enforcement Layers

1. **Passive Documentation** (CLAUDE.md in project root)
   - Provides context and references
   - Loaded automatically by Claude Code

2. **Active Rules** (`.claude/rules/*.md`)
   - Auto-loaded based on file paths
   - Always enforced

3. **Auto-Triggered Skills** (`.claude/skills/*/SKILL.md`)
   - Activate based on keywords
   - Provide role-specific guidance

4. **Manual Commands** (`.claude/commands/ai-pack/*.md`)
   - Explicit invocation by user
   - Discoverable via `/ai-pack help`

5. **Hook Enforcement** (`.claude/hooks/*.py` + `settings.json`)
   - Blocks operations that violate gates
   - Provides immediate feedback

### Example Flow

**User:** "Implement the login feature"

**What happens:**
1. **Hook fires** - `check-task-packet.py` verifies task packet exists
2. **Skill activates** - Engineer skill provides TDD guidance
3. **Rules apply** - Gates and standards are enforced
4. **Commands available** - User can run `/ai-pack test` when ready

## Setup for New Projects

When adding ai-pack to a new project:

1. **Add ai-pack as submodule:**
   ```bash
   git submodule add <ai-pack-url> .ai-pack
   git submodule update --init
   ```

2. **Run setup script:**
   ```bash
   python3 .ai-pack/templates/.claude-setup.py
   ```

3. **Copy and customize CLAUDE.md:**
   ```bash
   cp .ai-pack/templates/CLAUDE.md .
   # Edit with project-specific context
   ```

4. **Commit integration:**
   ```bash
   git add .claude/ .ai/ CLAUDE.md
   git commit -m "Add ai-pack Claude Code integration"
   ```

## Customization

### Adding Project-Specific Commands

Create new commands in `.claude/commands/ai-pack/`:

```bash
.claude/commands/ai-pack/my-command.md
```

### Adding Project-Specific Rules

Create new rules in `.claude/rules/`:

```yaml
---
paths: **/*.py
---

# Python-specific rules
...
```

### Extending Skills

Add more skills in `.claude/skills/`:

```
.claude/skills/reviewer/SKILL.md
```

See individual README files in each directory for details.

## Maintenance

### Updating ai-pack Framework

```bash
# Update submodule to latest
git submodule update --remote .ai-pack

# Re-run setup to get new templates
python3 .ai-pack/templates/.claude-setup.py

# Commit updates
git add .ai-pack .claude/
git commit -m "Update ai-pack framework"
```

### Testing Setup

```bash
# Test task status
python3 .claude/hooks/task-status.py

# Test task init
python3 .claude/hooks/task-init.py test-task

# Test gate enforcement
echo '{"user_input": "implement login"}' | python3 .claude/hooks/check-task-packet.py
```

## References

- **AI-Pack Framework:** `.ai-pack/README.md`
- **Claude Code Docs:** https://code.claude.com/docs/
- **Slash Commands:** https://code.claude.com/docs/en/slash-commands.md
- **Skills:** https://code.claude.com/docs/en/skills.md
- **Hooks:** https://code.claude.com/docs/en/hooks.md
- **Rules:** https://code.claude.com/docs/en/memory.md#modular-rules

## Support

For issues with:
- **Framework design:** See `.ai-pack/` repository
- **Claude Code integration:** See individual component READMEs
- **Project-specific setup:** Check `.ai/repo-overrides.md`
