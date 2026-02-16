# GoGrid Orchestration Patterns

GoGrid supports five composable orchestration patterns. This guide helps you choose the right one.

## Decision Table

| Scenario | Pattern | Why |
|----------|---------|-----|
| Single task, clear scope | **Single Agent** | Simplest option, start here |
| Multiple specialists, same input | **Team** | Concurrent execution, consensus |
| Sequential processing chain | **Pipeline** | Clear handoff, state ownership |
| Conditional branching or loops | **Graph** | Edges control flow dynamically |
| Unknown structure at design time | **Dynamic** | Spawn patterns at runtime |
| Code review by multiple experts | **Team** | Concurrent review, coordinator synthesis |
| Content: research → analyze → summarize | **Pipeline** | Each stage transforms the previous output |
| Draft → review → revise loop | **Graph** | Conditional edge loops back on rejection |
| Agent decides what sub-tasks to create | **Dynamic** | Runtime spawning with governance |

## Single Agent

One agent, a small tool set, and a well-defined scope. The recommended starting point.

```go
import "github.com/lonestarx1/gogrid/pkg/agent"

a := agent.New("my-agent",
    agent.WithProvider(provider),
    agent.WithModel("gpt-4o"),
    agent.WithInstructions("You are a helpful assistant."),
    agent.WithTools(&myTool{}),
    agent.WithConfig(agent.Config{MaxTurns: 5}),
)
result, err := a.Run(ctx, "user input")
```

See [`examples/single-agent/`](../examples/single-agent/) for a complete example.

## Team (Chat Room)

Multiple agents run concurrently on the same input. A message bus enables pub/sub communication. A consensus strategy decides when to stop. An optional coordinator synthesizes the final decision.

```go
import "github.com/lonestarx1/gogrid/pkg/orchestrator/team"

t := team.New("review-team",
    team.WithMembers(
        team.Member{Agent: reviewer1, Role: "security"},
        team.Member{Agent: reviewer2, Role: "performance"},
    ),
    team.WithCoordinator(coordinator),
    team.WithConfig(team.Config{MaxRounds: 3}),
)
result, err := t.Run(ctx, "Review this code")
```

**Consensus strategies:** `Unanimous` (all respond), `Majority` (>50%), `FirstResponse` (any one).

See [`examples/team-debate/`](../examples/team-debate/) for a complete example.

## Pipeline (Linear)

Sequential handoff between specialists. Each stage owns the state exclusively — ownership transfers cleanly between stages via generation-based handles.

```go
import "github.com/lonestarx1/gogrid/pkg/orchestrator/pipeline"

p := pipeline.New("content-pipeline",
    pipeline.WithStages(
        pipeline.Stage{Name: "research", Agent: researcher},
        pipeline.Stage{Name: "analyze", Agent: analyzer,
            InputTransform: func(in string) string { return "Analyze: " + in },
        },
        pipeline.Stage{Name: "summarize", Agent: summarizer,
            Retry: pipeline.RetryPolicy{MaxAttempts: 2},
        },
    ),
    pipeline.WithProgress(func(i, total int, sr pipeline.StageResult) {
        fmt.Printf("[%d/%d] %s done\n", i+1, total, sr.Name)
    }),
)
result, err := p.Run(ctx, "Research topic X")
```

See [`examples/pipeline-research/`](../examples/pipeline-research/) for a complete example.

## Graph

Directed graph with conditional edges, parallel fan-out, fan-in merging, and loops. Nodes execute concurrently in waves.

```go
import "github.com/lonestarx1/gogrid/pkg/orchestrator/graph"

g, err := graph.NewBuilder("review-loop").
    AddNode("writer", writer).
    AddNode("reviewer", reviewer).
    AddNode("reviser", reviser).
    AddEdge("writer", "reviewer").
    AddEdge("reviewer", "reviser", graph.When(func(out string) bool {
        return strings.Contains(out, "REVISE")
    })).
    AddEdge("reviser", "reviewer").
    Options(graph.WithConfig(graph.Config{MaxIterations: 5})).
    Build()

result, err := g.Run(ctx, "Write a technical summary")
fmt.Println(g.DOT()) // Graphviz DOT export
```

See [`examples/graph-review/`](../examples/graph-review/) for a complete example.

## Dynamic Orchestration

A runtime enables agents to spawn child agents, teams, pipelines, or graphs at runtime. Resource governance controls concurrency, nesting depth, and cost budgets.

```go
import "github.com/lonestarx1/gogrid/pkg/orchestrator/dynamic"

rt := dynamic.New("coordinator",
    dynamic.WithConfig(dynamic.Config{
        MaxConcurrent: 5,
        MaxDepth:      3,
        CostBudget:    10.00,
    }),
)
ctx = rt.Context(ctx)

// Spawn children at runtime.
result, err := rt.SpawnAgent(ctx, agent, "input")
teamResult, err := rt.SpawnTeam(ctx, team, "input")

// Async futures for parallel work.
future := rt.Go(ctx, "task-name", func(ctx context.Context) (string, error) {
    return "result", nil
})
output, err := future.Wait(ctx)
```

See [`examples/dynamic-spawn/`](../examples/dynamic-spawn/) for a complete example.

## Composability

All patterns are composable:

- A **team member** can be an agent backed by a pipeline
- A **graph node** can trigger a dynamic orchestrator
- A **dynamic runtime** can spawn teams, pipelines, and graphs as children

Cost budgets, timeouts, and tracing propagate through nested patterns via `context.Context`.

## Further Reading

- [Architecture Decision Record: Five Patterns](adr/002-five-orchestration-patterns.md) — Why five patterns instead of one
- [Website Documentation](https://gogrid.org/docs) — Comprehensive API reference for all patterns
- [Testing Guide](testing.md) — How to test each pattern with the mock provider
