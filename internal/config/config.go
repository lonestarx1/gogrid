// Package config handles GoGrid project configuration loading and validation.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// validProviders is the set of supported LLM provider names.
var validProviders = map[string]bool{
	"openai":    true,
	"anthropic": true,
	"gemini":    true,
}

// ProjectConfig is the top-level gogrid.yaml structure.
type ProjectConfig struct {
	// Version is the config schema version. Must be "1".
	Version string `yaml:"version"`
	// Agents maps agent names to their configurations.
	Agents map[string]AgentConfig `yaml:"agents"`
}

// AgentConfig defines a single agent's configuration.
type AgentConfig struct {
	// Model is the LLM model identifier (e.g. "gpt-4o", "claude-sonnet-4-5-20250929").
	Model string `yaml:"model"`
	// Provider is the LLM backend ("openai", "anthropic", or "gemini").
	Provider string `yaml:"provider"`
	// Instructions is the agent's system prompt.
	Instructions string `yaml:"instructions"`
	// Config holds execution parameters.
	Config RunConfig `yaml:"config"`
}

// RunConfig holds agent execution parameters.
type RunConfig struct {
	// MaxTurns limits the number of LLM round-trips. 0 means no limit.
	MaxTurns int `yaml:"max_turns"`
	// MaxTokens limits the LLM response length per turn.
	MaxTokens int `yaml:"max_tokens"`
	// Temperature controls LLM randomness (0.0-1.0). Nil means provider default.
	Temperature *float64 `yaml:"temperature"`
	// Timeout is the maximum wall-clock duration for a run (e.g. "60s", "5m").
	Timeout Duration `yaml:"timeout"`
	// CostBudget is the maximum cost in USD for a single run.
	CostBudget float64 `yaml:"cost_budget"`
}

// Duration wraps time.Duration with YAML string unmarshaling support.
type Duration struct {
	time.Duration
}

// UnmarshalYAML parses a duration string like "30s" or "5m".
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	if value.Value == "" {
		d.Duration = 0
		return nil
	}
	dur, err := time.ParseDuration(value.Value)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", value.Value, err)
	}
	d.Duration = dur
	return nil
}

// MarshalYAML writes the duration as a string.
func (d Duration) MarshalYAML() (any, error) {
	if d.Duration == 0 {
		return "", nil
	}
	return d.Duration.String(), nil
}

// Load reads a gogrid.yaml file, performs environment variable substitution,
// parses the YAML, and validates the result.
func Load(path string) (*ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}

	// Substitute environment variables before parsing.
	substituted := Substitute(string(data))

	var cfg ProjectConfig
	if err := yaml.Unmarshal([]byte(substituted), &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks that the configuration is well-formed.
func (c *ProjectConfig) Validate() error {
	if c.Version != "1" {
		return fmt.Errorf("config: unsupported version %q (expected \"1\")", c.Version)
	}
	if len(c.Agents) == 0 {
		return fmt.Errorf("config: at least one agent is required")
	}
	for name, agent := range c.Agents {
		if agent.Model == "" {
			return fmt.Errorf("config: agent %q: model is required", name)
		}
		if agent.Provider == "" {
			return fmt.Errorf("config: agent %q: provider is required", name)
		}
		if !validProviders[agent.Provider] {
			return fmt.Errorf("config: agent %q: unsupported provider %q (valid: openai, anthropic, gemini)", name, agent.Provider)
		}
	}
	return nil
}
