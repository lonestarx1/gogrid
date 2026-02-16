package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunList_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	yaml := `version: "1"
agents:
  researcher:
    model: claude-sonnet-4-5-20250929
    provider: anthropic
    instructions: Research things.
  summarizer:
    model: gpt-4o-mini
    provider: openai
    instructions: Summarize things.
`
	if err := os.WriteFile(filepath.Join(dir, "gogrid.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runList([]string{"-config", filepath.Join(dir, "gogrid.yaml")})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", code, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "NAME") {
		t.Error("expected table header")
	}
	if !strings.Contains(out, "researcher") {
		t.Error("expected researcher in output")
	}
	if !strings.Contains(out, "summarizer") {
		t.Error("expected summarizer in output")
	}
	if !strings.Contains(out, "anthropic") {
		t.Error("expected anthropic in output")
	}
	if !strings.Contains(out, "openai") {
		t.Error("expected openai in output")
	}
}

func TestRunList_MissingConfig(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runList([]string{"-config", "/nonexistent/gogrid.yaml"})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "Error") {
		t.Error("expected error in stderr")
	}
}
