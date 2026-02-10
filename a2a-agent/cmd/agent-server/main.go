package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/config"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/protocol_handler"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/server"
)

const (
	ServerPort           = "8080"
	agentURLFormatMsg    = "agentURLFormatMsg"
	contentTypeJSON      = "application/json"
	formatTaskID         = "   Task ID: %s\n"
	formatStatus         = "   Status: %s\n"
)

// checkServerRunning checks if the agent-server is already running
func checkServerRunning(serverURL string) bool {
	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// delegateToRunningServer sends the task to an already-running server
func delegateToRunningServer(serverURL, role, task string, async bool) error {
	// Create A2A execute request
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "execute",
		"params": map[string]interface{}{
			"role": role,
			"task": task,
		},
		"id": 1,
	}

	bodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send POST request to /a2a/execute
	resp, err := http.Post(serverURL+"/a2a/execute", contentTypeJSON, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract task ID and status
	resultData, ok := result["result"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected response format")
	}

	taskID, _ := resultData["task_id"].(string)
	status, _ := resultData["status"].(string)

	if async {
		fmt.Println("✅ Task spawned in background!")
		fmt.Printf(formatTaskID, taskID)
		fmt.Printf(formatStatus, status)
		fmt.Println()
		fmt.Printf("   Track progress: %s/stream/%s\n", serverURL, taskID)
	} else {
		fmt.Println("✅ Task delegated to server!")
		fmt.Printf(formatTaskID, taskID)
		fmt.Printf(formatStatus, status)
		fmt.Println()
		fmt.Printf("   Track progress: %s/stream/%s\n", serverURL, taskID)
	}

	return nil
}

// handleProtocolURL handles agent:// protocol URLs directly
func handleProtocolURL(agentURL, configPath string) {
	// Parse agent:// URL
	// Format: agent://role/task-description?async=true
	if !strings.HasPrefix(agentURL, "agent://") {
		fmt.Printf("❌ Invalid URL: %s\n", agentURL)
		fmt.Println(agentURLFormatMsg)
		os.Exit(1)
	}

	// Parse the URL
	parsedURL, err := url.Parse(agentURL)
	if err != nil {
		fmt.Printf("❌ Failed to parse URL: %v\n", err)
		os.Exit(1)
	}

	// Extract role from host
	role := parsedURL.Host
	if role == "" {
		fmt.Printf("❌ Invalid URL format: %s\n", agentURL)
		fmt.Println(agentURLFormatMsg)
		os.Exit(1)
	}

	// Extract task from path
	taskEncoded := strings.TrimPrefix(parsedURL.Path, "/")
	if taskEncoded == "" {
		fmt.Printf("❌ Invalid URL format: %s\n", agentURL)
		fmt.Println(agentURLFormatMsg)
		os.Exit(1)
	}

	// URL decode the task
	// Use PathUnescape to properly handle + characters (QueryUnescape treats + as space)
	task, err := url.PathUnescape(taskEncoded)
	if err != nil {
		task = taskEncoded // Use as-is if decode fails
	}

	// Check for async parameter
	async := parsedURL.Query().Get("async") == "true"

	fmt.Println("🔗 AI-Pack Agent Protocol Handler")
	fmt.Printf("   Role: %s\n", role)
	fmt.Printf("   Task: %s\n", task)
	if async {
		fmt.Printf("   Mode: Async (background)\n")
	} else {
		fmt.Printf("   Mode: Sync (wait for completion)\n")
	}
	fmt.Println()

	// Check if a server is already running
	serverURL := "http://localhost:" + ServerPort
	if checkServerRunning(serverURL) {
		fmt.Println("✓ Found running agent-server, delegating to it...")
		fmt.Println()
		if err := delegateToRunningServer(serverURL, role, task, async); err != nil {
			fmt.Printf("❌ Failed to delegate to running server: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// No server running, execute in single-shot mode
	fmt.Println("No running server found, executing in single-shot mode...")
	fmt.Println()

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize monitoring (minimal for single-shot mode)
	monitoring.InitLogger(slog.LevelInfo)
	monitoring.InitMetrics()

	// Get root directory
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Create server instance (but don't start HTTP server)
	s, err := server.NewAgentServer(
		rootDir,
		1, // Single task mode
		cfg.API.MaxTokens,
		cfg.API.AnthropicModel,
		cfg,
	)
	if err != nil {
		log.Fatalf("Failed to create agent server: %v", err)
	}

	fmt.Println("🚀 Executing task...")
	fmt.Println()

	if async {
		// Execute asynchronously - spawn and return immediately
		result, err := s.SpawnAgentTask(role, task)
		if err != nil {
			fmt.Printf("❌ Failed to spawn task: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ Task spawned in background!")
		fmt.Printf(formatTaskID, result.TaskID)
		fmt.Printf(formatStatus, result.Status)
		fmt.Println()
		fmt.Printf("   Results will be saved to: .beads/tasks/%s/\n", result.TaskID)
		fmt.Println()
		fmt.Println("   Task is running in the background.")
		fmt.Println("   Check the task directory for results.")
	} else {
		// Execute synchronously - wait for completion
		result, err := s.ExecuteTaskSync(role, task)
		if err != nil {
			fmt.Printf("❌ Task failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ Task completed successfully!")
		fmt.Printf(formatTaskID, result.TaskID)
		fmt.Printf(formatStatus, result.Status)
		fmt.Println()
		fmt.Printf("   Results saved to: .beads/tasks/%s/\n", result.TaskID)
		fmt.Println()
	}
}

// Health check endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":  "healthy",
		"version": server.Version,
		"server":  "ai-pack-agent-server",
		"features": map[string]bool{
			"a2a_protocol":       true,
			"sse_streaming":      true,
			"parallel_execution": true,
			"monitoring":         true,
		},
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	json.NewEncoder(w).Encode(health)
}

// Metrics endpoint
func handleMetrics(s *server.AgentServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshot := s.GetMetricsSnapshot()

		w.Header().Set("Content-Type", contentTypeJSON)
		json.NewEncoder(w).Encode(snapshot)
	}
}

func printUsage() {
	fmt.Println("AI-Pack Agent Server - Usage:")
	fmt.Println("")
	fmt.Println("  Server Mode:")
	fmt.Println("    agent-server --server [--config agent-server.json]")
	fmt.Println("")
	fmt.Println("  Then use the agent CLI to spawn tasks:")
	fmt.Println("    agent engineer <beads-task-id>")
	fmt.Println("    agent engineer <beads-task-id> --wait")
	fmt.Println("")
	fmt.Println("  Example:")
	fmt.Println("    # Start server")
	fmt.Println("    agent-server --server")
	fmt.Println("")
	fmt.Println("    # Spawn agent (in another terminal)")
	fmt.Println("    agent engineer bd-abc123")
	fmt.Println("")
}

func parseCommandLine() (configPath *string, maxConcurrent *int, port *int, serverMode *bool, migrateTasks *bool, args []string) {
	configPath = flag.String("config", "agent-server.json", "Path to configuration file")
	maxConcurrent = flag.Int("max-concurrent", 0, "Override max concurrent agents (0 = use config)")
	port = flag.Int("port", 0, "Override server port (0 = use config)")
	serverMode = flag.Bool("server", false, "Run in server mode (HTTP/A2A protocol). Default: protocol handler mode")
	migrateTasks = flag.Bool("migrate-tasks", false, "Migrate legacy task-* folders to beads-id-timestamp format")
	flag.Parse()
	args = flag.Args()
	return
}

func main() {
	configPath, maxConcurrent, port, serverMode, migrateTasks, args := parseCommandLine()

	// Check if we have a non-flag argument (agent:// URL)
	if len(args) > 0 && !*serverMode && !*migrateTasks {
		// Protocol handler mode: handle agent:// URL directly
		handleProtocolURL(args[0], *configPath)
		return
	}

	// If -server flag not set and no URL provided, show usage
	if !*serverMode && !*migrateTasks && len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Apply command-line overrides
	if *maxConcurrent > 0 {
		cfg.Server.MaxConcurrentAgents = *maxConcurrent
		cfg.Server.WorkerPoolSize = *maxConcurrent
	}
	if *port > 0 {
		cfg.Server.Port = *port
	}

	// Initialize monitoring
	var logLevel slog.Level
	switch cfg.Logging.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	monitoring.InitLogger(logLevel)
	monitoring.InitMetrics()

	monitoring.Logger.Info("server_starting",
		"version", server.Version,
		"max_concurrent", cfg.Server.MaxConcurrentAgents,
		"model", cfg.API.AnthropicModel)

	// Get root directory
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Create server with configuration
	s, err := server.NewAgentServer(
		rootDir,
		cfg.Server.MaxConcurrentAgents,
		cfg.API.MaxTokens,
		cfg.API.AnthropicModel,
		cfg,
	)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle migration mode
	if *migrateTasks {
		fmt.Println("🔄 Running task folder migration...")
		fmt.Println()

		renamed, archived, skipped, err := s.MigrateTaskFolders()
		if err != nil {
			log.Fatalf("❌ Migration failed: %v", err)
		}

		fmt.Println("✅ Migration complete!")
		fmt.Printf("   Renamed: %d folders\n", renamed)
		fmt.Printf("   Archived: %d folders (free-form tasks)\n", archived)
		if skipped > 0 {
			fmt.Printf("   Skipped: %d folders (errors)\n", skipped)
		}
		fmt.Println()
		fmt.Println("You can now start the server with: agent-server --server")
		return
	}

	// Check for legacy task folders before starting server
	hasLegacy, legacyFolders := s.DetectLegacyTaskFolders()
	if hasLegacy {
		fmt.Println()
		fmt.Println("⚠️  MIGRATION REQUIRED")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()
		fmt.Printf("Found %d legacy task folder(s) that need migration:\n", len(legacyFolders))
		fmt.Println()

		// Show first few examples
		maxShow := 5
		for i, folder := range legacyFolders {
			if i >= maxShow {
				fmt.Printf("   ... and %d more\n", len(legacyFolders)-maxShow)
				break
			}
			fmt.Printf("   %s\n", folder)
		}

		fmt.Println()
		fmt.Println("Legacy task folders use the format: task-{role}-{timestamp}")
		fmt.Println("They must be migrated to: {beads-id}-{timestamp}")
		fmt.Println()
		fmt.Println("To migrate:")
		fmt.Println("   agent-server --migrate-tasks")
		fmt.Println()
		fmt.Println("This will:")
		fmt.Println("   • Rename folders with Beads IDs to new format")
		fmt.Println("   • Archive folders with free-form descriptions")
		fmt.Println("   • Preserve all task data and history")
		fmt.Println()
		os.Exit(1)
	}

	// Setup routes with logging middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)                // Health check
	mux.HandleFunc("/metrics", handleMetrics(s))           // Metrics endpoint
	mux.HandleFunc("/a2a/discovery", s.HandleA2ADiscovery) // A2A discovery
	mux.HandleFunc("/a2a/execute", s.HandleA2AExecute)     // A2A execute
	mux.HandleFunc("/a2a/status/", s.HandleA2AStatus)      // A2A status (trailing slash for subpaths)
	mux.HandleFunc("/a2a/status", s.HandleA2AStatus)       // A2A status (POST JSON-RPC)
	mux.HandleFunc("/a2a/tasks", s.HandleTasksList)        // List all tasks (machine-wide)
	mux.HandleFunc("/a2a/tasks/", s.HandleTaskLogs)        // Task-specific logs (trailing slash for subpaths)
	mux.HandleFunc("/a2a/cancel/", s.HandleCancelTask)     // Cancel a running task
	mux.HandleFunc("/a2a/retry/", s.HandleRetryTask)       // Retry a failed task
	mux.HandleFunc("/a2a/start/", s.HandleStartTask)       // Start an agent for a task
	mux.HandleFunc("/stream/", s.HandleStream)             // SSE streaming (tasks)
	mux.HandleFunc("/logs/stream", s.HandleLogsStream)     // SSE streaming (logs)
	mux.HandleFunc("/logs/recent", s.HandleLogsRecent)     // Recent logs (JSON)
	mux.HandleFunc("/api/chat", s.HandleChat)              // Chat with Claude (SSE streaming)
	mux.HandleFunc("/api/chat/options", s.HandleChatOptions) // CORS preflight
	mux.HandleFunc("/api/browse-directories", s.HandleBrowseDirectories) // Directory autocomplete

	// Setup GraphQL endpoints
	s.SetupGraphQLHandlers(mux)

	// Wrap with logging middleware
	handler := server.LoggingMiddleware(mux)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("")
	log.Printf("🚀 AI-Pack Agent Server v%s starting on %s", server.Version, addr)
	log.Printf("")
	log.Printf("   📡 A2A Protocol Endpoints:")
	log.Printf("      - GET  /a2a/discovery      (Agent capabilities)")
	log.Printf("      - POST /a2a/execute        (Execute task - JSON-RPC 2.0)")
	log.Printf("      - POST /a2a/status         (Task status - JSON-RPC 2.0)")
	log.Printf("")
	log.Printf("   🔄 Streaming:")
	log.Printf("      - GET  /stream/:task_id            (Task progress - SSE)")
	log.Printf("      - GET  /logs/stream                (Server logs - SSE)")
	log.Printf("      - GET  /a2a/tasks/:task_id/logs    (Task logs - plain text)")
	log.Printf("      - GET  /a2a/tasks/:task_id/logs?stream=true (Task logs - SSE)")
	log.Printf("")
	log.Printf("   📊 GraphQL:")
	log.Printf("      - POST /graphql            (GraphQL API)")
	log.Printf("      - GET  /playground         (GraphQL Playground UI)")
	log.Printf("")
	log.Printf("   💬 Chat:")
	log.Printf("      - POST /api/chat           (Chat with Claude - SSE streaming)")
	log.Printf("")
	log.Printf("   🔧 Utility:")
	log.Printf("      - GET  /health             (Health check)")
	log.Printf("      - GET  /metrics            (Performance metrics)")
	log.Printf("      - GET  /logs/recent?limit=N (Recent logs - JSON)")
	log.Printf("")
	log.Printf("   🎯 Features:")
	log.Printf("      - A2A Protocol Compliance  ✅")
	log.Printf("      - SSE Streaming            ✅")
	log.Printf("      - Parallel Execution       ✅ (max %d concurrent)", s.GetMaxConcurrent())
	log.Printf("      - Structured Logging       ✅ (JSON format)")
	log.Printf("      - Performance Metrics      ✅")
	log.Printf("")
	log.Printf("   ⚙️  Configuration:")
	log.Printf("      - Model: %s", cfg.API.AnthropicModel)
	log.Printf("      - Max Tokens: %d", cfg.API.MaxTokens)
	log.Printf("      - API Mode: %s", cfg.API.Mode)
	if cfg.API.Mode == "proxy" && cfg.API.Proxy != nil {
		log.Printf("      - Proxy URL: %s", cfg.API.Proxy.BaseURL)
	}
	log.Printf("      - Config File: %s", *configPath)
	log.Printf("")
	log.Printf("   📂 Working Directory: %s", rootDir)
	log.Printf("      ⚠️  All spawned agents will execute in this directory")
	log.Printf("")

	// Register agent:// protocol handler (first run)
	if !protocol_handler.IsRegistered() {
		log.Printf("🔗 Registering agent:// protocol handler...")
		log.Printf("")
		if err := protocol_handler.Register(); err != nil {
			log.Printf("⚠️  Could not auto-register protocol handler: %v", err)
			log.Printf("   You can use the 'agent' CLI or set it up manually")
			log.Printf("   See: docs/content/framework/protocol-handler.md")
		}
		log.Printf("")
	}

	monitoring.Logger.Info("server_ready",
		"address", addr,
		"max_concurrent", s.GetMaxConcurrent(),
		"model", cfg.API.AnthropicModel)

	// Setup HTTP server with graceful shutdown
	httpServer := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		monitoring.Logger.Info("http_server_starting", "address", addr)
		serverErrors <- httpServer.ListenAndServe()
	}()

	// Channel to listen for interrupt or terminate signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		log.Fatalf("Server failed: %v", err)

	case sig := <-shutdown:
		monitoring.Logger.Info("shutdown_signal_received", "signal", sig.String())
		log.Printf("")
		log.Printf("🛑 Shutdown signal received (%s)", sig)

		// Check for active tasks before shutting down
		activeCount := s.GetActiveTaskCount()
		if activeCount > 0 {
			log.Printf("")
			log.Printf("⚠️  WARNING: %d active task(s) currently running:", activeCount)
			log.Printf("")

			activeTasks := s.GetActiveTaskIDs()
			for i, task := range activeTasks {
				log.Printf("   %d. Task: %s", i+1, task["task_id"])
				log.Printf("      Role: %s", task["role"])
				log.Printf("      Status: %s", task["status"])
			}

			log.Printf("")
			log.Printf("❌ Cannot shutdown with active tasks running")
			log.Printf("")
			log.Printf("Options:")
			log.Printf("  1. Wait for tasks to complete naturally")
			log.Printf("  2. Cancel tasks first:")
			for _, task := range activeTasks {
				log.Printf("     agent cancel %s", task["task_id"])
			}
			log.Printf("  3. Force shutdown (will kill active tasks): kill -9 %d", os.Getpid())
			log.Printf("")

			// Don't shutdown - keep server running
			monitoring.Logger.Warn("shutdown_cancelled_active_tasks", "count", activeCount)
			log.Printf("Server still running. Press Ctrl+C again after handling tasks.")
			log.Printf("")

			// Wait for another signal
			<-shutdown
			log.Printf("")
			log.Printf("🛑 Second shutdown signal received - forcing shutdown")
			log.Printf("")
		}

		// Give active tasks a moment to finish (5 seconds max)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		log.Printf("Shutting down server gracefully...")
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}

		// Check final state
		if err := s.Shutdown(ctx); err != nil {
			log.Printf("Warning during shutdown: %v", err)
		}

		monitoring.Logger.Info("server_shutdown_complete")
		log.Printf("✅ Server shutdown complete")
	}
}
