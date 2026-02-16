package file

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/memory"
)

// metadata is the sidecar structure persisted alongside each key's data.
type metadata struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	EntrySize []int     `json:"entry_sizes"`
}

// Memory is a file-backed implementation of memory.Memory.
// Each key maps to a JSON file containing the messages and a .meta.json
// sidecar containing timestamps and size information.
//
// Filenames are hex-encoded to avoid filesystem issues with special characters.
// Memory is safe for concurrent use.
type Memory struct {
	mu  sync.RWMutex
	dir string
}

// New creates a file-backed memory store at the given directory.
// The directory is created if it does not exist.
func New(dir string) (*Memory, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("file memory: create dir: %w", err)
	}
	return &Memory{dir: dir}, nil
}

// Load retrieves stored messages for the given key.
func (m *Memory) Load(_ context.Context, key string) ([]llm.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := os.ReadFile(m.dataPath(key))
	if os.IsNotExist(err) {
		return []llm.Message{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("file memory: read %q: %w", key, err)
	}

	var msgs []llm.Message
	if err := json.Unmarshal(data, &msgs); err != nil {
		return nil, fmt.Errorf("file memory: decode %q: %w", key, err)
	}
	return msgs, nil
}

// Save stores messages under the given key.
func (m *Memory) Save(_ context.Context, key string, messages []llm.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("file memory: encode %q: %w", key, err)
	}

	if err := os.WriteFile(m.dataPath(key), data, 0o644); err != nil {
		return fmt.Errorf("file memory: write %q: %w", key, err)
	}

	// Build and write metadata sidecar.
	now := time.Now()
	meta := metadata{
		UpdatedAt: now,
		EntrySize: make([]int, len(messages)),
	}

	// Preserve original created_at if the file already exists.
	existing, err := m.loadMeta(key)
	if err == nil {
		meta.CreatedAt = existing.CreatedAt
	} else {
		meta.CreatedAt = now
	}

	for i, msg := range messages {
		meta.EntrySize[i] = len(msg.Content)
	}

	metaData, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("file memory: encode meta %q: %w", key, err)
	}
	if err := os.WriteFile(m.metaPath(key), metaData, 0o644); err != nil {
		return fmt.Errorf("file memory: write meta %q: %w", key, err)
	}

	return nil
}

// Clear removes all messages for the given key.
func (m *Memory) Clear(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove data file.
	if err := os.Remove(m.dataPath(key)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("file memory: remove %q: %w", key, err)
	}
	// Remove metadata sidecar.
	if err := os.Remove(m.metaPath(key)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("file memory: remove meta %q: %w", key, err)
	}
	return nil
}

