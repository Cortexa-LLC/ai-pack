package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
)

// QwenFactory creates stream providers for Qwen models running locally via
// an OpenAI-compatible chat completions endpoint.
type QwenFactory struct {
	client openai.Client
}

// NewQwenFactory creates a new Qwen factory pointing at the local server.
// baseURL defaults to constants.QwenLocalBaseURL if empty.
func NewQwenFactory(baseURL string) *QwenFactory {
	if baseURL == "" {
		baseURL = constants.QwenLocalBaseURL
	}
	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey("none"), // local server requires no real key
		option.WithMaxRetries(2),
	)
	return &QwenFactory{client: client}
}

// GetProviderName returns the provider name.
func (f *QwenFactory) GetProviderName() string {
	return ProviderQwen
}

// SupportsModel returns true for any model whose name starts with "qwen".
func (f *QwenFactory) SupportsModel(model string) bool {
	return strings.HasPrefix(strings.ToLower(model), "qwen")
}

// CreateStream creates a streaming provider for a Qwen model.
func (f *QwenFactory) CreateStream(ctx context.Context, req StreamRequest) (StreamProvider, error) {
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages)+1)
	if req.SystemPrompt != "" {
		messages = append(messages, openai.SystemMessage(req.SystemPrompt))
	}
	for _, msg := range req.Messages {
		switch msg.Role {
		case "user":
			if len(msg.ToolResults) > 0 {
				for _, tr := range msg.ToolResults {
					messages = append(messages, openai.ToolMessage(tr.Content, tr.ToolUseID))
				}
			} else {
				messages = append(messages, openai.UserMessage(msg.Content))
			}
		case "assistant":
			if len(msg.ToolUses) > 0 {
				var toolCalls []openai.ChatCompletionMessageToolCallParam
				for _, tu := range msg.ToolUses {
					argsJSON, err := json.Marshal(tu.Input)
					if err != nil {
						return nil, fmt.Errorf("marshal tool input for %q: %w", tu.Name, err)
					}
					toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallParam{
						ID:   tu.ID,
						Type: "function",
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      tu.Name,
							Arguments: string(argsJSON),
						},
					})
				}
				messages = append(messages, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{ToolCalls: toolCalls},
				})
			} else {
				messages = append(messages, openai.AssistantMessage(msg.Content))
			}
		default:
			messages = append(messages, openai.UserMessage(msg.Content))
		}
	}

	chatParams := openai.ChatCompletionNewParams{
		Model:         openai.ChatModel(req.Model),
		Messages:      messages,
		StreamOptions: openai.ChatCompletionStreamOptionsParam{IncludeUsage: param.NewOpt(true)},
	}
	if req.MaxTokens > 0 {
		chatParams.MaxCompletionTokens = param.NewOpt(int64(req.MaxTokens))
	}
	if len(req.Tools) > 0 {
		chatTools := make([]openai.ChatCompletionToolParam, 0, len(req.Tools))
		for _, tool := range req.Tools {
			chatTools = append(chatTools, openai.ChatCompletionToolParam{
				Type: "function",
				Function: openai.FunctionDefinitionParam{
					Name:        tool.Name,
					Description: param.NewOpt(tool.Description),
					Parameters:  openai.FunctionParameters(tool.InputSchema),
				},
			})
		}
		chatParams.Tools = chatTools
	}

	stream := f.client.Chat.Completions.NewStreaming(ctx, chatParams)
	return &OpenAIChatStreamAdapter{
		stream:   stream,
		model:    req.Model,
		provider: ProviderQwen,
		message: &CompletedMessage{
			Provider: ProviderQwen,
			Model:    req.Model,
		},
	}, nil
}
