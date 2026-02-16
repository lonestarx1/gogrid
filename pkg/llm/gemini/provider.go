// Package gemini implements the GoGrid LLM provider for the Google Gemini API.
package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lonestarx1/gogrid/pkg/llm"
	"google.golang.org/genai"
)

// Provider implements llm.Provider using the official Google GenAI Go SDK.
type Provider struct {
	client *genai.Client
}

// Option configures the Gemini provider.
type Option func(*providerConfig)

type providerConfig struct {
	backend genai.Backend
}

// WithVertexAI configures the provider to use Vertex AI instead of the Gemini API.
func WithVertexAI() Option {
	return func(c *providerConfig) { c.backend = genai.BackendVertexAI }
}

// New creates a Gemini provider with the given API key.
func New(ctx context.Context, apiKey string, opts ...Option) (*Provider, error) {
	cfg := &providerConfig{
		backend: genai.BackendGeminiAPI,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: cfg.backend,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini: create client: %w", err)
	}

	return &Provider{client: client}, nil
}

// Complete sends a generateContent request to the Gemini API.
func (p *Provider) Complete(ctx context.Context, params llm.Params) (*llm.Response, error) {
	contents, config := toRequest(params)

	resp, err := p.client.Models.GenerateContent(ctx, params.Model, contents, config)
	if err != nil {
		return nil, fmt.Errorf("gemini: generate content: %w", err)
	}

	return fromResponse(resp, params.Model), nil
}

func toRequest(params llm.Params) ([]*genai.Content, *genai.GenerateContentConfig) {
	var sysInstr *genai.Content
	var contents []*genai.Content

	for _, m := range params.Messages {
		switch m.Role {
		case llm.RoleSystem:
			sysInstr = &genai.Content{
				Parts: []*genai.Part{genai.NewPartFromText(m.Content)},
			}

		case llm.RoleUser:
			contents = append(contents, &genai.Content{
				Role:  "user",
				Parts: []*genai.Part{genai.NewPartFromText(m.Content)},
			})

		case llm.RoleAssistant:
			c := &genai.Content{Role: "model"}
			if m.Content != "" {
				c.Parts = append(c.Parts, genai.NewPartFromText(m.Content))
			}
			for _, tc := range m.ToolCalls {
				var args map[string]any
				_ = json.Unmarshal(tc.Arguments, &args)
				c.Parts = append(c.Parts, genai.NewPartFromFunctionCall(tc.Function, args))
			}
			contents = append(contents, c)

		case llm.RoleTool:
			contents = append(contents, &genai.Content{
				Role: "user",
				Parts: []*genai.Part{
					genai.NewPartFromFunctionResponse(m.ToolCallID, map[string]any{
						"result": m.Content,
					}),
				},
			})
		}
	}

	// Build tools.
	var tools []*genai.Tool
	if len(params.Tools) > 0 {
		var decls []*genai.FunctionDeclaration
		for _, t := range params.Tools {
			fd := &genai.FunctionDeclaration{
				Name:        t.Name,
				Description: t.Description,
			}
			if len(t.Parameters) > 0 {
				var schema genai.Schema
				_ = json.Unmarshal(t.Parameters, &schema)
				fd.Parameters = &schema
			}
			decls = append(decls, fd)
		}
		tools = []*genai.Tool{{FunctionDeclarations: decls}}
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: sysInstr,
		Tools:             tools,
	}
	if params.Temperature != nil {
		t := float32(*params.Temperature)
		config.Temperature = &t
	}
	if params.MaxTokens > 0 {
		config.MaxOutputTokens = int32(params.MaxTokens)
	}
	if len(params.StopSequences) > 0 {
		config.StopSequences = params.StopSequences
	}

	return contents, config
}

func fromResponse(resp *genai.GenerateContentResponse, model string) *llm.Response {
	msg := llm.Message{Role: llm.RoleAssistant}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, p := range resp.Candidates[0].Content.Parts {
			if p.Text != "" {
				msg.Content += p.Text
			}
			if p.FunctionCall != nil {
				args, _ := json.Marshal(p.FunctionCall.Args)
				msg.ToolCalls = append(msg.ToolCalls, llm.ToolCall{
					ID:        p.FunctionCall.Name,
					Function:  p.FunctionCall.Name,
					Arguments: args,
				})
			}
		}
	}

	var usage llm.Usage
	if resp.UsageMetadata != nil {
		usage = llm.Usage{
			PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
			CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
		}
	}

	return &llm.Response{
		Message: msg,
		Usage:   usage,
		Model:   model,
	}
}
