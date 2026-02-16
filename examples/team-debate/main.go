// Example team-debate demonstrates a multi-agent team reaching consensus.
//
// Three agents with different roles (advocate, skeptic, moderator) discuss
// a topic over multiple rounds. The team uses the Unanimous consensus
// strategy and a coordinator to synthesize the final decision.
//
// Usage:
//
//	go run ./examples/team-debate
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/llm/mock"
	"github.com/lonestarx1/gogrid/pkg/orchestrator/team"
)

func main() {
	ctx := context.Background()

	// Create agents with different perspectives.
	advocate := agent.New("advocate",
		agent.WithProvider(mock.New(mock.WithFallback(&llm.Response{
			Message: llm.NewAssistantMessage("Go's type system catches errors at compile time, reducing production bugs. Static typing is essential for large-scale systems."),
			Usage:   llm.Usage{PromptTokens: 30, CompletionTokens: 25, TotalTokens: 55},
			Model:   "mock",
		}))),
		agent.WithModel("mock"),
		agent.WithInstructions("You advocate for the topic. Present strong arguments in favor."),
	)

	skeptic := agent.New("skeptic",
		agent.WithProvider(mock.New(mock.WithFallback(&llm.Response{
			Message: llm.NewAssistantMessage("Dynamic languages like Python offer faster iteration. Type inference in modern languages reduces the verbosity argument. Static typing adds ceremony without proportional value for small projects."),
			Usage:   llm.Usage{PromptTokens: 30, CompletionTokens: 30, TotalTokens: 60},
			Model:   "mock",
		}))),
		agent.WithModel("mock"),
		agent.WithInstructions("You are a skeptic. Challenge assumptions and present counterarguments."),
	)

	moderator := agent.New("moderator",
		agent.WithProvider(mock.New(mock.WithFallback(&llm.Response{
			Message: llm.NewAssistantMessage("Both perspectives have merit. Static typing shines in large codebases and team settings. Dynamic typing excels in prototyping and exploratory work. The choice depends on project scale and team composition."),
			Usage:   llm.Usage{PromptTokens: 30, CompletionTokens: 35, TotalTokens: 65},
			Model:   "mock",
		}))),
		agent.WithModel("mock"),
		agent.WithInstructions("You are a neutral moderator. Find common ground and synthesize balanced viewpoints."),
	)

	// Coordinator synthesizes the final team decision.
	coordinator := agent.New("coordinator",
		agent.WithProvider(mock.New(mock.WithFallback(&llm.Response{
			Message: llm.NewAssistantMessage("Team consensus: Static typing provides significant value for production systems at scale, while dynamic typing remains valuable for rapid prototyping. For GoGrid's target audience (infrastructure engineers), static typing is the right default."),
			Usage:   llm.Usage{PromptTokens: 100, CompletionTokens: 40, TotalTokens: 140},
			Model:   "mock",
		}))),
		agent.WithModel("mock"),
		agent.WithInstructions("Synthesize the team discussion into a clear, balanced final decision."),
	)

	// Build the team.
	t := team.New("debate-team",
		team.WithMembers(
			team.Member{Agent: advocate, Role: "advocate"},
			team.Member{Agent: skeptic, Role: "skeptic"},
			team.Member{Agent: moderator, Role: "moderator"},
		),
		team.WithCoordinator(coordinator),
		team.WithConfig(team.Config{
			MaxRounds:  2,
			CostBudget: 1.00,
		}),
	)

	// Run the debate.
	result, err := t.Run(ctx, "Is static typing essential for production software?")
	if err != nil {
		log.Fatalf("Run: %v", err)
	}

	// Display results.
	fmt.Printf("=== Team Debate Results ===\n\n")
	fmt.Printf("Decision: %s\n\n", result.Decision.Content)
	fmt.Printf("Rounds: %d\n", result.Rounds)
	fmt.Printf("Total cost: $%.6f\n", result.TotalCost)
	fmt.Printf("Total tokens: %d\n\n", result.TotalUsage.TotalTokens)

	fmt.Println("Member responses:")
	for name, r := range result.Responses {
		fmt.Printf("  %s: %s\n\n", name, r.Message.Content)
	}
}
