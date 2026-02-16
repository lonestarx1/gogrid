// Example dynamic-spawn demonstrates dynamic orchestration where a
// runtime spawns child agents and teams at runtime with resource governance.
//
// The runtime enforces concurrency limits, nesting depth, and cost budgets
// across all spawned children. Async futures allow parallel child execution.
//
// Usage:
//
//	go run ./examples/dynamic-spawn
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/llm/mock"
	"github.com/lonestarx1/gogrid/pkg/orchestrator/dynamic"
	"github.com/lonestarx1/gogrid/pkg/orchestrator/team"
)

func main() {
	ctx := context.Background()

	// Create a runtime with resource governance.
	rt := dynamic.New("coordinator",
		dynamic.WithConfig(dynamic.Config{
			MaxConcurrent: 3,
			MaxDepth:      2,
			CostBudget:    5.00,
		}),
	)

	// Embed runtime in context for child access.
	ctx = rt.Context(ctx)

	// Spawn a research agent.
	researcher := agent.New("researcher",
		agent.WithProvider(mock.New(mock.WithFallback(&llm.Response{
			Message: llm.NewAssistantMessage("Research: Go's concurrency model uses goroutines (lightweight threads) and channels (typed communication pipes). The runtime multiplexes goroutines onto OS threads."),
			Usage:   llm.Usage{PromptTokens: 20, CompletionTokens: 30, TotalTokens: 50},
			Model:   "mock",
		}))),
		agent.WithModel("mock"),
		agent.WithInstructions("Research the given topic thoroughly."),
	)

	fmt.Println("Spawning research agent...")
	researchResult, err := rt.SpawnAgent(ctx, researcher, "How does Go handle concurrency?")
	if err != nil {
		log.Fatalf("SpawnAgent: %v", err)
	}
	fmt.Printf("Research: %s\n\n", researchResult.Message.Content)

	// Spawn a review team using async futures.
	reviewer1 := agent.New("reviewer-1",
		agent.WithProvider(mock.New(mock.WithFallback(&llm.Response{
			Message: llm.NewAssistantMessage("The research accurately covers goroutines and channels. Could mention the sync package for mutex-based coordination."),
			Usage:   llm.Usage{PromptTokens: 40, CompletionTokens: 20, TotalTokens: 60},
			Model:   "mock",
		}))),
		agent.WithModel("mock"),
	)
	reviewer2 := agent.New("reviewer-2",
		agent.WithProvider(mock.New(mock.WithFallback(&llm.Response{
			Message: llm.NewAssistantMessage("Good coverage of core concepts. The context package for cancellation propagation deserves mention."),
			Usage:   llm.Usage{PromptTokens: 40, CompletionTokens: 18, TotalTokens: 58},
			Model:   "mock",
		}))),
		agent.WithModel("mock"),
	)

	reviewTeam := team.New("review-team",
		team.WithMembers(
			team.Member{Agent: reviewer1, Role: "technical reviewer"},
			team.Member{Agent: reviewer2, Role: "completeness reviewer"},
		),
		team.WithConfig(team.Config{MaxRounds: 1}),
	)

	fmt.Println("Spawning review team...")
	teamResult, err := rt.SpawnTeam(ctx, reviewTeam, researchResult.Message.Content)
	if err != nil {
		log.Fatalf("SpawnTeam: %v", err)
	}
	fmt.Printf("Team decision: %s\n\n", teamResult.Decision.Content)

	// Launch async futures for parallel work.
	fmt.Println("Launching async futures...")
	future1 := rt.Go(ctx, "format-markdown", func(ctx context.Context) (string, error) {
		return "## Go Concurrency\n\nFormatted research content...", nil
	})
	future2 := rt.Go(ctx, "generate-tags", func(ctx context.Context) (string, error) {
		return "tags: go, concurrency, goroutines, channels", nil
	})

	// Wait for futures.
	md, err := future1.Wait(ctx)
	if err != nil {
		log.Fatalf("future1: %v", err)
	}
	tags, err := future2.Wait(ctx)
	if err != nil {
		log.Fatalf("future2: %v", err)
	}
	fmt.Printf("Markdown: %s\nTags: %s\n\n", md, tags)

	// Wait for all background work.
	rt.Wait()

	// Display aggregate metrics.
	metrics := rt.Result()
	fmt.Printf("=== Runtime Metrics ===\n")
	fmt.Printf("Children spawned: %d\n", len(metrics.Children))
	fmt.Printf("Total cost: $%.6f\n", metrics.TotalCost)
	fmt.Printf("Total tokens: %d\n", metrics.TotalUsage.TotalTokens)
	fmt.Printf("Remaining budget: $%.2f\n", rt.RemainingBudget())
}
