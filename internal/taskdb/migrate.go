package taskdb

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/constants"
)

// beadsTask is the shape of a record in a Beads issues.jsonl file.
type beadsTask struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Priority    int     `json:"priority"`
	IssueType   string  `json:"issue_type"`
	Assignee    string  `json:"assignee"`
	Owner       string  `json:"owner"`
	CreatedBy   string  `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	ClosedAt    string  `json:"closed_at"`
	CloseReason string  `json:"close_reason"`
}

// MigrationBeadsJSONL is the key used to track the one-time Beads JSONL migration.
const MigrationBeadsJSONL = "beads_jsonl_v1"

// keep the unexported alias for internal use
const migrationBeadsJSONL = MigrationBeadsJSONL

// MigrateFromBeadsJSONL scans common project directories for .beads/issues.jsonl
// files and imports all Beads task records into the SQLite task database.
// It does not require the Dolt server to be running.
// This is a one-time migration: once recorded in schema_migrations it will not
// run again on subsequent server starts.
// Returns the total number of tasks imported across all projects.
func (db *DB) MigrateFromBeadsJSONL() (int, error) {
	// Ensure the migrations table exists (safe on older DBs that pre-date this table)
	if _, err := db.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name TEXT PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return 0, fmt.Errorf("ensure migrations table: %w", err)
	}

	// Check if already done
	var count int
	if err := db.db.QueryRow(
		`SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, migrationBeadsJSONL,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf("check migration status: %w", err)
	}
	if count > 0 {
		return 0, nil // already ran
	}

	homeDir := os.Getenv("HOME")
	searchRoots := []string{
		filepath.Join(homeDir, "Projects"),
		filepath.Join(homeDir, "Code"),
		filepath.Join(homeDir, "workspace"),
		filepath.Join(homeDir, "src"),
	}

	total := 0
	seen := map[string]bool{} // deduplicate JSONL paths

	for _, root := range searchRoots {
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // skip inaccessible paths
			}
			if !info.IsDir() || info.Name() != ".beads" {
				return nil // only act when we hit a .beads directory
			}

			// Found a .beads dir — look for issues.jsonl directly inside it.
			// Only take the top-level one (ignore .beads/backup/issues.jsonl etc.)
			jsonlPath := filepath.Join(path, "issues.jsonl")
			if !seen[jsonlPath] {
				if fi, err := os.Stat(jsonlPath); err == nil && !fi.IsDir() {
					seen[jsonlPath] = true
					projectRoot := filepath.Dir(path) // parent of .beads
					n, err := db.migrateBeadsJSONLFile(jsonlPath, projectRoot)
					if err != nil {
						fmt.Printf("  ⚠️  %s: %v\n", jsonlPath, err)
					} else if n > 0 {
						fmt.Printf("  ✅ %s: migrated %d tasks\n", projectRoot, n)
						total += n
					}
				}
			}

			return filepath.SkipDir // don't recurse into .beads subdirs
		})
		if err != nil {
			fmt.Printf("Warning: error scanning %s: %v\n", root, err)
		}
	}

	// Record migration as complete so it won't run again on future starts
	if _, err := db.db.Exec(
		`INSERT OR IGNORE INTO schema_migrations (name) VALUES (?)`, migrationBeadsJSONL,
	); err != nil {
		return total, fmt.Errorf("record migration: %w", err)
	}

	return total, nil
}

