package file

import (
	"context"
	"sync"
	"testing"

	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/memory"
)

func newTestMemory(t *testing.T) *Memory {
	t.Helper()
	m, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return m
}

func TestSaveAndLoad(t *testing.T) {
	ctx := context.Background()
	m := newTestMemory(t)

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

func TestLoadEmpty(t *testing.T) {
	ctx := context.Background()
	m := newTestMemory(t)

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

func TestClear(t *testing.T) {
	ctx := context.Background()
	m := newTestMemory(t)

	_ = m.Save(ctx, "k", []llm.Message{llm.NewUserMessage("data")})

	if err := m.Clear(ctx, "k"); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	loaded, err := m.Load(ctx, "k")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("Load after Clear len = %d, want 0", len(loaded))
	}
}

func TestClearNonexistent(t *testing.T) {
	m := newTestMemory(t)
	if err := m.Clear(context.Background(), "nope"); err != nil {
		t.Fatalf("Clear nonexistent: %v", err)
	}
}

func TestPersistence(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	// Write with one instance.
	m1, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_ = m1.Save(ctx, "k", []llm.Message{llm.NewUserMessage("persisted")})

	// Read with a new instance pointing at the same directory.
	m2, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	loaded, err := m2.Load(ctx, "k")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Content != "persisted" {
		t.Errorf("persisted data = %v, want [persisted]", loaded)
	}
}

func TestSpecialCharKeys(t *testing.T) {
	ctx := context.Background()
	m := newTestMemory(t)

	keys := []string{
		"key with spaces",
		"key/with/slashes",
		"key.with.dots",
		"键中文",
		"emoji🔑",
	}

	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			msgs := []llm.Message{llm.NewUserMessage("data for " + key)}
			if err := m.Save(ctx, key, msgs); err != nil {
				t.Fatalf("Save: %v", err)
			}
			loaded, err := m.Load(ctx, key)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if len(loaded) != 1 || loaded[0].Content != "data for "+key {
				t.Errorf("loaded = %v, want data for %s", loaded, key)
			}
		})
	}
}

func TestSearch(t *testing.T) {
	ctx := context.Background()
	m := newTestMemory(t)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := m.Search(ctx, tt.query)
			if err != nil {
				t.Fatalf("Search: %v", err)
			}
			if len(results) != tt.want {
				t.Errorf("Search(%q) = %d results, want %d", tt.query, len(results), tt.want)
			}
		})
	}
}

func TestPrune(t *testing.T) {
	ctx := context.Background()
	m := newTestMemory(t)

	_ = m.Save(ctx, "k", []llm.Message{
		llm.NewUserMessage("short"),
		llm.NewUserMessage("this is a much longer message for size pruning"),
	})

	removed, err := m.Prune(ctx, memory.NewMaxSize(10))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}

	loaded, _ := m.Load(ctx, "k")
	if len(loaded) != 1 || loaded[0].Content != "short" {
		t.Errorf("after prune = %v, want [short]", loaded)
	}
}

func TestPruneClearsEmptyKey(t *testing.T) {
	ctx := context.Background()
	m := newTestMemory(t)

	_ = m.Save(ctx, "k", []llm.Message{llm.NewUserMessage("x")})

	removed, err := m.Prune(ctx, memory.NewMaxSize(0))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}

	// Key should be gone (data file removed).
	loaded, err := m.Load(ctx, "k")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("after full prune len = %d, want 0", len(loaded))
	}
}

func TestStats(t *testing.T) {
	ctx := context.Background()
	m := newTestMemory(t)

	// Empty store.
	s, err := m.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if s.Keys != 0 || s.TotalEntries != 0 {
		t.Errorf("empty stats = %+v, want zeros", s)
	}

	_ = m.Save(ctx, "a", []llm.Message{
		llm.NewUserMessage("hello"),
		llm.NewAssistantMessage("world"),
	})
	_ = m.Save(ctx, "b", []llm.Message{
		llm.NewUserMessage("test"),
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
}

func TestConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	m := newTestMemory(t)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "concurrent"
			msgs := []llm.Message{llm.NewUserMessage("msg")}
			_ = m.Save(ctx, key, msgs)
			_, _ = m.Load(ctx, key)
			_, _ = m.Search(ctx, "msg")
			_, _ = m.Stats(ctx)
		}(i)
	}
	wg.Wait()
}

func TestImplementsInterfaces(t *testing.T) {
	m := &Memory{}
	var _ memory.Memory = m
	var _ memory.SearchableMemory = m
	var _ memory.PrunableMemory = m
	var _ memory.StatsMemory = m
}
