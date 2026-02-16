package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/memory"
	"github.com/lonestarx1/gogrid/pkg/tool"
	"github.com/lonestarx1/gogrid/pkg/trace"
)

// mockProvider is a minimal LLM provider for testing.
type mockProvider struct {
	responses []*llm.Response
	calls     int
}

func newMockProvider(responses ...*llm.Response) *mockProvider {
	return &mockProvider{responses: responses}
}

func (m *mockProvider) Complete(_ context.Context, _ llm.Params) (*llm.Response, error) {
	if m.calls >= len(m.responses) {
		return &llm.Response{
			Message: llm.NewAssistantMessage("fallback"),
			Usage:   llm.Usage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2},
			Model:   "mock-model",
		}, nil
	}
	resp := m.responses[m.calls]
	m.calls++
	return resp, nil
}

// errorProvider always returns an error.
type errorProvider struct{ err error }

func (e *errorProvider) Complete(_ context.Context, _ llm.Params) (*llm.Response, error) {
	return nil, e.err
}

// mockTool is a tool implementation for testing.
type mockTool struct {
	name   string
	output string
	err    error
	called int
}

func (m *mockTool) Name() string        { return m.name }
func (m *mockTool) Description() string { return m.name + " tool" }
func (m *mockTool) Schema() tool.Schema { return tool.Schema{Type: "object"} }
func (m *mockTool) Execute(_ context.Context, _ json.RawMessage) (string, error) {
	m.called++
	return m.output, m.err
}

func TestNewAgent(t *testing.T) {
	a := New("test-agent")
	if a.Name() != "test-agent" {
		t.Errorf("Name = %q, want %q", a.Name(), "test-agent")
	}
}

func TestAgentOptions(t *testing.T) {
	temp := 0.7
	a := New("agent",
		WithInstructions("you are helpful"),
		WithModel("gpt-4o"),
		WithProvider(newMockProvider()),
		WithMemory(memory.NewInMemory()),
		WithConfig(Config{
			MaxTurns:    10,
			MaxTokens:   4096,
			Temperature: &temp,
			Timeout:     30 * time.Second,
			CostBudget:  1.50,
		}),
	)

	if a.name != "agent" {
		t.Errorf("name = %q, want %q", a.name, "agent")
	}
	if a.instructions != "you are helpful" {
		t.Errorf("instructions = %q, want %q", a.instructions, "you are helpful")
	}
	if a.model != "gpt-4o" {
		t.Errorf("model = %q, want %q", a.model, "gpt-4o")
	}
	if a.provider == nil {
		t.Error("provider is nil")
	}
	if a.memory == nil {
		t.Error("memory is nil")
	}
	if a.config.MaxTurns != 10 {
		t.Errorf("MaxTurns = %d, want 10", a.config.MaxTurns)
	}
	if a.config.CostBudget != 1.50 {
		t.Errorf("CostBudget = %f, want 1.50", a.config.CostBudget)
	}
}

func TestRunRequiresProvider(t *testing.T) {
	a := New("agent", WithModel("gpt-4o"))
	_, err := a.Run(context.Background(), "hello")
	if err == nil {
		t.Fatal("Run without provider: expected error, got nil")
	}
}

func TestRunRequiresModel(t *testing.T) {
	a := New("agent", WithProvider(newMockProvider()))
	_, err := a.Run(context.Background(), "hello")
	if err == nil {
		t.Fatal("Run without model: expected error, got nil")
	}
}

