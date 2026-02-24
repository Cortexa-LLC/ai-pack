package proxy

import (
	"testing"
)

func TestExtractHost(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "HTTPS URL with path",
			url:      "https://proxy.example.com/gateway/v1/anthropic",
			expected: "proxy.example.com",
		},
		{
			name:     "HTTP URL with path",
			url:      "http://localhost:8080/api/v1",
			expected: "localhost:8080",
		},
		{
			name:     "URL without path",
			url:      "https://api.anthropic.com",
			expected: "api.anthropic.com",
		},
		{
			name:     "No scheme",
			url:      "example.com/path",
			expected: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractHost(tt.url)
			if result != tt.expected {
				t.Errorf("extractHost(%s) = %s, want %s", tt.url, result, tt.expected)
			}
		})
	}
}

func TestExtractPath(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "HTTPS URL with path",
			url:      "https://proxy.example.com/gateway/v1/anthropic",
			expected: "/gateway/v1/anthropic",
		},
		{
			name:     "HTTP URL with path",
			url:      "http://localhost:8080/api/v1",
			expected: "/api/v1",
		},
		{
			name:     "URL without path",
			url:      "https://api.anthropic.com",
			expected: "",
		},
		{
			name:     "URL with trailing slash",
			url:      "https://example.com/",
			expected: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPath(tt.url)
			if result != tt.expected {
				t.Errorf("extractPath(%s) = %s, want %s", tt.url, result, tt.expected)
			}
		})
	}
}
