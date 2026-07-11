// Package ollama implements ports.ProviderPort against the documented Ollama REST API on the
// local host (ADR-019): a thin hand-rolled client rather than importing the full ollama module.
package ollama

import (
	"context"
	"strings"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/provider"
)

// DefaultBaseURL is the documented local Ollama endpoint.
const DefaultBaseURL = "http://localhost:11434"

// Config configures the adapter.
type Config struct {
	BaseURL string // default DefaultBaseURL
	Client  *provider.Client
}

// Adapter is the Ollama provider adapter (local, no authentication).
type Adapter struct {
	client *provider.Client
}

// New builds an adapter.
func New(cfg Config) *Adapter {
	base := cfg.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	c := cfg.Client
	if c == nil {
		c = &provider.Client{}
	}
	c.BaseURL = strings.TrimRight(base, "/")
	return &Adapter{client: c}
}

var _ ports.ProviderPort = (*Adapter)(nil)

type chatReq struct {
	Model    string    `json:"model"`
	Messages []wireMsg `json:"messages"`
	Stream   bool      `json:"stream"`
}

type wireMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func build(req ports.ChatRequest) chatReq {
	cr := chatReq{Model: req.Model, Stream: false}
	for _, m := range req.Messages {
		var sb strings.Builder
		for _, p := range m.Parts {
			if p.Type == "" || p.Type == "text" {
				sb.WriteString(p.Text)
			}
		}
		cr.Messages = append(cr.Messages, wireMsg{Role: m.Role, Content: sb.String()})
	}
	return cr
}

type chatResp struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	PromptEvalCount int `json:"prompt_eval_count"`
	EvalCount       int `json:"eval_count"`
}

// Chat performs one non-streaming chat completion (/api/chat).
func (a *Adapter) Chat(ctx context.Context, req ports.ChatRequest) (ports.ChatResponse, error) {
	var resp chatResp
	if err := a.client.PostJSON(ctx, "/api/chat", build(req), &resp); err != nil {
		return ports.ChatResponse{}, err
	}
	return ports.ChatResponse{
		Message: ports.Message{Role: "assistant", Parts: []ports.ContentPart{{Type: "text", Text: resp.Message.Content}}},
		Usage:   ports.Usage{InputTokens: resp.PromptEvalCount, OutputTokens: resp.EvalCount, CostBasis: "reported"},
	}, nil
}

// ChatStream is not implemented at MVP for Ollama (NDJSON streaming); callers use Chat.
func (a *Adapter) ChatStream(_ context.Context, _ ports.ChatRequest) (ports.Stream[ports.ChatEvent], error) {
	return nil, provider.Unavailable("streaming")
}

// Embed produces embeddings (/api/embed).
func (a *Adapter) Embed(ctx context.Context, req ports.EmbedRequest) (ports.EmbedResponse, error) {
	body := map[string]any{"model": req.Model, "input": req.Inputs}
	var resp struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	if err := a.client.PostJSON(ctx, "/api/embed", body, &resp); err != nil {
		return ports.EmbedResponse{}, err
	}
	return ports.EmbedResponse{Vectors: resp.Embeddings, Usage: ports.Usage{CostBasis: "unavailable"}}, nil
}

// DiscoverModels lists locally available models (/api/tags).
func (a *Adapter) DiscoverModels(ctx context.Context) ([]ports.ModelDescriptor, error) {
	var resp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := a.client.GetJSON(ctx, "/api/tags", &resp); err != nil {
		return nil, err
	}
	out := make([]ports.ModelDescriptor, 0, len(resp.Models))
	for _, m := range resp.Models {
		out = append(out, ports.ModelDescriptor{ID: m.Name, Capabilities: caps()})
	}
	return out, nil
}

// Capabilities returns the declared capability set for local Ollama models. Tool calling and
// vision vary by model; the conservative baseline declares chat and embeddings only.
func (a *Adapter) Capabilities(_ context.Context, _ string) (ports.CapabilitySet, error) {
	return caps(), nil
}

func caps() core.Capabilities {
	return core.Capabilities{core.CapChat, core.CapEmbeddings, core.CapModelDiscovery, core.CapCancellation}
}

// CountTokens is unavailable; estimation is used upstream.
func (a *Adapter) CountTokens(_ context.Context, _ ports.TokenCountRequest) (ports.TokenCount, error) {
	return ports.TokenCount{}, provider.Unavailable("token_counting")
}
