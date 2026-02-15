package server

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
	"github.com/sashabaranov/go-openai"
)

// LLMProvider defines the interface for different LLM providers
type LLMProvider interface {
	// StreamCompletion streams a chat completion response
	StreamCompletion(ctx context.Context, messages []LLMMessage, systemPrompt string, tools []LLMTool) (*LLMStream, error)

	// GetProviderName returns the provider name (openai, anthropic)
	GetProviderName() string

	// GetModelName returns the model identifier
	GetModelName() string
}

// LLMMessage represents a message in the conversation (provider-agnostic)
type LLMMessage struct {
	Role    string // "user" or "assistant"
	Content string
}

// LLMTool represents a tool definition (provider-agnostic)
type LLMTool struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

// LLMStream represents a streaming response
type LLMStream struct {
	TextChan   chan string
	ErrorChan  chan error
	DoneChan   chan struct{}
	ToolCalls  []LLMToolCall
	InputTokens  int
	OutputTokens int
}

// LLMToolCall represents a tool call from the LLM
type LLMToolCall struct {
	ID    string
	Name  string
	Input map[string]interface{}
}

// AnthropicProvider implements LLMProvider for Anthropic Claude
type AnthropicProvider struct {
	client    anthropic.Client
	model     string
	maxTokens int
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(client anthropic.Client, model string, maxTokens int) *AnthropicProvider {
	return &AnthropicProvider{
		client:    client,
		model:     model,
		maxTokens: maxTokens,
	}
}

func (p *AnthropicProvider) GetProviderName() string {
	return "anthropic"
}

func (p *AnthropicProvider) GetModelName() string {
	return p.model
}

func (p *AnthropicProvider) StreamCompletion(ctx context.Context, messages []LLMMessage, systemPrompt string, tools []LLMTool) (*LLMStream, error) {
	// Convert to Anthropic format
	claudeMessages := make([]anthropic.MessageParam, len(messages))
	for i, msg := range messages {
		if msg.Role == "user" {
			claudeMessages[i] = anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content))
		} else {
			claudeMessages[i] = anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content))
		}
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: int64(p.maxTokens),
		Messages:  claudeMessages,
	}

	// Add system prompt if provided
	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{
				Type: "text",
				Text: systemPrompt,
			},
		}
	}

	// Convert tools to Anthropic format
	if len(tools) > 0 {
		claudeTools := make([]anthropic.ToolUnionParam, len(tools))
		for i, tool := range tools {
			claudeTools[i] = anthropic.ToolUnionParamOfTool(
				anthropic.ToolInputSchemaParam{
					Type:       "object",
					Properties: tool.InputSchema,
				},
				tool.Name,
			)
		}
		params.Tools = claudeTools
	}

	stream := p.client.Messages.NewStreaming(ctx, params)

	// Create response channels
	textChan := make(chan string, 100)
	errorChan := make(chan error, 1)
	doneChan := make(chan struct{})

	llmStream := &LLMStream{
		TextChan:  textChan,
		ErrorChan: errorChan,
		DoneChan:  doneChan,
	}

	// Stream in background
	go func() {
		defer close(textChan)
		defer close(errorChan)
		defer close(doneChan)

		var message anthropic.Message
		for stream.Next() {
			event := stream.Current()

			// Send text deltas
			if event.Type == "content_block_delta" {
				if event.Delta.Type == "text_delta" {
					textChan <- event.Delta.Text
				}
			}

			// Accumulate message
			if err := message.Accumulate(event); err != nil {
				monitoring.Logger.Error("anthropic_stream_accumulate_error", "error", err)
			}
		}

		if err := stream.Err(); err != nil {
			errorChan <- err
			return
		}

		// Capture usage
		llmStream.InputTokens = int(message.Usage.InputTokens)
		llmStream.OutputTokens = int(message.Usage.OutputTokens)

		// Extract tool calls
		for _, block := range message.Content {
			if block.Type == "tool_use" {
				var input map[string]interface{}
				if block.Input != nil {
					// block.Input is json.RawMessage
					// For now, we'll handle tool extraction later
				}
				llmStream.ToolCalls = append(llmStream.ToolCalls, LLMToolCall{
					ID:    block.ID,
					Name:  block.Name,
					Input: input,
				})
			}
		}
	}()

	return llmStream, nil
}

// OpenAIProvider implements LLMProvider for OpenAI GPT models
type OpenAIProvider struct {
	client    *openai.Client
	model     string
	maxTokens int
}

// NewOpenAIProvider creates a new OpenAI provider
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

func (p *OpenAIProvider) StreamCompletion(ctx context.Context, messages []LLMMessage, systemPrompt string, tools []LLMTool) (*LLMStream, error) {
	// Convert to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessage, 0, len(messages)+1)

	// Add system prompt as first message
	if systemPrompt != "" {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		})
	}

	// Add conversation messages
	for _, msg := range messages {
		role := openai.ChatMessageRoleUser
		if msg.Role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		}
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	req := openai.ChatCompletionRequest{
		Model:     p.model,
		Messages:  openaiMessages,
		MaxTokens: p.maxTokens,
		Stream:    true,
	}

	// Convert tools to OpenAI format
	if len(tools) > 0 {
		openaiTools := make([]openai.Tool, len(tools))
		for i, tool := range tools {
			openaiTools[i] = openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.InputSchema,
				},
			}
		}
		req.Tools = openaiTools
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openai stream creation failed: %w", err)
	}

	// Create response channels
	textChan := make(chan string, 100)
	errorChan := make(chan error, 1)
	doneChan := make(chan struct{})

	llmStream := &LLMStream{
		TextChan:  textChan,
		ErrorChan: errorChan,
		DoneChan:  doneChan,
	}

	// Stream in background
	go func() {
		defer close(textChan)
		defer close(errorChan)
		defer close(doneChan)
		defer stream.Close()

		var fullResponse strings.Builder
		var toolCallsBuffer []openai.ToolCall

		for {
			response, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				errorChan <- fmt.Errorf("openai stream error: %w", err)
				return
			}

			// Extract text deltas
			if len(response.Choices) > 0 {
				delta := response.Choices[0].Delta
				if delta.Content != "" {
					textChan <- delta.Content
					fullResponse.WriteString(delta.Content)
				}

				// Collect tool calls
				if len(delta.ToolCalls) > 0 {
					toolCallsBuffer = append(toolCallsBuffer, delta.ToolCalls...)
				}
			}

			// Capture usage (available in final chunk)
			if response.Usage != nil {
				llmStream.InputTokens = response.Usage.PromptTokens
				llmStream.OutputTokens = response.Usage.CompletionTokens
			}
		}

		// Convert OpenAI tool calls to our format
		for _, tc := range toolCallsBuffer {
			if tc.Function.Name != "" {
				llmStream.ToolCalls = append(llmStream.ToolCalls, LLMToolCall{
					ID:   tc.ID,
					Name: tc.Function.Name,
					// Input parsing would go here
				})
			}
		}
	}()

	return llmStream, nil
}

// SelectProvider chooses the appropriate provider based on model name
func SelectProvider(modelName string, anthropicProvider *AnthropicProvider, openaiProvider *OpenAIProvider) LLMProvider {
	// Auto-detect provider from model name
	if strings.HasPrefix(modelName, "gpt-") {
		// Create OpenAI provider with specific model
		return NewOpenAIProvider(openaiProvider.client, modelName, openaiProvider.maxTokens)
	}

	// Default to Anthropic (claude-*)
	return NewAnthropicProvider(anthropicProvider.client, modelName, anthropicProvider.maxTokens)
}
