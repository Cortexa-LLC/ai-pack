package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// openAIMaxCompletionTokens is the maximum completion tokens supported by gpt-4o / gpt-4o-mini.
// Requests with a higher max_tokens value are rejected by the API with a 400 error.
const openAIMaxCompletionTokens = 16384

// usesMaxCompletionTokens returns true for newer OpenAI models that require the
// max_completion_tokens field instead of the deprecated max_tokens field.
func usesMaxCompletionTokens(model string) bool {
	lower := strings.ToLower(model)
	return strings.HasPrefix(lower, "o1") ||
		strings.HasPrefix(lower, "o3") ||
		strings.HasPrefix(lower, "o4") ||
		strings.Contains(lower, "gpt-5")
}

// isCodexModel returns true for models in the codex family (e.g. gpt-5.1-codex, gpt-5.2-codex).
func isCodexModel(model string) bool {
	lower := strings.ToLower(model)
	return strings.Contains(lower, "codex")
}

// openaiPendingToolCall accumulates the streamed arguments for a single tool call.
// OpenAI sends function arguments as partial strings across multiple stream chunks,
// keyed by the tool call's index in the delta.ToolCalls slice.
type openaiPendingToolCall struct {
	id          string
	name        string
	argsBuffer  strings.Builder
}

// OpenAIStreamAdapter adapts OpenAI SDK streaming to our StreamProvider interface
type OpenAIStreamAdapter struct {
	stream           *openai.ChatCompletionStream
	current          StreamEvent
	message          *CompletedMessage
	err              error
	done             bool
	model            string // Model being used
	provider         string // Provider name
	pendingToolCalls map[int]*openaiPendingToolCall
	pendingEvents    []StreamEvent // tool-use events queued after EOF
}

// OpenAIFactory creates OpenAI stream providers
type OpenAIFactory struct {
	client    *openai.Client
	maxTokens int
}

// NewOpenAIFactory creates a new OpenAI provider factory
func NewOpenAIFactory(client *openai.Client, maxTokens int) *OpenAIFactory {
	return &OpenAIFactory{
		client:    client,
		maxTokens: maxTokens,
	}
}

// GetProviderName returns the provider name
func (f *OpenAIFactory) GetProviderName() string {
	return ProviderOpenAI
}

// SupportsModel checks if this is an OpenAI model
func (f *OpenAIFactory) SupportsModel(model string) bool {
	lower := strings.ToLower(model)
	return strings.HasPrefix(lower, "gpt-") ||
		strings.HasPrefix(lower, "o1") ||
		strings.HasPrefix(lower, "o3") ||
		strings.HasPrefix(lower, "o4")
}

// CreateStream creates an OpenAI streaming provider
func (f *OpenAIFactory) CreateStream(ctx context.Context, req StreamRequest) (StreamProvider, error) {
	// Convert generic messages to OpenAI format
	messages := make([]openai.ChatCompletionMessage, 0, len(req.Messages))

	// Add system prompt as first message if provided
	if req.SystemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.SystemPrompt,
		})
	}

	// Add conversation messages, preserving tool call history.
	for _, msg := range req.Messages {
		switch {
		case len(msg.ToolResults) > 0:
			// Each tool result is a separate "tool" role message in OpenAI format
			for _, tr := range msg.ToolResults {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    tr.Content,
					ToolCallID: tr.ToolUseID,
				})
			}

		case len(msg.ToolUses) > 0:
			// Assistant message with tool calls
			toolCalls := make([]openai.ToolCall, 0, len(msg.ToolUses))
			for _, tu := range msg.ToolUses {
				argsJSON, _ := json.Marshal(tu.Input)
				toolCalls = append(toolCalls, openai.ToolCall{
					ID:   tu.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      tu.Name,
						Arguments: string(argsJSON),
					},
				})
			}
			messages = append(messages, openai.ChatCompletionMessage{
				Role:      openai.ChatMessageRoleAssistant,
				Content:   msg.Content,
				ToolCalls: toolCalls,
			})

		case msg.Role == "user":
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: msg.Content,
			})

		default: // "assistant"
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: msg.Content,
			})
		}
	}

	// Cap tokens to avoid a 400 error from the API
	maxTokens := req.MaxTokens
	if maxTokens > openAIMaxCompletionTokens {
		maxTokens = openAIMaxCompletionTokens
	}

	// Build request — newer models (o1/o3/gpt-5.x) require MaxCompletionTokens;
	// legacy models (gpt-4o-mini, gpt-3.5, etc.) use the deprecated MaxTokens.
	request := openai.ChatCompletionRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   true,
		StreamOptions: &openai.StreamOptions{
			IncludeUsage: true,
		},
	}
	if usesMaxCompletionTokens(req.Model) {
		request.MaxCompletionTokens = maxTokens
	} else {
		request.MaxTokens = maxTokens
	}

	// Convert tools if provided
	if len(req.Tools) > 0 {
		tools := make([]openai.Tool, 0, len(req.Tools))
		for _, tool := range req.Tools {
			tools = append(tools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.InputSchema,
				},
			})
		}
		request.Tools = tools
	}

	// Create streaming request
	stream, err := f.client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		return nil, err
	}

	return &OpenAIStreamAdapter{
		stream:   stream,
		model:    req.Model,
		provider: ProviderOpenAI,
		message: &CompletedMessage{
			Provider: ProviderOpenAI,
			Model:    req.Model,
			Role:     "assistant",
		},
	}, nil
}

