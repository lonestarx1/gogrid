package memory

import (
	"context"
	"testing"

	"github.com/lonestarx1/gogrid/pkg/llm"
)

func TestInMemorySaveAndLoad(t *testing.T) {
	ctx := context.Background()
	m := NewInMemory()
	msgs := []llm.Message{
		llm.NewUserMessage("hello"),
		llm.NewAssistantMessage("hi"),
	}

	if err := m.Save(ctx, "conv-1", msgs); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := m.Load(ctx, "conv-1")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("Load len = %d, want 2", len(loaded))
	}
	if loaded[0].Content != "hello" {
		t.Errorf("loaded[0].Content = %q, want %q", loaded[0].Content, "hello")
	}
	if loaded[1].Content != "hi" {
		t.Errorf("loaded[1].Content = %q, want %q", loaded[1].Content, "hi")
	}
}

func TestInMemoryLoadEmpty(t *testing.T) {
	ctx := context.Background()
	m := NewInMemory()

	loaded, err := m.Load(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded == nil {
		t.Fatal("Load returned nil, want empty slice")
	}
	if len(loaded) != 0 {
		t.Errorf("Load len = %d, want 0", len(loaded))
	}
}

func TestInMemoryClear(t *testing.T) {
	ctx := context.Background()
	m := NewInMemory()
	msgs := []llm.Message{llm.NewUserMessage("hello")}

	_ = m.Save(ctx, "conv-1", msgs)

	if err := m.Clear(ctx, "conv-1"); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	loaded, err := m.Load(ctx, "conv-1")
	if err != nil {
		t.Fatalf("Load after Clear: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("Load after Clear len = %d, want 0", len(loaded))
	}
}

func TestInMemoryIsolation(t *testing.T) {
	ctx := context.Background()
	m := NewInMemory()

	msgs := []llm.Message{llm.NewUserMessage("original")}
	_ = m.Save(ctx, "conv-1", msgs)

	// Mutating the original slice should not affect stored data.
	msgs[0] = llm.NewUserMessage("mutated")

	loaded, _ := m.Load(ctx, "conv-1")
	if loaded[0].Content != "original" {
		t.Errorf("Content = %q, want %q (save did not copy)", loaded[0].Content, "original")
	}

	// Mutating the loaded slice should not affect stored data.
	loaded[0] = llm.NewUserMessage("also mutated")

	reloaded, _ := m.Load(ctx, "conv-1")
	if reloaded[0].Content != "original" {
		t.Errorf("Content = %q, want %q (load did not copy)", reloaded[0].Content, "original")
	}
}

func TestInMemoryMultipleKeys(t *testing.T) {
	ctx := context.Background()
	m := NewInMemory()

	_ = m.Save(ctx, "a", []llm.Message{llm.NewUserMessage("alpha")})
	_ = m.Save(ctx, "b", []llm.Message{llm.NewUserMessage("beta")})

	a, _ := m.Load(ctx, "a")
	b, _ := m.Load(ctx, "b")

	if a[0].Content != "alpha" {
		t.Errorf("a = %q, want %q", a[0].Content, "alpha")
	}
	if b[0].Content != "beta" {
		t.Errorf("b = %q, want %q", b[0].Content, "beta")
	}
}

func TestInMemoryKeys(t *testing.T) {
	m := NewInMemory()
	ctx := context.Background()

	_ = m.Save(ctx, "x", []llm.Message{llm.NewUserMessage("1")})
	_ = m.Save(ctx, "y", []llm.Message{llm.NewUserMessage("2")})

	keys := m.Keys()
	if len(keys) != 2 {
		t.Fatalf("Keys len = %d, want 2", len(keys))
	}

	keySet := map[string]bool{}
	for _, k := range keys {
		keySet[k] = true
	}
	if !keySet["x"] || !keySet["y"] {
		t.Errorf("Keys = %v, want [x, y]", keys)
	}
}

func TestReadOnlyLoad(t *testing.T) {
	ctx := context.Background()
	inner := NewInMemory()
	_ = inner.Save(ctx, "k", []llm.Message{llm.NewUserMessage("data")})

	ro := NewReadOnly(inner)

	loaded, err := ro.Load(ctx, "k")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded[0].Content != "data" {
		t.Errorf("Content = %q, want %q", loaded[0].Content, "data")
	}
}

func TestReadOnlySaveBlocked(t *testing.T) {
	ro := NewReadOnly(NewInMemory())
	err := ro.Save(context.Background(), "k", []llm.Message{llm.NewUserMessage("x")})
	if err != ErrReadOnly {
		t.Errorf("Save err = %v, want ErrReadOnly", err)
	}
}

func TestReadOnlyClearBlocked(t *testing.T) {
	ro := NewReadOnly(NewInMemory())
	err := ro.Clear(context.Background(), "k")
	if err != ErrReadOnly {
		t.Errorf("Clear err = %v, want ErrReadOnly", err)
	}
}

func TestInMemorySearch(t *testing.T) {
	ctx := context.Background()
	m := NewInMemory()

	_ = m.Save(ctx, "a", []llm.Message{
		llm.NewUserMessage("hello world"),
		llm.NewAssistantMessage("goodbye world"),
	})
	_ = m.Save(ctx, "b", []llm.Message{
		llm.NewUserMessage("foo bar"),
	})

	tests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "match multiple", query: "world", want: 2},
		{name: "case insensitive", query: "HELLO", want: 1},
		{name: "no match", query: "xyz", want: 0},
		{name: "partial match", query: "oo", want: 2}, // "goodbye" and "foo"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := m.Search(ctx, tt.query)
			if err != nil {
				t.Fatalf("Search: %v", err)
			}
			if len(results) != tt.want {
				t.Errorf("Search(%q) returned %d results, want %d", tt.query, len(results), tt.want)
			}
		})
	}
}

