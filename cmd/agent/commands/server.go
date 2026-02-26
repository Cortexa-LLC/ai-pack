package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	agentclient "github.com/cortexa-llc/ai-pack/cmd/agent/client"
	"github.com/spf13/cobra"
)

func newResumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "resume <task-id> [new-budget-tokens]",
		Short:        "Resume a paused or budget-exhausted task",
		Args:         cobra.RangeArgs(1, 2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var newBudget int64
			if len(args) >= 2 {
				fmt.Sscanf(args[1], "%d", &newBudget) //nolint:errcheck
			}
			return runResume(args[0], newBudget)
		},
	}
}

func runResume(taskID string, newBudget int64) error {
	url := fmt.Sprintf("%s/a2a/resume/%s", agentclient.ServerURL, taskID)

	var body io.Reader
	if newBudget > 0 {
		body = strings.NewReader(fmt.Sprintf(`{"new_budget":%d}`, newBudget))
	}

	resp, err := http.Post(url, "application/json", body) //nolint:gosec,noctx
	if err != nil {
		return fmt.Errorf("failed to resume task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to resume task: %s", string(respBody))
	}

	fmt.Printf("⏵  Task %s resuming from checkpoint\n", taskID)
	return nil
}

func newMetricsCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "metrics",
		Short:        "Show server performance metrics",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			showMetrics()
			return nil
		},
	}
}

func showMetrics() {
	resp, err := http.Get(fmt.Sprintf("%s/metrics", agentclient.ServerURL)) //nolint:gosec,noctx
	if err != nil {
		fmt.Printf("⚠️  Could not get metrics: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var metrics map[string]interface{}
	json.Unmarshal(body, &metrics) //nolint:errcheck

	fmt.Println()
	fmt.Println("📊 Server Metrics:")
	if spawned, ok := metrics["tasks_spawned"].(float64); ok {
		fmt.Printf("   Tasks spawned:    %d\n", int64(spawned))
	}
	if completed, ok := metrics["tasks_completed"].(float64); ok {
		fmt.Printf("   Tasks completed:  %d\n", int64(completed))
	}
	if failed, ok := metrics["tasks_failed"].(float64); ok {
		fmt.Printf("   Tasks failed:     %d\n", int64(failed))
	}
	if active, ok := metrics["active_tasks"].(float64); ok {
		fmt.Printf("   Active tasks:     %d\n", int64(active))
	}
}

func newPerformanceCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:          "performance",
		Aliases:      []string{"perf"},
		Short:        "Show detailed performance report",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPerformance(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, descOutputAsJSON)
	return cmd
}

func runPerformance(jsonOutput bool) error {
	resp, err := http.Get(fmt.Sprintf("%s/metrics", agentclient.ServerURL)) //nolint:gosec,noctx
	if err != nil {
		return fmt.Errorf("could not fetch metrics: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if jsonOutput {
		fmt.Println(string(body))
		return nil
	}

	var metrics map[string]interface{}
	if err := json.Unmarshal(body, &metrics); err != nil {
		return fmt.Errorf("failed to parse metrics: %w", err)
	}

	fmt.Println("======================================================================")
	fmt.Println("                    AGENT PERFORMANCE REPORT")
	fmt.Println("======================================================================")

	if model, ok := metrics["model"].(string); ok {
		fmt.Printf("Model: %s\n", model)
	}
	fmt.Println()

	if inputTokens, ok := metrics["total_input_tokens"].(float64); ok {
		if outputTokens, ok := metrics["total_output_tokens"].(float64); ok {
			fmt.Printf("Token Usage:\n")
			fmt.Printf("  Input:  %s\n", formatNumber(int64(inputTokens)))
			fmt.Printf("  Output: %s\n", formatNumber(int64(outputTokens)))
			fmt.Printf("  Total:  %s\n", formatNumber(int64(inputTokens+outputTokens)))
		}
	}

	if cost, ok := metrics["total_cost_usd"].(float64); ok {
		fmt.Printf("\nEstimated Cost: $%.4f USD\n", cost)
	}

	if tasksCompleted, ok := metrics["tasks_completed"].(float64); ok && tasksCompleted > 0 {
		if avgDuration, ok := metrics["avg_task_duration_seconds"].(float64); ok {
			fmt.Printf("\nTask Statistics:\n")
			fmt.Printf("  Completed: %.0f\n", tasksCompleted)
			fmt.Printf("  Avg Duration: %.1fs\n", avgDuration)
		}
	}

	fmt.Println("======================================================================")
	return nil
}

func newDiscoveryCmd() *cobra.Command {
	var jsonOutput bool
	var verbose bool

	cmd := &cobra.Command{
		Use:          "discovery",
		Short:        "Show A2A protocol discovery information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiscovery(jsonOutput, verbose)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, descOutputAsJSON)
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	return cmd
}

func runDiscovery(jsonOutput, verbose bool) error {
	resp, err := http.Get(fmt.Sprintf("%s/a2a/discovery", agentclient.ServerURL)) //nolint:gosec,noctx
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if jsonOutput {
		fmt.Println(string(body))
		return nil
	}

	var discovery map[string]interface{}
	if err := json.Unmarshal(body, &discovery); err != nil {
		return fmt.Errorf("failed to parse discovery response: %w", err)
	}

	fmt.Println("=== A2A Server Discovery ===")
	if name, ok := discovery["name"].(string); ok {
		fmt.Printf("Name: %s\n", name)
	}
	if version, ok := discovery["version"].(string); ok {
		fmt.Printf("Version: %s\n", version)
	}
	if endpoint, ok := discovery["endpoint"].(string); ok {
		fmt.Printf("Endpoint: %s\n", endpoint)
	}

	if verbose {
		if capabilities, ok := discovery["capabilities"].(map[string]interface{}); ok {
			fmt.Println("\nCapabilities:")
			for k, v := range capabilities {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}
		if roles, ok := discovery["supported_roles"].([]interface{}); ok {
			fmt.Println("\nSupported Roles:")
			for _, r := range roles {
				fmt.Printf("  - %s\n", r)
			}
		}
	}

	return nil
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "Show version information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			runVersion()
			return nil
		},
	}
}

func runVersion() {
	resp, err := http.Get(fmt.Sprintf("%s/health", agentclient.ServerURL)) //nolint:gosec,noctx
	if err != nil {
		fmt.Printf("CLI Version: %s\n", getVersion())
		fmt.Printf("Server:      (not running)\n")
		fmt.Printf("Platform:    %s\n", getRuntimeInfo())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var health map[string]interface{}
	json.Unmarshal(body, &health) //nolint:errcheck

	fmt.Printf("CLI Version:    %s\n", getVersion())
	if serverVersion, ok := health["version"].(string); ok {
		fmt.Printf("Server Version: %s\n", serverVersion)
	}
	fmt.Printf("Server URL:     %s\n", agentclient.ServerURL)
	fmt.Printf("Platform:       %s\n", getRuntimeInfo())
	fmt.Printf("Status:         %s\n", statusIcon("completed")+" connected")
}

func getVersion() string {
	return "1.0.0"
}
