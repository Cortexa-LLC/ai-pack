package commands

import (
	"github.com/cortexa-llc/ai-pack/internal/constants"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	agentclient "github.com/cortexa-llc/ai-pack/cmd/agent/client"
)

func runSpawn(role, taskInput string, wait, stream bool, inactiveTimeout time.Duration) error {
	// Validate task
	validateTaskOrExit(taskInput, role)

	fmt.Printf("🚀 Spawning %s agent...\n", role)
	fmt.Printf("🎯 task: %s\n", taskInput)

	projectRoot := detectProjectRoot()
	if projectRoot == "" {
		projectRoot = mustGetWorkingDir()
	}

	taskID, err := runSpawnHTTP(role, taskInput, projectRoot)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Agent spawned: %s\n", taskID)
	fmt.Printf("   Role: %s\n", role)
	fmt.Printf("   Beads: %s\n", taskInput)
	fmt.Println()

	if stream {
		if !streamTaskProgressWithInactivity(taskID, inactiveTimeout) {
			os.Exit(1)
		}
		return nil
	}

	if wait {
		waitForTaskCompletion(taskInput, 4*time.Hour)
		return nil
	}

	fmt.Printf("📋 Monitor: agent status %s\n", taskInput)
	fmt.Printf("📜 Logs:    agent logs %s\n", taskInput)
	fmt.Printf("⏳ Wait:    agent wait %s\n", taskInput)
	return nil
}

// runSpawnHTTP issues the POST /a2a/execute request and returns the task ID.
// It is extracted so that contract tests can redirect the HTTP call to an
// httptest.Server by setting agentclient.DefaultBaseURL.
func runSpawnHTTP(role, task, projectRoot string) (string, error) {
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "execute",
		"params": map[string]interface{}{
			"role":         role,
			"task":         task,
			"project_root": projectRoot,
		},
		"id": 1,
	}

	c := agentclient.Default()
	resp, err := c.PostJSON("/a2a/execute", requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to spawn agent: %w", err)
	}
	body := agentclient.RequireOK(resp, "spawn agent")

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for JSON-RPC error field first (server errors come back as HTTP 200
	// with an "error" object — RequireOK only catches non-200 status codes).
	if errObj, ok := result["error"].(map[string]interface{}); ok {
		msg, _ := errObj["message"].(string)
		data, _ := errObj["data"].(string)
		if data != "" {
			return "", fmt.Errorf("server error: %s: %s", msg, data)
		}
		return "", fmt.Errorf("server error: %s", msg)
	}

	taskID := ""
	if id, ok := result["task_id"].(string); ok {
		taskID = id
	} else if r, ok := result["result"].(map[string]interface{}); ok {
		if id, ok := r["task_id"].(string); ok {
			taskID = id
		}
	}

	if taskID == "" {
		return "", fmt.Errorf("server returned no task_id (raw response: %s)", string(body))
	}
	return taskID, nil
}

// validateTaskOrExit checks the task input looks like a task ID or a proper task string.
func validateTaskOrExit(taskInput, role string) {
	// If it looks like a common mistake (e.g. using a flag as the task)
	if strings.HasPrefix(taskInput, "-") {
		fmt.Printf("❌ Invalid task ID: %s\n", taskInput)
		fmt.Printf("   Usage: agent %s <task-id>\n", role)
		os.Exit(1)
	}
}

// detectProjectRoot looks for a git root or go.mod file.
func detectProjectRoot() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}
	// Fallback: look for go.mod
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for d := dir; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d
		}
		parent := filepath.Dir(d)
		if parent == d {
			break
		}
	}
	return ""
}

func mustGetWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}

// findInternalTaskIDFromServer asks the server for the internal task ID matching
// the given task ID.
func findInternalTaskIDFromServer(taskID string) string {
	c := agentclient.Default()
	id, _ := c.FindTaskByShortID(taskID)
	return id
}

// findTaskIDAndProjectFromServer returns (internalTaskID, projectRoot) from server.
func findTaskIDAndProjectFromServer(taskID string) (string, string) {
	c := agentclient.Default()
	return c.FindTaskByShortID(taskID)
}

// findInternalTaskID looks up the internal task ID from the local .beads directory.
func findInternalTaskID(taskID string) string {
	projectRoot := detectProjectRoot()
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}

	tasksDir := filepath.Join(projectRoot, constants.TaskRootDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		metaFile := filepath.Join(tasksDir, entry.Name(), "meta.json")
		data, err := os.ReadFile(metaFile)
		if err != nil {
			continue
		}
		var meta struct {
			BeadsTaskID string `json:"task_id"`
		}
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}
		if meta.BeadsTaskID == taskID {
			return entry.Name()
		}
	}
	return ""
}

