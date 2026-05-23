package taskdb

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// GenerateTaskID creates a unique task ID with format: {short-task-id}-{timestamp}-{random}
// Example: "ai-pack-abc123-20260427-153045-a1b2c3"
func GenerateTaskID(shortTaskID string) string {
	timestamp := time.Now().Format("20060102-150405")
	randomBytes := make([]byte, 3)
	rand.Read(randomBytes)
	randomSuffix := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s-%s-%s", shortTaskID, timestamp, randomSuffix)
}

// ExtractShortID extracts the short task ID from a full task ID
// Example: "ai-pack-abc123-20260427-153045-a1b2c3" -> "ai-pack-abc123"
func ExtractShortID(taskID string) string {
	return extractShortTaskID(taskID)
}

// extractShortTaskID splits by hyphen and returns the part before the timestamp.
func extractShortTaskID(taskID string) string {
	parts := strings.Split(taskID, "-")
	if len(parts) < 3 {
		return taskID
	}
	// Find the first part that looks like a timestamp (YYYYMMDD)
	for i := 2; i < len(parts); i++ {
		if len(parts[i]) == 8 && isNumeric(parts[i]) {
			return strings.Join(parts[:i], "-")
		}
	}
	return taskID
}

// isNumeric returns true if every character in s is a digit.
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
