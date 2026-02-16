package trace

import "context"

// Noop is a tracer that does nothing. Used as the default when
// no tracer is configured, avoiding nil checks throughout the code.
type Noop struct{}

// StartSpan returns the context and an empty span.
func (Noop) StartSpan(ctx context.Context, _ string) (context.Context, *Span) {
	return ctx, &Span{}
}

// EndSpan does nothing.
func (Noop) EndSpan(_ *Span) {}
