package streaming

import (
	"context"
	"encoding/json"
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
		var inputItems []responses.ResponseInputItemUnionParam
		for _, msg := range req.Messages {
			switch msg.Role {
			case "user":
				if len(msg.ToolResults) > 0 {
					for _, tr := range msg.ToolResults {
						inputItems = append(inputItems, responses.ResponseInputItemParamOfFunctionCallOutput(tr.ToolUseID, tr.Content))
					}
				} else {
					inputItems = append(inputItems, responses.ResponseInputItemUnionParam{OfMessage: &responses.EasyInputMessageParam{
						Role:    responses.EasyInputMessageRoleUser,
						Content: responses.EasyInputMessageContentUnionParam{OfString: param.NewOpt(msg.Content)},
					}})
				}
			case "assistant":
				if len(msg.ToolUses) > 0 {
					for _, tu := range msg.ToolUses {
						argsJSON, _ := json.Marshal(tu.Input)
						inputItems = append(inputItems, responses.ResponseInputItemParamOfFunctionCall(string(argsJSON), tu.ID, tu.Name))
					}
				} else {
					inputItems = append(inputItems, responses.ResponseInputItemUnionParam{OfMessage: &responses.EasyInputMessageParam{
						Role:    responses.EasyInputMessageRoleAssistant,
						Content: responses.EasyInputMessageContentUnionParam{OfString: param.NewOpt(msg.Content)},
					}})
				}
			}
		}
		params := responses.ResponseNewParams{
			Model: responses.ResponsesModel(req.Model),
			Input: responses.ResponseNewParamsInputUnion{OfInputItemList: inputItems},
		}
		if req.SystemPrompt != "" {
			params.Instructions = param.NewOpt(req.SystemPrompt)
		}
		if req.MaxTokens > 0 {
			params.MaxOutputTokens = param.NewOpt(int64(req.MaxTokens))
		}
		if len(req.Tools) > 0 {
			tools := make([]responses.ToolUnionParam, 0, len(req.Tools))
			for _, tool := range req.Tools {
				tools = append(tools, responses.ToolParamOfFunction(tool.Name, tool.InputSchema, false))
			}
			params.Tools = tools
		}

		stream := f.client.Responses.NewStreaming(ctx, params)
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
					argsJSON, _ := json.Marshal(tu.Input)
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
	chatStream := f.client.Chat.Completions.NewStreaming(ctx, chatParams)
	return &OpenAIChatStreamAdapter{
		stream:   chatStream,
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

	params.StreamOptions = openai.ChatCompletionStreamOptionsParam{IncludeUsage: param.NewOpt(true)}

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
	stream        *ssestream.Stream[openai.ChatCompletionChunk]
	acc           openai.ChatCompletionAccumulator
	pendingEvents []StreamEvent
	done          bool
	model         string
	provider      string
	current       StreamEvent
	message       *CompletedMessage
}

func (a *OpenAIChatStreamAdapter) Next() bool {
	if a.done {
		return false
	}
	if len(a.pendingEvents) > 0 {
		a.current = a.pendingEvents[0]
		a.pendingEvents = a.pendingEvents[1:]
		return true
	}
	for a.stream.Next() {
		chunk := a.stream.Current()
		a.acc.AddChunk(chunk)
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			text := chunk.Choices[0].Delta.Content
			a.message.Content += text
			a.current = StreamEvent{Type: "content_block_delta", Delta: &DeltaContent{Text: text, Type: "text_delta"}}
			return true
		}
	}
	if err := a.stream.Err(); err != nil {
		a.current = StreamEvent{Error: err}
		a.done = true
		return false
	}
	if len(a.acc.Choices) > 0 {
		for _, tc := range a.acc.Choices[0].Message.ToolCalls {
			var input map[string]interface{}
			json.Unmarshal([]byte(tc.Function.Arguments), &input)
			a.pendingEvents = append(a.pendingEvents, StreamEvent{Type: "tool_use", ToolUse: &ToolUse{ID: tc.ID, Name: tc.Function.Name, Input: input}})
		}
	}
	a.message.InputTokens = int(a.acc.Usage.PromptTokens)
	a.message.OutputTokens = int(a.acc.Usage.CompletionTokens)
	a.done = true
	if len(a.pendingEvents) > 0 {
		a.current = a.pendingEvents[0]
		a.pendingEvents = a.pendingEvents[1:]
		return true
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
	stream        *ssestream.Stream[responses.ResponseStreamEventUnion]
	pendingEvents []StreamEvent
	done          bool
	model         string
	provider      string
	current       StreamEvent
	message       *CompletedMessage
}

func (a *OpenAIResponsesStreamAdapter) Next() bool {
	if a.done {
		return false
	}
	if len(a.pendingEvents) > 0 {
		a.current = a.pendingEvents[0]
		a.pendingEvents = a.pendingEvents[1:]
		return true
	}
	for a.stream.Next() {
		event := a.stream.Current()
		switch event.Type {
		case "response.output_text.delta":
			delta := event.AsResponseOutputTextDelta()
			if delta.Delta == "" {
				continue
			}
			a.message.Content += delta.Delta
			a.current = StreamEvent{Type: "content_block_delta", Delta: &DeltaContent{Text: delta.Delta, Type: "text_delta"}}
			return true
		case "response.output_item.done":
			item := event.AsResponseOutputItemDone()
			if item.Item.Type == "function_call" {
				var input map[string]interface{}
				json.Unmarshal([]byte(item.Item.Arguments), &input)
				a.pendingEvents = append(a.pendingEvents, StreamEvent{Type: "tool_use", ToolUse: &ToolUse{ID: item.Item.CallID, Name: item.Item.Name, Input: input}})
			}
		case "response.completed":
			completed := event.AsResponseCompleted()
			a.message.InputTokens = int(completed.Response.Usage.InputTokens)
			a.message.OutputTokens = int(completed.Response.Usage.OutputTokens)
		}
	}
	if err := a.stream.Err(); err != nil {
		a.current = StreamEvent{Error: err}
	}
	a.done = true
	if len(a.pendingEvents) > 0 {
		a.current = a.pendingEvents[0]
		a.pendingEvents = a.pendingEvents[1:]
		return true
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
