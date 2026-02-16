# ADR-005: Evaluation Framework Design

## Status

Accepted

## Date

2025-06-15

## Context

Agent outputs are non-deterministic. A framework that cannot measure agent quality is incomplete. Existing evaluation tools are either Python-only, require external services, or focus narrowly on LLM output quality without considering operational metrics like cost and tool usage.

GoGrid needs an evaluation framework that:
- Works with Go's standard testing tools
- Supports both deterministic and LLM-based evaluation
- Measures operational metrics (cost, duration, tool usage), not just output quality
- Composes evaluators into suites for comprehensive testing

## Decision

GoGrid provides `pkg/eval/` with a composable evaluator architecture:

**Core types:**

- `Evaluator` interface — `Name()` and `Evaluate(ctx, *agent.Result) (Score, error)`
- `Score` — `Pass` (bool), `Value` (0.0-1.0), `Reason` (string)
- `Func` — wraps any function as an Evaluator for inline/one-off evaluations
- `Suite` — runs multiple evaluators against a single result, aggregates into `SuiteResult`

**Built-in evaluators:**

1. `ExactMatch` — output must equal an expected string
2. `Contains` — output must include specified substrings (Value = fraction found)
3. `CostWithin` — run cost must stay under a USD budget
4. `ToolUse` — conversation history must include expected tool calls with optional argument matching
5. `LLMJudge` — an LLM scores the output against a rubric (0-10 scale, >=7 passes)
6. `CompletedWithin` — standalone function for duration checks (use with `Func`)

**Benchmarks** (`pkg/eval/bench/`) measure performance across patterns using the mock provider.

Design decisions:

- **Score.Pass is the primary signal.** Value provides granularity for trend analysis, but pass/fail drives CI gates.
- **Evaluators receive the full `agent.Result`.** This gives access to cost, usage, turns, history, and memory stats — not just the output text.
- **LLMJudge uses a configurable prompt template.** Users can customize the rubric and scoring criteria while the framework handles response parsing.
- **Suite stops on errors but continues on failures.** A low score is not an error. Errors indicate infrastructure problems (e.g., the judge LLM is unreachable).

## Consequences

- **Go-native testing.** Evaluators work with `go test` and standard benchmarks. No external evaluation services required.
- **Operational metrics.** CostWithin and ToolUse evaluate aspects that traditional NLP metrics ignore. This aligns with GoGrid's production-first philosophy.
- **LLMJudge requires an LLM.** The most flexible evaluator needs an LLM provider. The mock provider can be used for testing the judge itself.
- **Extensibility.** The Evaluator interface is simple enough that users can implement domain-specific evaluators (e.g., JSON schema validation, sentiment analysis) with minimal effort.
