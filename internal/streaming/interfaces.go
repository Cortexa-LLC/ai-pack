package streaming

import (
	"context"

	"github.com/cortexa-llc/ai-pack/internal/constants"
)

// Re-export provider constants from the shared constants package so existing
// streaming-internal code can reference them without a package qualifier change.
const (
	ProviderAnthropic = constants.ProviderAnthropic
	ProviderOpenAI    = constants.ProviderOpenAI
	ProviderGemini    = constants.ProviderGemini
	ProviderQwen      = constants.ProviderQwen
)

// StreamEvent represents a single event in the stream
type StreamEvent struct {
	Type    string
	Delta   *DeltaContent
	ToolUse *ToolUse
	Message *CompletedMessage
	Error   error
}

// DeltaContent represents incremental content
type DeltaContent struct {
	Text string
	Type string
}

// CompletedMessage represents the final accumulated message
type CompletedMessage struct {
	ID           string
	Provider     string // Provider name (e.g., "anthropic", "openai")
	Model        string
	Role         string
	Content      string
	StopReason   string
	InputTokens  int
	OutputTokens int
}

// StreamProvider abstracts streaming from any LLM provider
type StreamProvider interface {
	// Next advances to the next event. Returns false when done or error.
	Next() bool

	// Current returns the current event
	Current() StreamEvent

	// Err returns any error that occurred during streaming
	Err() error

	// Close releases resources
	Close() error

	// GetMessage returns the final accumulated message
	GetMessage() *CompletedMessage

	// GetModel returns the model being used for this stream
	GetModel() string

	// GetProvider returns the provider name for this stream
	GetProvider() string
}

// StreamRequest represents a provider-agnostic streaming request
type StreamRequest struct {
	Messages         []Message
	SystemPrompt     string
	MaxTokens        int
	Tools            []Tool
	Model            string
	MinContextTokens int                    // Minimum context window required; filters out models that are too small (0 = no constraint)
	ProviderHints    map[string]interface{} // Provider-specific options
}

// Message represents a chat message with optional tool call history.
// Exactly one of the following shapes is expected per message:
//   - text-only:    Role + Content (no ToolUses/ToolResults)
//   - tool calls:   Role="assistant", ToolUses set (Content may be empty)
//   - tool results: Role="user", ToolResults set (Content empty)
type Message struct {
	Role        string       // "user" or "assistant"
	Content     string       // text content (may be empty when ToolUses is set)
	ToolUses    []ToolUse    // tool calls made by the assistant in this turn
	ToolResults []ToolResult // tool results returned by the user in this turn
}

// Tool represents a tool/function available to the model
type Tool struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

// ToolUse represents a tool invocation request from the model
type ToolUse struct {
	ID    string
	Name  string
	Input map[string]interface{}
}

// ToolResult represents the result of executing a tool
type ToolResult struct {
	ToolUseID string
	ToolName  string // function name; required for Gemini function responses
	Content   string
	IsError   bool
}

// ProviderFactory creates stream providers
type ProviderFactory interface {
	// CreateStream creates a streaming provider for the given request
	CreateStream(ctx context.Context, req StreamRequest) (StreamProvider, error)

	// GetProviderName returns the name of this provider
	GetProviderName() string

	// SupportsModel checks if this provider supports the given model
	SupportsModel(model string) bool
}
