package eval

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
)

// ToolExpectation defines an expected tool call.
type ToolExpectation struct {
	// Name is the tool function name to look for.
	Name string
	// ArgMatcher optionally validates call arguments. If nil, only the
	// tool name is checked.
	ArgMatcher func(json.RawMessage) bool
	// MinCalls is the minimum number of matching calls required.
	// 0 defaults to 1.
	MinCalls int
}

// ToolUse evaluates whether the agent's conversation history includes
// the expected tool calls.
type ToolUse struct {
	expectations []ToolExpectation
}

// NewToolUse creates a ToolUse evaluator with the given expectations.
func NewToolUse(expectations ...ToolExpectation) *ToolUse {
	return &ToolUse{expectations: expectations}
}

// Name returns "tool_use".
func (tu *ToolUse) Name() string { return "tool_use" }

// Evaluate inspects result.History for tool calls matching expectations.
func (tu *ToolUse) Evaluate(_ context.Context, result *agent.Result) (Score, error) {
	if len(tu.expectations) == 0 {
		return Score{Pass: true, Value: 1.0, Reason: "no tool expectations"}, nil
	}

	// Collect all tool calls by function name.
	calls := collectToolCalls(result.History)

	met := 0
	var unmet []string
	for _, exp := range tu.expectations {
		minCalls := exp.MinCalls
		if minCalls <= 0 {
			minCalls = 1
		}

		toolCalls := calls[exp.Name]
		if exp.ArgMatcher == nil {
			if len(toolCalls) >= minCalls {
				met++
			} else {
				unmet = append(unmet, fmt.Sprintf("%s (want >=%d, got %d)", exp.Name, minCalls, len(toolCalls)))
			}
		} else {
			matching := 0
			for _, args := range toolCalls {
				if exp.ArgMatcher(args) {
					matching++
				}
			}
			if matching >= minCalls {
				met++
			} else {
				unmet = append(unmet, fmt.Sprintf("%s (want >=%d matching, got %d)", exp.Name, minCalls, matching))
			}
		}
	}

	value := float64(met) / float64(len(tu.expectations))
	if met == len(tu.expectations) {
		return Score{
			Pass:   true,
			Value:  1.0,
			Reason: fmt.Sprintf("all %d tool expectations met", len(tu.expectations)),
		}, nil
	}
	return Score{
		Pass:   false,
		Value:  value,
		Reason: fmt.Sprintf("%d/%d expectations met, unmet: %v", met, len(tu.expectations), unmet),
	}, nil
}

// collectToolCalls extracts all tool calls from the conversation history,
// grouped by function name. Each entry is the raw JSON arguments.
func collectToolCalls(history []llm.Message) map[string][]json.RawMessage {
	calls := make(map[string][]json.RawMessage)
	for _, msg := range history {
		for _, tc := range msg.ToolCalls {
			calls[tc.Function] = append(calls[tc.Function], tc.Arguments)
		}
	}
	return calls
}
