package openaicompat

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/provider"
)

func msg(role, text string) ports.Message {
	return ports.Message{Role: role, Parts: []ports.ContentPart{{Type: "text", Text: text}}}
}

// A multi-turn tool exchange must serialize the assistant's tool_calls and the tool result's
// tool_call_id. Regression for E-PROV-006 "messages.N.tool.tool_call_id": buildMessages dropped
// tool-call linkage, so Groq/Cerebras rejected the second request the moment the agent used a tool.
func TestBuildMessagesSerializesToolExchange(t *testing.T) {
	msgs := []ports.Message{
		msg("user", "make a file"),
		{Role: "assistant", Parts: []ports.ContentPart{
			{Type: "text", Text: "sure"},
			{Type: "tool_call", ToolCallID: "call_1", ToolName: "write_file", ToolInput: ports.JSON(`{"path":"a.md"}`)},
		}},
		{Role: "tool", Parts: []ports.ContentPart{
			{Type: "tool_result", ToolCallID: "call_1", Text: "wrote a.md"},
		}},
	}
	wire := buildMessages(msgs)
	if len(wire) != 3 {
		t.Fatalf("wire messages = %d, want 3", len(wire))
	}
	asst := wire[1]
	if len(asst.ToolCalls) != 1 || asst.ToolCalls[0].ID != "call_1" || asst.ToolCalls[0].Type != "function" {
		t.Fatalf("assistant tool_calls = %+v", asst.ToolCalls)
	}
	if asst.ToolCalls[0].Function.Name != "write_file" || asst.ToolCalls[0].Function.Arguments != `{"path":"a.md"}` {
		t.Fatalf("assistant tool_call function = %+v", asst.ToolCalls[0].Function)
	}
	if tool := wire[2]; tool.Role != "tool" || tool.ToolCallID != "call_1" || tool.Content != "wrote a.md" {
		t.Fatalf("tool message = %+v", tool)
	}
}

// Empty tool arguments serialize as "{}" (the API requires a JSON string, never "").
func TestBuildMessagesEmptyToolArgs(t *testing.T) {
	wire := buildMessages([]ports.Message{{Role: "assistant", Parts: []ports.ContentPart{
		{Type: "tool_call", ToolCallID: "c", ToolName: "list", ToolInput: nil},
	}}})
	if got := wire[0].ToolCalls[0].Function.Arguments; got != "{}" {
		t.Fatalf("empty args = %q, want {}", got)
	}
}

// Tool declarations must carry their input schema as parameters so the model produces valid args.
func TestBuildRequestIncludesToolParameters(t *testing.T) {
	cr := buildRequest(ports.ChatRequest{
		Model: "m",
		Tools: []ports.ToolDeclaration{{Name: "write_file", Description: "d",
			InputSchema: ports.JSON(`{"type":"object","properties":{"path":{"type":"string"}}}`)}},
	}, false)
	if len(cr.Tools) != 1 || cr.Tools[0].Function.Parameters["type"] != "object" {
		t.Fatalf("tool parameters = %+v", cr.Tools)
	}
}

func TestChatParsesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer k" {
			t.Errorf("missing auth header")
		}
		var req chatReq
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "gpt-x" || len(req.Messages) != 1 {
			t.Errorf("request = %+v", req)
		}
		io.WriteString(w, `{"choices":[{"message":{"content":"hi there"}}],"usage":{"prompt_tokens":5,"completion_tokens":2}}`)
	}))
	defer srv.Close()

	a := New(Config{BaseURL: srv.URL, APIKey: "k", Client: &provider.Client{HTTP: srv.Client()}})
	resp, err := a.Chat(context.Background(), ports.ChatRequest{Model: "gpt-x", Messages: []ports.Message{msg("user", "hello")}})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Message.Parts[0].Text != "hi there" {
		t.Errorf("content = %q", resp.Message.Parts[0].Text)
	}
	if resp.Usage.InputTokens != 5 || resp.Usage.OutputTokens != 2 {
		t.Errorf("usage = %+v", resp.Usage)
	}
}

