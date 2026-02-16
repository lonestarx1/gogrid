package cost

import (
	"sync"
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

func TestAddForEntity(t *testing.T) {
	tr := NewTracker()
	c := tr.AddForEntity("gpt-4o", "agent-researcher", llm.Usage{
		PromptTokens:     1_000_000,
		CompletionTokens: 1_000_000,
		TotalTokens:      2_000_000,
	})

	// gpt-4o: 2.50 + 10.00 = 12.50
	if c != 12.50 {
		t.Errorf("AddForEntity cost = %f, want 12.50", c)
	}

	if got := tr.EntityCost("agent-researcher"); got != 12.50 {
		t.Errorf("EntityCost = %f, want 12.50", got)
	}
}

func TestEntityCostMultipleEntities(t *testing.T) {
	tr := NewTracker()
	tr.AddForEntity("gpt-4o", "agent-a", llm.Usage{PromptTokens: 1_000_000, TotalTokens: 1_000_000})
	tr.AddForEntity("gpt-4o", "agent-b", llm.Usage{PromptTokens: 1_000_000, TotalTokens: 1_000_000})
	tr.AddForEntity("gpt-4o", "agent-a", llm.Usage{CompletionTokens: 1_000_000, TotalTokens: 1_000_000})

	costA := tr.EntityCost("agent-a")
	// 2.50 (first call prompt) + 10.00 (third call completion) = 12.50
	if costA != 12.50 {
		t.Errorf("agent-a cost = %f, want 12.50", costA)
	}

	costB := tr.EntityCost("agent-b")
	// 2.50 (prompt only)
	if costB != 2.50 {
		t.Errorf("agent-b cost = %f, want 2.50", costB)
	}
}

func TestEntityCostUnknown(t *testing.T) {
	tr := NewTracker()
	if got := tr.EntityCost("nonexistent"); got != 0 {
		t.Errorf("EntityCost for unknown = %f, want 0", got)
	}
}

func TestAddForEntityEmptyEntityDoesNotTrack(t *testing.T) {
	tr := NewTracker()
	tr.AddForEntity("gpt-4o", "", llm.Usage{PromptTokens: 1000, TotalTokens: 1000})

	if got := tr.EntityCost(""); got != 0 {
		t.Errorf("empty entity cost = %f, want 0", got)
	}
}

func TestBudgetAlerts(t *testing.T) {
	tr := NewTracker()
	tr.SetPricing("test", ModelPricing{PromptPer1M: 10.0, CompletionPer1M: 0})
	tr.SetBudget(10.0) // $10 budget

	var mu sync.Mutex
	var alerts []float64
	tr.OnBudgetThreshold(func(threshold, current float64) {
		mu.Lock()
		alerts = append(alerts, threshold)
		mu.Unlock()
	}, 0.5, 0.8, 1.0)

	// Add $5 (50% of budget) — should trigger 0.5 alert.
	tr.Add("test", llm.Usage{PromptTokens: 500_000, TotalTokens: 500_000})

	mu.Lock()
	if len(alerts) != 1 || alerts[0] != 0.5 {
		t.Errorf("after $5: alerts = %v, want [0.5]", alerts)
	}
	mu.Unlock()

	// Add $4 (total $9, 90%) — should trigger 0.8 alert.
	tr.Add("test", llm.Usage{PromptTokens: 400_000, TotalTokens: 400_000})

	mu.Lock()
	if len(alerts) != 2 || alerts[1] != 0.8 {
		t.Errorf("after $9: alerts = %v, want [0.5, 0.8]", alerts)
	}
	mu.Unlock()

	// Add $2 (total $11, 110%) — should trigger 1.0 alert.
	tr.Add("test", llm.Usage{PromptTokens: 200_000, TotalTokens: 200_000})

	mu.Lock()
	if len(alerts) != 3 || alerts[2] != 1.0 {
		t.Errorf("after $11: alerts = %v, want [0.5, 0.8, 1.0]", alerts)
	}
	mu.Unlock()
}

func TestBudgetAlertsNotFiredTwice(t *testing.T) {
	tr := NewTracker()
	tr.SetPricing("test", ModelPricing{PromptPer1M: 10.0, CompletionPer1M: 0})
	tr.SetBudget(10.0)

	callCount := 0
	tr.OnBudgetThreshold(func(threshold, current float64) {
		callCount++
	}, 0.5)

	// Exceed 50% twice.
	tr.Add("test", llm.Usage{PromptTokens: 600_000, TotalTokens: 600_000})
	tr.Add("test", llm.Usage{PromptTokens: 100_000, TotalTokens: 100_000})

	if callCount != 1 {
		t.Errorf("alert fired %d times, want 1", callCount)
	}
}

func TestBudgetNoAlertWithoutBudget(t *testing.T) {
	tr := NewTracker()
	called := false
	tr.OnBudgetThreshold(func(threshold, current float64) {
		called = true
	}, 0.5)

	tr.Add("gpt-4o", llm.Usage{PromptTokens: 1_000_000, TotalTokens: 1_000_000})
	if called {
		t.Error("alert should not fire without budget set")
	}
}

func TestRecordHasTimestamp(t *testing.T) {
	tr := NewTracker()
	tr.Add("gpt-4o", llm.Usage{PromptTokens: 100, TotalTokens: 100})

	records := tr.Records()
	if records[0].Time.IsZero() {
		t.Error("Record.Time should not be zero")
	}
}

func TestReport(t *testing.T) {
	tr := NewTracker()
	tr.AddForEntity("gpt-4o", "agent-a", llm.Usage{PromptTokens: 1_000_000, CompletionTokens: 500_000, TotalTokens: 1_500_000})
	tr.AddForEntity("gpt-4o-mini", "agent-b", llm.Usage{PromptTokens: 2_000_000, CompletionTokens: 1_000_000, TotalTokens: 3_000_000})
	tr.AddForEntity("gpt-4o", "agent-a", llm.Usage{PromptTokens: 500_000, CompletionTokens: 250_000, TotalTokens: 750_000})

	rpt := tr.Report()

	if rpt.RecordCount != 3 {
		t.Errorf("RecordCount = %d, want 3", rpt.RecordCount)
	}

	// Check ByModel.
	gpt4o, ok := rpt.ByModel["gpt-4o"]
	if !ok {
		t.Fatal("missing gpt-4o in ByModel")
	}
	if gpt4o.Calls != 2 {
		t.Errorf("gpt-4o calls = %d, want 2", gpt4o.Calls)
	}
	if gpt4o.Usage.PromptTokens != 1_500_000 {
		t.Errorf("gpt-4o prompt tokens = %d, want 1500000", gpt4o.Usage.PromptTokens)
	}

	mini, ok := rpt.ByModel["gpt-4o-mini"]
	if !ok {
		t.Fatal("missing gpt-4o-mini in ByModel")
	}
	if mini.Calls != 1 {
		t.Errorf("gpt-4o-mini calls = %d, want 1", mini.Calls)
	}

	// Check ByEntity.
	if len(rpt.ByEntity) != 2 {
		t.Errorf("ByEntity len = %d, want 2", len(rpt.ByEntity))
	}
	if _, ok := rpt.ByEntity["agent-a"]; !ok {
		t.Error("missing agent-a in ByEntity")
	}
	if _, ok := rpt.ByEntity["agent-b"]; !ok {
		t.Error("missing agent-b in ByEntity")
	}

	// TotalCost should match.
	if rpt.TotalCost != tr.TotalCost() {
		t.Errorf("Report.TotalCost = %f, tracker.TotalCost = %f", rpt.TotalCost, tr.TotalCost())
	}
}

func TestReportEmpty(t *testing.T) {
	tr := NewTracker()
	rpt := tr.Report()

	if rpt.TotalCost != 0 {
		t.Errorf("TotalCost = %f, want 0", rpt.TotalCost)
	}
	if rpt.RecordCount != 0 {
		t.Errorf("RecordCount = %d, want 0", rpt.RecordCount)
	}
	if len(rpt.ByModel) != 0 {
		t.Errorf("ByModel len = %d, want 0", len(rpt.ByModel))
	}
}

func TestTrackerConcurrency(t *testing.T) {
	tr := NewTracker()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tr.AddForEntity("gpt-4o", "agent", llm.Usage{PromptTokens: 1000, TotalTokens: 1000})
		}()
	}
	wg.Wait()

	records := tr.Records()
	if len(records) != 100 {
		t.Errorf("Records len = %d, want 100", len(records))
	}

	usage := tr.TotalUsage()
	if usage.PromptTokens != 100_000 {
		t.Errorf("PromptTokens = %d, want 100000", usage.PromptTokens)
	}
}
