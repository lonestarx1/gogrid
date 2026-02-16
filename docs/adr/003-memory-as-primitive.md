# ADR-003: Memory as a First-Class Primitive

## Status

Accepted

## Date

February 2026

## Context

Most agent frameworks treat memory as an optional plugin — a vector store you can bolt on after the fact. This leads to agents that lose context, teams that cannot share state, and pipelines where data ownership is ambiguous.

GoGrid's orchestration patterns have distinct memory requirements:

- **Single agents** need conversation persistence across runs.
- **Teams** need shared memory with change notifications so agents can observe each other's state.
- **Pipelines** need state ownership transfer — one stage owns the state at a time, and ownership transfers cleanly to the next stage.

## Decision

Memory is a core primitive in GoGrid, not a plugin. The `pkg/memory/` package defines interfaces and provides multiple implementations:

1. **`memory.Memory`** — base interface with Load, Save, Clear operations.
2. **`memory.SearchableMemory`** — extends Memory with substring search across entries.
3. **`memory.PrunableMemory`** — extends Memory with policy-based pruning (by age, by count, by size).
4. **`memory.StatsMemory`** — extends Memory with aggregate statistics.
5. **`memory/file.Memory`** — file-backed persistence with hex-encoded filenames and JSON metadata sidecars.
6. **`memory/shared.Memory`** — thread-safe shared memory pool with pub/sub change notifications for team architectures.
7. **`memory/transfer.State`** — generation-based ownership transfer with audit trails for pipeline architectures.

State ownership is enforced at the type level: `transfer.Handle` methods fail if ownership has been transferred away (generation mismatch).

## Consequences

- **Always available.** Every agent can use memory without additional dependencies or configuration. File memory works out of the box.
- **Pattern-specific implementations.** Teams get shared memory with change events. Pipelines get transferable state with audit trails. This avoids one-size-fits-all compromises.
- **Interface hierarchy.** The SearchableMemory/PrunableMemory/StatsMemory interfaces add complexity but allow type assertions for optional capabilities.
- **No external dependencies.** All memory implementations use the Go standard library. No Redis, no vector databases. Users can implement custom Memory backends for those use cases.
