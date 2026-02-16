package shared

import "time"

// ChangeType describes the kind of mutation that occurred.
type ChangeType int

const (
	// ChangeSave indicates messages were saved.
	ChangeSave ChangeType = iota + 1
	// ChangeClear indicates messages were cleared.
	ChangeClear
)

// ChangeEvent describes a mutation to the shared memory store.
type ChangeEvent struct {
	// Type is the kind of change.
	Type ChangeType
	// Key is the memory key that was modified.
	Key string
	// Timestamp is when the change occurred.
	Timestamp time.Time
}

// subscriber holds a registered change notification channel.
type subscriber struct {
	ch chan ChangeEvent
}
