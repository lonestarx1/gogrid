package shared

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/memory"
)

func TestSaveAndLoad(t *testing.T) {
	ctx := context.Background()
	m := New()

	msgs := []llm.Message{
		llm.NewUserMessage("hello"),
		llm.NewAssistantMessage("hi"),
	}

	if err := m.Save(ctx, "conv", msgs); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := m.Load(ctx, "conv")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("len = %d, want 2", len(loaded))
	}
	if loaded[0].Content != "hello" {
		t.Errorf("content = %q, want %q", loaded[0].Content, "hello")
	}
}

func TestLoadEmpty(t *testing.T) {
	m := New()
	loaded, err := m.Load(context.Background(), "nope")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded == nil {
		t.Fatal("Load returned nil, want empty slice")
	}
	if len(loaded) != 0 {
		t.Errorf("len = %d, want 0", len(loaded))
	}
}

func TestClear(t *testing.T) {
	ctx := context.Background()
	m := New()
	_ = m.Save(ctx, "k", []llm.Message{llm.NewUserMessage("data")})

	if err := m.Clear(ctx, "k"); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	loaded, _ := m.Load(ctx, "k")
	if len(loaded) != 0 {
		t.Errorf("after Clear len = %d, want 0", len(loaded))
	}
}

func TestSubscribeNotifications(t *testing.T) {
	ctx := context.Background()
	m := New()

	ch := make(chan ChangeEvent, 10)
	unsub := m.Subscribe(ch)
	defer unsub()

	_ = m.Save(ctx, "k1", []llm.Message{llm.NewUserMessage("hello")})
	_ = m.Clear(ctx, "k1")

	// Should have received 2 events.
	select {
	case ev := <-ch:
		if ev.Type != ChangeSave || ev.Key != "k1" {
			t.Errorf("event 1 = %+v, want Save/k1", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for save event")
	}

	select {
	case ev := <-ch:
		if ev.Type != ChangeClear || ev.Key != "k1" {
			t.Errorf("event 2 = %+v, want Clear/k1", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for clear event")
	}
}

func TestUnsubscribe(t *testing.T) {
	ctx := context.Background()
	m := New()

	ch := make(chan ChangeEvent, 10)
	unsub := m.Subscribe(ch)
	unsub()

	_ = m.Save(ctx, "k", []llm.Message{llm.NewUserMessage("hello")})

	select {
	case ev := <-ch:
		t.Errorf("received event after unsubscribe: %+v", ev)
	default:
		// Expected: no event.
	}
}

func TestNonBlockingSend(t *testing.T) {
	ctx := context.Background()
	m := New()

	// Unbuffered channel — sends should be dropped, not block.
	ch := make(chan ChangeEvent)
	unsub := m.Subscribe(ch)
	defer unsub()

	// This should not block.
	_ = m.Save(ctx, "k", []llm.Message{llm.NewUserMessage("hello")})
}

func TestNamespacedView(t *testing.T) {
	ctx := context.Background()
	pool := New()

	agent1 := NewNamespacedView(pool, "agent1")
	agent2 := NewNamespacedView(pool, "agent2")

	_ = agent1.Save(ctx, "conv", []llm.Message{llm.NewUserMessage("from agent1")})
	_ = agent2.Save(ctx, "conv", []llm.Message{llm.NewUserMessage("from agent2")})

	// Each agent sees its own data.
	loaded1, _ := agent1.Load(ctx, "conv")
	loaded2, _ := agent2.Load(ctx, "conv")

	if len(loaded1) != 1 || loaded1[0].Content != "from agent1" {
		t.Errorf("agent1 = %v, want [from agent1]", loaded1)
	}
	if len(loaded2) != 1 || loaded2[0].Content != "from agent2" {
		t.Errorf("agent2 = %v, want [from agent2]", loaded2)
	}

	// Clear only affects namespaced key.
	_ = agent1.Clear(ctx, "conv")
	loaded1, _ = agent1.Load(ctx, "conv")
	loaded2, _ = agent2.Load(ctx, "conv")
	if len(loaded1) != 0 {
		t.Errorf("agent1 after clear len = %d, want 0", len(loaded1))
	}
	if len(loaded2) != 1 {
		t.Errorf("agent2 after agent1 clear len = %d, want 1", len(loaded2))
	}
}

func TestNamespacedViewNotifications(t *testing.T) {
	ctx := context.Background()
	pool := New()

	ch := make(chan ChangeEvent, 10)
	unsub := pool.Subscribe(ch)
	defer unsub()

	agent := NewNamespacedView(pool, "ns")
	_ = agent.Save(ctx, "k", []llm.Message{llm.NewUserMessage("data")})

	select {
	case ev := <-ch:
		if ev.Key != "ns/k" {
			t.Errorf("event key = %q, want %q", ev.Key, "ns/k")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestSearch(t *testing.T) {
	ctx := context.Background()
	m := New()

	_ = m.Save(ctx, "a", []llm.Message{llm.NewUserMessage("hello world")})
	_ = m.Save(ctx, "b", []llm.Message{llm.NewUserMessage("foo bar")})

	results, err := m.Search(ctx, "hello")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Search len = %d, want 1", len(results))
	}
}

func TestPrune(t *testing.T) {
	ctx := context.Background()
	m := New()

	_ = m.Save(ctx, "k", []llm.Message{
		llm.NewUserMessage("short"),
		llm.NewUserMessage("a longer message for pruning"),
	})

	removed, err := m.Prune(ctx, memory.NewMaxSize(10))
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}

	loaded, _ := m.Load(ctx, "k")
	if len(loaded) != 1 {
		t.Errorf("after prune len = %d, want 1", len(loaded))
	}
}

func TestStats(t *testing.T) {
	ctx := context.Background()
	m := New()

	_ = m.Save(ctx, "a", []llm.Message{
		llm.NewUserMessage("hello"),
		llm.NewAssistantMessage("world"),
	})

	s, err := m.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if s.Keys != 1 {
		t.Errorf("Keys = %d, want 1", s.Keys)
	}
	if s.TotalEntries != 2 {
		t.Errorf("TotalEntries = %d, want 2", s.TotalEntries)
	}
	if s.TotalSize != 10 {
		t.Errorf("TotalSize = %d, want 10", s.TotalSize)
	}
}

func TestConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	m := New()

	ch := make(chan ChangeEvent, 100)
	unsub := m.Subscribe(ch)
	defer unsub()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msgs := []llm.Message{llm.NewUserMessage("data")}
			_ = m.Save(ctx, "shared", msgs)
			_, _ = m.Load(ctx, "shared")
			_, _ = m.Search(ctx, "data")
			_, _ = m.Stats(ctx)
		}()
	}
	wg.Wait()
}

func TestImplementsInterfaces(t *testing.T) {
	m := New()
	var _ memory.Memory = m
	var _ memory.SearchableMemory = m
	var _ memory.PrunableMemory = m
	var _ memory.StatsMemory = m
}

func TestNamespacedViewImplementsMemory(t *testing.T) {
	v := NewNamespacedView(New(), "ns")
	var _ memory.Memory = v
}

func BenchmarkSharedMemorySaveLoad(b *testing.B) {
	ctx := context.Background()
	m := New()
	msgs := []llm.Message{llm.NewUserMessage("benchmark data")}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = m.Save(ctx, "bench", msgs)
			_, _ = m.Load(ctx, "bench")
		}
	})
}
