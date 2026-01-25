package beads

import (
	"testing"
)

func TestIsBeadsTaskID(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"bd-a1b2", true},
		{"bd-x7z9", true},
		{"bd-", false}, // Too short
		{"bd", false},
		{"create hello world", false},
		{"implement feature", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsBeadsTaskID(tt.input)
			if result != tt.expected {
				t.Errorf("IsBeadsTaskID(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsInstalled(t *testing.T) {
	// This test will pass or fail based on whether bd is installed
	// We just verify it doesn't crash
	installed := IsInstalled()
	t.Logf("Beads installed: %v", installed)
}

func TestClient_GetTaskDescription(t *testing.T) {
	client := NewClient()

	// Test free-form description
	desc, isBeads, err := client.GetTaskDescription("create hello world")
	if err != nil {
		t.Errorf("Unexpected error for free-form: %v", err)
	}
	if isBeads {
		t.Error("Expected isBeads=false for free-form description")
	}
	if desc != "create hello world" {
		t.Errorf("Expected description='create hello world', got '%s'", desc)
	}
}

// Note: TestClient_GetTask, TestClient_StartTask, etc. would require
// actual Beads installation and test tasks. Those are integration tests
// and should be run separately in a test environment with Beads installed.
