package bench

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/llm/mock"
	"github.com/lonestarx1/gogrid/pkg/tool"
)

func newTestAgent(name, content string) *agent.Agent {
	return agent.New(name,
		agent.WithProvider(mock.New(mock.WithFallback(&llm.Response{
			Message: llm.NewAssistantMessage(content),
			Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
			Model:   "mock-model",
		}))),
		agent.WithModel("mock-model"),
	)
}

// mockTool is a minimal tool implementation for benchmarks.
type mockTool struct {
	name   string
	output string
}

func (m *mockTool) Name() string        { return m.name }
func (m *mockTool) Description() string { return m.name + " tool" }
func (m *mockTool) Schema() tool.Schema { return tool.Schema{Type: "object"} }
func (m *mockTool) Execute(_ context.Context, _ json.RawMessage) (string, error) {
	return m.output, nil
}

func BenchmarkAgentRun(b *testing.B) {
	a := newTestAgent("bench-agent", "hello world")
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, err := a.Run(ctx, "benchmark input")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAgentRunWithToolUse(b *testing.B) {
	// Provider returns a tool call on first turn, then a final response.
	toolCallResp := &llm.Response{
		Message: llm.Message{
			Role: llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{
				{ID: "tc-1", Function: "search", Arguments: json.RawMessage(`{"q":"test"}`)},
			},
		},
		Usage: llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		Model: "mock-model",
	}
	finalResp := &llm.Response{
		Message: llm.NewAssistantMessage("found it"),
		Usage:   llm.Usage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30},
		Model:   "mock-model",
	}

	provider := mock.New(
		mock.WithResponses(toolCallResp, finalResp),
		mock.WithFallback(finalResp),
	)

	a := agent.New("bench-tool-agent",
		agent.WithProvider(provider),
		agent.WithModel("mock-model"),
		agent.WithTools(&mockTool{name: "search", output: "result"}),
	)

	ctx := context.Background()
	b.ResetTimer()
	for b.Loop() {
		provider.Reset()
		_, err := a.Run(ctx, "use the tool")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAgentRunParallel(b *testing.B) {
	a := newTestAgent("bench-parallel", "response")
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := a.Run(ctx, "parallel input")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
