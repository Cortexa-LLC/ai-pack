package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicStreamAdapter adapts Anthropic SDK streaming to our StreamProvider interface
type AnthropicStreamAdapter struct {
	stream  interface{} // Anthropic streaming type (inferred by Go)
	current StreamEvent
	message *CompletedMessage
	err     error
	done    bool
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
	// Convert generic messages to Anthropic format
	messages := make([]anthropic.MessageParam, 0, len(req.Messages))
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		} else if msg.Role == "assistant" {
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
			// Convert input schema to ToolInputSchemaParam
			schema := anthropic.ToolInputSchemaParam{
				Type:       "object",
				Properties: tool.InputSchema,
			}
			if required, ok := tool.InputSchema["required"].([]string); ok {
				schema.Required = required
			}

			tools = append(tools, anthropic.ToolUnionParamOfTool(
				schema,
				tool.Name,
			))
		}
		params.Tools = tools
	}

	// Create streaming request
	stream := f.client.Messages.NewStreaming(ctx, params)

	return &AnthropicStreamAdapter{
		stream:  stream,
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

// convertEvent converts Anthropic events to generic StreamEvent
func (a *AnthropicStreamAdapter) convertEvent(event interface{}) StreamEvent {
	// Marshal and unmarshal to extract fields
	eventJSON, _ := json.Marshal(event)
	var eventData map[string]interface{}
	json.Unmarshal(eventJSON, &eventData)

	eventType := ""
	if t, ok := eventData["type"].(string); ok {
		eventType = t
	}

	genericEvent := StreamEvent{
		Type: eventType,
	}

	// Handle different event types
	switch eventType {
	case "content_block_start":
		// Extract tool use blocks
		if contentBlock, ok := eventData["content_block"].(map[string]interface{}); ok {
			if blockType, ok := contentBlock["type"].(string); ok && blockType == "tool_use" {
				toolUse := &ToolUse{
					Input: make(map[string]interface{}),
				}
				if id, ok := contentBlock["id"].(string); ok {
					toolUse.ID = id
				}
				if name, ok := contentBlock["name"].(string); ok {
					toolUse.Name = name
				}
				if input, ok := contentBlock["input"].(map[string]interface{}); ok {
					toolUse.Input = input
				}
				genericEvent.ToolUse = toolUse
			}
		}

	case "content_block_delta":
		// Extract text delta
		eventJSON, _ := json.Marshal(event)
		var eventData map[string]interface{}
		json.Unmarshal(eventJSON, &eventData)

		if delta, ok := eventData["delta"].(map[string]interface{}); ok {
			if text, ok := delta["text"].(string); ok {
				genericEvent.Delta = &DeltaContent{
					Text: text,
					Type: "text",
				}
			}
		}

	case "message_start":
		// Extract message metadata
		eventJSON, _ := json.Marshal(event)
		var eventData map[string]interface{}
		json.Unmarshal(eventJSON, &eventData)

		if msgData, ok := eventData["message"].(map[string]interface{}); ok {
			if id, ok := msgData["id"].(string); ok {
				a.message.ID = id
			}
			if role, ok := msgData["role"].(string); ok {
				a.message.Role = role
			}
			// Extract input tokens from usage
			if usage, ok := msgData["usage"].(map[string]interface{}); ok {
				if inputTokens, ok := usage["input_tokens"].(float64); ok {
					a.message.InputTokens = int(inputTokens)
				}
			}
		}

	case "message_delta":
		// Extract token usage
		eventJSON, _ := json.Marshal(event)
		var eventData map[string]interface{}
		json.Unmarshal(eventJSON, &eventData)

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
