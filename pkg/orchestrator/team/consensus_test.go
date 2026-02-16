package team

import "testing"

func TestUnanimousEvaluate(t *testing.T) {
	s := Unanimous{}

	tests := []struct {
		name      string
		total     int
		responses map[string]string
		reached   bool
	}{
		{
			name:      "no responses",
			total:     3,
			responses: map[string]string{},
			reached:   false,
		},
		{
			name:      "partial responses",
			total:     3,
			responses: map[string]string{"a": "yes", "b": "no"},
			reached:   false,
		},
		{
			name:      "all responded",
			total:     3,
			responses: map[string]string{"a": "yes", "b": "no", "c": "maybe"},
			reached:   true,
		},
		{
			name:      "single agent",
			total:     1,
			responses: map[string]string{"a": "done"},
			reached:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, reached := s.Evaluate(tt.total, tt.responses)
			if reached != tt.reached {
				t.Errorf("reached = %v, want %v", reached, tt.reached)
			}
			if reached && decision == "" {
				t.Error("consensus reached but decision is empty")
			}
		})
	}
}

func TestUnanimousDecisionOrder(t *testing.T) {
	s := Unanimous{}
	responses := map[string]string{
		"charlie": "c-response",
		"alice":   "a-response",
		"bob":     "b-response",
	}

	decision, reached := s.Evaluate(3, responses)
	if !reached {
		t.Fatal("expected consensus")
	}

	// Should be alphabetical: alice, bob, charlie.
	want := "alice: a-response\n\nbob: b-response\n\ncharlie: c-response"
	if decision != want {
		t.Errorf("decision =\n%s\nwant:\n%s", decision, want)
	}
}

func TestMajorityEvaluate(t *testing.T) {
	s := Majority{}

	tests := []struct {
		name      string
		total     int
		responses map[string]string
		reached   bool
	}{
		{
			name:      "no responses",
			total:     3,
			responses: map[string]string{},
			reached:   false,
		},
		{
			name:      "exactly half (3 agents, 1 response)",
			total:     3,
			responses: map[string]string{"a": "yes"},
			reached:   false,
		},
		{
			name:      "majority (3 agents, 2 responses)",
			total:     3,
			responses: map[string]string{"a": "yes", "b": "no"},
			reached:   true,
		},
		{
			name:      "all responded",
			total:     3,
			responses: map[string]string{"a": "yes", "b": "no", "c": "maybe"},
			reached:   true,
		},
		{
			name:      "even split (4 agents, 2 responses)",
			total:     4,
			responses: map[string]string{"a": "yes", "b": "no"},
			reached:   false,
		},
		{
			name:      "majority of 4 (3 responses)",
			total:     4,
			responses: map[string]string{"a": "yes", "b": "no", "c": "maybe"},
			reached:   true,
		},
		{
			name:      "single agent",
			total:     1,
			responses: map[string]string{"a": "done"},
			reached:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, reached := s.Evaluate(tt.total, tt.responses)
			if reached != tt.reached {
				t.Errorf("reached = %v, want %v", reached, tt.reached)
			}
			if reached && decision == "" {
				t.Error("consensus reached but decision is empty")
			}
		})
	}
}

func TestFirstResponseEvaluate(t *testing.T) {
	s := FirstResponse{}

	tests := []struct {
		name      string
		responses map[string]string
		reached   bool
	}{
		{
			name:      "no responses",
			responses: map[string]string{},
			reached:   false,
		},
		{
			name:      "one response",
			responses: map[string]string{"a": "fast answer"},
			reached:   true,
		},
		{
			name:      "multiple responses",
			responses: map[string]string{"a": "first", "b": "second"},
			reached:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, reached := s.Evaluate(3, tt.responses)
			if reached != tt.reached {
				t.Errorf("reached = %v, want %v", reached, tt.reached)
			}
			if reached && decision == "" {
				t.Error("consensus reached but decision is empty")
			}
		})
	}
}

func TestStrategyNames(t *testing.T) {
	tests := []struct {
		strategy Strategy
		want     string
	}{
		{Unanimous{}, "unanimous"},
		{Majority{}, "majority"},
		{FirstResponse{}, "first_response"},
	}

	for _, tt := range tests {
		if got := tt.strategy.Name(); got != tt.want {
			t.Errorf("%T.Name() = %q, want %q", tt.strategy, got, tt.want)
		}
	}
}
