package transfer

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/memory"
)

func TestAcquireAndUse(t *testing.T) {
	ctx := context.Background()
	inner := memory.NewInMemory()
	state := NewState(inner)

	handle, err := state.Acquire("agent-a")
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	if handle.Owner() != "agent-a" {
		t.Errorf("Owner = %q, want %q", handle.Owner(), "agent-a")
	}

	// Save and load should work.
	msgs := []llm.Message{llm.NewUserMessage("hello")}
	if err := handle.Save(ctx, "k", msgs); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := handle.Load(ctx, "k")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Content != "hello" {
		t.Errorf("loaded = %v, want [hello]", loaded)
	}
}

func TestAcquireAlreadyOwned(t *testing.T) {
	inner := memory.NewInMemory()
	state := NewState(inner)

	_, err := state.Acquire("agent-a")
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}

	_, err = state.Acquire("agent-b")
	if !errors.Is(err, ErrAlreadyAcquired) {
		t.Errorf("second Acquire err = %v, want ErrAlreadyAcquired", err)
	}
}

func TestTransfer(t *testing.T) {
	ctx := context.Background()
	inner := memory.NewInMemory()
	state := NewState(inner)

	handleA, _ := state.Acquire("agent-a")

	// Save data as agent-a.
	_ = handleA.Save(ctx, "k", []llm.Message{llm.NewUserMessage("from a")})

	// Transfer to agent-b.
	handleB, err := state.Transfer("agent-a", "agent-b")
	if err != nil {
		t.Fatalf("Transfer: %v", err)
	}

	// agent-b can read what agent-a wrote.
	loaded, err := handleB.Load(ctx, "k")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Content != "from a" {
		t.Errorf("loaded = %v, want [from a]", loaded)
	}

	if state.Owner() != "agent-b" {
		t.Errorf("Owner = %q, want %q", state.Owner(), "agent-b")
	}
}

func TestStaleHandleRejected(t *testing.T) {
	ctx := context.Background()
	inner := memory.NewInMemory()
	state := NewState(inner)

	handleA, _ := state.Acquire("agent-a")
	_, _ = state.Transfer("agent-a", "agent-b")

	// handleA is now stale.
	_, err := handleA.Load(ctx, "k")
	if !errors.Is(err, ErrStateTransferred) {
		t.Errorf("stale Load err = %v, want ErrStateTransferred", err)
	}

	err = handleA.Save(ctx, "k", []llm.Message{llm.NewUserMessage("bad")})
	if !errors.Is(err, ErrStateTransferred) {
		t.Errorf("stale Save err = %v, want ErrStateTransferred", err)
	}

	err = handleA.Clear(ctx, "k")
	if !errors.Is(err, ErrStateTransferred) {
		t.Errorf("stale Clear err = %v, want ErrStateTransferred", err)
	}
}

func TestTransferWrongOwner(t *testing.T) {
	inner := memory.NewInMemory()
	state := NewState(inner)
	_, _ = state.Acquire("agent-a")

	_, err := state.Transfer("agent-b", "agent-c")
	if !errors.Is(err, ErrNotOwner) {
		t.Errorf("wrong owner Transfer err = %v, want ErrNotOwner", err)
	}
}

func TestChainTransfer(t *testing.T) {
	ctx := context.Background()
	inner := memory.NewInMemory()
	state := NewState(inner)

	h1, _ := state.Acquire("step-1")
	_ = h1.Save(ctx, "pipeline", []llm.Message{llm.NewUserMessage("step 1 data")})

	h2, _ := state.Transfer("step-1", "step-2")
	_ = h2.Save(ctx, "pipeline", []llm.Message{
		llm.NewUserMessage("step 1 data"),
		llm.NewUserMessage("step 2 data"),
	})

	h3, _ := state.Transfer("step-2", "step-3")
	loaded, _ := h3.Load(ctx, "pipeline")
	if len(loaded) != 2 {
		t.Fatalf("chain len = %d, want 2", len(loaded))
	}

	// All previous handles should be invalid.
	if _, err := h1.Load(ctx, "pipeline"); !errors.Is(err, ErrStateTransferred) {
		t.Errorf("h1 err = %v, want ErrStateTransferred", err)
	}
	if _, err := h2.Load(ctx, "pipeline"); !errors.Is(err, ErrStateTransferred) {
		t.Errorf("h2 err = %v, want ErrStateTransferred", err)
	}

	if state.Generation() != 3 {
		t.Errorf("Generation = %d, want 3", state.Generation())
	}
}

func TestValidationHook(t *testing.T) {
	inner := memory.NewInMemory()
	state := NewState(inner)

	hookErr := errors.New("transfer blocked by policy")
	state.OnTransfer(func(from, to string) error {
		if to == "blocked-agent" {
			return hookErr
		}
		return nil
	})

	_, _ = state.Acquire("agent-a")

	// Transfer to blocked agent should fail.
	_, err := state.Transfer("agent-a", "blocked-agent")
	if err == nil {
		t.Fatal("expected hook error, got nil")
	}
	if !errors.Is(err, hookErr) {
		t.Errorf("err = %v, want wrapped %v", err, hookErr)
	}

	// Owner should not change on failed transfer.
	if state.Owner() != "agent-a" {
		t.Errorf("Owner after failed transfer = %q, want %q", state.Owner(), "agent-a")
	}

	// Transfer to allowed agent should succeed.
	_, err = state.Transfer("agent-a", "allowed-agent")
	if err != nil {
		t.Fatalf("allowed transfer: %v", err)
	}
}

func TestAuditLog(t *testing.T) {
	inner := memory.NewInMemory()
	state := NewState(inner)

	_, _ = state.Acquire("a")
	_, _ = state.Transfer("a", "b")
	_, _ = state.Transfer("b", "c")

	log := state.AuditLog()
	if len(log) != 3 {
		t.Fatalf("audit log len = %d, want 3", len(log))
	}

	// First entry: acquire.
	if log[0].From != "" || log[0].To != "a" || log[0].Generation != 1 {
		t.Errorf("audit[0] = %+v", log[0])
	}
	// Second entry: a -> b.
	if log[1].From != "a" || log[1].To != "b" || log[1].Generation != 2 {
		t.Errorf("audit[1] = %+v", log[1])
	}
	// Third entry: b -> c.
	if log[2].From != "b" || log[2].To != "c" || log[2].Generation != 3 {
		t.Errorf("audit[2] = %+v", log[2])
	}

	// Mutating the returned log should not affect the internal state.
	log[0].To = "mutated"
	fresh := state.AuditLog()
	if fresh[0].To != "a" {
		t.Errorf("audit log not isolated: %+v", fresh[0])
	}
}

func TestHandleImplementsMemory(t *testing.T) {
	inner := memory.NewInMemory()
	state := NewState(inner)
	handle, _ := state.Acquire("agent")
	var _ memory.Memory = handle
}

func TestConcurrentTransfer(t *testing.T) {
	inner := memory.NewInMemory()
	state := NewState(inner)
	_, _ = state.Acquire("start")

	// Multiple goroutines try to transfer concurrently.
	// Only one should succeed per round.
	var wg sync.WaitGroup
	successes := make(chan string, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			owner := state.Owner()
			_, err := state.Transfer(owner, "contender")
			if err == nil {
				successes <- "ok"
			}
		}(i)
	}
	wg.Wait()
	close(successes)

	// At least one transfer should succeed (possibly more if they happen serially).
	count := 0
	for range successes {
		count++
	}
	if count == 0 {
		t.Error("no concurrent transfers succeeded")
	}
}
