package dynamic

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/orchestrator/graph"
	"github.com/lonestarx1/gogrid/pkg/orchestrator/pipeline"
	"github.com/lonestarx1/gogrid/pkg/orchestrator/team"
	"github.com/lonestarx1/gogrid/pkg/trace"
)

// --- Mock providers ---

type mockProvider struct {
	response *llm.Response
	calls    atomic.Int32
}

func newMockProvider(content string) *mockProvider {
	return &mockProvider{
		response: &llm.Response{
			Message: llm.NewAssistantMessage(content),
			Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
			Model:   "mock-model",
		},
	}
}

func (m *mockProvider) Complete(_ context.Context, _ llm.Params) (*llm.Response, error) {
	m.calls.Add(1)
	return m.response, nil
}

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

type errorProvider struct{ err error }

func (e *errorProvider) Complete(_ context.Context, _ llm.Params) (*llm.Response, error) {
	return nil, e.err
}

// costProvider returns a response with configurable cost.
type costProvider struct {
	content string
	usage   llm.Usage
}

func (c *costProvider) Complete(_ context.Context, _ llm.Params) (*llm.Response, error) {
	return &llm.Response{
		Message: llm.NewAssistantMessage(c.content),
		Usage:   c.usage,
		Model:   "mock-model",
	}, nil
}

// --- Helpers ---

func newTestAgent(name, content string) *agent.Agent {
	return agent.New(name,
		agent.WithProvider(newMockProvider(content)),
		agent.WithModel("mock-model"),
	)
}

func newSlowAgent(name string, delay time.Duration) *agent.Agent {
	return agent.New(name,
		agent.WithProvider(&slowProvider{
			delay: delay,
			response: &llm.Response{
				Message: llm.NewAssistantMessage(name + "-done"),
				Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
				Model:   "mock-model",
			},
		}),
		agent.WithModel("mock-model"),
	)
}

func newTestTeam(name string, agentContent string) *team.Team {
	return team.New(name,
		team.WithMembers(
			team.Member{Agent: newTestAgent("member-a", agentContent)},
			team.Member{Agent: newTestAgent("member-b", agentContent)},
		),
	)
}

func newTestPipeline(name string) *pipeline.Pipeline {
	return pipeline.New(name,
		pipeline.WithStages(
			pipeline.Stage{
				Name:  "stage-1",
				Agent: newTestAgent("s1", "stage-1-output"),
			},
			pipeline.Stage{
				Name:  "stage-2",
				Agent: newTestAgent("s2", "stage-2-output"),
			},
		),
	)
}

func newTestGraph(name string) *graph.Graph {
	g, _ := graph.NewBuilder(name).
		AddNode("draft", newTestAgent("draft", "draft-output")).
		AddNode("review", newTestAgent("review", "review-output")).
		AddEdge("draft", "review").
		Build()
	return g
}

// --- Tests ---

func TestNew(t *testing.T) {
	rt := New("test-runtime")
	if rt.Name() != "test-runtime" {
		t.Errorf("Name = %q, want %q", rt.Name(), "test-runtime")
	}
	if rt.config.MaxDepth != 10 {
		t.Errorf("MaxDepth = %d, want default 10", rt.config.MaxDepth)
	}
	if rt.sem != nil {
		t.Error("sem should be nil when MaxConcurrent is 0")
	}
}

func TestNewWithConfig(t *testing.T) {
	rt := New("rt",
		WithConfig(Config{
			MaxConcurrent: 3,
			MaxDepth:      5,
			CostBudget:    1.50,
		}),
	)
	if rt.config.MaxConcurrent != 3 {
		t.Errorf("MaxConcurrent = %d, want 3", rt.config.MaxConcurrent)
	}
	if rt.config.MaxDepth != 5 {
		t.Errorf("MaxDepth = %d, want 5", rt.config.MaxDepth)
	}
	if rt.config.CostBudget != 1.50 {
		t.Errorf("CostBudget = %f, want 1.50", rt.config.CostBudget)
	}
	if rt.sem == nil || cap(rt.sem) != 3 {
		t.Error("sem should have capacity 3")
	}
}

func TestContext(t *testing.T) {
	rt := New("rt")
	ctx := rt.Context(context.Background())

	got := FromContext(ctx)
	if got != rt {
		t.Error("FromContext returned different runtime")
	}

	// No runtime in plain context.
	if FromContext(context.Background()) != nil {
		t.Error("FromContext should return nil for plain context")
	}
}

