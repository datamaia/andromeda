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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

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

// defaultCodexModel is the newest model the ChatGPT-subscription Codex backend actually serves
// (used when no model is set). Verified against the live /responses endpoint 2026-07-13: the backend
// accepts only gpt-5.5 and gpt-5.4 — NOT gpt-5.6 ("not supported when using Codex with a ChatGPT
// account"), nor 5.3/older, nor any codex-suffixed variant. Re-verify as OpenAI enables new ones.
const defaultCodexModel = "gpt-5.5"

// legacyModels maps ids the Codex backend rejects onto the current default: the CLI's own "llama3"
// default, retired numbered bases, and the codex-suffixed variants the subscription backend has
// dropped (including the former default gpt-5.1-codex). Newer ids the backend has not enabled yet
// (gpt-5.6, gpt-5.7) are deliberately NOT remapped — they pass through so the backend's own clear
// "not supported … with a ChatGPT account" message surfaces, rather than a silent downgrade.
var legacyModels = map[string]string{
	"llama3":             defaultCodexModel, // the CLI's default model id; ChatGPT has no llama
	"gpt-5":              defaultCodexModel,
	"gpt-5.1":            defaultCodexModel,
	"gpt-5.2":            defaultCodexModel,
	"gpt-5.3":            defaultCodexModel,
	"gpt-5-codex":        defaultCodexModel,
	"gpt-5.1-codex":      defaultCodexModel,
	"gpt-5.1-codex-mini": defaultCodexModel,
	"codex-mini-latest":  defaultCodexModel,
}

