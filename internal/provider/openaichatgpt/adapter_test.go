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

func TestChatBuildsResponsesRequestAndParsesOutput(t *testing.T) {
	var gotPath string
	var gotHeaders http.Header
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotHeaders = r.Header
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"output": [
				{"type":"message","role":"assistant","content":[{"type":"output_text","text":"hello there"}]},
				{"type":"function_call","name":"fs_read","arguments":"{\"path\":\"x\"}","call_id":"call_1"}
			],
			"usage": {"input_tokens": 11, "output_tokens": 5}
		}`))
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
	if gotBody["model"] != "gpt-5.1-codex" {
		t.Errorf("model = %v, want remapped gpt-5.1-codex", gotBody["model"])
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

func TestChatWithoutTokenSourceIsActionable(t *testing.T) {
	a := New(Config{})
	_, err := a.Chat(context.Background(), ports.ChatRequest{})
	if err == nil || !strings.Contains(err.Error(), "auth login openai-chatgpt") {
		t.Errorf("error = %v, want a sign-in hint", err)
	}
}
