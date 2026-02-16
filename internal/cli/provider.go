package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/llm/anthropic"
	"github.com/lonestarx1/gogrid/pkg/llm/gemini"
	"github.com/lonestarx1/gogrid/pkg/llm/openai"
)

// envKeys maps provider names to their API key environment variables.
var envKeys = map[string]string{
	"openai":    "OPENAI_API_KEY",
	"anthropic": "ANTHROPIC_API_KEY",
	"gemini":    "GEMINI_API_KEY",
}

// defaultProviderFactory creates providers using API keys from environment variables.
func defaultProviderFactory(ctx context.Context, name string) (llm.Provider, error) {
	envKey, ok := envKeys[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider %q", name)
	}

	apiKey := os.Getenv(envKey)
	if apiKey == "" {
		return nil, fmt.Errorf("%s is not set (required for provider %q)", envKey, name)
	}

	switch name {
	case "openai":
		return openai.New(apiKey), nil
	case "anthropic":
		return anthropic.New(apiKey), nil
	case "gemini":
		return gemini.New(ctx, apiKey)
	default:
		return nil, fmt.Errorf("unknown provider %q", name)
	}
}
