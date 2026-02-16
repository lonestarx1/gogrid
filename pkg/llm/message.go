package llm

import "encoding/json"

// Role represents the role of a message participant.
type Role string

const (
	// RoleSystem is the system prompt role.
	RoleSystem Role = "system"
	// RoleUser is the user/human role.
	RoleUser Role = "user"
	// RoleAssistant is the LLM assistant role.
	RoleAssistant Role = "assistant"
	// RoleTool is the tool result role.
	RoleTool Role = "tool"
)

// Message represents a single message in a conversation.
type Message struct {
	// Role indicates who produced this message.
	Role Role `json:"role"`
	// Content is the text content of the message.
	Content string `json:"content"`
	// ToolCalls contains tool invocations requested by the assistant.
	// Only present when Role is RoleAssistant.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	// ToolCallID identifies which tool call this message is responding to.
	// Only present when Role is RoleTool.
	ToolCallID string `json:"tool_call_id,omitempty"`
	// Metadata holds arbitrary key-value pairs for extensibility.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ToolCall represents a tool invocation requested by the LLM.
type ToolCall struct {
	// ID is a unique identifier for this tool call.
	ID string `json:"id"`
	// Function is the name of the tool to invoke.
	Function string `json:"function"`
	// Arguments is the raw JSON arguments to pass to the tool.
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResult represents the output of a tool execution.
type ToolResult struct {
	// CallID references the ToolCall.ID this result responds to.
	CallID string `json:"call_id"`
	// Content is the string output of the tool.
	Content string `json:"content"`
	// Error is set if the tool execution failed.
	Error string `json:"error,omitempty"`
}

// NewSystemMessage creates a system prompt message.
func NewSystemMessage(content string) Message {
	return Message{Role: RoleSystem, Content: content}
}

// NewUserMessage creates a user message.
func NewUserMessage(content string) Message {
	return Message{Role: RoleUser, Content: content}
}

// NewAssistantMessage creates an assistant message.
func NewAssistantMessage(content string) Message {
	return Message{Role: RoleAssistant, Content: content}
}

// NewToolMessage creates a tool result message.
func NewToolMessage(callID, content string) Message {
	return Message{Role: RoleTool, Content: content, ToolCallID: callID}
}
