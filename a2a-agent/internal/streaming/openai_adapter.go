package streaming

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/packages/ssestream"
	"github.com/openai/openai-go/responses"
)

// OpenAIFactory creates OpenAI stream providers using the official SDK
// Supports chat completion and responses API for codex models
type OpenAIFactory struct {
	client openai.Client
}

// NewOpenAIFactory creates a new OpenAI factory with the given API key
func NewOpenAIFactory(apiKey string) *OpenAIFactory {
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &OpenAIFactory{client: client}
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
		strings.HasPrefix(lower, "o4") ||
		strings.HasPrefix(lower, "code-") ||
		strings.HasPrefix(lower, "davinci") ||
		strings.Contains(lower, "codex")
}

// CreateStream creates an OpenAI streaming provider
func (f *OpenAIFactory) CreateStream(ctx context.Context, req StreamRequest) (StreamProvider, error) {
	if isCodexModel(req.Model) {
		// Use Responses API for codex models
		// Build a single prompt string from messages
		var promptParts []string
		if req.SystemPrompt != "" {
			promptParts = append(promptParts, req.SystemPrompt)
		}
		for _, msg := range req.Messages {
			promptParts = append(promptParts, msg.Content)
		}
		prompt := strings.Join(promptParts, "\n")

		stream, err := f.StreamResponsesAPI(ctx, req.Model, prompt, req.MaxTokens)
		if err != nil {
			return nil, fmt.Errorf("openai responses stream: %w", err)
		}
		return &OpenAIResponsesStreamAdapter{
			stream:   stream,
			model:    req.Model,
			provider: ProviderOpenAI,
			message: &CompletedMessage{
				Provider: ProviderOpenAI,
				Model:    req.Model,
			},
		}, nil
	}

	// Use Chat Completions API for standard models
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages)+1)
	if req.SystemPrompt != "" {
		messages = append(messages, openai.SystemMessage(req.SystemPrompt))
	}
	for _, msg := range req.Messages {
		switch msg.Role {
		case "user":
			messages = append(messages, openai.UserMessage(msg.Content))
		case "assistant":
			messages = append(messages, openai.AssistantMessage(msg.Content))
		default:
			messages = append(messages, openai.UserMessage(msg.Content))
		}
	}

	stream, err := f.StreamChatCompletion(ctx, req.Model, messages, req.MaxTokens)
	if err != nil {
		return nil, fmt.Errorf("openai chat stream: %w", err)
	}
	return &OpenAIChatStreamAdapter{
		stream:   stream,
		model:    req.Model,
		provider: ProviderOpenAI,
		message: &CompletedMessage{
			Provider: ProviderOpenAI,
			Model:    req.Model,
		},
	}, nil
}

// StreamResponsesAPI streams using the Responses API for codex models
func (f *OpenAIFactory) StreamResponsesAPI(ctx context.Context, model string, prompt string, maxTokens int) (*ssestream.Stream[responses.ResponseStreamEventUnion], error) {
	params := responses.ResponseNewParams{
		Model: responses.ResponsesModel(model),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: responses.ResponseInputParam{
				responses.ResponseInputItemUnionParam{
					OfMessage: &responses.EasyInputMessageParam{
						Role:    responses.EasyInputMessageRoleUser,
						Content: responses.EasyInputMessageContentUnionParam{OfString: param.NewOpt(prompt)},
					},
				},
			},
		},
	}

	if maxTokens > 0 {
		params.MaxOutputTokens = param.NewOpt(int64(maxTokens))
	}

	stream := f.client.Responses.NewStreaming(ctx, params)
	return stream, nil
}

// StreamChatCompletion streams using the Chat Completions API
func (f *OpenAIFactory) StreamChatCompletion(ctx context.Context, model string, messages []openai.ChatCompletionMessageParamUnion, maxTokens int) (*ssestream.Stream[openai.ChatCompletionChunk], error) {
	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(model),
		Messages: messages,
	}

	if maxTokens > 0 {
		params.MaxCompletionTokens = param.NewOpt(int64(maxTokens))
	}

	stream := f.client.Chat.Completions.NewStreaming(ctx, params)
	return stream, nil
}