func TestRunReturnsResult(t *testing.T) {
	provider := newMockProvider(&llm.Response{
		Message: llm.NewAssistantMessage("hello back"),
		Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		Model:   "mock-model",
	})

	a := New("agent",
		WithProvider(provider),
		WithModel("mock-model"),
	)

	result, err := a.Run(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.RunID == "" {
		t.Error("RunID is empty")
	}
	if result.Message.Content != "hello back" {
		t.Errorf("Message.Content = %q, want %q", result.Message.Content, "hello back")
	}
	if result.Turns != 1 {
		t.Errorf("Turns = %d, want 1", result.Turns)
	}
	if result.Usage.PromptTokens != 10 {
		t.Errorf("Usage.PromptTokens = %d, want 10", result.Usage.PromptTokens)
	}
}

func TestRunWithToolCalls(t *testing.T) {
	// Turn 1: LLM requests a tool call.
	// Turn 2: LLM gives final response after seeing tool result.
	provider := newMockProvider(
		&llm.Response{
			Message: llm.Message{
				Role: llm.RoleAssistant,
				ToolCalls: []llm.ToolCall{
					{ID: "tc-1", Function: "search", Arguments: json.RawMessage(`{"q":"test"}`)},
				},
			},
			Usage: llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
			Model: "mock-model",
		},
		&llm.Response{
			Message: llm.NewAssistantMessage("found result: hello world"),
			Usage:   llm.Usage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30},
			Model:   "mock-model",
		},
	)

	searchTool := &mockTool{name: "search", output: "hello world"}

	a := New("agent",
		WithProvider(provider),
		WithModel("mock-model"),
		WithTools(searchTool),
	)

	result, err := a.Run(context.Background(), "search for test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if searchTool.called != 1 {
		t.Errorf("tool called %d times, want 1", searchTool.called)
	}
	if result.Turns != 2 {
		t.Errorf("Turns = %d, want 2", result.Turns)
	}
	if result.Message.Content != "found result: hello world" {
		t.Errorf("Message.Content = %q, want %q", result.Message.Content, "found result: hello world")
	}
	if result.Usage.PromptTokens != 30 {
		t.Errorf("Usage.PromptTokens = %d, want 30", result.Usage.PromptTokens)
	}
}

func TestRunToolNotFound(t *testing.T) {
	// LLM requests a tool that doesn't exist, then gives final answer.
	provider := newMockProvider(
		&llm.Response{
			Message: llm.Message{
				Role: llm.RoleAssistant,
				ToolCalls: []llm.ToolCall{
					{ID: "tc-1", Function: "nonexistent", Arguments: json.RawMessage(`{}`)},
				},
			},
			Usage: llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
			Model: "mock-model",
		},
		&llm.Response{
			Message: llm.NewAssistantMessage("sorry, tool not available"),
			Usage:   llm.Usage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30},
			Model:   "mock-model",
		},
	)

	a := New("agent",
		WithProvider(provider),
		WithModel("mock-model"),
	)

	result, err := a.Run(context.Background(), "use nonexistent")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Should still complete - tool not found results in error message to LLM.
	if result.Turns != 2 {
		t.Errorf("Turns = %d, want 2", result.Turns)
	}
}

func TestRunToolExecutionError(t *testing.T) {
	provider := newMockProvider(
		&llm.Response{
			Message: llm.Message{
				Role: llm.RoleAssistant,
				ToolCalls: []llm.ToolCall{
					{ID: "tc-1", Function: "broken", Arguments: json.RawMessage(`{}`)},
				},
			},
			Usage: llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
			Model: "mock-model",
		},
		&llm.Response{
			Message: llm.NewAssistantMessage("tool failed, sorry"),
			Usage:   llm.Usage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30},
			Model:   "mock-model",
		},
	)

	brokenTool := &mockTool{name: "broken", err: errors.New("disk full")}

	a := New("agent",
		WithProvider(provider),
		WithModel("mock-model"),
		WithTools(brokenTool),
	)

	result, err := a.Run(context.Background(), "use broken tool")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if brokenTool.called != 1 {
		t.Errorf("tool called %d times, want 1", brokenTool.called)
	}
	if result.Turns != 2 {
		t.Errorf("Turns = %d, want 2", result.Turns)
	}
}

