package shared

import (
	"context"

	"github.com/lonestarx1/gogrid/pkg/llm"
)

// NamespacedView provides a namespaced view of a SharedMemory by
// transparently prefixing all keys. This allows multiple agents to share
// a single SharedMemory pool without key collisions.
type NamespacedView struct {
	pool   *Memory
	prefix string
}

// NewNamespacedView creates a view that prefixes all keys with the given namespace.
// The separator "/" is automatically inserted between prefix and key.
func NewNamespacedView(pool *Memory, namespace string) *NamespacedView {
	return &NamespacedView{
		pool:   pool,
		prefix: namespace + "/",
	}
}

// Load retrieves stored messages for the namespaced key.
func (v *NamespacedView) Load(ctx context.Context, key string) ([]llm.Message, error) {
	return v.pool.Load(ctx, v.prefix+key)
}

// Save stores messages under the namespaced key.
func (v *NamespacedView) Save(ctx context.Context, key string, messages []llm.Message) error {
	return v.pool.Save(ctx, v.prefix+key, messages)
}

// Clear removes all messages for the namespaced key.
func (v *NamespacedView) Clear(ctx context.Context, key string) error {
	return v.pool.Clear(ctx, v.prefix+key)
}
