package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	agentclient "github.com/cortexa-llc/ai-pack/cmd/agent/client"
	"github.com/spf13/cobra"
)

func newDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "diff [task-id]",
		Short:        "Show git diff for changed files",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiff()
		},
	}
}

func runDiff() error {
	c := exec.Command("git", "diff")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func newFilesCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "files [task-id]",
		Short:        "Show modified files (git status --short)",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFiles()
		},
	}
}

func runFiles() error {
	c := exec.Command("git", "status", "--short")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func newCancelCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "cancel <task-id>",
		Short:        "Cancel a running agent task",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCancel(args[0])
		},
	}
}

func runCancel(taskID string) error {
	url := fmt.Sprintf("%s/a2a/cancel/%s", agentclient.DefaultBaseURL, taskID)
	resp, err := http.Post(url, "application/json", nil) //nolint:gosec,noctx
	if err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := agentclient.ReadBody(resp)
		if err != nil {
			return fmt.Errorf("failed to cancel task (and read response): %w", err)
		}
		return fmt.Errorf("failed to cancel task: %s", string(body))
	}

	fmt.Printf("✅ Task %s cancelled successfully\n", taskID)
	return nil
}

func newRetryCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "retry <task-id>",
		Short:        "Retry a failed agent task",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRetry(args[0])
		},
	}
}

func runRetry(taskID string) error {
	url := fmt.Sprintf("%s/a2a/retry/%s", agentclient.DefaultBaseURL, taskID)
	resp, err := http.Post(url, "application/json", nil) //nolint:gosec,noctx
	if err != nil {
		return fmt.Errorf("failed to retry task: %w", err)
	}
	defer resp.Body.Close()

	body, err := agentclient.ReadBody(resp)
	if err != nil {
		return fmt.Errorf("failed to retry task (and read response): %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to retry task: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err == nil {
		if newTaskID, ok := result["new_task_id"].(string); ok {
			fmt.Printf("✅ Task %s retried successfully\n", taskID)
			fmt.Printf("   New task ID: %s\n", newTaskID)
			return nil
		}
	}

	fmt.Printf("✅ Task %s retried successfully\n", taskID)
	return nil
}
