# GoGrid CLI Reference

The GoGrid CLI (`gogrid`) provides commands to scaffold projects, run agents, and inspect execution traces and costs from the command line. It reads agent definitions from a `gogrid.yaml` configuration file and resolves LLM provider credentials from environment variables.

## Installation

Build from source:

```bash
git clone https://github.com/lonestarx1/gogrid.git
cd gogrid
make build
```

The binary is written to `bin/gogrid`. Add it to your PATH or run it directly.

To embed a version string:

```bash
make build VERSION=1.0.0
bin/gogrid version
# gogrid 1.0.0 (darwin/arm64, go1.25.4)
```

---

## Commands

### `gogrid init`

Scaffold a new GoGrid project from a template.

```
gogrid init [flags] [directory]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-template` | `single` | Project template: `single`, `team`, or `pipeline` |
| `-name` | directory name | Project name used in generated files |

**Templates:**

- **`single`** — A single agent with instructions and configuration. The simplest starting point.
- **`team`** — Two agents (researcher + reviewer) collaborating via shared memory and consensus.
- **`pipeline`** — Two sequential stages (drafter + editor) with state transfer between stages.

Each template generates four files: `gogrid.yaml`, `main.go`, `Makefile`, and `README.md`.

**Examples:**

```bash
# Scaffold a single-agent project in a new directory
gogrid init --template single my-agent

# Scaffold a team project with a custom name
gogrid init --template team --name research-bot ./research

# Scaffold a pipeline project in the current (empty) directory
gogrid init --template pipeline
```

**What gets generated (single template):**

```
my-agent/
  gogrid.yaml   # Agent configuration
  main.go       # Programmatic entry point using GoGrid API
  Makefile      # Build targets
  README.md     # Setup instructions
```

After scaffolding:

```bash
cd my-agent
go mod init github.com/example/my-agent
go mod tidy
export OPENAI_API_KEY=sk-...
gogrid run assistant -input "Hello!"
```

---

### `gogrid list`

List all agents defined in the project's `gogrid.yaml`.

```
gogrid list [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-config` | `gogrid.yaml` | Path to configuration file |

**Example:**

```bash
$ gogrid list
NAME          PROVIDER    MODEL
researcher    anthropic   claude-sonnet-4-5-20250929
summarizer    openai      gpt-4o-mini
```

---

### `gogrid run`

Execute a named agent with the given input.

```
gogrid run <agent-name> [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-config` | `gogrid.yaml` | Path to configuration file |
| `-input` | (required) | Input text to send to the agent |
| `-timeout` | from config | Override the agent's timeout (e.g. `30s`, `5m`) |

The agent's response is printed to stdout. A run record is saved to `.gogrid/runs/<run-id>.json` for later inspection with `gogrid trace` and `gogrid cost`. The run ID is printed to stderr.

**Examples:**

```bash
# Run an agent with inline input
gogrid run researcher -input "Explain Go's context package"

# Override timeout
gogrid run summarizer -input "Summarize this paper..." -timeout 2m

# Use a different config file
gogrid run assistant -config staging.yaml -input "Hello"
```

**What happens during a run:**

