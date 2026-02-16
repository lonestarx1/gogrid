package cli

import (
	"context"
	"flag"
	"time"

	"github.com/lonestarx1/gogrid/internal/config"
	"github.com/lonestarx1/gogrid/internal/runrecord"
	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/trace"
)

func (a *App) runRun(args []string) int {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(a.stderr)
	configPath := fs.String("config", "gogrid.yaml", "path to gogrid.yaml")
	input := fs.String("input", "", "input text (reads stdin if empty)")
	timeout := fs.Duration("timeout", 0, "override timeout (e.g. 30s, 5m)")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if fs.NArg() == 0 {
		a.errf("Usage: gogrid run <agent-name> [flags]\n")
		return 1
	}
	agentName := fs.Arg(0)

	cfg, err := config.Load(*configPath)
	if err != nil {
		a.errf("Error: %v\n", err)
		return 1
	}

	agentCfg, ok := cfg.Agents[agentName]
	if !ok {
		a.errf("Error: unknown agent %q\n", agentName)
		a.errf("Available agents:\n")
		for name := range cfg.Agents {
			a.errf("  - %s\n", name)
		}
		return 1
	}

	// Read input from flag or stdin.
	inputText := *input
	if inputText == "" {
		a.errf("Error: no input provided (use -input flag or pipe via stdin)\n")
		return 1
	}

	// Resolve provider.
	ctx := context.Background()
	provider, err := a.providerFactory(ctx, agentCfg.Provider)
	if err != nil {
		a.errf("Error: %v\n", err)
		return 1
	}

	// Build agent options.
	agentTimeout := agentCfg.Config.Timeout.Duration
	if *timeout > 0 {
		agentTimeout = *timeout
	}

	tracer := trace.NewInMemory()

	ag := agent.New(agentName,
		agent.WithModel(agentCfg.Model),
		agent.WithProvider(provider),
		agent.WithInstructions(agentCfg.Instructions),
		agent.WithTracer(tracer),
		agent.WithConfig(agent.Config{
			MaxTurns:    agentCfg.Config.MaxTurns,
			MaxTokens:   agentCfg.Config.MaxTokens,
			Temperature: agentCfg.Config.Temperature,
			Timeout:     agentTimeout,
			CostBudget:  agentCfg.Config.CostBudget,
		}),
	)

	// Execute.
	start := time.Now()
	result, err := ag.Run(ctx, inputText)
	duration := time.Since(start)

	// Build run record.
	rec := &runrecord.Record{
		Agent:     agentName,
		Model:     agentCfg.Model,
		Provider:  agentCfg.Provider,
		Input:     inputText,
		StartTime: start,
		Duration:  duration,
	}

	if err != nil {
		rec.Error = err.Error()
		rec.RunID = "error-" + time.Now().Format("20060102-150405")
		rec.Spans = tracer.Spans()
		// Still save the record for debugging.
		_ = runrecord.Save(".", rec)
		a.errf("Error: %v\n", err)
		return 1
	}

	rec.RunID = result.RunID
	rec.Output = result.Message.Content
	rec.Turns = result.Turns
	rec.Usage = result.Usage
	rec.Cost = result.Cost
	rec.Spans = tracer.Spans()

	// Print response.
	a.outf("%s\n", result.Message.Content)

	// Save run record.
	if err := runrecord.Save(".", rec); err != nil {
		a.errf("Warning: failed to save run record: %v\n", err)
	} else {
		a.errf("\nRun ID: %s\n", rec.RunID)
	}

	return 0
}
