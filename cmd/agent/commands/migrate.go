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
	var force bool

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate tasks from Beads to SQLite task database",
		Long: `Migrate task data from Beads (issues.jsonl) into the SQLite task database.

This migration is normally run once automatically on server startup. Use this
command to run it manually or to re-run it with --force if something went wrong.

It will:
  1. Scan ~/Projects (and other common locations) for .beads/issues.jsonl files
  2. Import all Beads task records into ~/.ai-pack/tasks.db
  3. Skip tasks that are already in the database
  4. Record completion so it does not run again on startup`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrate(projectRoot, all, force)
		},
	}

	cmd.Flags().StringVar(&projectRoot, "project", ".", "Project root directory to migrate (default: current directory)")
	cmd.Flags().BoolVar(&all, "all", false, "Also scan per-project agent execution metadata (.beads/tasks/*/00-metadata.json)")
	cmd.Flags().BoolVar(&force, "force", false, "Re-run even if already marked complete")

	return cmd
}

func runMigrate(projectRoot string, all bool, force bool) error {
	// Open task database
	dbPath := filepath.Join(os.Getenv("HOME"), ".ai-pack", "tasks.db")
	db, err := taskdb.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open task database: %w", err)
	}
	defer db.Close()

	if force {
		if err := db.ResetMigration(taskdb.MigrationBeadsJSONL); err != nil {
			return fmt.Errorf("reset migration flag: %w", err)
		}
		fmt.Println("🔄 Migration flag reset — will re-run Beads import")
	}

	// Primary migration: import from Beads issues.jsonl files.
	// Does not require the Dolt server. One-time; skipped if already done.
	fmt.Println("📋 Migrating from Beads task history (issues.jsonl)...")
	beadsCount, err := db.MigrateFromBeadsJSONL()
	if err != nil {
		fmt.Printf("  ⚠️  Beads migration error: %v\n", err)
	} else if beadsCount == 0 {
		fmt.Println("  (no new Beads tasks to import, or already migrated)")
	} else {
		fmt.Printf("  ✅ Imported %d tasks from Beads\n", beadsCount)
	}

	// Secondary: migrate per-project agent execution metadata (.beads/tasks/*/00-metadata.json)
	if all {
		fmt.Println("📋 Migrating agent execution metadata...")
		return migrateAllProjects(db)
	}
	return nil
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
