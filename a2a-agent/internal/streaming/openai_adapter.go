package streaming

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// OpenAIStreamAdapter adapts OpenAI SDK streaming to our StreamProvider interface
type OpenAIStreamAdapter struct {
	stream   *openai.ChatCompletionStream
	current  StreamEvent
	message  *CompletedMessage
	err      error
	done     bool
	model    string // Model being used
	provider string // Provider name
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

// SupportsModel checks if this is a GPT model
func (f *OpenAIFactory) SupportsModel(model string) bool {
	return strings.HasPrefix(strings.ToLower(model), "gpt-") ||
		strings.Contains(strings.ToLower(model), "o1-") ||
		strings.Contains(strings.ToLower(model), "o3-")
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

	// Add conversation messages
	for _, msg := range req.Messages {
		role := openai.ChatMessageRoleUser
		if msg.Role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Build request
	request := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Stream:      true,
		StreamOptions: &openai.StreamOptions{
			IncludeUsage: true,
		},
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

	response, err := o.stream.Recv()
	if err != nil {
		if errors.Is(err, io.EOF) {
			o.done = true
			return false
		}
		o.err = err
		o.done = true
		return false
	}

	// Convert OpenAI response to generic event
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

// convertResponse converts OpenAI response to generic StreamEvent
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
