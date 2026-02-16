package id

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

// New returns a unique, time-sortable identifier.
//
// Format: 12-char hex timestamp (milliseconds since epoch) concatenated with
// 16-char hex random suffix, producing a 28-character string.
//
// IDs generated later in time sort lexicographically after earlier IDs.
// The random suffix ensures uniqueness across concurrent goroutines.
func New() string {
	ts := time.Now().UnixMilli()
	rb := make([]byte, 8)
	// crypto/rand.Read failing indicates a broken system (no entropy source).
	// Every major Go library (google/uuid, etc.) treats this as unrecoverable.
	if _, err := rand.Read(rb); err != nil {
		panic("id: crypto/rand read failed: " + err.Error())
	}
	return fmt.Sprintf("%012x%x", ts, binary.BigEndian.Uint64(rb))
}
