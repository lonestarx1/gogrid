// Example eval demonstrates GoGrid's evaluation framework.
//
// It runs an agent, then evaluates the result using multiple evaluators
// composed into a Suite: exact match, substring containment, cost budget,
// and a custom evaluator.
//
// Usage:
//
//	go run ./examples/eval
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/eval"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/llm/mock"
)

func main() {
	ctx := context.Background()

	// Create a mock agent that returns a known response.
	provider := mock.New(mock.WithFallback(&llm.Response{
		Message: llm.NewAssistantMessage("Go is a statically typed, compiled language designed for simplicity and efficiency."),
		Usage:   llm.Usage{PromptTokens: 15, CompletionTokens: 14, TotalTokens: 29},
		Model:   "mock",
	}))

	a := agent.New("eval-agent",
		agent.WithProvider(provider),
		agent.WithModel("mock"),
		agent.WithInstructions("You are a concise technical writer."),
	)

	// Run the agent.
	result, err := a.Run(ctx, "Describe Go in one sentence.")
	if err != nil {
		log.Fatalf("Run: %v", err)
	}
	fmt.Printf("Agent output: %s\n\n", result.Message.Content)

	// Build an evaluation suite with multiple evaluators.
	suite := eval.NewSuite(
		eval.NewContains("Go", "compiled", "simplicity"),
		eval.NewCostWithin(0.01),
		eval.NewFunc("min_length", func(_ context.Context, r *agent.Result) (eval.Score, error) {
			content := r.Message.Content
			if len(content) >= 20 {
				return eval.Score{Pass: true, Value: 1.0, Reason: fmt.Sprintf("length %d >= 20", len(content))}, nil
			}
			return eval.Score{Pass: false, Value: float64(len(content)) / 20.0, Reason: fmt.Sprintf("length %d < 20", len(content))}, nil
		}),
	)

	// Run the suite.
	sr, err := suite.Run(ctx, result)
	if err != nil {
		log.Fatalf("Suite.Run: %v", err)
	}

	// Display results.
	fmt.Printf("Suite passed: %v\n\n", sr.Pass)
	for name, score := range sr.Scores {
		status := "PASS"
		if !score.Pass {
			status = "FAIL"
		}
		fmt.Printf("  [%s] %s (%.2f): %s\n", status, name, score.Value, score.Reason)
	}
	for name, evalErr := range sr.Errors {
		fmt.Printf("  [ERROR] %s: %v\n", name, evalErr)
	}
}
