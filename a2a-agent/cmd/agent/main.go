package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const Version = "2.2.0"
const ServerURL = "http://localhost:8080"
const descOutputAsJSON = "Output as JSON"
const sseDataPrefix = "data: "
const errNoAgentForBeadsTask = "❌ No agent found for Beads task: %s\n"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "status":
		handleStatus(os.Args[2:])
	case "results":
		handleResults(os.Args[2:])
	case "logs":
		handleLogs(os.Args[2:])
	case "list":
		handleList(os.Args[2:])
	case "wait":
		handleWait(os.Args[2:])
	case "diff":
		handleDiff(os.Args[2:])
	case "files":
		handleFiles(os.Args[2:])
	case "cancel":
		handleCancel(os.Args[2:])
	case "retry":
		handleRetry(os.Args[2:])
	case "metrics":
		handleMetrics(os.Args[2:])
	case "performance", "perf":
		handlePerformance(os.Args[2:])
	case "discovery", "discover":
		handleDiscovery(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("AI-Pack Agent CLI v%s\n", Version)
	case "help", "--help", "-h":
		usage()
	default:
		// Assume it's a spawn command: agent <role> <task-id>
		handleSpawn(os.Args[1:])
	}
}

func parseSpawnFlags(args []string) (role, taskInput string, wait, stream bool) {
	var positionalArgs []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--wait" {
			wait = true
		} else if arg == "--stream" {
			stream = true
		} else if arg == "--async" {
			// Reserved for future use, ignore
		} else {
			positionalArgs = append(positionalArgs, arg)
		}
	}

	if len(positionalArgs) < 2 {
		fmt.Println("Usage: agent <role> <beads-task-id> [--wait|--stream]")
		os.Exit(1)
	}

	role = positionalArgs[0]
	taskInput = strings.Join(positionalArgs[1:], " ")
	return
}

func validateBeadsTaskOrExit(taskInput, role string) {
	if !isValidBeadsTask(taskInput) {
		fmt.Printf("❌ Error: '%s' is not a valid Beads task ID\n", taskInput)
		fmt.Println()
		fmt.Println("Create a Beads task first:")
		fmt.Printf("  bd create \"Working directory: %s\n", mustGetWorkingDir())
		fmt.Println("  Task packet: .ai/tasks/YYYY-MM-DD_task-name/")
		fmt.Println("  ")
		fmt.Println("  Task description...\" --priority high")
		fmt.Println()
		fmt.Println("Then spawn the agent:")
		fmt.Printf("  agent %s <task-id>\n", role)
		os.Exit(1)
	}
}

func handleSpawn(args []string) {
	role, taskInput, wait, stream := parseSpawnFlags(args)
	validateBeadsTaskOrExit(taskInput, role)

	fmt.Printf("🎯 Beads task: %s\n", taskInput)

	// Detect project root for Beads integration
	projectRoot := detectProjectRoot()
	if projectRoot == "" {
		projectRoot = mustGetWorkingDir()
	}

	// Execute task via HTTP API
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "execute",
		"params": map[string]interface{}{
			"role":         role,
			"task":         taskInput,
			"project_root": projectRoot,
		},
		"id": 1,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🚀 Spawning %s agent...\n", role)

	resp, err := http.Post(
		fmt.Sprintf("%s/a2a/execute", ServerURL),
		"application/json",
		strings.NewReader(string(jsonData)),
	)
	if err != nil {
		fmt.Printf("❌ Failed to execute task: %v\n", err)
		fmt.Printf("   Is the agent server running? (agent-server --server)\n")
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	// Check for JSON-RPC error
	if errObj, ok := result["error"].(map[string]interface{}); ok {
		fmt.Printf("❌ Server error: %v\n", errObj["message"])
		if data, ok := errObj["data"].(string); ok && data != "" {
			fmt.Printf("   %s\n", data)
		}
		os.Exit(1)
	}

	// Extract internal task ID from response
	var internalTaskID string
	if resultObj, ok := result["result"].(map[string]interface{}); ok {
		if taskID, ok := resultObj["task_id"].(string); ok {
			internalTaskID = taskID
		}
	}

	if internalTaskID == "" {
		fmt.Printf("❌ Failed to get task ID from server response\n")
		os.Exit(1)
	}

	fmt.Printf("   Internal task ID: %s\n", internalTaskID)

	// Wait a moment for the task to be registered
	time.Sleep(2 * time.Second)

	// Show initial metrics
	showMetrics()

	// If --stream flag, stream real-time progress via SSE
	if stream {
		fmt.Println()
		fmt.Println("📡 Streaming real-time progress...")
		// Use internal task ID directly for streaming
		streamURL := fmt.Sprintf("%s/stream/%s", ServerURL, internalTaskID)
		resp, err := http.Get(streamURL)
		if err != nil {
			fmt.Printf("❌ Failed to connect to stream: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Printf("❌ Stream connection failed with status %d\n", resp.StatusCode)
			os.Exit(1)
		}

		fmt.Printf("✓ Connected to task stream: %s\n", internalTaskID)
		fmt.Println()

		// Read SSE events
		buffer := make([]byte, 8192)
		dataBuffer := ""

		for {
			n, err := resp.Body.Read(buffer)
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Printf("⚠️  Stream read error: %v\n", err)
				break
			}

			dataBuffer += string(buffer[:n])

			for {
				idx := strings.Index(dataBuffer, "\n\n")
				if idx == -1 {
					break
				}

				message := dataBuffer[:idx]
				dataBuffer = dataBuffer[idx+2:]

				if strings.HasPrefix(message, sseDataPrefix) {
					jsonData := strings.TrimPrefix(message, sseDataPrefix)

					var event map[string]interface{}
					if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
						continue
					}

					eventType, _ := event["type"].(string)
					timestamp, _ := event["timestamp"].(string)
					data, _ := event["data"].(map[string]interface{})

					switch eventType {
					case "status_update":
						status, _ := data["status"].(string)
						progress, _ := data["progress"].(float64)
						fmt.Printf("[%s] Status: %s (%.0f%%)\n", timestamp, status, progress*100)
					case "completed":
						fmt.Printf("[%s] 🎉 Task completed!\n", timestamp)
						fmt.Println()
						showMetrics()
						return
					case "failed":
						errorMsg, _ := data["error"].(string)
						fmt.Printf("[%s] ❌ Task failed: %s\n", timestamp, errorMsg)
						os.Exit(1)
					}
				}
			}
		}
		fmt.Println()
		fmt.Println("✓ Stream closed")
		showMetrics()
		return
	}

	// If --wait flag, poll until complete
	if wait {
		fmt.Println()
		fmt.Println("⏳ Waiting for completion...")
		for {
			resp, err := http.Get(fmt.Sprintf("%s/a2a/status/%s", ServerURL, internalTaskID))
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}

			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			var status map[string]interface{}
			if err := json.Unmarshal(body, &status); err != nil {
				time.Sleep(2 * time.Second)
				continue
			}

			statusVal, ok := status["status"]
			if !ok || statusVal == nil {
				time.Sleep(2 * time.Second)
				continue
			}

			statusStr, ok := statusVal.(string)
			if !ok {
				time.Sleep(2 * time.Second)
				continue
			}

			if statusStr == "completed" {
				fmt.Println("✅ Agent completed!")
				showMetrics()
				break
			} else if statusStr == "failed" {
				fmt.Println("❌ Agent failed!")
				if status["error"] != nil {
					fmt.Printf("   Error: %v\n", status["error"])
				}
				break
			}

			time.Sleep(5 * time.Second)
		}
	}
}

