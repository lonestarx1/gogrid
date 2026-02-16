package team

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lonestarx1/gogrid/internal/id"
	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/memory"
	"github.com/lonestarx1/gogrid/pkg/memory/shared"
	"github.com/lonestarx1/gogrid/pkg/trace"
)

// Member represents an agent participating in the team.
type Member struct {
	// Agent is the configured agent instance.
	Agent *agent.Agent
	// Role is an optional description added to the round context
	// (e.g., "skeptic", "advocate").
	Role string
}

// Config controls team execution behavior.
type Config struct {
	// MaxRounds limits the number of discussion rounds. 0 defaults to 1.
	MaxRounds int
	// Timeout is the maximum wall-clock duration for the entire team run.
	// Zero means no timeout (relies on the caller's context).
	Timeout time.Duration
	// CostBudget is the maximum total cost in USD across all agents.
	// Zero means no budget limit.
	CostBudget float64
}

// Result is returned by Team.Run with the outcome of the team execution.
type Result struct {
	// RunID uniquely identifies this execution.
	RunID string
	// Decision is the team's consensus decision.
	Decision llm.Message
	// Responses maps agent name to that agent's final result.
	Responses map[string]*agent.Result
	// Rounds is the number of discussion rounds that occurred.
	Rounds int
	// TotalCost is the aggregate cost in USD across all agents.
	TotalCost float64
	// TotalUsage is the aggregate token usage across all agents.
	TotalUsage llm.Usage
	// MemoryStats contains shared memory statistics, if available.
	MemoryStats *memory.Stats
}

// Team orchestrates multiple agents in a chat-room style discussion.
// Agents run concurrently each round, communicate via a shared message bus,
// and reach decisions through a pluggable consensus strategy.
//
// An optional coordinator agent can be set to synthesize a final decision
// from the member responses instead of using string concatenation.
type Team struct {
	name        string
	members     []Member
	coordinator *agent.Agent
	strategy    Strategy
	memory      *shared.Memory
	bus         *Bus
	tracer      trace.Tracer
	config      Config
}

// Option is a functional option for configuring a Team.
type Option func(*Team)