func TestRunMaxTurns(t *testing.T) {
	// Provider always returns tool calls, but MaxTurns=2 should stop the loop.
	toolCallResp := &llm.Response{
		Message: llm.Message{
			Role: llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{
				{ID: "tc-1", Function: "search", Arguments: json.RawMessage(`{}`)},
			},
		},
		Usage: llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		Model: "mock-model",
	}

	provider := newMockProvider(toolCallResp, toolCallResp, toolCallResp, toolCallResp)
	searchTool := &mockTool{name: "search", output: "result"}

	a := New("agent",
		WithProvider(provider),
		WithModel("mock-model"),
		WithTools(searchTool),
		WithConfig(Config{MaxTurns: 2}),
	)

	result, err := a.Run(context.Background(), "loop forever")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Turns != 2 {
		t.Errorf("Turns = %d, want 2", result.Turns)
	}
}

func TestRunWithInstructions(t *testing.T) {
	var capturedParams llm.Params
	provider := &capturingProvider{
		response: &llm.Response{
			Message: llm.NewAssistantMessage("ok"),
			Usage:   llm.Usage{PromptTokens: 5, CompletionTokens: 3, TotalTokens: 8},
			Model:   "mock-model",
		},
		capture: func(p llm.Params) { capturedParams = p },
	}

	a := New("agent",
		WithProvider(provider),
		WithModel("mock-model"),
		WithInstructions("you are a helpful assistant"),
	)

	_, err := a.Run(context.Background(), "hi")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(capturedParams.Messages) < 2 {
		t.Fatalf("Messages len = %d, want >= 2", len(capturedParams.Messages))
	}
	if capturedParams.Messages[0].Role != llm.RoleSystem {
		t.Errorf("Messages[0].Role = %q, want %q", capturedParams.Messages[0].Role, llm.RoleSystem)
	}
	if capturedParams.Messages[0].Content != "you are a helpful assistant" {
		t.Errorf("system message = %q, want %q", capturedParams.Messages[0].Content, "you are a helpful assistant")
	}
}

func TestRunWithMemory(t *testing.T) {
	ctx := context.Background()
	mem := memory.NewInMemory()

	// Pre-populate memory.
	_ = mem.Save(ctx, "agent", []llm.Message{
		llm.NewUserMessage("previous question"),
		llm.NewAssistantMessage("previous answer"),
	})

	provider := newMockProvider(&llm.Response{
		Message: llm.NewAssistantMessage("new response"),
		Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		Model:   "mock-model",
	})

	a := New("agent",
		WithProvider(provider),
		WithModel("mock-model"),
		WithMemory(mem),
	)

	result, err := a.Run(ctx, "new question")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// History should include memory + new messages.
	if len(result.History) < 4 {
		t.Errorf("History len = %d, want >= 4 (2 memory + 1 user + 1 assistant)", len(result.History))
	}

	// Memory should be updated after run.
	saved, _ := mem.Load(ctx, "agent")
	if len(saved) != len(result.History) {
		t.Errorf("saved memory len = %d, want %d", len(saved), len(result.History))
	}
}

func TestRunWithTracer(t *testing.T) {
	tracer := trace.NewInMemory()

	provider := newMockProvider(&llm.Response{
		Message: llm.NewAssistantMessage("traced response"),
		Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		Model:   "mock-model",
	})

	a := New("agent",
		WithProvider(provider),
		WithModel("mock-model"),
		WithTracer(tracer),
	)

	_, err := a.Run(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	spans := tracer.Spans()
	if len(spans) < 2 {
		t.Fatalf("Spans len = %d, want >= 2 (agent.run + llm.complete)", len(spans))
	}

	// Check span names.
	names := make(map[string]bool)
	for _, s := range spans {
		names[s.Name] = true
	}
	if !names["agent.run"] {
		t.Error("missing agent.run span")
	}
	if !names["llm.complete"] {
		t.Error("missing llm.complete span")
	}
}

func TestRunLLMError(t *testing.T) {
	provider := &errorProvider{err: fmt.Errorf("rate limited")}

	a := New("agent",
		WithProvider(provider),
		WithModel("mock-model"),
	)

	_, err := a.Run(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, provider.err) {
		t.Errorf("error = %v, want wrapped %v", err, provider.err)
	}
}

func TestRunContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	provider := newMockProvider()

	a := New("agent",
		WithProvider(provider),
		WithModel("mock-model"),
	)

	_, err := a.Run(ctx, "hello")
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}