func printStatusUsage(fs *flag.FlagSet) {
	fmt.Println("Usage: agent status <task-id> [options]")
	fmt.Println()
	fmt.Println("Options:")
	fs.PrintDefaults()
	fmt.Println()
	fmt.Println("Exit codes:")
	fmt.Println("  0 - completed")
	fmt.Println("  1 - failed")
	fmt.Println("  2 - in_progress")
	fmt.Println("  3 - not found")
}

func parseStatusFlags(args []string) (taskID string, jsonOutput, quiet *bool) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	jsonOutput = fs.Bool("json", false, descOutputAsJSON)
	quiet = fs.Bool("quiet", false, "Output only status value")
	fs.Usage = func() { printStatusUsage(fs) }

	fs.Parse(args)
	positionalArgs := fs.Args()

	if len(positionalArgs) < 1 {
		fmt.Println("Usage: agent status <task-id> [options]")
		os.Exit(1)
	}

	taskID = positionalArgs[0]
	return
}

func handleStatus(args []string) {
	beadsTaskID, jsonOutput, quiet := parseStatusFlags(args)

	// Try to find task ID by querying server first (for cross-project tasks)
	internalTaskID := findInternalTaskIDFromServer(beadsTaskID)

	// Fallback: Look up internal task ID from local filesystem
	if internalTaskID == "" {
		internalTaskID = findInternalTaskID(beadsTaskID)
	}

	if internalTaskID == "" {
		if *jsonOutput {
			fmt.Println(`{"error":"not_found"}`)
		} else if !*quiet {
			fmt.Printf(errNoAgentForBeadsTask, beadsTaskID)
		}
		os.Exit(3) // not found
	}

	resp, err := http.Get(fmt.Sprintf("%s/a2a/status/%s", ServerURL, internalTaskID))
	if err != nil {
		if *jsonOutput {
			fmt.Printf(`{"error":"connection_failed","message":"%v"}\n`, err)
		} else if !*quiet {
			fmt.Printf("❌ Failed to get status: %v\n", err)
		}
		os.Exit(3)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var status map[string]interface{}
	json.Unmarshal(body, &status)

	statusStr, _ := status["status"].(string)

	// Output
	if *quiet {
		fmt.Println(statusStr)
	} else if *jsonOutput {
		// Add task_id to JSON output
		status["task_id"] = beadsTaskID
		jsonData, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(jsonData))
	} else {
		fmt.Printf("Task: %s\n", beadsTaskID)
		fmt.Printf("Status: %v\n", status["status"])
		if status["progress"] != nil {
			// Format progress as percentage (0.0-1.0 -> 0%-100%)
			if progress, ok := status["progress"].(float64); ok {
				fmt.Printf("Progress: %.0f%%\n", progress*100)
			} else {
				fmt.Printf("Progress: %v\n", status["progress"])
			}
		}
		if status["error"] != nil {
			fmt.Printf("Error: %v\n", status["error"])
		}
	}

	// Exit with semantic code
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
}

func handleResults(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: agent results <task-id>")
		os.Exit(1)
	}

	beadsTaskID := args[0]

	// Try to find task ID and project root by querying server first (for cross-project tasks)
	internalTaskID, projectRoot := findTaskIDAndProjectFromServer(beadsTaskID)

	// Fallback: Look up internal task ID from local filesystem
	if internalTaskID == "" {
		internalTaskID = findInternalTaskID(beadsTaskID)
		projectRoot = "." // Current directory
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
}

