#!/usr/bin/env python3
"""
Generate agent configuration files from role definitions.

Creates .ai-pack/agents/lightweight/*.yml configs based on roles/ definitions.
"""

import os
from pathlib import Path
from typing import Dict, List

# Agent configuration templates based on role characteristics
AGENT_CONFIGS = {
    "engineer": {
        "name": "engineer",
        "description": "Implementation specialist following TDD and code quality standards",
        "tier": "lightweight",
        "context": {
            "role_file": "roles/engineer.md",
            "gates": ["tdd-enforcement", "code-quality-review", "beads-enforcement"]
        },
        "delegation": {
            "mode": "delegate",
            "timeout": "5min"
        },
        "tools": ["Read", "Write", "Edit", "Bash", "Grep", "Glob"],
        "output": {
            "format": "task-packet",
            "artifacts": ["code", "tests"]
        }
    },
    "reviewer": {
        "name": "reviewer",
        "description": "Code review and quality assurance specialist",
        "tier": "lightweight",
        "context": {
            "role_file": "roles/reviewer.md",
            "gates": ["code-quality-review", "security-review", "beads-enforcement"]
        },
        "delegation": {
            "mode": "delegate",
            "timeout": "5min"
        },
        "tools": ["Read", "Grep", "Glob", "Bash"],
        "output": {
            "format": "task-packet",
            "artifacts": ["review-report", "recommendations"]
        }
    },
    "tester": {
        "name": "tester",
        "description": "Test creation and validation specialist",
        "tier": "lightweight",
        "context": {
            "role_file": "roles/tester.md",
            "gates": ["test-coverage", "tdd-enforcement", "beads-enforcement"]
        },
        "delegation": {
            "mode": "delegate",
            "timeout": "5min"
        },
        "tools": ["Read", "Write", "Edit", "Bash", "Grep", "Glob"],
        "output": {
            "format": "task-packet",
            "artifacts": ["tests", "test-report"]
        }
    },
    "architect": {
        "name": "architect",
        "description": "System design and architecture specialist",
        "tier": "lightweight",
        "context": {
            "role_file": "roles/architect.md",
            "gates": ["architecture-review", "beads-enforcement"]
        },
        "delegation": {
            "mode": "delegate",
            "timeout": "10min"
        },
        "tools": ["Read", "Grep", "Glob", "Task"],
        "output": {
            "format": "task-packet",
            "artifacts": ["architecture-doc", "design-decisions"]
        }
    },
    "cartographer": {
        "name": "cartographer",
        "description": "Codebase exploration and mapping specialist",
        "tier": "lightweight",
        "context": {
            "role_file": "roles/cartographer.md",
            "gates": ["beads-enforcement"]
        },
        "delegation": {
            "mode": "delegate",
            "timeout": "10min"
        },
        "tools": ["Read", "Grep", "Glob", "Task"],
        "output": {
            "format": "task-packet",
            "artifacts": ["codebase-map", "analysis-report"]
        }
    },
    "archaeologist": {
        "name": "archaeologist",
        "description": "Historical code analysis and investigation specialist",
        "tier": "lightweight",
        "context": {
            "role_file": "roles/archaeologist.md",
            "gates": ["beads-enforcement"]
        },
        "delegation": {
            "mode": "delegate",
            "timeout": "10min"
        },
        "tools": ["Read", "Grep", "Glob", "Bash", "Task"],
        "output": {
            "format": "task-packet",
            "artifacts": ["investigation-report", "findings"]
        }
    },
    "spelunker": {
        "name": "spelunker",
        "description": "Deep code investigation and exploration specialist",
        "tier": "lightweight",
        "context": {
            "role_file": "roles/spelunker.md",
            "gates": ["beads-enforcement"]
        },
        "delegation": {
            "mode": "delegate",
            "timeout": "10min"
        },
        "tools": ["Read", "Grep", "Glob", "Task"],
        "output": {
            "format": "task-packet",
            "artifacts": ["exploration-report", "insights"]
        }
    },
    "inspector": {
        "name": "inspector",
        "description": "Code quality and compliance inspection specialist",
        "tier": "lightweight",
        "context": {
            "role_file": "roles/inspector.md",
            "gates": ["quality-inspection", "beads-enforcement"]
        },
        "delegation": {
            "mode": "delegate",
            "timeout": "5min"
        },
        "tools": ["Read", "Grep", "Glob", "Bash"],
        "output": {
            "format": "task-packet",
            "artifacts": ["inspection-report", "violations"]
        }
    },
    "strategist": {
        "name": "strategist",
        "description": "Strategic planning and roadmap specialist",
        "tier": "lightweight",
        "context": {
            "role_file": "roles/strategist.md",
            "gates": ["beads-enforcement"]
        },
        "delegation": {
            "mode": "delegate",
            "timeout": "15min"
        },
        "tools": ["Read", "Grep", "Glob", "Task"],
        "output": {
            "format": "task-packet",
            "artifacts": ["strategy-doc", "roadmap"]
        }
    },
    "designer": {
        "name": "designer",
        "description": "UX/UI design and user experience specialist",
        "tier": "lightweight",
        "context": {
            "role_file": "roles/designer.md",
            "gates": ["design-review", "beads-enforcement"]
        },
        "delegation": {
            "mode": "delegate",
            "timeout": "10min"
        },
        "tools": ["Read", "Write", "Edit", "Grep", "Glob"],
        "output": {
            "format": "task-packet",
            "artifacts": ["design-doc", "mockups"]
        }
    }
}


