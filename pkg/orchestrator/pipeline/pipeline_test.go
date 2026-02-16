package pipeline

import (
	"context"
	"errors"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/trace"
)

// mockProvider returns a fixed response.
type mockProvider struct {
	response *llm.Response
	calls    atomic.Int32
}

func newMockProvider(content string) *mockProvider {
	return &mockProvider{
		response: &llm.Response{
			Message: llm.NewAssistantMessage(content),
			Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
			Model:   "gpt-4o",
		},
	}
}

func (m *mockProvider) Complete(_ context.Context, _ llm.Params) (*llm.Response, error) {
	m.calls.Add(1)
	return m.response, nil
}

// errorProvider always returns an error.
type errorProvider struct{ err error }

func (e *errorProvider) Complete(_ context.Context, _ llm.Params) (*llm.Response, error) {
	return nil, e.err
}

// countingErrorProvider fails the first N times then succeeds.
type countingErrorProvider struct {
	failCount int
	calls     atomic.Int32
	response  *llm.Response
}

func (c *countingErrorProvider) Complete(_ context.Context, _ llm.Params) (*llm.Response, error) {
	n := int(c.calls.Add(1))
	if n <= c.failCount {
		return nil, errors.New("transient error")
	}
	return c.response, nil
}

// slowProvider adds a delay before responding.
type slowProvider struct {
	delay    time.Duration
	response *llm.Response
}

