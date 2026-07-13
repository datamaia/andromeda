// Package openaichatgpt adapts a ChatGPT-subscription OAuth session to the ProviderPort by
// calling OpenAI's Codex backend Responses API (https://chatgpt.com/backend-api/codex/responses).
// This is the transport OpenAI's own Codex CLI uses for "sign in with ChatGPT"; it is NOT the
// public Chat Completions API. The subscription token is an OAuth session, never an API key.
//
// The request shape (Responses API with top-level `instructions`, `input` items, `store:false`,
// and encrypted-reasoning include), the required headers (ChatGPT-Account-ID, OpenAI-Beta), and
// the model set follow the verified Codex login behavior. End-to-end inference requires a live
// ChatGPT login; the request/response contract here is exercised by adapter_test.go against a
// stub server.
package openaichatgpt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/datamaia/andromeda/internal/ports"
)

// DefaultBaseURL is the Codex backend the ChatGPT subscription token authenticates against.
const DefaultBaseURL = "https://chatgpt.com/backend-api/codex"

// DefaultInstructions is the Codex identity the backend expects as the top-level `instructions`.
// The reference Codex login sends the full codex-rs/core/<model>_prompt.md; this identifying
// preamble is sent when no fuller prompt is supplied and can be overridden via Config.Instructions.
const DefaultInstructions = "You are Codex, based on GPT-5. You are running as a coding agent in the Codex CLI on a user's computer."

// TokenSource returns a currently-valid access token and the ChatGPT account id, refreshing the
// underlying OAuth token as needed. It is supplied by the composition root.
type TokenSource func(ctx context.Context) (accessToken, accountID string, err error)

// Config configures the adapter.
type Config struct {
	BaseURL      string
	Token        TokenSource
	Instructions string
	HTTPClient   *http.Client
}

// Adapter implements ports.ProviderPort over the Codex backend Responses API.
type Adapter struct {
	baseURL string
	token   TokenSource
	instr   string
	hc      *http.Client
}

var _ ports.ProviderPort = (*Adapter)(nil)

// New builds an adapter. A nil Token makes Chat fail with an actionable "sign in" error.
func New(cfg Config) *Adapter {
	a := &Adapter{baseURL: cfg.BaseURL, token: cfg.Token, instr: cfg.Instructions, hc: cfg.HTTPClient}
	if a.baseURL == "" {
		a.baseURL = DefaultBaseURL
	}
	if a.instr == "" {
		a.instr = DefaultInstructions
	}
	if a.hc == nil {
		a.hc = http.DefaultClient
	}
	return a
}

// legacyModels maps retired model ids to their current replacements (Codex backend rejects some).
var legacyModels = map[string]string{
	"gpt-5-codex":       "gpt-5.1-codex",
	"codex-mini-latest": "gpt-5.1-codex-mini",
	"gpt-5":             "gpt-5.1",
	"llama3":            "gpt-5.1-codex", // the CLI default; ChatGPT has no llama
}

func resolveModel(m string) string {
	if r, ok := legacyModels[m]; ok {
		return r
	}
	if m == "" {
		return "gpt-5.1-codex"
	}
	return m
}

// --- Responses API wire types ---

type responsesRequest struct {
	Model        string          `json:"model"`
	Instructions string          `json:"instructions"`
	Input        []responsesItem `json:"input"`
	Tools        []responsesTool `json:"tools,omitempty"`
	Store        bool            `json:"store"`
	Stream       bool            `json:"stream"`
	Include      []string        `json:"include,omitempty"`
}

type responsesItem struct {
	Type    string             `json:"type"` // "message"
	Role    string             `json:"role"`
	Content []responsesContent `json:"content"`
}

type responsesContent struct {
	Type string `json:"type"` // "input_text" (user) | "output_text" (assistant)
	Text string `json:"text"`
}

