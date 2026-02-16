package eval

import (
	"testing"
	"time"
)

func TestCompletedWithinLimit(t *testing.T) {
	score := CompletedWithin(100*time.Millisecond, 200*time.Millisecond)
	if !score.Pass {
		t.Error("expected Pass = true")
	}
	if score.Value != 1.0 {
		t.Errorf("Value = %f, want 1.0", score.Value)
	}
}

func TestCompletedOverLimit(t *testing.T) {
	score := CompletedWithin(300*time.Millisecond, 200*time.Millisecond)
	if score.Pass {
		t.Error("expected Pass = false")
	}
	// Value should be 200/300 ≈ 0.667
	want := 200.0 / 300.0
	if score.Value < want-0.01 || score.Value > want+0.01 {
		t.Errorf("Value = %f, want ~%f", score.Value, want)
	}
}

func TestCompletedExactlyAtLimit(t *testing.T) {
	score := CompletedWithin(200*time.Millisecond, 200*time.Millisecond)
	if !score.Pass {
		t.Error("expected Pass = true for exact limit")
	}
}
