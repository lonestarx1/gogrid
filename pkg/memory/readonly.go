package memory

import (
	"context"
	"errors"

	"github.com/lonestarx1/gogrid/pkg/llm"
)

// ErrReadOnly is returned when a write operation is attempted on read-only memory.
var ErrReadOnly = errors.New("memory is read-only")

// ReadOnly wraps a Memory implementation, permitting loads but rejecting
// saves and clears. Useful for giving agents access to reference data
// they should not modify.
type ReadOnly struct {
	inner Memory
}

// NewReadOnly wraps the given memory as read-only.
func NewReadOnly(inner Memory) *ReadOnly {
	return &ReadOnly{inner: inner}
}

// Load retrieves stored messages for the given key.
func (r *ReadOnly) Load(ctx context.Context, key string) ([]llm.Message, error) {
	return r.inner.Load(ctx, key)
}

// Save always returns ErrReadOnly.
func (r *ReadOnly) Save(_ context.Context, _ string, _ []llm.Message) error {
	return ErrReadOnly
}

// Clear always returns ErrReadOnly.
func (r *ReadOnly) Clear(_ context.Context, _ string) error {
	return ErrReadOnly
}
