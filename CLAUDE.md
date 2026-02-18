# CLAUDE.md — Project Instructions for Claude Code

## Project Overview

**GoGrid (G2)** — A unified system for developing and orchestrating production AI agents in Go. GoGrid fuses agent development with agent orchestration the way Kubernetes fused container packaging with container management.

**Project name:** GoGrid (shorthand: G2)
**Domain:** [gogrid.org](https://gogrid.org)
**Primary language:** Go
**Status:** Early development / foundational phase

## Core Philosophy

- **Clarity over cleverness.** Explicit, readable code. More lines of clear code is preferred over fewer lines of magic. Follow Go idioms and conventions.
- **Production-first.** Every GoGrid feature must consider monitoring, error handling, cost management, security, and scalability from the start.
- **Secure by default.** Never introduce code that is insecure by default. Security integration with Tenuo.io.
- **No vendor lock-in.** Model-agnostic, provider-agnostic. Swapping LLM providers must be a configuration change.
- **Backward compatibility.** API changes must be gradual and backward-compatible. Deprecate before removing.

## Architecture

GoGrid supports 5 orchestration patterns. All patterns are composable:

1. **Single Agent** — One agent, small tool set, well-defined scope
2. **Team (Chat Room)** — Pub/sub messaging, shared memory pool, concurrent execution, consensus-reaching
3. **Pipeline (Linear)** — Sequential handoff with state ownership transfer, clean termination
4. **Graph** — Like pipeline but with loops (re-do/clarify) and parallel branches that merge
5. **Dynamic Orchestration** — Agents can spawn child agents/teams/pipelines/graphs at runtime

## Go Conventions

- Follow standard Go project layout (`cmd/`, `internal/`, `pkg/`)
- Use `context.Context` for cancellation, timeouts, and request-scoped values
- Errors are values — return errors, don't panic. Use `fmt.Errorf` with `%w` for wrapping.
- Prefer interfaces at consumption site, not at definition site
- Use table-driven tests
- Run `go fmt`, `go vet`, and `golangci-lint` before committing
- No external dependencies unless clearly justified. Prefer the standard library.
- Exported types and functions must have doc comments

## File Organization

```
/
├── CLAUDE.md              # This file
├── README.md              # GoGrid project overview
├── docs/
│   └── manifesto.md       # The GoGrid Manifesto
├── cmd/                   # Binary entry points (future)
├── internal/              # Private packages (future)
└── pkg/                   # Public API packages (future)
```

## Key Design Decisions

- **Memory is a core primitive**, not a plugin. Design all GoGrid agent patterns with memory as a first-class concern.
- **Observability is built-in.** Structured tracing across agent invocations. Opt-in log files.
- **Cost governance is built-in.** Every LLM API call must be metered and budgetable.
- **State ownership matters.** In GoGrid's pipeline architecture, state transfers — it is not shared. Separate data ownership from authority to prevent confused deputy problems.
- **Shared memory for teams.** GoGrid's team/chat-room architectures require a shared memory pool or shared state.

## Communication Style

- When writing commit messages, docs, or comments: clear, direct, no fluff
- Use concrete examples over abstract descriptions
- Reference the GoGrid manifesto when making architectural decisions
- Refer to the project as "GoGrid" or "G2" — never as "the framework" in isolation. Use "system" when a generic noun is needed.

## What NOT to Do

- Do not add Python or TypeScript code to this project
- Do not introduce model-specific or vendor-specific hard dependencies
- Do not add abstractions that hide what the underlying LLM call is doing
- Do not break backward compatibility without a deprecation cycle
- Do not add dependencies where the standard library suffices
- **Do not mention Claude, AI assistants, or co-authored-by lines in commits, PRs, or any git history.** Commits and PRs should read as if written by a human contributor.
