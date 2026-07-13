// Package openaicompat implements ports.ProviderPort against any OpenAI-compatible Chat
// Completions API (ADR-019, FR-PROV-081): the generic adapter that covers OpenAI itself and
// the many services and local servers exposing the same surface.
package openaicompat

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/provider"
)

// Config configures the adapter.
type Config struct {
	BaseURL string // e.g. "https://api.openai.com/v1"
	APIKey  string // optional (local servers may need none)
	Client  *provider.Client
}

// Adapter is an OpenAI-compatible provider adapter.
type Adapter struct {
	client *provider.Client
}

// New builds an adapter from a Config.
func New(cfg Config) *Adapter {
	c := cfg.Client
	if c == nil {
		c = &provider.Client{}
	}
	c.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	if c.Headers == nil {
		c.Headers = map[string]string{}
	}
	if cfg.APIKey != "" {
		c.Headers["Authorization"] = "Bearer " + cfg.APIKey
	}
	return &Adapter{client: c}
}

var _ ports.ProviderPort = (*Adapter)(nil)

// wire types for the OpenAI chat-completions surface.
type chatReq struct {
	Model           string        `json:"model"`
	Messages        []wireMessage `json:"messages"`
	Tools           []wireTool    `json:"tools,omitempty"`
	Stream          bool          `json:"stream,omitempty"`
	Temperature     *float64      `json:"temperature,omitempty"`
	MaxTokens       *int          `json:"max_tokens,omitempty"`
	ReasoningEffort string        `json:"reasoning_effort,omitempty"` // reasoning models only
}

type wireMessage struct {
	Role       string         `json:"role"`
	Content    string         `json:"content"`
	ToolCalls  []wireToolCall `json:"tool_calls,omitempty"`   // assistant messages that called tools
	ToolCallID string         `json:"tool_call_id,omitempty"` // tool messages, correlating the result
}

type wireToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"` // always "function"
	Function wireToolCallFunc `json:"function"`
}

type wireToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // a JSON string, per the OpenAI schema
}

type wireTool struct {
	Type     string       `json:"type"`
	Function wireFunction `json:"function"`
}

type wireFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type chatResp struct {
	Choices []struct {
		Message struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func buildMessages(msgs []ports.Message) []wireMessage {
	out := make([]wireMessage, 0, len(msgs))
	for _, m := range msgs {
		wm := wireMessage{Role: m.Role}
		var sb strings.Builder
		for _, p := range m.Parts {
			switch p.Type {
			case "tool_call":
				args := string(p.ToolInput)
				if strings.TrimSpace(args) == "" {
					args = "{}" // the API requires a JSON string, never empty
				}
				wm.ToolCalls = append(wm.ToolCalls, wireToolCall{
					ID: p.ToolCallID, Type: "function",
					Function: wireToolCallFunc{Name: p.ToolName, Arguments: args},
				})
			case "tool_result":
				wm.ToolCallID = p.ToolCallID
				sb.WriteString(p.Text)
			default: // "" or "text"
				sb.WriteString(p.Text)
			}
		}
		wm.Content = sb.String()
		out = append(out, wm)
	}
	return out
}

func buildRequest(req ports.ChatRequest, stream bool) chatReq {
	cr := chatReq{Model: req.Model, Messages: buildMessages(req.Messages), Stream: stream}
	// Sampling controls and reasoning effort ride on ModelParams; omitempty keeps the body identical
	// to before when they are unset, so non-reasoning models are unaffected.
	cr.Temperature = req.Params.Temperature
	cr.MaxTokens = req.Params.MaxTokens
	if req.Params.Extra != nil {
		if v, ok := req.Params.Extra["reasoning_effort"].(string); ok && v != "" {
			cr.ReasoningEffort = v
		}
	}
	for _, t := range req.Tools {
		fn := wireFunction{Name: t.Name, Description: t.Description}
		// Advertise the tool's input schema so the model produces well-formed arguments; without it
		// the model is told a tool exists but not what to pass, and may skip or mis-call it.
		if len(t.InputSchema) > 0 {
			var params map[string]any
			if err := json.Unmarshal(t.InputSchema, &params); err == nil {
				fn.Parameters = params
			}
		}
		cr.Tools = append(cr.Tools, wireTool{Type: "function", Function: fn})
	}
	return cr
}

// Chat performs one non-streaming completion.
func (a *Adapter) Chat(ctx context.Context, req ports.ChatRequest) (ports.ChatResponse, error) {
	var resp chatResp
	if err := a.client.PostJSON(ctx, "/chat/completions", buildRequest(req, false), &resp); err != nil {
		return ports.ChatResponse{}, err
	}
	if len(resp.Choices) == 0 {
		return ports.ChatResponse{}, provider.Unavailable("empty completion")
	}
	ch := resp.Choices[0].Message
	out := ports.ChatResponse{
		Message: ports.Message{Role: "assistant", Parts: []ports.ContentPart{{Type: "text", Text: ch.Content}}},
		Usage: ports.Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
			CostBasis:    "reported",
		},
	}
	for _, tc := range ch.ToolCalls {
		out.ToolCalls = append(out.ToolCalls, ports.ToolCall{ID: tc.ID, Name: tc.Function.Name, Input: []byte(tc.Function.Arguments)})
	}
	return out, nil
}

// Embed produces embeddings for a batch.
func (a *Adapter) Embed(ctx context.Context, req ports.EmbedRequest) (ports.EmbedResponse, error) {
	body := map[string]any{"model": req.Model, "input": req.Inputs}
	var resp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
		Usage struct {
			PromptTokens int `json:"prompt_tokens"`
		} `json:"usage"`
	}
	if err := a.client.PostJSON(ctx, "/embeddings", body, &resp); err != nil {
		return ports.EmbedResponse{}, err
	}
	out := ports.EmbedResponse{Usage: ports.Usage{InputTokens: resp.Usage.PromptTokens, CostBasis: "reported"}}
	for _, d := range resp.Data {
		out.Vectors = append(out.Vectors, d.Embedding)
	}
	return out, nil
}

// DiscoverModels lists available models.
func (a *Adapter) DiscoverModels(ctx context.Context) ([]ports.ModelDescriptor, error) {
	var resp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := a.client.GetJSON(ctx, "/models", &resp); err != nil {
		return nil, err
	}
	out := make([]ports.ModelDescriptor, 0, len(resp.Data))
	for _, m := range resp.Data {
		out = append(out, ports.ModelDescriptor{ID: m.ID, Capabilities: declaredCaps()})
	}
	return out, nil
}

// Capabilities returns the conservative declared capability set for OpenAI-compatible models.
func (a *Adapter) Capabilities(_ context.Context, _ string) (ports.CapabilitySet, error) {
	return declaredCaps(), nil
}

func declaredCaps() core.Capabilities {
	return core.Capabilities{
		core.CapChat, core.CapStreaming, core.CapToolCalling,
		core.CapStructuredOutputs, core.CapEmbeddings, core.CapTokenUsageReporting,
		core.CapModelDiscovery, core.CapCancellation,
	}
}

// CountTokens is unavailable on the generic surface (no official counting endpoint); the
// Context Manager falls back to estimation.
func (a *Adapter) CountTokens(_ context.Context, _ ports.TokenCountRequest) (ports.TokenCount, error) {
	return ports.TokenCount{}, provider.Unavailable("token_counting")
}
