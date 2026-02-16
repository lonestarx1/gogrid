# ADR-004: No Vendor Lock-In

## Status

Accepted

## Date

February 2026

## Context

AI agent frameworks often have deep coupling to a single LLM provider. This creates risk: pricing changes, API deprecations, or model quality regressions force expensive rewrites. Users should be able to swap providers with a configuration change.

## Decision

GoGrid is model-agnostic and provider-agnostic through a clean provider interface:

```go
type Provider interface {
    Complete(ctx context.Context, params Params) (*Response, error)
}
```

All LLM interactions go through this interface. GoGrid ships three built-in providers:

1. **`llm/openai`** — OpenAI (GPT-4o, GPT-4.1, o3, etc.)
2. **`llm/anthropic`** — Anthropic (Claude Sonnet, Claude Opus, etc.)
3. **`llm/gemini`** — Google (Gemini 2.5 Pro, Gemini 2.5 Flash, etc.)
4. **`llm/mock`** — Mock provider for testing (no API keys needed)

Key design decisions:

- **Unified message format.** GoGrid defines its own `llm.Message` type. Providers translate to/from vendor-specific formats internally.
- **Unified tool calling.** `llm.ToolCall` and `llm.ToolResult` abstract over vendor differences in function calling.
- **No provider-specific features in the core.** If a provider offers a unique capability, it stays in that provider's package.
- **CLI resolves providers from config.** The `gogrid.yaml` `provider` field determines which implementation is used. Swapping is a one-line YAML change.

## Consequences

- **Lowest common denominator.** The unified interface cannot expose every provider-specific feature. Advanced users can access provider-specific packages directly.
- **Translation overhead.** Each provider must translate between GoGrid's message format and its own. This is minimal but adds maintenance cost when providers change their APIs.
- **Testing without API keys.** The mock provider enables complete test coverage and runnable examples without any external dependencies.
- **Easy migration.** Users can switch providers by changing one line of code or one YAML field. Multi-provider setups (e.g., different models for different agents) work naturally.
