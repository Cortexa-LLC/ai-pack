package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cortexa-llc/ai-pack/internal/mcp"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/streaming"
	"github.com/cortexa-llc/ai-pack/internal/tools"
)

func toolParamPreview(toolName string, input map[string]interface{}) string {
	const maxLen = 120
	truncate := func(s string) string {
		s = strings.ReplaceAll(s, "\n", " ")
		if len(s) > maxLen {
			return s[:maxLen] + "…"
		}
		return s
	}
	switch strings.ToLower(toolName) {
	case "bash":
		if cmd, ok := input["command"].(string); ok {
			return truncate(cmd)
		}
	case "read":
		if p, ok := input["file_path"].(string); ok {
			return truncate(p)
		}
	case "write":
		if p, ok := input["file_path"].(string); ok {
			return truncate(p)
		}
	case "edit", "multiedit":
		if p, ok := input["file_path"].(string); ok {
			return truncate(p)
		}
	case "grep":
		if pat, ok := input["pattern"].(string); ok {
			if path, ok2 := input["path"].(string); ok2 {
				return truncate(pat + " @ " + path)
			}
			return truncate(pat)
		}
	case "glob":
		if pat, ok := input["pattern"].(string); ok {
			return truncate(pat)
		}
	case "taskcomplete":
		if s, ok := input["summary"].(string); ok {
			return truncate(s)
		}
	}
	return ""
}

// extractContractSections pulls the meaningful sections out of a 00-contract.md file,
// skipping empty boilerplate placeholders ([Requirement X], [Assumption X], etc.).

func (s *AgentServer) executeTool(ctx context.Context, toolName string, toolInput map[string]interface{}, workingDir string, projectRoot string) (string, error) {
	// Handle search_knowledge_in_project: routes to another project's KG.
	if toolName == "search_knowledge_in_project" {
		return s.executeSearchKnowledgeInProject(ctx, toolInput)
	}

	if s.mcpManager != nil {
		// Determine if this is an MCP tool by checking project client first, then named clients.
		isProjectTool := false
		if projectRoot != "" {
			for _, t := range s.mcpManager.GetProjectTools(projectRoot) {
				if t.Name == toolName {
					isProjectTool = true
					break
				}
			}
		}

		isNamedTool := false
		if !isProjectTool {
			for _, serverTools := range s.mcpManager.GetAllTools() {
				for _, t := range serverTools {
					if t.Name == toolName {
						isNamedTool = true
						break
					}
				}
				if isNamedTool {
					break
				}
			}
		}

		if isProjectTool || isNamedTool {
			// Route via CallToolForProject: tries project client first, falls back to named clients.
			result, err := s.mcpManager.CallToolForProject(ctx, projectRoot, toolName, toolInput)
			if err != nil {
				return "", fmt.Errorf("MCP tool error: %w", err)
			}
			var resultText strings.Builder
			for _, block := range result.Content {
				if block.Type == "text" {
					resultText.WriteString(block.Text)
				}
			}
			return resultText.String(), nil
		}
	}

	// Not an MCP tool, execute as native tool
	return tools.ExecuteTool(toolName, toolInput, workingDir, s.claudeSettings)
}

// executeSearchKnowledgeInProject handles the search_knowledge_in_project synthetic tool.
// It ensures the target project's KG server is running, then calls search_nodes on it.
func (s *AgentServer) executeSearchKnowledgeInProject(ctx context.Context, toolInput map[string]interface{}) (string, error) {
	projectPath, _ := toolInput["project_path"].(string)
	query, _ := toolInput["query"].(string)

	if projectPath == "" {
		return "", fmt.Errorf("search_knowledge_in_project: project_path is required")
	}
	if query == "" {
		return "", fmt.Errorf("search_knowledge_in_project: query is required")
	}

	// Ensure the KG server is running for the target project.
	s.ensureKGForProject(projectPath)

	if s.mcpManager == nil {
		return "", fmt.Errorf("search_knowledge_in_project: MCP manager not available")
	}

	result, err := s.mcpManager.CallToolForProject(ctx, projectPath, "search_nodes", map[string]interface{}{
		"query": query,
	})
	if err != nil {
		return "", fmt.Errorf("search_knowledge_in_project: %w", err)
	}

	var resultText strings.Builder
	for _, block := range result.Content {
		if block.Type == "text" {
			resultText.WriteString(block.Text)
		}
	}
	monitoring.Logger.Info("search_knowledge_in_project",
		"project_path", projectPath,
		"query", query,
		"result_len", resultText.Len(),
	)
	return resultText.String(), nil
}

