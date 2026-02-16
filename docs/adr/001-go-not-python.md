# ADR-001: Why Go for an AI Agent Framework

## Status

Accepted

## Date

February 2026

## Context

The AI agent ecosystem is dominated by Python frameworks (LangChain, CrewAI, AutoGen). Starting a new framework in Go requires justification since Python has more LLM SDKs, a larger ML community, and faster prototyping velocity.

GoGrid targets production deployments, not notebook prototyping. The primary users are infrastructure engineers building agent systems that must run reliably at scale.

## Decision

GoGrid is written in Go. The core reasons:

1. **Concurrency model.** Goroutines handle 50,000+ concurrent agent workflows without the GIL bottleneck. Agent orchestration is inherently concurrent — teams run agents in parallel, graphs execute waves of nodes, dynamic runtimes spawn children concurrently.

2. **Deployment simplicity.** A single static binary with no runtime dependencies. No virtualenvs, no dependency conflicts, no `pip install` failures in CI. Agents deploy as containers with minimal images.

3. **Predictable performance.** Go's garbage collector has sub-millisecond pauses. Memory usage is predictable and profileable. This matters when running cost-governed agents where budget overruns are expensive.

4. **Operational fit.** The target audience (infrastructure engineers) already knows Go. Kubernetes, Docker, Prometheus, Terraform, and Vault are all Go. GoGrid fits the same toolchain.

5. **Type safety.** Compile-time type checking catches configuration errors before runtime. Agent interfaces, tool schemas, and orchestration patterns benefit from static typing.

## Consequences

- **Smaller LLM SDK ecosystem.** Go has fewer LLM client libraries than Python. GoGrid maintains its own provider implementations for OpenAI, Anthropic, and Gemini.
- **Longer initial development.** Go requires more boilerplate than Python for the same functionality. GoGrid accepts this trade-off because clarity over cleverness is a core principle.
- **Community adoption barrier.** Many AI practitioners are Python-first. GoGrid targets a different audience — infrastructure teams who deploy agents, not researchers who prototype them.
- **No notebook support.** Go does not run in Jupyter notebooks. GoGrid provides a CLI and programmatic API instead.
