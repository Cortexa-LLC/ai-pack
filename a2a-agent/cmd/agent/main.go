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

const Version = "2.0.0-phase2"

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
	task := strings.Join(args[1:], " ")

	// Build agent:// URL
	taskEncoded := url.QueryEscape(task)
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
	fmt.Println("  agent [flags] <role> <task description>")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  agent engineer \"create hello world function\"")
	fmt.Println("  agent tester \"run all unit tests\"")
	fmt.Println("  agent --async reviewer \"review latest PR\"")
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("This command generates an agent:// URL and opens it using your")
	fmt.Println("system's registered protocol handler (agent-server).")
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
