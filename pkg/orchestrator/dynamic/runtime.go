package dynamic

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/lonestarx1/gogrid/internal/id"
	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/orchestrator/graph"
	"github.com/lonestarx1/gogrid/pkg/orchestrator/pipeline"
	"github.com/lonestarx1/gogrid/pkg/orchestrator/team"
	"github.com/lonestarx1/gogrid/pkg/trace"
)

// Sentinel errors for resource governance violations.
var (
	// ErrMaxDepth is returned when a spawn would exceed the maximum
	// nesting depth.
	ErrMaxDepth = errors.New("dynamic: maximum nesting depth exceeded")
	// ErrCostBudget is returned when the runtime's cost budget is
	// exhausted and no more children can be spawned.
	ErrCostBudget = errors.New("dynamic: cost budget exhausted")
)

type runtimeKey struct{}
type depthKey struct{}

// FromContext retrieves the Runtime from the context, or nil if none.
func FromContext(ctx context.Context) *Runtime {
	r, _ := ctx.Value(runtimeKey{}).(*Runtime)
	return r
}

// DepthFromContext returns the current nesting depth from the context.
// Returns 0 if no depth has been set.
func DepthFromContext(ctx context.Context) int {
	d, _ := ctx.Value(depthKey{}).(int)
	return d
}

// Config controls resource governance for dynamic orchestration.
type Config struct {
	// MaxConcurrent is the maximum number of children that can execute
	// simultaneously. 0 means no limit.
	MaxConcurrent int
	// MaxDepth is the maximum nesting depth for recursive spawning.
	// 0 defaults to 10.
	MaxDepth int
	// CostBudget is the maximum total cost in USD across all children.
	// 0 means no limit.
	CostBudget float64
}

// ChildResult records the outcome of a single spawned child.
type ChildResult struct {
	// Name identifies the child.
	Name string
	// Type is "agent", "team", "pipeline", or "graph".
	Type string
	// Output is the child's final output content.
	Output string
	// Cost is the child's total cost in USD.
	Cost float64
	// Usage is the child's token usage.
	Usage llm.Usage
	// Error is non-nil if the child failed.
	Error error
}

// Result holds aggregate metrics from all children spawned by a Runtime.
type Result struct {
	// RunID uniquely identifies this runtime execution.
	RunID string
	// Children lists all spawned child results in order.
	Children []ChildResult
	// TotalCost is the aggregate cost across all children.
	TotalCost float64
	// TotalUsage is the aggregate token usage across all children.
	TotalUsage llm.Usage
}

// Option is a functional option for configuring a Runtime.
type Option func(*Runtime)

// WithConfig sets the runtime's resource governance configuration.
func WithConfig(c Config) Option {
	return func(r *Runtime) {
		r.config = c
	}
}

// WithTracer sets the tracer for observability.
func WithTracer(t trace.Tracer) Option {
	return func(r *Runtime) {
		r.tracer = t
	}
}

// Runtime enables dynamic spawning of child orchestrations with
// resource governance. It tracks concurrency, nesting depth, cost
// budgets, and aggregate metrics across all spawned children.
type Runtime struct {
	name   string
	tracer trace.Tracer
	config Config
	runID  string

	mu         sync.Mutex
	children   []ChildResult
	totalCost  float64
	totalUsage llm.Usage

	sem chan struct{} // concurrency semaphore, nil if unlimited
	wg  sync.WaitGroup
}

// New creates a Runtime with the given name and options.
func New(name string, opts ...Option) *Runtime {
	r := &Runtime{
		name:   name,
		tracer: trace.Noop{},
		runID:  id.New(),
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.config.MaxConcurrent > 0 {
		r.sem = make(chan struct{}, r.config.MaxConcurrent)
	}
	if r.config.MaxDepth <= 0 {
		r.config.MaxDepth = 10
	}
	return r
}

// Name returns the runtime's name.
func (r *Runtime) Name() string { return r.name }

// Context returns a new context with this runtime embedded.
// Child orchestrations can retrieve it with FromContext.
func (r *Runtime) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, runtimeKey{}, r)
}

