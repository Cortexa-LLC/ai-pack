package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var status string
	var result string
	var metadata string

	cmd := &cobra.Command{
		Use:   "update <task-id>",
		Short: "Update task fields",
		Long: `Update one or more fields of a task.

You can update the status, result message, or metadata (as JSON).

Example:
  agent update ai-pack-a1b2 --status in_progress
  agent update ai-pack-xyz --result "Partial completion"
  agent update ai-pack-abc --metadata '{"priority": "P0"}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			// Build request payload with only specified fields
			reqBody := make(map[string]interface{})

			if status != "" {
				// Validate status
				validStatuses := map[string]bool{
					"queued":      true,
					"in_progress": true,
					"completed":   true,
					"failed":      true,
					"cancelled":   true,
				}
				if !validStatuses[status] {
					return fmt.Errorf("invalid status: %s (must be one of: queued, in_progress, completed, failed, cancelled)", status)
				}
				reqBody["status"] = status
			}

			if result != "" {
				reqBody["result"] = result
			}

			if metadata != "" {
				// Validate JSON
				var metadataObj interface{}
				if err := json.Unmarshal([]byte(metadata), &metadataObj); err != nil {
					return fmt.Errorf("invalid JSON metadata: %w", err)
				}
				reqBody["metadata"] = metadata
			}

			if len(reqBody) == 0 {
				return fmt.Errorf("no fields specified to update (use --status, --result, or --metadata)")
			}

			// Send request to server
			serverURL := getServerURL()
			reqJSON, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			req, err := http.NewRequest(
				http.MethodPut,
				fmt.Sprintf("%s/a2a/tasks/%s", serverURL, taskID),
				bytes.NewReader(reqJSON),
			)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("failed to update task: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				return fmt.Errorf("task not found: %s", taskID)
			}

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
			}

			fmt.Printf("✅ Task %s updated successfully\n", taskID)

			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Update status (queued, in_progress, completed, failed, cancelled)")
	cmd.Flags().StringVar(&result, "result", "", "Set result message")
	cmd.Flags().StringVar(&metadata, "metadata", "", "Update metadata (JSON string)")

	return cmd
}
