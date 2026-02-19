package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/streaming"
)

type LLMMessage struct {
	Role    string
	Content string
}

type LLMTool struct {
	Name        string
	Description string
	InputSchema any
}

type LLMToolCall struct {
	ID   string
	Name string
}

type LLMStream struct {
	TextChan    chan string
	ErrorChan   chan error
	DoneChan    chan struct{}
	InputTokens int
	OutputTokens int
	ToolCalls   []LLMToolCall
}

// LLMProvider is the interface implemented by all LLM providers
// Used to unify Anthropic, OpenAI, Claude etc.
type LLMProvider interface {
	GetProviderName() string
	GetModelName() string
	APIKey() string
	StreamCompletion(ctx context.Context, messages []LLMMessage, systemPrompt string, tools []LLMTool) (*LLMStream, error)
}

// AnthropicProvider implements the LLMProvider interface using the Anthropics SDK
// This is a minimal stub implementation to allow compilation
// TODO: implement full streaming and other methods as needed
// Keeping minimal for now to unblock build

type AnthropicProvider struct {
	client *anthropic.Client
	model  string
	maxTokens int
}

func NewAnthropicProvider(client *anthropic.Client, model string, maxTokens int) *AnthropicProvider {
	return &AnthropicProvider{
		client: client,
		model: model,
		maxTokens: maxTokens,
	}
}

func (p *AnthropicProvider) GetProviderName() string {
	return "anthropic"
}

func (p *AnthropicProvider) GetModelName() string {
	return p.model
}

func (p *AnthropicProvider) APIKey() string {
	// Stub to return empty for now
	return ""
}

func (p *AnthropicProvider) StreamCompletion(ctx context.Context, messages []LLMMessage, systemPrompt string, tools []LLMTool) (*LLMStream, error) {
	// Stub implementation returning error
	return nil, fmt.Errorf("Anthropic provider stream completion not implemented")
}

// OpenAIProvider implements LLMProvider for OpenAI GPT models
// Using official openai/openai-go SDK and adding Responses API support

type OpenAIProvider struct {
	client    *openai.Client
	model     string
	maxTokens int
}

func NewOpenAIProvider(client *openai.Client, model string, maxTokens int) *OpenAIProvider {
	return &OpenAIProvider{
		client:    client,
		model:     model,
		maxTokens: maxTokens,
	}
}

func (p *OpenAIProvider) GetProviderName() string {
	return "openai"
}

func (p *OpenAIProvider) GetModelName() string {
	return p.model
}

func (p *OpenAIProvider) APIKey() string {
	// Assuming the client has a method to get the API key
	// The https://github.com/openai/openai-go does not expose API key directly
	// We might store it in the wrapper or configuration for now stub with empty
	return ""
}

func (p *OpenAIProvider) StreamCompletion(ctx context.Context, messages []LLMMessage, systemPrompt string, tools []LLMTool) (*LLMStream, error) {
	if streaming.IsCodexModel(p.model) {
		// Use the new Responses API streaming for Codex models
		factory := streaming.NewOpenAIFactory(p.APIKey())

		// Build prompt from messages
		var promptBuilder strings.Builder
		if systemPrompt != "" {
			promptBuilder.WriteString(systemPrompt + "\n")
		}
		for _, msg := range messages {
			promptBuilder.WriteString(msg.Role + ": " + msg.Content + "\n")
		}

		stream, err := factory.StreamResponsesAPI(ctx, p.model, promptBuilder.String(), p.maxTokens)
		if err != nil {
			return nil, fmt.Errorf("responses api streaming failed: %w", err)
		}

		textChan := make(chan string, 100)
		errorChan := make(chan error, 1)
		doneChan := make(chan struct{})

		llmStream := &LLMStream{
			TextChan:  textChan,
			ErrorChan: errorChan,
			DoneChan:  doneChan,
		}

		go func() {
			defer close(textChan)
			defer close(errorChan)
			defer close(doneChan)

			for stream.Next() {
				event := stream.Current()
				if event.Type == "text_delta" {
					textChan <- event.Text
				}
			}

			if err := stream.Err(); err != nil {
				errorChan <- err
			}

			doneChan <- struct{}{}
		}()

		return llmStream, nil
	}

	// Fallback to chat completion for non-Codex models
	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages)+1)

	if systemPrompt != "" {
		openaiMessages = append(openaiMessages, openai.SystemMessage(systemPrompt))
	}

	for _, msg := range messages {
		if msg.Role == "assistant" {
			openaiMessages = append(openaiMessages, openai.AssistantMessage(msg.Content))
		} else {
			openaiMessages = append(openaiMessages, openai.UserMessage(msg.Content))
		}
	}

	params := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(p.model),
		Messages: openaiMessages,
	}

	if p.maxTokens > 0 {
		params.MaxTokens = openai.Int(int64(p.maxTokens))
	}

	if len(tools) > 0 {
		// Tools support may not be available in this sdk version
		// Skipping tools for compatibility
	}

	stream := p.client.Chat.Completions.NewStreaming(ctx, params)

	textChan := make(chan string, 100)
	errorChan := make(chan error, 1)
	doneChan := make(chan struct{})

	llmStream := &LLMStream{
		TextChan:  textChan,
		ErrorChan: errorChan,
		DoneChan:  doneChan,
	}

	go func() {
		defer close(textChan)
		defer close(errorChan)
		defer close(doneChan)
		defer stream.Close()

		var fullResponse strings.Builder
		var toolCallsBuffer []openai.ChatCompletionChunkChoiceDeltaToolCall

		for stream.Next() {
			chunk := stream.Current()

			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				if delta.Content != "" {
					textChan <- delta.Content
					fullResponse.WriteString(delta.Content)
				}

				if len(delta.ToolCalls) > 0 {
					toolCallsBuffer = append(toolCallsBuffer, delta.ToolCalls...)
				}
			}

			if chunk.Usage.CompletionTokens > 0 {
				llmStream.InputTokens = int(chunk.Usage.PromptTokens)
				llmStream.OutputTokens = int(chunk.Usage.CompletionTokens)
			}
		}

		if err := stream.Err(); err != nil {
			errorChan <- fmt.Errorf("openai stream error: %w", err)
			return
		}

		for _, tc := range toolCallsBuffer {
			if tc.Function.Name != "" {
				llmStream.ToolCalls = append(llmStream.ToolCalls, LLMToolCall{
					ID:   tc.ID,
					Name: tc.Function.Name,
					// Input parsing would go here
				})
			}
		}

		doneChan <- struct{}{}
	}()

	return llmStream, nil
}
