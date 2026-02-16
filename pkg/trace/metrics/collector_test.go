package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/lonestarx1/gogrid/pkg/trace"
)

func TestCollectorDelegatesSpans(t *testing.T) {
	inner := trace.NewInMemory()
	reg := NewRegistry()
	c := NewCollector(inner, reg)

	ctx, span := c.StartSpan(context.Background(), "test.span")
	if span == nil {
		t.Fatal("span is nil")
	}
	if ctx == nil {
		t.Fatal("ctx is nil")
	}
	c.EndSpan(span)

	spans := inner.Spans()
	if len(spans) != 1 {
		t.Fatalf("inner spans = %d, want 1", len(spans))
	}
	if spans[0].Name != "test.span" {
		t.Errorf("span name = %q, want %q", spans[0].Name, "test.span")
	}
}

func TestCollectorAgentRunMetrics(t *testing.T) {
	inner := trace.NewInMemory()
	reg := NewRegistry()
	c := NewCollector(inner, reg)

	_, span := c.StartSpan(context.Background(), "agent.run")
	span.SetAttribute("agent.name", "researcher")
	span.SetAttribute("agent.cost_usd", "0.05")
	span.SetAttribute("agent.model", "gpt-4o")
	span.StartTime = time.Now().Add(-2 * time.Second)
	c.EndSpan(span)

	// Check agent runs counter.
	runs := c.agentRuns.Value(map[string]string{"agent": "researcher", "status": "ok"})
	if runs != 1 {
		t.Errorf("agent runs = %f, want 1", runs)
	}

	// Check duration was observed.
	count := c.agentDuration.Count(map[string]string{"agent": "researcher"})
	if count != 1 {
		t.Errorf("agent duration count = %d, want 1", count)
	}

	// Check cost.
	cost := c.costUSD.Value(map[string]string{"agent": "researcher", "model": "gpt-4o"})
	if cost != 0.05 {
		t.Errorf("cost = %f, want 0.05", cost)
	}
}

func TestCollectorAgentRunError(t *testing.T) {
	inner := trace.NewInMemory()
	reg := NewRegistry()
	c := NewCollector(inner, reg)

	_, span := c.StartSpan(context.Background(), "agent.run")
	span.SetAttribute("agent.name", "researcher")
	span.Status = trace.StatusError
	c.EndSpan(span)

	errRuns := c.agentRuns.Value(map[string]string{"agent": "researcher", "status": "error"})
	if errRuns != 1 {
		t.Errorf("error agent runs = %f, want 1", errRuns)
	}
}

func TestCollectorLLMMetrics(t *testing.T) {
	inner := trace.NewInMemory()
	reg := NewRegistry()
	c := NewCollector(inner, reg)

	_, span := c.StartSpan(context.Background(), "llm.complete")
	span.SetAttribute("llm.model", "gpt-4o")
	span.SetAttribute("llm.prompt_tokens", "100")
	span.SetAttribute("llm.completion_tokens", "50")
	span.StartTime = time.Now().Add(-500 * time.Millisecond)
	c.EndSpan(span)

	calls := c.llmCalls.Value(map[string]string{"model": "gpt-4o", "status": "ok"})
	if calls != 1 {
		t.Errorf("llm calls = %f, want 1", calls)
	}

	promptTokens := c.llmTokens.Value(map[string]string{"model": "gpt-4o", "type": "prompt"})
	if promptTokens != 100 {
		t.Errorf("prompt tokens = %f, want 100", promptTokens)
	}

	completionTokens := c.llmTokens.Value(map[string]string{"model": "gpt-4o", "type": "completion"})
	if completionTokens != 50 {
		t.Errorf("completion tokens = %f, want 50", completionTokens)
	}
}

func TestCollectorToolMetrics(t *testing.T) {
	inner := trace.NewInMemory()
	reg := NewRegistry()
	c := NewCollector(inner, reg)

	_, span := c.StartSpan(context.Background(), "tool.execute")
	span.SetAttribute("tool.name", "search")
	span.StartTime = time.Now().Add(-100 * time.Millisecond)
	c.EndSpan(span)

	execs := c.toolExecs.Value(map[string]string{"tool": "search", "status": "ok"})
	if execs != 1 {
		t.Errorf("tool execs = %f, want 1", execs)
	}

	count := c.toolDuration.Count(map[string]string{"tool": "search"})
	if count != 1 {
		t.Errorf("tool duration count = %d, want 1", count)
	}
}

func TestCollectorMemoryMetrics(t *testing.T) {
	inner := trace.NewInMemory()
	reg := NewRegistry()
	c := NewCollector(inner, reg)

	_, loadSpan := c.StartSpan(context.Background(), "memory.load")
	c.EndSpan(loadSpan)

	_, saveSpan := c.StartSpan(context.Background(), "memory.save")
	c.EndSpan(saveSpan)

	loads := c.memoryOps.Value(map[string]string{"operation": "load"})
	if loads != 1 {
		t.Errorf("memory loads = %f, want 1", loads)
	}

	saves := c.memoryOps.Value(map[string]string{"operation": "save"})
	if saves != 1 {
		t.Errorf("memory saves = %f, want 1", saves)
	}
}

func TestCollectorUnknownSpanName(t *testing.T) {
	inner := trace.NewInMemory()
	reg := NewRegistry()
	c := NewCollector(inner, reg)

	_, span := c.StartSpan(context.Background(), "unknown.operation")
	c.EndSpan(span)

	// Should not panic, no metrics recorded.
	out := reg.Export()
	if out != "" {
		t.Errorf("expected empty export for unknown span, got: %q", out)
	}
}

func TestCollectorMetricsViaExport(t *testing.T) {
	inner := trace.NewInMemory()
	reg := NewRegistry()
	c := NewCollector(inner, reg)

	_, span := c.StartSpan(context.Background(), "agent.run")
	span.SetAttribute("agent.name", "test")
	c.EndSpan(span)

	out := reg.Export()
	if out == "" {
		t.Error("expected non-empty export after recording metrics")
	}
}
