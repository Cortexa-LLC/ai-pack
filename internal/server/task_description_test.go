package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadTaskDescriptionFromContract(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()
	projectRoot := tmpDir
	taskPacketPath := ".ai/tasks/test-123"

	// Create task packet directory
	contractDir := filepath.Join(projectRoot, taskPacketPath)
	if err := os.MkdirAll(contractDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	tests := []struct {
		name           string
		contractContent string
		expected       string
	}{
		{
			name: "simple description",
			contractContent: `# Task Contract

**Task ID:** test-123

## Task Description

Implement the foo feature.
This is a multi-line description.

## Acceptance Criteria

- Feature works
`,
			expected: "Implement the foo feature.\nThis is a multi-line description.",
		},
		{
			name: "description with empty lines",
			contractContent: `# Task Contract

## Task Description


Fix the bug in authentication.


## Notes

Some notes here.
`,
			expected: "Fix the bug in authentication.",
		},
		{
			name: "long description gets truncated",
			contractContent: `# Task Contract

## Task Description

` + string(make([]byte, 250)) + `This should be truncated because it exceeds 200 characters and we want to keep the UI clean.

## Next Section
`,
			expected: "", // Will be truncated to 200 chars + "..."
		},
		{
			name: "no task description section",
			contractContent: `# Task Contract

**Task ID:** test-456

## Background

Some background info.
`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write contract file
			contractPath := filepath.Join(contractDir, "00-contract.md")
			if err := os.WriteFile(contractPath, []byte(tt.contractContent), 0644); err != nil {
				t.Fatalf("failed to write contract file: %v", err)
			}

			// Create server instance
			s := &AgentServer{}

			// Read description
			result := s.readTaskDescriptionFromContract(taskPacketPath, projectRoot)

			// For long description test, just verify it was truncated
			if tt.name == "long description gets truncated" {
				if len(result) > 203 { // 200 + "..."
					t.Errorf("description not truncated: got length %d", len(result))
				}
				if len(result) > 0 && result[len(result)-3:] != "..." {
					t.Errorf("truncated description should end with '...', got: %s", result[len(result)-10:])
				}
				return
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestReadTaskDescriptionFromContract_MissingFile(t *testing.T) {
	s := &AgentServer{}
	result := s.readTaskDescriptionFromContract(".ai/tasks/nonexistent", "/tmp")
	if result != "" {
		t.Errorf("expected empty string for missing file, got %q", result)
	}
}