// migrateBeadsJSONLFile reads a single issues.jsonl file and imports its tasks.
func (db *DB) migrateBeadsJSONLFile(jsonlPath, defaultProjectRoot string) (int, error) {
	f, err := os.Open(jsonlPath)
	if err != nil {
		return 0, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	migrated := 0
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1 MB per line

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var bt beadsTask
		if err := json.Unmarshal([]byte(line), &bt); err != nil {
			continue // skip malformed lines
		}
		if bt.ID == "" {
			continue
		}

		// Skip if already in the database
		if existing, _ := db.GetTask(bt.ID); existing != nil {
			continue
		}

		// Determine project root: prefer "Working directory:" from description
		projectRoot := extractWorkingDirectory(bt.Description)
		if projectRoot == "" {
			projectRoot = defaultProjectRoot
		}

		// Use description if available, otherwise fall back to title
		desc := bt.Description
		if desc == "" {
			desc = bt.Title
		}

		task := &Task{
			ID:              bt.ID,
			ProjectRoot:     projectRoot,
			Role:            "engineer", // Beads doesn't store role; default to engineer
			TaskDescription: desc,
			Status:          mapBeadsStatus(bt.Status),
		}

		// Parse timestamps (Beads uses RFC3339 with timezone offsets)
		if t, err := time.Parse(time.RFC3339Nano, bt.CreatedAt); err == nil {
			task.CreatedAt = t.UTC()
		} else if t, err := time.Parse(time.RFC3339, bt.CreatedAt); err == nil {
			task.CreatedAt = t.UTC()
		} else {
			task.CreatedAt = time.Now().UTC()
		}

		if t, err := time.Parse(time.RFC3339Nano, bt.UpdatedAt); err == nil {
			task.UpdatedAt = t.UTC()
		} else if t, err := time.Parse(time.RFC3339, bt.UpdatedAt); err == nil {
			task.UpdatedAt = t.UTC()
		} else {
			task.UpdatedAt = task.CreatedAt
		}

		if bt.ClosedAt != "" {
			if t, err := time.Parse(time.RFC3339Nano, bt.ClosedAt); err == nil {
				ct := t.UTC()
				task.CompletedAt = &ct
			} else if t, err := time.Parse(time.RFC3339, bt.ClosedAt); err == nil {
				ct := t.UTC()
				task.CompletedAt = &ct
			}
		}

		// Store Beads-specific fields in metadata JSON
		meta := map[string]interface{}{
			"source":       "beads",
			"title":        bt.Title,
			"priority":     bt.Priority,
			"issue_type":   bt.IssueType,
			"assignee":     bt.Assignee,
			"owner":        bt.Owner,
			"close_reason": bt.CloseReason,
		}
		metaJSON, _ := json.Marshal(meta)
		task.Metadata = string(metaJSON)

		query := `
			INSERT INTO tasks (
				id, project_root, role, task_description, status,
				created_at, updated_at, completed_at, metadata
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		if _, err := db.db.Exec(query,
			task.ID, task.ProjectRoot, task.Role,
			task.TaskDescription, task.Status,
			task.CreatedAt, task.UpdatedAt, task.CompletedAt,
			task.Metadata,
		); err != nil {
			// Log and continue — don't abort the whole migration on one bad row
			fmt.Printf("    skip %s: %v\n", task.ID, err)
			continue
		}
		migrated++
	}

	return migrated, scanner.Err()
}

// extractWorkingDirectory parses "Working directory: /path/to/project" from a
// Beads task description.
func extractWorkingDirectory(description string) string {
	for _, line := range strings.Split(description, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Working directory:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Working directory:"))
		}
	}
	return ""
}

// mapBeadsStatus maps Beads task status strings to taskdb status constants.
func mapBeadsStatus(s string) string {
	switch s {
	case "closed", "done":
		return StatusCompleted
	case "in_progress", "working", "running":
		return StatusInProgress
	case "failed", "error":
		return StatusFailed
	case "cancelled", "canceled":
		return StatusCancelled
	default: // "open", ""
		return StatusQueued
	}
}

// MigrationRunsV1 is the key used to track the one-time task-runs migration.
const MigrationRunsV1 = "beads_runs_v1"

// MigrateRunsFromBeads scans all .beads/tasks/ directories across common project
// locations and imports each execution attempt into the task_runs table.
// It is a one-time migration guarded by schema_migrations.
// Returns the total number of runs imported.
func (db *DB) MigrateRunsFromBeads() (int, error) {
	// Ensure migrations table exists
	if _, err := db.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name TEXT PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return 0, fmt.Errorf("ensure migrations table: %w", err)
	}

	// One-time guard
	var count int
	if err := db.db.QueryRow(
		`SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, MigrationRunsV1,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf("check migration status: %w", err)
	}
	if count > 0 {
		return 0, nil
	}

	homeDir := os.Getenv("HOME")
	searchRoots := []string{
		filepath.Join(homeDir, "Projects"),
		filepath.Join(homeDir, "Code"),
		filepath.Join(homeDir, "workspace"),
		filepath.Join(homeDir, "src"),
	}

	total := 0
	seen := map[string]bool{}

	for _, root := range searchRoots {
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() || info.Name() != ".beads" {
				return nil
			}

			// Found a .beads dir — scan its tasks/ subdirectory
			tasksDir := filepath.Join(path, "tasks")
			projectRoot := filepath.Dir(path)

			entries, err := os.ReadDir(tasksDir)
			if err != nil {
				return filepath.SkipDir
			}

			for _, entry := range entries {
				if !entry.IsDir() || entry.Name() == ".archive" {
					continue
				}
				runDir := filepath.Join(tasksDir, entry.Name())
				if seen[runDir] {
					continue
				}
				seen[runDir] = true

				n, err := db.migrateRunDir(runDir, entry.Name(), projectRoot)
				if err != nil {
					fmt.Printf("    skip run %s: %v\n", entry.Name(), err)
				} else {
					total += n
				}
			}

			return filepath.SkipDir
		})
		if err != nil {
			fmt.Printf("Warning: error scanning %s: %v\n", root, err)
		}
	}

	// Record completion
	if _, err := db.db.Exec(
		`INSERT OR IGNORE INTO schema_migrations (name) VALUES (?)`, MigrationRunsV1,
	); err != nil {
		return total, fmt.Errorf("record migration: %w", err)
	}

	return total, nil
}

