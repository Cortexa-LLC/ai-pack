package hello

import "testing"

// TestHelloWorld verifies that the HelloWorld function returns the expected greeting
func TestHelloWorld(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string returns default greeting",
			input:    "",
			expected: "Hello, World!",
		},
		{
			name:     "with name returns personalized greeting",
			input:    "Alice",
			expected: "Hello, Alice!",
		},
		{
			name:     "with name handles whitespace",
			input:    "  Bob  ",
			expected: "Hello, Bob!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HelloWorld(tt.input)
			if result != tt.expected {
				t.Errorf("HelloWorld(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
