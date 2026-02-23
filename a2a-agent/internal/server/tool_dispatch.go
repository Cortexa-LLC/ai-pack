package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/tools"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/streaming"
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
	}
	return ""
}

// extractContractSections pulls the meaningful sections out of a 00-contract.md file,
// skipping empty boilerplate placeholders ([Requirement X], [Assumption X], etc.).

func (s *AgentServer) executeTool(ctx context.Context, toolName string, toolInput map[string]interface{}, workingDir string) (string, error) {
	// Check if this is an MCP tool
	if s.mcpManager != nil {
		mcpTools := s.mcpManager.GetAllTools()
		for _, serverTools := range mcpTools {
			for _, tool := range serverTools {
				if tool.Name == toolName {
					// This is an MCP tool
					result, err := s.mcpManager.CallTool(ctx, toolName, toolInput)
					if err != nil {
						return "", fmt.Errorf("MCP tool error: %w", err)
					}

					// Convert MCP result to string
					var resultText strings.Builder
					for _, block := range result.Content {
						if block.Type == "text" {
							resultText.WriteString(block.Text)
						}
					}

					return resultText.String(), nil
				}
			}
		}
	}

	// Not an MCP tool, execute as native tool
	return tools.ExecuteTool(toolName, toolInput, workingDir, s.claudeSettings)
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

// getAllTools returns all available tools in provider-agnostic format.
// Includes native tools and any tools registered via MCP servers.
func (s *AgentServer) getAllTools() []streaming.Tool {
	// Start with native tools (already in streaming.Tool format)
	toolList := tools.DefineTools()

	// Add MCP tools if manager is initialized
	if s.mcpManager != nil {
		mcpTools := s.mcpManager.GetAllTools()

		for serverName, serverTools := range mcpTools {
			for _, tool := range serverTools {
				// MCP InputSchema structure varies:
				// - Some have {"properties": {...}, "required": [...]}
				// - Others have properties directly at root: {"param1": {...}, "param2": {...}}

				var properties map[string]interface{}
				var required []string

				// Check if schema has explicit "properties" field (JSON Schema standard format)
				if props, ok := tool.InputSchema["properties"].(map[string]interface{}); ok {
					properties = props

					// Extract required array if present
					if req, ok := tool.InputSchema["required"].([]interface{}); ok {
						for _, r := range req {
							if rStr, ok := r.(string); ok {
								required = append(required, rStr)
							}
						}
					}
				} else {
					// No "properties" field - MCP schema has properties at root level
					// Treat entire InputSchema as properties, excluding meta fields
					properties = make(map[string]interface{})
					for key, value := range tool.InputSchema {
						// Skip meta fields - only keep actual parameter definitions
						if key != "$schema" && key != "type" && key != "required" && key != "additionalProperties" {
							properties[key] = value
						}
					}

					// Extract required array if present at root
					if req, ok := tool.InputSchema["required"].([]interface{}); ok {
						for _, r := range req {
							if rStr, ok := r.(string); ok {
								required = append(required, rStr)
							}
						}
					}
				}

				// Clean properties recursively - remove $schema, additionalProperties, etc.
				cleanedProperties := cleanSchemaProperties(properties)

				// Debug log the cleaned schema
				if schemaJSON, err := json.MarshalIndent(cleanedProperties, "", "  "); err == nil {
					monitoring.Logger.Debug("mcp_tool_schema_cleaned",
						"tool", tool.Name,
						"schema", string(schemaJSON))
				}

				// Build provider-agnostic streaming.Tool with a complete JSON Schema
				schema := map[string]interface{}{
					"type":       "object",
					"properties": cleanedProperties,
				}
				if len(required) > 0 {
					schema["required"] = required
				}

				streamTool := streaming.Tool{
					Name:        tool.Name,
					Description: tool.Description,
					InputSchema: schema,
				}

				// Debug log the final schema
				if finalSchemaJSON, err := json.MarshalIndent(schema, "", "  "); err == nil {
					monitoring.Logger.Debug("mcp_tool_final_schema",
						"tool", tool.Name,
						"final_schema", string(finalSchemaJSON))
				}

				toolList = append(toolList, streamTool)

				monitoring.Logger.Debug("mcp_tool_registered",
					"server", serverName,
					"tool", tool.Name)
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

