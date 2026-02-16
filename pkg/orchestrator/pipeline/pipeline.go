package pipeline

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/lonestarx1/gogrid/internal/id"
	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/memory"
	"github.com/lonestarx1/gogrid/pkg/memory/transfer"
	"github.com/lonestarx1/gogrid/pkg/trace"
)

// ErrorAction specifies how a pipeline handles a stage failure.
type ErrorAction int

const (
	// Abort stops the pipeline and returns the error.
	Abort ErrorAction = iota
	// Skip marks the stage as skipped and continues with the previous output.
	Skip
)

// RetryPolicy controls retry behavior for a stage.
type RetryPolicy struct {
	// MaxAttempts is the maximum number of tries (including the initial).
	// 0 or 1 means no retry.
	MaxAttempts int
	// Delay is the wait time between retries.
	Delay time.Duration
}

// Stage is a single step in a pipeline.
type Stage struct {
	// Name identifies this stage in traces and progress reports.
	Name string
	// Agent is the agent that processes this stage.
	Agent *agent.Agent
	// InputTransform optionally transforms the previous stage's output
	// before passing it to this stage's agent. If nil, the previous
	// output is used as-is.
	InputTransform func(input string) string
	// OutputValidate optionally validates this stage's output.
	// Return a non-nil error to fail the stage.
	OutputValidate func(output string) error
	// OnError specifies the error handling action. Defaults to Abort.
	OnError ErrorAction
	// Retry controls retry behavior. Zero value means no retry.
	Retry RetryPolicy
	// Timeout is the maximum duration for this stage. Zero means no
	// per-stage timeout (inherits pipeline/context timeout).
	Timeout time.Duration
	// CostBudget is the maximum cost in USD for this stage.
	// Zero means no per-stage budget (inherits pipeline budget).
	CostBudget float64
}

// Config controls pipeline execution behavior.
type Config struct {
	// Timeout is the maximum wall-clock duration for the entire pipeline.
	// Zero means no timeout (relies on the caller's context).
	Timeout time.Duration
	// CostBudget is the maximum total cost in USD across all stages.
	// Zero means no budget limit.
	CostBudget float64
}

// StageResult holds the output of a single stage.
type StageResult struct {
	// Name is the stage name.
	Name string
	// AgentResult is the agent's execution result.
	AgentResult *agent.Result
	// Skipped is true if the stage was skipped due to OnError=Skip.
	Skipped bool
	// Attempts is the number of times the stage was executed.
	Attempts int
}

// Result is returned by Pipeline.Run with the outcome of the execution.
type Result struct {
	// RunID uniquely identifies this execution.
	RunID string
	// Output is the final stage's output content.
	Output string
	// Stages holds per-stage results in execution order.
	Stages []StageResult
	// TotalCost is the aggregate cost in USD across all stages.
	TotalCost float64
	// TotalUsage is the aggregate token usage.
	TotalUsage llm.Usage
	// TransferLog is the state ownership audit trail.
	TransferLog []transfer.AuditEntry
}

// ProgressFunc is called after each stage completes, reporting the
// stage index (0-based), total stages, and the stage result.
type ProgressFunc func(stageIndex, totalStages int, result StageResult)

// Pipeline orchestrates a sequential chain of agent stages with
// state ownership transfer between stages.
type Pipeline struct {
	name     string
	stages   []Stage
	tracer   trace.Tracer
	config   Config
	progress ProgressFunc
}

// Option is a functional option for configuring a Pipeline.
type Option func(*Pipeline)

