package graph

import (
	"context"
	"errors"
	"strings"
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

// --- Builder Tests ---

func TestBuilderSingleNode(t *testing.T) {
	g, err := NewBuilder("test").
		AddNode("only", newTestAgent("only", "output")).
		Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if g.Name() != "test" {
		t.Errorf("Name = %q, want %q", g.Name(), "test")
	}
	if len(g.Nodes()) != 1 {
		t.Errorf("Nodes = %d, want 1", len(g.Nodes()))
	}
}

func TestBuilderNoNodes(t *testing.T) {
	_, err := NewBuilder("empty").Build()
	if err == nil {
		t.Fatal("expected error for empty graph")
	}
}

func TestBuilderDuplicateNode(t *testing.T) {
	_, err := NewBuilder("dup").
		AddNode("a", newTestAgent("a", "")).
		AddNode("a", newTestAgent("a2", "")).
		Build()
	if err == nil {
		t.Fatal("expected error for duplicate node")
	}
}

func TestBuilderUnknownEdgeNode(t *testing.T) {
	_, err := NewBuilder("bad-edge").
		AddNode("a", newTestAgent("a", "")).
		AddEdge("a", "nonexistent").
		Build()
	if err == nil {
		t.Fatal("expected error for unknown edge target")
	}
}

func TestBuilderUnknownEdgeSource(t *testing.T) {
	_, err := NewBuilder("bad-source").
		AddNode("a", newTestAgent("a", "")).
		AddEdge("nonexistent", "a").
		Build()
	if err == nil {
		t.Fatal("expected error for unknown edge source")
	}
}

func TestBuilderLinearChain(t *testing.T) {
	g, err := NewBuilder("chain").
		AddNode("a", newTestAgent("a", "")).
		AddNode("b", newTestAgent("b", "")).
		AddNode("c", newTestAgent("c", "")).
		AddEdge("a", "b").
		AddEdge("b", "c").
		Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(g.Nodes()) != 3 {
		t.Errorf("Nodes = %d, want 3", len(g.Nodes()))
	}
	if len(g.Edges()) != 2 {
		t.Errorf("Edges = %d, want 2", len(g.Edges()))
	}
}

func TestBuilderWithOptions(t *testing.T) {
	tracer := trace.NewInMemory()
	g, err := NewBuilder("opts").
		AddNode("a", newTestAgent("a", "")).
		Options(
			WithTracer(tracer),
			WithConfig(Config{MaxIterations: 5}),
		).
		Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if g.config.MaxIterations != 5 {
		t.Errorf("MaxIterations = %d, want 5", g.config.MaxIterations)
	}
}

// --- DOT Export Tests ---

func TestDOT(t *testing.T) {
	g, _ := NewBuilder("my-graph").
		AddNode("a", newTestAgent("a", "")).
		AddNode("b", newTestAgent("b", "")).
		AddEdge("a", "b").
		Build()

	dot := g.DOT()
	if !strings.Contains(dot, "digraph") {
		t.Error("DOT missing digraph keyword")
	}
	if !strings.Contains(dot, `"a"`) {
		t.Error("DOT missing node a")
	}
	if !strings.Contains(dot, `"a" -> "b"`) {
		t.Error("DOT missing edge a -> b")
	}
}

func TestDOTConditionalEdge(t *testing.T) {
	g, _ := NewBuilder("cond").
		AddNode("a", newTestAgent("a", "")).
		AddNode("b", newTestAgent("b", "")).
		AddEdge("a", "b", Always()).
		Build()

	dot := g.DOT()
	if !strings.Contains(dot, "dashed") {
		t.Error("DOT missing dashed style for conditional edge")
	}
}

// --- Execution Tests ---

func TestRunSingleNode(t *testing.T) {
	g, _ := NewBuilder("single").
		AddNode("only", newTestAgent("only", "hello world")).
		Build()

	result, err := g.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Output != "hello world" {
		t.Errorf("Output = %q, want %q", result.Output, "hello world")
	}
	if result.RunID == "" {
		t.Error("RunID is empty")
	}
}

func TestRunLinearChain(t *testing.T) {
	g, _ := NewBuilder("chain").
		AddNode("a", newTestAgent("a", "from-a")).
		AddNode("b", newTestAgent("b", "from-b")).
		AddNode("c", newTestAgent("c", "from-c")).
		AddEdge("a", "b").
		AddEdge("b", "c").
		Build()

	result, err := g.Run(context.Background(), "start")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Final output should be from terminal node "c".
	if result.Output != "from-c" {
		t.Errorf("Output = %q, want %q", result.Output, "from-c")
	}

	// All three nodes should have results.
	for _, name := range []string{"a", "b", "c"} {
		if _, ok := result.NodeResults[name]; !ok {
			t.Errorf("missing result for node %q", name)
		}
	}
}

func TestRunFanOut(t *testing.T) {
	// a -> b, a -> c (parallel)
	g, _ := NewBuilder("fanout").
		AddNode("a", newTestAgent("a", "from-a")).
		AddNode("b", newTestAgent("b", "from-b")).
		AddNode("c", newTestAgent("c", "from-c")).
		AddEdge("a", "b").
		AddEdge("a", "c").
		Build()

	result, err := g.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Both b and c are terminal nodes.
	if !strings.Contains(result.Output, "from-b") {
		t.Error("Output missing from-b")
	}
	if !strings.Contains(result.Output, "from-c") {
		t.Error("Output missing from-c")
	}
}

func TestRunFanIn(t *testing.T) {
	// a -> c, b -> c (merge)
	g, _ := NewBuilder("fanin").
		AddNode("a", newTestAgent("a", "from-a")).
		AddNode("b", newTestAgent("b", "from-b")).
		AddNode("c", newTestAgent("c", "merged")).
		AddEdge("a", "c").
		AddEdge("b", "c").
		Build()

	result, err := g.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// c is the terminal node.
	if result.Output != "merged" {
		t.Errorf("Output = %q, want %q", result.Output, "merged")
	}

	// All three nodes should have run.
	for _, name := range []string{"a", "b", "c"} {
		if _, ok := result.NodeResults[name]; !ok {
			t.Errorf("missing result for node %q", name)
		}
	}
}

func TestRunDiamond(t *testing.T) {
	// a -> b, a -> c, b -> d, c -> d
	g, _ := NewBuilder("diamond").
		AddNode("a", newTestAgent("a", "from-a")).
		AddNode("b", newTestAgent("b", "from-b")).
		AddNode("c", newTestAgent("c", "from-c")).
		AddNode("d", newTestAgent("d", "from-d")).
		AddEdge("a", "b").
		AddEdge("a", "c").
		AddEdge("b", "d").
		AddEdge("c", "d").
		Build()

	result, err := g.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Output != "from-d" {
		t.Errorf("Output = %q, want %q", result.Output, "from-d")
	}

	for _, name := range []string{"a", "b", "c", "d"} {
		if _, ok := result.NodeResults[name]; !ok {
			t.Errorf("missing result for node %q", name)
		}
	}
}

func TestRunConditionalEdge(t *testing.T) {
	// a -> b (when output contains "yes"), a -> c (when output contains "no")
	g, _ := NewBuilder("conditional").
		AddNode("a", newTestAgent("a", "yes please")).
		AddNode("b", newTestAgent("b", "took-yes-path")).
		AddNode("c", newTestAgent("c", "took-no-path")).
		AddEdge("a", "b", When(func(out string) bool { return strings.Contains(out, "yes") })).
		AddEdge("a", "c", When(func(out string) bool { return strings.Contains(out, "no") })).
		Build()

	result, err := g.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Only b should have run (a outputs "yes please").
	if _, ok := result.NodeResults["b"]; !ok {
		t.Error("expected node b to run (yes path)")
	}
	if _, ok := result.NodeResults["c"]; ok {
		t.Error("expected node c NOT to run (no path)")
	}
}

func TestRunLoop(t *testing.T) {
	// draft -> review -> revise (conditional) -> review (loop back)
	//                  -> publish (conditional, when "done")
	callCount := &atomic.Int32{}

	dynamicProvider := &dynamicMockProvider{
		threshold: 2,
		before: &llm.Response{
			Message: llm.NewAssistantMessage("needs revision"),
			Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
			Model:   "gpt-4o",
		},
		after: &llm.Response{
			Message: llm.NewAssistantMessage("done reviewing"),
			Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
			Model:   "gpt-4o",
		},
		calls: callCount,
	}

	reviewAgent := agent.New("review",
		agent.WithProvider(dynamicProvider),
		agent.WithModel("gpt-4o"),
	)

	g, _ := NewBuilder("loop").
		AddNode("draft", newTestAgent("draft", "initial draft")).
		AddNode("review", reviewAgent).
		AddNode("revise", newTestAgent("revise", "revised draft")).
		AddNode("publish", newTestAgent("publish", "published")).
		AddEdge("draft", "review").
		AddEdge("review", "revise", When(func(out string) bool { return strings.Contains(out, "revision") })).
		AddEdge("review", "publish", When(func(out string) bool { return strings.Contains(out, "done") })).
		AddEdge("revise", "review").
		Options(WithConfig(Config{MaxIterations: 5})).
		Build()

	result, err := g.Run(context.Background(), "write something")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Output != "published" {
		t.Errorf("Output = %q, want %q", result.Output, "published")
	}

	// Review should have been visited more than once (loop).
	if len(result.NodeResults["review"]) < 2 {
		t.Errorf("review iterations = %d, want >= 2", len(result.NodeResults["review"]))
	}
}

// dynamicMockProvider returns different responses based on call count.
type dynamicMockProvider struct {
	threshold int
	before    *llm.Response
	after     *llm.Response
	calls     *atomic.Int32
}

func (d *dynamicMockProvider) Complete(_ context.Context, _ llm.Params) (*llm.Response, error) {
	n := int(d.calls.Add(1))
	if n <= d.threshold {
		return d.before, nil
	}
	return d.after, nil
}

func TestRunMaxIterations(t *testing.T) {
	// start -> a -> b -> a (loop, but capped by max iterations)
	g, _ := NewBuilder("infinite").
		AddNode("start", newTestAgent("start", "go")).
		AddNode("a", newTestAgent("a", "looping")).
		AddNode("b", newTestAgent("b", "looping")).
		AddEdge("start", "a").
		AddEdge("a", "b").
		AddEdge("b", "a").
		Options(WithConfig(Config{MaxIterations: 3})).
		Build()

	result, err := g.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// a and b should have been visited at most 3 times each.
	for _, name := range []string{"a", "b"} {
		rs := result.NodeResults[name]
		if len(rs) > 3 {
			t.Errorf("node %q visited %d times, want <= 3", name, len(rs))
		}
	}
}

func TestRunTracing(t *testing.T) {
	tracer := trace.NewInMemory()

	g, _ := NewBuilder("traced").
		AddNode("a", newTestAgent("a", "resp")).
		AddNode("b", newTestAgent("b", "resp")).
		AddEdge("a", "b").
		Options(WithTracer(tracer)).
		Build()

	_, err := g.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	spans := tracer.Spans()
	names := make(map[string]bool)
	for _, s := range spans {
		names[s.Name] = true
	}
	if !names["graph.run"] {
		t.Error("missing graph.run span")
	}
	if !names["graph.node"] {
		t.Error("missing graph.node span")
	}
}

func TestRunTimeout(t *testing.T) {
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

	g, _ := NewBuilder("timeout").
		AddNode("slow", slow).
		Options(WithConfig(Config{Timeout: 50 * time.Millisecond})).
		Build()

	_, err := g.Run(context.Background(), "input")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRunContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	g, _ := NewBuilder("canceled").
		AddNode("a", newTestAgent("a", "ok")).
		Build()

	_, err := g.Run(ctx, "hello")
	if err == nil {
		t.Fatal("expected context canceled error")
	}
}

func TestRunCostBudget(t *testing.T) {
	g, _ := NewBuilder("budget").
		AddNode("a", newTestAgent("a", "resp")).
		AddNode("b", newTestAgent("b", "resp")).
		AddNode("c", newTestAgent("c", "resp")).
		AddEdge("a", "b").
		AddEdge("b", "c").
		Options(WithConfig(Config{CostBudget: 0.000001})).
		Build()

	result, err := g.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Should stop early.
	total := 0
	for _, rs := range result.NodeResults {
		total += len(rs)
	}
	if total >= 3 {
		t.Errorf("total node executions = %d, expected < 3 with tiny budget", total)
	}
}

func TestRunAggregatesUsage(t *testing.T) {
	g, _ := NewBuilder("usage").
		AddNode("a", newTestAgent("a", "resp")).
		AddNode("b", newTestAgent("b", "resp")).
		AddEdge("a", "b").
		Build()

	result, err := g.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Each mock: 10 prompt + 5 completion = 15 total.
	if result.TotalUsage.TotalTokens != 30 {
		t.Errorf("TotalTokens = %d, want 30", result.TotalUsage.TotalTokens)
	}
}

func TestRunNodeError(t *testing.T) {
	bad := agent.New("bad",
		agent.WithProvider(&errorProvider{err: errors.New("node down")}),
		agent.WithModel("gpt-4o"),
	)

	g, _ := NewBuilder("error").
		AddNode("a", newTestAgent("a", "ok")).
		AddNode("bad", bad).
		AddEdge("a", "bad").
		Build()

	// Should still return a result (the error node just has no results).
	result, err := g.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if _, ok := result.NodeResults["bad"]; ok {
		t.Error("expected no result for failed node")
	}
}

func TestRunMultipleStartNodes(t *testing.T) {
	// a and b are both start nodes, both feed into c.
	g, _ := NewBuilder("multi-start").
		AddNode("a", newTestAgent("a", "from-a")).
		AddNode("b", newTestAgent("b", "from-b")).
		AddNode("c", newTestAgent("c", "merged")).
		AddEdge("a", "c").
		AddEdge("b", "c").
		Build()

	result, err := g.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Output != "merged" {
		t.Errorf("Output = %q, want %q", result.Output, "merged")
	}
}

func TestRunNoStartNodes(t *testing.T) {
	// All nodes have incoming edges (circular).
	g := &Graph{
		name:     "no-start",
		nodes:    map[string]*Node{"a": {name: "a", agent: newTestAgent("a", "")}},
		edges:    []Edge{{From: "a", To: "a"}},
		outgoing: map[string][]Edge{"a": {{From: "a", To: "a"}}},
		incoming: map[string][]string{"a": {"a"}},
		starts:   nil,
		ends:     nil,
		tracer:   trace.Noop{},
	}

	_, err := g.Run(context.Background(), "input")
	if err == nil {
		t.Fatal("expected error for no start nodes")
	}
}
