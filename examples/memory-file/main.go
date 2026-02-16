// Example memory-file demonstrates file-backed memory with pruning.
//
// It creates a file memory store, saves conversation entries, searches
// them, prunes old entries, and displays memory statistics.
//
// Usage:
//
//	go run ./examples/memory-file
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/llm/mock"
	"github.com/lonestarx1/gogrid/pkg/memory"
	"github.com/lonestarx1/gogrid/pkg/memory/file"
)

// agePolicy prunes entries older than a threshold.
type agePolicy struct {
	maxAge time.Duration
}

func (p agePolicy) ShouldPrune(e memory.Entry) bool {
	return time.Since(e.CreatedAt) > p.maxAge
}

func main() {
	ctx := context.Background()

	// Create a temporary directory for file memory.
	dir, err := os.MkdirTemp("", "gogrid-memory-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create file-backed memory.
	mem, err := file.New(dir)
	if err != nil {
		log.Fatalf("file.New: %v", err)
	}

	// Create an agent with file memory.
	provider := mock.New(mock.WithFallback(&llm.Response{
		Message: llm.NewAssistantMessage("Go was designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson, first appearing in 2009."),
		Usage:   llm.Usage{PromptTokens: 30, CompletionTokens: 20, TotalTokens: 50},
		Model:   "mock",
	}))

	a := agent.New("history-agent",
		agent.WithProvider(provider),
		agent.WithModel("mock"),
		agent.WithMemory(mem),
		agent.WithInstructions("You are a knowledgeable assistant with persistent memory."),
	)

	// Run the agent — memory is saved automatically.
	result, err := a.Run(ctx, "Tell me about Go's history.")
	if err != nil {
		log.Fatalf("Run: %v", err)
	}
	fmt.Printf("Agent: %s\n\n", result.Message.Content)

	// Show memory statistics.
	stats, err := mem.Stats(ctx)
	if err != nil {
		log.Fatalf("Stats: %v", err)
	}
	fmt.Printf("Memory stats: %d keys, %d entries, %d bytes\n",
		stats.Keys, stats.TotalEntries, stats.TotalSize)

	// Search memory.
	entries, err := mem.Search(ctx, "Go")
	if err != nil {
		log.Fatalf("Search: %v", err)
	}
	fmt.Printf("Search 'Go': found %d entries\n", len(entries))

	// Prune entries older than 1 hour (none will be pruned in this demo).
	pruned, err := mem.Prune(ctx, agePolicy{maxAge: time.Hour})
	if err != nil {
		log.Fatalf("Prune: %v", err)
	}
	fmt.Printf("Pruned: %d entries (older than 1h)\n", pruned)

	// Show final stats.
	stats, _ = mem.Stats(ctx)
	fmt.Printf("Final stats: %d keys, %d entries, %d bytes\n",
		stats.Keys, stats.TotalEntries, stats.TotalSize)

	fmt.Printf("\nMemory stored at: %s\n", dir)
}
