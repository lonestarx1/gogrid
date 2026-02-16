// Package eval provides an evaluation framework for scoring GoGrid
// agent outputs. It includes built-in evaluators for exact match,
// substring containment, cost budgets, tool usage, and LLM-as-judge,
// plus a Suite for running multiple evaluators against a single result.
package eval

import (
	"context"

	"github.com/lonestarx1/gogrid/pkg/agent"
)

// Evaluator scores an agent result against specific criteria.
type Evaluator interface {
	// Name returns a human-readable identifier for this evaluator.
	Name() string
	// Evaluate scores the result. Returns a Score and any error
	// encountered during evaluation (not a low score — that's Score.Pass).
	Evaluate(ctx context.Context, result *agent.Result) (Score, error)
}

// Score is the outcome of a single evaluation.
type Score struct {
	// Pass is the primary signal: did the result meet the criteria?
	Pass bool
	// Value is a normalized score from 0.0 to 1.0 for granularity.
	Value float64
	// Reason is a human-readable explanation of the outcome.
	Reason string
}

// Func wraps a function as an Evaluator. Use this for one-off or
// inline evaluations, especially when integrating standalone scoring
// functions like CompletedWithin into a Suite.
type Func struct {
	name string
	fn   func(context.Context, *agent.Result) (Score, error)
}

// NewFunc creates a Func evaluator with the given name and scoring function.
func NewFunc(name string, fn func(context.Context, *agent.Result) (Score, error)) *Func {
	return &Func{name: name, fn: fn}
}

// Name returns the evaluator's name.
func (f *Func) Name() string { return f.name }

// Evaluate calls the wrapped function.
func (f *Func) Evaluate(ctx context.Context, result *agent.Result) (Score, error) {
	return f.fn(ctx, result)
}

// Suite runs multiple evaluators against a single agent result.
type Suite struct {
	evaluators []Evaluator
}

// NewSuite creates a Suite from the given evaluators.
func NewSuite(evaluators ...Evaluator) *Suite {
	return &Suite{evaluators: evaluators}
}

// SuiteResult is the aggregate outcome of running a Suite.
type SuiteResult struct {
	// Scores maps evaluator name to its score.
	Scores map[string]Score
	// Pass is true only if every evaluator passed and no errors occurred.
	Pass bool
	// Errors maps evaluator name to any error encountered. Only evaluators
	// that returned errors are included.
	Errors map[string]error
}

// Run executes every evaluator in the suite against the result.
func (s *Suite) Run(ctx context.Context, result *agent.Result) (*SuiteResult, error) {
	sr := &SuiteResult{
		Scores: make(map[string]Score, len(s.evaluators)),
		Errors: make(map[string]error),
		Pass:   true,
	}

	for _, ev := range s.evaluators {
		score, err := ev.Evaluate(ctx, result)
		if err != nil {
			sr.Errors[ev.Name()] = err
			sr.Pass = false
			continue
		}
		sr.Scores[ev.Name()] = score
		if !score.Pass {
			sr.Pass = false
		}
	}

	return sr, nil
}