// resolveModel maps a requested model to one the ChatGPT-subscription /responses backend accepts,
// remapping known-rejected ids and defaulting an empty request; unknown ids pass through so models
// OpenAI enables later work without a code change (the backend validates and reports the rest).
func resolveModel(m string) string {
	if r, ok := legacyModels[m]; ok {
		return r
	}
	if m == "" {
		return defaultCodexModel
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
	Type    string             `json:"type"` // "message" | "function_call" | "function_call_output"
	Role    string             `json:"role,omitempty"`
	Content []responsesContent `json:"content,omitempty"`
	// function_call (assistant tool request) / function_call_output (tool result), correlated by CallID.
	CallID    string `json:"call_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
	Output    string `json:"output,omitempty"`
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
		switch m.Role {
		case "system":
			if text := plainText(m); text != "" {
				instr += "\n\n" + text
			}
		case "assistant":
			// Assistant text (if any) precedes its tool calls, each mapped to a function_call item so
			// the backend sees the request it must pair with the following function_call_output.
			if text := plainText(m); text != "" {
				input = append(input, responsesItem{Type: "message", Role: "assistant",
					Content: []responsesContent{{Type: "output_text", Text: text}}})
			}
			for _, p := range m.Parts {
				if p.Type == "tool_call" {
					args := string(p.ToolInput)
					if strings.TrimSpace(args) == "" {
						args = "{}"
					}
					input = append(input, responsesItem{Type: "function_call",
						CallID: p.ToolCallID, Name: p.ToolName, Arguments: args})
				}
			}
		case "tool":
			// Tool result → function_call_output correlated back to the call_id (the Responses API's
			// equivalent of a role:"tool" message; folding it into user text loses the linkage).
			for _, p := range m.Parts {
				if p.Type == "tool_result" {
					input = append(input, responsesItem{Type: "function_call_output",
						CallID: p.ToolCallID, Output: p.Text})
				}
			}
		default: // user
			input = append(input, responsesItem{Type: "message", Role: "user",
				Content: []responsesContent{{Type: "input_text", Text: plainText(m)}}})
		}
	}
	tools := make([]responsesTool, 0, len(req.Tools))
	for _, t := range req.Tools {
		tools = append(tools, responsesTool{Type: "function", Name: t.Name, Description: t.Description, Parameters: json.RawMessage(t.InputSchema)})
	}
	return responsesRequest{
		Model: resolveModel(req.Model), Instructions: instr, Input: input, Tools: tools,
		// The Codex backend is streaming-only ("Stream must be set to true"); Chat consumes the
		// stream and assembles the full result, while ChatStream forwards it live.
		Store: false, Stream: true, Include: []string{"reasoning.encrypted_content"},
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

// Chat performs one inference against the Codex backend. The backend is streaming-only, so Chat
// consumes the SSE stream and assembles the complete response (text, tool calls, usage).
func (a *Adapter) Chat(ctx context.Context, req ports.ChatRequest) (ports.ChatResponse, error) {
	st, err := a.ChatStream(ctx, req)
	if err != nil {
		return ports.ChatResponse{}, err
	}
	defer func() { _ = st.Close() }()
	var text strings.Builder
	var calls []ports.ToolCall
	var usage ports.Usage
	for {
		ev, err := st.Next(ctx)
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			return ports.ChatResponse{}, err
		}
		switch ev.Kind {
		case "content":
			text.WriteString(ev.ContentDelta)
		case "tool_call":
			if ev.ToolCall != nil {
				calls = append(calls, *ev.ToolCall)
			}
		default: // "usage" | "terminal"
			if ev.Usage != nil {
				usage = *ev.Usage
			}
		}
	}
	return ports.ChatResponse{
		Message:   ports.Message{Role: "assistant", Parts: []ports.ContentPart{{Type: "text", Text: text.String()}}},
		ToolCalls: calls,
		Usage:     usage,
	}, nil
}

// ChatStream opens the streaming /responses connection and parses its Server-Sent Events into
// ChatEvents (output_text deltas, completed function calls, and terminal usage).
func (a *Adapter) ChatStream(ctx context.Context, req ports.ChatRequest) (ports.Stream[ports.ChatEvent], error) {
	if a.token == nil {
		return nil, provErr("E-PROV-009", "not signed in to ChatGPT — run: andromeda auth login openai-chatgpt")
	}
	access, account, err := a.token(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(a.buildRequest(req))
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/responses", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Authorization", "Bearer "+access)
	httpReq.Header.Set("chatgpt-account-id", account)
	httpReq.Header.Set("OpenAI-Beta", "responses=experimental")
	httpReq.Header.Set("originator", "codex_cli_rs")

	resp, err := a.hc.Do(httpReq)
	if err != nil {
		return nil, provErr("E-PROV-001", "ChatGPT backend request failed: "+err.Error())
	}
	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()
		return nil, provErr("E-PROV-009", "ChatGPT session rejected (401) — sign in again: andromeda auth login openai-chatgpt")
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		_ = resp.Body.Close()
		return nil, provErr("E-PROV-001", fmt.Sprintf("ChatGPT backend error (%d): %s", resp.StatusCode, snippet(body)))
	}
	return newResponsesStream(resp.Body), nil
}

// responsesStreamEvent is one Server-Sent Event from the Responses API, keyed by its `type`.
type responsesStreamEvent struct {
	Type     string           `json:"type"`
	Delta    string           `json:"delta"` // output_text.delta
	Item     *responsesOutput `json:"item"`  // output_item.done (function_call)
	Response *struct {
		Usage *struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	} `json:"response"` // response.completed
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// responsesStream parses the Responses API SSE into ChatEvents.
type responsesStream struct {
	body       io.ReadCloser
	sc         *bufio.Scanner
	mu         sync.Mutex
	usage      ports.Usage
	terminated bool
	closed     bool
}

func newResponsesStream(body io.ReadCloser) *responsesStream {
	sc := bufio.NewScanner(body)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024) // reasoning payloads can be large
	return &responsesStream{body: body, sc: sc}
}

func (s *responsesStream) Next(ctx context.Context) (ports.ChatEvent, error) {
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
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue // skip SSE "event:" lines and blank separators
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			return s.terminal(), nil
		}
		var ev responsesStreamEvent
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			continue // tolerate unknown/partial event shapes
		}
		switch ev.Type {
		case "response.output_text.delta":
			if ev.Delta != "" {
				return ports.ChatEvent{Kind: "content", ContentDelta: ev.Delta}, nil
			}
		case "response.output_item.done":
			if ev.Item != nil && ev.Item.Type == "function_call" {
				return ports.ChatEvent{Kind: "tool_call", ToolCall: &ports.ToolCall{
					ID: ev.Item.CallID, Name: ev.Item.Name, Input: ports.JSON(ev.Item.Arguments),
				}}, nil
			}
		case "response.completed":
			if ev.Response != nil && ev.Response.Usage != nil {
				s.usage = ports.Usage{InputTokens: ev.Response.Usage.InputTokens, OutputTokens: ev.Response.Usage.OutputTokens, CostBasis: "reported"}
			}
			return s.terminal(), nil
		case "response.failed", "error":
			msg := "ChatGPT backend stream error"
			if ev.Error != nil && ev.Error.Message != "" {
				msg += ": " + ev.Error.Message
			}
			return ports.ChatEvent{}, provErr("E-PROV-001", msg)
		}
	}
	if err := s.sc.Err(); err != nil {
		return ports.ChatEvent{}, provErr("E-PROV-001", "ChatGPT stream read error")
	}
	return s.terminal(), nil
}

func (s *responsesStream) terminal() ports.ChatEvent {
	s.terminated = true
	usage := s.usage
	return ports.ChatEvent{Kind: "terminal", Terminal: true, Usage: &usage}
}

func (s *responsesStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.body.Close()
}

// Embed is unsupported: the ChatGPT backend exposes no embeddings endpoint here.
func (a *Adapter) Embed(context.Context, ports.EmbedRequest) (ports.EmbedResponse, error) {
	return ports.EmbedResponse{}, provErr("E-PROV-002", "the ChatGPT provider does not support embeddings")
}

// DiscoverModels reports the models the ChatGPT-subscription /responses backend accepts. The
// subscription backend has no models endpoint, so this list is curated and verified against the live
// endpoint (2026-07-13): it serves ONLY gpt-5.5 and gpt-5.4. gpt-5.6 exists on the public API but is
// rejected here ("not supported when using Codex with a ChatGPT account"), as are 5.3/older and every
// codex-suffixed variant. Re-verify and extend as OpenAI enables newer models on this backend.
func (a *Adapter) DiscoverModels(context.Context) ([]ports.ModelDescriptor, error) {
	ids := []string{"gpt-5.5", "gpt-5.4"}
	out := make([]ports.ModelDescriptor, 0, len(ids))
	for _, id := range ids {
		out = append(out, ports.ModelDescriptor{ID: id, DisplayName: id})
	}
	return out, nil
}

// Capabilities is unreported for the ChatGPT backend, which exposes no capabilities endpoint here.
func (a *Adapter) Capabilities(context.Context, string) (ports.CapabilitySet, error) {
	return ports.CapabilitySet{}, nil
}

// CountTokens is unsupported: the ChatGPT subscription backend exposes no token-counting endpoint.
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
