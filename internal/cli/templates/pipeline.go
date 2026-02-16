package templates

func init() {
	register(&Template{
		Name:        "pipeline",
		Description: "Sequential pipeline with state transfer",
		Files: []File{
			{Path: "gogrid.yaml", Content: pipelineConfig},
			{Path: "main.go", Content: pipelineMain},
			{Path: "Makefile", Content: pipelineMakefile},
			{Path: "README.md", Content: pipelineReadme},
		},
	})
}

const pipelineConfig = `version: "1"
agents:
  drafter:
    model: gpt-4o-mini
    provider: openai
    instructions: |
      You are a content drafter. Write a first draft based on the topic.
      Be thorough but don't worry about polish.
    config:
      max_turns: 5
      max_tokens: 4096
      timeout: 60s
  editor:
    model: gpt-4o-mini
    provider: openai
    instructions: |
      You are an editor. Improve the draft for clarity, grammar,
      and structure. Return the polished version.
    config:
      max_turns: 5
      max_tokens: 4096
      timeout: 60s
`

const pipelineMain = `package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm/openai"
	"github.com/lonestarx1/gogrid/pkg/orchestrator/pipeline"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}

	provider := openai.New(apiKey)

	drafter := agent.New("drafter",
		agent.WithModel("gpt-4o-mini"),
		agent.WithProvider(provider),
		agent.WithInstructions("Write a first draft on the given topic."),
	)

	editor := agent.New("editor",
		agent.WithModel("gpt-4o-mini"),
		agent.WithProvider(provider),
		agent.WithInstructions("Polish the draft for clarity and grammar."),
	)

	p := pipeline.New("content-pipeline",
		pipeline.WithStages(
			pipeline.Stage{Name: "draft", Agent: drafter},
			pipeline.Stage{Name: "edit", Agent: editor},
		),
	)

	result, err := p.Run(context.Background(), "Write a blog post about Go concurrency patterns")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.Output)
	fmt.Printf("\nStages: %d | Cost: $%.6f\n", len(result.Stages), result.TotalCost)
}
`

const pipelineMakefile = `.PHONY: build run clean

build:
	go build -o bin/{{.Name}} .

run: build
	./bin/{{.Name}}

clean:
	rm -rf bin/
`

const pipelineReadme = `# {{.Name}}

A GoGrid pipeline project with sequential agent stages.

## Setup

` + "```" + `bash
go mod tidy
export OPENAI_API_KEY=sk-...
` + "```" + `

## Run

` + "```" + `bash
# Using GoGrid CLI (runs individual agents)
gogrid run drafter -input "Go concurrency patterns"

# Or run the full pipeline directly
go run main.go
` + "```" + `
`
