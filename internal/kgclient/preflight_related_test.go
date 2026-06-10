package kgclient

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRelatedProjects(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "no related projects",
			content:  "# Contract\n\n**Task:** Do something\n",
			expected: nil,
		},
		{
			name:     "single related project",
			content:  "# Contract\n\nRelated Projects: /home/user/other-project\n",
			expected: []string{"/home/user/other-project"},
		},
		{
			name: "multiple related projects on separate lines",
			content: "# Contract\n\n" +
				"Related Projects: /projects/foo\n" +
				"Related Projects: /projects/bar\n",
			expected: []string{"/projects/foo", "/projects/bar"},
		},
		{
			name:     "related project with surrounding content",
			content:  "# Contract\n\n**Task:** Patch\n\nRelated Projects: /opt/project-x\n\n## Acceptance\n",
			expected: []string{"/opt/project-x"},
		},
		{
			name:     "related projects line with no path (blank after colon)",
			content:  "Related Projects:  \n",
			expected: nil,
		},
		{
			name:     "unrelated line with similar prefix",
			content:  "Not Related Projects: /some/path\n",
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Write to a temp file (using task.md format)
			dir := t.TempDir()
			path := filepath.Join(dir, "task.md")
			if err := os.WriteFile(path, []byte(tc.content), 0o644); err != nil {
				t.Fatalf("failed to write temp file: %v", err)
			}

			got := ParseRelatedProjects(path)

			if len(got) != len(tc.expected) {
				t.Errorf("got %v, want %v", got, tc.expected)
				return
			}
			for i := range got {
				if got[i] != tc.expected[i] {
					t.Errorf("got[%d] = %q, want %q", i, got[i], tc.expected[i])
				}
			}
		})
	}
}

func TestParseRelatedProjectsMissingFile(t *testing.T) {
	// Should return nil gracefully, not panic.
	result := ParseRelatedProjects("/nonexistent/path/task.md")
	if result != nil {
		t.Errorf("expected nil for missing file, got %v", result)
	}
}
