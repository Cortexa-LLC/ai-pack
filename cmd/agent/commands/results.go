package commands

import (
	"fmt"
	"os"
	"path/filepath"

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
	internalTaskID, projectRoot := findTaskIDAndProjectFromServer(beadsTaskID)
	if internalTaskID == "" {
		internalTaskID = findInternalTaskID(beadsTaskID)
		projectRoot = "."
	}

	if internalTaskID == "" {
		fmt.Printf(errNoAgentForBeadsTask, beadsTaskID)
		fmt.Printf("   Tip: Check 'agent list' for active agents or 'bd show %s' for task status\n", beadsTaskID)
		os.Exit(1)
	}

	resultsFile := filepath.Join(projectRoot, ".beads", "tasks", internalTaskID, "30-results.md")
	data, err := os.ReadFile(resultsFile)
	if err != nil {
		fmt.Printf("❌ No results found: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(data))
	return nil
}
