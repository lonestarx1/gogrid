package eval

import (
	"context"
	"errors"
	"testing"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
)

func testResult(content string) *agent.Result {
	return &agent.Result{
		RunID:   "test-run",
		Message: llm.NewAssistantMessage(content),
		History: []llm.Message{
			llm.NewUserMessage("input"),
			llm.NewAssistantMessage(content),
		},
		Usage: llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		Cost:  0.001,
		Turns: 1,
	}
}

func TestFuncEvaluator(t *testing.T) {
	ev := NewFunc("always-pass", func(_ context.Context, _ *agent.Result) (Score, error) {
		return Score{Pass: true, Value: 1.0, Reason: "always passes"}, nil
	})

	if ev.Name() != "always-pass" {
		t.Errorf("Name = %q, want %q", ev.Name(), "always-pass")
	}

	score, err := ev.Evaluate(context.Background(), testResult("hello"))
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

func TestSuiteAllPass(t *testing.T) {
	suite := NewSuite(
		NewFunc("a", func(_ context.Context, _ *agent.Result) (Score, error) {
			return Score{Pass: true, Value: 1.0, Reason: "a"}, nil
		}),
		NewFunc("b", func(_ context.Context, _ *agent.Result) (Score, error) {
			return Score{Pass: true, Value: 0.8, Reason: "b"}, nil
		}),
	)

	sr, err := suite.Run(context.Background(), testResult("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sr.Pass {
		t.Error("expected suite Pass = true")
	}
	if len(sr.Scores) != 2 {
		t.Errorf("Scores len = %d, want 2", len(sr.Scores))
	}
	if len(sr.Errors) != 0 {
		t.Errorf("Errors len = %d, want 0", len(sr.Errors))
	}
}

func TestSuiteSomeFail(t *testing.T) {
	suite := NewSuite(
		NewFunc("pass", func(_ context.Context, _ *agent.Result) (Score, error) {
			return Score{Pass: true, Value: 1.0, Reason: "ok"}, nil
		}),
		NewFunc("fail", func(_ context.Context, _ *agent.Result) (Score, error) {
			return Score{Pass: false, Value: 0.2, Reason: "bad"}, nil
		}),
	)

	sr, err := suite.Run(context.Background(), testResult("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sr.Pass {
		t.Error("expected suite Pass = false")
	}
	if !sr.Scores["pass"].Pass {
		t.Error("expected 'pass' evaluator to pass")
	}
	if sr.Scores["fail"].Pass {
		t.Error("expected 'fail' evaluator to fail")
	}
}

func TestSuiteWithErrors(t *testing.T) {
	evalErr := errors.New("evaluation failed")
	suite := NewSuite(
		NewFunc("pass", func(_ context.Context, _ *agent.Result) (Score, error) {
			return Score{Pass: true, Value: 1.0, Reason: "ok"}, nil
		}),
		NewFunc("error", func(_ context.Context, _ *agent.Result) (Score, error) {
			return Score{}, evalErr
		}),
	)

	sr, err := suite.Run(context.Background(), testResult("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sr.Pass {
		t.Error("expected suite Pass = false when evaluator errors")
	}
	if !errors.Is(sr.Errors["error"], evalErr) {
		t.Errorf("Errors[error] = %v, want %v", sr.Errors["error"], evalErr)
	}
	if _, ok := sr.Scores["error"]; ok {
		t.Error("errored evaluator should not have a score entry")
	}
}
