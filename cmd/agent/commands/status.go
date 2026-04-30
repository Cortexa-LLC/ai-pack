package commands

import (
	"encoding/json"
	"fmt"
	"os"

	agentclient "github.com/cortexa-llc/ai-pack/cmd/agent/client"
	"github.com/spf13/cobra"
)

const errNoAgentForBeadsTask = "❌ No agent found for task: %s\n"

func newStatusCmd() *cobra.Command {
	var jsonOutput bool
	var quiet bool

	cmd := &cobra.Command{
		Use:          "status <task-id>",
		Short:        "Show agent task status",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(args[0], jsonOutput, quiet)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Print only the status string")
	return cmd
}

func runStatus(taskID string, jsonOutput, quiet bool) error {
	internalTaskID := findInternalTaskIDFromServer(taskID)
	if internalTaskID == "" {
		internalTaskID = findInternalTaskID(taskID)
	}

	if internalTaskID == "" {
		if jsonOutput {
			fmt.Println(`{"error":"not_found"}`)
		} else if !quiet {
			fmt.Printf(errNoAgentForBeadsTask, taskID)
		}
		os.Exit(3) //nolint:gocritic // intentional semantic exit code
	}

	statusBody, err := runStatusByInternalID(internalTaskID)
	if err != nil {
		if jsonOutput {
			fmt.Printf(`{"error":"connection_failed","message":"%v"}`+"\n", err)
		} else if !quiet {
			fmt.Printf("❌ Failed to get status: %v\n", err)
		}
		os.Exit(3) //nolint:gocritic
	}

	var status map[string]interface{}
	json.Unmarshal(statusBody, &status) //nolint:errcheck

	statusStr, _ := status["status"].(string)

	if quiet {
		fmt.Println(statusStr)
	} else if jsonOutput {
		status["task_id"] = taskID
		jsonData, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(jsonData))
	} else {
		fmt.Printf("Task: %s\n", taskID)
		fmt.Printf("Status: %v\n", status["status"])
		if status["error"] != nil {
			fmt.Printf("Error: %v\n", status["error"])
		}
	}

	switch statusStr {
	case "completed":
		os.Exit(0)
	case "failed":
		os.Exit(1)
	case "in_progress":
		os.Exit(2)
	default:
		os.Exit(3)
	}
	return nil
}

// runStatusByInternalID issues GET /a2a/status/<internalTaskID> and returns the
// raw response body. It is extracted so that contract tests can verify the
// correct endpoint path without requiring a task lookup or calling os.Exit.
func runStatusByInternalID(internalTaskID string) ([]byte, error) {
	c := agentclient.Default()
	resp, err := c.Get(fmt.Sprintf("/a2a/status/%s", internalTaskID))
	if err != nil {
		return nil, err
	}
	return agentclient.ReadBody(resp)
}