def verify_role_exists(role_name: str, roles_dir: Path) -> bool:
    """Check if a role definition file exists."""
    role_file = roles_dir / f"{role_name}.md"
    return role_file.exists()


def dict_to_yaml(data: Dict, indent: int = 0) -> str:
    """Convert dict to YAML format (simple implementation)."""
    lines = []
    for key, value in data.items():
        prefix = "  " * indent
        if isinstance(value, dict):
            lines.append(f"{prefix}{key}:")
            lines.append(dict_to_yaml(value, indent + 1))
        elif isinstance(value, list):
            lines.append(f"{prefix}{key}:")
            for item in value:
                lines.append(f"{prefix}  - {item}")
        else:
            lines.append(f"{prefix}{key}: {value}")
    return "\n".join(lines)


def generate_agent_config(role_name: str, config: Dict, output_dir: Path) -> None:
    """Generate a single agent configuration file."""
    output_file = output_dir / f"{role_name}.yml"

    # Create YAML with nice formatting
    yaml_content = dict_to_yaml(config)

    with open(output_file, 'w') as f:
        f.write(f"# Agent Configuration: {role_name}\n")
        f.write(f"# Auto-generated from roles/{role_name}.md\n")
        f.write(f"# DO NOT EDIT - regenerate using scripts/generate-agent-configs.py\n\n")
        f.write(yaml_content)

    print(f"✅ Generated: {output_file}")


def main():
    """Generate all agent configuration files."""
    # Determine project root (script is in scripts/)
    script_dir = Path(__file__).parent
    project_root = script_dir.parent

    roles_dir = project_root / "roles"
    output_dir = project_root / ".ai-pack" / "agents" / "lightweight"

    # Create output directory
    output_dir.mkdir(parents=True, exist_ok=True)

    print(f"\n🔧 Generating agent configurations...")
    print(f"   Roles directory: {roles_dir}")
    print(f"   Output directory: {output_dir}\n")

    # Track success/failures
    generated = []
    skipped = []

    # Generate configs for roles that exist
    for role_name, config in AGENT_CONFIGS.items():
        if verify_role_exists(role_name, roles_dir):
            generate_agent_config(role_name, config, output_dir)
            generated.append(role_name)
        else:
            print(f"⚠️  Skipping {role_name} (role file not found)")
            skipped.append(role_name)

    # Summary
    print(f"\n📊 Summary:")
    print(f"   Generated: {len(generated)} configs")
    print(f"   Skipped: {len(skipped)} configs")

    if generated:
        print(f"\n✅ Generated configs for: {', '.join(generated)}")

    if skipped:
        print(f"\n⚠️  Skipped (missing role files): {', '.join(skipped)}")

    print(f"\n📁 Configuration directory: {output_dir}")
    print(f"   Use these configs with: ./bin/agent-server")


if __name__ == "__main__":
    main()
