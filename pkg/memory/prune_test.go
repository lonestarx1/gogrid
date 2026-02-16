package memory

import (
	"testing"
	"time"

	"github.com/lonestarx1/gogrid/pkg/llm"
)

func TestMaxEntriesShouldPrune(t *testing.T) {
	tests := []struct {
		name    string
		limit   int
		entries []Entry
		want    []bool
	}{
		{
			name:  "under limit",
			limit: 3,
			entries: []Entry{
				{Key: "a", Message: llm.NewUserMessage("1")},
				{Key: "a", Message: llm.NewUserMessage("2")},
			},
			want: []bool{false, false},
		},
		{
			name:  "at limit",
			limit: 2,
			entries: []Entry{
				{Key: "a", Message: llm.NewUserMessage("1")},
				{Key: "a", Message: llm.NewUserMessage("2")},
			},
			want: []bool{false, false},
		},
		{
			name:  "over limit",
			limit: 2,
			entries: []Entry{
				{Key: "a", Message: llm.NewUserMessage("1")},
				{Key: "a", Message: llm.NewUserMessage("2")},
				{Key: "a", Message: llm.NewUserMessage("3")},
			},
			want: []bool{false, false, true},
		},
		{
			name:  "separate keys",
			limit: 1,
			entries: []Entry{
				{Key: "a", Message: llm.NewUserMessage("1")},
				{Key: "b", Message: llm.NewUserMessage("2")},
				{Key: "a", Message: llm.NewUserMessage("3")},
			},
			want: []bool{false, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewMaxEntries(tt.limit)
			for i, e := range tt.entries {
				got := p.ShouldPrune(e)
				if got != tt.want[i] {
					t.Errorf("entry %d: ShouldPrune = %v, want %v", i, got, tt.want[i])
				}
			}
		})
	}
}

func TestMaxEntriesReset(t *testing.T) {
	p := NewMaxEntries(1)
	e := Entry{Key: "a", Message: llm.NewUserMessage("1")}
	p.ShouldPrune(e)
	p.ShouldPrune(e) // second call should return true

	p.Reset()
	if p.ShouldPrune(e) {
		t.Error("after Reset, first call should return false")
	}
}

func TestMaxAgeShouldPrune(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		age       time.Duration
		createdAt time.Time
		want      bool
	}{
		{
			name:      "recent entry kept",
			age:       1 * time.Hour,
			createdAt: now.Add(-30 * time.Minute),
			want:      false,
		},
		{
			name:      "old entry pruned",
			age:       1 * time.Hour,
			createdAt: now.Add(-2 * time.Hour),
			want:      true,
		},
		{
			name:      "exactly at boundary",
			age:       1 * time.Hour,
			createdAt: now.Add(-1 * time.Hour),
			want:      false,
		},
		{
			name:      "just past boundary",
			age:       1 * time.Hour,
			createdAt: now.Add(-1*time.Hour - 1*time.Nanosecond),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &MaxAge{Age: tt.age, now: now}
			e := Entry{Key: "k", CreatedAt: tt.createdAt}
			got := p.ShouldPrune(e)
			if got != tt.want {
				t.Errorf("ShouldPrune = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaxSizeShouldPrune(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		size  int
		want  bool
	}{
		{name: "under limit", limit: 100, size: 50, want: false},
		{name: "at limit", limit: 100, size: 100, want: false},
		{name: "over limit", limit: 100, size: 101, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewMaxSize(tt.limit)
			e := Entry{Key: "k", Size: tt.size}
			got := p.ShouldPrune(e)
			if got != tt.want {
				t.Errorf("ShouldPrune = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnyPolicyShouldPrune(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		entry Entry
		want  bool
	}{
		{
			name:  "no policy matches",
			entry: Entry{Key: "k", Size: 50, CreatedAt: now.Add(-10 * time.Minute)},
			want:  false,
		},
		{
			name:  "size matches",
			entry: Entry{Key: "k", Size: 200, CreatedAt: now.Add(-10 * time.Minute)},
			want:  true,
		},
		{
			name:  "age matches",
			entry: Entry{Key: "k", Size: 50, CreatedAt: now.Add(-2 * time.Hour)},
			want:  true,
		},
		{
			name:  "both match",
			entry: Entry{Key: "k", Size: 200, CreatedAt: now.Add(-2 * time.Hour)},
			want:  true,
		},
	}

	p := NewAnyPolicy(
		NewMaxSize(100),
		&MaxAge{Age: 1 * time.Hour, now: now},
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.ShouldPrune(tt.entry)
			if got != tt.want {
				t.Errorf("ShouldPrune = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnyPolicyEmpty(t *testing.T) {
	p := NewAnyPolicy()
	e := Entry{Key: "k", Size: 100}
	if p.ShouldPrune(e) {
		t.Error("empty AnyPolicy should not prune anything")
	}
}