1. Loads and validates `gogrid.yaml`
2. Looks up the agent by name
3. Resolves the LLM provider using environment variables (see [Environment Variables](#environment-variables))
4. Creates the agent with the configured model, instructions, and execution parameters
5. Calls `agent.Run()` with an in-memory tracer to capture spans
6. Prints the agent's response to stdout
7. Saves the full run record (spans, usage, cost) to `.gogrid/runs/`

---

### `gogrid trace`

Inspect execution traces for agent runs.

```
gogrid trace [run-id] [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-json` | `false` | Output spans as JSON instead of a tree |

With no arguments, lists the 10 most recent runs. With a run ID, renders the span tree.

**Examples:**

```bash
# List recent runs
$ gogrid trace
Recent runs:
  019479a3c4e80001  researcher  claude-sonnet-4-5-20250929  4.2s
  019479a1b2c70002  summarizer  gpt-4o-mini                1.1s

# View span tree for a specific run
$ gogrid trace 019479a3c4e80001
Run: 019479a3c4e80001
Agent: researcher | Model: claude-sonnet-4-5-20250929 | Duration: 4.2s

agent.run (4.2s)
├── memory.load (1ms)
├── llm.complete (2.1s) [prompt: 150, completion: 89]
├── tool.execute (1.8s) ["web_search"]
├── llm.complete (0.3s) [prompt: 280, completion: 45]
└── memory.save (2ms)

# Export as JSON for programmatic use
gogrid trace 019479a3c4e80001 -json
```

The span tree shows the hierarchical execution flow: LLM calls with token counts, tool executions, memory operations, and timing for each step.

---

### `gogrid cost`

View cost breakdown for agent runs.

```
gogrid cost [run-id] [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-json` | `false` | Output cost data as JSON |

With no arguments, lists all runs with their total cost. With a run ID, shows a per-model cost breakdown.

**Examples:**

```bash
# List all runs with costs
$ gogrid cost
RUN ID              AGENT       MODEL                         COST
019479a3c4e80001    researcher  claude-sonnet-4-5-20250929    $0.003280
019479a1b2c70002    summarizer  gpt-4o-mini                   $0.000150

# Detailed cost breakdown for a run
$ gogrid cost 019479a3c4e80001
Run: 019479a3c4e80001

MODEL                         CALLS  PROMPT  COMPLETION  COST
claude-sonnet-4-5-20250929    2      430     134         $0.003280
────────────────────────────────────────────────────────────────
TOTAL                         2      430     134         $0.003280

# Export as JSON
gogrid cost 019479a3c4e80001 -json
```

---

### `gogrid version`

Print the GoGrid version, platform, and Go version.

```bash
$ gogrid version
gogrid 1.0.0 (darwin/arm64, go1.25.4)
```

---

### `gogrid help`

Show the help message with all available commands.

```bash
gogrid help
gogrid -h
gogrid --help
```

---

## Configuration

GoGrid projects are configured via a `gogrid.yaml` file in the project root.

### Schema

```yaml
version: "1"                    # Required. Config schema version.

agents:
  <agent-name>:                 # Unique agent identifier.
    model: <string>             # Required. LLM model ID.
    provider: <string>          # Required. One of: openai, anthropic, gemini.
    instructions: <string>      # System prompt for the agent.
    config:
      max_turns: <int>          # Max LLM round-trips. 0 = unlimited.
      max_tokens: <int>         # Max response tokens per turn.
      temperature: <float>      # LLM randomness (0.0-1.0). Omit for provider default.
      timeout: <duration>       # Wall-clock limit (e.g. "30s", "5m", "1h").
      cost_budget: <float>      # Max cost in USD for a single run.
```

### Full Example

```yaml
version: "1"

agents:
  researcher:
    model: claude-sonnet-4-5-20250929
    provider: anthropic
    instructions: |
      You are a research assistant. When given a topic, provide a thorough
      analysis with key findings, supporting evidence, and areas that need
      further investigation.
    config:
      max_turns: 10
      max_tokens: 4096
      temperature: 0.7
      timeout: 2m
      cost_budget: 0.50

  summarizer:
    model: gpt-4o-mini
    provider: openai
    instructions: |
      You are a summarizer. Take the provided text and produce a concise
      summary that captures the key points in 3-5 bullet points.
    config:
      max_turns: 3
      max_tokens: 1024
      timeout: 30s
      cost_budget: 0.10

  translator:
    model: gemini-2.5-flash
    provider: gemini
    instructions: |
      You are a translator. Translate the input text to the requested
      language while preserving tone and meaning.
    config:
      max_turns: 3
      max_tokens: 4096
      timeout: 30s
```

### Environment Variable Substitution

Config values support `${VAR}` and `${VAR:-default}` syntax for environment variable substitution. This is processed before YAML parsing.

```yaml
version: "1"

agents:
  assistant:
    model: ${MODEL:-gpt-4o-mini}
    provider: ${PROVIDER:-openai}
    instructions: ${AGENT_INSTRUCTIONS:-You are a helpful assistant.}
    config:
      max_turns: 10
      timeout: ${TIMEOUT:-60s}
```

This lets you change model, provider, or other settings per environment without modifying the config file:

```bash
MODEL=claude-sonnet-4-5-20250929 PROVIDER=anthropic gogrid run assistant -input "Hello"
```

### Validation Rules

The config is validated on load. The following rules apply:

- `version` must be `"1"`
- At least one agent must be defined
- Each agent must have `model` and `provider`
- `provider` must be one of: `openai`, `anthropic`, `gemini`

---

## Environment Variables

The CLI resolves LLM provider credentials from environment variables. No secrets are stored in config files.

| Provider | Environment Variable | Example |
|----------|---------------------|---------|
| OpenAI | `OPENAI_API_KEY` | `sk-proj-...` |
| Anthropic | `ANTHROPIC_API_KEY` | `sk-ant-...` |
| Gemini | `GEMINI_API_KEY` | `AIza...` |

Set the variable for whichever provider your agents use:

```bash
# For OpenAI models
export OPENAI_API_KEY=sk-proj-...

# For Anthropic models
export ANTHROPIC_API_KEY=sk-ant-...

# For Google Gemini models
export GEMINI_API_KEY=AIza...

# Multiple providers at once (for projects with mixed providers)
export OPENAI_API_KEY=sk-proj-...
export ANTHROPIC_API_KEY=sk-ant-...
```

---

## Run Records

Every `gogrid run` invocation saves a JSON record to `.gogrid/runs/<run-id>.json`. Run IDs are time-sortable, so newer runs sort after older ones.

A run record contains:

| Field | Description |
|-------|-------------|
| `run_id` | Unique, time-sortable identifier |
| `agent` | Agent name from config |
| `model` | LLM model used |
| `provider` | LLM provider used |
| `input` | User input text |
| `output` | Agent's final response |
| `turns` | Number of LLM round-trips |
| `usage` | Token counts (prompt, completion, total) |
| `cost` | Estimated cost in USD |
| `spans` | Execution trace spans (LLM calls, tool executions, memory operations) |
| `cost_records` | Per-call cost breakdown |
| `start_time` | When the run started |
| `duration` | Wall-clock duration |
| `error` | Error message if the run failed |

Run records are plain JSON files. You can inspect them directly, back them up, or pipe them to other tools:

```bash
# View raw record
cat .gogrid/runs/019479a3c4e80001.json | jq .

# Extract just the cost from all runs
ls .gogrid/runs/*.json | xargs -I{} jq -r '[.run_id, .agent, .cost] | @tsv' {}

# Total cost across all runs
ls .gogrid/runs/*.json | xargs -I{} jq '.cost' {} | paste -sd+ | bc
```

Add `.gogrid/` to your `.gitignore` — run records are local development artifacts:

```
# .gitignore
.gogrid/
```

---

## Supported Models

GoGrid includes built-in pricing for cost tracking. Any model string is accepted — these are the ones with pre-configured pricing:

**OpenAI:** `gpt-4o`, `gpt-4o-mini`, `gpt-4.1`, `gpt-4.1-mini`, `gpt-4.1-nano`, `o3`, `o4-mini`

**Anthropic:** `claude-opus-4-6-20250827`, `claude-opus-4-5-20250620`, `claude-sonnet-4-5-20250929`, `claude-sonnet-4-0-20250514`, `claude-haiku-4-5-20251001`

**Google Gemini:** `gemini-3-pro`, `gemini-3-flash`, `gemini-2.5-pro`, `gemini-2.5-flash`, `gemini-2.0-flash`

Models not in this list work fine — cost tracking will report $0.00 until custom pricing is configured via the Go API (`cost.Tracker.SetPricing`).

---

## Typical Workflow

```bash
# 1. Scaffold a project
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
gogrid run assistant -input "Explain the CAP theorem in simple terms"

# 6. Inspect the trace
gogrid trace    # list recent runs, copy a run ID
gogrid trace <run-id>

# 7. Check costs
gogrid cost <run-id>
gogrid cost     # summary of all runs
```

---

## Project Structure

A typical GoGrid project:

```
my-project/
  gogrid.yaml         # Agent configuration
  main.go             # Programmatic entry point (optional — for custom logic)
  Makefile             # Build targets
  .gogrid/
    runs/              # Run records (auto-created by gogrid run)
      019479a3c4e8.json
      019479a1b2c7.json
```

The CLI and the Go API are complementary. Use the CLI for quick iteration and inspection. Use `main.go` with the Go API when you need tools, custom orchestration patterns, or programmatic control.
