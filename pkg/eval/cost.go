package eval

import (
	"context"
	"fmt"

	"github.com/lonestarx1/gogrid/pkg/agent"
)

// CostWithin evaluates whether the agent's run cost stayed within
// a USD budget.
type CostWithin struct {
	maxCostUSD float64
}

// NewCostWithin creates a CostWithin evaluator with the given budget.
func NewCostWithin(maxCostUSD float64) *CostWithin {
	return &CostWithin{maxCostUSD: maxCostUSD}
}

// Name returns "cost_within".
func (c *CostWithin) Name() string { return "cost_within" }

// Evaluate checks if result.Cost <= maxCostUSD.
func (c *CostWithin) Evaluate(_ context.Context, result *agent.Result) (Score, error) {
	if result.Cost <= c.maxCostUSD {
		return Score{
			Pass:   true,
			Value:  1.0,
			Reason: fmt.Sprintf("cost $%.6f within budget $%.6f", result.Cost, c.maxCostUSD),
		}, nil
	}
	// Value is the fraction of budget used (capped at 0).
	var value float64
	if c.maxCostUSD > 0 {
		value = c.maxCostUSD / result.Cost
	}
	return Score{
		Pass:   false,
		Value:  value,
		Reason: fmt.Sprintf("cost $%.6f exceeds budget $%.6f", result.Cost, c.maxCostUSD),
	}, nil
}
