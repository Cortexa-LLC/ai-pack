package server

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

// isShortTaskID checks if a string matches the task ID format
// Valid formats: project-id, project-id.subtask, project++-id, project++-id.subtask
func isShortTaskID(id string) bool {
	// task ID pattern: starts with letter, contains letters/numbers/+/_, has a dash, then alphanumeric, optional .subtask
	pattern := `^[a-zA-Z][a-zA-Z0-9+_-]*-[a-zA-Z0-9]+(\.[a-zA-Z0-9]+)?$`
	matched, _ := regexp.MatchString(pattern, id)
	return matched
}

// extractTimestampFromFolderName extracts YYYYMMDD-HHMMSS from task-{role}-YYYYMMDD-HHMMSS-000000
func extractTimestampFromFolderName(folderName string) (string, error) {
	pattern := `task-[^-]+-([0-9]{8})-([0-9]{6})-[0-9]{6}`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(folderName)
	if len(matches) < 3 {
		return "", fmt.Errorf("cannot parse timestamp from folder name: %s", folderName)
	}
	return matches[1] + "-" + matches[2], nil
}

// extractShortTaskIDFromPrompt reads agent-prompt.txt and extracts the task ID
func extractShortTaskIDFromPrompt(taskDir string) (string, error) {
	promptPath := filepath.Join(taskDir, "agent-prompt.txt")
	data, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	foundTaskLine := false
	for _, line := range lines {
		if strings.Contains(line, "**Your Task:**") {
			foundTaskLine = true
			continue
		}
		if foundTaskLine && strings.TrimSpace(line) != "" {
			// Extract first word as potential task ID
			words := strings.Fields(line)
			if len(words) > 0 {
				return words[0], nil
			}
			break
		}
	}
	return "", fmt.Errorf("no task found in prompt")
}

// DetectLegacyTaskFoldersInProject checks if legacy task-* folders exist in a specific project
func DetectLegacyTaskFoldersInProject(projectRoot string) (bool, []string) {
	var legacyFolders []string
	tasksDir := filepath.Join(projectRoot, ".beads", "tasks")

	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return false, nil // No tasks directory means no legacy folders
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "task-") {
			legacyFolders = append(legacyFolders, filepath.Join(tasksDir, entry.Name()))
		}
	}

	return len(legacyFolders) > 0, legacyFolders
}

// DetectLegacyTaskFolders checks if any legacy task-* folders exist in registered projects
func (s *AgentServer) DetectLegacyTaskFolders() (bool, []string) {
	var legacyFolders []string
	projectRoots := s.GetProjectRoots()

	for _, projectRoot := range projectRoots {
		hasLegacy, folders := DetectLegacyTaskFoldersInProject(projectRoot)
		if hasLegacy {
			legacyFolders = append(legacyFolders, folders...)
		}
	}

	return len(legacyFolders) > 0, legacyFolders
}

