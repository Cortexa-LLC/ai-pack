// Package commands contains route contract tests that verify every CLI command
// hits the correct server endpoint with the correct HTTP method.
//
// These tests guard against regressions like the one in commit 9c84947b, where
// the "spawn" command was inadvertently changed to POST to /a2a/tasks (the
// GET-only list endpoint) instead of /a2a/execute, causing silent 405 failures.
//
// Each test:
//  1. Starts a local httptest.Server that records the method + path of every
//     request it receives.
//  2. Overrides agentclient.DefaultBaseURL to point at that server.
//  3. Invokes the run* function under test.
//  4. Asserts that the recorded method+path matches the expected contract.
package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	agentclient "github.com/cortexa-llc/ai-pack/cmd/agent/client"
)

// capturedRequest holds the method and path of a single HTTP request.
type capturedRequest struct {
	Method string
	Path   string
}

// requestCaptor records every HTTP request that arrives at a test server.
type requestCaptor struct {
	mu       sync.Mutex
	requests []capturedRequest
}

func (rc *requestCaptor) add(method, path string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.requests = append(rc.requests, capturedRequest{Method: method, Path: path})
}

// first returns the first captured request, or an error if none were captured.
func (rc *requestCaptor) first() (capturedRequest, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if len(rc.requests) == 0 {
		return capturedRequest{}, fmt.Errorf("no requests captured")
	}
	return rc.requests[0], nil
}

// newTestServer starts an httptest.Server that:
//   - records every request via captor
//   - returns a minimal valid JSON body so callers don't crash on parse
func newTestServer(t *testing.T, captor *requestCaptor, responseBody string, statusCode int) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captor.add(r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if responseBody != "" {
			fmt.Fprint(w, responseBody)
		}
	}))
	t.Cleanup(ts.Close)
	return ts
}

// redirectToTestServer overrides agentclient.DefaultBaseURL to the given server
// URL and returns a restore function. Use with defer:
//
//	defer redirectToTestServer(t, ts.URL)()
func redirectToTestServer(t *testing.T, baseURL string) (restore func()) {
	t.Helper()
	orig := agentclient.DefaultBaseURL
	agentclient.DefaultBaseURL = baseURL
	return func() { agentclient.DefaultBaseURL = orig }
}

// ──────────────────────────────────────────────────────────────────────────────
// spawn → POST /a2a/execute
// ──────────────────────────────────────────────────────────────────────────────

