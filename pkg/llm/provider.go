package llm

import (
	"context"
	"encoding/json"
)

// Provider is the interface that LLM backends must implement.
// GoGrid is model-agnostic — swapping providers is a configuration change.
type Provider interface {
	// Complete sends a completion request to the LLM and returns the response.
	Complete(ctx context.Context, params Params) (*Response, error)
}

// StreamProvider is an optional extension for providers that support streaming.
type StreamProvider interface {
	Provider
	// Stream sends a completion request and returns a stream of events.
	Stream(ctx context.Context, params Params) (Stream, error)
}

// Stream is an iterator over streaming LLM response events.
type Stream interface {
	// Next advances to the next event. Returns false when the stream is exhausted.
	Next() bool
	// Event returns the current stream event. Only valid after Next returns true.
	Event() StreamEvent
	// Err returns the first error encountered during streaming.
	Err() error
	// Close releases resources associated with the stream.
	Close() error
}

// StreamEvent represents a single event in a streaming LLM response.
type StreamEvent struct {
	// Delta is the incremental text content in this event.
	Delta string
	// Message is the complete assembled message, populated on the final event.
	Message *Message
	// Usage is populated on the final event.
	Usage *Usage
}

// Params configures a single LLM completion request.
type Params struct {
	// Model is the model identifier (e.g. "gpt-4o", "claude-sonnet-4-5-20250929").
	Model string `json:"model"`
	// Messages is the conversation history to send.
	Messages []Message `json:"messages"`
	// Tools is the set of tools available to the LLM.
	Tools []ToolDefinition `json:"tools,omitempty"`
	// Temperature controls randomness. 0.0 = deterministic, 1.0 = creative.
	Temperature *float64 `json:"temperature,omitempty"`
	// MaxTokens limits the response length.
	MaxTokens int `json:"max_tokens,omitempty"`
	// StopSequences causes the LLM to stop generating when encountered.
	StopSequences []string `json:"stop_sequences,omitempty"`
}

// ToolDefinition describes a tool available to the LLM during completion.
type ToolDefinition struct {
	// Name is the tool's unique identifier.
	Name string `json:"name"`
	// Description explains what the tool does, helping the LLM decide when to use it.
	Description string `json:"description"`
	// Parameters is the JSON Schema describing the tool's input parameters.
	Parameters json.RawMessage `json:"parameters"`
}

// Response is the result of an LLM completion request.
type Response struct {
	// Message is the LLM's response message.
	Message Message `json:"message"`
	// Usage reports token consumption for this request.
	Usage Usage `json:"usage"`
	// Model is the actual model used (may differ from requested if aliased).
	Model string `json:"model"`
	// Metadata holds provider-specific response data.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Usage reports token consumption for a single LLM request.
type Usage struct {
	// PromptTokens is the number of tokens in the request.
	PromptTokens int `json:"prompt_tokens"`
	// CompletionTokens is the number of tokens in the response.
	CompletionTokens int `json:"completion_tokens"`
	// TotalTokens is PromptTokens + CompletionTokens.
	TotalTokens int `json:"total_tokens"`
}
