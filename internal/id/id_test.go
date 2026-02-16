package id

import (
	"testing"
	"time"
)

func TestNewReturnsNonEmpty(t *testing.T) {
	got := New()
	if got == "" {
		t.Fatal("New returned empty string")
	}
}

func TestNewUniqueness(t *testing.T) {
	const count = 1000
	seen := make(map[string]bool, count)
	for i := 0; i < count; i++ {
		id := New()
		if seen[id] {
			t.Fatalf("duplicate ID after %d generations: %s", i, id)
		}
		seen[id] = true
	}
}

func TestNewSortability(t *testing.T) {
	id1 := New()
	time.Sleep(2 * time.Millisecond)
	id2 := New()

	if id1 >= id2 {
		t.Errorf("IDs not sortable: %q should be < %q", id1, id2)
	}
}

func TestNewFormat(t *testing.T) {
	id := New()
	// 12 hex chars (timestamp) + 1-16 hex chars (random) = at least 13 chars.
	// In practice, the random part is always 16 chars since we format a uint64.
	if len(id) < 13 {
		t.Errorf("ID too short: %q (len=%d)", id, len(id))
	}

	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("ID contains non-hex character %q in %q", string(c), id)
			break
		}
	}
}
