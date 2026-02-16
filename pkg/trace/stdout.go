package trace

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"
)

// Stdout writes completed spans as JSON to a writer. Suitable for
// structured logging during development.
type Stdout struct {
	mu  sync.Mutex
	enc *json.Encoder
}

// NewStdout creates a tracer that writes JSON spans to w.
func NewStdout(w io.Writer) *Stdout {
	return &Stdout{enc: json.NewEncoder(w)}
}

// StartSpan begins a new span linked to any parent span in the context.
func (t *Stdout) StartSpan(ctx context.Context, name string) (context.Context, *Span) {
	return newSpan(ctx, name)
}

// EndSpan records the span end time and writes it as JSON.
func (t *Stdout) EndSpan(span *Span) {
	span.EndTime = time.Now()
	t.mu.Lock()
	_ = t.enc.Encode(span)
	t.mu.Unlock()
}
