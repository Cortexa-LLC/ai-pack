package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var (
	serverURL     = "http://localhost:8888"
	testServerCmd = getProjectRoot() + "/bin/agent-server"
	testAgentCmd  = getProjectRoot() + "/bin/agent"
)

// getProjectRoot returns the absolute path to the project root
func getProjectRoot() string {
	// Get the absolute path of the current test file's directory
	wd, _ := os.Getwd()
	// From test/integration, go up two levels to project root
	return filepath.Join(wd, "..", "..")
}

// TestE2EOrchestratorSpawnAndWait tests the full orchestrator workflow:
// 1. Start agent-server
// 2. Create Beads task
// 3. Spawn agent via CLI
// 4. Wait for completion
// 5. Verify results
func TestE2EOrchestratorSpawnAndWait(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup test environment
	testDir := setupTestEnvironment(t)
	defer os.RemoveAll(testDir)

	// Start agent-server in background
	serverCmd := startAgentServer(t, testDir)
	defer stopAgentServer(serverCmd)

	// Wait for server to be ready
	waitForServer(t, serverURL)

	// Create a mock Beads task
	beadsTaskID := createMockBeadsTask(t, testDir, "Test orchestrator spawn")

	t.Run("SpawnAgentAndWait", func(t *testing.T) {
		// Spawn agent using CLI
		spawnCmd := exec.Command(testAgentCmd, "engineer", beadsTaskID)
		spawnCmd.Dir = testDir
		spawnOutput, err := spawnCmd.CombinedOutput()
		if err != nil {
			t.Logf("Spawn output: %s", spawnOutput)
			t.Fatalf("Failed to spawn agent: %v", err)
		}

		t.Logf("Agent spawned: %s", spawnOutput)

		// Wait for completion using agent wait
		waitCmd := exec.Command(testAgentCmd, "wait", beadsTaskID)
		waitCmd.Dir = testDir

		// Set timeout
		done := make(chan error, 1)
		go func() {
			output, err := waitCmd.CombinedOutput()
			if err != nil {
				t.Logf("Wait output: %s", output)
			}
			done <- err
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("Wait command failed: %v", err)
			}
			t.Log("Agent completed successfully")
		case <-time.After(3 * time.Minute):
			t.Fatal("Wait timed out after 3 minutes")
		}

		// Verify status
		statusCmd := exec.Command(testAgentCmd, "status", beadsTaskID)
		statusCmd.Dir = testDir
		statusOutput, err := statusCmd.CombinedOutput()
		if err != nil {
			t.Logf("Status output: %s", statusOutput)
		}

		statusStr := string(statusOutput)
		if !strings.Contains(statusStr, "completed") && !strings.Contains(statusStr, "failed") {
			t.Errorf("Expected terminal status, got: %s", statusStr)
		}
	})
}