// MigrateTaskFolders migrates legacy task-* folders to beads-id-timestamp format
// Returns counts of renamed, archived, and skipped folders
func (s *AgentServer) MigrateTaskFolders() (renamed, archived, skipped int, err error) {
	projectRoots := s.GetProjectRoots()

	for _, projectRoot := range projectRoots {
		tasksDir := filepath.Join(projectRoot, ".beads", "tasks")
		archiveDir := filepath.Join(tasksDir, ".archive", "legacy")

		// Create archive directory
		if err := os.MkdirAll(archiveDir, 0755); err != nil {
			return renamed, archived, skipped, fmt.Errorf("failed to create archive directory: %w", err)
		}

		// Create README if it doesn't exist
		readmePath := filepath.Join(archiveDir, "README.md")
		if _, err := os.Stat(readmePath); os.IsNotExist(err) {
			readme := `# Legacy Task Archive

## Overview
This directory contains agent tasks that were created before the system standardized on task IDs.

## What's Archived
Tasks with the naming format: ` + "`task-{role}-{timestamp}`" + `

These tasks were created using either:
- Free-form descriptions (not tasks)
- tasks but with old folder naming (duplicates after migration)

## Why Archived
As of February 2026, the agent-server system was refactored to:
- Require all tasks to be created in Beads first (` + "`bd create '<description>'`" + `)
- Use task IDs as the single source of truth
- Use consistent timestamped folder names: ` + "`{short-task-id}-{YYYYMMDD}-{HHMMSS}`" + `

Legacy tasks have been archived to:
1. Prevent confusion between old and new task formats
2. Maintain system consistency going forward
3. Preserve historical records for reference

## Task Contents
Each archived task folder contains:
- ` + "`00-metadata.json`" + ` - Agent configuration and metadata
- ` + "`agent-prompt.txt`" + ` - The original task prompt
- ` + "`execution.log`" + ` - Execution history
- ` + "`10-plan.md`" + ` - Planning phase output (if applicable)
- ` + "`30-results.md`" + ` - Task results (if completed)

## Accessing Archived Tasks
These tasks are read-only archives. If you need to perform similar work:
1. Create a new task: ` + "`bd create '<description>'`" + `
2. Spawn an agent with the task ID: ` + "`agent {role} {beads-task-id}`" + `

## Migration Date
Migrated on: ` + time.Now().Format("2006-01-02") + `
`
			if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
				monitoring.Logger.Warn("failed_to_create_readme", "error", err.Error())
			}
		}

		entries, err := os.ReadDir(tasksDir)
		if err != nil {
			continue // Skip if tasks directory doesn't exist
		}

		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "task-") {
				continue
			}

			folderName := entry.Name()
			folderPath := filepath.Join(tasksDir, folderName)

			// Extract timestamp
			timestamp, err := extractTimestampFromFolderName(folderName)
			if err != nil {
				monitoring.Logger.Warn("migration_skip_bad_timestamp",
					"folder", folderName,
					"error", err.Error())
				skipped++
				continue
			}

			// Try to extract task ID from prompt
			shortTaskID, err := extractShortTaskIDFromPrompt(folderPath)
			if err != nil || shortTaskID == "" {
				// No prompt or no task - archive
				monitoring.Logger.Info("migration_archive_no_prompt",
					"folder", folderName,
					"reason", "no_beads_id")
				if err := os.Rename(folderPath, filepath.Join(archiveDir, folderName)); err != nil {
					monitoring.Logger.Error("migration_archive_failed", "folder", folderName, "error", err.Error())
					skipped++
				} else {
					archived++
				}
				continue
			}

			// Validate it's actually a task ID
			if !isShortTaskID(shortTaskID) {
				// Free-form description - archive
				monitoring.Logger.Info("migration_archive_freeform",
					"folder", folderName,
					"task", shortTaskID,
					"reason", "not_beads_id_format")
				if err := os.Rename(folderPath, filepath.Join(archiveDir, folderName)); err != nil {
					monitoring.Logger.Error("migration_archive_failed", "folder", folderName, "error", err.Error())
					skipped++
				} else {
					archived++
				}
				continue
			}

			// Rename to beads-id-timestamp format
			newName := fmt.Sprintf("%s-%s", shortTaskID, timestamp)
			newPath := filepath.Join(tasksDir, newName)

			// Check if destination already exists
			if _, err := os.Stat(newPath); err == nil {
				// Destination exists - archive the duplicate
				monitoring.Logger.Info("migration_archive_duplicate",
					"folder", folderName,
					"beads_id", shortTaskID,
					"reason", "destination_exists")
				if err := os.Rename(folderPath, filepath.Join(archiveDir, folderName)); err != nil {
					monitoring.Logger.Error("migration_archive_failed", "folder", folderName, "error", err.Error())
					skipped++
				} else {
					archived++
				}
				continue
			}

			// Perform the rename
			monitoring.Logger.Info("migration_rename",
				"from", folderName,
				"to", newName,
				"beads_id", shortTaskID)
			if err := os.Rename(folderPath, newPath); err != nil {
				monitoring.Logger.Error("migration_rename_failed",
					"folder", folderName,
					"new_name", newName,
					"error", err.Error())
				skipped++
			} else {
				renamed++
			}
		}
	}

	monitoring.Logger.Info("migration_complete",
		"renamed", renamed,
		"archived", archived,
		"skipped", skipped)

	return renamed, archived, skipped, nil
}
