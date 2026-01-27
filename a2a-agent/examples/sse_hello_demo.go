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

func connectToSSE(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.Header.Get("Content-Type") != "text/event-stream" {
		resp.Body.Close()
		return nil, fmt.Errorf("invalid content type: %s", resp.Header.Get("Content-Type"))
	}

	return resp, nil
}

func processSSEStream(resp *http.Response) {
	scanner := bufio.NewScanner(resp.Body)
	var currentEvent, currentData string

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			if currentEvent != "" && currentData != "" {
				fmt.Printf("Event: %s\n", currentEvent)
				fmt.Printf("Data: %s\n", currentData)
				fmt.Println()

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

// consumeSSE connects to an SSE endpoint and prints events
func consumeSSE(url string) {
	resp, err := connectToSSE(url)
	if err != nil {
		log.Printf("Error connecting to SSE: %v\n", err)
		return
	}
	defer resp.Body.Close()

	processSSEStream(resp)
}
