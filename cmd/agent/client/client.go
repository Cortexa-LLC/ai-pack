// Package client provides an HTTP client for the ai-pack agent server.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// ServerURL is the default agent server base URL.
const ServerURL = "http://localhost:8080"

// SSEDataPrefix is the standard Server-Sent Events data line prefix ("data: ").
const SSEDataPrefix = "data: "

// ParseSSELine checks whether line is an SSE data line. If so it returns the
// payload (everything after the "data: " prefix) and true. Otherwise it returns
// "", false. Trailing carriage-returns are stripped before the check.
func ParseSSELine(line string) (data string, ok bool) {
	// Normalise Windows-style line endings.
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}
	if len(line) >= len(SSEDataPrefix) && line[:len(SSEDataPrefix)] == SSEDataPrefix {
		return line[len(SSEDataPrefix):], true
	}
	return "", false
}

// Client wraps the HTTP connection to the agent server.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// New creates a Client using the given base URL.
func New(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

// Default returns a Client pointed at the default server URL.
func Default() *Client {
	return New(ServerURL)
}

// Get performs a GET request to path and returns the raw response body.
// Caller is responsible for checking status codes as needed.
func (c *Client) Get(path string) (*http.Response, error) {
	return c.HTTPClient.Get(c.BaseURL + path)
}

// PostJSON marshals body to JSON, POSTs it to path, and returns the raw response.
func (c *Client) PostJSON(path string, body interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	return c.HTTPClient.Post(c.BaseURL+path, "application/json", bytes.NewBuffer(jsonData))
}

// ReadBody reads the full response body and closes it.
func ReadBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// RequireOK checks that the response status is 200. On failure it prints the error
// body and calls os.Exit(1).
func RequireOK(resp *http.Response, action string) []byte {
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ %s: %s\n", action, string(body))
		os.Exit(1)
	}
	return body
}

// TaskInfo is the minimal shape returned by /a2a/tasks.
type TaskInfo struct {
	TaskID      string `json:"task_id"`
	BeadsTaskID string `json:"beads_task_id"`
	ProjectRoot string `json:"project_root"`
}

// TaskListResponse is the envelope returned by GET /a2a/tasks.
type TaskListResponse struct {
	Tasks []TaskInfo `json:"tasks"`
}

// FetchTaskResults queries the server for a task's 30-results.md content.
// It returns the result string and true when successful; empty string and false
// when the server is unreachable, returns a non-200 status, or the task has no
// results yet.
func (c *Client) FetchTaskResults(internalTaskID string) (result string, ok bool) {
	resp, err := c.Get("/a2a/tasks/" + internalTaskID + "/results")
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", false
	}
	body, _ := io.ReadAll(resp.Body)
	var r struct {
		TaskID string `json:"task_id"`
		Result string `json:"result"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return "", false
	}
	return r.Result, true
}

// FindTaskByBeadsID queries the server for a task matching beadsTaskID and returns
// (internalTaskID, projectRoot). Both are empty strings when not found.
func (c *Client) FindTaskByBeadsID(beadsTaskID string) (taskID, projectRoot string) {
	resp, err := c.Get("/a2a/tasks")
	if err != nil {
		return "", ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", ""
	}
	body, _ := io.ReadAll(resp.Body)
	var result TaskListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", ""
	}
	for _, t := range result.Tasks {
		if t.BeadsTaskID == beadsTaskID {
			return t.TaskID, t.ProjectRoot
		}
	}
	return "", ""
}
