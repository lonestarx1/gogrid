package trace

import (
	"context"
	"sync"
	"time"
)

// InMemory collects spans in memory. Useful for testing and debugging.
type InMemory struct {
	mu    sync.Mutex
	spans []*Span
}

// NewInMemory creates an in-memory tracer.
func NewInMemory() *InMemory {
	return &InMemory{}
}

// StartSpan begins a new span linked to any parent span in the context.
func (t *InMemory) StartSpan(ctx context.Context, name string) (context.Context, *Span) {
	return NewSpan(ctx, name)
}

// EndSpan records the completed span.
func (t *InMemory) EndSpan(span *Span) {
	span.EndTime = time.Now()
	t.mu.Lock()
	t.spans = append(t.spans, span)
	t.mu.Unlock()
}

// Spans returns a copy of all recorded spans.
func (t *InMemory) Spans() []*Span {
	t.mu.Lock()
	defer t.mu.Unlock()
	cp := make([]*Span, len(t.spans))
	copy(cp, t.spans)
	return cp
}
