"use client";

import { motion } from "framer-motion";
import CodeBlock from "@/components/CodeBlock";

const examples = [
  {
    id: "single-agent",
    title: "Single Agent with Tools",
    desc: "A calculator agent that uses tool calling to evaluate math expressions.",
  },
  {
    id: "multi-provider",
    title: "Multi-Provider Swap",
    desc: "Same agent, different LLM providers — swap with one line.",
  },
  {
    id: "memory-persistence",
    title: "Persistent Memory",
    desc: "File-backed memory that survives process restarts.",
  },
  {
    id: "code-review-team",
    title: "Code Review Team",
    desc: "Three specialized agents review code concurrently.",
  },
  {
    id: "debate-team",
    title: "Multi-Round Debate",
    desc: "Agents debate a topic over multiple rounds, seeing each other's responses.",
  },
  {
    id: "coordinated-team",
    title: "Coordinated Team",
    desc: "A leader agent synthesizes team responses into a final decision.",
  },
  {
    id: "review-graph",
    title: "Review Loop Graph",
    desc: "Draft, review, revise loop with conditional edges and DOT export.",
  },
  {
    id: "research-pipeline",
    title: "Research Pipeline",
    desc: "Three-stage pipeline: research, analyze, summarize — with retry and progress.",
  },
  {
    id: "pipeline-state",
    title: "Pipeline State Transfer",
    desc: "State ownership enforced across pipeline stages with audit trail.",
  },
  {
    id: "dynamic-research",
    title: "Dynamic Research Coordinator",
    desc: "An agent dynamically spawns teams, pipelines, and sub-agents at runtime.",
  },
  {
    id: "observability",
    title: "Full Observability Stack",
    desc: "OTLP export, structured logging, and Prometheus metrics — all wired together.",
  },
  {
    id: "cli-multi-agent",
    title: "CLI Multi-Agent Project",
    desc: "Define and run multiple agents from YAML — no Go code required.",
  },
  {
    id: "cli-trace-cost",
    title: "CLI Trace & Cost Inspection",
    desc: "Inspect execution traces and cost breakdowns from the command line.",
  },
  {
    id: "cli-scaffold",
    title: "Project Scaffolding",
    desc: "Scaffold a complete GoGrid project from a template in one command.",
  },
  {
    id: "eval-suite",
    title: "Evaluation Suite",
    desc: "Score agent outputs with composable evaluators — exact match, cost budgets, and custom checks.",
  },
  {
    id: "mock-testing",
    title: "Mock Provider Testing",
    desc: "Test agents with sequential responses, error injection, and call recording.",
  },
];

