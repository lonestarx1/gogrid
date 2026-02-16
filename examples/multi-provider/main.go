// Example multi-provider demonstrates swapping LLM providers without
// changing agent logic.
//
// The same agent runs with two different mock providers to show that
// provider selection is a configuration concern, not a code change.
//
// Usage:
//
//	go run ./examples/multi-provider
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/llm/mock"
)

func main() {
	ctx := context.Background()

	// Define two providers with different responses (simulating different LLMs).
	providerA := mock.New(mock.WithFallback(&llm.Response{
		Message: llm.NewAssistantMessage("Provider A: Go's goroutines enable lightweight concurrency with channels for communication."),
		Usage:   llm.Usage{PromptTokens: 20, CompletionTokens: 15, TotalTokens: 35},
		Model:   "model-a",
	}))

	providerB := mock.New(mock.WithFallback(&llm.Response{
		Message: llm.NewAssistantMessage("Provider B: Go achieves concurrency through goroutines — lightweight threads managed by the runtime."),
		Usage:   llm.Usage{PromptTokens: 20, CompletionTokens: 18, TotalTokens: 38},
		Model:   "model-b",
	}))

	// Same agent configuration, different providers.
	instructions := "You are a Go expert. Explain concepts clearly and concisely."
	input := "Explain Go's concurrency model in one sentence."

	for i, p := range []llm.Provider{providerA, providerB} {
		a := agent.New(fmt.Sprintf("go-expert-%d", i+1),
			agent.WithProvider(p),
			agent.WithModel("mock-model"),
			agent.WithInstructions(instructions),
			agent.WithConfig(agent.Config{MaxTurns: 3, MaxTokens: 256}),
		)

		result, err := a.Run(ctx, input)
		if err != nil {
			log.Fatalf("Run with provider %d: %v", i+1, err)
		}

		fmt.Printf("=== Provider %d ===\n", i+1)
		fmt.Printf("Response: %s\n", result.Message.Content)
		fmt.Printf("Model: %s | Tokens: %d\n\n", result.Message.Content[:10], result.Usage.TotalTokens)
	}

	fmt.Println("Both runs used the same agent logic — only the provider changed.")
}
