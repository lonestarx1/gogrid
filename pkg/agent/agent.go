package agent

import (
	"context"
	"errors"
	"time"

	"github.com/lonestarx1/gogrid/internal/id"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/memory"
	"github.com/lonestarx1/gogrid/pkg/tool"
)

// Agent is the fundamental unit of work in GoGrid.
// It combines an LLM provider, tools, memory, and configuration
// to execute a task through an iterative tool-use loop.
type Agent struct {
	name         string
	instructions string
	tools        []tool.Tool
	model        string
	provider     llm.Provider
	memory       memory.Memory
	config       Config
}

// Config controls agent execution behavior.
type Config struct {
	// MaxTurns limits the number of LLM round-trips. 0 means no limit.
	MaxTurns int
	// MaxTokens limits the LLM response length per turn.
	MaxTokens int
	// Temperature controls LLM randomness (0.0-1.0).
	Temperature *float64
	// Timeout is the maximum wall-clock duration for a single Run call.
	// Zero means no timeout (relies on the caller's context).
	Timeout time.Duration
	// CostBudget is the maximum cost in USD for a single Run call.
	// Zero means no budget limit.
	CostBudget float64
}

// Result is returned by Agent.Run with the outcome of the execution.
type Result struct {
	// RunID uniquely identifies this execution.
	RunID string
	// Message is the agent's final response.
	Message llm.Message
	// History is the full conversation including tool calls and results.
	History []llm.Message
	// Usage is the aggregate token usage across all LLM calls in this run.
	Usage llm.Usage
	// Cost is the total estimated cost in USD for this run.
	Cost float64
	// Turns is the number of LLM round-trips that occurred.
	Turns int
}

// Option is a functional option for configuring an Agent.
type Option func(*Agent)

// New creates an Agent with the given name and options.
// The name is required and identifies the agent in traces and logs.
func New(name string, opts ...Option) *Agent {
	a := &Agent{
		name: name,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// WithInstructions sets the agent's system prompt.
func WithInstructions(instructions string) Option {
	return func(a *Agent) {
		a.instructions = instructions
	}
}

// WithTools sets the tools available to the agent.
func WithTools(tools ...tool.Tool) Option {
	return func(a *Agent) {
		a.tools = tools
	}
}

// WithModel sets the LLM model identifier.
func WithModel(model string) Option {
	return func(a *Agent) {
		a.model = model
	}
}

// WithProvider sets the LLM provider.
func WithProvider(provider llm.Provider) Option {
	return func(a *Agent) {
		a.provider = provider
	}
}

// WithMemory sets the agent's conversation memory.
func WithMemory(mem memory.Memory) Option {
	return func(a *Agent) {
		a.memory = mem
	}
}

// WithConfig sets the agent's execution configuration.
func WithConfig(config Config) Option {
	return func(a *Agent) {
		a.config = config
	}
}

// Name returns the agent's name.
func (a *Agent) Name() string { return a.name }

// Run executes the agent with the given user input.
// The agent loop (LLM calls, tool execution, iteration) is implemented in Phase 1.
func (a *Agent) Run(ctx context.Context, input string) (*Result, error) {
	if a.provider == nil {
		return nil, errors.New("agent: provider is required")
	}
	if a.model == "" {
		return nil, errors.New("agent: model is required")
	}

	return &Result{
		RunID:   id.New(),
		Message: llm.NewAssistantMessage(""),
		History: []llm.Message{},
	}, nil
}
