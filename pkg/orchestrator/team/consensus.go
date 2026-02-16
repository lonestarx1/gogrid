package team

import (
	"sort"
	"strings"
)

// Strategy determines when a team has reached consensus and how to form
// the final decision from agent responses.
type Strategy interface {
	// Name returns the strategy name for tracing and logging.
	Name() string
	// Evaluate checks whether consensus has been reached.
	// total is the number of team members. responses maps agent name to
	// that agent's response content for the current round.
	// Returns the decision text and whether consensus was reached.
	Evaluate(total int, responses map[string]string) (decision string, reached bool)
}

// Unanimous reaches consensus when all agents have responded.
// The decision concatenates all responses in alphabetical order by agent name.
type Unanimous struct{}

// Name returns "unanimous".
func (u Unanimous) Name() string { return "unanimous" }

// Evaluate returns consensus when all agents have responded.
func (u Unanimous) Evaluate(total int, responses map[string]string) (string, bool) {
	if len(responses) < total {
		return "", false
	}
	return combineResponses(responses), true
}

// Majority reaches consensus when more than half of the agents have responded.
// The decision concatenates available responses in alphabetical order.
type Majority struct{}

// Name returns "majority".
func (m Majority) Name() string { return "majority" }

// Evaluate returns consensus when more than half of agents have responded.
func (m Majority) Evaluate(total int, responses map[string]string) (string, bool) {
	if len(responses) > total/2 {
		return combineResponses(responses), true
	}
	return "", false
}

// FirstResponse reaches consensus as soon as any agent responds.
// The decision is the first agent's response.
type FirstResponse struct{}

// Name returns "first_response".
func (f FirstResponse) Name() string { return "first_response" }

// Evaluate returns consensus as soon as any response is available.
func (f FirstResponse) Evaluate(_ int, responses map[string]string) (string, bool) {
	for _, content := range responses {
		return content, true
	}
	return "", false
}

// combineResponses joins all responses in alphabetical order by agent name.
func combineResponses(responses map[string]string) string {
	names := make([]string, 0, len(responses))
	for name := range responses {
		names = append(names, name)
	}
	sort.Strings(names)

	var parts []string
	for _, name := range names {
		parts = append(parts, name+": "+responses[name])
	}
	return strings.Join(parts, "\n\n")
}