func TestChatStreamParsesSSE(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hel\"}}]}\n\n")
		io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"lo\"}}]}\n\n")
		io.WriteString(w, "data: {\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":1}}\n\n")
		io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	a := New(Config{BaseURL: srv.URL, Client: &provider.Client{HTTP: srv.Client()}})
	st, err := a.ChatStream(context.Background(), ports.ChatRequest{Model: "m", Messages: []ports.Message{msg("user", "hi")}})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	var content strings.Builder
	var sawTerminal bool
	for {
		ev, err := st.Next(context.Background())
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		switch ev.Kind {
		case "content":
			content.WriteString(ev.ContentDelta)
		case "terminal":
			sawTerminal = true
			if ev.Usage == nil || ev.Usage.OutputTokens != 1 {
				t.Errorf("terminal usage = %+v", ev.Usage)
			}
		}
	}
	if content.String() != "Hello" {
		t.Errorf("streamed content = %q", content.String())
	}
	if !sawTerminal {
		t.Error("expected a terminal event")
	}
}

func TestAuthErrorMapsToE_PROV_002(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":"bad key"}`)
	}))
	defer srv.Close()
	a := New(Config{BaseURL: srv.URL, Client: &provider.Client{HTTP: srv.Client()}})
	_, err := a.Chat(context.Background(), ports.ChatRequest{Model: "m"})
	pe, ok := err.(*ports.PortError)
	if !ok || pe.Code != provider.CodeAuth || pe.Retryable {
		t.Fatalf("want non-retryable E-PROV-002, got %v", err)
	}
}

// A streamed tool call arrives as fragmented deltas (id/name first, arguments across chunks). The
// parser must reassemble them and emit a tool_call event before the terminal. Regression: streaming
// dropped tool calls entirely, so a tool-using turn could not work while streaming.
func TestChatStreamAccumulatesToolCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","function":{"name":"fs_write","arguments":""}}]}}]}`+"\n\n")
		io.WriteString(w, `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"path\":"}}]}}]}`+"\n\n")
		io.WriteString(w, `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"a.md\"}"}}]}}]}`+"\n\n")
		io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	a := New(Config{BaseURL: srv.URL, Client: &provider.Client{HTTP: srv.Client()}})
	st, err := a.ChatStream(context.Background(), ports.ChatRequest{Model: "m", Messages: []ports.Message{msg("user", "hi")}})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	var calls []ports.ToolCall
	for {
		ev, err := st.Next(context.Background())
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if ev.Kind == "tool_call" && ev.ToolCall != nil {
			calls = append(calls, *ev.ToolCall)
		}
	}
	if len(calls) != 1 {
		t.Fatalf("expected 1 reassembled tool call, got %d", len(calls))
	}
	if calls[0].ID != "call_1" || calls[0].Name != "fs_write" || string(calls[0].Input) != `{"path":"a.md"}` {
		t.Fatalf("reassembled tool call = %+v (args %q)", calls[0], calls[0].Input)
	}
}

func TestCapabilitiesAndCountTokens(t *testing.T) {
	a := New(Config{BaseURL: "http://x"})
	caps, _ := a.Capabilities(context.Background(), "m")
	if !caps.Has("tool_calling") {
		t.Error("expected tool_calling capability")
	}
	if _, err := a.CountTokens(context.Background(), ports.TokenCountRequest{}); err == nil {
		t.Error("CountTokens should be unavailable")
	}
}

// reasoning_effort from ModelParams.Extra is forwarded in the request body (reasoning models),
// and omitted entirely when unset so ordinary models are unaffected.
func TestBuildRequestReasoningEffort(t *testing.T) {
	with := buildRequest(ports.ChatRequest{
		Model:    "o4-mini",
		Messages: []ports.Message{{Role: "user", Parts: []ports.ContentPart{{Type: "text", Text: "hi"}}}},
		Params:   ports.ModelParams{Extra: map[string]any{"reasoning_effort": "high"}},
	}, false)
	if with.ReasoningEffort != "high" {
		t.Fatalf("reasoning_effort = %q, want high", with.ReasoningEffort)
	}
	data, _ := json.Marshal(with)
	if !strings.Contains(string(data), `"reasoning_effort":"high"`) {
		t.Errorf("body missing reasoning_effort: %s", data)
	}
	without := buildRequest(ports.ChatRequest{Model: "llama3"}, false)
	data, _ = json.Marshal(without)
	if strings.Contains(string(data), "reasoning_effort") {
		t.Errorf("reasoning_effort must be omitted when unset: %s", data)
	}
}
