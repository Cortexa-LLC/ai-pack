package mcp

import (
	"context"
	"fmt"
	"sync"
)

// Manager manages multiple MCP server connections
type Manager struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

// NewManager creates a new MCP manager
func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*Client),
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

// GetAllTools returns all tools from all connected MCP servers
func (m *Manager) GetAllTools() map[string][]Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]Tool)
	for name, client := range m.clients {
		result[name] = client.GetTools()
	}

	return result
}

// CallTool executes a tool on the appropriate MCP server
// Tool names should be prefixed with server name, e.g. "memory:create_entities"
func (m *Manager) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*CallToolResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Find which server has this tool
	var targetClient *Client
	var actualToolName string

	for _, client := range m.clients {
		for _, tool := range client.GetTools() {
			if tool.Name == toolName {
				targetClient = client
				actualToolName = tool.Name
				break
			}
		}
		if targetClient != nil {
			break
		}
	}

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
	defer m.mu.RUnlock()

	var targetClient *Client
	var actualToolName string

	for _, client := range m.clients {
		for _, tool := range client.GetTools() {
			if tool.Name == toolName {
				targetClient = client
				actualToolName = tool.Name
				break
			}
		}
		if targetClient != nil {
			break
		}
	}

	if targetClient == nil {
		return fmt.Errorf("no MCP server found with tool: %s", toolName)
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
	return firstErr
}

// GetActiveServers returns a list of active server names
func (m *Manager) GetActiveServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.clients))
	for name := range m.clients {
		names = append(names, name)
	}

	return names
}
