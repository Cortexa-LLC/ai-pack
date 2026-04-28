package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	agentclient "github.com/cortexa-llc/ai-pack/cmd/agent/client"
)

func TestDeleteCmd(t *testing.T) {
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
			name:   "successful_delete",
			taskID: "test-abc",
			args:   []string{"test-abc", "--force"},
			serverResponse: map[string]interface{}{
				"task_id": "test-abc",
				"status":  "deleted",
			},
			serverStatus: http.StatusOK,
			expectError:  false,
			checkRequest: func(t *testing.T, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("Expected DELETE method, got %s", r.Method)
				}
			},
		},
		{
			name:         "task_not_found",
			taskID:       "nonexistent",
			args:         []string{"nonexistent", "--force"},
			serverStatus: http.StatusNotFound,
			expectError:  true,
		},
		{
			name:         "server_error",
			taskID:       "test-error",
			args:         []string{"test-error", "--force"},
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

			cmd := newDeleteCmd()
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

func TestDeleteCmd_NoArgs(t *testing.T) {
	cmd := newDeleteCmd()
	cmd.SetArgs([]string{}) // No task ID

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no task ID provided")
	}
}

func TestDeleteCmd_JsonOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"task_id": "test-json",
			"status":  "deleted",
		})
	}))
	defer server.Close()

	oldURL := agentclient.DefaultBaseURL
	agentclient.DefaultBaseURL = server.URL
	defer func() { agentclient.DefaultBaseURL = oldURL }()

	cmd := newDeleteCmd()
	cmd.SetArgs([]string{"test-json", "--force", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
