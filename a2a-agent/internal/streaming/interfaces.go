package streaming

import "context"

// StreamEvent represents a single event in the stream
type StreamEvent struct {
	Type    string
	Delta   *DeltaContent
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
}

// StreamRequest represents a provider-agnostic streaming request
type StreamRequest struct {
	Messages      []Message
	SystemPrompt  string
	MaxTokens     int
	Tools         []Tool
	Model         string
	ProviderHints map[string]interface{} // Provider-specific options
}

// Message represents a chat message
type Message struct {
	Role    string // "user" or "assistant"
	Content string
}

// Tool represents a tool/function available to the model
type Tool struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
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
