package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	agentclient "github.com/cortexa-llc/ai-pack/cmd/agent/client"
	"github.com/spf13/cobra"
)

const descOutputAsJSON = "Output as JSON"

func newLogsCmd() *cobra.Command {
	var tailLines int
	var follow bool
	var serverLogs bool
	var allLogs bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "logs [beads-task-id]",
		Short: "Show or stream task / server logs",
		Long: `Show logs for a task, stream the server log, or show all logs.

Examples:
  agent logs xasm++-vp5
  agent logs xasm++-vp5 --follow
  agent logs --server --follow
  agent logs --all`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogs(args, tailLines, follow, serverLogs, allLogs, jsonOutput)
		},
	}

	cmd.Flags().IntVar(&tailLines, "tail", 0, "Show last N lines")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Stream new log lines")
	cmd.Flags().BoolVar(&serverLogs, "server", false, "Show server logs instead of task logs")
	cmd.Flags().BoolVar(&allLogs, "all", false, "Show all logs (server + all tasks)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, descOutputAsJSON)
	return cmd
}

func runLogs(args []string, tailLines int, follow, serverLogs, allLogs, jsonOutput bool) error {
	if serverLogs {
		if follow {
			streamServerLogs(jsonOutput)
		} else {
			fetchRecentServerLogs(tailLines, jsonOutput)
		}
		return nil
	}

	if allLogs {
		if follow {
			fmt.Println("❌ --all with --follow not yet supported")
			os.Exit(1)
		}
		fetchAllLogs(tailLines, jsonOutput)
		return nil
	}

	if len(args) < 1 {
		return fmt.Errorf("usage: agent logs <beads-task-id> [options]\n       agent logs --server")
	}

	beadsTaskID := args[0]
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

	logFile := filepath.Join(projectRoot, ".beads", "tasks", internalTaskID, "execution.log")
	if follow {
		followLogFile(logFile, jsonOutput)
	} else {
		displayLogFile(logFile, tailLines, jsonOutput)
	}
	return nil
}

func displayLogFile(logFile string, tailLines int, jsonOutput bool) {
	data, err := os.ReadFile(logFile)
	if err != nil {
		fmt.Printf("❌ No logs found: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(data), "\n")
	if tailLines > 0 && len(lines) > tailLines {
		lines = lines[len(lines)-tailLines:]
	}

	if jsonOutput {
		jsonData, _ := json.Marshal(lines)
		fmt.Println(string(jsonData))
	} else {
		fmt.Println(strings.Join(lines, "\n"))
	}
}

func followLogFile(logFile string, jsonOutput bool) {
	initialData, err := os.ReadFile(logFile)
	if err != nil {
		fmt.Printf("❌ No logs found: %v\n", err)
		os.Exit(1)
	}

	if !jsonOutput {
		fmt.Print(string(initialData))
	} else {
		lines := strings.Split(string(initialData), "\n")
		for _, line := range lines {
			if line != "" {
				jsonLine, _ := json.Marshal(map[string]string{"line": line})
				fmt.Println(string(jsonLine))
			}
		}
	}

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

			file.Seek(lastSize, 0) //nolint:errcheck
			newData, _ := io.ReadAll(file)
			file.Close()

			if jsonOutput {
				for _, line := range strings.Split(string(newData), "\n") {
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
			if noChangeCount > 60 {
				noChangeCount = 0
			}
		}
	}
}

func streamServerLogs(jsonOutput bool) {
	resp, err := http.Get(fmt.Sprintf("%s/logs/stream", agentclient.ServerURL))
	if err != nil {
		fmt.Printf("❌ Failed to connect to log stream: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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

			if jsonData, ok := agentclient.ParseSSELine(message); ok {
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
	if err := runLogsRecent(agentclient.DefaultBaseURL, tailLines, jsonOutput); err != nil {
		os.Exit(1)
	}
}

// runLogsRecent issues GET /logs/recent against serverBaseURL. It is extracted
// so that contract tests can pass an httptest.Server URL and verify the path.
func runLogsRecent(serverBaseURL string, tailLines int, jsonOutput bool) error {
	url := fmt.Sprintf("%s/logs/recent", serverBaseURL)
	if tailLines > 0 {
		url += fmt.Sprintf("?limit=%d", tailLines)
	}

	resp, err := http.Get(url) //nolint:gosec,noctx
	if err != nil {
		fmt.Printf("❌ Failed to fetch server logs: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	body, err := agentclient.ReadBody(resp)
	if err != nil {
		fmt.Printf("❌ Failed to read server logs: %v\n", err)
		return err
	}

	if jsonOutput {
		fmt.Println(string(body))
		return nil
	}

	var response struct {
		Logs  []map[string]interface{} `json:"logs"`
		Count int                      `json:"count"`
		Limit int                      `json:"limit"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("❌ Failed to parse logs: %v\n", err)
		return err
	}

	for _, entry := range response.Logs {
		timestamp, _ := entry["timestamp"].(string)
		level, _ := entry["level"].(string)
		msg, _ := entry["message"].(string)
		fmt.Printf("[%s] %s: %s\n", timestamp, level, msg)
	}
	return nil
}

func fetchAllLogs(tailLines int, jsonOutput bool) {
	fetchRecentServerLogs(tailLines, jsonOutput)

	matches, _ := filepath.Glob(".beads/tasks/task-*/execution.log")
	for _, logFile := range matches {
		taskID := filepath.Base(filepath.Dir(logFile))
		if !jsonOutput {
			fmt.Printf("\n=== Task: %s ===\n", taskID)
		}
		displayLogFile(logFile, tailLines, jsonOutput)
	}
}
