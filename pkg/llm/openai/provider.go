// Package openai implements the GoGrid LLM provider for OpenAI-compatible APIs.
package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

// Provider implements llm.Provider using the official OpenAI Go SDK.
type Provider struct {
	client openai.Client
}

// Option configures the OpenAI provider.
type Option func(*providerConfig)

type providerConfig struct {
	baseURL    string
	httpClient *http.Client
}

// WithBaseURL sets a custom API base URL (for Azure, local models, etc.).
func WithBaseURL(url string) Option {
	return func(c *providerConfig) { c.baseURL = url }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *providerConfig) { c.httpClient = client }
}

// New creates an OpenAI provider with the given API key.
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
		client: openai.NewClient(clientOpts...),
	}
}

// Complete sends a chat completion request to the OpenAI API.
func (p *Provider) Complete(ctx context.Context, params llm.Params) (*llm.Response, error) {
	req, err := toRequest(params)
	if err != nil {
		return nil, fmt.Errorf("openai: build request: %w", err)
	}

	completion, err := p.client.Chat.Completions.New(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openai: completion: %w", err)
	}

	return fromResponse(completion)
}

func toRequest(params llm.Params) (openai.ChatCompletionNewParams, error) {
	var msgs []openai.ChatCompletionMessageParamUnion
	for _, m := range params.Messages {
		switch m.Role {
		case llm.RoleSystem:
			msgs = append(msgs, openai.SystemMessage(m.Content))
		case llm.RoleUser:
			msgs = append(msgs, openai.UserMessage(m.Content))
		case llm.RoleAssistant:
			if len(m.ToolCalls) > 0 {
				var tcs []openai.ChatCompletionMessageToolCallParam
				for _, tc := range m.ToolCalls {
					tcs = append(tcs, openai.ChatCompletionMessageToolCallParam{
						ID: tc.ID,
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      tc.Function,
							Arguments: string(tc.Arguments),
						},
					})
				}
				msgs = append(msgs, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						ToolCalls: tcs,
					},
				})
			} else {
				msgs = append(msgs, openai.AssistantMessage(m.Content))
			}
		case llm.RoleTool:
			msgs = append(msgs, openai.ToolMessage(m.ToolCallID, m.Content))
		}
	}

	var tools []openai.ChatCompletionToolParam
	for _, t := range params.Tools {
		var paramMap openai.FunctionParameters
		if err := json.Unmarshal(t.Parameters, &paramMap); err != nil {
			return openai.ChatCompletionNewParams{}, fmt.Errorf("unmarshal tool %q params: %w", t.Name, err)
		}
		tools = append(tools, openai.ChatCompletionToolParam{
			Function: shared.FunctionDefinitionParam{
				Name:        t.Name,
				Description: openai.String(t.Description),
				Parameters:  paramMap,
			},
		})
	}

	req := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(params.Model),
		Messages: msgs,
	}
	if len(tools) > 0 {
		req.Tools = tools
	}
	if params.Temperature != nil {
		req.Temperature = openai.Float(*params.Temperature)
	}
	if params.MaxTokens > 0 {
		req.MaxCompletionTokens = openai.Int(int64(params.MaxTokens))
	}
	if len(params.StopSequences) > 0 {
		req.Stop = openai.ChatCompletionNewParamsStopUnion{
			OfString: openai.String(params.StopSequences[0]),
		}
	}

	return req, nil
}

func fromResponse(c *openai.ChatCompletion) (*llm.Response, error) {
	if len(c.Choices) == 0 {
		return nil, fmt.Errorf("openai: response contains no choices")
	}

	choice := c.Choices[0]
	msg := llm.Message{
		Role:    llm.RoleAssistant,
		Content: choice.Message.Content,
	}

	for _, tc := range choice.Message.ToolCalls {
		msg.ToolCalls = append(msg.ToolCalls, llm.ToolCall{
			ID:        tc.ID,
			Function:  tc.Function.Name,
			Arguments: json.RawMessage(tc.Function.Arguments),
		})
	}

	return &llm.Response{
		Message: msg,
		Usage: llm.Usage{
			PromptTokens:     int(c.Usage.PromptTokens),
			CompletionTokens: int(c.Usage.CompletionTokens),
			TotalTokens:      int(c.Usage.TotalTokens),
		},
		Model: c.Model,
	}, nil
}
