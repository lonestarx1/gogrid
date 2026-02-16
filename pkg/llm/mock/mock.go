package mock

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lonestarx1/gogrid/pkg/llm"
)

// Provider is a configurable mock implementation of llm.Provider.
// It is safe for concurrent use.
type Provider struct {
	mu        sync.Mutex
	responses []*llm.Response
	fallback  *llm.Response
	calls     atomic.Int32
	history   []llm.Params
	err       error
	failCount int
	delay     time.Duration
	onCall    func(llm.Params)
}

// New creates a mock Provider with the given options.
func New(opts ...Option) *Provider {
	p := &Provider{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Complete implements llm.Provider. It records the call, applies any
// configured delay and error injection, then returns the next response
// in the sequence or the fallback.
func (p *Provider) Complete(ctx context.Context, params llm.Params) (*llm.Response, error) {
	callNum := int(p.calls.Add(1))

	// Record params.
	p.mu.Lock()
	p.history = append(p.history, params)
	onCall := p.onCall
	p.mu.Unlock()

	// Invoke callback.
	if onCall != nil {
		onCall(params)
	}

	// Simulate latency.
	if p.delay > 0 {
		select {
		case <-time.After(p.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Error injection.
	if p.err != nil {
		if p.failCount == 0 {
			// Constant error mode: always fail.
			return nil, p.err
		}
		if callNum <= p.failCount {
			return nil, p.err
		}
	} else if p.failCount > 0 && callNum <= p.failCount {
		return nil, errors.New("mock: injected error")
	}

	// Determine response index, adjusting for failed calls.
	idx := callNum - 1
	if p.failCount > 0 {
		idx = callNum - p.failCount - 1
	}

	if idx >= 0 && idx < len(p.responses) {
		return p.responses[idx], nil
	}

	// Fallback response.
	if p.fallback != nil {
		return p.fallback, nil
	}

	return &llm.Response{
		Message: llm.NewAssistantMessage(""),
		Usage:   llm.Usage{},
		Model:   "mock",
	}, nil
}

// Calls returns the total number of Complete calls made.
func (p *Provider) Calls() int {
	return int(p.calls.Load())
}

// History returns a copy of all recorded call params.
func (p *Provider) History() []llm.Params {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]llm.Params, len(p.history))
	copy(out, p.history)
	return out
}

// Reset clears the call counter and history. The configured responses,
// fallback, error, and delay are preserved.
func (p *Provider) Reset() {
	p.calls.Store(0)
	p.mu.Lock()
	p.history = nil
	p.mu.Unlock()
}
