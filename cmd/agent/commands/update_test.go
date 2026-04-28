package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	agentclient "github.com/cortexa-llc/ai-pack/cmd/agent/client"
)

func TestUpdateCmd(t *testing.T) {
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
			name:   "update_status",
			taskID: "test-abc",
			args:   []string{"test-abc", "--status", "in_progress"},
			serverResponse: map[string]interface{}{
				"task_id": "test-abc",
				"status":  "updated",
			},
			serverStatus: http.StatusOK,
			expectError:  false,
			checkRequest: func(t *testing.T, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				var req map[string]interface{}
				json.Unmarshal(body, &req)
				if req["status"] != "in_progress" {
					t.Errorf("Expected status in request, got: %v", req)
				}
			},
		},
		{
			name:   "update_result",
			taskID: "test-def",
			args:   []string{"test-def", "--result", "Partial completion"},
			serverResponse: map[string]interface{}{
				"task_id": "test-def",
				"status":  "updated",
			},
			serverStatus: http.StatusOK,
			expectError:  false,
			checkRequest: func(t *testing.T, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				var req map[string]interface{}
				json.Unmarshal(body, &req)
				if req["result"] != "Partial completion" {
					t.Errorf("Expected result in request, got: %v", req)
				}
			},
		},
		{
			name:   "update_metadata",
			taskID: "test-ghi",
			args:   []string{"test-ghi", "--metadata", `{"priority":"P0"}`},
			serverResponse: map[string]interface{}{
				"task_id": "test-ghi",
				"status":  "updated",
			},
			serverStatus: http.StatusOK,
			expectError:  false,
			checkRequest: func(t *testing.T, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				var req map[string]interface{}
				json.Unmarshal(body, &req)
				if req["metadata"] != `{"priority":"P0"}` {
					t.Errorf("Expected metadata in request, got: %v", req)
				}
			},
		},
		{
			name:        "invalid_status",
			taskID:      "test-invalid",
			args:        []string{"test-invalid", "--status", "invalid_status"},
			expectError: true,
		},
		{
			name:        "invalid_json_metadata",
			taskID:      "test-json-err",
			args:        []string{"test-json-err", "--metadata", "not-json"},
			expectError: true,
		},
		{
			name:        "no_fields_specified",
			taskID:      "test-empty",
			args:        []string{"test-empty"},
			expectError: true,
		},
		{
			name:         "task_not_found",
			taskID:       "nonexistent",
			args:         []string{"nonexistent", "--status", "completed"},
			serverStatus: http.StatusNotFound,
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

			cmd := newUpdateCmd()
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

func TestUpdateCmd_MultipleFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		json.Unmarshal(body, &req)

		// Verify multiple fields were sent
		if req["status"] != "completed" {
			t.Errorf("Expected status field, got: %v", req)
		}
		if req["result"] != "All tests pass" {
			t.Errorf("Expected result field, got: %v", req)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"task_id": "test-multi",
			"status":  "updated",
		})
	}))
	defer server.Close()

	oldURL := agentclient.DefaultBaseURL
	agentclient.DefaultBaseURL = server.URL
	defer func() { agentclient.DefaultBaseURL = oldURL }()

	cmd := newUpdateCmd()
	cmd.SetArgs([]string{"test-multi", "--status", "completed", "--result", "All tests pass"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
