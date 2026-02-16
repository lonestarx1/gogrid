// Example graph-review demonstrates a graph with conditional edges and loops.
//
// The graph implements a write-review-revise workflow: a writer produces
// the initial draft, a reviewer evaluates it, and if revision is needed
// a reviser improves it and sends it back to the reviewer. The loop
// continues until the reviewer approves or the iteration limit is reached.
//
// Usage:
//
//	go run ./examples/graph-review
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/llm/mock"
	"github.com/lonestarx1/gogrid/pkg/orchestrator/graph"
)

func main() {
	ctx := context.Background()

	// Writer produces the initial draft.
	writer := agent.New("writer",
		agent.WithProvider(mock.New(mock.WithFallback(&llm.Response{
			Message: llm.NewAssistantMessage("Draft: GoGrid is a framework for building AI agents in Go. It supports five orchestration patterns."),
			Usage:   llm.Usage{PromptTokens: 20, CompletionTokens: 20, TotalTokens: 40},
			Model:   "mock",
		}))),
		agent.WithModel("mock"),
		agent.WithInstructions("You are a technical writer. Write clear, concise content."),
	)

	// Reviewer evaluates content. Includes "REVISE" or "APPROVED" in response.
	reviewer := agent.New("reviewer",
		agent.WithProvider(mock.New(mock.WithResponses(
			&llm.Response{
				Message: llm.NewAssistantMessage("REVISE: Too vague. Add specifics about the patterns and emphasize production-grade quality."),
				Usage:   llm.Usage{PromptTokens: 30, CompletionTokens: 20, TotalTokens: 50},
				Model:   "mock",
			},
			&llm.Response{
				Message: llm.NewAssistantMessage("APPROVED: Clear, specific, and well-structured. Good to publish."),
				Usage:   llm.Usage{PromptTokens: 50, CompletionTokens: 12, TotalTokens: 62},
				Model:   "mock",
			},
		))),
		agent.WithModel("mock"),
		agent.WithInstructions("Review content. Respond with APPROVED or REVISE with feedback."),
	)

	// Reviser improves content based on feedback, then goes back to reviewer.
	reviser := agent.New("reviser",
		agent.WithProvider(mock.New(mock.WithFallback(&llm.Response{
			Message: llm.NewAssistantMessage("GoGrid is a production-grade AI agent framework in Go with five composable orchestration patterns: single agent, team, pipeline, graph, and dynamic orchestration."),
			Usage:   llm.Usage{PromptTokens: 40, CompletionTokens: 30, TotalTokens: 70},
			Model:   "mock",
		}))),
		agent.WithModel("mock"),
		agent.WithInstructions("Revise the content based on reviewer feedback."),
	)

	// Build the graph: writer -> reviewer -> (reviser -> reviewer loop).
	// writer is the start node (no incoming edges).
	// reviewer is a terminal node when it approves (no unconditional outgoing).
	g, err := graph.NewBuilder("review-loop").
		AddNode("writer", writer).
		AddNode("reviewer", reviewer).
		AddNode("reviser", reviser).
		AddEdge("writer", "reviewer").
		AddEdge("reviewer", "reviser", graph.When(func(output string) bool {
			return strings.Contains(output, "REVISE")
		})).
		AddEdge("reviser", "reviewer").
		Options(graph.WithConfig(graph.Config{MaxIterations: 3})).
		Build()
	if err != nil {
		log.Fatalf("Build: %v", err)
	}

	// Export the graph structure as DOT format.
	fmt.Println("=== Graph Structure (DOT) ===")
	fmt.Println(g.DOT())

	// Run the graph.
	result, err := g.Run(ctx, "Write a one-paragraph description of GoGrid.")
	if err != nil {
		log.Fatalf("Run: %v", err)
	}

	// Display results.
	fmt.Printf("=== Graph Results ===\n\n")
	fmt.Printf("Final output: %s\n\n", result.Output)
	fmt.Printf("Total cost: $%.6f\n", result.TotalCost)
	fmt.Printf("Total tokens: %d\n\n", result.TotalUsage.TotalTokens)

	fmt.Println("Node execution history:")
	for name, results := range result.NodeResults {
		for _, nr := range results {
			fmt.Printf("  %s (iteration %d): %s\n", name, nr.Iteration, truncate(nr.Output, 80))
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