// New creates a Pipeline with the given name and options.
func New(name string, opts ...Option) *Pipeline {
	p := &Pipeline{
		name:   name,
		tracer: trace.Noop{},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// WithStages sets the pipeline stages in execution order.
func WithStages(stages ...Stage) Option {
	return func(p *Pipeline) {
		p.stages = stages
	}
}

// WithTracer sets the tracer for observability.
func WithTracer(t trace.Tracer) Option {
	return func(p *Pipeline) {
		p.tracer = t
	}
}

// WithConfig sets the pipeline's execution configuration.
func WithConfig(c Config) Option {
	return func(p *Pipeline) {
		p.config = c
	}
}

// WithProgress sets a callback for stage completion progress reporting.
func WithProgress(fn ProgressFunc) Option {
	return func(p *Pipeline) {
		p.progress = fn
	}
}

// Name returns the pipeline's name.
func (p *Pipeline) Name() string { return p.name }

// Run executes the pipeline stages sequentially, transferring state
// ownership between each stage.
//
// The pipeline loop:
//  1. Create transferable state, acquire for the first stage.
//  2. Run the stage's agent with the current input.
//  3. Apply output validation if configured.
//  4. Transfer state ownership to the next stage.
//  5. Apply input transform if configured.
//  6. Repeat until all stages complete or a failure occurs.
func (p *Pipeline) Run(ctx context.Context, input string) (*Result, error) {
	if len(p.stages) == 0 {
		return nil, errors.New("pipeline: at least one stage is required")
	}

	// Apply pipeline timeout.
	if p.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.config.Timeout)
		defer cancel()
	}

	runID := id.New()

	// Start pipeline span.
	ctx, runSpan := p.tracer.StartSpan(ctx, "pipeline.run")
	runSpan.SetAttribute("pipeline.name", p.name)
	runSpan.SetAttribute("pipeline.run_id", runID)
	runSpan.SetAttribute("pipeline.stages", strconv.Itoa(len(p.stages)))
	defer p.tracer.EndSpan(runSpan)

	// Create transferable state.
	state := transfer.NewState(memory.NewInMemory())

	var totalCost float64
	var totalUsage llm.Usage
	stageResults := make([]StageResult, 0, len(p.stages))
	currentInput := input
	var prevOwner string

	for i, stage := range p.stages {
		if err := ctx.Err(); err != nil {
			runSpan.SetError(err)
			return nil, fmt.Errorf("pipeline: %w", err)
		}

		// Check pipeline cost budget.
		if p.config.CostBudget > 0 && totalCost >= p.config.CostBudget {
			runSpan.SetAttribute("pipeline.stopped_reason", "cost_budget")
			break
		}

		sr, err := p.runStage(ctx, state, i, stage, currentInput, &prevOwner)
		if err != nil {
			runSpan.SetError(err)
			return nil, err
		}

		stageResults = append(stageResults, sr)

		if sr.AgentResult != nil {
			totalCost += sr.AgentResult.Cost
			totalUsage.PromptTokens += sr.AgentResult.Usage.PromptTokens
			totalUsage.CompletionTokens += sr.AgentResult.Usage.CompletionTokens
			totalUsage.TotalTokens += sr.AgentResult.Usage.TotalTokens
		}

		// Progress callback.
		if p.progress != nil {
			p.progress(i, len(p.stages), sr)
		}

		// Advance input for next stage.
		if !sr.Skipped && sr.AgentResult != nil {
			currentInput = sr.AgentResult.Message.Content
		}
	}

	// Determine final output.
	output := currentInput
	for i := len(stageResults) - 1; i >= 0; i-- {
		if !stageResults[i].Skipped && stageResults[i].AgentResult != nil {
			output = stageResults[i].AgentResult.Message.Content
			break
		}
	}

	runSpan.SetAttribute("pipeline.cost_usd", fmt.Sprintf("%.6f", totalCost))
	runSpan.SetAttribute("pipeline.stages_completed", strconv.Itoa(len(stageResults)))

	return &Result{
		RunID:       runID,
		Output:      output,
		Stages:      stageResults,
		TotalCost:   totalCost,
		TotalUsage:  totalUsage,
		TransferLog: state.AuditLog(),
	}, nil
}

