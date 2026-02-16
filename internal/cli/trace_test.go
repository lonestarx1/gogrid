package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lonestarx1/gogrid/internal/runrecord"
	"github.com/lonestarx1/gogrid/pkg/trace"
)

func saveTestRecord(t *testing.T, dir string) *runrecord.Record {
	t.Helper()
	now := time.Now()
	rec := &runrecord.Record{
		RunID:    "test-run-001",
		Agent:    "researcher",
		Model:    "claude-sonnet-4-5-20250929",
		Provider: "anthropic",
		Input:    "hello",
		Output:   "world",
		Turns:    2,
		Duration: 4200 * time.Millisecond,
		Spans: []*trace.Span{
			{
				ID:        "span-root",
				Name:      "agent.run",
				StartTime: now,
				EndTime:   now.Add(4200 * time.Millisecond),
			},
			{
				ID:        "span-mem",
				ParentID:  "span-root",
				Name:      "memory.load",
				StartTime: now,
				EndTime:   now.Add(1 * time.Millisecond),
			},
			{
				ID:        "span-llm",
				ParentID:  "span-root",
				Name:      "llm.complete",
				StartTime: now.Add(1 * time.Millisecond),
				EndTime:   now.Add(2100 * time.Millisecond),
				Attributes: map[string]string{
					"llm.prompt_tokens":     "150",
					"llm.completion_tokens": "89",
				},
			},
			{
				ID:        "span-tool",
				ParentID:  "span-root",
				Name:      "tool.execute",
				StartTime: now.Add(2100 * time.Millisecond),
				EndTime:   now.Add(3900 * time.Millisecond),
				Attributes: map[string]string{
					"tool.name": "web_search",
				},
			},
		},
	}
	if err := runrecord.Save(dir, rec); err != nil {
		t.Fatal(err)
	}
	return rec
}

func TestRunTrace_SpanTree(t *testing.T) {
	dir := t.TempDir()
	saveTestRecord(t, dir)

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runTrace([]string{"test-run-001"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", code, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "test-run-001") {
		t.Error("expected run ID in output")
	}
	if !strings.Contains(out, "agent.run") {
		t.Error("expected agent.run span")
	}
	if !strings.Contains(out, "memory.load") {
		t.Error("expected memory.load span")
	}
	if !strings.Contains(out, "llm.complete") {
		t.Error("expected llm.complete span")
	}
	if !strings.Contains(out, "tool.execute") {
		t.Error("expected tool.execute span")
	}
	if !strings.Contains(out, "prompt: 150") {
		t.Error("expected prompt token count")
	}
	if !strings.Contains(out, "web_search") {
		t.Error("expected tool name")
	}
}

func TestRunTrace_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	saveTestRecord(t, dir)

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runTrace([]string{"-json", "test-run-001"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", code, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, `"name"`) {
		t.Error("expected JSON output with name field")
	}
	if !strings.Contains(out, "agent.run") {
		t.Error("expected agent.run in JSON")
	}
}

func TestRunTrace_MissingRunID(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runTrace([]string{"nonexistent"})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestRunTrace_NoArgs_ListRecent(t *testing.T) {
	dir := t.TempDir()
	saveTestRecord(t, dir)

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runTrace(nil)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "test-run-001") {
		t.Error("expected run ID in recent runs list")
	}
}