func TestDepthFromContext(t *testing.T) {
	ctx := context.Background()
	if DepthFromContext(ctx) != 0 {
		t.Error("DepthFromContext should return 0 for plain context")
	}

	ctx = context.WithValue(ctx, depthKey{}, 3)
	if DepthFromContext(ctx) != 3 {
		t.Errorf("DepthFromContext = %d, want 3", DepthFromContext(ctx))
	}
}

func TestSpawnAgent(t *testing.T) {
	rt := New("rt")
	ctx := context.Background()

	a := newTestAgent("researcher", "research-result")
	result, err := rt.SpawnAgent(ctx, a, "find papers")
	if err != nil {
		t.Fatalf("SpawnAgent error: %v", err)
	}
	if result.Message.Content != "research-result" {
		t.Errorf("output = %q, want %q", result.Message.Content, "research-result")
	}

	// Check child recorded.
	res := rt.Result()
	if len(res.Children) != 1 {
		t.Fatalf("Children = %d, want 1", len(res.Children))
	}
	if res.Children[0].Name != "researcher" {
		t.Errorf("child name = %q, want %q", res.Children[0].Name, "researcher")
	}
	if res.Children[0].Type != "agent" {
		t.Errorf("child type = %q, want %q", res.Children[0].Type, "agent")
	}
	if res.Children[0].Output != "research-result" {
		t.Errorf("child output = %q, want %q", res.Children[0].Output, "research-result")
	}
}

func TestSpawnAgentError(t *testing.T) {
	rt := New("rt")
	ctx := context.Background()

	a := agent.New("broken",
		agent.WithProvider(&errorProvider{err: errors.New("llm down")}),
		agent.WithModel("mock-model"),
	)

	_, err := rt.SpawnAgent(ctx, a, "input")
	if err == nil {
		t.Fatal("expected error from SpawnAgent")
	}

	// Error child should be recorded.
	res := rt.Result()
	if len(res.Children) != 1 {
		t.Fatalf("Children = %d, want 1", len(res.Children))
	}
	if res.Children[0].Error == nil {
		t.Error("child error should be non-nil")
	}
}

func TestSpawnTeam(t *testing.T) {
	rt := New("rt")
	ctx := context.Background()

	tm := newTestTeam("debate", "team-response")
	result, err := rt.SpawnTeam(ctx, tm, "discuss topic")
	if err != nil {
		t.Fatalf("SpawnTeam error: %v", err)
	}
	if result.Decision.Content == "" {
		t.Error("team decision should not be empty")
	}

	res := rt.Result()
	if len(res.Children) != 1 {
		t.Fatalf("Children = %d, want 1", len(res.Children))
	}
	if res.Children[0].Type != "team" {
		t.Errorf("child type = %q, want %q", res.Children[0].Type, "team")
	}
}

func TestSpawnPipeline(t *testing.T) {
	rt := New("rt")
	ctx := context.Background()

	p := newTestPipeline("research-pipeline")
	result, err := rt.SpawnPipeline(ctx, p, "research input")
	if err != nil {
		t.Fatalf("SpawnPipeline error: %v", err)
	}
	if result.Output != "stage-2-output" {
		t.Errorf("output = %q, want %q", result.Output, "stage-2-output")
	}

	res := rt.Result()
	if len(res.Children) != 1 {
		t.Fatalf("Children = %d, want 1", len(res.Children))
	}
	if res.Children[0].Type != "pipeline" {
		t.Errorf("child type = %q, want %q", res.Children[0].Type, "pipeline")
	}
}

func TestSpawnGraph(t *testing.T) {
	rt := New("rt")
	ctx := context.Background()

	g := newTestGraph("review-graph")
	result, err := rt.SpawnGraph(ctx, g, "write something")
	if err != nil {
		t.Fatalf("SpawnGraph error: %v", err)
	}
	if result.Output != "review-output" {
		t.Errorf("output = %q, want %q", result.Output, "review-output")
	}

	res := rt.Result()
	if len(res.Children) != 1 {
		t.Fatalf("Children = %d, want 1", len(res.Children))
	}
	if res.Children[0].Type != "graph" {
		t.Errorf("child type = %q, want %q", res.Children[0].Type, "graph")
	}
}

