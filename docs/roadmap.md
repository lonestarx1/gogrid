# GoGrid (G2) — Implementation Plan (plan0)

> Multi-phase roadmap from empty repo to production-grade AI agent framework.
> Each phase builds on the previous. No phase ships without tests.

---

## Phase 0: Project Scaffold & Core Types

**Goal:** Establish the Go module, project layout, build tooling, and foundational types that everything else depends on.

### 0.1 — Go Module & Project Structure

- `go mod init github.com/gogridai/gogrid`
- Create directory skeleton:
  ```
  cmd/gogrid/          # CLI entry point (stub)
  pkg/agent/           # Public: agent definition, lifecycle
  pkg/tool/            # Public: tool interface, registry
  pkg/memory/          # Public: memory interfaces and implementations
  pkg/llm/             # Public: LLM provider interface, message types
  pkg/orchestrator/    # Public: orchestration patterns
  pkg/cost/            # Public: cost tracking primitives
  pkg/trace/           # Public: observability primitives
  internal/config/     # Private: configuration loading
  internal/id/         # Private: ID generation
  examples/            # Example programs
  ```
- Add `Makefile` with targets: `build`, `test`, `lint`, `fmt`, `vet`
- Add `.golangci.yml` linter configuration
- Add `go.sum` / dependency management

### 0.2 — Core Types & Interfaces

These are the foundational types the entire framework builds on.

**Message types** (`pkg/llm/`):
- `Role` type (system, user, assistant, tool)
- `Message` struct (role, content, tool calls, tool results, metadata)
- `ToolCall` struct (ID, function name, arguments)
- `ToolResult` struct (ID, content, error)

**LLM Provider interface** (`pkg/llm/`):
- `Provider` interface: `Complete(ctx, params) (Response, error)`
- `Params` struct: model, messages, tools, temperature, max tokens, stop sequences
- `Response` struct: message, usage (prompt tokens, completion tokens), model, provider metadata
- `Usage` struct: prompt tokens, completion tokens, total tokens
- `StreamProvider` interface (optional extension): `Stream(ctx, params) (Stream, error)`

**Tool interface** (`pkg/tool/`):
- `Tool` interface: `Name() string`, `Description() string`, `Schema() Schema`, `Execute(ctx, input) (output, error)`
- `Schema` struct wrapping JSON Schema for tool parameters
- `Registry` for registering and looking up tools by name

**Agent definition** (`pkg/agent/`):
- `Agent` struct: name, instructions (system prompt), tools, model, provider, memory, config
- `AgentConfig`: max turns, max tokens, temperature, timeout, cost budget
- `AgentBuilder` — functional options pattern for constructing agents
- `Run(ctx) (Result, error)` — execute the agent loop
- `Result` struct: final message, conversation history, usage summary, cost

**Memory interface** (`pkg/memory/`):
- `Memory` interface: `Load(ctx, key) ([]Message, error)`, `Save(ctx, key, []Message) error`, `Clear(ctx, key) error`
- `InMemory` implementation (for development/testing)
- `ReadOnlyMemory` wrapper

**ID generation** (`internal/id/`):
- Short, sortable unique IDs for agents, runs, traces (ULID or similar)

### 0.3 — Deliverables

- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes with tests for core types
- [ ] `make lint` passes
- [ ] All exported types have doc comments
- [ ] Table-driven tests for message construction, tool registry, memory operations

---

## Phase 1: Single Agent Pattern

**Goal:** A single GoGrid agent can take a system prompt, a set of tools, call an LLM, execute tool calls, and return a result. This is the foundational loop that everything else composes.

### 1.1 — Agent Execution Loop

- Implement the core agent loop in `pkg/agent/`:
  1. Load system prompt + memory → build initial messages
  2. Call LLM provider with messages + tool definitions
  3. If response contains tool calls → execute tools → append results → loop
  4. If response is a final message → return result
  5. Respect max turns, timeout (via `context.Context`), and cost budget
- Error handling: retry with exponential backoff on transient LLM errors
- Clean cancellation via `context.Context`

### 1.2 — OpenAI-Compatible Provider

- Implement `Provider` for OpenAI-compatible APIs (`pkg/llm/openai/`)
- HTTP client using `net/http` (no SDK dependency)
- Support: chat completions, tool use, streaming
- API key from config/environment
- Respect rate limits (429 → backoff)

### 1.3 — Anthropic Provider

- Implement `Provider` for Anthropic API (`pkg/llm/anthropic/`)
- HTTP client using `net/http`
- Map GoGrid message format ↔ Anthropic message format
- Support: messages API, tool use, streaming

