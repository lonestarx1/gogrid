package eval

import (
	"context"
	"testing"
)

func TestExactMatchPass(t *testing.T) {
	ev := NewExactMatch("hello world")
	score, err := ev.Evaluate(context.Background(), testResult("hello world"))
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

func TestExactMatchFail(t *testing.T) {
	ev := NewExactMatch("hello world")
	score, err := ev.Evaluate(context.Background(), testResult("hello"))
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

func TestExactMatchName(t *testing.T) {
	ev := NewExactMatch("x")
	if ev.Name() != "exact_match" {
		t.Errorf("Name = %q, want %q", ev.Name(), "exact_match")
	}
}