func TestMaxDepthEnforcement(t *testing.T) {
	rt := New("rt", WithConfig(Config{MaxDepth: 2}))

	// Depth 0 → OK.
	ctx := context.Background()
	_, err := rt.SpawnAgent(ctx, newTestAgent("a1", "ok"), "input")
	if err != nil {
		t.Fatalf("depth 0 should succeed: %v", err)
	}

	// Depth 1 → OK.
	ctx1 := context.WithValue(ctx, depthKey{}, 1)
	_, err = rt.SpawnAgent(ctx1, newTestAgent("a2", "ok"), "input")
	if err != nil {
		t.Fatalf("depth 1 should succeed: %v", err)
	}

	// Depth 2 → exceeds MaxDepth=2.
	ctx2 := context.WithValue(ctx, depthKey{}, 2)
	_, err = rt.SpawnAgent(ctx2, newTestAgent("a3", "fail"), "input")
	if !errors.Is(err, ErrMaxDepth) {
		t.Errorf("depth 2 error = %v, want ErrMaxDepth", err)
	}
}

func TestCostBudgetEnforcement(t *testing.T) {
	rt := New("rt", WithConfig(Config{CostBudget: 0.001}))
	ctx := context.Background()

	// First spawn succeeds (cost is 0 before).
	_, err := rt.SpawnAgent(ctx, newTestAgent("a1", "ok"), "input")
	if err != nil {
		t.Fatalf("first spawn should succeed: %v", err)
	}

	// The mock agent has some cost from the cost tracker.
	// Second spawn should fail because budget is tiny and first spawn used some.
	// Force the cost to exceed budget.
	rt.mu.Lock()
	rt.totalCost = 0.001 // Force to budget limit.
	rt.mu.Unlock()

	_, err = rt.SpawnAgent(ctx, newTestAgent("a2", "fail"), "input")
	if !errors.Is(err, ErrCostBudget) {
		t.Errorf("second spawn error = %v, want ErrCostBudget", err)
	}
}

func TestMaxConcurrentEnforcement(t *testing.T) {
	rt := New("rt", WithConfig(Config{MaxConcurrent: 2}))
	ctx := context.Background()

	// Fill both slots.
	slow1 := newSlowAgent("slow1", 200*time.Millisecond)
	slow2 := newSlowAgent("slow2", 200*time.Millisecond)

	done := make(chan struct{})
	go func() {
		_, _ = rt.SpawnAgent(ctx, slow1, "input")
		done <- struct{}{}
	}()
	go func() {
		_, _ = rt.SpawnAgent(ctx, slow2, "input")
		done <- struct{}{}
	}()

	// Give the goroutines time to start and acquire slots.
	time.Sleep(50 * time.Millisecond)

	// Third spawn with a short timeout should fail because slots are full.
	shortCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	_, err := rt.SpawnAgent(shortCtx, newTestAgent("blocked", "fail"), "input")
	if err == nil {
		t.Error("expected error when all slots are full")
	}

	// Wait for slow agents to finish.
	<-done
	<-done
}

func TestCascadingCancellation(t *testing.T) {
	rt := New("rt")
	ctx, cancel := context.WithCancel(context.Background())

	slow := newSlowAgent("slow", 5*time.Second)

	errCh := make(chan error, 1)
	go func() {
		_, err := rt.SpawnAgent(ctx, slow, "input")
		errCh <- err
	}()

	// Cancel the parent context.
	time.Sleep(50 * time.Millisecond)
	cancel()

	err := <-errCh
	if err == nil {
		t.Fatal("expected error from canceled context")
	}
}

func TestGoAndFuture(t *testing.T) {
	rt := New("rt")
	ctx := context.Background()

	f := rt.Go(ctx, "async-task", func(ctx context.Context) (string, error) {
		return "async-result", nil
	})

	output, err := f.Wait(ctx)
	if err != nil {
		t.Fatalf("Future.Wait error: %v", err)
	}
	if output != "async-result" {
		t.Errorf("output = %q, want %q", output, "async-result")
	}
}

func TestGoWithSpawn(t *testing.T) {
	rt := New("rt")
	ctx := context.Background()

	f := rt.Go(ctx, "spawn-async", func(ctx context.Context) (string, error) {
		result, err := rt.SpawnAgent(ctx, newTestAgent("inner", "inner-result"), "input")
		if err != nil {
			return "", err
		}
		return result.Message.Content, nil
	})

	output, err := f.Wait(ctx)
	if err != nil {
		t.Fatalf("Future.Wait error: %v", err)
	}
	if output != "inner-result" {
		t.Errorf("output = %q, want %q", output, "inner-result")
	}

	// The SpawnAgent call should have recorded a child.
	res := rt.Result()
	if len(res.Children) != 1 {
		t.Fatalf("Children = %d, want 1", len(res.Children))
	}
}