// TestE2EOrchestratorParallelAgents tests spawning multiple agents in parallel
func TestE2EOrchestratorParallelAgents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	testDir := setupTestEnvironment(t)
	defer os.RemoveAll(testDir)

	serverCmd := startAgentServer(t, testDir)
	defer stopAgentServer(serverCmd)

	waitForServer(t, serverURL)

	// Create 3 Beads tasks for parallel execution
	tasks := []struct {
		id          string
		description string
	}{
		{createMockBeadsTask(t, testDir, "Parallel task 1"), "Task 1"},
		{createMockBeadsTask(t, testDir, "Parallel task 2"), "Task 2"},
		{createMockBeadsTask(t, testDir, "Parallel task 3"), "Task 3"},
	}

	t.Run("SpawnParallelAgents", func(t *testing.T) {
		// Spawn all agents in parallel
		for _, task := range tasks {
			spawnCmd := exec.Command(testAgentCmd, "engineer", task.id)
			spawnCmd.Dir = testDir
			output, err := spawnCmd.CombinedOutput()
			if err != nil {
				t.Logf("Spawn output for %s: %s", task.id, output)
				t.Errorf("Failed to spawn agent for %s: %v", task.id, err)
			}
			t.Logf("Spawned agent for %s", task.id)
		}

		// Verify all are running via agent list
		listCmd := exec.Command(testAgentCmd, "list", "--server")
		listCmd.Dir = testDir
		listOutput, err := listCmd.CombinedOutput()
		if err != nil {
			t.Logf("List output: %s", listOutput)
		}

		listStr := string(listOutput)
		runningCount := strings.Count(listStr, "RUNNING")
		if runningCount < 2 {
			t.Logf("List output:\n%s", listStr)
			t.Errorf("Expected at least 2 running agents, found %d", runningCount)
		}

		// Wait for all to complete
		for _, task := range tasks {
			waitCmd := exec.Command(testAgentCmd, "wait", task.id)
			waitCmd.Dir = testDir

			done := make(chan error, 1)
			go func(taskID string) {
				output, err := waitCmd.CombinedOutput()
				if err != nil {
					t.Logf("Wait output for %s: %s", taskID, output)
				}
				done <- err
			}(task.id)

			select {
			case err := <-done:
				if err != nil {
					t.Logf("Wait failed for %s", task.id)
				} else {
					t.Logf("Agent %s completed", task.id)
				}
			case <-time.After(3 * time.Minute):
				t.Logf("Wait timed out for %s after 3 minutes", task.id)
			}
		}
	})
}

// TestE2EOrchestratorStreamMode tests using --stream for real-time monitoring
func TestE2EOrchestratorStreamMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	testDir := setupTestEnvironment(t)
	defer os.RemoveAll(testDir)

	serverCmd := startAgentServer(t, testDir)
	defer stopAgentServer(serverCmd)

	waitForServer(t, serverURL)

	beadsTaskID := createMockBeadsTask(t, testDir, "Test streaming")

	t.Run("SpawnWithStream", func(t *testing.T) {
		// Spawn with --stream (blocks until completion)
		streamCmd := exec.Command(testAgentCmd, "engineer", beadsTaskID, "--stream")
		streamCmd.Dir = testDir

		done := make(chan error, 1)
		go func() {
			output, err := streamCmd.CombinedOutput()
			t.Logf("Stream output: %s", output)
			done <- err
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("Stream command failed: %v", err)
			}
			t.Log("Agent completed via stream")
		case <-time.After(3 * time.Minute):
			t.Fatal("Stream timed out after 3 minutes")
		}

		// Verify final status
		statusCmd := exec.Command(testAgentCmd, "status", beadsTaskID)
		statusCmd.Dir = testDir
		statusOutput, _ := statusCmd.CombinedOutput()

		if !strings.Contains(string(statusOutput), "completed") &&
		   !strings.Contains(string(statusOutput), "failed") {
			t.Errorf("Expected terminal status, got: %s", statusOutput)
		}
	})
}

