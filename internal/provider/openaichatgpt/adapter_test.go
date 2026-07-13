package openaichatgpt

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

// The Codex/ChatGPT-subscription backend serves only gpt-5.5 and gpt-5.4 (verified live): supported
// ids pass through, rejected ids (retired bases, dropped codex variants, the CLI's llama3 default)
// remap to the default, and newer-but-not-yet-enabled ids (gpt-5.6) pass through so the backend's own
// "not supported … with a ChatGPT account" error surfaces instead of a silent downgrade.
func TestResolveModel(t *testing.T) {
	cases := map[string]string{
		"gpt-5.5":       "gpt-5.5", // supported — unchanged
		"gpt-5.4":       "gpt-5.4", // supported — unchanged
		"gpt-5.6":       "gpt-5.6", // newer, not yet on this backend — passes through to the clear 400
		"gpt-5.1-codex": "gpt-5.5", // codex variant dropped by the backend → default
		"gpt-5.2":       "gpt-5.5", // retired → default
		"gpt-5-codex":   "gpt-5.5", // legacy alias → default
		"llama3":        "gpt-5.5", // the CLI default id → default
		"":              "gpt-5.5", // empty → default
	}
	for in, want := range cases {
		if got := resolveModel(in); got != want {
			t.Errorf("resolveModel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestChatBuildsResponsesRequestAndParsesOutput(t *testing.T) {
	var gotPath string
	var gotHeaders http.Header
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotHeaders = r.Header
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		// The Codex backend is streaming-only: respond with Responses-API SSE events.
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"hello there\"}\n\n")
		_, _ = io.WriteString(w, "data: {\"type\":\"response.output_item.done\",\"item\":{\"type\":\"function_call\",\"name\":\"fs_read\",\"arguments\":\"{\\\"path\\\":\\\"x\\\"}\",\"call_id\":\"call_1\"}}\n\n")
		_, _ = io.WriteString(w, "data: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"input_tokens\":11,\"output_tokens\":5}}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	a := New(Config{
		BaseURL: srv.URL,
		Token:   func(context.Context) (string, string, error) { return "tok_abc", "acc_9", nil },
	})
	resp, err := a.Chat(context.Background(), ports.ChatRequest{
		Model: "gpt-5-codex", // legacy → should be remapped
		Messages: []ports.Message{
			{Role: "system", Parts: []ports.ContentPart{{Type: "text", Text: "be terse"}}},
			{Role: "user", Parts: []ports.ContentPart{{Type: "text", Text: "hi"}}},
		},
		Tools: []ports.ToolDeclaration{{Name: "fs_read", Description: "read", InputSchema: ports.JSON(`{"type":"object"}`)}},
	})
	if err != nil {
		t.Fatal(err)
	}

	// endpoint + headers
	if gotPath != "/responses" {
		t.Errorf("path = %q, want /responses", gotPath)
	}
	if got := gotHeaders.Get("Authorization"); got != "Bearer tok_abc" {
		t.Errorf("Authorization = %q", got)
	}
	if got := gotHeaders.Get("chatgpt-account-id"); got != "acc_9" {
		t.Errorf("account header = %q", got)
	}
	if got := gotHeaders.Get("OpenAI-Beta"); got != "responses=experimental" {
		t.Errorf("OpenAI-Beta = %q", got)
	}

	// body contract
	if store, _ := gotBody["store"].(bool); store {
		t.Error("store must be false for the ChatGPT backend")
	}
	if stream, _ := gotBody["stream"].(bool); !stream {
		t.Error("stream must be true — the Codex backend is streaming-only")
	}
	if gotBody["model"] != "gpt-5.5" {
		t.Errorf("model = %v, want remapped gpt-5.5", gotBody["model"])
	}
	if instr, _ := gotBody["instructions"].(string); !strings.Contains(instr, "Codex") || !strings.Contains(instr, "be terse") {
		t.Errorf("instructions should carry the Codex prompt + folded system message: %q", instr)
	}
	if input, ok := gotBody["input"].([]any); !ok || len(input) != 1 {
		t.Errorf("input items = %v (want 1 user item; system folds into instructions)", gotBody["input"])
	}
	if _, ok := gotBody["tools"].([]any); !ok {
		t.Error("tools should be forwarded in Responses format")
	}

	// response parsing
	if got := plainText(resp.Message); got != "hello there" {
		t.Errorf("assistant text = %q", got)
	}
	if len(resp.ToolCalls) != 1 || resp.ToolCalls[0].Name != "fs_read" || resp.ToolCalls[0].ID != "call_1" {
		t.Errorf("tool calls = %+v", resp.ToolCalls)
	}
	if resp.Usage.InputTokens != 11 || resp.Usage.OutputTokens != 5 {
		t.Errorf("usage = %+v", resp.Usage)
	}
}

// ChatStream parses Responses-API SSE into live content deltas, a completed tool call, and a
// terminal event with usage. Regression: the backend requires stream:true, so a non-streaming
// request 400'd ("Stream must be set to true").
func TestChatStreamParsesResponsesSSE(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"Ho\"}\n\n")
		_, _ = io.WriteString(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"la\"}\n\n")
		_, _ = io.WriteString(w, "data: {\"type\":\"response.output_item.done\",\"item\":{\"type\":\"function_call\",\"name\":\"fs_write\",\"arguments\":\"{}\",\"call_id\":\"c1\"}}\n\n")
		_, _ = io.WriteString(w, "data: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"input_tokens\":3,\"output_tokens\":2}}}\n\n")
	}))
	defer srv.Close()

	a := New(Config{BaseURL: srv.URL, Token: func(context.Context) (string, string, error) { return "t", "acc", nil }})
	st, err := a.ChatStream(context.Background(), ports.ChatRequest{Model: "gpt-5.5", Messages: []ports.Message{{Role: "user", Parts: []ports.ContentPart{{Type: "text", Text: "hi"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	var content strings.Builder
	var calls, terminals int
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
		case "tool_call":
			if ev.ToolCall != nil && ev.ToolCall.Name == "fs_write" {
				calls++
			}
		case "terminal":
			terminals++
			if ev.Usage == nil || ev.Usage.OutputTokens != 2 {
				t.Errorf("terminal usage = %+v", ev.Usage)
			}
		}
	}
	if content.String() != "Hola" {
		t.Errorf("streamed content = %q, want Hola", content.String())
	}
	if calls != 1 {
		t.Errorf("tool calls = %d, want 1", calls)
	}
	if terminals != 1 {
		t.Errorf("terminal events = %d, want 1", terminals)
	}
}

func TestChatWithoutTokenSourceIsActionable(t *testing.T) {
	a := New(Config{})
	_, err := a.Chat(context.Background(), ports.ChatRequest{})
	if err == nil || !strings.Contains(err.Error(), "auth login openai-chatgpt") {
		t.Errorf("error = %v, want a sign-in hint", err)
	}
}
