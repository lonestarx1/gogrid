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