type responsesTool struct {
	Type        string          `json:"type"` // "function"
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

type responsesResult struct {
	Output []responsesOutput `json:"output"`
	Usage  *struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type responsesOutput struct {
	Type      string             `json:"type"` // "message" | "function_call"
	Role      string             `json:"role"`
	Content   []responsesContent `json:"content"`
	Name      string             `json:"name"`      // function_call
	Arguments string             `json:"arguments"` // function_call (JSON string)
	CallID    string             `json:"call_id"`   // function_call
}

// buildRequest maps a ports.ChatRequest into a Responses request. Any system message is folded
// into the top-level instructions (the backend's own Codex prompt must remain the base).
func (a *Adapter) buildRequest(req ports.ChatRequest) responsesRequest {
	instr := a.instr
	input := make([]responsesItem, 0, len(req.Messages))
	for _, m := range req.Messages {
		text := plainText(m)
		switch m.Role {
		case "system":
			if text != "" {
				instr += "\n\n" + text
			}
		case "assistant":
			input = append(input, responsesItem{Type: "message", Role: "assistant",
				Content: []responsesContent{{Type: "output_text", Text: text}}})
		default: // user (and tool results surfaced as user text)
			input = append(input, responsesItem{Type: "message", Role: "user",
				Content: []responsesContent{{Type: "input_text", Text: text}}})
		}
	}
	tools := make([]responsesTool, 0, len(req.Tools))
	for _, t := range req.Tools {
		tools = append(tools, responsesTool{Type: "function", Name: t.Name, Description: t.Description, Parameters: json.RawMessage(t.InputSchema)})
	}
	return responsesRequest{
		Model: resolveModel(req.Model), Instructions: instr, Input: input, Tools: tools,
		Store: false, Stream: false, Include: []string{"reasoning.encrypted_content"},
	}
}

func plainText(m ports.Message) string {
	var b strings.Builder
	for _, p := range m.Parts {
		if p.Type == "" || p.Type == "text" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}

// Chat performs one non-streaming inference against the Codex backend.
func (a *Adapter) Chat(ctx context.Context, req ports.ChatRequest) (ports.ChatResponse, error) {
	if a.token == nil {
		return ports.ChatResponse{}, provErr("E-PROV-009", "not signed in to ChatGPT — run: andromeda auth login openai-chatgpt")
	}
	access, account, err := a.token(ctx)
	if err != nil {
		return ports.ChatResponse{}, err
	}
	payload, err := json.Marshal(a.buildRequest(req))
	if err != nil {
		return ports.ChatResponse{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/responses", bytes.NewReader(payload))
	if err != nil {
		return ports.ChatResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+access)
	httpReq.Header.Set("chatgpt-account-id", account)
	httpReq.Header.Set("OpenAI-Beta", "responses=experimental")
	httpReq.Header.Set("originator", "codex_cli_rs")

	resp, err := a.hc.Do(httpReq)
	if err != nil {
		return ports.ChatResponse{}, provErr("E-PROV-001", "ChatGPT backend request failed: "+err.Error())
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode == http.StatusUnauthorized {
		return ports.ChatResponse{}, provErr("E-PROV-009", "ChatGPT session rejected (401) — sign in again: andromeda auth login openai-chatgpt")
	}
	if resp.StatusCode >= 400 {
		return ports.ChatResponse{}, provErr("E-PROV-001", fmt.Sprintf("ChatGPT backend error (%d): %s", resp.StatusCode, snippet(body)))
	}
	var out responsesResult
	if err := json.Unmarshal(body, &out); err != nil {
		return ports.ChatResponse{}, provErr("E-PROV-001", "ChatGPT backend response was not valid JSON")
	}
	if out.Error != nil && out.Error.Message != "" {
		return ports.ChatResponse{}, provErr("E-PROV-001", "ChatGPT backend error: "+out.Error.Message)
	}
	return assembleResponse(out), nil
}

func assembleResponse(out responsesResult) ports.ChatResponse {
	var text strings.Builder
	var calls []ports.ToolCall
	for _, o := range out.Output {
		switch o.Type {
		case "function_call":
			calls = append(calls, ports.ToolCall{ID: o.CallID, Name: o.Name, Input: ports.JSON(o.Arguments)})
		default: // "message"
			for _, c := range o.Content {
				if c.Type == "output_text" || c.Type == "text" {
					text.WriteString(c.Text)
				}
			}
		}
	}
	resp := ports.ChatResponse{
		Message:   ports.Message{Role: "assistant", Parts: []ports.ContentPart{{Type: "text", Text: text.String()}}},
		ToolCalls: calls,
	}
	if out.Usage != nil {
		resp.Usage = ports.Usage{InputTokens: out.Usage.InputTokens, OutputTokens: out.Usage.OutputTokens, CostBasis: "reported"}
	}
	return resp
}

// ChatStream falls back to a single Chat result delivered as one content event then terminal.
func (a *Adapter) ChatStream(ctx context.Context, req ports.ChatRequest) (ports.Stream[ports.ChatEvent], error) {
	resp, err := a.Chat(ctx, req)
	if err != nil {
		return nil, err
	}
	usage := resp.Usage
	return &sliceStream{events: []ports.ChatEvent{
		{Kind: "content", ContentDelta: plainText(resp.Message)},
		{Kind: "usage", Usage: &usage},
		{Kind: "terminal", Terminal: true},
	}}, nil
}

// sliceStream replays a fixed slice of events as a ports.Stream (Chat is non-streaming upstream).
type sliceStream struct {
	events []ports.ChatEvent
	i      int
}

func (s *sliceStream) Next(context.Context) (ports.ChatEvent, error) {
	if s.i >= len(s.events) {
		return ports.ChatEvent{}, ports.ErrEndOfStream
	}
	e := s.events[s.i]
	s.i++
	return e, nil
}

func (s *sliceStream) Close() error { return nil }

// Embed is unsupported: the ChatGPT backend exposes no embeddings endpoint here.
func (a *Adapter) Embed(context.Context, ports.EmbedRequest) (ports.EmbedResponse, error) {
	return ports.EmbedResponse{}, provErr("E-PROV-002", "the ChatGPT provider does not support embeddings")
}

// DiscoverModels reports the Codex models a ChatGPT subscription can address (plan-gated at use).
func (a *Adapter) DiscoverModels(context.Context) ([]ports.ModelDescriptor, error) {
	ids := []string{"gpt-5.1-codex", "gpt-5.1-codex-max", "gpt-5.1-codex-mini", "gpt-5.2-codex", "gpt-5.1", "gpt-5.2"}
	out := make([]ports.ModelDescriptor, 0, len(ids))
	for _, id := range ids {
		out = append(out, ports.ModelDescriptor{ID: id, DisplayName: id})
	}
	return out, nil
}

func (a *Adapter) Capabilities(context.Context, string) (ports.CapabilitySet, error) {
	return ports.CapabilitySet{}, nil
}

func (a *Adapter) CountTokens(context.Context, ports.TokenCountRequest) (ports.TokenCount, error) {
	return ports.TokenCount{}, nil
}

func provErr(code, msg string) error {
	return &ports.PortError{Code: code, Category: "provider", Severity: "error", Message: msg}
}

func snippet(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 300 {
		return s[:300] + "…"
	}
	return s
}
