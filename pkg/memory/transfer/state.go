package transfer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/memory"
)

// Sentinel errors for ownership enforcement.
var (
	// ErrStateTransferred is returned when a Handle is used after its
	// generation has been superseded by a transfer.
	ErrStateTransferred = errors.New("state has been transferred to a new owner")
	// ErrNotOwner is returned when a non-owner attempts a transfer.
	ErrNotOwner = errors.New("caller is not the current owner")
	// ErrAlreadyAcquired is returned when Acquire is called but the
	// state already has an owner.
	ErrAlreadyAcquired = errors.New("state is already acquired by another owner")
)

// State manages ownership transfer of a memory.Memory instance.
// Only one owner can hold the state at a time, enforced by a generation
// counter. When state is transferred, previous Handles are invalidated.
type State struct {
	mu         sync.RWMutex
	inner      memory.Memory
	owner      string
	generation uint64
	audit      []AuditEntry
	hooks      []ValidationHook
}

// NewState creates a TransferableState wrapping the given memory.
// The state starts with no owner; call Acquire to take initial ownership.
func NewState(inner memory.Memory) *State {
	return &State{inner: inner}
}

// OnTransfer registers a validation hook that is called before each transfer.
// If any hook returns an error, the transfer is rejected.
func (s *State) OnTransfer(hook ValidationHook) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hooks = append(s.hooks, hook)
}

// Acquire takes initial ownership of the state. Returns ErrAlreadyAcquired
// if the state already has an owner.
func (s *State) Acquire(owner string) (*Handle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.owner != "" {
		return nil, ErrAlreadyAcquired
	}

	s.generation++
	s.owner = owner
	s.audit = append(s.audit, AuditEntry{
		From:       "",
		To:         owner,
		Generation: s.generation,
		Timestamp:  time.Now(),
	})

	return &Handle{
		state:      s,
		owner:      owner,
		generation: s.generation,
	}, nil
}

// Transfer moves ownership from the current owner to a new owner.
// The caller must provide the current owner name for verification.
// Returns a new Handle for the new owner. The old Handle becomes invalid.
func (s *State) Transfer(from, to string) (*Handle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.owner != from {
		return nil, fmt.Errorf("%w: current owner is %q, not %q", ErrNotOwner, s.owner, from)
	}

	// Run validation hooks.
	for _, hook := range s.hooks {
		if err := hook(from, to); err != nil {
			return nil, fmt.Errorf("transfer validation: %w", err)
		}
	}

	s.generation++
	s.owner = to
	s.audit = append(s.audit, AuditEntry{
		From:       from,
		To:         to,
		Generation: s.generation,
		Timestamp:  time.Now(),
	})

	return &Handle{
		state:      s,
		owner:      to,
		generation: s.generation,
	}, nil
}

// Owner returns the current owner name. Empty string if unowned.
func (s *State) Owner() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.owner
}

// Generation returns the current generation number.
func (s *State) Generation() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.generation
}

// AuditLog returns a copy of the transfer audit trail.
func (s *State) AuditLog() []AuditEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make([]AuditEntry, len(s.audit))
	copy(cp, s.audit)
	return cp
}

// Handle provides memory.Memory access that is valid only while the
// owning generation is current. All operations return ErrStateTransferred
// if a transfer has occurred since this Handle was created.
type Handle struct {
	state      *State
	owner      string
	generation uint64
}

// validate checks that this Handle's generation is still current.
// Caller must hold at least a read lock on s.state.mu.
func (h *Handle) validate() error {
	if h.generation != h.state.generation {
		return ErrStateTransferred
	}
	return nil
}

// Load retrieves stored messages if the Handle is still valid.
func (h *Handle) Load(ctx context.Context, key string) ([]llm.Message, error) {
	h.state.mu.RLock()
	if err := h.validate(); err != nil {
		h.state.mu.RUnlock()
		return nil, err
	}
	h.state.mu.RUnlock()
	return h.state.inner.Load(ctx, key)
}

// Save stores messages if the Handle is still valid.
func (h *Handle) Save(ctx context.Context, key string, messages []llm.Message) error {
	h.state.mu.RLock()
	if err := h.validate(); err != nil {
		h.state.mu.RUnlock()
		return err
	}
	h.state.mu.RUnlock()
	return h.state.inner.Save(ctx, key, messages)
}

// Clear removes messages if the Handle is still valid.
func (h *Handle) Clear(ctx context.Context, key string) error {
	h.state.mu.RLock()
	if err := h.validate(); err != nil {
		h.state.mu.RUnlock()
		return err
	}
	h.state.mu.RUnlock()
	return h.state.inner.Clear(ctx, key)
}

// Owner returns the owner name of this Handle.
func (h *Handle) Owner() string {
	return h.owner
}
