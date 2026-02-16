package memory

import (
	"context"

	"github.com/lonestarx1/gogrid/pkg/llm"
)

// Memory is the interface for agent conversation memory.
// Memory is a first-class primitive in GoGrid, not a plugin.
type Memory interface {
	// Load retrieves stored messages for the given key.
	// Returns an empty slice (not nil) if no messages exist for the key.
	Load(ctx context.Context, key string) ([]llm.Message, error)
	// Save stores messages under the given key, replacing any existing messages.
	Save(ctx context.Context, key string, messages []llm.Message) error
	// Clear removes all messages for the given key.
	Clear(ctx context.Context, key string) error
}
