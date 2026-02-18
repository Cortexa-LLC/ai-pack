package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
)

// pendingToolUse tracks an in-progress streaming Anthropic tool call.
// Anthropic sends tool input as incremental JSON deltas (input_json_delta)
// across multiple content_block_delta events; we accumulate them here and
// emit the completed ToolUse on content_block_stop.
type pendingToolUse struct {
	id         string
	name       string
	jsonBuffer strings.Builder
}

// AnthropicStreamAdapter adapts Anthropic SDK streaming to our StreamProvider interface
type AnthropicStreamAdapter struct {
	stream           interface{} // Anthropic streaming type (inferred by Go)
	current          StreamEvent
	message          *CompletedMessage
	err              error
	done             bool
	model            string // Model being used
	provider         string // Provider name
	pendingToolCalls map[int]*pendingToolUse
}

// AnthropicFactory creates Anthropic stream providers
type AnthropicFactory struct {
	client    anthropic.Client
	apiKey    string
	maxTokens int
}

// NewAnthropicFactory creates a new Anthropic provider factory
func NewAnthropicFactory(apiKey string, maxTokens int, opts ...option.RequestOption) *AnthropicFactory {
	client := anthropic.NewClient(opts...)
	return &AnthropicFactory{
		client:    client,
		apiKey:    apiKey,
		maxTokens: maxTokens,
	}
}

// GetProviderName returns the provider name
func (f *AnthropicFactory) GetProviderName() string {
	return ProviderAnthropic
}

// SupportsModel checks if this is a Claude model
func (f *AnthropicFactory) SupportsModel(model string) bool {
	return strings.Contains(strings.ToLower(model), "claude")
}

// CreateStream creates an Anthropic streaming provider
func (f *AnthropicFactory) CreateStream(ctx context.Context, req StreamRequest) (StreamProvider, error) {
	// Convert generic messages to Anthropic format, preserving tool call history.
	messages := make([]anthropic.MessageParam, 0, len(req.Messages))
	for _, msg := range req.Messages {
		switch {
		case len(msg.ToolResults) > 0:
			// User message carrying tool results
			var blocks []anthropic.ContentBlockParamUnion
			for _, tr := range msg.ToolResults {
				blocks = append(blocks, anthropic.NewToolResultBlock(tr.ToolUseID, tr.Content, tr.IsError))
			}
			messages = append(messages, anthropic.NewUserMessage(blocks...))

		case len(msg.ToolUses) > 0:
			// Assistant message with tool calls (and optional preceding text)
			var blocks []anthropic.ContentBlockParamUnion
			if msg.Content != "" {
				blocks = append(blocks, anthropic.NewTextBlock(msg.Content))
			}
			for _, tu := range msg.ToolUses {
				inputJSON, _ := json.Marshal(tu.Input)
				blocks = append(blocks, anthropic.NewToolUseBlock(tu.ID, json.RawMessage(inputJSON), tu.Name))
			}
			messages = append(messages, anthropic.NewAssistantMessage(blocks...))

		case msg.Role == "user":
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))

		case msg.Role == "assistant":
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
		}
	}

	// Build request params
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(req.Model),
		MaxTokens: int64(req.MaxTokens),
		Messages:  messages,
	}

	// Add system prompt if provided
	if req.SystemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{
				Text: req.SystemPrompt,
				Type: "text",
			},
		}
	}

	// Convert tools if provided
	if len(req.Tools) > 0 {
		tools := make([]anthropic.ToolUnionParam, 0, len(req.Tools))
		for _, tool := range req.Tools {
			// Extract properties from the schema map
			// tool.InputSchema is already a complete schema with "type", "properties", "required"
			// We need to extract just the properties field
			properties, ok := tool.InputSchema["properties"].(map[string]interface{})
			if !ok {
				properties = make(map[string]interface{})
			}

			// Convert input schema to ToolInputSchemaParam
			schema := anthropic.ToolInputSchemaParam{
				Type:       "object",
				Properties: properties,
			}

			// Extract required array - handle both []string and []interface{} types
			if required, ok := tool.InputSchema["required"].([]string); ok {
				schema.Required = required
			} else if requiredIface, ok := tool.InputSchema["required"].([]interface{}); ok {
				reqStrings := make([]string, 0, len(requiredIface))
				for _, r := range requiredIface {
					if rStr, ok := r.(string); ok {
						reqStrings = append(reqStrings, rStr)
					}
				}
				schema.Required = reqStrings
			}

			// Create ToolParam manually to ensure Type field is set correctly
			toolParam := anthropic.ToolParam{
				Name:        tool.Name,
				InputSchema: schema,
				Type:        anthropic.ToolTypeCustom,
			}
			// Add description if available
			if tool.Description != "" {
				toolParam.Description = param.NewOpt(tool.Description)
			}

			// Create ToolUnionParam from ToolParam
			toolUnion := anthropic.ToolUnionParam{
				OfTool: &toolParam,
			}

			tools = append(tools, toolUnion)
		}
		params.Tools = tools
	}

	// Create streaming request
	stream := f.client.Messages.NewStreaming(ctx, params)

	return &AnthropicStreamAdapter{
		stream:   stream,
		model:    req.Model,
		provider: ProviderAnthropic,
		message: &CompletedMessage{
			Provider: ProviderAnthropic,
			Model:    req.Model,
		},
	}, nil
}

