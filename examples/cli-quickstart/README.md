# CLI Quickstart Example

This example demonstrates how to use the GoGrid CLI to define, run, and inspect agents using a `gogrid.yaml` configuration file.

## Overview

The `gogrid.yaml` in this directory defines three agents with different roles, providers, and configurations:

| Agent | Provider | Model | Role |
|-------|----------|-------|------|
| `researcher` | Anthropic | claude-sonnet-4-5-20250929 | In-depth technical analysis |
| `code-reviewer` | OpenAI | gpt-4o-mini | Go code review |
| `summarizer` | OpenAI | gpt-4o-mini | Text summarization |

## Prerequisites

1. Build the GoGrid CLI:

```bash
# From the repository root
make build
```

2. Set your API keys for the providers you want to use:

```bash
# For the researcher agent (Anthropic)
export ANTHROPIC_API_KEY=sk-ant-...

# For the code-reviewer and summarizer agents (OpenAI)
export OPENAI_API_KEY=sk-proj-...
```

## Usage

All commands below should be run from this directory (`examples/cli-quickstart/`).

### List agents

```bash
$ gogrid list
NAME             PROVIDER    MODEL
code-reviewer    openai      gpt-4o-mini
researcher       anthropic   claude-sonnet-4-5-20250929
summarizer       openai      gpt-4o-mini
```

### Run the researcher agent

```bash
gogrid run researcher -input "Explain how garbage collection works in Go"
```

The agent will respond with a structured analysis. After the response, you'll see a run ID printed to stderr:

```
Run ID: 019479a3c4e8...
```

### Run the code reviewer agent

```bash
gogrid run code-reviewer -input "Review this Go function:

func fetchUser(id string) (*User, error) {
    resp, err := http.Get(\"https://api.example.com/users/\" + id)
    if err != nil {
        return nil, err
    }
    var user User
    json.NewDecoder(resp.Body).Decode(&user)
    return &user, nil
}"
```

### Run the summarizer agent

```bash
gogrid run summarizer -input "Go is a statically typed, compiled programming language designed at Google. It is syntactically similar to C, but with memory safety, garbage collection, structural typing, and CSP-style concurrency. It was designed by Robert Griesemer, Rob Pike, and Ken Thompson. Go was publicly announced in November 2009, and version 1.0 was released in March 2012."
```

### Inspect execution traces

After running agents, inspect what happened under the hood:

```bash
# List recent runs
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

# Export as JSON for scripts or other tools
gogrid trace 019479a3c4e80001 -json | jq '.[].name'
```

### View cost breakdown

```bash
# Overview of all runs
$ gogrid cost
RUN ID              AGENT           MODEL                         COST
019479a3c4e80001    researcher      claude-sonnet-4-5-20250929    $0.003280
019479a1b2c70002    code-reviewer   gpt-4o-mini                   $0.000150
019479a0a1b60003    summarizer      gpt-4o-mini                   $0.000090

# Detailed breakdown for a specific run
$ gogrid cost 019479a3c4e80001
Run: 019479a3c4e80001

MODEL                         CALLS  PROMPT  COMPLETION  COST
claude-sonnet-4-5-20250929    2      430     234         $0.003280
────────────────────────────────────────────────────────────────
TOTAL                         2      430     234         $0.003280

# Export as JSON
gogrid cost -json
```

## Overriding configuration with environment variables

The `gogrid.yaml` uses environment variable substitution. You can swap models without editing the config:

```bash
# Use a different Anthropic model
ANTHROPIC_MODEL=claude-haiku-4-5-20251001 gogrid run researcher -input "What is a goroutine?"

# Use a different OpenAI model
OPENAI_MODEL=gpt-4o gogrid run code-reviewer -input "Review: func main() { fmt.Println(\"hello\") }"
```

## Run records

All run results are saved as JSON files under `.gogrid/runs/`. You can inspect them directly:

```bash
# List run records
ls .gogrid/runs/

# View a raw record
cat .gogrid/runs/019479a3c4e80001.json | jq .

# Extract agent names and costs from all runs
for f in .gogrid/runs/*.json; do
  echo "$(jq -r '.agent' $f): $(jq -r '.cost' $f)"
done
```

## What to explore next

- **Add tools:** Create a programmatic `main.go` using the GoGrid API to give agents tools (see `examples/single-agent/`)
- **Team orchestration:** Use `gogrid init --template team` to scaffold a multi-agent collaboration project
- **Pipeline orchestration:** Use `gogrid init --template pipeline` to scaffold a sequential processing project
- **Full API docs:** See the Go package documentation for `pkg/agent`, `pkg/orchestrator/team`, `pkg/orchestrator/pipeline`, `pkg/orchestrator/graph`, and `pkg/orchestrator/dynamic`
