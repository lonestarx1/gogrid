package eval

import (
	"context"
	"fmt"
	"strings"

	"github.com/lonestarx1/gogrid/pkg/agent"
)

// Contains evaluates whether the agent's output contains all of the
// expected substrings. Value reflects the fraction of substrings found.
type Contains struct {
	substrings []string
}

// NewContains creates a Contains evaluator that checks for the
// presence of all given substrings.
func NewContains(substrings ...string) *Contains {
	return &Contains{substrings: substrings}
}

// Name returns "contains".
func (c *Contains) Name() string { return "contains" }

// Evaluate checks how many of the expected substrings are present.
func (c *Contains) Evaluate(_ context.Context, result *agent.Result) (Score, error) {
	if len(c.substrings) == 0 {
		return Score{Pass: true, Value: 1.0, Reason: "no substrings to check"}, nil
	}

	content := result.Message.Content
	found := 0
	var missing []string
	for _, sub := range c.substrings {
		if strings.Contains(content, sub) {
			found++
		} else {
			missing = append(missing, sub)
		}
	}

	value := float64(found) / float64(len(c.substrings))
	if found == len(c.substrings) {
		return Score{
			Pass:   true,
			Value:  1.0,
			Reason: fmt.Sprintf("all %d substrings found", len(c.substrings)),
		}, nil
	}
	return Score{
		Pass:   false,
		Value:  value,
		Reason: fmt.Sprintf("%d/%d substrings found, missing: %v", found, len(c.substrings), missing),
	}, nil
}