// cleanSchemaProperties recursively cleans schema properties to be Anthropic-compatible
// Removes: $schema, additionalProperties, and other non-standard fields
// Preserves: type, description, properties, items, enum, etc.
func cleanSchemaProperties(properties map[string]interface{}) map[string]interface{} {
	cleaned := make(map[string]interface{})

	for key, value := range properties {
		// Skip fields that Anthropic doesn't support in nested schemas
		if key == "$schema" || key == "additionalProperties" {
			continue
		}

		// Handle nested objects recursively
		if valueMap, ok := value.(map[string]interface{}); ok {
			cleaned[key] = cleanSchemaProperties(valueMap)
		} else {
			cleaned[key] = value
		}
	}

	return cleaned
}

// buildMCPStreamTool converts an mcp.Tool to a streaming.Tool, cleaning the schema.
func buildMCPStreamTool(tool mcp.Tool, serverName string) streaming.Tool {
	var properties map[string]interface{}
	var required []string

	if props, ok := tool.InputSchema["properties"].(map[string]interface{}); ok {
		properties = props
		if req, ok := tool.InputSchema["required"].([]interface{}); ok {
			for _, r := range req {
				if rStr, ok := r.(string); ok {
					required = append(required, rStr)
				}
			}
		}
	} else {
		properties = make(map[string]interface{})
		for key, value := range tool.InputSchema {
			if key != "$schema" && key != "type" && key != "required" && key != "additionalProperties" {
				properties[key] = value
			}
		}
		if req, ok := tool.InputSchema["required"].([]interface{}); ok {
			for _, r := range req {
				if rStr, ok := r.(string); ok {
					required = append(required, rStr)
				}
			}
		}
	}

	cleanedProperties := cleanSchemaProperties(properties)
	schema := map[string]interface{}{
		"type":       "object",
		"properties": cleanedProperties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}

	monitoring.Logger.Debug("mcp_tool_registered", "server", serverName, "tool", tool.Name)
	return streaming.Tool{
		Name:        tool.Name,
		Description: tool.Description,
		InputSchema: schema,
	}
}

// getAllTools returns all available tools in provider-agnostic format.
// Includes native tools, named MCP server tools, and project-specific KG tools.
// Project tools (KG) are included once, deduplicated against named-client tools.
func (s *AgentServer) getAllTools(projectRoot string) []streaming.Tool {
	// Start with native tools (already in streaming.Tool format)
	toolList := tools.DefineTools()

	// Track tool names already added to prevent duplicates
	seen := make(map[string]bool)
	for _, t := range toolList {
		seen[t.Name] = true
	}

	if s.mcpManager != nil {
		// Add project-specific KG tools first (highest priority for dedup)
		if projectRoot != "" {
			for _, tool := range s.mcpManager.GetProjectTools(projectRoot) {
				if !seen[tool.Name] {
					toolList = append(toolList, buildMCPStreamTool(tool, "kg:"+projectRoot))
					seen[tool.Name] = true
				}
			}
		}

		// Add named client tools
		mcpTools := s.mcpManager.GetAllTools()

		for serverName, serverTools := range mcpTools {
			for _, tool := range serverTools {
				if !seen[tool.Name] {
					toolList = append(toolList, buildMCPStreamTool(tool, serverName))
					seen[tool.Name] = true
				}
			}
		}
	}

	// Debug log all tools being returned
	if len(toolList) > 0 {
		if tools0JSON, err := json.MarshalIndent(toolList[0].InputSchema, "", "  "); err == nil {
			monitoring.Logger.Debug("tools_array_first_tool",
				"name", toolList[0].Name,
				"input_schema", string(tools0JSON))
		}
	}
	monitoring.Logger.Debug("tools_array_count", "total", len(toolList))

	return toolList
}
