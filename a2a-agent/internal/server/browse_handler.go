package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// isPathAllowed checks whether path is allowed given the set of allowed roots.
// A path is allowed if it IS one of the roots, is under one of the roots, or is
// a prefix of one of the roots (so the user can navigate up to reach a root).
func isPathAllowed(path string, allowedRoots []string) bool {
	cleanPath := filepath.Clean(path)
	for _, root := range allowedRoots {
		cleanRoot := filepath.Clean(root)
		// Allow path if it IS an allowed root, is under one, or is a prefix of one
		if strings.HasPrefix(cleanPath, cleanRoot) || strings.HasPrefix(cleanRoot, cleanPath) {
			return true
		}
	}
	return false
}

// HandleBrowseDirectories returns directory suggestions for autocomplete
func (s *AgentServer) HandleBrowseDirectories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// Get partial path from query
	partialPath := r.URL.Query().Get("path")
	if partialPath == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"directories": []string{},
		})
		return
	}

	// Expand home directory
	if strings.HasPrefix(partialPath, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			partialPath = filepath.Join(home, partialPath[2:])
		}
	}

	// Restrict browsing to allowed project roots (and their prefixes/ancestors)
	allowedRoots := s.GetProjectRoots()
	if len(allowedRoots) > 0 && !isPathAllowed(partialPath, allowedRoots) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "path not within allowed project roots",
		})
		return
	}

	// Get directory to search and prefix
	var searchDir string
	var prefix string

	// If path ends with separator, search that directory
	if strings.HasSuffix(partialPath, string(os.PathSeparator)) {
		searchDir = partialPath
		prefix = ""
	} else {
		// Otherwise, search parent directory with prefix filter
		searchDir = filepath.Dir(partialPath)
		prefix = filepath.Base(partialPath)
	}

	// List directories
	var directories []string
	entries, err := os.ReadDir(searchDir)
	if err != nil {
		// Directory doesn't exist or can't be read - return empty
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"directories": []string{},
		})
		return
	}

	// Filter for directories matching prefix
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip hidden directories unless user is typing them
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".") {
			continue
		}

		// Filter by prefix
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			continue
		}

		// Build full path
		fullPath := filepath.Join(searchDir, name)

		// Convert back to ~/ notation if under home
		home, _ := os.UserHomeDir()
		if home != "" && strings.HasPrefix(fullPath, home) {
			fullPath = "~" + strings.TrimPrefix(fullPath, home)
		}

		directories = append(directories, fullPath)

		// Limit results
		if len(directories) >= 20 {
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"directories": directories,
	})
}
