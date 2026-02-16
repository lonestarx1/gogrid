package runrecord

import (
	"testing"
	"time"

	"github.com/lonestarx1/gogrid/pkg/cost"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/trace"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	rec := &Record{
		RunID:    "019479a3c4e80001",
		Agent:    "researcher",
		Model:    "claude-sonnet-4-5-20250929",
		Provider: "anthropic",
		Input:    "test input",
		Output:   "test output",
		Turns:    2,
		Usage:    llm.Usage{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150},
		Cost:     0.001,
		Spans: []*trace.Span{
			{ID: "span1", Name: "agent.run", StartTime: time.Now(), EndTime: time.Now()},
		},
		CostRecords: []cost.Record{
			{Model: "claude-sonnet-4-5-20250929", Usage: llm.Usage{PromptTokens: 100}, Cost: 0.001},
		},
		StartTime: time.Now().Truncate(time.Millisecond),
		Duration:  2 * time.Second,
	}

	if err := Save(dir, rec); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(dir, rec.RunID)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.RunID != rec.RunID {
		t.Errorf("RunID = %q, want %q", loaded.RunID, rec.RunID)
	}
	if loaded.Agent != rec.Agent {
		t.Errorf("Agent = %q, want %q", loaded.Agent, rec.Agent)
	}
	if loaded.Model != rec.Model {
		t.Errorf("Model = %q, want %q", loaded.Model, rec.Model)
	}
	if loaded.Provider != rec.Provider {
		t.Errorf("Provider = %q, want %q", loaded.Provider, rec.Provider)
	}
	if loaded.Input != rec.Input {
		t.Errorf("Input = %q, want %q", loaded.Input, rec.Input)
	}
	if loaded.Output != rec.Output {
		t.Errorf("Output = %q, want %q", loaded.Output, rec.Output)
	}
	if loaded.Turns != rec.Turns {
		t.Errorf("Turns = %d, want %d", loaded.Turns, rec.Turns)
	}
	if loaded.Usage.PromptTokens != rec.Usage.PromptTokens {
		t.Errorf("PromptTokens = %d, want %d", loaded.Usage.PromptTokens, rec.Usage.PromptTokens)
	}
	if loaded.Cost != rec.Cost {
		t.Errorf("Cost = %f, want %f", loaded.Cost, rec.Cost)
	}
	if len(loaded.Spans) != 1 {
		t.Errorf("Spans len = %d, want 1", len(loaded.Spans))
	}
	if len(loaded.CostRecords) != 1 {
		t.Errorf("CostRecords len = %d, want 1", len(loaded.CostRecords))
	}
}

func TestSave_MissingID(t *testing.T) {
	dir := t.TempDir()
	rec := &Record{Agent: "test"}

	err := Save(dir, rec)
	if err == nil {
		t.Fatal("expected error for missing run ID")
	}
}

func TestLoad_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(dir, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing record")
	}
}

func TestSave_AutoCreateDir(t *testing.T) {
	dir := t.TempDir()
	rec := &Record{RunID: "test-run-001", Agent: "test"}

	if err := Save(dir, rec); err != nil {
		t.Fatalf("Save should auto-create .gogrid/runs: %v", err)
	}

	// Verify the file exists by loading it.
	loaded, err := Load(dir, "test-run-001")
	if err != nil {
		t.Fatalf("Load after auto-create: %v", err)
	}
	if loaded.RunID != "test-run-001" {
		t.Errorf("RunID = %q, want %q", loaded.RunID, "test-run-001")
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()

	// Empty list when no runs exist.
	ids, err := List(dir)
	if err != nil {
		t.Fatalf("List empty: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected empty list, got %d items", len(ids))
	}

	// Save a few records in non-sorted order.
	records := []*Record{
		{RunID: "aaa", Agent: "a"},
		{RunID: "ccc", Agent: "c"},
		{RunID: "bbb", Agent: "b"},
	}
	for _, rec := range records {
		if err := Save(dir, rec); err != nil {
			t.Fatalf("Save %s: %v", rec.RunID, err)
		}
	}

	ids, err = List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(ids) != 3 {
		t.Fatalf("expected 3 ids, got %d", len(ids))
	}
	// Should be sorted descending.
	if ids[0] != "ccc" || ids[1] != "bbb" || ids[2] != "aaa" {
		t.Errorf("expected [ccc bbb aaa], got %v", ids)
	}
}

func TestSaveAndLoad_WithError(t *testing.T) {
	dir := t.TempDir()
	rec := &Record{
		RunID: "error-run",
		Agent: "test",
		Error: "something went wrong",
	}

	if err := Save(dir, rec); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(dir, "error-run")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Error != "something went wrong" {
		t.Errorf("Error = %q, want %q", loaded.Error, "something went wrong")
	}
}
