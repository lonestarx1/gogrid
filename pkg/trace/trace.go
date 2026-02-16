package trace

import (
	"context"
	"time"

	"github.com/lonestarx1/gogrid/internal/id"
)

// Tracer creates and manages trace spans for GoGrid agent execution.
type Tracer interface {
	// StartSpan begins a new span. The returned context carries the span
	// so child spans can reference it as their parent.
	StartSpan(ctx context.Context, name string) (context.Context, *Span)
	// EndSpan completes the span and records it.
	EndSpan(span *Span)
}

// Status indicates whether a span completed successfully.
type Status int

const (
	// StatusOK indicates the span completed without error.
	StatusOK Status = iota
	// StatusError indicates the span encountered an error.
	StatusError
)

// Span represents a unit of work within a trace.
type Span struct {
	// ID uniquely identifies this span.
	ID string `json:"id"`
	// ParentID is the ID of the parent span, empty for root spans.
	ParentID string `json:"parent_id,omitempty"`
	// Name describes the operation (e.g. "agent.run", "llm.complete", "tool.execute").
	Name string `json:"name"`
	// StartTime is when the span began.
	StartTime time.Time `json:"start_time"`
	// EndTime is when the span ended.
	EndTime time.Time `json:"end_time,omitempty"`
	// Attributes holds key-value metadata about the span.
	Attributes map[string]string `json:"attributes,omitempty"`
	// Status indicates success or failure.
	Status Status `json:"status"`
	// Error is the error message if Status is StatusError.
	Error string `json:"error,omitempty"`
}

// SetAttribute adds a key-value attribute to the span.
func (s *Span) SetAttribute(key, value string) {
	if s.Attributes == nil {
		s.Attributes = make(map[string]string)
	}
	s.Attributes[key] = value
}

// SetError marks the span as failed with the given error.
func (s *Span) SetError(err error) {
	s.Status = StatusError
	s.Error = err.Error()
}

type spanContextKey struct{}

// SpanFromContext returns the current span from the context, or nil.
func SpanFromContext(ctx context.Context) *Span {
	span, _ := ctx.Value(spanContextKey{}).(*Span)
	return span
}

func newSpan(ctx context.Context, name string) (context.Context, *Span) {
	span := &Span{
		ID:        id.New(),
		Name:      name,
		StartTime: time.Now(),
	}
	if parent := SpanFromContext(ctx); parent != nil {
		span.ParentID = parent.ID
	}
	return context.WithValue(ctx, spanContextKey{}, span), span
}
