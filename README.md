# GoGrid (G2) — Kubernetes for AI Agents

> **[gogrid.org](https://gogrid.org)**
> A production-grade AI agent framework written in Go.
> Built for infrastructure engineers, not notebook demos.

---

## Why Another Agent Framework?

Every existing agent framework gets something right — and gets something fundamentally wrong.

LangChain has the ecosystem but drowns you in abstraction. CrewAI has the mental model but falls apart in production. AutoGen pioneered multi-agent conversations but got rewritten twice. Most are Python-only, prototype-first, and treat production as an afterthought.

**80-90% of AI agent projects never leave the pilot phase.** (RAND, 2025)

GoGrid is the framework for the other side of that gap.

GoGrid is not a wrapper around LLM APIs. It is infrastructure for running agents at scale — the way Kubernetes is infrastructure for running containers at scale.

## Core Principles

- **Go-native.** GoGrid agents compile to a single static binary. No runtime dependencies. No `pip install` nightmares. Agents that just work.
- **Right-sized abstraction.** In the AI-assisted coding era, typing speed isn't the bottleneck — understanding and maintaining code is. GoGrid optimizes for clarity over cleverness, even if it means more lines of code.
- **Production-first.** Monitoring, cost tracking, error recovery, graceful degradation, and security are built into GoGrid from day one, not bolted on later.
- **Model-agnostic, vendor-agnostic.** Swap models with a config change. No subtle lock-in. Open source, forever.
- **Secure by default.** Integrated security posture with [Tenio.ai](https://tenio.ai). Prompt injection protection, tool-use authorization, and data exfiltration prevention as first-class concerns.

## Architecture

GoGrid supports five orchestration patterns, because different problems demand different architectures:

### Single Agent
A well-scoped agent with a small number of tools. The recommended starting point for any GoGrid project.

### Team (Chat Room)
Multiple domain experts collaborating in real-time — like a meeting where participants work concurrently, debate, and reach consensus. Built on pub/sub messaging with shared memory.

### Pipeline (Linear)
Sequential handoff between specialists. Each agent completes its work, yields its state to the next agent, and terminates cleanly. Clear ownership, predictable execution.

### Graph
Like a pipeline, but with the ability to loop back (re-do, clarify) or branch into parallel paths that merge later. Best for workflows with a bounded number of agents where you want to visualize the data flow.

### Dynamic Orchestration
GoGrid's most powerful pattern. Agents can spawn child agents, child teams, child pipelines, or child graphs dynamically at runtime. Unlimited scaling with minimal assumptions about how a problem gets solved. For when the developer doesn't know — or shouldn't hardcode — the exact steps to a solution.

> All GoGrid patterns are composable. A team can contain pipelines. A graph node can spawn a dynamic orchestrator. The architecture adapts to the problem, not the other way around.

## Why Go?

AI agents are **infrastructure**, not scripts.

| Requirement | Go's Answer |
|---|---|
| 50,000+ concurrent agent workflows | Goroutines — lightweight, true parallelism, no GIL |
| Network-heavy I/O (LLM calls, APIs, webhooks) | Built for large-scale networked services |
| Predictable production behavior | Predictable memory, CPU, excellent profiling |
| Simple deployment | Single static binary, tiny containers, fast startup |
| Minimal dependencies | Rich standard library (HTTP, JSON, context, testing) |
| Operational simplicity | Easy cross-compilation, DevOps teams love Go |

Python works for prototyping. Go works for production — the same way Kubernetes, Docker, Prometheus, and Terraform are all written in Go. That's why GoGrid is built on Go.

## Features

- **Built-in observability** — Structured tracing across GoGrid agent invocations with opt-in log files
- **Cost governance** — LLM API cost monitoring and budgeting as a first-class GoGrid primitive
- **Shared memory** — Optimized, monitorable shared memory pool for multi-agent architectures
- **Evaluation framework** — Built-in metrics for agent performance and scalability
- **Security integration** — Secure-by-default design with Tenio.ai integration
- **Backward compatibility** — Gradual, backward-compatible updates. Your GoGrid agents won't break on upgrade.

## Project Status

**Early development.** We're building the foundation of GoGrid. Star the repo and watch for updates.

## Documentation

- [Manifesto](docs/manifesto.md) — Why GoGrid exists and what we believe

## License

TBD

## Contributing

TBD — Contribution guidelines coming soon. Visit [gogrid.org](https://gogrid.org) for updates.
