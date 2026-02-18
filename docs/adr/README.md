# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for GoGrid.

ADRs capture significant architectural decisions along with their context and consequences. They are numbered sequentially and immutable once accepted — superseded decisions get a new ADR that references the old one.

## Index

| ADR | Title | Status |
|-----|-------|--------|
| [001](001-go-not-python.md) | Why Go for an AI agent system | Accepted |
| [002](002-five-orchestration-patterns.md) | Five orchestration patterns | Accepted |
| [003](003-memory-as-primitive.md) | Memory as a first-class primitive | Accepted |
| [004](004-no-vendor-lock-in.md) | No vendor lock-in | Accepted |
| [005](005-eval-framework-design.md) | Evaluation framework design | Accepted |

## Format

Each ADR follows this structure:

- **Title** — short descriptive name
- **Status** — Proposed, Accepted, Deprecated, or Superseded
- **Date** — when the decision was made
- **Context** — what prompted the decision
- **Decision** — what was decided and why
- **Consequences** — trade-offs and implications