// New creates a Team with the given name and options.
func New(name string, opts ...Option) *Team {
	t := &Team{
		name:     name,
		strategy: Unanimous{},
		memory:   shared.New(),
		bus:      NewBus(),
		tracer:   trace.Noop{},
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// WithMembers sets the team's agent members.
func WithMembers(members ...Member) Option {
	return func(t *Team) {
		t.members = members
	}
}

// WithStrategy sets the consensus strategy.
func WithStrategy(s Strategy) Option {
	return func(t *Team) {
		t.strategy = s
	}
}

// WithSharedMemory sets the shared memory pool.
func WithSharedMemory(m *shared.Memory) Option {
	return func(t *Team) {
		t.memory = m
	}
}

// WithBus sets the message bus.
func WithBus(b *Bus) Option {
	return func(t *Team) {
		t.bus = b
	}
}

// WithTracer sets the tracer for observability.
func WithTracer(tr trace.Tracer) Option {
	return func(t *Team) {
		t.tracer = tr
	}
}

// WithConfig sets the team's execution configuration.
func WithConfig(c Config) Option {
	return func(t *Team) {
		t.config = c
	}
}

// WithCoordinator sets a coordinator agent that synthesizes the final
// team decision. After all rounds complete, the coordinator receives
// every member's response and produces a single coherent answer.
// Without a coordinator, member responses are concatenated.
func WithCoordinator(a *agent.Agent) Option {
	return func(t *Team) {
		t.coordinator = a
	}
}

// Name returns the team's name.
func (t *Team) Name() string { return t.name }

// Bus returns the team's message bus for external subscribers.
func (t *Team) Bus() *Bus { return t.bus }

// SharedMemory returns the team's shared memory pool.
func (t *Team) SharedMemory() *shared.Memory { return t.memory }

// Run executes the team discussion with the given input.
//
// The team loop:
//  1. Each round, all agents receive the input (plus prior round context).
//  2. Agents run concurrently within a round.
//  3. Responses are published to the bus and stored in shared memory.
//  4. The consensus strategy evaluates after each agent completes.
//  5. If consensus is reached, the round ends early.
//  6. If not, the next round begins with updated context.
func (t *Team) Run(ctx context.Context, input string) (*Result, error) {
	if len(t.members) == 0 {
		return nil, errors.New("team: at least one member is required")
	}

	// Apply timeout.
	if t.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, t.config.Timeout)
		defer cancel()
	}

	runID := id.New()

	// Start run span.
	ctx, runSpan := t.tracer.StartSpan(ctx, "team.run")
	runSpan.SetAttribute("team.name", t.name)
	runSpan.SetAttribute("team.run_id", runID)
	runSpan.SetAttribute("team.members", strconv.Itoa(len(t.members)))
	runSpan.SetAttribute("team.strategy", t.strategy.Name())
	if t.coordinator != nil {
		runSpan.SetAttribute("team.coordinator", t.coordinator.Name())
	}
	defer t.tracer.EndSpan(runSpan)

	maxRounds := t.config.MaxRounds
	if maxRounds <= 0 {
		maxRounds = 1
	}

	var totalCost float64
	var totalUsage llm.Usage
	agentResults := make(map[string]*agent.Result)
	var decision string
	rounds := 0

	for round := 1; round <= maxRounds; round++ {
		if err := ctx.Err(); err != nil {
			runSpan.SetError(err)
			return nil, fmt.Errorf("team: %w", err)
		}
		if t.config.CostBudget > 0 && totalCost >= t.config.CostBudget {
			break
		}

		rounds = round

		roundCtx, cancelRound := context.WithCancel(ctx)
		_, roundSpan := t.tracer.StartSpan(roundCtx, "team.round")
		roundSpan.SetAttribute("team.round", strconv.Itoa(round))

		// Build input for this round.
		roundInput := input
		if round > 1 {
			roundInput = t.buildRoundInput(input, round, agentResults)
		}

		// Run all agents concurrently.
		type agentResponse struct {
			name   string
			result *agent.Result
			err    error
		}
		responseCh := make(chan agentResponse, len(t.members))
		for _, m := range t.members {
			go func(member Member) {
				result, err := member.Agent.Run(roundCtx, roundInput)
				responseCh <- agentResponse{
					name:   member.Agent.Name(),
					result: result,
					err:    err,
				}
			}(m)
		}

		// Collect responses and check consensus after each.
		responses := make(map[string]string)
		consensusReached := false

		for i := 0; i < len(t.members); i++ {
			select {
			case resp := <-responseCh:
				if resp.err != nil {
					roundSpan.SetAttribute("agent."+resp.name+".error", resp.err.Error())
					continue
				}

				agentResults[resp.name] = resp.result
				responses[resp.name] = resp.result.Message.Content
				totalCost += resp.result.Cost
				totalUsage.PromptTokens += resp.result.Usage.PromptTokens
				totalUsage.CompletionTokens += resp.result.Usage.CompletionTokens
				totalUsage.TotalTokens += resp.result.Usage.TotalTokens

				// Publish to bus.
				t.bus.Publish(Message{
					From:    resp.name,
					Topic:   "team.response",
					Content: resp.result.Message.Content,
					Metadata: map[string]string{
						"round":  strconv.Itoa(round),
						"run_id": runID,
					},
				})

				// Store in shared memory.
				memKey := resp.name + "/round/" + strconv.Itoa(round)
				_ = t.memory.Save(roundCtx, memKey, []llm.Message{resp.result.Message})

				// Check consensus.
				d, reached := t.strategy.Evaluate(len(t.members), responses)
				if reached {
					decision = d
					consensusReached = true
					cancelRound()
				}

			case <-ctx.Done():
				cancelRound()
				t.tracer.EndSpan(roundSpan)
				runSpan.SetError(ctx.Err())
				return nil, fmt.Errorf("team: %w", ctx.Err())
			}

			if consensusReached {
				break
			}
		}

		roundSpan.SetAttribute("team.responses", strconv.Itoa(len(responses)))
		roundSpan.SetAttribute("team.consensus", strconv.FormatBool(consensusReached))
		t.tracer.EndSpan(roundSpan)
		cancelRound()

		if consensusReached {
			break
		}
	}

	// If no consensus after all rounds, combine whatever we have.
	if decision == "" && len(agentResults) > 0 {
		responses := make(map[string]string)
		for name, r := range agentResults {
			responses[name] = r.Message.Content
		}
		decision = combineResponses(responses)
	}

	// Run coordinator to synthesize a final decision from member responses.
	if t.coordinator != nil && len(agentResults) > 0 {
		coordResult, coordErr := t.runCoordinator(ctx, input, agentResults)
		if coordErr != nil {
			runSpan.SetAttribute("team.coordinator.error", coordErr.Error())
		} else {
			decision = coordResult.Message.Content
			agentResults[t.coordinator.Name()] = coordResult
			totalCost += coordResult.Cost
			totalUsage.PromptTokens += coordResult.Usage.PromptTokens
			totalUsage.CompletionTokens += coordResult.Usage.CompletionTokens
			totalUsage.TotalTokens += coordResult.Usage.TotalTokens
		}
	}

	// Collect shared memory stats.
	var memStats *memory.Stats
	if s, err := t.memory.Stats(ctx); err == nil {
		memStats = s
	}

	runSpan.SetAttribute("team.rounds", strconv.Itoa(rounds))
	runSpan.SetAttribute("team.cost_usd", fmt.Sprintf("%.6f", totalCost))

	return &Result{
		RunID:       runID,
		Decision:    llm.NewAssistantMessage(decision),
		Responses:   agentResults,
		Rounds:      rounds,
		TotalCost:   totalCost,
		TotalUsage:  totalUsage,
		MemoryStats: memStats,
	}, nil
}

