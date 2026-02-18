"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import CodeBlock from "@/components/CodeBlock";

const sections = [
  { id: "getting-started", label: "Getting Started" },
  { id: "cli", label: "CLI" },
  { id: "cli-config", label: "Configuration (YAML)" },
  { id: "cli-commands", label: "CLI Commands" },
  { id: "cli-run-records", label: "Run Records" },
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
  { id: "pipeline", label: "Pipeline" },
  { id: "pipeline-stages", label: "Pipeline Stages" },
  { id: "pipeline-retry", label: "Retry & Error Handling" },
  { id: "graph", label: "Graph" },
  { id: "graph-builder", label: "Graph Builder" },
  { id: "graph-advanced", label: "Loops & Conditions" },
  { id: "dynamic", label: "Dynamic Orchestration" },
  { id: "dynamic-governance", label: "Resource Governance" },
  { id: "dynamic-async", label: "Async & Futures" },
  { id: "tracing", label: "Tracing" },
  { id: "otel", label: "OpenTelemetry Export" },
  { id: "structured-logging", label: "Structured Logging" },
  { id: "metrics", label: "Metrics" },
  { id: "cost-tracking", label: "Cost Tracking" },
  { id: "cost-governance", label: "Cost Governance" },
  { id: "mock-provider", label: "Mock Provider" },
  { id: "evaluation", label: "Evaluation Framework" },
  { id: "benchmarks", label: "Benchmarks" },
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
                GoGrid (G2) is a unified system for developing and orchestrating production
                AI agents in Go. It supports five composable workload types: Single Agent,
                Team, Pipeline, Graph, and Dynamic Orchestration.
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
│   │   ├── gemini/     # Google Gemini provider
│   │   └── mock/       # Mock provider for testing
│   ├── memory/         # Memory interfaces and implementations
│   │   ├── file/       # File-backed memory
│   │   ├── shared/     # Shared memory for teams
│   │   └── transfer/   # Transferable state for pipelines
│   ├── tool/           # Tool interface and registry
│   ├── trace/          # Tracing and observability
│   │   ├── otel/       # OTLP JSON exporter
│   │   ├── log/        # Structured JSON logging
│   │   └── metrics/    # Prometheus-compatible metrics
│   ├── cost/           # Cost tracking, budgets, and governance
│   ├── eval/           # Evaluation framework and benchmarks
│   │   └── bench/      # Performance benchmarks
│   └── orchestrator/
│       ├── team/       # Team (chat room) orchestrator
│       ├── pipeline/   # Pipeline (linear) orchestrator
│       ├── graph/      # Graph orchestrator
│       └── dynamic/    # Dynamic orchestration runtime
├── internal/
│   ├── id/             # ID generation
│   ├── config/         # YAML configuration loading
│   ├── runrecord/      # Run record persistence
│   └── cli/            # CLI implementation
│       └── templates/  # Project scaffolding templates
└── cmd/
    └── gogrid/         # CLI entry point`}
                filename="project layout"
              />
            </Section>

            {/* CLI */}
            <Section id="cli" title="CLI">
              <P>
                The <Code>gogrid</Code> CLI is the primary interface for working with GoGrid
                projects. Define agents declaratively in <Code>gogrid.yaml</Code>, run them
                from the command line, and inspect execution traces and costs — all without
                writing Go code.
              </P>
              <H3>Installation</H3>
              <P>
                Build the CLI from source. The binary is written to <Code>bin/gogrid</Code>.
              </P>
              <CodeBlock
                code={`git clone https://github.com/lonestarx1/gogrid.git
cd gogrid
make build

# Verify the installation
bin/gogrid version
# gogrid dev (darwin/arm64, go1.25.4)

# Optional: embed a version string
make build VERSION=1.0.0`}
                filename="terminal"
              />
              <H3>Quick Start</H3>
              <P>
                Scaffold a project, set an API key, and run your first agent in under a minute.
              </P>
              <CodeBlock
                code={`# Scaffold a new project
gogrid init --template single my-agent
cd my-agent

# Set up the project
go mod init github.com/example/my-agent
go mod tidy
export OPENAI_API_KEY=sk-proj-...

# Run your agent
gogrid run assistant -input "Explain Go's concurrency model"`}
                filename="terminal"
              />
              <H3>Project Templates</H3>
              <P>
                Three templates are available for <Code>gogrid init</Code>, each generating
                a working project with <Code>gogrid.yaml</Code>, <Code>main.go</Code>,
                <Code>Makefile</Code>, and <Code>README.md</Code>.
              </P>
              <div className="space-y-4 mb-6">
                <Card title="single" desc="A single agent with instructions and configuration. The simplest starting point for any GoGrid project." />
                <Card title="team" desc="Two agents (researcher + reviewer) collaborating via shared memory and consensus. Demonstrates multi-agent coordination." />
                <Card title="pipeline" desc="Two sequential stages (drafter + editor) with state transfer between stages. Demonstrates linear workflows." />
              </div>
              <CodeBlock
                code={`# Scaffold each template type
gogrid init --template single my-agent
gogrid init --template team my-research-team
gogrid init --template pipeline my-content-pipeline

# Use a custom name
gogrid init --template team --name research-bot ./research`}
                filename="terminal"
              />
            </Section>

            {/* CLI Configuration */}
            <Section id="cli-config" title="Configuration (YAML)">
              <P>
                GoGrid projects are configured via a <Code>gogrid.yaml</Code> file in the
                project root. Each agent is defined with a model, provider, system prompt,
                and execution parameters.
              </P>
              <H3>Schema</H3>
              <CodeBlock
                code={`version: "1"                    # Required. Config schema version.

agents:
  <agent-name>:                 # Unique agent identifier.
    model: <string>             # Required. LLM model ID.
    provider: <string>          # Required. One of: openai, anthropic, gemini.
    instructions: <string>      # System prompt for the agent.
    config:
      max_turns: <int>          # Max LLM round-trips. 0 = unlimited.
      max_tokens: <int>         # Max response tokens per turn.
      temperature: <float>      # LLM randomness (0.0-1.0). Omit for default.
      timeout: <duration>       # Wall-clock limit (e.g. "30s", "5m", "1h").
      cost_budget: <float>      # Max cost in USD for a single run.`}
                filename="gogrid.yaml schema"
              />
              <H3>Full Example</H3>
              <CodeBlock
                code={`version: "1"

agents:
  researcher:
    model: claude-sonnet-4-5-20250929
    provider: anthropic
    instructions: |
      You are a technical researcher. When given a topic:
      1. Explain the core concepts clearly
      2. Provide concrete examples
      3. Discuss trade-offs and alternatives
      4. Mention common pitfalls
    config:
      max_turns: 10
      max_tokens: 4096
      temperature: 0.7
      timeout: 2m
      cost_budget: 0.50

  code-reviewer:
    model: gpt-4o-mini
    provider: openai
    instructions: |
      You are a senior Go code reviewer. Check for correctness,
      error handling, naming, and Go idioms.
    config:
      max_turns: 5
      max_tokens: 2048
      timeout: 60s
      cost_budget: 0.10

  summarizer:
    model: gpt-4o-mini
    provider: openai
    instructions: |
      Condense the input into 3-5 bullet points. Under 200 words.
    config:
      max_turns: 3
      max_tokens: 1024
      timeout: 30s
      cost_budget: 0.05`}
                filename="gogrid.yaml"
              />
              <H3>Environment Variable Substitution</H3>
              <P>
                Config values support <Code>{"\u0024{VAR}"}</Code> and <Code>{"\u0024{VAR:-default}"}</Code> syntax.
                This is processed before YAML parsing, so you can override any value per
                environment without editing the config file.
              </P>
              <CodeBlock
                code={`version: "1"

agents:
  assistant:
    model: \${MODEL:-gpt-4o-mini}
    provider: \${PROVIDER:-openai}
    instructions: \${AGENT_INSTRUCTIONS:-You are a helpful assistant.}
    config:
      max_turns: 10
      timeout: \${TIMEOUT:-60s}`}
                filename="gogrid.yaml with env vars"
              />
              <CodeBlock
                code={`# Override model and provider at runtime
MODEL=claude-sonnet-4-5-20250929 PROVIDER=anthropic gogrid run assistant -input "Hello"

# Use a different model for cost savings
MODEL=gpt-4o-mini gogrid run assistant -input "Quick question"`}
                filename="terminal"
              />
              <H3>Environment Variables</H3>
              <P>
                API keys are resolved from environment variables — never stored in config files.
              </P>
              <div className="space-y-4 mb-6">
                <Card title="OPENAI_API_KEY" desc="API key for OpenAI models (gpt-4o, gpt-4o-mini, gpt-4.1, o3, etc.)" />
                <Card title="ANTHROPIC_API_KEY" desc="API key for Anthropic models (claude-sonnet-4-5, claude-opus-4-6, etc.)" />
                <Card title="GEMINI_API_KEY" desc="API key for Google Gemini models (gemini-2.5-pro, gemini-2.5-flash, etc.)" />
              </div>
              <H3>Validation</H3>
              <P>
                The config is validated on load. The CLI will report clear errors for
                missing fields, invalid providers, or malformed YAML.
              </P>
              <ol className="list-decimal list-inside space-y-2 text-text-muted mb-6">
                <li><Code>version</Code> must be <Code>&quot;1&quot;</Code></li>
                <li>At least one agent must be defined</li>
                <li>Each agent must have <Code>model</Code> and <Code>provider</Code></li>
                <li><Code>provider</Code> must be one of: <Code>openai</Code>, <Code>anthropic</Code>, <Code>gemini</Code></li>
              </ol>
            </Section>

            {/* CLI Commands */}
            <Section id="cli-commands" title="CLI Commands">
              <H3>gogrid init</H3>
              <P>
                Scaffold a new GoGrid project from a template. Creates a directory with
                <Code>gogrid.yaml</Code>, <Code>main.go</Code>, <Code>Makefile</Code>, and <Code>README.md</Code>.
              </P>
              <CodeBlock
                code={`gogrid init [flags] [directory]

Flags:
  -template string   Project template: single, team, pipeline (default "single")
  -name string       Project name (defaults to directory name)`}
                filename="gogrid init"
              />
              <CodeBlock
                code={`# Scaffold a single-agent project
$ gogrid init --template single my-agent

Created GoGrid project in my-agent/
  gogrid.yaml   Agent configuration
  main.go       Programmatic entry point
  Makefile      Build targets
  README.md     Setup instructions

Next steps:
  cd my-agent
  go mod init github.com/example/my-agent
  go mod tidy
  export OPENAI_API_KEY=sk-...
  gogrid run assistant -input "Hello!"`}
                filename="terminal"
              />

              <H3>gogrid list</H3>
              <P>
                List all agents defined in the project&apos;s <Code>gogrid.yaml</Code>.
              </P>
              <CodeBlock
                code={`$ gogrid list
NAME             PROVIDER    MODEL
code-reviewer    openai      gpt-4o-mini
researcher       anthropic   claude-sonnet-4-5-20250929
summarizer       openai      gpt-4o-mini`}
                filename="terminal"
              />

              <H3>gogrid run</H3>
              <P>
                Execute a named agent with the given input. The agent&apos;s response is printed
                to stdout. A run record is saved for later inspection with <Code>gogrid trace</Code> and <Code>gogrid cost</Code>.
              </P>
              <CodeBlock
                code={`gogrid run <agent-name> [flags]

Flags:
  -config string    Path to config file (default "gogrid.yaml")
  -input string     Input text to send to the agent (required)
  -timeout string   Override the agent's timeout (e.g. "30s", "5m")`}
                filename="gogrid run"
              />
              <CodeBlock
                code={`# Run an agent
$ gogrid run researcher -input "Explain Go's context package"
Go's context package provides a way to carry deadlines, cancellation signals,
and request-scoped values across API boundaries...

# The run ID is printed to stderr for later inspection
Run ID: 019479a3c4e80001

# Override timeout
$ gogrid run summarizer -input "Summarize this paper..." -timeout 2m`}
                filename="terminal"
              />
              <P>
                What happens during a run:
              </P>
              <ol className="list-decimal list-inside space-y-2 text-text-muted mb-6">
                <li>Loads and validates <Code>gogrid.yaml</Code></li>
                <li>Looks up the agent by name</li>
                <li>Resolves the LLM provider using environment variables</li>
                <li>Creates the agent with configured model, instructions, and parameters</li>
                <li>Calls <Code>agent.Run()</Code> with an in-memory tracer</li>
                <li>Prints the response to stdout</li>
                <li>Saves the run record to <Code>.gogrid/runs/</Code></li>
              </ol>

              <H3>gogrid trace</H3>
              <P>
                Inspect execution traces. With no arguments, lists recent runs. With a run ID,
                renders the span tree showing the full execution flow.
              </P>
              <CodeBlock
                code={`# List recent runs
$ gogrid trace
Recent runs:
  019479a3c4e80001  researcher     claude-sonnet-4-5-20250929  4.2s
  019479a1b2c70002  code-reviewer  gpt-4o-mini                1.1s
  019479a0a1b60003  summarizer     gpt-4o-mini                0.8s

# View span tree for a specific run
$ gogrid trace 019479a3c4e80001
Run: 019479a3c4e80001
Agent: researcher | Model: claude-sonnet-4-5-20250929 | Duration: 4.2s

agent.run (4.2s)
├── memory.load (1ms)
├── llm.complete (2.1s) [prompt: 150, completion: 89]
├── llm.complete (1.8s) [prompt: 280, completion: 145]
└── memory.save (2ms)

# Export as JSON for programmatic use
$ gogrid trace 019479a3c4e80001 -json | jq '.[].name'`}
                filename="terminal"
              />

              <H3>gogrid cost</H3>
              <P>
                View cost breakdown for agent runs. With no arguments, lists all runs with
                their total cost. With a run ID, shows a per-model cost breakdown.
              </P>
              <CodeBlock
                code={`# List all runs with costs
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

# Export as JSON
$ gogrid cost -json`}
                filename="terminal"
              />

              <H3>gogrid version</H3>
              <CodeBlock
                code={`$ gogrid version
gogrid 1.0.0 (darwin/arm64, go1.25.4)`}
                filename="terminal"
              />

              <H3>Supported Models</H3>
              <P>
                GoGrid includes built-in pricing for cost tracking. Any model string is
                accepted — these have pre-configured pricing:
              </P>
              <div className="space-y-4 mb-6">
                <Card title="OpenAI" desc="gpt-4o, gpt-4o-mini, gpt-4.1, gpt-4.1-mini, gpt-4.1-nano, o3, o4-mini" />
                <Card title="Anthropic" desc="claude-opus-4-6, claude-opus-4-5, claude-sonnet-4-5, claude-sonnet-4-0, claude-haiku-4-5" />
                <Card title="Google Gemini" desc="gemini-3-pro, gemini-3-flash, gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash" />
              </div>
              <P>
                Models not in this list work fine — cost tracking will report $0.00 until
                custom pricing is configured via the Go API.
              </P>
            </Section>

            {/* Run Records */}
            <Section id="cli-run-records" title="Run Records">
              <P>
                Every <Code>gogrid run</Code> invocation saves a JSON record to
                <Code>.gogrid/runs/&lt;run-id&gt;.json</Code>. Run IDs are time-sortable, so
                newer runs always sort after older ones.
              </P>
              <H3>Record Fields</H3>
              <div className="space-y-4 mb-6">
                <Card title="run_id" desc="Unique, time-sortable identifier for the run." />
                <Card title="agent / model / provider" desc="Which agent ran, which model and provider were used." />
                <Card title="input / output" desc="The user's input and the agent's final response." />
                <Card title="turns / usage" desc="Number of LLM round-trips and token counts (prompt, completion, total)." />
                <Card title="cost" desc="Estimated cost in USD based on built-in model pricing." />
                <Card title="spans" desc="Execution trace spans — LLM calls, tool executions, memory operations." />
                <Card title="duration / error" desc="Wall-clock duration and error message if the run failed." />
              </div>
              <H3>Inspecting Run Records</H3>
              <P>
                Run records are plain JSON files. You can inspect them directly, back them up,
                or pipe them to other tools.
              </P>
              <CodeBlock
                code={`# List run records
ls .gogrid/runs/

# View a raw record
cat .gogrid/runs/019479a3c4e80001.json | jq .

# Extract agent names and costs from all runs
for f in .gogrid/runs/*.json; do
  echo "$(jq -r '.agent' $f): $(jq -r '.cost' $f)"
done

# Total cost across all runs
ls .gogrid/runs/*.json | xargs -I{} jq '.cost' {} | paste -sd+ | bc`}
                filename="terminal"
              />
              <P>
                Add <Code>.gogrid/</Code> to your <Code>.gitignore</Code> — run records are
                local development artifacts.
              </P>
              <H3>Typical Workflow</H3>
              <CodeBlock
                code={`# 1. Scaffold a project
gogrid init --template single my-project
cd my-project

# 2. Set up Go module and dependencies
go mod init github.com/example/my-project
go mod tidy

# 3. Set your API key
export OPENAI_API_KEY=sk-proj-...

# 4. List agents
gogrid list

# 5. Run an agent
gogrid run assistant -input "Explain the CAP theorem"

# 6. Inspect the trace
gogrid trace          # list recent runs, copy a run ID
gogrid trace <run-id> # view execution span tree

# 7. Check costs
gogrid cost <run-id>  # detailed breakdown
gogrid cost           # summary of all runs`}
                filename="terminal"
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

            {/* Pipeline */}
            <Section id="pipeline" title="Pipeline (Linear)">
              <P>
                The Pipeline pattern chains agents sequentially — each stage processes
                input, produces output, and transfers state ownership to the next stage.
                Previous stages lose access to the data once ownership is transferred.
              </P>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/orchestrator/pipeline"

p := pipeline.New("research-pipeline",
    pipeline.WithStages(
        pipeline.Stage{Name: "collect", Agent: collector},
        pipeline.Stage{Name: "analyze", Agent: analyzer},
        pipeline.Stage{Name: "summarize", Agent: summarizer},
    ),
    pipeline.WithConfig(pipeline.Config{
        Timeout:    5 * time.Minute,
        CostBudget: 2.00,
    }),
)

result, err := p.Run(ctx, "Research the impact of AI on healthcare")
fmt.Println(result.Output) // Final stage's output
fmt.Printf("Stages: %d, Cost: $%.4f\\n", len(result.Stages), result.TotalCost)`}
                filename="pipeline"
              />
              <H3>State Ownership Transfer</H3>
              <P>
                Pipelines integrate with GoGrid&apos;s <Code>memory/transfer</Code> package.
                Each stage gets an owned handle — when state transfers to the next stage,
                the previous handle is invalidated. The audit trail records every transfer.
              </P>
              <CodeBlock
                code={`// The transfer log shows ownership history
for _, entry := range result.TransferLog {
    fmt.Printf("%s -> %s (generation %d)\\n",
        entry.From, entry.To, entry.Generation)
}
// Output:
//  -> collect (generation 1)
// collect -> analyze (generation 2)
// analyze -> summarize (generation 3)`}
                filename="transfer log"
              />
            </Section>

            {/* Pipeline Stages */}
            <Section id="pipeline-stages" title="Pipeline Stages">
              <P>
                Each stage can optionally transform its input and validate its output.
              </P>
              <CodeBlock
                code={`pipeline.Stage{
    Name:  "analyze",
    Agent: analyzer,
    // Transform the previous stage's output before passing to this agent
    InputTransform: func(input string) string {
        return "Analyze the following data:\\n\\n" + input
    },
    // Validate the output before proceeding to the next stage
    OutputValidate: func(output string) error {
        if len(output) < 100 {
            return errors.New("analysis too short")
        }
        return nil
    },
    // Per-stage timeout and cost budget
    Timeout:    30 * time.Second,
    CostBudget: 0.50,
}`}
                filename="stage options"
              />
              <H3>Progress Reporting</H3>
              <P>
                Track pipeline progress with a callback function.
              </P>
              <CodeBlock
                code={`p := pipeline.New("tracked",
    pipeline.WithStages(...),
    pipeline.WithProgress(func(idx, total int, sr pipeline.StageResult) {
        fmt.Printf("[%d/%d] Stage %q completed\\n", idx+1, total, sr.Name)
    }),
)`}
                filename="progress"
              />
            </Section>

            {/* Pipeline Retry & Error Handling */}
            <Section id="pipeline-retry" title="Retry & Error Handling">
              <P>
                Stages support configurable retry policies and error actions.
              </P>
              <CodeBlock
                code={`// Retry up to 3 times with a delay between attempts
pipeline.Stage{
    Name:  "flaky-api",
    Agent: apiAgent,
    Retry: pipeline.RetryPolicy{
        MaxAttempts: 3,
        Delay:       2 * time.Second,
    },
}

// Skip on failure instead of aborting the pipeline
pipeline.Stage{
    Name:    "optional-enrichment",
    Agent:   enrichAgent,
    OnError: pipeline.Skip,
}

// Default: abort the pipeline on failure
pipeline.Stage{
    Name:    "critical-step",
    Agent:   criticalAgent,
    OnError: pipeline.Abort, // default behavior
}`}
                filename="retry and error handling"
              />
              <H3>Error Actions</H3>
              <div className="space-y-4 mb-6">
                <Card title="Abort" desc="Stop the pipeline and return the error. This is the default behavior." />
                <Card title="Skip" desc="Mark the stage as skipped and continue with the previous stage's output. The pipeline proceeds to the next stage." />
              </div>
              <P>
                When a stage with retry fails all attempts and has <Code>OnError: Skip</Code>,
                the stage is skipped after exhausting retries. The <Code>StageResult.Attempts</Code> field
                records how many times the stage was executed.
              </P>
            </Section>

            {/* Graph */}
            <Section id="graph" title="Graph">
              <P>
                The Graph pattern extends pipelines with conditional branches, parallel
                execution (fan-out/fan-in), and loops. Nodes wrap agents and edges connect
                them with optional condition functions.
              </P>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/orchestrator/graph"

g, err := graph.NewBuilder("review-pipeline").
    AddNode("draft", draftAgent).
    AddNode("review", reviewAgent).
    AddNode("publish", publishAgent).
    AddEdge("draft", "review").
    AddEdge("review", "publish").
    Options(
        graph.WithConfig(graph.Config{
            MaxIterations: 10,
            Timeout:       5 * time.Minute,
            CostBudget:    3.00,
        }),
        graph.WithTracer(tracer),
    ).
    Build()

result, err := g.Run(ctx, "Write about AI agents")
fmt.Println(result.Output)
fmt.Printf("Nodes: %d, Cost: $%.4f\\n", len(result.NodeResults), result.TotalCost)`}
                filename="graph"
              />
              <H3>Fan-Out / Fan-In</H3>
              <P>
                Independent branches run concurrently. When multiple edges point to the
                same node, it waits for all incoming sources to complete before running.
              </P>
              <CodeBlock
                code={`// Diamond pattern: a -> b, a -> c, b -> d, c -> d
// b and c run in parallel; d waits for both before running
g, _ := graph.NewBuilder("diamond").
    AddNode("split", splitAgent).
    AddNode("path-a", agentA).
    AddNode("path-b", agentB).
    AddNode("merge", mergeAgent).
    AddEdge("split", "path-a").
    AddEdge("split", "path-b").
    AddEdge("path-a", "merge").
    AddEdge("path-b", "merge").
    Build()`}
                filename="fan-out/fan-in"
              />
            </Section>

            {/* Graph Builder */}
            <Section id="graph-builder" title="Graph Builder">
              <P>
                The fluent builder API validates the graph at build time — duplicate nodes,
                missing edge targets, and other structural errors are caught before execution.
              </P>
              <CodeBlock
                code={`b := graph.NewBuilder("my-graph")

// Add nodes (each wraps an agent)
b.AddNode("research", researchAgent)
b.AddNode("analyze", analyzeAgent)
b.AddNode("report", reportAgent)

// Add edges (optionally with conditions)
b.AddEdge("research", "analyze")
b.AddEdge("analyze", "report")

// Apply options
b.Options(graph.WithTracer(tracer))

// Build validates and returns the graph
g, err := b.Build()
if err != nil {
    log.Fatal(err) // e.g., "edge references unknown node"
}`}
                filename="builder"
              />
              <H3>DOT Export</H3>
              <P>
                Export the graph to Graphviz DOT format for visualization.
              </P>
              <CodeBlock
                code={`dot := g.DOT()
// Output:
// digraph "my-graph" {
//   rankdir=LR;
//   node [shape=box, style=rounded];
//   "research";
//   "analyze";
//   "report";
//   "research" -> "analyze";
//   "analyze" -> "report";
// }`}
                filename="DOT export"
              />
            </Section>

            {/* Graph Advanced */}
            <Section id="graph-advanced" title="Loops & Conditions">
              <P>
                Conditional edges use <Code>graph.When</Code> to control routing based on
                a node&apos;s output. Backward edges create loops with a configurable
                max iteration guard.
              </P>
              <CodeBlock
                code={`g, _ := graph.NewBuilder("review-loop").
    AddNode("draft", draftAgent).
    AddNode("review", reviewAgent).
    AddNode("revise", reviseAgent).
    AddNode("publish", publishAgent).
    AddEdge("draft", "review").
    // Conditional: route based on review output
    AddEdge("review", "revise", graph.When(func(out string) bool {
        return strings.Contains(out, "needs revision")
    })).
    AddEdge("review", "publish", graph.When(func(out string) bool {
        return strings.Contains(out, "approved")
    })).
    // Loop back: revise -> review
    AddEdge("revise", "review").
    Options(graph.WithConfig(graph.Config{
        MaxIterations: 5, // prevent infinite loops
    })).
    Build()`}
                filename="loops and conditions"
              />
              <H3>Edge Helpers</H3>
              <CodeBlock
                code={`// When wraps a simple output check
graph.When(func(output string) bool {
    return strings.Contains(output, "approved")
})

// Always is an unconditional edge (same as no condition)
graph.Always()`}
                filename="edge helpers"
              />
            </Section>

            {/* Dynamic Orchestration */}
            <Section id="dynamic" title="Dynamic Orchestration">
              <P>
                Dynamic Orchestration is GoGrid&apos;s most powerful pattern. A <Code>Runtime</Code> enables
                agents to spawn child agents, teams, pipelines, or graphs at runtime — the executing
                agent decides which orchestration to use based on the problem at hand.
              </P>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/orchestrator/dynamic"

rt := dynamic.New("coordinator",
    dynamic.WithConfig(dynamic.Config{
        MaxConcurrent: 5,
        MaxDepth:      3,
        CostBudget:    2.00,
    }),
    dynamic.WithTracer(tracer),
)

// Embed the runtime in context for child access
ctx := rt.Context(ctx)

// Spawn any orchestration pattern as a child
agentResult, _ := rt.SpawnAgent(ctx, researchAgent, "Find papers on X")
teamResult, _  := rt.SpawnTeam(ctx, reviewTeam, agentResult.Message.Content)
pipeResult, _  := rt.SpawnPipeline(ctx, summarizePipeline, teamResult.Decision.Content)
graphResult, _ := rt.SpawnGraph(ctx, publishGraph, pipeResult.Output)

// Aggregate metrics across all children
res := rt.Result()
fmt.Printf("Children: %d, Cost: $%.4f\\n", len(res.Children), res.TotalCost)`}
                filename="dynamic orchestration"
              />
              <H3>Spawning Children</H3>
              <P>
                Four spawn methods correspond to GoGrid&apos;s four orchestration patterns. Each
                blocks until the child completes, inherits the parent&apos;s tracing context,
                and records cost/usage metrics.
              </P>
              <CodeBlock
                code={`// Spawn a single agent
result, err := rt.SpawnAgent(ctx, agent, "input")

// Spawn a team
result, err := rt.SpawnTeam(ctx, team, "discuss this")

// Spawn a pipeline
result, err := rt.SpawnPipeline(ctx, pipeline, "process this")

// Spawn a graph
result, err := rt.SpawnGraph(ctx, graph, "route this")`}
                filename="spawn methods"
              />
              <H3>Context Propagation</H3>
              <P>
                The runtime is stored in context so nested orchestrations can dynamically
                spawn further children up to the configured depth limit.
              </P>
              <CodeBlock
                code={`// Retrieve runtime from context (e.g., inside a tool)
rt := dynamic.FromContext(ctx)
if rt != nil {
    // Spawn a sub-task from within a tool execution
    result, _ := rt.SpawnAgent(ctx, helperAgent, "assist with this")
}

// Check current nesting depth
depth := dynamic.DepthFromContext(ctx)`}
                filename="context propagation"
              />
            </Section>

            {/* Resource Governance */}
            <Section id="dynamic-governance" title="Resource Governance">
              <P>
                The runtime enforces resource limits to prevent runaway costs, infinite recursion,
                and resource exhaustion.
              </P>
              <div className="space-y-4 mb-6">
                <Card title="MaxConcurrent" desc="Maximum number of children executing simultaneously. Uses a semaphore — excess spawns block until a slot is available." />
                <Card title="MaxDepth" desc="Maximum nesting depth for recursive spawning. Prevents infinite recursion when children spawn further children. Defaults to 10." />
                <Card title="CostBudget" desc="Maximum total cost in USD across all children. New spawns are rejected once the budget is exhausted." />
              </div>
              <CodeBlock
                code={`rt := dynamic.New("governed",
    dynamic.WithConfig(dynamic.Config{
        MaxConcurrent: 3,  // at most 3 children running at once
        MaxDepth:      4,  // max 4 levels of nesting
        CostBudget:    1.00, // $1.00 total across all children
    }),
)

// Check remaining budget before expensive operations
remaining := rt.RemainingBudget() // -1 if unlimited`}
                filename="resource governance"
              />
              <H3>Cascading Cancellation</H3>
              <P>
                All children use derived contexts. Canceling the parent context automatically
                cancels all running children — no orphaned goroutines or wasted LLM calls.
              </P>
              <CodeBlock
                code={`ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()

ctx = rt.Context(ctx)

// All spawned children will be canceled if the 30s timeout fires
rt.SpawnAgent(ctx, slowAgent, "this might take a while")`}
                filename="cascading cancellation"
              />
            </Section>

            {/* Async & Futures */}
            <Section id="dynamic-async" title="Async & Futures">
              <P>
                Use <Code>Go</Code> to launch children in the background. Returns a <Code>Future</Code> that
                can be awaited or polled.
              </P>
              <CodeBlock
                code={`// Launch multiple children concurrently
f1 := rt.Go(ctx, "research", func(ctx context.Context) (string, error) {
    r, err := rt.SpawnAgent(ctx, researchAgent, "find papers")
    if err != nil {
        return "", err
    }
    return r.Message.Content, nil
})

f2 := rt.Go(ctx, "analyze", func(ctx context.Context) (string, error) {
    r, err := rt.SpawnAgent(ctx, analysisAgent, "analyze trends")
    if err != nil {
        return "", err
    }
    return r.Message.Content, nil
})

// Await results
research, _ := f1.Wait(ctx)
analysis, _ := f2.Wait(ctx)

// Or wait for all background children at once
rt.Wait()`}
                filename="async spawning"
              />
              <H3>Future API</H3>
              <CodeBlock
                code={`// Wait blocks until the future completes or context is canceled
output, err := future.Wait(ctx)

// Done returns a channel that closes when the future completes
select {
case <-future.Done():
    output, err := future.Wait(ctx) // returns immediately
case <-ctx.Done():
    // timeout
}`}
                filename="future API"
              />
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
              <H3>Pipeline Trace Spans</H3>
              <P>
                Pipeline execution produces a trace tree with pipeline and per-stage spans.
              </P>
              <CodeBlock
                code={`// Trace tree for a pipeline run:
// pipeline.run
//   ├─ pipeline.stage (collect)
//   │   ├─ agent.run
//   │   │   ├─ memory.load
//   │   │   ├─ llm.complete
//   │   │   └─ memory.save
//   ├─ pipeline.stage (analyze)
//   │   └─ agent.run ...
//   └─ pipeline.stage (summarize)
//       └─ agent.run ...`}
                filename="pipeline trace tree"
              />
              <H3>Graph Trace Spans</H3>
              <CodeBlock
                code={`// Trace tree for a graph run:
// graph.run
//   ├─ graph.node (draft, iteration=1)
//   │   └─ agent.run ...
//   ├─ graph.node (review, iteration=1)
//   │   └─ agent.run ...
//   ├─ graph.node (revise, iteration=1)
//   │   └─ agent.run ...
//   ├─ graph.node (review, iteration=2)
//   │   └─ agent.run ...
//   └─ graph.node (publish, iteration=1)
//       └─ agent.run ...`}
                filename="graph trace tree"
              />
              <H3>Dynamic Trace Spans</H3>
              <CodeBlock
                code={`// Trace tree for dynamic orchestration:
// dynamic.spawn_agent (research)
//   └─ agent.run ...
// dynamic.spawn_team (debate)
//   └─ team.run ...
// dynamic.go (background-task)
//   └─ dynamic.spawn_pipeline (summarize)
//       └─ pipeline.run ...`}
                filename="dynamic trace tree"
              />
            </Section>

            {/* OpenTelemetry Export */}
            <Section id="otel" title="OpenTelemetry Export">
              <P>
                The OTLP exporter sends GoGrid trace spans to any OpenTelemetry-compatible
                backend — Jaeger, Zipkin, Grafana Tempo, and others. Built entirely with
                the Go standard library — no external OTel SDK required.
              </P>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/trace/otel"

exporter := otel.NewExporter(
    otel.WithEndpoint("http://localhost:4318/v1/traces"),
    otel.WithServiceName("my-agent-service"),
    otel.WithServiceVersion("1.0.0"),
    otel.WithBatchSize(100),
    otel.WithFlushInterval(5 * time.Second),
)
defer exporter.Shutdown()

// Use as any tracer
a := agent.New("assistant",
    agent.WithTracer(exporter),
    agent.WithProvider(provider),
    agent.WithModel("gpt-4o"),
)
result, _ := a.Run(ctx, "hello")`}
                filename="OTLP exporter"
              />
              <H3>Batching & Flushing</H3>
              <P>
                Spans are batched in memory and flushed periodically or when the batch
                size is reached. <Code>Shutdown</Code> sends all remaining spans.
              </P>
              <CodeBlock
                code={`// Spans are sent as OTLP JSON over HTTP POST
// Content-Type: application/json
// Payload follows the OTLP JSON span format with:
//   - resourceSpans[].resource.attributes: service.name, service.version
//   - scopeSpans[].spans[]: traceId, spanId, parentSpanId, name, kind, timestamps
//   - Attributes prefixed with "gogrid." (e.g., gogrid.agent.name)
//   - exception.message for error spans`}
                filename="OTLP format"
              />
              <H3>Semantic Conventions</H3>
              <P>
                GoGrid spans follow semantic conventions for attribute naming.
              </P>
              <div className="space-y-4 mb-6">
                <Card title="gogrid.agent.name" desc="The agent's name. Attached to agent.run spans." />
                <Card title="gogrid.llm.model" desc="The LLM model used. Attached to llm.complete spans." />
                <Card title="gogrid.tool.name" desc="The tool's name. Attached to tool.execute spans." />
                <Card title="gogrid.cost.usd" desc="Cost in USD. Attached to agent.run spans." />
              </div>
            </Section>

            {/* Structured Logging */}
            <Section id="structured-logging" title="Structured Logging">
              <P>
                The <Code>trace/log</Code> package provides structured JSON logging with
                automatic trace correlation. When a span exists in the context, the logger
                includes <Code>trace_id</Code> and <Code>span_id</Code> in every log line.
              </P>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/trace/log"

logger := log.New(os.Stdout, log.Info)

// Basic logging
logger.Info("agent started", "agent", "researcher", "model", "gpt-4o")

// Context-aware logging with trace correlation
logger.InfoCtx(ctx, "LLM call complete", "tokens", "150", "model", "gpt-4o")

// Output (JSON):
// {"level":"info","time":"2026-02-16T12:00:00Z","msg":"LLM call complete",
//  "trace_id":"abc123","span_id":"def456",
//  "fields":{"tokens":"150","model":"gpt-4o"}}`}
                filename="structured logging"
              />
              <H3>Log Levels</H3>
              <div className="space-y-4 mb-6">
                <Card title="Debug" desc="Most verbose. Use for detailed diagnostic information." />
                <Card title="Info" desc="Default level. Normal operational messages." />
                <Card title="Warn" desc="Potential issues that don't prevent operation." />
                <Card title="Error" desc="Failures that need attention." />
              </div>
              <H3>File Logging with Rotation</H3>
              <P>
                <Code>FileWriter</Code> writes to disk with automatic size-based rotation.
                Configurable max file size and number of rotated files to keep.
              </P>
              <CodeBlock
                code={`fw, err := log.NewFileWriter("/var/log/gogrid.log", log.FileConfig{
    MaxSize:  10 * 1024 * 1024, // 10 MB per file
    MaxFiles: 5,                // keep 5 rotated files
})
if err != nil {
    panic(err)
}
defer fw.Close()

logger := log.New(fw, log.Debug)
logger.Info("agent started")
// Rotated files: gogrid.log, gogrid.log.1, gogrid.log.2, ...`}
                filename="file logging"
              />
            </Section>

            {/* Metrics */}
            <Section id="metrics" title="Metrics">
              <P>
                GoGrid provides Prometheus-compatible metrics with no external dependencies.
                The <Code>metrics.Collector</Code> wraps any tracer and automatically populates
                counters, gauges, and histograms from trace spans.
              </P>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/trace/metrics"

reg := metrics.NewRegistry()
collector := metrics.NewCollector(innerTracer, reg)

// Use collector as the tracer — metrics are recorded automatically
a := agent.New("assistant",
    agent.WithTracer(collector),
    agent.WithProvider(provider),
    agent.WithModel("gpt-4o"),
)

// Expose metrics for Prometheus scraping
http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; version=0.0.4")
    fmt.Fprint(w, reg.Export())
})`}
                filename="metrics"
              />
              <H3>Auto-Collected Metrics</H3>
              <div className="space-y-4 mb-6">
                <Card title="gogrid_agent_runs_total" desc="Total agent runs, labeled by agent name and status (ok/error)." />
                <Card title="gogrid_agent_run_duration_seconds" desc="Histogram of agent run durations." />
                <Card title="gogrid_llm_calls_total" desc="Total LLM calls, labeled by model and status." />
                <Card title="gogrid_llm_call_duration_seconds" desc="Histogram of LLM call latencies." />
                <Card title="gogrid_llm_tokens_total" desc="Total tokens consumed, labeled by model and type (prompt/completion)." />
                <Card title="gogrid_tool_executions_total" desc="Total tool executions, labeled by tool name and status." />
                <Card title="gogrid_tool_execution_duration_seconds" desc="Histogram of tool execution durations." />
                <Card title="gogrid_memory_operations_total" desc="Total memory operations, labeled by operation type (load/save)." />
                <Card title="gogrid_cost_usd_total" desc="Total cost in USD, labeled by agent and model." />
              </div>
              <H3>Custom Metrics</H3>
              <P>
                Create your own counters, gauges, and histograms for application-specific metrics.
              </P>
              <CodeBlock
                code={`reg := metrics.NewRegistry()

