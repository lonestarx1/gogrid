# GoGrid (G2) — Develop and Orchestrate AI Agents

> **[gogrid.org](https://gogrid.org)**
> A unified system for developing and orchestrating production AI agents in Go.
> Built for infrastructure engineers, not notebook demos.

---

## Why Another Agent System?

Every existing agent framework gets something right — and gets something fundamentally wrong.

LangChain has the ecosystem but drowns you in abstraction. CrewAI has the mental model but falls apart in production. AutoGen pioneered multi-agent conversations but got rewritten twice. Most are Python-only, prototype-first, and treat production as an afterthought.

**80-90% of AI agent projects never leave the pilot phase.** (RAND, 2025)

GoGrid is the system for the other side of that gap.

GoGrid is not a wrapper around LLM APIs. It fuses agent development with agent orchestration into a single system — the way Kubernetes fused container packaging with container management. The agent you define IS the agent that gets orchestrated, monitored, and governed in production.

## Core Principles

- **Go-native.** GoGrid agents compile to a single static binary. No runtime dependencies. No `pip install` nightmares. Agents that just work.
- **Right-sized abstraction.** In the AI-assisted coding era, typing speed isn't the bottleneck — understanding and maintaining code is. GoGrid optimizes for clarity over cleverness, even if it means more lines of code.
- **Production-first.** Monitoring, cost tracking, error recovery, graceful degradation, and security are built into GoGrid from day one, not bolted on later.
- **Model-agnostic, vendor-agnostic.** Swap models with a config change. No subtle lock-in. Open source, forever.
- **Secure by default.** Planned [Tenuo.io](https://tenuo.io) integration for prompt injection protection, tool-use authorization, and data exfiltration prevention. *(Coming when the Tenuo Go SDK is available.)*

## Architecture

GoGrid supports five workload types — composable orchestration primitives with built-in governance:

### Single Agent
A well-scoped agent with a small number of tools. The recommended starting point for any GoGrid project.

### Team (Chat Room)
Multiple domain experts collaborating in real-time — like a meeting where participants work concurrently, debate, and reach consensus. Built on pub/sub messaging with shared memory. An optional coordinator agent can synthesize a final decision from all member responses.

### Pipeline (Linear)
Sequential handoff between specialists. Each agent completes its work, yields its state to the next agent, and terminates cleanly. Supports stage-level retry policies, input transforms, output validation, and progress reporting. Clear ownership, predictable execution.

### Graph
Like a pipeline, but with conditional edges, parallel fan-out, fan-in merging, and loops. Agents execute concurrently in waves — when a node completes, its outgoing edges are evaluated and successor nodes fire when all dependencies are satisfied. Supports configurable iteration limits, cost budgets, timeouts, and exports to Graphviz DOT format for visualization.

### Dynamic Orchestration
GoGrid's most powerful pattern. A Runtime enables agents to spawn child agents, teams, pipelines, or graphs dynamically at runtime. Resource governance controls concurrency limits, nesting depth, and cost budgets across all spawned children. Async futures allow parallel child execution with aggregate metrics tracking.

> All GoGrid patterns are composable. A team can contain pipelines. A graph node can spawn a dynamic orchestrator. The architecture adapts to the problem, not the other way around.

## Why Go?

AI agents are **infrastructure** — building them and running them should be one system.

| Requirement | Go's Answer |
|---|---|
| 50,000+ concurrent agent workflows | Goroutines — lightweight, true parallelism, no GIL |
| Network-heavy I/O (LLM calls, APIs, webhooks) | Built for large-scale networked services |
| Predictable production behavior | Predictable memory, CPU, excellent profiling |
| Simple deployment | Single static binary, tiny containers, fast startup |
| Minimal dependencies | Rich standard library (HTTP, JSON, context, testing) |
| Operational simplicity | Easy cross-compilation, DevOps teams love Go |

Python works for prototyping. Go works for production — the same way Kubernetes, Docker, Prometheus, and Terraform are all written in Go. That's why GoGrid is built on Go.

## Quick Start

Install GoGrid and scaffold a project in under a minute:

```bash
# Build the CLI
git clone https://github.com/lonestarx1/gogrid.git
cd gogrid
make build

# Scaffold a new project
bin/gogrid init --template single my-agent
cd my-agent

# Set up the project
go mod init github.com/example/my-agent
go mod tidy
export OPENAI_API_KEY=sk-proj-...

# Run your agent
gogrid run assistant -input "Explain Go's concurrency model"
```

GoGrid also supports `team` and `pipeline` templates:

```bash
gogrid init --template team my-research-team
gogrid init --template pipeline my-content-pipeline
```

## CLI

The `gogrid` CLI is the primary interface for working with GoGrid projects. Define agents in `gogrid.yaml`, run them from the command line, and inspect execution traces and costs.

### Define agents in `gogrid.yaml`

```yaml
version: "1"

agents:
  researcher:
    model: claude-sonnet-4-5-20250929
    provider: anthropic
    instructions: |
      You are a research assistant. Provide thorough analysis with
      key findings, supporting evidence, and areas for further investigation.
    config:
      max_turns: 10
      max_tokens: 4096
      temperature: 0.7
      timeout: 2m
      cost_budget: 0.50

  summarizer:
    model: gpt-4o-mini
    provider: openai
    instructions: |
      Condense the input into 3-5 bullet points. Keep it under 200 words.
    config:
      max_turns: 3
      max_tokens: 1024
      timeout: 30s
```

Config values support environment variable substitution (`${VAR}`, `${VAR:-default}`). API keys are resolved from `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, or `GEMINI_API_KEY` environment variables — never stored in config files.

### Run agents

```bash
gogrid list                                           # List defined agents
gogrid run researcher -input "Explain the CAP theorem" # Execute an agent
```

### Inspect traces and costs

Every run is recorded. Inspect what happened and what it cost:

```bash
# View execution span tree
$ gogrid trace <run-id>
agent.run (4.2s)
├── memory.load (1ms)
├── llm.complete (2.1s) [prompt: 150, completion: 89]
├── llm.complete (1.8s) [prompt: 280, completion: 145]
└── memory.save (2ms)

# View cost breakdown
$ gogrid cost <run-id>
MODEL                         CALLS  PROMPT  COMPLETION  COST
claude-sonnet-4-5-20250929    2      430     234         $0.003280
────────────────────────────────────────────────────────────────
TOTAL                         2      430     234         $0.003280
```

Both commands support `-json` for programmatic use. Run `gogrid trace` or `gogrid cost` with no arguments to list all recorded runs.

For full CLI documentation, see [docs/cli.md](docs/cli.md).

## Features

- **CLI and project scaffolding** — `gogrid init` generates working projects from templates. `gogrid run`, `gogrid trace`, and `gogrid cost` provide a complete development workflow from the command line.
- **YAML-based configuration** — Define agents declaratively with environment variable substitution. No secrets in config files.
- **Built-in observability** — Structured tracing with OTLP export (Jaeger, Tempo, Zipkin), structured JSON logging with trace correlation, and Prometheus-compatible metrics — all using the Go standard library
- **Cost governance** — Budget alerts, per-entity cost allocation, aggregate reporting, and built-in pricing for popular models. Every LLM call is metered and budgetable.
- **Shared memory** — Optimized, monitorable shared memory pool for multi-agent architectures
- **Evaluation framework** — Composable evaluators (exact match, cost budgets, tool usage, LLM-as-judge) with suites for scoring agent outputs. Performance benchmarks for all patterns.
- **Security integration** *(planned)* — [Tenuo.io](https://tenuo.io) integration for prompt injection protection, tool authorization, and data exfiltration prevention. Waiting on the Tenuo Go SDK.
- **Backward compatibility** — Gradual, backward-compatible updates. Your GoGrid agents won't break on upgrade.

## Project Status

**Early development.** We're building the foundation of GoGrid. Star the repo and watch for updates.

## Documentation

- [Getting Started](docs/getting-started.md) — Install, first agent, first CLI project
- [Patterns Guide](docs/patterns.md) — When to use each orchestration pattern
- [Testing Guide](docs/testing.md) — Mock provider, evaluation framework, benchmarks
- [Observability Guide](docs/observability.md) — Tracing, logging, metrics, cost governance
- [CLI Reference](docs/cli.md) — Full CLI command reference, config format, and environment setup
- [Manifesto](docs/manifesto.md) — Why GoGrid exists and what we believe
- [Architecture Decisions](docs/adr/README.md) — ADRs for key design choices

## Examples

All examples use the mock provider and run without API keys (`go run ./examples/<name>/`).

- [`examples/single-agent/`](examples/single-agent/) — Single agent with tool use (Go API)
- [`examples/cli-quickstart/`](examples/cli-quickstart/) — Multi-agent project using the CLI
- [`examples/multi-provider/`](examples/multi-provider/) — Same agent, different providers
- [`examples/memory-file/`](examples/memory-file/) — File-backed memory with search and pruning
- [`examples/eval/`](examples/eval/) — Evaluation suite with composable evaluators
- [`examples/team-debate/`](examples/team-debate/) — Multi-agent team with consensus and coordinator
- [`examples/pipeline-research/`](examples/pipeline-research/) — Three-stage pipeline with state transfer
- [`examples/graph-review/`](examples/graph-review/) — Graph with conditional edges and review loop
- [`examples/dynamic-spawn/`](examples/dynamic-spawn/) — Dynamic orchestration with resource governance
- [`examples/observability/`](examples/observability/) — Full observability stack (tracing, logging, metrics, cost)

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.

## Contributing

TBD — Contribution guidelines coming soon. Visit [gogrid.org](https://gogrid.org) for updates.
