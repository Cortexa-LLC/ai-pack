package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

// SearchRequest represents a codebase search request
type SearchRequest struct {
	Query       string `json:"query"`        // Search query/pattern
	ProjectRoot string `json:"project_root"` // Directory to search
	MaxResults  int    `json:"max_results"`  // Maximum number of results (default 50)
}

// SearchResult represents a single search result
type SearchResult struct {
	File    string `json:"file"`     // Relative file path
	Line    int    `json:"line"`     // Line number
	Content string `json:"content"`  // Line content
	Match   string `json:"match"`    // Matched text
}

// SearchResponse represents the search response
type SearchResponse struct {
	Success bool           `json:"success"`
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Count   int            `json:"count"`
	Error   string         `json:"error,omitempty"`
}

// HandleSearch handles codebase search requests
func (s *AgentServer) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	// Use provided project root or default to server root
	projectRoot := req.ProjectRoot
	if projectRoot == "" {
		projectRoot = s.rootDir
	}

	// Default max results
	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = 50
	}

	monitoring.Logger.Info("codebase_search", "query", req.Query, "project_root", projectRoot)

	// Use ripgrep (rg) for fast searching
	// Format: rg --json "query" path
	cmd := exec.Command("rg",
		"--json",
		"--max-count", fmt.Sprintf("%d", maxResults),
		"--max-columns", "500",
		"--no-heading",
		"--with-filename",
		"--line-number",
		"--case-sensitive",
		req.Query,
		projectRoot,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// rg exits with code 1 if no matches found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// No matches - return empty results
			response := SearchResponse{
				Success: true,
				Query:   req.Query,
				Results: []SearchResult{},
				Count:   0,
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Check if rg is installed
		if strings.Contains(err.Error(), "executable file not found") {
			monitoring.Logger.Error("search_rg_not_found", "error", err)
			http.Error(w, "ripgrep (rg) is not installed. Please install it to use search.", http.StatusInternalServerError)
			return
		}

		monitoring.Logger.Error("search_exec_error", "error", err, "output", string(output))
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse ripgrep JSON output
	results := []SearchResult{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		var rgResult map[string]interface{}
		if err := json.Unmarshal([]byte(line), &rgResult); err != nil {
			continue
		}

		// Only process "match" type results
		if rgResult["type"] != "match" {
			continue
		}

		data, ok := rgResult["data"].(map[string]interface{})
		if !ok {
			continue
		}

		// Extract file path
		pathData, ok := data["path"].(map[string]interface{})
		if !ok {
			continue
		}
		filePath, ok := pathData["text"].(string)
		if !ok {
			continue
		}

		// Make path relative to project root
		relPath, err := filepath.Rel(projectRoot, filePath)
		if err != nil {
			relPath = filePath
		}

		// Extract line number
		lineNum, ok := data["line_number"].(float64)
		if !ok {
			continue
		}

		// Extract line content
		lines, ok := data["lines"].(map[string]interface{})
		if !ok {
			continue
		}
		content, ok := lines["text"].(string)
		if !ok {
			continue
		}

		// Extract matched text from submatches
		match := req.Query // Default to query
		if submatches, ok := data["submatches"].([]interface{}); ok && len(submatches) > 0 {
			if submatch, ok := submatches[0].(map[string]interface{}); ok {
				if matchData, ok := submatch["match"].(map[string]interface{}); ok {
					if matchText, ok := matchData["text"].(string); ok {
						match = matchText
					}
				}
			}
		}

		results = append(results, SearchResult{
			File:    relPath,
			Line:    int(lineNum),
			Content: strings.TrimSpace(content),
			Match:   match,
		})

		if len(results) >= maxResults {
			break
		}
	}

	response := SearchResponse{
		Success: true,
		Query:   req.Query,
		Results: results,
		Count:   len(results),
	}

	monitoring.Logger.Info("search_completed", "query", req.Query, "results", len(results))

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
}

// HandleSearchOptions handles CORS preflight for search endpoint
func (s *AgentServer) HandleSearchOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)
}