func TestGoError(t *testing.T) {
	rt := New("rt")
	ctx := context.Background()

	f := rt.Go(ctx, "fail-task", func(ctx context.Context) (string, error) {
		return "", errors.New("task failed")
	})

	_, err := f.Wait(ctx)
	if err == nil || err.Error() != "task failed" {
		t.Errorf("error = %v, want 'task failed'", err)
	}
}

func TestFutureContextCancel(t *testing.T) {
	rt := New("rt")
	goCtx, goCancel := context.WithCancel(context.Background())

	f := rt.Go(goCtx, "slow-task", func(ctx context.Context) (string, error) {
		select {
		case <-time.After(10 * time.Second):
			return "done", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	})

	// Wait with a short timeout.
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer waitCancel()

	_, err := f.Wait(waitCtx)
	if err == nil {
		t.Fatal("expected context deadline exceeded")
	}

	// Clean up: cancel the goroutine's context so it exits promptly.
	goCancel()
	rt.Wait()
}

func TestFutureDone(t *testing.T) {
	rt := New("rt")
	ctx := context.Background()

	f := rt.Go(ctx, "quick", func(ctx context.Context) (string, error) {
		return "ok", nil
	})

	// Wait for completion via Done channel.
	select {
	case <-f.Done():
		// OK
	case <-time.After(2 * time.Second):
		t.Fatal("Future.Done did not close")
	}
}

func TestWait(t *testing.T) {
	rt := New("rt")
	ctx := context.Background()

	var count atomic.Int32

	for i := 0; i < 5; i++ {
		rt.Go(ctx, "task", func(ctx context.Context) (string, error) {
			time.Sleep(50 * time.Millisecond)
			count.Add(1)
			return "done", nil
		})
	}

	rt.Wait()

	if count.Load() != 5 {
		t.Errorf("count = %d, want 5", count.Load())
	}
}

func TestMultipleSpawns(t *testing.T) {
	rt := New("rt")
	ctx := context.Background()

	// Spawn several children of different types.
	_, err := rt.SpawnAgent(ctx, newTestAgent("agent1", "a1-out"), "input")
	if err != nil {
		t.Fatalf("SpawnAgent error: %v", err)
	}

	_, err = rt.SpawnAgent(ctx, newTestAgent("agent2", "a2-out"), "input")
	if err != nil {
		t.Fatalf("SpawnAgent error: %v", err)
	}

	res := rt.Result()
	if len(res.Children) != 2 {
		t.Fatalf("Children = %d, want 2", len(res.Children))
	}
	if res.Children[0].Output != "a1-out" {
		t.Errorf("child[0] output = %q, want %q", res.Children[0].Output, "a1-out")
	}
	if res.Children[1].Output != "a2-out" {
		t.Errorf("child[1] output = %q, want %q", res.Children[1].Output, "a2-out")
	}
}

func TestResultAggregation(t *testing.T) {
	rt := New("rt")
	ctx := context.Background()

	// Spawn two agents.
	_, _ = rt.SpawnAgent(ctx, newTestAgent("a1", "out1"), "input")
	_, _ = rt.SpawnAgent(ctx, newTestAgent("a2", "out2"), "input")

	res := rt.Result()
	if res.RunID == "" {
		t.Error("RunID should not be empty")
	}
	if len(res.Children) != 2 {
		t.Fatalf("Children = %d, want 2", len(res.Children))
	}
	// Usage should be aggregated from both agents.
	if res.TotalUsage.PromptTokens != 20 {
		t.Errorf("TotalUsage.PromptTokens = %d, want 20", res.TotalUsage.PromptTokens)
	}
	if res.TotalUsage.CompletionTokens != 10 {
		t.Errorf("TotalUsage.CompletionTokens = %d, want 10", res.TotalUsage.CompletionTokens)
	}
	if res.TotalUsage.TotalTokens != 30 {
		t.Errorf("TotalUsage.TotalTokens = %d, want 30", res.TotalUsage.TotalTokens)
	}
}

func TestRemainingBudget(t *testing.T) {
	// No budget → returns -1.
	rt := New("rt")
	if rt.RemainingBudget() != -1 {
		t.Errorf("RemainingBudget = %f, want -1", rt.RemainingBudget())
	}

	// With budget.
	rt2 := New("rt2", WithConfig(Config{CostBudget: 1.00}))
	if rt2.RemainingBudget() != 1.00 {
		t.Errorf("RemainingBudget = %f, want 1.00", rt2.RemainingBudget())
	}

	// After spending.
	rt2.mu.Lock()
	rt2.totalCost = 0.75
	rt2.mu.Unlock()
	if rt2.RemainingBudget() != 0.25 {
		t.Errorf("RemainingBudget = %f, want 0.25", rt2.RemainingBudget())
	}
}

func TestTraceSpans(t *testing.T) {
	tracer := trace.NewInMemory()
	rt := New("rt", WithTracer(tracer))
	ctx := context.Background()

	_, _ = rt.SpawnAgent(ctx, newTestAgent("traced", "out"), "input")

	spans := tracer.Spans()
	found := false
	for _, s := range spans {
		if s.Name == "dynamic.spawn_agent" {
			found = true
			if s.Attributes["dynamic.child.name"] != "traced" {
				t.Errorf("child name attr = %q, want %q",
					s.Attributes["dynamic.child.name"], "traced")
			}
			if s.Attributes["dynamic.child.type"] != "agent" {
				t.Errorf("child type attr = %q, want %q",
					s.Attributes["dynamic.child.type"], "agent")
			}
		}
	}
	if !found {
		t.Error("dynamic.spawn_agent span not found")
	}
}

func TestTraceSpanGo(t *testing.T) {
	tracer := trace.NewInMemory()
	rt := New("rt", WithTracer(tracer))
	ctx := context.Background()

	f := rt.Go(ctx, "bg-task", func(ctx context.Context) (string, error) {
		return "done", nil
	})
	_, _ = f.Wait(ctx)

	spans := tracer.Spans()
	found := false
	for _, s := range spans {
		if s.Name == "dynamic.go" {
			found = true
			if s.Attributes["dynamic.child.name"] != "bg-task" {
				t.Errorf("child name = %q, want %q",
					s.Attributes["dynamic.child.name"], "bg-task")
			}
		}
	}
	if !found {
		t.Error("dynamic.go span not found")
	}
}

func TestParallelSpawns(t *testing.T) {
	rt := New("rt", WithConfig(Config{MaxConcurrent: 3}))
	ctx := context.Background()

	// Launch 3 concurrent agents via Go.
	var futures []*Future
	for i := 0; i < 3; i++ {
		a := newTestAgent("parallel", "parallel-out")
		f := rt.Go(ctx, "parallel", func(ctx context.Context) (string, error) {
			r, err := rt.SpawnAgent(ctx, a, "input")
			if err != nil {
				return "", err
			}
			return r.Message.Content, nil
		})
		futures = append(futures, f)
	}

	for _, f := range futures {
		out, err := f.Wait(ctx)
		if err != nil {
			t.Fatalf("parallel wait error: %v", err)
		}
		if out != "parallel-out" {
			t.Errorf("parallel output = %q, want %q", out, "parallel-out")
		}
	}

	res := rt.Result()
	if len(res.Children) != 3 {
		t.Errorf("Children = %d, want 3", len(res.Children))
	}
}

func TestChildContextIncreasesDepth(t *testing.T) {
	rt := New("rt")
	ctx := context.Background()

	// Depth starts at 0.
	childCtx := rt.childContext(ctx)
	if DepthFromContext(childCtx) != 1 {
		t.Errorf("child depth = %d, want 1", DepthFromContext(childCtx))
	}

	// Nested.
	grandchildCtx := rt.childContext(childCtx)
	if DepthFromContext(grandchildCtx) != 2 {
		t.Errorf("grandchild depth = %d, want 2", DepthFromContext(grandchildCtx))
	}
}

func TestSpawnAllTypes(t *testing.T) {
	rt := New("rt")
	ctx := context.Background()

	// Agent.
	_, err := rt.SpawnAgent(ctx, newTestAgent("a", "agent-out"), "input")
	if err != nil {
		t.Fatalf("SpawnAgent: %v", err)
	}

	// Team.
	_, err = rt.SpawnTeam(ctx, newTestTeam("t", "team-out"), "input")
	if err != nil {
		t.Fatalf("SpawnTeam: %v", err)
	}

	// Pipeline.
	_, err = rt.SpawnPipeline(ctx, newTestPipeline("p"), "input")
	if err != nil {
		t.Fatalf("SpawnPipeline: %v", err)
	}

	// Graph.
	_, err = rt.SpawnGraph(ctx, newTestGraph("g"), "input")
	if err != nil {
		t.Fatalf("SpawnGraph: %v", err)
	}

	res := rt.Result()
	if len(res.Children) != 4 {
		t.Fatalf("Children = %d, want 4", len(res.Children))
	}

	types := map[string]bool{}
	for _, c := range res.Children {
		types[c.Type] = true
	}
	for _, typ := range []string{"agent", "team", "pipeline", "graph"} {
		if !types[typ] {
			t.Errorf("missing child type %q", typ)
		}
	}
}
