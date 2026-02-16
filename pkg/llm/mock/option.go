package mock

import (
	"time"

	"github.com/lonestarx1/gogrid/pkg/llm"
)

// Option configures a mock Provider.
type Option func(*Provider)

// WithResponses sets a sequence of pre-programmed responses.
// The provider returns them in order, one per Complete call.
// After the sequence is exhausted, the fallback response is used.
func WithResponses(responses ...*llm.Response) Option {
	return func(p *Provider) {
		p.responses = responses
	}
}

// WithFallback sets the response returned after the response sequence
// is exhausted. If no fallback is set, a default empty assistant message
// is returned.
func WithFallback(resp *llm.Response) Option {
	return func(p *Provider) {
		p.fallback = resp
	}
}

// WithError configures the provider to always return the given error.
// If WithFailCount is also set, the error is only returned for the
// first N calls.
func WithError(err error) Option {
	return func(p *Provider) {
		p.err = err
	}
}

// WithFailCount causes the first n calls to Complete to return
// the configured error (defaults to "mock: injected error" if no
// error is set). Subsequent calls succeed normally.
func WithFailCount(n int) Option {
	return func(p *Provider) {
		p.failCount = n
	}
}

// WithDelay adds simulated latency to each Complete call.
// The delay respects context cancellation.
func WithDelay(d time.Duration) Option {
	return func(p *Provider) {
		p.delay = d
	}
}

// WithCallback registers a function invoked on every Complete call
// with the received params. Useful for asserting call arguments in tests.
func WithCallback(fn func(llm.Params)) Option {
	return func(p *Provider) {
		p.onCall = fn
	}
}
