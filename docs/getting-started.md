# Getting Started with GoGrid

This guide gets you from zero to a running agent in under five minutes.

## Prerequisites

- Go 1.22 or later
- Git

## Installation

```bash
go get github.com/lonestarx1/gogrid
```

Or clone the repository:

```bash
git clone https://github.com/lonestarx1/gogrid.git
cd gogrid
```

## Your First Agent (Go API)

Create a simple agent using the mock provider — no API keys needed:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/lonestarx1/gogrid/pkg/agent"
    "github.com/lonestarx1/gogrid/pkg/llm"
    "github.com/lonestarx1/gogrid/pkg/llm/mock"
)

func main() {
    provider := mock.New(mock.WithFallback(&llm.Response{
        Message: llm.NewAssistantMessage("Hello from GoGrid!"),
        Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
        Model:   "mock",
    }))

    a := agent.New("my-agent",
        agent.WithProvider(provider),
        agent.WithModel("mock"),
        agent.WithInstructions("You are a helpful assistant."),
    )

    result, err := a.Run(context.Background(), "Say hello")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.Message.Content)
}
```

For a real LLM provider, swap `mock.New(...)` with `openai.New(os.Getenv("OPENAI_API_KEY"))` and set the appropriate model.

## Your First CLI Project

Build the CLI and scaffold a project:

```bash
make build
bin/gogrid init --template single my-agent
cd my-agent
```

Edit `gogrid.yaml` to configure your agent, then run it:

```bash
gogrid run assistant -input "What is Go?"
```

See [docs/cli.md](cli.md) for the full CLI reference.

## Using Real LLM Providers

GoGrid supports three built-in providers. Set the appropriate environment variable:

| Provider | Package | Environment Variable |
|----------|---------|---------------------|
| OpenAI | `pkg/llm/openai` | `OPENAI_API_KEY` |
| Anthropic | `pkg/llm/anthropic` | `ANTHROPIC_API_KEY` |
| Google Gemini | `pkg/llm/gemini` | `GOOGLE_API_KEY` |

## Runnable Examples

Every example uses the mock provider and runs without API keys:

| Example | Command |
|---------|---------|
| Single agent with tools | `go run ./examples/single-agent/` |
| Provider swapping | `go run ./examples/multi-provider/` |
| File memory + pruning | `go run ./examples/memory-file/` |
| Evaluation framework | `go run ./examples/eval/` |
| Team debate | `go run ./examples/team-debate/` |
| Pipeline with state transfer | `go run ./examples/pipeline-research/` |
| Graph with review loop | `go run ./examples/graph-review/` |
| Dynamic orchestration | `go run ./examples/dynamic-spawn/` |
| Full observability stack | `go run ./examples/observability/` |

## What's Next

- **[Patterns Guide](patterns.md)** — When to use each orchestration pattern
- **[Testing Guide](testing.md)** — Mock provider, evaluation, benchmarks
- **[Observability Guide](observability.md)** — Tracing, logging, metrics, cost governance
- **[CLI Reference](cli.md)** — Full command reference and YAML config format
- **[Website](https://gogrid.org/docs)** — Comprehensive API documentation
- **[Architecture Decisions](adr/README.md)** — Why GoGrid is built the way it is
