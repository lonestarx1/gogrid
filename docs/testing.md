# Testing GoGrid Agents

This guide covers the mock provider, evaluation framework, and benchmarks.

## Mock Provider

The `pkg/llm/mock` package provides a configurable mock LLM provider for testing. No API keys needed.

### Basic Usage

```go
import (
    "github.com/lonestarx1/gogrid/pkg/llm"
    "github.com/lonestarx1/gogrid/pkg/llm/mock"
)

// Fixed response for all calls.
provider := mock.New(mock.WithFallback(&llm.Response{
    Message: llm.NewAssistantMessage("mock response"),
    Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
    Model:   "mock",
}))
```

### Sequential Responses

```go
// First call returns response1, second returns response2, then fallback.
provider := mock.New(
    mock.WithResponses(response1, response2),
    mock.WithFallback(fallbackResponse),
)
```

### Error Injection

```go
// Always fail.
provider := mock.New(mock.WithError(errors.New("api unavailable")))

// Fail first 2 calls, then succeed.
provider := mock.New(
    mock.WithFailCount(2),
    mock.WithFallback(successResponse),
)
```

### Latency Simulation

```go
// Simulate 100ms LLM latency (respects context cancellation).
provider := mock.New(
    mock.WithDelay(100 * time.Millisecond),
    mock.WithFallback(response),
)
```

### Call Recording

```go
provider := mock.New(mock.WithFallback(response))
// ... run agent ...
fmt.Println(provider.Calls())   // Number of Complete calls
history := provider.History()    // All recorded call params
provider.Reset()                 // Clear history, keep config
```

See [`examples/multi-provider/`](../examples/multi-provider/) and [`examples/eval/`](../examples/eval/) for runnable examples.

## Evaluation Framework

The `pkg/eval` package provides composable evaluators for scoring agent outputs.

### Built-in Evaluators

| Evaluator | What it checks |
|-----------|---------------|
| `ExactMatch` | Output exactly equals expected string |
| `Contains` | Output includes specified substrings |
| `CostWithin` | Run cost stays under a USD budget |
| `ToolUse` | History includes expected tool calls |
| `LLMJudge` | An LLM scores output against a rubric |
| `CompletedWithin` | Run completed within a duration (use with `Func`) |

### Evaluation Suite

Compose multiple evaluators into a suite:

```go
import "github.com/lonestarx1/gogrid/pkg/eval"

suite := eval.NewSuite(
    eval.NewContains("Go", "concurrency"),
    eval.NewCostWithin(0.05),
    eval.NewFunc("min_length", func(_ context.Context, r *agent.Result) (eval.Score, error) {
        if len(r.Message.Content) >= 20 {
            return eval.Score{Pass: true, Value: 1.0, Reason: "sufficient length"}, nil
        }
        return eval.Score{Pass: false, Value: 0.0, Reason: "too short"}, nil
    }),
)

result, err := suite.Run(ctx, agentResult)
fmt.Println(result.Pass) // true if all evaluators passed
```

### Score Structure

Every evaluator returns a `Score`:

- **Pass** (bool) — primary signal for CI gates
- **Value** (float64, 0.0-1.0) — granularity for trend analysis
- **Reason** (string) — human-readable explanation

### LLM-as-Judge

Use an LLM to evaluate output quality:

```go
judge := eval.NewLLMJudge(provider, "gpt-4o", "Rate the clarity and accuracy of this response.")
score, err := judge.Evaluate(ctx, agentResult)
// Score.Value = LLM rating / 10, Pass = rating >= 7
```

See [`examples/eval/`](../examples/eval/) for a runnable example.

## Benchmarks

The `pkg/eval/bench` package measures performance across GoGrid patterns.

### Running Benchmarks

```bash
go test -bench=. ./pkg/eval/bench/

# With memory profiling
go test -bench=. -benchmem ./pkg/eval/bench/

# Specific benchmark
go test -bench=BenchmarkAgentRun ./pkg/eval/bench/
```

### Available Benchmarks

| Benchmark | What it measures |
|-----------|-----------------|
| `BenchmarkAgentRun` | Basic agent execution |
| `BenchmarkAgentRunWithToolUse` | Agent with tool calling loop |
| `BenchmarkAgentRunParallel` | Concurrent agent execution |
| `BenchmarkPipelineThreeStages` | Fixed three-stage pipeline |
| `BenchmarkPipelineScaling` | Pipeline with 1, 3, 5, 10 stages |
| `BenchmarkTeamTwoMembers` | Two-agent team execution |
| `BenchmarkTeamScaling` | Team with 1, 2, 5, 10, 20 members |
| `BenchmarkSharedMemorySaveLoad` | Memory load/save operations |
| `BenchmarkSharedMemoryContention` | Memory with 1, 2, 5, 10 concurrent writers |

All benchmarks use the mock provider to measure framework overhead, not LLM latency.

## Further Reading

- [ADR-005: Evaluation Framework Design](adr/005-eval-framework-design.md) — Design decisions
- [ADR-004: No Vendor Lock-In](adr/004-no-vendor-lock-in.md) — Mock provider rationale
- [Website Documentation](https://gogrid.org/docs) — Comprehensive API reference
