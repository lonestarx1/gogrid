// Example observability demonstrates GoGrid's full observability stack:
// OTLP tracing, structured JSON logging, Prometheus metrics, and cost
// governance — all wired together.
//
// Usage:
//
//	go run ./examples/observability
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/cost"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/llm/mock"
	"github.com/lonestarx1/gogrid/pkg/trace"
	tracelog "github.com/lonestarx1/gogrid/pkg/trace/log"
	"github.com/lonestarx1/gogrid/pkg/trace/metrics"
)

func main() {
	ctx := context.Background()

	// --- Structured Logging ---
	logger := tracelog.New(os.Stderr, tracelog.Info)
	logger.Info("starting observability example", "component", "main")

	// --- Metrics ---
	// Create a registry and wrap a base tracer with metrics collection.
	// In production, use otel.NewExporter() instead of trace.Noop{}.
	registry := metrics.NewRegistry()
	metricsCollector := metrics.NewCollector(trace.Noop{}, registry)

	// --- Cost Governance ---
	tracker := cost.NewTracker()
	tracker.SetBudget(1.00)
	tracker.OnBudgetThreshold(func(threshold, current float64) {
		logger.Warn("budget alert",
			"threshold", fmt.Sprintf("%.0f%%", threshold*100),
			"current", fmt.Sprintf("$%.4f", current),
		)
	}, 0.5, 0.8, 1.0)

	// --- Agent with Observability ---
	provider := mock.New(mock.WithFallback(&llm.Response{
		Message: llm.NewAssistantMessage("Observability in GoGrid combines structured tracing, JSON logging, Prometheus metrics, and cost governance into a unified stack."),
		Usage:   llm.Usage{PromptTokens: 25, CompletionTokens: 22, TotalTokens: 47},
		Model:   "mock",
	}))

	a := agent.New("observed-agent",
		agent.WithProvider(provider),
		agent.WithModel("mock"),
		agent.WithTracer(metricsCollector),
		agent.WithInstructions("You are an observability expert."),
		agent.WithConfig(agent.Config{MaxTurns: 3}),
	)

	// Run the agent.
	logger.Info("running agent", "agent", "observed-agent")
	result, err := a.Run(ctx, "What is observability in GoGrid?")
	if err != nil {
		log.Fatalf("Run: %v", err)
	}

	// Record cost.
	agentCost := tracker.AddForEntity("mock", "observed-agent", result.Usage)
	logger.Info("agent complete",
		"turns", fmt.Sprintf("%d", result.Turns),
		"cost", fmt.Sprintf("$%.6f", agentCost),
		"tokens", fmt.Sprintf("%d", result.Usage.TotalTokens),
	)

	fmt.Printf("\nAgent: %s\n\n", result.Message.Content)

	// --- Cost Report ---
	report := tracker.Report()
	fmt.Printf("=== Cost Report ===\n")
	fmt.Printf("Total: $%.6f across %d calls\n", report.TotalCost, report.RecordCount)
	for model, mr := range report.ByModel {
		fmt.Printf("  %s: %d calls, $%.6f, %d tokens\n",
			model, mr.Calls, mr.Cost, mr.Usage.TotalTokens)
	}

	// --- Metrics Snapshot (Prometheus format) ---
	fmt.Printf("\n=== Metrics ===\n")
	fmt.Print(registry.Export())
}
