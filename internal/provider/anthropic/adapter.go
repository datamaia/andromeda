// Package anthropic implements ports.ProviderPort against the Anthropic Messages API
// (ADR-019). Only the documented public surface is used; no undocumented behavior is assumed.
package anthropic

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strings"
	"sync"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/provider"
)

// APIVersion is the documented anthropic-version header value used by the adapter.
const APIVersion = "2023-06-01"

// DefaultMaxTokens is used when a request does not set a max-tokens parameter (the Messages
// API requires max_tokens).
const DefaultMaxTokens = 4096

// Config configures the adapter.
type Config struct {
	BaseURL string // default "https://api.anthropic.com/v1"
	APIKey  string
	Client  *provider.Client
}

// Adapter is the Anthropic provider adapter.
type Adapter struct {
	client *provider.Client
}

// New builds an adapter.
func New(cfg Config) *Adapter {
	base := cfg.BaseURL
	if base == "" {
		base = "https://api.anthropic.com/v1"
	}
	c := cfg.Client
	if c == nil {
		c = &provider.Client{}
	}
	c.BaseURL = strings.TrimRight(base, "/")
	if c.Headers == nil {
		c.Headers = map[string]string{}
	}
	if cfg.APIKey != "" {
		c.Headers["x-api-key"] = cfg.APIKey
	}
	c.Headers["anthropic-version"] = APIVersion
	return &Adapter{client: c}
}

var _ ports.ProviderPort = (*Adapter)(nil)

type msgReq struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []wireMsg `json:"messages"`
	Stream    bool      `json:"stream,omitempty"`
}

type wireMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func build(req ports.ChatRequest, stream bool) msgReq {
	mr := msgReq{Model: req.Model, MaxTokens: DefaultMaxTokens, Stream: stream}
	if req.Params.MaxTokens != nil {
		mr.MaxTokens = *req.Params.MaxTokens
	}
	for _, m := range req.Messages {
		var sb strings.Builder
		for _, p := range m.Parts {
			if p.Type == "" || p.Type == "text" {
				sb.WriteString(p.Text)
			}
		}
		if m.Role == "system" {
			mr.System = sb.String()
			continue
		}
		mr.Messages = append(mr.Messages, wireMsg{Role: m.Role, Content: sb.String()})
	}
	return mr
}

type msgResp struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Chat performs one non-streaming message completion.
func (a *Adapter) Chat(ctx context.Context, req ports.ChatRequest) (ports.ChatResponse, error) {
	var resp msgResp
	if err := a.client.PostJSON(ctx, "/messages", build(req, false), &resp); err != nil {
		return ports.ChatResponse{}, err
	}
	var sb strings.Builder
	for _, c := range resp.Content {
		if c.Type == "text" {
			sb.WriteString(c.Text)
		}
	}
	return ports.ChatResponse{
		Message: ports.Message{Role: "assistant", Parts: []ports.ContentPart{{Type: "text", Text: sb.String()}}},
		Usage:   ports.Usage{InputTokens: resp.Usage.InputTokens, OutputTokens: resp.Usage.OutputTokens, CostBasis: "reported"},
	}, nil
}

// ChatStream streams a message completion, parsing Anthropic SSE content_block_delta events.
func (a *Adapter) ChatStream(ctx context.Context, req ports.ChatRequest) (ports.Stream[ports.ChatEvent], error) {
	body, err := a.client.PostStream(ctx, "/messages", build(req, true))
	if err != nil {
		return nil, err
	}
	return &stream{body: body, sc: bufio.NewScanner(body)}, nil
}

type stream struct {
	body       io.ReadCloser
	sc         *bufio.Scanner
	mu         sync.Mutex
	usage      ports.Usage
	terminated bool
	closed     bool
}

func (s *stream) Next(ctx context.Context) (ports.ChatEvent, error) {
	if err := ctx.Err(); err != nil {
		return ports.ChatEvent{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.terminated {
		return ports.ChatEvent{}, ports.ErrEndOfStream
	}
	for s.sc.Scan() {
		line := strings.TrimSpace(s.sc.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		var ev struct {
			Type  string `json:"type"`
			Delta struct {
				Text string `json:"text"`
			} `json:"delta"`
			Usage *struct {
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}
		if json.Unmarshal([]byte(data), &ev) != nil {
			continue
		}
		switch ev.Type {
		case "content_block_delta":
			if ev.Delta.Text != "" {
				return ports.ChatEvent{Kind: "content", ContentDelta: ev.Delta.Text}, nil
			}
		case "message_delta":
			if ev.Usage != nil {
				s.usage.OutputTokens = ev.Usage.OutputTokens
			}
		case "message_stop":
			s.terminated = true
			u := s.usage
			return ports.ChatEvent{Kind: "terminal", Terminal: true, Usage: &u}, nil
		}
	}
	s.terminated = true
	u := s.usage
	return ports.ChatEvent{Kind: "terminal", Terminal: true, Usage: &u}, nil
}

func (s *stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.body.Close()
}

// Embed is unavailable: the Messages API does not provide embeddings.
func (a *Adapter) Embed(_ context.Context, _ ports.EmbedRequest) (ports.EmbedResponse, error) {
	return ports.EmbedResponse{}, provider.Unavailable("embeddings")
}

// DiscoverModels returns no models: the adapter relies on configured model identifiers rather
// than assuming an enumeration endpoint shape.
func (a *Adapter) DiscoverModels(_ context.Context) ([]ports.ModelDescriptor, error) {
	return nil, nil
}

// Capabilities returns the declared capability set for Anthropic models.
func (a *Adapter) Capabilities(_ context.Context, _ string) (ports.CapabilitySet, error) {
	return core.Capabilities{
		core.CapChat, core.CapStreaming, core.CapToolCalling, core.CapVision,
		core.CapReasoning, core.CapStructuredOutputs, core.CapTokenUsageReporting,
		core.CapCancellation,
	}, nil
}

// CountTokens is not wired to the count-tokens endpoint at MVP; estimation is used upstream.
func (a *Adapter) CountTokens(_ context.Context, _ ports.TokenCountRequest) (ports.TokenCount, error) {
	return ports.TokenCount{}, provider.Unavailable("token_counting")
}
