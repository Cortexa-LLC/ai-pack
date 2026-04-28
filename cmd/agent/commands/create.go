package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	var priority string
	var role string
	var projectRoot string
	var packetPath string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "create <description>",
		Short: "Create a new agent task",
		Long: `Create a new agent task in the task database.

The task will be assigned a unique ID and stored in the SQLite database.
You can then spawn an agent to work on it using:
  agent <role> <task-id> --stream

Example:
  agent create "Fix authentication bug" --priority P1 --role engineer
  agent create "Add user registration" --packet .ai/tasks/reg-task/`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			description := args[0]

			// Resolve project root to absolute path
			absProjectRoot, err := filepath.Abs(projectRoot)
			if err != nil {
				return fmt.Errorf("failed to resolve project root: %w", err)
			}

			// Build request payload
			reqBody := map[string]interface{}{
				"description":  description,
				"project_root": absProjectRoot,
			}

			if role != "" {
				reqBody["role"] = role
			}

			// Store priority and packet path in metadata
			metadata := make(map[string]string)
			if priority != "" {
				metadata["priority"] = priority
			}
			if packetPath != "" {
				metadata["task_packet"] = packetPath
			}
			if len(metadata) > 0 {
				metadataJSON, _ := json.Marshal(metadata)
				reqBody["metadata"] = string(metadataJSON)
			}

			// Send request to server
			serverURL := getServerURL()
			reqJSON, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			resp, err := http.Post(
				fmt.Sprintf("%s/a2a/tasks/create", serverURL),
				"application/json",
				bytes.NewReader(reqJSON),
			)
			if err != nil {
				return fmt.Errorf("failed to create task: %w", err)
			}
			defer resp.Body.Close()

			// Read response body first
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
			}

			// Parse response
			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			taskID, ok := result["task_id"].(string)
			if !ok {
				return fmt.Errorf("server did not return task_id")
			}

			if jsonOutput {
				output, _ := json.Marshal(result)
				fmt.Println(string(output))
			} else {
				fmt.Printf("✅ Task created: %s\n", taskID)
				if role != "" {
					fmt.Printf("\nNext steps:\n")
					fmt.Printf("  agent %s %s --stream\n", role, taskID)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&priority, "priority", "p", "P2", "Task priority (P0-P4)")
	cmd.Flags().StringVarP(&role, "role", "r", "", "Assign to role (engineer, reviewer, tester, etc.)")
	cmd.Flags().StringVar(&projectRoot, "project", ".", "Project root directory")
	cmd.Flags().StringVar(&packetPath, "packet", "", "Task packet directory path")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
