package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	agentclient "github.com/cortexa-llc/ai-pack/cmd/agent/client"
	"github.com/spf13/cobra"
)

func newWaitCmd() *cobra.Command {
	var timeout time.Duration
	var stream bool
	var inactiveTimeout time.Duration

	cmd := &cobra.Command{
		Use:          "wait <beads-task-id>",
		Short:        "Wait for a task to complete",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWait(args[0], timeout, stream, inactiveTimeout)
		},
	}

	cmd.Flags().DurationVar(&timeout, "timeout", 4*time.Hour, "Maximum wait time (e.g. 30m, 4h)")
	cmd.Flags().BoolVar(&stream, "stream", false, "Stream live output while waiting")
	cmd.Flags().DurationVar(&inactiveTimeout, "inactive-timeout", 2*time.Minute, "Stream inactivity timeout")
	return cmd
}

func runWait(beadsTaskID string, timeout time.Duration, stream bool, inactiveTimeout time.Duration) error {
	if stream {
		internalTaskID := findInternalTaskIDFromServer(beadsTaskID)
		if internalTaskID == "" {
			internalTaskID = findInternalTaskID(beadsTaskID)
		}
		if internalTaskID == "" {
			fmt.Printf("❌ No agent found for Beads task: %s\n", beadsTaskID)
			fmt.Printf("   Tip: Check 'agent list' for active agents\n")
			os.Exit(1)
		}
		if !streamTaskProgressWithInactivity(internalTaskID, inactiveTimeout) {
			os.Exit(1)
		}
		return nil
	}

	fmt.Printf("⏳ Waiting for task %s (timeout: %v)...\n", beadsTaskID, timeout)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		internalTaskID := findInternalTaskIDFromServer(beadsTaskID)
		if internalTaskID == "" {
			internalTaskID = findInternalTaskID(beadsTaskID)
		}
		if internalTaskID == "" {
			fmt.Printf("⚠️  Task not found yet, retrying...\n")
			time.Sleep(5 * time.Second)
			continue
		}

		if err := runWaitHTTP(internalTaskID, timeout, inactiveTimeout); err == nil {
			return nil
		}
	}
	fmt.Printf("⏰ Timeout waiting for task %s\n", beadsTaskID)
	os.Exit(1)
	return nil
}

// runWaitHTTP polls GET /a2a/tasks/<internalTaskID> until the task reaches a
// terminal state or the timeout elapses. It is extracted so that contract tests
// can verify the correct endpoint path without a Beads task lookup.
func runWaitHTTP(internalTaskID string, timeout, _ time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c := agentclient.Default()
		resp, err := c.Get(fmt.Sprintf("/a2a/tasks/%s", internalTaskID))
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		body, _ := agentclient.ReadBody(resp)
		var status map[string]interface{}
		json.Unmarshal(body, &status) //nolint:errcheck

		statusStr, _ := status["status"].(string)
		switch statusStr {
		case "completed":
			fmt.Println("✅ Agent completed!")
			showMetrics()
			return nil
		case "failed":
			fmt.Println("❌ Agent failed!")
			os.Exit(1)
		default:
			fmt.Printf("  Status: %s\n", statusStr)
			time.Sleep(10 * time.Second)
		}
	}
	return fmt.Errorf("timeout waiting for task %s", internalTaskID)
}