export default function ExamplesPage() {
  return (
    <div className="pt-14 min-h-screen">
      <main className="max-w-4xl mx-auto px-6 py-12">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
        >
          <h1 className="font-mono text-4xl md:text-5xl font-bold text-white mb-4">
            Examples
          </h1>
          <p className="text-text-muted text-lg mb-6 max-w-2xl">
            Complete, runnable examples showing GoGrid patterns in action.
          </p>

          {/* Quick nav */}
          <div className="flex flex-wrap gap-2 mb-12">
            {examples.map((ex) => (
              <a
                key={ex.id}
                href={`#${ex.id}`}
                className="font-mono text-xs px-3 py-1.5 rounded border border-border text-text-muted hover:text-accent hover:border-accent transition-colors"
              >
                {ex.title}
              </a>
            ))}
          </div>

          {/* Single Agent with Tools */}
          <ExampleSection
            id="single-agent"
            title="Single Agent with Tools"
            desc="A calculator agent that uses tool calling to evaluate math expressions. Demonstrates agent creation, tool definition, and the iterative tool-use loop."
          >
            <CodeBlock
              filename="examples/calculator/main.go"
              code={`package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strconv"

    "github.com/lonestarx1/gogrid/pkg/agent"
    "github.com/lonestarx1/gogrid/pkg/llm/openai"
    "github.com/lonestarx1/gogrid/pkg/memory"
    "github.com/lonestarx1/gogrid/pkg/tool"
    "github.com/lonestarx1/gogrid/pkg/trace"
)

// CalcTool evaluates simple math expressions.
type CalcTool struct{}

func (c *CalcTool) Name() string        { return "calculate" }
func (c *CalcTool) Description() string { return "Evaluate a math expression" }
func (c *CalcTool) Schema() tool.Schema {
    return tool.Schema{
        Type: "object",
        Properties: map[string]*tool.Schema{
            "a":  {Type: "number", Description: "First operand"},
            "b":  {Type: "number", Description: "Second operand"},
            "op": {Type: "string", Description: "Operator: +, -, *, /"},
        },
        Required: []string{"a", "b", "op"},
    }
}

func (c *CalcTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
    var args struct {
        A  float64 \`json:"a"\`
        B  float64 \`json:"b"\`
        Op string  \`json:"op"\`
    }
    if err := json.Unmarshal(input, &args); err != nil {
        return "", err
    }
    var result float64
    switch args.Op {
    case "+":
        result = args.A + args.B
    case "-":
        result = args.A - args.B
    case "*":
        result = args.A * args.B
    case "/":
        if args.B == 0 {
            return "error: division by zero", nil
        }
        result = args.A / args.B
    default:
        return "error: unknown operator " + args.Op, nil
    }
    return strconv.FormatFloat(result, 'f', -1, 64), nil
}

func main() {
    ctx := context.Background()
    provider := openai.New(os.Getenv("OPENAI_API_KEY"))

    a := agent.New("calculator",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("You are a calculator. Use the calculate tool for math."),
        agent.WithTools(&CalcTool{}),
        agent.WithMemory(memory.NewInMemory()),
        agent.WithTracer(trace.NewStdout(os.Stdout)),
        agent.WithConfig(agent.Config{MaxTurns: 5}),
    )

    result, err := a.Run(ctx, "What is 42 * 17 + 3?")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Answer: %s\\n", result.Message.Content)
    fmt.Printf("Turns: %d | Cost: $%.6f | Tokens: %d\\n",
        result.Turns, result.Cost, result.Usage.TotalTokens)
}`}
            />
          </ExampleSection>

          {/* Multi-Provider Swap */}
          <ExampleSection
            id="multi-provider"
            title="Multi-Provider Swap"
            desc="The same agent logic works with any LLM provider. Swap the provider and model — everything else stays the same."
          >
            <CodeBlock
              filename="examples/multi-provider/main.go"
              code={`package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lonestarx1/gogrid/pkg/agent"
    "github.com/lonestarx1/gogrid/pkg/llm"
    "github.com/lonestarx1/gogrid/pkg/llm/anthropic"
    "github.com/lonestarx1/gogrid/pkg/llm/openai"
)

func runWith(ctx context.Context, name string, provider llm.Provider, model string) {
    a := agent.New(name,
        agent.WithProvider(provider),
        agent.WithModel(model),
        agent.WithInstructions("Answer concisely in one sentence."),
    )

    result, err := a.Run(ctx, "What is the capital of France?")
    if err != nil {
        log.Printf("[%s] error: %v", name, err)
        return
    }
    fmt.Printf("[%s] %s (cost: $%.6f)\\n", name, result.Message.Content, result.Cost)
}

func main() {
    ctx := context.Background()

    // OpenAI
    oai := openai.New(os.Getenv("OPENAI_API_KEY"))
    runWith(ctx, "openai", oai, "gpt-4o")

    // Anthropic
    ant := anthropic.New(os.Getenv("ANTHROPIC_API_KEY"))
    runWith(ctx, "anthropic", ant, "claude-sonnet-4-5-20250929")
}`}
            />
          </ExampleSection>

          {/* Persistent Memory */}
          <ExampleSection
            id="memory-persistence"
            title="Persistent Memory"
            desc="File-backed memory that persists across process restarts. The agent remembers previous conversations."
          >
            <CodeBlock
              filename="examples/persistent-memory/main.go"
              code={`package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lonestarx1/gogrid/pkg/agent"
    "github.com/lonestarx1/gogrid/pkg/llm/openai"
    "github.com/lonestarx1/gogrid/pkg/memory/file"
)

func main() {
    ctx := context.Background()

    // File memory persists to disk — survives restarts
    mem, err := file.New("./agent-data")
    if err != nil {
        log.Fatal(err)
    }

    provider := openai.New(os.Getenv("OPENAI_API_KEY"))

    a := agent.New("assistant",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("You are a helpful assistant with persistent memory."),
        agent.WithMemory(mem),
    )

    // First run: tell it something
    result, _ := a.Run(ctx, "My name is Alice and I like Go programming.")
    fmt.Println(result.Message.Content)

    // Second run (even after restart): it remembers
    result, _ = a.Run(ctx, "What's my name and what do I like?")
    fmt.Println(result.Message.Content)

    // Check memory stats
    stats, _ := mem.Stats(ctx)
    fmt.Printf("Memory: %d keys, %d entries, %d bytes\\n",
        stats.Keys, stats.TotalEntries, stats.TotalSize)
}`}
            />
          </ExampleSection>

          {/* Code Review Team */}
          <ExampleSection
            id="code-review-team"
            title="Code Review Team"
            desc="Three agents review code concurrently — one for correctness, one for security, one for performance. All responses are combined using the Unanimous strategy."
          >
            <CodeBlock
              filename="examples/code-review/main.go"
              code={`package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lonestarx1/gogrid/pkg/agent"
    "github.com/lonestarx1/gogrid/pkg/llm/openai"
    "github.com/lonestarx1/gogrid/pkg/orchestrator/team"
    "github.com/lonestarx1/gogrid/pkg/trace"
)

func main() {
    ctx := context.Background()
    provider := openai.New(os.Getenv("OPENAI_API_KEY"))
    tracer := trace.NewStdout(os.Stdout)

    reviewer := agent.New("reviewer",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("Review the code for correctness, readability, and Go idioms."),
        agent.WithTracer(tracer),
    )

    security := agent.New("security",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("Review the code for security vulnerabilities (injection, auth, data exposure)."),
        agent.WithTracer(tracer),
    )

    perf := agent.New("performance",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("Review the code for performance issues (allocations, complexity, concurrency)."),
        agent.WithTracer(tracer),
    )

    t := team.New("code-review",
        team.WithMembers(
            team.Member{Agent: reviewer, Role: "correctness"},
            team.Member{Agent: security, Role: "security"},
            team.Member{Agent: perf, Role: "performance"},
        ),
        team.WithStrategy(team.Unanimous{}),
        team.WithTracer(tracer),
        team.WithConfig(team.Config{CostBudget: 2.00}),
    )

    code := \`func handleLogin(w http.ResponseWriter, r *http.Request) {
    username := r.FormValue("username")
    password := r.FormValue("password")
    query := fmt.Sprintf("SELECT * FROM users WHERE name='%s' AND pass='%s'", username, password)
    rows, _ := db.Query(query)
    // ...
}\`

    result, err := t.Run(ctx, "Review this Go code:\\n\\n"+code)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("=== Team Decision ===")
    fmt.Println(result.Decision.Content)
    fmt.Printf("\\nCost: $%.4f | Rounds: %d\\n", result.TotalCost, result.Rounds)

    // Individual agent responses
    for name, r := range result.Responses {
        fmt.Printf("\\n--- %s ---\\n%s\\n", name, r.Message.Content)
    }
}`}
            />
          </ExampleSection>

          {/* Multi-Round Debate */}
          <ExampleSection
            id="debate-team"
            title="Multi-Round Debate"
            desc="Two agents debate a topic over multiple rounds. Each round, agents see the other's previous response and refine their position."
          >
            <CodeBlock
              filename="examples/debate/main.go"
              code={`package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lonestarx1/gogrid/pkg/agent"
    "github.com/lonestarx1/gogrid/pkg/llm/openai"
    "github.com/lonestarx1/gogrid/pkg/orchestrator/team"
)

func main() {
    ctx := context.Background()
    provider := openai.New(os.Getenv("OPENAI_API_KEY"))

    advocate := agent.New("advocate",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("You argue IN FAVOR of the topic. Be concise but persuasive."),
    )

    skeptic := agent.New("skeptic",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("You argue AGAINST the topic. Be concise but rigorous."),
    )

    // Subscribe to the bus to watch the debate in real time
    bus := team.NewBus()
    ch, unsub := bus.Subscribe("team.response", 20)
    defer unsub()

    go func() {
        for msg := range ch {
            fmt.Printf("  [Round %s] %s: %s\\n\\n",
                msg.Metadata["round"], msg.From, msg.Content[:min(len(msg.Content), 100)]+"...")
        }
    }()

    t := team.New("debate",
        team.WithMembers(
            team.Member{Agent: advocate, Role: "advocate"},
            team.Member{Agent: skeptic, Role: "skeptic"},
        ),
        team.WithBus(bus),
        team.WithStrategy(team.Unanimous{}),
        team.WithConfig(team.Config{MaxRounds: 3}),
    )

    result, err := t.Run(ctx, "Should companies adopt AI agents in production?")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("\\n=== Final Decision ===")
    fmt.Println(result.Decision.Content)
    fmt.Printf("Rounds: %d | Cost: $%.4f\\n", result.Rounds, result.TotalCost)
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}`}
            />
          </ExampleSection>

          {/* Coordinated Team */}
          <ExampleSection
            id="coordinated-team"
            title="Coordinated Team"
            desc="A coordinator agent listens to all team members and synthesizes a single, coherent decision — instead of concatenating responses."
          >
            <CodeBlock
              filename="examples/coordinated-team/main.go"
              code={`package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lonestarx1/gogrid/pkg/agent"
    "github.com/lonestarx1/gogrid/pkg/llm/openai"
    "github.com/lonestarx1/gogrid/pkg/orchestrator/team"
)

func main() {
    ctx := context.Background()
    provider := openai.New(os.Getenv("OPENAI_API_KEY"))

    // Domain experts
    frontend := agent.New("frontend",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("You are a frontend engineer. Evaluate proposals from a UI/UX perspective."),
    )
    backend := agent.New("backend",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("You are a backend engineer. Evaluate proposals from an architecture perspective."),
    )
    security := agent.New("security",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("You are a security engineer. Evaluate proposals for security implications."),
    )

    // Team lead synthesizes all perspectives
    lead := agent.New("tech-lead",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("You are the tech lead. Synthesize all perspectives into a clear, actionable decision."),
    )

    t := team.New("design-review",
        team.WithMembers(
            team.Member{Agent: frontend, Role: "frontend"},
            team.Member{Agent: backend, Role: "backend"},
            team.Member{Agent: security, Role: "security"},
        ),
        team.WithCoordinator(lead),
        team.WithStrategy(team.Unanimous{}),
        team.WithConfig(team.Config{MaxRounds: 2}),
    )

    result, err := t.Run(ctx, "Should we add WebSocket support for real-time notifications?")
    if err != nil {
        log.Fatal(err)
    }

    // The decision comes from the coordinator, not concatenated responses
    fmt.Println("=== Tech Lead Decision ===")
    fmt.Println(result.Decision.Content)

    fmt.Printf("\\nRounds: %d | Cost: $%.4f\\n", result.Rounds, result.TotalCost)

    // Individual perspectives are still available
    for name, r := range result.Responses {
        fmt.Printf("\\n--- %s ---\\n%s\\n", name, r.Message.Content[:min(len(r.Message.Content), 100)])
    }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}`}
            />
          </ExampleSection>

          {/* Review Loop Graph */}
          <ExampleSection
            id="review-graph"
            title="Review Loop Graph"
            desc="A graph with conditional edges: draft is reviewed, and if it needs revision, it loops back through revise → review until approved, then publishes."
          >
            <CodeBlock
              filename="examples/review-graph/main.go"
              code={`package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "strings"
    "time"

    "github.com/lonestarx1/gogrid/pkg/agent"
    "github.com/lonestarx1/gogrid/pkg/llm/openai"
    "github.com/lonestarx1/gogrid/pkg/orchestrator/graph"
    "github.com/lonestarx1/gogrid/pkg/trace"
)

func main() {
    ctx := context.Background()
    provider := openai.New(os.Getenv("OPENAI_API_KEY"))
    tracer := trace.NewStdout(os.Stdout)

    draft := agent.New("draft",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("Write a short blog post about the given topic."),
        agent.WithTracer(tracer),
    )
    review := agent.New("review",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("Review this draft. If it needs work, say 'needs revision' and explain. If good, say 'approved'."),
        agent.WithTracer(tracer),
    )
    revise := agent.New("revise",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("Revise the draft based on the review feedback."),
        agent.WithTracer(tracer),
    )
    publish := agent.New("publish",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("Format the approved draft for publication. Add a title and summary."),
        agent.WithTracer(tracer),
    )

    g, err := graph.NewBuilder("review-pipeline").
        AddNode("draft", draft).
        AddNode("review", review).
        AddNode("revise", revise).
        AddNode("publish", publish).
        AddEdge("draft", "review").
        AddEdge("review", "revise", graph.When(func(out string) bool {
            return strings.Contains(strings.ToLower(out), "needs revision")
        })).
        AddEdge("review", "publish", graph.When(func(out string) bool {
            return strings.Contains(strings.ToLower(out), "approved")
        })).
        AddEdge("revise", "review").
        Options(
            graph.WithConfig(graph.Config{
                MaxIterations: 3,
                Timeout:       2 * time.Minute,
                CostBudget:    2.00,
            }),
            graph.WithTracer(tracer),
        ).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    // Export DOT for visualization
    fmt.Println("=== Graph Structure (DOT) ===")
    fmt.Println(g.DOT())

    result, err := g.Run(ctx, "The future of AI agents in production systems")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("=== Published Output ===")
    fmt.Println(result.Output)
    fmt.Printf("\\nCost: $%.4f | Tokens: %d\\n", result.TotalCost, result.TotalUsage.TotalTokens)

    // Show iteration counts
    for name, results := range result.NodeResults {
        fmt.Printf("  %s: %d iteration(s)\\n", name, len(results))
    }
}`}
            />
          </ExampleSection>

          {/* Research Pipeline */}
          <ExampleSection
            id="research-pipeline"
            title="Research Pipeline"
            desc="A three-stage pipeline with retry, input transforms, and progress reporting. Each stage processes output from the previous one."
          >
            <CodeBlock
              filename="examples/research-pipeline/main.go"
              code={`package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/lonestarx1/gogrid/pkg/agent"
    "github.com/lonestarx1/gogrid/pkg/llm/openai"
    "github.com/lonestarx1/gogrid/pkg/orchestrator/pipeline"
    "github.com/lonestarx1/gogrid/pkg/trace"
)

func main() {
    ctx := context.Background()
    provider := openai.New(os.Getenv("OPENAI_API_KEY"))
    tracer := trace.NewStdout(os.Stdout)

    researcher := agent.New("researcher",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("You are a research assistant. Find key facts about the topic."),
        agent.WithTracer(tracer),
    )

    analyst := agent.New("analyst",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("You are a data analyst. Identify patterns and insights."),
        agent.WithTracer(tracer),
    )

    writer := agent.New("writer",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("You are a technical writer. Produce a clear, concise summary."),
        agent.WithTracer(tracer),
    )

    p := pipeline.New("research",
        pipeline.WithStages(
            pipeline.Stage{
                Name:  "research",
                Agent: researcher,
                Retry: pipeline.RetryPolicy{MaxAttempts: 2, Delay: time.Second},
            },
            pipeline.Stage{
                Name:  "analyze",
                Agent: analyst,
                InputTransform: func(input string) string {
                    return "Analyze the following research:\\n\\n" + input
                },
            },
            pipeline.Stage{
                Name:  "summarize",
                Agent: writer,
                InputTransform: func(input string) string {
                    return "Summarize these findings in 3 bullet points:\\n\\n" + input
                },
            },
        ),
        pipeline.WithTracer(tracer),
        pipeline.WithConfig(pipeline.Config{
            Timeout:    2 * time.Minute,
            CostBudget: 1.00,
        }),
        pipeline.WithProgress(func(idx, total int, sr pipeline.StageResult) {
            fmt.Printf("[%d/%d] Stage %q completed\\n", idx+1, total, sr.Name)
        }),
    )

    result, err := p.Run(ctx, "Impact of large language models on software development")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("\\n=== Final Output ===")
    fmt.Println(result.Output)
    fmt.Printf("\\nStages: %d | Cost: $%.4f | Tokens: %d\\n",
        len(result.Stages), result.TotalCost, result.TotalUsage.TotalTokens)

    fmt.Println("\\n=== State Transfer Log ===")
    for _, entry := range result.TransferLog {
        fmt.Printf("  %s -> %s (generation %d)\\n",
            entry.From, entry.To, entry.Generation)
    }
}`}
            />
          </ExampleSection>

          {/* Pipeline State Transfer */}
          <ExampleSection
            id="pipeline-state"
            title="Pipeline State Transfer"
            desc="Demonstrates ownership enforcement for pipeline stages. Each stage processes data and transfers state to the next — previous stages can no longer modify the data."
          >
            <CodeBlock
              filename="examples/pipeline-state/main.go"
              code={`package main

import (
    "context"
    "fmt"
    "log"

    "github.com/lonestarx1/gogrid/pkg/llm"
    "github.com/lonestarx1/gogrid/pkg/memory"
    "github.com/lonestarx1/gogrid/pkg/memory/transfer"
)

func main() {
    ctx := context.Background()

    // Create transferable state wrapping an in-memory store
    state := transfer.NewState(memory.NewInMemory())

    // Register a validation hook
    state.OnTransfer(func(from, to string) error {
        fmt.Printf("  Transfer: %s -> %s\\n", from, to)
        return nil
    })

    // Stage 1: Data collection
    h1, err := state.Acquire("collector")
    if err != nil {
        log.Fatal(err)
    }
    _ = h1.Save(ctx, "pipeline", []llm.Message{
        llm.NewUserMessage("raw data from source"),
    })
    fmt.Println("[collector] Saved raw data")

    // Stage 2: Data processing
    h2, _ := state.Transfer("collector", "processor")
    msgs, _ := h2.Load(ctx, "pipeline")
    msgs = append(msgs, llm.NewAssistantMessage("processed: "+msgs[0].Content))
    _ = h2.Save(ctx, "pipeline", msgs)
    fmt.Println("[processor] Processed data")

    // Stage 1 can no longer access the data
    _, err = h1.Load(ctx, "pipeline")
    fmt.Printf("[collector] Tried to read: %v\\n", err)
    // Output: state has been transferred to a new owner

    // Stage 3: Data output
    h3, _ := state.Transfer("processor", "output")
    final, _ := h3.Load(ctx, "pipeline")
    for _, m := range final {
        fmt.Printf("[output] %s: %s\\n", m.Role, m.Content)
    }

    // Print audit trail
    fmt.Println("\\n=== Audit Trail ===")
    for _, entry := range state.AuditLog() {
        fmt.Printf("  %s -> %s (generation %d)\\n",
            entry.From, entry.To, entry.Generation)
    }
}`}
            />
          </ExampleSection>

          {/* Dynamic Research Coordinator */}
          <ExampleSection
            id="dynamic-research"
            title="Dynamic Research Coordinator"
            desc="A coordinator agent dynamically spawns sub-agents and a pipeline at runtime based on the task. Demonstrates resource governance, async futures, and aggregate metrics."
          >
            <CodeBlock
              filename="examples/dynamic-research/main.go"
              code={`package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/lonestarx1/gogrid/pkg/agent"
    "github.com/lonestarx1/gogrid/pkg/llm/openai"
    "github.com/lonestarx1/gogrid/pkg/orchestrator/dynamic"
    "github.com/lonestarx1/gogrid/pkg/orchestrator/pipeline"
    "github.com/lonestarx1/gogrid/pkg/trace"
)

func main() {
    ctx := context.Background()
    provider := openai.New(os.Getenv("OPENAI_API_KEY"))
    tracer := trace.NewStdout(os.Stdout)

    // Create a runtime with resource governance
    rt := dynamic.New("coordinator",
        dynamic.WithConfig(dynamic.Config{
            MaxConcurrent: 3,
            MaxDepth:      2,
            CostBudget:    5.00,
        }),
        dynamic.WithTracer(tracer),
    )
    ctx = rt.Context(ctx)

    // Define child agents
    researcher := agent.New("researcher",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("Find key facts and data about the given topic."),
    )
    analyst := agent.New("analyst",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("Analyze data and identify trends."),
    )
    writer := agent.New("writer",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("Write a clear summary from the analysis."),
    )

    topic := "Impact of AI agents on enterprise software"

    // Phase 1: Spawn two researchers concurrently using Go
    fmt.Println("Phase 1: Research (parallel)")
    f1 := rt.Go(ctx, "research-papers", func(ctx context.Context) (string, error) {
        r, err := rt.SpawnAgent(ctx, researcher, "Find academic papers about: "+topic)
        if err != nil {
            return "", err
        }
        return r.Message.Content, nil
    })
    f2 := rt.Go(ctx, "research-industry", func(ctx context.Context) (string, error) {
        r, err := rt.SpawnAgent(ctx, analyst, "Find industry reports about: "+topic)
        if err != nil {
            return "", err
        }
        return r.Message.Content, nil
    })

    papers, err := f1.Wait(ctx)
    if err != nil {
        log.Fatal(err)
    }
    industry, err := f2.Wait(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // Phase 2: Feed into a summarization pipeline
    fmt.Println("\\nPhase 2: Summarize (pipeline)")
    combined := "Academic Research:\\n" + papers + "\\n\\nIndustry Analysis:\\n" + industry

    summarizer := pipeline.New("summarize",
        pipeline.WithStages(
            pipeline.Stage{
                Name:  "synthesize",
                Agent: analyst,
                InputTransform: func(in string) string {
                    return "Synthesize these findings into key insights:\\n\\n" + in
                },
            },
            pipeline.Stage{
                Name:  "write",
                Agent: writer,
                InputTransform: func(in string) string {
                    return "Write an executive summary from these insights:\\n\\n" + in
                },
            },
        ),
        pipeline.WithConfig(pipeline.Config{Timeout: time.Minute}),
    )

    pResult, err := rt.SpawnPipeline(ctx, summarizer, combined)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("\\n=== Executive Summary ===")
    fmt.Println(pResult.Output)

    // Aggregate metrics from all children
    res := rt.Result()
    fmt.Printf("\\nChildren: %d | Total Cost: $%.4f | Tokens: %d\\n",
        len(res.Children), res.TotalCost, res.TotalUsage.TotalTokens)
    fmt.Printf("Remaining Budget: $%.4f\\n", rt.RemainingBudget())

    for _, child := range res.Children {
        fmt.Printf("  [%s] %s — $%.4f\\n", child.Type, child.Name, child.Cost)
    }
}`}
            />
          </ExampleSection>

          {/* CLI Multi-Agent Project */}
          <ExampleSection
            id="cli-multi-agent"
            title="CLI Multi-Agent Project"
            desc="Define multiple agents with different providers and models in gogrid.yaml, then run them from the command line. No Go code required — just YAML and API keys."
          >
            <CodeBlock
              filename="gogrid.yaml"
              code={`version: "1"

agents:
  # Research agent using Anthropic's Claude
  researcher:
    model: \${ANTHROPIC_MODEL:-claude-sonnet-4-5-20250929}
    provider: anthropic
    instructions: |
      You are a technical researcher. When given a topic:
      1. Explain the core concepts clearly
      2. Provide concrete examples
      3. Discuss trade-offs and alternatives
      4. Mention common pitfalls
      Be thorough but structured. Use headings and bullet points.
    config:
      max_turns: 10
      max_tokens: 4096
      temperature: 0.7
      timeout: 2m
      cost_budget: 0.50

  # Code review agent using OpenAI's GPT-4o-mini
  code-reviewer:
    model: \${OPENAI_MODEL:-gpt-4o-mini}
    provider: openai
    instructions: |
      You are a senior Go code reviewer. When given code:
      1. Check for correctness and edge cases
      2. Evaluate error handling
      3. Assess naming and readability
      4. Suggest improvements with code examples
      Be constructive. Follow Go conventions and idioms.
    config:
      max_turns: 5
      max_tokens: 2048
      timeout: 60s
      cost_budget: 0.10

  # Summarization agent — fast and cheap
  summarizer:
    model: \${OPENAI_MODEL:-gpt-4o-mini}
    provider: openai
    instructions: |
      Condense the input into 3-5 bullet points. Under 200 words.
    config:
      max_turns: 3
      max_tokens: 1024
      timeout: 30s
      cost_budget: 0.05`}
            />
            <CodeBlock
              filename="terminal"
              code={`# Set API keys for the providers you use
export ANTHROPIC_API_KEY=sk-ant-...
export OPENAI_API_KEY=sk-proj-...

# List all defined agents
$ gogrid list
NAME             PROVIDER    MODEL
code-reviewer    openai      gpt-4o-mini
researcher       anthropic   claude-sonnet-4-5-20250929
summarizer       openai      gpt-4o-mini

# Run the researcher agent
$ gogrid run researcher -input "Explain how garbage collection works in Go"

# Run the code reviewer
$ gogrid run code-reviewer -input "Review this Go function:

func fetchUser(id string) (*User, error) {
    resp, err := http.Get(\\"https://api.example.com/users/\\" + id)
    if err != nil {
        return nil, err
    }
    var user User
    json.NewDecoder(resp.Body).Decode(&user)
    return &user, nil
}"

# Run the summarizer
$ gogrid run summarizer -input "Go is a statically typed, compiled language..."

# Override the model at runtime without editing the config
$ ANTHROPIC_MODEL=claude-haiku-4-5-20251001 gogrid run researcher -input "What is a goroutine?"`}
            />
          </ExampleSection>

          {/* CLI Trace & Cost Inspection */}
          <ExampleSection
            id="cli-trace-cost"
            title="CLI Trace & Cost Inspection"
            desc="After running agents, inspect what happened under the hood. View execution span trees to understand timing, and cost breakdowns to track spend per model."
          >
            <CodeBlock
              filename="terminal"
              code={`# List recent runs
$ gogrid trace
Recent runs:
  019479a3c4e80001  researcher     claude-sonnet-4-5-20250929  4.2s
  019479a1b2c70002  code-reviewer  gpt-4o-mini                1.1s
  019479a0a1b60003  summarizer     gpt-4o-mini                0.8s

# View the span tree for a specific run
$ gogrid trace 019479a3c4e80001
Run: 019479a3c4e80001
Agent: researcher | Model: claude-sonnet-4-5-20250929 | Duration: 4.2s

agent.run (4.2s)
├── memory.load (1ms)
├── llm.complete (2.1s) [prompt: 150, completion: 89]
├── llm.complete (1.8s) [prompt: 280, completion: 145]
└── memory.save (2ms)

# Export spans as JSON for scripts or other tools
$ gogrid trace 019479a3c4e80001 -json | jq '.[].name'
"agent.run"
"memory.load"
"llm.complete"
"llm.complete"
"memory.save"`}
            />
            <CodeBlock
              filename="terminal"
              code={`# View cost overview of all runs
$ gogrid cost
RUN ID              AGENT           MODEL                         COST
019479a3c4e80001    researcher      claude-sonnet-4-5-20250929    $0.003280
019479a1b2c70002    code-reviewer   gpt-4o-mini                   $0.000150
019479a0a1b60003    summarizer      gpt-4o-mini                   $0.000090

# Detailed cost breakdown for a specific run
$ gogrid cost 019479a3c4e80001
Run: 019479a3c4e80001

MODEL                         CALLS  PROMPT  COMPLETION  COST
claude-sonnet-4-5-20250929    2      430     234         $0.003280
────────────────────────────────────────────────────────────────
TOTAL                         2      430     234         $0.003280

# Export costs as JSON
$ gogrid cost -json | jq '.[] | {agent, cost}'

# Inspect raw run records directly
$ cat .gogrid/runs/019479a3c4e80001.json | jq '{agent, model, turns, cost}'
{
  "agent": "researcher",
  "model": "claude-sonnet-4-5-20250929",
  "turns": 2,
  "cost": 0.00328
}`}
            />
          </ExampleSection>

          {/* Project Scaffolding */}
          <ExampleSection
            id="cli-scaffold"
            title="Project Scaffolding"
            desc="Use gogrid init to scaffold a complete project from a template. Each template generates a working project with gogrid.yaml, main.go, Makefile, and README — ready to run."
          >
            <CodeBlock
              filename="terminal"
              code={`# Scaffold a single-agent project
$ gogrid init --template single my-agent
Created GoGrid project in my-agent/
  gogrid.yaml   Agent configuration
  main.go       Programmatic entry point
  Makefile      Build targets
  README.md     Setup instructions

$ cd my-agent && cat gogrid.yaml`}
            />
            <CodeBlock
              filename="my-agent/gogrid.yaml (generated)"
              code={`version: "1"

agents:
  assistant:
    model: gpt-4o-mini
    provider: openai
    instructions: |
      You are a helpful assistant for the my-agent project.
      Provide clear, concise, and accurate responses.
    config:
      max_turns: 10
      max_tokens: 4096
      timeout: 60s`}
            />
            <CodeBlock
              filename="terminal"
              code={`# Scaffold a team project
$ gogrid init --template team my-research-team
$ cat my-research-team/gogrid.yaml`}
            />
            <CodeBlock
              filename="my-research-team/gogrid.yaml (generated)"
              code={`version: "1"

agents:
  researcher:
    model: gpt-4o
    provider: openai
    instructions: |
      You are a research specialist. Investigate the given topic
      thoroughly and provide detailed findings with evidence.
    config:
      max_turns: 10
      max_tokens: 4096
      timeout: 2m
      cost_budget: 1.00

  reviewer:
    model: gpt-4o
    provider: openai
    instructions: |
      You are a critical reviewer. Evaluate the research for accuracy,
      completeness, and potential biases. Suggest improvements.
    config:
      max_turns: 5
      max_tokens: 2048
      timeout: 60s
      cost_budget: 0.50`}
            />
            <CodeBlock
              filename="terminal"
              code={`# Scaffold a pipeline project
$ gogrid init --template pipeline my-content-pipeline

# Set up and run any scaffolded project
$ cd my-content-pipeline
$ go mod init github.com/example/my-content-pipeline
$ go mod tidy
$ export OPENAI_API_KEY=sk-proj-...
$ gogrid list
$ gogrid run drafter -input "Write about AI agents in production"`}
            />
          </ExampleSection>

          {/* Full Observability Stack */}
          <ExampleSection
            id="observability"
            title="Full Observability Stack"
            desc="Wire up OTLP trace export, structured JSON logging with trace correlation, Prometheus-compatible metrics, and cost governance with budget alerts — all using the Go standard library."
          >
            <CodeBlock
              filename="examples/observability/main.go"
              code={`package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/lonestarx1/gogrid/pkg/agent"
    "github.com/lonestarx1/gogrid/pkg/cost"
    "github.com/lonestarx1/gogrid/pkg/llm/openai"
    tracelog "github.com/lonestarx1/gogrid/pkg/trace/log"
    "github.com/lonestarx1/gogrid/pkg/trace/metrics"
    "github.com/lonestarx1/gogrid/pkg/trace/otel"
)

func main() {
    ctx := context.Background()

    // 1. OTLP Exporter — sends spans to Jaeger/Tempo/etc.
    exporter := otel.NewExporter(
        otel.WithEndpoint("http://localhost:4318/v1/traces"),
        otel.WithServiceName("my-agent-service"),
        otel.WithServiceVersion("1.0.0"),
        otel.WithBatchSize(100),
        otel.WithFlushInterval(5 * time.Second),
    )
    defer exporter.Shutdown()

    // 2. Metrics Collector — wraps the exporter, auto-populates metrics
    reg := metrics.NewRegistry()
    collector := metrics.NewCollector(exporter, reg)

    // 3. Structured Logger — JSON logging with trace correlation
    logger := tracelog.New(os.Stdout, tracelog.Info)

    // 4. Cost Governance — budget with threshold alerts
    tracker := cost.NewTracker()
    tracker.SetBudget(5.00)
    tracker.OnBudgetThreshold(func(threshold, current float64) {
        logger.Warn("cost alert",
            "threshold", fmt.Sprintf("%.0f%%", threshold*100),
            "current", fmt.Sprintf("$%.4f", current),
        )
    }, 0.5, 0.8, 1.0)

    // 5. Prometheus metrics endpoint
    go func() {
        http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Content-Type", "text/plain; version=0.0.4")
            fmt.Fprint(w, reg.Export())
        })
        _ = http.ListenAndServe(":9090", nil)
    }()

    // 6. Create agent with the full observability stack
    provider := openai.New(os.Getenv("OPENAI_API_KEY"))
    a := agent.New("assistant",
        agent.WithProvider(provider),
        agent.WithModel("gpt-4o"),
        agent.WithInstructions("You are a helpful assistant."),
        agent.WithTracer(collector), // OTLP export + auto-metrics
    )

    // Run the agent
    logger.Info("starting agent run", "agent", "assistant")
    result, err := a.Run(ctx, "What are the key benefits of observability?")
    if err != nil {
        log.Fatal(err)
    }

    // Log the result with trace correlation
    logger.InfoCtx(ctx, "agent run complete",
        "turns", fmt.Sprintf("%d", result.Turns),
        "cost", fmt.Sprintf("$%.6f", result.Cost),
        "tokens", fmt.Sprintf("%d", result.Usage.TotalTokens),
    )

    // Record cost for governance
    tracker.AddForEntity("gpt-4o", "assistant", result.Usage)

    // Generate cost report
    report := tracker.Report()
    fmt.Printf("\\nCost Report: $%.4f across %d calls\\n", report.TotalCost, report.RecordCount)
    for model, mr := range report.ByModel {
        fmt.Printf("  %s: %d calls, $%.4f\\n", model, mr.Calls, mr.Cost)
    }

    fmt.Printf("\\nMetrics available at http://localhost:9090/metrics\\n")
    fmt.Println(result.Message.Content)
}`}
            />
          </ExampleSection>

          <ExampleSection
            id="eval-suite"
            title="Evaluation Suite"
            desc="Score agent outputs with composable evaluators — exact match, cost budgets, and custom checks."
          >
            <CodeBlock
              code={`package main

import (
    "context"
    "fmt"
    "log"

    "github.com/lonestarx1/gogrid/pkg/agent"
    "github.com/lonestarx1/gogrid/pkg/eval"
    "github.com/lonestarx1/gogrid/pkg/llm"
    "github.com/lonestarx1/gogrid/pkg/llm/mock"
)

func main() {
    provider := mock.New(mock.WithFallback(&llm.Response{
        Message: llm.NewAssistantMessage("Go is a statically typed, compiled language designed for simplicity."),
        Usage:   llm.Usage{PromptTokens: 15, CompletionTokens: 14, TotalTokens: 29},
        Model:   "mock",
    }))

    a := agent.New("eval-agent",
        agent.WithProvider(provider),
        agent.WithModel("mock"),
    )

    result, err := a.Run(context.Background(), "Describe Go in one sentence.")
    if err != nil {
        log.Fatal(err)
    }

    // Compose evaluators into a suite.
    suite := eval.NewSuite(
        eval.NewContains("Go", "compiled", "simplicity"),
        eval.NewCostWithin(0.01),
        eval.NewFunc("min_length", func(_ context.Context, r *agent.Result) (eval.Score, error) {
            if len(r.Message.Content) >= 20 {
                return eval.Score{Pass: true, Value: 1.0, Reason: "sufficient length"}, nil
            }
            return eval.Score{Pass: false, Value: 0.0, Reason: "too short"}, nil
        }),
    )

    sr, _ := suite.Run(context.Background(), result)
    fmt.Printf("Suite passed: %v\\n", sr.Pass)
    for name, score := range sr.Scores {
        fmt.Printf("  %s: pass=%v value=%.2f reason=%s\\n",
            name, score.Pass, score.Value, score.Reason)
    }
}`}
            />
          </ExampleSection>

          <ExampleSection
            id="mock-testing"
            title="Mock Provider Testing"
            desc="Test agents with sequential responses, error injection, and call recording — no API keys needed."
          >
            <CodeBlock
              code={`package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/lonestarx1/gogrid/pkg/agent"
    "github.com/lonestarx1/gogrid/pkg/llm"
    "github.com/lonestarx1/gogrid/pkg/llm/mock"
)

func main() {
    // Queue a tool call response followed by a final answer.
    toolCallResp := &llm.Response{
        Message: llm.Message{
            Role: llm.RoleAssistant,
            ToolCalls: []llm.ToolCall{
                {ID: "tc-1", Function: "search", Arguments: json.RawMessage(\`{"q":"Go"}\`)},
            },
        },
        Usage: llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
        Model: "mock",
    }
    finalResp := &llm.Response{
        Message: llm.NewAssistantMessage("Go was created at Google in 2007."),
        Usage:   llm.Usage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30},
        Model:   "mock",
    }

    provider := mock.New(
        mock.WithResponses(toolCallResp, finalResp),
        mock.WithFallback(finalResp),
    )

    a := agent.New("test-agent",
        agent.WithProvider(provider),
        agent.WithModel("mock"),
    )

    result, err := a.Run(context.Background(), "When was Go created?")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %s\\n", result.Message.Content)
    fmt.Printf("Calls: %d\\n", provider.Calls())
    fmt.Printf("History: %d params recorded\\n", len(provider.History()))
}`}
            />
          </ExampleSection>
        </motion.div>
      </main>
    </div>
  );
}

function ExampleSection({
  id,
  title,
  desc,
  children,
}: {
  id: string;
  title: string;
  desc: string;
  children: React.ReactNode;
}) {
  return (
    <section id={id} className="mb-16 scroll-mt-20">
      <h2 className="font-mono text-2xl font-bold text-white mb-2">{title}</h2>
      <p className="text-text-muted mb-6">{desc}</p>
      {children}
    </section>
  );
}
