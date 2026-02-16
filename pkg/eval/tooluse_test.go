package eval

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
)

func resultWithToolCalls(calls ...llm.ToolCall) *agent.Result {
	return &agent.Result{
		RunID:   "test-run",
		Message: llm.NewAssistantMessage("done"),
		History: []llm.Message{
			llm.NewUserMessage("do something"),
			{
				Role:      llm.RoleAssistant,
				ToolCalls: calls,
			},
			llm.NewToolMessage("tc-1", "result"),
			llm.NewAssistantMessage("done"),
		},
	}
}

func TestToolUseCalled(t *testing.T) {
	ev := NewToolUse(ToolExpectation{Name: "search"})
	result := resultWithToolCalls(llm.ToolCall{
		ID:        "tc-1",
		Function:  "search",
		Arguments: json.RawMessage(`{"q":"test"}`),
	})

	score, err := ev.Evaluate(context.Background(), result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !score.Pass {
		t.Error("expected Pass = true")
	}
}

func TestToolUseNotCalled(t *testing.T) {
	ev := NewToolUse(ToolExpectation{Name: "search"})
	result := resultWithToolCalls(llm.ToolCall{
		ID:        "tc-1",
		Function:  "calculate",
		Arguments: json.RawMessage(`{}`),
	})

	score, err := ev.Evaluate(context.Background(), result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Pass {
		t.Error("expected Pass = false")
	}
	if score.Value != 0.0 {
		t.Errorf("Value = %f, want 0.0", score.Value)
	}
}

func TestToolUseArgMatcher(t *testing.T) {
	matcher := func(args json.RawMessage) bool {
		return strings.Contains(string(args), "test")
	}

	ev := NewToolUse(ToolExpectation{
		Name:       "search",
		ArgMatcher: matcher,
	})

	t.Run("matching args", func(t *testing.T) {
		result := resultWithToolCalls(llm.ToolCall{
			ID:        "tc-1",
			Function:  "search",
			Arguments: json.RawMessage(`{"q":"test query"}`),
		})
		score, _ := ev.Evaluate(context.Background(), result)
		if !score.Pass {
			t.Error("expected Pass = true")
		}
	})

	t.Run("non-matching args", func(t *testing.T) {
		result := resultWithToolCalls(llm.ToolCall{
			ID:        "tc-1",
			Function:  "search",
			Arguments: json.RawMessage(`{"q":"other"}`),
		})
		score, _ := ev.Evaluate(context.Background(), result)
		if score.Pass {
			t.Error("expected Pass = false")
		}
	})
}

func TestToolUseMultipleCalls(t *testing.T) {
	ev := NewToolUse(ToolExpectation{Name: "search", MinCalls: 2})
	result := resultWithToolCalls(
		llm.ToolCall{ID: "tc-1", Function: "search", Arguments: json.RawMessage(`{}`)},
		llm.ToolCall{ID: "tc-2", Function: "search", Arguments: json.RawMessage(`{}`)},
	)

	score, err := ev.Evaluate(context.Background(), result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !score.Pass {
		t.Error("expected Pass = true")
	}
}

func TestToolUseMinCallsNotMet(t *testing.T) {
	ev := NewToolUse(ToolExpectation{Name: "search", MinCalls: 3})
	result := resultWithToolCalls(
		llm.ToolCall{ID: "tc-1", Function: "search", Arguments: json.RawMessage(`{}`)},
	)

	score, err := ev.Evaluate(context.Background(), result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Pass {
		t.Error("expected Pass = false")
	}
}

func TestToolUseMultipleExpectations(t *testing.T) {
	ev := NewToolUse(
		ToolExpectation{Name: "search"},
		ToolExpectation{Name: "calculate"},
	)

	t.Run("both met", func(t *testing.T) {
		result := resultWithToolCalls(
			llm.ToolCall{ID: "tc-1", Function: "search", Arguments: json.RawMessage(`{}`)},
			llm.ToolCall{ID: "tc-2", Function: "calculate", Arguments: json.RawMessage(`{}`)},
		)
		score, _ := ev.Evaluate(context.Background(), result)
		if !score.Pass {
			t.Error("expected Pass = true")
		}
	})

	t.Run("one met", func(t *testing.T) {
		result := resultWithToolCalls(
			llm.ToolCall{ID: "tc-1", Function: "search", Arguments: json.RawMessage(`{}`)},
		)
		score, _ := ev.Evaluate(context.Background(), result)
		if score.Pass {
			t.Error("expected Pass = false")
		}
		if score.Value != 0.5 {
			t.Errorf("Value = %f, want 0.5", score.Value)
		}
	})
}

func TestToolUseEmptyExpectations(t *testing.T) {
	ev := NewToolUse()
	score, err := ev.Evaluate(context.Background(), testResult("anything"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !score.Pass {
		t.Error("expected Pass = true for empty expectations")
	}
}
