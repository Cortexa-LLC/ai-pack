package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

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

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "agent-server.json", "Path to configuration file")
	maxConcurrent := flag.Int("max-concurrent", 0, "Override max concurrent agents (0 = use config)")
	port := flag.Int("port", 0, "Override server port (0 = use config)")
	serverMode := flag.Bool("server", false, "Run in server mode (HTTP/A2A protocol). Default: protocol handler mode")
	flag.Parse()

	// Check if we have a non-flag argument (agent:// URL)
	args := flag.Args()
	if len(args) > 0 && !*serverMode {
		// Protocol handler mode: handle agent:// URL directly
		handleProtocolURL(args[0], *configPath)
		return
	}

	// If -server flag not set and no URL provided, show usage
	if !*serverMode && len(args) == 0 {
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

	// Setup routes with logging middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)                // Health check
	mux.HandleFunc("/metrics", handleMetrics(s))           // Metrics endpoint
	mux.HandleFunc("/a2a/discovery", s.HandleA2ADiscovery) // A2A discovery
	mux.HandleFunc("/a2a/execute", s.HandleA2AExecute)     // A2A execute
	mux.HandleFunc("/a2a/status", s.HandleA2AStatus)       // A2A status
	mux.HandleFunc("/stream/", s.HandleStream)             // SSE streaming (tasks)
	mux.HandleFunc("/logs/stream", s.HandleLogsStream)     // SSE streaming (logs)
	mux.HandleFunc("/logs/recent", s.HandleLogsRecent)     // Recent logs (JSON)

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
	log.Printf("      - GET  /stream/:task_id    (Task progress - SSE)")
	log.Printf("      - GET  /logs/stream        (Realtime logs - SSE)")
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

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
