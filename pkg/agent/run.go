package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/lonestarx1/gogrid/internal/id"
	"github.com/lonestarx1/gogrid/pkg/cost"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/tool"
)

// Run executes the agent with the given user input.
//
// The agent loop:
//  1. Builds initial messages from system prompt, memory, and user input.
//  2. Calls the LLM with messages and tool definitions.
//  3. If the LLM responds with tool calls, executes them and loops.
//  4. If the LLM responds with a final message, returns the result.
//  5. Respects max turns, timeout, and cost budget.
func (a *Agent) Run(ctx context.Context, input string) (*Result, error) {
	if a.provider == nil {
		return nil, errors.New("agent: provider is required")
	}
	if a.model == "" {
		return nil, errors.New("agent: model is required")
	}

	// Apply timeout if configured.
	if a.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, a.config.Timeout)
		defer cancel()
	}

	runID := id.New()

	// Start run span.
	ctx, runSpan := a.tracer.StartSpan(ctx, "agent.run")
	runSpan.SetAttribute("agent.name", a.name)
	runSpan.SetAttribute("agent.run_id", runID)
	runSpan.SetAttribute("agent.model", a.model)
	defer a.tracer.EndSpan(runSpan)

	// Build initial messages.
	var messages []llm.Message
	if a.instructions != "" {
		messages = append(messages, llm.NewSystemMessage(a.instructions))
	}

	// Load conversation history from memory.
	if a.memory != nil {
		history, err := a.memory.Load(ctx, a.name)
		if err != nil {
			runSpan.SetError(err)
			return nil, fmt.Errorf("agent: load memory: %w", err)
		}
		messages = append(messages, history...)
	}

	messages = append(messages, llm.NewUserMessage(input))

	// Convert tools to LLM definitions.
	toolDefs, err := toolsToDefinitions(a.tools)
	if err != nil {
		runSpan.SetError(err)
		return nil, fmt.Errorf("agent: %w", err)
	}

	// Build tool lookup map.
	toolMap := make(map[string]tool.Tool)
	for _, t := range a.tools {
		toolMap[t.Name()] = t
	}

	// Cost tracking.
	tracker := cost.NewTracker()
	var totalCost float64
	turns := 0

	// Agent loop.
	for {
		if a.config.MaxTurns > 0 && turns >= a.config.MaxTurns {
			break
		}
		if err := ctx.Err(); err != nil {
			runSpan.SetError(err)
			return nil, fmt.Errorf("agent: %w", err)
		}
		if a.config.CostBudget > 0 && totalCost >= a.config.CostBudget {
			break
		}

		// Call LLM.
		params := llm.Params{
			Model:    a.model,
			Messages: messages,
			Tools:    toolDefs,
		}
		if a.config.Temperature != nil {
			params.Temperature = a.config.Temperature
		}
		if a.config.MaxTokens > 0 {
			params.MaxTokens = a.config.MaxTokens
		}

		_, llmSpan := a.tracer.StartSpan(ctx, "llm.complete")
		llmSpan.SetAttribute("llm.model", a.model)
		llmSpan.SetAttribute("llm.turn", strconv.Itoa(turns+1))

		resp, err := a.provider.Complete(ctx, params)
		if err != nil {
			llmSpan.SetError(err)
			a.tracer.EndSpan(llmSpan)
			runSpan.SetError(err)
			return nil, fmt.Errorf("agent: llm complete (turn %d): %w", turns+1, err)
		}

		llmSpan.SetAttribute("llm.prompt_tokens", strconv.Itoa(resp.Usage.PromptTokens))
		llmSpan.SetAttribute("llm.completion_tokens", strconv.Itoa(resp.Usage.CompletionTokens))
		a.tracer.EndSpan(llmSpan)

		// Track cost.
		callCost := tracker.Add(resp.Model, resp.Usage)
		totalCost += callCost
		turns++

		// Append assistant message.
		messages = append(messages, resp.Message)

		// If no tool calls, we're done.
		if len(resp.Message.ToolCalls) == 0 {
			break
		}

		// Execute tool calls.
		for _, tc := range resp.Message.ToolCalls {
			_, toolSpan := a.tracer.StartSpan(ctx, "tool.execute")
			toolSpan.SetAttribute("tool.name", tc.Function)
			toolSpan.SetAttribute("tool.call_id", tc.ID)

			t, ok := toolMap[tc.Function]
			if !ok {
				errMsg := fmt.Sprintf("tool %q not found", tc.Function)
				toolSpan.SetAttribute("tool.error", errMsg)
				a.tracer.EndSpan(toolSpan)
				messages = append(messages, llm.NewToolMessage(tc.ID, "error: "+errMsg))
				continue
			}

			output, err := t.Execute(ctx, tc.Arguments)
			if err != nil {
				toolSpan.SetError(err)
				a.tracer.EndSpan(toolSpan)
				messages = append(messages, llm.NewToolMessage(tc.ID, "error: "+err.Error()))
				continue
			}

			toolSpan.SetAttribute("tool.output_len", strconv.Itoa(len(output)))
			a.tracer.EndSpan(toolSpan)
			messages = append(messages, llm.NewToolMessage(tc.ID, output))
		}
	}

	// Determine final message.
	var finalMessage llm.Message
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == llm.RoleAssistant {
			finalMessage = messages[i]
			break
		}
	}

	// Save to memory.
	if a.memory != nil {
		if err := a.memory.Save(ctx, a.name, messages); err != nil {
			runSpan.SetError(err)
			return nil, fmt.Errorf("agent: save memory: %w", err)
		}
	}

	runSpan.SetAttribute("agent.turns", strconv.Itoa(turns))
	runSpan.SetAttribute("agent.cost_usd", fmt.Sprintf("%.6f", totalCost))

	return &Result{
		RunID:   runID,
		Message: finalMessage,
		History: messages,
		Usage:   tracker.TotalUsage(),
		Cost:    totalCost,
		Turns:   turns,
	}, nil
}

func toolsToDefinitions(tools []tool.Tool) ([]llm.ToolDefinition, error) {
	defs := make([]llm.ToolDefinition, 0, len(tools))
	for _, t := range tools {
		params, err := json.Marshal(t.Schema())
		if err != nil {
			return nil, fmt.Errorf("marshal schema for tool %q: %w", t.Name(), err)
		}
		defs = append(defs, llm.ToolDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  params,
		})
	}
	return defs, nil
}
