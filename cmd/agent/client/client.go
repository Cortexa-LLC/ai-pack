// Package client provides an HTTP client for the ai-pack agent server.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// loadServerURL reads the server URL from configuration sources.
// Priority order (highest to lowest):
// 1. AGENT_SERVER_URL environment variable
// 2. ~/.ai-pack/config.json server.host:port
// 3. Hardcoded fallback (only if config file missing)
func loadServerURL() string {
	// 1. Environment variable (highest priority)
	if v := os.Getenv("AGENT_SERVER_URL"); v != "" {
		return v
	}

	// 2. Read from config file
	if url := readConfigServerURL(); url != "" {
		return url
	}

	// 3. Fallback only if config missing (should never happen after install)
	return "http://localhost:8080"
}

// readConfigServerURL reads server URL from ~/.ai-pack/config.json
func readConfigServerURL() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	configPath := filepath.Join(home, ".ai-pack", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}

	var config struct {
		Server struct {
			Host string `json:"host"`
			Port int    `json:"port"`
		} `json:"server"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return ""
	}

	if config.Server.Host != "" && config.Server.Port > 0 {
		return fmt.Sprintf("http://%s:%d", config.Server.Host, config.Server.Port)
	}

	return ""
}

// DefaultBaseURL is the base URL used by Default().
// It is loaded from AGENT_SERVER_URL env var, or ~/.ai-pack/config.json.
// Tests may override this variable directly.
var DefaultBaseURL = loadServerURL()

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

// DefaultTimeout is the default HTTP client timeout for non-streaming requests.
const DefaultTimeout = 30 * time.Second

// New creates a Client using the given base URL.
func New(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: DefaultTimeout},
	}
}

// Default returns a Client pointed at DefaultBaseURL (normally ServerURL).
// Tests may redirect CLI calls by setting DefaultBaseURL before invoking
// a run* function.
func Default() *Client {
	return New(DefaultBaseURL)
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
// NOTE: the server serialises fields in camelCase — these tags must match.
type TaskInfo struct {
	// TaskID is the short logical task ID (e.g. "ai-pack-aa0").
	TaskID      string `json:"taskId"`
	ProjectRoot string `json:"projectRoot"`
	// LatestRunID is the most recent timestamped execution run ID.
	// When non-empty, use this for log file lookups instead of TaskID.
	LatestRunID string `json:"latestRunId"`
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

// FindTaskByShortID queries the server for a task matching taskID and returns
// (internalTaskID, projectRoot). Both are empty strings when not found.
// internalTaskID is the latest_run_id when available (for log file lookups),
// otherwise the short task_id.
func (c *Client) FindTaskByShortID(taskID string) (internalTaskID, projectRoot string) {
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
		if t.TaskID == taskID {
			// Prefer latest_run_id for log file resolution; fall back to task_id
			id := t.LatestRunID
			if id == "" {
				id = t.TaskID
			}
			return id, t.ProjectRoot
		}
	}
	return "", ""
}
