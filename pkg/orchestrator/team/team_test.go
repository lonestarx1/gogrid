package team

import (
	"context"
	"errors"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/memory/shared"
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
			Model:   "mock-model",
		},
	}
}

func (m *mockProvider) Complete(_ context.Context, _ llm.Params) (*llm.Response, error) {
	m.calls.Add(1)
	return m.response, nil
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

// errorProvider always returns an error.
type errorProvider struct{ err error }

func (e *errorProvider) Complete(_ context.Context, _ llm.Params) (*llm.Response, error) {
	return nil, e.err
}

func newTestAgent(name, content string) *agent.Agent {
	return agent.New(name,
		agent.WithProvider(newMockProvider(content)),
		agent.WithModel("mock-model"),
	)
}

func TestNewTeam(t *testing.T) {
	tm := New("test-team")
	if tm.Name() != "test-team" {
		t.Errorf("Name = %q, want %q", tm.Name(), "test-team")
	}
	if tm.Bus() == nil {
		t.Error("Bus is nil")
	}
	if tm.SharedMemory() == nil {
		t.Error("SharedMemory is nil")
	}
}

func TestTeamOptions(t *testing.T) {
	mem := shared.New()
	bus := NewBus()
	tracer := trace.NewInMemory()

	tm := New("team",
		WithMembers(
			Member{Agent: newTestAgent("a", "response-a")},
			Member{Agent: newTestAgent("b", "response-b"), Role: "critic"},
		),
		WithStrategy(FirstResponse{}),
		WithSharedMemory(mem),
		WithBus(bus),
		WithTracer(tracer),
		WithConfig(Config{MaxRounds: 3, Timeout: 10 * time.Second, CostBudget: 5.0}),
	)

	if len(tm.members) != 2 {
		t.Errorf("members len = %d, want 2", len(tm.members))
	}
	if tm.strategy.Name() != "first_response" {
		t.Errorf("strategy = %q, want %q", tm.strategy.Name(), "first_response")
	}
	if tm.memory != mem {
		t.Error("shared memory not set")
	}
	if tm.bus != bus {
		t.Error("bus not set")
	}
}

func TestRunRequiresMembers(t *testing.T) {
	tm := New("empty-team")
	_, err := tm.Run(context.Background(), "hello")
	if err == nil {
		t.Fatal("Run without members: expected error")
	}
}

func TestRunUnanimous(t *testing.T) {
	tm := New("debate-team",
		WithMembers(
			Member{Agent: newTestAgent("alice", "I agree")},
			Member{Agent: newTestAgent("bob", "Me too")},
			Member{Agent: newTestAgent("charlie", "Same here")},
		),
		WithStrategy(Unanimous{}),
	)

	result, err := tm.Run(context.Background(), "should we proceed?")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.RunID == "" {
		t.Error("RunID is empty")
	}
	if result.Rounds != 1 {
		t.Errorf("Rounds = %d, want 1", result.Rounds)
	}
	if len(result.Responses) != 3 {
		t.Errorf("Responses len = %d, want 3", len(result.Responses))
	}
	if result.Decision.Content == "" {
		t.Error("Decision is empty")
	}
	// Decision should contain all three agents' responses.
	for _, name := range []string{"alice", "bob", "charlie"} {
		if _, ok := result.Responses[name]; !ok {
			t.Errorf("missing response from %q", name)
		}
	}
}

func TestRunFirstResponse(t *testing.T) {
	fast := agent.New("fast",
		agent.WithProvider(&slowProvider{
			delay: 10 * time.Millisecond,
			response: &llm.Response{
				Message: llm.NewAssistantMessage("quick answer"),
				Usage:   llm.Usage{PromptTokens: 5, CompletionTokens: 3, TotalTokens: 8},
				Model:   "mock-model",
			},
		}),
		agent.WithModel("mock-model"),
	)

	slow := agent.New("slow",
		agent.WithProvider(&slowProvider{
			delay: 5 * time.Second,
			response: &llm.Response{
				Message: llm.NewAssistantMessage("slow answer"),
				Usage:   llm.Usage{PromptTokens: 5, CompletionTokens: 3, TotalTokens: 8},
				Model:   "mock-model",
			},
		}),
		agent.WithModel("mock-model"),
	)

	tm := New("race-team",
		WithMembers(
			Member{Agent: fast},
			Member{Agent: slow},
		),
		WithStrategy(FirstResponse{}),
		WithConfig(Config{Timeout: 5 * time.Second}),
	)

	result, err := tm.Run(context.Background(), "who responds first?")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// FirstResponse should return after the fast agent.
	if result.Decision.Content != "quick answer" {
		t.Errorf("Decision = %q, want %q", result.Decision.Content, "quick answer")
	}
}

func TestRunMajority(t *testing.T) {
	tm := New("vote-team",
		WithMembers(
			Member{Agent: newTestAgent("a", "yes")},
			Member{Agent: newTestAgent("b", "yes")},
			Member{Agent: newTestAgent("c", "no")},
		),
		WithStrategy(Majority{}),
	)

	result, err := tm.Run(context.Background(), "vote please")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Majority requires >50%, so at least 2 of 3.
	if len(result.Responses) < 2 {
		t.Errorf("Responses len = %d, want >= 2", len(result.Responses))
	}
}

func TestRunMultipleRounds(t *testing.T) {
	// Custom strategy that requires 2 rounds.
	roundTracker := &roundCountingStrategy{targetRound: 2}

	tm := New("multi-round",
		WithMembers(
			Member{Agent: newTestAgent("a", "round response")},
			Member{Agent: newTestAgent("b", "round response")},
		),
		WithStrategy(roundTracker),
		WithConfig(Config{MaxRounds: 3}),
	)

	result, err := tm.Run(context.Background(), "discuss this")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Rounds != 2 {
		t.Errorf("Rounds = %d, want 2", result.Rounds)
	}
}

// roundCountingStrategy only reaches consensus on the target round.
type roundCountingStrategy struct {
	targetRound int
	calls       int
}

func (s *roundCountingStrategy) Name() string { return "round_counting" }
func (s *roundCountingStrategy) Evaluate(total int, responses map[string]string) (string, bool) {
	// Only reach consensus when we've been called targetRound times with full responses.
	if len(responses) < total {
		return "", false
	}
	s.calls++
	if s.calls >= s.targetRound {
		return combineResponses(responses), true
	}
	return "", false
}

func TestRunWithTracer(t *testing.T) {
	tracer := trace.NewInMemory()

	tm := New("traced-team",
		WithMembers(
			Member{Agent: newTestAgent("a", "response")},
		),
		WithStrategy(Unanimous{}),
		WithTracer(tracer),
	)

	_, err := tm.Run(context.Background(), "test tracing")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	spans := tracer.Spans()
	names := make(map[string]bool)
	for _, s := range spans {
		names[s.Name] = true
	}
	if !names["team.run"] {
		t.Error("missing team.run span")
	}
	if !names["team.round"] {
		t.Error("missing team.round span")
	}
}

func TestRunTracerAttributes(t *testing.T) {
	tracer := trace.NewInMemory()

	tm := New("attr-team",
		WithMembers(
			Member{Agent: newTestAgent("a", "resp")},
			Member{Agent: newTestAgent("b", "resp")},
		),
		WithStrategy(Unanimous{}),
		WithTracer(tracer),
	)

	_, err := tm.Run(context.Background(), "hi")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	spans := tracer.Spans()
	for _, s := range spans {
		if s.Name == "team.run" {
			if s.Attributes["team.name"] != "attr-team" {
				t.Errorf("team.name attr = %q, want %q", s.Attributes["team.name"], "attr-team")
			}
			if s.Attributes["team.members"] != "2" {
				t.Errorf("team.members attr = %q, want %q", s.Attributes["team.members"], "2")
			}
			if s.Attributes["team.strategy"] != "unanimous" {
				t.Errorf("team.strategy attr = %q, want %q", s.Attributes["team.strategy"], "unanimous")
			}
		}
	}
}

func TestRunSharedMemory(t *testing.T) {
	ctx := context.Background()
	mem := shared.New()

	tm := New("shared-mem-team",
		WithMembers(
			Member{Agent: newTestAgent("writer-a", "data from a")},
			Member{Agent: newTestAgent("writer-b", "data from b")},
		),
		WithSharedMemory(mem),
	)

	_, err := tm.Run(ctx, "write something")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Shared memory should have entries for each agent's round.
	msgA, _ := mem.Load(ctx, "writer-a/round/1")
	if len(msgA) != 1 {
		t.Errorf("shared memory for writer-a: len = %d, want 1", len(msgA))
	}
	msgB, _ := mem.Load(ctx, "writer-b/round/1")
	if len(msgB) != 1 {
		t.Errorf("shared memory for writer-b: len = %d, want 1", len(msgB))
	}
}

func TestRunBusNotifications(t *testing.T) {
	bus := NewBus()
	ch, unsub := bus.Subscribe("team.response", 10)
	defer unsub()

	tm := New("bus-team",
		WithMembers(
			Member{Agent: newTestAgent("a", "hello")},
			Member{Agent: newTestAgent("b", "world")},
		),
		WithBus(bus),
	)

	_, err := tm.Run(context.Background(), "test bus")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Should have received 2 messages on the bus.
	received := 0
	for {
		select {
		case msg := <-ch:
			received++
			if msg.Metadata["round"] != "1" {
				t.Errorf("round metadata = %q, want %q", msg.Metadata["round"], "1")
			}
		default:
			goto DONE
		}
	}
DONE:
	if received != 2 {
		t.Errorf("bus messages = %d, want 2", received)
	}
}

func TestRunCostBudget(t *testing.T) {
	// Use a model with known pricing so agents report nonzero cost.
	// gpt-4o: 2.50/1M prompt + 10.00/1M completion.
	// Each agent: 10 prompt + 5 completion = ~0.000075 USD.
	// Two agents per round = ~0.000150 USD.
	costProvider := func(content string) *mockProvider {
		return &mockProvider{
			response: &llm.Response{
				Message: llm.NewAssistantMessage(content),
				Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
				Model:   "gpt-4o",
			},
		}
	}

	agentA := agent.New("a",
		agent.WithProvider(costProvider("cheap")),
		agent.WithModel("gpt-4o"),
	)
	agentB := agent.New("b",
		agent.WithProvider(costProvider("cheap")),
		agent.WithModel("gpt-4o"),
	)

	tm := New("budget-team",
		WithMembers(
			Member{Agent: agentA},
			Member{Agent: agentB},
		),
		WithConfig(Config{MaxRounds: 5, CostBudget: 0.000001}),
		WithStrategy(&neverStrategy{}),
	)

	result, err := tm.Run(context.Background(), "stay cheap")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Should stop early because cost exceeds tiny budget after round 1.
	if result.Rounds > 2 {
		t.Errorf("Rounds = %d, expected <= 2 with tiny budget", result.Rounds)
	}
	if result.TotalCost == 0 {
		t.Error("TotalCost should be nonzero with gpt-4o pricing")
	}
}

// neverStrategy never reaches consensus.
type neverStrategy struct{}

func (n *neverStrategy) Name() string { return "never" }
func (n *neverStrategy) Evaluate(_ int, _ map[string]string) (string, bool) {
	return "", false
}

func TestRunTimeout(t *testing.T) {
	slow := agent.New("slow",
		agent.WithProvider(&slowProvider{
			delay: 10 * time.Second,
			response: &llm.Response{
				Message: llm.NewAssistantMessage("slow"),
				Usage:   llm.Usage{PromptTokens: 5, CompletionTokens: 3, TotalTokens: 8},
				Model:   "mock-model",
			},
		}),
		agent.WithModel("mock-model"),
	)

	tm := New("timeout-team",
		WithMembers(Member{Agent: slow}),
		WithConfig(Config{Timeout: 50 * time.Millisecond}),
	)

	_, err := tm.Run(context.Background(), "hurry")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRunContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tm := New("canceled-team",
		WithMembers(Member{Agent: newTestAgent("a", "ok")}),
	)

	_, err := tm.Run(ctx, "hello")
	if err == nil {
		t.Fatal("expected context canceled error")
	}
}

func TestRunAgentError(t *testing.T) {
	good := newTestAgent("good", "I'm fine")
	bad := agent.New("bad",
		agent.WithProvider(&errorProvider{err: errors.New("provider down")}),
		agent.WithModel("mock-model"),
	)

	tm := New("mixed-team",
		WithMembers(
			Member{Agent: good},
			Member{Agent: bad},
		),
		WithStrategy(Unanimous{}),
	)

	result, err := tm.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// The good agent should have a response, the bad one should not.
	if _, ok := result.Responses["good"]; !ok {
		t.Error("missing response from good agent")
	}
	if _, ok := result.Responses["bad"]; ok {
		t.Error("bad agent should not have a response")
	}
	// Decision should still be formed from available responses.
	if result.Decision.Content == "" {
		t.Error("Decision should not be empty when at least one agent succeeded")
	}
}

func TestRunAggregatesUsage(t *testing.T) {
	tm := New("usage-team",
		WithMembers(
			Member{Agent: newTestAgent("a", "resp")},
			Member{Agent: newTestAgent("b", "resp")},
		),
	)

	result, err := tm.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Each mock agent uses 10 prompt + 5 completion = 15 total.
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

func TestRunMemoryStats(t *testing.T) {
	tm := New("stats-team",
		WithMembers(
			Member{Agent: newTestAgent("a", "data")},
		),
	)

	result, err := tm.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.MemoryStats == nil {
		t.Fatal("MemoryStats is nil")
	}
	if result.MemoryStats.Keys < 1 {
		t.Errorf("MemoryStats.Keys = %d, want >= 1", result.MemoryStats.Keys)
	}
}

func TestRunSingleAgent(t *testing.T) {
	tm := New("solo",
		WithMembers(Member{Agent: newTestAgent("solo", "just me")}),
	)

	result, err := tm.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.Responses) != 1 {
		t.Errorf("Responses len = %d, want 1", len(result.Responses))
	}
}

func TestBuildRoundInput(t *testing.T) {
	tm := New("test",
		WithMembers(
			Member{Agent: newTestAgent("alice", ""), Role: "advocate"},
			Member{Agent: newTestAgent("bob", "")},
		),
	)

	results := map[string]*agent.Result{
		"alice": {Message: llm.NewAssistantMessage("alice says yes")},
		"bob":   {Message: llm.NewAssistantMessage("bob says no")},
	}

	input := tm.buildRoundInput("original question", 2, results)

	if !contains(input, "original question") {
		t.Error("missing original question")
	}
	if !contains(input, "Round 2") {
		t.Error("missing round number")
	}
	if !contains(input, "alice (advocate)") {
		t.Error("missing alice with role")
	}
	if !contains(input, "alice says yes") {
		t.Error("missing alice's response")
	}
	if !contains(input, "bob: bob says no") {
		t.Error("missing bob's response (bob has no role)")
	}
}

func TestRunMaxRoundsDefault(t *testing.T) {
	tm := New("default-rounds",
		WithMembers(
			Member{Agent: newTestAgent("a", "resp")},
		),
		WithStrategy(&neverStrategy{}),
		// MaxRounds defaults to 1.
	)

	result, err := tm.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Rounds != 1 {
		t.Errorf("Rounds = %d, want 1 (default)", result.Rounds)
	}
}

func TestRunMultiRoundBusMessages(t *testing.T) {
	bus := NewBus()
	ch, unsub := bus.Subscribe("team.response", 20)
	defer unsub()

	tm := New("multi-round-bus",
		WithMembers(
			Member{Agent: newTestAgent("a", "resp")},
		),
		WithBus(bus),
		WithStrategy(&neverStrategy{}),
		WithConfig(Config{MaxRounds: 3}),
	)

	_, err := tm.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	received := 0
	rounds := make(map[string]bool)
	for {
		select {
		case msg := <-ch:
			received++
			rounds[msg.Metadata["round"]] = true
		default:
			goto DONE
		}
	}
DONE:
	if received != 3 {
		t.Errorf("bus messages = %d, want 3 (one per round)", received)
	}
	for i := 1; i <= 3; i++ {
		if !rounds[strconv.Itoa(i)] {
			t.Errorf("missing bus message for round %d", i)
		}
	}
}

func TestRunWithCoordinator(t *testing.T) {
	coordinator := newTestAgent("coordinator", "synthesized decision from all inputs")

	tm := New("coordinated-team",
		WithMembers(
			Member{Agent: newTestAgent("alice", "I think yes"), Role: "advocate"},
			Member{Agent: newTestAgent("bob", "I think no"), Role: "skeptic"},
		),
		WithCoordinator(coordinator),
		WithStrategy(Unanimous{}),
	)

	result, err := tm.Run(context.Background(), "should we proceed?")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Decision should come from coordinator, not concatenated responses.
	if result.Decision.Content != "synthesized decision from all inputs" {
		t.Errorf("Decision = %q, want coordinator's response", result.Decision.Content)
	}

	// Coordinator result should be in Responses.
	if _, ok := result.Responses["coordinator"]; !ok {
		t.Error("missing coordinator in Responses")
	}

	// Member responses should also be present.
	if _, ok := result.Responses["alice"]; !ok {
		t.Error("missing alice in Responses")
	}
	if _, ok := result.Responses["bob"]; !ok {
		t.Error("missing bob in Responses")
	}
}

func TestRunCoordinatorCostIncluded(t *testing.T) {
	coordinator := newTestAgent("coord", "final answer")

	tm := New("cost-coord",
		WithMembers(
			Member{Agent: newTestAgent("a", "resp")},
		),
		WithCoordinator(coordinator),
	)

	result, err := tm.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Two agents ran (member + coordinator), each with 15 total tokens.
	if result.TotalUsage.TotalTokens != 30 {
		t.Errorf("TotalTokens = %d, want 30 (member + coordinator)", result.TotalUsage.TotalTokens)
	}
}

func TestRunCoordinatorTraceSpan(t *testing.T) {
	tracer := trace.NewInMemory()
	coordinator := newTestAgent("lead", "decision")

	tm := New("traced-coord",
		WithMembers(
			Member{Agent: newTestAgent("a", "resp")},
		),
		WithCoordinator(coordinator),
		WithTracer(tracer),
	)

	_, err := tm.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	spans := tracer.Spans()
	found := false
	for _, s := range spans {
		if s.Name == "team.coordinator" {
			found = true
			if s.Attributes["team.coordinator.name"] != "lead" {
				t.Errorf("coordinator name attr = %q, want %q",
					s.Attributes["team.coordinator.name"], "lead")
			}
		}
	}
	if !found {
		t.Error("missing team.coordinator span")
	}
}

func TestRunCoordinatorBusMessage(t *testing.T) {
	bus := NewBus()
	ch, unsub := bus.Subscribe("team.coordinator", 10)
	defer unsub()

	coordinator := newTestAgent("lead", "final call")

	tm := New("bus-coord",
		WithMembers(
			Member{Agent: newTestAgent("a", "input")},
		),
		WithCoordinator(coordinator),
		WithBus(bus),
	)

	_, err := tm.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	select {
	case msg := <-ch:
		if msg.From != "lead" {
			t.Errorf("bus msg From = %q, want %q", msg.From, "lead")
		}
		if msg.Content != "final call" {
			t.Errorf("bus msg Content = %q, want %q", msg.Content, "final call")
		}
	default:
		t.Error("no coordinator message on bus")
	}
}

func TestRunCoordinatorError(t *testing.T) {
	badCoord := agent.New("bad-coord",
		agent.WithProvider(&errorProvider{err: errors.New("coord down")}),
		agent.WithModel("mock-model"),
	)

	tm := New("error-coord",
		WithMembers(
			Member{Agent: newTestAgent("a", "resp")},
		),
		WithCoordinator(badCoord),
	)

	result, err := tm.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// When coordinator fails, decision falls back to combined member responses.
	if result.Decision.Content == "" {
		t.Error("Decision should not be empty when coordinator fails")
	}
	if _, ok := result.Responses["bad-coord"]; ok {
		t.Error("failed coordinator should not be in Responses")
	}
}

func TestRunWithoutCoordinator(t *testing.T) {
	// Verify existing behavior is unchanged when no coordinator is set.
	tm := New("no-coord",
		WithMembers(
			Member{Agent: newTestAgent("a", "response a")},
			Member{Agent: newTestAgent("b", "response b")},
		),
		WithStrategy(Unanimous{}),
	)

	result, err := tm.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Decision should be combined responses (no coordinator).
	if !contains(result.Decision.Content, "response a") {
		t.Error("Decision missing agent a's response")
	}
	if !contains(result.Decision.Content, "response b") {
		t.Error("Decision missing agent b's response")
	}
}

func TestBuildCoordinatorPrompt(t *testing.T) {
	tm := New("test",
		WithMembers(
			Member{Agent: newTestAgent("alice", ""), Role: "advocate"},
			Member{Agent: newTestAgent("bob", "")},
		),
	)

	results := map[string]*agent.Result{
		"alice": {Message: llm.NewAssistantMessage("alice says yes")},
		"bob":   {Message: llm.NewAssistantMessage("bob says no")},
	}

	prompt := tm.buildCoordinatorPrompt("original question", results)

	if !contains(prompt, "original question") {
		t.Error("missing original question")
	}
	if !contains(prompt, "alice (advocate)") {
		t.Error("missing alice with role")
	}
	if !contains(prompt, "alice says yes") {
		t.Error("missing alice's response")
	}
	if !contains(prompt, "bob") {
		t.Error("missing bob")
	}
	if !contains(prompt, "bob says no") {
		t.Error("missing bob's response")
	}
	if !contains(prompt, "Synthesize") {
		t.Error("missing synthesis instruction")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
