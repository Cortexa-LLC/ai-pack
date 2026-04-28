# Agent Configurations

This directory contains shared agent configuration files used by all execution methods (Task tool, A2A server, etc.).

## Structure

```
agents/
├── engineer.yml        # Implementation specialist
├── tester.yml          # Testing specialist
├── reviewer.yml        # Code review specialist
├── architect.yml       # System architecture specialist
├── product-manager.yml    # Requirements and user stories
└── README.md          # This file
```

## Configuration Format

Each `.yml` file defines:

- **name**: Agent identifier
- **description**: Brief role summary
- **delegation**: Execution parameters
  - `mode`: How the agent operates (e.g., `delegate`)
  - `timeout`: Maximum execution time (e.g., `10min`, `15min`)
  - `max_context`: Maximum context window size
- **tools**: Permitted tool access
  - `read`, `write`, `edit` - File operations
  - `bash` - Command execution
  - `grep`, `glob` - Code search
  - `webfetch` - Web access
- **context**: Role behavior
  - `role_file`: Path to role definition (e.g., `roles/engineer.md`)
  - `gates`: Quality gates to enforce
  - `additional_instructions`: Extra guidance
- **success_criteria**: What defines successful completion
- **metadata**: Version and compatibility info

## Role Definitions vs Agent Configs

**Agent Config (.yml)** - Technical execution parameters:
- Which tools the agent can use
- Timeout settings
- Which role.md file to load
- Which gates to enforce

**Role Definition (.md)** - Agent behavior and instructions:
- How the agent should think and act
- Workflow steps and processes
- Best practices and examples
- Quality standards

**Example Flow:**
```
1. Load agent config: agents/engineer.yml
   → timeout: 10min
   → tools: [read, write, edit, bash, grep, glob]
   → role_file: roles/engineer.md

2. Load role definition: roles/engineer.md
   → "You are an Engineer. Follow TDD..."
   → Detailed instructions and examples

3. Inject role into system prompt
4. Execute with configured tools and timeout
```

## Execution Methods

These configs work with multiple execution methods:

### 1. Task Tool (Lightweight)
- Claude Code's built-in Task tool
- Sequential execution
- Foreground only

### 2. A2A Server (Production)
- Go-based dedicated server
- Parallel execution
- SSE streaming
- Background execution

Both use the same config files - no duplication needed!

## Project-Specific Overrides

Adopter projects can override agent configurations without modifying the submodule:

### Override Pattern

```
your-project/
├── .ai-pack/              # Read-only submodule
│   └── agents/
│       └── engineer.yml   # Framework default
└── .ai/                   # Project-specific (writable)
    └── agents/
        └── engineer.yml   # Your override (takes precedence)
```

### Creating an Override

1. **Create directory**: `mkdir -p .ai/agents`
2. **Copy framework config**: `cp .ai-pack/agents/engineer.yml .ai/agents/engineer.yml`
3. **Modify as needed**:
   ```yaml
   name: engineer
   delegation:
     timeout: 20min  # Increased for complex project
   tools:
     - read
     - write
     - edit
     - bash
     - grep
     - glob
     - webfetch    # Added for API docs
   context:
     role_file: .ai/roles/engineer.md  # Use extended role
     gates:
       - tdd-enforcement
       - code-quality-review
       - custom-security-scan  # Project-specific gate
   ```

### Resolution Order

1. **First**: Check `.ai/agents/<role>.yml` (project override)
2. **Fallback**: Use `.ai-pack/agents/<role>.yml` (framework default)

This allows projects to:
- Adjust timeouts for their needs
- Add project-specific tools
- Use extended role definitions (`.ai/roles/`)
- Add custom quality gates
- Modify success criteria

## Adding New Agents

To add a new agent:

1. **Create role definition**: `roles/newagent.md`
   - Define responsibilities, workflow, best practices

2. **Create agent config**: `agents/newagent.yml`
   - Set tools, timeout, gates
   - Reference role file: `role_file: roles/newagent.md`

3. **Use the agent**:
   ```bash
   # With Beads
   agent create "Task description"
   agent newagent bd-xxxx

   # Direct (deprecated)
   agent newagent "task description"
   ```

## Available Agents

| Agent | Purpose | Typical Timeout | Key Tools |
|-------|---------|----------------|-----------|
| **engineer** | Implementation | 10min | read, write, edit, bash, grep, glob |
| **tester** | Testing | 10min | read, write, edit, bash, grep, glob |
| **reviewer** | Code review | 10min | read, grep, glob, bash |
| **architect** | System design | 15min | read, write, edit, bash, grep, glob, webfetch |
| **product-manager** | Requirements | 15min | read, write, edit, bash, grep, glob, webfetch |

## Best Practices

1. **Keep configs DRY** - Use shared role files, don't duplicate instructions
2. **Minimal tools** - Only grant tools the agent actually needs
3. **Reasonable timeouts** - Balance completion vs. cost
4. **Clear success criteria** - Make it measurable
5. **Document gates** - Explain what quality checks apply

## Version Compatibility

All agent configs include metadata indicating compatibility:

```yaml
metadata:
  version: "2.0"
  compatible_with:
    - lightweight  # Task tool execution
    - a2a          # A2A server execution
```

This ensures configs work across execution methods.

---

**Last Updated**: 2026-01-24
**Version**: 2.0.0
