package eval

import (
	"context"
	"testing"
)

func TestContainsAllFound(t *testing.T) {
	ev := NewContains("hello", "world")
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

func TestContainsSomeMissing(t *testing.T) {
	ev := NewContains("hello", "world", "missing")
	score, err := ev.Evaluate(context.Background(), testResult("hello world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Pass {
		t.Error("expected Pass = false")
	}
	// 2 out of 3 found.
	want := 2.0 / 3.0
	if score.Value < want-0.01 || score.Value > want+0.01 {
		t.Errorf("Value = %f, want ~%f", score.Value, want)
	}
}

func TestContainsNoneFound(t *testing.T) {
	ev := NewContains("foo", "bar")
	score, err := ev.Evaluate(context.Background(), testResult("hello world"))
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

func TestContainsEmptySubstrings(t *testing.T) {
	ev := NewContains()
	score, err := ev.Evaluate(context.Background(), testResult("anything"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !score.Pass {
		t.Error("expected Pass = true for empty substrings")
	}
}
