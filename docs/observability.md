# GoGrid Observability

GoGrid provides built-in observability through structured tracing, JSON logging, Prometheus-compatible metrics, and cost governance. All components use the Go standard library.

## Architecture Overview

```
Agent.Run() → Tracer.StartSpan() → span attributes → Tracer.EndSpan()
                  ↓                                        ↓
            otel.Exporter (OTLP)                  metrics.Collector
            log.Logger (JSON)                     cost.Tracker
```

Every agent, orchestrator, and infrastructure operation creates trace spans. These spans flow through the tracer pipeline and are consumed by exporters, loggers, metrics collectors, and cost trackers.

## Tracing

### Stdout Tracer (Development)

```go
import "github.com/lonestarx1/gogrid/pkg/trace"

tracer := trace.NewStdout(os.Stderr)
a := agent.New("my-agent", agent.WithTracer(tracer))
```

### OTLP Exporter (Production)

Export spans to Jaeger, Tempo, Zipkin, or any OTLP-compatible backend:

```go
import "github.com/lonestarx1/gogrid/pkg/trace/otel"

exporter := otel.NewExporter(
    otel.WithEndpoint("http://localhost:4318/v1/traces"),
    otel.WithServiceName("my-service"),
    otel.WithServiceVersion("1.0.0"),
    otel.WithBatchSize(100),
    otel.WithFlushInterval(5 * time.Second),
)
defer exporter.Shutdown()

a := agent.New("my-agent", agent.WithTracer(exporter))
```

### Span Attributes

GoGrid automatically sets span attributes for:

- `agent.name`, `agent.model`, `agent.cost_usd`, `agent.turns`
- `llm.model`, `llm.prompt_tokens`, `llm.completion_tokens`
- `tool.name`, `tool.status`
- `memory.operation`, `memory.key`
- `pipeline.stage.name`, `pipeline.stage.index`
- `team.name`, `team.round`, `team.consensus`
- `graph.node.name`, `graph.node.iteration`
- `dynamic.child.name`, `dynamic.child.type`, `dynamic.depth`

## Structured Logging

JSON-formatted logs with optional trace correlation.

```go
import tracelog "github.com/lonestarx1/gogrid/pkg/trace/log"

logger := tracelog.New(os.Stderr, tracelog.Info)
logger.Info("agent started", "agent", "my-agent", "model", "gpt-4o")

// With trace correlation (extracts span/trace IDs from context).
logger.InfoCtx(ctx, "llm call complete", "tokens", "150")
```

### Log Levels

`tracelog.Debug`, `tracelog.Info`, `tracelog.Warn`, `tracelog.Error`

### File Rotation

```go
writer, err := tracelog.NewFileWriter("/var/log/gogrid.log", tracelog.FileConfig{
    MaxSize:  10 * 1024 * 1024, // 10 MB
    MaxFiles: 5,
})
defer writer.Close()

logger := tracelog.New(writer, tracelog.Info)
```

## Prometheus Metrics

Automatic metrics collection from trace spans.

```go
import "github.com/lonestarx1/gogrid/pkg/trace/metrics"

registry := metrics.NewRegistry()
collector := metrics.NewCollector(innerTracer, registry)

a := agent.New("my-agent", agent.WithTracer(collector))
// ... run agents ...

// Export in Prometheus exposition format.
fmt.Print(registry.Export())
```

### Auto-populated Metrics

| Metric | Type | Labels |
|--------|------|--------|
| `gogrid_agent_runs_total` | Counter | agent, status |
| `gogrid_agent_run_duration_seconds` | Histogram | agent |
| `gogrid_llm_calls_total` | Counter | model, status |
| `gogrid_llm_call_duration_seconds` | Histogram | model |
| `gogrid_llm_tokens_total` | Counter | model, type |
| `gogrid_tool_executions_total` | Counter | tool, status |
| `gogrid_tool_execution_duration_seconds` | Histogram | tool |
| `gogrid_memory_operations_total` | Counter | operation |
| `gogrid_cost_usd_total` | Counter | agent, model |

## Cost Governance

Track costs, set budgets, and allocate costs to entities.

```go
import "github.com/lonestarx1/gogrid/pkg/cost"

tracker := cost.NewTracker()

// Budget with threshold alerts.
tracker.SetBudget(10.00)
tracker.OnBudgetThreshold(func(threshold, current float64) {
    log.Printf("ALERT: %.0f%% of budget ($%.2f)", threshold*100, current)
}, 0.5, 0.8, 1.0)

// Record usage (returns cost).
cost := tracker.Add("gpt-4o", usage)

// Per-entity allocation.
tracker.AddForEntity("gpt-4o", "research-agent", usage)

// Reports.
report := tracker.Report()
```

Cost budgets are also supported at every orchestration level:

- `agent.Config{CostBudget: 0.50}`
- `team.Config{CostBudget: 2.00}`
- `pipeline.Config{CostBudget: 1.00}` and per-stage `Stage{CostBudget: 0.25}`
- `graph.Config{CostBudget: 5.00}`
- `dynamic.Config{CostBudget: 10.00}`

See [`examples/observability/`](../examples/observability/) for a runnable example combining all components.

## Further Reading

- [Website Documentation](https://gogrid.org/docs) — Comprehensive sections on tracing, logging, metrics, and cost governance
- [Testing Guide](testing.md) — Benchmarks for measuring framework overhead
