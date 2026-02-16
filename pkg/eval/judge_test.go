package eval

import (
	"context"
	"testing"

	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/llm/mock"
)

func judgeResp(content string) *llm.Response {
	return &llm.Response{
		Message: llm.NewAssistantMessage(content),
		Usage:   llm.Usage{PromptTokens: 50, CompletionTokens: 20, TotalTokens: 70},
		Model:   "judge-model",
	}
}

func TestLLMJudgeGoodScore(t *testing.T) {
	provider := mock.New(mock.WithResponses(
		judgeResp("SCORE: 9\nREASON: Excellent output"),
	))

	ev := NewLLMJudge(provider, "judge-model", "Be helpful and accurate")
	score, err := ev.Evaluate(context.Background(), testResult("The answer is 42"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !score.Pass {
		t.Error("expected Pass = true for score 9")
	}
	if score.Value != 0.9 {
		t.Errorf("Value = %f, want 0.9", score.Value)
	}
	if score.Reason != "Excellent output" {
		t.Errorf("Reason = %q, want %q", score.Reason, "Excellent output")
	}
}

func TestLLMJudgeLowScore(t *testing.T) {
	provider := mock.New(mock.WithResponses(
		judgeResp("SCORE: 3\nREASON: Incomplete and inaccurate"),
	))

	ev := NewLLMJudge(provider, "judge-model", "Be helpful")
	score, err := ev.Evaluate(context.Background(), testResult("idk"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Pass {
		t.Error("expected Pass = false for score 3")
	}
	if score.Value != 0.3 {
		t.Errorf("Value = %f, want 0.3", score.Value)
	}
}

func TestLLMJudgeBorderlineScore(t *testing.T) {
	provider := mock.New(mock.WithResponses(
		judgeResp("SCORE: 7\nREASON: Adequate"),
	))

	ev := NewLLMJudge(provider, "judge-model", "Be helpful")
	score, err := ev.Evaluate(context.Background(), testResult("decent answer"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !score.Pass {
		t.Error("expected Pass = true for score 7 (threshold)")
	}
}

func TestLLMJudgeParseError(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"no score line", "This is just text\nREASON: something"},
		{"invalid score number", "SCORE: abc\nREASON: bad"},
		{"score out of range", "SCORE: 15\nREASON: too high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := mock.New(mock.WithResponses(judgeResp(tt.content)))
			ev := NewLLMJudge(provider, "judge-model", "rubric")
			_, err := ev.Evaluate(context.Background(), testResult("output"))
			if err == nil {
				t.Error("expected error for malformed response")
			}
		})
	}
}

func TestLLMJudgeCaseInsensitive(t *testing.T) {
	provider := mock.New(mock.WithResponses(
		judgeResp("score: 8\nreason: good work"),
	))

	ev := NewLLMJudge(provider, "judge-model", "rubric")
	score, err := ev.Evaluate(context.Background(), testResult("output"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !score.Pass {
		t.Error("expected Pass = true")
	}
	if score.Reason != "good work" {
		t.Errorf("Reason = %q, want %q", score.Reason, "good work")
	}
}

func TestLLMJudgeCustomPromptTemplate(t *testing.T) {
	var captured llm.Params
	provider := mock.New(
		mock.WithResponses(judgeResp("SCORE: 8\nREASON: ok")),
		mock.WithCallback(func(p llm.Params) { captured = p }),
	)

	ev := NewLLMJudge(provider, "judge-model", "my rubric").
		WithPromptTemplate("Custom: {rubric} | Output: {output}")

	_, _ = ev.Evaluate(context.Background(), testResult("test output"))

	msg := captured.Messages[0].Content
	if msg != "Custom: my rubric | Output: test output" {
		t.Errorf("prompt = %q, want custom template applied", msg)
	}
}

func TestLLMJudgeProviderError(t *testing.T) {
	provider := mock.New(mock.WithError(context.DeadlineExceeded))
	ev := NewLLMJudge(provider, "judge-model", "rubric")
	_, err := ev.Evaluate(context.Background(), testResult("output"))
	if err == nil {
		t.Error("expected error when provider fails")
	}
}

func TestLLMJudgeMissingReason(t *testing.T) {
	provider := mock.New(mock.WithResponses(
		judgeResp("SCORE: 8"),
	))

	ev := NewLLMJudge(provider, "judge-model", "rubric")
	score, err := ev.Evaluate(context.Background(), testResult("output"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Reason != "no reason provided" {
		t.Errorf("Reason = %q, want %q", score.Reason, "no reason provided")
	}
}

func TestLLMJudgeName(t *testing.T) {
	ev := NewLLMJudge(nil, "", "")
	if ev.Name() != "llm_judge" {
		t.Errorf("Name = %q, want %q", ev.Name(), "llm_judge")
	}
}