### 1.4 — Cost Tracking (Basic)

- `pkg/cost/` — `Tracker` that accumulates token usage per run
- Pricing table per model (configurable)
- `CostBudget` on `AgentConfig` — agent stops if budget exceeded
- Cost data included in `Result`

### 1.5 — Tracing (Basic)

- `pkg/trace/` — `Span` struct with start/end, parent span, agent name, attributes
- `Tracer` interface: `StartSpan(ctx, name) (ctx, Span)`, `EndSpan(Span)`
- In-memory tracer implementation for testing/debugging
- Stdout tracer implementation (structured JSON logs)
- Every LLM call and tool execution creates a span

### 1.6 — Deliverables

- [ ] Working single agent that can converse and use tools
- [ ] At least two LLM providers (OpenAI, Anthropic)
- [ ] Cost tracking with budget enforcement
- [ ] Basic tracing with structured output
- [ ] Example: `examples/single-agent/` — a simple agent with calculator tools
- [ ] Example: `examples/web-search/` — agent that uses a web search tool
- [ ] Integration tests (with mock LLM provider)
- [ ] Unit tests for each component

---

## Phase 2: Memory System

**Goal:** Upgrade memory from a basic interface to a production-grade primitive. Memory is a first-class citizen in GoGrid.

### 2.1 — Memory Backends

- `InMemory` — already exists from Phase 0, production-hardened
- `FileMemory` (`pkg/memory/file/`) — JSON file-backed, for local development
- `Memory` interface extended:
  - `Search(ctx, query, limit) ([]Entry, error)` — semantic or keyword search
  - `Prune(ctx, policy) error` — remove old/irrelevant entries
  - `Stats(ctx) (Stats, error)` — memory size, entry count, cost of storage

### 2.2 — Shared Memory Pool

- `SharedMemory` (`pkg/memory/shared/`) — concurrent-safe memory for team pattern
- Read-write locking with configurable isolation levels
- Namespaced access — agents can have scoped views of shared memory
- Change notification (observer pattern) — agents notified when shared memory changes

### 2.3 — State Ownership Transfer

- `TransferableState` (`pkg/memory/transfer/`) — for pipeline pattern
- State is moved, not copied — previous owner loses access after transfer
- Validation hooks on transfer (pre-transfer checks)
- Audit trail — who owned what state and when

### 2.4 — Memory Monitoring

- Memory operations emit trace spans
- Memory size/cost visible in agent `Result`
- Configurable pruning policies (max entries, max age, max tokens)

### 2.5 — Deliverables

- [ ] At least 3 memory backends (in-memory, file, shared)
- [ ] State ownership transfer with safety guarantees
- [ ] Memory monitoring and pruning
- [ ] Comprehensive tests for concurrency safety
- [ ] Benchmarks for shared memory under contention

---

## Phase 3: Team (Chat Room) Pattern

**Goal:** Multiple agents collaborate concurrently via pub/sub messaging with shared memory, reaching consensus on a task.

### 3.1 — Message Bus

- `pkg/orchestrator/team/` — team orchestrator
- `MessageBus` interface: `Publish(ctx, topic, Message) error`, `Subscribe(ctx, topic, handler) error`
- In-memory message bus implementation
- Topics: general (all agents), direct (agent-to-agent), system (orchestrator commands)

### 3.2 — Team Orchestrator

- `Team` struct: name, agents, shared memory, message bus, config
- `TeamConfig`: max rounds, consensus strategy, timeout, cost budget
- Execution model:
  1. Orchestrator posts the task to the general topic
  2. Agents run concurrently (goroutines), each receiving messages via subscription
  3. Agents read/write shared memory, publish responses
  4. Orchestrator evaluates consensus strategy each round
  5. Team terminates when consensus reached, max rounds hit, or timeout/budget exceeded

### 3.3 — Consensus Strategies

- `Consensus` interface: `Evaluate(ctx, messages) (done bool, result Message, error)`
- Built-in strategies:
  - `UnanimousConsensus` — all agents agree
  - `MajorityConsensus` — majority of agents agree
  - `OrchestratorDecision` — a designated "lead" agent decides when done
  - `MaxRounds` — fixed number of rounds, take last state

### 3.4 — Deliverables

- [ ] Working team of agents that debate and reach consensus
- [ ] Shared memory integration
- [ ] Multiple consensus strategies
- [ ] Example: `examples/team-debate/` — agents with different perspectives reach a decision
- [ ] Tests for concurrent message passing and consensus
- [ ] Tests for team cost budget enforcement