func (s *slowProvider) Complete(ctx context.Context, _ llm.Params) (*llm.Response, error) {
	select {
	case <-time.After(s.delay):
		return s.response, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func newTestAgent(name, content string) *agent.Agent {
	return agent.New(name,
		agent.WithProvider(newMockProvider(content)),
		agent.WithModel("gpt-4o"),
	)
}

func TestNewPipeline(t *testing.T) {
	p := New("test-pipeline")
	if p.Name() != "test-pipeline" {
		t.Errorf("Name = %q, want %q", p.Name(), "test-pipeline")
	}
}

func TestRunRequiresStages(t *testing.T) {
	p := New("empty")
	_, err := p.Run(context.Background(), "hello")
	if err == nil {
		t.Fatal("Run without stages: expected error")
	}
}

func TestRunSingleStage(t *testing.T) {
	p := New("single",
		WithStages(Stage{
			Name:  "process",
			Agent: newTestAgent("processor", "processed output"),
		}),
	)

	result, err := p.Run(context.Background(), "raw input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.RunID == "" {
		t.Error("RunID is empty")
	}
	if result.Output != "processed output" {
		t.Errorf("Output = %q, want %q", result.Output, "processed output")
	}
	if len(result.Stages) != 1 {
		t.Errorf("Stages len = %d, want 1", len(result.Stages))
	}
	if result.Stages[0].Name != "process" {
		t.Errorf("Stage name = %q, want %q", result.Stages[0].Name, "process")
	}
}

func TestRunMultipleStages(t *testing.T) {
	p := New("three-stage",
		WithStages(
			Stage{Name: "collect", Agent: newTestAgent("collector", "collected data")},
			Stage{Name: "analyze", Agent: newTestAgent("analyzer", "analysis result")},
			Stage{Name: "summarize", Agent: newTestAgent("summarizer", "final summary")},
		),
	)

	result, err := p.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Output != "final summary" {
		t.Errorf("Output = %q, want %q", result.Output, "final summary")
	}
	if len(result.Stages) != 3 {
		t.Errorf("Stages len = %d, want 3", len(result.Stages))
	}
}

func TestRunStateTransfer(t *testing.T) {
	p := New("transfer",
		WithStages(
			Stage{Name: "stage-1", Agent: newTestAgent("a", "output-1")},
			Stage{Name: "stage-2", Agent: newTestAgent("b", "output-2")},
			Stage{Name: "stage-3", Agent: newTestAgent("c", "output-3")},
		),
	)

	result, err := p.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Should have transfer log: acquire stage-1, transfer to stage-2, transfer to stage-3.
	if len(result.TransferLog) != 3 {
		t.Fatalf("TransferLog len = %d, want 3", len(result.TransferLog))
	}
	if result.TransferLog[0].To != "stage-1" {
		t.Errorf("Transfer[0].To = %q, want %q", result.TransferLog[0].To, "stage-1")
	}
	if result.TransferLog[1].From != "stage-1" || result.TransferLog[1].To != "stage-2" {
		t.Errorf("Transfer[1] = %q->%q, want stage-1->stage-2",
			result.TransferLog[1].From, result.TransferLog[1].To)
	}
	if result.TransferLog[2].From != "stage-2" || result.TransferLog[2].To != "stage-3" {
		t.Errorf("Transfer[2] = %q->%q, want stage-2->stage-3",
			result.TransferLog[2].From, result.TransferLog[2].To)
	}
}

func TestRunInputTransform(t *testing.T) {
	provider := newMockProvider("echoed")
	a := agent.New("echo",
		agent.WithProvider(provider),
		agent.WithModel("gpt-4o"),
	)

	p := New("transform",
		WithStages(
			Stage{Name: "first", Agent: newTestAgent("a", "raw")},
			Stage{
				Name:  "second",
				Agent: a,
				InputTransform: func(input string) string {
					return "TRANSFORMED: " + input
				},
			},
		),
	)

	_, err := p.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// The provider was called, meaning the agent ran with transformed input.
	if provider.calls.Load() != 1 {
		t.Errorf("provider calls = %d, want 1", provider.calls.Load())
	}
}

func TestRunOutputValidation(t *testing.T) {
	p := New("validate",
		WithStages(Stage{
			Name:  "validated",
			Agent: newTestAgent("a", "bad output"),
			OutputValidate: func(output string) error {
				if output == "bad output" {
					return errors.New("output is invalid")
				}
				return nil
			},
		}),
	)

	_, err := p.Run(context.Background(), "input")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRunOutputValidationPass(t *testing.T) {
	p := New("validate-pass",
		WithStages(Stage{
			Name:  "validated",
			Agent: newTestAgent("a", "good output"),
			OutputValidate: func(output string) error {
				if output == "good output" {
					return nil
				}
				return errors.New("unexpected")
			},
		}),
	)

	result, err := p.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Output != "good output" {
		t.Errorf("Output = %q, want %q", result.Output, "good output")
	}
}

func TestRunRetrySuccess(t *testing.T) {
	provider := &countingErrorProvider{
		failCount: 2,
		response: &llm.Response{
			Message: llm.NewAssistantMessage("success after retries"),
			Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
			Model:   "gpt-4o",
		},
	}
	a := agent.New("retryable",
		agent.WithProvider(provider),
		agent.WithModel("gpt-4o"),
	)

	p := New("retry",
		WithStages(Stage{
			Name:  "flaky",
			Agent: a,
			Retry: RetryPolicy{MaxAttempts: 3, Delay: time.Millisecond},
		}),
	)

	result, err := p.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Output != "success after retries" {
		t.Errorf("Output = %q, want %q", result.Output, "success after retries")
	}
	if result.Stages[0].Attempts != 3 {
		t.Errorf("Attempts = %d, want 3", result.Stages[0].Attempts)
	}
}

func TestRunRetryExhausted(t *testing.T) {
	a := agent.New("fail",
		agent.WithProvider(&errorProvider{err: errors.New("always fails")}),
		agent.WithModel("gpt-4o"),
	)

	p := New("retry-exhaust",
		WithStages(Stage{
			Name:  "fail",
			Agent: a,
			Retry: RetryPolicy{MaxAttempts: 3, Delay: time.Millisecond},
		}),
	)

	_, err := p.Run(context.Background(), "input")
	if err == nil {
		t.Fatal("expected error after retry exhaustion")
	}
}

func TestRunErrorActionSkip(t *testing.T) {
	bad := agent.New("bad",
		agent.WithProvider(&errorProvider{err: errors.New("fails")}),
		agent.WithModel("gpt-4o"),
	)

	p := New("skip-error",
		WithStages(
			Stage{Name: "first", Agent: newTestAgent("a", "from first")},
			Stage{Name: "flaky", Agent: bad, OnError: Skip},
			Stage{Name: "final", Agent: newTestAgent("c", "final output")},
		),
	)

	result, err := p.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Output != "final output" {
		t.Errorf("Output = %q, want %q", result.Output, "final output")
	}
	if !result.Stages[1].Skipped {
		t.Error("stage 1 should be skipped")
	}
	if len(result.Stages) != 3 {
		t.Errorf("Stages len = %d, want 3", len(result.Stages))
	}
}

func TestRunErrorActionAbort(t *testing.T) {
	bad := agent.New("bad",
		agent.WithProvider(&errorProvider{err: errors.New("fails")}),
		agent.WithModel("gpt-4o"),
	)

	p := New("abort-error",
		WithStages(
			Stage{Name: "first", Agent: newTestAgent("a", "ok")},
			Stage{Name: "fail", Agent: bad, OnError: Abort},
			Stage{Name: "never", Agent: newTestAgent("c", "never reached")},
		),
	)

	_, err := p.Run(context.Background(), "input")
	if err == nil {
		t.Fatal("expected abort error")
	}
}

func TestRunTracing(t *testing.T) {
	tracer := trace.NewInMemory()

	p := New("traced",
		WithStages(
			Stage{Name: "s1", Agent: newTestAgent("a", "resp")},
			Stage{Name: "s2", Agent: newTestAgent("b", "resp")},
		),
		WithTracer(tracer),
	)

	_, err := p.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	spans := tracer.Spans()
	names := make(map[string]bool)
	for _, s := range spans {
		names[s.Name] = true
	}
	if !names["pipeline.run"] {
		t.Error("missing pipeline.run span")
	}
	if !names["pipeline.stage"] {
		t.Error("missing pipeline.stage span")
	}
}

func TestRunTracerAttributes(t *testing.T) {
	tracer := trace.NewInMemory()

	p := New("attrs",
		WithStages(
			Stage{Name: "only", Agent: newTestAgent("a", "resp")},
		),
		WithTracer(tracer),
	)

	_, err := p.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	for _, s := range tracer.Spans() {
		if s.Name == "pipeline.run" {
			if s.Attributes["pipeline.name"] != "attrs" {
				t.Errorf("pipeline.name = %q, want %q", s.Attributes["pipeline.name"], "attrs")
			}
			if s.Attributes["pipeline.stages"] != "1" {
				t.Errorf("pipeline.stages = %q, want %q", s.Attributes["pipeline.stages"], "1")
			}
		}
		if s.Name == "pipeline.stage" {
			if s.Attributes["pipeline.stage.name"] != "only" {
				t.Errorf("stage.name = %q, want %q", s.Attributes["pipeline.stage.name"], "only")
			}
		}
	}
}

func TestRunPipelineTimeout(t *testing.T) {
	slow := agent.New("slow",
		agent.WithProvider(&slowProvider{
			delay: 10 * time.Second,
			response: &llm.Response{
				Message: llm.NewAssistantMessage("slow"),
				Usage:   llm.Usage{TotalTokens: 10},
				Model:   "gpt-4o",
			},
		}),
		agent.WithModel("gpt-4o"),
	)

	p := New("timeout",
		WithStages(Stage{Name: "slow", Agent: slow}),
		WithConfig(Config{Timeout: 50 * time.Millisecond}),
	)

	_, err := p.Run(context.Background(), "input")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRunStageTimeout(t *testing.T) {
	slow := agent.New("slow",
		agent.WithProvider(&slowProvider{
			delay: 10 * time.Second,
			response: &llm.Response{
				Message: llm.NewAssistantMessage("slow"),
				Usage:   llm.Usage{TotalTokens: 10},
				Model:   "gpt-4o",
			},
		}),
		agent.WithModel("gpt-4o"),
	)

	p := New("stage-timeout",
		WithStages(Stage{
			Name:    "slow",
			Agent:   slow,
			Timeout: 50 * time.Millisecond,
		}),
	)

	_, err := p.Run(context.Background(), "input")
	if err == nil {
		t.Fatal("expected stage timeout error")
	}
}

func TestRunContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := New("canceled",
		WithStages(Stage{Name: "a", Agent: newTestAgent("a", "ok")}),
	)

	_, err := p.Run(ctx, "hello")
	if err == nil {
		t.Fatal("expected context canceled error")
	}
}

func TestRunCostBudget(t *testing.T) {
	p := New("budget",
		WithStages(
			Stage{Name: "s1", Agent: newTestAgent("a", "cheap")},
			Stage{Name: "s2", Agent: newTestAgent("b", "cheap")},
			Stage{Name: "s3", Agent: newTestAgent("c", "never")},
		),
		WithConfig(Config{CostBudget: 0.000001}),
	)

	result, err := p.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Should stop early after cost exceeds tiny budget.
	if len(result.Stages) >= 3 {
		t.Errorf("Stages = %d, expected < 3 with tiny budget", len(result.Stages))
	}
}

func TestRunStageCostBudget(t *testing.T) {
	p := New("stage-budget",
		WithStages(Stage{
			Name:       "expensive",
			Agent:      newTestAgent("a", "output"),
			CostBudget: 0.000001, // Tiny budget, will be exceeded.
		}),
	)

	_, err := p.Run(context.Background(), "input")
	if err == nil {
		t.Fatal("expected stage cost budget error")
	}
}

func TestRunAggregatesUsage(t *testing.T) {
	p := New("usage",
		WithStages(
			Stage{Name: "s1", Agent: newTestAgent("a", "resp")},
			Stage{Name: "s2", Agent: newTestAgent("b", "resp")},
		),
	)

	result, err := p.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Each mock: 10 prompt + 5 completion = 15 total. Two stages.
	if result.TotalUsage.PromptTokens != 20 {
		t.Errorf("PromptTokens = %d, want 20", result.TotalUsage.PromptTokens)
	}
	if result.TotalUsage.CompletionTokens != 10 {
		t.Errorf("CompletionTokens = %d, want 10", result.TotalUsage.CompletionTokens)
	}
	if result.TotalUsage.TotalTokens != 30 {
		t.Errorf("TotalTokens = %d, want 30", result.TotalUsage.TotalTokens)
	}
}

func TestRunProgress(t *testing.T) {
	var progress []string

	p := New("progress",
		WithStages(
			Stage{Name: "s1", Agent: newTestAgent("a", "r1")},
			Stage{Name: "s2", Agent: newTestAgent("b", "r2")},
			Stage{Name: "s3", Agent: newTestAgent("c", "r3")},
		),
		WithProgress(func(idx, total int, sr StageResult) {
			progress = append(progress, sr.Name+":"+strconv.Itoa(idx)+"/"+strconv.Itoa(total))
		}),
	)

	_, err := p.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(progress) != 3 {
		t.Fatalf("progress callbacks = %d, want 3", len(progress))
	}
	expected := []string{"s1:0/3", "s2:1/3", "s3:2/3"}
	for i, want := range expected {
		if progress[i] != want {
			t.Errorf("progress[%d] = %q, want %q", i, progress[i], want)
		}
	}
}

func TestRunStageNameFallback(t *testing.T) {
	// When Stage.Name is empty, it should use agent name.
	p := New("fallback",
		WithStages(Stage{
			Agent: newTestAgent("my-agent", "resp"),
		}),
	)

	result, err := p.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Stages[0].Name != "my-agent" {
		t.Errorf("Stage name = %q, want %q", result.Stages[0].Name, "my-agent")
	}
}

func TestRunRetryWithSkipOnExhaustion(t *testing.T) {
	bad := agent.New("bad",
		agent.WithProvider(&errorProvider{err: errors.New("fails")}),
		agent.WithModel("gpt-4o"),
	)

	p := New("retry-skip",
		WithStages(
			Stage{
				Name:    "flaky",
				Agent:   bad,
				OnError: Skip,
				Retry:   RetryPolicy{MaxAttempts: 2, Delay: time.Millisecond},
			},
			Stage{Name: "final", Agent: newTestAgent("a", "done")},
		),
	)

	result, err := p.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Stages[0].Skipped {
		t.Error("stage 0 should be skipped")
	}
	if result.Stages[0].Attempts != 2 {
		t.Errorf("Attempts = %d, want 2", result.Stages[0].Attempts)
	}
	if result.Output != "done" {
		t.Errorf("Output = %q, want %q", result.Output, "done")
	}
}

func TestRunTotalCost(t *testing.T) {
	p := New("cost",
		WithStages(
			Stage{Name: "s1", Agent: newTestAgent("a", "resp")},
			Stage{Name: "s2", Agent: newTestAgent("b", "resp")},
		),
	)

	result, err := p.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.TotalCost == 0 {
		t.Error("TotalCost should be nonzero with gpt-4o pricing")
	}
}

func TestRunSkippedStagePreservesInput(t *testing.T) {
	bad := agent.New("bad",
		agent.WithProvider(&errorProvider{err: errors.New("fails")}),
		agent.WithModel("gpt-4o"),
	)

	p := New("preserve",
		WithStages(
			Stage{Name: "first", Agent: newTestAgent("a", "stage-1-output")},
			Stage{Name: "skip-me", Agent: bad, OnError: Skip},
			Stage{Name: "final", Agent: newTestAgent("c", "final")},
		),
	)

	result, err := p.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// The final stage should have run (even though middle was skipped).
	if len(result.Stages) != 3 {
		t.Errorf("Stages len = %d, want 3", len(result.Stages))
	}
	if result.Output != "final" {
		t.Errorf("Output = %q, want %q", result.Output, "final")
	}
}