// Search returns entries whose message content contains the query string.
func (m *Memory) Search(_ context.Context, query string) ([]memory.Entry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys, err := m.listKeys()
	if err != nil {
		return nil, err
	}

	lower := strings.ToLower(query)
	var results []memory.Entry
	for _, key := range keys {
		msgs, meta, err := m.loadKeyData(key)
		if err != nil {
			return nil, err
		}
		for i, msg := range msgs {
			if strings.Contains(strings.ToLower(msg.Content), lower) {
				size := 0
				if i < len(meta.EntrySize) {
					size = meta.EntrySize[i]
				}
				results = append(results, memory.Entry{
					Key:       key,
					Message:   msg,
					CreatedAt: meta.CreatedAt,
					Size:      size,
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

	keys, err := m.listKeys()
	if err != nil {
		return 0, err
	}

	removed := 0
	for _, key := range keys {
		msgs, meta, err := m.loadKeyData(key)
		if err != nil {
			return removed, err
		}

		var kept []llm.Message
		var keptSizes []int
		for i, msg := range msgs {
			size := 0
			if i < len(meta.EntrySize) {
				size = meta.EntrySize[i]
			}
			e := memory.Entry{
				Key:       key,
				Message:   msg,
				CreatedAt: meta.CreatedAt,
				Size:      size,
			}
			if policy.ShouldPrune(e) {
				removed++
			} else {
				kept = append(kept, msg)
				keptSizes = append(keptSizes, size)
			}
		}

		if len(kept) == 0 {
			// Remove both files; ignore not-exist errors.
			if err := os.Remove(m.dataPath(key)); err != nil && !os.IsNotExist(err) {
				return removed, fmt.Errorf("file memory: remove %q: %w", key, err)
			}
			if err := os.Remove(m.metaPath(key)); err != nil && !os.IsNotExist(err) {
				return removed, fmt.Errorf("file memory: remove meta %q: %w", key, err)
			}
			continue
		}

		if len(kept) < len(msgs) {
			// Rewrite with remaining entries.
			data, err := json.Marshal(kept)
			if err != nil {
				return removed, fmt.Errorf("file memory: encode %q: %w", key, err)
			}
			if err := os.WriteFile(m.dataPath(key), data, 0o644); err != nil {
				return removed, fmt.Errorf("file memory: write %q: %w", key, err)
			}
			meta.EntrySize = keptSizes
			meta.UpdatedAt = time.Now()
			metaData, err := json.Marshal(meta)
			if err != nil {
				return removed, fmt.Errorf("file memory: encode meta %q: %w", key, err)
			}
			if err := os.WriteFile(m.metaPath(key), metaData, 0o644); err != nil {
				return removed, fmt.Errorf("file memory: write meta %q: %w", key, err)
			}
		}
	}
	return removed, nil
}

// Stats returns aggregate statistics about the memory store.
func (m *Memory) Stats(_ context.Context) (*memory.Stats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys, err := m.listKeys()
	if err != nil {
		return nil, err
	}

	s := &memory.Stats{Keys: len(keys)}
	for _, key := range keys {
		msgs, meta, err := m.loadKeyData(key)
		if err != nil {
			return nil, err
		}
		s.TotalEntries += len(msgs)
		for i := range msgs {
			size := 0
			if i < len(meta.EntrySize) {
				size = meta.EntrySize[i]
			}
			s.TotalSize += int64(size)
		}
		if !meta.CreatedAt.IsZero() {
			if s.OldestEntry.IsZero() || meta.CreatedAt.Before(s.OldestEntry) {
				s.OldestEntry = meta.CreatedAt
			}
			if s.NewestEntry.IsZero() || meta.CreatedAt.After(s.NewestEntry) {
				s.NewestEntry = meta.CreatedAt
			}
		}
	}
	return s, nil
}

// dataPath returns the file path for a key's message data.
func (m *Memory) dataPath(key string) string {
	return filepath.Join(m.dir, hex.EncodeToString([]byte(key))+".json")
}

// metaPath returns the file path for a key's metadata sidecar.
func (m *Memory) metaPath(key string) string {
	return filepath.Join(m.dir, hex.EncodeToString([]byte(key))+".meta.json")
}

// loadMeta reads and decodes the metadata sidecar for a key.
// Caller must hold at least a read lock.
func (m *Memory) loadMeta(key string) (*metadata, error) {
	data, err := os.ReadFile(m.metaPath(key))
	if err != nil {
		return nil, err
	}
	var meta metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// loadKeyData reads messages and metadata for a key.
// Caller must hold at least a read lock.
func (m *Memory) loadKeyData(key string) ([]llm.Message, *metadata, error) {
	data, err := os.ReadFile(m.dataPath(key))
	if err != nil {
		return nil, nil, fmt.Errorf("file memory: read %q: %w", key, err)
	}
	var msgs []llm.Message
	if err := json.Unmarshal(data, &msgs); err != nil {
		return nil, nil, fmt.Errorf("file memory: decode %q: %w", key, err)
	}
	meta, err := m.loadMeta(key)
	if err != nil {
		// If metadata is missing, use a zero-value.
		meta = &metadata{}
	}
	return msgs, meta, nil
}

// listKeys returns all keys that have stored data files.
// Caller must hold at least a read lock.
func (m *Memory) listKeys() ([]string, error) {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil, fmt.Errorf("file memory: list dir: %w", err)
	}

	var keys []string
	for _, e := range entries {
		name := e.Name()
		// Only consider data files (not .meta.json).
		if strings.HasSuffix(name, ".meta.json") {
			continue
		}
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		hexKey := strings.TrimSuffix(name, ".json")
		decoded, err := hex.DecodeString(hexKey)
		if err != nil {
			continue // Skip files with non-hex names.
		}
		keys = append(keys, string(decoded))
	}
	return keys, nil
}
