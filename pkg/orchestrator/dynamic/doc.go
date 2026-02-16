// Package dynamic implements GoGrid's Dynamic Orchestration pattern.
//
// Dynamic orchestration enables agents to spawn child agents, teams,
// pipelines, or graphs at runtime. This is GoGrid's most powerful
// pattern — the executing agent decides which orchestration to use
// based on the problem at hand.
//
// A Runtime manages resource governance: concurrency limits, maximum
// nesting depth, cost budgets, and cascading cancellation. Child
// orchestrations inherit the parent's tracing context and are tracked
// for aggregate cost and usage metrics.
//
// Usage:
//
//	rt := dynamic.New("coordinator",
//	    dynamic.WithConfig(dynamic.Config{
//	        MaxConcurrent: 5,
//	        MaxDepth:      3,
//	        CostBudget:    1.00,
//	    }),
//	)
//	ctx := rt.Context(ctx)
//	result, err := rt.SpawnAgent(ctx, researchAgent, "Find papers on X")
//
// For async spawning, use Go to launch children in the background:
//
//	f := rt.Go(ctx, "research", func(ctx context.Context) (string, error) {
//	    r, err := rt.SpawnAgent(ctx, researchAgent, input)
//	    if err != nil {
//	        return "", err
//	    }
//	    return r.Message.Content, nil
//	})
//	output, err := f.Wait(ctx)
//
// The Runtime is made available to child orchestrations via context,
// enabling nested dynamic spawning up to the configured MaxDepth.
package dynamic
