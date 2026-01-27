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
	case "version", "--version", "-v":
		fmt.Printf("AI-Pack Agent CLI v%s\n", Version)
	case "help", "--help", "-h":
		usage()
	default:
		// Assume it's a spawn command: agent <role> <task-id>
		handleSpawn(os.Args[1:])
	}
}

func handleSpawn(args []string) {
	// Parse flags manually to allow them anywhere in arguments
	// agent engineer xasm++-m94 --wait should work
	wait := false
	stream := false
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

	role := positionalArgs[0]
	taskInput := strings.Join(positionalArgs[1:], " ")

	// Validate Beads task ID
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

	fmt.Printf("🎯 Beads task: %s\n", taskInput)

	// Execute task via HTTP API
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "execute",
		"params": map[string]string{
			"role": role,
			"task": taskInput,
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

	// Wait a moment for the task to be registered
	time.Sleep(2 * time.Second)

	// Show initial metrics
	showMetrics()

	// If --stream flag, stream real-time progress via SSE
	if stream {
		fmt.Println()
		fmt.Println("📡 Streaming real-time progress...")
		streamTaskProgress(taskInput)
		return
	}

	// If --wait flag, poll until complete
	if wait {
		fmt.Println()
		fmt.Println("⏳ Waiting for completion...")
		waitForTaskCompletion(taskInput)
	}
}

func handleStatus(args []string) {
	// Parse flags
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output as JSON")
	quiet := fs.Bool("quiet", false, "Output only status value")

	fs.Usage = func() {
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

	fs.Parse(args)
	positionalArgs := fs.Args()

	if len(positionalArgs) < 1 {
		fmt.Println("Usage: agent status <task-id> [options]")
		os.Exit(1)
	}

	beadsTaskID := positionalArgs[0]

	// Look up internal task ID from Beads task ID
	internalTaskID := findInternalTaskID(beadsTaskID)
	if internalTaskID == "" {
		if *jsonOutput {
			fmt.Println(`{"error":"not_found"}`)
		} else if !*quiet {
			fmt.Printf("❌ No agent found for Beads task: %s\n", beadsTaskID)
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
			fmt.Printf("Progress: %v\n", status["progress"])
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
	internalTaskID := findInternalTaskID(beadsTaskID)
	if internalTaskID == "" {
		fmt.Printf("❌ No agent found for Beads task: %s\n", beadsTaskID)
		fmt.Printf("   Tip: Check 'agent list' for active agents or 'bd show %s' for task status\n", beadsTaskID)
		os.Exit(1)
	}

	resultsFile := filepath.Join(".beads", "tasks", internalTaskID, "30-results.md")
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
	jsonOutput := fs.Bool("json", false, "Output as JSON")

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
	internalTaskID := findInternalTaskID(beadsTaskID)
	if internalTaskID == "" {
		fmt.Printf("❌ No agent found for Beads task: %s\n", beadsTaskID)
		fmt.Printf("   Tip: Check 'agent list' for active agents or 'bd show %s' for task status\n", beadsTaskID)
		os.Exit(1)
	}

	logFile := filepath.Join(".beads", "tasks", internalTaskID, "execution.log")

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
	}

	// Follow new lines
	lastSize := int64(len(initialData))

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

			if strings.HasPrefix(message, "data: ") {
				jsonData := strings.TrimPrefix(message, "data: ")

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
		var logs []map[string]interface{}
		if err := json.Unmarshal(body, &logs); err != nil {
			fmt.Printf("❌ Failed to parse logs: %v\n", err)
			os.Exit(1)
		}

		for _, entry := range logs {
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

func handleList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	running := fs.Bool("running", false, "Show only running agents")
	completed := fs.Bool("completed", false, "Show only completed agents")
	failed := fs.Bool("failed", false, "Show only failed agents")
	jsonOutput := fs.Bool("json", false, "Output as JSON")
	fs.Parse(args)

	// List .beads/tasks/task-* directories
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
		if *running && status != "in_progress" {
			continue
		}
		if *completed && status != "completed" {
			continue
		}
		if *failed && status != "failed" {
			continue
		}

		taskID := filepath.Base(taskDir)
		role, _ := meta["role"].(string)
		description := fmt.Sprintf("%v", meta["description"])
		beadsTaskID := ""
		if config, ok := meta["config"].(map[string]interface{}); ok {
			if md, ok := config["metadata"].(map[string]interface{}); ok {
				if btid, ok := md["beads_task_id"].(string); ok {
					beadsTaskID = btid
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
	} else {
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

func streamTaskProgress(beadsTaskID string) {
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
			if strings.HasPrefix(message, "data: ") {
				jsonData := strings.TrimPrefix(message, "data: ")

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

func findInternalTaskID(beadsTaskID string) string {
	// Search .beads/tasks/ for metadata matching this Beads task ID
	matches, _ := filepath.Glob(".beads/tasks/task-*")
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

		// Check config.metadata.beads_task_id
		if config, ok := meta["config"].(map[string]interface{}); ok {
			if metadata, ok := config["metadata"].(map[string]interface{}); ok {
				if btid, ok := metadata["beads_task_id"].(string); ok && btid == beadsTaskID {
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
	fmt.Println("  agent list [--running|--completed|--failed] [--json] List agents")
	fmt.Println("  agent wait <task-id>                                Wait for completion")
	fmt.Println("  agent diff <task-id>                                Show git diff")
	fmt.Println("  agent files <task-id>                               List modified files")
	fmt.Println("  agent metrics                                       Show server metrics")
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
	fmt.Println("  # List running agents as JSON")
	fmt.Println("  agent list --running --json")
	fmt.Println()
}
