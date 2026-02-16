"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import CodeBlock from "@/components/CodeBlock";

const sections = [
  { id: "getting-started", label: "Getting Started" },
  { id: "core-types", label: "Core Types" },
  { id: "single-agent", label: "Single Agent" },
  { id: "providers", label: "LLM Providers" },
  { id: "tools", label: "Tools" },
  { id: "memory", label: "Memory" },
  { id: "file-memory", label: "File Memory" },
  { id: "shared-memory", label: "Shared Memory" },
  { id: "transferable-state", label: "Transferable State" },
  { id: "prune-policies", label: "Prune Policies" },
  { id: "team", label: "Team (Chat Room)" },
  { id: "message-bus", label: "Message Bus" },
  { id: "consensus", label: "Consensus Strategies" },
  { id: "coordinator", label: "Coordinator" },
  { id: "tracing", label: "Tracing" },
  { id: "cost-tracking", label: "Cost Tracking" },
];

export default function DocsPage() {
  const [active, setActive] = useState("getting-started");
  // Scroll-spy: pick the section whose top has most recently scrolled past the threshold
  useEffect(() => {
    const ids = sections.map((s) => s.id);
    const threshold = 100; // px from top of viewport

    const onScroll = () => {
      // If scrolled to bottom, activate last section
      if (window.innerHeight + window.scrollY >= document.body.scrollHeight - 2) {
        setActive(ids[ids.length - 1]);
        return;
      }

      let current = ids[0];
      for (const id of ids) {
        const el = document.getElementById(id);
        if (!el) continue;
        if (el.getBoundingClientRect().top <= threshold) {
          current = id;
        } else {
          break;
        }
      }
      setActive(current);
    };

    window.addEventListener("scroll", onScroll, { passive: true });
    onScroll();
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  const scrollTo = (id: string) => {
    document.getElementById(id)?.scrollIntoView({ behavior: "smooth" });
  };

  return (
    <div className="pt-14 min-h-screen">
      <div className="max-w-7xl mx-auto flex">
        {/* Sidebar */}
        <aside className="hidden lg:block w-64 shrink-0 border-r border-border sticky top-14 h-[calc(100vh-3.5rem)] overflow-y-auto py-8 px-4">
          <p className="font-mono text-xs text-accent uppercase tracking-widest mb-6">
            Documentation
          </p>
          <nav className="space-y-1">
            {sections.map((s) => (
              <button
                key={s.id}
                onClick={() => scrollTo(s.id)}
                className={`block w-full text-left px-3 py-1.5 rounded text-sm font-mono transition-colors ${
                  active === s.id
                    ? "text-accent bg-accent/10"
                    : "text-text-muted hover:text-text hover:bg-bg-card"
                }`}
              >
                {s.label}
              </button>
            ))}
          </nav>
        </aside>

        {/* Content */}
        <main className="flex-1 min-w-0 px-6 lg:px-12 py-12 max-w-4xl">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
          >
            <h1 className="font-mono text-4xl md:text-5xl font-bold text-white mb-4">
              Documentation
            </h1>
            <p className="text-text-muted text-lg mb-12 max-w-2xl">
              Everything you need to build production-grade AI agents with GoGrid.
            </p>

            {/* Getting Started */}
            <Section id="getting-started" title="Getting Started">
              <P>
                GoGrid (G2) is an open-source AI agent framework written in Go. It provides
                five composable orchestration patterns: Single Agent, Team, Pipeline, Graph,
                and Dynamic Orchestration.
              </P>
              <H3>Installation</H3>
              <CodeBlock
                code={`go get github.com/lonestarx1/gogrid`}
                filename="terminal"
              />
              <H3>Project Structure</H3>
              <CodeBlock
                code={`gogrid/
├── pkg/
│   ├── agent/          # Agent creation and execution
│   ├── llm/            # LLM provider interfaces and implementations
│   │   ├── openai/     # OpenAI provider
│   │   ├── anthropic/  # Anthropic provider
│   │   └── gemini/     # Google Gemini provider
│   ├── memory/         # Memory interfaces and implementations
│   │   ├── file/       # File-backed memory
│   │   ├── shared/     # Shared memory for teams
│   │   └── transfer/   # Transferable state for pipelines
│   ├── tool/           # Tool interface and registry
│   ├── trace/          # Tracing and observability
│   ├── cost/           # Cost tracking and budgets
│   └── orchestrator/
│       └── team/       # Team (chat room) orchestrator
├── internal/
│   └── id/             # ID generation
└── cmd/
    └── gogrid/         # CLI entry point`}
                filename="project layout"
              />
            </Section>

            {/* Core Types */}
            <Section id="core-types" title="Core Types">
              <P>
                GoGrid is built on a small set of composable interfaces. These are the
                building blocks for all orchestration patterns.
              </P>
              <H3>Message</H3>
              <P>
                Messages represent conversation turns. Every LLM interaction flows through
                the <Code>llm.Message</Code> type.
              </P>
              <CodeBlock
                code={`// Roles: system, user, assistant, tool
msg := llm.NewUserMessage("What is the weather?")
msg := llm.NewAssistantMessage("The weather is sunny.")
msg := llm.NewSystemMessage("You are a helpful assistant.")
msg := llm.NewToolMessage(callID, "72°F and sunny")`}
                filename="pkg/llm/message.go"
              />

              <H3>Provider</H3>
              <P>
                The <Code>llm.Provider</Code> interface abstracts LLM backends. Swapping
                providers is a configuration change — no code changes required.
              </P>
              <CodeBlock
                code={`type Provider interface {
    Complete(ctx context.Context, params Params) (*Response, error)
}

// Params includes Model, Messages, Tools, Temperature, MaxTokens
// Response includes Message, Usage (tokens), and Model`}
                filename="pkg/llm/provider.go"
              />

              <H3>Tool</H3>
              <P>
                Tools are functions that agents can call. Each tool has a name, description,
                JSON Schema for parameters, and an Execute method.
              </P>
              <CodeBlock
                code={`type Tool interface {
    Name() string
    Description() string
    Schema() Schema
    Execute(ctx context.Context, input json.RawMessage) (string, error)
}`}
                filename="pkg/tool/tool.go"
              />

              <H3>Memory</H3>
              <P>
                Memory is a first-class primitive, not a plugin. Every agent pattern is
                designed with memory as a core concern.
              </P>
              <CodeBlock
                code={`type Memory interface {
    Load(ctx context.Context, key string) ([]llm.Message, error)
    Save(ctx context.Context, key string, messages []llm.Message) error
    Clear(ctx context.Context, key string) error
}`}
                filename="pkg/memory/memory.go"
              />
            </Section>

            {/* Single Agent */}
            <Section id="single-agent" title="Single Agent">
              <P>
                The Single Agent pattern is the fundamental unit of work in GoGrid. It
                combines an LLM provider, tools, memory, and configuration to execute
                tasks through an iterative tool-use loop.
              </P>
              <H3>Creating an Agent</H3>
              <P>
                Agents are created with functional options. At minimum, you need a name,
                provider, and model.
              </P>
              <CodeBlock
                code={`import (
    "github.com/lonestarx1/gogrid/pkg/agent"
    "github.com/lonestarx1/gogrid/pkg/llm/openai"
    "github.com/lonestarx1/gogrid/pkg/memory"
)

provider := openai.New(os.Getenv("OPENAI_API_KEY"))

a := agent.New("assistant",
    agent.WithProvider(provider),
    agent.WithModel("gpt-4o"),
    agent.WithInstructions("You are a helpful assistant."),
    agent.WithMemory(memory.NewInMemory()),
    agent.WithTools(mySearchTool, myCalcTool),
    agent.WithConfig(agent.Config{
        MaxTurns:   10,
        MaxTokens:  4096,
        CostBudget: 1.00, // USD
    }),
)`}
                filename="main.go"
              />
              <H3>Running an Agent</H3>
              <P>
                Call <Code>Run</Code> with a context and user input. The agent loops:
                call LLM, execute tool calls, repeat until the LLM gives a final response
                or limits are hit.
              </P>
              <CodeBlock
                code={`result, err := a.Run(ctx, "What is 42 * 17?")
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Message.Content) // "42 * 17 = 714"
fmt.Printf("Turns: %d, Cost: $%.4f\\n", result.Turns, result.Cost)
fmt.Printf("Tokens: %d prompt, %d completion\\n",
    result.Usage.PromptTokens, result.Usage.CompletionTokens)`}
                filename="main.go"
              />
              <H3>Agent Loop</H3>
              <P>The execution flow inside <Code>Agent.Run</Code>:</P>
              <ol className="list-decimal list-inside space-y-2 text-text-muted mb-6">
                <li>Build messages from system prompt, memory history, and user input</li>
                <li>Call the LLM with messages and tool definitions</li>
                <li>If the LLM responds with tool calls, execute them and loop</li>
                <li>If the LLM responds with a final message, return the result</li>
                <li>Respect max turns, timeout, and cost budget at each iteration</li>
              </ol>
            </Section>

            {/* LLM Providers */}
            <Section id="providers" title="LLM Providers">
              <P>
                GoGrid ships with three built-in providers. All implement the same
                <Code>llm.Provider</Code> interface, so swapping is a one-line change.
              </P>
              <H3>OpenAI</H3>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/llm/openai"

provider := openai.New(os.Getenv("OPENAI_API_KEY"))
// Models: "gpt-4o", "gpt-4o-mini", "gpt-4.1", "o3", etc.`}
                filename="openai"
              />
              <H3>Anthropic</H3>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/llm/anthropic"

provider := anthropic.New(os.Getenv("ANTHROPIC_API_KEY"))
// Models: "claude-sonnet-4-5-20250929", "claude-opus-4-6-20250827", etc.`}
                filename="anthropic"
              />
              <H3>Google Gemini</H3>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/llm/gemini"

provider, err := gemini.New(ctx, os.Getenv("GOOGLE_API_KEY"))
// Models: "gemini-2.5-pro", "gemini-2.5-flash", etc.`}
                filename="gemini"
              />
            </Section>

            {/* Tools */}
            <Section id="tools" title="Tools">
              <P>
                Tools give agents the ability to take actions. Define the tool interface
                and the agent will call it when appropriate.
              </P>
              <CodeBlock
                code={`type CalculatorTool struct{}

func (t *CalculatorTool) Name() string        { return "calculator" }
func (t *CalculatorTool) Description() string  { return "Evaluate math expressions" }
func (t *CalculatorTool) Schema() tool.Schema {
    return tool.Schema{
        Type: "object",
        Properties: map[string]*tool.Schema{
            "expression": {Type: "string", Description: "Math expression to evaluate"},
        },
        Required: []string{"expression"},
    }
}

func (t *CalculatorTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
    var args struct {
        Expression string \`json:"expression"\`
    }
    if err := json.Unmarshal(input, &args); err != nil {
        return "", err
    }
    // Evaluate the expression...
    return result, nil
}`}
                filename="tools/calculator.go"
              />
              <H3>Tool Registry</H3>
              <P>
                The <Code>tool.Registry</Code> provides centralized tool management with
                duplicate-name prevention.
              </P>
              <CodeBlock
                code={`registry := tool.NewRegistry()
registry.Register(&CalculatorTool{})
registry.Register(&SearchTool{})

t, err := registry.Get("calculator") // retrieve by name
names := registry.List()             // ["calculator", "search"]`}
                filename="registry"
              />
            </Section>

            {/* Memory */}
            <Section id="memory" title="Memory">
              <P>
                Memory is a first-class primitive in GoGrid. The base <Code>Memory</Code> interface
                provides Load/Save/Clear. Extended interfaces add search, pruning, and statistics.
              </P>
              <H3>InMemory Store</H3>
              <P>
                The default memory implementation. Thread-safe, suitable for development and
                short-lived sessions. Implements all extended interfaces.
              </P>
              <CodeBlock
                code={`mem := memory.NewInMemory()

// Save conversation
_ = mem.Save(ctx, "session-1", []llm.Message{
    llm.NewUserMessage("Hello"),
    llm.NewAssistantMessage("Hi there!"),
})

// Load conversation
msgs, _ := mem.Load(ctx, "session-1")

// Search across all keys (case-insensitive)
entries, _ := mem.Search(ctx, "hello")

// Get aggregate statistics
stats, _ := mem.Stats(ctx)
fmt.Printf("Keys: %d, Entries: %d, Size: %d bytes\\n",
    stats.Keys, stats.TotalEntries, stats.TotalSize)`}
                filename="memory usage"
              />
              <H3>ReadOnly Wrapper</H3>
              <P>
                Wraps any Memory, permitting loads but rejecting saves and clears with
                <Code>ErrReadOnly</Code>. Useful for reference data.
              </P>
              <CodeBlock
                code={`ro := memory.NewReadOnly(inner)
msgs, _ := ro.Load(ctx, "k")  // works
err := ro.Save(ctx, "k", msgs) // returns memory.ErrReadOnly`}
                filename="readonly"
              />
              <H3>Extended Interfaces</H3>
              <CodeBlock
                code={`// SearchableMemory adds keyword search
type SearchableMemory interface {
    Memory
    Search(ctx context.Context, query string) ([]Entry, error)
}

// PrunableMemory adds policy-based pruning
type PrunableMemory interface {
    Memory
    Prune(ctx context.Context, policy PrunePolicy) (int, error)
}

// StatsMemory adds aggregate statistics
type StatsMemory interface {
    Memory
    Stats(ctx context.Context) (*Stats, error)
}`}
                filename="pkg/memory/memory.go"
              />
            </Section>

            {/* File Memory */}
            <Section id="file-memory" title="File Memory">
              <P>
                File-backed memory persists data to disk as JSON files. Each key maps to a
                separate file (hex-encoded filename) with a <Code>.meta.json</Code> sidecar
                for timestamps and sizes. Easy to inspect, no contention between keys.
              </P>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/memory/file"

mem, err := file.New("/var/data/agent-memory")
if err != nil {
    log.Fatal(err)
}

// Use like any other Memory
_ = mem.Save(ctx, "session-1", messages)
loaded, _ := mem.Load(ctx, "session-1")

// Also supports Search, Prune, and Stats
results, _ := mem.Search(ctx, "keyword")
stats, _ := mem.Stats(ctx)`}
                filename="file memory"
              />
              <P>
                On disk, keys are hex-encoded for filesystem safety. A key <Code>session-1</Code> becomes:
              </P>
              <CodeBlock
                code={`/var/data/agent-memory/
├── 73657373696f6e2d31.json       # message data
└── 73657373696f6e2d31.meta.json  # metadata sidecar`}
                filename="filesystem layout"
              />
            </Section>

            {/* Shared Memory */}
            <Section id="shared-memory" title="Shared Memory">
              <P>
                Shared memory is designed for GoGrid&apos;s team/chat-room patterns. Multiple
                agents read and write concurrently, with optional change notifications via
                Go channels.
              </P>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/memory/shared"

pool := shared.New()

// Subscribe to changes (non-blocking sends)
ch := make(chan shared.ChangeEvent, 10)
unsub := pool.Subscribe(ch)
defer unsub()

// Agents write to the pool
_ = pool.Save(ctx, "agent-a/result", messages)

// Receive notifications
event := <-ch
fmt.Printf("Change: %v on key %q\\n", event.Type, event.Key)`}
                filename="shared memory"
              />
              <H3>Namespaced Views</H3>
              <P>
                <Code>NamespacedView</Code> transparently prefixes keys so agents can share
                a pool without collisions.
              </P>
              <CodeBlock
                code={`pool := shared.New()

// Each agent gets an isolated view of the same pool
viewA := shared.NewNamespacedView(pool, "agent-a")
viewB := shared.NewNamespacedView(pool, "agent-b")

// Both use key "result" but they don't collide
_ = viewA.Save(ctx, "result", msgsA) // stored as "agent-a/result"
_ = viewB.Save(ctx, "result", msgsB) // stored as "agent-b/result"`}
                filename="namespaced views"
              />
            </Section>

            {/* Transferable State */}
            <Section id="transferable-state" title="Transferable State">
              <P>
                For pipeline patterns, state must transfer between stages — not be shared.
                <Code>TransferableState</Code> uses a generation counter so that when
                ownership moves to the next agent, the previous owner&apos;s handle becomes
                invalid.
              </P>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/memory/transfer"

// Create transferable state wrapping any Memory
state := transfer.NewState(memory.NewInMemory())

// Stage 1 acquires ownership
h1, _ := state.Acquire("stage-1")
_ = h1.Save(ctx, "pipeline", messages)

// Transfer to stage 2
h2, _ := state.Transfer("stage-1", "stage-2")
loaded, _ := h2.Load(ctx, "pipeline") // works

// Stage 1's handle is now invalid
_, err := h1.Load(ctx, "pipeline")
// err == transfer.ErrStateTransferred`}
                filename="transferable state"
              />
              <H3>Validation Hooks</H3>
              <P>
                Register hooks to enforce policies before transfers occur.
              </P>
              <CodeBlock
                code={`state.OnTransfer(func(from, to string) error {
    if to == "untrusted-agent" {
        return errors.New("transfer to untrusted agent denied")
    }
    return nil
})

// Audit trail
log := state.AuditLog()
for _, entry := range log {
    fmt.Printf("%s -> %s (gen %d)\\n", entry.From, entry.To, entry.Generation)
}`}
                filename="hooks and audit"
              />
            </Section>

            {/* Prune Policies */}
            <Section id="prune-policies" title="Prune Policies">
              <P>
                Prune policies control which memory entries get removed. Policies implement
                the <Code>PrunePolicy</Code> interface and compose with <Code>AnyPolicy</Code>.
              </P>
              <CodeBlock
                code={`// Remove entries older than 1 hour
removed, _ := mem.Prune(ctx, memory.NewMaxAge(1 * time.Hour))

// Remove entries larger than 10KB
removed, _ := mem.Prune(ctx, memory.NewMaxSize(10_000))

// Keep only the last 100 entries per key
removed, _ := mem.Prune(ctx, memory.NewMaxEntries(100))

// Compose: prune if old OR too large
policy := memory.NewAnyPolicy(
    memory.NewMaxAge(24 * time.Hour),
    memory.NewMaxSize(50_000),
)
removed, _ := mem.Prune(ctx, policy)`}
                filename="prune policies"
              />
            </Section>

            {/* Team */}
            <Section id="team" title="Team (Chat Room)">
              <P>
                The Team pattern orchestrates multiple agents working concurrently on the
                same input. Agents communicate via a shared message bus, store results in
                shared memory, and reach decisions through pluggable consensus strategies.
              </P>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/orchestrator/team"

// Create agents with different perspectives
reviewer := agent.New("reviewer",
    agent.WithProvider(provider),
    agent.WithModel("gpt-4o"),
    agent.WithInstructions("You review code for correctness and clarity."),
)
security := agent.New("security",
    agent.WithProvider(provider),
    agent.WithModel("gpt-4o"),
    agent.WithInstructions("You review code for security vulnerabilities."),
)
perf := agent.New("performance",
    agent.WithProvider(provider),
    agent.WithModel("gpt-4o"),
    agent.WithInstructions("You review code for performance issues."),
)

// Create the team
t := team.New("code-review",
    team.WithMembers(
        team.Member{Agent: reviewer, Role: "code quality"},
        team.Member{Agent: security, Role: "security"},
        team.Member{Agent: perf, Role: "performance"},
    ),
    team.WithStrategy(team.Unanimous{}),
    team.WithConfig(team.Config{
        MaxRounds:  1,
        CostBudget: 5.00,
    }),
)

result, err := t.Run(ctx, "Review this function: ...")
fmt.Println(result.Decision.Content)
fmt.Printf("Cost: $%.4f, Rounds: %d\\n", result.TotalCost, result.Rounds)`}
                filename="team orchestrator"
              />
              <H3>Multi-Round Discussions</H3>
              <P>
                With <Code>MaxRounds &gt; 1</Code>, agents see each other&apos;s previous
                responses and can refine their answers across rounds. The team continues
                until the consensus strategy is satisfied or max rounds are reached.
              </P>
              <CodeBlock
                code={`t := team.New("debate",
    team.WithMembers(
        team.Member{Agent: proAgent, Role: "advocate"},
        team.Member{Agent: conAgent, Role: "skeptic"},
    ),
    team.WithStrategy(team.Unanimous{}),
    team.WithConfig(team.Config{MaxRounds: 3}),
)

// Round 1: Both agents respond independently
// Round 2: Each sees the other's Round 1 response
// Round 3: Each sees all prior responses
result, _ := t.Run(ctx, "Should we adopt microservices?")`}
                filename="multi-round"
              />
              <H3>Per-Agent Results</H3>
              <CodeBlock
                code={`for name, agentResult := range result.Responses {
    fmt.Printf("[%s] %s\\n", name, agentResult.Message.Content)
    fmt.Printf("  Tokens: %d, Cost: $%.4f\\n",
        agentResult.Usage.TotalTokens, agentResult.Cost)
}`}
                filename="inspecting results"
              />
            </Section>

            {/* Message Bus */}
            <Section id="message-bus" title="Message Bus">
              <P>
                The message bus provides pub/sub communication within a team. Subscribe to
                topics to monitor agent responses in real time. Sends are non-blocking —
                messages are dropped if a subscriber&apos;s channel is full.
              </P>
              <CodeBlock
                code={`bus := team.NewBus()

// Subscribe to agent responses
ch, unsub := bus.Subscribe("team.response", 100)
defer unsub()

// Monitor in a goroutine
go func() {
    for msg := range ch {
        fmt.Printf("[%s] %s (round %s)\\n",
            msg.From, msg.Content, msg.Metadata["round"])
    }
}()

// Pass the bus to a team
t := team.New("my-team",
    team.WithBus(bus),
    team.WithMembers(...),
)
t.Run(ctx, "discuss this topic")

// Retrieve full history
history := bus.History()`}
                filename="message bus"
              />
            </Section>

            {/* Consensus */}
            <Section id="consensus" title="Consensus Strategies">
              <P>
                Strategies determine when a team has reached a decision and how to form
                the final answer.
              </P>
              <div className="space-y-4 mb-6">
                <Card title="Unanimous" desc="Waits for all agents to respond. Combines all responses in alphabetical order. Default strategy." />
                <Card title="Majority" desc="Waits for more than half of agents. Returns as soon as a majority have responded." />
                <Card title="FirstResponse" desc="Returns immediately when any agent completes. Cancels remaining agents." />
              </div>
              <H3>Custom Strategies</H3>
              <P>
                Implement the <Code>Strategy</Code> interface for domain-specific consensus logic.
              </P>
              <CodeBlock
                code={`type Strategy interface {
    Name() string
    Evaluate(total int, responses map[string]string) (decision string, reached bool)
}

// Example: converge when all agents agree on a keyword
type KeywordConsensus struct {
    Keyword string
}

func (k KeywordConsensus) Name() string { return "keyword" }
func (k KeywordConsensus) Evaluate(total int, responses map[string]string) (string, bool) {
    if len(responses) < total {
        return "", false
    }
    for _, content := range responses {
        if !strings.Contains(strings.ToLower(content), k.Keyword) {
            return "", false
        }
    }
    return combineResponses(responses), true
}`}
                filename="custom strategy"
              />
            </Section>

            {/* Coordinator */}
            <Section id="coordinator" title="Coordinator">
              <P>
                By default, team decisions are formed by concatenating agent responses.
                A <Code>coordinator</Code> is an optional leader agent that receives all
                member responses and synthesizes a single, coherent final decision — like
                a team lead who listens to everyone before making the call.
              </P>
              <CodeBlock
                code={`coordinator := agent.New("lead",
    agent.WithProvider(provider),
    agent.WithModel("gpt-4o"),
    agent.WithInstructions("You are the team lead. Synthesize all perspectives into a clear decision."),
)

t := team.New("code-review",
    team.WithMembers(
        team.Member{Agent: reviewer, Role: "correctness"},
        team.Member{Agent: security, Role: "security"},
        team.Member{Agent: perf, Role: "performance"},
    ),
    team.WithCoordinator(coordinator),
    team.WithStrategy(team.Unanimous{}),
    team.WithConfig(team.Config{MaxRounds: 1}),
)

result, _ := t.Run(ctx, "Review this function: ...")
// result.Decision.Content is the coordinator's synthesized answer
// result.Responses includes both member and coordinator results`}
                filename="coordinator"
              />
              <H3>How It Works</H3>
              <ol className="list-decimal list-inside space-y-2 text-text-muted mb-6">
                <li>Member agents run their rounds as normal (controlled by the consensus strategy)</li>
                <li>After rounds complete, the coordinator receives the original input and all member responses</li>
                <li>The coordinator produces the final team decision</li>
                <li>If the coordinator fails, the team falls back to the combined member responses</li>
              </ol>
              <H3>Coordinator Costs</H3>
              <P>
                The coordinator&apos;s LLM call is included in the team&apos;s total cost and
                token usage. Its result appears in <Code>result.Responses</Code> alongside
                member results, and a <Code>team.coordinator</Code> trace span is emitted.
              </P>
            </Section>

            {/* Tracing */}
            <Section id="tracing" title="Tracing">
              <P>
                GoGrid has built-in structured tracing. Every agent run, LLM call, tool
                execution, and memory operation produces spans with parent-child relationships.
              </P>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/trace"

// In-memory tracer for testing/debugging
tracer := trace.NewInMemory()

a := agent.New("agent",
    agent.WithProvider(provider),
    agent.WithModel("gpt-4o"),
    agent.WithTracer(tracer),
)

result, _ := a.Run(ctx, "hello")

// Inspect spans
for _, span := range tracer.Spans() {
    fmt.Printf("[%s] %s (%v)\\n", span.Name, span.ID, span.Attributes)
}
// Output:
// [llm.complete] abc123 {llm.model: gpt-4o, llm.turn: 1}
// [memory.load] def456 {memory.key: agent}
// [memory.save] ghi789 {memory.key: agent, memory.entries: 3}
// [agent.run] jkl012 {agent.name: agent, agent.turns: 1}`}
                filename="tracing"
              />
              <H3>Stdout Tracer</H3>
              <P>
                For structured logging, the stdout tracer writes spans as JSON lines.
              </P>
              <CodeBlock
                code={`tracer := trace.NewStdout(os.Stdout)
// Each completed span is written as a JSON line:
// {"id":"...","name":"agent.run","start_time":"...","attributes":{...}}`}
                filename="stdout tracer"
              />
              <H3>Team Trace Spans</H3>
              <P>
                Team execution produces a trace tree with team, round, and per-agent spans.
              </P>
              <CodeBlock
                code={`// Trace tree for a team run:
// team.run
//   └─ team.round (round=1)
//       ├─ agent.run (agent-a)
//       │   ├─ memory.load
//       │   ├─ llm.complete
//       │   └─ memory.save
//       └─ agent.run (agent-b)
//           ├─ memory.load
//           ├─ llm.complete
//           └─ memory.save`}
                filename="team trace tree"
              />
            </Section>

            {/* Cost Tracking */}
            <Section id="cost-tracking" title="Cost Tracking">
              <P>
                Every LLM call is metered. GoGrid includes default pricing for popular
                models and tracks costs per agent and per team.
              </P>
              <CodeBlock
                code={`// Per-agent cost in Result
result, _ := agent.Run(ctx, "hello")
fmt.Printf("Cost: $%.6f\\n", result.Cost)

// Per-team aggregate cost
teamResult, _ := t.Run(ctx, "discuss")
fmt.Printf("Team cost: $%.6f\\n", teamResult.TotalCost)

// Budget enforcement — agent stops when budget is exceeded
a := agent.New("agent",
    agent.WithConfig(agent.Config{CostBudget: 0.50}),
    // ...
)

// Team budget — team stops starting new rounds when exceeded
t := team.New("team",
    team.WithConfig(team.Config{CostBudget: 5.00}),
    // ...
)`}
                filename="cost tracking"
              />
              <H3>Memory Stats in Results</H3>
              <P>
                When the agent&apos;s memory implements <Code>StatsMemory</Code>, the result
                includes aggregate memory statistics.
              </P>
              <CodeBlock
                code={`if result.MemoryStats != nil {
    fmt.Printf("Memory: %d keys, %d entries, %d bytes\\n",
        result.MemoryStats.Keys,
        result.MemoryStats.TotalEntries,
        result.MemoryStats.TotalSize)
}`}
                filename="memory stats"
              />
            </Section>
          </motion.div>
        </main>
      </div>
    </div>
  );
}

/* ---------- reusable sub-components ---------- */

function Section({
  id,
  title,
  children,
}: {
  id: string;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section id={id} className="mb-16 scroll-mt-20">
      <h2 className="font-mono text-2xl md:text-3xl font-bold text-white mb-6 pb-3 border-b border-border">
        {title}
      </h2>
      {children}
    </section>
  );
}

function H3({ children }: { children: React.ReactNode }) {
  return (
    <h3 className="font-mono text-lg font-semibold text-white mt-8 mb-3">
      {children}
    </h3>
  );
}

function P({ children }: { children: React.ReactNode }) {
  return <p className="text-text-muted leading-relaxed mb-4">{children}</p>;
}

function Code({ children }: { children: React.ReactNode }) {
  return (
    <code className="font-mono text-sm text-accent bg-bg-card px-1.5 py-0.5 rounded">
      {children}
    </code>
  );
}

function Card({ title, desc }: { title: string; desc: string }) {
  return (
    <div className="bg-bg-card border border-border rounded-lg p-4">
      <p className="font-mono text-sm font-semibold text-accent mb-1">{title}</p>
      <p className="text-text-muted text-sm">{desc}</p>
    </div>
  );
}
