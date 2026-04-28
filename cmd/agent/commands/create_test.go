package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	agentclient "github.com/cortexa-llc/ai-pack/cmd/agent/client"
)

func TestCreateCmd(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		serverResponse map[string]interface{}
		serverStatus   int
		expectError    bool
	}{
		{
			name: "successful_creation",
			args: []string{"Test task description"},
			serverResponse: map[string]interface{}{
				"task_id":      "test-abc",
				"status":       "queued",
				"project_root": "/test/project",
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name: "creation_with_role",
			args: []string{"Test task", "--role", "engineer"},
			serverResponse: map[string]interface{}{
				"task_id": "test-def",
				"status":  "queued",
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name: "creation_with_priority",
			args: []string{"Test task", "--priority", "P0"},
			serverResponse: map[string]interface{}{
				"task_id": "test-ghi",
				"status":  "queued",
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "server_error",
			args:         []string{"Test task"},
			serverStatus: http.StatusInternalServerError,
			expectError:  true,
		},
		{
			name: "invalid_response",
			args: []string{"Test task"},
			serverResponse: map[string]interface{}{
				// Missing task_id
				"status": "queued",
			},
			serverStatus: http.StatusOK,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/a2a/tasks/create" {
					t.Errorf("Expected path /a2a/tasks/create, got %s", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST method, got %s", r.Method)
				}

				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			// Override server URL for test
			oldURL := agentclient.DefaultBaseURL
			agentclient.DefaultBaseURL = server.URL
			defer func() { agentclient.DefaultBaseURL = oldURL }()

			// Create command
			cmd := newCreateCmd()
			cmd.SetArgs(tt.args)

			// Execute
			err := cmd.Execute()

			// Verify
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCreateCmd_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"task_id":      "test-json",
			"status":       "queued",
			"project_root": "/test",
		})
	}))
	defer server.Close()

	oldURL := agentclient.DefaultBaseURL
	agentclient.DefaultBaseURL = server.URL
	defer func() { agentclient.DefaultBaseURL = oldURL }()

	cmd := newCreateCmd()
	cmd.SetArgs([]string{"Test task", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
