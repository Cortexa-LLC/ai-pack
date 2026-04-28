package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

func newCloseCmd() *cobra.Command {
	var result string

	cmd := &cobra.Command{
		Use:   "close <task-id>",
		Short: "Mark a task as completed",
		Long: `Mark a task as completed with an optional result message.

This updates the task status to 'completed' and sets the completion timestamp.

Example:
  agent close ai-pack-a1b2
  agent close ai-pack-xyz --result "Successfully implemented feature"
  agent close ai-pack-abc -r "Bug fixed and tests passing"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			// Build request payload
			reqBody := map[string]interface{}{
				"result": result,
			}

			// Send request to server
			serverURL := getServerURL()
			reqJSON, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			req, err := http.NewRequest(
				http.MethodPut,
				fmt.Sprintf("%s/a2a/tasks/%s/close", serverURL, taskID),
				bytes.NewReader(reqJSON),
			)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("failed to close task: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				return fmt.Errorf("task not found: %s", taskID)
			}

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
			}

			fmt.Printf("✅ Task %s marked as completed\n", taskID)
			if result != "" {
				fmt.Printf("   Result: %s\n", result)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&result, "result", "r", "", "Completion message or result")

	return cmd
}
