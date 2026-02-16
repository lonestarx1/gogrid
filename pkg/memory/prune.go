package memory

import "time"

// PrunePolicy determines which memory entries should be pruned.
type PrunePolicy interface {
	// ShouldPrune returns true if the given entry should be removed.
	ShouldPrune(entry Entry) bool
}

// MaxEntries prunes entries beyond a maximum count per key.
// When applied, the oldest entries (by CreatedAt) are pruned first.
// The pruning loop is responsible for ordering; this policy simply
// marks entries beyond the limit.
type MaxEntries struct {
	// Limit is the maximum number of entries to keep per key.
	Limit int
	// seen tracks how many entries have been evaluated per key.
	seen map[string]int
}

// NewMaxEntries creates a MaxEntries policy with the given limit.
func NewMaxEntries(limit int) *MaxEntries {
	return &MaxEntries{
		Limit: limit,
		seen:  make(map[string]int),
	}
}

// ShouldPrune returns true if the key has exceeded its entry limit.
// Entries are expected to be presented oldest-first; entries beyond
// the limit are pruned.
func (p *MaxEntries) ShouldPrune(entry Entry) bool {
	p.seen[entry.Key]++
	return p.seen[entry.Key] > p.Limit
}

// Reset clears the internal counter so the policy can be reused.
func (p *MaxEntries) Reset() {
	p.seen = make(map[string]int)
}

// MaxAge prunes entries older than a specified duration.
type MaxAge struct {
	// Age is the maximum age of entries to keep.
	Age time.Duration
	// now is the reference time. If zero, time.Now() is used.
	now time.Time
}

// NewMaxAge creates a MaxAge policy with the given duration.
func NewMaxAge(age time.Duration) *MaxAge {
	return &MaxAge{Age: age}
}

// ShouldPrune returns true if the entry is older than the configured age.
func (p *MaxAge) ShouldPrune(entry Entry) bool {
	ref := p.now
	if ref.IsZero() {
		ref = time.Now()
	}
	return ref.Sub(entry.CreatedAt) > p.Age
}

// MaxSize prunes entries whose individual content size exceeds a byte limit.
type MaxSize struct {
	// ByteLimit is the maximum size in bytes for a single entry's content.
	ByteLimit int
}

// NewMaxSize creates a MaxSize policy with the given byte limit.
func NewMaxSize(byteLimit int) *MaxSize {
	return &MaxSize{ByteLimit: byteLimit}
}

// ShouldPrune returns true if the entry's size exceeds the byte limit.
func (p *MaxSize) ShouldPrune(entry Entry) bool {
	return entry.Size > p.ByteLimit
}

// AnyPolicy composes multiple policies with OR logic: an entry is pruned
// if any of the sub-policies says it should be pruned.
type AnyPolicy struct {
	// Policies is the set of policies to evaluate.
	Policies []PrunePolicy
}

// NewAnyPolicy creates a composite policy that prunes if any sub-policy matches.
func NewAnyPolicy(policies ...PrunePolicy) *AnyPolicy {
	return &AnyPolicy{Policies: policies}
}

// ShouldPrune returns true if any sub-policy returns true.
func (p *AnyPolicy) ShouldPrune(entry Entry) bool {
	for _, pol := range p.Policies {
		if pol.ShouldPrune(entry) {
			return true
		}
	}
	return false
}