// RemainingBudget returns the remaining cost budget in USD.
// Returns -1 if no budget limit is configured.
func (r *Runtime) RemainingBudget() float64 {
	if r.config.CostBudget <= 0 {
		return -1
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	remaining := r.config.CostBudget - r.totalCost
	if remaining < 0 {
		return 0
	}
	return remaining
}

// SpawnAgent runs a single agent as a child of this runtime.
func (r *Runtime) SpawnAgent(ctx context.Context, a *agent.Agent, input string) (*agent.Result, error) {
	if err := r.checkLimits(ctx); err != nil {
		return nil, err
	}
	if err := r.acquireSlot(ctx); err != nil {
		return nil, err
	}
	defer r.releaseSlot()

	ctx, span := r.tracer.StartSpan(ctx, "dynamic.spawn_agent")
	span.SetAttribute("dynamic.child.name", a.Name())
	span.SetAttribute("dynamic.child.type", "agent")
	span.SetAttribute("dynamic.depth", strconv.Itoa(DepthFromContext(ctx)))
	defer r.tracer.EndSpan(span)

	childCtx := r.childContext(ctx)
	result, err := a.Run(childCtx, input)
	if err != nil {
		span.SetError(err)
		r.recordChild(a.Name(), "agent", "", 0, llm.Usage{}, err)
		return nil, fmt.Errorf("dynamic spawn agent %q: %w", a.Name(), err)
	}

	r.recordChild(a.Name(), "agent", result.Message.Content, result.Cost, result.Usage, nil)
	span.SetAttribute("dynamic.child.cost_usd", fmt.Sprintf("%.6f", result.Cost))

	return result, nil
}

// SpawnTeam runs a team as a child of this runtime.
func (r *Runtime) SpawnTeam(ctx context.Context, t *team.Team, input string) (*team.Result, error) {
	if err := r.checkLimits(ctx); err != nil {
		return nil, err
	}
	if err := r.acquireSlot(ctx); err != nil {
		return nil, err
	}
	defer r.releaseSlot()

	ctx, span := r.tracer.StartSpan(ctx, "dynamic.spawn_team")
	span.SetAttribute("dynamic.child.name", t.Name())
	span.SetAttribute("dynamic.child.type", "team")
	span.SetAttribute("dynamic.depth", strconv.Itoa(DepthFromContext(ctx)))
	defer r.tracer.EndSpan(span)

	childCtx := r.childContext(ctx)
	result, err := t.Run(childCtx, input)
	if err != nil {
		span.SetError(err)
		r.recordChild(t.Name(), "team", "", 0, llm.Usage{}, err)
		return nil, fmt.Errorf("dynamic spawn team %q: %w", t.Name(), err)
	}

	r.recordChild(t.Name(), "team", result.Decision.Content, result.TotalCost, result.TotalUsage, nil)
	span.SetAttribute("dynamic.child.cost_usd", fmt.Sprintf("%.6f", result.TotalCost))

	return result, nil
}

// SpawnPipeline runs a pipeline as a child of this runtime.
func (r *Runtime) SpawnPipeline(ctx context.Context, p *pipeline.Pipeline, input string) (*pipeline.Result, error) {
	if err := r.checkLimits(ctx); err != nil {
		return nil, err
	}
	if err := r.acquireSlot(ctx); err != nil {
		return nil, err
	}
	defer r.releaseSlot()

	ctx, span := r.tracer.StartSpan(ctx, "dynamic.spawn_pipeline")
	span.SetAttribute("dynamic.child.name", p.Name())
	span.SetAttribute("dynamic.child.type", "pipeline")
	span.SetAttribute("dynamic.depth", strconv.Itoa(DepthFromContext(ctx)))
	defer r.tracer.EndSpan(span)

	childCtx := r.childContext(ctx)
	result, err := p.Run(childCtx, input)
	if err != nil {
		span.SetError(err)
		r.recordChild(p.Name(), "pipeline", "", 0, llm.Usage{}, err)
		return nil, fmt.Errorf("dynamic spawn pipeline %q: %w", p.Name(), err)
	}

	r.recordChild(p.Name(), "pipeline", result.Output, result.TotalCost, result.TotalUsage, nil)
	span.SetAttribute("dynamic.child.cost_usd", fmt.Sprintf("%.6f", result.TotalCost))

	return result, nil
}

// SpawnGraph runs a graph as a child of this runtime.
func (r *Runtime) SpawnGraph(ctx context.Context, g *graph.Graph, input string) (*graph.Result, error) {
	if err := r.checkLimits(ctx); err != nil {
		return nil, err
	}
	if err := r.acquireSlot(ctx); err != nil {
		return nil, err
	}
	defer r.releaseSlot()

	ctx, span := r.tracer.StartSpan(ctx, "dynamic.spawn_graph")
	span.SetAttribute("dynamic.child.name", g.Name())
	span.SetAttribute("dynamic.child.type", "graph")
	span.SetAttribute("dynamic.depth", strconv.Itoa(DepthFromContext(ctx)))
	defer r.tracer.EndSpan(span)

	childCtx := r.childContext(ctx)
	result, err := g.Run(childCtx, input)
	if err != nil {
		span.SetError(err)
		r.recordChild(g.Name(), "graph", "", 0, llm.Usage{}, err)
		return nil, fmt.Errorf("dynamic spawn graph %q: %w", g.Name(), err)
	}

	r.recordChild(g.Name(), "graph", result.Output, result.TotalCost, result.TotalUsage, nil)
	span.SetAttribute("dynamic.child.cost_usd", fmt.Sprintf("%.6f", result.TotalCost))

	return result, nil
}

// Future represents an asynchronous child operation launched via Go.
type Future struct {
	ch     chan struct{}
	once   sync.Once
	output string
	err    error
}

func newFuture() *Future {
	return &Future{ch: make(chan struct{})}
}

func (f *Future) complete(output string, err error) {
	f.once.Do(func() {
		f.output = output
		f.err = err
		close(f.ch)
	})
}

// Wait blocks until the future completes or the context is canceled.
// Returns the output string and any error from the spawned function.
func (f *Future) Wait(ctx context.Context) (string, error) {
	select {
	case <-f.ch:
		return f.output, f.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// Done returns a channel that is closed when the future completes.
func (f *Future) Done() <-chan struct{} {
	return f.ch
}

// Go launches fn as a background child of this runtime.
// The function runs in a new goroutine. Use the returned Future to
// await the result, or call Wait to block until all background
// children complete.
func (r *Runtime) Go(ctx context.Context, name string, fn func(ctx context.Context) (string, error)) *Future {
	f := newFuture()
	_, span := r.tracer.StartSpan(ctx, "dynamic.go")
	span.SetAttribute("dynamic.child.name", name)

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer r.tracer.EndSpan(span)

		output, err := fn(ctx)
		if err != nil {
			span.SetError(err)
		}
		f.complete(output, err)
	}()

	return f
}

// Wait blocks until all background children launched via Go complete.
func (r *Runtime) Wait() {
	r.wg.Wait()
}

// Result returns aggregate metrics from all spawned children.
func (r *Runtime) Result() *Result {
	r.mu.Lock()
	defer r.mu.Unlock()
	children := make([]ChildResult, len(r.children))
	copy(children, r.children)
	return &Result{
		RunID:      r.runID,
		Children:   children,
		TotalCost:  r.totalCost,
		TotalUsage: r.totalUsage,
	}
}

// checkLimits verifies that depth and budget constraints are met.
func (r *Runtime) checkLimits(ctx context.Context) error {
	depth := DepthFromContext(ctx)
	if depth >= r.config.MaxDepth {
		return ErrMaxDepth
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.config.CostBudget > 0 && r.totalCost >= r.config.CostBudget {
		return ErrCostBudget
	}
	return nil
}

// acquireSlot acquires a concurrency slot, blocking until one is
// available or the context is canceled.
func (r *Runtime) acquireSlot(ctx context.Context) error {
	if r.sem == nil {
		return nil
	}
	select {
	case r.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("dynamic: waiting for concurrency slot: %w", ctx.Err())
	}
}

func (r *Runtime) releaseSlot() {
	if r.sem != nil {
		<-r.sem
	}
}

// childContext creates a child context with incremented depth.
func (r *Runtime) childContext(ctx context.Context) context.Context {
	depth := DepthFromContext(ctx)
	return context.WithValue(ctx, depthKey{}, depth+1)
}

// recordChild records a child execution result and updates aggregates.
func (r *Runtime) recordChild(name, typ, output string, cost float64, usage llm.Usage, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.children = append(r.children, ChildResult{
		Name:   name,
		Type:   typ,
		Output: output,
		Cost:   cost,
		Usage:  usage,
		Error:  err,
	})
	r.totalCost += cost
	r.totalUsage.PromptTokens += usage.PromptTokens
	r.totalUsage.CompletionTokens += usage.CompletionTokens
	r.totalUsage.TotalTokens += usage.TotalTokens
}