// Counters
requests := reg.Counter("http_requests_total", "Total HTTP requests")
requests.Inc(map[string]string{"method": "GET", "status": "200"})

// Gauges
connections := reg.Gauge("active_connections", "Active connections")
connections.Set(42, map[string]string{"server": "web-1"})

// Histograms (custom buckets)
latency := reg.Histogram("request_duration_seconds", "Request latency",
    0.01, 0.05, 0.1, 0.5, 1, 5)
latency.Observe(0.123, map[string]string{"handler": "api"})`}
                filename="custom metrics"
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

            {/* Cost Governance */}
            <Section id="cost-governance" title="Cost Governance">
              <P>
                Advanced cost governance features for production deployments. Set budgets
                with threshold alerts, allocate costs to specific entities, and generate
                aggregate reports.
              </P>
              <H3>Budget Alerts</H3>
              <P>
                Register a callback that fires when cost crosses threshold fractions of
                a configured budget. Each threshold fires at most once.
              </P>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/cost"

tracker := cost.NewTracker()
tracker.SetBudget(10.00) // $10 budget

tracker.OnBudgetThreshold(func(threshold, current float64) {
    fmt.Printf("ALERT: %.0f%% of budget reached ($%.2f)\\n",
        threshold*100, current)
}, 0.5, 0.8, 1.0) // alerts at 50%, 80%, 100%

// As costs accumulate, alerts fire automatically
tracker.Add("gpt-4o", usage) // triggers 50% alert at $5`}
                filename="budget alerts"
              />
              <H3>Cost Allocation</H3>
              <P>
                Attribute costs to specific entities — agents, teams, pipelines, or any
                named component. Track which parts of your system spend the most.
              </P>
              <CodeBlock
                code={`// Attribute costs to specific entities
tracker.AddForEntity("gpt-4o", "research-agent", usage1)
tracker.AddForEntity("gpt-4o", "summary-agent", usage2)
tracker.AddForEntity("gpt-4o-mini", "research-agent", usage3)

// Query per-entity costs
researchCost := tracker.EntityCost("research-agent")
summaryCost := tracker.EntityCost("summary-agent")

fmt.Printf("Research: $%.4f, Summary: $%.4f\\n", researchCost, summaryCost)`}
                filename="cost allocation"
              />
              <H3>Cost Reports</H3>
              <P>
                Generate aggregate reports breaking down costs by model and entity.
              </P>
              <CodeBlock
                code={`report := tracker.Report()
fmt.Printf("Total: $%.4f across %d calls\\n", report.TotalCost, report.RecordCount)

// By model
for model, mr := range report.ByModel {
    fmt.Printf("  %s: %d calls, $%.4f, %d tokens\\n",
        model, mr.Calls, mr.Cost, mr.Usage.TotalTokens)
}

// By entity
for entity, cost := range report.ByEntity {
    fmt.Printf("  %s: $%.4f\\n", entity, cost)
}`}
                filename="cost reports"
              />
            </Section>

            {/* Mock Provider */}
            <Section id="mock-provider" title="Mock Provider">
              <P>
                The <Code>pkg/llm/mock</Code> package provides a configurable mock LLM
                provider for testing GoGrid agents, teams, pipelines, and graphs without
                API keys.
              </P>
              <H3>Basic Usage</H3>
              <CodeBlock
                code={`import (
    "github.com/lonestarx1/gogrid/pkg/llm"
    "github.com/lonestarx1/gogrid/pkg/llm/mock"
)

// Fixed response for all calls.
provider := mock.New(mock.WithFallback(&llm.Response{
    Message: llm.NewAssistantMessage("mock response"),
    Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
    Model:   "mock",
}))

agent.New("my-agent", agent.WithProvider(provider), agent.WithModel("mock"))`}
                filename="mock provider"
              />
              <H3>Sequential Responses</H3>
              <P>
                Queue multiple responses for multi-turn conversations. The provider
                returns them in order, then falls back to the fallback response.
              </P>
              <CodeBlock
                code={`toolCallResp := &llm.Response{
    Message: llm.Message{
        Role:      llm.RoleAssistant,
        ToolCalls: []llm.ToolCall{{ID: "tc-1", Function: "search", Arguments: []byte(\`{"q":"test"}\`)}},
    },
    Usage: llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
    Model: "mock",
}
finalResp := &llm.Response{
    Message: llm.NewAssistantMessage("found it"),
    Usage:   llm.Usage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30},
    Model:   "mock",
}

provider := mock.New(
    mock.WithResponses(toolCallResp, finalResp),
    mock.WithFallback(finalResp),
)`}
                filename="sequential responses"
              />
              <H3>Error Injection & Latency</H3>
              <CodeBlock
                code={`// Always fail.
provider := mock.New(mock.WithError(errors.New("api unavailable")))

// Fail first 2 calls, then succeed.
provider = mock.New(
    mock.WithFailCount(2),
    mock.WithFallback(successResponse),
)

// Simulate 100ms latency (respects context cancellation).
provider = mock.New(
    mock.WithDelay(100 * time.Millisecond),
    mock.WithFallback(response),
)`}
                filename="error injection"
              />
              <H3>Call Recording</H3>
              <P>
                The mock provider records all calls for assertions in tests.
              </P>
              <CodeBlock
                code={`provider := mock.New(mock.WithFallback(response))
// ... run agent ...

fmt.Println(provider.Calls())    // Number of Complete calls
history := provider.History()     // All recorded call params
provider.Reset()                  // Clear history, keep config`}
                filename="call recording"
              />
            </Section>

            {/* Evaluation Framework */}
            <Section id="evaluation" title="Evaluation Framework">
              <P>
                The <Code>pkg/eval</Code> package provides composable evaluators for
                scoring agent outputs. Evaluators measure both output quality and
                operational metrics like cost and tool usage.
              </P>
              <H3>Evaluator Interface</H3>
              <CodeBlock
                code={`type Evaluator interface {
    Name() string
    Evaluate(ctx context.Context, result *agent.Result) (Score, error)
}

type Score struct {
    Pass   bool    // Primary signal: did the result meet criteria?
    Value  float64 // Normalized 0.0-1.0 for trend analysis
    Reason string  // Human-readable explanation
}`}
                filename="eval.go"
              />
              <H3>Built-in Evaluators</H3>
              <CodeBlock
                code={`import "github.com/lonestarx1/gogrid/pkg/eval"

// Exact string match.
eval.NewExactMatch("expected output")

// Substring containment (Value = fraction of substrings found).
eval.NewContains("Go", "concurrency", "goroutines")

// Cost budget check.
eval.NewCostWithin(0.05) // $0.05 USD max

// Tool usage expectations.
eval.NewToolUse(eval.ToolExpectation{
    Name:     "search",
    MinCalls: 1,
})

// LLM-as-judge (scores 0-10, passes at >= 7).
eval.NewLLMJudge(provider, "gpt-4o", "Rate clarity and accuracy.")`}
                filename="built-in evaluators"
              />
              <H3>Evaluation Suite</H3>
              <P>
                Compose multiple evaluators into a suite. The suite runs all evaluators
                and aggregates results. <Code>SuiteResult.Pass</Code> is true only if
                every evaluator passed.
              </P>
              <CodeBlock
                code={`suite := eval.NewSuite(
    eval.NewContains("Go", "compiled"),
    eval.NewCostWithin(0.05),
    eval.NewFunc("min_length", func(_ context.Context, r *agent.Result) (eval.Score, error) {
        if len(r.Message.Content) >= 20 {
            return eval.Score{Pass: true, Value: 1.0, Reason: "sufficient"}, nil
        }
        return eval.Score{Pass: false, Value: 0.0, Reason: "too short"}, nil
    }),
)

result, err := suite.Run(ctx, agentResult)
fmt.Println(result.Pass) // true if all passed

for name, score := range result.Scores {
    fmt.Printf("%s: pass=%v value=%.2f reason=%s\\n",
        name, score.Pass, score.Value, score.Reason)
}`}
                filename="evaluation suite"
              />
            </Section>

            {/* Benchmarks */}
            <Section id="benchmarks" title="Benchmarks">
              <P>
                The <Code>pkg/eval/bench</Code> package provides benchmarks for GoGrid{"'"}s
                core patterns using the mock provider. All benchmarks measure framework
                overhead, not LLM latency.
              </P>
              <CodeBlock
                code={`# Run all benchmarks
go test -bench=. ./pkg/eval/bench/

# With memory profiling
go test -bench=. -benchmem ./pkg/eval/bench/

# Available benchmarks:
# BenchmarkAgentRun              - Basic agent execution
# BenchmarkAgentRunWithToolUse   - Agent with tool calling
# BenchmarkAgentRunParallel      - Concurrent agent execution
# BenchmarkPipelineThreeStages   - Fixed three-stage pipeline
# BenchmarkPipelineScaling       - Pipeline: 1, 3, 5, 10 stages
# BenchmarkTeamTwoMembers        - Two-agent team
# BenchmarkTeamScaling           - Team: 1, 2, 5, 10, 20 members
# BenchmarkSharedMemorySaveLoad  - Memory load/save ops
# BenchmarkSharedMemoryContention - Memory: 1, 2, 5, 10 writers`}
                filename="terminal"
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
