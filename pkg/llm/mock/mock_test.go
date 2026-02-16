package mock

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/lonestarx1/gogrid/pkg/llm"
)

func resp(content string) *llm.Response {
	return &llm.Response{
		Message: llm.NewAssistantMessage(content),
		Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		Model:   "mock-model",
	}
}

func params(content string) llm.Params {
	return llm.Params{
		Model:    "mock-model",
		Messages: []llm.Message{llm.NewUserMessage(content)},
	}
}

func TestSequentialResponses(t *testing.T) {
	p := New(WithResponses(resp("first"), resp("second"), resp("third")))
	ctx := context.Background()

	tests := []struct {
		want string
	}{
		{"first"},
		{"second"},
		{"third"},
	}
	for i, tt := range tests {
		got, err := p.Complete(ctx, params("q"))
		if err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
		if got.Message.Content != tt.want {
			t.Errorf("call %d: got %q, want %q", i, got.Message.Content, tt.want)
		}
	}
	if p.Calls() != 3 {
		t.Errorf("Calls = %d, want 3", p.Calls())
	}
}

func TestFallbackAfterSequenceExhausted(t *testing.T) {
	p := New(
		WithResponses(resp("first")),
		WithFallback(resp("fallback")),
	)
	ctx := context.Background()

	// Consume the one sequenced response.
	r, _ := p.Complete(ctx, params("q"))
	if r.Message.Content != "first" {
		t.Errorf("got %q, want %q", r.Message.Content, "first")
	}

	// Now should get fallback.
	r, _ = p.Complete(ctx, params("q"))
	if r.Message.Content != "fallback" {
		t.Errorf("got %q, want %q", r.Message.Content, "fallback")
	}
}

func TestDefaultFallback(t *testing.T) {
	p := New()
	ctx := context.Background()

	r, err := p.Complete(ctx, params("q"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Message.Content != "" {
		t.Errorf("default fallback Content = %q, want empty", r.Message.Content)
	}
}

func TestConstantError(t *testing.T) {
	injected := errors.New("rate limited")
	p := New(WithError(injected))
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := p.Complete(ctx, params("q"))
		if !errors.Is(err, injected) {
			t.Errorf("call %d: err = %v, want %v", i, err, injected)
		}
	}
	if p.Calls() != 3 {
		t.Errorf("Calls = %d, want 3", p.Calls())
	}
}

func TestFailCountThenSucceed(t *testing.T) {
	injected := errors.New("transient")
	p := New(
		WithError(injected),
		WithFailCount(2),
		WithResponses(resp("success")),
	)
	ctx := context.Background()

	// First 2 calls fail.
	for i := 0; i < 2; i++ {
		_, err := p.Complete(ctx, params("q"))
		if !errors.Is(err, injected) {
			t.Errorf("call %d: err = %v, want %v", i, err, injected)
		}
	}

	// Third call succeeds.
	r, err := p.Complete(ctx, params("q"))
	if err != nil {
		t.Fatalf("call 3: unexpected error: %v", err)
	}
	if r.Message.Content != "success" {
		t.Errorf("call 3: got %q, want %q", r.Message.Content, "success")
	}
}

func TestFailCountWithDefaultError(t *testing.T) {
	p := New(
		WithFailCount(1),
		WithResponses(resp("ok")),
	)
	ctx := context.Background()

	_, err := p.Complete(ctx, params("q"))
	if err == nil {
		t.Fatal("expected error on first call")
	}

	r, err := p.Complete(ctx, params("q"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Message.Content != "ok" {
		t.Errorf("got %q, want %q", r.Message.Content, "ok")
	}
}

func TestDelay(t *testing.T) {
	p := New(
		WithDelay(50*time.Millisecond),
		WithResponses(resp("delayed")),
	)
	ctx := context.Background()

	start := time.Now()
	r, err := p.Complete(ctx, params("q"))
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Message.Content != "delayed" {
		t.Errorf("got %q, want %q", r.Message.Content, "delayed")
	}
	if elapsed < 40*time.Millisecond {
		t.Errorf("elapsed = %v, want >= 40ms", elapsed)
	}
}

func TestDelayRespectsContextCancellation(t *testing.T) {
	p := New(WithDelay(5 * time.Second))
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, err := p.Complete(ctx, params("q"))
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestCallback(t *testing.T) {
	var captured llm.Params
	p := New(
		WithCallback(func(params llm.Params) {
			captured = params
		}),
	)
	ctx := context.Background()

	input := params("capture me")
	_, _ = p.Complete(ctx, input)

	if captured.Model != "mock-model" {
		t.Errorf("captured Model = %q, want %q", captured.Model, "mock-model")
	}
	if len(captured.Messages) != 1 || captured.Messages[0].Content != "capture me" {
		t.Errorf("captured Messages = %v, want single user message", captured.Messages)
	}
}

func TestHistory(t *testing.T) {
	p := New()
	ctx := context.Background()

	_, _ = p.Complete(ctx, params("first"))
	_, _ = p.Complete(ctx, params("second"))

	history := p.History()
	if len(history) != 2 {
		t.Fatalf("History len = %d, want 2", len(history))
	}
	if history[0].Messages[0].Content != "first" {
		t.Errorf("history[0] = %q, want %q", history[0].Messages[0].Content, "first")
	}
	if history[1].Messages[0].Content != "second" {
		t.Errorf("history[1] = %q, want %q", history[1].Messages[0].Content, "second")
	}
}

func TestReset(t *testing.T) {
	p := New(WithResponses(resp("a"), resp("b")))
	ctx := context.Background()

	_, _ = p.Complete(ctx, params("q"))
	_, _ = p.Complete(ctx, params("q"))

	p.Reset()

	if p.Calls() != 0 {
		t.Errorf("after Reset: Calls = %d, want 0", p.Calls())
	}
	if len(p.History()) != 0 {
		t.Errorf("after Reset: History len = %d, want 0", len(p.History()))
	}
}

func TestConcurrentSafety(t *testing.T) {
	p := New(WithFallback(resp("concurrent")))
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := p.Complete(ctx, params("q"))
			if err != nil {
				t.Errorf("concurrent call error: %v", err)
				return
			}
			if r.Message.Content != "concurrent" {
				t.Errorf("concurrent call: got %q, want %q", r.Message.Content, "concurrent")
			}
		}()
	}
	wg.Wait()

	if p.Calls() != 100 {
		t.Errorf("Calls = %d, want 100", p.Calls())
	}
	if len(p.History()) != 100 {
		t.Errorf("History len = %d, want 100", len(p.History()))
	}
}
