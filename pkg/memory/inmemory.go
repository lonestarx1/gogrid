package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lonestarx1/gogrid/pkg/llm"
)

// entry is the internal representation of a stored message with metadata.
type entry struct {
	msg       llm.Message
	createdAt time.Time
	size      int
}

// InMemory is a thread-safe, in-memory implementation of Memory.
// It also implements SearchableMemory, PrunableMemory, and StatsMemory.
// Suitable for development, testing, and short-lived agent sessions.
// Data does not survive process restarts.
type InMemory struct {
	mu   sync.RWMutex
	data map[string][]entry
}

// NewInMemory creates an empty in-memory store.
func NewInMemory() *InMemory {
	return &InMemory{
		data: make(map[string][]entry),
	}
}

// Load retrieves stored messages for the given key.
func (m *InMemory) Load(_ context.Context, key string) ([]llm.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries, ok := m.data[key]
	if !ok {
		return []llm.Message{}, nil
	}
	msgs := make([]llm.Message, len(entries))
	for i, e := range entries {
		msgs[i] = e.msg
	}
	return msgs, nil
}

// Save stores messages under the given key.
func (m *InMemory) Save(_ context.Context, key string, messages []llm.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	entries := make([]entry, len(messages))
	for i, msg := range messages {
		entries[i] = entry{
			msg:       msg,
			createdAt: now,
			size:      len(msg.Content),
		}
	}
	m.data[key] = entries
	return nil
}

// Clear removes all messages for the given key.
func (m *InMemory) Clear(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

// Keys returns all keys that have stored messages.
func (m *InMemory) Keys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

// Search returns entries whose message content contains the query string.
// The search is case-insensitive substring matching across all keys.
func (m *InMemory) Search(_ context.Context, query string) ([]Entry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lower := strings.ToLower(query)
	var results []Entry
	for key, entries := range m.data {
		for _, e := range entries {
			if strings.Contains(strings.ToLower(e.msg.Content), lower) {
				results = append(results, Entry{
					Key:       key,
					Message:   e.msg,
					CreatedAt: e.createdAt,
					Size:      e.size,
				})
			}
		}
	}
	// Sort by creation time for deterministic output.
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.Before(results[j].CreatedAt)
	})
	return results, nil
}

// Prune removes entries matching the given policy and returns the count removed.
// For MaxEntries policies, entries are evaluated oldest-first within each key.
func (m *InMemory) Prune(_ context.Context, policy PrunePolicy) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	removed := 0
	for key, entries := range m.data {
		var kept []entry
		for _, e := range entries {
			ext := Entry{
				Key:       key,
				Message:   e.msg,
				CreatedAt: e.createdAt,
				Size:      e.size,
			}
			if policy.ShouldPrune(ext) {
				removed++
			} else {
				kept = append(kept, e)
			}
		}
		if len(kept) == 0 {
			delete(m.data, key)
		} else {
			m.data[key] = kept
		}
	}
	return removed, nil
}

// Stats returns aggregate statistics about the memory store.
func (m *InMemory) Stats(_ context.Context) (*Stats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s := &Stats{
		Keys: len(m.data),
	}
	for _, entries := range m.data {
		s.TotalEntries += len(entries)
		for _, e := range entries {
			s.TotalSize += int64(e.size)
			if s.OldestEntry.IsZero() || e.createdAt.Before(s.OldestEntry) {
				s.OldestEntry = e.createdAt
			}
			if s.NewestEntry.IsZero() || e.createdAt.After(s.NewestEntry) {
				s.NewestEntry = e.createdAt
			}
		}
	}
	return s, nil
}