// runCoordinator runs the coordinator agent with a prompt that includes the
// original input and all member responses, asking it to synthesize a decision.
func (t *Team) runCoordinator(ctx context.Context, input string, results map[string]*agent.Result) (*agent.Result, error) {
	_, coordSpan := t.tracer.StartSpan(ctx, "team.coordinator")
	coordSpan.SetAttribute("team.coordinator.name", t.coordinator.Name())
	defer t.tracer.EndSpan(coordSpan)

	prompt := t.buildCoordinatorPrompt(input, results)

	result, err := t.coordinator.Run(ctx, prompt)
	if err != nil {
		coordSpan.SetError(err)
		return nil, fmt.Errorf("coordinator: %w", err)
	}

	// Publish to bus.
	t.bus.Publish(Message{
		From:    t.coordinator.Name(),
		Topic:   "team.coordinator",
		Content: result.Message.Content,
	})

	return result, nil
}

// buildCoordinatorPrompt builds the coordinator's input with the original
// question and all member responses.
func (t *Team) buildCoordinatorPrompt(original string, results map[string]*agent.Result) string {
	var b strings.Builder
	b.WriteString("You are the team coordinator. Review the following team discussion and synthesize a final decision.\n\n")
	b.WriteString("## Original Task\n")
	b.WriteString(original)
	b.WriteString("\n\n## Team Member Responses\n\n")

	for _, m := range t.members {
		r, ok := results[m.Agent.Name()]
		if !ok {
			continue
		}
		b.WriteString("### ")
		b.WriteString(m.Agent.Name())
		if m.Role != "" {
			b.WriteString(" (")
			b.WriteString(m.Role)
			b.WriteString(")")
		}
		b.WriteString("\n")
		b.WriteString(r.Message.Content)
		b.WriteString("\n\n")
	}

	b.WriteString("## Your Task\nSynthesize the above responses into a single, coherent final decision.")
	return b.String()
}

// buildRoundInput formats the input for round 2+ with context from prior rounds.
func (t *Team) buildRoundInput(original string, round int, results map[string]*agent.Result) string {
	var b strings.Builder
	b.WriteString(original)
	b.WriteString("\n\n--- Team Discussion (Round ")
	b.WriteString(strconv.Itoa(round))
	b.WriteString(") ---\nPrevious responses from team members:\n\n")

	for _, m := range t.members {
		r, ok := results[m.Agent.Name()]
		if !ok {
			continue
		}
		b.WriteString(m.Agent.Name())
		if m.Role != "" {
			b.WriteString(" (")
			b.WriteString(m.Role)
			b.WriteString(")")
		}
		b.WriteString(": ")
		b.WriteString(r.Message.Content)
		b.WriteString("\n\n")
	}

	b.WriteString("Please provide your updated response considering the above discussion.")
	return b.String()
}
