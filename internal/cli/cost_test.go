package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lonestarx1/gogrid/internal/runrecord"
	"github.com/lonestarx1/gogrid/pkg/cost"
	"github.com/lonestarx1/gogrid/pkg/llm"
)

func saveCostTestRecord(t *testing.T, dir string) {
	t.Helper()
	rec := &runrecord.Record{
		RunID:    "cost-run-001",
		Agent:    "researcher",
		Model:    "claude-sonnet-4-5-20250929",
		Provider: "anthropic",
		Turns:    2,
		Usage:    llm.Usage{PromptTokens: 430, CompletionTokens: 134, TotalTokens: 564},
		Cost:     0.003280,
		CostRecords: []cost.Record{
			{
				Model: "claude-sonnet-4-5-20250929",
				Usage: llm.Usage{PromptTokens: 200, CompletionTokens: 89, TotalTokens: 289},
				Cost:  0.001935,
				Time:  time.Now(),
			},
			{
				Model: "claude-sonnet-4-5-20250929",
				Usage: llm.Usage{PromptTokens: 230, CompletionTokens: 45, TotalTokens: 275},
				Cost:  0.001345,
				Time:  time.Now(),
			},
		},
		StartTime: time.Now(),
		Duration:  3 * time.Second,
	}
	if err := runrecord.Save(dir, rec); err != nil {
		t.Fatal(err)
	}
}

func TestRunCost_Table(t *testing.T) {
	dir := t.TempDir()
	saveCostTestRecord(t, dir)

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runCost([]string{"cost-run-001"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", code, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "cost-run-001") {
		t.Error("expected run ID")
	}
	if !strings.Contains(out, "claude-sonnet-4-5-20250929") {
		t.Error("expected model name")
	}
	if !strings.Contains(out, "TOTAL") {
		t.Error("expected TOTAL row")
	}
	if !strings.Contains(out, "MODEL") {
		t.Error("expected table header")
	}
}

func TestRunCost_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	saveCostTestRecord(t, dir)

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runCost([]string{"-json", "cost-run-001"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", code, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, `"model"`) {
		t.Error("expected JSON with model field")
	}
	if !strings.Contains(out, `"cost"`) {
		t.Error("expected JSON with cost field")
	}
}

func TestRunCost_MissingRunID(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runCost([]string{"nonexistent"})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestRunCost_NoArgs_ListAll(t *testing.T) {
	dir := t.TempDir()
	saveCostTestRecord(t, dir)

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runCost(nil)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", code, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "cost-run-001") {
		t.Error("expected run ID in list")
	}
}
