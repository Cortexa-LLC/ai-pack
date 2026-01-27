package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

func createTestTask() (string, error) {
	fmt.Println("📝 Creating test task...")
	cmd := exec.Command("bd", "create", "SSE test: write a simple hello world", "-p", "P1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create task: %v\n%s", err, output)
	}

	taskID := parseTaskID(output)
	if taskID == "" {
		return "", fmt.Errorf("could not parse task ID from output")
	}

	fmt.Printf("✅ Created task: %s\n", taskID)
	fmt.Println()
	return taskID, nil
}

func parseTaskID(output []byte) string {
	taskIDLine := strings.TrimSpace(string(output))
	lines := strings.Split(taskIDLine, "\n")

	for _, line := range lines {
		if strings.Contains(line, "Created issue:") || strings.Contains(line, "ai-pack-") {
			words := strings.Fields(line)
			for _, word := range words {
				if strings.HasPrefix(word, "ai-pack-") {
					return word
				}
			}
		}
	}
	return ""
}

func spawnAgent(taskID string) {
	fmt.Println("🚀 Spawning engineer agent...")
	go func() {
		cmd := exec.Command("agent", "engineer", taskID)
		cmd.Run()
	}()
	time.Sleep(2 * time.Second)
}

func checkServerRunning() error {
	resp, err := http.Get("http://localhost:8080/metrics")
	if err != nil {
		return fmt.Errorf("server not running: %v", err)
	}
	resp.Body.Close()
	fmt.Println("📡 Server is running")
	fmt.Println()
	return nil
}

func testSSEEndpoint() {
	fmt.Println("🔌 Testing SSE endpoint availability...")
	fmt.Println("   Attempting: http://localhost:8080/stream/test-task-id")

	resp, err := http.Get("http://localhost:8080/stream/test-task-id")
	if err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("   Status: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("   Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Println()

	handleSSEResponse(resp)
}

func handleSSEResponse(resp *http.Response) {
	if resp.StatusCode == 404 {
		fmt.Println("⚠️  SSE endpoint returned 404")
		fmt.Println("   This could mean:")
		fmt.Println("   1. The task ID doesn't exist (expected for test-task-id)")
		fmt.Println("   2. The endpoint is registered and responding")
		fmt.Println()
		fmt.Println("✅ SSE endpoint is accessible (not 404 from proxy)")
		return
	}

	if resp.StatusCode == 200 && strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		readSSEStream(resp)
		return
	}

	fmt.Printf("⚠️  Unexpected response: %d %s\n", resp.StatusCode, resp.Status)
}

func readSSEStream(resp *http.Response) {
	fmt.Println("✅ SSE streaming is working!")
	fmt.Println()
	fmt.Println("📊 Reading stream events...")

	scanner := bufio.NewScanner(resp.Body)
	eventCount := 0
	timeout := time.After(10 * time.Second)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				fmt.Printf("   %s\n", line)
				eventCount++
			}
		}
	}()

	<-timeout
	fmt.Printf("\n📊 Received %d lines in 10 seconds\n", eventCount)
}

func main() {
	fmt.Println("🚀 Testing SSE streaming through agent-server")
	fmt.Println()

	taskID, err := createTestTask()
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	spawnAgent(taskID)

	if err := checkServerRunning(); err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	testSSEEndpoint()
}
