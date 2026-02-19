# Orchestrator-Engineer Handoff Protocol

**Purpose:** To provide a clear structure for orchestrators when spawning engineers on well-scoped tasks. This ensures smooth, effective execution without unnecessary delays due to vague instructions.

## Pre-Cooked Brief Format
When creating a task assignation for engineers, the following elements must be included:
1. **Exact File Paths:** Specify the exact files the engineer will be working on.
2. **Specific Changes:** Outline the specific changes required, avoiding vague descriptions.
3. **Explicit Context Signal:** Clearly state, **"All context provided"** to indicate that supporting documents have been reviewed and are not necessary for the engineer.
4. **Acceptance Criteria:** Express these as shell commands that need to pass (e.g., `go build ./... must pass`).

### Example of a Well-Structured Handoff Brief:
```plaintext
Task: Implement fix for multi-turn tool use.

Working Directory: /Users/bryanw/Projects/Vibe/ai-pack/

File Paths:
- `file_path_a.go`
- `file_path_b.go`

Specific Changes:
- Modify method XYZ in `file_path_a.go` to improve its logic based on previous feedback.
- Update corresponding methods in `file_path_b.go` as needed.

All context provided.

Acceptance Criteria:
- Ensure `go build ./...` passes without errors.
```