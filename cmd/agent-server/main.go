package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/cortexa-llc/ai-pack/internal/config"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/server"
)

const (
	ServerPort = "8080"
)

// handleProtocolURL handles agent:// protocol URLs directly
func handleProtocolURL(agentURL, configPath string) {
	// Parse agent:// URL
	// Format: agent://role/task-description?async=true
	if !strings.HasPrefix(agentURL, "agent://") {
		fmt.Printf("❌ Invalid URL: %s\n", agentURL)
		fmt.Println("   Expected format: agent://role/task-description")
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
		fmt.Println("   Expected format: agent://role/task-description")
		os.Exit(1)
	}

	// Extract task from path
	taskEncoded := strings.TrimPrefix(parsedURL.Path, "/")
	if taskEncoded == "" {
		fmt.Printf("❌ Invalid URL format: %s\n", agentURL)
		fmt.Println("   Expected format: agent://role/task-description")
		os.Exit(1)
	}

	// URL decode the task
	task, err := url.QueryUnescape(taskEncoded)
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
		&cfg.API,
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
		fmt.Printf("   Task ID: %s\n", result.TaskID)
		fmt.Printf("   Status: %s\n", result.Status)
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
		fmt.Printf("   Task ID: %s\n", result.TaskID)
		fmt.Printf("   Status: %s\n", result.Status)
		fmt.Println()
		fmt.Printf("   Results saved to: .beads/tasks/%s/\n", result.TaskID)
		fmt.Println()
	}
}

// Legacy /spawn endpoint for backward compatibility
func handleLegacySpawn(s *server.AgentServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Role string `json:"role"`
			Task string `json:"task"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		monitoring.Logger.Info("legacy_spawn_request", "role", req.Role, "task", req.Task)

		result, err := s.SpawnAgentTask(req.Role, req.Task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// Metrics endpoint
func handleMetrics(s *server.AgentServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshot := s.GetMetricsSnapshot()

		w.Header().Set("Content-Type", "application/json")
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
		fmt.Println("AI-Pack Agent - Usage:")
		fmt.Println("")
		fmt.Println("  Protocol Handler Mode (default):")
		fmt.Println("    agent-server agent://role/task-description")
		fmt.Println("    Example: agent-server agent://engineer/create-hello-world")
		fmt.Println("")
		fmt.Println("  Server Mode (optional - for A2A protocol):")
		fmt.Println("    agent-server -server [-config agent-server.json]")
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
		&cfg.API,
	)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup routes with logging middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/spawn", handleLegacySpawn(s))          // Legacy endpoint
	mux.HandleFunc("/health", handleHealth)                  // Health check
	mux.HandleFunc("/metrics", handleMetrics(s))            // Metrics endpoint
	mux.HandleFunc("/a2a/discovery", s.HandleA2ADiscovery)  // A2A discovery
	mux.HandleFunc("/a2a/execute", s.HandleA2AExecute)      // A2A execute
	mux.HandleFunc("/a2a/status", s.HandleA2AStatus)        // A2A status
	mux.HandleFunc("/stream/", s.HandleStream)              // SSE streaming

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
	log.Printf("      - GET  /stream/:task_id    (Real-time progress - SSE)")
	log.Printf("")
	log.Printf("   🔧 Legacy & Utility:")
	log.Printf("      - POST /spawn              (Legacy spawn endpoint)")
	log.Printf("      - GET  /health             (Health check)")
	log.Printf("      - GET  /metrics            (Performance metrics)")
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
		log.Printf("      - Proxy Type: %s", cfg.API.Proxy.Type)
		log.Printf("      - Proxy URL: %s", cfg.API.Proxy.BaseURL)
	}
	log.Printf("      - Config File: %s", *configPath)
	log.Printf("")
	log.Printf("   📂 Root directory: %s", rootDir)
	log.Printf("")

	monitoring.Logger.Info("server_ready",
		"address", addr,
		"max_concurrent", s.GetMaxConcurrent(),
		"model", cfg.API.AnthropicModel)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
