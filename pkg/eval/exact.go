package eval

import (
	"context"
	"fmt"

	"github.com/lonestarx1/gogrid/pkg/agent"
)

// ExactMatch evaluates whether the agent's final output exactly matches
// the expected string.
type ExactMatch struct {
	expected string
}

// NewExactMatch creates an ExactMatch evaluator for the given string.
func NewExactMatch(expected string) *ExactMatch {
	return &ExactMatch{expected: expected}
}

// Name returns "exact_match".
func (e *ExactMatch) Name() string { return "exact_match" }

// Evaluate checks if result.Message.Content exactly equals the expected string.
func (e *ExactMatch) Evaluate(_ context.Context, result *agent.Result) (Score, error) {
	if result.Message.Content == e.expected {
		return Score{
			Pass:   true,
			Value:  1.0,
			Reason: "output matches expected string",
		}, nil
	}
	return Score{
		Pass:   false,
		Value:  0.0,
		Reason: fmt.Sprintf("output %q does not match expected %q", result.Message.Content, e.expected),
	}, nil
}
