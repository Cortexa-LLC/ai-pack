package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const Version = "2.1.0"

func main() {
	// Parse command-line flags
	async := flag.Bool("async", false, "Execute task asynchronously (return immediately)")
	version := flag.Bool("version", false, "Show version information")
	flag.Usage = usage
	flag.Parse()

	if *version {
		fmt.Printf("AI-Pack Agent CLI v%s\n", Version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) < 2 {
		usage()
		os.Exit(1)
	}

	role := args[0]
	taskInput := strings.Join(args[1:], " ")

	// Check if input is a Beads task ID by querying Beads
	// This works with ANY Beads prefix (bd-, xasm++-, etc.)
	isBeadsTask := isValidBeadsTask(taskInput)

	if isBeadsTask {
		fmt.Printf("🎯 Beads task: %s\n", taskInput)
	} else {
		// Free-form description - show deprecation warning
		fmt.Printf("⚠️  Using free-form description (deprecated)\n")
		fmt.Printf("   For better task tracking, use Beads:\n")
		fmt.Printf("     bd create \"%s\"\n", taskInput)
		fmt.Printf("     agent %s <task-id>\n", role)
		fmt.Println()
	}

	// Build agent:// URL
	// Use PathEscape for URL path segments (preserves + and other special chars)
	taskEncoded := url.PathEscape(taskInput)
	agentURL := fmt.Sprintf("agent://%s/%s", role, taskEncoded)

	// Add async parameter if requested
	if *async {
		agentURL += "?async=true"
	}

	fmt.Printf("🔗 Opening: %s\n", agentURL)
	fmt.Println()

	// Open the URL (OS will invoke the registered protocol handler)
	if err := openURL(agentURL); err != nil {
		fmt.Printf("❌ Failed to open agent:// URL: %v\n", err)
		fmt.Println()
		fmt.Println("Make sure the agent:// protocol is registered.")
		fmt.Println("See PROTOCOL-REGISTRATION.md for setup instructions.")
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("AI-Pack Agent CLI - Convenience wrapper for agent:// protocol")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  agent [flags] <role> <beads-task-id>          (recommended)")
	fmt.Println("  agent [flags] <role> <task description>       (deprecated)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Recommended: Use Beads task tracking")
	fmt.Println("  bd create \"Create hello world function\"")
	fmt.Println("  agent engineer bd-a1b2")
	fmt.Println()
	fmt.Println("  # Deprecated: Free-form description")
	fmt.Println("  agent engineer \"create hello world function\"")
	fmt.Println()
	fmt.Println("  # Async execution")
	fmt.Println("  agent --async tester bd-x7z9")
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("This command generates an agent:// URL and opens it using your")
	fmt.Println("system's registered protocol handler (agent-server).")
	fmt.Println()
	fmt.Println("For best results, create tasks in Beads first:")
	fmt.Println("  bd create \"Task description\"")
	fmt.Println("  bd show bd-xxxx")
	fmt.Println("  agent <role> bd-xxxx")
}

func isValidBeadsTask(taskID string) bool {
	// Quick format check - must contain a separator like - or +
	if !strings.Contains(taskID, "-") && !strings.Contains(taskID, "+") {
		return false
	}

	// Ask Beads if this task exists (works with any prefix)
	cmd := exec.Command("bd", "show", taskID)
	err := cmd.Run()
	return err == nil
}

func openURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Run()
}
