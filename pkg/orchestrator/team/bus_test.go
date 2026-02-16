package team

import (
	"sync"
	"testing"
	"time"
)

func TestBusPublishAndSubscribe(t *testing.T) {
	bus := NewBus()

	ch, unsub := bus.Subscribe("topic-a", 10)
	defer unsub()

	bus.Publish(Message{From: "agent-1", Topic: "topic-a", Content: "hello"})

	select {
	case msg := <-ch:
		if msg.From != "agent-1" {
			t.Errorf("From = %q, want %q", msg.From, "agent-1")
		}
		if msg.Content != "hello" {
			t.Errorf("Content = %q, want %q", msg.Content, "hello")
		}
		if msg.Timestamp.IsZero() {
			t.Error("Timestamp is zero")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestBusMultipleSubscribers(t *testing.T) {
	bus := NewBus()

	ch1, unsub1 := bus.Subscribe("topic", 10)
	defer unsub1()
	ch2, unsub2 := bus.Subscribe("topic", 10)
	defer unsub2()

	bus.Publish(Message{From: "sender", Topic: "topic", Content: "broadcast"})

	for _, ch := range []<-chan Message{ch1, ch2} {
		select {
		case msg := <-ch:
			if msg.Content != "broadcast" {
				t.Errorf("Content = %q, want %q", msg.Content, "broadcast")
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for message")
		}
	}
}

func TestBusMultipleTopics(t *testing.T) {
	bus := NewBus()

	chA, unsubA := bus.Subscribe("topic-a", 10)
	defer unsubA()
	chB, unsubB := bus.Subscribe("topic-b", 10)
	defer unsubB()

	bus.Publish(Message{From: "sender", Topic: "topic-a", Content: "for A"})
	bus.Publish(Message{From: "sender", Topic: "topic-b", Content: "for B"})

	select {
	case msg := <-chA:
		if msg.Content != "for A" {
			t.Errorf("topic-a content = %q, want %q", msg.Content, "for A")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout on topic-a")
	}

	select {
	case msg := <-chB:
		if msg.Content != "for B" {
			t.Errorf("topic-b content = %q, want %q", msg.Content, "for B")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout on topic-b")
	}

	// topic-a subscriber should not get topic-b messages.
	select {
	case msg := <-chA:
		t.Errorf("topic-a received unexpected message: %+v", msg)
	default:
	}
}

func TestBusUnsubscribe(t *testing.T) {
	bus := NewBus()

	ch, unsub := bus.Subscribe("topic", 10)
	unsub()

	bus.Publish(Message{From: "sender", Topic: "topic", Content: "after unsub"})

	select {
	case msg := <-ch:
		t.Errorf("received message after unsubscribe: %+v", msg)
	default:
	}
}

func TestBusNonBlockingSend(t *testing.T) {
	bus := NewBus()

	// Unbuffered channel — sends should be dropped, not block.
	ch, unsub := bus.Subscribe("topic", 0)
	defer unsub()

	// This must not block.
	bus.Publish(Message{From: "sender", Topic: "topic", Content: "dropped"})

	select {
	case <-ch:
		t.Error("expected message to be dropped on unbuffered channel")
	default:
	}
}

func TestBusHistory(t *testing.T) {
	bus := NewBus()

	bus.Publish(Message{From: "a", Topic: "t", Content: "first"})
	bus.Publish(Message{From: "b", Topic: "t", Content: "second"})

	history := bus.History()
	if len(history) != 2 {
		t.Fatalf("History len = %d, want 2", len(history))
	}
	if history[0].Content != "first" {
		t.Errorf("history[0].Content = %q, want %q", history[0].Content, "first")
	}
	if history[1].Content != "second" {
		t.Errorf("history[1].Content = %q, want %q", history[1].Content, "second")
	}

	// Mutating returned history should not affect bus state.
	history[0].Content = "mutated"
	fresh := bus.History()
	if fresh[0].Content != "first" {
		t.Error("History not isolated from mutations")
	}
}

func TestBusConcurrentPublish(t *testing.T) {
	bus := NewBus()

	ch, unsub := bus.Subscribe("topic", 100)
	defer unsub()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Publish(Message{From: "agent", Topic: "topic", Content: "data"})
		}()
	}
	wg.Wait()

	history := bus.History()
	if len(history) != 20 {
		t.Errorf("History len = %d, want 20", len(history))
	}

	received := 0
	for {
		select {
		case <-ch:
			received++
		default:
			goto DONE
		}
	}
DONE:
	if received != 20 {
		t.Errorf("received = %d, want 20", received)
	}
}
