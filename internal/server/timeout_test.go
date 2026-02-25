package server

import (
	"testing"
	"time"
)

func TestParseRoleTimeout(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		// Standard Go duration strings
		{"10m", 10 * time.Minute},
		{"1h", time.Hour},
		{"30s", 30 * time.Second},
		{"1h30m", 90 * time.Minute},

		// Human-friendly long-form suffixes
		{"10min", 10 * time.Minute},
		{"30sec", 30 * time.Second},
		{"2min", 2 * time.Minute},

		// Edge cases
		{"", defaultRoleTimeout},
		{"garbage", defaultRoleTimeout},
		{"0m", defaultRoleTimeout},    // zero duration falls back
		{"-5min", defaultRoleTimeout}, // negative duration falls back
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := parseRoleTimeout(tc.input)
			if got != tc.want {
				t.Errorf("parseRoleTimeout(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
