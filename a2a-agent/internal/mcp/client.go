package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
)

// Client represents an MCP client connected to a server
type Client struct {
	name    string
	config  ServerConfig
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	scanner *bufio.Scanner

	// Request ID counter
	requestID atomic.Int64

	// Response tracking
	responseMu sync.Mutex
	responses  map[int64]chan *JSONRPCResponse

	// Server info and capabilities
	serverInfo   ServerInfo
	capabilities ServerCapabilities
	tools        []Tool

	// Lifecycle
	initialized bool
	closed      bool
	closeMu     sync.Mutex
}

// NewClient creates a new MCP client
func NewClient(name string, config ServerConfig) *Client {
	return &Client{
		name:      name,
		config:    config,
		responses: make(map[int64]chan *JSONRPCResponse),
	}
}

// Start launches the MCP server process and initializes the connection
func (c *Client) Start(ctx context.Context) error {
	// Create command
	c.cmd = exec.CommandContext(ctx, c.config.Command, c.config.Args...)

	// Set environment
	c.cmd.Env = os.Environ()
	for key, value := range c.config.Env {
		c.cmd.Env = append(c.cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Set up pipes
	stdin, err := c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	c.stdin = stdin

	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	c.stdout = stdout

	// Capture stderr for debugging
	c.cmd.Stderr = os.Stderr

	// Start process
	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	// Start reading responses
	c.scanner = bufio.NewScanner(c.stdout)
	go c.readLoop()

	// Initialize connection
	if err := c.initialize(ctx); err != nil {
		c.Close()
		return fmt.Errorf("failed to initialize MCP connection: %w", err)
	}

	// List tools
	if err := c.listTools(ctx); err != nil {
		c.Close()
		return fmt.Errorf("failed to list MCP tools: %w", err)
	}

	c.initialized = true
	return nil
}

// initialize sends the initialize request
func (c *Client) initialize(ctx context.Context) error {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    Capabilities{},
		ClientInfo: ClientInfo{
			Name:    "a2a-agent",
			Version: "1.0.0",
		},
	}

	var result InitializeResult
	if err := c.call(ctx, "initialize", params, &result); err != nil {
		return err
	}

	c.serverInfo = result.ServerInfo
	c.capabilities = result.Capabilities

	// Send initialized notification
	notification := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal initialized notification: %w", err)
	}

	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}

	return nil
}

// listTools requests the list of tools from the server
func (c *Client) listTools(ctx context.Context) error {
	var result ListToolsResult
	if err := c.call(ctx, "tools/list", nil, &result); err != nil {
		return err
	}

	c.tools = result.Tools
	return nil
}

// CallTool executes a tool on the MCP server
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*CallToolResult, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	params := CallToolParams{
		Name:      name,
		Arguments: arguments,
	}

	var result CallToolResult
	if err := c.call(ctx, "tools/call", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTools returns the list of available tools
func (c *Client) GetTools() []Tool {
	return c.tools
}

// GetServerInfo returns the server information
func (c *Client) GetServerInfo() ServerInfo {
	return c.serverInfo
}

// call sends a JSON-RPC request and waits for the response
func (c *Client) call(ctx context.Context, method string, params interface{}, result interface{}) error {
	// Generate request ID
	id := c.requestID.Add(1)

	// Create response channel
	responseChan := make(chan *JSONRPCResponse, 1)
	c.responseMu.Lock()
	c.responses[id] = responseChan
	c.responseMu.Unlock()

	// Clean up response channel after request
	defer func() {
		c.responseMu.Lock()
		delete(c.responses, id)
		c.responseMu.Unlock()
	}()

	// Build request
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	// Marshal and send
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response
	select {
	case response := <-responseChan:
		if response.Error != nil {
			return fmt.Errorf("JSON-RPC error %d: %s", response.Error.Code, response.Error.Message)
		}

		// Unmarshal result
		if result != nil {
			resultData, err := json.Marshal(response.Result)
			if err != nil {
				return fmt.Errorf("failed to marshal response result: %w", err)
			}

			if err := json.Unmarshal(resultData, result); err != nil {
				return fmt.Errorf("failed to unmarshal response result: %w", err)
			}
		}

		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

// readLoop continuously reads responses from the MCP server
func (c *Client) readLoop() {
	for c.scanner.Scan() {
		line := c.scanner.Bytes()

		var response JSONRPCResponse
		if err := json.Unmarshal(line, &response); err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "MCP client %s: failed to unmarshal response: %v\n", c.name, err)
			continue
		}

		// Route response to waiting caller
		if response.ID != nil {
			// Convert ID to int64
			var id int64
			switch v := response.ID.(type) {
			case float64:
				id = int64(v)
			case int64:
				id = v
			case int:
				id = int64(v)
			default:
				fmt.Fprintf(os.Stderr, "MCP client %s: unexpected ID type: %T\n", c.name, v)
				continue
			}

			c.responseMu.Lock()
			if ch, ok := c.responses[id]; ok {
				ch <- &response
			}
			c.responseMu.Unlock()
		}
	}

	if err := c.scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "MCP client %s: scanner error: %v\n", c.name, err)
	}
}

// Close shuts down the MCP client
func (c *Client) Close() error {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	// Close stdin (signals server to shut down)
	if c.stdin != nil {
		c.stdin.Close()
	}

	// Wait for process to exit
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Wait()
	}

	// Close stdout
	if c.stdout != nil {
		c.stdout.Close()
	}

	return nil
}

// Name returns the client name
func (c *Client) Name() string {
	return c.name
}
