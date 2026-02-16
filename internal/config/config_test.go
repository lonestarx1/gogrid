package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		env     map[string]string
		wantErr string
	}{
		{
			name: "valid config",
			yaml: `version: "1"
agents:
  researcher:
    model: claude-sonnet-4-5-20250929
    provider: anthropic
    instructions: You are a researcher.
    config:
      max_turns: 5
      max_tokens: 4096
      temperature: 0.7
      timeout: 60s
      cost_budget: 1.0
`,
		},
		{
			name: "multiple agents",
			yaml: `version: "1"
agents:
  writer:
    model: gpt-4o
    provider: openai
    instructions: You write content.
  reviewer:
    model: gemini-2.5-pro
    provider: gemini
    instructions: You review content.
`,
		},
		{
			name: "env substitution",
			yaml: `version: "1"
agents:
  test:
    model: ${TEST_MODEL}
    provider: openai
`,
			env: map[string]string{"TEST_MODEL": "gpt-4o-mini"},
		},
		{
			name: "env substitution with default",
			yaml: `version: "1"
agents:
  test:
    model: ${TEST_MODEL:-gpt-4o}
    provider: openai
`,
		},
		{
			name:    "bad version",
			yaml:    `version: "2"`,
			wantErr: `unsupported version "2"`,
		},
		{
			name:    "missing version",
			yaml:    `agents: {}`,
			wantErr: `unsupported version ""`,
		},
		{
			name: "no agents",
			yaml: `version: "1"
agents: {}
`,
			wantErr: "at least one agent is required",
		},
		{
			name: "missing model",
			yaml: `version: "1"
agents:
  test:
    provider: openai
`,
			wantErr: `agent "test": model is required`,
		},
		{
			name: "missing provider",
			yaml: `version: "1"
agents:
  test:
    model: gpt-4o
`,
			wantErr: `agent "test": provider is required`,
		},
		{
			name: "invalid provider",
			yaml: `version: "1"
agents:
  test:
    model: some-model
    provider: invalid
`,
			wantErr: `unsupported provider "invalid"`,
		},
		{
			name:    "bad yaml",
			yaml:    `{{{`,
			wantErr: "parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			dir := t.TempDir()
			path := filepath.Join(dir, "gogrid.yaml")
			if err := os.WriteFile(path, []byte(tt.yaml), 0o644); err != nil {
				t.Fatal(err)
			}

			cfg, err := Load(path)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Version != "1" {
				t.Errorf("version = %q, want %q", cfg.Version, "1")
			}
			if len(cfg.Agents) == 0 {
				t.Error("expected at least one agent")
			}
		})
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/gogrid.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestDuration_Parsing(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantSec float64
		wantErr bool
	}{
		{name: "seconds", yaml: "30s", wantSec: 30},
		{name: "minutes", yaml: "5m", wantSec: 300},
		{name: "complex", yaml: "1m30s", wantSec: 90},
		{name: "empty", yaml: "", wantSec: 0},
		{name: "invalid", yaml: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgYAML := `version: "1"
agents:
  test:
    model: gpt-4o
    provider: openai
    config:
      timeout: ` + tt.yaml + "\n"

			dir := t.TempDir()
			path := filepath.Join(dir, "gogrid.yaml")
			if err := os.WriteFile(path, []byte(cfgYAML), 0o644); err != nil {
				t.Fatal(err)
			}

			cfg, err := Load(path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := cfg.Agents["test"].Config.Timeout.Seconds()
			if got != tt.wantSec {
				t.Errorf("timeout = %vs, want %vs", got, tt.wantSec)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
