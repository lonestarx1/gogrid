// Example single-agent demonstrates a basic GoGrid agent with tool use.
//
// It creates an agent with a simple calculator tool and runs it with a
// user prompt. Set OPENAI_API_KEY in your environment to run this example.
//
// Usage:
//
//	go run ./examples/single-agent
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm/openai"
	"github.com/lonestarx1/gogrid/pkg/tool"
	"github.com/lonestarx1/gogrid/pkg/trace"
)

// calculator is a simple tool that evaluates arithmetic expressions.
type calculator struct{}

func (c *calculator) Name() string        { return "calculate" }
func (c *calculator) Description() string { return "Evaluate a simple arithmetic expression (a op b)" }
func (c *calculator) Schema() tool.Schema {
	return tool.Schema{
		Type: "object",
		Properties: map[string]*tool.Schema{
			"a":  {Type: "number", Description: "First operand"},
			"b":  {Type: "number", Description: "Second operand"},
			"op": {Type: "string", Description: "Operator: +, -, *, /", Enum: []string{"+", "-", "*", "/"}},
		},
		Required: []string{"a", "b", "op"},
	}
}

func (c *calculator) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var params struct {
		A  float64 `json:"a"`
		B  float64 `json:"b"`
		Op string  `json:"op"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("parse input: %w", err)
	}

	var result float64
	switch params.Op {
	case "+":
		result = params.A + params.B
	case "-":
		result = params.A - params.B
	case "*":
		result = params.A * params.B
	case "/":
		if params.B == 0 {
			return "error: division by zero", nil
		}
		result = params.A / params.B
	default:
		return fmt.Sprintf("error: unknown operator %q", params.Op), nil
	}

	return strconv.FormatFloat(result, 'f', -1, 64), nil
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	provider := openai.New(apiKey)
	tracer := trace.NewStdout(os.Stderr)

	a := agent.New("calculator-agent",
		agent.WithProvider(provider),
		agent.WithModel("gpt-4o-mini"),
		agent.WithInstructions("You are a helpful calculator assistant. Use the calculate tool to evaluate math expressions."),
		agent.WithTools(&calculator{}),
		agent.WithTracer(tracer),
		agent.WithConfig(agent.Config{
			MaxTurns:  5,
			MaxTokens: 1024,
		}),
	)

	result, err := a.Run(context.Background(), "What is 42 * 17 + 3?")
	if err != nil {
		log.Fatalf("Run: %v", err)
	}

	fmt.Printf("\nAgent: %s\n", result.Message.Content)
	fmt.Printf("Turns: %d | Cost: $%.6f | Tokens: %d prompt, %d completion\n",
		result.Turns, result.Cost,
		result.Usage.PromptTokens, result.Usage.CompletionTokens)
}
