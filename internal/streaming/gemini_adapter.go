package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	genai "google.golang.org/genai"
)

// GeminiFactory creates Gemini stream providers using the official google.golang.org/genai SDK.
type GeminiFactory struct {
	apiKey    string
	maxTokens int
}

// NewGeminiFactory creates a new GeminiFactory with the given API key and max output tokens.
func NewGeminiFactory(apiKey string, maxTokens int) *GeminiFactory {
	return &GeminiFactory{
		apiKey:    apiKey,
		maxTokens: maxTokens,
	}
}

// GetProviderName returns the provider identifier.
func (f *GeminiFactory) GetProviderName() string {
	return ProviderGemini
}

// SupportsModel returns true if the model name contains "gemini".
func (f *GeminiFactory) SupportsModel(model string) bool {
	return strings.Contains(strings.ToLower(model), "gemini")
}

// CreateStream converts messages/tools into a Gemini streaming session.
func (f *GeminiFactory) CreateStream(ctx context.Context, req StreamRequest) (StreamProvider, error) {
	// Build the Gemini content history from the canonical message list.
	contents, err := buildGeminiContents(req.Messages)
	if err != nil {
		return nil, fmt.Errorf("gemini: build contents: %w", err)
	}

	// Build generation config.
	config := &genai.GenerateContentConfig{
		MaxOutputTokens: int32(f.maxTokens),
	}
	if req.SystemPrompt != "" {
		config.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{{Text: req.SystemPrompt}},
		}
	}
	if len(req.Tools) > 0 {
		decls, err := buildGeminiFunctionDeclarations(req.Tools)
		if err != nil {
			return nil, fmt.Errorf("gemini: build tools: %w", err)
		}
		config.Tools = []*genai.Tool{{FunctionDeclarations: decls}}
	}

	// Create client inside the stream so it is bound to the request context.
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  f.apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini: create client: %w", err)
	}

	eventCh := make(chan StreamEvent, 64)

	// Run the streaming call in a background goroutine.
	go func() {
		defer close(eventCh)

		for resp, iterErr := range client.Models.GenerateContentStream(ctx, req.Model, contents, config) {
			if iterErr != nil {
				select {
				case eventCh <- StreamEvent{Error: fmt.Errorf("gemini stream error: %w", iterErr)}:
				case <-ctx.Done():
				}
				return
			}

			// Emit text delta if present.
			if text := resp.Text(); text != "" {
				select {
				case eventCh <- StreamEvent{
					Type:  "content_block_delta",
					Delta: &DeltaContent{Text: text, Type: "text_delta"},
				}:
				case <-ctx.Done():
					return
				}
			}

			// Emit tool-use events for any function calls.
			if fcs := resp.FunctionCalls(); len(fcs) > 0 {
				for idx, fc := range fcs {
					toolID := fmt.Sprintf("gemini-fc-%s-%d", fc.Name, idx)
					// Marshal args back to JSON for the canonical ToolUse type.
					inputJSON, _ := json.Marshal(fc.Args)
					var inputMap map[string]interface{}
					_ = json.Unmarshal(inputJSON, &inputMap)

					select {
					case eventCh <- StreamEvent{
						Type: "tool_use",
						ToolUse: &ToolUse{
							ID:    toolID,
							Name:  fc.Name,
							Input: inputMap,
						},
					}:
					case <-ctx.Done():
						return
					}
				}
			}

			// On the final chunk emit a completed message with usage metadata.
			if resp.UsageMetadata != nil && len(resp.Candidates) > 0 {
				cand := resp.Candidates[0]
				stopReason := ""
				if cand.FinishReason != "" {
					stopReason = string(cand.FinishReason)
				}

				select {
				case eventCh <- StreamEvent{
					Type: "message_stop",
					Message: &CompletedMessage{
						StopReason:   stopReason,
						InputTokens:  int(resp.UsageMetadata.PromptTokenCount),
						OutputTokens: int(resp.UsageMetadata.CandidatesTokenCount),
					},
				}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return &GeminiStreamProvider{
		eventCh:  eventCh,
		model:    req.Model,
		provider: ProviderGemini,
		message:  &CompletedMessage{},
	}, nil
}

// GeminiStreamProvider reads events produced by the background streaming goroutine
// and implements the StreamProvider interface.
type GeminiStreamProvider struct {
	eventCh  chan StreamEvent
	model    string
	provider string
	current  StreamEvent
	message  *CompletedMessage
	err      error
	done     bool
}

// Next advances to the next event. Returns false when done or error.
func (p *GeminiStreamProvider) Next() bool {
	if p.done {
		return false
	}
	event, ok := <-p.eventCh
	if !ok {
		p.done = true
		return false
	}
	if event.Error != nil {
		p.err = event.Error
		p.done = true
		return false
	}
	// Accumulate the final message metadata when message_stop arrives.
	if event.Type == "message_stop" && event.Message != nil {
		p.message.StopReason = event.Message.StopReason
		p.message.InputTokens = event.Message.InputTokens
		p.message.OutputTokens = event.Message.OutputTokens
		p.done = true
		return false
	}
	// Accumulate content.
	if event.Delta != nil {
		p.message.Content += event.Delta.Text
	}
	p.current = event
	return true
}

// Current returns the current event.
func (p *GeminiStreamProvider) Current() StreamEvent {
	return p.current
}

// Err returns any error that occurred during streaming.
func (p *GeminiStreamProvider) Err() error {
	return p.err
}

// Close drains the event channel so the goroutine can exit cleanly.
func (p *GeminiStreamProvider) Close() error {
	for range p.eventCh {
	}
	return nil
}

// GetMessage returns the final accumulated message.
func (p *GeminiStreamProvider) GetMessage() *CompletedMessage {
	return p.message
}

// GetModel returns the model being used for this stream.
func (p *GeminiStreamProvider) GetModel() string {
	return p.model
}

// GetProvider returns the provider name for this stream.
func (p *GeminiStreamProvider) GetProvider() string {
	return p.provider
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildGeminiContents converts the canonical message list to Gemini Content objects.
func buildGeminiContents(messages []Message) ([]*genai.Content, error) {
	var contents []*genai.Content
	for _, msg := range messages {
		switch msg.Role {
		case "tool_result":
			// Tool results come back as user-role function responses.
			for _, tr := range msg.ToolResults {
				c := genai.NewContentFromFunctionResponse(
					tr.ToolUseID,
					map[string]any{"output": tr.Content},
					genai.RoleUser,
				)
				contents = append(contents, c)
			}
		case "assistant":
			if len(msg.ToolUses) > 0 {
				// Assistant tool calls.
				for _, tu := range msg.ToolUses {
					c := genai.NewContentFromFunctionCall(tu.Name, tu.Input, genai.RoleModel)
					contents = append(contents, c)
				}
			} else {
				contents = append(contents, genai.NewContentFromText(msg.Content, genai.RoleModel))
			}
		default:
			// user / system treated as user
			contents = append(contents, genai.NewContentFromText(msg.Content, genai.RoleUser))
		}
	}
	return contents, nil
}

// buildGeminiFunctionDeclarations converts canonical Tool definitions to Gemini FunctionDeclarations.
func buildGeminiFunctionDeclarations(tools []Tool) ([]*genai.FunctionDeclaration, error) {
	decls := make([]*genai.FunctionDeclaration, 0, len(tools))
	for _, t := range tools {
		schema, err := buildGeminiSchema(t.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("tool %q: %w", t.Name, err)
		}
		decls = append(decls, &genai.FunctionDeclaration{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  schema,
		})
	}
	return decls, nil
}

// buildGeminiSchema converts a map[string]interface{} JSON-schema representation
// into a *genai.Schema recursively.
func buildGeminiSchema(raw map[string]interface{}) (*genai.Schema, error) {
	if raw == nil {
		return &genai.Schema{Type: genai.TypeObject}, nil
	}

	schema := &genai.Schema{}

	if t, ok := raw["type"].(string); ok {
		switch strings.ToLower(t) {
		case "object":
			schema.Type = genai.TypeObject
		case "array":
			schema.Type = genai.TypeArray
		case "string":
			schema.Type = genai.TypeString
		case "number":
			schema.Type = genai.TypeNumber
		case "integer":
			schema.Type = genai.TypeInteger
		case "boolean":
			schema.Type = genai.TypeBoolean
		default:
			schema.Type = genai.TypeString
		}
	} else {
		schema.Type = genai.TypeObject
	}

	if desc, ok := raw["description"].(string); ok {
		schema.Description = desc
	}

	// Properties (for object types)
	if props, ok := raw["properties"].(map[string]interface{}); ok {
		schema.Properties = make(map[string]*genai.Schema, len(props))
		for name, val := range props {
			propMap, ok := val.(map[string]interface{})
			if !ok {
				continue
			}
			propSchema, err := buildGeminiSchema(propMap)
			if err != nil {
				return nil, err
			}
			schema.Properties[name] = propSchema
		}
	}

	// Required fields
	if req, ok := raw["required"].([]interface{}); ok {
		for _, r := range req {
			if s, ok := r.(string); ok {
				schema.Required = append(schema.Required, s)
			}
		}
	}

	// Items schema (for array types)
	if items, ok := raw["items"].(map[string]interface{}); ok {
		itemSchema, err := buildGeminiSchema(items)
		if err != nil {
			return nil, err
		}
		schema.Items = itemSchema
	}

	return schema, nil
}
