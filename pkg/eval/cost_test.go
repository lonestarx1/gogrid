package eval

import (
	"context"
	"testing"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
)

func resultWithCost(cost float64) *agent.Result {
	return &agent.Result{
		RunID:   "test-run",
		Message: llm.NewAssistantMessage("ok"),
		Cost:    cost,
	}
}

func TestCostWithinBudget(t *testing.T) {
	ev := NewCostWithin(1.0)
	score, err := ev.Evaluate(context.Background(), resultWithCost(0.50))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !score.Pass {
		t.Error("expected Pass = true")
	}
	if score.Value != 1.0 {
		t.Errorf("Value = %f, want 1.0", score.Value)
	}
}

func TestCostOverBudget(t *testing.T) {
	ev := NewCostWithin(1.0)
	score, err := ev.Evaluate(context.Background(), resultWithCost(2.0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Pass {
		t.Error("expected Pass = false")
	}
	if score.Value != 0.5 {
		t.Errorf("Value = %f, want 0.5", score.Value)
	}
}

func TestCostZeroCost(t *testing.T) {
	ev := NewCostWithin(1.0)
	score, err := ev.Evaluate(context.Background(), resultWithCost(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !score.Pass {
		t.Error("expected Pass = true for zero cost")
	}
}

func TestCostZeroBudget(t *testing.T) {
	ev := NewCostWithin(0)
	score, err := ev.Evaluate(context.Background(), resultWithCost(0.001))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Pass {
		t.Error("expected Pass = false when budget is zero and cost > 0")
	}
}

func TestCostZeroBudgetZeroCost(t *testing.T) {
	ev := NewCostWithin(0)
	score, err := ev.Evaluate(context.Background(), resultWithCost(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !score.Pass {
		t.Error("expected Pass = true when both budget and cost are zero")
	}
}