// Next advances to the next event
func (o *OpenAIStreamAdapter) Next() bool {
	if o.done {
		return false
	}

	// Drain any tool-use events queued after the stream ended
	if len(o.pendingEvents) > 0 {
		o.current = o.pendingEvents[0]
		o.pendingEvents = o.pendingEvents[1:]
		return true
	}

	response, err := o.stream.Recv()
	if err != nil {
		if errors.Is(err, io.EOF) {
			// Flush accumulated tool calls into pendingEvents before stopping
			o.flushPendingToolCalls()
			if len(o.pendingEvents) > 0 {
				o.current = o.pendingEvents[0]
				o.pendingEvents = o.pendingEvents[1:]
				return true
			}
			o.done = true
			return false
		}
		o.err = err
		o.done = true
		return false
	}

	// Accumulate tool call argument fragments by index
	if len(response.Choices) > 0 {
		for _, tc := range response.Choices[0].Delta.ToolCalls {
			if tc.Index == nil {
				continue
			}
			idx := *tc.Index
			if o.pendingToolCalls == nil {
				o.pendingToolCalls = make(map[int]*openaiPendingToolCall)
			}
			pending, exists := o.pendingToolCalls[idx]
			if !exists {
				pending = &openaiPendingToolCall{}
				o.pendingToolCalls[idx] = pending
			}
			if tc.ID != "" {
				pending.id = tc.ID
			}
			if tc.Function.Name != "" {
				pending.name = tc.Function.Name
			}
			if tc.Function.Arguments != "" {
				pending.argsBuffer.WriteString(tc.Function.Arguments)
			}
		}
	}

	// Convert OpenAI response to generic event (text only; tool calls emitted on flush)
	o.current = o.convertResponse(response)

	// Accumulate message content
	if o.current.Delta != nil && o.current.Delta.Text != "" {
		o.message.Content += o.current.Delta.Text
	}

	// Update token usage if available
	if response.Usage != nil {
		o.message.InputTokens = response.Usage.PromptTokens
		o.message.OutputTokens = response.Usage.CompletionTokens
	}

	// Check for completion
	if len(response.Choices) > 0 && response.Choices[0].FinishReason != "" {
		o.message.StopReason = string(response.Choices[0].FinishReason)
	}

	return true
}

// flushPendingToolCalls converts accumulated tool call data into StreamEvents.
// Called once after the stream ends so that each tool call has its complete arguments.
func (o *OpenAIStreamAdapter) flushPendingToolCalls() {
	if len(o.pendingToolCalls) == 0 {
		return
	}

	// Process tool calls in index order for deterministic output
	indices := make([]int, 0, len(o.pendingToolCalls))
	for idx := range o.pendingToolCalls {
		indices = append(indices, idx)
	}
	sort.Ints(indices)

	for _, idx := range indices {
		pending := o.pendingToolCalls[idx]
		toolUse := &ToolUse{
			ID:    pending.id,
			Name:  pending.name,
			Input: make(map[string]interface{}),
		}
		if argsStr := pending.argsBuffer.String(); argsStr != "" {
			json.Unmarshal([]byte(argsStr), &toolUse.Input) //nolint:errcheck
		}
		o.pendingEvents = append(o.pendingEvents, StreamEvent{
			Type:    "tool_use",
			ToolUse: toolUse,
		})
	}

	o.pendingToolCalls = nil
}

// Current returns the current event
func (o *OpenAIStreamAdapter) Current() StreamEvent {
	return o.current
}

// Err returns any streaming error
func (o *OpenAIStreamAdapter) Err() error {
	return o.err
}

// Close releases resources
func (o *OpenAIStreamAdapter) Close() error {
	o.stream.Close()
	return nil
}

// GetMessage returns the accumulated message
func (o *OpenAIStreamAdapter) GetMessage() *CompletedMessage {
	return o.message
}

// GetModel returns the model being used
func (o *OpenAIStreamAdapter) GetModel() string {
	return o.model
}

// GetProvider returns the provider name
func (o *OpenAIStreamAdapter) GetProvider() string {
	return o.provider
}

// convertResponse converts OpenAI response to generic StreamEvent (text content only).
// Tool call events are emitted separately via flushPendingToolCalls after EOF.
func (o *OpenAIStreamAdapter) convertResponse(response openai.ChatCompletionStreamResponse) StreamEvent {
	event := StreamEvent{
		Type: "content_block_delta", // Use Anthropic-compatible event type
	}

	// Extract delta content
	if len(response.Choices) > 0 {
		delta := response.Choices[0].Delta
		if delta.Content != "" {
			event.Delta = &DeltaContent{
				Text: delta.Content,
				Type: "text",
			}
		}
	}

	return event
}