// TestSpawnUsesPostExecute is the primary regression guard for commit 9c84947b.
// It asserts that runSpawnHTTP POSTs to /a2a/execute, NOT to /a2a/tasks.
func TestSpawnUsesPostExecute(t *testing.T) {
	captor := &requestCaptor{}

	// Minimal valid A2A execute response body
	body, _ := json.Marshal(map[string]interface{}{
		"task_id": "test-task-id",
		"status":  map[string]string{"state": "submitted"},
	})
	ts := newTestServer(t, captor, string(body), http.StatusOK)
	defer redirectToTestServer(t, ts.URL)()

	_, _ = runSpawnHTTP("engineer", "test task description", "test-project-root")

	got, err := captor.first()
	if err != nil {
		t.Fatalf("spawn: no HTTP request captured — did the function exit before making the call? %v", err)
	}

	if got.Method != http.MethodPost {
		t.Errorf("spawn: expected method POST, got %s", got.Method)
	}
	if got.Path != "/a2a/execute" {
		t.Errorf("spawn: expected path /a2a/execute, got %q — regression: must not POST to /a2a/tasks", got.Path)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// list → GET /a2a/tasks
// ──────────────────────────────────────────────────────────────────────────────

func TestListUsesGetTasks(t *testing.T) {
	captor := &requestCaptor{}

	body, _ := json.Marshal(map[string]interface{}{"tasks": []interface{}{}})
	ts := newTestServer(t, captor, string(body), http.StatusOK)
	defer redirectToTestServer(t, ts.URL)()

	// all=true to skip filtering; jsonOutput=false to avoid output noise.
	_ = runList(false, false, false, true, false, false)

	got, err := captor.first()
	if err != nil {
		t.Fatalf("list: no HTTP request captured: %v", err)
	}

	if got.Method != http.MethodGet {
		t.Errorf("list: expected method GET, got %s", got.Method)
	}
	if got.Path != "/a2a/tasks" {
		t.Errorf("list: expected path /a2a/tasks, got %s", got.Path)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// status → GET /a2a/status/<taskID>
// ──────────────────────────────────────────────────────────────────────────────

func TestStatusUsesGetStatus(t *testing.T) {
	captor := &requestCaptor{}

	body, _ := json.Marshal(map[string]interface{}{
		"id":     "internal-task-id",
		"status": map[string]string{"state": "working"},
	})
	ts := newTestServer(t, captor, string(body), http.StatusOK)
	defer redirectToTestServer(t, ts.URL)()

	_, _ = runStatusByInternalID("internal-task-id")

	got, err := captor.first()
	if err != nil {
		t.Fatalf("status: no HTTP request captured: %v", err)
	}

	if got.Method != http.MethodGet {
		t.Errorf("status: expected method GET, got %s", got.Method)
	}
	expectedPath := "/a2a/status/internal-task-id"
	if got.Path != expectedPath {
		t.Errorf("status: expected path %s, got %s", expectedPath, got.Path)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// cancel → POST /a2a/cancel/<taskID>
// ──────────────────────────────────────────────────────────────────────────────

func TestCancelUsesPostCancel(t *testing.T) {
	captor := &requestCaptor{}

	body, _ := json.Marshal(map[string]interface{}{"status": "cancelled"})
	ts := newTestServer(t, captor, string(body), http.StatusOK)
	defer redirectToTestServer(t, ts.URL)()

	_ = runCancel("internal-task-id")

	got, err := captor.first()
	if err != nil {
		t.Fatalf("cancel: no HTTP request captured: %v", err)
	}

	if got.Method != http.MethodPost {
		t.Errorf("cancel: expected method POST, got %s", got.Method)
	}
	expectedPath := "/a2a/cancel/internal-task-id"
	if got.Path != expectedPath {
		t.Errorf("cancel: expected path %s, got %s", expectedPath, got.Path)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// retry → POST /a2a/retry/<taskID>
// ──────────────────────────────────────────────────────────────────────────────

func TestRetryUsesPostRetry(t *testing.T) {
	captor := &requestCaptor{}

	body, _ := json.Marshal(map[string]interface{}{"status": "submitted"})
	ts := newTestServer(t, captor, string(body), http.StatusOK)
	defer redirectToTestServer(t, ts.URL)()

	_ = runRetry("internal-task-id")

	got, err := captor.first()
	if err != nil {
		t.Fatalf("retry: no HTTP request captured: %v", err)
	}

	if got.Method != http.MethodPost {
		t.Errorf("retry: expected method POST, got %s", got.Method)
	}
	expectedPath := "/a2a/retry/internal-task-id"
	if got.Path != expectedPath {
		t.Errorf("retry: expected path %s, got %s", expectedPath, got.Path)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// resume → POST /a2a/resume/<taskID>
// ──────────────────────────────────────────────────────────────────────────────

func TestResumeUsesPostResume(t *testing.T) {
	captor := &requestCaptor{}

	body, _ := json.Marshal(map[string]interface{}{"status": "working"})
	ts := newTestServer(t, captor, string(body), http.StatusOK)
	defer redirectToTestServer(t, ts.URL)()

	_ = runResumeHTTP("internal-task-id", 0, "") // extendBudget=0, extendDuration=""

	got, err := captor.first()
	if err != nil {
		t.Fatalf("resume: no HTTP request captured: %v", err)
	}

	if got.Method != http.MethodPost {
		t.Errorf("resume: expected method POST, got %s", got.Method)
	}
	expectedPath := "/a2a/resume/internal-task-id"
	if got.Path != expectedPath {
		t.Errorf("resume: expected path %s, got %s", expectedPath, got.Path)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// wait (polling) → GET /a2a/tasks/<taskID>
// ──────────────────────────────────────────────────────────────────────────────

func TestWaitUsesGetTaskByID(t *testing.T) {
	captor := &requestCaptor{}

	// Return a completed state so the wait loop exits immediately.
	body, _ := json.Marshal(map[string]interface{}{
		"task_id": "internal-task-id",
		"status":  "completed",
	})
	ts := newTestServer(t, captor, string(body), http.StatusOK)
	defer redirectToTestServer(t, ts.URL)()

	_ = runWaitHTTP("internal-task-id", 5*time.Second, 60*time.Second)

	got, err := captor.first()
	if err != nil {
		t.Fatalf("wait: no HTTP request captured: %v", err)
	}

	if got.Method != http.MethodGet {
		t.Errorf("wait: expected method GET, got %s", got.Method)
	}
	expectedPath := "/a2a/tasks/internal-task-id"
	if got.Path != expectedPath {
		t.Errorf("wait: expected path %s, got %s", expectedPath, got.Path)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// logs (recent) → GET /logs/recent
// ──────────────────────────────────────────────────────────────────────────────

func TestLogsRecentUsesGetLogsRecent(t *testing.T) {
	captor := &requestCaptor{}

	body, _ := json.Marshal(map[string]interface{}{"logs": []interface{}{}, "count": 0, "limit": 50})
	ts := newTestServer(t, captor, string(body), http.StatusOK)
	defer redirectToTestServer(t, ts.URL)()

	_ = runLogsRecent(ts.URL, 50, false)

	got, err := captor.first()
	if err != nil {
		t.Fatalf("logs recent: no HTTP request captured: %v", err)
	}

	if got.Method != http.MethodGet {
		t.Errorf("logs recent: expected method GET, got %s", got.Method)
	}
	if got.Path != "/logs/recent" {
		t.Errorf("logs recent: expected path /logs/recent, got %s", got.Path)
	}
}
