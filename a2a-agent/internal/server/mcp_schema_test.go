package server

import (
	"encoding/json"
	"testing"
)

// TestCleanSchemaProperties validates that MCP tool schemas are cleaned correctly
func TestCleanSchemaProperties(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
		invalid  []string // Fields that should be removed
	}{
		{
			name: "remove $schema and additionalProperties at root",
			input: map[string]interface{}{
				"$schema":              "http://json-schema.org/draft-07/schema#",
				"additionalProperties": false,
				"type":                 "object",
				"description":          "Test schema",
			},
			expected: map[string]interface{}{
				"type":        "object",
				"description": "Test schema",
			},
			invalid: []string{"$schema", "additionalProperties"},
		},
		{
			name: "remove $schema in nested properties",
			input: map[string]interface{}{
				"entities": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"$schema":              "http://json-schema.org/draft-07/schema#",
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Entity name",
							},
						},
					},
				},
			},
			expected: map[string]interface{}{
				"entities": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Entity name",
							},
						},
					},
				},
			},
			invalid: []string{"$schema", "additionalProperties"},
		},
		{
			name: "real MCP memory tool schema",
			input: map[string]interface{}{
				"entities": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "The name of the entity",
							},
							"entityType": map[string]interface{}{
								"type":        "string",
								"description": "The type of the entity",
							},
							"observations": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "string",
								},
								"description": "An array of observation contents associated with the entity",
							},
						},
						"required": []interface{}{"name", "entityType", "observations"},
					},
					"description": "An array of entities to create",
				},
			},
			expected: map[string]interface{}{
				"entities": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "The name of the entity",
							},
							"entityType": map[string]interface{}{
								"type":        "string",
								"description": "The type of the entity",
							},
							"observations": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "string",
								},
								"description": "An array of observation contents associated with the entity",
							},
						},
						"required": []interface{}{"name", "entityType", "observations"},
					},
					"description": "An array of entities to create",
				},
			},
			invalid: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanSchemaProperties(tt.input)

			// Verify invalid fields are removed
			for _, invalidField := range tt.invalid {
				if containsField(result, invalidField) {
					t.Errorf("Field %s should have been removed but is still present", invalidField)
				}
			}

			// Verify expected structure matches
			resultJSON, _ := json.MarshalIndent(result, "", "  ")
			expectedJSON, _ := json.MarshalIndent(tt.expected, "", "  ")

			if string(resultJSON) != string(expectedJSON) {
				t.Errorf("Schema mismatch:\nExpected:\n%s\n\nGot:\n%s", string(expectedJSON), string(resultJSON))
			}
		})
	}
}

// containsField recursively checks if a field name exists in the schema
func containsField(schema map[string]interface{}, fieldName string) bool {
	for key, value := range schema {
		if key == fieldName {
			return true
		}
		if nested, ok := value.(map[string]interface{}); ok {
			if containsField(nested, fieldName) {
				return true
			}
		}
	}
	return false
}

// TestMCPToolSchemaConversion validates complete MCP tool to Anthropic conversion
func TestMCPToolSchemaConversion(t *testing.T) {
	// Simulate an MCP tool InputSchema
	mcpInputSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"entities": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$schema": "http://json-schema.org/draft-07/schema#",
					"type":    "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
					},
					"additionalProperties": false,
				},
			},
		},
		"required": []interface{}{"entities"},
	}

	// Extract and clean properties
	properties, _ := mcpInputSchema["properties"].(map[string]interface{})
	cleanedProperties := cleanSchemaProperties(properties)

	// Verify no invalid fields remain
	invalidFields := []string{"$schema", "additionalProperties"}
	for _, field := range invalidFields {
		if containsField(cleanedProperties, field) {
			t.Errorf("Invalid field %s found in cleaned schema", field)
		}
	}

	// Verify required array extraction
	required := []string{}
	if req, ok := mcpInputSchema["required"].([]interface{}); ok {
		for _, r := range req {
			if rStr, ok := r.(string); ok {
				required = append(required, rStr)
			}
		}
	}

	if len(required) != 1 || required[0] != "entities" {
		t.Errorf("Expected required=['entities'], got %v", required)
	}

	// Verify the schema structure is valid JSON Schema draft 2020-12
	// (simplified check - ensures it can be marshaled and has correct structure)
	if _, err := json.Marshal(cleanedProperties); err != nil {
		t.Errorf("Cleaned schema is not valid JSON: %v", err)
	}
}