func handleLogs(args []string) {
	// Parse flags
	fs := flag.NewFlagSet("logs", flag.ExitOnError)
	tailLines := fs.Int("tail", 0, "Show last N lines")
	follow := fs.Bool("follow", false, "Stream new log lines")
	serverLogs := fs.Bool("server", false, "Show server logs instead of task logs")
	allLogs := fs.Bool("all", false, "Show all logs (server + all tasks)")
	jsonOutput := fs.Bool("json", false, descOutputAsJSON)

	fs.Usage = func() {
		fmt.Println("Usage: agent logs <task-id> [options]")
		fmt.Println("       agent logs --server [options]")
		fmt.Println("       agent logs --all [options]")
		fmt.Println()
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	fs.Parse(args)
	positionalArgs := fs.Args()

	// Server logs
	if *serverLogs {
		if *follow {
			streamServerLogs(*jsonOutput)
		} else {
			fetchRecentServerLogs(*tailLines, *jsonOutput)
		}
		return
	}

	// All logs
	if *allLogs {
		if *follow {
			fmt.Println("❌ --all with --follow not yet supported")
			os.Exit(1)
		}
		fetchAllLogs(*tailLines, *jsonOutput)
		return
	}

	// Task logs (default)
	if len(positionalArgs) < 1 {
		fmt.Println("Usage: agent logs <task-id> [options]")
		fmt.Println("       agent logs --server [options]")
		os.Exit(1)
	}

	beadsTaskID := positionalArgs[0]

	// Try to find task ID and project root by querying server first (for cross-project tasks)
	internalTaskID, projectRoot := findTaskIDAndProjectFromServer(beadsTaskID)

	// Fallback: Look up internal task ID from local filesystem
	if internalTaskID == "" {
		internalTaskID = findInternalTaskID(beadsTaskID)
		projectRoot = "." // Current directory
	}

	if internalTaskID == "" {
		fmt.Printf(errNoAgentForBeadsTask, beadsTaskID)
		fmt.Printf("   Tip: Check 'agent list' for active agents or 'bd show %s' for task status\n", beadsTaskID)
		os.Exit(1)
	}

	// Construct log file path using project root
	logFile := filepath.Join(projectRoot, ".beads", "tasks", internalTaskID, "execution.log")

	if *follow {
		followLogFile(logFile, *jsonOutput)
	} else {
		displayLogFile(logFile, *tailLines, *jsonOutput)
	}
}

func displayLogFile(logFile string, tailLines int, jsonOutput bool) {
	data, err := os.ReadFile(logFile)
	if err != nil {
		fmt.Printf("❌ No logs found: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(data), "\n")

	// Apply tail
	if tailLines > 0 && len(lines) > tailLines {
		lines = lines[len(lines)-tailLines:]
	}

	if jsonOutput {
		// Output as JSON array of lines
		jsonData, _ := json.Marshal(lines)
		fmt.Println(string(jsonData))
	} else {
		fmt.Println(strings.Join(lines, "\n"))
	}
}

func followLogFile(logFile string, jsonOutput bool) {
	// Read initial content
	initialData, err := os.ReadFile(logFile)
	if err != nil {
		fmt.Printf("❌ No logs found: %v\n", err)
		os.Exit(1)
	}

	if !jsonOutput {
		fmt.Println(string(initialData))
		fmt.Println()
		fmt.Println("📡 Following log file (Ctrl+C to stop)...")
	}

	// Follow new lines
	lastSize := int64(len(initialData))
	noChangeCount := 0

	for {
		time.Sleep(1 * time.Second)

		stat, err := os.Stat(logFile)
		if err != nil {
			continue
		}

		if stat.Size() > lastSize {
			file, err := os.Open(logFile)
			if err != nil {
				continue
			}

			file.Seek(lastSize, 0)
			newData, _ := io.ReadAll(file)
			file.Close()

			if jsonOutput {
				lines := strings.Split(string(newData), "\n")
				for _, line := range lines {
					if line != "" {
						jsonLine, _ := json.Marshal(map[string]string{"line": line})
						fmt.Println(string(jsonLine))
					}
				}
			} else {
				fmt.Print(string(newData))
			}

			lastSize = stat.Size()
			noChangeCount = 0

			// Check if log contains completion marker
			logContent := string(newData)
			if strings.Contains(logContent, "✅ Agent completed") ||
				strings.Contains(logContent, "🎉 Task completed successfully") ||
				strings.Contains(logContent, "❌ Task failed") {
				if !jsonOutput {
					fmt.Println()
					fmt.Println("✓ Task completed, exiting follow mode")
				}
				return
			}
		} else {
			noChangeCount++
			// If no changes for 60 seconds, check if task still exists
			if noChangeCount > 60 {
				// Could query server here to check task status
				noChangeCount = 0
			}
		}
	}
}

func streamServerLogs(jsonOutput bool) {
	resp, err := http.Get(fmt.Sprintf("%s/logs/stream", ServerURL))
	if err != nil {
		fmt.Printf("❌ Failed to connect to log stream: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("❌ Server returned status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Println("📡 Streaming server logs (Ctrl+C to stop)...")
	if !jsonOutput {
		fmt.Println()
	}

	buffer := make([]byte, 8192)
	dataBuffer := ""

	for {
		n, err := resp.Body.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("⚠️  Stream read error: %v\n", err)
			break
		}

		dataBuffer += string(buffer[:n])

		for {
			idx := strings.Index(dataBuffer, "\n\n")
			if idx == -1 {
				break
			}

			message := dataBuffer[:idx]
			dataBuffer = dataBuffer[idx+2:]

			if strings.HasPrefix(message, sseDataPrefix) {
				jsonData := strings.TrimPrefix(message, sseDataPrefix)

				if jsonOutput {
					fmt.Println(jsonData)
				} else {
					var logEntry map[string]interface{}
					if err := json.Unmarshal([]byte(jsonData), &logEntry); err == nil {
						timestamp, _ := logEntry["timestamp"].(string)
						level, _ := logEntry["level"].(string)
						msg, _ := logEntry["message"].(string)
						fmt.Printf("[%s] %s: %s\n", timestamp, level, msg)
					}
				}
			}
		}
	}
}

func fetchRecentServerLogs(tailLines int, jsonOutput bool) {
	url := fmt.Sprintf("%s/logs/recent", ServerURL)
	if tailLines > 0 {
		url += fmt.Sprintf("?limit=%d", tailLines)
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ Failed to fetch server logs: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if jsonOutput {
		fmt.Println(string(body))
	} else {
		// Server returns: {"logs": [...], "count": N, "limit": N}
		var response struct {
			Logs  []map[string]interface{} `json:"logs"`
			Count int                      `json:"count"`
			Limit int                      `json:"limit"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Printf("❌ Failed to parse logs: %v\n", err)
			os.Exit(1)
		}

		for _, entry := range response.Logs {
			timestamp, _ := entry["timestamp"].(string)
			level, _ := entry["level"].(string)
			msg, _ := entry["message"].(string)
			fmt.Printf("[%s] %s: %s\n", timestamp, level, msg)
		}
	}
}

func fetchAllLogs(tailLines int, jsonOutput bool) {
	// Fetch server logs
	fetchRecentServerLogs(tailLines, jsonOutput)

	// Fetch all task logs
	matches, _ := filepath.Glob(".beads/tasks/task-*/execution.log")
	for _, logFile := range matches {
		taskID := filepath.Base(filepath.Dir(logFile))

		if !jsonOutput {
			fmt.Printf("\n=== Task: %s ===\n", taskID)
		}

		displayLogFile(logFile, tailLines, jsonOutput)
	}
}

func handleListServer(running, completed, failed, all, jsonOutput, verboseOutput *bool) {
	// Determine if we should show only active tasks
	showOnlyActive := !*running && !*completed && !*failed && !*all
	resp, err := http.Get(fmt.Sprintf("%s/a2a/tasks", ServerURL))
	if err != nil {
		fmt.Printf("❌ Failed to query server: %v\n", err)
		fmt.Printf("   Is the agent server running? (agent-server --server)\n")
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var response struct {
		Tasks []struct {
			TaskID      string  `json:"task_id"`
			BeadsTaskID string  `json:"beads_task_id"`
			Status      string  `json:"status"`
			Role        string  `json:"role"`
			Description string  `json:"description"`
			ProjectRoot string  `json:"project_root"`
			Progress    float64 `json:"progress"`
			Error       string  `json:"error"`
		} `json:"tasks"`
		Count int `json:"count"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("❌ Failed to parse server response: %v\n", err)
		os.Exit(1)
	}

	// Filter tasks
	var filteredTasks []struct {
		TaskID      string
		BeadsTaskID string
		Status      string
		Role        string
		Description string
		ProjectRoot string
	}

	for _, task := range response.Tasks {
		// Apply filters
		if showOnlyActive {
			// Default: show only active work (running, queued, in_progress)
			if task.Status == "completed" || task.Status == "failed" {
				continue
			}
		} else {
			// Specific filters
			if *running && task.Status != "in_progress" {
				continue
			}
			if *completed && task.Status != "completed" {
				continue
			}
			if *failed && task.Status != "failed" {
				continue
			}
			// If --all, show everything (no filtering)
		}

		filteredTasks = append(filteredTasks, struct {
			TaskID      string
			BeadsTaskID string
			Status      string
			Role        string
			Description string
			ProjectRoot string
		}{
			TaskID:      task.TaskID,
			BeadsTaskID: task.BeadsTaskID,
			Status:      task.Status,
			Role:        task.Role,
			Description: task.Description,
			ProjectRoot: task.ProjectRoot,
		})
	}

	// Output
	if *jsonOutput {
		jsonData, _ := json.MarshalIndent(filteredTasks, "", "  ")
		fmt.Println(string(jsonData))
		return
	}

	if *verboseOutput {
		fmt.Println("Machine-Wide Agent Status (from server):")
		fmt.Println()
		for _, task := range filteredTasks {
			fmt.Printf("  %s [%s]\n", task.TaskID, task.Status)
			if task.BeadsTaskID != "" {
				fmt.Printf("    Beads ID: %s\n", task.BeadsTaskID)
			}
			fmt.Printf("    Role: %s\n", task.Role)
			fmt.Printf("    Project: %s\n", task.ProjectRoot)
			fmt.Printf("    Task: %s\n", task.Description)
			fmt.Println()
		}
		return
	}

	// Compact format (default)
	if len(filteredTasks) > 0 {
		fmt.Println("STATUS      BEADS-ID      INTERNAL-ID                             DESCRIPTION")
		fmt.Println("----------  ------------  --------------------------------------  -----------")
	}

	for _, task := range filteredTasks {
		statusText := ""
		switch task.Status {
		case "in_progress":
			statusText = "RUNNING"
		case "completed":
			statusText = "COMPLETED"
		case "failed":
			statusText = "FAILED"
		default:
			statusText = task.Status
		}

		description := truncateDescription(task.Description, 50)

		beadsID := task.BeadsTaskID
		if beadsID == "" {
			beadsID = "(none)"
		}

		internalID := task.TaskID
		if len(internalID) > 38 {
			internalID = internalID[:35] + "..."
		}

		fmt.Printf("%-10s  %-12s  %-38s  %s\n", statusText, beadsID, internalID, description)
	}

	if len(filteredTasks) == 0 {
		fmt.Println("No agents found")
	}
}

func handleList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	running := fs.Bool("running", false, "Show only running agents")
	completed := fs.Bool("completed", false, "Show only completed agents")
	failed := fs.Bool("failed", false, "Show only failed agents")
	all := fs.Bool("all", false, "Show all agents (including completed/failed)")
	jsonOutput := fs.Bool("json", false, descOutputAsJSON)
	verboseOutput := fs.Bool("verbose", false, "Verbose output (show role and full details)")
	serverQuery := fs.Bool("server", false, "Query server for machine-wide tasks (default: local project only)")
	fs.Parse(args)

	// Determine if we should show only active tasks
	showOnlyActive := !*running && !*completed && !*failed && !*all

	// If --server flag, query the API for machine-wide view
	if *serverQuery {
		handleListServer(running, completed, failed, all, jsonOutput, verboseOutput)
		return
	}

	// Otherwise, list local .beads/tasks/task-* directories (project-specific)
	matches, _ := filepath.Glob(".beads/tasks/task-*")

	type AgentInfo struct {
		TaskID      string `json:"task_id"`
		BeadsTaskID string `json:"beads_task_id,omitempty"`
		Status      string `json:"status"`
		Role        string `json:"role"`
		Description string `json:"description"`
	}

	var agents []AgentInfo

	for _, taskDir := range matches {
		metaFile := filepath.Join(taskDir, "00-metadata.json")
		data, err := os.ReadFile(metaFile)
		if err != nil {
			continue
		}

		var meta map[string]interface{}
		json.Unmarshal(data, &meta)

		status, _ := meta["status"].(string)

		// Filter by status
		if showOnlyActive {
			// Default: show only active work (running, queued, in_progress)
			// Hide completed and failed
			if status == "completed" || status == "failed" {
				continue
			}
		} else {
			// Specific filters
			if *running && status != "in_progress" {
				continue
			}
			if *completed && status != "completed" {
				continue
			}
			if *failed && status != "failed" {
				continue
			}
			// If --all, show everything (no filtering)
		}

		taskID := filepath.Base(taskDir)
		role, _ := meta["role"].(string)
		description := fmt.Sprintf("%v", meta["description"])
		beadsTaskID := ""

		// Check metadata at root level (new location)
		if metadata, ok := meta["metadata"].(map[string]interface{}); ok {
			if btid, ok := metadata["beads_task_id"].(string); ok {
				beadsTaskID = btid
			}
		}

		// Fallback: check old location for backward compatibility
		if beadsTaskID == "" {
			if config, ok := meta["config"].(map[string]interface{}); ok {
				if md, ok := config["metadata"].(map[string]interface{}); ok {
					if btid, ok := md["beads_task_id"].(string); ok {
						beadsTaskID = btid
					}
				}
			}
		}

		agents = append(agents, AgentInfo{
			TaskID:      taskID,
			BeadsTaskID: beadsTaskID,
			Status:      status,
			Role:        role,
			Description: description,
		})
	}

	if *jsonOutput {
		jsonData, _ := json.MarshalIndent(agents, "", "  ")
		fmt.Println(string(jsonData))
	} else if *verboseOutput {
		// Verbose format: full details with role and status
		fmt.Println("Active Agents:")
		fmt.Println()
		for _, agent := range agents {
			// Show Beads task ID as primary identifier
			displayID := agent.BeadsTaskID
			if displayID == "" {
				displayID = agent.TaskID // Fallback to internal ID if no Beads task
			}

			fmt.Printf("  %s [%s]\n", displayID, agent.Status)
			fmt.Printf("    Role: %s\n", agent.Role)
			fmt.Printf("    Task: %s\n", agent.Description)
			fmt.Println()
		}
	} else {
		// Compact format (default): Status, Beads ID, Internal ID, and description
		if len(agents) > 0 {
			fmt.Println("STATUS      BEADS-ID      INTERNAL-ID                             DESCRIPTION")
			fmt.Println("----------  ------------  --------------------------------------  -----------")
		}

		for _, agent := range agents {
			// Format status with text
			statusText := ""
			switch agent.Status {
			case "in_progress":
				statusText = "RUNNING"
			case "completed":
				statusText = "COMPLETED"
			case "failed":
				statusText = "FAILED"
			default:
				statusText = agent.Status
			}

			// Truncate long descriptions
			description := truncateDescription(agent.Description, 50)

			// Show both IDs
			beadsID := agent.BeadsTaskID
			if beadsID == "" {
				beadsID = "(none)"
			}
			internalID := agent.TaskID
			if len(internalID) > 38 {
				internalID = internalID[:35] + "..."
			}

			fmt.Printf("%-10s  %-12s  %-38s  %s\n", statusText, beadsID, internalID, description)
		}

		if len(agents) == 0 {
			fmt.Println("No agents found")
		}
	}
}

func handleWait(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: agent wait <task-id>")
		os.Exit(1)
	}

	taskID := args[0]
	waitForTaskCompletion(taskID)
}

func handleDiff(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: agent diff <task-id>")
		os.Exit(1)
	}

	// Show git diff
	cmd := exec.Command("git", "diff")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func handleFiles(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: agent files <task-id>")
		os.Exit(1)
	}

	// Show modified files
	cmd := exec.Command("git", "status", "--short")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func handleCancel(args []string) {
	fmt.Println("❌ Cancel not yet implemented")
	os.Exit(1)
}

func handleRetry(args []string) {
	fmt.Println("❌ Retry not yet implemented")
	os.Exit(1)
}

func handleMetrics(args []string) {
	showMetrics()
}

func showMetrics() {
	resp, err := http.Get(fmt.Sprintf("%s/metrics", ServerURL))
	if err != nil {
		fmt.Printf("⚠️  Could not get metrics: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var metrics map[string]interface{}
	json.Unmarshal(body, &metrics)

	fmt.Println()
	fmt.Println("📊 Server Metrics:")
	fmt.Printf("   Tasks spawned: %.0f\n", metrics["tasks_spawned"])
	fmt.Printf("   In progress: %.0f\n", metrics["tasks_in_progress"])
	fmt.Printf("   Completed: %.0f\n", metrics["tasks_completed"])
	fmt.Printf("   Failed: %.0f\n", metrics["tasks_failed"])
	fmt.Println()
}

func handlePerformance(args []string) {
	resp, err := http.Get(fmt.Sprintf("%s/metrics", ServerURL))
	if err != nil {
		fmt.Printf("❌ Could not fetch metrics: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var metrics map[string]interface{}
	if err := json.Unmarshal(body, &metrics); err != nil {
		fmt.Printf("❌ Failed to parse metrics: %v\n", err)
		os.Exit(1)
	}

	// Display performance report
	fmt.Println("======================================================================")
	fmt.Println("A2A AGENT SERVER PERFORMANCE REPORT")
	fmt.Println("======================================================================")
	fmt.Printf("Generated: %v\n", time.Now().Format(time.RFC3339))
	fmt.Printf("Server: %s\n", ServerURL)
	fmt.Println()

	// Server Metrics
	fmt.Println("📊 SERVER METRICS")
	fmt.Println("----------------------------------------------------------------------")
	fmt.Printf("  Tasks Spawned:     %.0f\n", metrics["tasks_spawned"])
	fmt.Printf("  Tasks Completed:   %.0f ✅\n", metrics["tasks_completed"])
	fmt.Printf("  Tasks Failed:      %.0f\n", metrics["tasks_failed"])
	fmt.Printf("  Tasks In Progress: %.0f\n", metrics["tasks_in_progress"])

	if avgDur, ok := metrics["avg_duration_ms"].(float64); ok && avgDur > 0 {
		fmt.Printf("  Average Duration:  %.1fs\n", avgDur/1000)
	}

	fmt.Printf("  API Calls Total:   %.0f\n", metrics["api_calls_total"])
	fmt.Printf("  API Calls Success: %.0f\n", metrics["api_calls_success"])
	fmt.Printf("  API Calls Failed:  %.0f\n", metrics["api_calls_failed"])

	if apiTotal, ok := metrics["api_calls_total"].(float64); ok && apiTotal > 0 {
		apiSuccess, _ := metrics["api_calls_success"].(float64)
		successRate := (apiSuccess / apiTotal) * 100
		fmt.Printf("  Success Rate:      %.1f%%\n", successRate)
	}

	fmt.Printf("  Rate Limit Hits:   %.0f\n", metrics["rate_limit_violations"])
	fmt.Println()

	// Token Usage
	totalInput, hasInput := metrics["total_input_tokens"].(float64)
	totalOutput, hasOutput := metrics["total_output_tokens"].(float64)
	tasksCompleted, _ := metrics["tasks_completed"].(float64)

	if hasInput && hasOutput && tasksCompleted > 0 {
		fmt.Println("💾 TOKEN USAGE")
		fmt.Println("----------------------------------------------------------------------")
		fmt.Printf("  Total Input:       %s tokens\n", formatNumber(int64(totalInput)))
		fmt.Printf("  Total Output:      %s tokens\n", formatNumber(int64(totalOutput)))
		fmt.Printf("  Avg Input/Task:    %s tokens\n", formatNumber(int64(totalInput/tasksCompleted)))
		fmt.Printf("  Avg Output/Task:   %s tokens\n", formatNumber(int64(totalOutput/tasksCompleted)))

		if totalOutput > 0 {
			ratio := totalInput / totalOutput
			fmt.Printf("  Input/Output Ratio: %.1f:1\n", ratio)

			// Cache efficiency heuristic
			if ratio > 50 {
				fmt.Println("  Cache Efficiency:  Excellent (likely caching)")
			} else if ratio > 20 {
				fmt.Println("  Cache Efficiency:  Good")
			} else {
				fmt.Println("  Cache Efficiency:  Low (check caching)")
			}
		}
		fmt.Println()
	}

	// Per-Turn Statistics (if available)
	totalTurns, hasTurns := metrics["total_turns"].(float64)
	avgInputPerTurn, hasAvgInput := metrics["avg_input_per_turn"].(float64)
	avgOutputPerTurn, hasAvgOutput := metrics["avg_output_per_turn"].(float64)

	if hasTurns && hasAvgInput && hasAvgOutput && totalTurns > 0 {
		fmt.Println("🔄 PER-TURN AVERAGES")
		fmt.Println("----------------------------------------------------------------------")
		fmt.Printf("  Total Turns:       %s\n", formatNumber(int64(totalTurns)))
		fmt.Printf("  Avg Input/Turn:    %s tokens\n", formatNumber(int64(avgInputPerTurn)))
		fmt.Printf("  Avg Output/Turn:   %s tokens\n", formatNumber(int64(avgOutputPerTurn)))

		if avgOutputPerTurn > 0 {
			turnRatio := avgInputPerTurn / avgOutputPerTurn
			fmt.Printf("  Avg Turn Ratio:    %.1f:1\n", turnRatio)
		}
		fmt.Println()

		// Recent Turn Details (last 10 turns)
		if turnDataRaw, ok := metrics["turn_token_data"].([]interface{}); ok && len(turnDataRaw) > 0 {
			fmt.Printf("📊 RECENT TURNS (last %d)\n", min(10, len(turnDataRaw)))
			fmt.Println("----------------------------------------------------------------------")
			fmt.Println("  Turn  Duration  Input      Output     Task")
			fmt.Println("  ----  --------  ---------  ---------  ----")

			startIdx := 0
			if len(turnDataRaw) > 10 {
				startIdx = len(turnDataRaw) - 10
			}

			for _, turnRaw := range turnDataRaw[startIdx:] {
				turn, ok := turnRaw.(map[string]interface{})
				if !ok {
					continue
				}

				turnNum, _ := turn["Turn"].(float64)
				durationMs, _ := turn["DurationMs"].(float64)
				inputTokens, _ := turn["InputTokens"].(float64)
				outputTokens, _ := turn["OutputTokens"].(float64)
				taskID, _ := turn["TaskID"].(string)

				// Truncate task ID for display
				displayTaskID := taskID
				if len(displayTaskID) > 20 {
					displayTaskID = displayTaskID[:17] + "..."
				}

				fmt.Printf("  %-4.0f  %6.1fs  %9s  %9s  %s\n",
					turnNum,
					durationMs/1000,
					formatNumber(int64(inputTokens)),
					formatNumber(int64(outputTokens)),
					displayTaskID)
			}
			fmt.Println()
		}
	}

	// Recent Sessions (if available)
	if taskUsageRaw, ok := metrics["task_token_usage"].([]interface{}); ok && len(taskUsageRaw) > 0 {
		fmt.Printf("📋 RECENT SESSIONS (last %d)\n", len(taskUsageRaw))
		fmt.Println("----------------------------------------------------------------------")

		for i, taskRaw := range taskUsageRaw {
			if i >= 10 { // Show max 10
				break
			}

			task, ok := taskRaw.(map[string]interface{})
			if !ok {
				continue
			}

			taskID, _ := task["TaskID"].(string)
			inputTokens, _ := task["InputTokens"].(float64)
			outputTokens, _ := task["OutputTokens"].(float64)
			turnCount, _ := task["TurnCount"].(float64)

			fmt.Printf("  %s\n", taskID)
			fmt.Printf("    Turns: %.0f | Input: %s | Output: %s\n",
				turnCount,
				formatNumber(int64(inputTokens)),
				formatNumber(int64(outputTokens)))
		}
		fmt.Println()
	}

	fmt.Println("======================================================================")
}

func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%s", humanizeNumber(n))
}

func humanizeNumber(n int64) string {
	// Simple comma formatting
	str := fmt.Sprintf("%d", n)
	var result string
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}

func handleDiscovery(args []string) {
	// Parse flags
	fs := flag.NewFlagSet("discovery", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output as JSON")
	verboseOutput := fs.Bool("verbose", false, "Show detailed information")
	fs.Usage = func() {
		fmt.Println("Usage: agent discovery [options]")
		fmt.Println()
		fmt.Println("Options:")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	resp, err := http.Get(fmt.Sprintf("%s/a2a/discovery", ServerURL))
	if err != nil {
		fmt.Printf("❌ Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if *jsonOutput {
		fmt.Println(string(body))
		return
	}

	var discovery map[string]interface{}
	if err := json.Unmarshal(body, &discovery); err != nil {
		fmt.Printf("❌ Failed to parse discovery response: %v\n", err)
		os.Exit(1)
	}

	// Display discovery info
	fmt.Println("======================================================================")
	fmt.Println("AI-PACK AGENT SERVER DISCOVERY")
	fmt.Println("======================================================================")
	fmt.Println()

	// Server info
	fmt.Printf("Server: %s\n", discovery["name"])
	fmt.Printf("Version: %s\n", discovery["version"])
	if desc, ok := discovery["description"].(string); ok {
		fmt.Printf("Description: %s\n", desc)
	}
	fmt.Println()

	// Capabilities
	if caps, ok := discovery["capabilities"].(map[string]interface{}); ok {
		fmt.Println("🎯 CAPABILITIES")
		fmt.Println("----------------------------------------------------------------------")
		if streaming, ok := caps["streaming"].(bool); ok && streaming {
			fmt.Println("  ✓ SSE Streaming")
		}
		if parallel, ok := caps["parallel"].(bool); ok && parallel {
			fmt.Println("  ✓ Parallel Execution")
		}
		if maxConcurrent, ok := caps["max_concurrent"].(float64); ok {
			fmt.Printf("  ✓ Max Concurrent: %.0f agents\n", maxConcurrent)
		}
		if models, ok := caps["supported_models"].([]interface{}); ok && len(models) > 0 {
			fmt.Printf("  ✓ Supported Models: ")
			for i, model := range models {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(model)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	// Available agents
	if agents, ok := discovery["agents"].([]interface{}); ok {
		fmt.Printf("🤖 AVAILABLE AGENTS (%d)\n", len(agents))
		fmt.Println("----------------------------------------------------------------------")
		for _, agent := range agents {
			if agentMap, ok := agent.(map[string]interface{}); ok {
				role, _ := agentMap["role"].(string)
				description, _ := agentMap["description"].(string)

				fmt.Printf("\n  • %s\n", role)
				if description != "" {
					// Wrap description at 68 chars for readability
					wrapped := wrapText(description, 68, "    ")
					fmt.Print(wrapped)
				}

				if *verboseOutput {
					if tools, ok := agentMap["tools"].([]interface{}); ok && len(tools) > 0 {
						fmt.Print("    Tools: ")
						for i, tool := range tools {
							if i > 0 {
								fmt.Print(", ")
							}
							fmt.Print(tool)
						}
						fmt.Println()
					}
					if timeout, ok := agentMap["timeout"].(string); ok && timeout != "" {
						fmt.Printf("    Timeout: %s\n", timeout)
					}
				}
			}
		}
		fmt.Println()
	}

	// Endpoints
	if endpoints, ok := discovery["endpoints"].(map[string]interface{}); ok {
		fmt.Printf("📡 ENDPOINTS (%d)\n", len(endpoints))
		fmt.Println("----------------------------------------------------------------------")
		for name, endpoint := range endpoints {
			if ep, ok := endpoint.(map[string]interface{}); ok {
				path, _ := ep["path"].(string)
				method, _ := ep["method"].(string)
				desc, _ := ep["description"].(string)

				fmt.Printf("\n  %s\n", name)
				fmt.Printf("    Path: %s\n", path)
				fmt.Printf("    Method: %s\n", method)
				if desc != "" {
					fmt.Printf("    Description: %s\n", desc)
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("======================================================================")
}

func wrapText(text string, width int, indent string) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return indent + "\n"
	}

	var result strings.Builder
	result.WriteString(indent)
	lineLen := 0

	for i, word := range words {
		wordLen := len(word)
		if i > 0 {
			if lineLen+wordLen+1 > width {
				result.WriteString("\n" + indent)
				lineLen = 0
			} else {
				result.WriteString(" ")
				lineLen++
			}
		}
		result.WriteString(word)
		lineLen += wordLen
	}
	result.WriteString("\n")

	return result.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// truncateDescription takes a multi-line description and returns only the first line, truncated if needed
func truncateDescription(desc string, maxLen int) string {
	// Get only the first line (before any newline)
	lines := strings.Split(desc, "\n")
	firstLine := strings.TrimSpace(lines[0])

	// Truncate if too long
	if len(firstLine) > maxLen {
		return firstLine[:maxLen-3] + "..."
	}

	return firstLine
}

func streamTaskProgress(beadsTaskID string) {
	// Try to find task ID by querying server first (for cross-project tasks)
	internalTaskID := findInternalTaskIDFromServer(beadsTaskID)

	// Fallback: Look up internal task ID from local filesystem
	if internalTaskID == "" {
		internalTaskID = findInternalTaskID(beadsTaskID)
	}
	if internalTaskID == "" {
		// Check if this is a Beads task that's already completed
		if isValidBeadsTask(beadsTaskID) {
			cmd := exec.Command("bd", "show", beadsTaskID, "--json")
			output, err := cmd.Output()
			if err == nil {
				var beadsTask map[string]interface{}
				if json.Unmarshal(output, &beadsTask) == nil {
					if status, ok := beadsTask["status"].(string); ok {
						if status == "closed" || status == "completed" {
							fmt.Printf("✅ Task already completed: %s\n", beadsTaskID)
							os.Exit(0)
						}
					}
				}
			}
		}
		fmt.Printf("❌ No agent found for Beads task: %s\n", beadsTaskID)
		fmt.Printf("   Tip: Check 'agent list' for active agents or 'bd show %s' for task status\n", beadsTaskID)
		os.Exit(1)
	}

	// Connect to SSE stream endpoint
	streamURL := fmt.Sprintf("%s/stream/%s", ServerURL, internalTaskID)
	resp, err := http.Get(streamURL)
	if err != nil {
		fmt.Printf("❌ Failed to connect to stream: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("❌ Stream connection failed with status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Printf("✓ Connected to task stream: %s\n", internalTaskID)
	fmt.Println()

	// Read SSE events line by line
	// SSE format: "data: {...}\n\n"
	buffer := make([]byte, 8192)
	dataBuffer := ""

	for {
		n, err := resp.Body.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("⚠️  Stream read error: %v\n", err)
			break
		}

		dataBuffer += string(buffer[:n])

		// Process complete SSE messages (terminated by "\n\n")
		for {
			idx := strings.Index(dataBuffer, "\n\n")
			if idx == -1 {
				break
			}

			message := dataBuffer[:idx]
			dataBuffer = dataBuffer[idx+2:]

			// Parse SSE event
			if strings.HasPrefix(message, sseDataPrefix) {
				jsonData := strings.TrimPrefix(message, sseDataPrefix)

				var event map[string]interface{}
				if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
					continue
				}

				// Display event based on type
				eventType, _ := event["type"].(string)
				timestamp, _ := event["timestamp"].(string)
				data, _ := event["data"].(map[string]interface{})

				switch eventType {
				case "status_update":
					status, _ := data["status"].(string)
					progress, _ := data["progress"].(float64)
					fmt.Printf("[%s] Status: %s (%.0f%%)\n", timestamp, status, progress*100)

				case "api_call_start":
					fmt.Printf("[%s] 🔄 API call starting...\n", timestamp)

				case "api_call_complete":
					fmt.Printf("[%s] ✅ API call complete\n", timestamp)

				case "completed":
					fmt.Printf("[%s] 🎉 Task completed!\n", timestamp)
					fmt.Println()
					showMetrics()
					return

				case "failed":
					errorMsg, _ := data["error"].(string)
					fmt.Printf("[%s] ❌ Task failed: %s\n", timestamp, errorMsg)
					os.Exit(1)

				default:
					// Log unknown events for debugging
					fmt.Printf("[%s] %s\n", timestamp, eventType)
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("✓ Stream closed")
	showMetrics()
}

func waitForTaskCompletion(beadsTaskID string) {
	// Look up internal task ID from Beads task ID
	internalTaskID := findInternalTaskID(beadsTaskID)
	if internalTaskID == "" {
		// Check if this is a Beads task that's already completed
		if isValidBeadsTask(beadsTaskID) {
			cmd := exec.Command("bd", "show", beadsTaskID, "--json")
			output, err := cmd.Output()
			if err == nil {
				var beadsTask map[string]interface{}
				if json.Unmarshal(output, &beadsTask) == nil {
					if status, ok := beadsTask["status"].(string); ok {
						if status == "closed" || status == "completed" {
							fmt.Printf("✅ Task already completed: %s\n", beadsTaskID)
							os.Exit(0)
						}
					}
				}
			}
		}
		fmt.Printf("❌ No agent found for Beads task: %s\n", beadsTaskID)
		fmt.Printf("   Tip: Check 'agent list' for active agents or 'bd show %s' for task status\n", beadsTaskID)
		os.Exit(1)
	}

	for {
		resp, err := http.Get(fmt.Sprintf("%s/a2a/status/%s", ServerURL, internalTaskID))
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var status map[string]interface{}
		if err := json.Unmarshal(body, &status); err != nil {
			fmt.Printf("⚠️  Failed to parse status response: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Check if status field exists and is a string
		statusVal, ok := status["status"]
		if !ok || statusVal == nil {
			// Task might not be registered yet, wait and retry
			time.Sleep(2 * time.Second)
			continue
		}

		statusStr, ok := statusVal.(string)
		if !ok {
			fmt.Printf("⚠️  Unexpected status type: %v\n", statusVal)
			time.Sleep(2 * time.Second)
			continue
		}

		if statusStr == "completed" {
			fmt.Println("✅ Agent completed!")
			showMetrics()
			break
		} else if statusStr == "failed" {
			fmt.Println("❌ Agent failed!")
			if status["error"] != nil {
				fmt.Printf("   Error: %v\n", status["error"])
			}
			break
		}

		time.Sleep(5 * time.Second)
	}
}

func findInternalTaskIDFromServer(beadsTaskID string) string {
	taskID, _ := findTaskIDAndProjectFromServer(beadsTaskID)
	return taskID
}

func findTaskIDAndProjectFromServer(beadsTaskID string) (string, string) {
	// Query server's /a2a/tasks endpoint for the beads task ID
	resp, err := http.Get(fmt.Sprintf("%s/a2a/tasks", ServerURL))
	if err != nil {
		return "", "" // Server not available, will fallback to local search
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", ""
	}

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Tasks []struct {
			TaskID      string `json:"task_id"`
			BeadsTaskID string `json:"beads_task_id"`
			ProjectRoot string `json:"project_root"`
		} `json:"tasks"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", ""
	}

	// Find matching beads task ID
	for _, task := range result.Tasks {
		if task.BeadsTaskID == beadsTaskID {
			return task.TaskID, task.ProjectRoot
		}
	}

	return "", ""
}

func findInternalTaskID(beadsTaskID string) string {
	// First try: Search server's .beads/tasks/ directory
	// (Server creates tasks in its own working directory)
	serverTasksDir := "/Users/bryanw/Projects/Vibe/ai-pack/a2a-agent/.beads/tasks"
	taskID := searchTasksDir(serverTasksDir, beadsTaskID)
	if taskID != "" {
		return taskID
	}

	// Fallback: Search current directory's .beads/tasks/
	taskID = searchTasksDir(".beads/tasks", beadsTaskID)
	return taskID
}

func searchTasksDir(tasksDir string, beadsTaskID string) string {
	matches, _ := filepath.Glob(filepath.Join(tasksDir, "task-*"))
	for _, taskDir := range matches {
		metaFile := filepath.Join(taskDir, "00-metadata.json")
		data, err := os.ReadFile(metaFile)
		if err != nil {
			continue
		}

		var meta map[string]interface{}
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}

		// Check metadata.beads_task_id (new location)
		if metadata, ok := meta["metadata"].(map[string]interface{}); ok {
			if btid, ok := metadata["beads_task_id"].(string); ok && btid == beadsTaskID {
				return filepath.Base(taskDir)
			}
		}

		// Fallback: Check config.metadata.beads_task_id (old location)
		if config, ok := meta["config"].(map[string]interface{}); ok {
			if configMeta, ok := config["metadata"].(map[string]interface{}); ok {
				if btid, ok := configMeta["beads_task_id"].(string); ok && btid == beadsTaskID {
					return filepath.Base(taskDir)
				}
			}
		}
	}
	return ""
}

func isValidBeadsTask(taskID string) bool {
	if !strings.Contains(taskID, "-") && !strings.Contains(taskID, "+") {
		return false
	}

	cmd := exec.Command("bd", "show", taskID)
	err := cmd.Run()
	return err == nil
}

func mustGetWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "/path/to/project"
	}
	return wd
}

// detectProjectRoot detects the project root using git
// Tries in order:
// 1. git rev-parse --show-superproject-working-tree (for submodules)
// 2. git rev-parse --show-toplevel (for regular repos)
// 3. Current working directory (fallback)
func detectProjectRoot() string {
	// Try superproject working tree first (for submodules)
	cmd := exec.Command("git", "rev-parse", "--show-superproject-working-tree")
	output, err := cmd.Output()
	if err == nil {
		root := strings.TrimSpace(string(output))
		if root != "" {
			return root
		}
	}

	// Try regular git root
	cmd = exec.Command("git", "rev-parse", "--show-toplevel")
	output, err = cmd.Output()
	if err == nil {
		root := strings.TrimSpace(string(output))
		if root != "" {
			return root
		}
	}

	// Fallback to current directory
	return ""
}

func openURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Run()
}

func usage() {
	fmt.Println("AI-Pack Agent CLI v" + Version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  agent <role> <beads-task-id> [--stream|--wait]      Spawn an agent")
	fmt.Println("  agent status <task-id> [--json] [--quiet]           Check agent status")
	fmt.Println("  agent results <task-id>                             Show agent results")
	fmt.Println("  agent logs <task-id> [--tail N] [--follow] [--json] Show agent logs")
	fmt.Println("  agent logs --server [--tail N] [--follow] [--json]  Show server logs")
	fmt.Println("  agent list [--all|--running|--completed|--failed] [--verbose|--json|--server] List agents")
	fmt.Println("  agent wait <task-id>                                Wait for completion")
	fmt.Println("  agent diff <task-id>                                Show git diff")
	fmt.Println("  agent files <task-id>                               List modified files")
	fmt.Println("  agent metrics                                       Show server metrics")
	fmt.Println("  agent performance                                   Performance report with token usage")
	fmt.Println("  agent discovery [--verbose] [--json]                Show server capabilities and available agents")
	fmt.Println("  agent version                                       Show version")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Create Beads task with proper format")
	fmt.Println("  bd create \"Feature title")
	fmt.Println()
	fmt.Println("  Working directory: /Users/you/Projects/project")
	fmt.Println("  Task packet: .ai/tasks/2026-01-24_feature/")
	fmt.Println()
	fmt.Println("  Description...\" --priority high")
	fmt.Println()
	fmt.Println("  # Spawn and stream real-time progress (RECOMMENDED)")
	fmt.Println("  agent engineer xasm++-vp5 --stream")
	fmt.Println()
	fmt.Println("  # Check status with JSON output")
	fmt.Println("  agent status xasm++-vp5 --json")
	fmt.Println()
	fmt.Println("  # Tail task logs")
	fmt.Println("  agent logs xasm++-vp5 --tail 50")
	fmt.Println()
	fmt.Println("  # Follow server logs in real-time")
	fmt.Println("  agent logs --server --follow")
	fmt.Println()
	fmt.Println("  # List active agents (default: running + queued only)")
	fmt.Println("  agent list")
	fmt.Println()
	fmt.Println("  # List all agents including completed/failed")
	fmt.Println("  agent list --all")
	fmt.Println()
	fmt.Println("  # List all agents machine-wide (query server)")
	fmt.Println("  agent list --server")
	fmt.Println()
	fmt.Println("  # List with full details")
	fmt.Println("  agent list --verbose")
	fmt.Println()
	fmt.Println("  # List running agents as JSON")
	fmt.Println("  agent list --running --json")
	fmt.Println()
}