// migrateRunDir imports a single .beads/tasks/<run-id>/ directory into task_runs.
func (db *DB) migrateRunDir(runDir, runDirName, defaultProjectRoot string) (int, error) {
	metaPath := filepath.Join(runDir, "00-metadata.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return 0, nil // no metadata — skip silently
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return 0, nil
	}

	// run_id: use the task_id field if present, otherwise the directory name
	runID := getString(raw, "task_id")
	if runID == "" {
		runID = runDirName
	}

	// Skip if already imported
	var exists int
	if err := db.db.QueryRow(`SELECT COUNT(*) FROM task_runs WHERE run_id = ?`, runID).Scan(&exists); err == nil && exists > 0 {
		return 0, nil
	}

	// task_id: the short Beads ID from the nested metadata map
	shortTaskID := ""
	if meta, ok := raw["metadata"].(map[string]interface{}); ok {
		shortTaskID = getString(meta, "beads_task_id")
	}
	if shortTaskID == "" {
		// Fall back: strip timestamp suffix from run ID (e.g. "ai-pack-0a6-20260225-211529" → "ai-pack-0a6")
		shortTaskID = extractShortTaskID(runID)
	}

	// project_root: prefer metadata.project_root
	projectRoot := defaultProjectRoot
	if meta, ok := raw["metadata"].(map[string]interface{}); ok {
		if pr := getString(meta, "project_root"); pr != "" {
			projectRoot = pr
		}
	}

	role := getString(raw, "role")
	if role == "" {
		role = "engineer"
	}

	status := mapStatus(getString(raw, "status"))

	var createdAt, updatedAt time.Time
	if t, err := time.Parse(time.RFC3339Nano, getString(raw, "spawned_at")); err == nil {
		createdAt = t.UTC()
	} else if t, err := time.Parse(time.RFC3339, getString(raw, "spawned_at")); err == nil {
		createdAt = t.UTC()
	} else {
		createdAt = time.Now().UTC()
	}

	if t, err := time.Parse(time.RFC3339Nano, getString(raw, "updated_at")); err == nil {
		updatedAt = t.UTC()
	} else if t, err := time.Parse(time.RFC3339, getString(raw, "updated_at")); err == nil {
		updatedAt = t.UTC()
	} else {
		updatedAt = createdAt
	}

	var completedAt *time.Time
	if status == StatusCompleted || status == StatusFailed || status == StatusCancelled {
		t := updatedAt
		completedAt = &t
	}

	// Read result from 30-results.md if present
	result := ""
	if rd, err := os.ReadFile(filepath.Join(runDir, "30-results.md")); err == nil {
		result = string(rd)
	}

	// Store model/provider/tier in metadata
	runMeta := map[string]interface{}{
		"model":      getString(raw, "model"),
		"provider":   getString(raw, "provider"),
		"tier":       getString(raw, "tier"),
		"spawned_by": getString(raw, "spawned_by"),
	}
	runMetaJSON, _ := json.Marshal(runMeta)

	// Ensure the parent task row exists (create a stub if the JSONL migration
	// didn't produce one — e.g. tasks created before Beads was in use)
	if err := db.ensureTaskStub(shortTaskID, projectRoot, role, getString(raw, "description"), createdAt, updatedAt); err != nil {
		return 0, fmt.Errorf("ensure task stub: %w", err)
	}

	_, err = db.db.Exec(`
		INSERT OR IGNORE INTO task_runs (
			run_id, task_id, project_root, role, status,
			created_at, updated_at, completed_at,
			result, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		runID, shortTaskID, projectRoot, role, status,
		createdAt, updatedAt, completedAt,
		result, string(runMetaJSON),
	)
	if err != nil {
		return 0, err
	}

	// Update parent task's latest_run_id if this is the most recent run
	db.db.Exec(`
		UPDATE tasks SET latest_run_id = ?, status = ?, updated_at = ?
		WHERE id = ? AND (latest_run_id IS NULL OR updated_at <= ?)`,
		runID, status, updatedAt, shortTaskID, updatedAt,
	) //nolint:errcheck — best-effort update

	return 1, nil
}

// ensureTaskStub creates a minimal tasks row if none exists for shortTaskID.
// This handles runs whose Beads task record was never in issues.jsonl.
func (db *DB) ensureTaskStub(taskID, projectRoot, role, description string, createdAt, updatedAt time.Time) error {
	var exists int
	if err := db.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE id = ?`, taskID).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	if description == "" {
		description = taskID
	}
	_, err := db.db.Exec(`
		INSERT OR IGNORE INTO tasks (id, project_root, role, task_description, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		taskID, projectRoot, role, description, StatusQueued, createdAt, updatedAt,
	)
	return err
}

// MigrateFromLegacy imports tasks from .beads/tasks directories into the SQLite database.
// It scans all task directories and creates corresponding database entries.
func (db *DB) MigrateFromLegacy(projectRoot string) (int, error) {
	tasksDir := filepath.Join(projectRoot, constants.TaskRootDir, "tasks")

	// Check if tasks directory exists
	if _, err := os.Stat(tasksDir); os.IsNotExist(err) {
		return 0, nil // No tasks to migrate
	}

	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read tasks directory: %w", err)
	}

	migrated := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		taskID := entry.Name()
		metadataPath := filepath.Join(tasksDir, taskID, "00-metadata.json")

		// Read metadata file
		data, err := os.ReadFile(metadataPath)
		if err != nil {
			// Skip tasks without metadata
			continue
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(data, &metadata); err != nil {
			continue
		}

		// Check if task already exists in database
		existing, _ := db.GetTask(taskID)
		if existing != nil {
			continue // Skip already migrated tasks
		}

		// Map metadata to Task struct
		task := &Task{
			ID:              taskID,
			ProjectRoot:     projectRoot,
			Role:            getString(metadata, "role"),
			TaskDescription: getString(metadata, "task"),
			Status:          mapStatus(getString(metadata, "status")),
		}

		// Parse timestamps
		if createdStr := getString(metadata, "spawned_at"); createdStr != "" {
			if t, err := time.Parse(time.RFC3339, createdStr); err == nil {
				task.CreatedAt = t
			}
		}
		if task.CreatedAt.IsZero() {
			task.CreatedAt = time.Now()
		}

		if updatedStr := getString(metadata, "updated_at"); updatedStr != "" {
			if t, err := time.Parse(time.RFC3339, updatedStr); err == nil {
				task.UpdatedAt = t
			}
		}
		if task.UpdatedAt.IsZero() {
			task.UpdatedAt = task.CreatedAt
		}

		// Extract error message if present
		if errorMsg := getString(metadata, "error"); errorMsg != "" {
			task.Error = errorMsg
		}

		// Read result if task is completed
		if task.Status == StatusCompleted || task.Status == StatusFailed {
			resultPath := filepath.Join(tasksDir, taskID, "result.txt")
			if resultData, err := os.ReadFile(resultPath); err == nil {
				task.Result = string(resultData)
			}

			completedTime := task.UpdatedAt
			task.CompletedAt = &completedTime
		}

		// Insert task (bypass CreateTask to preserve original timestamps)
		query := `
			INSERT INTO tasks (
				id, project_root, role, task_description, status,
				created_at, updated_at, completed_at, result, error
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		_, err = db.db.Exec(query,
			task.ID, task.ProjectRoot, task.Role,
			task.TaskDescription, task.Status,
			task.CreatedAt, task.UpdatedAt, task.CompletedAt,
			task.Result, task.Error,
		)

		if err != nil {
			return migrated, fmt.Errorf("failed to insert task %s: %w", taskID, err)
		}

		migrated++
	}

	return migrated, nil
}

// extractShortTaskID extracts the short task ID from a full task ID.
// Example: "listingsgql-5b6-20260424-162611" -> "listingsgql-5b6"
func extractShortTaskID(taskID string) string {
	// Split by hyphen and take everything before the timestamp
	parts := strings.Split(taskID, "-")
	if len(parts) < 3 {
		return taskID
	}

	// Find the first part that looks like a timestamp (YYYYMMDD)
	for i := 2; i < len(parts); i++ {
		if len(parts[i]) == 8 && isNumeric(parts[i]) {
			// Everything before this is the task ID
			return strings.Join(parts[:i], "-")
		}
	}

	return taskID
}

// mapStatus maps metadata status to database status constants.
func mapStatus(metaStatus string) string {
	switch metaStatus {
	case "in_progress", "working", "running":
		return StatusInProgress
	case "completed", "done", "success":
		return StatusCompleted
	case "failed", "error":
		return StatusFailed
	case "cancelled", "canceled":
		return StatusCancelled
	default:
		return StatusQueued
	}
}

// getString safely extracts a string from a map.
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// isNumeric checks if a string contains only digits.
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
