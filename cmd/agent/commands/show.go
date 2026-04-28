package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show <task-id>",
		Short: "Display detailed task information",
		Long: `Display detailed information about a task including its status, timestamps,
result, and error messages.

Example:
  agent show ai-pack-a1b2
  agent show ai-pack-xyz --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			// Query server for task details
			serverURL := getServerURL()
			resp, err := http.Get(fmt.Sprintf("%s/a2a/tasks/%s", serverURL, taskID))
			if err != nil {
				return fmt.Errorf("failed to query task: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				return fmt.Errorf("task not found: %s", taskID)
			}

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
			}

			// Parse response
			var task map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(task, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			// Format human-readable output
			fmt.Printf("Task Details\n")
			fmt.Printf("════════════\n\n")
			fmt.Printf("Task ID:       %s\n", task["id"])
			fmt.Printf("Status:        %s\n", formatStatus(task["status"]))

			if role, ok := task["role"].(string); ok && role != "" {
				fmt.Printf("Role:          %s\n", role)
			}

			fmt.Printf("Description:   %s\n", task["description"])

			if projectRoot, ok := task["project_root"].(string); ok && projectRoot != "" {
				fmt.Printf("Project Root:  %s\n", projectRoot)
			}

			fmt.Printf("\nTimestamps:\n")
			if createdAt, ok := task["created_at"].(string); ok {
				t, _ := time.Parse(time.RFC3339, createdAt)
				fmt.Printf("  Created:     %s\n", t.Format("2006-01-02 15:04:05"))
			}
			if startedAt, ok := task["started_at"].(string); ok && startedAt != "" {
				t, _ := time.Parse(time.RFC3339, startedAt)
				fmt.Printf("  Started:     %s\n", t.Format("2006-01-02 15:04:05"))
			}
			if completedAt, ok := task["completed_at"].(string); ok && completedAt != "" {
				t, _ := time.Parse(time.RFC3339, completedAt)
				fmt.Printf("  Completed:   %s\n", t.Format("2006-01-02 15:04:05"))
			}

			if result, ok := task["result"].(string); ok && result != "" {
				fmt.Printf("\nResult:\n%s\n", result)
			}

			if errMsg, ok := task["error"].(string); ok && errMsg != "" {
				fmt.Printf("\nError:\n%s\n", errMsg)
			}

			if metadata, ok := task["metadata"].(string); ok && metadata != "" && metadata != "{}" {
				fmt.Printf("\nMetadata:\n%s\n", metadata)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func formatStatus(status interface{}) string {
	s, ok := status.(string)
	if !ok {
		return "unknown"
	}

	switch s {
	case "queued":
		return "⏳ queued"
	case "in_progress":
		return "🔄 in_progress"
	case "completed":
		return "✅ completed"
	case "failed":
		return "❌ failed"
	case "cancelled":
		return "❓ cancelled"
	default:
		return s
	}
}
