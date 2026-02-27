package commands

import (
	"fmt"
	"os"
	"path/filepath"

	agentclient "github.com/cortexa-llc/ai-pack/cmd/agent/client"
	"github.com/spf13/cobra"
)

func newResultsCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "results <beads-task-id>",
		Short:        "Show the results for a completed task",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResults(args[0])
		},
	}
}

func runResults(beadsTaskID string) error {
	// Resolve the internal task ID (execution folder name) via the server or
	// local .beads directory scan.
	internalTaskID, projectRoot := findTaskIDAndProjectFromServer(beadsTaskID)
	if internalTaskID == "" {
		internalTaskID = findInternalTaskID(beadsTaskID)
		projectRoot = ""
	}

	if internalTaskID == "" {
		return fmt.Errorf("no task found for beads ID %s – check 'agent list' or 'bd show %s'", beadsTaskID, beadsTaskID)
	}

	// 1. Try the server API – works whether or not the task files are present
	//    on this machine.
	c := agentclient.Default()
	if result, ok := c.FetchTaskResults(internalTaskID); ok {
		fmt.Print(result)
		return nil
	}

	// 2. Disk fallback – for cases where the server is not running or the task
	//    was created locally without a running server.
	if projectRoot == "" {
		projectRoot = detectProjectRoot()
		if projectRoot == "" {
			projectRoot, _ = os.Getwd()
		}
	}

	resultsFile := filepath.Join(projectRoot, ".beads", "tasks", internalTaskID, "30-results.md")
	data, err := os.ReadFile(resultsFile)
	if err != nil {
		return fmt.Errorf("no results found for task %s: %w", beadsTaskID, err)
	}

	fmt.Print(string(data))
	return nil
}
