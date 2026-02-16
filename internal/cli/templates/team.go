package templates

func init() {
	register(&Template{
		Name:        "team",
		Description: "Team of agents with shared memory",
		Files: []File{
			{Path: "gogrid.yaml", Content: teamConfig},
			{Path: "main.go", Content: teamMain},
			{Path: "Makefile", Content: teamMakefile},
			{Path: "README.md", Content: teamReadme},
		},
	})
}

const teamConfig = `version: "1"
agents:
  researcher:
    model: gpt-4o-mini
    provider: openai
    instructions: |
      You are a researcher. Analyze the topic thoroughly and provide
      detailed findings with sources.
    config:
      max_turns: 5
      max_tokens: 4096
      timeout: 60s
  reviewer:
    model: gpt-4o-mini
    provider: openai
    instructions: |
      You are a critical reviewer. Evaluate the research for accuracy,
      completeness, and potential biases.
    config:
      max_turns: 5
      max_tokens: 4096
      timeout: 60s
`

const teamMain = `package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm/openai"
	"github.com/lonestarx1/gogrid/pkg/memory/shared"
	"github.com/lonestarx1/gogrid/pkg/orchestrator/team"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}

	provider := openai.New(apiKey)
	sharedMem := shared.New()

	researcher := agent.New("researcher",
		agent.WithModel("gpt-4o-mini"),
		agent.WithProvider(provider),
		agent.WithInstructions("You are a researcher. Provide detailed analysis."),
		agent.WithMemory(sharedMem.For("researcher")),
	)

	reviewer := agent.New("reviewer",
		agent.WithModel("gpt-4o-mini"),
		agent.WithProvider(provider),
		agent.WithInstructions("You are a reviewer. Evaluate for accuracy and gaps."),
		agent.WithMemory(sharedMem.For("reviewer")),
	)

	t := team.New("research-team",
		team.WithMembers(
			team.Member{Agent: researcher, Role: "researcher"},
			team.Member{Agent: reviewer, Role: "reviewer"},
		),
		team.WithStrategy(team.Unanimous),
		team.WithMaxRounds(3),
	)

	result, err := t.Run(context.Background(), "Analyze the impact of AI on software development")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.Decision.Content)
	fmt.Printf("\nRounds: %d | Cost: $%.6f\n", result.Rounds, result.TotalCost)
}
`

const teamMakefile = `.PHONY: build run clean

build:
	go build -o bin/{{.Name}} .

run: build
	./bin/{{.Name}}

clean:
	rm -rf bin/
`

const teamReadme = `# {{.Name}}

A GoGrid team project with multiple collaborating agents.

## Setup

` + "```" + `bash
go mod tidy
export OPENAI_API_KEY=sk-...
` + "```" + `

## Run

` + "```" + `bash
# Using GoGrid CLI (runs individual agents)
gogrid run researcher -input "Analyze AI impact"

# Or run the team directly
go run main.go
` + "```" + `
`