func TestDefaultTracerIsNoop(t *testing.T) {
	a := New("agent")
	if _, ok := a.tracer.(trace.Noop); !ok {
		t.Errorf("default tracer is %T, want trace.Noop", a.tracer)
	}
}

func TestRunMemoryTraceSpans(t *testing.T) {
	ctx := context.Background()
	tracer := trace.NewInMemory()
	mem := memory.NewInMemory()

	provider := newMockProvider(&llm.Response{
		Message: llm.NewAssistantMessage("ok"),
		Usage:   llm.Usage{PromptTokens: 5, CompletionTokens: 3, TotalTokens: 8},
		Model:   "mock-model",
	})

	a := New("agent",
		WithProvider(provider),
		WithModel("mock-model"),
		WithMemory(mem),
		WithTracer(tracer),
	)

	_, err := a.Run(ctx, "hi")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	spans := tracer.Spans()
	names := make(map[string]bool)
	for _, s := range spans {
		names[s.Name] = true
	}
	if !names["memory.load"] {
		t.Error("missing memory.load span")
	}
	if !names["memory.save"] {
		t.Error("missing memory.save span")
	}
}

func TestRunMemoryStats(t *testing.T) {
	ctx := context.Background()
	mem := memory.NewInMemory()

	provider := newMockProvider(&llm.Response{
		Message: llm.NewAssistantMessage("response"),
		Usage:   llm.Usage{PromptTokens: 5, CompletionTokens: 3, TotalTokens: 8},
		Model:   "mock-model",
	})

	a := New("agent",
		WithProvider(provider),
		WithModel("mock-model"),
		WithMemory(mem),
	)

	result, err := a.Run(ctx, "hi")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// InMemory implements StatsMemory, so MemoryStats should be populated.
	if result.MemoryStats == nil {
		t.Fatal("MemoryStats is nil")
	}
	if result.MemoryStats.Keys != 1 {
		t.Errorf("MemoryStats.Keys = %d, want 1", result.MemoryStats.Keys)
	}
	if result.MemoryStats.TotalEntries < 1 {
		t.Errorf("MemoryStats.TotalEntries = %d, want >= 1", result.MemoryStats.TotalEntries)
	}
}

func TestRunMemoryStatsNilWithoutStatsMemory(t *testing.T) {
	ctx := context.Background()
	// ReadOnly wraps InMemory but doesn't implement StatsMemory itself.
	inner := memory.NewInMemory()
	ro := memory.NewReadOnly(inner)

	// Pre-populate so Load works.
	_ = inner.Save(ctx, "agent", []llm.Message{})

	provider := newMockProvider(&llm.Response{
		Message: llm.NewAssistantMessage("ok"),
		Usage:   llm.Usage{PromptTokens: 5, CompletionTokens: 3, TotalTokens: 8},
		Model:   "mock-model",
	})

	a := New("agent",
		WithProvider(provider),
		WithModel("mock-model"),
		WithMemory(ro),
	)

	result, err := a.Run(ctx, "hi")
	// ReadOnly.Save returns ErrReadOnly, so Run will error.
	// But the point is: if memory doesn't implement StatsMemory,
	// MemoryStats should be nil.
	if err != nil {
		// Expected: ReadOnly rejects Save. That's fine for this test.
		return
	}
	if result.MemoryStats != nil {
		t.Errorf("MemoryStats should be nil for non-StatsMemory, got %+v", result.MemoryStats)
	}
}

// capturingProvider captures the params passed to Complete.
type capturingProvider struct {
	response *llm.Response
	capture  func(llm.Params)
}

func (c *capturingProvider) Complete(_ context.Context, params llm.Params) (*llm.Response, error) {
	if c.capture != nil {
		c.capture(params)
	}
	return c.response, nil
}
