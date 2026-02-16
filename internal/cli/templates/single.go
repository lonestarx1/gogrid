package templates

func init() {
	register(&Template{
		Name:        "single",
		Description: "Single agent with tool use",
		Files: []File{
			{Path: "gogrid.yaml", Content: singleConfig},
			{Path: "main.go", Content: singleMain},
			{Path: "Makefile", Content: singleMakefile},
			{Path: "README.md", Content: singleReadme},
		},
	})
}

const singleConfig = `version: "1"
agents:
  assistant:
    model: gpt-4o-mini
    provider: openai
    instructions: |
      You are a helpful assistant. Answer questions clearly and concisely.
    config:
      max_turns: 10
      max_tokens: 4096
      timeout: 60s
      cost_budget: 0.50
`

const singleMain = `package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm/openai"
	"github.com/lonestarx1/gogrid/pkg/trace"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}

	provider := openai.New(apiKey)
	tracer := trace.NewInMemory()

	a := agent.New("assistant",
		agent.WithModel("gpt-4o-mini"),
		agent.WithProvider(provider),
		agent.WithInstructions("You are a helpful assistant."),
		agent.WithTracer(tracer),
		agent.WithConfig(agent.Config{
			MaxTurns:  10,
			MaxTokens: 4096,
		}),
	)

	result, err := a.Run(context.Background(), "Hello! What can you help me with?")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.Message.Content)
	fmt.Printf("\nTokens: %d prompt, %d completion | Cost: $%.6f\n",
		result.Usage.PromptTokens, result.Usage.CompletionTokens, result.Cost)
}
`

const singleMakefile = `.PHONY: build run clean

build:
	go build -o bin/{{.Name}} .

run: build
	./bin/{{.Name}}

clean:
	rm -rf bin/
`

const singleReadme = `# {{.Name}}

A GoGrid single-agent project.

## Setup

` + "```" + `bash
go mod tidy
export OPENAI_API_KEY=sk-...
` + "```" + `

## Run

` + "```" + `bash
# Using GoGrid CLI
gogrid run assistant -input "Hello!"

# Or directly
go run main.go
` + "```" + `
`
