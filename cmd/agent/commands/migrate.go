package commands

import (
	"github.com/cortexa-llc/ai-pack/internal/constants"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cortexa-llc/ai-pack/internal/taskdb"
	"github.com/spf13/cobra"
)

func newMigrateCmd() *cobra.Command {
	var projectRoot string
	var all bool

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate tasks from Beads to SQLite task database",
		Long: `Migrate task data from .beads/tasks directories to the new SQLite task database.

This is a one-time migration from the old Beads-based task tracking to the new
SQLite-based system. It will:
  1. Scan .beads/tasks directories for task metadata
  2. Import tasks into ~/.ai-pack/tasks.db
  3. Skip tasks that are already in the database

Run this after upgrading to the new task tracking system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrate(projectRoot, all)
		},
	}

	cmd.Flags().StringVar(&projectRoot, "project", ".", "Project root directory to migrate (default: current directory)")
	cmd.Flags().BoolVar(&all, "all", false, "Migrate all projects (scan common project locations)")

	return cmd
}

func runMigrate(projectRoot string, all bool) error {
	// Open task database
	dbPath := filepath.Join(os.Getenv("HOME"), ".ai-pack", "tasks.db")
	db, err := taskdb.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open task database: %w", err)
	}
	defer db.Close()

	if all {
		return migrateAllProjects(db)
	}

	return migrateSingleProject(db, projectRoot)
}

func migrateSingleProject(db *taskdb.DB, projectRoot string) error {
	// Resolve absolute path
	absPath, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve project path: %w", err)
	}

	fmt.Printf("Migrating tasks from: %s\n", absPath)

	count, err := db.MigrateFromLegacy(absPath)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	if count == 0 {
		fmt.Println("✅ No tasks to migrate (or all tasks already migrated)")
	} else {
		fmt.Printf("✅ Migrated %d tasks to SQLite database\n", count)
	}

	return nil
}

func migrateAllProjects(db *taskdb.DB) error {
	// Common project locations to scan
	homeDir := os.Getenv("HOME")
	projectDirs := []string{
		filepath.Join(homeDir, "Projects"),
		filepath.Join(homeDir, "Code"),
		filepath.Join(homeDir, "workspace"),
		filepath.Join(homeDir, "src"),
	}

	totalMigrated := 0
	projectsScanned := 0

	for _, baseDir := range projectDirs {
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			continue
		}

		// Walk subdirectories looking for .beads directories
		err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip inaccessible directories
			}

			// Look for .beads/tasks directories
			if info.IsDir() && info.Name() == constants.TaskRootDir {
				projectRoot := filepath.Dir(path)
				tasksDir := filepath.Join(path, "tasks")

				if _, err := os.Stat(tasksDir); err == nil {
					fmt.Printf("Found project: %s\n", projectRoot)
					count, err := db.MigrateFromLegacy(projectRoot)
					if err != nil {
						fmt.Printf("  ⚠️  Migration error: %v\n", err)
					} else if count > 0 {
						fmt.Printf("  ✅ Migrated %d tasks\n", count)
						totalMigrated += count
					}
					projectsScanned++
				}

				return filepath.SkipDir // Don't recurse into .beads
			}

			return nil
		})

		if err != nil {
			fmt.Printf("Warning: error scanning %s: %v\n", baseDir, err)
		}
	}

	fmt.Printf("\n📊 Migration complete:\n")
	fmt.Printf("   Projects scanned: %d\n", projectsScanned)
	fmt.Printf("   Tasks migrated:   %d\n", totalMigrated)

	return nil
}
