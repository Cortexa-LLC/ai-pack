package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	var force bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "delete <task-id>",
		Short: "Delete a task from the database",
		Long: `Delete a task from the task database.

WARNING: This permanently removes the task from the database.
The task execution directory (if any) will remain.

Example:
  agent delete ai-pack-abc
  agent delete ai-pack-abc --force    # Skip confirmation
  agent delete ai-pack-abc --json     # JSON output`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			// Confirmation prompt unless --force
			if !force {
				fmt.Printf("⚠️  Delete task %s? This cannot be undone. (y/N): ", taskID)
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			// Send DELETE request
			serverURL := getServerURL()
			req, err := http.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("%s/a2a/tasks/%s", serverURL, taskID),
				nil,
			)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("failed to delete task: %w", err)
			}
			defer resp.Body.Close()

			// Read response body
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

			if jsonOutput {
				output, _ := json.Marshal(result)
				fmt.Println(string(output))
			} else {
				fmt.Printf("✅ Task deleted: %s\n", taskID)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
