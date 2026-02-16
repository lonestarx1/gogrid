package team

import (
	"sync"
	"time"
)

// Message is a single message on the bus.
type Message struct {
	// From identifies the sender (typically an agent name).
	From string
	// Topic is the message topic/channel.
	Topic string
	// Content is the message payload.
	Content string
	// Timestamp is when the message was published.
	Timestamp time.Time
	// Metadata holds arbitrary key-value pairs.
	Metadata map[string]string
}

// Bus is a pub/sub message bus for inter-agent communication.
// Subscribers receive messages on buffered channels with non-blocking sends.
// Safe for concurrent use.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Message
	history     []Message
}

// NewBus creates a new message bus.
func NewBus() *Bus {
	return &Bus{
		subscribers: make(map[string][]chan Message),
	}
}

// Publish sends a message to all subscribers of the given topic.
// The timestamp is set automatically. Sends are non-blocking: if a
// subscriber's channel is full, the message is dropped for that subscriber.
func (b *Bus) Publish(msg Message) {
	msg.Timestamp = time.Now()

	b.mu.Lock()
	b.history = append(b.history, msg)
	subs := make([]chan Message, len(b.subscribers[msg.Topic]))
	copy(subs, b.subscribers[msg.Topic])
	b.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- msg:
		default:
		}
	}
}

// Subscribe registers a channel to receive messages on the given topic.
// The bufferSize controls how many messages can queue before drops occur.
// Returns a receive-only channel and an unsubscribe function.
func (b *Bus) Subscribe(topic string, bufferSize int) (<-chan Message, func()) {
	ch := make(chan Message, bufferSize)

	b.mu.Lock()
	b.subscribers[topic] = append(b.subscribers[topic], ch)
	b.mu.Unlock()

	unsub := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		subs := b.subscribers[topic]
		for i, s := range subs {
			if s == ch {
				b.subscribers[topic] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}

	return ch, unsub
}

// History returns a copy of all published messages in order.
func (b *Bus) History() []Message {
	b.mu.RLock()
	defer b.mu.RUnlock()
	cp := make([]Message, len(b.history))
	copy(cp, b.history)
	return cp
}
