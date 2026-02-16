package cost

import (
	"sort"
	"sync"
	"time"

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
	Model  string
	Entity string
	Usage  llm.Usage
	Cost   float64
	Time   time.Time
}

// AlertFunc is called when cost crosses a threshold. It receives the
// threshold value and the current total cost.
type AlertFunc func(threshold, current float64)

// Tracker accumulates token usage and cost across LLM calls.
// It supports budget alerts, cost allocation per entity, and reporting.
type Tracker struct {
	mu      sync.Mutex
	pricing map[string]ModelPricing
	records []Record
	total   float64
	usage   llm.Usage

	// Alert thresholds and callback.
	budget     float64
	thresholds []float64 // sorted ascending (fractions 0-1)
	nextAlert  int       // index into thresholds of next alert to fire
	alertFn    AlertFunc

	// Per-entity cost allocation.
	entityCost map[string]float64
}

// NewTracker creates a Tracker preloaded with default model pricing.
func NewTracker() *Tracker {
	t := &Tracker{
		pricing:    make(map[string]ModelPricing),
		entityCost: make(map[string]float64),
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

// SetBudget configures a total cost budget in USD. Alerts are triggered
// when cost crosses the configured threshold fractions of this budget.
func (t *Tracker) SetBudget(budget float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.budget = budget
}

// OnBudgetThreshold registers an alert callback that fires when the
// accumulated cost crosses each threshold fraction (0.0–1.0) of the
// configured budget. Thresholds are checked in ascending order; each
// fires at most once.
//
// Example:
//
//	tracker.SetBudget(10.0)
//	tracker.OnBudgetThreshold(func(threshold, current float64) {
//	    log.Printf("cost alert: %.0f%% of budget (%.2f USD)", threshold*100, current)
//	}, 0.5, 0.8, 1.0)
func (t *Tracker) OnBudgetThreshold(fn AlertFunc, thresholds ...float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	sorted := make([]float64, len(thresholds))
	copy(sorted, thresholds)
	sort.Float64s(sorted)
	t.thresholds = sorted
	t.nextAlert = 0
	t.alertFn = fn
}

// Add records token usage for a model and returns the computed cost.
// If no pricing is configured for the model, cost is zero.
func (t *Tracker) Add(model string, usage llm.Usage) float64 {
	return t.AddForEntity(model, "", usage)
}

// AddForEntity records token usage attributed to a named entity (agent,
// team, pipeline, etc.) and returns the computed cost.
func (t *Tracker) AddForEntity(model, entity string, usage llm.Usage) float64 {
	t.mu.Lock()

	var c float64
	if p, ok := t.pricing[model]; ok {
		c = float64(usage.PromptTokens)/1_000_000*p.PromptPer1M +
			float64(usage.CompletionTokens)/1_000_000*p.CompletionPer1M
	}

	t.records = append(t.records, Record{
		Model:  model,
		Entity: entity,
		Usage:  usage,
		Cost:   c,
		Time:   time.Now(),
	})
	t.total += c
	t.usage.PromptTokens += usage.PromptTokens
	t.usage.CompletionTokens += usage.CompletionTokens
	t.usage.TotalTokens += usage.TotalTokens

	if entity != "" {
		t.entityCost[entity] += c
	}

	// Check alerts (collect any that need firing while holding the lock).
	var alerts []float64
	if t.budget > 0 && t.alertFn != nil {
		for t.nextAlert < len(t.thresholds) {
			threshold := t.thresholds[t.nextAlert]
			if t.total >= t.budget*threshold {
				alerts = append(alerts, threshold)
				t.nextAlert++
			} else {
				break
			}
		}
	}

	current := t.total
	fn := t.alertFn
	t.mu.Unlock()

	// Fire alerts outside the lock to prevent deadlocks.
	for _, threshold := range alerts {
		fn(threshold, current)
	}

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

// EntityCost returns the accumulated cost for a specific entity.
func (t *Tracker) EntityCost(entity string) float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.entityCost[entity]
}

// Report generates a cost report summarizing usage by model and entity.
type Report struct {
	TotalCost   float64
	TotalUsage  llm.Usage
	ByModel     map[string]ModelReport
	ByEntity    map[string]float64
	RecordCount int
}

// ModelReport summarizes cost and usage for a single model.
type ModelReport struct {
	Cost  float64
	Usage llm.Usage
	Calls int
}

// Report generates an aggregate cost report.
func (t *Tracker) Report() Report {
	t.mu.Lock()
	defer t.mu.Unlock()

	r := Report{
		TotalCost:   t.total,
		TotalUsage:  t.usage,
		ByModel:     make(map[string]ModelReport),
		ByEntity:    make(map[string]float64, len(t.entityCost)),
		RecordCount: len(t.records),
	}

	for _, rec := range t.records {
		mr := r.ByModel[rec.Model]
		mr.Cost += rec.Cost
		mr.Usage.PromptTokens += rec.Usage.PromptTokens
		mr.Usage.CompletionTokens += rec.Usage.CompletionTokens
		mr.Usage.TotalTokens += rec.Usage.TotalTokens
		mr.Calls++
		r.ByModel[rec.Model] = mr
	}

	for entity, cost := range t.entityCost {
		r.ByEntity[entity] = cost
	}

	return r
}
