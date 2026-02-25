package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	agentclient "github.com/cortexa-llc/ai-pack/cmd/agent/client"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var running bool
	var completed bool
	var failed bool
	var all bool
	var jsonOutput bool
	var verbose bool

	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List agent tasks",
		Aliases:      []string{"ls"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(running, completed, failed, all, jsonOutput, verbose)
		},
	}

	cmd.Flags().BoolVarP(&running, "running", "r", false, "Show only running tasks")
	cmd.Flags().BoolVarP(&completed, "completed", "c", false, "Show only completed tasks")
	cmd.Flags().BoolVarP(&failed, "failed", "F", false, "Show only failed tasks")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Show all tasks")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, descOutputAsJSON)
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	return cmd
}

func runList(running, completed, failed, all, jsonOutput, verbose bool) error {
	showOnlyActive := !running && !completed && !failed && !all

	resp, err := http.Get(fmt.Sprintf("%s/a2a/tasks", agentclient.ServerURL))
	if err != nil {
		fmt.Printf("❌ Failed to query server: %v\n", err)
		fmt.Printf("   Is the agent server running? (agent-server --server)\n")
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var response struct {
		Tasks []struct {
			TaskID      string   `json:"task_id"`
			BeadsTaskID string   `json:"beads_task_id"`
			Status      string   `json:"status"`
			Role        string   `json:"role"`
			Description string   `json:"description"`
			ProjectRoot string   `json:"project_root"`
			StartedAt   string   `json:"started_at"`
			CompletedAt string   `json:"completed_at"`
			Tags        []string `json:"tags"`
			Error       string   `json:"error"`
			Duration    float64  `json:"duration_seconds"`
		} `json:"tasks"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		if jsonOutput {
			fmt.Println(string(body))
			return nil
		}
		fmt.Printf("❌ Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	// Filter tasks
	var tasks []struct {
		TaskID      string   `json:"task_id"`
		BeadsTaskID string   `json:"beads_task_id"`
		Status      string   `json:"status"`
		Role        string   `json:"role"`
		Description string   `json:"description"`
		ProjectRoot string   `json:"project_root"`
		StartedAt   string   `json:"started_at"`
		CompletedAt string   `json:"completed_at"`
		Tags        []string `json:"tags"`
		Error       string   `json:"error"`
		Duration    float64  `json:"duration_seconds"`
	}

	for _, t := range response.Tasks {
		switch {
		case showOnlyActive && (t.Status == "in_progress" || t.Status == "queued"):
			tasks = append(tasks, t)
		case running && t.Status == "in_progress":
			tasks = append(tasks, t)
		case completed && t.Status == "completed":
			tasks = append(tasks, t)
		case failed && t.Status == "failed":
			tasks = append(tasks, t)
		case all:
			tasks = append(tasks, t)
		}
	}

	if jsonOutput {
		jsonData, _ := json.MarshalIndent(map[string]interface{}{"tasks": tasks}, "", "  ")
		fmt.Println(string(jsonData))
		return nil
	}

	if len(tasks) == 0 {
		if showOnlyActive {
			fmt.Println("No active agents running. Use 'agent list --all' to see all tasks.")
		} else {
			fmt.Println("No tasks found matching filters.")
		}
		return nil
	}

	fmt.Println("Tasks:")
	fmt.Printf("%-25s %-12s %-12s %-45s\n", "BEADS ID", "ROLE", "STATUS", "DESCRIPTION")
	fmt.Println("────────────────────────────────────────────────────────────────────────────────────────────────")

	for _, t := range tasks {
		statusEmoji := statusIcon(t.Status)
		desc := truncateDescription(t.Description, 44)
		beadsID := t.BeadsTaskID
		if beadsID == "" {
			beadsID = t.TaskID[:minInt(12, len(t.TaskID))]
		}
		fmt.Printf("%-25s %-12s %s %-10s %-45s\n",
			beadsID,
			t.Role,
			statusEmoji,
			t.Status,
			desc,
		)

		if verbose {
			fmt.Printf("  Internal ID: %s\n", t.TaskID)
			fmt.Printf("  Project:     %s\n", t.ProjectRoot)
			if t.StartedAt != "" {
				started, err := time.Parse(time.RFC3339, t.StartedAt)
				if err == nil {
					fmt.Printf("  Started:     %s\n", started.Format("2006-01-02 15:04:05"))
				}
			}
			if t.Error != "" {
				fmt.Printf("  Error:       %s\n", t.Error)
			}
		}
	}
	return nil
}

func statusIcon(status string) string {
	switch status {
	case "completed":
		return "✅"
	case "failed":
		return "❌"
	case "in_progress":
		return "⚙️ "
	case "queued":
		return "⏳"
	default:
		return "❓"
	}
}
