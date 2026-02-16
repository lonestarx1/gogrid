// Example pipeline-research demonstrates a three-stage pipeline with
// state ownership transfer, input transforms, and progress reporting.
//
// The pipeline flows: research -> analyze -> summarize. Each stage
// receives the previous stage's output, and state ownership transfers
// cleanly between stages.
//
// Usage:
//
//	go run ./examples/pipeline-research
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/llm/mock"
	"github.com/lonestarx1/gogrid/pkg/orchestrator/pipeline"
)

func main() {
	ctx := context.Background()

	// Stage 1: Research — gathers raw information.
	researcher := agent.New("researcher",
		agent.WithProvider(mock.New(mock.WithFallback(&llm.Response{
			Message: llm.NewAssistantMessage("Research findings: Go was created at Google in 2007 by Griesemer, Pike, and Thompson. It features goroutines for concurrency, a garbage collector, and compiles to native code. Go 1.0 was released in 2012. The language emphasizes simplicity and readability."),
			Usage:   llm.Usage{PromptTokens: 20, CompletionTokens: 50, TotalTokens: 70},
			Model:   "mock",
		}))),
		agent.WithModel("mock"),
		agent.WithInstructions("You are a thorough researcher. Gather key facts and findings."),
	)

	// Stage 2: Analyze — structures and evaluates the research.
	analyzer := agent.New("analyzer",
		agent.WithProvider(mock.New(mock.WithFallback(&llm.Response{
			Message: llm.NewAssistantMessage("Analysis: Go's design priorities are (1) compilation speed, (2) runtime efficiency, (3) developer productivity. The goroutine model solves C10K-class concurrency problems. Trade-off: generics were added late (Go 1.18), limiting early adoption for some use cases."),
			Usage:   llm.Usage{PromptTokens: 60, CompletionTokens: 45, TotalTokens: 105},
			Model:   "mock",
		}))),
		agent.WithModel("mock"),
		agent.WithInstructions("You are an analyst. Structure findings into key themes with trade-offs."),
	)

	// Stage 3: Summarize — produces a concise summary.
	summarizer := agent.New("summarizer",
		agent.WithProvider(mock.New(mock.WithFallback(&llm.Response{
			Message: llm.NewAssistantMessage("Summary: Go is a compiled language built for production infrastructure. Its goroutine-based concurrency model, fast compilation, and simple syntax make it ideal for networked services. Late generics adoption was a trade-off for initial simplicity."),
			Usage:   llm.Usage{PromptTokens: 80, CompletionTokens: 40, TotalTokens: 120},
			Model:   "mock",
		}))),
		agent.WithModel("mock"),
		agent.WithInstructions("You are a summarizer. Condense analysis into 2-3 sentences."),
	)

	// Build the pipeline.
	p := pipeline.New("research-pipeline",
		pipeline.WithStages(
			pipeline.Stage{
				Name:  "research",
				Agent: researcher,
			},
			pipeline.Stage{
				Name:  "analyze",
				Agent: analyzer,
				InputTransform: func(input string) string {
					return "Analyze the following research:\n\n" + input
				},
			},
			pipeline.Stage{
				Name:  "summarize",
				Agent: summarizer,
				InputTransform: func(input string) string {
					return "Summarize this analysis in 2-3 sentences:\n\n" + input
				},
				OutputValidate: func(output string) error {
					if len(strings.Fields(output)) < 5 {
						return fmt.Errorf("summary too short: %d words", len(strings.Fields(output)))
					}
					return nil
				},
			},
		),
		pipeline.WithProgress(func(i, total int, sr pipeline.StageResult) {
			fmt.Printf("[%d/%d] Stage %q complete (attempts: %d)\n",
				i+1, total, sr.Name, sr.Attempts)
		}),
		pipeline.WithConfig(pipeline.Config{CostBudget: 1.00}),
	)

	// Run the pipeline.
	result, err := p.Run(ctx, "Research the Go programming language.")
	if err != nil {
		log.Fatalf("Run: %v", err)
	}

	// Display results.
	fmt.Printf("\n=== Pipeline Results ===\n\n")
	fmt.Printf("Final output: %s\n\n", result.Output)
	fmt.Printf("Total cost: $%.6f\n", result.TotalCost)
	fmt.Printf("Total tokens: %d\n", result.TotalUsage.TotalTokens)
	fmt.Printf("Stages: %d\n\n", len(result.Stages))

	fmt.Println("State transfer audit log:")
	for _, entry := range result.TransferLog {
		fmt.Printf("  %s -> %s (gen %d)\n", entry.From, entry.To, entry.Generation)
	}
}
