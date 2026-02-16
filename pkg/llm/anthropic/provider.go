// Package anthropic implements the GoGrid LLM provider for the Anthropic Messages API.
package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/lonestarx1/gogrid/pkg/llm"
)

// Provider implements llm.Provider using the official Anthropic Go SDK.
type Provider struct {
	client anthropic.Client
}

// Option configures the Anthropic provider.
type Option func(*providerConfig)

type providerConfig struct {
	baseURL    string
	httpClient *http.Client
}

// WithBaseURL sets a custom API base URL.
func WithBaseURL(url string) Option {
	return func(c *providerConfig) { c.baseURL = url }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *providerConfig) { c.httpClient = client }
}

// New creates an Anthropic provider with the given API key.
func New(apiKey string, opts ...Option) *Provider {
	cfg := &providerConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	clientOpts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}
	if cfg.baseURL != "" {
		clientOpts = append(clientOpts, option.WithBaseURL(cfg.baseURL))
	}
	if cfg.httpClient != nil {
		clientOpts = append(clientOpts, option.WithHTTPClient(cfg.httpClient))
	}

	return &Provider{
		client: anthropic.NewClient(clientOpts...),
	}
}

// Complete sends a message request to the Anthropic API.
func (p *Provider) Complete(ctx context.Context, params llm.Params) (*llm.Response, error) {
	req := toRequest(params)

	message, err := p.client.Messages.New(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("anthropic: messages: %w", err)
	}

	return fromResponse(message), nil
}

func toRequest(params llm.Params) anthropic.MessageNewParams {
	maxTokens := int64(params.MaxTokens)
	if maxTokens == 0 {
		maxTokens = 4096
	}

	// Extract system message and build conversation messages.
	var system []anthropic.TextBlockParam
	var msgs []anthropic.MessageParam

	for _, m := range params.Messages {
		switch m.Role {
		case llm.RoleSystem:
			system = append(system, anthropic.TextBlockParam{
				Text: m.Content,
			})

		case llm.RoleUser:
			msgs = append(msgs, anthropic.NewUserMessage(
				anthropic.NewTextBlock(m.Content),
			))

		case llm.RoleAssistant:
			if len(m.ToolCalls) > 0 {
				var blocks []anthropic.ContentBlockParamUnion
				if m.Content != "" {
					blocks = append(blocks, anthropic.NewTextBlock(m.Content))
				}
				for _, tc := range m.ToolCalls {
					var input map[string]any
					_ = json.Unmarshal(tc.Arguments, &input)
					blocks = append(blocks, anthropic.ContentBlockParamUnion{
						OfToolUse: &anthropic.ToolUseBlockParam{
							ID:    tc.ID,
							Name:  tc.Function,
							Input: input,
						},
					})
				}
				msgs = append(msgs, anthropic.NewAssistantMessage(blocks...))
			} else {
				msgs = append(msgs, anthropic.NewAssistantMessage(
					anthropic.NewTextBlock(m.Content),
				))
			}

		case llm.RoleTool:
			msgs = append(msgs, anthropic.NewUserMessage(
				anthropic.NewToolResultBlock(m.ToolCallID, m.Content, false),
			))
		}
	}

	var tools []anthropic.ToolUnionParam
	for _, t := range params.Tools {
		var props map[string]any
		_ = json.Unmarshal(t.Parameters, &props)
		tools = append(tools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        t.Name,
				Description: anthropic.String(t.Description),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: props["properties"],
				},
			},
		})
	}

	req := anthropic.MessageNewParams{
		Model:     anthropic.Model(params.Model),
		MaxTokens: maxTokens,
		Messages:  msgs,
	}
	if len(system) > 0 {
		req.System = system
	}
	if len(tools) > 0 {
		req.Tools = tools
	}
	if params.Temperature != nil {
		req.Temperature = anthropic.Float(*params.Temperature)
	}
	if len(params.StopSequences) > 0 {
		req.StopSequences = params.StopSequences
	}

	return req
}

func fromResponse(msg *anthropic.Message) *llm.Response {
	result := llm.Message{Role: llm.RoleAssistant}

	for _, block := range msg.Content {
		switch block.Type {
		case "text":
			result.Content += block.AsText().Text
		case "tool_use":
			tu := block.AsToolUse()
			result.ToolCalls = append(result.ToolCalls, llm.ToolCall{
				ID:        tu.ID,
				Function:  tu.Name,
				Arguments: tu.Input,
			})
		}
	}

	return &llm.Response{
		Message: result,
		Usage: llm.Usage{
			PromptTokens:     int(msg.Usage.InputTokens),
			CompletionTokens: int(msg.Usage.OutputTokens),
			TotalTokens:      int(msg.Usage.InputTokens + msg.Usage.OutputTokens),
		},
		Model: string(msg.Model),
	}
}
