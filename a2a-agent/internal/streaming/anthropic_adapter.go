package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
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

// AnthropicStreamAdapter adapts Anthropic SDK streaming to our StreamProvider interface.
// It uses typed SDK calls instead of reflection or JSON round-trips to dispatch events.
type AnthropicStreamAdapter struct {
	stream          *ssestream.Stream[anthropic.MessageStreamEventUnion]
	current         StreamEvent
	message         *CompletedMessage
	err             error
	done            bool
	model           string
	provider        string
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
				inputJSON, err := json.Marshal(tu.Input)
				if err != nil {
					return nil, fmt.Errorf("marshal tool input for %q: %w", tu.Name, err)
				}
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
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = f.maxTokens
	}
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(req.Model),
		MaxTokens: int64(maxTokens),
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
			// tool.InputSchema is a complete schema with "type", "properties", "required".
			// Extract just the properties field for ToolInputSchemaParam.
			properties, ok := tool.InputSchema["properties"].(map[string]interface{})
			if !ok {
				properties = make(map[string]interface{})
			}

			schema := anthropic.ToolInputSchemaParam{
				Type:       "object",
				Properties: properties,
			}

			// Extract required array — handle both []string and []interface{} types
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

			toolParam := anthropic.ToolParam{
				Name:        tool.Name,
				InputSchema: schema,
				Type:        anthropic.ToolTypeCustom,
			}
			if tool.Description != "" {
				toolParam.Description = param.NewOpt(tool.Description)
			}

			tools = append(tools, anthropic.ToolUnionParam{OfTool: &toolParam})
		}
		params.Tools = tools
	}

	stream := f.client.Messages.NewStreaming(ctx, params)

	return &AnthropicStreamAdapter{
		stream:   stream,
		model:    req.Model,
		provider: ProviderAnthropic,
		message: &CompletedMessage{
			Provider: ProviderAnthropic,
			Model:    req.Model,
		},
		pendingToolCalls: make(map[int]*pendingToolUse),
	}, nil
}

// Next advances to the next event using typed SDK calls (no reflection).
func (a *AnthropicStreamAdapter) Next() bool {
	if a.done {
		return false
	}

	for a.stream.Next() {
		ev := a.convertEvent(a.stream.Current())

		// Accumulate text content into the completed message.
		if ev.Delta != nil && ev.Delta.Text != "" && ev.Delta.Type == "text" {
			a.message.Content += ev.Delta.Text
		}

		// Skip internal bookkeeping events that have no consumer-facing payload,
		// except ones that carry meaningful data (tool_use, message_start/delta/stop).
		// All event types are forwarded; callers decide what to act on.
		a.current = ev
		return true
	}

	// Stream exhausted — capture any terminal error.
	if err := a.stream.Err(); err != nil {
		a.err = err
	}
	a.done = true
	return false
}

// Current returns the current event.
func (a *AnthropicStreamAdapter) Current() StreamEvent {
	return a.current
}

// Err returns any streaming error.
func (a *AnthropicStreamAdapter) Err() error {
	return a.err
}

// Close releases resources held by the underlying stream.
func (a *AnthropicStreamAdapter) Close() error {
	return a.stream.Close()
}

// GetMessage returns the accumulated message (valid after the stream is consumed).
func (a *AnthropicStreamAdapter) GetMessage() *CompletedMessage {
	return a.message
}

// GetModel returns the model being used.
func (a *AnthropicStreamAdapter) GetModel() string {
	return a.model
}

// GetProvider returns the provider name.
func (a *AnthropicStreamAdapter) GetProvider() string {
	return a.provider
}

// convertEvent maps a typed Anthropic SDK event onto the generic StreamEvent.
// All dispatch is done via AsAny() — no reflection, no JSON round-trip.
func (a *AnthropicStreamAdapter) convertEvent(raw anthropic.MessageStreamEventUnion) StreamEvent {
	switch ev := raw.AsAny().(type) {

	// ── message_start ──────────────────────────────────────────────────────
	case anthropic.MessageStartEvent:
		if a.message != nil {
			a.message.ID = ev.Message.ID
			a.message.Role = string(ev.Message.Role)
			a.message.InputTokens = int(ev.Message.Usage.InputTokens)
		}
		return StreamEvent{Type: "message_start"}

	// ── content_block_start ────────────────────────────────────────────────
	case anthropic.ContentBlockStartEvent:
		idx := int(ev.Index)
		switch cb := ev.ContentBlock.AsAny().(type) {
		case anthropic.ToolUseBlock:
			// Register pending tool call so we can accumulate input_json_delta events.
			a.pendingToolCalls[idx] = &pendingToolUse{
				id:   cb.ID,
				name: cb.Name,
			}
		}
		return StreamEvent{Type: "content_block_start"}

	// ── content_block_delta ────────────────────────────────────────────────
	case anthropic.ContentBlockDeltaEvent:
		idx := int(ev.Index)
		switch delta := ev.Delta.AsAny().(type) {
		case anthropic.TextDelta:
			return StreamEvent{
				Type:  "content_block_delta",
				Delta: &DeltaContent{Type: "text", Text: delta.Text},
			}
		case anthropic.InputJSONDelta:
			if pending, ok := a.pendingToolCalls[idx]; ok {
				pending.jsonBuffer.WriteString(delta.PartialJSON)
			}
			return StreamEvent{Type: "content_block_delta"}
		}
		return StreamEvent{Type: "content_block_delta"}

	// ── content_block_stop ─────────────────────────────────────────────────
	case anthropic.ContentBlockStopEvent:
		idx := int(ev.Index)
		if pending, ok := a.pendingToolCalls[idx]; ok {
			delete(a.pendingToolCalls, idx)
			// Guard: skip emission if tool call metadata is incomplete.
			if pending.id == "" || pending.name == "" {
				return StreamEvent{Type: "content_block_stop"}
			}
			input := make(map[string]interface{})
			if jsonStr := pending.jsonBuffer.String(); jsonStr != "" {
				if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
					input = map[string]interface{}{"_args": jsonStr}
				}
			}
			return StreamEvent{
				Type: "content_block_stop",
				ToolUse: &ToolUse{
					ID:    pending.id,
					Name:  pending.name,
					Input: input,
				},
			}
		}
		return StreamEvent{Type: "content_block_stop"}

	// ── message_delta ──────────────────────────────────────────────────────
	case anthropic.MessageDeltaEvent:
		if a.message != nil {
			if ev.Delta.StopReason != "" {
				a.message.StopReason = string(ev.Delta.StopReason)
			}
			a.message.OutputTokens = int(ev.Usage.OutputTokens)
		}
		return StreamEvent{Type: "message_delta"}

	// ── message_stop ───────────────────────────────────────────────────────
	case anthropic.MessageStopEvent:
		return StreamEvent{
			Type:    "message_stop",
			Message: a.message,
		}
	}

	// Unknown / future event type — return zero value; caller will skip it.
	return StreamEvent{}
}