// Next advances to the next event
func (a *AnthropicStreamAdapter) Next() bool {
	if a.done {
		return false
	}

	// Use reflection to call methods on the stream
	streamVal := reflect.ValueOf(a.stream)
	if !streamVal.IsValid() {
		a.err = fmt.Errorf("invalid stream")
		a.done = true
		return false
	}

	// Call Next() method
	nextMethod := streamVal.MethodByName("Next")
	if !nextMethod.IsValid() {
		a.err = fmt.Errorf("stream missing Next method")
		a.done = true
		return false
	}

	nextResult := nextMethod.Call(nil)
	if len(nextResult) == 0 || !nextResult[0].Bool() {
		// Stream ended, get error
		errMethod := streamVal.MethodByName("Err")
		if errMethod.IsValid() {
			errResult := errMethod.Call(nil)
			if len(errResult) > 0 && !errResult[0].IsNil() {
				a.err = errResult[0].Interface().(error)
			}
		}
		a.done = true
		return false
	}

	// Call Current() method to get event
	currentMethod := streamVal.MethodByName("Current")
	if !currentMethod.IsValid() {
		a.err = fmt.Errorf("stream missing Current method")
		a.done = true
		return false
	}

	currentResult := currentMethod.Call(nil)
	if len(currentResult) == 0 {
		a.err = fmt.Errorf("Current() returned no value")
		a.done = true
		return false
	}

	// Convert Anthropic event to generic event
	event := currentResult[0].Interface()
	a.current = a.convertEvent(event)

	// Accumulate message content
	if a.current.Delta != nil && a.current.Delta.Text != "" {
		a.message.Content += a.current.Delta.Text
	}

	return true
}

// Current returns the current event
func (a *AnthropicStreamAdapter) Current() StreamEvent {
	return a.current
}

// Err returns any streaming error
func (a *AnthropicStreamAdapter) Err() error {
	return a.err
}

// Close releases resources
func (a *AnthropicStreamAdapter) Close() error {
	// Anthropic stream doesn't require explicit close
	return nil
}

// GetMessage returns the accumulated message
func (a *AnthropicStreamAdapter) GetMessage() *CompletedMessage {
	return a.message
}

// GetModel returns the model being used
func (a *AnthropicStreamAdapter) GetModel() string {
	return a.model
}

// GetProvider returns the provider name
func (a *AnthropicStreamAdapter) GetProvider() string {
	return a.provider
}

// convertEvent converts Anthropic events to generic StreamEvent
func (a *AnthropicStreamAdapter) convertEvent(event interface{}) StreamEvent {
	// Marshal and unmarshal to extract fields
	eventJSON, _ := json.Marshal(event)
	var eventData map[string]interface{}
	json.Unmarshal(eventJSON, &eventData) //nolint:errcheck

	eventType := ""
	if t, ok := eventData["type"].(string); ok {
		eventType = t
	}

	genericEvent := StreamEvent{
		Type: eventType,
	}

	// Get block index if present (used to correlate tool call events)
	blockIndex := 0
	if idxFloat, ok := eventData["index"].(float64); ok {
		blockIndex = int(idxFloat)
	}

	// Handle different event types
	switch eventType {
	case "content_block_start":
		if contentBlock, ok := eventData["content_block"].(map[string]interface{}); ok {
			if blockType, ok := contentBlock["type"].(string); ok && blockType == "tool_use" {
				// Store pending tool call — input comes in subsequent input_json_delta events
				if a.pendingToolCalls == nil {
					a.pendingToolCalls = make(map[int]*pendingToolUse)
				}
				pending := &pendingToolUse{}
				if id, ok := contentBlock["id"].(string); ok {
					pending.id = id
				}
				if name, ok := contentBlock["name"].(string); ok {
					pending.name = name
				}
				a.pendingToolCalls[blockIndex] = pending
				// Don't emit ToolUse yet; wait for content_block_stop
			}
		}

	case "content_block_delta":
		if delta, ok := eventData["delta"].(map[string]interface{}); ok {
			deltaType, _ := delta["type"].(string)
			switch deltaType {
			case "text_delta":
				if text, ok := delta["text"].(string); ok {
					genericEvent.Delta = &DeltaContent{
						Text: text,
						Type: "text",
					}
				}
			case "input_json_delta":
				// Accumulate JSON fragment for the pending tool call
				if partialJSON, ok := delta["partial_json"].(string); ok {
					if a.pendingToolCalls != nil {
						if pending, exists := a.pendingToolCalls[blockIndex]; exists {
							pending.jsonBuffer.WriteString(partialJSON)
						}
					}
				}
			}
		}

	case "content_block_stop":
		// If this block was a tool call, emit the completed ToolUse now
		if a.pendingToolCalls != nil {
			if pending, exists := a.pendingToolCalls[blockIndex]; exists {
				toolUse := &ToolUse{
					ID:    pending.id,
					Name:  pending.name,
					Input: make(map[string]interface{}),
				}
				if jsonStr := pending.jsonBuffer.String(); jsonStr != "" {
					json.Unmarshal([]byte(jsonStr), &toolUse.Input) //nolint:errcheck
				}
				genericEvent.ToolUse = toolUse
				delete(a.pendingToolCalls, blockIndex)
			}
		}

	case "message_start":
		if msgData, ok := eventData["message"].(map[string]interface{}); ok {
			if id, ok := msgData["id"].(string); ok {
				a.message.ID = id
			}
			if role, ok := msgData["role"].(string); ok {
				a.message.Role = role
			}
			if usage, ok := msgData["usage"].(map[string]interface{}); ok {
				if inputTokens, ok := usage["input_tokens"].(float64); ok {
					a.message.InputTokens = int(inputTokens)
				}
			}
		}

	case "message_delta":
		if usage, ok := eventData["usage"].(map[string]interface{}); ok {
			if outputTokens, ok := usage["output_tokens"].(float64); ok {
				a.message.OutputTokens = int(outputTokens)
			}
		}
		if delta, ok := eventData["delta"].(map[string]interface{}); ok {
			if stopReason, ok := delta["stop_reason"].(string); ok {
				a.message.StopReason = stopReason
			}
		}
	}

	return genericEvent
}
