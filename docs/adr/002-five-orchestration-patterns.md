# ADR-002: Five Orchestration Patterns

## Status

Accepted

## Date

February 2026

## Context

Most agent frameworks offer one or two orchestration modes — typically a single-agent loop and some form of multi-agent chat. This forces users to shoehorn problems into the wrong pattern or build custom orchestration on top of the framework.

Different problems demand different architectures. A code review needs concurrent specialists (team). A content pipeline needs sequential handoff (pipeline). A research workflow needs conditional branching and loops (graph). A complex coordinator needs to spawn sub-workflows at runtime (dynamic).

## Decision

GoGrid supports five composable orchestration patterns:

1. **Single Agent** — one agent, small tool set, well-defined scope. The recommended starting point. Implemented in `pkg/agent/`.

2. **Team (Chat Room)** — concurrent agents with pub/sub messaging, shared memory, and pluggable consensus strategies (Unanimous, Majority, FirstResponse). An optional coordinator agent synthesizes the final decision. Implemented in `pkg/orchestrator/team/`.

3. **Pipeline (Linear)** — sequential handoff with state ownership transfer. Each stage has input transforms, output validation, retry policies, and per-stage cost budgets. State ownership is enforced via generation-based handles. Implemented in `pkg/orchestrator/pipeline/`.

4. **Graph** — directed graph with conditional edges, parallel fan-out, fan-in merging, and loops. Nodes execute concurrently in waves. Supports iteration limits, cost budgets, timeouts, and Graphviz DOT export. Implemented in `pkg/orchestrator/graph/`.

5. **Dynamic Orchestration** — a runtime that enables agents to spawn child agents, teams, pipelines, or graphs at runtime. Resource governance controls concurrency, nesting depth, and cost budgets. Async futures allow parallel child execution. Implemented in `pkg/orchestrator/dynamic/`.

All patterns are composable: a team can contain pipelines, a graph node can be a dynamic orchestrator.

## Consequences

- **Larger API surface.** Five patterns means more code to maintain and document. Each pattern has its own package with distinct types and options.
- **Clear pattern selection.** Users can pick the right tool for the job instead of forcing everything into a single abstraction. The [patterns guide](../patterns.md) provides a decision table.
- **Composability complexity.** Nested patterns (e.g., a dynamic runtime spawning a graph that contains a team) are powerful but require careful cost budget and timeout management.
- **Consistent interfaces.** All patterns follow the same conventions: functional options, `Config` structs, `Result` types with cost/usage tracking, and tracer integration. This reduces the learning curve across patterns.
