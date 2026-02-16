package memory

import (
	"context"
	"time"

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

// Entry holds a stored message along with its metadata.
type Entry struct {
	// Key is the memory key this entry belongs to.
	Key string
	// Message is the stored conversation message.
	Message llm.Message
	// CreatedAt is when the entry was first stored.
	CreatedAt time.Time
	// Size is the byte size of the message content.
	Size int
}

// Stats reports aggregate statistics about a memory store.
type Stats struct {
	// Keys is the number of distinct keys stored.
	Keys int
	// TotalEntries is the total number of messages across all keys.
	TotalEntries int
	// TotalSize is the total byte size of all stored message content.
	TotalSize int64
	// OldestEntry is the timestamp of the oldest stored entry.
	// Zero value if no entries exist.
	OldestEntry time.Time
	// NewestEntry is the timestamp of the newest stored entry.
	// Zero value if no entries exist.
	NewestEntry time.Time
}

// SearchableMemory extends Memory with keyword search.
type SearchableMemory interface {
	Memory
	// Search returns entries whose message content contains the query string.
	// The search is case-insensitive substring matching.
	Search(ctx context.Context, query string) ([]Entry, error)
}

// PrunableMemory extends Memory with policy-based pruning.
type PrunableMemory interface {
	Memory
	// Prune removes entries that match the given policy and returns the
	// number of entries removed.
	Prune(ctx context.Context, policy PrunePolicy) (int, error)
}

// StatsMemory extends Memory with aggregate statistics.
type StatsMemory interface {
	Memory
	// Stats returns aggregate statistics about the memory store.
	Stats(ctx context.Context) (*Stats, error)
}