// streamTaskProgressWithInactivity streams SSE events from the server.
// Returns true on successful completion, false on failure or cancellation.
func streamTaskProgressWithInactivity(internalTaskID string, inactiveTimeout time.Duration) bool {
	streamURL := fmt.Sprintf("%s/stream/%s", agentclient.DefaultBaseURL, internalTaskID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, nil)
	if err != nil {
		fmt.Printf("❌ Failed to build stream request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("❌ Failed to connect to stream: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Stream connection failed with status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Printf("✓ Connected to task stream: %s\n", internalTaskID)
	fmt.Printf("  (inactivity timeout: %v)\n", inactiveTimeout)
	fmt.Println()

	inactivityTimer := time.AfterFunc(inactiveTimeout, func() {
		fmt.Printf("\n⏰ Stream inactive for %v — disconnecting\n", inactiveTimeout)
		cancel()
	})
	defer inactivityTimer.Stop()

	buffer := make([]byte, 8192)
	dataBuffer := ""

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			inactivityTimer.Reset(inactiveTimeout)
			dataBuffer += string(buffer[:n])

			for {
				lineEnd := strings.Index(dataBuffer, "\n")
				if lineEnd < 0 {
					break
				}
				line := dataBuffer[:lineEnd]
				dataBuffer = dataBuffer[lineEnd+1:]
				line = strings.TrimRight(line, "\r")

				if data, ok := agentclient.ParseSSELine(line); ok {
					if data == "[DONE]" {
						fmt.Println("\n✅ Stream complete")
						return true
					}
					// Try to parse as JSON event.
					// Server sends StreamEvent{Type, TaskID, Timestamp, Data} so the
					// terminal signal is in the "type" field, not "status".
					var event map[string]interface{}
					if jsonErr := json.Unmarshal([]byte(data), &event); jsonErr == nil {
						if msg, ok := event["message"].(string); ok {
							fmt.Print(msg)
						} else if eventType, ok := event["type"].(string); ok {
							switch eventType {
							case "completed":
								fmt.Println("\n✅ Agent completed!")
								showMetrics()
								return true
							case "failed":
								fmt.Println("\n❌ Agent failed!")
								return false
							}
						}
					} else {
						// Plain text data
						fmt.Print(data)
					}
				}
				// SSE comments (heartbeats) like ": heartbeat" are intentionally ignored
			}
		}
		if err != nil {
			if ctx.Err() != nil {
				// Cancelled by inactivity timer — task may still be running on server.
				return false
			}
			if err.Error() == "EOF" {
				fmt.Println("\n✅ Stream ended")
				return true
			}
			fmt.Printf("\n⚠️  Stream error: %v\n", err)
			return false
		}
	}
}

// waitForTaskCompletion polls the server until the task is done or timeout expires.
func waitForTaskCompletion(taskID string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	fmt.Printf("⏳ Waiting for task %s to complete (timeout: %v)...\n", taskID, timeout)

	for time.Now().Before(deadline) {
		internalTaskID := findInternalTaskIDFromServer(taskID)
		if internalTaskID == "" {
			internalTaskID = findInternalTaskID(taskID)
		}
		if internalTaskID == "" {
			fmt.Printf("⚠️  Task not found yet, retrying...\n")
			time.Sleep(5 * time.Second)
			continue
		}

		c := agentclient.Default()
		resp, err := c.Get(fmt.Sprintf("/a2a/tasks/%s", internalTaskID))
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		body, _ := agentclient.ReadBody(resp)
		var status map[string]interface{}
		if err := json.Unmarshal(body, &status); err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		statusVal, ok := status["status"]
		if !ok || statusVal == nil {
			time.Sleep(2 * time.Second)
			continue
		}

		statusStr, ok := statusVal.(string)
		if !ok {
			time.Sleep(2 * time.Second)
			continue
		}

		switch statusStr {
		case "completed":
			fmt.Println("✅ Agent completed!")
			showMetrics()
			return
		case "failed":
			fmt.Println("❌ Agent failed!")
			os.Exit(1)
		default:
			fmt.Printf("  Status: %s\n", statusStr)
			time.Sleep(10 * time.Second)
		}
	}
	fmt.Printf("⏰ Timeout waiting for task %s\n", taskID)
	os.Exit(1)
}

// formatNumber formats a large integer with commas.
func formatNumber(n int64) string {
	s := fmt.Sprintf("%d", n)
	result := ""
	for i, c := range reverseString(s) {
		if i > 0 && i%3 == 0 {
			result = "," + result
		}
		result = string(c) + result
	}
	return result
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// truncateDescription returns only the first line of desc, truncated to maxLen.
func truncateDescription(desc string, maxLen int) string {
	lines := strings.Split(desc, "\n")
	firstLine := strings.TrimSpace(lines[0])
	if len(firstLine) > maxLen {
		return firstLine[:maxLen-3] + "..."
	}
	return firstLine
}

// wrapText wraps text at width, indenting continuation lines with indent.
func wrapText(text string, width int, indent string) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return indent + "\n"
	}

	var result strings.Builder
	result.WriteString(indent)
	lineLen := 0

	for i, word := range words {
		wordLen := len(word)
		if i > 0 {
			if lineLen+wordLen+1 > width {
				result.WriteString("\n" + indent)
				lineLen = 0
			} else {
				result.WriteString(" ")
				lineLen++
			}
		}
		result.WriteString(word)
		lineLen += wordLen
	}
	result.WriteString("\n")
	return result.String()
}

// minInt returns the smaller of a and b.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getRuntimeInfo returns OS/arch info.
func getRuntimeInfo() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}
