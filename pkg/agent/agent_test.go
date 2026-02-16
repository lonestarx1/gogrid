package agent

import (
	"context"
	"testing"
	"time"

	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/memory"
)

// mockProvider is a minimal LLM provider for testing agent construction.
type mockProvider struct{}

func (m *mockProvider) Complete(_ context.Context, _ llm.Params) (*llm.Response, error) {
	return &llm.Response{
		Message: llm.NewAssistantMessage("mock response"),
		Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		Model:   "mock-model",
	}, nil
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
		WithProvider(&mockProvider{}),
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
	a := New("agent", WithProvider(&mockProvider{}))
	_, err := a.Run(context.Background(), "hello")
	if err == nil {
		t.Fatal("Run without model: expected error, got nil")
	}
}

func TestRunReturnsResult(t *testing.T) {
	a := New("agent",
		WithProvider(&mockProvider{}),
		WithModel("mock-model"),
	)

	result, err := a.Run(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.RunID == "" {
		t.Error("RunID is empty")
	}
}
