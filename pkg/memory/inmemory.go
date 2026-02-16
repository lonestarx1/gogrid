package memory

import (
	"context"
	"sync"

	"github.com/lonestarx1/gogrid/pkg/llm"
)

// InMemory is a thread-safe, in-memory implementation of Memory.
// Suitable for development, testing, and short-lived agent sessions.
// Data does not survive process restarts.
type InMemory struct {
	mu   sync.RWMutex
	data map[string][]llm.Message
}

// NewInMemory creates an empty in-memory store.
func NewInMemory() *InMemory {
	return &InMemory{
		data: make(map[string][]llm.Message),
	}
}

// Load retrieves stored messages for the given key.
func (m *InMemory) Load(_ context.Context, key string) ([]llm.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	msgs, ok := m.data[key]
	if !ok {
		return []llm.Message{}, nil
	}
	// Return a copy to prevent external mutation.
	cp := make([]llm.Message, len(msgs))
	copy(cp, msgs)
	return cp, nil
}

// Save stores messages under the given key.
func (m *InMemory) Save(_ context.Context, key string, messages []llm.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store a copy to prevent external mutation.
	cp := make([]llm.Message, len(messages))
	copy(cp, messages)
	m.data[key] = cp
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
