package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	agentclient "github.com/cortexa-llc/ai-pack/cmd/agent/client"
)

func TestShowCmd(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		serverResponse map[string]interface{}
		serverStatus   int
		expectError    bool
	}{
		{
			name:   "successful_show",
			taskID: "test-abc",
			serverResponse: map[string]interface{}{
				"id":          "test-abc",
				"status":      "queued",
				"role":        "engineer",
				"description": "Test task",
				"created_at":  time.Now().Format(time.RFC3339),
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:   "show_completed_task",
			taskID: "test-def",
			serverResponse: map[string]interface{}{
				"id":           "test-def",
				"status":       "completed",
				"role":         "tester",
				"description":  "Completed task",
				"created_at":   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				"completed_at": time.Now().Format(time.RFC3339),
				"result":       "Task completed successfully",
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:   "show_failed_task",
			taskID: "test-ghi",
			serverResponse: map[string]interface{}{
				"id":          "test-ghi",
				"status":      "failed",
				"description": "Failed task",
				"error":       "Task failed: timeout",
				"created_at":  time.Now().Format(time.RFC3339),
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "task_not_found",
			taskID:       "nonexistent",
			serverStatus: http.StatusNotFound,
			expectError:  true,
		},
		{
			name:         "server_error",
			taskID:       "test-error",
			serverStatus: http.StatusInternalServerError,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/a2a/tasks/" + tt.taskID
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET method, got %s", r.Method)
				}

				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			oldURL := agentclient.DefaultBaseURL
			agentclient.DefaultBaseURL = server.URL
			defer func() { agentclient.DefaultBaseURL = oldURL }()

			cmd := newShowCmd()
			cmd.SetArgs([]string{tt.taskID})

			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestShowCmd_JSONOutput(t *testing.T) {
	taskID := "test-json"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          taskID,
			"status":      "in_progress",
			"description": "JSON output test",
			"created_at":  time.Now().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	oldURL := agentclient.DefaultBaseURL
	agentclient.DefaultBaseURL = server.URL
	defer func() { agentclient.DefaultBaseURL = oldURL }()

	cmd := newShowCmd()
	cmd.SetArgs([]string{taskID, "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestShowCmd_NoArgs(t *testing.T) {
	cmd := newShowCmd()
	cmd.SetArgs([]string{}) // No task ID provided

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no task ID provided")
	}
}