// TestE2EOrchestratorCrossProjectCoordination tests agents in different projects
func TestE2EOrchestratorCrossProjectCoordination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Create two separate project directories
	project1 := setupTestEnvironment(t)
	project2 := setupTestEnvironment(t)
	defer os.RemoveAll(project1)
	defer os.RemoveAll(project2)

	// Start single server (machine-wide)
	serverCmd := startAgentServer(t, project1)
	defer stopAgentServer(serverCmd)

	waitForServer(t, serverURL)

	// Create tasks in different projects
	task1 := createMockBeadsTask(t, project1, "Project 1 task")
	task2 := createMockBeadsTask(t, project2, "Project 2 task")

	t.Run("CoordinateCrossProject", func(t *testing.T) {
		// Spawn agent in project 1
		spawn1Cmd := exec.Command(testAgentCmd, "engineer", task1)
		spawn1Cmd.Dir = project1
		output1, err := spawn1Cmd.CombinedOutput()
		if err != nil {
			t.Logf("Spawn 1 output: %s", output1)
			t.Fatalf("Failed to spawn in project1: %v", err)
		}

		// Spawn agent in project 2
		spawn2Cmd := exec.Command(testAgentCmd, "engineer", task2)
		spawn2Cmd.Dir = project2
		output2, err := spawn2Cmd.CombinedOutput()
		if err != nil {
			t.Logf("Spawn 2 output: %s", output2)
			t.Fatalf("Failed to spawn in project2: %v", err)
		}

		// List from project1 - should see both agents
		listCmd := exec.Command(testAgentCmd, "list", "--server")
		listCmd.Dir = project1
		listOutput, _ := listCmd.CombinedOutput()
		listStr := string(listOutput)

		if !strings.Contains(listStr, task1) || !strings.Contains(listStr, task2) {
			t.Logf("List output:\n%s", listStr)
			t.Error("Expected to see both cross-project agents in list")
		}

		// Status from project1 should find project2 task
		statusCmd := exec.Command(testAgentCmd, "status", task2)
		statusCmd.Dir = project1
		statusOutput, err := statusCmd.CombinedOutput()
		if err != nil {
			t.Logf("Status output: %s", statusOutput)
			t.Error("Failed to get status of cross-project task")
		}
	})
}

// TestE2EOrchestratorLogsFollowing tests log following
func TestE2EOrchestratorLogsFollowing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	testDir := setupTestEnvironment(t)
	defer os.RemoveAll(testDir)

	serverCmd := startAgentServer(t, testDir)
	defer stopAgentServer(serverCmd)

	waitForServer(t, serverURL)

	beadsTaskID := createMockBeadsTask(t, testDir, "Test log following")

	t.Run("FollowLogsInRealtime", func(t *testing.T) {
		// Spawn agent (don't wait)
		spawnCmd := exec.Command(testAgentCmd, "engineer", beadsTaskID)
		spawnCmd.Dir = testDir
		spawnCmd.Run()

		// Follow logs with timeout
		logsCmd := exec.Command(testAgentCmd, "logs", beadsTaskID, "--follow")
		logsCmd.Dir = testDir

		done := make(chan error, 1)
		var logsOutput []byte
		go func() {
			output, err := logsCmd.CombinedOutput()
			logsOutput = output
			done <- err
		}()

		select {
		case <-done:
			// Logs should show execution details
			logsStr := string(logsOutput)
			if len(logsStr) < 100 {
				t.Errorf("Expected substantial log output, got: %s", logsStr)
			}
			t.Logf("Logs captured: %d bytes", len(logsOutput))
		case <-time.After(15 * time.Second):
			// Timeout is okay - just stop following
			logsCmd.Process.Kill()
			t.Log("Log following stopped after timeout (expected)")
		}
	})
}

// TestE2EOrchestratorAgentDiscovery tests agent discovery command
func TestE2EOrchestratorAgentDiscovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	testDir := setupTestEnvironment(t)
	defer os.RemoveAll(testDir)

	serverCmd := startAgentServer(t, testDir)
	defer stopAgentServer(serverCmd)

	waitForServer(t, serverURL)

	t.Run("DiscoverCapabilities", func(t *testing.T) {
		// Run discovery command
		discoveryCmd := exec.Command(testAgentCmd, "discovery", "--json")
		discoveryCmd.Dir = testDir
		discoveryOutput, err := discoveryCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Discovery command failed: %v\nOutput: %s", err, discoveryOutput)
		}

		// Parse JSON
		var discovery map[string]interface{}
		if err := json.Unmarshal(discoveryOutput, &discovery); err != nil {
			t.Fatalf("Failed to parse discovery JSON: %v", err)
		}

		// Verify capabilities
		caps, ok := discovery["capabilities"].(map[string]interface{})
		if !ok {
			t.Fatal("No capabilities in discovery response")
		}

		if maxConcurrent, ok := caps["max_concurrent"].(float64); ok {
			t.Logf("Max concurrent agents: %.0f", maxConcurrent)
			if maxConcurrent < 1 {
				t.Error("Expected max_concurrent >= 1")
			}
		}

		// Verify agents are discoverable
		agents, ok := discovery["agents"].([]interface{})
		if !ok {
			t.Fatal("No agents in discovery response")
		}

		if len(agents) < 1 {
			t.Error("Expected at least 1 available agent role")
		}

		t.Logf("Discovered %d agent roles", len(agents))
	})
}

