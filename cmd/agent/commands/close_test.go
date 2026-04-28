package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	agentclient "github.com/cortexa-llc/ai-pack/cmd/agent/client"
)

func TestCloseCmd(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		args           []string
		serverResponse map[string]interface{}
		serverStatus   int
		expectError    bool
		checkRequest   func(*testing.T, *http.Request)
	}{
		{
			name:   "successful_close",
			taskID: "test-abc",
			args:   []string{"test-abc"},
			serverResponse: map[string]interface{}{
				"task_id": "test-abc",
				"status":  "completed",
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:   "close_with_result",
			taskID: "test-def",
			args:   []string{"test-def", "--result", "Task completed successfully"},
			serverResponse: map[string]interface{}{
				"task_id": "test-def",
				"status":  "completed",
			},
			serverStatus: http.StatusOK,
			expectError:  false,
			checkRequest: func(t *testing.T, r *http.Request) {
				// Verify result was sent in request body
				body, _ := io.ReadAll(r.Body)
				var req map[string]interface{}
				json.Unmarshal(body, &req)
				if req["result"] != "Task completed successfully" {
					t.Errorf("Expected result in request body, got: %v", req)
				}
			},
		},
		{
			name:         "task_not_found",
			taskID:       "nonexistent",
			args:         []string{"nonexistent"},
			serverStatus: http.StatusNotFound,
			expectError:  true,
		},
		{
			name:         "server_error",
			taskID:       "test-error",
			args:         []string{"test-error"},
			serverStatus: http.StatusInternalServerError,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/a2a/tasks/" + tt.taskID + "/close"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodPut {
					t.Errorf("Expected PUT method, got %s", r.Method)
				}

				if tt.checkRequest != nil {
					tt.checkRequest(t, r)
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

			cmd := newCloseCmd()
			cmd.SetArgs(tt.args)

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

func TestCloseCmd_NoArgs(t *testing.T) {
	cmd := newCloseCmd()
	cmd.SetArgs([]string{}) // No task ID

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no task ID provided")
	}
}
