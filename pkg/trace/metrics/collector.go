package metrics

import (
	"context"
	"strconv"

	"github.com/lonestarx1/gogrid/pkg/trace"
)

// Collector wraps a trace.Tracer and automatically populates metrics
// from GoGrid trace spans. Use it as a drop-in replacement for any
// tracer to gain automatic metrics collection.
type Collector struct {
	inner trace.Tracer
	reg   *Registry

	agentRuns     *Counter
	agentDuration *Histogram
	llmCalls      *Counter
	llmDuration   *Histogram
	llmTokens     *Counter
	toolExecs     *Counter
	toolDuration  *Histogram
	memoryOps     *Counter
	costUSD       *Counter
}

// NewCollector creates a Collector that delegates span management to
// inner and records metrics in reg.
func NewCollector(inner trace.Tracer, reg *Registry) *Collector {
	return &Collector{
		inner:         inner,
		reg:           reg,
		agentRuns:     reg.Counter("gogrid_agent_runs_total", "Total number of agent runs"),
		agentDuration: reg.Histogram("gogrid_agent_run_duration_seconds", "Agent run duration in seconds"),
		llmCalls:      reg.Counter("gogrid_llm_calls_total", "Total number of LLM calls"),
		llmDuration:   reg.Histogram("gogrid_llm_call_duration_seconds", "LLM call duration in seconds"),
		llmTokens:     reg.Counter("gogrid_llm_tokens_total", "Total LLM tokens consumed"),
		toolExecs:     reg.Counter("gogrid_tool_executions_total", "Total tool executions"),
		toolDuration:  reg.Histogram("gogrid_tool_execution_duration_seconds", "Tool execution duration in seconds"),
		memoryOps:     reg.Counter("gogrid_memory_operations_total", "Total memory operations"),
		costUSD:       reg.Counter("gogrid_cost_usd_total", "Total cost in USD"),
	}
}

// StartSpan delegates to the inner tracer.
func (c *Collector) StartSpan(ctx context.Context, name string) (context.Context, *trace.Span) {
	return c.inner.StartSpan(ctx, name)
}

// EndSpan delegates to the inner tracer and records metrics.
func (c *Collector) EndSpan(span *trace.Span) {
	c.inner.EndSpan(span)
	c.record(span)
}

func (c *Collector) record(span *trace.Span) {
	duration := span.EndTime.Sub(span.StartTime).Seconds()
	status := "ok"
	if span.Status == trace.StatusError {
		status = "error"
	}

	switch span.Name {
	case "agent.run":
		agentName := span.Attributes["agent.name"]
		c.agentRuns.Inc(map[string]string{"agent": agentName, "status": status})
		c.agentDuration.Observe(duration, map[string]string{"agent": agentName})
		if costStr, ok := span.Attributes["agent.cost_usd"]; ok {
			if cost, err := strconv.ParseFloat(costStr, 64); err == nil && cost > 0 {
				model := span.Attributes["agent.model"]
				c.costUSD.Add(cost, map[string]string{"agent": agentName, "model": model})
			}
		}

	case "llm.complete":
		model := span.Attributes["llm.model"]
		c.llmCalls.Inc(map[string]string{"model": model, "status": status})
		c.llmDuration.Observe(duration, map[string]string{"model": model})
		if pt, err := strconv.Atoi(span.Attributes["llm.prompt_tokens"]); err == nil {
			c.llmTokens.Add(float64(pt), map[string]string{"model": model, "type": "prompt"})
		}
		if ct, err := strconv.Atoi(span.Attributes["llm.completion_tokens"]); err == nil {
			c.llmTokens.Add(float64(ct), map[string]string{"model": model, "type": "completion"})
		}

	case "tool.execute":
		toolName := span.Attributes["tool.name"]
		c.toolExecs.Inc(map[string]string{"tool": toolName, "status": status})
		c.toolDuration.Observe(duration, map[string]string{"tool": toolName})

	case "memory.load":
		c.memoryOps.Inc(map[string]string{"operation": "load"})
	case "memory.save":
		c.memoryOps.Inc(map[string]string{"operation": "save"})
	}
}
