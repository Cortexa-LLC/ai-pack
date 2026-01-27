package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// ExecuteResponse matches A2A protocol response
type ExecuteResponse struct {
	Result struct {
		TaskID string `json:"task_id"`
		Status string `json:"status"`
	} `json:"result"`
}

func main() {
	fmt.Println("🧪 SSE Streaming Test - Direct Connection")
	fmt.Println("==========================================")
	fmt.Println()

	// 1. Spawn a task
	fmt.Println("📝 Step 1: Spawning a test task via A2A protocol...")
	reqBody := `{"jsonrpc":"2.0","method":"execute_task","params":{"role":"engineer","task":"Write a simple hello world to test SSE"},"id":1}`

	resp, err := http.Post("http://localhost:8080/a2a/execute", "application/json", strings.NewReader(reqBody))
	if err != nil {
		fmt.Printf("❌ Failed to spawn task: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var execResp ExecuteResponse
	if err := json.NewDecoder(resp.Body).Decode(&execResp); err != nil {
		fmt.Printf("❌ Failed to parse response: %v\n", err)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   Response: %s\n", body)
		return
	}

	taskID := execResp.Result.TaskID
	fmt.Printf("✅ Task spawned: %s\n", taskID)
	fmt.Printf("   Status: %s\n", execResp.Result.Status)
	fmt.Println()

	// 2. Give agent a moment to start
	fmt.Println("⏳ Waiting for agent to start...")
	time.Sleep(2 * time.Second)
	fmt.Println()

	// 3. Connect to SSE stream
	fmt.Printf("📡 Step 2: Connecting to SSE stream for task %s...\n", taskID)
	streamURL := fmt.Sprintf("http://localhost:8080/stream/%s", taskID)
	fmt.Printf("   URL: %s\n", streamURL)

	req, _ := http.NewRequest("GET", streamURL, nil)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	client := &http.Client{
		Timeout: 0, // No timeout for streaming
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("❌ Failed to connect to stream: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("   Status: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("   Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Println()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Stream connection failed\n")
		fmt.Printf("   Response: %s\n", body)
		return
	}

	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		fmt.Printf("⚠️  Warning: Content-Type is not text/event-stream\n")
		fmt.Println()
	}

	// 4. Read stream events
	fmt.Println("📊 Step 3: Reading stream events...")
	fmt.Println("   (Will read for 15 seconds or until stream closes)")
	fmt.Println()

	scanner := bufio.NewScanner(resp.Body)
	eventCount := 0
	dataCount := 0
	currentEvent := ""

	// Read events for 15 seconds
	done := make(chan bool)
	go func() {
		time.Sleep(15 * time.Second)
		done <- true
	}()

	go func() {
		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "event:") {
				currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
				eventCount++
				fmt.Printf("   🔔 Event: %s\n", currentEvent)
			} else if strings.HasPrefix(line, "data:") {
				data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				dataCount++
				fmt.Printf("      Data: %s\n", data)
			} else if line == "" && currentEvent != "" {
				// Empty line marks end of event
				fmt.Println()
				currentEvent = ""
			}
		}
		done <- true
	}()

	<-done

	fmt.Println()
	fmt.Println("📊 Summary:")
	fmt.Printf("   Events received: %d\n", eventCount)
	fmt.Printf("   Data lines: %d\n", dataCount)
	fmt.Println()

	if eventCount > 0 {
		fmt.Println("✅ SSE streaming is working through the server!")
	} else {
		fmt.Println("⚠️  No events received - streaming may not be enabled or task completed too quickly")
	}

	// Clean up - kill the spawned task
	exec.Command("pkill", "-f", taskID).Run()
}
