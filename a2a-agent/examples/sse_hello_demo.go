package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"time"
)

// This is a standalone demo showing how to consume SSE Hello World events

func main() {
	fmt.Println("SSE Hello World Demo")
	fmt.Println("====================")
	fmt.Println()

	// Example 1: Single event
	fmt.Println("Example 1: Single hello event")
	fmt.Println("------------------------------")
	consumeSSE("http://localhost:8080/sse/hello")
	fmt.Println()

	// Example 2: Multiple events
	fmt.Println("Example 2: Three hello events")
	fmt.Println("------------------------------")
	consumeSSE("http://localhost:8080/sse/hello?count=3")
}

// consumeSSE connects to an SSE endpoint and prints events
func consumeSSE(url string) {
	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return
	}

	// Set Accept header for SSE
	req.Header.Set("Accept", "text/event-stream")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error connecting to SSE: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Verify content type
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		log.Printf("Invalid content type: %s\n", resp.Header.Get("Content-Type"))
		return
	}

	// Read events from stream
	scanner := bufio.NewScanner(resp.Body)
	var currentEvent string
	var currentData string

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line marks end of event
			if currentEvent != "" && currentData != "" {
				fmt.Printf("Event: %s\n", currentEvent)
				fmt.Printf("Data: %s\n", currentData)
				fmt.Println()

				// Close after receiving 'complete' event
				if currentEvent == "complete" {
					break
				}
			}
			currentEvent = ""
			currentData = ""
		} else if len(line) > 7 && line[:7] == "event: " {
			currentEvent = line[7:]
		} else if len(line) > 6 && line[:6] == "data: " {
			currentData = line[6:]
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading stream: %v\n", err)
	}
}
