package streaming

import (
	"context"
	"fmt"
	"strings"

	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
)

// buildQwenXMLToolCall reconstructs a ToolUse in Qwen's native XML tool call format.
// This keeps conversation history in the format the model was trained on, preventing
// confusion from seeing OpenAI-format tool_calls in the history.
func buildQwenXMLToolCall(tu ToolUse) string {
	var sb strings.Builder
	sb.WriteString("<tool_call>\n<function=")
	sb.WriteString(tu.Name)
	sb.WriteString(">\n")
	for k, v := range tu.Input {
		sb.WriteString("<parameter=")
		sb.WriteString(k)
		sb.WriteString(">")
		sb.WriteString(fmt.Sprintf("%v", v))
		sb.WriteString("</parameter>\n")
	}
	sb.WriteString("</function>\n</tool_call>")
	return sb.String()
}

// QwenFactory creates stream providers for Qwen models running locally via
// an OpenAI-compatible chat completions endpoint.
type QwenFactory struct {
	client openai.Client
}

// NewQwenFactory creates a new Qwen factory pointing at the local server.
// baseURL defaults to constants.QwenLocalBaseURL if empty.
// apiKey is passed to the local server; use "none" for servers that require no auth.
func NewQwenFactory(baseURL, apiKey string) *QwenFactory {
	if baseURL == "" {
		baseURL = constants.QwenLocalBaseURL
	}
	if apiKey == "" {
		apiKey = "none"
	}
	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(apiKey),
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
				// Use native Qwen tool_response format rather than OpenAI tool-role messages.
				// This keeps history in the format the model was trained on, preventing
				// confusion from mismatched message formats across long conversations.
				var sb strings.Builder
				for _, tr := range msg.ToolResults {
					sb.WriteString("<tool_response>\n")
					sb.WriteString(tr.Content)
					sb.WriteString("\n</tool_response>\n")
				}
				messages = append(messages, openai.UserMessage(strings.TrimSpace(sb.String())))
			} else {
				messages = append(messages, openai.UserMessage(msg.Content))
			}
		case "assistant":
			// Reconstruct tool calls as XML text so the history matches Qwen's
			// training format. OpenAI-format tool_calls in history confuse the model
			// and cause it to generate malformed output after many turns.
			var assistantText strings.Builder
			if msg.Content != "" {
				assistantText.WriteString(msg.Content)
			}
			for _, tu := range msg.ToolUses {
				if assistantText.Len() > 0 {
					assistantText.WriteString("\n")
				}
				assistantText.WriteString(buildQwenXMLToolCall(tu))
			}
			messages = append(messages, openai.AssistantMessage(assistantText.String()))
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
