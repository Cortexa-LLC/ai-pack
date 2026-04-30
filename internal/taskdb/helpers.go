package taskdb

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateTaskID creates a unique task ID with format: {beads-id}-{timestamp}-{random}
// Example: "ai-pack-abc123-20260427-153045-a1b2c3"
func GenerateTaskID(beadsID string) string {
	timestamp := time.Now().Format("20060102-150405")
	randomBytes := make([]byte, 3)
	rand.Read(randomBytes)
	randomSuffix := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s-%s-%s", beadsID, timestamp, randomSuffix)
}

// ExtractShortID extracts the short task ID from a full task ID
// Example: "ai-pack-abc123-20260427-153045-a1b2c3" -> "ai-pack-abc123"
func ExtractShortID(taskID string) string {
	return extractBeadsID(taskID) // reuse existing helper
}
