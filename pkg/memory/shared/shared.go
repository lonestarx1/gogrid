package shared

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/memory"
)

// entry is the internal representation of a stored message with metadata.
type entry struct {
	msg       llm.Message
	createdAt time.Time
	size      int
}

// Memory is a thread-safe shared memory pool for GoGrid's team patterns.
// It implements memory.Memory, memory.SearchableMemory, memory.PrunableMemory,
// and memory.StatsMemory.
//
// Multiple agents can read and write concurrently. Change notifications are
// delivered via Go channels using non-blocking sends.
type Memory struct {
	mu          sync.RWMutex
	data        map[string][]entry
	subscribers []subscriber
	subMu       sync.Mutex
}

// New creates an empty shared memory pool.
func New() *Memory {
	return &Memory{
		data: make(map[string][]entry),
	}
}

// Subscribe registers a channel to receive change events.
// The buffer size of the channel determines how many events can queue.
// Events are dropped (non-blocking send) if the channel is full.
// Returns an unsubscribe function.
func (m *Memory) Subscribe(ch chan ChangeEvent) func() {
	m.subMu.Lock()
	sub := subscriber{ch: ch}
	m.subscribers = append(m.subscribers, sub)
	m.subMu.Unlock()

	return func() {
		m.subMu.Lock()
		defer m.subMu.Unlock()
		for i, s := range m.subscribers {
			if s.ch == ch {
				m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
				break
			}
		}
	}
}

// notify sends a change event to all subscribers using non-blocking send.
func (m *Memory) notify(event ChangeEvent) {
	m.subMu.Lock()
	subs := make([]subscriber, len(m.subscribers))
	copy(subs, m.subscribers)
	m.subMu.Unlock()

	for _, s := range subs {
		select {
		case s.ch <- event:
		default:
			// Drop event if channel is full.
		}
	}
}

// Load retrieves stored messages for the given key.
func (m *Memory) Load(_ context.Context, key string) ([]llm.Message, error) {
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

// Save stores messages under the given key and notifies subscribers.
func (m *Memory) Save(_ context.Context, key string, messages []llm.Message) error {
	m.mu.Lock()
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
	m.mu.Unlock()

	m.notify(ChangeEvent{Type: ChangeSave, Key: key, Timestamp: now})
	return nil
}

// Clear removes all messages for the given key and notifies subscribers.
func (m *Memory) Clear(_ context.Context, key string) error {
	m.mu.Lock()
	delete(m.data, key)
	m.mu.Unlock()

	m.notify(ChangeEvent{Type: ChangeClear, Key: key, Timestamp: time.Now()})
	return nil
}

// Search returns entries whose message content contains the query string.
func (m *Memory) Search(_ context.Context, query string) ([]memory.Entry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lower := strings.ToLower(query)
	var results []memory.Entry
	for key, entries := range m.data {
		for _, e := range entries {
			if strings.Contains(strings.ToLower(e.msg.Content), lower) {
				results = append(results, memory.Entry{
					Key:       key,
					Message:   e.msg,
					CreatedAt: e.createdAt,
					Size:      e.size,
				})
			}
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.Before(results[j].CreatedAt)
	})
	return results, nil
}

// Prune removes entries matching the given policy and returns the count removed.
func (m *Memory) Prune(_ context.Context, policy memory.PrunePolicy) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	removed := 0
	for key, entries := range m.data {
		var kept []entry
		for _, e := range entries {
			ext := memory.Entry{
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
func (m *Memory) Stats(_ context.Context) (*memory.Stats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s := &memory.Stats{Keys: len(m.data)}
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
