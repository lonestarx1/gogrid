package cost

import (
	"sync"

	"github.com/lonestarx1/gogrid/pkg/llm"
)

// ModelPricing defines the cost per 1 million tokens for a model.
type ModelPricing struct {
	// PromptPer1M is the cost in USD per 1M prompt/input tokens.
	PromptPer1M float64
	// CompletionPer1M is the cost in USD per 1M completion/output tokens.
	CompletionPer1M float64
}

// Record captures a single cost event.
type Record struct {
	Model string
	Usage llm.Usage
	Cost  float64
}

// Tracker accumulates token usage and cost across LLM calls.
type Tracker struct {
	mu      sync.Mutex
	pricing map[string]ModelPricing
	records []Record
	total   float64
	usage   llm.Usage
}

// NewTracker creates a Tracker preloaded with default model pricing.
func NewTracker() *Tracker {
	t := &Tracker{
		pricing: make(map[string]ModelPricing),
	}
	for model, price := range DefaultPricing {
		t.pricing[model] = price
	}
	return t
}

// SetPricing sets or overrides pricing for a model.
func (t *Tracker) SetPricing(model string, p ModelPricing) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.pricing[model] = p
}

// Add records token usage for a model and returns the computed cost.
// If no pricing is configured for the model, cost is zero.
func (t *Tracker) Add(model string, usage llm.Usage) float64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	var c float64
	if p, ok := t.pricing[model]; ok {
		c = float64(usage.PromptTokens)/1_000_000*p.PromptPer1M +
			float64(usage.CompletionTokens)/1_000_000*p.CompletionPer1M
	}

	t.records = append(t.records, Record{Model: model, Usage: usage, Cost: c})
	t.total += c
	t.usage.PromptTokens += usage.PromptTokens
	t.usage.CompletionTokens += usage.CompletionTokens
	t.usage.TotalTokens += usage.TotalTokens

	return c
}

// TotalCost returns the accumulated cost in USD.
func (t *Tracker) TotalCost() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.total
}

// TotalUsage returns the accumulated token usage.
func (t *Tracker) TotalUsage() llm.Usage {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.usage
}

// Records returns a copy of all recorded cost events.
func (t *Tracker) Records() []Record {
	t.mu.Lock()
	defer t.mu.Unlock()
	cp := make([]Record, len(t.records))
	copy(cp, t.records)
	return cp
}