// IsCodexModel is an exported alias for isCodexModel for use outside this package
func IsCodexModel(model string) bool {
	return isCodexModel(model)
}

// isCodexModel determines if a model should use the Responses API
func isCodexModel(model string) bool {
	lower := strings.ToLower(model)
	codexPrefixes := []string{
		"code-",
		"davinci-codex",
		"gpt-5.1-codex",
	}

	for _, prefix := range codexPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}

	// Also match any model containing "codex"
	return strings.Contains(lower, "codex")
}

// usesMaxCompletionTokens returns true for models that use max_completion_tokens instead of max_tokens
func usesMaxCompletionTokens(model string) bool {
	lower := strings.ToLower(model)
	// o1, o3, o4 series use max_completion_tokens
	if strings.HasPrefix(lower, "o1") ||
		strings.HasPrefix(lower, "o3") ||
		strings.HasPrefix(lower, "o4") {
		return true
	}
	// gpt-5 and above (gpt-5, gpt-5-mini, gpt-5.x, gpt-5.x-codex, etc.)
	if strings.HasPrefix(lower, "gpt-5") {
		return true
	}
	return false
}

// OpenAIChatStreamAdapter wraps an OpenAI chat completion stream to implement StreamProvider
type OpenAIChatStreamAdapter struct {
	stream   *ssestream.Stream[openai.ChatCompletionChunk]
	model    string
	provider string
	current  StreamEvent
	message  *CompletedMessage
}

func (a *OpenAIChatStreamAdapter) Next() bool {
	for a.stream.Next() {
		chunk := a.stream.Current()
		if len(chunk.Choices) == 0 {
			continue
		}
		choice := chunk.Choices[0]
		if choice.Delta.Content == "" {
			continue
		}
		text := choice.Delta.Content
		a.message.Content += text
		a.current = StreamEvent{
			Type: "content_block_delta",
			Delta: &DeltaContent{
				Text: text,
				Type: "text_delta",
			},
		}
		return true
	}
	if err := a.stream.Err(); err != nil {
		a.current = StreamEvent{Error: err}
	}
	return false
}

func (a *OpenAIChatStreamAdapter) Current() StreamEvent {
	return a.current
}

func (a *OpenAIChatStreamAdapter) Err() error {
	return a.stream.Err()
}

func (a *OpenAIChatStreamAdapter) Close() error {
	return a.stream.Close()
}

func (a *OpenAIChatStreamAdapter) GetMessage() *CompletedMessage {
	return a.message
}

func (a *OpenAIChatStreamAdapter) GetModel() string {
	return a.model
}

func (a *OpenAIChatStreamAdapter) GetProvider() string {
	return a.provider
}

// OpenAIResponsesStreamAdapter wraps an OpenAI Responses API stream to implement StreamProvider
type OpenAIResponsesStreamAdapter struct {
	stream   *ssestream.Stream[responses.ResponseStreamEventUnion]
	model    string
	provider string
	current  StreamEvent
	message  *CompletedMessage
}

func (a *OpenAIResponsesStreamAdapter) Next() bool {
	for a.stream.Next() {
		event := a.stream.Current()
		// Extract text delta from responses stream event
		if event.Type == "response.output_text.delta" {
			delta := event.AsResponseOutputTextDelta()
			text := delta.Delta
			if text == "" {
				continue
			}
			a.message.Content += text
			a.current = StreamEvent{
				Type: "content_block_delta",
				Delta: &DeltaContent{
					Text: text,
					Type: "text_delta",
				},
			}
			return true
		}
	}
	if err := a.stream.Err(); err != nil {
		a.current = StreamEvent{Error: err}
	}
	return false
}

func (a *OpenAIResponsesStreamAdapter) Current() StreamEvent {
	return a.current
}

func (a *OpenAIResponsesStreamAdapter) Err() error {
	return a.stream.Err()
}

func (a *OpenAIResponsesStreamAdapter) Close() error {
	return a.stream.Close()
}

func (a *OpenAIResponsesStreamAdapter) GetMessage() *CompletedMessage {
	return a.message
}

func (a *OpenAIResponsesStreamAdapter) GetModel() string {
	return a.model
}

func (a *OpenAIResponsesStreamAdapter) GetProvider() string {
	return a.provider
}
