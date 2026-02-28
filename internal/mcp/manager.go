package mcp

import (
	"context"
	"fmt"
	"sync"
)

// Manager manages multiple MCP server connections
type Manager struct {
	mu      sync.RWMutex
	clients map[string]*Client // named clients (from config)

	projectMu      sync.RWMutex
	projectClients map[string]*Client // keyed by absolute projectRoot path
}

// NewManager creates a new MCP manager
func NewManager() *Manager {
	return &Manager{
		clients:        make(map[string]*Client),
		projectClients: make(map[string]*Client),
	}
}

// StartServer starts an MCP server with the given configuration
func (m *Manager) StartServer(ctx context.Context, name string, config ServerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already started
	if _, exists := m.clients[name]; exists {
		return fmt.Errorf("MCP server '%s' already started", name)
	}

	// Create and start client
	client := NewClient(name, config)
	if err := client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start MCP server '%s': %w", name, err)
	}

	m.clients[name] = client
	return nil
}

// EnsureProjectServer starts a KG server for projectRoot if not already running.
// cfg.Dir is overridden with projectRoot. Idempotent — no-op if already running.
func (m *Manager) EnsureProjectServer(ctx context.Context, projectRoot string, cfg ServerConfig) error {
	// Fast path: already running
	m.projectMu.RLock()
	_, exists := m.projectClients[projectRoot]
	m.projectMu.RUnlock()
	if exists {
		return nil
	}

	m.projectMu.Lock()
	defer m.projectMu.Unlock()

	// Double-check after acquiring write lock
	if _, exists := m.projectClients[projectRoot]; exists {
		return nil
	}

	cfg.Dir = projectRoot
	client := NewClient("kg:"+projectRoot, cfg)
	if err := client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start KG server for project '%s': %w", projectRoot, err)
	}

	m.projectClients[projectRoot] = client
	return nil
}

// GetClient returns a client by name
func (m *Manager) GetClient(name string) (*Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, ok := m.clients[name]
	if !ok {
		return nil, fmt.Errorf("MCP server '%s' not found", name)
	}

	return client, nil
}

// GetAllTools returns all tools from all connected named MCP servers
func (m *Manager) GetAllTools() map[string][]Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]Tool)
	for name, client := range m.clients {
		result[name] = client.GetTools()
	}

	return result
}

// GetProjectTools returns tools from the project-specific KG client for projectRoot.
// Returns nil if no client has been started for that root yet.
func (m *Manager) GetProjectTools(projectRoot string) []Tool {
	m.projectMu.RLock()
	defer m.projectMu.RUnlock()

	client, ok := m.projectClients[projectRoot]
	if !ok {
		return nil
	}
	return client.GetTools()
}

// findClientWithTool returns the first named client that has the given tool.
// Caller must hold m.mu.RLock.
func (m *Manager) findClientWithTool(toolName string) (*Client, string) {
	for _, client := range m.clients {
		for _, tool := range client.GetTools() {
			if tool.Name == toolName {
				return client, tool.Name
			}
		}
	}
	return nil, ""
}

// CallTool executes a tool on the appropriate MCP server
// Tool names should be prefixed with server name, e.g. "memory:create_entities"
func (m *Manager) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*CallToolResult, error) {
	m.mu.RLock()
	targetClient, actualToolName := m.findClientWithTool(toolName)
	m.mu.RUnlock()

	if targetClient == nil {
		return nil, fmt.Errorf("tool '%s' not found in any MCP server", toolName)
	}

	return targetClient.CallTool(ctx, actualToolName, arguments)
}

// CallToolInto finds the server hosting toolName and executes the tool,
// unmarshalling the raw JSON-RPC result into dest. Use this when the server
// returns a plain struct (not wrapped in CallToolResult).
func (m *Manager) CallToolInto(ctx context.Context, toolName string, arguments map[string]interface{}, dest interface{}) error {
	m.mu.RLock()
	targetClient, actualToolName := m.findClientWithTool(toolName)
	m.mu.RUnlock()

	if targetClient == nil {
		return fmt.Errorf("no MCP server found with tool: %s", toolName)
	}

	return targetClient.CallToolInto(ctx, actualToolName, arguments, dest)
}

// CallToolForProject routes to the project-specific KG client for projectRoot,
// falling back to named clients if no project client is found.
func (m *Manager) CallToolForProject(ctx context.Context, projectRoot string, toolName string, arguments map[string]interface{}) (*CallToolResult, error) {
	// Try project-specific client first
	m.projectMu.RLock()
	client, ok := m.projectClients[projectRoot]
	m.projectMu.RUnlock()

	if ok {
		return client.CallTool(ctx, toolName, arguments)
	}

	// Fall back to named clients
	m.mu.RLock()
	targetClient, actualToolName := m.findClientWithTool(toolName)
	m.mu.RUnlock()

	if targetClient == nil {
		return nil, fmt.Errorf("tool '%s' not found: no project client for '%s' and not in any named server", toolName, projectRoot)
	}

	return targetClient.CallTool(ctx, actualToolName, arguments)
}

// CallToolIntoForProject routes to the project-specific KG client for projectRoot,
// falling back to named clients, and unmarshals the result into dest.
func (m *Manager) CallToolIntoForProject(ctx context.Context, projectRoot string, toolName string, arguments map[string]interface{}, dest interface{}) error {
	// Try project-specific client first
	m.projectMu.RLock()
	client, ok := m.projectClients[projectRoot]
	m.projectMu.RUnlock()

	if ok {
		return client.CallToolInto(ctx, toolName, arguments, dest)
	}

	// Fall back to named clients
	m.mu.RLock()
	targetClient, actualToolName := m.findClientWithTool(toolName)
	m.mu.RUnlock()

	if targetClient == nil {
		return fmt.Errorf("tool '%s' not found: no project client for '%s' and not in any named server", toolName, projectRoot)
	}

	return targetClient.CallToolInto(ctx, actualToolName, arguments, dest)
}

// Close shuts down all MCP server connections
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	for name, client := range m.clients {
		if err := client.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("failed to close MCP server '%s': %w", name, err)
		}
	}
	m.clients = make(map[string]*Client)

	m.projectMu.Lock()
	defer m.projectMu.Unlock()

	for projectRoot, client := range m.projectClients {
		if err := client.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("failed to close KG server for project '%s': %w", projectRoot, err)
		}
	}
	m.projectClients = make(map[string]*Client)

	return firstErr
}

// GetActiveServers returns a list of active server names (named + project paths)
func (m *Manager) GetActiveServers() []string {
	m.mu.RLock()
	names := make([]string, 0, len(m.clients))
	for name := range m.clients {
		names = append(names, name)
	}
	m.mu.RUnlock()

	m.projectMu.RLock()
	for projectRoot := range m.projectClients {
		names = append(names, projectRoot)
	}
	m.projectMu.RUnlock()

	return names
}