func TestInMemoryPrune(t *testing.T) {
	ctx := context.Background()
	m := NewInMemory()

	_ = m.Save(ctx, "k", []llm.Message{
		llm.NewUserMessage("short"),
		llm.NewUserMessage("this is a much longer message for testing size pruning"),
	})

	// Prune entries larger than 10 bytes.
	removed, err := m.Prune(ctx, NewMaxSize(10))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if removed != 1 {
		t.Errorf("Prune removed = %d, want 1", removed)
	}

	loaded, _ := m.Load(ctx, "k")
	if len(loaded) != 1 {
		t.Fatalf("after Prune len = %d, want 1", len(loaded))
	}
	if loaded[0].Content != "short" {
		t.Errorf("remaining = %q, want %q", loaded[0].Content, "short")
	}
}

func TestInMemoryPruneClearsEmptyKey(t *testing.T) {
	ctx := context.Background()
	m := NewInMemory()

	_ = m.Save(ctx, "k", []llm.Message{llm.NewUserMessage("x")})

	// Prune everything (size limit 0).
	removed, err := m.Prune(ctx, NewMaxSize(0))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}
	if len(m.Keys()) != 0 {
		t.Errorf("Keys after full prune = %v, want empty", m.Keys())
	}
}

func TestInMemoryStats(t *testing.T) {
	ctx := context.Background()
	m := NewInMemory()

	// Empty store.
	s, err := m.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if s.Keys != 0 || s.TotalEntries != 0 || s.TotalSize != 0 {
		t.Errorf("empty stats = %+v, want zeros", s)
	}

	_ = m.Save(ctx, "a", []llm.Message{
		llm.NewUserMessage("hello"),      // 5 bytes
		llm.NewAssistantMessage("world"), // 5 bytes
	})
	_ = m.Save(ctx, "b", []llm.Message{
		llm.NewUserMessage("test"), // 4 bytes
	})

	s, err = m.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if s.Keys != 2 {
		t.Errorf("Keys = %d, want 2", s.Keys)
	}
	if s.TotalEntries != 3 {
		t.Errorf("TotalEntries = %d, want 3", s.TotalEntries)
	}
	if s.TotalSize != 14 {
		t.Errorf("TotalSize = %d, want 14", s.TotalSize)
	}
	if s.OldestEntry.IsZero() {
		t.Error("OldestEntry is zero")
	}
	if s.NewestEntry.IsZero() {
		t.Error("NewestEntry is zero")
	}
}

func TestInMemoryImplementsExtendedInterfaces(t *testing.T) {
	m := NewInMemory()
	var _ SearchableMemory = m
	var _ PrunableMemory = m
	var _ StatsMemory = m
}