// runStage executes a single stage with retry logic, state transfer,
// input transformation, and output validation.
func (p *Pipeline) runStage(
	ctx context.Context,
	state *transfer.State,
	index int,
	stage Stage,
	input string,
	prevOwner *string,
) (StageResult, error) {
	stageName := stage.Name
	if stageName == "" {
		stageName = stage.Agent.Name()
	}

	// Start stage span.
	stageCtx, stageSpan := p.tracer.StartSpan(ctx, "pipeline.stage")
	stageSpan.SetAttribute("pipeline.stage.name", stageName)
	stageSpan.SetAttribute("pipeline.stage.index", strconv.Itoa(index))
	defer p.tracer.EndSpan(stageSpan)

	// Apply per-stage timeout.
	if stage.Timeout > 0 {
		var cancel context.CancelFunc
		stageCtx, cancel = context.WithTimeout(stageCtx, stage.Timeout)
		defer cancel()
	}

	// Acquire or transfer state ownership.
	var handle *transfer.Handle
	var err error
	if *prevOwner == "" {
		handle, err = state.Acquire(stageName)
	} else {
		handle, err = state.Transfer(*prevOwner, stageName)
	}
	if err != nil {
		stageSpan.SetError(err)
		return StageResult{}, fmt.Errorf("pipeline stage %q: state transfer: %w", stageName, err)
	}

	// Apply input transform.
	stageInput := input
	if stage.InputTransform != nil {
		stageInput = stage.InputTransform(input)
	}

	// Save input to state.
	_ = handle.Save(stageCtx, "input", []llm.Message{llm.NewUserMessage(stageInput)})

	// Execute with retries.
	maxAttempts := stage.Retry.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	var lastErr error
	var agentResult *agent.Result
	attempts := 0
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		attempts = attempt
		stageSpan.SetAttribute("pipeline.stage.attempt", strconv.Itoa(attempt))

		agentResult, lastErr = stage.Agent.Run(stageCtx, stageInput)
		if lastErr == nil {
			// Apply output validation.
			if stage.OutputValidate != nil {
				lastErr = stage.OutputValidate(agentResult.Message.Content)
			}
		}

		if lastErr == nil {
			// Check per-stage cost budget.
			if stage.CostBudget > 0 && agentResult.Cost > stage.CostBudget {
				lastErr = fmt.Errorf("stage %q exceeded cost budget: $%.6f > $%.6f",
					stageName, agentResult.Cost, stage.CostBudget)
			}
		}

		if lastErr == nil {
			break
		}

		// Wait before retry (except on last attempt).
		if attempt < maxAttempts && stage.Retry.Delay > 0 {
			select {
			case <-time.After(stage.Retry.Delay):
			case <-stageCtx.Done():
				stageSpan.SetError(stageCtx.Err())
				return StageResult{}, fmt.Errorf("pipeline stage %q: %w", stageName, stageCtx.Err())
			}
		}
	}

	// Handle failure.
	if lastErr != nil {
		stageSpan.SetError(lastErr)
		stageSpan.SetAttribute("pipeline.stage.attempts", strconv.Itoa(attempts))

		switch stage.OnError {
		case Skip:
			stageSpan.SetAttribute("pipeline.stage.skipped", "true")
			*prevOwner = stageName
			return StageResult{
				Name:     stageName,
				Skipped:  true,
				Attempts: attempts,
			}, nil
		default: // Abort
			return StageResult{}, fmt.Errorf("pipeline stage %q: %w", stageName, lastErr)
		}
	}

	// Save output to state.
	_ = handle.Save(stageCtx, "output", []llm.Message{agentResult.Message})

	stageSpan.SetAttribute("pipeline.stage.cost_usd", fmt.Sprintf("%.6f", agentResult.Cost))
	*prevOwner = stageName

	return StageResult{
		Name:        stageName,
		AgentResult: agentResult,
		Attempts:    attempts,
	}, nil
}
