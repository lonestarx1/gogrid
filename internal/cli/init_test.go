package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInit_SingleTemplate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "myproject")

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runInit([]string{"-template", "single", "-name", "myproject", dir})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", code, stderr.String())
	}

	// Check files exist.
	for _, f := range []string{"gogrid.yaml", "main.go", "Makefile", "README.md"} {
		path := filepath.Join(dir, f)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s to exist: %v", f, err)
		}
	}

	// Verify config is valid YAML with version "1".
	data, _ := os.ReadFile(filepath.Join(dir, "gogrid.yaml"))
	if !strings.Contains(string(data), `version: "1"`) {
		t.Error("expected version 1 in generated config")
	}

	if !strings.Contains(stdout.String(), "Created GoGrid project") {
		t.Error("expected success message")
	}
}

func TestRunInit_TeamTemplate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "teamproject")

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runInit([]string{"-template", "team", dir})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", code, stderr.String())
	}

	data, _ := os.ReadFile(filepath.Join(dir, "gogrid.yaml"))
	if !strings.Contains(string(data), "researcher") {
		t.Error("expected researcher agent in team config")
	}
	if !strings.Contains(string(data), "reviewer") {
		t.Error("expected reviewer agent in team config")
	}
}

func TestRunInit_PipelineTemplate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "pipeproject")

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runInit([]string{"-template", "pipeline", dir})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", code, stderr.String())
	}

	data, _ := os.ReadFile(filepath.Join(dir, "gogrid.yaml"))
	if !strings.Contains(string(data), "drafter") {
		t.Error("expected drafter agent in pipeline config")
	}
	if !strings.Contains(string(data), "editor") {
		t.Error("expected editor agent in pipeline config")
	}
}

func TestRunInit_NonEmptyDir(t *testing.T) {
	dir := t.TempDir()
	// Create a visible file.
	if err := os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runInit([]string{dir})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "not empty") {
		t.Errorf("expected non-empty error, got: %s", stderr.String())
	}
}

func TestRunInit_InvalidTemplate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "proj")

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runInit([]string{"-template", "invalid", dir})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "unknown template") {
		t.Errorf("expected unknown template error, got: %s", stderr.String())
	}
}

func TestRunInit_DefaultsToDirectoryName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my-cool-project")

	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runInit([]string{dir})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", code, stderr.String())
	}

	if !strings.Contains(stdout.String(), "my-cool-project") {
		t.Error("expected directory name as project name")
	}
}