// Helper: Setup test environment
func setupTestEnvironment(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "e2e-orchestrator-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create .beads directory structure
	beadsDir := filepath.Join(tmpDir, ".beads")
	os.MkdirAll(filepath.Join(beadsDir, "tasks"), 0755)

	// Initialize empty beads database
	issuesFile := filepath.Join(beadsDir, "issues.jsonl")
	os.WriteFile(issuesFile, []byte(""), 0644)

	// Create .ai/tasks directory
	os.MkdirAll(filepath.Join(tmpDir, ".ai", "tasks"), 0755)

	return tmpDir
}

// Helper: Start agent-server
func startAgentServer(t *testing.T, workDir string) *exec.Cmd {
	// Check if server binary exists
	if _, err := os.Stat(testServerCmd); os.IsNotExist(err) {
		t.Skip("Agent server binary not found - run 'go build -o bin/agent-server ./cmd/agent-server' first")
	}

	cmd := exec.Command(testServerCmd, "--server", "--port", "8888")
	cmd.Dir = workDir
	// Inherit environment (ANTHROPIC_API_KEY or ANTHROPIC_API_TOKEN)
	cmd.Env = os.Environ()

	// Capture output for debugging
	logFile := filepath.Join(workDir, "server.log")
	outFile, err := os.Create(logFile)
	if err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}
	cmd.Stdout = outFile
	cmd.Stderr = outFile

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	t.Logf("Started agent-server (PID %d) in %s (logs: %s)", cmd.Process.Pid, workDir, logFile)
	return cmd
}

// Helper: Stop agent-server
func stopAgentServer(cmd *exec.Cmd) {
	if cmd != nil && cmd.Process != nil {
		cmd.Process.Kill()
		cmd.Wait()
	}
}

// Helper: Wait for server to be ready
func waitForServer(t *testing.T, url string) {
	// Give server a moment to start binding to port
	time.Sleep(500 * time.Millisecond)

	for i := 0; i < 30; i++ {
		checkCmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", url+"/health")
		output, _ := checkCmd.Output()

		if string(output) == "200" {
			t.Log("Server is ready")
			return
		}

		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("Server did not become ready in time")
}

// Helper: Create mock Beads task
func createMockBeadsTask(t *testing.T, dir, description string) string {
	// Generate unique task ID
	taskID := fmt.Sprintf("e2e-%d", time.Now().UnixNano()%1000000)

	// Create Beads task directory
	taskDir := filepath.Join(dir, ".beads", "tasks", fmt.Sprintf("task-engineer-%s", taskID))
	os.MkdirAll(taskDir, 0755)

	// Create metadata.json
	metadata := map[string]interface{}{
		"task_id":     taskID,
		"description": description,
		"status":      "open",
		"created_at":  time.Now().Format(time.RFC3339),
		"metadata": map[string]string{
			"beads_task_id": taskID,
		},
	}

	metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
	metadataFile := filepath.Join(taskDir, "00-metadata.json")
	os.WriteFile(metadataFile, metadataJSON, 0644)

	// Add to issues.jsonl
	issueEntry := map[string]interface{}{
		"id":          taskID,
		"title":       description,
		"status":      "open",
		"created_at":  time.Now().Format(time.RFC3339),
	}
	issueJSON, _ := json.Marshal(issueEntry)
	issuesFile := filepath.Join(dir, ".beads", "issues.jsonl")
	f, _ := os.OpenFile(issuesFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	f.WriteString(string(issueJSON) + "\n")
	f.Close()

	t.Logf("Created mock Beads task: %s - %s", taskID, description)
	return taskID
}
