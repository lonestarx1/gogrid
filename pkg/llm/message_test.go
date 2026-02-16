package llm

import (
	"encoding/json"
	"testing"
)

func TestRoleConstants(t *testing.T) {
	tests := []struct {
		role Role
		want string
	}{
		{RoleSystem, "system"},
		{RoleUser, "user"},
		{RoleAssistant, "assistant"},
		{RoleTool, "tool"},
	}
	for _, tt := range tests {
		if string(tt.role) != tt.want {
			t.Errorf("Role %v = %q, want %q", tt.role, string(tt.role), tt.want)
		}
	}
}

func TestMessageConstructors(t *testing.T) {
	tests := []struct {
		name    string
		msg     Message
		role    Role
		content string
	}{
		{"system", NewSystemMessage("you are helpful"), RoleSystem, "you are helpful"},
		{"user", NewUserMessage("hello"), RoleUser, "hello"},
		{"assistant", NewAssistantMessage("hi there"), RoleAssistant, "hi there"},
		{"tool", NewToolMessage("call-1", "result data"), RoleTool, "result data"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.msg.Role != tt.role {
				t.Errorf("Role = %q, want %q", tt.msg.Role, tt.role)
			}
			if tt.msg.Content != tt.content {
				t.Errorf("Content = %q, want %q", tt.msg.Content, tt.content)
			}
		})
	}
}

func TestToolMessageHasCallID(t *testing.T) {
	msg := NewToolMessage("call-42", "output")
	if msg.ToolCallID != "call-42" {
		t.Errorf("ToolCallID = %q, want %q", msg.ToolCallID, "call-42")
	}
}

func TestMessageJSONRoundTrip(t *testing.T) {
	original := Message{
		Role:    RoleAssistant,
		Content: "hello",
		ToolCalls: []ToolCall{
			{ID: "tc-1", Function: "search", Arguments: json.RawMessage(`{"q":"test"}`)},
		},
		Metadata: map[string]string{"key": "value"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Role != original.Role {
		t.Errorf("Role = %q, want %q", decoded.Role, original.Role)
	}
	if decoded.Content != original.Content {
		t.Errorf("Content = %q, want %q", decoded.Content, original.Content)
	}
	if len(decoded.ToolCalls) != 1 {
		t.Fatalf("ToolCalls len = %d, want 1", len(decoded.ToolCalls))
	}
	if decoded.ToolCalls[0].Function != "search" {
		t.Errorf("ToolCalls[0].Function = %q, want %q", decoded.ToolCalls[0].Function, "search")
	}
	if decoded.Metadata["key"] != "value" {
		t.Errorf("Metadata[key] = %q, want %q", decoded.Metadata["key"], "value")
	}
}

func TestToolCallArgumentsPreserved(t *testing.T) {
	tc := ToolCall{
		ID:        "tc-1",
		Function:  "calculate",
		Arguments: json.RawMessage(`{"expression":"2+2"}`),
	}

	data, err := json.Marshal(tc)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded ToolCall
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if string(decoded.Arguments) != `{"expression":"2+2"}` {
		t.Errorf("Arguments = %s, want %s", decoded.Arguments, `{"expression":"2+2"}`)
	}
}

func TestToolResultJSONRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		result ToolResult
	}{
		{
			name:   "success",
			result: ToolResult{CallID: "tc-1", Content: "42"},
		},
		{
			name:   "error",
			result: ToolResult{CallID: "tc-2", Error: "division by zero"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.result)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			var decoded ToolResult
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if decoded.CallID != tt.result.CallID {
				t.Errorf("CallID = %q, want %q", decoded.CallID, tt.result.CallID)
			}
			if decoded.Content != tt.result.Content {
				t.Errorf("Content = %q, want %q", decoded.Content, tt.result.Content)
			}
			if decoded.Error != tt.result.Error {
				t.Errorf("Error = %q, want %q", decoded.Error, tt.result.Error)
			}
		})
	}
}