---

## Phase 4: Pipeline (Linear) Pattern

**Goal:** Sequential chain of agents where each agent completes its work, transfers state ownership to the next, and terminates.

### 4.1 — Pipeline Orchestrator

- `pkg/orchestrator/pipeline/` — pipeline orchestrator
- `Pipeline` struct: name, stages, config
- `Stage` struct: agent, input transform, output validation, retry policy
- Execution model:
  1. First agent receives initial input
  2. Agent processes, produces output
  3. State ownership transfers to next agent (previous agent's state is sealed)
  4. Repeat until final stage
  5. Pipeline returns final stage output

### 4.2 — State Transfer Protocol

- Integrates with `memory/transfer/` from Phase 2
- Each stage receives owned state — can read and modify
- On handoff: state is validated, sealed, transferred
- Failed stage can retry without re-running previous stages (checkpoint)

### 4.3 — Pipeline Control

- Stage-level error handling: retry, skip, abort pipeline
- Stage-level timeouts and cost budgets (independent of pipeline-level)
- Pipeline-level timeout and cost budget (aggregate)
- Progress reporting — which stage is active, what has completed

### 4.4 — Deliverables

- [ ] Working pipeline with state transfer between agents
- [ ] Checkpointing and retry at stage level
- [ ] Example: `examples/pipeline-research/` — research → analyze → summarize pipeline
- [ ] Tests for state ownership enforcement
- [ ] Tests for failure recovery and checkpointing

---

## Phase 5: Graph Pattern

**Goal:** Like pipeline but with conditional branches, parallel execution paths, and loops (re-do / clarify).

### 5.1 — Graph Orchestrator

- `pkg/orchestrator/graph/` — graph orchestrator
- `Graph` struct: name, nodes, edges, config
- `Node`: wraps an agent (or sub-orchestrator)
- `Edge`: source node → target node, with condition function
- `EdgeCondition`: `func(ctx, NodeResult) (bool, error)`

### 5.2 — Graph Execution Engine

- Topological execution respecting dependencies
- Parallel execution of independent branches (fan-out)
- Fan-in: merge results from parallel branches
- Loops: edges can point backward (with max iteration guard)
- Conditional routing: edge conditions determine which path to take

### 5.3 — Graph Builder

- Fluent API for constructing graphs:
  ```go
  g := graph.New("review-pipeline").
      AddNode("draft", draftAgent).
      AddNode("review", reviewAgent).
      AddNode("revise", reviseAgent).
      AddNode("publish", publishAgent).
      AddEdge("draft", "review").
      AddEdge("review", "revise", graph.When(needsRevision)).
      AddEdge("review", "publish", graph.When(approved)).
      AddEdge("revise", "review").  // loop back
      Build()
  ```

### 5.4 — Deliverables

- [ ] Working graph with branches, parallel paths, and loops
- [ ] Graph visualization output (DOT format for Graphviz)
- [ ] Example: `examples/graph-review/` — draft → review → revise loop → publish
- [ ] Tests for cycle detection, max iteration enforcement
- [ ] Tests for parallel fan-out/fan-in correctness

---

## Phase 6: Dynamic Orchestration

**Goal:** Agents can spawn child agents, teams, pipelines, or graphs at runtime. The most powerful GoGrid pattern.

### 6.1 — Runtime Spawning

- `pkg/orchestrator/dynamic/` — dynamic orchestrator
- `Runtime` interface available to agents:
  - `SpawnAgent(ctx, config) (Result, error)`
  - `SpawnTeam(ctx, config) (Result, error)`
  - `SpawnPipeline(ctx, config) (Result, error)`
  - `SpawnGraph(ctx, config) (Result, error)`
- Spawned children inherit parent's tracing context, cost budget (subdivided), and security policy
- Parent agent can await child results or fire-and-forget

### 6.2 — Resource Governance

- Child cost budgets are carved from parent budget
- Max concurrent children per agent
- Max depth of dynamic spawning (prevent infinite recursion)
- Resource cleanup on parent cancellation (cascading context cancellation)

### 6.3 — Deliverables

- [ ] Agents can dynamically spawn any orchestration pattern
- [ ] Budget subdivision and resource governance
- [ ] Example: `examples/dynamic-research/` — agent decides at runtime whether to spawn a team, pipeline, or single sub-agent
- [ ] Tests for resource limits, cascading cancellation, budget enforcement

---

## Phase 7: Observability & Cost Governance (Production Grade)

**Goal:** Upgrade the basic tracing and cost tracking from Phase 1 into production-grade observability.

### 7.1 — OpenTelemetry Integration

- `pkg/trace/otel/` — OpenTelemetry-compatible tracer
- Export spans to any OTLP-compatible backend (Jaeger, Zipkin, Grafana Tempo)
- Semantic conventions for GoGrid spans:
  - `gogrid.agent.name`, `gogrid.agent.run_id`
  - `gogrid.llm.model`, `gogrid.llm.provider`
  - `gogrid.llm.tokens.prompt`, `gogrid.llm.tokens.completion`
  - `gogrid.tool.name`, `gogrid.tool.duration`
  - `gogrid.cost.usd`

### 7.2 — Structured Logging

- `pkg/trace/log/` — structured logging (JSON) integrated with tracing
- Log levels: debug, info, warn, error
- Correlation IDs linking logs to trace spans
- Opt-in file logging with rotation

### 7.3 — Cost Governance (Advanced)

- Real-time cost dashboarding data (expose via metrics)
- Cost alerts — callback when approaching budget
- Cost allocation — attribute costs to specific agents, teams, or orchestrations
- Historical cost reporting per run, per agent, per model

### 7.4 — Metrics

- `pkg/trace/metrics/` — Prometheus-compatible metrics
- Agent run duration, success/failure rate
- LLM call latency, token throughput
- Tool execution duration
- Memory size and operation count
- Cost per run, per model

### 7.5 — Deliverables

- [ ] OpenTelemetry export working with at least one backend
- [ ] Prometheus metrics endpoint
- [ ] Structured logging with trace correlation
- [ ] Advanced cost governance with alerts and allocation
- [ ] Example: `examples/observability/` — fully instrumented agent with dashboard config

---

## Phase 8: Security Integration

**Goal:** Secure-by-default with Tenio.ai integration. Prompt injection protection, tool authorization, data exfiltration prevention.

### 8.1 — Tool Authorization

- `pkg/security/` — security middleware
- `ToolPolicy`: allow/deny lists for tool access per agent
- `ToolApproval` hook: require approval before certain tool executions
- Audit log for all tool executions

### 8.2 — Prompt Injection Protection

- Input sanitization hooks (pre-LLM-call)
- Output validation hooks (post-LLM-call)
- Tenio.ai integration for automated prompt injection detection
- Configurable: strict mode (block suspicious), warn mode (log and continue)

### 8.3 — Data Exfiltration Prevention

- Output filters: prevent agents from leaking sensitive data in tool calls or responses
- Redaction rules: configurable patterns to redact (SSN, API keys, etc.)
- Network boundary enforcement: restrict which URLs/endpoints tools can access

### 8.4 — Deliverables

- [ ] Tool authorization with policy enforcement
- [ ] Prompt injection detection (built-in heuristics + Tenio.ai)
- [ ] Data exfiltration prevention
- [ ] Security audit logging
- [ ] Example: `examples/secure-agent/` — agent with full security policy
- [ ] Tests for each security layer

---

## Phase 9: CLI & Developer Experience

**Goal:** A `gogrid` CLI that makes creating, running, testing, and monitoring agents frictionless.

### 9.1 — CLI Foundation

- `cmd/gogrid/` — CLI entry point
- Subcommands:
  - `gogrid init` — scaffold a new GoGrid project
  - `gogrid run <agent>` — run an agent from config
  - `gogrid list` — list available agents in project
  - `gogrid trace <run-id>` — view trace for a past run
  - `gogrid cost <run-id>` — view cost breakdown for a run
  - `gogrid version` — print version info

### 9.2 — Agent Configuration

- YAML/TOML configuration for agents:
  ```yaml
  agent:
    name: "researcher"
    model: "claude-sonnet-4-5-20250929"
    provider: "anthropic"
    instructions: "You are a research assistant..."
    tools:
      - web_search
      - file_read
    config:
      max_turns: 10
      cost_budget: 0.50
      timeout: 60s
  ```
- Environment variable substitution in config
- Config validation on load

### 9.3 — Project Templates

- `gogrid init --template single` — single agent project
- `gogrid init --template team` — team collaboration project
- `gogrid init --template pipeline` — pipeline project
- Each template includes working example code, config, and tests

### 9.4 — Deliverables

- [ ] Working CLI with init, run, list, trace, cost subcommands
- [ ] YAML-based agent configuration
- [ ] Project templates
- [ ] `gogrid` compiles to a single binary
- [ ] Tests for CLI commands

---

## Phase 10: Evaluation Framework & Testing Utilities

**Goal:** Built-in tools for evaluating agent performance, running benchmarks, and testing agent behavior deterministically.

### 10.1 — Mock Provider

- `pkg/llm/mock/` — deterministic LLM provider for testing
- Pre-programmed responses, can assert on inputs
- Latency simulation for performance testing
- Error injection for resilience testing

### 10.2 — Evaluation Framework

- `pkg/eval/` — agent evaluation primitives
- `Evaluator` interface: `Evaluate(ctx, result, criteria) (Score, error)`
- Built-in evaluators:
  - `ExactMatch` — output matches expected string
  - `Contains` — output contains expected substrings
  - `LLMJudge` — use an LLM to evaluate another agent's output
  - `ToolUse` — assert specific tools were called with specific arguments
  - `CostWithin` — assert cost stayed within budget
  - `CompletedWithin` — assert run completed within time limit

### 10.3 — Benchmark Suite

- `pkg/eval/bench/` — standardized benchmarks
- Agent throughput (runs/second)
- Concurrent agent scaling
- Memory usage under load
- LLM call latency distribution

### 10.4 — Deliverables

- [ ] Mock provider for deterministic testing
- [ ] Evaluation framework with built-in evaluators
- [ ] Benchmark suite
- [ ] Example: `examples/eval/` — evaluating an agent with multiple criteria
- [ ] Documentation for writing custom evaluators

---

## Phase 11: Documentation & Examples

**Goal:** Comprehensive documentation that matches the quality bar of the framework itself.

### 11.1 — API Documentation

- Godoc for all exported types and functions (ongoing from Phase 0)
- Package-level documentation explaining design intent

### 11.2 — Guides

- Getting Started guide
- Single Agent tutorial
- Team Pattern tutorial
- Pipeline Pattern tutorial
- Graph Pattern tutorial
- Dynamic Orchestration tutorial
- Memory deep-dive
- Observability setup guide
- Security configuration guide
- Cost management guide
- Testing and evaluation guide

### 11.3 — Examples

Accumulated across all phases:
- `examples/single-agent/` — basic agent with tools
- `examples/web-search/` — agent with web search
- `examples/team-debate/` — team reaching consensus
- `examples/pipeline-research/` — research pipeline
- `examples/graph-review/` — review loop graph
- `examples/dynamic-research/` — dynamic orchestration
- `examples/observability/` — full observability setup
- `examples/secure-agent/` — security policy enforcement
- `examples/eval/` — evaluation framework usage
- `examples/multi-provider/` — swapping between LLM providers

### 11.4 — Deliverables

- [ ] Complete Godoc coverage
- [ ] Getting Started guide on gogrid.org
- [ ] Tutorial for each orchestration pattern
- [ ] 10+ working examples
- [ ] Architecture decision records (ADRs) for key decisions

---

## Phase Dependency Graph

```
Phase 0 (Scaffold & Core Types)
    │
    ▼
Phase 1 (Single Agent) ──────────────────────┐
    │                                         │
    ▼                                         │
Phase 2 (Memory System)                       │
    │                                         │
    ├──────────────┬──────────────┐           │
    ▼              ▼              ▼           │
Phase 3        Phase 4        Phase 5         │
(Team)         (Pipeline)     (Graph)         │
    │              │              │           │
    └──────────────┴──────────────┘           │
                   │                          │
                   ▼                          │
            Phase 6 (Dynamic)                 │
                   │                          │
         ┌────────┼─────────┐                │
         ▼        ▼         ▼                ▼
    Phase 7    Phase 8    Phase 9     Phase 10
    (Observ)   (Security)  (CLI)      (Eval)
         │        │         │            │
         └────────┴─────────┴────────────┘
                        │
                        ▼
                  Phase 11 (Docs)
```

---

## Guiding Principles Across All Phases

1. **Tests ship with code.** No phase is complete without unit and integration tests.
2. **Every LLM call is metered.** Cost tracking from Phase 1 onward.
3. **Every operation is traceable.** Spans from Phase 1 onward.
4. **`context.Context` everywhere.** Cancellation, timeouts, and request-scoped values in every public API.
5. **No external dependencies unless justified.** Standard library first.
6. **Backward compatibility.** Public APIs are stable once shipped. Deprecate before removing.
7. **Interfaces at consumption site.** Define interfaces where they are used, not where they are implemented.
8. **Error wrapping with `%w`.** Every returned error includes context about what failed and why.

---

*This plan is a living document. Update it as GoGrid evolves.*
