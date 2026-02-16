package cost

import (
	"testing"

	"github.com/lonestarx1/gogrid/pkg/llm"
)

func TestNewTrackerHasDefaultPricing(t *testing.T) {
	tr := NewTracker()
	if len(tr.pricing) == 0 {
		t.Fatal("NewTracker has no default pricing")
	}
	if _, ok := tr.pricing["gpt-4o"]; !ok {
		t.Error("default pricing missing gpt-4o")
	}
}

func TestTrackerAdd(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		usage    llm.Usage
		wantCost float64
	}{
		{
			name:  "known model",
			model: "gpt-4o",
			usage: llm.Usage{PromptTokens: 1_000_000, CompletionTokens: 1_000_000, TotalTokens: 2_000_000},
			// gpt-4o: 2.50 prompt + 10.00 completion = 12.50
			wantCost: 12.50,
		},
		{
			name:     "unknown model returns zero cost",
			model:    "unknown-model",
			usage:    llm.Usage{PromptTokens: 1000, CompletionTokens: 500, TotalTokens: 1500},
			wantCost: 0,
		},
		{
			name:  "fractional tokens",
			model: "gpt-4o",
			usage: llm.Usage{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150},
			// 100/1M * 2.50 + 50/1M * 10.00 = 0.00025 + 0.0005 = 0.00075
			wantCost: 0.00075,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewTracker()
			got := tr.Add(tt.model, tt.usage)
			if got != tt.wantCost {
				t.Errorf("Add cost = %f, want %f", got, tt.wantCost)
			}
		})
	}
}

func TestTrackerTotalCost(t *testing.T) {
	tr := NewTracker()
	tr.Add("gpt-4o", llm.Usage{PromptTokens: 1_000_000, CompletionTokens: 0, TotalTokens: 1_000_000})
	tr.Add("gpt-4o", llm.Usage{PromptTokens: 0, CompletionTokens: 1_000_000, TotalTokens: 1_000_000})

	// 2.50 + 10.00 = 12.50
	want := 12.50
	if got := tr.TotalCost(); got != want {
		t.Errorf("TotalCost = %f, want %f", got, want)
	}
}

func TestTrackerTotalUsage(t *testing.T) {
	tr := NewTracker()
	tr.Add("gpt-4o", llm.Usage{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150})
	tr.Add("gpt-4o", llm.Usage{PromptTokens: 200, CompletionTokens: 100, TotalTokens: 300})

	usage := tr.TotalUsage()
	if usage.PromptTokens != 300 {
		t.Errorf("PromptTokens = %d, want 300", usage.PromptTokens)
	}
	if usage.CompletionTokens != 150 {
		t.Errorf("CompletionTokens = %d, want 150", usage.CompletionTokens)
	}
	if usage.TotalTokens != 450 {
		t.Errorf("TotalTokens = %d, want 450", usage.TotalTokens)
	}
}

func TestTrackerRecords(t *testing.T) {
	tr := NewTracker()
	tr.Add("gpt-4o", llm.Usage{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150})
	tr.Add("gpt-4o-mini", llm.Usage{PromptTokens: 200, CompletionTokens: 100, TotalTokens: 300})

	records := tr.Records()
	if len(records) != 2 {
		t.Fatalf("Records len = %d, want 2", len(records))
	}
	if records[0].Model != "gpt-4o" {
		t.Errorf("records[0].Model = %q, want %q", records[0].Model, "gpt-4o")
	}
	if records[1].Model != "gpt-4o-mini" {
		t.Errorf("records[1].Model = %q, want %q", records[1].Model, "gpt-4o-mini")
	}
}

func TestTrackerRecordsCopy(t *testing.T) {
	tr := NewTracker()
	tr.Add("gpt-4o", llm.Usage{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150})

	records := tr.Records()
	records[0].Model = "mutated"

	records2 := tr.Records()
	if records2[0].Model != "gpt-4o" {
		t.Errorf("Records did not return a copy: Model = %q, want %q", records2[0].Model, "gpt-4o")
	}
}

func TestTrackerSetPricing(t *testing.T) {
	tr := NewTracker()
	tr.SetPricing("custom-model", ModelPricing{PromptPer1M: 1.0, CompletionPer1M: 2.0})

	cost := tr.Add("custom-model", llm.Usage{PromptTokens: 1_000_000, CompletionTokens: 1_000_000, TotalTokens: 2_000_000})
	want := 3.0 // 1.0 + 2.0
	if cost != want {
		t.Errorf("custom model cost = %f, want %f", cost, want)
	}
}
