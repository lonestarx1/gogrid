package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lonestarx1/gogrid/pkg/llm"
)

// mockProvider returns a canned response for testing.
type mockProvider struct {
	response string
}

func (m *mockProvider) Complete(_ context.Context, params llm.Params) (*llm.Response, error) {
	return &llm.Response{
		Message: llm.Message{
			Role:    llm.RoleAssistant,
			Content: m.response,
		},
		Usage: llm.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
		Model: params.Model,
	}, nil
}

func newMockFactory(resp string) ProviderFactory {
	return func(_ context.Context, _ string) (llm.Provider, error) {
		return &mockProvider{response: resp}, nil
	}
}

func newFailingFactory(msg string) ProviderFactory {
	return func(_ context.Context, _ string) (llm.Provider, error) {
		return nil, fmt.Errorf("%s", msg)
	}
}

func writeTestConfig(t *testing.T, dir string) string {
	t.Helper()
	yaml := `version: "1"
agents:
  helper:
    model: test-model
    provider: openai
    instructions: You are helpful.
    config:
      max_turns: 1
`
	path := filepath.Join(dir, "gogrid.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunRun_Success(t *testing.T) {
	dir := t.TempDir()
	configPath := writeTestConfig(t, dir)

	// Change to temp dir so run record is saved there.
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)
	app.SetProviderFactory(newMockFactory("Hello from mock!"))

	code := app.runRun([]string{"-config", configPath, "-input", "test", "helper"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", code, stderr.String())
	}

	if !strings.Contains(stdout.String(), "Hello from mock!") {
		t.Errorf("expected mock response in stdout, got: %s", stdout.String())
	}

	// Verify run record was saved.
	if !strings.Contains(stderr.String(), "Run ID:") {
		t.Errorf("expected run ID in stderr, got: %s", stderr.String())
	}
}

func TestRunRun_UnknownAgent(t *testing.T) {
	dir := t.TempDir()
	configPath := writeTestConfig(t, dir)

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)
	app.SetProviderFactory(newMockFactory(""))

	code := app.runRun([]string{"-config", configPath, "-input", "test", "nonexistent"})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "unknown agent") {
		t.Errorf("expected unknown agent error, got: %s", stderr.String())
	}
}

func TestRunRun_MissingAPIKey(t *testing.T) {
	dir := t.TempDir()
	configPath := writeTestConfig(t, dir)

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)
	app.SetProviderFactory(newFailingFactory("OPENAI_API_KEY is not set"))

	code := app.runRun([]string{"-config", configPath, "-input", "test", "helper"})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "API_KEY") {
		t.Errorf("expected API key error, got: %s", stderr.String())
	}
}

func TestRunRun_NoAgentName(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runRun([]string{"-input", "test"})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestRunRun_NoInput(t *testing.T) {
	dir := t.TempDir()
	configPath := writeTestConfig(t, dir)

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)
	app.SetProviderFactory(newMockFactory(""))

	code := app.runRun([]string{"-config", configPath, "helper"})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "no input") {
		t.Errorf("expected no input error, got: %s", stderr.String())
	}
}

func TestRunRun_RecordSaved(t *testing.T) {
	dir := t.TempDir()
	configPath := writeTestConfig(t, dir)

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)
	app.SetProviderFactory(newMockFactory("response"))

	code := app.runRun([]string{"-config", configPath, "-input", "hello", "helper"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", code, stderr.String())
	}

	// Check that a run record file was created.
	entries, err := os.ReadDir(filepath.Join(dir, ".gogrid", "runs"))
	if err != nil {
		t.Fatalf("failed to read runs dir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 run record, got %d", len(entries))
	}
}
