// Package commands contains all cobra subcommands for the agent CLI.
package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// Version info injected at build time via ldflags.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// spawnFlags are registered on rootCmd so that `agent engineer xasm++-vp5 --stream` works.
var (
	rootSpawnWait            bool
	rootSpawnStream          bool
	rootSpawnInactiveTimeout time.Duration
)

// rootCmd is the parent cobra command. All subcommands are registered here.
var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "AI-Pack Agent CLI",
	Long: `AI-Pack Agent CLI — spawn and manage AI agents.

Examples:
  # Spawn an agent and stream real-time progress (RECOMMENDED)
  agent engineer xasm++-vp5 --stream

  # Check status
  agent status xasm++-vp5

  # Tail live logs
  agent logs xasm++-vp5 --follow`,
	// Allow unknown commands so that "agent engineer <task>" falls through to RunE.
	Args: cobra.ArbitraryArgs,
	// Do not print usage on error — keeps output clean.
	SilenceUsage: true,
	// RunE handles the "agent <role> <task>" dynamic spawn pattern.
	// Known subcommands are matched first by cobra; anything else lands here.
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		if len(args) < 2 {
			return fmt.Errorf("usage: agent <role> <task-id> [flags]\n\nExample: agent engineer xasm++-vp5 --stream")
		}
		role := args[0]
		taskInput := args[1]
		return runSpawn(role, taskInput, rootSpawnWait, rootSpawnStream, rootSpawnInactiveTimeout)
	},
}

// Execute is the entry point called from main.
func Execute(version, commit, buildTime string) {
	Version = version
	Commit = commit
	BuildTime = buildTime

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Register spawn flags on root so `agent engineer <task> --stream` works.
	rootCmd.Flags().BoolVar(&rootSpawnWait, "wait", false, "Wait for completion before exiting")
	rootCmd.Flags().BoolVar(&rootSpawnStream, "stream", false, "Stream live output while running")
	rootCmd.Flags().DurationVar(&rootSpawnInactiveTimeout, "inactive-timeout", 2*time.Minute, "Disconnect stream after this much inactivity (e.g. 5m, 30s)")

	rootCmd.AddCommand(
		newCreateCmd(),
		newShowCmd(),
		newUpdateCmd(),
		newCloseCmd(),
		newDeleteCmd(),
		newStatusCmd(),
		newResultsCmd(),
		newLogsCmd(),
		newListCmd(),
		newWaitCmd(),
		newDiffCmd(),
		newFilesCmd(),
		newCancelCmd(),
		newRetryCmd(),
		newResumeCmd(),
		newMetricsCmd(),
		newPerformanceCmd(),
		newDiscoveryCmd(),
		newVersionCmd(),
	)
}
