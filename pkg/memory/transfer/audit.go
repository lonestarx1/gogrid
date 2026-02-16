package transfer

import "time"

// AuditEntry records a state ownership transfer event.
type AuditEntry struct {
	// From is the owner who released the state.
	From string
	// To is the owner who acquired the state.
	To string
	// Generation is the generation number after transfer.
	Generation uint64
	// Timestamp is when the transfer occurred.
	Timestamp time.Time
}

// ValidationHook is called before a transfer is executed.
// If it returns a non-nil error, the transfer is rejected.
type ValidationHook func(from, to string) error
